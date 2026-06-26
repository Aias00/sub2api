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

	mock.ExpectQuery("INSERT INTO wechat_articles").
		WithArgs(article.UserID, article.AccountFakeID, article.SourceType, article.Title, article.Author, article.Link, article.Cover, article.Digest, article.PublishAt, article.IsOriginal, article.IsPaySubscribe, article.ContentStatus, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(9), now, now))

	err := repo.UpsertArticle(context.Background(), article)
	require.NoError(t, err)
	require.Equal(t, int64(9), article.ID)
	require.Equal(t, now, article.CreatedAt)
	require.Equal(t, now, article.UpdatedAt)
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
		FormatsJSON:          `["html"]`,
		PayloadJSON:          `{"article_ids":[3],"formats":["html"]}`,
		ResultManifestJSON:   "{}",
		RetentionDays:        7,
	}

	mock.ExpectQuery("INSERT INTO wechat_export_tasks").
		WithArgs(task.UserID, task.Status, task.SelectedArticleCount, task.FormatsJSON, task.IncludeEngagement, task.PayloadJSON, task.ResultManifestJSON, task.RetentionDays).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(11), now, now))

	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)
	require.Equal(t, int64(11), task.ID)

	articleRows := sqlmock.NewRows([]string{
		"id", "user_id", "account_fakeid", "source_type", "title", "author", "link", "cover", "digest",
		"publish_at", "is_original", "is_pay_subscribe", "content_status", "metadata_json", "created_at", "updated_at",
	}).AddRow(int64(3), int64(42), "", service.WeChatArticleSourceDirectLink, "Demo", "", "https://mp.weixin.qq.com/s/demo", "", "", nil, false, false, "pending", []byte(`{}`), now, now)

	mock.ExpectQuery("UPDATE wechat_export_tasks AS tasks").
		WithArgs(service.WeChatExportTaskStatusQueued, service.WeChatExportTaskStatusRunning, int64(60), service.WeChatExportTaskStatusRunning).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "created_at", "updated_at",
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusRunning, 1, 0, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[3],"formats":["html"]}`), []byte(`{}`), "", now.Add(time.Minute), 7, nil, now, now))
	mock.ExpectQuery("SELECT id, user_id, account_fakeid, source_type").
		WithArgs(int64(42), int64(3)).
		WillReturnRows(articleRows)

	claimed, articles, err := repo.ClaimNextTask(context.Background(), 60)
	require.NoError(t, err)
	require.Equal(t, int64(11), claimed.ID)
	require.Equal(t, []int64{3}, claimed.ArticleIDs)
	require.Len(t, articles, 1)
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
	mock.ExpectQuery("UPDATE wechat_export_tasks").
		WithArgs(service.WeChatExportTaskStatusCompleted, 0, `{"ok":true}`, int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "selected_article_count", "successful_article_count", "failed_article_count",
			"formats_json", "include_engagement", "payload_json", "result_manifest_json", "error_message",
			"worker_lease_until", "retention_days", "expires_at", "created_at", "updated_at",
		}).AddRow(int64(11), int64(42), service.WeChatExportTaskStatusCompleted, 1, 1, 0, []byte(`["html"]`), false, []byte(`{"article_ids":[3]}`), []byte(`{"ok":true}`), "", nil, 7, nil, now, now))
	mock.ExpectQuery("INSERT INTO wechat_export_artifacts").
		WithArgs(int64(11), int64(42), artifact.Format, artifact.StorageProvider, artifact.StorageKey, artifact.DownloadURL, artifact.FileName, artifact.FileSize, artifact.Checksum, artifact.ExpiresAt).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(99), now, now))
	mock.ExpectCommit()

	completed, err := repo.CompleteTask(context.Background(), 11, []service.WeChatExportArtifact{artifact}, `{"ok":true}`)
	require.NoError(t, err)
	require.Equal(t, service.WeChatExportTaskStatusCompleted, completed.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWeChatExportTaskPayloadRoundTrip(t *testing.T) {
	payload := wechatExportTaskPayload{ArticleIDs: []int64{1, 2}, Formats: []service.WeChatExportFormat{service.WeChatExportFormatJSON}}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	task := service.WeChatExportTask{PayloadJSON: string(raw), FormatsJSON: `["json"]`}
	require.NoError(t, hydrateWeChatTaskPayload(&task))
	require.Equal(t, []int64{1, 2}, task.ArticleIDs)
	require.Equal(t, []service.WeChatExportFormat{service.WeChatExportFormatJSON}, task.Formats)
}

func TestWeChatExportRepositoryListArticlesEmpty(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &wechatExportRepository{db: db, sql: db}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM wechat_articles").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	items, result, err := repo.ListArticles(context.Background(), 42, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Empty(t, items)
	require.Equal(t, int64(0), result.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}
