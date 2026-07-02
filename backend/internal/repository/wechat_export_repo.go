package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/cloudbase/internal/pkg/errors"
	"github.com/Wei-Shaw/cloudbase/internal/pkg/pagination"
	"github.com/Wei-Shaw/cloudbase/internal/service"
)

type wechatExportRepository struct {
	db  *sql.DB
	sql sqlExecutor
}

func NewWeChatExportRepository(db *sql.DB) service.WeChatExportRepository {
	return &wechatExportRepository{db: db, sql: db}
}

type wechatExportTaskPayload struct {
	ArticleIDs []int64                      `json:"article_ids"`
	Formats    []service.WeChatExportFormat `json:"formats"`
}

func (r *wechatExportRepository) GetActiveSession(ctx context.Context, userID int64) (*service.WeChatSession, error) {
	session, err := scanWeChatSession(ctx, r.sql, `
		SELECT id, user_id, status, login_token, cookies_encrypted, login_account_name,
			last_validated_at, expires_at, created_at, updated_at
		FROM wechat_sessions
		WHERE user_id = $1
			AND status IN ($2, $3, $4)
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY CASE WHEN status = $4 THEN 0 ELSE 1 END, updated_at DESC, id DESC
		LIMIT 1
	`, userID, service.WeChatSessionStatusPending, service.WeChatSessionStatusScanConfirmed, service.WeChatSessionStatusReady)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return session, err
}

