//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSortUserProfileTimelineOrdersDescendingAndCaps(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, time.July, 7, 10, 0, 0, 0, time.UTC)
	summary := &UserProfileSummary{}
	for i := 0; i < 205; i++ {
		summary.Timeline = append(summary.Timeline, UserProfileTimelineEvent{
			OccurredAt: base.Add(time.Duration(i) * time.Minute),
			Source:     "test",
			Action:     "event",
			Title:      "event",
			RecordID:   int64RecordID(int64(i)),
		})
	}

	sortUserProfileTimeline(summary)

	require.Len(t, summary.Timeline, 200)
	require.Equal(t, "204", summary.Timeline[0].RecordID)
	require.Equal(t, "5", summary.Timeline[len(summary.Timeline)-1].RecordID)
	for i := 1; i < len(summary.Timeline); i++ {
		require.False(t, summary.Timeline[i].OccurredAt.After(summary.Timeline[i-1].OccurredAt))
	}
}
