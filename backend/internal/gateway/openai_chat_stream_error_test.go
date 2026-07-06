package gateway

import (
	"strings"
	"testing"
)

func TestBuildOpenAIChatStreamErrorSSE(t *testing.T) {
	got := BuildOpenAIChatStreamErrorSSE("cyber_policy", "blocked")
	if !strings.HasPrefix(got, "data: ") || !strings.HasSuffix(got, "\n\n") {
		t.Fatalf("invalid SSE frame: %q", got)
	}
	if !strings.Contains(got, `"code":"cyber_policy"`) {
		t.Fatalf("missing code: %q", got)
	}
	if !strings.Contains(got, `"message":"blocked"`) {
		t.Fatalf("missing message: %q", got)
	}
}
