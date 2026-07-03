package ops

import (
	"math"
	"time"
)

type Percentiles struct {
	P99 *int
}

type SystemMetricsSnapshot struct {
	DBOK               *bool
	RedisOK            *bool
	CPUUsagePercent    *float64
	MemoryUsagePercent *float64
}

type JobHeartbeat struct {
	LastSuccessAt *time.Time
	LastErrorAt   *time.Time
}

type DashboardHealthInput struct {
	RequestCountSLA   int64
	RequestCountTotal int64
	ErrorCountTotal   int64

	ErrorRate         float64
	UpstreamErrorRate float64

	TTFT Percentiles

	SystemMetrics *SystemMetricsSnapshot
	JobHeartbeats []*JobHeartbeat
}

func ComputeDashboardHealthScore(now time.Time, input *DashboardHealthInput) int {
	if input == nil {
		return 0
	}

	if input.RequestCountSLA <= 0 && input.RequestCountTotal <= 0 && input.ErrorCountTotal <= 0 {
		return 100
	}

	businessHealth := ComputeBusinessHealth(input)
	infraHealth := ComputeInfraHealth(now, input)
	score := businessHealth*0.7 + infraHealth*0.3
	return int(math.Round(ClampFloat64(score, 0, 100)))
}

func ComputeBusinessHealth(input *DashboardHealthInput) float64 {
	errorScore := 100.0
	errorPct := ClampFloat64(input.ErrorRate*100, 0, 100)
	upstreamPct := ClampFloat64(input.UpstreamErrorRate*100, 0, 100)
	combinedErrorPct := math.Max(errorPct, upstreamPct)
	if combinedErrorPct > 1.0 {
		if combinedErrorPct <= 10.0 {
			errorScore = (10.0 - combinedErrorPct) / 9.0 * 100
		} else {
			errorScore = 0
		}
	}

	ttftScore := 100.0
	if input.TTFT.P99 != nil {
		p99 := float64(*input.TTFT.P99)
		if p99 > 1000 {
			if p99 <= 3000 {
				ttftScore = (3000 - p99) / 2000 * 100
			} else {
				ttftScore = 0
			}
		}
	}

	return errorScore*0.5 + ttftScore*0.5
}

func ComputeInfraHealth(now time.Time, input *DashboardHealthInput) float64 {
	storageScore := 100.0
	if input.SystemMetrics != nil {
		if input.SystemMetrics.DBOK != nil && !*input.SystemMetrics.DBOK {
			storageScore = 0
		} else if input.SystemMetrics.RedisOK != nil && !*input.SystemMetrics.RedisOK {
			storageScore = 50
		}
	}

	computeScore := 100.0
	if input.SystemMetrics != nil {
		cpuScore := 100.0
		if input.SystemMetrics.CPUUsagePercent != nil {
			cpuPct := ClampFloat64(*input.SystemMetrics.CPUUsagePercent, 0, 100)
			if cpuPct > 80 {
				if cpuPct <= 100 {
					cpuScore = (100 - cpuPct) / 20 * 100
				} else {
					cpuScore = 0
				}
			}
		}

		memScore := 100.0
		if input.SystemMetrics.MemoryUsagePercent != nil {
			memPct := ClampFloat64(*input.SystemMetrics.MemoryUsagePercent, 0, 100)
			if memPct > 85 {
				if memPct <= 100 {
					memScore = (100 - memPct) / 15 * 100
				} else {
					memScore = 0
				}
			}
		}

		computeScore = (cpuScore + memScore) / 2
	}

	jobScore := 100.0
	failedJobs := 0
	totalJobs := 0
	for _, hb := range input.JobHeartbeats {
		if hb == nil {
			continue
		}
		totalJobs++
		if hb.LastErrorAt != nil && (hb.LastSuccessAt == nil || hb.LastErrorAt.After(*hb.LastSuccessAt)) {
			failedJobs++
		} else if hb.LastSuccessAt != nil && now.Sub(*hb.LastSuccessAt) > 15*time.Minute {
			failedJobs++
		}
	}
	if totalJobs > 0 && failedJobs > 0 {
		jobScore = (1 - float64(failedJobs)/float64(totalJobs)) * 100
	}

	return storageScore*0.4 + computeScore*0.3 + jobScore*0.3
}

func ClampFloat64(v float64, min float64, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
