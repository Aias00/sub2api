package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	dbent "github.com/Aias00/cloudbase/ent"
	billingctx "github.com/Aias00/cloudbase/internal/billing"
	"github.com/Aias00/cloudbase/internal/domain"
	"github.com/Aias00/cloudbase/internal/gateway"
	"github.com/Aias00/cloudbase/internal/identity"
	"github.com/Aias00/cloudbase/internal/pkg/logger"
	"github.com/Aias00/cloudbase/internal/pkg/timezone"
	"github.com/Aias00/cloudbase/internal/service"
)

type usageBillingRepository struct {
	db *sql.DB
}

func NewUsageBillingRepository(_ *dbent.Client, sqlDB *sql.DB) billingctx.UsageBillingRepository {
	return &usageBillingRepository{db: sqlDB}
}

func (r *usageBillingRepository) Apply(ctx context.Context, cmd *billingctx.UsageBillingCommand) (_ *billingctx.UsageBillingApplyResult, err error) {
	if cmd == nil {
		return &billingctx.UsageBillingApplyResult{}, nil
	}
	if r == nil || r.db == nil {
		return nil, errors.New("usage billing repository db is nil")
	}

	cmd.Normalize()
	if cmd.RequestID == "" {
		return nil, billingctx.ErrUsageBillingRequestIDRequired
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	applied, err := r.claimUsageBillingKey(ctx, tx, cmd)
	if err != nil {
		return nil, err
	}
	if !applied {
		return &billingctx.UsageBillingApplyResult{Applied: false}, nil
	}

	result := &billingctx.UsageBillingApplyResult{Applied: true}
	if err := r.applyUsageBillingEffects(ctx, tx, cmd, result); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	return result, nil
}

func (r *usageBillingRepository) claimUsageBillingKey(ctx context.Context, tx *sql.Tx, cmd *billingctx.UsageBillingCommand) (bool, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO usage_billing_dedup (request_id, api_key_id, request_fingerprint)
		VALUES ($1, $2, $3)
		ON CONFLICT (request_id, api_key_id) DO NOTHING
		RETURNING id
	`, cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		var existingFingerprint string
		if err := tx.QueryRowContext(ctx, `
			SELECT request_fingerprint
			FROM usage_billing_dedup
			WHERE request_id = $1 AND api_key_id = $2
		`, cmd.RequestID, cmd.APIKeyID).Scan(&existingFingerprint); err != nil {
			return false, err
		}
		if strings.TrimSpace(existingFingerprint) != strings.TrimSpace(cmd.RequestFingerprint) {
			return false, billingctx.ErrUsageBillingRequestConflict
		}
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var archivedFingerprint string
	err = tx.QueryRowContext(ctx, `
		SELECT request_fingerprint
		FROM usage_billing_dedup_archive
		WHERE request_id = $1 AND api_key_id = $2
	`, cmd.RequestID, cmd.APIKeyID).Scan(&archivedFingerprint)
	if err == nil {
		if strings.TrimSpace(archivedFingerprint) != strings.TrimSpace(cmd.RequestFingerprint) {
			return false, billingctx.ErrUsageBillingRequestConflict
		}
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return true, nil
}

func (r *usageBillingRepository) applyUsageBillingEffects(ctx context.Context, tx *sql.Tx, cmd *billingctx.UsageBillingCommand, result *billingctx.UsageBillingApplyResult) error {
	if cmd.SubscriptionCost > 0 && cmd.SubscriptionID != nil {
		if err := incrementUsageBillingSubscription(ctx, tx, *cmd.SubscriptionID, cmd.SubscriptionCost); err != nil {
			return err
		}
	}

	if cmd.BalanceCost > 0 {
		newBalance, sufficient, err := deductUsageBillingBalance(ctx, tx, cmd.UserID, cmd.BalanceCost, signupGiftBalanceEligible(cmd))
		if err != nil {
			return err
		}
		result.NewBalance = &newBalance
		result.BalanceOverdrafted = !sufficient
	}

	if cmd.APIKeyQuotaCost > 0 {
		exhausted, err := incrementUsageBillingAPIKeyQuota(ctx, tx, cmd.APIKeyID, cmd.APIKeyQuotaCost)
		if err != nil {
			return err
		}
		result.APIKeyQuotaExhausted = exhausted
	}

	if cmd.APIKeyRateLimitCost > 0 {
		if err := incrementUsageBillingAPIKeyRateLimit(ctx, tx, cmd.APIKeyID, cmd.APIKeyRateLimitCost); err != nil {
			return err
		}
	}

	if cmd.AccountQuotaCost > 0 && (strings.EqualFold(cmd.AccountType, domain.AccountTypeAPIKey) || strings.EqualFold(cmd.AccountType, domain.AccountTypeBedrock)) {
		quotaState, err := incrementUsageBillingAccountQuota(ctx, tx, cmd.AccountID, cmd.AccountQuotaCost)
		if err != nil {
			return err
		}
		result.QuotaState = quotaState
	}

	return nil
}

func incrementUsageBillingSubscription(ctx context.Context, tx *sql.Tx, subscriptionID int64, costUSD float64) error {
	// 1. 锁行读取订阅状态 + 分组限额
	const selectSQL = `
		SELECT us.daily_usage_usd, us.weekly_usage_usd, us.monthly_usage_usd,
		       us.daily_window_start, us.weekly_window_start, us.monthly_window_start,
		       us.starts_at, us.expires_at,
		       g.daily_limit_usd, g.weekly_limit_usd, g.monthly_limit_usd
		FROM user_subscriptions us
		JOIN groups g ON us.group_id = g.id
		WHERE us.id = $1 AND us.deleted_at IS NULL AND g.deleted_at IS NULL
		FOR UPDATE OF us`

	var s subscriptionWindowState
	err := tx.QueryRowContext(ctx, selectSQL, subscriptionID).Scan(
		&s.DailyUsageUSD, &s.WeeklyUsageUSD, &s.MonthlyUsageUSD,
		&s.DailyWindowStart, &s.WeeklyWindowStart, &s.MonthlyWindowStart,
		&s.StartsAt, &s.ExpiresAt,
		&s.DailyLimitUSD, &s.WeeklyLimitUSD, &s.MonthlyLimitUSD,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.ErrSubscriptionNotFound
		}
		return err
	}

	// 1b. 检查订阅是否已过期
	now := time.Now()
	if s.ExpiresAt.Before(now) {
		return service.ErrSubscriptionExpired
	}
	if s.StartsAt.After(now) {
		return service.ErrSubscriptionNotActive
	}

	// 2. 计算窗口重置后的新用量
	oneTimeDailyQuota := isOneTimeDailyQuota(s.StartsAt, s.ExpiresAt)

	newDaily, newDailyWindow := computeWindowedUsage(
		s.DailyUsageUSD, s.DailyWindowStart,
		now, 24*time.Hour,
		!oneTimeDailyQuota, // 一次性日配额不重置
		timezone.StartOfDay(now),
		costUSD,
	)
	newWeekly, newWeeklyWindow := computeWindowedUsage(
		s.WeeklyUsageUSD, s.WeeklyWindowStart,
		now, 7*24*time.Hour,
		true,
		timezone.StartOfWeek(now),
		costUSD,
	)
	newMonthly, newMonthlyWindow := computeWindowedUsage(
		s.MonthlyUsageUSD, s.MonthlyWindowStart,
		now, 30*24*time.Hour,
		true,
		timezone.StartOfDay(now),
		costUSD,
	)

	// 3. 校验限额
	if s.DailyLimitUSD != nil && *s.DailyLimitUSD > 0 && newDaily > *s.DailyLimitUSD {
		return service.ErrDailyLimitExceeded
	}
	if s.WeeklyLimitUSD != nil && *s.WeeklyLimitUSD > 0 && newWeekly > *s.WeeklyLimitUSD {
		return service.ErrWeeklyLimitExceeded
	}
	if s.MonthlyLimitUSD != nil && *s.MonthlyLimitUSD > 0 && newMonthly > *s.MonthlyLimitUSD {
		return service.ErrMonthlyLimitExceeded
	}

	// 4. 写入计算后的新值
	const updateSQL = `
		UPDATE user_subscriptions
		SET daily_usage_usd = $1,
		    weekly_usage_usd = $2,
		    monthly_usage_usd = $3,
		    daily_window_start = $4,
		    weekly_window_start = $5,
		    monthly_window_start = $6,
		    updated_at = NOW()
		WHERE id = $7 AND deleted_at IS NULL`

	res, err := tx.ExecContext(ctx, updateSQL,
		newDaily, newWeekly, newMonthly,
		newDailyWindow, newWeeklyWindow, newMonthlyWindow,
		subscriptionID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrSubscriptionNotFound
	}
	return nil
}

// subscriptionWindowState 保存 SELECT FOR UPDATE 读取的订阅窗口状态和分组限额。
type subscriptionWindowState struct {
	DailyUsageUSD   float64
	WeeklyUsageUSD  float64
	MonthlyUsageUSD float64

	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time

	StartsAt  time.Time
	ExpiresAt time.Time

	DailyLimitUSD   *float64
	WeeklyLimitUSD  *float64
	MonthlyLimitUSD *float64
}

// computeWindowedUsage 判断窗口是否过期并计算新用量。
// 若 needsReset=true 且窗口已过期，用量重置为 costUSD 并更新窗口起始；
// 否则在当前用量上累加 costUSD。
func computeWindowedUsage(
	currentUsage float64, windowStart *time.Time,
	now time.Time, windowDuration time.Duration,
	needsReset bool, resetTarget time.Time,
	costUSD float64,
) (float64, *time.Time) {
	if windowStart == nil {
		// 窗口从未激活，从当前时刻开始
		return costUSD, &resetTarget
	}
	if needsReset && now.Sub(*windowStart) >= windowDuration {
		// 窗口已过期，重置用量并开始新窗口
		return costUSD, &resetTarget
	}
	// 窗口仍在有效期内，累加用量
	return currentUsage + costUSD, windowStart
}

// isOneTimeDailyQuota 判断订阅是否为一次性日配额（有效期 ≤ 24h），
// 此类订阅的日窗口不应重置。
func isOneTimeDailyQuota(startsAt, expiresAt time.Time) bool {
	return !expiresAt.After(startsAt.AddDate(0, 0, 1))
}

func deductUsageBillingBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64, giftEligible bool) (float64, bool, error) {
	var newBalance float64
	if giftEligible {
		err := tx.QueryRowContext(ctx, `
			UPDATE users
			SET balance = balance - $1,
				gift_balance = GREATEST(gift_balance - $1, 0),
				paid_balance = paid_balance - GREATEST($1 - gift_balance, 0),
				updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL AND balance >= $1
			RETURNING balance
		`, amount, userID).Scan(&newBalance)
		if err == nil {
			return newBalance, true, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, false, err
		}

		err = tx.QueryRowContext(ctx, `
				UPDATE users
				SET balance = balance - $1,
					gift_balance = 0,
					paid_balance = balance - $1,
					updated_at = NOW()
				WHERE id = $2 AND deleted_at IS NULL
				RETURNING balance
		`, amount, userID).Scan(&newBalance)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, identity.ErrUserNotFound
		}
		if err != nil {
			return 0, false, err
		}
		return newBalance, false, nil
	}

	err := tx.QueryRowContext(ctx, `
		UPDATE users
		SET balance = balance - $1,
			paid_balance = paid_balance - $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND paid_balance >= $1
		RETURNING balance
	`, amount, userID).Scan(&newBalance)
	if err == nil {
		return newBalance, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE users
		SET balance = balance - $1,
			paid_balance = paid_balance - $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING balance
	`, amount, userID).Scan(&newBalance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, identity.ErrUserNotFound
	}
	if err != nil {
		return 0, false, err
	}
	return newBalance, false, nil
}

