package vendorpreset

import (
	"strings"
	"testing"

	"github.com/Aias00/cloudbase/internal/domain"
)

func TestAllReturnsCatalog(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("expected non-empty vendor catalog")
	}
}

func TestAllReturnsCopy(t *testing.T) {
	a := All()
	if len(a) == 0 {
		t.Fatal("empty catalog")
	}
	original := a[0].ID
	a[0].ID = "mutated"
	b := All()
	if b[0].ID != original {
		t.Fatalf("All() must return a copy; catalog was mutated: got %q want %q", b[0].ID, original)
	}
}

func TestByID(t *testing.T) {
	p, ok := ByID("deepseek")
	if !ok {
		t.Fatal("expected to find deepseek preset")
	}
	if p.DisplayName != "DeepSeek" {
		t.Fatalf("unexpected display name: %q", p.DisplayName)
	}

	if _, ok := ByID("does-not-exist"); ok {
		t.Fatal("expected missing preset to return ok=false")
	}
}

// TestCatalogIntegrity enforces invariants every preset must satisfy so a bad
// entry fails the build instead of shipping a broken vendor.
func TestCatalogIntegrity(t *testing.T) {
	seen := map[string]bool{}
	validPlatforms := map[string]bool{
		domain.PlatformOpenAI:      true,
		domain.PlatformGrok:        true,
		domain.PlatformAnthropic:   true,
		domain.PlatformGemini:      true,
		domain.PlatformAntigravity: true,
	}

	for _, p := range All() {
		if p.ID == "" {
			t.Errorf("preset %q has empty ID", p.DisplayName)
		}
		if seen[p.ID] {
			t.Errorf("duplicate preset ID: %q", p.ID)
		}
		seen[p.ID] = true

		if p.DisplayName == "" {
			t.Errorf("preset %q has empty DisplayName", p.ID)
		}
		if !validPlatforms[p.Platform] {
			t.Errorf("preset %q has invalid platform %q", p.ID, p.Platform)
		}
		if p.AccountType == "" {
			t.Errorf("preset %q has empty AccountType", p.ID)
		}
		if len(p.DefaultModels) == 0 {
			t.Errorf("preset %q has no default models", p.ID)
		}
		if !strings.HasPrefix(p.BaseURL, "https://") {
			t.Errorf("preset %q base URL must be https: %q", p.ID, p.BaseURL)
		}
		if strings.HasSuffix(p.BaseURL, "/") {
			t.Errorf("preset %q base URL must not have trailing slash: %q", p.ID, p.BaseURL)
		}
	}
}

// TestOpenAIStyleBaseURLContract verifies every OpenAI-style preset's base URL
// yields a correct chat/completions URL under the gateway's URL builder rule:
//   - ends with "/chat/completions" -> used as-is
//   - ends with "/v1"               -> "/chat/completions" appended
//   - otherwise (bare/other)        -> "/v1/chat/completions" appended
//
// A base like "https://host/api/paas/v4" would silently produce
// ".../v4/v1/chat/completions", so we require OpenAI-style bases to end in a
// recognizable version segment or /chat/completions.
func TestOpenAIStyleBaseURLContract(t *testing.T) {
	for _, p := range All() {
		if p.APIStyle != APIStyleOpenAI {
			continue
		}
		if strings.HasSuffix(p.BaseURL, "/chat/completions") {
			continue
		}
		if strings.HasSuffix(p.BaseURL, "/v1") {
			continue
		}
		t.Errorf("preset %q (openai style) base URL %q does not end in /v1 or /chat/completions; "+
			"forwarding URL builder would produce an incorrect path", p.ID, p.BaseURL)
	}
}

// TestAnthropicStyleBaseURLContract verifies every Anthropic-style preset's
// base URL is a ROOT: the Anthropic API-key passthrough path appends
// "/v1/messages" to account.GetBaseURL(), so a base that already contains
// "/v1/messages" (or a trailing slash) would produce a broken upstream URL.
func TestAnthropicStyleBaseURLContract(t *testing.T) {
	for _, p := range All() {
		if p.APIStyle != APIStyleAnthropic {
			continue
		}
		if strings.Contains(p.BaseURL, "/v1/messages") {
			t.Errorf("preset %q (anthropic style) base URL %q must not contain /v1/messages; "+
				"the passthrough path appends it", p.ID, p.BaseURL)
		}
		if strings.HasSuffix(p.BaseURL, "/v1") {
			t.Errorf("preset %q (anthropic style) base URL %q must be a root (no /v1 suffix); "+
				"appending /v1/messages would yield /v1/v1/messages", p.ID, p.BaseURL)
		}
	}
}

// TestHasAnthropicStylePresets guards that the Anthropic-style catalog is
// actually populated (phase 2 deliverable), so a future refactor that drops
// them fails loudly.
func TestHasAnthropicStylePresets(t *testing.T) {
	count := 0
	for _, p := range All() {
		if p.APIStyle == APIStyleAnthropic {
			count++
			if p.Platform != domain.PlatformAnthropic {
				t.Errorf("anthropic-style preset %q must map to platform %q, got %q",
					p.ID, domain.PlatformAnthropic, p.Platform)
			}
		}
	}
	if count == 0 {
		t.Fatal("expected at least one anthropic-style preset")
	}
}