func (r *wechatExportRepository) CreateSession(ctx context.Context, session *service.WeChatSession) error {
	if session == nil {
		return nil
	}
	return scanSingleRow(ctx, r.sql, `
		INSERT INTO wechat_sessions (user_id, status, login_token, cookies_encrypted, login_account_name, last_validated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, []any{
		session.UserID,
		session.Status,
		session.LoginToken,
		session.CookiesEncrypted,
		session.LoginAccountName,
		session.LastValidatedAt,
		session.ExpiresAt,
	}, &session.ID, &session.CreatedAt, &session.UpdatedAt)
}

func (r *wechatExportRepository) UpdateSession(ctx context.Context, session *service.WeChatSession) error {
	if session == nil {
		return nil
	}
	updated, err := scanWeChatSession(ctx, r.sql, `
		UPDATE wechat_sessions
		SET status = $3,
			login_token = $4,
			cookies_encrypted = $5,
			login_account_name = $6,
			last_validated_at = $7,
			expires_at = $8,
			updated_at = NOW()
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, status, login_token, cookies_encrypted, login_account_name,
			last_validated_at, expires_at, created_at, updated_at
	`, session.UserID, session.ID, session.Status, session.LoginToken, session.CookiesEncrypted,
		session.LoginAccountName, session.LastValidatedAt, session.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrWeChatSessionNotFound
	}
	if err != nil {
		return err
	}
	*session = *updated
	return nil
}

func (r *wechatExportRepository) GetSession(ctx context.Context, userID int64, sessionID int64) (*service.WeChatSession, error) {
	session, err := scanWeChatSession(ctx, r.sql, `
		SELECT id, user_id, status, login_token, cookies_encrypted, login_account_name,
			last_validated_at, expires_at, created_at, updated_at
		FROM wechat_sessions
		WHERE user_id = $1 AND id = $2
	`, userID, sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatSessionNotFound
	}
	return session, err
}

func (r *wechatExportRepository) ExpireUserSessions(ctx context.Context, userID int64) error {
	_, err := r.sql.ExecContext(ctx, `
		UPDATE wechat_sessions
		SET status = $1,
			updated_at = NOW()
		WHERE user_id = $2
			AND status IN ($3, $4, $5)
	`, service.WeChatSessionStatusExpired, userID, service.WeChatSessionStatusPending, service.WeChatSessionStatusScanConfirmed, service.WeChatSessionStatusReady)
	return err
}

func (r *wechatExportRepository) ExpireLoginAttemptSessions(ctx context.Context, userID int64) error {
	_, err := r.sql.ExecContext(ctx, `
		UPDATE wechat_sessions
		SET status = $1,
			updated_at = NOW()
		WHERE user_id = $2
			AND status IN ($3, $4)
	`, service.WeChatSessionStatusExpired, userID, service.WeChatSessionStatusPending, service.WeChatSessionStatusScanConfirmed)
	return err
}

func (r *wechatExportRepository) SearchAccounts(ctx context.Context, userID int64, query string, limit int) ([]service.WeChatAccount, error) {
	args := []any{userID, limit}
	where := "binding.user_id = $1"
	if strings.TrimSpace(query) != "" {
		args = append(args, "%"+strings.TrimSpace(query)+"%")
		where += " AND (account.fakeid ILIKE $3 OR account.nickname ILIKE $3 OR account.alias ILIKE $3)"
	}
	rows, err := r.sql.QueryContext(ctx, fmt.Sprintf(`
		SELECT binding.id, binding.user_id, account.fakeid, account.nickname, account.alias,
			account.avatar, account.description, binding.is_active, binding.last_synced_at,
			binding.created_at, GREATEST(binding.updated_at, account.updated_at)
		FROM wechat_account_bindings binding
		INNER JOIN wechat_public_accounts account ON account.id = binding.account_id
		WHERE %s
		ORDER BY binding.updated_at DESC, binding.id DESC
		LIMIT $2
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanWeChatAccountRows(rows)
}

func (r *wechatExportRepository) GetAccount(ctx context.Context, userID int64, fakeID string) (*service.WeChatAccount, error) {
	account, err := scanWeChatAccount(ctx, r.sql, `
		SELECT binding.id, binding.user_id, account.fakeid, account.nickname, account.alias,
			account.avatar, account.description, binding.is_active, binding.last_synced_at,
			binding.created_at, GREATEST(binding.updated_at, account.updated_at)
		FROM wechat_account_bindings binding
		INNER JOIN wechat_public_accounts account ON account.id = binding.account_id
		WHERE binding.user_id = $1 AND account.fakeid = $2
	`, userID, fakeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatAccountNotFound
	}
	return account, err
}

func (r *wechatExportRepository) UpsertAccount(ctx context.Context, account *service.WeChatAccount) error {
	if account == nil {
		return nil
	}
	return scanSingleRow(ctx, r.sql, `
		WITH public_account AS (
			INSERT INTO wechat_public_accounts (
				fakeid, nickname, alias, avatar, description
			)
			VALUES ($2, $3, $4, $5, $6)
			ON CONFLICT (fakeid) DO UPDATE
			SET nickname = COALESCE(NULLIF(EXCLUDED.nickname, ''), wechat_public_accounts.nickname),
				alias = COALESCE(NULLIF(EXCLUDED.alias, ''), wechat_public_accounts.alias),
				avatar = COALESCE(NULLIF(EXCLUDED.avatar, ''), wechat_public_accounts.avatar),
				description = COALESCE(NULLIF(EXCLUDED.description, ''), wechat_public_accounts.description),
				updated_at = NOW()
			RETURNING id, fakeid, nickname, alias, avatar, description, updated_at
		),
		binding AS (
			INSERT INTO wechat_account_bindings (
				user_id, account_id, is_active
			)
			SELECT $1, id, $7
			FROM public_account
			ON CONFLICT (user_id, account_id) DO UPDATE
			SET is_active = EXCLUDED.is_active,
				updated_at = NOW()
			RETURNING id, user_id, is_active, last_synced_at, created_at, updated_at
		)
		SELECT binding.id, binding.created_at, GREATEST(binding.updated_at, public_account.updated_at)
		FROM binding, public_account
	`, []any{
		account.UserID,
		account.FakeID,
		account.Nickname,
		account.Alias,
		account.Avatar,
		account.Description,
		account.IsActive,
	}, &account.ID, &account.CreatedAt, &account.UpdatedAt)
}

func (r *wechatExportRepository) MarkAccountSynced(ctx context.Context, userID int64, fakeID string) (*service.WeChatAccount, error) {
	account, err := scanWeChatAccount(ctx, r.sql, `
		WITH updated AS (
			UPDATE wechat_account_bindings binding
			SET last_synced_at = NOW(),
				updated_at = NOW()
			FROM wechat_public_accounts account
			WHERE binding.account_id = account.id
				AND binding.user_id = $1
				AND account.fakeid = $2
			RETURNING binding.id, binding.user_id, binding.is_active, binding.last_synced_at,
				binding.created_at, binding.updated_at, account.fakeid, account.nickname,
				account.alias, account.avatar, account.description, account.updated_at AS account_updated_at
		)
		SELECT id, user_id, fakeid, nickname, alias, avatar, description,
			is_active, last_synced_at, created_at, GREATEST(updated_at, account_updated_at)
		FROM updated
	`, userID, fakeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatAccountNotFound
	}
	return account, err
}

func (r *wechatExportRepository) UpsertArticle(ctx context.Context, article *service.WeChatArticle) error {
	if article == nil {
		return nil
	}
	if strings.TrimSpace(article.MetadataJSON) == "" {
		article.MetadataJSON = "{}"
	}
	// Preserve enriched data on the global article row. User-specific ownership is
	// represented by wechat_article_bindings.
	query := `
		WITH public_account AS (
			INSERT INTO wechat_public_accounts (fakeid, nickname)
			SELECT $2, $2
			WHERE $2 <> ''
			ON CONFLICT (fakeid) DO UPDATE SET updated_at = NOW()
			RETURNING id
		),
		public_article AS (
			INSERT INTO wechat_public_articles (
				account_id, title, author, link, cover, digest,
				publish_at, is_original, is_pay_subscribe, content_status, metadata_json
			)
			VALUES (
				(SELECT id FROM public_account),
				$4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb
			)
			ON CONFLICT (link) DO UPDATE
			SET account_id = COALESCE(EXCLUDED.account_id, wechat_public_articles.account_id),
				title = COALESCE(NULLIF(EXCLUDED.title, ''), wechat_public_articles.title),
				author = COALESCE(NULLIF(EXCLUDED.author, ''), wechat_public_articles.author),
				cover = COALESCE(NULLIF(EXCLUDED.cover, ''), wechat_public_articles.cover),
				digest = COALESCE(NULLIF(EXCLUDED.digest, ''), wechat_public_articles.digest),
				publish_at = COALESCE(EXCLUDED.publish_at, wechat_public_articles.publish_at),
				is_original = EXCLUDED.is_original,
				is_pay_subscribe = EXCLUDED.is_pay_subscribe,
				content_status = COALESCE(NULLIF(EXCLUDED.content_status, ''), wechat_public_articles.content_status),
				metadata_json = CASE
					WHEN wechat_public_articles.metadata_json = '{}'::jsonb OR wechat_public_articles.metadata_json IS NULL
					THEN EXCLUDED.metadata_json
					WHEN EXCLUDED.metadata_json = '{}'::jsonb OR EXCLUDED.metadata_json IS NULL
					THEN wechat_public_articles.metadata_json
					ELSE wechat_public_articles.metadata_json || EXCLUDED.metadata_json
				END,
				updated_at = NOW()
			RETURNING id, created_at, updated_at
		),
		binding AS (
			INSERT INTO wechat_article_bindings (user_id, article_id, source_type)
			SELECT $1, id, $3
			FROM public_article
			ON CONFLICT (user_id, article_id) DO UPDATE
			SET source_type = COALESCE(NULLIF(EXCLUDED.source_type, ''), wechat_article_bindings.source_type),
				updated_at = NOW()
			RETURNING updated_at
		)
		SELECT public_article.id, public_article.created_at, GREATEST(public_article.updated_at, binding.updated_at)
		FROM public_article, binding
	`
	return scanSingleRow(ctx, r.sql, query, []any{
		article.UserID,
		article.AccountFakeID,
		article.SourceType,
		article.Title,
		article.Author,
		article.Link,
		article.Cover,
		article.Digest,
		article.PublishAt,
		article.IsOriginal,
		article.IsPaySubscribe,
		article.ContentStatus,
		article.MetadataJSON,
	}, &article.ID, &article.CreatedAt, &article.UpdatedAt)
}

func (r *wechatExportRepository) UpdateArticleEnrichment(ctx context.Context, article *service.WeChatArticle) error {
	if article == nil {
		return nil
	}
	if strings.TrimSpace(article.MetadataJSON) == "" {
		article.MetadataJSON = "{}"
	}
	return scanSingleRow(ctx, r.sql, `
		WITH public_account AS (
			INSERT INTO wechat_public_accounts (fakeid, nickname)
			SELECT $3, $3
			WHERE $3 <> ''
			ON CONFLICT (fakeid) DO UPDATE SET updated_at = NOW()
			RETURNING id
		),
		updated AS (
			UPDATE wechat_public_articles article
			SET account_id = COALESCE((SELECT id FROM public_account), article.account_id),
				title = COALESCE(NULLIF($4, ''), article.title),
				author = COALESCE(NULLIF($5, ''), article.author),
				cover = COALESCE(NULLIF($6, ''), article.cover),
				digest = COALESCE(NULLIF($7, ''), article.digest),
				publish_at = COALESCE($8, article.publish_at),
				is_original = $9,
				is_pay_subscribe = $10,
				content_status = COALESCE(NULLIF($11, ''), article.content_status),
				metadata_json = $12::jsonb,
				updated_at = NOW()
			WHERE article.id = $1
				AND EXISTS (
					SELECT 1
					FROM wechat_article_bindings binding
					WHERE binding.article_id = article.id AND binding.user_id = $2
				)
			RETURNING article.created_at, article.updated_at
		)
		SELECT created_at, updated_at
		FROM updated
	`, []any{
		article.ID,
		article.UserID,
		article.AccountFakeID,
		article.Title,
		article.Author,
		article.Cover,
		article.Digest,
		article.PublishAt,
		article.IsOriginal,
		article.IsPaySubscribe,
		article.ContentStatus,
		article.MetadataJSON,
	}, &article.CreatedAt, &article.UpdatedAt)
}

func (r *wechatExportRepository) ListArticles(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.WeChatArticle, *pagination.PaginationResult, error) {
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM wechat_article_bindings WHERE user_id = $1", []any{userID}, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []service.WeChatArticle{}, paginationResultFromTotal(0, params), nil
	}
	query := `
		SELECT article.id, binding.user_id, COALESCE(account.fakeid, ''), binding.source_type,
			article.title, article.author, article.link, article.cover, article.digest,
			article.publish_at, article.is_original, article.is_pay_subscribe,
			article.content_status, article.metadata_json,
			LEAST(article.created_at, binding.created_at), GREATEST(article.updated_at, binding.updated_at)
		FROM wechat_article_bindings binding
		INNER JOIN wechat_public_articles article ON article.id = binding.article_id
		LEFT JOIN wechat_public_accounts account ON account.id = article.account_id
		WHERE binding.user_id = $1
		ORDER BY COALESCE(article.publish_at, article.created_at) DESC, article.id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.sql.QueryContext(ctx, query, userID, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	articles, err := scanWeChatArticleRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return articles, paginationResultFromTotal(total, params), nil
}

func (r *wechatExportRepository) GetArticleByID(ctx context.Context, articleID int64) (*service.WeChatArticle, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT article.id, binding.user_id, COALESCE(account.fakeid, ''), binding.source_type,
			article.title, article.author, article.link, article.cover, article.digest,
			article.publish_at, article.is_original, article.is_pay_subscribe,
			article.content_status, article.metadata_json,
			LEAST(article.created_at, binding.created_at), GREATEST(article.updated_at, binding.updated_at)
		FROM wechat_public_articles article
		INNER JOIN wechat_article_bindings binding ON binding.article_id = article.id
		LEFT JOIN wechat_public_accounts account ON account.id = article.account_id
		WHERE article.id = $1
		ORDER BY binding.updated_at DESC
		LIMIT 1
	`, articleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	articles, err := scanWeChatArticleRows(rows)
	if err != nil {
		return nil, err
	}
	if len(articles) == 0 {
		return nil, service.ErrWeChatArticleNotFound
	}
	return &articles[0], nil
}