func signupGiftBalanceEligible(cmd *billingctx.UsageBillingCommand) bool {
	if cmd == nil || cmd.BalanceCost <= 0 {
		return false
	}
	model := strings.ToLower(strings.TrimSpace(cmd.Model))
	if model == "" {
		return false
	}
	if !signupGiftAllowedModel(model) {
		return false
	}
	if cmd.ImageCount > 1 {
		return false
	}
	if cmd.ImageCount == 1 && !signupGiftAllowedImageSize(cmd.ImageSize) {
		return false
	}
	return true
}

func signupGiftAllowedModel(model string) bool {
	allowed := []string{
		"gpt-5.4-mini",
		"gpt-5.3-codex-spark",
		"gpt-image-2",
	}
	for _, item := range allowed {
		if strings.EqualFold(model, item) {
			return true
		}
	}
	return false
}

func signupGiftAllowedImageSize(raw string) bool {
	size := strings.ToLower(strings.TrimSpace(raw))
	if size == "" {
		return true
	}
	size = strings.ReplaceAll(size, " ", "")
	switch size {
	case "1024x1024", "1024*1024", "1k", "low":
		return true
	default:
		return false
	}
}

func incrementUsageBillingAPIKeyQuota(ctx context.Context, tx *sql.Tx, apiKeyID int64, amount float64) (bool, error) {
	var exhausted bool
	err := tx.QueryRowContext(ctx, `
		UPDATE api_keys
		SET quota_used = quota_used + $1,
			status = CASE
				WHEN quota > 0
					AND status = $3
					AND quota_used < quota
					AND quota_used + $1 >= quota
				THEN $4
				ELSE status
			END,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING quota > 0 AND quota_used >= quota AND quota_used - $1 < quota
	`, amount, apiKeyID, service.StatusAPIKeyActive, service.StatusAPIKeyQuotaExhausted).Scan(&exhausted)
	if errors.Is(err, sql.ErrNoRows) {
		return false, gateway.ErrAPIKeyNotFound
	}
	if err != nil {
		return false, err
	}
	return exhausted, nil
}

