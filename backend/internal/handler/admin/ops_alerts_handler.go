package admin

import (
	"encoding/json"
	"time"

	opsctx "github.com/Aias00/cloudbase/internal/ops"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/server/middleware"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type opsAlertEventStatusRequest struct {
	Status string `json:"status"`
}

type opsAlertSilenceRequest struct {
	RuleID   int64   `json:"rule_id"`
	Platform string  `json:"platform"`
	GroupID  *int64  `json:"group_id"`
	Region   *string `json:"region"`
	Until    string  `json:"until"`
	Reason   string  `json:"reason"`
}

// ListAlertRules returns all ops alert rules.
// GET /api/v1/admin/ops/alert-rules
func (h *OpsHandler) ListAlertRules(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	rules, err := opsService.ListAlertRules(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, rules)
}

// CreateAlertRule creates an ops alert rule.
// POST /api/v1/admin/ops/alert-rules
func (h *OpsHandler) CreateAlertRule(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	rule, ok := bindOpsAlertRuleRequest(c, 0)
	if !ok {
		return
	}

	created, err := opsService.CreateAlertRule(c.Request.Context(), &rule)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, created)
}

// UpdateAlertRule updates an existing ops alert rule.
// PUT /api/v1/admin/ops/alert-rules/:id
func (h *OpsHandler) UpdateAlertRule(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	id, err := opsctx.ParsePositiveID(c.Param("id"), "Invalid rule ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	rule, ok := bindOpsAlertRuleRequest(c, id)
	if !ok {
		return
	}

	updated, err := opsService.UpdateAlertRule(c.Request.Context(), &rule)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, updated)
}

func bindOpsAlertRuleRequest(c *gin.Context, id int64) (service.OpsAlertRule, bool) {
	var raw map[string]json.RawMessage
	if err := c.ShouldBindBodyWith(&raw, binding.JSON); err != nil {
		response.BadRequest(c, "Invalid request body")
		return service.OpsAlertRule{}, false
	}
	validated, err := opsctx.ValidateAlertRulePayload(raw)
	if err != nil {
		response.BadRequestWithError(c, err)
		return service.OpsAlertRule{}, false
	}

	var rule service.OpsAlertRule
	if err := c.ShouldBindBodyWith(&rule, binding.JSON); err != nil {
		response.BadRequest(c, "Invalid request body")
		return service.OpsAlertRule{}, false
	}

	return opsAlertRuleFromValidated(id, validated), true
}

func opsAlertRuleFromValidated(id int64, validated *opsctx.AlertRuleValidatedInput) service.OpsAlertRule {
	if validated == nil {
		return service.OpsAlertRule{ID: id}
	}
	return service.OpsAlertRule{
		ID:               id,
		Name:             validated.Name,
		MetricType:       validated.MetricType,
		Operator:         validated.Operator,
		Threshold:        validated.Threshold,
		WindowMinutes:    validated.WindowMinutes,
		SustainedMinutes: validated.SustainedMinutes,
		CooldownMinutes:  validated.CooldownMinutes,
		Severity:         validated.Severity,
		Enabled:          validated.Enabled,
		NotifyEmail:      validated.NotifyEmail,
	}
}

// DeleteAlertRule deletes an ops alert rule.
// DELETE /api/v1/admin/ops/alert-rules/:id
func (h *OpsHandler) DeleteAlertRule(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	id, err := opsctx.ParsePositiveID(c.Param("id"), "Invalid rule ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := opsService.DeleteAlertRule(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

// GetAlertEvent returns a single ops alert event.
// GET /api/v1/admin/ops/alert-events/:id
func (h *OpsHandler) GetAlertEvent(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	id, err := opsctx.ParsePositiveID(c.Param("id"), "Invalid event ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ev, err := opsService.GetAlertEventByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, ev)
}

// UpdateAlertEventStatus updates an ops alert event status.
// PUT /api/v1/admin/ops/alert-events/:id/status
func (h *OpsHandler) UpdateAlertEventStatus(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	id, err := opsctx.ParsePositiveID(c.Param("id"), "Invalid event ID")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	payload, ok := bindOpsJSON[opsAlertEventStatusRequest](c)
	if !ok {
		return
	}
	status, err := opsctx.ParseAlertEventStatus(payload.Status)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resolvedAt := opsctx.AlertEventResolvedAt(status, time.Now())
	if err := opsService.UpdateAlertEventStatus(c.Request.Context(), id, status, resolvedAt); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

// CreateAlertSilence creates a scoped silence for ops alerts.
// POST /api/v1/admin/ops/alert-silences
func (h *OpsHandler) CreateAlertSilence(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	payload, ok := bindOpsJSON[opsAlertSilenceRequest](c)
	if !ok {
		return
	}
	parsed, err := opsctx.ParseAlertSilence(opsctx.AlertSilenceInput{
		RuleID:      payload.RuleID,
		PlatformRaw: payload.Platform,
		GroupID:     payload.GroupID,
		Region:      payload.Region,
		UntilRaw:    payload.Until,
		ReasonRaw:   payload.Reason,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	createdBy := (*int64)(nil)
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok {
		uid := subject.UserID
		createdBy = &uid
	}

	silence := &service.OpsAlertSilence{
		RuleID:    parsed.RuleID,
		Platform:  parsed.Platform,
		GroupID:   parsed.GroupID,
		Region:    parsed.Region,
		Until:     parsed.Until,
		Reason:    parsed.Reason,
		CreatedBy: createdBy,
	}

	created, err := opsService.CreateAlertSilence(c.Request.Context(), silence)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, created)
}

func opsAlertEventFilterFromParsed(parsed *opsctx.AlertEventFilter) *service.OpsAlertEventFilter {
	if parsed == nil {
		return &service.OpsAlertEventFilter{}
	}
	return &service.OpsAlertEventFilter{
		Limit:         parsed.Limit,
		BeforeFiredAt: parsed.BeforeFiredAt,
		BeforeID:      parsed.BeforeID,
		Status:        parsed.Status,
		Severity:      parsed.Severity,
		EmailSent:     parsed.EmailSent,
		StartTime:     parsed.StartTime,
		EndTime:       parsed.EndTime,
		Platform:      parsed.Platform,
		GroupID:       parsed.GroupID,
	}
}

// ListAlertEvents lists recent ops alert events.
// GET /api/v1/admin/ops/alert-events
func (h *OpsHandler) ListAlertEvents(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	timeRangeInput := opsTimeRangeInputFromQuery(c, "24h")
	hasTimeRange := timeRangeInput.HasExplicitTimeRange()
	var startTime, endTime time.Time
	startTime, endTime, err := parseOpsTimeRange(c, "24h")
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	parsed, err := opsctx.ParseAlertEventFilter(opsctx.AlertEventFilterInput{
		LimitRaw:         c.Query("limit"),
		StatusRaw:        c.Query("status"),
		SeverityRaw:      c.Query("severity"),
		EmailSentRaw:     c.Query("email_sent"),
		BeforeFiredAtRaw: c.Query("before_fired_at"),
		BeforeIDRaw:      c.Query("before_id"),
		PlatformRaw:      c.Query("platform"),
		GroupIDRaw:       c.Query("group_id"),
		StartTime:        startTime,
		EndTime:          endTime,
		HasTimeRange:     hasTimeRange,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	filter := opsAlertEventFilterFromParsed(parsed)

	events, err := opsService.ListAlertEvents(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, events)
}
