package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type HomeBusinessCapabilityHandler struct {
	promptCatalog  *PromptCatalogHandler
	weChatExport   *WeChatExportHandler
	imageWorkspace *ImageWorkspaceHandler
	hotContent     *HotContentHandler
}

const (
	workerNodeBusiness = "business-worker"
	workerNodeContent  = "content-worker"
)

func NewHomeBusinessCapabilityHandler(
	promptCatalog *PromptCatalogHandler,
	weChatExport *WeChatExportHandler,
	imageWorkspace *ImageWorkspaceHandler,
	hotContent *HotContentHandler,
) *HomeBusinessCapabilityHandler {
	return &HomeBusinessCapabilityHandler{
		promptCatalog:  promptCatalog,
		weChatExport:   weChatExport,
		imageWorkspace: imageWorkspace,
		hotContent:     hotContent,
	}
}

type homeBusinessCapabilityStatusDTO struct {
	Status      string `json:"status"`
	StatusLabel string `json:"status_label,omitempty"`
	Message     string `json:"message,omitempty"`
	Count       int64  `json:"count,omitempty"`
}

type adminWorkerRuntimeStatusDTO struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	NodeID           string   `json:"node_id,omitempty"`
	ContainerName    string   `json:"container_name,omitempty"`
	ContainerState   string   `json:"container_state,omitempty"`
	Health           string   `json:"health"`
	Message          string   `json:"message,omitempty"`
	Queue            int64    `json:"queue,omitempty"`
	Running          int64    `json:"running,omitempty"`
	Stale            int64    `json:"stale,omitempty"`
	Failed           int64    `json:"failed,omitempty"`
	Succeeded        int64    `json:"succeeded,omitempty"`
	Total            int64    `json:"total,omitempty"`
	LastUpdatedAt    string   `json:"last_updated_at,omitempty"`
	LastAgeSeconds   *int64   `json:"last_age_seconds,omitempty"`
	OldestQueuedAt   string   `json:"oldest_queued_at,omitempty"`
	AttentionReasons []string `json:"attention_reasons,omitempty"`
	Configured       bool     `json:"configured"`
	StatusPath       string   `json:"status_path,omitempty"`
	Manageable       bool     `json:"manageable"`
	ManagementReason string   `json:"management_reason,omitempty"`
	DeployCommand    string   `json:"deploy_command,omitempty"`
	Actions          []string `json:"actions,omitempty"`
}

func (h *HomeBusinessCapabilityHandler) GetStatuses(c *gin.Context) {
	response.Success(c, gin.H{
		"wechat-export":   h.weChatExportStatus(c.Request.Context()),
		"prompt-catalog":  h.promptCatalogStatus(c.Request.Context()),
		"image-workspace": h.imageWorkspaceStatus(c.Request.Context()),
		"hot-topics":      h.hotContentStatus(c.Request.Context()),
	})
}

func (h *HomeBusinessCapabilityHandler) GetAdminWorkerRuntimeStatuses(c *gin.Context) {
	workers := []adminWorkerRuntimeStatusDTO{
		h.adminImageWorkspaceWorkerStatus(c.Request.Context()),
		h.adminWeChatExportWorkerStatus(c.Request.Context()),
		h.adminHotCollectorWorkerStatus(),
		h.adminXAutoWorkerStatus(c.Request.Context()),
	}
	manager := newRuntimeWorkerDockerManager()
	for i := range workers {
		workers[i] = manager.enrich(workers[i])
	}
	response.Success(c, gin.H{
		"workers": workers,
		"management": gin.H{
			"enabled": manager.enabled,
			"reason":  manager.disabledReason(),
			"socket":  manager.socketPath,
		},
	})
}