func incrementUsageBillingAPIKeyRateLimit(ctx context.Context, tx *sql.Tx, apiKeyID int64, cost float64) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE api_keys SET
			usage_5h = CASE WHEN window_5h_start IS NOT NULL AND window_5h_start + INTERVAL '5 hours' <= NOW() THEN $1 ELSE usage_5h + $1 END,
			usage_1d = CASE WHEN window_1d_start IS NOT NULL AND window_1d_start + INTERVAL '24 hours' <= NOW() THEN $1 ELSE usage_1d + $1 END,
			usage_7d = CASE WHEN window_7d_start IS NOT NULL AND window_7d_start + INTERVAL '7 days' <= NOW() THEN $1 ELSE usage_7d + $1 END,
			window_5h_start = CASE WHEN window_5h_start IS NULL OR window_5h_start + INTERVAL '5 hours' <= NOW() THEN NOW() ELSE window_5h_start END,
			window_1d_start = CASE WHEN window_1d_start IS NULL OR window_1d_start + INTERVAL '24 hours' <= NOW() THEN date_trunc('day', NOW()) ELSE window_1d_start END,
			window_7d_start = CASE WHEN window_7d_start IS NULL OR window_7d_start + INTERVAL '7 days' <= NOW() THEN date_trunc('day', NOW()) ELSE window_7d_start END,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, cost, apiKeyID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return gateway.ErrAPIKeyNotFound
	}
	return nil
}

