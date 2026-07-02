package handler

import (
	"archive/zip"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

func (h *WeChatExportHandler) ValidateSession(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	session, err := h.service.ValidateSession(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, session)
}

func (h *WeChatExportHandler) SearchAccounts(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	limit := maxInt(1, parseIntDefault(c.Query("limit"), 20))
	var items []service.WeChatAccount
	if strings.EqualFold(strings.TrimSpace(c.Query("remote")), "true") {
		items, err = h.service.SearchRemoteAccounts(c.Request.Context(), subject.UserID, c.Query("q"), limit)
	} else {
		items, err = h.service.SearchAccounts(c.Request.Context(), subject.UserID, c.Query("q"), limit)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"items": items,
		"query": strings.TrimSpace(c.Query("q")),
	})
}

func (h *WeChatExportHandler) BindAccount(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req service.BindWeChatAccountInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	account, err := h.service.BindAccount(c.Request.Context(), subject.UserID, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	// API contract: return sync_required flag to indicate client should trigger sync
	// This allows mobile/script clients to decide whether to auto-sync based on their UX
	response.Created(c, gin.H{
		"account":       account,
		"sync_required": true, // Client should sync articles after binding
	})
}

func (h *WeChatExportHandler) SyncAccount(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	// Parse optional begin parameter for continuation sync
	beginFrom := 0
	if beginStr := c.Query("begin"); beginStr != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(beginStr)); err == nil && parsed >= 0 {
			beginFrom = parsed
		}
	}
	account, result, err := h.service.SyncAccount(c.Request.Context(), subject.UserID, c.Param("accountID"), beginFrom)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"account": account,
		"status":  "synced",
		"result":  result,
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

