package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	weChatDirectImportProbe = nil
	os.Exit(m.Run())
}

type wechatExportRepoFake struct {
	sessions  map[int64]WeChatSession
	accounts  map[int64]WeChatAccount
	articles  map[int64]WeChatArticle
	tasks     map[int64]WeChatExportTask
	artifacts map[int64]WeChatExportArtifact
	taskLogs  map[int64][]WeChatExportTaskLog
	nextID    int64
	// Mock error fields for testing
	createTaskErr error
}

func newWeChatExportRepoFake() *wechatExportRepoFake {
	return &wechatExportRepoFake{
		sessions:  map[int64]WeChatSession{},
		accounts:  map[int64]WeChatAccount{},
		articles:  map[int64]WeChatArticle{},
		tasks:     map[int64]WeChatExportTask{},
		artifacts: map[int64]WeChatExportArtifact{},
		taskLogs:  map[int64][]WeChatExportTaskLog{},
		nextID:    1,
	}
}

func (r *wechatExportRepoFake) GetActiveSession(ctx context.Context, userID int64) (*WeChatSession, error) {
	for _, session := range r.sessions {
		if session.UserID == userID && session.Status != WeChatSessionStatusExpired {
			return &session, nil
		}
	}
	return nil, nil
}

func (r *wechatExportRepoFake) CreateSession(ctx context.Context, session *WeChatSession) error {
	session.ID = r.nextID
	r.nextID++
	r.sessions[session.ID] = *session
	return nil
}

func (r *wechatExportRepoFake) UpdateSession(ctx context.Context, session *WeChatSession) error {
	if session == nil {
		return nil
	}
	existing, ok := r.sessions[session.ID]
	if !ok || existing.UserID != session.UserID {
		return ErrWeChatSessionNotFound
	}
	r.sessions[session.ID] = *session
	return nil
}

func (r *wechatExportRepoFake) GetSession(ctx context.Context, userID int64, sessionID int64) (*WeChatSession, error) {
	session, ok := r.sessions[sessionID]
	if !ok || session.UserID != userID {
		return nil, ErrWeChatSessionNotFound
	}
	return &session, nil
}

func (r *wechatExportRepoFake) ExpireUserSessions(ctx context.Context, userID int64) error {
	for id, session := range r.sessions {
		if session.UserID == userID {
			session.Status = WeChatSessionStatusExpired
			r.sessions[id] = session
		}
	}
	return nil
}

func (r *wechatExportRepoFake) ExpireLoginAttemptSessions(ctx context.Context, userID int64) error {
	for id, session := range r.sessions {
		if session.UserID == userID && (session.Status == WeChatSessionStatusPending || session.Status == WeChatSessionStatusScanConfirmed) {
			session.Status = WeChatSessionStatusExpired
			r.sessions[id] = session
		}
	}
	return nil
}

