package identity

import (
	"net/http"
	"testing"
)

func TestFingerprintFromHeadersUsesDefaults(t *testing.T) {
	defaults := Fingerprint{
		UserAgent:               "claude-cli/1.0.0",
		StainlessLang:           "js",
		StainlessPackageVersion: "0.94.0",
		StainlessOS:             "Linux",
		StainlessArch:           "arm64",
		StainlessRuntime:        "node",
		StainlessRuntimeVersion: "v24.3.0",
	}
	headers := http.Header{"User-Agent": []string{"claude-cli/1.2.3"}}

	got := FingerprintFromHeaders(headers, defaults)
	if got.UserAgent != "claude-cli/1.2.3" || got.StainlessLang != "js" || got.StainlessOS != "Linux" {
		t.Fatalf("FingerprintFromHeaders() = %#v", got)
	}
}

func TestMergeHeadersIntoFingerprintOnlyOverwritesPresentHeaders(t *testing.T) {
	fp := Fingerprint{
		UserAgent:     "claude-cli/1.0.0",
		StainlessLang: "js",
		StainlessOS:   "Linux",
	}
	headers := http.Header{
		"User-Agent":     []string{"claude-cli/1.2.0"},
		"X-Stainless-OS": []string{"Darwin"},
	}

	MergeHeadersIntoFingerprint(&fp, headers)
	if fp.UserAgent != "claude-cli/1.2.0" || fp.StainlessLang != "js" || fp.StainlessOS != "Darwin" {
		t.Fatalf("MergeHeadersIntoFingerprint() = %#v", fp)
	}
}

func TestApplyFingerprintHeadersUsesRawCasing(t *testing.T) {
	headers := http.Header{"X-Stainless-Os": []string{"old"}}
	ApplyFingerprintHeaders(headers, &Fingerprint{
		UserAgent:               "claude-cli/1.2.0",
		StainlessOS:             "Darwin",
		StainlessPackageVersion: "1.0.0",
	})

	if _, ok := headers["X-Stainless-Os"]; ok {
		t.Fatal("canonicalized old casing should be removed")
	}
	rawOSHeader := "X-Stainless-OS"
	if got := headers[rawOSHeader]; len(got) != 1 || got[0] != "Darwin" {
		t.Fatalf("X-Stainless-OS = %#v", got)
	}
}

func TestIsNewerUserAgentVersion(t *testing.T) {
	tests := []struct {
		newUA    string
		cachedUA string
		want     bool
	}{
		{"claude-cli/1.2.1", "claude-cli/1.2.0", true},
		{"claude-cli/1.3.0", "claude-cli/1.2.9", true},
		{"claude-cli/2.0.0", "claude-cli/1.9.9", true},
		{"claude-cli/1.2.0", "claude-cli/1.2.1", false},
		{"Mozilla/5.0", "claude-cli/1.2.1", false},
		{"other/2.0.0", "claude-cli/1.2.1", false},
	}

	for _, tt := range tests {
		if got := IsNewerUserAgentVersion(tt.newUA, tt.cachedUA); got != tt.want {
			t.Fatalf("IsNewerUserAgentVersion(%q, %q) = %v, want %v", tt.newUA, tt.cachedUA, got, tt.want)
		}
	}
}
