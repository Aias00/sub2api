package ops

import "testing"

func TestParsePlatformGroupFilter(t *testing.T) {
	got, err := ParsePlatformGroupFilter(PlatformGroupFilterInput{
		PlatformRaw: " openai ",
		GroupIDRaw:  "42",
	})
	if err != nil {
		t.Fatalf("ParsePlatformGroupFilter() error = %v", err)
	}
	if got.Platform != "openai" {
		t.Fatalf("platform = %q, want openai", got.Platform)
	}
	if got.GroupID == nil || *got.GroupID != 42 {
		t.Fatalf("group id = %v, want 42", got.GroupID)
	}
}

func TestParsePlatformGroupFilterOptionalGroup(t *testing.T) {
	got, err := ParsePlatformGroupFilter(PlatformGroupFilterInput{PlatformRaw: " gemini "})
	if err != nil {
		t.Fatalf("ParsePlatformGroupFilter() error = %v", err)
	}
	if got.Platform != "gemini" || got.GroupID != nil {
		t.Fatalf("unexpected filter: %+v", got)
	}
}

func TestParsePlatformGroupFilterInvalidGroup(t *testing.T) {
	tests := []string{"0", "-1", "bad"}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParsePlatformGroupFilter(PlatformGroupFilterInput{GroupIDRaw: raw}); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
