package admin

import (
	opsctx "github.com/Aias00/cloudbase/internal/ops"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
)

type opsSystemLogCleanupRequest struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`

	Level           string `json:"level"`
	Component       string `json:"component"`
	RequestID       string `json:"request_id"`
	ClientRequestID string `json:"client_request_id"`
	UserID          *int64 `json:"user_id"`
	APIKeyID        *int64 `json:"api_key_id"`
	AccountID       *int64 `json:"account_id"`
	Platform        string `json:"platform"`
	Model           string `json:"model"`
	Query           string `json:"q"`
}

func opsSystemLogFilterFromParsed(parsed *opsctx.SystemLogListFilter) *service.OpsSystemLogFilter {
	if parsed == nil {
		return &service.OpsSystemLogFilter{}
	}
	return &service.OpsSystemLogFilter{
		Page:            parsed.Page,
		PageSize:        parsed.PageSize,
		StartTime:       parsed.StartTime,
		EndTime:         parsed.EndTime,
		Level:           parsed.Level,
		Component:       parsed.Component,
		RequestID:       parsed.RequestID,
		ClientRequestID: parsed.ClientRequestID,
		UserID:          parsed.UserID,
		APIKeyID:        parsed.APIKeyID,
		AccountID:       parsed.AccountID,
		Platform:        parsed.Platform,
		Model:           parsed.Model,
		Query:           parsed.Query,
	}
}

func opsSystemLogCleanupFilterFromParsed(parsed *opsctx.SystemLogCleanupFilter) *service.OpsSystemLogCleanupFilter {
	if parsed == nil {
		return &service.OpsSystemLogCleanupFilter{}
	}
	return &service.OpsSystemLogCleanupFilter{
		StartTime:       parsed.StartTime,
		EndTime:         parsed.EndTime,
		Level:           parsed.Level,
		Component:       parsed.Component,
		RequestID:       parsed.RequestID,
		ClientRequestID: parsed.ClientRequestID,
		UserID:          parsed.UserID,
		APIKeyID:        parsed.APIKeyID,
		AccountID:       parsed.AccountID,
		Platform:        parsed.Platform,
		Model:           parsed.Model,
		Query:           parsed.Query,
	}
}

// ListSystemLogs returns indexed system logs.
// GET /api/v1/admin/ops/system-logs
func (h *OpsHandler) ListSystemLogs(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	page, pageSize := response.ParsePagination(c)
	pageSize = opsctx.CapPageSize(pageSize, 200)

	start, end, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	parsed, err := opsctx.ParseSystemLogListFilter(opsctx.SystemLogListFilterInput{
		StartTime:          start,
		EndTime:            end,
		Page:               page,
		PageSize:           pageSize,
		LevelRaw:           c.Query("level"),
		ComponentRaw:       c.Query("component"),
		RequestIDRaw:       c.Query("request_id"),
		ClientRequestIDRaw: c.Query("client_request_id"),
		UserIDRaw:          c.Query("user_id"),
		APIKeyIDRaw:        c.Query("api_key_id"),
		AccountIDRaw:       c.Query("account_id"),
		PlatformRaw:        c.Query("platform"),
		ModelRaw:           c.Query("model"),
		QueryRaw:           c.Query("q"),
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	filter := opsSystemLogFilterFromParsed(parsed)

	result, err := opsService.ListSystemLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Logs, int64(result.Total), result.Page, result.PageSize)
}

// CleanupSystemLogs deletes indexed system logs by filter.
// POST /api/v1/admin/ops/system-logs/cleanup
func (h *OpsHandler) CleanupSystemLogs(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	uid, ok := requireOpsUserID(c)
	if !ok {
		return
	}

	req, ok := bindOpsJSON[opsSystemLogCleanupRequest](c)
	if !ok {
		return
	}

	parsed, err := opsctx.ParseSystemLogCleanupFilter(opsctx.SystemLogCleanupFilterInput{
		StartTimeRaw:       req.StartTime,
		EndTimeRaw:         req.EndTime,
		LevelRaw:           req.Level,
		ComponentRaw:       req.Component,
		RequestIDRaw:       req.RequestID,
		ClientRequestIDRaw: req.ClientRequestID,
		UserID:             req.UserID,
		APIKeyID:           req.APIKeyID,
		AccountID:          req.AccountID,
		PlatformRaw:        req.Platform,
		ModelRaw:           req.Model,
		QueryRaw:           req.Query,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := opsSystemLogCleanupFilterFromParsed(parsed)

	deleted, err := opsService.CleanupSystemLogs(c.Request.Context(), filter, uid)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": deleted})
}

// GetSystemLogIngestionHealth returns sink health metrics.
// GET /api/v1/admin/ops/system-logs/health
func (h *OpsHandler) GetSystemLogIngestionHealth(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}
	response.Success(c, opsService.GetSystemLogSinkHealth())
}
