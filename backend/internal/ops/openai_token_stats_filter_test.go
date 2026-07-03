package ops

import (
	"testing"
	"time"
)

func TestParseOpenAITokenStatsFilterDefaultPagination(t *testing.T) {
	now := mustParseTimeRangeTime(t, "2026-01-02T04:04:05Z")

	got, err := ParseOpenAITokenStatsFilter(OpenAITokenStatsFilterInput{
		PlatformRaw: " openai ",
		Now:         now,
	})
	if err != nil {
		t.Fatalf("ParseOpenAITokenStatsFilter() error = %v", err)
	}
	if got.TimeRange != "30d" {
		t.Fatalf("time range = %q, want 30d", got.TimeRange)
	}
	if got.Platform != "openai" {
		t.Fatalf("platform = %q, want openai", got.Platform)
	}
	if got.Page != 1 || got.PageSize != 20 {
		t.Fatalf("page = %d page_size = %d, want 1/20", got.Page, got.PageSize)
	}
	if !got.EndTime.Equal(now.UTC()) {
		t.Fatalf("end = %s, want %s", got.EndTime, now.UTC())
	}
	if got.EndTime.Sub(got.StartTime) != 30*24*time.Hour {
		t.Fatalf("window = %s, want 30d", got.EndTime.Sub(got.StartTime))
	}
}

func TestParseOpenAITokenStatsFilterGroupTopN(t *testing.T) {
	got, err := ParseOpenAITokenStatsFilter(OpenAITokenStatsFilterInput{
		TimeRangeRaw: " 1h ",
		GroupIDRaw:   "42",
		TopNRaw:      "10",
		Now:          mustParseTimeRangeTime(t, "2026-01-02T04:04:05Z"),
	})
	if err != nil {
		t.Fatalf("ParseOpenAITokenStatsFilter() error = %v", err)
	}
	if got.TimeRange != "1h" || got.TopN != 10 {
		t.Fatalf("time_range/top_n = %q/%d, want 1h/10", got.TimeRange, got.TopN)
	}
	if got.GroupID == nil || *got.GroupID != 42 {
		t.Fatalf("group id = %v, want 42", got.GroupID)
	}
	if got.Page != 0 || got.PageSize != 0 {
		t.Fatalf("top_n mode should not set pagination, got %d/%d", got.Page, got.PageSize)
	}
}

func TestParseOpenAITokenStatsFilterPagination(t *testing.T) {
	got, err := ParseOpenAITokenStatsFilter(OpenAITokenStatsFilterInput{
		TimeRangeRaw: "15d",
		PageRaw:      "3",
		PageSizeRaw:  "50",
		Now:          mustParseTimeRangeTime(t, "2026-01-02T04:04:05Z"),
	})
	if err != nil {
		t.Fatalf("ParseOpenAITokenStatsFilter() error = %v", err)
	}
	if got.Page != 3 || got.PageSize != 50 {
		t.Fatalf("page = %d page_size = %d, want 3/50", got.Page, got.PageSize)
	}
}

func TestParseOpenAITokenStatsFilterErrors(t *testing.T) {
	tests := []struct {
		name  string
		input OpenAITokenStatsFilterInput
	}{
		{name: "time_range", input: OpenAITokenStatsFilterInput{TimeRangeRaw: "7d"}},
		{name: "group_id", input: OpenAITokenStatsFilterInput{GroupIDRaw: "0"}},
		{name: "top_n_with_page", input: OpenAITokenStatsFilterInput{TopNRaw: "5", PageRaw: "1"}},
		{name: "top_n", input: OpenAITokenStatsFilterInput{TopNRaw: "101"}},
		{name: "page", input: OpenAITokenStatsFilterInput{PageRaw: "0"}},
		{name: "page_size", input: OpenAITokenStatsFilterInput{PageSizeRaw: "101"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseOpenAITokenStatsFilter(tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
