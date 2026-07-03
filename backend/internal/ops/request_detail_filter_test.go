package ops

import "testing"

func TestParseRequestDetailFilter(t *testing.T) {
	start := mustParseTimeRangeTime(t, "2026-01-02T03:04:05Z")
	end := mustParseTimeRangeTime(t, "2026-01-02T04:04:05Z")

	got, err := ParseRequestDetailFilter(RequestDetailFilterInput{
		StartTime:        start,
		EndTime:          end,
		Page:             2,
		PageSize:         100,
		KindRaw:          " error ",
		PlatformRaw:      " openai ",
		ModelRaw:         " gpt ",
		RequestIDRaw:     " req ",
		QueryRaw:         " user ",
		SortRaw:          " duration_desc ",
		UserIDRaw:        "1",
		APIKeyIDRaw:      "2",
		AccountIDRaw:     "3",
		GroupIDRaw:       "4",
		MinDurationMsRaw: "0",
		MaxDurationMsRaw: "1200",
	})
	if err != nil {
		t.Fatalf("ParseRequestDetailFilter() error = %v", err)
	}
	if got.Page != 2 || got.PageSize != 100 {
		t.Fatalf("unexpected pagination: %+v", got)
	}
	if got.StartTime == nil || !got.StartTime.Equal(start) || got.EndTime == nil || !got.EndTime.Equal(end) {
		t.Fatalf("unexpected time range: %+v", got)
	}
	if got.Kind != "error" || got.Platform != "openai" || got.Model != "gpt" || got.RequestID != "req" || got.Query != "user" || got.Sort != "duration_desc" {
		t.Fatalf("string fields not trimmed: %+v", got)
	}
	if got.UserID == nil || *got.UserID != 1 || got.APIKeyID == nil || *got.APIKeyID != 2 || got.AccountID == nil || *got.AccountID != 3 || got.GroupID == nil || *got.GroupID != 4 {
		t.Fatalf("ids not parsed: %+v", got)
	}
	if got.MinDurationMs == nil || *got.MinDurationMs != 0 || got.MaxDurationMs == nil || *got.MaxDurationMs != 1200 {
		t.Fatalf("duration filters not parsed: %+v", got)
	}
}

func TestParseRequestDetailFilterErrors(t *testing.T) {
	tests := []struct {
		name  string
		input RequestDetailFilterInput
	}{
		{name: "user", input: RequestDetailFilterInput{UserIDRaw: "0"}},
		{name: "api_key", input: RequestDetailFilterInput{APIKeyIDRaw: "bad"}},
		{name: "account", input: RequestDetailFilterInput{AccountIDRaw: "-1"}},
		{name: "group", input: RequestDetailFilterInput{GroupIDRaw: "bad"}},
		{name: "min_duration", input: RequestDetailFilterInput{MinDurationMsRaw: "-1"}},
		{name: "max_duration", input: RequestDetailFilterInput{MaxDurationMsRaw: "bad"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseRequestDetailFilter(tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