func (h *WeChatExportHandler) GetWorkerStatus(c *gin.Context) {
	subject, err := h.authSubject(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	status, err := h.service.GetWorkerStatus(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
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

func (h *WeChatExportHandler) CancelTask(c *gin.Context) {
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
	task, err := h.service.CancelTask(c.Request.Context(), subject.UserID, taskID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *WeChatExportHandler) RetryTask(c *gin.Context) {
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
	task, err := h.service.RetryTask(c.Request.Context(), subject.UserID, taskID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *WeChatExportHandler) ListTaskLogs(c *gin.Context) {
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
	items, err := h.service.ListTaskLogs(c.Request.Context(), subject.UserID, taskID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
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

func (h *WeChatExportHandler) DownloadTaskArtifactsZip(c *gin.Context) {
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
	entries := make([]wechatArtifactZipEntry, 0, len(items))
	seenNames := map[string]int{}
	for _, artifact := range items {
		storagePath, ok := resolveWeChatExportStoragePath(artifact.StorageKey)
		if ok {
			info, err := os.Stat(storagePath)
			if err == nil && !info.IsDir() {
				entries = append(entries, wechatArtifactZipEntry{
					Artifact: artifact,
					Path:     storagePath,
					Name:     uniqueWeChatArtifactZipName(artifact.FileName, artifact.Format, seenNames),
				})
				continue
			}
		}
		if isWeChatArtifactRemoteZipAllowed(artifact.DownloadURL) {
			entries = append(entries, wechatArtifactZipEntry{
				Artifact:  artifact,
				RemoteURL: strings.TrimSpace(artifact.DownloadURL),
				Name:      uniqueWeChatArtifactZipName(artifact.FileName, artifact.Format, seenNames),
			})
		}
	}
	if len(entries) == 0 {
		response.NotFound(c, "No local or allowed remote artifact files are available for zip download")
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="wechat-export-task-%d.zip"`, taskID))
	zipWriter := zip.NewWriter(c.Writer)
	defer func() { _ = zipWriter.Close() }()
	for _, entry := range entries {
		if err := writeWeChatArtifactZipEntry(zipWriter, entry); err != nil {
			_ = c.Error(err)
			return
		}
	}
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
	if isWeChatArtifactRemoteDownloadAllowed(artifact.DownloadURL) {
		c.Redirect(http.StatusFound, strings.TrimSpace(artifact.DownloadURL))
		return
	}
	if strings.TrimSpace(artifact.StorageKey) == "" {
		response.NotFound(c, "Artifact file not found")
		return
	}
	storagePath, ok := resolveWeChatExportStoragePath(artifact.StorageKey)
	if !ok {
		response.NotFound(c, "Artifact file not found")
		return
	}
	if _, err := os.Stat(storagePath); err != nil {
		response.NotFound(c, "Artifact file not found")
		return
	}
	c.FileAttachment(storagePath, artifact.FileName)
}

type wechatArtifactZipEntry struct {
	Artifact  service.WeChatExportArtifact
	Path      string
	RemoteURL string
	Name      string
}

func writeWeChatArtifactZipEntry(zipWriter *zip.Writer, entry wechatArtifactZipEntry) error {
	header := &zip.FileHeader{
		Name:   entry.Name,
		Method: zip.Deflate,
	}
	header.SetMode(0o600)
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if entry.RemoteURL != "" {
		return copyWeChatArtifactRemoteZipEntry(writer, entry.RemoteURL)
	}
	file, err := os.Open(entry.Path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	_, err = io.Copy(writer, file)
	return err
}

func copyWeChatArtifactRemoteZipEntry(writer io.Writer, remoteURL string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(remoteURL)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("remote artifact returned status %d", resp.StatusCode)
	}
	_, err = io.Copy(writer, resp.Body)
	return err
}

func isWeChatArtifactRemoteZipAllowed(rawURL string) bool {
	return isWeChatArtifactRemoteDownloadAllowed(rawURL)
}

func isWeChatArtifactRemoteDownloadAllowed(rawURL string) bool {
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
	for _, allowed := range allowedWeChatArtifactZipHosts() {
		if host == allowed || strings.ToLower(parsed.Host) == allowed {
			return true
		}
	}
	return false
}

func allowedWeChatArtifactZipHosts() []string {
	values := []string{
		os.Getenv("WECHAT_EXPORT_ZIP_REMOTE_HOST_ALLOWLIST"),
		os.Getenv("WECHAT_EXPORT_ARTIFACT_PUBLIC_BASE_URL"),
		os.Getenv("WECHAT_EXPORT_PUBLIC_ARTIFACT_BASE_URL"),
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

func uniqueWeChatArtifactZipName(fileName string, format string, seen map[string]int) string {
	name := sanitizeWeChatArtifactFileName(fileName)
	if name == "" {
		name = "wechat-export." + sanitizeWeChatArtifactFileName(format)
	}
	if name == "wechat-export." {
		name = "wechat-export.txt"
	}
	count := seen[name]
	seen[name] = count + 1
	if count == 0 {
		return name
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s-%d%s", base, count+1, ext)
}

func sanitizeWeChatArtifactFileName(value string) string {
	name := strings.TrimSpace(filepath.Base(value))
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.Trim(name, ". ")
	return name
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
	task, articles, leaseToken, err := h.service.ClaimNextTask(c.Request.Context(), req.LeaseSeconds)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if task == nil {
		response.Success(c, gin.H{"task": nil, "articles": []any{}, "lease_token": ""})
		return
	}
	// Phase 4：返回leaseToken给worker
	response.Success(c, gin.H{
		"task":        task,
		"articles":    articles,
		"lease_token": leaseToken,
	})
}

func (h *WeChatExportHandler) WorkerHealth(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"ok":      true,
		"service": "wechat-export-worker",
	})
}

func (h *WeChatExportHandler) WorkerRuntimeConfig(c *gin.Context) {
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

func (h *WeChatExportHandler) WorkerEnrichArticle(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	articleID, err := strconv.ParseInt(c.Param("articleID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid article id")
		return
	}
	var req struct {
		UserID         int64   `json:"user_id"`
		AccountFakeID  string  `json:"account_fakeid"`
		AccountName    string  `json:"account_name"`
		AccountAlias   string  `json:"account_alias"`
		AccountAvatar  string  `json:"account_avatar"`
		AccountDesc    string  `json:"account_description"`
		Title          string  `json:"title"`
		Author         string  `json:"author"`
		Cover          string  `json:"cover"`
		Digest         string  `json:"digest"`
		PublishAt      *string `json:"publish_at"`
		IsOriginal     bool    `json:"is_original"`
		IsPaySubscribe bool    `json:"is_pay_subscribe"`
		ContentStatus  string  `json:"content_status"`
		MetadataJSON   string  `json:"metadata_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	var publishAt *time.Time
	if req.PublishAt != nil && strings.TrimSpace(*req.PublishAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.PublishAt))
		if err != nil {
			response.BadRequest(c, "Invalid publish_at")
			return
		}
		publishAt = &parsed
	}
	article, err := h.service.EnrichArticle(c.Request.Context(), service.EnrichWeChatArticleInput{
		ArticleID:      articleID,
		UserID:         req.UserID,
		AccountFakeID:  req.AccountFakeID,
		AccountName:    req.AccountName,
		AccountAlias:   req.AccountAlias,
		AccountAvatar:  req.AccountAvatar,
		AccountDesc:    req.AccountDesc,
		Title:          req.Title,
		Author:         req.Author,
		Cover:          req.Cover,
		Digest:         req.Digest,
		PublishAt:      publishAt,
		IsOriginal:     req.IsOriginal,
		IsPaySubscribe: req.IsPaySubscribe,
		ContentStatus:  req.ContentStatus,
		MetadataJSON:   req.MetadataJSON,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, article)
}

func (h *WeChatExportHandler) WorkerFetchArticleEngagement(c *gin.Context) {
	if err := h.ensureWorkerReady(c); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	articleID, err := strconv.ParseInt(c.Param("articleID"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid article id")
		return
	}
	var req struct {
		UserID       int64  `json:"user_id"`
		Link         string `json:"link"`
		MetadataJSON string `json:"metadata_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	result, err := h.service.FetchArticleEngagement(c.Request.Context(), service.FetchWeChatArticleEngagementInput{
		ArticleID:    articleID,
		UserID:       req.UserID,
		Link:         req.Link,
		MetadataJSON: req.MetadataJSON,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *WeChatExportHandler) WorkerAddTaskLog(c *gin.Context) {
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
		LeaseToken string `json:"lease_token" binding:"required"` // Phase 4：新增且required（强制验证）
		Event      string `json:"event" binding:"required"`
		Status     string `json:"status"`
		Message    string `json:"message"`
		MetaJSON   string `json:"meta_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err) // lease_token missing → 400
		return
	}
	log, err := h.service.AddTaskLog(c.Request.Context(), taskID, req.LeaseToken, service.AddWeChatExportTaskLogInput{
		Event:    req.Event,
		Status:   req.Status,
		Message:  req.Message,
		MetaJSON: req.MetaJSON,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, log)
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
		LeaseToken         string                              `json:"lease_token" binding:"required"` // Phase 4：新增且required
		Artifacts          []service.WeChatExportArtifactInput `json:"artifacts"`
		ResultManifestJSON string                              `json:"result_manifest_json"`
		FailedArticleCount int                                 `json:"failed_article_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	task, err := h.service.CompleteTask(c.Request.Context(), taskID, req.LeaseToken, service.CompleteWeChatExportTaskInput{
		Artifacts:          req.Artifacts,
		ResultManifestJSON: req.ResultManifestJSON,
		FailedArticleCount: req.FailedArticleCount,
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
		LeaseToken string `json:"lease_token" binding:"required"` // Phase 4：新增且required
		Message    string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	task, err := h.service.FailTask(c.Request.Context(), taskID, req.LeaseToken, req.Message)
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
		tokenHash := sha256.Sum256([]byte(token))
		gotHash := sha256.Sum256([]byte(got))
		if subtle.ConstantTimeCompare(gotHash[:], tokenHash[:]) != 1 {
			return infraerrors.Unauthorized("WECHAT_WORKER_UNAUTHORIZED", "wechat worker token is invalid")
		}
		return h.service.Health(c.Request.Context())
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN")), "true") {
		return infraerrors.Unauthorized("WECHAT_WORKER_UNAUTHORIZED", "wechat worker token is required")
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

func parseIntDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func resolveWeChatExportStoragePath(storageKey string) (string, bool) {
	cleanKey := filepath.Clean(strings.TrimSpace(storageKey))
	if cleanKey == "." || cleanKey == "" || !filepath.IsAbs(cleanKey) {
		return "", false
	}
	root := strings.TrimSpace(os.Getenv("WECHAT_EXPORT_STORAGE_ROOT"))
	if root == "" {
		root = strings.TrimSpace(os.Getenv("WECHAT_EXPORT_OUTPUT_DIR"))
	}
	if root == "" {
		root = "/app/data/wechat-export"
	}
	cleanRoot := filepath.Clean(root)
	relative, err := filepath.Rel(cleanRoot, cleanKey)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || relative == ".." {
		return "", false
	}
	return cleanKey, true
}