func incrementUsageBillingAccountQuota(ctx context.Context, tx *sql.Tx, accountID int64, amount float64) (*billingctx.AccountQuotaState, error) {
	rows, err := tx.QueryContext(ctx,
		`UPDATE accounts SET extra = (
			COALESCE(extra, '{}'::jsonb)
			|| jsonb_build_object('quota_used', COALESCE((extra->>'quota_used')::numeric, 0) + $1)
			|| CASE WHEN COALESCE((extra->>'quota_daily_limit')::numeric, 0) > 0 THEN
				jsonb_build_object(
					'quota_daily_used',
					CASE WHEN `+dailyExpiredExpr+`
					THEN $1
					ELSE COALESCE((extra->>'quota_daily_used')::numeric, 0) + $1 END,
					'quota_daily_start',
					CASE WHEN `+dailyExpiredExpr+`
					THEN `+nowUTC+`
					ELSE COALESCE(extra->>'quota_daily_start', `+nowUTC+`) END
				)
				|| CASE WHEN `+dailyExpiredExpr+` AND `+nextDailyResetAtExpr+` IS NOT NULL
				   THEN jsonb_build_object('quota_daily_reset_at', `+nextDailyResetAtExpr+`)
				   ELSE '{}'::jsonb END
			ELSE '{}'::jsonb END
			|| CASE WHEN COALESCE((extra->>'quota_weekly_limit')::numeric, 0) > 0 THEN
				jsonb_build_object(
					'quota_weekly_used',
					CASE WHEN `+weeklyExpiredExpr+`
					THEN $1
					ELSE COALESCE((extra->>'quota_weekly_used')::numeric, 0) + $1 END,
					'quota_weekly_start',
					CASE WHEN `+weeklyExpiredExpr+`
					THEN `+nowUTC+`
					ELSE COALESCE(extra->>'quota_weekly_start', `+nowUTC+`) END
				)
				|| CASE WHEN `+weeklyExpiredExpr+` AND `+nextWeeklyResetAtExpr+` IS NOT NULL
				   THEN jsonb_build_object('quota_weekly_reset_at', `+nextWeeklyResetAtExpr+`)
				   ELSE '{}'::jsonb END
			ELSE '{}'::jsonb END
		), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING
			COALESCE((extra->>'quota_used')::numeric, 0),
			COALESCE((extra->>'quota_limit')::numeric, 0),
			COALESCE((extra->>'quota_daily_used')::numeric, 0),
			COALESCE((extra->>'quota_daily_limit')::numeric, 0),
			COALESCE((extra->>'quota_weekly_used')::numeric, 0),
			COALESCE((extra->>'quota_weekly_limit')::numeric, 0)`,
		amount, accountID)
	if err != nil {
		return nil, err
	}

	var state billingctx.AccountQuotaState
	if rows.Next() {
		if err := rows.Scan(
			&state.TotalUsed, &state.TotalLimit,
			&state.DailyUsed, &state.DailyLimit,
			&state.WeeklyUsed, &state.WeeklyLimit,
		); err != nil {
			_ = rows.Close()
			return nil, err
		}
	} else {
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return nil, err
		}
		_ = rows.Close()
		return nil, gateway.ErrAccountNotFound
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	// 必须在执行下一条 SQL 前显式关闭 rows：pq 驱动在同一连接上
	// 不允许前一条查询的结果集未耗尽时启动新查询，否则会返回
	// "unexpected Parse response" 错误。
	if err := rows.Close(); err != nil {
		return nil, err
	}
	// 任意维度额度在本次递增中从"未超"跨越到"已超"时，必须刷新调度快照，
	// 否则 Redis 中缓存的 Account 仍显示旧的 used 值，后续请求会继续选中本账号，
	// 最终观察到 daily_used / weekly_used 大幅超过配置的 limit。
	// 对于日/周额度，即使本次触发了周期重置（pre=0、post=amount），
	// 判定式 (post-amount) < limit 同样成立，逻辑与总额度保持一致。
	crossedTotal := state.TotalLimit > 0 && state.TotalUsed >= state.TotalLimit && (state.TotalUsed-amount) < state.TotalLimit
	crossedDaily := state.DailyLimit > 0 && state.DailyUsed >= state.DailyLimit && (state.DailyUsed-amount) < state.DailyLimit
	crossedWeekly := state.WeeklyLimit > 0 && state.WeeklyUsed >= state.WeeklyLimit && (state.WeeklyUsed-amount) < state.WeeklyLimit
	if crossedTotal || crossedDaily || crossedWeekly {
		if err := enqueueSchedulerOutbox(ctx, tx, service.SchedulerOutboxEventAccountChanged, &accountID, nil, nil); err != nil {
			logger.LegacyPrintf("repository.usage_billing", "[SchedulerOutbox] enqueue quota exceeded failed: account=%d err=%v", accountID, err)
			return nil, err
		}
	}
	return &state, nil
}
