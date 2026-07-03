package ops

import "testing"

func TestComputeGroupAvailableRatio(t *testing.T) {
	tests := []struct {
		name  string
		group *GroupAvailabilityCounts
		want  float64
	}{
		{name: "nil group", want: 0},
		{name: "zero total", group: &GroupAvailabilityCounts{TotalAccounts: 0, AvailableCount: 8}, want: 0},
		{name: "zero available", group: &GroupAvailabilityCounts{TotalAccounts: 10, AvailableCount: 0}, want: 0},
		{name: "normal ratio", group: &GroupAvailabilityCounts{TotalAccounts: 10, AvailableCount: 8}, want: 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeGroupAvailableRatio(tt.group)
			if got != tt.want {
				t.Fatalf("ComputeGroupAvailableRatio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountByCondition(t *testing.T) {
	type item struct {
		enabled bool
	}

	items := map[int64]*item{
		1: {enabled: true},
		2: {enabled: false},
		3: {enabled: true},
		4: nil,
	}

	got := CountByCondition(items, func(v *item) bool {
		return v.enabled
	})
	if got != 2 {
		t.Fatalf("CountByCondition() = %d, want 2", got)
	}

	if got := CountByCondition(map[int64]*item{}, func(v *item) bool { return v.enabled }); got != 0 {
		t.Fatalf("empty CountByCondition() = %d, want 0", got)
	}
	if got := CountByCondition(items, nil); got != 0 {
		t.Fatalf("nil condition CountByCondition() = %d, want 0", got)
	}
}
