package ops

import (
	"testing"
	"time"
)

func TestComputeDashboardHealthScore(t *testing.T) {
	t.Parallel()

	if got := ComputeDashboardHealthScore(time.Now().UTC(), nil); got != 0 {
		t.Fatalf("nil health score = %d, want 0", got)
	}
	if got := ComputeDashboardHealthScore(time.Now().UTC(), &DashboardHealthInput{}); got != 100 {
		t.Fatalf("idle health score = %d, want 100", got)
	}

	score := ComputeDashboardHealthScore(time.Now().UTC(), &DashboardHealthInput{
		RequestCountTotal: 100,
		RequestCountSLA:   100,
		ErrorCountTotal:   10,
		ErrorRate:         0.10,
		UpstreamErrorRate: 0.08,
		TTFT:              Percentiles{P99: intPtr(2_000)},
		SystemMetrics: &SystemMetricsSnapshot{
			DBOK:               boolPtr(false),
			RedisOK:            boolPtr(false),
			CPUUsagePercent:    float64Ptr(98),
			MemoryUsagePercent: float64Ptr(97),
		},
		JobHeartbeats: []*JobHeartbeat{{LastErrorAt: timePtr(time.Now().UTC())}},
	})
	if score >= 80 || score < 0 {
		t.Fatalf("degraded health score = %d, want [0,80)", score)
	}
}

func TestComputeBusinessHealth(t *testing.T) {
	t.Parallel()

	score := ComputeBusinessHealth(&DashboardHealthInput{
		ErrorRate:         0.05,
		UpstreamErrorRate: 0,
		TTFT:              Percentiles{P99: intPtr(500)},
	})
	if score < 77 || score > 78 {
		t.Fatalf("business health = %.2f, want around 77-78", score)
	}
}

func TestComputeInfraHealth(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	score := ComputeInfraHealth(now, &DashboardHealthInput{
		SystemMetrics: &SystemMetricsSnapshot{
			DBOK:               boolPtr(true),
			RedisOK:            boolPtr(true),
			CPUUsagePercent:    float64Ptr(30),
			MemoryUsagePercent: float64Ptr(40),
		},
		JobHeartbeats: []*JobHeartbeat{{LastErrorAt: &now}},
	})
	if score < 70 || score > 90 {
		t.Fatalf("infra health = %.2f, want 70-90", score)
	}
}

func intPtr(v int) *int { return &v }

func boolPtr(v bool) *bool { return &v }

func float64Ptr(v float64) *float64 { return &v }

func timePtr(v time.Time) *time.Time { return &v }