func (h *HomeBusinessCapabilityHandler) ManageAdminRuntimeWorker(c *gin.Context) {
	workerID := strings.TrimSpace(c.Param("id"))
	action := strings.ToLower(strings.TrimSpace(c.Param("action")))
	manager := newRuntimeWorkerDockerManager()
	target, ok := runtimeWorkerTarget(workerID)
	if !ok {
		response.NotFound(c, "worker not found")
		return
	}
	if action == "deploy" {
		response.Success(c, gin.H{
			"ok":              false,
			"action":          action,
			"worker_id":       workerID,
			"container_name":  target.ContainerName,
			"message":         "Deployment is intentionally exposed as an operator command, not executed by the API server.",
			"deploy_command":  target.DeployCommand,
			"management_note": manager.disabledReason(),
		})
		return
	}
	if !manager.enabled {
		response.Forbidden(c, manager.disabledReason())
		return
	}
	var err error
	switch action {
	case "restart":
		err = manager.postDocker(c.Request.Context(), fmt.Sprintf("/containers/%s/restart?t=10", target.ContainerName))
	case "start", "online":
		action = "start"
		err = manager.postDocker(c.Request.Context(), fmt.Sprintf("/containers/%s/start", target.ContainerName))
	case "stop", "offline":
		action = "stop"
		err = manager.postDocker(c.Request.Context(), fmt.Sprintf("/containers/%s/stop?t=10", target.ContainerName))
	default:
		response.BadRequest(c, "unsupported worker action")
		return
	}
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"ok":             true,
		"action":         action,
		"worker_id":      workerID,
		"container_name": target.ContainerName,
	})
}

func (h *HomeBusinessCapabilityHandler) promptCatalogStatus(ctx context.Context) homeBusinessCapabilityStatusDTO {
	if h == nil || h.promptCatalog == nil || h.promptCatalog.service == nil {
		return homeBusinessInProgress("Prompt Catalog service is not configured.")
	}
	hasImage := true
	summary, err := h.promptCatalog.service.GetCaseSummary(ctx, service.PromptCatalogListFilters{
		SourceType: "case",
		HasImage:   &hasImage,
	})
	if err != nil {
		return homeBusinessInProgress("Prompt Catalog data is not reachable.")
	}
	count := summaryCount(summary)
	if count <= 0 {
		return homeBusinessInProgress("Prompt Catalog has no published image prompt cases.")
	}
	return homeBusinessAvailable(count)
}

func (h *HomeBusinessCapabilityHandler) hotContentStatus(ctx context.Context) homeBusinessCapabilityStatusDTO {
	if h == nil || h.hotContent == nil || h.hotContent.service == nil {
		return homeBusinessInProgress("Hot content service is not configured.")
	}
	_, result, err := h.hotContent.service.ListItems(ctx, pagination.PaginationParams{
		Page:     1,
		PageSize: 1,
	}, service.HotContentListFilters{Status: "published"})
	if err != nil {
		return homeBusinessInProgress("Hot content data is not reachable.")
	}
	count := int64(0)
	if result != nil {
		count = result.Total
	}
	if count <= 0 {
		return homeBusinessInProgress("Hot content API has no published items.")
	}
	if reasons := hotContentRuntimeReadinessGaps(); len(reasons) > 0 {
		return homeBusinessInProgress(strings.Join(reasons, " "))
	}
	return homeBusinessAvailable(count)
}

