package admin

import (
	"net/http"
	"time"

	opsctx "github.com/Aias00/cloudbase/internal/ops"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/server/middleware"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
)

type OpsHandler struct {
	opsService *service.OpsService
}

// GetErrorLogByID returns ops error log detail.
// GET /api/v1/admin/ops/errors/:id
func (h *OpsHandler) GetErrorLogByID(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	id, err := opsctx.ParsePositiveID(c.Param("id"), "Invalid error id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	detail, err := opsService.GetErrorLogByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, detail)
}

type opsErrorLogFilterOptions struct {
	IncludeUserAPI       bool
	IncludeModel         bool
	ClearUpstreamPhase   bool
	FixedView            string
	FixedPhase           string
	FixedOwner           string
	FixedRequestID       string
	FixedClientRequestID string
}

func parseOpsErrorLogFilter(c *gin.Context, page, pageSize int, startTime, endTime time.Time, opts opsErrorLogFilterOptions) (*service.OpsErrorLogFilter, error) {
	viewRaw := c.Query("view")
	if opts.FixedView != "" {
		viewRaw = opts.FixedView
	}
	phaseRaw := c.Query("phase")
	if opts.FixedPhase != "" {
		phaseRaw = opts.FixedPhase
	}
	ownerRaw := c.Query("error_owner")
	if opts.FixedOwner != "" {
		ownerRaw = opts.FixedOwner
	}
	modelRaw := ""
	if opts.IncludeModel {
		modelRaw = c.Query("model")
	}
	userIDRaw := ""
	apiKeyIDRaw := ""
	if opts.IncludeUserAPI {
		userIDRaw = c.Query("user_id")
		apiKeyIDRaw = c.Query("api_key_id")
	}

	parsed, err := opsctx.ParseErrorLogFilter(opsctx.ErrorLogFilterInput{
		StartTime:          startTime,
		EndTime:            endTime,
		Page:               page,
		PageSize:           pageSize,
		ViewRaw:            viewRaw,
		PhaseRaw:           phaseRaw,
		OwnerRaw:           ownerRaw,
		SourceRaw:          c.Query("error_source"),
		QueryRaw:           c.Query("q"),
		UserQueryRaw:       c.Query("user_query"),
		ModelRaw:           modelRaw,
		RequestIDRaw:       opts.FixedRequestID,
		ClientRequestIDRaw: opts.FixedClientRequestID,
		PlatformRaw:        c.Query("platform"),
		GroupIDRaw:         c.Query("group_id"),
		AccountIDRaw:       c.Query("account_id"),
		UserIDRaw:          userIDRaw,
		APIKeyIDRaw:        apiKeyIDRaw,
		ResolvedRaw:        c.Query("resolved"),
		StatusCodesRaw:     c.Query("status_codes"),
		ClearUpstreamPhase: opts.ClearUpstreamPhase,
	})
	if err != nil {
		return nil, err
	}
	return opsErrorLogFilterFromParsed(parsed), nil
}

func opsErrorLogFilterFromParsed(parsed *opsctx.ErrorLogFilter) *service.OpsErrorLogFilter {
	if parsed == nil {
		return &service.OpsErrorLogFilter{}
	}
	return &service.OpsErrorLogFilter{
		StartTime:       parsed.StartTime,
		EndTime:         parsed.EndTime,
		Page:            parsed.Page,
		PageSize:        parsed.PageSize,
		Platform:        parsed.Platform,
		GroupID:         parsed.GroupID,
		AccountID:       parsed.AccountID,
		StatusCodes:     parsed.StatusCodes,
		Phase:           parsed.Phase,
		Owner:           parsed.Owner,
		Source:          parsed.Source,
		Resolved:        parsed.Resolved,
		Query:           parsed.Query,
		UserQuery:       parsed.UserQuery,
		RequestID:       parsed.RequestID,
		ClientRequestID: parsed.ClientRequestID,
		UserID:          parsed.UserID,
		APIKeyID:        parsed.APIKeyID,
		Model:           parsed.Model,
		View:            parsed.View,
	}
}

func NewOpsHandler(opsService *service.OpsService) *OpsHandler {
	return &OpsHandler{opsService: opsService}
}

func (h *OpsHandler) requireOpsService(c *gin.Context) (*service.OpsService, bool) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return nil, false
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return nil, false
	}
	return h.opsService, true
}

// applyOpsErrorSortParams reads sort_by/sort_order query params into the filter.
// Column whitelist and order normalization live in the repository; unknown
// values degrade to the default (created_at DESC), mirroring the usage list.
func applyOpsErrorSortParams(c *gin.Context, filter *service.OpsErrorLogFilter) {
	filter.SetSort(c.Query("sort_by"), c.Query("sort_order"))
}