func (r *wechatExportRepoFake) SearchAccounts(ctx context.Context, userID int64, query string, limit int) ([]WeChatAccount, error) {
	items := make([]WeChatAccount, 0, len(r.accounts))
	for _, account := range r.accounts {
		if account.UserID == userID {
			items = append(items, account)
		}
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (r *wechatExportRepoFake) GetAccount(ctx context.Context, userID int64, fakeID string) (*WeChatAccount, error) {
	for _, account := range r.accounts {
		if account.UserID == userID && account.FakeID == fakeID {
			return &account, nil
		}
	}
	return nil, ErrWeChatAccountNotFound
}

func (r *wechatExportRepoFake) UpsertAccount(ctx context.Context, account *WeChatAccount) error {
	for id, existing := range r.accounts {
		if existing.UserID == account.UserID && existing.FakeID == account.FakeID {
			account.ID = id
			r.accounts[id] = *account
			return nil
		}
	}
	account.ID = r.nextID
	r.nextID++
	r.accounts[account.ID] = *account
	return nil
}

func (r *wechatExportRepoFake) MarkAccountSynced(ctx context.Context, userID int64, fakeID string) (*WeChatAccount, error) {
	for id, account := range r.accounts {
		if account.UserID == userID && account.FakeID == fakeID {
			r.accounts[id] = account
			return &account, nil
		}
	}
	return nil, ErrWeChatAccountNotFound
}

func (r *wechatExportRepoFake) UpsertArticle(ctx context.Context, article *WeChatArticle) error {
	if article.ID == 0 {
		article.ID = r.nextID
		r.nextID++
	}
	r.articles[article.ID] = *article
	return nil
}

func (r *wechatExportRepoFake) UpdateArticleEnrichment(ctx context.Context, article *WeChatArticle) error {
	existing, ok := r.articles[article.ID]
	if !ok || existing.UserID != article.UserID {
		return ErrWeChatArticleNotFound
	}
	if article.AccountFakeID != "" {
		existing.AccountFakeID = article.AccountFakeID
	}
	if article.Title != "" {
		existing.Title = article.Title
	}
	existing.Author = article.Author
	existing.Cover = article.Cover
	existing.Digest = article.Digest
	existing.PublishAt = article.PublishAt
	existing.IsOriginal = article.IsOriginal
	existing.IsPaySubscribe = article.IsPaySubscribe
	existing.ContentStatus = article.ContentStatus
	existing.MetadataJSON = article.MetadataJSON
	r.articles[article.ID] = existing
	*article = existing
	return nil
}

func (r *wechatExportRepoFake) ListArticles(ctx context.Context, userID int64, params pagination.PaginationParams) ([]WeChatArticle, *pagination.PaginationResult, error) {
	items := make([]WeChatArticle, 0, len(r.articles))
	for _, article := range r.articles {
		if article.UserID == userID {
			items = append(items, article)
		}
	}
	return items, &pagination.PaginationResult{Total: int64(len(items)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (r *wechatExportRepoFake) GetArticleByID(ctx context.Context, articleID int64) (*WeChatArticle, error) {
	article, ok := r.articles[articleID]
	if !ok {
		return nil, ErrWeChatArticleNotFound
	}
	return &article, nil
}

func (r *wechatExportRepoFake) ListArticlesByIDs(ctx context.Context, userID int64, articleIDs []int64) ([]WeChatArticle, error) {
	items := make([]WeChatArticle, 0, len(articleIDs))
	for _, id := range articleIDs {
		article, ok := r.articles[id]
		if ok && article.UserID == userID {
			items = append(items, article)
		}
	}
	return items, nil
}

func (r *wechatExportRepoFake) CreateTask(ctx context.Context, task *WeChatExportTask) error {
	// Return mock error if set (for testing)
	if r.createTaskErr != nil {
		return r.createTaskErr
	}
	task.ID = r.nextID
	r.nextID++
	r.tasks[task.ID] = *task
	r.appendTaskLog(task.ID, task.UserID, "task_created", task.Status, "Task queued for export.")
	return nil
}

func (r *wechatExportRepoFake) ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams) ([]WeChatExportTask, *pagination.PaginationResult, error) {
	items := make([]WeChatExportTask, 0, len(r.tasks))
	for _, task := range r.tasks {
		if task.UserID == userID {
			items = append(items, task)
		}
	}
	return items, &pagination.PaginationResult{Total: int64(len(items)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (r *wechatExportRepoFake) GetWorkerStatus(ctx context.Context, userID int64) (*WeChatExportWorkerStatus, error) {
	status := &WeChatExportWorkerStatus{}
	for _, task := range r.tasks {
		if task.UserID != userID {
			continue
		}
		status.TotalCount++
		switch task.Status {
		case WeChatExportTaskStatusQueued:
			status.QueuedCount++
			if status.OldestQueuedAt == nil || task.CreatedAt.Before(*status.OldestQueuedAt) {
				createdAt := task.CreatedAt
				status.OldestQueuedAt = &createdAt
			}
		case WeChatExportTaskStatusRunning:
			status.RunningCount++
			if task.WorkerLeaseUntil == nil || task.WorkerLeaseUntil.Before(time.Now().UTC()) {
				status.StaleRunningCount++
			}
		case WeChatExportTaskStatusFailed, WeChatExportTaskStatusCompletedWithErrors:
			status.FailedCount++
		case WeChatExportTaskStatusCompleted:
			status.CompletedCount++
		case WeChatExportTaskStatusCancelled:
			status.CancelledCount++
		}
		if status.LastTaskUpdatedAt == nil || task.UpdatedAt.After(*status.LastTaskUpdatedAt) {
			updatedAt := task.UpdatedAt
			status.LastTaskUpdatedAt = &updatedAt
		}
	}
	return status, nil
}

func (r *wechatExportRepoFake) GetTask(ctx context.Context, userID int64, taskID int64) (*WeChatExportTask, error) {
	task, ok := r.tasks[taskID]
	if !ok || (userID > 0 && task.UserID != userID) {
		return nil, ErrWeChatTaskNotFound
	}
	return &task, nil
}

func (r *wechatExportRepoFake) CancelTask(ctx context.Context, userID int64, taskID int64) (*WeChatExportTask, error) {
	task, ok := r.tasks[taskID]
	if !ok || task.UserID != userID || (task.Status != WeChatExportTaskStatusQueued && task.Status != WeChatExportTaskStatusRunning) {
		return nil, ErrWeChatTaskNotFound
	}
	task.Status = WeChatExportTaskStatusCancelled
	task.ErrorMessage = "cancelled by user"
	task.WorkerLeaseUntil = nil
	r.tasks[taskID] = task
	r.appendTaskLog(task.ID, task.UserID, "task_cancelled", task.Status, task.ErrorMessage)
	return &task, nil
}

func (r *wechatExportRepoFake) RetryTask(ctx context.Context, userID int64, taskID int64) (*WeChatExportTask, error) {
	task, ok := r.tasks[taskID]
	if !ok || task.UserID != userID {
		return nil, ErrWeChatTaskNotFound
	}
	switch task.Status {
	case WeChatExportTaskStatusFailed, WeChatExportTaskStatusCompletedWithErrors, WeChatExportTaskStatusCancelled, WeChatExportTaskStatusCompleted:
	default:
		return nil, ErrWeChatTaskNotFound
	}
	task.Status = WeChatExportTaskStatusQueued
	task.SuccessfulArticleCount = 0
	task.FailedArticleCount = 0
	task.ResultManifestJSON = "{}"
	task.ErrorMessage = ""
	task.WorkerLeaseUntil = nil
	task.ExpiresAt = nil
	r.tasks[taskID] = task
	r.appendTaskLog(task.ID, task.UserID, "task_retried", task.Status, "Task reset and queued for retry.")
	for id, artifact := range r.artifacts {
		if artifact.UserID == userID && artifact.TaskID == taskID {
			now := time.Now().UTC()
			artifact.DeletedAt = &now
			r.artifacts[id] = artifact
		}
	}
	return &task, nil
}

func (r *wechatExportRepoFake) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*WeChatExportTask, []WeChatArticle, string, error) {
	for id, task := range r.tasks {
		if task.Status == WeChatExportTaskStatusQueued {
			task.Status = WeChatExportTaskStatusRunning
			r.tasks[id] = task
			r.appendTaskLog(task.ID, task.UserID, "task_claimed", task.Status, "Worker claimed the task.")
			articles, err := r.ListArticlesByIDs(ctx, task.UserID, task.ArticleIDs)
			return &task, articles, "test-lease-token", err
		}
	}
	return nil, nil, "", nil
}

func (r *wechatExportRepoFake) CompleteTask(ctx context.Context, taskID int64, leaseToken string, artifacts []WeChatExportArtifact, resultManifestJSON string, failedArticleCount int, actualCost float64) (*WeChatExportTask, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, ErrWeChatTaskNotFound
	}
	if failedArticleCount > 0 {
		task.Status = WeChatExportTaskStatusCompletedWithErrors
	} else {
		task.Status = WeChatExportTaskStatusCompleted
	}
	task.FailedArticleCount = failedArticleCount
	task.SuccessfulArticleCount = maxIntValue(task.SelectedArticleCount-failedArticleCount, 0)
	task.ResultManifestJSON = resultManifestJSON
	r.tasks[taskID] = task
	r.appendTaskLog(task.ID, task.UserID, "task_completed", task.Status, "Export completed.")
	for _, artifact := range artifacts {
		artifact.ID = r.nextID
		r.nextID++
		artifact.TaskID = taskID
		artifact.UserID = task.UserID
		r.artifacts[artifact.ID] = artifact
	}
	return &task, nil
}

func (r *wechatExportRepoFake) FailTask(ctx context.Context, taskID int64, leaseToken string, message string) (*WeChatExportTask, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, ErrWeChatTaskNotFound
	}
	task.Status = WeChatExportTaskStatusFailed
	task.ErrorMessage = message
	r.tasks[taskID] = task
	r.appendTaskLog(task.ID, task.UserID, "task_failed", task.Status, message)
	return &task, nil
}

func (r *wechatExportRepoFake) AddTaskLog(ctx context.Context, taskID int64, leaseToken string, log WeChatExportTaskLog) (*WeChatExportTaskLog, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, ErrWeChatTaskNotFound
	}
	log.ID = r.nextID
	r.nextID++
	log.TaskID = taskID
	log.UserID = task.UserID
	if log.MetaJSON == "" {
		log.MetaJSON = "{}"
	}
	log.CreatedAt = time.Now().UTC()
	r.taskLogs[taskID] = append(r.taskLogs[taskID], log)
	return &log, nil
}

func (r *wechatExportRepoFake) ListTaskLogs(ctx context.Context, userID int64, taskID int64) ([]WeChatExportTaskLog, error) {
	task, ok := r.tasks[taskID]
	if !ok || task.UserID != userID {
		return nil, ErrWeChatTaskNotFound
	}
	return append([]WeChatExportTaskLog(nil), r.taskLogs[taskID]...), nil
}

func (r *wechatExportRepoFake) appendTaskLog(taskID int64, userID int64, event string, status string, message string) {
	log := WeChatExportTaskLog{
		ID:        r.nextID,
		TaskID:    taskID,
		UserID:    userID,
		Event:     event,
		Status:    status,
		Message:   message,
		MetaJSON:  "{}",
		CreatedAt: time.Now().UTC(),
	}
	r.nextID++
	r.taskLogs[taskID] = append(r.taskLogs[taskID], log)
}

func (r *wechatExportRepoFake) ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]WeChatExportArtifact, error) {
	items := make([]WeChatExportArtifact, 0, len(r.artifacts))
	for _, artifact := range r.artifacts {
		if artifact.UserID == userID && artifact.TaskID == taskID && artifact.DeletedAt == nil {
			items = append(items, artifact)
		}
	}
	return items, nil
}

func (r *wechatExportRepoFake) GetArtifact(ctx context.Context, userID int64, artifactID int64) (*WeChatExportArtifact, error) {
	artifact, ok := r.artifacts[artifactID]
	if !ok || artifact.UserID != userID {
		return nil, ErrWeChatTaskNotFound
	}
	return &artifact, nil
}

func TestWeChatExportServiceImportLinkAndCreateTask(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)

	article, err := svc.ImportArticleLink(context.Background(), 42, " https://mp.weixin.qq.com/s/demo ")
	require.NoError(t, err)
	require.Equal(t, int64(42), article.UserID)
	require.Equal(t, WeChatArticleSourceDirectLink, article.SourceType)
	require.Equal(t, "https://mp.weixin.qq.com/s/demo", article.Link)

	task, err := svc.CreateTask(context.Background(), 42, CreateWeChatExportTaskInput{
		ArticleIDs:        []int64{article.ID},
		Formats:           []string{"markdown", "html", "markdown"},
		IncludeEngagement: true,
	})
	require.NoError(t, err)
	require.Equal(t, WeChatExportTaskStatusQueued, task.Status)
	require.Equal(t, 1, task.SelectedArticleCount)
	require.Equal(t, []WeChatExportFormat{WeChatExportFormatMarkdown, WeChatExportFormatHTML}, task.Formats)
	require.True(t, task.IncludeEngagement)
}