func (h *HomeBusinessCapabilityHandler) imageWorkspaceStatus(ctx context.Context) homeBusinessCapabilityStatusDTO {
	if h == nil || h.imageWorkspace == nil || h.imageWorkspace.service == nil {
		return homeBusinessInProgress("Image Workspace service is not configured.")
	}
	if err := h.imageWorkspace.service.Health(); err != nil {
		return homeBusinessInProgress("Image Workspace repository is not configured.")
	}
	models, err := h.imageWorkspace.service.ListModels(ctx)
	if err != nil {
		return homeBusinessInProgress("Image Workspace model config is not reachable.")
	}
	enabledCount := int64(0)
	for _, model := range models {
		if model.Enabled && strings.TrimSpace(model.ID) != "" {
			enabledCount++
		}
	}
	if enabledCount <= 0 {
		return homeBusinessInProgress("Image Workspace has no enabled image models.")
	}
	status, err := h.imageWorkspace.service.GetWorkerStatus(ctx)
	if err != nil {
		return homeBusinessInProgress("Image Workspace worker status is not reachable.")
	}
	if status != nil {
		if status.StaleRunningCount > 0 {
			return homeBusinessInProgress("Image Workspace has stale running tasks waiting for worker recovery.")
		}
		if strings.EqualFold(status.Health, "attention") && imageWorkspaceStatusHasAttention(status, "recent_runtime_failure") {
			return homeBusinessInProgress("Image Workspace worker runtime needs attention after a recent upstream/provider failure.")
		}
		if status.QueuedCount > 0 && status.RunningCount == 0 && status.OldestQueuedSeconds != nil && *status.OldestQueuedSeconds >= 300 {
			return homeBusinessInProgress("Image Workspace queued tasks have been waiting for a worker for over 5 minutes.")
		}
	}
	if reasons := imageWorkspaceRuntimeReadinessGaps(ctx, h.imageWorkspace.service); len(reasons) > 0 {
		return homeBusinessInProgress(strings.Join(reasons, " "))
	}
	return homeBusinessAvailable(enabledCount)
}

func imageWorkspaceStatusHasAttention(status *service.ImageWorkspaceWorkerStatus, reason string) bool {
	if status == nil {
		return false
	}
	for _, item := range status.AttentionReasons {
		if strings.EqualFold(strings.TrimSpace(item), reason) {
			return true
		}
	}
	return false
}

func (h *HomeBusinessCapabilityHandler) weChatExportStatus(ctx context.Context) homeBusinessCapabilityStatusDTO {
	if h == nil || h.weChatExport == nil || h.weChatExport.service == nil {
		return homeBusinessInProgress("WeChat Export service is not configured.")
	}
	if err := h.weChatExport.service.Health(ctx); err != nil {
		return homeBusinessInProgress("WeChat Export repository is not configured.")
	}
	if reasons := weChatExportRuntimeReadinessGaps(ctx, h.weChatExport.service); len(reasons) > 0 {
		return homeBusinessInProgress(strings.Join(reasons, " "))
	}
	return homeBusinessAvailable(0)
}

func summaryCount(summary *service.PromptCatalogSummary) int64 {
	if summary == nil {
		return 0
	}
	if summary.CaseCount > 0 {
		return summary.CaseCount
	}
	return summary.Total
}

func homeBusinessAvailable(count int64) homeBusinessCapabilityStatusDTO {
	return homeBusinessCapabilityStatusDTO{
		Status: "available",
		Count:  count,
	}
}

func homeBusinessInProgress(message string) homeBusinessCapabilityStatusDTO {
	return homeBusinessCapabilityStatusDTO{
		Status:  "in_progress",
		Message: message,
	}
}

func (h *HomeBusinessCapabilityHandler) adminImageWorkspaceWorkerStatus(ctx context.Context) adminWorkerRuntimeStatusDTO {
	result := adminWorkerRuntimeStatusDTO{
		ID:         "image-workspace",
		Name:       "生图工作台 Worker",
		NodeID:     workerNodeBusiness,
		Health:     "unknown",
		Configured: h != nil && h.imageWorkspace != nil && h.imageWorkspace.service != nil,
	}
	if !result.Configured {
		result.Health = "not_configured"
		result.Message = "Image Workspace service is not configured."
		return result
	}
	status, err := h.imageWorkspace.service.GetWorkerStatus(ctx)
	if err != nil {
		result.Health = "attention"
		result.Message = "Image Workspace worker status is not reachable."
		result.AttentionReasons = []string{err.Error()}
		return result
	}
	if status == nil {
		result.Health = "idle"
		return result
	}
	result.Health = status.Health
	result.Message = status.Message
	result.Queue = status.QueuedCount
	result.Running = status.RunningCount
	result.Stale = status.StaleRunningCount
	result.Failed = status.FailedCount
	result.Succeeded = status.SucceededCount
	result.Total = status.TotalCount
	result.LastAgeSeconds = status.LastTaskAgeSeconds
	result.AttentionReasons = append([]string(nil), status.AttentionReasons...)
	if status.LastTaskUpdatedAt != nil {
		result.LastUpdatedAt = status.LastTaskUpdatedAt.UTC().Format(time.RFC3339)
	}
	if status.OldestQueuedAt != nil {
		result.OldestQueuedAt = status.OldestQueuedAt.UTC().Format(time.RFC3339)
	}
	return result
}

