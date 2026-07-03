package ops

import (
	"strings"
	"testing"
)

func TestDefaultMetricThresholds(t *testing.T) {
	cfg := DefaultMetricThresholds()

	if cfg.SLAPercentMin == nil || *cfg.SLAPercentMin != 99.5 {
		t.Fatalf("expected default sla min 99.5, got %#v", cfg.SLAPercentMin)
	}
	if cfg.TTFTp99MsMax == nil || *cfg.TTFTp99MsMax != 500.0 {
		t.Fatalf("expected default ttft p99 max 500, got %#v", cfg.TTFTp99MsMax)
	}
	if cfg.RequestErrorRatePercentMax == nil || *cfg.RequestErrorRatePercentMax != 5.0 {
		t.Fatalf("expected default request error max 5, got %#v", cfg.RequestErrorRatePercentMax)
	}
	if cfg.UpstreamErrorRatePercentMax == nil || *cfg.UpstreamErrorRatePercentMax != 5.0 {
		t.Fatalf("expected default upstream error max 5, got %#v", cfg.UpstreamErrorRatePercentMax)
	}
}

func TestValidateMetricThresholds(t *testing.T) {
	validPercent := 99.0
	validTTFT := 1.0
	if err := ValidateMetricThresholds(&MetricThresholds{
		SLAPercentMin:               &validPercent,
		TTFTp99MsMax:                &validTTFT,
		RequestErrorRatePercentMax:  &validPercent,
		UpstreamErrorRatePercentMax: &validPercent,
	}); err != nil {
		t.Fatalf("expected valid thresholds, got %v", err)
	}

	if err := ValidateMetricThresholds(nil); err == nil || !strings.Contains(err.Error(), "invalid config") {
		t.Fatalf("expected nil config error, got %v", err)
	}

	negative := -1.0
	tooHigh := 100.1
	tests := []struct {
		name string
		cfg  *MetricThresholds
		want string
	}{
		{
			name: "sla percent min range",
			cfg:  &MetricThresholds{SLAPercentMin: &tooHigh},
			want: "sla_percent_min",
		},
		{
			name: "ttft nonnegative",
			cfg:  &MetricThresholds{TTFTp99MsMax: &negative},
			want: "ttft_p99_ms_max",
		},
		{
			name: "request error rate range",
			cfg:  &MetricThresholds{RequestErrorRatePercentMax: &negative},
			want: "request_error_rate_percent_max",
		},
		{
			name: "upstream error rate range",
			cfg:  &MetricThresholds{UpstreamErrorRatePercentMax: &tooHigh},
			want: "upstream_error_rate_percent_max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricThresholds(tt.cfg)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
}