func TestWeChatExportServiceRejectsNonWeChatArticleURL(t *testing.T) {
	svc := NewWeChatExportService(newWeChatExportRepoFake(), nil)

	_, err := svc.ImportArticleLink(context.Background(), 42, "https://example.com/s/demo")
	require.ErrorIs(t, err, ErrWeChatInvalidInput)

	_, err = svc.ImportArticleLink(context.Background(), 42, "http://mp.weixin.qq.com/s/demo")
	require.ErrorIs(t, err, ErrWeChatInvalidInput)
}

func TestWeChatExportServiceImportLinkRejectsVerifyPage(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)
	previousProbe := weChatDirectImportProbe
	weChatDirectImportProbe = func(ctx context.Context, link string) error {
		return ErrWeChatArticleVerifyPage
	}
	t.Cleanup(func() {
		weChatDirectImportProbe = previousProbe
	})

	_, err := svc.ImportArticleLink(context.Background(), 42, "https://mp.weixin.qq.com/s?__biz=biz&mid=123&idx=1&sn=abc")

	require.ErrorIs(t, err, ErrWeChatArticleVerifyPage)
	require.Empty(t, repo.articles)
}

func TestIsWeChatVerifyPageHTML(t *testing.T) {
	require.True(t, isWeChatVerifyPageHTML(`<script>var PAGE_MID='mmbizwap:secitptpage/verify.html';</script>`))
	require.True(t, isWeChatVerifyPageHTML(`<script>window.cgiData={cap_sid:"sid",poc_token:"token",target_url:"/s/demo"}</script>`))
	require.False(t, isWeChatVerifyPageHTML(`<h1 id="activity-name">正常文章</h1><div id="js_content">正文</div>`))
}

