package ops

type GroupAvailabilityCounts struct {
	TotalAccounts  int64
	AvailableCount int64
}

func ComputeGroupAvailableRatio(group *GroupAvailabilityCounts) float64 {
	if group == nil || group.TotalAccounts <= 0 {
		return 0
	}
	return (float64(group.AvailableCount) / float64(group.TotalAccounts)) * 100
}

func CountByCondition[T any](items map[int64]*T, condition func(*T) bool) int64 {
	if len(items) == 0 || condition == nil {
		return 0
	}
	var count int64
	for _, item := range items {
		if item != nil && condition(item) {
			count++
		}
	}
	return count
}
