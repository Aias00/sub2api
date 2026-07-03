package ops

import "testing"

func TestParseErrorLogFilter(t *testing.T) {
	start := mustParseTimeRangeTime(t, "2026-01-02T03:04:05Z")
	end := mustParseTimeRangeTime(t, "2026-01-02T04:04:05Z")

	got, err := ParseErrorLogFilter(ErrorLogFilterInput{
		StartTime:          start,
		EndTime:            end,
		Page:               2,
		PageSize:           100,
		ViewRaw:            "all",
		PhaseRaw:           " upstream ",
		OwnerRaw:           " provider ",
		SourceRaw:          " openai ",
		QueryRaw:           " timeout ",
		UserQueryRaw:       " user@example.com ",
		ModelRaw:           " gpt ",
		RequestIDRaw:       " req ",
		ClientRequestIDRaw: " creq ",
		PlatformRaw:        " openai ",
		GroupIDRaw:         "1",
		AccountIDRaw:       "2",
		UserIDRaw:          "3",
		APIKeyIDRaw:        "4",
		ResolvedRaw:        "yes",
		StatusCodesRaw:     " 400, 500,,0 ",
		ClearUpstreamPhase: true,
	})
	if err != nil {
		t.Fatalf("ParseErrorLogFilter() error = %v", err)
	}
	if got.Page != 2 || got.PageSize != 100 || got.View != ListViewAll {
		t.Fatalf("unexpected pagination/view: %+v", got)
	}
	if got.StartTime == nil || !got.StartTime.Equal(start) || got.EndTime == nil || !got.EndTime.Equal(end) {
		t.Fatalf("unexpected time range: %+v", got)
	}
	if got.Phase != "" || got.Owner != "provider" || got.Source != "openai" || got.Query != "timeout" || got.UserQuery != "user@example.com" || got.Model != "gpt" || got.RequestID != "req" || got.ClientRequestID != "creq" || got.Platform != "openai" {
		t.Fatalf("string fields not normalized: %+v", got)
	}
	if got.GroupID == nil || *got.GroupID != 1 || got.AccountID == nil || *got.AccountID != 2 || got.UserID == nil || *got.UserID != 3 || got.APIKeyID == nil || *got.APIKeyID != 4 {
		t.Fatalf("ids not parsed: %+v", got)
	}
	if got.Resolved == nil || !*got.Resolved {
		t.Fatalf("resolved = %v, want true", got.Resolved)
	}
	if len(got.StatusCodes) != 3 || got.StatusCodes[0] != 400 || got.StatusCodes[1] != 500 || got.StatusCodes[2] != 0 {
		t.Fatalf("status codes = %v", got.StatusCodes)
	}
}

func TestParseErrorLogFilterErrors(t *testing.T) {
	tests := []struct {
		name  string
		input ErrorLogFilterInput
	}{
		{name: "group", input: ErrorLogFilterInput{GroupIDRaw: "0"}},
		{name: "account", input: ErrorLogFilterInput{AccountIDRaw: "bad"}},
		{name: "user", input: ErrorLogFilterInput{UserIDRaw: "-1"}},
		{name: "api_key", input: ErrorLogFilterInput{APIKeyIDRaw: "bad"}},
		{name: "resolved", input: ErrorLogFilterInput{ResolvedRaw: "maybe"}},
		{name: "status", input: ErrorLogFilterInput{StatusCodesRaw: "400,bad"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseErrorLogFilter(tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
