package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type wechatExportRepoFake struct {
	sessions  map[int64]WeChatSession
	articles  map[int64]WeChatArticle
	tasks     map[int64]WeChatExportTask
	artifacts map[int64]WeChatExportArtifact
	nextID    int64
}

func newWeChatExportRepoFake() *wechatExportRepoFake {
	return &wechatExportRepoFake{
		sessions:  map[int64]WeChatSession{},
		articles:  map[int64]WeChatArticle{},
		tasks:     map[int64]WeChatExportTask{},
		artifacts: map[int64]WeChatExportArtifact{},
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

func (r *wechatExportRepoFake) UpsertArticle(ctx context.Context, article *WeChatArticle) error {
	if article.ID == 0 {
		article.ID = r.nextID
		r.nextID++
	}
	r.articles[article.ID] = *article
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
	task.ID = r.nextID
	r.nextID++
	r.tasks[task.ID] = *task
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

func (r *wechatExportRepoFake) GetTask(ctx context.Context, userID int64, taskID int64) (*WeChatExportTask, error) {
	task, ok := r.tasks[taskID]
	if !ok || task.UserID != userID {
		return nil, ErrWeChatTaskNotFound
	}
	return &task, nil
}

func (r *wechatExportRepoFake) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*WeChatExportTask, []WeChatArticle, error) {
	for id, task := range r.tasks {
		if task.Status == WeChatExportTaskStatusQueued {
			task.Status = WeChatExportTaskStatusRunning
			r.tasks[id] = task
			articles, err := r.ListArticlesByIDs(ctx, task.UserID, task.ArticleIDs)
			return &task, articles, err
		}
	}
	return nil, nil, nil
}

func (r *wechatExportRepoFake) CompleteTask(ctx context.Context, taskID int64, artifacts []WeChatExportArtifact, resultManifestJSON string) (*WeChatExportTask, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, ErrWeChatTaskNotFound
	}
	task.Status = WeChatExportTaskStatusCompleted
	task.SuccessfulArticleCount = task.SelectedArticleCount
	task.ResultManifestJSON = resultManifestJSON
	r.tasks[taskID] = task
	for _, artifact := range artifacts {
		artifact.ID = r.nextID
		r.nextID++
		artifact.TaskID = taskID
		artifact.UserID = task.UserID
		r.artifacts[artifact.ID] = artifact
	}
	return &task, nil
}

func (r *wechatExportRepoFake) FailTask(ctx context.Context, taskID int64, message string) (*WeChatExportTask, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, ErrWeChatTaskNotFound
	}
	task.Status = WeChatExportTaskStatusFailed
	task.ErrorMessage = message
	r.tasks[taskID] = task
	return &task, nil
}

func (r *wechatExportRepoFake) ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]WeChatExportArtifact, error) {
	items := make([]WeChatExportArtifact, 0, len(r.artifacts))
	for _, artifact := range r.artifacts {
		if artifact.UserID == userID && artifact.TaskID == taskID {
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
	svc := NewWeChatExportService(repo)

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

func TestWeChatExportServiceRejectsUnknownArticles(t *testing.T) {
	svc := NewWeChatExportService(newWeChatExportRepoFake())

	_, err := svc.CreateTask(context.Background(), 42, CreateWeChatExportTaskInput{
		ArticleIDs: []int64{999},
		Formats:    []string{"html"},
	})

	require.ErrorIs(t, err, ErrWeChatArticleNotFound)
}

func TestWeChatExportServiceWorkerClaimAndComplete(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo)
	article, err := svc.ImportArticleLink(context.Background(), 7, "https://mp.weixin.qq.com/s/demo")
	require.NoError(t, err)
	task, err := svc.CreateTask(context.Background(), 7, CreateWeChatExportTaskInput{
		ArticleIDs: []int64{article.ID},
		Formats:    []string{"json"},
	})
	require.NoError(t, err)

	claimed, articles, err := svc.ClaimNextTask(context.Background(), 60)
	require.NoError(t, err)
	require.Equal(t, task.ID, claimed.ID)
	require.Equal(t, WeChatExportTaskStatusRunning, claimed.Status)
	require.Len(t, articles, 1)

	completed, err := svc.CompleteTask(context.Background(), task.ID, CompleteWeChatExportTaskInput{
		Artifacts: []WeChatExportArtifactInput{{
			Format:      "json",
			StorageKey:  "/tmp/wechat-export.json",
			DownloadURL: "/api/v1/wechat/artifacts/1/download",
			FileName:    "wechat-export.json",
			FileSize:    128,
		}},
		ResultManifestJSON: `{"ok":true}`,
	})
	require.NoError(t, err)
	require.Equal(t, WeChatExportTaskStatusCompleted, completed.Status)

	artifacts, err := svc.ListArtifacts(context.Background(), 7, task.ID)
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	require.Equal(t, "wechat-export.json", artifacts[0].FileName)
}
