package handler

import (
	"context"
	"encoding/json"
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
	response.Success(c, gin.H{
		"workers": []adminWorkerRuntimeStatusDTO{
			h.adminImageWorkspaceWorkerStatus(c.Request.Context()),
			h.adminWeChatExportWorkerStatus(c.Request.Context()),
			h.adminHotCollectorWorkerStatus(),
		},
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
	if reasons := imageWorkspaceRuntimeReadinessGaps(); len(reasons) > 0 {
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

func imageWorkspaceRuntimeReadinessGaps() []string {
	reasons := make([]string, 0, 4)
	if !homeBusinessEnvConfigured("IMAGE_WORKSPACE_WORKER_TOKEN") &&
		!homeBusinessEnvEnabled("IMAGE_WORKSPACE_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN") {
		reasons = append(reasons, "Image Workspace worker authentication is not configured.")
	}
	if !homeBusinessEnvConfigured("IMAGE_WORKSPACE_UPSTREAM_API_KEY") &&
		!homeBusinessEnvEnabled("IMAGE_WORKSPACE_CAPABILITY_ASSUME_EXTERNAL_WORKER_READY") {
		reasons = append(reasons, "Image Workspace upstream image provider is not configured.")
	}
	if homeBusinessEnvEnabled("IMAGE_WORKSPACE_OBJECT_STORAGE_ENABLED") {
		if !homeBusinessAnyEnvConfigured("IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT", "IMAGE_WORKSPACE_R2_ENDPOINT", "IMAGE_WORKSPACE_R2_ACCOUNT_ID") {
			reasons = append(reasons, "Image Workspace object storage endpoint/account is not configured.")
		}
		if !homeBusinessAnyEnvConfigured("IMAGE_WORKSPACE_OBJECT_STORAGE_BUCKET", "IMAGE_WORKSPACE_R2_BUCKET", "IMAGE_WORKSPACE_R2_BUCKET_NAME") {
			reasons = append(reasons, "Image Workspace object storage bucket is not configured.")
		}
		if !homeBusinessAnyEnvConfigured("IMAGE_WORKSPACE_OBJECT_STORAGE_PUBLIC_BASE_URL", "IMAGE_WORKSPACE_R2_PUBLIC_BASE_URL", "IMAGE_WORKSPACE_R2_DOMAIN") {
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

func positiveDurationMillisEnv(key string) time.Duration {
	value, err := strconv.ParseInt(strings.TrimSpace(os.Getenv(key)), 10, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return time.Duration(value) * time.Millisecond
}
