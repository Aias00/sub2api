package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"
)

// GetHealthSnapshot returns the operational health view for one monitor.
func (s *ChannelMonitorService) GetHealthSnapshot(ctx context.Context, id int64) (*ChannelMonitorHealthSnapshot, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	entries, err := s.repo.ListHistory(ctx, id, "", monitorHealthHistoryLimit)
	if err != nil {
		return nil, fmt.Errorf("list monitor health history: %w", err)
	}
	return buildChannelMonitorHealthSnapshot(m, entries, time.Now()), nil
}

// evaluateHealthTransition updates monitor-level auto-disabled/recovered state.
// It deliberately does not flip Enabled, so the runner can keep probing and
// auto-recovery can happen without manual intervention.
func (s *ChannelMonitorService) evaluateHealthTransition(ctx context.Context, m *ChannelMonitor) {
	if m == nil {
		return
	}
	snapshot, err := s.GetHealthSnapshot(ctx, m.ID)
	if err != nil {
		slog.Warn("channel_monitor: evaluate health failed", "monitor_id", m.ID, "error", err)
		return
	}

	autoDisabled := m.AutoDisabled
	reason := m.AutoDisabledReason
	disabledAt := m.AutoDisabledAt
	recoveredAt := m.AutoRecoveredAt
	wasAutoDisabled := m.AutoDisabled
	now := time.Now()

	if snapshot.ConsecutiveFailedRuns >= monitorAutoDisableConsecutiveFailures {
		autoDisabled = true
		reason = buildMonitorAutoDisableReason(snapshot)
		if !m.AutoDisabled {
			disabledAt = &now
		}
	} else if m.AutoDisabled && snapshot.ConsecutiveSuccessfulRuns >= monitorAutoRecoverConsecutiveSuccesses {
		autoDisabled = false
		reason = ""
		disabledAt = nil
		recoveredAt = &now
	}

	if autoDisabled == m.AutoDisabled &&
		reason == m.AutoDisabledReason &&
		healthStatusOrUnknown(m.LastHealthStatus) == snapshot.HealthStatus {
		return
	}
	if err := s.repo.UpdateHealthState(ctx, m.ID, snapshot.HealthStatus, autoDisabled, reason, disabledAt, recoveredAt); err != nil {
		slog.Warn("channel_monitor: update health state failed",
			"monitor_id", m.ID, "health_status", snapshot.HealthStatus, "auto_disabled", autoDisabled, "error", err)
		return
	}

	switch {
	case !wasAutoDisabled && autoDisabled:
		s.emitMonitorHealthAlert(ctx, m, snapshot, OpsAlertStatusFiring, "P1", "Channel monitor auto-disabled", reason)
	case wasAutoDisabled && !autoDisabled:
		s.emitMonitorHealthAlert(ctx, m, snapshot, OpsAlertStatusResolved, "P2", "Channel monitor auto-recovered", "monitor recovered after successful checks")
	}
}

func (s *ChannelMonitorService) emitMonitorHealthAlert(
	ctx context.Context,
	m *ChannelMonitor,
	snapshot *ChannelMonitorHealthSnapshot,
	status string,
	severity string,
	title string,
	reason string,
) {
	if s == nil || s.alertSink == nil || m == nil || snapshot == nil {
		return
	}
	alertCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	now := time.Now().UTC()
	event := &OpsAlertEvent{
		Severity:    severity,
		Status:      status,
		Title:       fmt.Sprintf("%s: %s", title, monitorDisplayName(m)),
		Description: buildMonitorHealthAlertDescription(snapshot, reason),
		MetricValue: float64Ptr(snapshot.SuccessRatePct),
		Dimensions: map[string]any{
			"source":                      "channel_monitor",
			"monitor_id":                  m.ID,
			"monitor_name":                monitorDisplayName(m),
			"provider":                    m.Provider,
			"health_status":               snapshot.HealthStatus,
			"latest_status":               snapshot.LatestStatus,
			"latest_error_category":       snapshot.LatestErrorCategory,
			"consecutive_failed_runs":     snapshot.ConsecutiveFailedRuns,
			"consecutive_successful_runs": snapshot.ConsecutiveSuccessfulRuns,
			"success_rate_pct":            snapshot.SuccessRatePct,
			"auto_disabled":               status == OpsAlertStatusFiring,
		},
		FiredAt:   now,
		CreatedAt: now,
	}
	if status == OpsAlertStatusResolved {
		event.ResolvedAt = &now
	}
	if _, err := s.alertSink.CreateAlertEvent(alertCtx, event); err != nil {
		slog.Warn("channel_monitor: emit health alert failed",
			"monitor_id", m.ID, "status", status, "health_status", snapshot.HealthStatus, "error", err)
	}
}