func (h *HomeBusinessCapabilityHandler) adminWeChatExportWorkerStatus(ctx context.Context) adminWorkerRuntimeStatusDTO {
	result := adminWorkerRuntimeStatusDTO{
		ID:         "wechat-export",
		Name:       "微信导出 Worker",
		NodeID:     workerNodeBusiness,
		Health:     "unknown",
		Configured: h != nil && h.weChatExport != nil && h.weChatExport.service != nil,
	}
	if !result.Configured {
		result.Health = "not_configured"
		result.Message = "WeChat Export service is not configured."
		return result
	}
	if !homeBusinessEnvConfigured("WECHAT_EXPORT_WORKER_TOKEN") && !homeBusinessEnvEnabled("WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN") {
		result.Health = "not_configured"
		result.Message = "WeChat Export worker authentication is not configured."
		result.AttentionReasons = []string{"worker_auth_not_configured"}
		return result
	}
	userIDRaw := strings.TrimSpace(os.Getenv("WECHAT_EXPORT_CAPABILITY_STATUS_USER_ID"))
	if userIDRaw == "" {
		result.Health = "unknown"
		result.Message = "Set WECHAT_EXPORT_CAPABILITY_STATUS_USER_ID to show user-scoped WeChat export queue metrics."
		result.AttentionReasons = []string{"status_user_id_not_configured"}
		return result
	}
	userID, err := strconv.ParseInt(userIDRaw, 10, 64)
	if err != nil || userID <= 0 {
		result.Health = "attention"
		result.Message = "WECHAT_EXPORT_CAPABILITY_STATUS_USER_ID is invalid."
		result.AttentionReasons = []string{"status_user_id_invalid"}
		return result
	}
	status, err := h.weChatExport.service.GetWorkerStatus(ctx, userID)
	if err != nil {
		result.Health = "attention"
		result.Message = "WeChat Export worker status is not reachable."
		result.AttentionReasons = []string{err.Error()}
		return result
	}
	if status == nil {
		result.Health = "idle"
		return result
	}
	result.Health = status.Health
	result.Message = status.Message
	result.Queue = status.QueuedCount
	result.Running = status.RunningCount
	result.Stale = status.StaleRunningCount
	result.Failed = status.FailedCount
	result.Succeeded = status.CompletedCount
	result.Total = status.TotalCount
	result.LastAgeSeconds = status.LastTaskAgeSeconds
	result.AttentionReasons = append([]string(nil), status.AttentionReasons...)
	if status.LastTaskUpdatedAt != nil {
		result.LastUpdatedAt = status.LastTaskUpdatedAt.UTC().Format(time.RFC3339)
	}
	if status.OldestQueuedAt != nil {
		result.OldestQueuedAt = status.OldestQueuedAt.UTC().Format(time.RFC3339)
	}
	return result
}

