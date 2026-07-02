package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/cloudbase/internal/pkg/errors"
	"github.com/Wei-Shaw/cloudbase/internal/pkg/pagination"
	"github.com/Wei-Shaw/cloudbase/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/cloudbase/internal/server/middleware"
	"github.com/Wei-Shaw/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
)

type ImageWorkspaceHandler struct {
	service *service.ImageWorkspaceService
}

func NewImageWorkspaceHandler(service *service.ImageWorkspaceService) *ImageWorkspaceHandler {
	return &ImageWorkspaceHandler{service: service}
}

func (h *ImageWorkspaceHandler) CreateTask(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	var req struct {
		Prompt         string `json:"prompt" binding:"required"`
		NegativePrompt string `json:"negative_prompt"`
		Model          string `json:"model"`
		Provider       string `json:"provider"`
		Size           string `json:"size"`
		Quality        string `json:"quality"`
		Style          string `json:"style"`
		Seed           *int64 `json:"seed"`
		BatchSize      int    `json:"batch_size"`
		TemplateID     *int64 `json:"template_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	task, err := h.service.CreateTask(c.Request.Context(), subject.UserID, service.CreateImageWorkspaceTaskInput{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Model:          req.Model,
		Provider:       req.Provider,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		Seed:           req.Seed,
		BatchSize:      req.BatchSize,
		TemplateID:     req.TemplateID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Accepted(c, task)
}

func (h *ImageWorkspaceHandler) ListTasks(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")
	items, result, err := h.service.ListTasks(c.Request.Context(), subject.UserID, pagination.PaginationParams{Page: page, PageSize: pageSize}, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, &response.PaginationResult{
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Pages:    result.Pages,
	})
}

func (h *ImageWorkspaceHandler) ListModels(c *gin.Context) {
	if _, ok := middleware2.GetAuthSubjectFromContext(c); !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	models, err := h.service.ListModels(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"models": models})
}

func (h *ImageWorkspaceHandler) GetTask(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	task, err := h.service.GetTask(c.Request.Context(), subject.UserID, taskID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *ImageWorkspaceHandler) CancelTask(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	task, err := h.service.CancelTask(c.Request.Context(), subject.UserID, taskID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *ImageWorkspaceHandler) RetryTask(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	task, err := h.service.RetryTask(c.Request.Context(), subject.UserID, taskID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Accepted(c, task)
}

func (h *ImageWorkspaceHandler) DownloadArtifact(c *gin.Context) {
	started := time.Now()
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	artifactID, err := strconv.ParseInt(c.Param("artifactID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid artifact id")
		return
	}
	artifact, err := h.service.GetArtifact(c.Request.Context(), subject.UserID, artifactID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	filename := imageWorkspaceArtifactDownloadName(artifact)
	storagePath, ok := resolveImageWorkspaceStoragePath(artifact.StorageKey)
	if !ok {
		if h.redirectImageWorkspaceRemoteArtifact(c, artifact, filename, started) {
			return
		}
		if h.serveImageWorkspaceRemoteArtifact(c, artifact.ImageURL, filename) {
			return
		}
		response.NotFound(c, "Artifact file not found")
		return
	}
	if _, err := os.Stat(storagePath); err != nil {
		if h.redirectImageWorkspaceRemoteArtifact(c, artifact, filename, started) {
			return
		}
		if h.serveImageWorkspaceRemoteArtifact(c, artifact.ImageURL, filename) {
			return
		}
		response.NotFound(c, "Artifact file not found")
		return
	}
	slog.Info("image_workspace.artifact_download",
		"mode", "local_file",
		"artifact_id", artifact.ID,
		"task_id", artifact.TaskID,
		"storage_provider", artifact.StorageProvider,
		"duration_ms", time.Since(started).Milliseconds(),
	)
	c.FileAttachment(storagePath, filename)
}

func (h *ImageWorkspaceHandler) ListTemplates(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	items, err := h.service.ListTemplates(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *ImageWorkspaceHandler) UpsertTemplate(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	var req struct {
		ID             int64  `json:"id"`
		Title          string `json:"title" binding:"required"`
		Description    string `json:"description"`
		Prompt         string `json:"prompt" binding:"required"`
		NegativePrompt string `json:"negative_prompt"`
		Model          string `json:"model"`
		Size           string `json:"size"`
		Quality        string `json:"quality"`
		Style          string `json:"style"`
		IsDefault      bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	template, err := h.service.UpsertTemplate(c.Request.Context(), subject.UserID, service.UpsertImageWorkspaceTemplateInput{
		ID:             req.ID,
		Title:          req.Title,
		Description:    req.Description,
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Model:          req.Model,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		IsDefault:      req.IsDefault,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, template)
}

func (h *ImageWorkspaceHandler) DeleteTemplate(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	templateID, err := strconv.ParseInt(c.Param("templateID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid template id")
		return
	}
	if err := h.service.DeleteTemplate(c.Request.Context(), subject.UserID, templateID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ImageWorkspaceHandler) ListUsageRecords(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListUsageRecords(c.Request.Context(), subject.UserID, pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, &response.PaginationResult{
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Pages:    result.Pages,
	})
}

func (h *ImageWorkspaceHandler) WorkerHealth(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

func (h *ImageWorkspaceHandler) WorkerStatus(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	status, err := h.service.GetWorkerStatus(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
}

func (h *ImageWorkspaceHandler) WorkerRuntimeConfig(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	cfg, err := h.service.GetWorkerRuntimeConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

func (h *ImageWorkspaceHandler) WorkerClaimTask(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req struct {
		LeaseSeconds int64 `json:"lease_seconds"`
	}
	_ = c.ShouldBindJSON(&req)
	task, err := h.service.ClaimNextTask(c.Request.Context(), req.LeaseSeconds)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if task == nil {
		response.Success(c, gin.H{"task": nil})
		return
	}
	response.Success(c, gin.H{"task": task})
}

func (h *ImageWorkspaceHandler) WorkerCompleteTask(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	var req struct {
		Artifacts  []service.ImageWorkspaceArtifactInput `json:"artifacts"`
		ResultJSON string                                `json:"result_json"`
		Cost       float64                               `json:"cost"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	task, err := h.service.CompleteTask(c.Request.Context(), taskID, service.CompleteImageWorkspaceTaskInput{
		Artifacts:  req.Artifacts,
		ResultJSON: req.ResultJSON,
		Cost:       req.Cost,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *ImageWorkspaceHandler) WorkerFailTask(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	var req struct {
		Message    string `json:"message"`
		ResultJSON string `json:"result_json"`
	}
	_ = c.ShouldBindJSON(&req)
	task, err := h.service.FailTask(c.Request.Context(), taskID, service.FailImageWorkspaceTaskInput{Message: req.Message, ResultJSON: req.ResultJSON})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *ImageWorkspaceHandler) ensureWorkerReady(c *gin.Context) error {
	if h == nil || h.service == nil {
		return infraerrors.InternalServer("IMAGE_WORKSPACE_NOT_CONFIGURED", "image workspace capability is not configured")
	}
	token := strings.TrimSpace(os.Getenv("IMAGE_WORKSPACE_WORKER_TOKEN"))
	if token != "" {
		got := strings.TrimSpace(c.GetHeader("X-Image-Workspace-Worker-Token"))
		if got == "" || got != token {
			return infraerrors.Unauthorized("IMAGE_WORKSPACE_WORKER_UNAUTHORIZED", "image workspace worker token is invalid")
		}
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("IMAGE_WORKSPACE_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN")), "true") {
		return infraerrors.Unauthorized("IMAGE_WORKSPACE_WORKER_UNAUTHORIZED", "image workspace worker token is required")
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		host = c.ClientIP()
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() && !ip.IsPrivate() {
		return infraerrors.Unauthorized("IMAGE_WORKSPACE_WORKER_UNAUTHORIZED", "image workspace worker token is required")
	}
	return nil
}

func resolveImageWorkspaceStoragePath(storageKey string) (string, bool) {
	cleanKey := filepath.Clean(strings.TrimSpace(storageKey))
	if cleanKey == "." || cleanKey == "" || !filepath.IsAbs(cleanKey) {
		return "", false
	}
	root := strings.TrimSpace(os.Getenv("IMAGE_WORKSPACE_STORAGE_ROOT"))
	if root == "" {
		root = strings.TrimSpace(os.Getenv("IMAGE_WORKSPACE_OUTPUT_DIR"))
	}
	if root == "" {
		root = "/app/data/image-workspace"
	}
	cleanRoot := filepath.Clean(root)
	relative, err := filepath.Rel(cleanRoot, cleanKey)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || relative == ".." {
		return "", false
	}
	return cleanKey, true
}

func (h *ImageWorkspaceHandler) serveImageWorkspaceRemoteArtifact(c *gin.Context, rawURL string, filename string) bool {
	started := time.Now()
	if !h.isImageWorkspaceRemoteArtifactAllowed(c.Request.Context(), rawURL) {
		return false
	}
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, strings.TrimSpace(rawURL), nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("image_workspace.artifact_download",
			"mode", "remote_proxy",
			"status", "fetch_failed",
			"duration_ms", time.Since(started).Milliseconds(),
			"error", err.Error(),
		)
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "IMAGE_WORKSPACE_ARTIFACT_FETCH_FAILED", "failed to fetch artifact file").WithCause(err))
		return true
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		slog.Warn("image_workspace.artifact_download",
			"mode", "remote_proxy",
			"status", "bad_status",
			"http_status", resp.StatusCode,
			"duration_ms", time.Since(started).Milliseconds(),
		)
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "IMAGE_WORKSPACE_ARTIFACT_FETCH_FAILED", fmt.Sprintf("artifact file returned status %d", resp.StatusCode)))
		return true
	}
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Status(http.StatusOK)
	bytesWritten, copyErr := io.Copy(c.Writer, resp.Body)
	if copyErr != nil {
		slog.Warn("image_workspace.artifact_download",
			"mode", "remote_proxy",
			"status", "copy_failed",
			"bytes", bytesWritten,
			"duration_ms", time.Since(started).Milliseconds(),
			"error", copyErr.Error(),
		)
		return true
	}
	slog.Info("image_workspace.artifact_download",
		"mode", "remote_proxy",
		"status", "ok",
		"bytes", bytesWritten,
		"duration_ms", time.Since(started).Milliseconds(),
	)
	return true
}

