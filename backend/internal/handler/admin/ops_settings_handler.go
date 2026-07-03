package admin

import (
	"net/http"

	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
)

// GetEmailNotificationConfig returns Ops email notification config (DB-backed).
// GET /api/v1/admin/ops/email-notification/config
func (h *OpsHandler) GetEmailNotificationConfig(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	cfg, err := opsService.GetEmailNotificationConfig(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get email notification config")
		return
	}
	response.Success(c, cfg)
}

// UpdateEmailNotificationConfig updates Ops email notification config (DB-backed).
// PUT /api/v1/admin/ops/email-notification/config
func (h *OpsHandler) UpdateEmailNotificationConfig(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	req, ok := bindOpsJSON[service.OpsEmailNotificationConfigUpdateRequest](c)
	if !ok {
		return
	}

	updated, err := opsService.UpdateEmailNotificationConfig(c.Request.Context(), req)
	if err != nil {
		// Most failures here are validation errors from request payload; treat as 400.
		response.BadRequestWithError(c, err)
		return
	}
	response.Success(c, updated)
}

// GetAlertRuntimeSettings returns Ops alert evaluator runtime settings (DB-backed).
// GET /api/v1/admin/ops/runtime/alert
func (h *OpsHandler) GetAlertRuntimeSettings(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	cfg, err := opsService.GetOpsAlertRuntimeSettings(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get alert runtime settings")
		return
	}
	response.Success(c, cfg)
}

// UpdateAlertRuntimeSettings updates Ops alert evaluator runtime settings (DB-backed).
// PUT /api/v1/admin/ops/runtime/alert
func (h *OpsHandler) UpdateAlertRuntimeSettings(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	req, ok := bindOpsJSON[service.OpsAlertRuntimeSettings](c)
	if !ok {
		return
	}

	updated, err := opsService.UpdateOpsAlertRuntimeSettings(c.Request.Context(), req)
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	response.Success(c, updated)
}

// GetRuntimeLogConfig returns runtime log config (DB-backed).
// GET /api/v1/admin/ops/runtime/logging
func (h *OpsHandler) GetRuntimeLogConfig(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	cfg, err := opsService.GetRuntimeLogConfig(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get runtime log config")
		return
	}
	response.Success(c, cfg)
}

// UpdateRuntimeLogConfig updates runtime log config and applies changes immediately.
// PUT /api/v1/admin/ops/runtime/logging
func (h *OpsHandler) UpdateRuntimeLogConfig(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	req, ok := bindOpsJSON[service.OpsRuntimeLogConfig](c)
	if !ok {
		return
	}

	uid, ok := requireOpsUserID(c)
	if !ok {
		return
	}

	updated, err := opsService.UpdateRuntimeLogConfig(c.Request.Context(), req, uid)
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	response.Success(c, updated)
}

// ResetRuntimeLogConfig removes runtime override and falls back to env/yaml baseline.
// POST /api/v1/admin/ops/runtime/logging/reset
func (h *OpsHandler) ResetRuntimeLogConfig(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	uid, ok := requireOpsUserID(c)
	if !ok {
		return
	}

	updated, err := opsService.ResetRuntimeLogConfig(c.Request.Context(), uid)
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	response.Success(c, updated)
}

// GetAdvancedSettings returns Ops advanced settings (DB-backed).
// GET /api/v1/admin/ops/advanced-settings
func (h *OpsHandler) GetAdvancedSettings(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	cfg, err := opsService.GetOpsAdvancedSettings(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get advanced settings")
		return
	}
	response.Success(c, cfg)
}

// UpdateAdvancedSettings updates Ops advanced settings (DB-backed).
// PUT /api/v1/admin/ops/advanced-settings
func (h *OpsHandler) UpdateAdvancedSettings(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	req, ok := bindOpsJSON[service.OpsAdvancedSettings](c)
	if !ok {
		return
	}

	updated, err := opsService.UpdateOpsAdvancedSettings(c.Request.Context(), req)
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	response.Success(c, updated)
}

// GetMetricThresholds returns Ops metric thresholds (DB-backed).
// GET /api/v1/admin/ops/settings/metric-thresholds
func (h *OpsHandler) GetMetricThresholds(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	cfg, err := opsService.GetMetricThresholds(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get metric thresholds")
		return
	}
	response.Success(c, cfg)
}

// UpdateMetricThresholds updates Ops metric thresholds (DB-backed).
// PUT /api/v1/admin/ops/settings/metric-thresholds
func (h *OpsHandler) UpdateMetricThresholds(c *gin.Context) {
	opsService, ok := h.requireOpsService(c)
	if !ok {
		return
	}

	req, ok := bindOpsJSON[service.OpsMetricThresholds](c)
	if !ok {
		return
	}

	updated, err := opsService.UpdateMetricThresholds(c.Request.Context(), req)
	if err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	response.Success(c, updated)
}
