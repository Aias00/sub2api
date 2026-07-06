package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	opsctx "github.com/Aias00/cloudbase/internal/ops"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
)

func opsRealtimePayload(enabled bool, fields gin.H, timestamp *time.Time) gin.H {
	payload := gin.H{"enabled": enabled}
	for key, value := range fields {
		payload[key] = value
	}
	if timestamp != nil {
		payload["timestamp"] = timestamp.UTC()
	}
	return payload
}

func opsRealtimeTimestamp(t time.Time) *time.Time {
	t = t.UTC()
	return &t
}

// GetConcurrencyStats returns real-time concurrency usage aggregated by platform/group/account.
// GET /api/v1/admin/ops/concurrency
func (h *OpsHandler) GetConcurrencyStats(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	if !opsService.IsRealtimeMonitoringEnabled(c.Request.Context()) {
		response.Success(c, opsRealtimePayload(false, gin.H{
			"platform": map[string]*service.PlatformConcurrencyInfo{},
			"group":    map[int64]*service.GroupConcurrencyInfo{},
			"account":  map[int64]*service.AccountConcurrencyInfo{},
		}, opsRealtimeTimestamp(time.Now())))
		return
	}

	filter, err := parseOpsPlatformGroupFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	platform, group, account, collectedAt, err := opsService.GetConcurrencyStats(c.Request.Context(), filter.Platform, filter.GroupID)
	if err != nil {
		if isOpsRealtimeRequestCanceled(c, err) {
			return
		}
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, opsRealtimePayload(true, gin.H{
		"platform": platform,
		"group":    group,
		"account":  account,
	}, collectedAt))
}

// GetUserConcurrencyStats returns real-time concurrency usage for all active users.
// GET /api/v1/admin/ops/user-concurrency
func (h *OpsHandler) GetUserConcurrencyStats(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	if !opsService.IsRealtimeMonitoringEnabled(c.Request.Context()) {
		response.Success(c, opsRealtimePayload(false, gin.H{
			"user": map[int64]*service.UserConcurrencyInfo{},
		}, opsRealtimeTimestamp(time.Now())))
		return
	}

	users, collectedAt, err := opsService.GetUserConcurrencyStats(c.Request.Context())
	if err != nil {
		if isOpsRealtimeRequestCanceled(c, err) {
			return
		}
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, opsRealtimePayload(true, gin.H{"user": users}, collectedAt))
}

// GetAccountAvailability returns account availability statistics.
// GET /api/v1/admin/ops/account-availability
//
// Query params:
// - platform: optional
// - group_id: optional
func (h *OpsHandler) GetAccountAvailability(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	if !opsService.IsRealtimeMonitoringEnabled(c.Request.Context()) {
		response.Success(c, opsRealtimePayload(false, gin.H{
			"platform": map[string]*service.PlatformAvailability{},
			"group":    map[int64]*service.GroupAvailability{},
			"account":  map[int64]*service.AccountAvailability{},
		}, opsRealtimeTimestamp(time.Now())))
		return
	}

	parsed, err := parseOpsPlatformGroupFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	platformStats, groupStats, accountStats, collectedAt, err := opsService.GetAccountAvailabilityStats(c.Request.Context(), parsed.Platform, parsed.GroupID)
	if err != nil {
		if isOpsRealtimeRequestCanceled(c, err) {
			return
		}
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, opsRealtimePayload(true, gin.H{
		"platform": platformStats,
		"group":    groupStats,
		"account":  accountStats,
	}, collectedAt))
}

func isOpsRealtimeRequestCanceled(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	if c != nil && c.Request != nil && errors.Is(c.Request.Context().Err(), context.Canceled) {
		return true
	}
	return strings.Contains(err.Error(), "canceling statement due to user request")
}

func parseOpsRealtimeWindow(v string) (time.Duration, string, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "1min", "1m":
		return 1 * time.Minute, "1min", true
	case "5min", "5m":
		return 5 * time.Minute, "5min", true
	case "30min", "30m":
		return 30 * time.Minute, "30min", true
	case "1h", "60m", "60min":
		return 1 * time.Hour, "1h", true
	default:
		return 0, "", false
	}
}

// GetRealtimeTrafficSummary returns QPS/TPS current/peak/avg for the selected window.
// GET /api/v1/admin/ops/realtime-traffic
//
// Query params:
// - window: 1min|5min|30min|1h (default: 1min)
// - platform: optional
// - group_id: optional
func (h *OpsHandler) GetRealtimeTrafficSummary(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	windowRange, ok := opsctx.ParseRealtimeWindowRange(c.Query("window"), time.Now())
	if !ok {
		response.BadRequest(c, "Invalid window")
		return
	}

	parsed, err := parseOpsPlatformGroupFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if !opsService.IsRealtimeMonitoringEnabled(c.Request.Context()) {
		disabledSummary := &service.OpsRealtimeTrafficSummary{
			Window:    windowRange.Label,
			StartTime: windowRange.StartTime,
			EndTime:   windowRange.EndTime,
			Platform:  parsed.Platform,
			GroupID:   parsed.GroupID,
			QPS:       service.OpsRateSummary{},
			TPS:       service.OpsRateSummary{},
		}
		response.Success(c, opsRealtimePayload(false, gin.H{
			"summary": disabledSummary,
		}, &windowRange.EndTime))
		return
	}

	filter := opsDashboardFilterFromScope(windowRange.StartTime, windowRange.EndTime, parsed, service.OpsQueryModeRaw)

	summary, err := opsService.GetRealtimeTrafficSummary(c.Request.Context(), filter)
	if err != nil {
		if isOpsRealtimeRequestCanceled(c, err) {
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	if summary != nil {
		summary.Window = windowRange.Label
	}
	response.Success(c, opsRealtimePayload(true, gin.H{
		"summary": summary,
	}, &windowRange.EndTime))
}
