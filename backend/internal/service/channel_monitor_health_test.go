//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestClassifyMonitorError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		statusCode int
		body       string
		want       string
	}{
		{name: "timeout", err: context.DeadlineExceeded, want: MonitorErrorCategoryTimeout},
		{name: "network", err: errors.New("connection refused"), want: MonitorErrorCategoryNetwork},
		{name: "auth", statusCode: 401, body: `{"error":"invalid api key"}`, want: MonitorErrorCategoryAuth},
		{name: "rate limit", statusCode: 429, body: `{"error":"rate limit exceeded"}`, want: MonitorErrorCategoryRateLimit},
		{name: "quota", statusCode: 400, body: `{"error":"insufficient credit balance"}`, want: MonitorErrorCategoryQuota},
		{name: "server", statusCode: 503, body: `temporarily unavailable`, want: MonitorErrorCategoryServer},
		{name: "invalid request", statusCode: 404, body: `not found`, want: MonitorErrorCategoryInvalid},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyMonitorError(tt.err, tt.statusCode, tt.body); got != tt.want {
				t.Fatalf("classifyMonitorError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildChannelMonitorHealthSnapshot_UnhealthyAfterConsecutiveFailedRuns(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	entries := []*ChannelMonitorHistoryEntry{
		{Status: MonitorStatusError, ErrorCategory: MonitorErrorCategoryAuth, Message: "invalid key", CheckedAt: now},
		{Status: MonitorStatusError, ErrorCategory: MonitorErrorCategoryRateLimit, CheckedAt: now.Add(-10 * time.Second)},
		{Status: MonitorStatusFailed, ErrorCategory: MonitorErrorCategoryChallenge, CheckedAt: now.Add(-20 * time.Second)},
	}

	snapshot := buildChannelMonitorHealthSnapshot(&ChannelMonitor{ID: 7}, entries, now)

	if snapshot.HealthStatus != MonitorHealthUnhealthy {
		t.Fatalf("HealthStatus = %q, want %q", snapshot.HealthStatus, MonitorHealthUnhealthy)
	}
	if snapshot.ConsecutiveFailedRuns != 3 {
		t.Fatalf("ConsecutiveFailedRuns = %d, want 3", snapshot.ConsecutiveFailedRuns)
	}
	if snapshot.SuccessRatePct != 0 {
		t.Fatalf("SuccessRatePct = %f, want 0", snapshot.SuccessRatePct)
	}
	if len(snapshot.TopErrorCategories) != 3 {
		t.Fatalf("TopErrorCategories length = %d, want 3", len(snapshot.TopErrorCategories))
	}
	if snapshot.LatestErrorCategory != MonitorErrorCategoryAuth {
		t.Fatalf("LatestErrorCategory = %q, want %q", snapshot.LatestErrorCategory, MonitorErrorCategoryAuth)
	}
}

func TestBuildChannelMonitorHealthSnapshot_HealthyAfterSuccessfulRuns(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	fast := 120
	entries := []*ChannelMonitorHistoryEntry{
		{Status: MonitorStatusOperational, LatencyMs: &fast, CheckedAt: now},
		{Status: MonitorStatusOperational, LatencyMs: &fast, CheckedAt: now.Add(-10 * time.Second)},
	}

	snapshot := buildChannelMonitorHealthSnapshot(&ChannelMonitor{ID: 7, AutoDisabled: true}, entries, now)

	if snapshot.HealthStatus != MonitorHealthHealthy {
		t.Fatalf("HealthStatus = %q, want %q", snapshot.HealthStatus, MonitorHealthHealthy)
	}
	if snapshot.ConsecutiveSuccessfulRuns != 2 {
		t.Fatalf("ConsecutiveSuccessfulRuns = %d, want 2", snapshot.ConsecutiveSuccessfulRuns)
	}
	if snapshot.SuccessRatePct != 100 {
		t.Fatalf("SuccessRatePct = %f, want 100", snapshot.SuccessRatePct)
	}
	if snapshot.AutoDisabled != true {
		t.Fatal("snapshot should reflect persisted AutoDisabled before transition update")
	}
}