func TestWeChatExportServiceBindSearchAndSyncAccount(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)

	account, err := svc.BindAccount(context.Background(), 42, BindWeChatAccountInput{
		FakeID:   "biz-001",
		Nickname: "Demo Account",
	})
	require.NoError(t, err)
	require.Equal(t, "Demo Account", account.Nickname)

	items, err := svc.SearchAccounts(context.Background(), 42, "demo", 20)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "biz-001", items[0].FakeID)

	_, result, err := svc.SyncAccount(context.Background(), 42, "biz-001")
	require.ErrorIs(t, err, ErrWeChatSessionNotReady)
	require.Nil(t, result)
}

func TestWeChatExportServiceSyncAccountRejectsUnboundFakeidInMain(t *testing.T) {
	// Regression test: SyncAccount must reject unbound fakeid without writing articles
	// This prevents orphan article records when sync is called before BindAccount
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)

	// Setup a ready session to ensure we test account check, not session check
	repo.sessions[1] = WeChatSession{
		ID:     1,
		UserID: 42,
		Status: WeChatSessionStatusReady,
	}

	// Do NOT call BindAccount - test with completely unbound fakeid
	account, result, err := svc.SyncAccount(context.Background(), 42, "missing-fakeid")

	// Must return ErrWeChatAccountNotFound, not ErrWeChatSessionNotReady or success
	require.ErrorIs(t, err, ErrWeChatAccountNotFound)
	require.Nil(t, account)
	require.Nil(t, result)

	// Critical: verify no articles were written for unbound fakeid
	require.Empty(t, repo.articles, "SyncAccount must not write articles for unbound fakeid")
}

