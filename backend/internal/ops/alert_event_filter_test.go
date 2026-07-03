package ops

import (
	"testing"
	"time"
)

func TestParseAlertEventFilterDefaultsAndTrims(t *testing.T) {
	got, err := ParseAlertEventFilter(AlertEventFilterInput{
		StatusRaw:   " firing ",
		SeverityRaw: " P1 ",
		PlatformRaw: " openai ",
	})
	if err != nil {
		t.Fatalf("ParseAlertEventFilter() error = %v", err)
	}
	if got.Limit != 20 {
		t.Fatalf("limit = %d, want 20", got.Limit)
	}
	if got.Status != "firing" || got.Severity != "P1" || got.Platform != "openai" {
		t.Fatalf("filters were not trimmed: %+v", got)
	}
}

func TestParseAlertEventFilterEmailCursorGroupAndTimeRange(t *testing.T) {
	start := mustParseTimeRangeTime(t, "2026-01-02T03:04:05Z")
	end := mustParseTimeRangeTime(t, "2026-01-02T04:04:05Z")

	got, err := ParseAlertEventFilter(AlertEventFilterInput{
		LimitRaw:         "50",
		EmailSentRaw:     "1",
		BeforeFiredAtRaw: "2026-01-02T04:04:05.123Z",
		BeforeIDRaw:      "99",
		GroupIDRaw:       "42",
		StartTime:        start,
		EndTime:          end,
		HasTimeRange:     true,
	})
	if err != nil {
		t.Fatalf("ParseAlertEventFilter() error = %v", err)
	}
	if got.Limit != 50 {
		t.Fatalf("limit = %d, want 50", got.Limit)
	}
	if got.EmailSent == nil || !*got.EmailSent {
		t.Fatalf("email_sent = %v, want true", got.EmailSent)
	}
	if got.BeforeFiredAt == nil || got.BeforeFiredAt.Format("2006-01-02T15:04:05.999Z07:00") != "2026-01-02T04:04:05.123Z" {
		t.Fatalf("before_fired_at = %v", got.BeforeFiredAt)
	}
	if got.BeforeID == nil || *got.BeforeID != 99 || got.GroupID == nil || *got.GroupID != 42 {
		t.Fatalf("ids were not parsed: %+v", got)
	}
	if got.StartTime == nil || !got.StartTime.Equal(start) || got.EndTime == nil || !got.EndTime.Equal(end) {
		t.Fatalf("time range was not applied: %+v", got)
	}
}

func TestParseAlertEventFilterErrors(t *testing.T) {
	tests := []struct {
		name  string
		input AlertEventFilterInput
	}{
		{name: "limit", input: AlertEventFilterInput{LimitRaw: "0"}},
		{name: "email", input: AlertEventFilterInput{EmailSentRaw: "maybe"}},
		{name: "cursor_pair", input: AlertEventFilterInput{BeforeFiredAtRaw: "2026-01-02T04:04:05Z"}},
		{name: "cursor_time", input: AlertEventFilterInput{BeforeFiredAtRaw: "bad", BeforeIDRaw: "1"}},
		{name: "cursor_id", input: AlertEventFilterInput{BeforeFiredAtRaw: "2026-01-02T04:04:05Z", BeforeIDRaw: "0"}},
		{name: "group", input: AlertEventFilterInput{GroupIDRaw: "bad"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseAlertEventFilter(tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseAlertEventStatus(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: " resolved ", want: AlertStatusResolved},
		{raw: "manual_resolved", want: AlertStatusManualResolved},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := ParseAlertEventStatus(tt.raw)
			if err != nil {
				t.Fatalf("ParseAlertEventStatus() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("status = %q, want %q", got, tt.want)
			}
		})
	}

	if _, err := ParseAlertEventStatus("firing"); err == nil {
		t.Fatal("expected invalid status error")
	}
	if _, err := ParseAlertEventStatus(" "); err == nil {
		t.Fatal("expected blank status error")
	}
}

func TestAlertEventResolvedAt(t *testing.T) {
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.FixedZone("test", 8*60*60))
	for _, status := range []string{AlertStatusResolved, AlertStatusManualResolved} {
		t.Run(status, func(t *testing.T) {
			got := AlertEventResolvedAt(" "+status+" ", now)
			if got == nil {
				t.Fatal("resolved at = nil, want timestamp")
			}
			if !got.Equal(now.UTC()) {
				t.Fatalf("resolved at = %v, want %v", got, now.UTC())
			}
			if got.Location() != time.UTC {
				t.Fatalf("resolved at location = %v, want UTC", got.Location())
			}
		})
	}

	if got := AlertEventResolvedAt("firing", now); got != nil {
		t.Fatalf("resolved at = %v, want nil", got)
	}
}

func TestParseAlertSilence(t *testing.T) {
	groupID := int64(42)
	region := " us-east "

	got, err := ParseAlertSilence(AlertSilenceInput{
		RuleID:      7,
		PlatformRaw: " openai ",
		GroupID:     &groupID,
		Region:      &region,
		UntilRaw:    " 2026-01-02T04:04:05Z ",
		ReasonRaw:   " deploy ",
	})
	if err != nil {
		t.Fatalf("ParseAlertSilence() error = %v", err)
	}
	if got.RuleID != 7 || got.Platform != "openai" || got.Reason != "deploy" {
		t.Fatalf("unexpected silence: %+v", got)
	}
	if got.GroupID == nil || *got.GroupID != 42 {
		t.Fatalf("group id = %v, want 42", got.GroupID)
	}
	if got.Region == nil || *got.Region != region {
		t.Fatalf("region pointer should be preserved, got %v", got.Region)
	}
	if got.Until.Format(time.RFC3339) != "2026-01-02T04:04:05Z" {
		t.Fatalf("until = %s", got.Until.Format(time.RFC3339))
	}
}

func TestParseAlertSilenceInvalidUntil(t *testing.T) {
	if _, err := ParseAlertSilence(AlertSilenceInput{UntilRaw: "bad"}); err == nil {
		t.Fatal("expected invalid until error")
	}
}
