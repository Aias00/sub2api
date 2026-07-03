package ops

import (
	"errors"
	"testing"
)

func TestParseQueryMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want QueryMode
	}{
		{raw: "", want: QueryModeAuto},
		{raw: " raw ", want: QueryModeRaw},
		{raw: "PREAGG", want: QueryModePreagg},
		{raw: "unknown", want: QueryModeAuto},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()

			if got := ParseQueryMode(tt.raw); got != tt.want {
				t.Fatalf("ParseQueryMode(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseOptionalQueryMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want QueryMode
	}{
		{raw: "", want: ""},
		{raw: "   ", want: ""},
		{raw: " raw ", want: QueryModeRaw},
		{raw: "unknown", want: QueryModeAuto},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()

			if got := ParseOptionalQueryMode(tt.raw); got != tt.want {
				t.Fatalf("ParseOptionalQueryMode(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestQueryModeIsValid(t *testing.T) {
	t.Parallel()

	if !QueryModeAuto.IsValid() || !QueryModeRaw.IsValid() || !QueryModePreagg.IsValid() {
		t.Fatal("known query modes should be valid")
	}
	if QueryMode("custom").IsValid() {
		t.Fatal("custom query mode should be invalid")
	}
}

func TestShouldFallbackPreagg(t *testing.T) {
	t.Parallel()

	otherErr := errors.New("other")
	if !ShouldFallbackPreagg(QueryModeAuto, ErrPreaggregatedNotPopulated) {
		t.Fatal("auto mode should fallback for preagg population error")
	}
	if !ShouldFallbackPreagg(QueryModeAuto, errors.Join(ErrPreaggregatedNotPopulated, otherErr)) {
		t.Fatal("auto mode should fallback for wrapped preagg population error")
	}
	if ShouldFallbackPreagg(QueryModeRaw, ErrPreaggregatedNotPopulated) {
		t.Fatal("raw mode should not fallback")
	}
	if ShouldFallbackPreagg(QueryModeAuto, otherErr) {
		t.Fatal("other errors should not fallback")
	}
}

func TestParseListView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want string
	}{
		{raw: "", want: ListViewErrors},
		{raw: " errors ", want: ListViewErrors},
		{raw: "EXCLUDED", want: ListViewExcluded},
		{raw: "all", want: ListViewAll},
		{raw: "unknown", want: ListViewErrors},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()

			if got := ParseListView(tt.raw); got != tt.want {
				t.Fatalf("ParseListView(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
