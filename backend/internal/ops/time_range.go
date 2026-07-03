package ops

import (
	"fmt"
	"strings"
	"time"
)

type RealtimeWindow struct {
	Duration time.Duration
	Label    string
}

type RealtimeWindowRange struct {
	Duration  time.Duration
	Label     string
	StartTime time.Time
	EndTime   time.Time
}

type TimeRangeInput struct {
	StartTimeRaw string
	EndTimeRaw   string
	TimeRangeRaw string
	DefaultRange string
	Now          time.Time
}

func (input TimeRangeInput) HasExplicitTimeRange() bool {
	return strings.TrimSpace(input.StartTimeRaw) != "" ||
		strings.TrimSpace(input.EndTimeRaw) != "" ||
		strings.TrimSpace(input.TimeRangeRaw) != ""
}

func ParseDuration(raw string) (time.Duration, bool) {
	switch strings.TrimSpace(raw) {
	case "5m":
		return 5 * time.Minute, true
	case "30m":
		return 30 * time.Minute, true
	case "1h":
		return time.Hour, true
	case "6h":
		return 6 * time.Hour, true
	case "24h":
		return 24 * time.Hour, true
	case "7d":
		return 7 * 24 * time.Hour, true
	case "30d":
		return 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func ParseOpenAITokenStatsDuration(raw string) (time.Duration, bool) {
	switch strings.TrimSpace(raw) {
	case "30m":
		return 30 * time.Minute, true
	case "1h":
		return time.Hour, true
	case "1d":
		return 24 * time.Hour, true
	case "15d":
		return 15 * 24 * time.Hour, true
	case "30d":
		return 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func ParseRealtimeWindow(raw string) (RealtimeWindow, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "1min", "1m":
		return RealtimeWindow{Duration: time.Minute, Label: "1min"}, true
	case "5min", "5m":
		return RealtimeWindow{Duration: 5 * time.Minute, Label: "5min"}, true
	case "30min", "30m":
		return RealtimeWindow{Duration: 30 * time.Minute, Label: "30min"}, true
	case "1h", "60m", "60min":
		return RealtimeWindow{Duration: time.Hour, Label: "1h"}, true
	default:
		return RealtimeWindow{}, false
	}
}

func ParseRealtimeWindowRange(raw string, now time.Time) (RealtimeWindowRange, bool) {
	window, ok := ParseRealtimeWindow(raw)
	if !ok {
		return RealtimeWindowRange{}, false
	}
	if now.IsZero() {
		now = time.Now()
	}
	endTime := now.UTC()
	return RealtimeWindowRange{
		Duration:  window.Duration,
		Label:     window.Label,
		StartTime: endTime.Add(-window.Duration),
		EndTime:   endTime,
	}, true
}

func PickThroughputBucketSeconds(window time.Duration) int {
	switch {
	case window <= 2*time.Hour:
		return 60
	case window <= 24*time.Hour:
		return 300
	default:
		return 3600
	}
}

func ParseTimeRange(input TimeRangeInput) (time.Time, time.Time, error) {
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}

	startStr := strings.TrimSpace(input.StartTimeRaw)
	endStr := strings.TrimSpace(input.EndTimeRaw)

	start, err := parseTimestamp(startStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseTimestamp(endStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if startStr != "" || endStr != "" {
		if end.IsZero() {
			end = now
		}
		if start.IsZero() {
			dur, _ := ParseDuration(input.DefaultRange)
			start = end.Add(-dur)
		}
		return validateTimeRange(start, end)
	}

	rangeRaw := strings.TrimSpace(input.TimeRangeRaw)
	if rangeRaw == "" {
		rangeRaw = input.DefaultRange
	}
	dur, ok := ParseDuration(rangeRaw)
	if !ok {
		dur, _ = ParseDuration(input.DefaultRange)
	}

	end = now
	start = end.Add(-dur)
	return validateTimeRange(start, end)
}

func parseTimestamp(raw string) (time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, raw)
}

func validateTimeRange(start time.Time, end time.Time) (time.Time, time.Time, error) {
	if start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range: start_time must be <= end_time")
	}
	if end.Sub(start) > 30*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range: max window is 30 days")
	}
	return start, end, nil
}
