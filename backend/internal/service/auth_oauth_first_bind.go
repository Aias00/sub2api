package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dbent "github.com/Aias00/cloudbase/ent"
	billingctx "github.com/Aias00/cloudbase/internal/billing"
	"github.com/Aias00/cloudbase/internal/pkg/logger"

	entsql "entgo.io/ent/dialect/sql"
)

// ApplyProviderDefaultSettingsOnFirstBind applies provider-specific bootstrap
// settings the first time a user binds a third-party identity. The grant is
// idempotent per user/provider pair.
//
// email is the user's bound email, passed in by the caller so the signup-grant
// risk control can key on it WITHOUT a GetByID lookup. That lookup is forbidden
// here because this method may run inside an ent.Tx (e.g. BindEmailIdentity →
// updateBoundEmailIdentityTx leaves the tx in ctx): userRepository.GetByID uses
// the root ent client and a raw *sql.DB, neither of which is tx-aware, so on
// SQLite shared-cache it deadlocks the connection the tx already holds (and on
// Postgres it reads outside the tx — an isolation hazard). Every call site
// already has the email in hand, so we thread it through instead.
//
// emailVerified tells the signup-grant risk control's RequireVerifiedEmail rule
// whether to allow the grant. Every first-bind path only fires after the email
// has been authenticated (verify code, OAuth verified email, or password login),
// so callers pass true; the rule then passes and the configured limits run.
func (s *AuthService) ApplyProviderDefaultSettingsOnFirstBind(
	ctx context.Context,
	userID int64,
	providerType string,
	email string,
	emailVerified bool,
) error {
	if s == nil || s.entClient == nil || s.settingService == nil || userID <= 0 {
		return nil
	}

	if dbent.TxFromContext(ctx) != nil {
		return s.applyProviderDefaultSettingsOnFirstBind(ctx, userID, providerType, email, emailVerified)
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin first bind defaults transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	if err := s.applyProviderDefaultSettingsOnFirstBind(txCtx, userID, providerType, email, emailVerified); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *AuthService) applyProviderDefaultSettingsOnFirstBind(
	ctx context.Context,
	userID int64,
	providerType string,
	email string,
	emailVerified bool,
) error {
	providerDefaults, enabled, err := s.settingService.ResolveAuthSourceGrantSettings(ctx, providerType, true)
	if err != nil {
		return fmt.Errorf("load auth source defaults: %w", err)
	}
	if !enabled {
		return nil
	}

	client := s.entClient
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}

	var result entsql.Result
	if err := client.Driver().Exec(
		ctx,
		`INSERT INTO user_provider_default_grants (user_id, provider_type, grant_reason)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, provider_type, grant_reason) DO NOTHING`,
		[]any{userID, strings.TrimSpace(providerType), "first_bind"},
		&result,
	); err != nil {
		return fmt.Errorf("record first bind provider grant: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read first bind provider grant result: %w", err)
	}
	if affected == 0 {
		return nil
	}

	// 首绑余额发放前接入注册风控：复用 signup_grant_claims 限额表与 override 名单，
	// force=true 绕过来源白名单（linuxdo/wechat 等非白名单来源仍跑全部限额）。
	// 命中风控时跳过余额/gift 发放，但保留上面的去重行（user_provider_default_grants），
	// 使被封身份换 IP 也无法再触发首绑奖励。concurrency/subscription 不受影响（滥用向量是余额套现）。
	if providerDefaults.Balance != 0 {
		// email 由调用方传入（见上 ApplyProviderDefaultSettingsOnFirstBind 注释），
		// 不在此处 GetByID —— 那会触发 ent.Tx 内的死锁。
		// emailVerified 同样由调用方传入并合入 risk input，供 RequireVerifiedEmail 规则判定；
		// 与注册路径一致（auth_oauth_email_flow.go 等），首绑只对已认证邮箱放行。
		riskCtx := WithSignupGrantRiskInput(ctx, mergeSignupGrantRiskInput(signupGrantRiskInputFromContext(ctx), SignupGrantRiskInput{
			EmailVerified: signupGrantEmailVerified(emailVerified),
		}))
		_, claim := s.applySignupGrantRiskControlEx(riskCtx, email, providerType, signupGrantPlan{Balance: providerDefaults.Balance}, true /*force*/)
		if claim != nil {
			// markGift=false：首绑余额由下方 ApplyBalanceChangeCtx（ADD + GiftDelta）追加，
			// 已自行处理 gift 组件；若此处再 SET 标记 gift 会导致 gift_balance 双记。
			// 仅做 claim→user 审计关联，gift 组件交给 ApplyBalanceChangeCtx。
			s.attachSignupGrantClaim(ctx, claim, userID, false)
			if claim.Blocked {
				logger.LegacyPrintf("service.auth", "[FirstBindGrant] risk-blocked user=%d provider=%s reason=%s", userID, providerType, claim.Reason)
				providerDefaults.Balance = 0
			}
		}
	}

	if providerDefaults.Balance != 0 {
		if s.ledgerService != nil {
			// Participates in the surrounding ent tx (present in ctx here) so the
			// balance + gift grant and its ledger row commit atomically with the
			// provider-grant record above.
			if _, err := s.ledgerService.ApplyBalanceChangeCtx(ctx, billingctx.BalanceChangeCommand{
				UserID:        userID,
				Delta:         providerDefaults.Balance,
				GiftDelta:     providerDefaults.Balance,
				EntryType:     billingctx.EntryTypeOAuthBindBonus,
				SourceType:    billingctx.SourceTypeOAuthBinding,
				Description:   fmt.Sprintf("first bind grant (%s)", strings.TrimSpace(providerType)),
				AllowNegative: true,
			}); err != nil {
				if errors.Is(err, billingctx.ErrLedgerUserNotFound) {
					return ErrUserNotFound
				}
				return fmt.Errorf("apply first bind balance default: %w", err)
			}
		} else {
			var balanceResult entsql.Result
			if err := client.Driver().Exec(ctx, `
			UPDATE users
			SET balance = balance + $1,
				gift_balance = gift_balance + $1,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND deleted_at IS NULL
		`, []any{providerDefaults.Balance, userID}, &balanceResult); err != nil {
				return fmt.Errorf("apply first bind balance default: %w", err)
			}
			affected, err := balanceResult.RowsAffected()
			if err != nil {
				return fmt.Errorf("read first bind balance default result: %w", err)
			}
			if affected == 0 {
				return ErrUserNotFound
			}
		}
	}
	if providerDefaults.Concurrency != 0 {
		if err := client.User.UpdateOneID(userID).AddConcurrency(providerDefaults.Concurrency).Exec(ctx); err != nil {
			return fmt.Errorf("apply first bind concurrency default: %w", err)
		}
	}
	if s.defaultSubAssigner != nil {
		for _, item := range providerDefaults.Subscriptions {
			if _, _, err := s.defaultSubAssigner.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
				UserID:       userID,
				GroupID:      item.GroupID,
				ValidityDays: item.ValidityDays,
				Notes:        "auto assigned by first bind defaults",
			}); err != nil {
				return fmt.Errorf("apply first bind subscription default: %w", err)
			}
		}
	}

	return nil
}