func (h *ImageWorkspaceHandler) redirectImageWorkspaceRemoteArtifact(c *gin.Context, artifact *service.ImageWorkspaceArtifact, filename string, started time.Time) bool {
	if artifact == nil || strings.EqualFold(strings.TrimSpace(artifact.StorageProvider), "upstream") {
		return false
	}
	rawURL := strings.TrimSpace(artifact.ImageURL)
	if !h.isImageWorkspaceRemoteArtifactAllowed(c.Request.Context(), rawURL) {
		return false
	}
	slog.Info("image_workspace.artifact_download",
		"mode", "remote_redirect",
		"artifact_id", artifact.ID,
		"task_id", artifact.TaskID,
		"storage_provider", artifact.StorageProvider,
		"filename", filename,
		"duration_ms", time.Since(started).Milliseconds(),
	)
	c.Redirect(http.StatusFound, rawURL)
	return true
}

func redirectImageWorkspaceRemoteArtifact(c *gin.Context, artifact *service.ImageWorkspaceArtifact, filename string, started time.Time) bool {
	return ((*ImageWorkspaceHandler)(nil)).redirectImageWorkspaceRemoteArtifact(c, artifact, filename, started)
}

func (h *ImageWorkspaceHandler) isImageWorkspaceRemoteArtifactAllowed(ctx context.Context, rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return false
	}
	for _, allowed := range h.allowedImageWorkspaceArtifactHosts(ctx) {
		if host == allowed || strings.ToLower(parsed.Host) == allowed {
			return true
		}
	}
	return false
}

