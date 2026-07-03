package ops

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
		ok    bool
	}{
		{input: "5m", want: 5 * time.Minute, ok: true},
		{input: "30m", want: 30 * time.Minute, ok: true},
		{input: "1h", want: time.Hour, ok: true},
		{input: "6h", want: 6 * time.Hour, ok: true},
		{input: "24h", want: 24 * time.Hour, ok: true},
		{input: "7d", want: 7 * 24 * time.Hour, ok: true},
		{input: "30d", want: 30 * 24 * time.Hour, ok: true},
		{input: " 1h ", want: time.Hour, ok: true},
		{input: "invalid", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseDuration(tt.input)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("duration = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOpenAITokenStatsDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
		ok    bool
	}{
		{input: "30m", want: 30 * time.Minute, ok: true},
		{input: "1h", want: time.Hour, ok: true},
		{input: "1d", want: 24 * time.Hour, ok: true},
		{input: "15d", want: 15 * 24 * time.Hour, ok: true},
		{input: "30d", want: 30 * 24 * time.Hour, ok: true},
		{input: " 1h ", want: time.Hour, ok: true},
		{input: "7d", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseOpenAITokenStatsDuration(tt.input)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("duration = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRealtimeWindow(t *testing.T) {
	tests := []struct {
		input string
		want  RealtimeWindow
		ok    bool
	}{
		{input: "", want: RealtimeWindow{Duration: time.Minute, Label: "1min"}, ok: true},
		{input: "1min", want: RealtimeWindow{Duration: time.Minute, Label: "1min"}, ok: true},
		{input: "1m", want: RealtimeWindow{Duration: time.Minute, Label: "1min"}, ok: true},
		{input: "5m", want: RealtimeWindow{Duration: 5 * time.Minute, Label: "5min"}, ok: true},
		{input: "30MIN", want: RealtimeWindow{Duration: 30 * time.Minute, Label: "30min"}, ok: true},
		{input: "60m", want: RealtimeWindow{Duration: time.Hour, Label: "1h"}, ok: true},
		{input: " invalid ", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseRealtimeWindow(tt.input)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("window = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseRealtimeWindowRange(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("UTC+8", 8*60*60))
	got, ok := ParseRealtimeWindowRange("5m", now)
	if !ok {
		t.Fatal("expected valid realtime window range")
	}
	if got.Duration != 5*time.Minute {
		t.Fatalf("duration = %v, want 5m", got.Duration)
	}
	if got.Label != "5min" {
		t.Fatalf("label = %q, want 5min", got.Label)
	}
	if got.EndTime.Location() != time.UTC {
		t.Fatalf("end location = %v, want UTC", got.EndTime.Location())
	}
	if got.EndTime.Sub(got.StartTime) != 5*time.Minute {
		t.Fatalf("range duration = %v, want 5m", got.EndTime.Sub(got.StartTime))
	}

	if _, ok := ParseRealtimeWindowRange("bad", now); ok {
		t.Fatal("expected invalid realtime window range")
	}
}

func TestPickThroughputBucketSeconds(t *testing.T) {
	tests := []struct {
		window time.Duration
		want   int
	}{
		{window: 30 * time.Minute, want: 60},
		{window: 2 * time.Hour, want: 60},
		{window: 6 * time.Hour, want: 300},
		{window: 24 * time.Hour, want: 300},
		{window: 48 * time.Hour, want: 3600},
	}

	for _, tt := range tests {
		t.Run(tt.window.String(), func(t *testing.T) {
			got := PickThroughputBucketSeconds(tt.window)
			if got != tt.want {
				t.Fatalf("bucket seconds = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseTimeRangeExplicitStartEnd(t *testing.T) {
	startRaw := "2026-01-02T03:04:05Z"
	endRaw := "2026-01-02T04:04:05Z"

	start, end, err := ParseTimeRange(TimeRangeInput{
		StartTimeRaw: startRaw,
		EndTimeRaw:   endRaw,
		DefaultRange: "1h",
		Now:          mustParseTimeRangeTime(t, "2026-01-02T05:04:05Z"),
	})
	if err != nil {
		t.Fatalf("ParseTimeRange() error = %v", err)
	}
	if start.Format(time.RFC3339) != startRaw {
		t.Fatalf("start = %s, want %s", start.Format(time.RFC3339), startRaw)
	}
	if end.Format(time.RFC3339) != endRaw {
		t.Fatalf("end = %s, want %s", end.Format(time.RFC3339), endRaw)
	}
}

func TestTimeRangeInputHasExplicitTimeRange(t *testing.T) {
	tests := []struct {
		name  string
		input TimeRangeInput
		want  bool
	}{
		{name: "blank", input: TimeRangeInput{}, want: false},
		{name: "start", input: TimeRangeInput{StartTimeRaw: " 2026-01-02T03:04:05Z "}, want: true},
		{name: "end", input: TimeRangeInput{EndTimeRaw: "2026-01-02T04:04:05Z"}, want: true},
		{name: "range", input: TimeRangeInput{TimeRangeRaw: " 1h "}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.HasExplicitTimeRange(); got != tt.want {
				t.Fatalf("HasExplicitTimeRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimeRangePartialExplicitUsesDefaultRange(t *testing.T) {
	now := mustParseTimeRangeTime(t, "2026-01-02T05:04:05Z")

	start, end, err := ParseTimeRange(TimeRangeInput{
		EndTimeRaw:   "2026-01-02T04:04:05Z",
		DefaultRange: "1h",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("ParseTimeRange() error = %v", err)
	}
	if end.Format(time.RFC3339) != "2026-01-02T04:04:05Z" {
		t.Fatalf("end = %s", end.Format(time.RFC3339))
	}
	if end.Sub(start) != time.Hour {
		t.Fatalf("duration = %v, want 1h", end.Sub(start))
	}

	start, end, err = ParseTimeRange(TimeRangeInput{
		StartTimeRaw: "2026-01-02T03:04:05Z",
		DefaultRange: "1h",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("ParseTimeRange() error = %v", err)
	}
	if end != now {
		t.Fatalf("end = %s, want now %s", end, now)
	}
	if start.Format(time.RFC3339) != "2026-01-02T03:04:05Z" {
		t.Fatalf("start = %s", start.Format(time.RFC3339))
	}
}

func TestParseTimeRangeTimeRangeFallback(t *testing.T) {
	now := mustParseTimeRangeTime(t, "2026-01-02T05:04:05Z")

	start, end, err := ParseTimeRange(TimeRangeInput{
		TimeRangeRaw: "6h",
		DefaultRange: "1h",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("ParseTimeRange() error = %v", err)
	}
	if end != now {
		t.Fatalf("end = %s, want now %s", end, now)
	}
	if end.Sub(start) != 6*time.Hour {
		t.Fatalf("duration = %v, want 6h", end.Sub(start))
	}

	start, end, err = ParseTimeRange(TimeRangeInput{
		TimeRangeRaw: "bad",
		DefaultRange: "1h",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("ParseTimeRange() fallback error = %v", err)
	}
	if end.Sub(start) != time.Hour {
		t.Fatalf("fallback duration = %v, want 1h", end.Sub(start))
	}
}

func TestParseTimeRangeErrors(t *testing.T) {
	now := mustParseTimeRangeTime(t, "2026-01-02T05:04:05Z")

	if _, _, err := ParseTimeRange(TimeRangeInput{
		StartTimeRaw: "bad",
		DefaultRange: "1h",
		Now:          now,
	}); err == nil {
		t.Fatalf("expected invalid start time error")
	}

	if _, _, err := ParseTimeRange(TimeRangeInput{
		StartTimeRaw: "2026-01-02T06:04:05Z",
		EndTimeRaw:   "2026-01-02T05:04:05Z",
		DefaultRange: "1h",
		Now:          now,
	}); err == nil {
		t.Fatalf("expected start after end error")
	}

	if _, _, err := ParseTimeRange(TimeRangeInput{
		StartTimeRaw: "2026-01-01T00:00:00Z",
		EndTimeRaw:   "2026-02-02T00:00:00Z",
		DefaultRange: "1h",
		Now:          now,
	}); err == nil {
		t.Fatalf("expected max window error")
	}
}

func mustParseTimeRangeTime(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed
}
