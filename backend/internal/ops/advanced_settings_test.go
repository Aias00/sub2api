package ops

import (
	"strings"
	"testing"
)

func TestDefaultAdvancedSettings(t *testing.T) {
	cfg := DefaultAdvancedSettings("0 2 * * *")

	if cfg.DataRetention.CleanupSchedule != "0 2 * * *" {
		t.Fatalf("unexpected cleanup schedule %q", cfg.DataRetention.CleanupSchedule)
	}
	if cfg.DataRetention.ErrorLogRetentionDays != 30 ||
		cfg.DataRetention.MinuteMetricsRetentionDays != 30 ||
		cfg.DataRetention.HourlyMetricsRetentionDays != 30 {
		t.Fatalf("unexpected retention defaults: %+v", cfg.DataRetention)
	}
	if !cfg.IgnoreCountTokensErrors {
		t.Fatalf("expected count_tokens errors ignored by default")
	}
	if !cfg.IgnoreContextCanceled {
		t.Fatalf("expected context canceled ignored by default")
	}
	if cfg.IgnoreNoAvailableAccounts {
		t.Fatalf("expected no available accounts not ignored by default")
	}
	if cfg.DisplayOpenAITokenStats {
		t.Fatalf("expected OpenAI token stats hidden by default")
	}
	if !cfg.DisplayAlertEvents {
		t.Fatalf("expected alert events shown by default")
	}
	if cfg.AutoRefreshIntervalSec != 30 {
		t.Fatalf("expected auto refresh interval 30, got %d", cfg.AutoRefreshIntervalSec)
	}
}

func TestNormalizeAdvancedSettings(t *testing.T) {
	cfg := &AdvancedSettings{
		DataRetention: DataRetentionSettings{
			CleanupSchedule:            " ",
			ErrorLogRetentionDays:      -1,
			MinuteMetricsRetentionDays: -2,
			HourlyMetricsRetentionDays: -3,
		},
		OpenAIAccountQuotaAutoPause: OpenAIAccountQuotaAutoPauseSettings{
			DefaultThreshold5h: -0.5,
			DefaultThreshold7d: 1.5,
		},
		AutoRefreshIntervalSec: 0,
	}

	NormalizeAdvancedSettings(cfg, "0 3 * * *")

	if cfg.DataRetention.CleanupSchedule != "0 3 * * *" {
		t.Fatalf("expected default cleanup schedule, got %q", cfg.DataRetention.CleanupSchedule)
	}
	if cfg.DataRetention.ErrorLogRetentionDays != 30 ||
		cfg.DataRetention.MinuteMetricsRetentionDays != 30 ||
		cfg.DataRetention.HourlyMetricsRetentionDays != 30 {
		t.Fatalf("expected negative retention backfilled to 30, got %+v", cfg.DataRetention)
	}
	if cfg.OpenAIAccountQuotaAutoPause.DefaultThreshold5h != 0 {
		t.Fatalf("expected 5h threshold clamped to 0, got %f", cfg.OpenAIAccountQuotaAutoPause.DefaultThreshold5h)
	}
	if cfg.OpenAIAccountQuotaAutoPause.DefaultThreshold7d != 1 {
		t.Fatalf("expected 7d threshold clamped to 1, got %f", cfg.OpenAIAccountQuotaAutoPause.DefaultThreshold7d)
	}
	if cfg.AutoRefreshIntervalSec != 30 {
		t.Fatalf("expected auto refresh interval backfilled to 30, got %d", cfg.AutoRefreshIntervalSec)
	}
}

func TestNormalizeAdvancedSettingsTrimsCleanupScheduleAndKeepsZeroRetention(t *testing.T) {
	cfg := &AdvancedSettings{
		DataRetention: DataRetentionSettings{
			CleanupSchedule:            " 0 4 * * * ",
			ErrorLogRetentionDays:      0,
			MinuteMetricsRetentionDays: 0,
			HourlyMetricsRetentionDays: 0,
		},
		AutoRefreshIntervalSec: 15,
	}

	NormalizeAdvancedSettings(cfg, "0 3 * * *")

	if cfg.DataRetention.CleanupSchedule != "0 4 * * *" {
		t.Fatalf("expected trimmed cleanup schedule, got %q", cfg.DataRetention.CleanupSchedule)
	}
	if cfg.DataRetention.ErrorLogRetentionDays != 0 ||
		cfg.DataRetention.MinuteMetricsRetentionDays != 0 ||
		cfg.DataRetention.HourlyMetricsRetentionDays != 0 {
		t.Fatalf("expected zero retention preserved, got %+v", cfg.DataRetention)
	}
}

func TestClampQuotaAutoPauseThreshold(t *testing.T) {
	tests := []struct {
		value float64
		want  float64
	}{
		{value: -1, want: 0},
		{value: 0, want: 0},
		{value: 0.75, want: 0.75},
		{value: 2, want: 1},
	}

	for _, tt := range tests {
		got := ClampQuotaAutoPauseThreshold(tt.value)
		if got != tt.want {
			t.Fatalf("ClampQuotaAutoPauseThreshold(%v) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestValidateAdvancedSettings(t *testing.T) {
	valid := DefaultAdvancedSettings("0 2 * * *")
	if err := ValidateAdvancedSettings(valid); err != nil {
		t.Fatalf("expected default advanced settings valid, got %v", err)
	}
	if err := ValidateAdvancedSettings(nil); err == nil || !strings.Contains(err.Error(), "invalid config") {
		t.Fatalf("expected nil config error, got %v", err)
	}

	tests := []struct {
		name string
		edit func(*AdvancedSettings)
		want string
	}{
		{
			name: "error log retention range",
			edit: func(cfg *AdvancedSettings) {
				cfg.DataRetention.ErrorLogRetentionDays = 366
			},
			want: "error_log_retention_days",
		},
		{
			name: "minute metrics retention range",
			edit: func(cfg *AdvancedSettings) {
				cfg.DataRetention.MinuteMetricsRetentionDays = -1
			},
			want: "minute_metrics_retention_days",
		},
		{
			name: "hourly metrics retention range",
			edit: func(cfg *AdvancedSettings) {
				cfg.DataRetention.HourlyMetricsRetentionDays = 366
			},
			want: "hourly_metrics_retention_days",
		},
		{
			name: "auto refresh interval range",
			edit: func(cfg *AdvancedSettings) {
				cfg.AutoRefreshIntervalSec = 301
			},
			want: "auto_refresh_interval_seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultAdvancedSettings("0 2 * * *")
			tt.edit(cfg)
			err := ValidateAdvancedSettings(cfg)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
}
