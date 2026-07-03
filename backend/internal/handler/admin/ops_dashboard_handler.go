package admin

import (
	"fmt"
	"time"

	opsctx "github.com/Aias00/cloudbase/internal/ops"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
)

func parseOpsDashboardFilter(c *gin.Context, startTime, endTime time.Time) (*service.OpsDashboardFilter, error) {
	scoped, err := parseOpsPlatformGroupFilter(c)
	if err != nil {
		return nil, err
	}
	return opsDashboardFilterFromScope(startTime, endTime, scoped, parseOpsQueryMode(c)), nil
}

func opsDashboardFilterFromScope(startTime, endTime time.Time, scoped *opsctx.PlatformGroupFilter, queryMode service.OpsQueryMode) *service.OpsDashboardFilter {
	if scoped == nil {
		scoped = &opsctx.PlatformGroupFilter{}
	}
	return &service.OpsDashboardFilter{
		StartTime: startTime,
		EndTime:   endTime,
		Platform:  scoped.Platform,
		GroupID:   scoped.GroupID,
		QueryMode: queryMode,
	}
}

func parseOpsPlatformGroupFilter(c *gin.Context) (*opsctx.PlatformGroupFilter, error) {
	return opsctx.ParsePlatformGroupFilter(opsctx.PlatformGroupFilterInput{
		PlatformRaw: c.Query("platform"),
		GroupIDRaw:  c.Query("group_id"),
	})
}

// GetDashboardOverview returns vNext ops dashboard overview (raw path).
// GET /api/v1/admin/ops/dashboard/overview
func (h *OpsHandler) GetDashboardOverview(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	filter, err := parseOpsDashboardFilter(c, startTime, endTime)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := opsService.GetDashboardOverview(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

// GetDashboardThroughputTrend returns throughput time series (raw path).
// GET /api/v1/admin/ops/dashboard/throughput-trend
func (h *OpsHandler) GetDashboardThroughputTrend(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	filter, err := parseOpsDashboardFilter(c, startTime, endTime)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	bucketSeconds := opsctx.PickThroughputBucketSeconds(endTime.Sub(startTime))
	data, err := opsService.GetThroughputTrend(c.Request.Context(), filter, bucketSeconds)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

// GetDashboardLatencyHistogram returns the latency distribution histogram (success requests).
// GET /api/v1/admin/ops/dashboard/latency-histogram
func (h *OpsHandler) GetDashboardLatencyHistogram(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	filter, err := parseOpsDashboardFilter(c, startTime, endTime)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := opsService.GetLatencyHistogram(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

// GetDashboardErrorTrend returns error counts time series (raw path).
// GET /api/v1/admin/ops/dashboard/error-trend
func (h *OpsHandler) GetDashboardErrorTrend(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	filter, err := parseOpsDashboardFilter(c, startTime, endTime)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	bucketSeconds := opsctx.PickThroughputBucketSeconds(endTime.Sub(startTime))
	data, err := opsService.GetErrorTrend(c.Request.Context(), filter, bucketSeconds)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

// GetDashboardErrorDistribution returns error distribution by status code (raw path).
// GET /api/v1/admin/ops/dashboard/error-distribution
func (h *OpsHandler) GetDashboardErrorDistribution(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	filter, err := parseOpsDashboardFilter(c, startTime, endTime)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := opsService.GetErrorDistribution(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

// GetDashboardOpenAITokenStats returns OpenAI token efficiency stats grouped by model.
// GET /api/v1/admin/ops/dashboard/openai-token-stats
func (h *OpsHandler) GetDashboardOpenAITokenStats(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	filter, err := parseOpsOpenAITokenStatsFilter(c)
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	data, err := opsService.GetOpenAITokenStats(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

func parseOpsOpenAITokenStatsFilter(c *gin.Context) (*service.OpsOpenAITokenStatsFilter, error) {
	if c == nil {
		return nil, fmt.Errorf("invalid request")
	}
	filter, err := opsctx.ParseOpenAITokenStatsFilter(opsctx.OpenAITokenStatsFilterInput{
		TimeRangeRaw: c.Query("time_range"),
		PlatformRaw:  c.Query("platform"),
		GroupIDRaw:   c.Query("group_id"),
		TopNRaw:      c.Query("top_n"),
		PageRaw:      c.Query("page"),
		PageSizeRaw:  c.Query("page_size"),
		Now:          time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return opsOpenAITokenStatsFilterFromParsed(filter), nil
}

func opsOpenAITokenStatsFilterFromParsed(filter *opsctx.OpenAITokenStatsFilter) *service.OpsOpenAITokenStatsFilter {
	if filter == nil {
		return &service.OpsOpenAITokenStatsFilter{}
	}
	return &service.OpsOpenAITokenStatsFilter{
		TimeRange: filter.TimeRange,
		StartTime: filter.StartTime,
		EndTime:   filter.EndTime,
		Platform:  filter.Platform,
		GroupID:   filter.GroupID,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		TopN:      filter.TopN,
	}
}

func parseOpsQueryMode(c *gin.Context) service.OpsQueryMode {
	if c == nil {
		return ""
	}
	// Empty means "use server default" (DB setting ops_query_mode_default).
	return service.OpsQueryMode(opsctx.ParseOptionalQueryMode(c.Query("mode")))
}
