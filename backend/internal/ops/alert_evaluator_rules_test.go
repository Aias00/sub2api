package ops

import (
	"testing"
	"time"
)

func TestParseAlertRuleScope(t *testing.T) {
	tests := []struct {
		name         string
		filters      map[string]any
		wantPlatform string
		wantGroupID  *int64
		wantRegion   *string
	}{
		{name: "nil filters"},
		{
			name: "trims platform region and parses int group",
			filters: map[string]any{
				"platform": " openai ",
				"group_id": 12,
				"region":   " us-east ",
			},
			wantPlatform: "openai",
			wantGroupID:  alertEvaluatorRulesInt64Ptr(12),
			wantRegion:   alertEvaluatorRulesStringPtr("us-east"),
		},
		{
			name: "parses float group",
			filters: map[string]any{
				"group_id": float64(7),
			},
			wantGroupID: alertEvaluatorRulesInt64Ptr(7),
		},
		{
			name: "parses int64 group",
			filters: map[string]any{
				"group_id": int64(9),
			},
			wantGroupID: alertEvaluatorRulesInt64Ptr(9),
		},
		{
			name: "parses string group",
			filters: map[string]any{
				"group_id": " 15 ",
			},
			wantGroupID: alertEvaluatorRulesInt64Ptr(15),
		},
		{
			name: "ignores invalid group and empty region",
			filters: map[string]any{
				"platform": 123,
				"group_id": "bad",
				"region":   " ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAlertRuleScope(tt.filters)
			if got.Platform != tt.wantPlatform {
				t.Fatalf("Platform = %q, want %q", got.Platform, tt.wantPlatform)
			}
			assertInt64PtrEqual(t, "GroupID", got.GroupID, tt.wantGroupID)
			assertStringPtrEqual(t, "Region", got.Region, tt.wantRegion)
		})
	}
}

func TestBuildAlertDimensions(t *testing.T) {
	groupID := int64(42)

	if got := BuildAlertDimensions("", nil); got != nil {
		t.Fatalf("expected nil dimensions, got %#v", got)
	}

	got := BuildAlertDimensions(" openai ", &groupID)
	if got["platform"] != "openai" {
		t.Fatalf("platform dimension = %#v", got["platform"])
	}
	if got["group_id"] != groupID {
		t.Fatalf("group_id dimension = %#v", got["group_id"])
	}
}

func TestBuildAlertDescription(t *testing.T) {
	groupID := int64(42)
	rule := AlertDescriptionRule{
		MetricType: " error_rate ",
		Operator:   ">=",
		Threshold:  5,
	}

	got := BuildAlertDescription(rule, 6.789, 0, " openai ", &groupID)
	want := "error_rate >= 5.00 (current 6.79) over last 1m (platform=openai group_id=42)"
	if got != want {
		t.Fatalf("description = %q, want %q", got, want)
	}

	got = BuildAlertDescription(rule, 6, 5, "", nil)
	want = "error_rate >= 5.00 (current 6.00) over last 5m (overall)"
	if got != want {
		t.Fatalf("description = %q, want %q", got, want)
	}
}

func TestRequiredSustainedBreaches(t *testing.T) {
	tests := []struct {
		name             string
		sustainedMinutes int
		interval         time.Duration
		want             int
	}{
		{name: "nonpositive sustained defaults to one", sustainedMinutes: 0, interval: time.Minute, want: 1},
		{name: "nonpositive interval uses sustained minutes", sustainedMinutes: 3, interval: 0, want: 3},
		{name: "one minute at thirty seconds needs two", sustainedMinutes: 1, interval: 30 * time.Second, want: 2},
		{name: "rounds up partial interval", sustainedMinutes: 2, interval: 45 * time.Second, want: 3},
		{name: "minimum one for long interval", sustainedMinutes: 1, interval: 2 * time.Minute, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequiredSustainedBreaches(tt.sustainedMinutes, tt.interval)
			if got != tt.want {
				t.Fatalf("RequiredSustainedBreaches() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompareAlertMetric(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		operator  string
		threshold float64
		want      bool
	}{
		{name: "greater", value: 2, operator: ">", threshold: 1, want: true},
		{name: "greater false on equal", value: 1, operator: ">", threshold: 1, want: false},
		{name: "greater equal", value: 1, operator: ">=", threshold: 1, want: true},
		{name: "less", value: 0, operator: "<", threshold: 1, want: true},
		{name: "less false on equal", value: 1, operator: "<", threshold: 1, want: false},
		{name: "less equal", value: 1, operator: "<=", threshold: 1, want: true},
		{name: "equal", value: 1, operator: "==", threshold: 1, want: true},
		{name: "not equal", value: 2, operator: "!=", threshold: 1, want: true},
		{name: "unknown operator", value: 2, operator: "~", threshold: 1, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareAlertMetric(tt.value, tt.operator, tt.threshold)
			if got != tt.want {
				t.Fatalf("CompareAlertMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func alertEvaluatorRulesInt64Ptr(v int64) *int64 {
	return &v
}

func alertEvaluatorRulesStringPtr(v string) *string {
	return &v
}

func assertInt64PtrEqual(t *testing.T, label string, got *int64, want *int64) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("%s = %#v, want %#v", label, got, want)
		}
		return
	}
	if *got != *want {
		t.Fatalf("%s = %d, want %d", label, *got, *want)
	}
}

func assertStringPtrEqual(t *testing.T, label string, got *string, want *string) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("%s = %#v, want %#v", label, got, want)
		}
		return
	}
	if *got != *want {
		t.Fatalf("%s = %q, want %q", label, *got, *want)
	}
}