func bindOpsJSON[T any](c *gin.Context) (*T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return nil, false
	}
	return &req, true
}

func requireOpsUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return 0, false
	}
	return subject.UserID, true
}

func opsRequestDetailFilterFromParsed(parsed *opsctx.RequestDetailFilter) *service.OpsRequestDetailFilter {
	if parsed == nil {
		return &service.OpsRequestDetailFilter{}
	}
	return &service.OpsRequestDetailFilter{
		StartTime:     parsed.StartTime,
		EndTime:       parsed.EndTime,
		Kind:          parsed.Kind,
		Platform:      parsed.Platform,
		GroupID:       parsed.GroupID,
		UserID:        parsed.UserID,
		APIKeyID:      parsed.APIKeyID,
		AccountID:     parsed.AccountID,
		Model:         parsed.Model,
		RequestID:     parsed.RequestID,
		Query:         parsed.Query,
		MinDurationMs: parsed.MinDurationMs,
		MaxDurationMs: parsed.MaxDurationMs,
		Sort:          parsed.Sort,
		Page:          parsed.Page,
		PageSize:      parsed.PageSize,
	}
}

// GetErrorLogs lists ops error logs.
// GET /api/v1/admin/ops/errors
func (h *OpsHandler) GetErrorLogs(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	page, pageSize := response.ParsePagination(c)
	// Ops list can be larger than standard admin tables.
	pageSize = opsctx.CapPageSize(pageSize, 500)

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	filter, err := parseOpsErrorLogFilter(c, page, pageSize, startTime, endTime, opsErrorLogFilterOptions{
		IncludeUserAPI:     true,
		IncludeModel:       true,
		ClearUpstreamPhase: true,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	applyOpsErrorSortParams(c, filter)
	result, err := opsService.GetErrorLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Errors, int64(result.Total), result.Page, result.PageSize)
}

// ListRequestErrors lists client-visible request errors.
// GET /api/v1/admin/ops/request-errors
func (h *OpsHandler) ListRequestErrors(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	page, pageSize := response.ParsePagination(c)
	pageSize = opsctx.CapPageSize(pageSize, 500)
	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	filter, err := parseOpsErrorLogFilter(c, page, pageSize, startTime, endTime, opsErrorLogFilterOptions{
		IncludeModel:       true,
		ClearUpstreamPhase: true,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	applyOpsErrorSortParams(c, filter)
	result, err := opsService.GetErrorLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Errors, int64(result.Total), result.Page, result.PageSize)
}

// GetRequestError returns request error detail.
// GET /api/v1/admin/ops/request-errors/:id
func (h *OpsHandler) GetRequestError(c *gin.Context) {
	// same storage; just proxy to existing detail
	h.GetErrorLogByID(c)
}

// ListRequestErrorUpstreamErrors lists upstream error logs correlated to a request error.
// GET /api/v1/admin/ops/request-errors/:id/upstream-errors
func (h *OpsHandler) ListRequestErrorUpstreamErrors(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	id, err := opsctx.ParsePositiveID(c.Param("id"), "Invalid error id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Load request error to get correlation keys.
	detail, err := opsService.GetErrorLogByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Correlate by request_id/client_request_id.
	correlationKey := opsctx.PickErrorCorrelationKey(detail.RequestID, detail.ClientRequestID)
	if correlationKey.RequestID == "" && correlationKey.ClientRequestID == "" {
		response.Paginated(c, []*service.OpsErrorLog{}, 0, 1, 10)
		return
	}

	page, pageSize := response.ParsePagination(c)
	pageSize = opsctx.CapPageSize(pageSize, 500)

	// Keep correlation window wide enough so linked upstream errors
	// are discoverable even when UI defaults to 1h elsewhere.
	startTime, endTime, err := parseOpsTimeRange(c, "30d")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	opts := opsErrorLogFilterOptions{
		FixedView:  "all",
		FixedPhase: "upstream",
		FixedOwner: "provider",
	}
	if correlationKey.RequestID != "" {
		opts.FixedRequestID = correlationKey.RequestID
	} else {
		opts.FixedClientRequestID = correlationKey.ClientRequestID
	}
	filter, err := parseOpsErrorLogFilter(c, page, pageSize, startTime, endTime, opts)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	applyOpsErrorSortParams(c, filter)
	result, err := opsService.GetErrorLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// If client asks for details, expand each upstream error log to include upstream response fields.
	if opsctx.ParseTruthyFlag(c.Query("include_detail")) {
		details := make([]*service.OpsErrorLogDetail, 0, len(result.Errors))
		for _, item := range result.Errors {
			if item == nil {
				continue
			}
			d, err := opsService.GetErrorLogByID(c.Request.Context(), item.ID)
			if err != nil || d == nil {
				continue
			}
			details = append(details, d)
		}
		response.Paginated(c, details, int64(result.Total), result.Page, result.PageSize)
		return
	}

	response.Paginated(c, result.Errors, int64(result.Total), result.Page, result.PageSize)
}

// ResolveRequestError toggles resolved status.
// PUT /api/v1/admin/ops/request-errors/:id/resolve
func (h *OpsHandler) ResolveRequestError(c *gin.Context) {
	h.UpdateErrorResolution(c)
}

// ListUpstreamErrors lists independent upstream errors.
// GET /api/v1/admin/ops/upstream-errors
func (h *OpsHandler) ListUpstreamErrors(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	page, pageSize := response.ParsePagination(c)
	pageSize = opsctx.CapPageSize(pageSize, 500)
	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	filter, err := parseOpsErrorLogFilter(c, page, pageSize, startTime, endTime, opsErrorLogFilterOptions{
		FixedPhase: "upstream",
		FixedOwner: "provider",
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	applyOpsErrorSortParams(c, filter)
	result, err := opsService.GetErrorLogs(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Errors, int64(result.Total), result.Page, result.PageSize)
}

// GetUpstreamError returns upstream error detail.
// GET /api/v1/admin/ops/upstream-errors/:id
func (h *OpsHandler) GetUpstreamError(c *gin.Context) {
	h.GetErrorLogByID(c)
}

// ResolveUpstreamError toggles resolved status.
// PUT /api/v1/admin/ops/upstream-errors/:id/resolve
func (h *OpsHandler) ResolveUpstreamError(c *gin.Context) {
	h.UpdateErrorResolution(c)
}

// ==================== Existing endpoints ====================

// ListRequestDetails returns a request-level list (success + error) for drill-down.
// GET /api/v1/admin/ops/requests
func (h *OpsHandler) ListRequestDetails(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	page, pageSize := response.ParsePagination(c)
	pageSize = opsctx.CapPageSize(pageSize, 100)

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	parsed, err := opsctx.ParseRequestDetailFilter(opsctx.RequestDetailFilterInput{
		StartTime:        startTime,
		EndTime:          endTime,
		Page:             page,
		PageSize:         pageSize,
		KindRaw:          c.Query("kind"),
		PlatformRaw:      c.Query("platform"),
		ModelRaw:         c.Query("model"),
		RequestIDRaw:     c.Query("request_id"),
		QueryRaw:         c.Query("q"),
		SortRaw:          c.Query("sort"),
		UserIDRaw:        c.Query("user_id"),
		APIKeyIDRaw:      c.Query("api_key_id"),
		AccountIDRaw:     c.Query("account_id"),
		GroupIDRaw:       c.Query("group_id"),
		MinDurationMsRaw: c.Query("min_duration_ms"),
		MaxDurationMsRaw: c.Query("max_duration_ms"),
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	filter := opsRequestDetailFilterFromParsed(parsed)

	out, err := opsService.ListRequestDetails(c.Request.Context(), filter)
	if err != nil {
		// Invalid sort/kind/platform etc should be a bad request; keep it simple.
		if opsctx.IsInvalidInputError(err) {
			response.BadRequestWithError(c, err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to list request details")
		return
	}

	response.Paginated(c, out.Items, out.Total, out.Page, out.PageSize)
}

type opsResolveRequest struct {
	Resolved bool `json:"resolved"`
}

// UpdateErrorResolution allows manual resolve/unresolve.
// PUT /api/v1/admin/ops/errors/:id/resolve
func (h *OpsHandler) UpdateErrorResolution(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	uid, ok := requireOpsUserID(c)
	if !ok {
		return
	}

	id, err := opsctx.ParsePositiveID(c.Param("id"), "Invalid error id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var req opsResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	if err := opsService.UpdateErrorResolution(c.Request.Context(), id, req.Resolved, &uid); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func parseOpsTimeRange(c *gin.Context, defaultRange string) (time.Time, time.Time, error) {
	input := opsTimeRangeInputFromQuery(c, defaultRange)
	return opsctx.ParseTimeRange(input)
}

func opsTimeRangeInputFromQuery(c *gin.Context, defaultRange string) opsctx.TimeRangeInput {
	return opsctx.TimeRangeInput{
		StartTimeRaw: c.Query("start_time"),
		EndTimeRaw:   c.Query("end_time"),
		TimeRangeRaw: c.Query("time_range"),
		DefaultRange: defaultRange,
		Now:          time.Now(),
	}
}
