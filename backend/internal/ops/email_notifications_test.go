package ops

import (
	"strings"
	"testing"
)

func TestDefaultEmailNotificationConfig(t *testing.T) {
	cfg := DefaultEmailNotificationConfig()

	if !cfg.Alert.Enabled {
		t.Fatalf("expected alert enabled by default")
	}
	if cfg.Alert.Recipients == nil || len(cfg.Alert.Recipients) != 0 {
		t.Fatalf("expected empty alert recipients slice, got %#v", cfg.Alert.Recipients)
	}
	if cfg.Report.Recipients == nil || len(cfg.Report.Recipients) != 0 {
		t.Fatalf("expected empty report recipients slice, got %#v", cfg.Report.Recipients)
	}
	if cfg.Report.DailySummarySchedule != "0 9 * * *" {
		t.Fatalf("unexpected daily schedule %q", cfg.Report.DailySummarySchedule)
	}
	if cfg.Report.WeeklySummarySchedule != "0 9 * * 1" {
		t.Fatalf("unexpected weekly schedule %q", cfg.Report.WeeklySummarySchedule)
	}
	if cfg.Report.ErrorDigestMinCount != 10 {
		t.Fatalf("unexpected error digest min count %d", cfg.Report.ErrorDigestMinCount)
	}
	if cfg.Report.AccountHealthErrorRateThreshold != 10.0 {
		t.Fatalf("unexpected account health threshold %f", cfg.Report.AccountHealthErrorRateThreshold)
	}
}

func TestNormalizeEmailNotificationConfig(t *testing.T) {
	cfg := &EmailNotificationConfig{
		Alert: EmailAlertConfig{
			MinSeverity: " warning ",
		},
		Report: EmailReportConfig{
			DailySummarySchedule:  " ",
			WeeklySummarySchedule: " 0 10 * * 1 ",
			ErrorDigestSchedule:   "",
			AccountHealthSchedule: " 0 8 * * * ",
			ErrorDigestMinCount:   10,
			AccountHealthEnabled:  true,
			DailySummaryEnabled:   true,
			WeeklySummaryEnabled:  true,
			ErrorDigestEnabled:    true,
		},
	}

	NormalizeEmailNotificationConfig(cfg)

	if cfg.Alert.Recipients == nil {
		t.Fatalf("expected alert recipients initialized")
	}
	if cfg.Report.Recipients == nil {
		t.Fatalf("expected report recipients initialized")
	}
	if cfg.Alert.MinSeverity != "warning" {
		t.Fatalf("expected trimmed min severity, got %q", cfg.Alert.MinSeverity)
	}
	if cfg.Report.DailySummarySchedule != "0 9 * * *" {
		t.Fatalf("expected default daily schedule, got %q", cfg.Report.DailySummarySchedule)
	}
	if cfg.Report.WeeklySummarySchedule != "0 10 * * 1" {
		t.Fatalf("expected trimmed weekly schedule, got %q", cfg.Report.WeeklySummarySchedule)
	}
	if cfg.Report.ErrorDigestSchedule != "0 9 * * *" {
		t.Fatalf("expected default error digest schedule, got %q", cfg.Report.ErrorDigestSchedule)
	}
	if cfg.Report.AccountHealthSchedule != "0 8 * * *" {
		t.Fatalf("expected trimmed account health schedule, got %q", cfg.Report.AccountHealthSchedule)
	}
}

func TestValidateEmailNotificationConfig(t *testing.T) {
	if err := ValidateEmailNotificationConfig(DefaultEmailNotificationConfig()); err != nil {
		t.Fatalf("expected default config valid, got %v", err)
	}
	if err := ValidateEmailNotificationConfig(nil); err == nil || !strings.Contains(err.Error(), "invalid config") {
		t.Fatalf("expected nil config error, got %v", err)
	}

	tests := []struct {
		name string
		cfg  *EmailNotificationConfig
		want string
	}{
		{
			name: "negative rate limit",
			cfg:  &EmailNotificationConfig{Alert: EmailAlertConfig{RateLimitPerHour: -1}},
			want: "alert.rate_limit_per_hour",
		},
		{
			name: "negative batching window",
			cfg:  &EmailNotificationConfig{Alert: EmailAlertConfig{BatchingWindowSeconds: -1}},
			want: "alert.batching_window_seconds",
		},
		{
			name: "invalid min severity",
			cfg:  &EmailNotificationConfig{Alert: EmailAlertConfig{MinSeverity: "p1"}},
			want: "alert.min_severity",
		},
		{
			name: "negative error digest min count",
			cfg:  &EmailNotificationConfig{Report: EmailReportConfig{ErrorDigestMinCount: -1}},
			want: "report.error_digest_min_count",
		},
		{
			name: "account health threshold range",
			cfg:  &EmailNotificationConfig{Report: EmailReportConfig{AccountHealthErrorRateThreshold: 100.1}},
			want: "report.account_health_error_rate_threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmailNotificationConfig(tt.cfg)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
}