func (h *HomeBusinessCapabilityHandler) adminHotCollectorWorkerStatus() adminWorkerRuntimeStatusDTO {
	statusPath := strings.TrimSpace(os.Getenv("HOT_WORKER_STATUS_PATH"))
	result := adminWorkerRuntimeStatusDTO{
		ID:         "hot-collector",
		Name:       "热点采集 Worker",
		NodeID:     workerNodeContent,
		Health:     "unknown",
		Configured: statusPath != "",
		StatusPath: statusPath,
	}
	if statusPath == "" {
		result.Health = "not_configured"
		result.Message = "HOT_WORKER_STATUS_PATH is not configured."
		return result
	}
	data, err := os.ReadFile(statusPath)
	if err != nil {
		result.Health = "attention"
		result.Message = "Hot collector worker status is not reachable."
		result.AttentionReasons = []string{err.Error()}
		return result
	}
	var status hotWorkerStatusFile
	if err := json.Unmarshal(data, &status); err != nil {
		result.Health = "attention"
		result.Message = "Hot collector worker status is invalid."
		result.AttentionReasons = []string{err.Error()}
		return result
	}
	switch strings.ToLower(strings.TrimSpace(status.Status)) {
	case "ok", "running":
		result.Health = "active"
	default:
		result.Health = "attention"
	}
	result.Total = status.RunCount
	result.Succeeded = status.SuccessCount
	result.Failed = status.FailureCount
	result.LastUpdatedAt = strings.TrimSpace(status.UpdatedAt)
	if !status.Apply {
		result.AttentionReasons = append(result.AttentionReasons, "dry_run_mode")
	}
	if status.FailureCount > 0 {
		result.AttentionReasons = append(result.AttentionReasons, "worker_failures")
	}
	if updatedAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(status.UpdatedAt)); err == nil {
		ageSeconds := int64(time.Since(updatedAt).Seconds())
		if ageSeconds < 0 {
			ageSeconds = 0
		}
		result.LastAgeSeconds = &ageSeconds
		if time.Since(updatedAt) > hotWorkerStatusMaxAge() {
			result.Health = "attention"
			result.AttentionReasons = append(result.AttentionReasons, "status_stale")
		}
	} else {
		result.Health = "attention"
		result.AttentionReasons = append(result.AttentionReasons, "status_timestamp_invalid")
	}
	if len(result.AttentionReasons) == 0 && result.Message == "" {
		result.Message = "Hot collector worker status file is current."
	}
	return result
}

