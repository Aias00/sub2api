package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
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
			AND status IN ($2, $3)
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, userID, service.WeChatSessionStatusPending, service.WeChatSessionStatusReady)
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
		INSERT INTO wechat_sessions (user_id, status, login_token, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, []any{
		session.UserID,
		session.Status,
		session.LoginToken,
		session.ExpiresAt,
	}, &session.ID, &session.CreatedAt, &session.UpdatedAt)
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
			AND status IN ($3, $4)
	`, service.WeChatSessionStatusExpired, userID, service.WeChatSessionStatusPending, service.WeChatSessionStatusReady)
	return err
}

func (r *wechatExportRepository) UpsertArticle(ctx context.Context, article *service.WeChatArticle) error {
	if article == nil {
		return nil
	}
	if strings.TrimSpace(article.MetadataJSON) == "" {
		article.MetadataJSON = "{}"
	}
	query := `
		INSERT INTO wechat_articles (
			user_id, account_fakeid, source_type, title, author, link, cover, digest,
			publish_at, is_original, is_pay_subscribe, content_status, metadata_json
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb)
		ON CONFLICT (user_id, link) DO UPDATE
		SET account_fakeid = EXCLUDED.account_fakeid,
			source_type = EXCLUDED.source_type,
			title = EXCLUDED.title,
			author = EXCLUDED.author,
			cover = EXCLUDED.cover,
			digest = EXCLUDED.digest,
			publish_at = EXCLUDED.publish_at,
			is_original = EXCLUDED.is_original,
			is_pay_subscribe = EXCLUDED.is_pay_subscribe,
			content_status = EXCLUDED.content_status,
			metadata_json = EXCLUDED.metadata_json,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
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

func (r *wechatExportRepository) ListArticles(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.WeChatArticle, *pagination.PaginationResult, error) {
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM wechat_articles WHERE user_id = $1", []any{userID}, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []service.WeChatArticle{}, paginationResultFromTotal(0, params), nil
	}
	query := `
		SELECT id, user_id, account_fakeid, source_type, title, author, link, cover, digest,
			publish_at, is_original, is_pay_subscribe, content_status, metadata_json, created_at, updated_at
		FROM wechat_articles
		WHERE user_id = $1
		ORDER BY COALESCE(publish_at, created_at) DESC, id DESC
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
		SELECT id, user_id, account_fakeid, source_type, title, author, link, cover, digest,
			publish_at, is_original, is_pay_subscribe, content_status, metadata_json, created_at, updated_at
		FROM wechat_articles
		WHERE user_id = $1 AND id IN (%s)
		ORDER BY id ASC
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
	query := `
		INSERT INTO wechat_export_tasks (
			user_id, status, selected_article_count, formats_json, include_engagement,
			payload_json, result_manifest_json, retention_days
		)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6::jsonb, $7::jsonb, $8)
		RETURNING id, created_at, updated_at
	`
	return scanSingleRow(ctx, r.sql, query, []any{
		task.UserID,
		task.Status,
		task.SelectedArticleCount,
		task.FormatsJSON,
		task.IncludeEngagement,
		task.PayloadJSON,
		task.ResultManifestJSON,
		task.RetentionDays,
	}, &task.ID, &task.CreatedAt, &task.UpdatedAt)
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
			worker_lease_until, retention_days, expires_at, created_at, updated_at
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

func (r *wechatExportRepository) GetTask(ctx context.Context, userID int64, taskID int64) (*service.WeChatExportTask, error) {
	query := `
		SELECT id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, created_at, updated_at
		FROM wechat_export_tasks
		WHERE user_id = $1 AND id = $2
	`
	task, err := scanWeChatTask(ctx, r.sql, query, userID, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	return task, err
}

func (r *wechatExportRepository) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*service.WeChatExportTask, []service.WeChatArticle, error) {
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
			error_message = '',
			updated_at = NOW()
		FROM next
		WHERE tasks.id = next.id
		RETURNING tasks.id, tasks.user_id, tasks.status, tasks.selected_article_count, tasks.successful_article_count, tasks.failed_article_count,
			tasks.formats_json, tasks.include_engagement, tasks.payload_json, tasks.result_manifest_json, tasks.error_message,
			tasks.worker_lease_until, tasks.retention_days, tasks.expires_at, tasks.created_at, tasks.updated_at
	`
	task, err := scanWeChatTask(ctx, r.sql, query, service.WeChatExportTaskStatusQueued, service.WeChatExportTaskStatusRunning, leaseSeconds, service.WeChatExportTaskStatusRunning)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	articles, err := r.ListArticlesByIDs(ctx, task.UserID, task.ArticleIDs)
	if err != nil {
		return nil, nil, err
	}
	return task, articles, nil
}

func (r *wechatExportRepository) CompleteTask(ctx context.Context, taskID int64, artifacts []service.WeChatExportArtifact, resultManifestJSON string) (*service.WeChatExportTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	task, err := scanWeChatTask(ctx, tx, `
		UPDATE wechat_export_tasks
		SET status = $1,
			successful_article_count = selected_article_count,
			failed_article_count = $2,
			result_manifest_json = $3::jsonb,
			worker_lease_until = NULL,
			expires_at = NOW() + (retention_days * interval '1 day'),
			updated_at = NOW()
		WHERE id = $4
		RETURNING id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, created_at, updated_at
	`, service.WeChatExportTaskStatusCompleted, 0, resultManifestJSON, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	for i := range artifacts {
		artifact := &artifacts[i]
		artifact.TaskID = task.ID
		artifact.UserID = task.UserID
		if err := insertWeChatArtifact(ctx, tx, artifact); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *wechatExportRepository) FailTask(ctx context.Context, taskID int64, message string) (*service.WeChatExportTask, error) {
	task, err := scanWeChatTask(ctx, r.sql, `
		UPDATE wechat_export_tasks
		SET status = $1,
			error_message = $2,
			worker_lease_until = NULL,
			updated_at = NOW()
		WHERE id = $3
		RETURNING id, user_id, status, selected_article_count, successful_article_count, failed_article_count,
			formats_json, include_engagement, payload_json, result_manifest_json, error_message,
			worker_lease_until, retention_days, expires_at, created_at, updated_at
	`, service.WeChatExportTaskStatusFailed, message, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrWeChatTaskNotFound
	}
	return task, err
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
		item.FormatsJSON = string(formatsJSON)
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
	if strings.TrimSpace(task.FormatsJSON) != "" {
		if err := json.Unmarshal([]byte(task.FormatsJSON), &task.Formats); err != nil {
			return fmt.Errorf("parse wechat task formats: %w", err)
		}
	}
	if strings.TrimSpace(task.PayloadJSON) == "" {
		return nil
	}
	var payload wechatExportTaskPayload
	if err := json.Unmarshal([]byte(task.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("parse wechat task payload: %w", err)
	}
	task.ArticleIDs = payload.ArticleIDs
	if len(task.Formats) == 0 {
		task.Formats = payload.Formats
	}
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
