package ops

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateAlertRulePayload(t *testing.T) {
	t.Parallel()

	raw := map[string]json.RawMessage{
		"name":        json.RawMessage(`" High error rate "`),
		"metric_type": json.RawMessage(`"error_rate"`),
		"operator":    json.RawMessage(`">"`),
		"threshold":   json.RawMessage(`90`),
	}

	validated, err := ValidateAlertRulePayload(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if validated.Name != "High error rate" ||
		validated.Severity != "P2" ||
		validated.Enabled != true ||
		validated.NotifyEmail != true ||
		validated.WindowMinutes != 1 ||
		validated.SustainedMinutes != 1 ||
		validated.CooldownMinutes != 0 {
		t.Fatalf("validated = %+v", validated)
	}
}

func TestValidateAlertRulePayloadOptionalFields(t *testing.T) {
	t.Parallel()

	raw := map[string]json.RawMessage{
		"name":              json.RawMessage(`"CPU high"`),
		"metric_type":       json.RawMessage(`"cpu_usage_percent"`),
		"operator":          json.RawMessage(`">="`),
		"threshold":         json.RawMessage(`95`),
		"severity":          json.RawMessage(`"p1"`),
		"enabled":           json.RawMessage(`false`),
		"notify_email":      json.RawMessage(`false`),
		"window_minutes":    json.RawMessage(`5`),
		"sustained_minutes": json.RawMessage(`10`),
		"cooldown_minutes":  json.RawMessage(`30`),
	}

	validated, err := ValidateAlertRulePayload(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if validated.Severity != "P1" ||
		validated.Enabled ||
		validated.NotifyEmail ||
		validated.WindowMinutes != 5 ||
		validated.SustainedMinutes != 10 ||
		validated.CooldownMinutes != 30 {
		t.Fatalf("validated = %+v", validated)
	}
}

func TestValidateAlertRulePayloadErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  map[string]json.RawMessage
		want string
	}{
		{name: "missing required", raw: map[string]json.RawMessage{}, want: "name is required"},
		{
			name: "bad operator",
			raw: map[string]json.RawMessage{
				"name":        json.RawMessage(`"x"`),
				"metric_type": json.RawMessage(`"error_rate"`),
				"operator":    json.RawMessage(`"between"`),
				"threshold":   json.RawMessage(`1`),
			},
			want: "operator must be one of",
		},
		{
			name: "percent out of range",
			raw: map[string]json.RawMessage{
				"name":        json.RawMessage(`"x"`),
				"metric_type": json.RawMessage(`"error_rate"`),
				"operator":    json.RawMessage(`">"`),
				"threshold":   json.RawMessage(`101`),
			},
			want: "between 0 and 100",
		},
		{
			name: "negative count",
			raw: map[string]json.RawMessage{
				"name":        json.RawMessage(`"x"`),
				"metric_type": json.RawMessage(`"concurrency_queue_depth"`),
				"operator":    json.RawMessage(`">"`),
				"threshold":   json.RawMessage(`-1`),
			},
			want: "threshold must be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ValidateAlertRulePayload(tt.raw)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestIsPercentOrRateMetric(t *testing.T) {
	t.Parallel()

	if !IsPercentOrRateMetric("error_rate") {
		t.Fatal("error_rate should be percent/rate metric")
	}
	if IsPercentOrRateMetric("concurrency_queue_depth") {
		t.Fatal("concurrency_queue_depth should not be percent/rate metric")
	}
}