func TestWeChatExportServiceSearchRemoteAccountsRequiresReadySession(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)

	_, err := svc.SearchRemoteAccounts(context.Background(), 42, "demo", 5)
	require.ErrorIs(t, err, ErrWeChatSessionNotReady)

	repo.sessions[1] = WeChatSession{
		ID:     1,
		UserID: 42,
		Status: WeChatSessionStatusPending,
	}
	_, err = svc.SearchRemoteAccounts(context.Background(), 42, "demo", 5)
	require.ErrorIs(t, err, ErrWeChatSessionNotReady)
}

func TestWeChatExportServiceEnrichArticleUpsertsAccountAndArticleMetadata(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)
	article, err := svc.ImportArticleLink(context.Background(), 7, "https://mp.weixin.qq.com/s/demo")
	require.NoError(t, err)

	enriched, err := svc.EnrichArticle(context.Background(), EnrichWeChatArticleInput{
		ArticleID:      article.ID,
		UserID:         7,
		AccountFakeID:  "biz-demo",
		AccountName:    "Demo Account",
		Title:          "Parsed title",
		Author:         "Parsed author",
		ContentStatus:  "fetched",
		MetadataJSON:   `{"source":"worker"}`,
		IsOriginal:     true,
		IsPaySubscribe: true,
	})
	require.NoError(t, err)
	require.Equal(t, "Parsed title", enriched.Title)
	require.Equal(t, "biz-demo", enriched.AccountFakeID)
	require.True(t, enriched.IsOriginal)
	require.True(t, enriched.IsPaySubscribe)

	accounts, err := svc.SearchAccounts(context.Background(), 7, "Demo", 20)
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	require.Equal(t, "Demo Account", accounts[0].Nickname)
}

