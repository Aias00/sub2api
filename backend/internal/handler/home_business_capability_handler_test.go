package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Aias00/cloudbase/internal/hot"
	imagectx "github.com/Aias00/cloudbase/internal/image"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type homeBusinessPromptRepoStub struct {
	summary *imagectx.PromptCatalogSummary
}

func (s *homeBusinessPromptRepoStub) ListCases(context.Context, pagination.PaginationParams, imagectx.PromptCatalogListFilters) ([]imagectx.PromptCatalogCase, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (s *homeBusinessPromptRepoStub) GetCaseSummary(context.Context, imagectx.PromptCatalogListFilters) (*imagectx.PromptCatalogSummary, error) {
	if s.summary != nil {
		return s.summary, nil
	}
	return &imagectx.PromptCatalogSummary{}, nil
}

func (s *homeBusinessPromptRepoStub) GetCaseByID(context.Context, string) (*imagectx.PromptCatalogCase, error) {
	return nil, imagectx.ErrPromptCatalogNotFound
}

func (s *homeBusinessPromptRepoStub) UpsertCase(context.Context, *imagectx.PromptCatalogCase) error {
	return nil
}

type homeBusinessHotRepoStub struct {
	total int64
}

func (s *homeBusinessHotRepoStub) ListSources(context.Context) ([]hot.Source, error) {
	return []hot.Source{}, nil
}

func (s *homeBusinessHotRepoStub) ListItems(context.Context, pagination.PaginationParams, hot.ListFilters) ([]hot.Item, *pagination.PaginationResult, error) {
	return []hot.Item{}, &pagination.PaginationResult{Total: s.total, Page: 1, PageSize: 1, Pages: 1}, nil
}

func (s *homeBusinessHotRepoStub) ListRunEvents(context.Context, string, pagination.PaginationParams) ([]hot.RunEvent, *pagination.PaginationResult, error) {
	return []hot.RunEvent{}, &pagination.PaginationResult{}, nil
}

type homeBusinessImageWorkspaceRepoStub struct {
	status *service.ImageWorkspaceWorkerStatus
}

func (s *homeBusinessImageWorkspaceRepoStub) CreateTask(context.Context, *service.ImageWorkspaceTask) error {
	return nil
}

func (s *homeBusinessImageWorkspaceRepoStub) ListTasks(context.Context, int64, pagination.PaginationParams, service.ImageWorkspaceTaskFilters) ([]service.ImageWorkspaceTask, *pagination.PaginationResult, error) {
	return []service.ImageWorkspaceTask{}, &pagination.PaginationResult{}, nil
}

func (s *homeBusinessImageWorkspaceRepoStub) GetTask(context.Context, int64, int64) (*service.ImageWorkspaceTask, error) {
	return nil, service.ErrImageWorkspaceTaskNotFound
}

func (s *homeBusinessImageWorkspaceRepoStub) ListArtifacts(context.Context, int64, int64) ([]service.ImageWorkspaceArtifact, error) {
	return []service.ImageWorkspaceArtifact{}, nil
}

func (s *homeBusinessImageWorkspaceRepoStub) GetArtifact(context.Context, int64, int64) (*service.ImageWorkspaceArtifact, error) {
	return nil, service.ErrImageWorkspaceTaskNotFound
}

func (s *homeBusinessImageWorkspaceRepoStub) ListTemplates(context.Context, int64) ([]service.ImageWorkspaceTemplate, error) {
	return []service.ImageWorkspaceTemplate{}, nil
}

func (s *homeBusinessImageWorkspaceRepoStub) UpsertTemplate(context.Context, *service.ImageWorkspaceTemplate) error {
	return nil
}

func (s *homeBusinessImageWorkspaceRepoStub) DeleteTemplate(context.Context, int64, int64) error {
	return nil
}

func (s *homeBusinessImageWorkspaceRepoStub) ListUsageRecords(context.Context, int64, pagination.PaginationParams) ([]service.ImageWorkspaceUsageRecord, *pagination.PaginationResult, error) {
	return []service.ImageWorkspaceUsageRecord{}, &pagination.PaginationResult{}, nil
}

func (s *homeBusinessImageWorkspaceRepoStub) ClaimNextTask(context.Context, int64) (*service.ImageWorkspaceTask, error) {
	return nil, nil
}

func (s *homeBusinessImageWorkspaceRepoStub) CompleteTask(context.Context, int64, []service.ImageWorkspaceArtifact, string, float64) (*service.ImageWorkspaceTask, error) {
	return nil, service.ErrImageWorkspaceTaskNotFound
}

func (s *homeBusinessImageWorkspaceRepoStub) FailTask(context.Context, int64, string, string) (*service.ImageWorkspaceTask, error) {
	return nil, service.ErrImageWorkspaceTaskNotFound
}

func (s *homeBusinessImageWorkspaceRepoStub) CancelTask(context.Context, int64, int64) (*service.ImageWorkspaceTask, error) {
	return nil, service.ErrImageWorkspaceTaskNotFound
}

func (s *homeBusinessImageWorkspaceRepoStub) GetWorkerStatus(context.Context) (*service.ImageWorkspaceWorkerStatus, error) {
	if s.status != nil {
		return s.status, nil
	}
	return &service.ImageWorkspaceWorkerStatus{}, nil
}

type homeBusinessWeChatExportRepoStub struct {
	status *service.WeChatExportWorkerStatus
}

func (s *homeBusinessWeChatExportRepoStub) GetActiveSession(context.Context, int64) (*service.WeChatSession, error) {
	return nil, service.ErrWeChatSessionNotFound
}
func (s *homeBusinessWeChatExportRepoStub) CreateSession(context.Context, *service.WeChatSession) error {
	return nil
}
func (s *homeBusinessWeChatExportRepoStub) UpdateSession(context.Context, *service.WeChatSession) error {
	return nil
}
func (s *homeBusinessWeChatExportRepoStub) GetSession(context.Context, int64, int64) (*service.WeChatSession, error) {
	return nil, service.ErrWeChatSessionNotFound
}
func (s *homeBusinessWeChatExportRepoStub) ExpireUserSessions(context.Context, int64) error {
	return nil
}
func (s *homeBusinessWeChatExportRepoStub) ExpireLoginAttemptSessions(context.Context, int64) error {
	return nil
}
func (s *homeBusinessWeChatExportRepoStub) SearchAccounts(context.Context, int64, string, int) ([]service.WeChatAccount, error) {
	return []service.WeChatAccount{}, nil
}
func (s *homeBusinessWeChatExportRepoStub) GetAccount(context.Context, int64, string) (*service.WeChatAccount, error) {
	return nil, service.ErrWeChatAccountNotFound
}
func (s *homeBusinessWeChatExportRepoStub) UpsertAccount(context.Context, *service.WeChatAccount) error {
	return nil
}
func (s *homeBusinessWeChatExportRepoStub) MarkAccountSynced(context.Context, int64, string) (*service.WeChatAccount, error) {
	return nil, service.ErrWeChatAccountNotFound
}
func (s *homeBusinessWeChatExportRepoStub) UpsertArticle(context.Context, *service.WeChatArticle) error {
	return nil
}
func (s *homeBusinessWeChatExportRepoStub) UpdateArticleEnrichment(context.Context, *service.WeChatArticle) error {
	return nil
}
func (s *homeBusinessWeChatExportRepoStub) ListArticles(context.Context, int64, pagination.PaginationParams) ([]service.WeChatArticle, *pagination.PaginationResult, error) {
	return []service.WeChatArticle{}, &pagination.PaginationResult{}, nil
}
func (s *homeBusinessWeChatExportRepoStub) GetArticleByID(context.Context, int64) (*service.WeChatArticle, error) {
	return nil, service.ErrWeChatArticleNotFound
}
func (s *homeBusinessWeChatExportRepoStub) ListArticlesByIDs(context.Context, int64, []int64) ([]service.WeChatArticle, error) {
	return []service.WeChatArticle{}, nil
}
func (s *homeBusinessWeChatExportRepoStub) CreateTask(context.Context, *service.WeChatExportTask) error {
	return nil
}
func (s *homeBusinessWeChatExportRepoStub) ListTasks(context.Context, int64, pagination.PaginationParams) ([]service.WeChatExportTask, *pagination.PaginationResult, error) {
	return []service.WeChatExportTask{}, &pagination.PaginationResult{}, nil
}
func (s *homeBusinessWeChatExportRepoStub) GetWorkerStatus(context.Context, int64) (*service.WeChatExportWorkerStatus, error) {
	if s.status != nil {
		return s.status, nil
	}
	return &service.WeChatExportWorkerStatus{}, nil
}
func (s *homeBusinessWeChatExportRepoStub) GetTask(context.Context, int64, int64) (*service.WeChatExportTask, error) {
	return nil, service.ErrWeChatTaskNotFound
}
func (s *homeBusinessWeChatExportRepoStub) CancelTask(context.Context, int64, int64) (*service.WeChatExportTask, error) {
	return nil, service.ErrWeChatTaskNotFound
}
func (s *homeBusinessWeChatExportRepoStub) RetryTask(context.Context, int64, int64) (*service.WeChatExportTask, error) {
	return nil, service.ErrWeChatTaskNotFound
}
func (s *homeBusinessWeChatExportRepoStub) AddTaskLog(context.Context, int64, string, service.WeChatExportTaskLog) (*service.WeChatExportTaskLog, error) {
	return nil, service.ErrWeChatTaskNotFound
}
func (s *homeBusinessWeChatExportRepoStub) ListTaskLogs(context.Context, int64, int64) ([]service.WeChatExportTaskLog, error) {
	return []service.WeChatExportTaskLog{}, nil
}
func (s *homeBusinessWeChatExportRepoStub) ClaimNextTask(context.Context, int64) (*service.WeChatExportTask, []service.WeChatArticle, string, error) {
	return nil, nil, "", nil
}
func (s *homeBusinessWeChatExportRepoStub) CompleteTask(context.Context, int64, string, []service.WeChatExportArtifact, string, int, float64) (*service.WeChatExportTask, error) {
	return nil, service.ErrWeChatTaskNotFound
}
func (s *homeBusinessWeChatExportRepoStub) FailTask(context.Context, int64, string, string) (*service.WeChatExportTask, error) {
	return nil, service.ErrWeChatTaskNotFound
}
func (s *homeBusinessWeChatExportRepoStub) ListArtifacts(context.Context, int64, int64) ([]service.WeChatExportArtifact, error) {
	return []service.WeChatExportArtifact{}, nil
}
func (s *homeBusinessWeChatExportRepoStub) GetArtifact(context.Context, int64, int64) (*service.WeChatExportArtifact, error) {
	return nil, service.ErrWeChatTaskNotFound
}
func (s *homeBusinessWeChatExportRepoStub) GetWeChatExportCostPerArticle() float64 {
	return 0
}

func TestHomeBusinessCapabilityStatuses(t *testing.T) {
	t.Setenv("HOT_CONTENT_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY", "true")

	gin.SetMode(gin.TestMode)
	prompt := NewPromptCatalogHandler(imagectx.NewPromptCatalogService(&homeBusinessPromptRepoStub{
		summary: &imagectx.PromptCatalogSummary{Total: 3, CaseCount: 3},
	}))
	hotHandler := NewHotContentHandler(hot.NewService(&homeBusinessHotRepoStub{total: 2}))
	h := NewHomeBusinessCapabilityHandler(prompt, nil, nil, hotHandler)

	router := gin.New()
	router.GET("/api/v1/home/business-capabilities", h.GetStatuses)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/home/business-capabilities", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Code int `json:"code"`
		Data map[string]struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Count   int64  `json:"count"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, "available", body.Data["prompt-catalog"].Status)
	require.Equal(t, int64(3), body.Data["prompt-catalog"].Count)
	require.Equal(t, "available", body.Data["hot-topics"].Status)
	require.Equal(t, int64(2), body.Data["hot-topics"].Count)
	require.Equal(t, "in_progress", body.Data["wechat-export"].Status)
	require.NotEmpty(t, body.Data["wechat-export"].Message)
	require.Equal(t, "in_progress", body.Data["image-workspace"].Status)
	require.NotEmpty(t, body.Data["image-workspace"].Message)
}

func TestHomeBusinessCapabilityHotContentRequiresWorkerStatus(t *testing.T) {
	t.Setenv("HOT_CONTENT_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY", "false")
	t.Setenv("HOT_WORKER_STATUS_PATH", "")

	hotHandler := NewHotContentHandler(hot.NewService(&homeBusinessHotRepoStub{total: 2}))
	h := NewHomeBusinessCapabilityHandler(nil, nil, nil, hotHandler)

	status := h.hotContentStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "status path")
}

func TestHomeBusinessCapabilityHotContentReportsUnhealthyWorker(t *testing.T) {
	statusPath := writeHotWorkerStatus(t, `{"status":"error","updated_at":"`+time.Now().UTC().Format(time.RFC3339Nano)+`"}`)
	t.Setenv("HOT_CONTENT_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY", "false")
	t.Setenv("HOT_WORKER_STATUS_PATH", statusPath)

	hotHandler := NewHotContentHandler(hot.NewService(&homeBusinessHotRepoStub{total: 2}))
	h := NewHomeBusinessCapabilityHandler(nil, nil, nil, hotHandler)

	status := h.hotContentStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "unhealthy")
}

func TestHomeBusinessCapabilityHotContentReportsStaleWorker(t *testing.T) {
	statusPath := writeHotWorkerStatus(t, `{"status":"ok","apply":true,"run_count":1,"success_count":1,"failure_count":0,"updated_at":"`+time.Now().Add(-2*time.Hour).UTC().Format(time.RFC3339Nano)+`"}`)
	t.Setenv("HOT_CONTENT_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY", "false")
	t.Setenv("HOT_WORKER_STATUS_PATH", statusPath)
	t.Setenv("HOT_WORKER_HEALTH_MAX_AGE_MS", "1000")

	hotHandler := NewHotContentHandler(hot.NewService(&homeBusinessHotRepoStub{total: 2}))
	h := NewHomeBusinessCapabilityHandler(nil, nil, nil, hotHandler)

	status := h.hotContentStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "stale")
}

func TestHomeBusinessCapabilityHotContentRejectsDryRunWorker(t *testing.T) {
	statusPath := writeHotWorkerStatus(t, `{"status":"ok","apply":false,"run_count":1,"success_count":1,"failure_count":0,"updated_at":"`+time.Now().UTC().Format(time.RFC3339Nano)+`"}`)
	t.Setenv("HOT_CONTENT_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY", "false")
	t.Setenv("HOT_WORKER_STATUS_PATH", statusPath)

	hotHandler := NewHotContentHandler(hot.NewService(&homeBusinessHotRepoStub{total: 2}))
	h := NewHomeBusinessCapabilityHandler(nil, nil, nil, hotHandler)

	status := h.hotContentStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "dry-run")
}

func TestHomeBusinessCapabilityHotContentReportsWorkerFailures(t *testing.T) {
	statusPath := writeHotWorkerStatus(t, `{"status":"ok","apply":true,"run_count":2,"success_count":1,"failure_count":1,"updated_at":"`+time.Now().UTC().Format(time.RFC3339Nano)+`"}`)
	t.Setenv("HOT_CONTENT_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY", "false")
	t.Setenv("HOT_WORKER_STATUS_PATH", statusPath)

	hotHandler := NewHotContentHandler(hot.NewService(&homeBusinessHotRepoStub{total: 2}))
	h := NewHomeBusinessCapabilityHandler(nil, nil, nil, hotHandler)

	status := h.hotContentStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "failures")
}

func TestHomeBusinessCapabilityHotContentAvailableWhenWorkerReady(t *testing.T) {
	statusPath := writeHotWorkerStatus(t, `{"status":"ok","apply":true,"run_count":1,"success_count":1,"failure_count":0,"updated_at":"`+time.Now().UTC().Format(time.RFC3339Nano)+`"}`)
	t.Setenv("HOT_CONTENT_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY", "false")
	t.Setenv("HOT_WORKER_STATUS_PATH", statusPath)

	hotHandler := NewHotContentHandler(hot.NewService(&homeBusinessHotRepoStub{total: 2}))
	h := NewHomeBusinessCapabilityHandler(nil, nil, nil, hotHandler)

	status := h.hotContentStatus(context.Background())

	require.Equal(t, "available", status.Status)
	require.Equal(t, int64(2), status.Count)
	require.Empty(t, status.Message)
}

func writeHotWorkerStatus(t *testing.T, payload string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hot-worker-status.json")
	require.NoError(t, os.WriteFile(path, []byte(payload), 0o600))
	return path
}

func TestHomeBusinessCapabilityImageWorkspaceRequiresRuntimeReadiness(t *testing.T) {
	t.Setenv("IMAGE_WORKSPACE_WORKER_TOKEN", "worker-token")
	t.Setenv("IMAGE_WORKSPACE_UPSTREAM_API_KEY", "")
	t.Setenv("IMAGE_WORKSPACE_OUTPUT_DIR", "/tmp/image-workspace")
	t.Setenv("IMAGE_WORKSPACE_STORAGE_ROOT", "/tmp/image-workspace")

	imageWorkspace := NewImageWorkspaceHandler(service.NewImageWorkspaceService(&homeBusinessImageWorkspaceRepoStub{}, nil))
	h := NewHomeBusinessCapabilityHandler(nil, nil, imageWorkspace, nil)

	status := h.imageWorkspaceStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "upstream image provider")
}

func TestHomeBusinessCapabilityImageWorkspaceAvailableWhenRuntimeReady(t *testing.T) {
	t.Setenv("IMAGE_WORKSPACE_WORKER_TOKEN", "worker-token")
	t.Setenv("IMAGE_WORKSPACE_UPSTREAM_API_KEY", "provider-token")
	t.Setenv("IMAGE_WORKSPACE_OUTPUT_DIR", "/tmp/image-workspace")
	t.Setenv("IMAGE_WORKSPACE_STORAGE_ROOT", "/tmp/image-workspace")

	imageWorkspace := NewImageWorkspaceHandler(service.NewImageWorkspaceService(&homeBusinessImageWorkspaceRepoStub{}, nil))
	h := NewHomeBusinessCapabilityHandler(nil, nil, imageWorkspace, nil)

	status := h.imageWorkspaceStatus(context.Background())

	require.Equal(t, "available", status.Status)
	require.Positive(t, status.Count)
	require.Empty(t, status.Message)
}

func TestHomeBusinessCapabilityImageWorkspaceReportsRecentRuntimeFailure(t *testing.T) {
	t.Setenv("IMAGE_WORKSPACE_WORKER_TOKEN", "worker-token")
	t.Setenv("IMAGE_WORKSPACE_UPSTREAM_API_KEY", "provider-token")
	t.Setenv("IMAGE_WORKSPACE_OUTPUT_DIR", "/tmp/image-workspace")
	t.Setenv("IMAGE_WORKSPACE_STORAGE_ROOT", "/tmp/image-workspace")

	imageWorkspace := NewImageWorkspaceHandler(service.NewImageWorkspaceService(&homeBusinessImageWorkspaceRepoStub{
		status: &service.ImageWorkspaceWorkerStatus{
			RecentFailedCount:  1,
			LastFailureMessage: "upstream 404: /v1/images/generations is not available",
		},
	}, nil))
	h := NewHomeBusinessCapabilityHandler(nil, nil, imageWorkspace, nil)

	status := h.imageWorkspaceStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "worker runtime")
}

func TestHomeBusinessCapabilityWeChatExportRequiresWorkerToken(t *testing.T) {
	t.Setenv("WECHAT_EXPORT_WORKER_TOKEN", "")
	t.Setenv("WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN", "false")

	weChat := NewWeChatExportHandler(service.NewWeChatExportService(&homeBusinessWeChatExportRepoStub{}, nil))
	h := NewHomeBusinessCapabilityHandler(nil, weChat, nil, nil)

	status := h.weChatExportStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "worker authentication")
}

func TestHomeBusinessCapabilityWeChatExportRejectsPrivateWorkerBypass(t *testing.T) {
	t.Setenv("WECHAT_EXPORT_WORKER_TOKEN", "")
	t.Setenv("WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN", "true")

	weChat := NewWeChatExportHandler(service.NewWeChatExportService(&homeBusinessWeChatExportRepoStub{}, nil))
	h := NewHomeBusinessCapabilityHandler(nil, weChat, nil, nil)

	status := h.weChatExportStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "without a token")
}

func TestHomeBusinessCapabilityWeChatExportUsesQueueStatusWhenConfigured(t *testing.T) {
	t.Setenv("WECHAT_EXPORT_WORKER_TOKEN", "worker-token")
	t.Setenv("WECHAT_EXPORT_CAPABILITY_STATUS_USER_ID", "42")
	oldestQueued := time.Now().Add(-10 * time.Minute)

	weChat := NewWeChatExportHandler(service.NewWeChatExportService(&homeBusinessWeChatExportRepoStub{
		status: &service.WeChatExportWorkerStatus{
			QueuedCount:    1,
			OldestQueuedAt: &oldestQueued,
		},
	}, nil))
	h := NewHomeBusinessCapabilityHandler(nil, weChat, nil, nil)

	status := h.weChatExportStatus(context.Background())

	require.Equal(t, "in_progress", status.Status)
	require.Contains(t, status.Message, "queued tasks")
}

func TestHomeBusinessCapabilityWeChatExportAvailableWhenRuntimeReady(t *testing.T) {
	t.Setenv("WECHAT_EXPORT_WORKER_TOKEN", "worker-token")

	weChat := NewWeChatExportHandler(service.NewWeChatExportService(&homeBusinessWeChatExportRepoStub{}, nil))
	h := NewHomeBusinessCapabilityHandler(nil, weChat, nil, nil)

	status := h.weChatExportStatus(context.Background())

	require.Equal(t, "available", status.Status)
	require.Empty(t, status.Message)
}
