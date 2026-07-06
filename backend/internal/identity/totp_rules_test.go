package identity

import (
	"encoding/hex"
	"testing"
)

func TestResolveTotpIssuer(t *testing.T) {
	tests := []struct {
		name        string
		frontendURL string
		siteName    string
		want        string
	}{
		{name: "frontend host", frontendURL: "https://cloudbase.eu.org/login", siteName: "cloudbase", want: "cloudbase.eu.org"},
		{name: "site fallback", siteName: "cloudbase", want: "cloudbase"},
		{name: "default", want: DefaultTotpIssuer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveTotpIssuer(tt.frontendURL, tt.siteName); got != tt.want {
				t.Fatalf("ResolveTotpIssuer() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConstantTimeEqual(t *testing.T) {
	if !ConstantTimeEqual("same", "same") {
		t.Fatal("expected equal")
	}
	if ConstantTimeEqual("same", "different") {
		t.Fatal("expected not equal")
	}
}

func TestTotpSecretPrefix(t *testing.T) {
	if got := TotpSecretPrefix("ABCD1234"); got != "ABCD" {
		t.Fatalf("TotpSecretPrefix() = %q", got)
	}
	if got := TotpSecretPrefix("ABC"); got != "N/A" {
		t.Fatalf("TotpSecretPrefix(short) = %q", got)
	}
}

func TestGenerateRandomHexToken(t *testing.T) {
	got, err := GenerateRandomHexToken(16)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 32 {
		t.Fatalf("token len = %d", len(got))
	}
	if _, err := hex.DecodeString(got); err != nil {
		t.Fatal(err)
	}
}

func TestMaskEmail(t *testing.T) {
	tests := map[string]string{
		"ab":               "***",
		"abc":              "a***",
		"a@example.com":    "a***@example.com",
		"ab@example.com":   "a***@example.com",
		"aias@example.com": "a***s@example.com",
	}
	for input, want := range tests {
		if got := MaskEmail(input); got != want {
			t.Fatalf("MaskEmail(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestTotpVerificationMethod(t *testing.T) {
	if got := TotpVerificationMethod(true); got != "email" {
		t.Fatalf("TotpVerificationMethod(true) = %q", got)
	}
	if got := TotpVerificationMethod(false); got != "password" {
		t.Fatalf("TotpVerificationMethod(false) = %q", got)
	}
}