func TestWeChatExportServiceFetchArticleEngagementRequiresReadySessionWithoutFailingTask(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)
	article, err := svc.ImportArticleLink(context.Background(), 7, "https://mp.weixin.qq.com/s?__biz=biz&mid=123&idx=1&sn=abc")
	require.NoError(t, err)

	result, err := svc.FetchArticleEngagement(context.Background(), FetchWeChatArticleEngagementInput{
		ArticleID:    article.ID,
		MetadataJSON: `{"appmsg_token":"token"}`,
	})
	require.NoError(t, err)
	require.Equal(t, "unavailable", result.Status)
	require.Contains(t, result.Message, "ready WeChat session")
}

func TestWeChatExportServiceRejectsUnknownArticles(t *testing.T) {
	svc := NewWeChatExportService(newWeChatExportRepoFake(), nil)

	_, err := svc.CreateTask(context.Background(), 42, CreateWeChatExportTaskInput{
		ArticleIDs: []int64{999},
		Formats:    []string{"html"},
	})

	require.ErrorIs(t, err, ErrWeChatArticleNotFound)
}

func TestWeChatExportServiceWorkerClaimAndComplete(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)
	article, err := svc.ImportArticleLink(context.Background(), 7, "https://mp.weixin.qq.com/s/demo")
	require.NoError(t, err)
	task, err := svc.CreateTask(context.Background(), 7, CreateWeChatExportTaskInput{
		ArticleIDs: []int64{article.ID},
		Formats:    []string{"html"},
	})
	require.NoError(t, err)

	claimed, articles, _, err := svc.ClaimNextTask(context.Background(), 60)
	require.NoError(t, err)
	require.Equal(t, task.ID, claimed.ID)
	require.Equal(t, WeChatExportTaskStatusRunning, claimed.Status)
	require.Len(t, articles, 1)

	completed, err := svc.CompleteTask(context.Background(), task.ID, "test-lease-token", CompleteWeChatExportTaskInput{
		Artifacts: []WeChatExportArtifactInput{{
			Format:      "html",
			StorageKey:  "/tmp/wechat-export.html",
			DownloadURL: "/api/v1/wechat/artifacts/1/download",
			FileName:    "wechat-export.html",
			FileSize:    128,
		}},
		ResultManifestJSON: `{"ok":true}`,
	})
	require.NoError(t, err)
	require.Equal(t, WeChatExportTaskStatusCompleted, completed.Status)

	logs, err := svc.ListTaskLogs(context.Background(), 7, task.ID)
	require.NoError(t, err)
	require.Len(t, logs, 3)
	require.Equal(t, "task_created", logs[0].Event)
	require.Equal(t, "task_claimed", logs[1].Event)
	require.Equal(t, "task_completed", logs[2].Event)

	artifacts, err := svc.ListArtifacts(context.Background(), 7, task.ID)
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	require.Equal(t, "wechat-export.html", artifacts[0].FileName)
}

