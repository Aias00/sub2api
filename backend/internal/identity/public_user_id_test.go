package identity

import (
	"strings"
	"testing"
	"time"
)

func TestNewPublicUserIDIntIsMonotonicWithinSameMillisecond(t *testing.T) {
	publicUserIDState.Lock()
	publicUserIDState.lastMillis = 0
	publicUserIDState.sequence = 0
	publicUserIDState.Unlock()

	fixed := time.Date(2026, 7, 8, 0, 0, 0, 123_000_000, time.UTC)
	first := newPublicUserIDInt(func() time.Time { return fixed })
	second := newPublicUserIDInt(func() time.Time { return fixed })

	if second <= first {
		t.Fatalf("second id = %d, want greater than %d", second, first)
	}
}

func TestNewPublicUserIDReturnsPrefixedLongString(t *testing.T) {
	got := NewPublicUserID()
	if !strings.HasPrefix(got, PublicUserIDPrefix) {
		t.Fatalf("public user id = %q, want prefix %q", got, PublicUserIDPrefix)
	}
	if len(got) < 18 {
		t.Fatalf("public user id length = %d, want a long decimal id: %q", len(got), got)
	}
}
