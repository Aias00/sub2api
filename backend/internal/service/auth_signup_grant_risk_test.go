package service

import "testing"

func TestSignupGrantRiskAppliesToEmailSourcesOnly(t *testing.T) {
	allowed := []string{"email", "github", "google"}
	for _, source := range allowed {
		if !signupGrantRiskAppliesToSource(source) {
			t.Fatalf("source %q should be eligible for signup grant risk control", source)
		}
	}

	blocked := []string{"", "touch", "linuxdo", "wechat", "oidc", "dingtalk"}
	for _, source := range blocked {
		if signupGrantRiskAppliesToSource(source) {
			t.Fatalf("source %q should not be eligible for signup grant risk control", source)
		}
	}
}

func TestNormalizeSignupGrantRiskEmail(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantEmail  string
		wantDomain string
	}{
		{
			name:       "lowercase and trim",
			input:      "  User@Example.COM  ",
			wantEmail:  "user@example.com",
			wantDomain: "example.com",
		},
		{
			name:       "strip plus alias",
			input:      "user+promo@example.com",
			wantEmail:  "user@example.com",
			wantDomain: "example.com",
		},
		{
			name:       "gmail dots and googlemail",
			input:      "u.ser+promo@googlemail.com",
			wantEmail:  "user@gmail.com",
			wantDomain: "gmail.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEmail, gotDomain := normalizeSignupGrantRiskEmail(tt.input)
			if gotEmail != tt.wantEmail || gotDomain != tt.wantDomain {
				t.Fatalf("normalizeSignupGrantRiskEmail(%q) = (%q, %q), want (%q, %q)", tt.input, gotEmail, gotDomain, tt.wantEmail, tt.wantDomain)
			}
		})
	}
}

func TestNormalizeDomainListSetting(t *testing.T) {
	got := normalizeDomainListSetting(" @TempMail.COM\n*.Trash.test, qq.com；QQ.com，corp.test ")
	want := "tempmail.com,trash.test,qq.com,corp.test"
	if got != want {
		t.Fatalf("normalizeDomainListSetting() = %q, want %q", got, want)
	}
}

func TestSignupGrantRiskDomainLimitFor(t *testing.T) {
	cfg := signupGrantRiskConfig{
		DomainLimit:     10,
		FreeDomainLimit: 3,
		FreeDomains:     parseSignupGrantRiskDomainSet("gmail.com,qq.com"),
		TrustedDomains:  parseSignupGrantRiskDomainSet("example.com"),
	}

	tests := []struct {
		domain string
		want   int
	}{
		{domain: "gmail.com", want: 3},
		{domain: "qq.com", want: 3},
		{domain: "example.com", want: 0},
		{domain: "corp.test", want: 10},
	}

	for _, tt := range tests {
		if got := cfg.domainLimitFor(tt.domain); got != tt.want {
			t.Fatalf("domainLimitFor(%q) = %d, want %d", tt.domain, got, tt.want)
		}
	}
}

func TestSignupGrantDeviceSignalPrefersExplicitFingerprint(t *testing.T) {
	input := SignupGrantRiskInput{
		UserAgent:         "Mozilla/5.0",
		AcceptLanguage:    "zh-CN",
		DeviceFingerprint: " fp-123 ",
	}
	if got := signupGrantDeviceSignal(input); got != "fp-123" {
		t.Fatalf("signupGrantDeviceSignal() = %q, want explicit fingerprint", got)
	}

	input.DeviceFingerprint = ""
	if got := signupGrantDeviceSignal(input); got != "Mozilla/5.0\x00zh-CN" {
		t.Fatalf("signupGrantDeviceSignal() = %q, want UA/language fallback", got)
	}
}