func buildMonitorHealthAlertDescription(snapshot *ChannelMonitorHealthSnapshot, reason string) string {
	if snapshot == nil {
		return strings.TrimSpace(reason)
	}
	parts := []string{
		fmt.Sprintf("health=%s", snapshot.HealthStatus),
		fmt.Sprintf("success_rate=%.2f%%", snapshot.SuccessRatePct),
		fmt.Sprintf("failed_runs=%d", snapshot.ConsecutiveFailedRuns),
		fmt.Sprintf("successful_runs=%d", snapshot.ConsecutiveSuccessfulRuns),
	}
	if snapshot.AvgLatencyMs != nil {
		parts = append(parts, fmt.Sprintf("avg_latency_ms=%d", *snapshot.AvgLatencyMs))
	}
	if strings.TrimSpace(snapshot.LatestErrorCategory) != "" {
		parts = append(parts, fmt.Sprintf("latest_error=%s", snapshot.LatestErrorCategory))
	}
	if strings.TrimSpace(reason) != "" {
		parts = append(parts, fmt.Sprintf("reason=%s", strings.TrimSpace(reason)))
	}
	return strings.Join(parts, "; ")
}

func monitorDisplayName(m *ChannelMonitor) string {
	if m == nil {
		return "unknown"
	}
	if name := strings.TrimSpace(m.Name); name != "" {
		return name
	}
	return fmt.Sprintf("monitor-%d", m.ID)
}

func buildChannelMonitorHealthSnapshot(
	m *ChannelMonitor,
	entries []*ChannelMonitorHistoryEntry,
	now time.Time,
) *ChannelMonitorHealthSnapshot {
	snapshot := &ChannelMonitorHealthSnapshot{
		HealthStatus:       MonitorHealthUnknown,
		WindowMinutes:      int(monitorHealthWindow / time.Minute),
		TopErrorCategories: []MonitorErrorCategoryCount{},
	}
	if m != nil {
		snapshot.MonitorID = m.ID
		snapshot.AutoDisabled = m.AutoDisabled
		snapshot.AutoDisabledAt = m.AutoDisabledAt
		snapshot.AutoDisabledReason = m.AutoDisabledReason
		snapshot.AutoRecoveredAt = m.AutoRecoveredAt
	}
	if len(entries) == 0 {
		return snapshot
	}

	latest := entries[0]
	snapshot.LatestStatus = latest.Status
	snapshot.LatestErrorCategory = normalizeMonitorErrorCategory(latest.Status, latest.ErrorCategory)
	snapshot.LatestMessage = latest.Message
	latestAt := latest.CheckedAt
	snapshot.LatestCheckedAt = &latestAt

	recent := filterRecentMonitorHistory(entries, now.Add(-monitorHealthWindow))
	fillMonitorHealthWindowStats(snapshot, recent)

	runs := groupMonitorHistoryRuns(entries)
	snapshot.ConsecutiveFailedRuns = countConsecutiveRuns(runs, monitorRunFailed)
	snapshot.ConsecutiveSuccessfulRuns = countConsecutiveRuns(runs, monitorRunSuccessful)
	snapshot.HealthStatus = deriveMonitorHealthStatus(snapshot, recent)
	return snapshot
}

func filterRecentMonitorHistory(entries []*ChannelMonitorHistoryEntry, cutoff time.Time) []*ChannelMonitorHistoryEntry {
	out := make([]*ChannelMonitorHistoryEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.CheckedAt.Before(cutoff) {
			out = append(out, entry)
		}
	}
	return out
}

func fillMonitorHealthWindowStats(snapshot *ChannelMonitorHealthSnapshot, entries []*ChannelMonitorHistoryEntry) {
	if len(entries) == 0 {
		return
	}
	categoryCounts := make(map[string]int)
	var latencySum, latencyCount int
	for _, entry := range entries {
		snapshot.TotalChecks++
		if monitorStatusOK(entry.Status) {
			snapshot.SuccessfulChecks++
		}
		if entry.LatencyMs != nil {
			latencySum += *entry.LatencyMs
			latencyCount++
		}
		category := normalizeMonitorErrorCategory(entry.Status, entry.ErrorCategory)
		if category != MonitorErrorCategoryNone {
			categoryCounts[category]++
		}
	}
	if snapshot.TotalChecks > 0 {
		snapshot.SuccessRatePct = float64(snapshot.SuccessfulChecks) * 100 / float64(snapshot.TotalChecks)
	}
	if latencyCount > 0 {
		avg := latencySum / latencyCount
		snapshot.AvgLatencyMs = &avg
	}
	snapshot.TopErrorCategories = topMonitorErrorCategories(categoryCounts, 5)
}

