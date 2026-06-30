package repository

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestImageWorkspaceRepositoryCreateTaskReservesBalance(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := NewImageWorkspaceRepository(db)
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	task := &service.ImageWorkspaceTask{
		UserID:         42,
		Status:         service.ImageWorkspaceTaskStatusQueued,
		Prompt:         "A clean product render",
		NegativePrompt: "blurry",
		Model:          "gpt-image-2",
		Provider:       "openai",
		Size:           "1024x1024",
		Quality:        "standard",
		Style:          "editorial",
		BatchSize:      1,
		CostEstimate:   0.5,
		ResultJSON:     "{}",
	}

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("UPDATE users\\s+SET balance = balance -").
		WithArgs(0.5, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(9.5))
	mock.ExpectQuery("INSERT INTO image_workspace_tasks").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(101), now, now))
	mock.ExpectCommit()

	require.NoError(t, repo.CreateTask(context.Background(), task))
	require.Equal(t, int64(101), task.ID)
	require.Equal(t, 9.5, task.BalanceSnapshot)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImageWorkspaceRepositoryCompleteTaskSettlesArtifactsAndUsage(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := NewImageWorkspaceRepository(db)
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	artifact := service.ImageWorkspaceArtifact{
		StorageProvider: "local",
		StorageKey:      "/tmp/sub2api-image-workspace/demo.png",
		ImageURL:        "/api/v1/image-workspace/artifacts/900/download",
		Prompt:          "A clean product render",
		MimeType:        "image/png",
		Width:           1024,
		Height:          1024,
		FileSize:        68,
		Checksum:        "sha256-demo",
		MetadataJSON:    `{"source":"worker"}`,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM image_workspace_tasks\\s+WHERE id = \\$1\\s+FOR UPDATE").
		WithArgs(int64(101)).
		WillReturnRows(imageWorkspaceTaskRows(now).
			AddRow(int64(101), int64(42), service.ImageWorkspaceTaskStatusRunning, "A clean product render", "blurry", "gpt-image-2", "openai", "1024x1024", "standard", "editorial", nil, 1, nil, now.Add(time.Minute), 0.5, 9.5, "", []byte(`{}`), now, now))
	mock.ExpectQuery("UPDATE users\\s+SET balance = balance -").
		WithArgs(0.25, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(9.25))
	mock.ExpectQuery("UPDATE image_workspace_tasks\\s+SET status = \\$1").
		WithArgs(service.ImageWorkspaceTaskStatusSucceeded, 0.75, `{"artifact_count":1}`, int64(101)).
		WillReturnRows(imageWorkspaceTaskRows(now).
			AddRow(int64(101), int64(42), service.ImageWorkspaceTaskStatusSucceeded, "A clean product render", "blurry", "gpt-image-2", "openai", "1024x1024", "standard", "editorial", nil, 1, nil, nil, 0.75, 9.5, "", []byte(`{"artifact_count":1}`), now, now))
	mock.ExpectExec("INSERT INTO image_workspace_artifacts").
		WithArgs(int64(101), int64(42), artifact.StorageProvider, artifact.StorageKey, artifact.ImageURL, artifact.Prompt, artifact.MimeType, artifact.Width, artifact.Height, artifact.FileSize, artifact.Checksum, artifact.MetadataJSON).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE image_workspace_tasks\\s+SET balance_snapshot = \\$1").
		WithArgs(9.25, int64(101)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO image_workspace_usage_records").
		WithArgs(int64(101), int64(42), "openai", "gpt-image-2", "1024x1024", "standard", 1, 0.5, 0.75, 9.25, "settled", `{"artifact_count":1}`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	task, err := repo.CompleteTask(context.Background(), 101, []service.ImageWorkspaceArtifact{artifact}, `{"artifact_count":1}`, 0.75)
	require.NoError(t, err)
	require.Equal(t, service.ImageWorkspaceTaskStatusSucceeded, task.Status)
	require.Equal(t, 9.25, task.BalanceSnapshot)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImageWorkspaceRepositoryCompleteTaskRejectsAdditionalCostWhenBalanceInsufficient(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := NewImageWorkspaceRepository(db)
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	artifact := service.ImageWorkspaceArtifact{
		StorageProvider: "local",
		StorageKey:      "/tmp/sub2api-image-workspace/demo.png",
		ImageURL:        "/api/v1/image-workspace/artifacts/900/download",
		Prompt:          "A clean product render",
		MimeType:        "image/png",
		Width:           1024,
		Height:          1024,
		FileSize:        68,
		Checksum:        "sha256-demo",
		MetadataJSON:    `{"source":"worker"}`,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("FROM image_workspace_tasks\\s+WHERE id = \\$1\\s+FOR UPDATE").
		WithArgs(int64(101)).
		WillReturnRows(imageWorkspaceTaskRows(now).
			AddRow(int64(101), int64(42), service.ImageWorkspaceTaskStatusRunning, "A clean product render", "blurry", "gpt-image-2", "openai", "1024x1024", "standard", "editorial", nil, 1, nil, now.Add(time.Minute), 0.5, 0, "", []byte(`{}`), now, now))
	mock.ExpectQuery("UPDATE users\\s+SET balance = balance -").
		WithArgs(0.25, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}))
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\$1\\)").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	task, err := repo.CompleteTask(context.Background(), 101, []service.ImageWorkspaceArtifact{artifact}, `{"artifact_count":1}`, 0.75)
	require.ErrorIs(t, err, service.ErrInsufficientBalance)
	require.Nil(t, task)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImageWorkspaceRepositoryClaimNextTaskUsesQualifiedReturningColumns(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := NewImageWorkspaceRepository(db)
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectQuery("RETURNING tasks\\.id, tasks\\.user_id, tasks\\.status").
		WithArgs(service.ImageWorkspaceTaskStatusQueued, service.ImageWorkspaceTaskStatusRunning, int64(300)).
		WillReturnRows(imageWorkspaceTaskRows(now).
			AddRow(int64(101), int64(42), service.ImageWorkspaceTaskStatusRunning, "A clean product render", "blurry", "gpt-image-2", "openai", "1024x1024", "standard", "editorial", nil, 1, nil, now.Add(5*time.Minute), 0.5, 9.5, "", []byte(`{}`), now, now))

	task, err := repo.ClaimNextTask(context.Background(), 300)
	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, int64(101), task.ID)
	require.Equal(t, service.ImageWorkspaceTaskStatusRunning, task.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImageWorkspaceRepositoryFailTaskRefundsReservedBalance(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := NewImageWorkspaceRepository(db)
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("FROM image_workspace_tasks\\s+WHERE id = \\$1\\s+FOR UPDATE").
		WithArgs(int64(101)).
		WillReturnRows(imageWorkspaceTaskRows(now).
			AddRow(int64(101), int64(42), service.ImageWorkspaceTaskStatusRunning, "A clean product render", "blurry", "gpt-image-2", "openai", "1024x1024", "standard", "editorial", nil, 1, nil, now.Add(time.Minute), 0.5, 9.5, "", []byte(`{}`), now, now))
	mock.ExpectQuery("UPDATE image_workspace_tasks\\s+SET status = \\$1").
		WithArgs(service.ImageWorkspaceTaskStatusFailed, "upstream timeout", `{"failure":{"upstream_status":504}}`, int64(101)).
		WillReturnRows(imageWorkspaceTaskRows(now).
			AddRow(int64(101), int64(42), service.ImageWorkspaceTaskStatusFailed, "A clean product render", "blurry", "gpt-image-2", "openai", "1024x1024", "standard", "editorial", nil, 1, nil, nil, 0.5, 9.5, "upstream timeout", []byte(`{"failure":{"upstream_status":504}}`), now, now))
	mock.ExpectQuery("UPDATE users\\s+SET balance = balance \\+").
		WithArgs(0.5, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(10.0))
	mock.ExpectExec("UPDATE image_workspace_tasks\\s+SET balance_snapshot = \\$1").
		WithArgs(10.0, int64(101)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO image_workspace_usage_records").
		WithArgs(int64(101), int64(42), "openai", "gpt-image-2", "1024x1024", "standard", 0, 0.5, 0.0, 10.0, "refunded", `{"error_message":"upstream timeout","status":"failed"}`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	task, err := repo.FailTask(context.Background(), 101, "upstream timeout", `{"failure":{"upstream_status":504}}`)
	require.NoError(t, err)
	require.Equal(t, service.ImageWorkspaceTaskStatusFailed, task.Status)
	require.JSONEq(t, `{"failure":{"upstream_status":504}}`, task.ResultJSON)
	require.Equal(t, 10.0, task.BalanceSnapshot)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImageWorkspaceRepositoryListUsageRecords(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := NewImageWorkspaceRepository(db)
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM image_workspace_usage_records WHERE user_id = \\$1").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT id, task_id, user_id, provider, model").
		WithArgs(int64(42), 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"task_id",
			"user_id",
			"provider",
			"model",
			"size",
			"quality",
			"image_count",
			"reserved_cost",
			"actual_cost",
			"balance_snapshot",
			"billing_status",
			"metadata_json",
			"created_at",
			"updated_at",
		}).AddRow(int64(9), int64(101), int64(42), "openai", "gpt-image-2", "1024x1024", "standard", 2, 0.5, 0.75, 9.25, "settled", []byte(`{"artifact_count":2}`), now, now))

	items, result, err := repo.ListUsageRecords(context.Background(), 42, pagination.PaginationParams{Page: 1, PageSize: 20})

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), result.Total)
	require.Equal(t, int64(101), items[0].TaskID)
	require.Equal(t, "gpt-image-2", items[0].Model)
	require.Equal(t, 0.75, items[0].ActualCost)
	require.Equal(t, 9.25, items[0].BalanceSnapshot)
	require.JSONEq(t, `{"artifact_count":2}`, items[0].MetadataJSON)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImageWorkspaceRepositoryGetWorkerStatusAggregatesQueue(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := NewImageWorkspaceRepository(db)
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\) AS total_count").
		WithArgs(
			service.ImageWorkspaceTaskStatusQueued,
			service.ImageWorkspaceTaskStatusRunning,
			service.ImageWorkspaceTaskStatusFailed,
			service.ImageWorkspaceTaskStatusSucceeded,
			service.ImageWorkspaceTaskStatusCancelled,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_count",
			"queued_count",
			"running_count",
			"stale_running_count",
			"failed_count",
			"recent_failed_count",
			"succeeded_count",
			"cancelled_count",
			"artifact_count",
			"last_task_updated_at",
			"last_failed_at",
			"last_failure_message",
			"oldest_queued_at",
		}).AddRow(
			int64(7),
			int64(2),
			int64(1),
			int64(1),
			int64(1),
			int64(1),
			int64(2),
			int64(1),
			int64(5),
			now,
			now.Add(-time.Minute),
			"upstream 404",
			now.Add(-6*time.Minute),
		))

	status, err := repo.GetWorkerStatus(context.Background())

	require.NoError(t, err)
	require.Equal(t, int64(7), status.TotalCount)
	require.Equal(t, int64(2), status.QueuedCount)
	require.Equal(t, int64(1), status.RunningCount)
	require.Equal(t, int64(1), status.StaleRunningCount)
	require.Equal(t, int64(1), status.RecentFailedCount)
	require.Equal(t, int64(5), status.ArtifactCount)
	require.NotNil(t, status.LastTaskUpdatedAt)
	require.NotNil(t, status.LastFailedAt)
	require.Equal(t, "upstream 404", status.LastFailureMessage)
	require.NotNil(t, status.OldestQueuedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func imageWorkspaceTaskRows(_ time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "user_id", "status", "prompt", "negative_prompt", "model", "provider", "size", "quality", "style",
		"seed", "batch_size", "template_id", "worker_lease_until", "cost_estimate", "balance_snapshot",
		"error_message", "result_json", "created_at", "updated_at",
	})
}
