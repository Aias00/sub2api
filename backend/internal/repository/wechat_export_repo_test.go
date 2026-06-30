package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestWeChatExportRepositoryUpsertArticle(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	article := &service.WeChatArticle{
		UserID:        42,
		SourceType:    service.WeChatArticleSourceDirectLink,
		Title:         "Demo",
		Link:          "https://mp.weixin.qq.com/s/demo",
		ContentStatus: "pending",
		MetadataJSON:  "{}",
	}

	mock.ExpectQuery("WITH public_account AS .*INSERT INTO wechat_public_articles").
		WithArgs(article.UserID, article.AccountFakeID, article.SourceType, article.Title, article.Author, article.Link, article.Cover, article.Digest, article.PublishAt, article.IsOriginal, article.IsPaySubscribe, article.ContentStatus, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(9), now, now))

	err := repo.UpsertArticle(context.Background(), article)
	require.NoError(t, err)
	require.Equal(t, int64(9), article.ID)
	require.Equal(t, now, article.CreatedAt)
	require.Equal(t, now, article.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryUpdateArticleEnrichment(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	article := &service.WeChatArticle{
		ID:             9,
		UserID:         42,
		AccountFakeID:  "biz-demo",
		Title:          "Parsed title",
		Author:         "Parsed author",
		Cover:          "https://example.com/cover.png",
		Digest:         "Digest",
		IsOriginal:     true,
		IsPaySubscribe: true,
		ContentStatus:  "fetched",
		MetadataJSON:   `{"source":"worker"}`,
	}

	mock.ExpectQuery("WITH public_account AS .*UPDATE wechat_public_articles").
		WithArgs(article.ID, article.UserID, article.AccountFakeID, article.Title, article.Author, article.Cover, article.Digest, article.PublishAt, article.IsOriginal, article.IsPaySubscribe, article.ContentStatus, article.MetadataJSON).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))

	require.NoError(t, repo.UpdateArticleEnrichment(context.Background(), article))
	require.Equal(t, now, article.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryAccountLifecycle(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	account := &service.WeChatAccount{
		UserID:      42,
		FakeID:      "biz-001",
		Nickname:    "Demo Account",
		Alias:       "demo",
		Description: "local account",
		IsActive:    true,
	}

	mock.ExpectQuery("WITH public_account AS .*INSERT INTO wechat_public_accounts").
		WithArgs(account.UserID, account.FakeID, account.Nickname, account.Alias, account.Avatar, account.Description, account.IsActive).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(12), now, now))

	require.NoError(t, repo.UpsertAccount(context.Background(), account))
	require.Equal(t, int64(12), account.ID)

	mock.ExpectQuery("FROM wechat_account_bindings binding").
		WithArgs(int64(42), 20, "%Demo%").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "fakeid", "nickname", "alias", "avatar", "description",
			"is_active", "last_synced_at", "created_at", "updated_at",
		}).AddRow(int64(12), int64(42), "biz-001", "Demo Account", "demo", "", "local account", true, nil, now, now))

	items, err := repo.SearchAccounts(context.Background(), 42, "Demo", 20)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "biz-001", items[0].FakeID)

	mock.ExpectQuery("UPDATE wechat_account_bindings").
		WithArgs(int64(42), "biz-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "fakeid", "nickname", "alias", "avatar", "description",
			"is_active", "last_synced_at", "created_at", "updated_at",
		}).AddRow(int64(12), int64(42), "biz-001", "Demo Account", "demo", "", "local account", true, now, now, now))

	synced, err := repo.MarkAccountSynced(context.Background(), 42, "biz-001")
	require.NoError(t, err)
	require.NotNil(t, synced.LastSyncedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryGetActiveSessionPrefersReadyOverPending(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 30, 7, 29, 0, 0, time.UTC)

	mock.ExpectQuery("ORDER BY CASE WHEN status = \\$4 THEN 0 ELSE 1 END, updated_at DESC, id DESC").
		WithArgs(int64(42), service.WeChatSessionStatusPending, service.WeChatSessionStatusScanConfirmed, service.WeChatSessionStatusReady).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "login_token", "cookies_encrypted", "login_account_name",
			"last_validated_at", "expires_at", "created_at", "updated_at",
		}).AddRow(int64(7), int64(42), service.WeChatSessionStatusReady, "ready-token", "cookie", "ready account", now, now.Add(4*24*time.Hour), now.Add(-time.Hour), now))

	session, err := repo.GetActiveSession(context.Background(), 42)

	require.NoError(t, err)
	require.Equal(t, service.WeChatSessionStatusReady, session.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryExpireLoginAttemptSessionsPreservesReady(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}

	mock.ExpectExec("UPDATE wechat_sessions").
		WithArgs(service.WeChatSessionStatusExpired, int64(42), service.WeChatSessionStatusPending, service.WeChatSessionStatusScanConfirmed).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.ExpireLoginAttemptSessions(context.Background(), 42))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryCreateAndClaimTask(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	task := &service.WeChatExportTask{
		UserID:               42,
		Status:               service.WeChatExportTaskStatusQueued,
		ArticleIDs:           []int64{3},
		Formats:              []service.WeChatExportFormat{service.WeChatExportFormatHTML},
		SelectedArticleCount: 1,
		PayloadJSON:          `{"article_ids":[3],"formats":["html"]}`,
		ResultManifestJSON:   "{}",
		RetentionDays:        7,
		CostEstimate:         0,
		BalanceSnapshot:      0,
	}

	// Repository will extract formats from PayloadJSON and store in formats_json for database
	formatsForDB := []byte(`["html"]`)

	mock.ExpectQuery("INSERT INTO wechat_export_tasks").
		WithArgs(task.UserID, task.Status, task.SelectedArticleCount, formatsForDB, task.IncludeEngagement, task.PayloadJSON, task.ResultManifestJSON, task.RetentionDays, task.CostEstimate, task.BalanceSnapshot).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(11), now, now))
	mock.ExpectQuery("INSERT INTO wechat_export_task_logs").
		WithArgs(int64(11), task.UserID, "task_created", task.Status, "Task queued for export.", "{}").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(12), now))

	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)
	require.Equal(t, int64(11), task.ID)

	articleRows := sqlmock.NewRows([]string{
		"id", "user_id", "account_fakeid", "source_type", "title", "author", "link", "cover", "digest",
		"publish_at", "is_original", "is_pay_subscribe", "content_status", "metadata_json", "created_at", "updated_at",
	}).AddRow(int64(3), int64(42), "", service.WeChatArticleSourceDirectLink, "Demo", "", "https://mp.weixin.qq.com/s/demo", "", "", nil, false, false, "pending", []byte(`{}`), now, now)

	mock.ExpectQuery("UPDATE wechat_export_tasks AS tasks").
		WithArgs(service.WeChatExportTaskStatusQueued, service.WeChatExportTaskStatusRunning, int64(60), service.WeChatExportTaskStatusRunning, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at", // Phase 6：新增字段
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusRunning, 1, 0, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[3],"formats":["html"]}`), []byte(`{}`), "", now.Add(time.Minute), 7, nil, 0.0, 0.0,
			"test_lease_token_12345", "test_run_id_67890", now, now)) // Phase 6：新增token和run_id
	mock.ExpectQuery("INSERT INTO wechat_export_task_logs").
		WithArgs(int64(11), int64(42), "task_claimed", service.WeChatExportTaskStatusRunning, "Worker claimed the task.", "{}").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(13), now))
	mock.ExpectQuery("FROM wechat_article_bindings binding").
		WithArgs(int64(42), int64(3)).
		WillReturnRows(articleRows)

	claimed, articles, leaseToken, err := repo.ClaimNextTask(context.Background(), 60) // Phase 6：新增leaseToken返回值
	require.NoError(t, err)
	require.Equal(t, int64(11), claimed.ID)
	require.Equal(t, []int64{3}, claimed.ArticleIDs)
	require.Len(t, articles, 1)
	require.NotEmpty(t, leaseToken, "lease token should be returned") // Phase 6：验证token非空
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryCompleteTaskStoresArtifacts(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	artifact := service.WeChatExportArtifact{
		Format:      "html",
		StorageKey:  "/tmp/demo.html",
		DownloadURL: "/api/v1/wechat/artifacts/99/download",
		FileName:    "demo.html",
		FileSize:    88,
	}

	mock.ExpectBegin()
	// Phase 6：新增SELECT验证lease_token的ExpectQuery
	mock.ExpectQuery("SELECT id, user_id, status.*FROM wechat_export_tasks WHERE id = \\$1 FOR UPDATE").
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at",
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusRunning, 1, 0, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[3]}`), []byte(`{}`), "", nil, 7, nil, 0.0, 0.0, "test_lease_token", "", now, now))
	mock.ExpectQuery("UPDATE wechat_export_tasks").
		WithArgs(service.WeChatExportTaskStatusCompleted, 0, `{"ok":true}`, int64(11), service.WeChatExportTaskStatusRunning).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at", // Phase 6：新增字段
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusCompleted, 1, 1, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[3]}`), []byte(`{"ok":true}`), "", nil, 7, nil, 0.0, 0.0, "", "", now, now)) // Phase 6：token清空
	// CostEstimate == 0, actualCost == 0, so adjustment == 0 — no balance update expected
	mock.ExpectExec("INSERT INTO wechat_export_usage_records").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO wechat_export_artifacts").
		WithArgs(int64(11), int64(42), artifact.Format, artifact.StorageProvider, artifact.StorageKey, artifact.DownloadURL, artifact.FileName, artifact.FileSize, artifact.Checksum, artifact.ExpiresAt).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(99), now, now))
	mock.ExpectQuery("INSERT INTO wechat_export_task_logs").
		WithArgs(int64(11), int64(42), "task_completed", service.WeChatExportTaskStatusCompleted, "Export completed with 1 artifact(s) and 0 failed article(s).", `{"artifact_count":1,"failed_article_count":0}`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(100), now))
	mock.ExpectCommit()

	completed, err := repo.CompleteTask(context.Background(), 11, "test_lease_token", []service.WeChatExportArtifact{artifact}, `{"ok":true}`, 0, 0.0) // Phase 6：新增leaseToken参数
	require.NoError(t, err)
	require.Equal(t, service.WeChatExportTaskStatusCompleted, completed.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryGetTaskAllowsWorkerInternalLookup(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectQuery("FROM wechat_export_tasks\\s+WHERE id = \\$1").
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at",
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusRunning, 1, 0, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[3]}`), []byte(`{}`), "", nil, 7, nil, 0.10, 99.90, "", "", now, now))

	task, err := repo.GetTask(context.Background(), 0, 11)
	require.NoError(t, err)
	require.Equal(t, int64(11), task.ID)
	require.Equal(t, int64(42), task.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryGetWorkerStatusAggregatesMetrics(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	oldestQueued := now.Add(-10 * time.Minute)

	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\) AS total_count").
		WithArgs(
			int64(42),
			service.WeChatExportTaskStatusQueued,
			service.WeChatExportTaskStatusRunning,
			service.WeChatExportTaskStatusFailed,
			service.WeChatExportTaskStatusCompletedWithErrors,
			service.WeChatExportTaskStatusCompleted,
			service.WeChatExportTaskStatusCancelled,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_count",
			"queued_count",
			"running_count",
			"stale_running_count",
			"failed_count",
			"completed_count",
			"cancelled_count",
			"last_task_updated_at",
			"oldest_queued_at",
		}).AddRow(int64(9), int64(2), int64(1), int64(1), int64(3), int64(2), int64(1), now, oldestQueued))

	status, err := repo.GetWorkerStatus(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, int64(9), status.TotalCount)
	require.Equal(t, int64(2), status.QueuedCount)
	require.Equal(t, int64(1), status.RunningCount)
	require.Equal(t, int64(1), status.StaleRunningCount)
	require.Equal(t, int64(3), status.FailedCount)
	require.Equal(t, int64(2), status.CompletedCount)
	require.Equal(t, int64(1), status.CancelledCount)
	require.Equal(t, now, *status.LastTaskUpdatedAt)
	require.Equal(t, oldestQueued, *status.OldestQueuedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryCompleteTaskWithErrors(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectBegin()
	// Phase 6：新增SELECT验证lease_token的ExpectQuery
	mock.ExpectQuery("SELECT id, user_id, status.*FROM wechat_export_tasks WHERE id = \\$1 FOR UPDATE").
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at",
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusRunning, 3, 0, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[3,4,5]}`), []byte(`{}`), "", nil, 7, nil, 0.0, 0.0, "test_lease_token", "", now, now))
	mock.ExpectQuery("UPDATE wechat_export_tasks").
		WithArgs(service.WeChatExportTaskStatusCompletedWithErrors, 1, `{"ok":true}`, int64(11), service.WeChatExportTaskStatusRunning).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at", // Phase 6：新增字段
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusCompletedWithErrors, 3, 2, 1, []byte(`["html"]`), false, []byte(`{"article_ids":[3,4,5]}`), []byte(`{"ok":true}`), "", nil, 7, nil, 0.0, 0.0, "", "", now, now)) // Phase 6：token清空
	// CostEstimate == 0, actualCost == 0, so adjustment == 0 — no balance update expected
	mock.ExpectExec("INSERT INTO wechat_export_usage_records").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO wechat_export_task_logs").
		WithArgs(int64(11), int64(42), "task_completed", service.WeChatExportTaskStatusCompletedWithErrors, "Export completed with 0 artifact(s) and 1 failed article(s).", `{"artifact_count":0,"failed_article_count":1}`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(100), now))
	mock.ExpectCommit()

	completed, err := repo.CompleteTask(context.Background(), 11, "test_lease_token", nil, `{"ok":true}`, 1, 0.0) // Phase 6：新增leaseToken参数
	require.NoError(t, err)
	require.Equal(t, service.WeChatExportTaskStatusCompletedWithErrors, completed.Status)
	require.Equal(t, 2, completed.SuccessfulArticleCount)
	require.Equal(t, 1, completed.FailedArticleCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryListTaskLogs(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectQuery("SELECT logs.id, logs.task_id, logs.user_id").
		WithArgs(int64(42), int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "task_id", "user_id", "event", "status", "message", "meta_json", "created_at",
		}).AddRow(int64(1), int64(11), int64(42), "task_created", service.WeChatExportTaskStatusQueued, "Task queued for export.", []byte(`{}`), now))

	logs, err := repo.ListTaskLogs(context.Background(), 42, 11)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "task_created", logs[0].Event)
	require.Equal(t, "{}", logs[0].MetaJSON)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryAddTaskLog(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	// Phase 6：新增SELECT验证lease_token的ExpectQuery
	mock.ExpectQuery("SELECT status, worker_lease_token, worker_lease_until FROM wechat_export_tasks WHERE id = \\$1").
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"status", "worker_lease_token", "worker_lease_until",
		}).AddRow(service.WeChatExportTaskStatusRunning, "test_lease_token", nil))

	mock.ExpectQuery("INSERT INTO wechat_export_task_logs").
		WithArgs(int64(11), "article_fetched", service.WeChatExportTaskStatusRunning, "Fetched article HTML.", `{"article_id":3}`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "task_id", "user_id", "event", "status", "message", "meta_json", "created_at",
		}).AddRow(int64(21), int64(11), int64(42), "article_fetched", service.WeChatExportTaskStatusRunning, "Fetched article HTML.", []byte(`{"article_id":3}`), now))

	// Phase 6：新增leaseToken参数
	log, err := repo.AddTaskLog(context.Background(), 11, "test_lease_token", service.WeChatExportTaskLog{
		Event:    "article_fetched",
		Status:   service.WeChatExportTaskStatusRunning,
		Message:  "Fetched article HTML.",
		MetaJSON: `{"article_id":3}`,
	})
	require.NoError(t, err)
	require.Equal(t, int64(42), log.UserID)
	require.Equal(t, "article_fetched", log.Event)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportTaskPayloadRoundTrip(t *testing.T) {
	payload := wechatExportTaskPayload{ArticleIDs: []int64{1, 2}, Formats: []service.WeChatExportFormat{service.WeChatExportFormatHTML}}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	task := service.WeChatExportTask{PayloadJSON: string(raw)}
	require.NoError(t, hydrateWeChatTaskPayload(&task))
	require.Equal(t, []int64{1, 2}, task.ArticleIDs)
	require.Equal(t, []service.WeChatExportFormat{service.WeChatExportFormatHTML}, task.Formats)
}

func TestWeChatExportRepositoryListArticlesEmpty(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM wechat_article_bindings").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	items, result, err := repo.ListArticles(context.Background(), 42, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Empty(t, items)
	require.Equal(t, int64(0), result.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryCreateTaskInsufficientBalance(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}

	task := &service.WeChatExportTask{
		UserID:               42,
		Status:               service.WeChatExportTaskStatusQueued,
		SelectedArticleCount: 3,
		IncludeEngagement:    false,
		PayloadJSON:          `{"article_ids":[1,2,3],"formats":["html","markdown"]}`,
		ResultManifestJSON:   "{}",
		RetentionDays:        7,
		CostEstimate:         0.45, // 3 articles * 3 formats * 0.05
	}

	// 预留余额失败（余额不足）
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE users SET balance = balance -").
		WithArgs(0.45, int64(42)).
		WillReturnError(sqlmock.ErrCancelled) // 模拟余额不足
	mock.ExpectRollback()

	err := repo.CreateTask(context.Background(), task)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryCreateTaskWithBalanceReservation(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	task := &service.WeChatExportTask{
		UserID:               42,
		Status:               service.WeChatExportTaskStatusQueued,
		SelectedArticleCount: 3,
		IncludeEngagement:    false,
		PayloadJSON:          `{"article_ids":[1,2,3],"formats":["html","markdown"]}`,
		ResultManifestJSON:   "{}",
		RetentionDays:        7,
		CostEstimate:         0.45,
	}

	// 预留余额成功
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE users SET balance = balance -").
		WithArgs(0.45, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(9.55)) // 余额从10扣到9.55
	// 插入任务记录
	mock.ExpectQuery("INSERT INTO wechat_export_tasks").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(11), now, now))
	// 插入任务日志（使用Query因为有RETURNING）
	mock.ExpectQuery("INSERT INTO wechat_export_task_logs").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(1), now))
	mock.ExpectCommit()

	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)
	require.Equal(t, int64(11), task.ID)
	require.Equal(t, 9.55, task.BalanceSnapshot)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryCreateTaskZeroCost(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	task := &service.WeChatExportTask{
		UserID:               42,
		Status:               service.WeChatExportTaskStatusQueued,
		SelectedArticleCount: 0,
		IncludeEngagement:    false,
		PayloadJSON:          `{"article_ids":[],"formats":[]}`,
		ResultManifestJSON:   "{}",
		RetentionDays:        7,
		CostEstimate:         0, // 费用为0，不需要预留余额
	}

	// CostEstimate == 0，不开启事务，直接插入
	mock.ExpectQuery("INSERT INTO wechat_export_tasks").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(11), now, now))
	// 插入任务日志（使用Query因为有RETURNING）
	mock.ExpectQuery("INSERT INTO wechat_export_task_logs").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(1), now))

	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)
	require.Equal(t, int64(11), task.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// Phase 3: Worker Lease验证测试
func TestWeChatExportRepositoryCompleteTask_InvalidLeaseToken(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectBegin()
	// SELECT验证lease_token（返回正确的token）
	mock.ExpectQuery("SELECT id.*FROM wechat_export_tasks WHERE id = \\$1 FOR UPDATE").
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at",
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusRunning, 3, 0, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[1,2,3],"formats":["html"]}`), []byte(`{}`), "", nil, 7, nil, 0.45, 9.55, "correct_lease_token", "", now, now))
	mock.ExpectRollback()

	// 使用错误的lease_token
	task, err := repo.CompleteTask(context.Background(), int64(11), "wrong_lease_token", []service.WeChatExportArtifact{}, "{}", 0, 0.45)
	require.Error(t, err)
	require.Nil(t, task)
	require.Contains(t, err.Error(), "LEASE_TOKEN_MISMATCH")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryCompleteTask_ExpiredLease(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)
	expiredLease := now.Add(-time.Hour) // Lease已过期1小时

	mock.ExpectBegin()
	// SELECT验证lease_token（返回已过期的lease）
	mock.ExpectQuery("SELECT id.*FROM wechat_export_tasks WHERE id = \\$1 FOR UPDATE").
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at",
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusRunning, 3, 0, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[1,2,3],"formats":["html"]}`), []byte(`{}`), "", &expiredLease, 7, nil, 0.45, 9.55, "test_lease_token", "", now, now))
	mock.ExpectRollback()

	// 使用正确的lease_token但lease已过期
	task, err := repo.CompleteTask(context.Background(), int64(11), "test_lease_token", []service.WeChatExportArtifact{}, "{}", 0, 0.45)
	require.Error(t, err)
	require.Nil(t, task)
	require.Contains(t, err.Error(), "LEASE_EXPIRED")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportRepositoryFailTask_InvalidLeaseToken(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}
	now := time.Date(2026, 6, 26, 1, 2, 3, 0, time.UTC)

	mock.ExpectBegin()
	// SELECT验证lease_token
	mock.ExpectQuery("SELECT id.*FROM wechat_export_tasks WHERE id = \\$1 FOR UPDATE").
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "cost_estimate", "balance_snapshot",
			"worker_lease_token", "worker_run_id", "created_at", "updated_at",
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusRunning, 3, 0, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[1,2,3],"formats":["html"]}`), []byte(`{}`), "", nil, 7, nil, 0.45, 9.55, "correct_lease_token", "", now, now))
	mock.ExpectRollback()

	// 使用错误的lease_token调用FailTask
	task, err := repo.FailTask(context.Background(), int64(11), "wrong_lease_token", "Test error")
	require.Error(t, err)
	require.Nil(t, task)
	require.Contains(t, err.Error(), "LEASE_TOKEN_MISMATCH")
	require.NoError(t, mock.ExpectationsWereMet())
}