func (h *HomeBusinessCapabilityHandler) adminXAutoWorkerStatus(ctx context.Context) adminWorkerRuntimeStatusDTO {
	baseURL := xAutoBaseURL()
	result := adminWorkerRuntimeStatusDTO{
		ID:         "x-auto",
		Name:       "X Auto Worker",
		NodeID:     workerNodeContent,
		Health:     "unknown",
		Configured: baseURL != "",
		StatusPath: baseURL,
	}
	if baseURL == "" {
		result.Health = "not_configured"
		result.Message = "X_AUTO_BASE_URL is not configured."
		return result
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/healthz", nil)
	if err != nil {
		result.Health = "attention"
		result.Message = "X Auto worker health URL is invalid."
		result.AttentionReasons = []string{err.Error()}
		return result
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Health = "attention"
		result.Message = "X Auto worker is not reachable."
		result.AttentionReasons = []string{err.Error()}
		return result
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Health = "attention"
		result.Message = fmt.Sprintf("X Auto worker healthcheck returned HTTP %d.", resp.StatusCode)
		return result
	}
	var payload struct {
		Status    string `json:"status"`
		CheckedAt string `json:"checked_at"`
		DBPath    string `json:"db_path"`
	}
	content, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if len(content) > 0 {
		_ = json.Unmarshal(content, &payload)
	}
	result.Health = "active"
	result.Message = "X Auto worker is reachable."
	if strings.TrimSpace(payload.CheckedAt) != "" {
		result.LastUpdatedAt = strings.TrimSpace(payload.CheckedAt)
	}
	if strings.TrimSpace(payload.DBPath) != "" {
		result.AttentionReasons = []string{"db_path=" + strings.TrimSpace(payload.DBPath)}
	}
	return result
}

type hotWorkerStatusFile struct {
	Status       string `json:"status"`
	Apply        bool   `json:"apply"`
	RunCount     int64  `json:"run_count"`
	SuccessCount int64  `json:"success_count"`
	FailureCount int64  `json:"failure_count"`
	UpdatedAt    string `json:"updated_at"`
}

func hotContentRuntimeReadinessGaps() []string {
	if homeBusinessEnvEnabled("HOT_CONTENT_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY") ||
		homeBusinessEnvEnabled("HOT_COLLECTOR_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY") {
		return nil
	}
	reasons := make([]string, 0, 4)
	statusPath := strings.TrimSpace(os.Getenv("HOT_WORKER_STATUS_PATH"))
	if statusPath == "" {
		reasons = append(reasons, "Hot collector worker status path is not configured.")
		return reasons
	}
	data, err := os.ReadFile(statusPath)
	if err != nil {
		reasons = append(reasons, "Hot collector worker status is not reachable.")
		return reasons
	}
	var status hotWorkerStatusFile
	if err := json.Unmarshal(data, &status); err != nil {
		reasons = append(reasons, "Hot collector worker status is invalid.")
		return reasons
	}
	switch strings.ToLower(strings.TrimSpace(status.Status)) {
	case "ok", "running":
	default:
		reasons = append(reasons, "Hot collector worker is unhealthy.")
	}
	if !status.Apply {
		reasons = append(reasons, "Hot collector worker is running in dry-run mode.")
	}
	if status.RunCount <= 0 {
		reasons = append(reasons, "Hot collector worker has not completed any run.")
	}
	if status.SuccessCount <= 0 {
		reasons = append(reasons, "Hot collector worker has no successful runs.")
	}
	if status.FailureCount > 0 {
		reasons = append(reasons, "Hot collector worker has recorded failures.")
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(status.UpdatedAt))
	if err != nil {
		reasons = append(reasons, "Hot collector worker status timestamp is invalid.")
		return reasons
	}
	if time.Since(updatedAt) > hotWorkerStatusMaxAge() {
		reasons = append(reasons, "Hot collector worker status is stale.")
	}
	return reasons
}

func imageWorkspaceRuntimeReadinessGaps(ctx context.Context, svc *service.ImageWorkspaceService) []string {
	reasons := make([]string, 0, 4)
	if !homeBusinessEnvConfigured("IMAGE_WORKSPACE_WORKER_TOKEN") &&
		!homeBusinessEnvEnabled("IMAGE_WORKSPACE_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN") {
		reasons = append(reasons, "Image Workspace worker authentication is not configured.")
	}
	runtimeConfig, _ := svc.GetWorkerRuntimeConfig(ctx)
	if !homeBusinessEnvConfigured("IMAGE_WORKSPACE_UPSTREAM_API_KEY") && !runtimeConfig.AssumeWorkerReady {
		reasons = append(reasons, "Image Workspace upstream image provider is not configured.")
	}
	if runtimeConfig.ObjectStorage.Enabled {
		if !homeBusinessAnyEnvConfigured("IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT", "IMAGE_WORKSPACE_R2_ENDPOINT", "IMAGE_WORKSPACE_R2_ACCOUNT_ID") {
			reasons = append(reasons, "Image Workspace object storage endpoint/account is not configured.")
		}
		if strings.TrimSpace(runtimeConfig.ObjectStorage.Bucket) == "" {
			reasons = append(reasons, "Image Workspace object storage bucket is not configured.")
		}
		if strings.TrimSpace(runtimeConfig.ObjectStorage.PublicBaseURL) == "" {
			reasons = append(reasons, "Image Workspace object storage public URL is not configured.")
		}
		return reasons
	}
	if !homeBusinessAnyEnvConfigured("IMAGE_WORKSPACE_STORAGE_ROOT", "IMAGE_WORKSPACE_OUTPUT_DIR") {
		reasons = append(reasons, "Image Workspace local artifact storage root is not configured.")
	}
	return reasons
}

func weChatExportRuntimeReadinessGaps(ctx context.Context, svc *service.WeChatExportService) []string {
	reasons := make([]string, 0, 4)
	if !homeBusinessEnvConfigured("WECHAT_EXPORT_WORKER_TOKEN") {
		if homeBusinessEnvEnabled("WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN") {
			reasons = append(reasons, "WeChat Export is using private worker access without a token.")
		} else {
			reasons = append(reasons, "WeChat Export worker authentication is not configured.")
		}
	}
	userIDRaw := strings.TrimSpace(os.Getenv("WECHAT_EXPORT_CAPABILITY_STATUS_USER_ID"))
	if userIDRaw == "" {
		return reasons
	}
	userID, err := strconv.ParseInt(userIDRaw, 10, 64)
	if err != nil || userID <= 0 {
		reasons = append(reasons, "WeChat Export capability status user id is invalid.")
		return reasons
	}
	status, err := svc.GetWorkerStatus(ctx, userID)
	if err != nil {
		reasons = append(reasons, "WeChat Export worker queue status is not reachable.")
		return reasons
	}
	if status == nil {
		return reasons
	}
	if status.StaleRunningCount > 0 {
		reasons = append(reasons, "WeChat Export has stale running tasks waiting for worker recovery.")
	}
	if status.QueuedCount > 0 && status.RunningCount == 0 && status.OldestQueuedSeconds != nil && *status.OldestQueuedSeconds >= 300 {
		reasons = append(reasons, "WeChat Export queued tasks have been waiting for a worker for over 5 minutes.")
	}
	if strings.EqualFold(status.Health, "attention") && len(status.AttentionReasons) > 0 {
		reasons = append(reasons, "WeChat Export worker status requires attention.")
	}
	return reasons
}

func homeBusinessEnvConfigured(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}

func homeBusinessAnyEnvConfigured(keys ...string) bool {
	for _, key := range keys {
		if homeBusinessEnvConfigured(key) {
			return true
		}
	}
	return false
}

func homeBusinessEnvEnabled(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func hotWorkerStatusMaxAge() time.Duration {
	if value := positiveDurationMillisEnv("HOT_WORKER_HEALTH_MAX_AGE_MS"); value > 0 {
		return value
	}
	interval := positiveDurationMillisEnv("HOT_RSS_COLLECT_INTERVAL_MS")
	if interval <= 0 {
		interval = 30 * time.Minute
	}
	backoff := positiveDurationMillisEnv("HOT_RSS_COLLECT_MAX_BACKOFF_MS")
	if backoff <= 0 {
		backoff = 10 * time.Minute
	}
	maxAge := interval + backoff
	if interval*2 > maxAge {
		maxAge = interval * 2
	}
	return maxAge
}

type runtimeWorkerTargetInfo struct {
	ID            string
	ContainerName string
	DeployCommand string
}

type runtimeWorkerDockerManager struct {
	enabled    bool
	socketPath string
	reason     string
	client     *http.Client
}

func runtimeWorkerTarget(id string) (runtimeWorkerTargetInfo, bool) {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case workerNodeBusiness, "wechat-export", "image-workspace":
		return runtimeWorkerTargetInfo{
			ID:            workerNodeBusiness,
			ContainerName: envOrDefault("BUSINESS_WORKER_CONTAINER_NAME", "sub2api-business-worker"),
			DeployCommand: "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.business-worker.yml --profile business-worker up -d --build",
		}, true
	case workerNodeContent, "hot-collector", "hot-rss-collector", "x-auto", "xauto":
		return runtimeWorkerTargetInfo{
			ID:            workerNodeContent,
			ContainerName: envOrDefault("CONTENT_WORKER_CONTAINER_NAME", "sub2api-content-worker"),
			DeployCommand: "docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.content-worker.yml --profile content-worker up -d --build",
		}, true
	default:
		return runtimeWorkerTargetInfo{}, false
	}
}

func newRuntimeWorkerDockerManager() runtimeWorkerDockerManager {
	socketPath := envOrDefault("WORKER_MANAGER_DOCKER_SOCKET", "/var/run/docker.sock")
	manager := runtimeWorkerDockerManager{
		enabled:    homeBusinessEnvEnabled("WORKER_MANAGER_ENABLED"),
		socketPath: socketPath,
	}
	if !manager.enabled {
		manager.reason = "Worker management is disabled. Set WORKER_MANAGER_ENABLED=true and mount the Docker socket to enable restart/online/offline actions."
		return manager
	}
	if _, err := os.Stat(socketPath); err != nil {
		manager.enabled = false
		manager.reason = "Docker socket is not available at " + socketPath + "."
		return manager
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}
	manager.client = &http.Client{Transport: transport, Timeout: 15 * time.Second}
	return manager
}

func (m runtimeWorkerDockerManager) disabledReason() string {
	if m.reason != "" {
		return m.reason
	}
	if m.enabled {
		return ""
	}
	return "Worker management is disabled."
}

func (m runtimeWorkerDockerManager) enrich(worker adminWorkerRuntimeStatusDTO) adminWorkerRuntimeStatusDTO {
	target, ok := runtimeWorkerTarget(worker.NodeID)
	if !ok {
		target, ok = runtimeWorkerTarget(worker.ID)
	}
	if !ok {
		worker.ManagementReason = "Worker is not mapped to a managed container."
		return worker
	}
	worker.NodeID = target.ID
	worker.ContainerName = target.ContainerName
	worker.DeployCommand = target.DeployCommand
	worker.Actions = []string{"deploy", "restart", "start", "stop"}
	worker.Manageable = m.enabled
	if !m.enabled {
		worker.ManagementReason = m.disabledReason()
		return worker
	}
	state, err := m.containerState(context.Background(), target.ContainerName)
	if err != nil {
		worker.ContainerState = "unknown"
		worker.ManagementReason = err.Error()
		return worker
	}
	worker.ContainerState = state
	return worker
}

func (m runtimeWorkerDockerManager) containerState(ctx context.Context, containerName string) (string, error) {
	if m.client == nil {
		return "", errors.New(m.disabledReason())
	}
	status := struct {
		State struct {
			Status string `json:"Status"`
		} `json:"State"`
	}{}
	if err := m.doDocker(ctx, http.MethodGet, "/containers/"+containerName+"/json", nil, &status); err != nil {
		return "", err
	}
	if strings.TrimSpace(status.State.Status) == "" {
		return "unknown", nil
	}
	return status.State.Status, nil
}

func (m runtimeWorkerDockerManager) postDocker(ctx context.Context, path string) error {
	return m.doDocker(ctx, http.MethodPost, path, nil, nil)
}

func (m runtimeWorkerDockerManager) doDocker(ctx context.Context, method, path string, body io.Reader, out any) error {
	if m.client == nil {
		return errors.New(m.disabledReason())
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://docker"+path, body)
	if err != nil {
		return err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	content, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotModified {
		message := strings.TrimSpace(string(content))
		if message == "" {
			message = resp.Status
		}
		return fmt.Errorf("docker api %s %s failed: %s", method, path, message)
	}
	if out != nil && len(content) > 0 {
		if err := json.Unmarshal(content, out); err != nil {
			return err
		}
	}
	return nil
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func xAutoBaseURL() string {
	value := strings.TrimRight(strings.TrimSpace(os.Getenv("X_AUTO_BASE_URL")), "/")
	if value != "" {
		return value
	}
	return strings.TrimRight(strings.TrimSpace(os.Getenv("X_ATUO_BASE_URL")), "/")
}

func positiveDurationMillisEnv(key string) time.Duration {
	value, err := strconv.ParseInt(strings.TrimSpace(os.Getenv(key)), 10, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return time.Duration(value) * time.Millisecond
}
