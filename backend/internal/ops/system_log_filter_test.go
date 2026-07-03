package ops

import (
	"testing"
	"time"
)

func TestParseSystemLogListFilter(t *testing.T) {
	start := mustParseTimeRangeTime(t, "2026-01-02T03:04:05Z")
	end := mustParseTimeRangeTime(t, "2026-01-02T04:04:05Z")

	got, err := ParseSystemLogListFilter(SystemLogListFilterInput{
		StartTime:          start,
		EndTime:            end,
		Page:               2,
		PageSize:           100,
		LevelRaw:           " warn ",
		ComponentRaw:       " gateway ",
		RequestIDRaw:       " req ",
		ClientRequestIDRaw: " creq ",
		UserIDRaw:          " 11 ",
		APIKeyIDRaw:        "12",
		AccountIDRaw:       "13",
		PlatformRaw:        " openai ",
		ModelRaw:           " gpt ",
		QueryRaw:           " error ",
	})
	if err != nil {
		t.Fatalf("ParseSystemLogListFilter() error = %v", err)
	}
	if got.Page != 2 || got.PageSize != 100 {
		t.Fatalf("page = %d page_size = %d, want 2/100", got.Page, got.PageSize)
	}
	if got.StartTime == nil || !got.StartTime.Equal(start) || got.EndTime == nil || !got.EndTime.Equal(end) {
		t.Fatalf("unexpected time range: %v %v", got.StartTime, got.EndTime)
	}
	if got.Level != "warn" || got.Component != "gateway" || got.RequestID != "req" || got.ClientRequestID != "creq" {
		t.Fatalf("string filters were not trimmed: %+v", got)
	}
	if got.Platform != "openai" || got.Model != "gpt" || got.Query != "error" {
		t.Fatalf("query filters were not trimmed: %+v", got)
	}
	if got.UserID == nil || *got.UserID != 11 || got.APIKeyID == nil || *got.APIKeyID != 12 || got.AccountID == nil || *got.AccountID != 13 {
		t.Fatalf("ids were not parsed: %+v", got)
	}
}

func TestParseSystemLogListFilterOptionalFields(t *testing.T) {
	got, err := ParseSystemLogListFilter(SystemLogListFilterInput{})
	if err != nil {
		t.Fatalf("ParseSystemLogListFilter() error = %v", err)
	}
	if got.StartTime != nil || got.EndTime != nil || got.UserID != nil || got.APIKeyID != nil || got.AccountID != nil {
		t.Fatalf("blank optional fields should stay nil: %+v", got)
	}
}

func TestParseSystemLogListFilterInvalidIDs(t *testing.T) {
	tests := []struct {
		name  string
		input SystemLogListFilterInput
	}{
		{name: "user", input: SystemLogListFilterInput{UserIDRaw: "0"}},
		{name: "api_key", input: SystemLogListFilterInput{APIKeyIDRaw: "-1"}},
		{name: "account", input: SystemLogListFilterInput{AccountIDRaw: "bad"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseSystemLogListFilter(tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseSystemLogCleanupFilter(t *testing.T) {
	userID := int64(0)
	apiKeyID := int64(123)
	accountID := int64(-1)

	got, err := ParseSystemLogCleanupFilter(SystemLogCleanupFilterInput{
		StartTimeRaw:       " 2026-01-02T03:04:05Z ",
		EndTimeRaw:         "2026-01-02T04:04:05.123Z",
		LevelRaw:           " warn ",
		ComponentRaw:       " gateway ",
		RequestIDRaw:       " req ",
		ClientRequestIDRaw: " creq ",
		UserID:             &userID,
		APIKeyID:           &apiKeyID,
		AccountID:          &accountID,
		PlatformRaw:        " openai ",
		ModelRaw:           " gpt ",
		QueryRaw:           " error ",
	})
	if err != nil {
		t.Fatalf("ParseSystemLogCleanupFilter() error = %v", err)
	}
	if got.StartTime == nil || got.StartTime.Format(time.RFC3339) != "2026-01-02T03:04:05Z" {
		t.Fatalf("unexpected start_time: %v", got.StartTime)
	}
	if got.EndTime == nil || got.EndTime.Format(time.RFC3339Nano) != "2026-01-02T04:04:05.123Z" {
		t.Fatalf("unexpected end_time: %v", got.EndTime)
	}
	if got.Level != "warn" || got.Component != "gateway" || got.RequestID != "req" || got.ClientRequestID != "creq" {
		t.Fatalf("string filters were not trimmed: %+v", got)
	}
	if got.UserID == nil || *got.UserID != 0 || got.APIKeyID == nil || *got.APIKeyID != 123 || got.AccountID == nil || *got.AccountID != -1 {
		t.Fatalf("id pointers should be preserved except api key validation: %+v", got)
	}
}

func TestParseSystemLogCleanupFilterErrors(t *testing.T) {
	badAPIKeyID := int64(0)
	tests := []struct {
		name  string
		input SystemLogCleanupFilterInput
	}{
		{name: "start", input: SystemLogCleanupFilterInput{StartTimeRaw: "bad"}},
		{name: "end", input: SystemLogCleanupFilterInput{EndTimeRaw: "bad"}},
		{name: "api_key", input: SystemLogCleanupFilterInput{APIKeyID: &badAPIKeyID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseSystemLogCleanupFilter(tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
