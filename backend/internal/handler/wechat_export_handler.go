package handler

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type WeChatExportHandler struct {
	service *service.WeChatExportService
}

func NewWeChatExportHandler(service *service.WeChatExportService) *WeChatExportHandler {
	return &WeChatExportHandler{service: service}
}

func (h *WeChatExportHandler) GetSession(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	session, err := h.service.GetSession(c.Request.Context(), subject.UserID)
	if err != nil {
		if err == service.ErrWeChatSessionNotFound {
			response.Success(c, gin.H{"status": "not_connected"})
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, session)
}

func (h *WeChatExportHandler) CreateQRCodeSession(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	session, qrcodeURL, err := h.service.CreateQRCodeSession(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, gin.H{
		"session":    session,
		"qrcode_url": qrcodeURL,
	})
}

func (h *WeChatExportHandler) PollSession(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	sessionID, err := strconv.ParseInt(c.Param("sessionID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid session id")
		return
	}
	session, err := h.service.PollSession(c.Request.Context(), subject.UserID, sessionID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, session)
}

func (h *WeChatExportHandler) LogoutSession(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.service.LogoutSession(c.Request.Context(), subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func (h *WeChatExportHandler) SearchAccounts(c *gin.Context) {
	if err := h.ensureReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"items": []any{},
		"query": strings.TrimSpace(c.Query("q")),
	})
}

func (h *WeChatExportHandler) BindAccount(c *gin.Context) {
	if err := h.ensureReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req struct {
		FakeID string `json:"fakeid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	response.Success(c, gin.H{
		"fakeid": req.FakeID,
		"status": "bound_placeholder",
	})
}

func (h *WeChatExportHandler) SyncAccount(c *gin.Context) {
	if err := h.ensureReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"fakeid": c.Param("accountID"),
		"status": "sync_queued_placeholder",
	})
}

func (h *WeChatExportHandler) ListArticles(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListArticles(c.Request.Context(), subject.UserID, pagination.PaginationParams{Page: page, PageSize: pageSize})
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

func (h *WeChatExportHandler) ImportArticleLink(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req struct {
		Link string `json:"link" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	article, err := h.service.ImportArticleLink(c.Request.Context(), subject.UserID, req.Link)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, article)
}

func (h *WeChatExportHandler) QuoteTask(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req struct {
		ArticleIDs        []int64  `json:"article_ids"`
		Formats           []string `json:"formats"`
		IncludeEngagement bool     `json:"include_engagement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	quote, err := h.service.QuoteTask(c.Request.Context(), subject.UserID, service.CreateWeChatExportTaskInput{
		ArticleIDs:        req.ArticleIDs,
		Formats:           req.Formats,
		IncludeEngagement: req.IncludeEngagement,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, quote)
}

func (h *WeChatExportHandler) CreateTask(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req struct {
		ArticleIDs        []int64  `json:"article_ids"`
		Formats           []string `json:"formats"`
		IncludeEngagement bool     `json:"include_engagement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	task, err := h.service.CreateTask(c.Request.Context(), subject.UserID, service.CreateWeChatExportTaskInput{
		ArticleIDs:        req.ArticleIDs,
		Formats:           req.Formats,
		IncludeEngagement: req.IncludeEngagement,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Accepted(c, task)
}

func (h *WeChatExportHandler) ListTasks(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListTasks(c.Request.Context(), subject.UserID, pagination.PaginationParams{Page: page, PageSize: pageSize})
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

func (h *WeChatExportHandler) GetTask(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
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

func (h *WeChatExportHandler) ListArtifacts(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid task id")
		return
	}
	items, err := h.service.ListArtifacts(c.Request.Context(), subject.UserID, taskID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *WeChatExportHandler) DownloadArtifact(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
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
	if strings.TrimSpace(artifact.DownloadURL) != "" && strings.HasPrefix(artifact.DownloadURL, "http") {
		c.Redirect(http.StatusFound, artifact.DownloadURL)
		return
	}
	if strings.TrimSpace(artifact.StorageKey) == "" {
		response.NotFound(c, "Artifact file not found")
		return
	}
	c.FileAttachment(artifact.StorageKey, artifact.FileName)
}

func (h *WeChatExportHandler) WorkerClaimTask(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req struct {
		LeaseSeconds int64 `json:"lease_seconds"`
	}
	_ = c.ShouldBindJSON(&req)
	task, articles, err := h.service.ClaimNextTask(c.Request.Context(), req.LeaseSeconds)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if task == nil {
		response.Success(c, gin.H{"task": nil, "articles": []any{}})
		return
	}
	response.Success(c, gin.H{"task": task, "articles": articles})
}

func (h *WeChatExportHandler) WorkerCompleteTask(c *gin.Context) {
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
		Artifacts          []service.WeChatExportArtifactInput `json:"artifacts"`
		ResultManifestJSON string                              `json:"result_manifest_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	task, err := h.service.CompleteTask(c.Request.Context(), taskID, service.CompleteWeChatExportTaskInput{
		Artifacts:          req.Artifacts,
		ResultManifestJSON: req.ResultManifestJSON,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *WeChatExportHandler) WorkerFailTask(c *gin.Context) {
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
		Message string `json:"message"`
	}
	_ = c.ShouldBindJSON(&req)
	task, err := h.service.FailTask(c.Request.Context(), taskID, req.Message)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *WeChatExportHandler) ensureReady(c *gin.Context) error {
	if h == nil || h.service == nil {
		return service.ErrWeChatExportNotConfigured
	}
	if _, ok := middleware2.GetAuthSubjectFromContext(c); !ok {
		return infraerrors.Unauthorized("AUTH_REQUIRED", "user not authenticated")
	}
	return h.service.Health(c.Request.Context())
}

func (h *WeChatExportHandler) authSubject(c *gin.Context) (middleware2.AuthSubject, error) {
	if err := h.ensureReady(c); err != nil {
		return middleware2.AuthSubject{}, err
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		return middleware2.AuthSubject{}, infraerrors.Unauthorized("AUTH_REQUIRED", "user not authenticated")
	}
	return subject, nil
}

func (h *WeChatExportHandler) ensureWorkerReady(c *gin.Context) error {
	if h == nil || h.service == nil {
		return service.ErrWeChatExportNotConfigured
	}
	token := strings.TrimSpace(os.Getenv("WECHAT_EXPORT_WORKER_TOKEN"))
	if token != "" {
		got := strings.TrimSpace(c.GetHeader("X-WeChat-Worker-Token"))
		if got != token {
			return infraerrors.Unauthorized("WECHAT_WORKER_UNAUTHORIZED", "wechat worker token is invalid")
		}
		return h.service.Health(c.Request.Context())
	}
	ip := net.ParseIP(c.ClientIP())
	if ip == nil || (!ip.IsLoopback() && !ip.IsPrivate()) {
		return infraerrors.Unauthorized("WECHAT_WORKER_UNAUTHORIZED", "wechat worker token is required")
	}
	return h.service.Health(c.Request.Context())
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