func isImageWorkspaceRemoteArtifactAllowed(rawURL string) bool {
	return ((*ImageWorkspaceHandler)(nil)).isImageWorkspaceRemoteArtifactAllowed(context.Background(), rawURL)
}

func (h *ImageWorkspaceHandler) allowedImageWorkspaceArtifactHosts(ctx context.Context) []string {
	values := []string{
		os.Getenv("IMAGE_WORKSPACE_ARTIFACT_REMOTE_HOST_ALLOWLIST"),
		os.Getenv("IMAGE_WORKSPACE_PUBLIC_ARTIFACT_BASE_URL"),
		os.Getenv("MEDIA_CDN_BASE_URL"),
	}
	if h != nil && h.service != nil {
		if cfg, err := h.service.GetWorkerRuntimeConfig(ctx); err == nil {
			values = append(values, cfg.ObjectStorage.PublicBaseURL, cfg.MediaCDNBaseURL)
		}
	}
	hosts := make([]string, 0)
	seen := map[string]bool{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			host := strings.ToLower(trimmed)
			if parsed, err := url.Parse(trimmed); err == nil && parsed.Hostname() != "" {
				host = strings.ToLower(parsed.Hostname())
			}
			if !seen[host] {
				seen[host] = true
				hosts = append(hosts, host)
			}
		}
	}
	return hosts
}

func imageWorkspaceArtifactDownloadName(artifact *service.ImageWorkspaceArtifact) string {
	if artifact == nil {
		return "image-workspace-artifact.bin"
	}
	var extension string
	switch strings.ToLower(strings.TrimSpace(artifact.MimeType)) {
	case "image/jpeg", "image/jpg":
		extension = "jpg"
	case "image/webp":
		extension = "webp"
	case "image/png", "":
		extension = "png"
	default:
		extension = "bin"
	}
	if artifact.TaskID > 0 && artifact.ID > 0 {
		return fmt.Sprintf("image-task-%d-%d.%s", artifact.TaskID, artifact.ID, extension)
	}
	if artifact.ID > 0 {
		return fmt.Sprintf("image-artifact-%d.%s", artifact.ID, extension)
	}
	return "image-workspace-artifact." + extension
}
