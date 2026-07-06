package identity

import "testing"

func TestSignupGrantRiskAppliesToEmailSourcesOnly(t *testing.T) {
	allowed := []string{"email", "github", "google"}
	for _, source := range allowed {
		if !SignupGrantRiskAppliesToSource(source) {
			t.Fatalf("source %q should be eligible for signup grant risk control", source)
		}
	}

	blocked := []string{"", "touch", "linuxdo", "wechat", "oidc", "dingtalk"}
	for _, source := range blocked {
		if SignupGrantRiskAppliesToSource(source) {
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
		{name: "lowercase and trim", input: "  User@Example.COM  ", wantEmail: "user@example.com", wantDomain: "example.com"},
		{name: "strip plus alias", input: "user+promo@example.com", wantEmail: "user@example.com", wantDomain: "example.com"},
		{name: "gmail dots and googlemail", input: "u.ser+promo@googlemail.com", wantEmail: "user@gmail.com", wantDomain: "gmail.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEmail, gotDomain := NormalizeSignupGrantRiskEmail(tt.input)
			if gotEmail != tt.wantEmail || gotDomain != tt.wantDomain {
				t.Fatalf("NormalizeSignupGrantRiskEmail(%q) = (%q, %q), want (%q, %q)", tt.input, gotEmail, gotDomain, tt.wantEmail, tt.wantDomain)
			}
		})
	}
}

func TestNormalizeDomainListSetting(t *testing.T) {
	got := NormalizeDomainListSetting(" @TempMail.COM\n*.Trash.test, qq.com；QQ.com，corp.test ")
	want := "tempmail.com,trash.test,qq.com,corp.test"
	if got != want {
		t.Fatalf("NormalizeDomainListSetting() = %q, want %q", got, want)
	}
}

func TestSignupGrantRiskDomainLimitFor(t *testing.T) {
	cfg := SignupGrantRiskConfig{
		DomainLimit:     10,
		FreeDomainLimit: 3,
		FreeDomains:     ParseSignupGrantRiskDomainSet("gmail.com,qq.com"),
		TrustedDomains:  ParseSignupGrantRiskDomainSet("example.com"),
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
		if got := cfg.DomainLimitFor(tt.domain); got != tt.want {
			t.Fatalf("DomainLimitFor(%q) = %d, want %d", tt.domain, got, tt.want)
		}
	}
}

func TestSignupGrantDeviceSignalPrefersExplicitFingerprint(t *testing.T) {
	input := SignupGrantRiskInput{
		UserAgent:         "Mozilla/5.0",
		AcceptLanguage:    "zh-CN",
		DeviceFingerprint: " fp-123 ",
	}
	if got := SignupGrantDeviceSignal(input); got != "fp-123" {
		t.Fatalf("SignupGrantDeviceSignal() = %q, want explicit fingerprint", got)
	}

	input.DeviceFingerprint = ""
	if got := SignupGrantDeviceSignal(input); got != "Mozilla/5.0\x00zh-CN" {
		t.Fatalf("SignupGrantDeviceSignal() = %q, want UA/language fallback", got)
	}
}

func TestNormalizeSignupGrantRiskIP(t *testing.T) {
	if got := NormalizeSignupGrantRiskIP("127.0.0.1:1234"); got != "127.0.0.1" {
		t.Fatalf("NormalizeSignupGrantRiskIP() = %q", got)
	}
	if got := NormalizeSignupGrantRiskIP("not-ip"); got != "" {
		t.Fatalf("NormalizeSignupGrantRiskIP(invalid) = %q", got)
	}
}

func TestSignupGrantRiskHash(t *testing.T) {
	a := SignupGrantRiskHash("salt", " User@Example.COM ")
	b := SignupGrantRiskHash("salt", "user@example.com")
	if a == "" || a != b {
		t.Fatalf("SignupGrantRiskHash should normalize value, got %q and %q", a, b)
	}
}
