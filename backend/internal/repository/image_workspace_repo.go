package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/Aias00/cloudbase/internal/service"
)

type imageWorkspaceRepository struct {
	db  *sql.DB
	sql sqlExecutor
}

func NewImageWorkspaceRepository(db *sql.DB) service.ImageWorkspaceRepository {
	return &imageWorkspaceRepository{db: db, sql: db}
}

func (r *imageWorkspaceRepository) CreateTask(ctx context.Context, task *service.ImageWorkspaceTask) error {
	if task == nil {
		return nil
	}
	q := r.sql
	var tx *sql.Tx
	if r.db != nil && task.CostEstimate > 0 {
		var err error
		tx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		q = tx
		// Acquire per-user advisory lock to serialize concurrent task creation
		if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", task.UserID); err != nil {
			return err
		}
		reservation, err := reserveImageWorkspaceBalance(ctx, q, task.UserID, task.CostEstimate)
		if err != nil {
			return err
		}
		task.BalanceSnapshot = reservation.BalanceSnapshot
		task.ReservedPaidBalance = reservation.Paid
		task.ReservedGiftBalance = reservation.Gift
	}
	if err := scanSingleRow(ctx, q, `
		INSERT INTO image_workspace_tasks (
			user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot,
			reserved_paid_balance, reserved_gift_balance, result_json
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18::jsonb)
		RETURNING id, created_at, updated_at
	`, []any{
		task.UserID,
		task.Status,
		task.Prompt,
		task.NegativePrompt,
		task.Model,
		task.Provider,
		task.Size,
		task.Quality,
		task.Style,
		task.Seed,
		task.BatchSize,
		task.TemplateID,
		task.WorkerLeaseUntil,
		task.CostEstimate,
		task.BalanceSnapshot,
		task.ReservedPaidBalance,
		task.ReservedGiftBalance,
		normalizeJSONText(task.ResultJSON),
	}, &task.ID, &task.CreatedAt, &task.UpdatedAt); err != nil {
		return err
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (r *imageWorkspaceRepository) ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams, filters service.ImageWorkspaceTaskFilters) ([]service.ImageWorkspaceTask, *pagination.PaginationResult, error) {
	whereClauses := []string{"user_id = $1"}
	args := []any{userID}
	argIdx := 2
	if filters.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	whereClause := strings.Join(whereClauses, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM image_workspace_tasks WHERE %s", whereClause)
	if err := scanSingleRow(ctx, r.sql, countQuery, args, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []service.ImageWorkspaceTask{}, paginationResultFromTotal(0, params), nil
	}
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance, error_message,
			result_json, created_at, updated_at
		FROM image_workspace_tasks
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, params.Limit(), params.Offset())
	rows, err := r.sql.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items, err := scanImageWorkspaceTaskRows(rows)
	if err != nil {
		return nil, nil, err
	}
	if err := r.hydrateTaskArtifacts(ctx, userID, items); err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *imageWorkspaceRepository) GetTask(ctx context.Context, userID int64, taskID int64) (*service.ImageWorkspaceTask, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance, error_message,
			result_json, created_at, updated_at
		FROM image_workspace_tasks
		WHERE user_id = $1 AND id = $2
	`, userID, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items, err := scanImageWorkspaceTaskRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrImageWorkspaceTaskNotFound
	}
	return &items[0], nil
}

func (r *imageWorkspaceRepository) ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]service.ImageWorkspaceArtifact, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, task_id, user_id, storage_provider, storage_key, image_url, prompt,
			mime_type, width, height, file_size, checksum, metadata_json, created_at
		FROM image_workspace_artifacts
		WHERE user_id = $1 AND task_id = $2
		ORDER BY created_at ASC, id ASC
	`, userID, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanImageWorkspaceArtifactRows(rows)
}

