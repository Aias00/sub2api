package service

import (
	"time"

	opsctx "github.com/Aias00/cloudbase/internal/ops"
)

func computeDashboardHealthScore(now time.Time, overview *OpsDashboardOverview) int {
	return opsctx.ComputeDashboardHealthScore(now, opsHealthInputFromOverview(overview))
}

func opsHealthInputFromOverview(overview *OpsDashboardOverview) *opsctx.DashboardHealthInput {
	if overview == nil {
		return nil
	}
	input := &opsctx.DashboardHealthInput{
		RequestCountSLA:   overview.RequestCountSLA,
		RequestCountTotal: overview.RequestCountTotal,
		ErrorCountTotal:   overview.ErrorCountTotal,
		ErrorRate:         overview.ErrorRate,
		UpstreamErrorRate: overview.UpstreamErrorRate,
		TTFT:              opsctx.Percentiles{P99: overview.TTFT.P99},
	}
	if overview.SystemMetrics != nil {
		input.SystemMetrics = &opsctx.SystemMetricsSnapshot{
			DBOK:               overview.SystemMetrics.DBOK,
			RedisOK:            overview.SystemMetrics.RedisOK,
			CPUUsagePercent:    overview.SystemMetrics.CPUUsagePercent,
			MemoryUsagePercent: overview.SystemMetrics.MemoryUsagePercent,
		}
	}
	if len(overview.JobHeartbeats) > 0 {
		input.JobHeartbeats = make([]*opsctx.JobHeartbeat, 0, len(overview.JobHeartbeats))
		for _, hb := range overview.JobHeartbeats {
			if hb == nil {
				input.JobHeartbeats = append(input.JobHeartbeats, nil)
				continue
			}
			input.JobHeartbeats = append(input.JobHeartbeats, &opsctx.JobHeartbeat{
				LastSuccessAt: hb.LastSuccessAt,
				LastErrorAt:   hb.LastErrorAt,
			})
		}
	}
	return input
}
