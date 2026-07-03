package service

import opsctx "github.com/Aias00/cloudbase/internal/ops"

// Ops settings models stored in DB `settings` table (JSON blobs).

type OpsEmailNotificationConfig = opsctx.EmailNotificationConfig
type OpsEmailAlertConfig = opsctx.EmailAlertConfig
type OpsEmailReportConfig = opsctx.EmailReportConfig

// OpsEmailNotificationConfigUpdateRequest allows partial updates, while the
// frontend can still send the full config shape.
type OpsEmailNotificationConfigUpdateRequest struct {
	Alert  *OpsEmailAlertConfig  `json:"alert"`
	Report *OpsEmailReportConfig `json:"report"`
}

type OpsDistributedLockSettings = opsctx.DistributedLockSettings
type OpsAlertSilenceEntry = opsctx.AlertSilenceEntry
type OpsAlertSilencingSettings = opsctx.AlertSilencingSettings
type OpsMetricThresholds = opsctx.MetricThresholds
type OpsRuntimeLogConfig = opsctx.RuntimeLogConfig
type OpsAlertRuntimeSettings = opsctx.AlertRuntimeSettings

type OpsAdvancedSettings = opsctx.AdvancedSettings
type OpsOpenAIAccountQuotaAutoPauseSettings = opsctx.OpenAIAccountQuotaAutoPauseSettings
type OpsDataRetentionSettings = opsctx.DataRetentionSettings
type OpsAggregationSettings = opsctx.AggregationSettings