func (r *wechatExportRepository) ListArticlesByIDs(ctx context.Context, userID int64, articleIDs []int64) ([]service.WeChatArticle, error) {
	if len(articleIDs) == 0 {
		return []service.WeChatArticle{}, nil
	}
	placeholders := make([]string, 0, len(articleIDs))
	args := []any{userID}
	for i, id := range articleIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		args = append(args, id)
	}
	query := fmt.Sprintf(`
		SELECT article.id, binding.user_id, COALESCE(account.fakeid, ''), binding.source_type,
			article.title, article.author, article.link, article.cover, article.digest,
			article.publish_at, article.is_original, article.is_pay_subscribe,
			article.content_status, article.metadata_json,
			LEAST(article.created_at, binding.created_at), GREATEST(article.updated_at, binding.updated_at)
		FROM wechat_article_bindings binding
		INNER JOIN wechat_public_articles article ON article.id = binding.article_id
		LEFT JOIN wechat_public_accounts account ON account.id = article.account_id
		WHERE binding.user_id = $1 AND article.id IN (%s)
		ORDER BY article.id ASC
	`, strings.Join(placeholders, ", "))
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanWeChatArticleRows(rows)
}

func (r *wechatExportRepository) CreateTask(ctx context.Context, task *service.WeChatExportTask) error {
	if task == nil {
		return nil
	}
	q := r.sql
	var tx *sql.Tx
	// 预留余额：如果费用估算 > 0，先预留余额，余额不足则拒绝创建
	if r.db != nil && task.CostEstimate > 0 {
		var err error
		tx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		q = tx
		// 预留余额，余额不足则返回错误（任务创建失败）
		reservation, err := reserveWeChatExportBalance(ctx, q, task.UserID, task.CostEstimate)
		if err != nil {
			return err
		}
		task.BalanceSnapshot = reservation.BalanceSnapshot
		task.ReservedPaidBalance = reservation.Paid
		task.ReservedGiftBalance = reservation.Gift
	}
	// Extract formats from PayloadJSON for database storage (formats_json field is kept for SQL queries)
	var formatsForDB []byte
	if task.PayloadJSON != "" {
		var payload struct {
			Formats []service.WeChatExportFormat `json:"formats"`
		}
		if err := json.Unmarshal([]byte(task.PayloadJSON), &payload); err == nil && len(payload.Formats) > 0 {
			formatsForDB, _ = json.Marshal(payload.Formats)
		}
	}
	if len(formatsForDB) == 0 {
		formatsForDB = []byte("[]")
	}

	query := `
		INSERT INTO wechat_export_tasks (
			user_id, status, selected_article_count, formats_json, include_engagement,
			payload_json, result_manifest_json, retention_days, cost_estimate, balance_snapshot,
			reserved_paid_balance, reserved_gift_balance
		)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6::jsonb, $7::jsonb, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	if err := scanSingleRow(ctx, q, query, []any{
		task.UserID,
		task.Status,
		task.SelectedArticleCount,
		formatsForDB,
		task.IncludeEngagement,
		task.PayloadJSON,
		task.ResultManifestJSON,
		task.RetentionDays,
		task.CostEstimate,
		task.BalanceSnapshot,
		task.ReservedPaidBalance,
		task.ReservedGiftBalance,
	}, &task.ID, &task.CreatedAt, &task.UpdatedAt); err != nil {
		return err
	}
	if err := insertWeChatTaskLog(ctx, q, service.WeChatExportTaskLog{
		TaskID:  task.ID,
		UserID:  task.UserID,
		Event:   "task_created",
		Status:  task.Status,
		Message: "Task queued for export.",
	}); err != nil {
		return err
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (r *wechatExportRepository) ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.WeChatExportTask, *pagination.PaginationResult, error) {
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM wechat_export_tasks WHERE user_id = $1", []any{userID}, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []service.WeChatExportTask{}, paginationResultFromTotal(0, params), nil
	}
	query := `
		SELECT id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
		FROM wechat_export_tasks
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.sql.QueryContext(ctx, query, userID, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	tasks, err := scanWeChatTaskRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return tasks, paginationResultFromTotal(total, params), nil
}

func (r *wechatExportRepository) GetWorkerStatus(ctx context.Context, userID int64) (*service.WeChatExportWorkerStatus, error) {
	var status service.WeChatExportWorkerStatus
	var lastTaskUpdatedAt sql.NullTime
	var oldestQueuedAt sql.NullTime
	err := scanSingleRow(ctx, r.sql, `
		SELECT
			COUNT(*) AS total_count,
			COUNT(*) FILTER (WHERE status = $2) AS queued_count,
			COUNT(*) FILTER (WHERE status = $3) AS running_count,
			COUNT(*) FILTER (WHERE status = $3 AND (worker_lease_until IS NULL OR worker_lease_until < NOW())) AS stale_running_count,
			COUNT(*) FILTER (WHERE status IN ($4, $5)) AS failed_count,
			COUNT(*) FILTER (WHERE status = $6) AS completed_count,
			COUNT(*) FILTER (WHERE status = $7) AS cancelled_count,
			MAX(updated_at) AS last_task_updated_at,
			MIN(created_at) FILTER (WHERE status = $2) AS oldest_queued_at
		FROM wechat_export_tasks
		WHERE user_id = $1
	`, []any{
		userID,
		service.WeChatExportTaskStatusQueued,
		service.WeChatExportTaskStatusRunning,
		service.WeChatExportTaskStatusFailed,
		service.WeChatExportTaskStatusCompletedWithErrors,
		service.WeChatExportTaskStatusCompleted,
		service.WeChatExportTaskStatusCancelled,
	},
		&status.TotalCount,
		&status.QueuedCount,
		&status.RunningCount,
		&status.StaleRunningCount,
		&status.FailedCount,
		&status.CompletedCount,
		&status.CancelledCount,
		&lastTaskUpdatedAt,
		&oldestQueuedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastTaskUpdatedAt.Valid {
		status.LastTaskUpdatedAt = &lastTaskUpdatedAt.Time
	}
	if oldestQueuedAt.Valid {
		status.OldestQueuedAt = &oldestQueuedAt.Time
	}
	return &status, nil
}

func (r *wechatExportRepository) GetTask(ctx context.Context, userID int64, taskID int64) (*service.WeChatExportTask, error) {
	where := "id = $1"
	args := []any{taskID}
	if userID > 0 {
		where = "user_id = $1 AND id = $2"
		args = []any{userID, taskID}
	}
	query := `
		SELECT id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
		FROM wechat_export_tasks
		WHERE ` + where + `
	`
	task, err := scanWeChatTask(ctx, r.sql, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	return task, err
}

func (r *wechatExportRepository) CancelTask(ctx context.Context, userID int64, taskID int64) (*service.WeChatExportTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock and fetch the task to check for reserved credits
	task, err := scanWeChatTask(ctx, tx, `
		SELECT id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
		FROM wechat_export_tasks
		WHERE user_id = $1 AND id = $2
		FOR UPDATE
	`, userID, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	// Only allow cancellation of queued or running tasks
	if task.Status != service.WeChatExportTaskStatusQueued && task.Status != service.WeChatExportTaskStatusRunning {
		return nil, service.ErrWeChatTaskNotFound
	}
	// Refund reserved credits if any
	if task.CostEstimate > 0 {
		refunded, refundErr := refundFullBalanceReservation(ctx, tx, task.UserID, weChatExportTaskReservation(task).Paid, weChatExportTaskReservation(task).Gift)
		if refundErr != nil {
			return nil, refundErr
		}
		task.BalanceSnapshot = refunded.BalanceSnapshot
		task.ReservedPaidBalance = 0
		task.ReservedGiftBalance = 0
		if _, updateErr := tx.ExecContext(ctx, `
			UPDATE wechat_export_tasks
			SET balance_snapshot = $1,
				reserved_paid_balance = 0,
				reserved_gift_balance = 0,
				updated_at = NOW()
			WHERE id = $2
		`, task.BalanceSnapshot, task.ID); updateErr != nil {
			return nil, updateErr
		}
		if usageErr := upsertWeChatExportUsageRecord(ctx, tx, task, task.SelectedArticleCount, task.CostEstimate, 0, "refunded", "{}"); usageErr != nil {
			return nil, usageErr
		}
	}
	// Update task status
	task, err = scanWeChatTask(ctx, tx, `
		UPDATE wechat_export_tasks
		SET status = $1,
			error_message = $2,
			worker_lease_until = NULL,
			worker_lease_token = '',  -- Phase 3：清空token（防止worker继续操作）
			worker_run_id = '',       -- Phase 3：清空run_id
			updated_at = NOW()
		WHERE user_id = $3
			AND id = $4
			AND status IN ($5, $6)
		RETURNING id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
	`, service.WeChatExportTaskStatusCancelled, "cancelled by user", userID, taskID, service.WeChatExportTaskStatusQueued, service.WeChatExportTaskStatusRunning)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	if logErr := insertWeChatTaskLog(ctx, tx, service.WeChatExportTaskLog{
		TaskID:  task.ID,
		UserID:  task.UserID,
		Event:   "task_cancelled",
		Status:  task.Status,
		Message: task.ErrorMessage,
	}); logErr != nil {
		return nil, logErr
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *wechatExportRepository) RetryTask(ctx context.Context, userID int64, taskID int64) (*service.WeChatExportTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock the task and verify it is in a retryable state
	task, err := scanWeChatTask(ctx, tx, `
		SELECT id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
		FROM wechat_export_tasks
		WHERE user_id = $1 AND id = $2
		FOR UPDATE
	`, userID, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	retryable := task.Status == service.WeChatExportTaskStatusFailed ||
		task.Status == service.WeChatExportTaskStatusCompletedWithErrors ||
		task.Status == service.WeChatExportTaskStatusCancelled ||
		task.Status == service.WeChatExportTaskStatusCompleted
	if !retryable {
		return nil, service.ErrWeChatTaskConflict
	}
	// For completed/failed-with-errors tasks, credits were already settled by
	// CompleteTask (actual cost adjustment) or FailTask/CancelTask (full refund).
	// Re-estimate and re-reserve credits for the retry.
	var costEstimate float64
	var reservation userBalanceReservation
	if task.CostEstimate > 0 {
		// Re-reserve the same estimated cost for the retry.
		// If the user's balance is insufficient, the retry fails.
		reservation, err = reserveWeChatExportBalance(ctx, tx, task.UserID, task.CostEstimate)
		if err != nil {
			return nil, fmt.Errorf("re-reserve balance for retry: %w", err)
		}
		costEstimate = task.CostEstimate
	}
	// Update task status and reset work state
	task, err = scanWeChatTask(ctx, tx, `
		UPDATE wechat_export_tasks
		SET status = $1,
			successful_article_count = 0,
			failed_article_count = 0,
			result_manifest_json = '{}'::jsonb,
			error_message = '',
			worker_lease_until = NULL,
			worker_lease_token = '',  -- Phase 3：清空旧token
			worker_run_id = '',       -- Phase 3：清空旧run_id
			expires_at = NULL,
			cost_estimate = $2,
			balance_snapshot = $3,
			reserved_paid_balance = $4,
			reserved_gift_balance = $5,
			updated_at = NOW()
		WHERE id = $6
			AND status IN ($7, $8, $9, $10)
		RETURNING id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
	`, service.WeChatExportTaskStatusQueued, costEstimate, reservation.BalanceSnapshot, reservation.Paid, reservation.Gift, task.ID,
		service.WeChatExportTaskStatusFailed, service.WeChatExportTaskStatusCompletedWithErrors,
		service.WeChatExportTaskStatusCancelled, service.WeChatExportTaskStatusCompleted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskConflict
	}
	if err != nil {
		return nil, err
	}
	task.CostEstimate = costEstimate
	task.BalanceSnapshot = reservation.BalanceSnapshot
	task.ReservedPaidBalance = reservation.Paid
	task.ReservedGiftBalance = reservation.Gift
	if _, err := tx.ExecContext(ctx, `
		UPDATE wechat_export_artifacts
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE user_id = $1
			AND task_id = $2
			AND deleted_at IS NULL
	`, userID, taskID); err != nil {
		return nil, err
	}
	if err := insertWeChatTaskLog(ctx, tx, service.WeChatExportTaskLog{
		TaskID:  task.ID,
		UserID:  task.UserID,
		Event:   "task_retried",
		Status:  task.Status,
		Message: "Task reset and queued for retry.",
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *wechatExportRepository) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*service.WeChatExportTask, []service.WeChatArticle, string, error) {
	// Phase 3：生成lease_token和run_id
	leaseToken, err := service.GenerateWorkerLeaseToken()
	if err != nil {
		return nil, nil, "", err
	}
	runID, err := service.GenerateWorkerRunID()
	if err != nil {
		return nil, nil, "", err
	}

	query := `
		WITH next AS (
			SELECT id
			FROM wechat_export_tasks
			WHERE status = $1
				OR (status = $2 AND (worker_lease_until IS NULL OR worker_lease_until < NOW()))
			ORDER BY created_at ASC, id ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE wechat_export_tasks AS tasks
		SET status = $4,
			worker_lease_until = NOW() + ($3 * interval '1 second'),
			worker_lease_token = $5,  -- Phase 3：生成并存储token
			worker_run_id = $6,       -- Phase 3：生成并存储run_id
			error_message = '',
			updated_at = NOW()
		FROM next
		WHERE tasks.id = next.id
		RETURNING tasks.id, tasks.user_id, tasks.status, tasks.selected_article_count, tasks.successful_article_count, tasks.failed_article_count,
			tasks.formats_json, tasks.include_engagement, tasks.payload_json, tasks.result_manifest_json, tasks.error_message,
			tasks.worker_lease_until, tasks.retention_days, tasks.expires_at, tasks.cost_estimate, tasks.balance_snapshot,
			tasks.reserved_paid_balance, tasks.reserved_gift_balance,
			tasks.worker_lease_token, tasks.worker_run_id, tasks.created_at, tasks.updated_at
	`
	task, err := scanWeChatTask(ctx, r.sql, query, service.WeChatExportTaskStatusQueued, service.WeChatExportTaskStatusRunning, leaseSeconds, service.WeChatExportTaskStatusRunning, leaseToken, runID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, "", nil
	}
	if err != nil {
		return nil, nil, "", err
	}
	if err := insertWeChatTaskLog(ctx, r.sql, service.WeChatExportTaskLog{
		TaskID:  task.ID,
		UserID:  task.UserID,
		Event:   "task_claimed",
		Status:  task.Status,
		Message: "Worker claimed the task.",
	}); err != nil {
		return nil, nil, "", err
	}
	articles, err := r.ListArticlesByIDs(ctx, task.UserID, task.ArticleIDs)
	if err != nil {
		return nil, nil, "", err
	}
	// Phase 3：返回leaseToken给worker
	return task, articles, leaseToken, nil
}

func (r *wechatExportRepository) CompleteTask(ctx context.Context, taskID int64, leaseToken string, artifacts []service.WeChatExportArtifact, resultManifestJSON string, failedArticleCount int, actualCost float64) (*service.WeChatExportTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	// Lock and validate the task before completing — only running tasks may be completed.
	existing, err := scanWeChatTask(ctx, tx, `
		SELECT id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
		FROM wechat_export_tasks
		WHERE id = $1
		FOR UPDATE
	`, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	if existing.Status != service.WeChatExportTaskStatusRunning {
		return nil, fmt.Errorf("cannot complete task in status %q: %w", existing.Status, service.ErrWeChatTaskConflict)
	}
	// Phase 3：新增lease_token验证
	if existing.WorkerLeaseToken != leaseToken {
		return nil, infraerrors.Unauthorized("LEASE_TOKEN_MISMATCH",
			"lease token does not match the claimed task")
	}
	// Phase 3：新增lease未过期验证（防止late callback）
	if existing.WorkerLeaseUntil != nil && existing.WorkerLeaseUntil.Before(time.Now()) {
		return nil, infraerrors.Conflict("LEASE_EXPIRED",
			"worker lease has expired, task may have been reclaimed")
	}
	status := service.WeChatExportTaskStatusCompleted
	if failedArticleCount > 0 {
		status = service.WeChatExportTaskStatusCompletedWithErrors
	}
	task, err := scanWeChatTask(ctx, tx, `
		UPDATE wechat_export_tasks
		SET status = $1,
			successful_article_count = GREATEST(selected_article_count - $2, 0),
			failed_article_count = $2,
			result_manifest_json = $3::jsonb,
			worker_lease_until = NULL,
			worker_lease_token = '',  -- Phase 3：清空token
			worker_run_id = '',       -- Phase 3：清空run_id
			expires_at = NOW() + (retention_days * interval '1 day'),
			updated_at = NOW()
		WHERE id = $4 AND status = $5
		RETURNING id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
	`, status, failedArticleCount, resultManifestJSON, taskID, service.WeChatExportTaskStatusRunning)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskConflict
	}
	if err != nil {
		return nil, err
	}
	// Billing adjustment: settle the difference between reserved and actual cost
	adjustment := actualCost - task.CostEstimate
	reservation := weChatExportTaskReservation(task)
	if adjustment > 0 {
		delta, reserveErr := reserveWeChatExportBalance(ctx, tx, task.UserID, adjustment)
		if reserveErr != nil {
			return nil, reserveErr
		}
		reservation = mergeBalanceReservation(reservation, delta)
	} else if adjustment < 0 {
		refunded, refundErr := refundBalanceReservation(ctx, tx, task.UserID, -adjustment, reservation.Paid, reservation.Gift)
		if refundErr != nil {
			return nil, refundErr
		}
		reservation = reduceBalanceReservation(reservation, refunded)
	}
	if adjustment != 0 {
		task.BalanceSnapshot = reservation.BalanceSnapshot
		task.ReservedPaidBalance = reservation.Paid
		task.ReservedGiftBalance = reservation.Gift
		if _, updateErr := tx.ExecContext(ctx, `
			UPDATE wechat_export_tasks
			SET balance_snapshot = $1,
				reserved_paid_balance = $2,
				reserved_gift_balance = $3,
				updated_at = NOW()
			WHERE id = $4
		`, task.BalanceSnapshot, task.ReservedPaidBalance, task.ReservedGiftBalance, task.ID); updateErr != nil {
			return nil, updateErr
		}
	}
	if usageErr := upsertWeChatExportUsageRecord(ctx, tx, task, task.SelectedArticleCount, task.CostEstimate, actualCost, "settled", fmt.Sprintf(`{"artifact_count":%d,"failed_article_count":%d}`, len(artifacts), failedArticleCount)); usageErr != nil {
		return nil, usageErr
	}
	for i := range artifacts {
		artifact := &artifacts[i]
		artifact.TaskID = task.ID
		artifact.UserID = task.UserID
		if err := insertWeChatArtifact(ctx, tx, artifact); err != nil {
			return nil, err
		}
	}
	if err := insertWeChatTaskLog(ctx, tx, service.WeChatExportTaskLog{
		TaskID:  task.ID,
		UserID:  task.UserID,
		Event:   "task_completed",
		Status:  task.Status,
		Message: fmt.Sprintf("Export completed with %d artifact(s) and %d failed article(s).", len(artifacts), failedArticleCount),
		MetaJSON: fmt.Sprintf(`{"artifact_count":%d,"failed_article_count":%d}`,
			len(artifacts),
			failedArticleCount,
		),
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *wechatExportRepository) FailTask(ctx context.Context, taskID int64, leaseToken string, message string) (*service.WeChatExportTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock and fetch the task to check status and reserved credits
	task, err := scanWeChatTask(ctx, tx, `
		SELECT id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
		FROM wechat_export_tasks
		WHERE id = $1
		FOR UPDATE
	`, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	// Only allow failing tasks that are still running or queued.
	// This prevents double-refunding already completed or cancelled tasks.
	if task.Status != service.WeChatExportTaskStatusRunning && task.Status != service.WeChatExportTaskStatusQueued {
		return nil, fmt.Errorf("cannot fail task in status %q: %w", task.Status, service.ErrWeChatTaskConflict)
	}
	// Phase 3：新增lease_token验证（仅running状态）
	if task.Status == service.WeChatExportTaskStatusRunning {
		if task.WorkerLeaseToken != leaseToken {
			return nil, infraerrors.Unauthorized("LEASE_TOKEN_MISMATCH",
				"lease token does not match the running task")
		}
		if task.WorkerLeaseUntil != nil && task.WorkerLeaseUntil.Before(time.Now()) {
			return nil, infraerrors.Conflict("LEASE_EXPIRED",
				"worker lease has expired")
		}
	}
	// Refund reserved credits if any (before the status update, within the same
	// transaction and row lock, so it's safe from concurrent FailTask calls).
	if task.CostEstimate > 0 {
		refunded, refundErr := refundFullBalanceReservation(ctx, tx, task.UserID, weChatExportTaskReservation(task).Paid, weChatExportTaskReservation(task).Gift)
		if refundErr != nil {
			return nil, refundErr
		}
		task.BalanceSnapshot = refunded.BalanceSnapshot
		task.ReservedPaidBalance = 0
		task.ReservedGiftBalance = 0
		if usageErr := upsertWeChatExportUsageRecord(ctx, tx, task, task.SelectedArticleCount, task.CostEstimate, 0, "refunded", "{}"); usageErr != nil {
			return nil, usageErr
		}
	}
	// Update task status — only allow failing tasks that are still running or queued
	task, err = scanWeChatTask(ctx, tx, `
		UPDATE wechat_export_tasks
		SET status = $1,
			error_message = $2,
			worker_lease_until = NULL,
			worker_lease_token = '',  -- Phase 3：清空token
			worker_run_id = '',       -- Phase 3：清空run_id
			balance_snapshot = $4,
			reserved_paid_balance = 0,
			reserved_gift_balance = 0,
			updated_at = NOW()
		WHERE id = $3
			AND status IN ($5, $6)
		RETURNING id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance,
			worker_lease_token, worker_run_id, created_at, updated_at
	`, service.WeChatExportTaskStatusFailed, message, taskID, task.BalanceSnapshot, service.WeChatExportTaskStatusRunning, service.WeChatExportTaskStatusQueued)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskConflict
	}
	if err != nil {
		return nil, err
	}
	if logErr := insertWeChatTaskLog(ctx, tx, service.WeChatExportTaskLog{
		TaskID:  task.ID,
		UserID:  task.UserID,
		Event:   "task_failed",
		Status:  task.Status,
		Message: task.ErrorMessage,
	}); logErr != nil {
		return nil, logErr
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *wechatExportRepository) AddTaskLog(ctx context.Context, taskID int64, leaseToken string, log service.WeChatExportTaskLog) (*service.WeChatExportTaskLog, error) {
	// Phase 3：验证lease_token（强制验证，用户决策）
	if strings.TrimSpace(leaseToken) == "" {
		return nil, infraerrors.BadRequest("LEASE_TOKEN_REQUIRED", "lease token is required for adding task log")
	}

	// Phase 3：验证token匹配和lease状态
	var status string
	var actualToken string
	var leaseUntil sql.NullTime
	err := scanSingleRow(ctx, r.sql, `
		SELECT status, worker_lease_token, worker_lease_until
		FROM wechat_export_tasks
		WHERE id = $1
	`, []any{taskID}, &status, &actualToken, &leaseUntil)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	// Phase 3：running状态强制验证lease_token
	if status == service.WeChatExportTaskStatusRunning {
		if actualToken != leaseToken {
			return nil, infraerrors.Unauthorized("LEASE_TOKEN_MISMATCH",
				"lease token does not match the running task")
		}
		// 检查lease未过期
		if leaseUntil.Valid && leaseUntil.Time.Before(time.Now()) {
			return nil, infraerrors.Conflict("LEASE_EXPIRED",
				"worker lease has expired")
		}
	}

	// 原有INSERT逻辑
	var inserted service.WeChatExportTaskLog
	var metaJSON []byte
	err = scanSingleRow(ctx, r.sql, `
		INSERT INTO wechat_export_task_logs (
			task_id, user_id, event, status, message, meta_json
		)
		SELECT id, user_id, $2, $3, $4, $5::jsonb
		FROM wechat_export_tasks
		WHERE id = $1
		RETURNING id, task_id, user_id, event, status, message, meta_json, created_at
	`, []any{
		taskID,
		strings.TrimSpace(log.Event),
		strings.TrimSpace(log.Status),
		strings.TrimSpace(log.Message),
		normalizeWeChatTaskLogMetaJSON(log.MetaJSON),
	},
		&inserted.ID,
		&inserted.TaskID,
		&inserted.UserID,
		&inserted.Event,
		&inserted.Status,
		&inserted.Message,
		&metaJSON,
		&inserted.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	inserted.MetaJSON = string(metaJSON)
	return &inserted, nil
}

func (r *wechatExportRepository) ListTaskLogs(ctx context.Context, userID int64, taskID int64) ([]service.WeChatExportTaskLog, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT logs.id, logs.task_id, logs.user_id, logs.event, logs.status, logs.message,
			logs.meta_json, logs.created_at
		FROM wechat_export_task_logs logs
		INNER JOIN wechat_export_tasks tasks ON tasks.id = logs.task_id
		WHERE tasks.user_id = $1 AND logs.user_id = $1 AND logs.task_id = $2
		ORDER BY logs.created_at ASC, logs.id ASC
	`, userID, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanWeChatTaskLogRows(rows)
}

func (r *wechatExportRepository) ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]service.WeChatExportArtifact, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, task_id, user_id, format, storage_provider, storage_key, download_url,
			file_name, file_size, checksum, expires_at, deleted_at, created_at, updated_at
		FROM wechat_export_artifacts
		WHERE user_id = $1 AND task_id = $2 AND deleted_at IS NULL
		ORDER BY created_at ASC, id ASC
	`, userID, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanWeChatArtifactRows(rows)
}

func (r *wechatExportRepository) GetArtifact(ctx context.Context, userID int64, artifactID int64) (*service.WeChatExportArtifact, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, task_id, user_id, format, storage_provider, storage_key, download_url,
			file_name, file_size, checksum, expires_at, deleted_at, created_at, updated_at
		FROM wechat_export_artifacts
		WHERE user_id = $1 AND id = $2 AND deleted_at IS NULL
	`, userID, artifactID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	artifacts, err := scanWeChatArtifactRows(rows)
	if err != nil {
		return nil, err
	}
	if len(artifacts) == 0 {
		return nil, service.ErrWeChatTaskNotFound
	}
	return &artifacts[0], nil
}

func scanWeChatTask(ctx context.Context, q sqlQueryer, query string, args ...any) (*service.WeChatExportTask, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	tasks, err := scanWeChatTaskRows(rows)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, sql.ErrNoRows
	}
	return &tasks[0], nil
}

func reserveWeChatExportBalance(ctx context.Context, q sqlExecutor, userID int64, amount float64) (userBalanceReservation, error) {
	return reserveUserBalanceWithComponents(ctx, q, userID, amount, service.ErrWeChatInsufficientBalance)
}

func weChatExportTaskReservation(task *service.WeChatExportTask) userBalanceReservation {
	if task == nil {
		return userBalanceReservation{}
	}
	return userBalanceReservation{
		BalanceSnapshot: task.BalanceSnapshot,
		Paid:            task.ReservedPaidBalance,
		Gift:            task.ReservedGiftBalance,
	}
}

func upsertWeChatExportUsageRecord(ctx context.Context, q sqlExecutor, task *service.WeChatExportTask, articleCount int, reservedCost float64, actualCost float64, billingStatus string, metadataJSON string) error {
	if task == nil {
		return nil
	}
	_, err := q.ExecContext(ctx, `
		INSERT INTO wechat_export_usage_records (
			task_id, user_id, article_count, format_count, include_engagement,
			reserved_cost, actual_cost, balance_snapshot, billing_status, metadata_json
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		ON CONFLICT (task_id) DO UPDATE SET
			article_count = EXCLUDED.article_count,
			format_count = EXCLUDED.format_count,
			include_engagement = EXCLUDED.include_engagement,
			reserved_cost = EXCLUDED.reserved_cost,
			actual_cost = EXCLUDED.actual_cost,
			balance_snapshot = EXCLUDED.balance_snapshot,
			billing_status = EXCLUDED.billing_status,
			metadata_json = EXCLUDED.metadata_json,
			updated_at = NOW()
	`, task.ID, task.UserID, articleCount, len(task.Formats), task.IncludeEngagement,
		reservedCost, actualCost, task.BalanceSnapshot, billingStatus, metadataJSON)
	return err
}

func scanWeChatSession(ctx context.Context, q sqlQueryer, query string, args ...any) (*service.WeChatSession, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	var session service.WeChatSession
	var lastValidatedAt sql.NullTime
	var expiresAt sql.NullTime
	if err := rows.Scan(
		&session.ID,
		&session.UserID,
		&session.Status,
		&session.LoginToken,
		&session.CookiesEncrypted,
		&session.LoginAccountName,
		&lastValidatedAt,
		&expiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if lastValidatedAt.Valid {
		session.LastValidatedAt = &lastValidatedAt.Time
	}
	if expiresAt.Valid {
		session.ExpiresAt = &expiresAt.Time
	}
	return &session, rows.Err()
}

func scanWeChatAccount(ctx context.Context, q sqlQueryer, query string, args ...any) (*service.WeChatAccount, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	accounts, err := scanWeChatAccountRows(rows)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, sql.ErrNoRows
	}
	return &accounts[0], nil
}

func scanWeChatAccountRows(rows *sql.Rows) ([]service.WeChatAccount, error) {
	items := make([]service.WeChatAccount, 0)
	for rows.Next() {
		var item service.WeChatAccount
		var lastSyncedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.FakeID,
			&item.Nickname,
			&item.Alias,
			&item.Avatar,
			&item.Description,
			&item.IsActive,
			&lastSyncedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if lastSyncedAt.Valid {
			item.LastSyncedAt = &lastSyncedAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanWeChatArticleRows(rows *sql.Rows) ([]service.WeChatArticle, error) {
	items := make([]service.WeChatArticle, 0)
	for rows.Next() {
		var item service.WeChatArticle
		var publishAt sql.NullTime
		var metadataJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.AccountFakeID,
			&item.SourceType,
			&item.Title,
			&item.Author,
			&item.Link,
			&item.Cover,
			&item.Digest,
			&publishAt,
			&item.IsOriginal,
			&item.IsPaySubscribe,
			&item.ContentStatus,
			&metadataJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if publishAt.Valid {
			item.PublishAt = &publishAt.Time
		}
		item.MetadataJSON = string(metadataJSON)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanWeChatTaskRows(rows *sql.Rows) ([]service.WeChatExportTask, error) {
	items := make([]service.WeChatExportTask, 0)
	for rows.Next() {
		var item service.WeChatExportTask
		var formatsJSON []byte
		var payloadJSON []byte
		var manifestJSON []byte
		var leaseUntil sql.NullTime
		var expiresAt sql.NullTime
		// Phase 3：新增字段扫描
		var leaseToken string
		var runID string
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Status,
			&item.SelectedArticleCount,
			&item.SuccessfulArticleCount,
			&item.FailedArticleCount,
			&formatsJSON,
			&item.IncludeEngagement,
			&payloadJSON,
			&manifestJSON,
			&item.ErrorMessage,
			&leaseUntil,
			&item.RetentionDays,
			&expiresAt,
			&item.CostEstimate,
			&item.BalanceSnapshot,
			&item.ReservedPaidBalance,
			&item.ReservedGiftBalance,
			&leaseToken, // Phase 3：新增
			&runID,      // Phase 3：新增
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if leaseUntil.Valid {
			item.WorkerLeaseUntil = &leaseUntil.Time
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}
		// Phase 3：赋值新字段
		item.WorkerLeaseToken = leaseToken
		item.WorkerRunID = runID
		item.PayloadJSON = string(payloadJSON)
		item.ResultManifestJSON = string(manifestJSON)
		if err := hydrateWeChatTaskPayload(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func hydrateWeChatTaskPayload(task *service.WeChatExportTask) error {
	if task == nil {
		return nil
	}
	if strings.TrimSpace(task.PayloadJSON) == "" {
		return nil
	}
	var payload wechatExportTaskPayload
	if err := json.Unmarshal([]byte(task.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("parse wechat task payload: %w", err)
	}
	task.ArticleIDs = payload.ArticleIDs
	task.Formats = payload.Formats
	return nil
}

func insertWeChatArtifact(ctx context.Context, q sqlQueryer, artifact *service.WeChatExportArtifact) error {
	return scanSingleRow(ctx, q, `
		INSERT INTO wechat_export_artifacts (
			task_id, user_id, format, storage_provider, storage_key, download_url,
			file_name, file_size, checksum, expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`, []any{
		artifact.TaskID,
		artifact.UserID,
		artifact.Format,
		artifact.StorageProvider,
		artifact.StorageKey,
		artifact.DownloadURL,
		artifact.FileName,
		artifact.FileSize,
		artifact.Checksum,
		artifact.ExpiresAt,
	}, &artifact.ID, &artifact.CreatedAt, &artifact.UpdatedAt)
}

func insertWeChatTaskLog(ctx context.Context, q sqlQueryer, log service.WeChatExportTaskLog) error {
	message := strings.TrimSpace(log.Message)
	return scanSingleRow(ctx, q, `
		INSERT INTO wechat_export_task_logs (
			task_id, user_id, event, status, message, meta_json
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING id, created_at
	`, []any{
		log.TaskID,
		log.UserID,
		strings.TrimSpace(log.Event),
		strings.TrimSpace(log.Status),
		message,
		normalizeWeChatTaskLogMetaJSON(log.MetaJSON),
	}, &log.ID, &log.CreatedAt)
}

func normalizeWeChatTaskLogMetaJSON(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "{}"
	}
	return value
}

func scanWeChatTaskLogRows(rows *sql.Rows) ([]service.WeChatExportTaskLog, error) {
	items := make([]service.WeChatExportTaskLog, 0)
	for rows.Next() {
		var item service.WeChatExportTaskLog
		var metaJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.TaskID,
			&item.UserID,
			&item.Event,
			&item.Status,
			&item.Message,
			&metaJSON,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.MetaJSON = string(metaJSON)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanWeChatArtifactRows(rows *sql.Rows) ([]service.WeChatExportArtifact, error) {
	items := make([]service.WeChatExportArtifact, 0)
	for rows.Next() {
		var item service.WeChatExportArtifact
		var expiresAt sql.NullTime
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.TaskID,
			&item.UserID,
			&item.Format,
			&item.StorageProvider,
			&item.StorageKey,
			&item.DownloadURL,
			&item.FileName,
			&item.FileSize,
			&item.Checksum,
			&expiresAt,
			&deletedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}
		if deletedAt.Valid {
			item.DeletedAt = &deletedAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