func TestWeChatExportServiceAddsWorkerArticleLog(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)
	repo.tasks[1] = WeChatExportTask{ID: 1, UserID: 7, Status: WeChatExportTaskStatusRunning}

	log, err := svc.AddTaskLog(context.Background(), 1, "test-lease-token", AddWeChatExportTaskLogInput{
		Event:    "article_fetched",
		Status:   "running",
		Message:  "Fetched article HTML.",
		MetaJSON: `{"article_id":3,"title":"Demo"}`,
	})

	require.NoError(t, err)
	require.Equal(t, int64(7), log.UserID)
	require.Equal(t, "article_fetched", log.Event)
	require.JSONEq(t, `{"article_id":3,"title":"Demo"}`, log.MetaJSON)
}

func TestWeChatExportServiceWorkerStatusReportsQueuedAndStaleTasks(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)
	now := time.Now().UTC()
	repo.tasks[1] = WeChatExportTask{
		ID:        1,
		UserID:    7,
		Status:    WeChatExportTaskStatusQueued,
		CreatedAt: now.Add(-2 * time.Minute),
		UpdatedAt: now.Add(-2 * time.Minute),
	}
	expiredLease := now.Add(-time.Minute)
	repo.tasks[2] = WeChatExportTask{
		ID:               2,
		UserID:           7,
		Status:           WeChatExportTaskStatusRunning,
		WorkerLeaseUntil: &expiredLease,
		CreatedAt:        now.Add(-time.Minute),
		UpdatedAt:        now.Add(-time.Minute),
	}

	status, err := svc.GetWorkerStatus(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, int64(1), status.QueuedCount)
	require.Equal(t, int64(1), status.RunningCount)
	require.Equal(t, int64(1), status.StaleRunningCount)
	require.Equal(t, int64(2), status.TotalCount)
	require.Equal(t, "attention", status.Health)
	require.Contains(t, status.AttentionReasons, "stale_running_tasks")
	require.NotNil(t, status.OldestQueuedSeconds)
	require.GreaterOrEqual(t, *status.OldestQueuedSeconds, int64(100))
	require.NotNil(t, status.LastTaskAgeSeconds)
}

func TestWeChatExportServiceCompleteTaskWithErrors(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)
	articleOne, err := svc.ImportArticleLink(context.Background(), 7, "https://mp.weixin.qq.com/s/demo-one")
	require.NoError(t, err)
	articleTwo, err := svc.ImportArticleLink(context.Background(), 7, "https://mp.weixin.qq.com/s/demo-two")
	require.NoError(t, err)
	task, err := svc.CreateTask(context.Background(), 7, CreateWeChatExportTaskInput{
		ArticleIDs: []int64{articleOne.ID, articleTwo.ID},
		Formats:    []string{"html"},
	})
	require.NoError(t, err)

	completed, err := svc.CompleteTask(context.Background(), task.ID, "test-lease-token", CompleteWeChatExportTaskInput{
		FailedArticleCount: 1,
		ResultManifestJSON: `{"failed_articles":[{"link":"https://mp.weixin.qq.com/s/demo-two"}]}`,
	})
	require.NoError(t, err)
	require.Equal(t, WeChatExportTaskStatusCompletedWithErrors, completed.Status)
	require.Equal(t, 1, completed.SuccessfulArticleCount)
	require.Equal(t, 1, completed.FailedArticleCount)
}

