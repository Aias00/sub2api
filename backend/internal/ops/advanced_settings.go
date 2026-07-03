package ops

import (
	"errors"
	"strings"
)

type AdvancedSettings struct {
	DataRetention                   DataRetentionSettings               `json:"data_retention"`
	Aggregation                     AggregationSettings                 `json:"aggregation"`
	OpenAIAccountQuotaAutoPause     OpenAIAccountQuotaAutoPauseSettings `json:"openai_account_quota_auto_pause"`
	IgnoreCountTokensErrors         bool                                `json:"ignore_count_tokens_errors"`
	IgnoreContextCanceled           bool                                `json:"ignore_context_canceled"`
	IgnoreNoAvailableAccounts       bool                                `json:"ignore_no_available_accounts"`
	IgnoreInvalidApiKeyErrors       bool                                `json:"ignore_invalid_api_key_errors"`
	IgnoreInsufficientBalanceErrors bool                                `json:"ignore_insufficient_balance_errors"`
	DisplayOpenAITokenStats         bool                                `json:"display_openai_token_stats"`
	DisplayAlertEvents              bool                                `json:"display_alert_events"`
	AutoRefreshEnabled              bool                                `json:"auto_refresh_enabled"`
	AutoRefreshIntervalSec          int                                 `json:"auto_refresh_interval_seconds"`
}

type OpenAIAccountQuotaAutoPauseSettings struct {
	DefaultThreshold5h float64 `json:"default_threshold_5h"`
	DefaultThreshold7d float64 `json:"default_threshold_7d"`
}

type DataRetentionSettings struct {
	CleanupEnabled             bool   `json:"cleanup_enabled"`
	CleanupSchedule            string `json:"cleanup_schedule"`
	ErrorLogRetentionDays      int    `json:"error_log_retention_days"`
	MinuteMetricsRetentionDays int    `json:"minute_metrics_retention_days"`
	HourlyMetricsRetentionDays int    `json:"hourly_metrics_retention_days"`
}

type AggregationSettings struct {
	AggregationEnabled bool `json:"aggregation_enabled"`
}

func DefaultAdvancedSettings(defaultCleanupSchedule string) *AdvancedSettings {
	return &AdvancedSettings{
		DataRetention: DataRetentionSettings{
			CleanupEnabled:             false,
			CleanupSchedule:            defaultCleanupSchedule,
			ErrorLogRetentionDays:      30,
			MinuteMetricsRetentionDays: 30,
			HourlyMetricsRetentionDays: 30,
		},
		Aggregation: AggregationSettings{
			AggregationEnabled: false,
		},
		OpenAIAccountQuotaAutoPause:     OpenAIAccountQuotaAutoPauseSettings{},
		IgnoreCountTokensErrors:         true,
		IgnoreContextCanceled:           true,
		IgnoreNoAvailableAccounts:       false,
		IgnoreInsufficientBalanceErrors: false,
		DisplayOpenAITokenStats:         false,
		DisplayAlertEvents:              true,
		AutoRefreshEnabled:              false,
		AutoRefreshIntervalSec:          30,
	}
}

func NormalizeAdvancedSettings(cfg *AdvancedSettings, defaultCleanupSchedule string) {
	if cfg == nil {
		return
	}
	cfg.OpenAIAccountQuotaAutoPause.DefaultThreshold5h = ClampQuotaAutoPauseThreshold(cfg.OpenAIAccountQuotaAutoPause.DefaultThreshold5h)
	cfg.OpenAIAccountQuotaAutoPause.DefaultThreshold7d = ClampQuotaAutoPauseThreshold(cfg.OpenAIAccountQuotaAutoPause.DefaultThreshold7d)
	cfg.DataRetention.CleanupSchedule = strings.TrimSpace(cfg.DataRetention.CleanupSchedule)
	if cfg.DataRetention.CleanupSchedule == "" {
		cfg.DataRetention.CleanupSchedule = defaultCleanupSchedule
	}
	if cfg.DataRetention.ErrorLogRetentionDays < 0 {
		cfg.DataRetention.ErrorLogRetentionDays = 30
	}
	if cfg.DataRetention.MinuteMetricsRetentionDays < 0 {
		cfg.DataRetention.MinuteMetricsRetentionDays = 30
	}
	if cfg.DataRetention.HourlyMetricsRetentionDays < 0 {
		cfg.DataRetention.HourlyMetricsRetentionDays = 30
	}
	if cfg.AutoRefreshIntervalSec <= 0 {
		cfg.AutoRefreshIntervalSec = 30
	}
}

func ClampQuotaAutoPauseThreshold(value float64) float64 {
	if value <= 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func ValidateAdvancedSettings(cfg *AdvancedSettings) error {
	if cfg == nil {
		return errors.New("invalid config")
	}
	if cfg.DataRetention.ErrorLogRetentionDays < 0 || cfg.DataRetention.ErrorLogRetentionDays > 365 {
		return errors.New("error_log_retention_days must be between 0 and 365")
	}
	if cfg.DataRetention.MinuteMetricsRetentionDays < 0 || cfg.DataRetention.MinuteMetricsRetentionDays > 365 {
		return errors.New("minute_metrics_retention_days must be between 0 and 365")
	}
	if cfg.DataRetention.HourlyMetricsRetentionDays < 0 || cfg.DataRetention.HourlyMetricsRetentionDays > 365 {
		return errors.New("hourly_metrics_retention_days must be between 0 and 365")
	}
	if cfg.AutoRefreshIntervalSec < 15 || cfg.AutoRefreshIntervalSec > 300 {
		return errors.New("auto_refresh_interval_seconds must be between 15 and 300")
	}
	return nil
}