type monitorHistoryRun []*ChannelMonitorHistoryEntry

func groupMonitorHistoryRuns(entries []*ChannelMonitorHistoryEntry) []monitorHistoryRun {
	runs := make([]monitorHistoryRun, 0, len(entries))
	for _, entry := range entries {
		if len(runs) == 0 || !sameMonitorHistoryRun(runs[len(runs)-1][0].CheckedAt, entry.CheckedAt) {
			runs = append(runs, monitorHistoryRun{entry})
			continue
		}
		runs[len(runs)-1] = append(runs[len(runs)-1], entry)
	}
	return runs
}

func sameMonitorHistoryRun(anchor, checkedAt time.Time) bool {
	diff := anchor.Sub(checkedAt)
	if diff < 0 {
		diff = -diff
	}
	return diff <= monitorHealthRunGroupWindow
}

func countConsecutiveRuns(runs []monitorHistoryRun, pred func(monitorHistoryRun) bool) int {
	count := 0
	for _, run := range runs {
		if !pred(run) {
			break
		}
		count++
	}
	return count
}

func monitorRunFailed(run monitorHistoryRun) bool {
	if len(run) == 0 {
		return false
	}
	for _, entry := range run {
		if monitorStatusOK(entry.Status) {
			return false
		}
	}
	return true
}

func monitorRunSuccessful(run monitorHistoryRun) bool {
	if len(run) == 0 {
		return false
	}
	for _, entry := range run {
		if !monitorStatusOK(entry.Status) {
			return false
		}
	}
	return true
}

func deriveMonitorHealthStatus(snapshot *ChannelMonitorHealthSnapshot, recent []*ChannelMonitorHistoryEntry) string {
	if snapshot.TotalChecks == 0 {
		return MonitorHealthUnknown
	}
	if snapshot.ConsecutiveFailedRuns >= monitorAutoDisableConsecutiveFailures ||
		(snapshot.TotalChecks >= 3 && snapshot.SuccessRatePct < 50) {
		return MonitorHealthUnhealthy
	}
	if snapshot.ConsecutiveFailedRuns > 0 ||
		snapshot.SuccessRatePct < 99 ||
		hasRecentNonOperationalStatus(recent) ||
		(snapshot.AvgLatencyMs != nil && *snapshot.AvgLatencyMs >= int(monitorDegradedThreshold/time.Millisecond)) {
		return MonitorHealthDegraded
	}
	return MonitorHealthHealthy
}

func hasRecentNonOperationalStatus(entries []*ChannelMonitorHistoryEntry) bool {
	for _, entry := range entries {
		if entry.Status != MonitorStatusOperational {
			return true
		}
	}
	return false
}

func monitorStatusOK(status string) bool {
	return status == MonitorStatusOperational || status == MonitorStatusDegraded
}

func normalizeMonitorErrorCategory(status, category string) string {
	if category != "" {
		return category
	}
	switch status {
	case MonitorStatusOperational:
		return MonitorErrorCategoryNone
	case MonitorStatusDegraded:
		return MonitorErrorCategorySlow
	case MonitorStatusFailed, MonitorStatusError:
		return MonitorErrorCategoryUnknown
	default:
		return MonitorErrorCategoryUnknown
	}
}

func topMonitorErrorCategories(counts map[string]int, limit int) []MonitorErrorCategoryCount {
	out := make([]MonitorErrorCategoryCount, 0, len(counts))
	for category, count := range counts {
		if category == MonitorErrorCategoryNone || count <= 0 {
			continue
		}
		out = append(out, MonitorErrorCategoryCount{Category: category, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Category < out[j].Category
		}
		return out[i].Count > out[j].Count
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func buildMonitorAutoDisableReason(snapshot *ChannelMonitorHealthSnapshot) string {
	category := snapshot.LatestErrorCategory
	if category == MonitorErrorCategoryNone && len(snapshot.TopErrorCategories) > 0 {
		category = snapshot.TopErrorCategories[0].Category
	}
	if category == MonitorErrorCategoryNone {
		category = MonitorErrorCategoryUnknown
	}
	message := snapshot.LatestMessage
	if message == "" {
		message = category
	}
	return truncateMessage(fmt.Sprintf("auto disabled after %d consecutive failed checks (%s): %s",
		snapshot.ConsecutiveFailedRuns, category, message))
}

func healthStatusOrUnknown(status string) string {
	if status == "" {
		return MonitorHealthUnknown
	}
	return status
}