func TestWeChatExportServiceCancelAndRetryTask(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)
	article, err := svc.ImportArticleLink(context.Background(), 7, "https://mp.weixin.qq.com/s/demo")
	require.NoError(t, err)
	task, err := svc.CreateTask(context.Background(), 7, CreateWeChatExportTaskInput{
		ArticleIDs: []int64{article.ID},
		Formats:    []string{"html"},
	})
	require.NoError(t, err)

	cancelled, err := svc.CancelTask(context.Background(), 7, task.ID)
	require.NoError(t, err)
	require.Equal(t, WeChatExportTaskStatusCancelled, cancelled.Status)

	retried, err := svc.RetryTask(context.Background(), 7, task.ID)
	require.NoError(t, err)
	require.Equal(t, WeChatExportTaskStatusQueued, retried.Status)
	require.Equal(t, 0, retried.SuccessfulArticleCount)
	require.Equal(t, 0, retried.FailedArticleCount)
}

func TestWeChatExportServiceRejectsExpiredArtifact(t *testing.T) {
	repo := newWeChatExportRepoFake()
	expiredAt := time.Now().Add(-time.Minute)
	repo.artifacts[1] = WeChatExportArtifact{
		ID:        1,
		UserID:    42,
		TaskID:    9,
		FileName:  "expired.json",
		ExpiresAt: &expiredAt,
	}
	svc := NewWeChatExportService(repo, nil)

	_, err := svc.GetArtifact(context.Background(), 42, 1)
	require.ErrorIs(t, err, ErrWeChatTaskNotFound)
}

func TestWeChatCookiePayloadEncryptionRoundTrip(t *testing.T) {
	t.Setenv("WECHAT_EXPORT_SESSION_SECRET", "test-secret")

	encrypted, err := encryptWeChatCookiePayload(wechatCookiePayload{CookieHeader: "foo=bar; baz=qux"})
	require.NoError(t, err)
	require.NotContains(t, encrypted, "foo=bar")

	decrypted, err := decryptWeChatCookiePayload(encrypted)
	require.NoError(t, err)
	require.Equal(t, "foo=bar; baz=qux", decrypted.CookieHeader)
}

func TestWeChatCookieHeaderAndRedirectParsing(t *testing.T) {
	merged := mergeCookieHeaders("foo=new; baz=qux", "foo=old; sid=1")
	require.Equal(t, "foo=new; sid=1; baz=qux", merged)
	require.Equal(t, "123", tokenFromWeChatRedirect("https://mp.weixin.qq.com/cgi-bin/home?t=home/index&token=123&lang=zh_CN"))
	require.Empty(t, tokenFromWeChatRedirect("://bad-url"))
}

// Phase 3: Service层Billing边界测试
func TestWeChatExportServiceCreateTask_InsufficientBalance(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)

	// 创建Session和Articles
	session := &WeChatSession{
		UserID: 42,
		Status: WeChatSessionStatusReady,
	}
	require.NoError(t, repo.CreateSession(context.Background(), session))

	article := &WeChatArticle{
		UserID: 42,
		Title:  "Test Article",
		Link:   "https://mp.weixin.qq.com/s/test",
	}
	require.NoError(t, repo.UpsertArticle(context.Background(), article))

	// Mock CreateTask返回余额不足错误
	repo.createTaskErr = ErrWeChatInsufficientBalance

	input := CreateWeChatExportTaskInput{
		ArticleIDs:        []int64{article.ID},
		Formats:           []string{"html"},
		IncludeEngagement: false,
	}

	task, err := svc.CreateTask(context.Background(), session.UserID, input)
	require.Error(t, err)
	require.Equal(t, ErrWeChatInsufficientBalance, err)
	require.Nil(t, task)
}