func (r *imageWorkspaceRepository) GetArtifact(ctx context.Context, userID int64, artifactID int64) (*service.ImageWorkspaceArtifact, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, task_id, user_id, storage_provider, storage_key, image_url, prompt,
			mime_type, width, height, file_size, checksum, metadata_json, created_at
		FROM image_workspace_artifacts
		WHERE user_id = $1 AND id = $2
	`, userID, artifactID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items, err := scanImageWorkspaceArtifactRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, service.ErrImageWorkspaceTaskNotFound
	}
	return &items[0], nil
}

func (r *imageWorkspaceRepository) ListTemplates(ctx context.Context, userID int64) ([]service.ImageWorkspaceTemplate, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, user_id, title, description, prompt, negative_prompt, model, size,
			quality, style, is_default, created_at, updated_at
		FROM image_workspace_templates
		WHERE user_id = $1
		ORDER BY is_default DESC, updated_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanImageWorkspaceTemplateRows(rows)
}

func (r *imageWorkspaceRepository) UpsertTemplate(ctx context.Context, template *service.ImageWorkspaceTemplate) error {
	if template == nil {
		return nil
	}
	if template.ID > 0 {
		err := scanSingleRow(ctx, r.sql, `
			UPDATE image_workspace_templates
			SET title = $3,
				description = $4,
				prompt = $5,
				negative_prompt = $6,
				model = $7,
				size = $8,
				quality = $9,
				style = $10,
				is_default = $11,
				updated_at = NOW()
			WHERE user_id = $1 AND id = $2
			RETURNING created_at, updated_at
		`, []any{
			template.UserID,
			template.ID,
			template.Title,
			template.Description,
			template.Prompt,
			template.NegativePrompt,
			template.Model,
			template.Size,
			template.Quality,
			template.Style,
			template.IsDefault,
		}, &template.CreatedAt, &template.UpdatedAt)
		if errors.Is(err, sql.ErrNoRows) {
			return service.ErrImageWorkspaceTaskNotFound
		}
		return err
	}
	return scanSingleRow(ctx, r.sql, `
		INSERT INTO image_workspace_templates (
			user_id, title, description, prompt, negative_prompt, model, size, quality, style, is_default
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`, []any{
		template.UserID,
		template.Title,
		template.Description,
		template.Prompt,
		template.NegativePrompt,
		template.Model,
		template.Size,
		template.Quality,
		template.Style,
		template.IsDefault,
	}, &template.ID, &template.CreatedAt, &template.UpdatedAt)
}

func (r *imageWorkspaceRepository) DeleteTemplate(ctx context.Context, userID int64, templateID int64) error {
	result, err := r.sql.ExecContext(ctx, `
		DELETE FROM image_workspace_templates
		WHERE user_id = $1 AND id = $2
	`, userID, templateID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return service.ErrImageWorkspaceTaskNotFound
	}
	return nil
}

func (r *imageWorkspaceRepository) ListUsageRecords(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.ImageWorkspaceUsageRecord, *pagination.PaginationResult, error) {
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM image_workspace_usage_records WHERE user_id = $1", []any{userID}, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []service.ImageWorkspaceUsageRecord{}, paginationResultFromTotal(0, params), nil
	}
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, task_id, user_id, provider, model, size, quality, image_count,
			reserved_cost, actual_cost, balance_snapshot, billing_status, metadata_json,
			created_at, updated_at
		FROM image_workspace_usage_records
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, userID, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items, err := scanImageWorkspaceUsageRecordRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *imageWorkspaceRepository) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*service.ImageWorkspaceTask, error) {
	query := `
		WITH next AS (
			SELECT id
			FROM image_workspace_tasks
			WHERE status = $1
				OR (status = $2 AND (worker_lease_until IS NULL OR worker_lease_until < NOW()))
			ORDER BY created_at ASC, id ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE image_workspace_tasks AS tasks
		SET status = $2,
			worker_lease_until = NOW() + ($3 * interval '1 second'),
			error_message = '',
			updated_at = NOW()
		FROM next
		WHERE tasks.id = next.id
		RETURNING tasks.id, tasks.user_id, tasks.status, tasks.prompt, tasks.negative_prompt, tasks.model, tasks.provider, tasks.size, tasks.quality, tasks.style,
			tasks.seed, tasks.batch_size, tasks.template_id, tasks.worker_lease_until, tasks.cost_estimate, tasks.balance_snapshot,
			tasks.reserved_paid_balance, tasks.reserved_gift_balance, tasks.error_message,
			tasks.result_json, tasks.created_at, tasks.updated_at
	`
	task, err := scanImageWorkspaceTask(ctx, r.sql, query, service.ImageWorkspaceTaskStatusQueued, service.ImageWorkspaceTaskStatusRunning, leaseSeconds)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return task, err
}

func (r *imageWorkspaceRepository) CompleteTask(ctx context.Context, taskID int64, artifacts []service.ImageWorkspaceArtifact, resultJSON string, cost float64) (*service.ImageWorkspaceTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	current, err := scanImageWorkspaceTask(ctx, tx, `
		SELECT id, user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance, error_message,
			result_json, created_at, updated_at
		FROM image_workspace_tasks
		WHERE id = $1
		FOR UPDATE
	`, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrImageWorkspaceTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	if current.Status == service.ImageWorkspaceTaskStatusFailed ||
		current.Status == service.ImageWorkspaceTaskStatusCancelled ||
		current.Status == service.ImageWorkspaceTaskStatusSucceeded {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return current, nil
	}
	originalEstimate := current.CostEstimate
	// Validate worker-reported cost is not suspiciously lower than the estimate
	// Allow up to 50% reduction for legitimate variance (e.g., smaller output sizes)
	// but reject obvious bypass attempts like cost=0 for paid tasks
	minAllowedCost := originalEstimate * 0.5
	if originalEstimate > 0 && cost < minAllowedCost {
		return nil, service.ErrImageWorkspaceInvalidCost
	}
	adjustment := cost - originalEstimate
	reservation := imageWorkspaceTaskReservation(current)
	if adjustment > 0 {
		delta, err := reserveImageWorkspaceBalance(ctx, tx, current.UserID, adjustment)
		if err != nil {
			return nil, err
		}
		reservation = mergeBalanceReservation(reservation, delta)
	} else if adjustment < 0 {
		refunded, err := refundBalanceReservation(ctx, tx, current.UserID, -adjustment, reservation.Paid, reservation.Gift)
		if err != nil {
			return nil, err
		}
		reservation = reduceBalanceReservation(reservation, refunded)
	}
	task, err := scanImageWorkspaceTask(ctx, tx, `
		UPDATE image_workspace_tasks
		SET status = $1,
			cost_estimate = $2,
			result_json = $3::jsonb,
			error_message = '',
			worker_lease_until = NULL,
			updated_at = NOW()
		WHERE id = $4
		RETURNING id, user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance, error_message,
			result_json, created_at, updated_at
	`, service.ImageWorkspaceTaskStatusSucceeded, cost, normalizeJSONText(resultJSON), taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrImageWorkspaceTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	for i := range artifacts {
		artifact := &artifacts[i]
		artifact.TaskID = task.ID
		artifact.UserID = task.UserID
		if err := insertImageWorkspaceArtifact(ctx, tx, artifact); err != nil {
			return nil, err
		}
	}
	if adjustment != 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE image_workspace_tasks
			SET balance_snapshot = $1,
				reserved_paid_balance = $2,
				reserved_gift_balance = $3
			WHERE id = $4
		`, reservation.BalanceSnapshot, reservation.Paid, reservation.Gift, task.ID); err != nil {
			return nil, err
		}
		task.BalanceSnapshot = reservation.BalanceSnapshot
		task.ReservedPaidBalance = reservation.Paid
		task.ReservedGiftBalance = reservation.Gift
	}
	if err := upsertImageWorkspaceUsageRecord(ctx, tx, task, len(artifacts), originalEstimate, cost, resultJSON); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *imageWorkspaceRepository) FailTask(ctx context.Context, taskID int64, message string, resultJSON string) (*service.ImageWorkspaceTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	current, err := scanImageWorkspaceTask(ctx, tx, `
		SELECT id, user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance, error_message,
			result_json, created_at, updated_at
		FROM image_workspace_tasks
		WHERE id = $1
		FOR UPDATE
	`, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrImageWorkspaceTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	if current.Status == service.ImageWorkspaceTaskStatusFailed ||
		current.Status == service.ImageWorkspaceTaskStatusCancelled ||
		current.Status == service.ImageWorkspaceTaskStatusSucceeded {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return current, nil
	}
	task, err := scanImageWorkspaceTask(ctx, tx, `
		UPDATE image_workspace_tasks
		SET status = $1,
			error_message = $2,
			result_json = $3::jsonb,
			worker_lease_until = NULL,
			updated_at = NOW()
		WHERE id = $4
		RETURNING id, user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance, error_message,
			result_json, created_at, updated_at
	`, service.ImageWorkspaceTaskStatusFailed, message, normalizeJSONText(resultJSON), taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrImageWorkspaceTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	if task.CostEstimate > 0 {
		refunded, err := refundFullBalanceReservation(ctx, tx, task.UserID, imageWorkspaceTaskReservation(task).Paid, imageWorkspaceTaskReservation(task).Gift)
		if err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE image_workspace_tasks
			SET balance_snapshot = $1,
				reserved_paid_balance = 0,
				reserved_gift_balance = 0
			WHERE id = $2
		`, refunded.BalanceSnapshot, task.ID); err != nil {
			return nil, err
		}
		task.BalanceSnapshot = refunded.BalanceSnapshot
		task.ReservedPaidBalance = 0
		task.ReservedGiftBalance = 0
	}
	if task.CostEstimate > 0 {
		metadataJSON, _ := json.Marshal(map[string]string{
			"error_message": message,
			"status":        service.ImageWorkspaceTaskStatusFailed,
		})
		if err := upsertImageWorkspaceUsageRecordWithStatus(ctx, tx, task, 0, task.CostEstimate, 0, string(metadataJSON), "refunded"); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *imageWorkspaceRepository) CancelTask(ctx context.Context, taskID int64, userID int64) (*service.ImageWorkspaceTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	current, err := scanImageWorkspaceTask(ctx, tx, `
		SELECT id, user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance, error_message,
			result_json, created_at, updated_at
		FROM image_workspace_tasks
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, taskID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrImageWorkspaceTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	// Allow cancelling queued tasks or running tasks with expired lease
	now := time.Now()
	if current.Status != service.ImageWorkspaceTaskStatusQueued {
		if current.Status != service.ImageWorkspaceTaskStatusRunning {
			return nil, service.ErrImageWorkspaceInvalidInput
		}
		if current.WorkerLeaseUntil != nil && current.WorkerLeaseUntil.After(now) {
			return nil, service.ErrImageWorkspaceInvalidInput
		}
	}
	task, err := scanImageWorkspaceTask(ctx, tx, `
		UPDATE image_workspace_tasks
		SET status = $1,
			error_message = 'cancelled by user',
			worker_lease_until = NULL,
			updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, status, prompt, negative_prompt, model, provider, size, quality, style,
			seed, batch_size, template_id, worker_lease_until, cost_estimate, balance_snapshot, reserved_paid_balance, reserved_gift_balance, error_message,
			result_json, created_at, updated_at
	`, service.ImageWorkspaceTaskStatusCancelled, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrImageWorkspaceTaskNotFound
	}
	if err != nil {
		return nil, err
	}
	if task.CostEstimate > 0 {
		refunded, err := refundFullBalanceReservation(ctx, tx, task.UserID, imageWorkspaceTaskReservation(task).Paid, imageWorkspaceTaskReservation(task).Gift)
		if err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE image_workspace_tasks
			SET balance_snapshot = $1,
				reserved_paid_balance = 0,
				reserved_gift_balance = 0
			WHERE id = $2
		`, refunded.BalanceSnapshot, task.ID); err != nil {
			return nil, err
		}
		task.BalanceSnapshot = refunded.BalanceSnapshot
		task.ReservedPaidBalance = 0
		task.ReservedGiftBalance = 0
	}
	if task.CostEstimate > 0 {
		metadataJSON, _ := json.Marshal(map[string]string{
			"status": service.ImageWorkspaceTaskStatusCancelled,
		})
		if err := upsertImageWorkspaceUsageRecordWithStatus(ctx, tx, task, 0, task.CostEstimate, 0, string(metadataJSON), "refunded"); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *imageWorkspaceRepository) GetWorkerStatus(ctx context.Context) (*service.ImageWorkspaceWorkerStatus, error) {
	var status service.ImageWorkspaceWorkerStatus
	var lastTaskUpdatedAt sql.NullTime
	var lastFailedAt sql.NullTime
	var lastFailureMessage sql.NullString
	var oldestQueuedAt sql.NullTime
	err := scanSingleRow(ctx, r.sql, `
		SELECT
			COUNT(*) AS total_count,
			COUNT(*) FILTER (WHERE status = $1) AS queued_count,
			COUNT(*) FILTER (WHERE status = $2) AS running_count,
			COUNT(*) FILTER (WHERE status = $2 AND (worker_lease_until IS NULL OR worker_lease_until < NOW())) AS stale_running_count,
			COUNT(*) FILTER (WHERE status = $3) AS failed_count,
			COUNT(*) FILTER (WHERE status = $3 AND updated_at >= NOW() - INTERVAL '30 minutes') AS recent_failed_count,
			COUNT(*) FILTER (WHERE status = $4) AS succeeded_count,
			COUNT(*) FILTER (WHERE status = $5) AS cancelled_count,
			COALESCE((SELECT COUNT(*) FROM image_workspace_artifacts), 0) AS artifact_count,
			MAX(updated_at) AS last_task_updated_at,
			(
				SELECT updated_at
				FROM image_workspace_tasks
				WHERE status = $3
				ORDER BY updated_at DESC, id DESC
				LIMIT 1
			) AS last_failed_at,
			(
				SELECT error_message
				FROM image_workspace_tasks
				WHERE status = $3
				ORDER BY updated_at DESC, id DESC
				LIMIT 1
			) AS last_failure_message,
			MIN(created_at) FILTER (WHERE status = $1) AS oldest_queued_at
		FROM image_workspace_tasks
	`, []any{
		service.ImageWorkspaceTaskStatusQueued,
		service.ImageWorkspaceTaskStatusRunning,
		service.ImageWorkspaceTaskStatusFailed,
		service.ImageWorkspaceTaskStatusSucceeded,
		service.ImageWorkspaceTaskStatusCancelled,
	},
		&status.TotalCount,
		&status.QueuedCount,
		&status.RunningCount,
		&status.StaleRunningCount,
		&status.FailedCount,
		&status.RecentFailedCount,
		&status.SucceededCount,
		&status.CancelledCount,
		&status.ArtifactCount,
		&lastTaskUpdatedAt,
		&lastFailedAt,
		&lastFailureMessage,
		&oldestQueuedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastTaskUpdatedAt.Valid {
		status.LastTaskUpdatedAt = &lastTaskUpdatedAt.Time
	}
	if lastFailedAt.Valid {
		status.LastFailedAt = &lastFailedAt.Time
	}
	if lastFailureMessage.Valid {
		status.LastFailureMessage = lastFailureMessage.String
	}
	if oldestQueuedAt.Valid {
		status.OldestQueuedAt = &oldestQueuedAt.Time
	}
	return &status, nil
}

func reserveImageWorkspaceBalance(ctx context.Context, q sqlExecutor, userID int64, amount float64) (userBalanceReservation, error) {
	return reserveUserBalanceWithComponents(ctx, q, userID, amount, service.ErrInsufficientBalance)
}

func imageWorkspaceTaskReservation(task *service.ImageWorkspaceTask) userBalanceReservation {
	if task == nil {
		return userBalanceReservation{}
	}
	return userBalanceReservation{
		BalanceSnapshot: task.BalanceSnapshot,
		Paid:            task.ReservedPaidBalance,
		Gift:            task.ReservedGiftBalance,
	}
}

func scanImageWorkspaceTask(ctx context.Context, q sqlQueryer, query string, args ...any) (*service.ImageWorkspaceTask, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items, err := scanImageWorkspaceTaskRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	return &items[0], nil
}

func insertImageWorkspaceArtifact(ctx context.Context, q sqlExecutor, artifact *service.ImageWorkspaceArtifact) error {
	if artifact == nil {
		return nil
	}
	_, err := q.ExecContext(ctx, `
		INSERT INTO image_workspace_artifacts (
			task_id, user_id, storage_provider, storage_key, image_url, prompt,
			mime_type, width, height, file_size, checksum, metadata_json
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb)
	`, artifact.TaskID,
		artifact.UserID,
		artifact.StorageProvider,
		artifact.StorageKey,
		artifact.ImageURL,
		artifact.Prompt,
		artifact.MimeType,
		artifact.Width,
		artifact.Height,
		artifact.FileSize,
		artifact.Checksum,
		normalizeJSONText(artifact.MetadataJSON),
	)
	if err != nil {
		return fmt.Errorf("insert image workspace artifact: %w", err)
	}
	return nil
}

func upsertImageWorkspaceUsageRecord(ctx context.Context, q sqlExecutor, task *service.ImageWorkspaceTask, imageCount int, reservedCost float64, actualCost float64, metadataJSON string) error {
	return upsertImageWorkspaceUsageRecordWithStatus(ctx, q, task, imageCount, reservedCost, actualCost, metadataJSON, "settled")
}

func upsertImageWorkspaceUsageRecordWithStatus(ctx context.Context, q sqlExecutor, task *service.ImageWorkspaceTask, imageCount int, reservedCost float64, actualCost float64, metadataJSON string, billingStatus string) error {
	if task == nil {
		return nil
	}
	if billingStatus == "" {
		billingStatus = "settled"
	}
	_, err := q.ExecContext(ctx, `
		INSERT INTO image_workspace_usage_records (
			task_id, user_id, provider, model, size, quality, image_count,
			reserved_cost, actual_cost, balance_snapshot, billing_status, metadata_json
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb)
		ON CONFLICT (task_id) DO UPDATE
		SET provider = EXCLUDED.provider,
			model = EXCLUDED.model,
			size = EXCLUDED.size,
			quality = EXCLUDED.quality,
			image_count = EXCLUDED.image_count,
			reserved_cost = EXCLUDED.reserved_cost,
			actual_cost = EXCLUDED.actual_cost,
			balance_snapshot = EXCLUDED.balance_snapshot,
			billing_status = EXCLUDED.billing_status,
			metadata_json = EXCLUDED.metadata_json,
			updated_at = NOW()
	`, task.ID,
		task.UserID,
		task.Provider,
		task.Model,
		task.Size,
		task.Quality,
		imageCount,
		reservedCost,
		actualCost,
		task.BalanceSnapshot,
		billingStatus,
		normalizeJSONText(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("upsert image workspace usage record: %w", err)
	}
	return nil
}

func (r *imageWorkspaceRepository) hydrateTaskArtifacts(ctx context.Context, userID int64, tasks []service.ImageWorkspaceTask) error {
	succeededIDs := make([]int64, 0)
	for _, task := range tasks {
		if task.Status == service.ImageWorkspaceTaskStatusSucceeded {
			succeededIDs = append(succeededIDs, task.ID)
		}
	}
	if len(succeededIDs) == 0 {
		return nil
	}
	placeholders := make([]string, len(succeededIDs))
	args := make([]any, len(succeededIDs)+1)
	args[0] = userID
	for i, id := range succeededIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}
	query := fmt.Sprintf(`
		SELECT id, task_id, user_id, storage_provider, storage_key, image_url, prompt,
			mime_type, width, height, file_size, checksum, metadata_json, created_at
		FROM image_workspace_artifacts
		WHERE user_id = $1 AND task_id IN (%s)
		ORDER BY created_at ASC, id ASC
	`, strings.Join(placeholders, ", "))
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	artifactsByTask := make(map[int64][]service.ImageWorkspaceArtifact)
	for rows.Next() {
		artifact, err := scanImageWorkspaceArtifactRow(rows)
		if err != nil {
			return err
		}
		artifactsByTask[artifact.TaskID] = append(artifactsByTask[artifact.TaskID], artifact)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range tasks {
		if artifacts, ok := artifactsByTask[tasks[i].ID]; ok {
			tasks[i].Artifacts = artifacts
		}
	}
	return nil
}

func scanImageWorkspaceTaskRows(rows *sql.Rows) ([]service.ImageWorkspaceTask, error) {
	items := make([]service.ImageWorkspaceTask, 0)
	for rows.Next() {
		var item service.ImageWorkspaceTask
		var seed sql.NullInt64
		var templateID sql.NullInt64
		var workerLeaseUntil sql.NullTime
		var resultJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Status,
			&item.Prompt,
			&item.NegativePrompt,
			&item.Model,
			&item.Provider,
			&item.Size,
			&item.Quality,
			&item.Style,
			&seed,
			&item.BatchSize,
			&templateID,
			&workerLeaseUntil,
			&item.CostEstimate,
			&item.BalanceSnapshot,
			&item.ReservedPaidBalance,
			&item.ReservedGiftBalance,
			&item.ErrorMessage,
			&resultJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if seed.Valid {
			item.Seed = &seed.Int64
		}
		if templateID.Valid {
			item.TemplateID = &templateID.Int64
		}
		if workerLeaseUntil.Valid {
			item.WorkerLeaseUntil = &workerLeaseUntil.Time
		}
		item.ResultJSON = string(resultJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanImageWorkspaceArtifactRow(rows *sql.Rows) (service.ImageWorkspaceArtifact, error) {
	var item service.ImageWorkspaceArtifact
	var metadataJSON []byte
	if err := rows.Scan(
		&item.ID,
		&item.TaskID,
		&item.UserID,
		&item.StorageProvider,
		&item.StorageKey,
		&item.ImageURL,
		&item.Prompt,
		&item.MimeType,
		&item.Width,
		&item.Height,
		&item.FileSize,
		&item.Checksum,
		&metadataJSON,
		&item.CreatedAt,
	); err != nil {
		return item, err
	}
	item.MetadataJSON = string(metadataJSON)
	return item, nil
}

func scanImageWorkspaceArtifactRows(rows *sql.Rows) ([]service.ImageWorkspaceArtifact, error) {
	items := make([]service.ImageWorkspaceArtifact, 0)
	for rows.Next() {
		item, err := scanImageWorkspaceArtifactRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanImageWorkspaceTemplateRows(rows *sql.Rows) ([]service.ImageWorkspaceTemplate, error) {
	items := make([]service.ImageWorkspaceTemplate, 0)
	for rows.Next() {
		var item service.ImageWorkspaceTemplate
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Title,
			&item.Description,
			&item.Prompt,
			&item.NegativePrompt,
			&item.Model,
			&item.Size,
			&item.Quality,
			&item.Style,
			&item.IsDefault,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanImageWorkspaceUsageRecordRows(rows *sql.Rows) ([]service.ImageWorkspaceUsageRecord, error) {
	items := make([]service.ImageWorkspaceUsageRecord, 0)
	for rows.Next() {
		var item service.ImageWorkspaceUsageRecord
		var metadataJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.TaskID,
			&item.UserID,
			&item.Provider,
			&item.Model,
			&item.Size,
			&item.Quality,
			&item.ImageCount,
			&item.ReservedCost,
			&item.ActualCost,
			&item.BalanceSnapshot,
			&item.BillingStatus,
			&metadataJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.MetadataJSON = string(metadataJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}

func normalizeJSONText(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}
