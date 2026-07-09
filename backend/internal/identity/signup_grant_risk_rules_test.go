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

func TestNormalizeSignupGrantRiskIPForHash(t *testing.T) {
	// 用户报告的真实地址：隐私扩展 IPv6，后 64 位会轮换。
	const rotatingV6 = "2408:8215:5413:4700:cd91:7fd7:23c8:c71"
	tests := []struct {
		name   string
		raw    string
		prefix int
		want   string
	}{
		// IPv4 透传，不截断（无论 prefix 设多少）。
		{name: "ipv4 passthrough", raw: "203.0.113.7", prefix: 64, want: "203.0.113.7"},
		{name: "ipv4 with port passthrough", raw: "203.0.113.7:8080", prefix: 64, want: "203.0.113.7"},
		{name: "ipv4 prefix ignored", raw: "203.0.113.7", prefix: 32, want: "203.0.113.7"},

		// IPv6 /64 截断（默认值）。
		{name: "ipv6 /64 truncates interface id", raw: rotatingV6, prefix: 64, want: "2408:8215:5413:4700::"},
		{name: "ipv6 /64 with port", raw: "[2408:8215:5413:4700:cd91:7fd7:23c8:c71]:443", prefix: 64, want: "2408:8215:5413:4700::"},
		{name: "ipv6 /64 loopback", raw: "::1", prefix: 64, want: "::"},

		// 更紧前缀。
		{name: "ipv6 /56", raw: "2408:8215:5413:4700:cd91:7fd7:23c8:c71", prefix: 56, want: "2408:8215:5413:4700::"},
		{name: "ipv6 /48", raw: "2408:8215:5413:4700:cd91:7fd7:23c8:c71", prefix: 48, want: "2408:8215:5413::"},
		{name: "ipv6 /32", raw: "2408:8215:5413:4700:cd91:7fd7:23c8:c71", prefix: 32, want: "2408:8215::"},

		// /0 禁用截断，保留全量（等价于 NormalizeSignupGrantRiskIP）。
		{name: "ipv6 prefix 0 keeps full", raw: rotatingV6, prefix: 0, want: "2408:8215:5413:4700:cd91:7fd7:23c8:c71"},
		{name: "ipv4 prefix 0 keeps full", raw: "203.0.113.7", prefix: 0, want: "203.0.113.7"},

		// 非法 prefixBits 回退默认 /64。
		{name: "ipv6 illegal negative prefix falls back /64", raw: rotatingV6, prefix: -1, want: "2408:8215:5413:4700::"},
		{name: "ipv6 illegal over-128 prefix falls back /64", raw: rotatingV6, prefix: 200, want: "2408:8215:5413:4700::"},
		{name: "ipv6 exactly 128 keeps full", raw: rotatingV6, prefix: 128, want: "2408:8215:5413:4700:cd91:7fd7:23c8:c71"},

		// IPv4-mapped IPv6 视作 IPv4 透传。
		{name: "ipv4-mapped ipv6 treated as ipv4", raw: "::ffff:203.0.113.7", prefix: 64, want: "203.0.113.7"},

		// 非法 / 空输入。
		{name: "empty returns empty", raw: "", prefix: 64, want: ""},
		{name: "whitespace only returns empty", raw: "   ", prefix: 64, want: ""},
		{name: "invalid ip returns empty", raw: "not-an-ip", prefix: 64, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSignupGrantRiskIPForHash(tt.raw, tt.prefix)
			if got != tt.want {
				t.Fatalf("NormalizeSignupGrantRiskIPForHash(%q, %d) = %q, want %q", tt.raw, tt.prefix, got, tt.want)
			}
		})
	}

	// 同一 /64 内的不同接口标识符应归一化到同一哈希键——这是 IPv6 轮换绕过的核心防御。
	const otherIfaceV6 = "2408:8215:5413:4700:1:2:3:4"
	a := NormalizeSignupGrantRiskIPForHash(rotatingV6, 64)
	b := NormalizeSignupGrantRiskIPForHash(otherIfaceV6, 64)
	if a != b {
		t.Fatalf("same /64 different iface id should collide: %q vs %q", a, b)
	}
	// 但全量（prefix=0）下应不同，证明截断确实是聚合的原因。
	fullA := NormalizeSignupGrantRiskIPForHash(rotatingV6, 0)
	fullB := NormalizeSignupGrantRiskIPForHash(otherIfaceV6, 0)
	if fullA == fullB {
		t.Fatalf("different iface ids should differ under full IP: %q == %q", fullA, fullB)
	}
}

func TestSignupGrantRiskHash(t *testing.T) {
	a := SignupGrantRiskHash("salt", " User@Example.COM ")
	b := SignupGrantRiskHash("salt", "user@example.com")
	if a == "" || a != b {
		t.Fatalf("SignupGrantRiskHash should normalize value, got %q and %q", a, b)
	}
}
