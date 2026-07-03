package ops

import (
	"net/netip"
	"testing"
)

func TestParseWSOriginPolicy(t *testing.T) {
	got, ok := ParseWSOriginPolicy(" STRICT ")
	if !ok || got != WSOriginPolicyStrict {
		t.Fatalf("policy = %q ok=%v, want strict true", got, ok)
	}
	got, ok = ParseWSOriginPolicy("permissive")
	if !ok || got != WSOriginPolicyPermissive {
		t.Fatalf("policy = %q ok=%v, want permissive true", got, ok)
	}
	if _, ok := ParseWSOriginPolicy("unknown"); ok {
		t.Fatal("expected invalid origin policy")
	}
}

func TestRoundTo1DP(t *testing.T) {
	tests := []struct {
		name string
		raw  float64
		want float64
	}{
		{name: "round up", raw: 1.25, want: 1.3},
		{name: "round down", raw: 1.24, want: 1.2},
		{name: "zero", raw: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoundTo1DP(tt.raw); got != tt.want {
				t.Fatalf("RoundTo1DP(%v) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestWSClientIPSlotKey(t *testing.T) {
	if got, ok := WSClientIPSlotKey(" 203.0.113.1 "); !ok || got != "203.0.113.1" {
		t.Fatalf("slot key = %q ok=%v, want trimmed true", got, ok)
	}
	if got, ok := WSClientIPSlotKey(" "); ok || got != "" {
		t.Fatalf("slot key = %q ok=%v, want blank false", got, ok)
	}
}

func TestParseWSProxyConfig(t *testing.T) {
	defaultProxies, invalidDefault := ParseWSTrustedProxyList("127.0.0.0/8")
	if len(invalidDefault) > 0 {
		t.Fatalf("invalid default proxies = %#v", invalidDefault)
	}

	cfg, invalid := ParseWSProxyConfig(WSProxyConfigInput{
		TrustProxyRaw:     "false",
		TrustedProxiesRaw: "10.0.0.0/8, bad",
		OriginPolicyRaw:   "strict",
		DefaultTrustProxy: true,
		DefaultProxies:    defaultProxies,
		DefaultPolicy:     WSOriginPolicyPermissive,
	})

	if cfg.TrustProxy {
		t.Fatal("trust proxy = true, want false")
	}
	if len(cfg.TrustedProxies) != 1 || cfg.TrustedProxies[0].String() != "10.0.0.0/8" {
		t.Fatalf("trusted proxies = %#v, want 10.0.0.0/8", cfg.TrustedProxies)
	}
	if cfg.OriginPolicy != WSOriginPolicyStrict {
		t.Fatalf("origin policy = %q, want strict", cfg.OriginPolicy)
	}
	if invalid.TrustProxy || invalid.OriginPolicy {
		t.Fatalf("unexpected bool/policy invalid flags: %#v", invalid)
	}
	if len(invalid.TrustedProxies) != 1 || invalid.TrustedProxies[0] != "bad" {
		t.Fatalf("invalid trusted proxies = %#v, want [bad]", invalid.TrustedProxies)
	}
}

func TestParseWSProxyConfigInvalidKeepsDefaults(t *testing.T) {
	defaultProxies, _ := ParseWSTrustedProxyList("127.0.0.0/8")
	cfg, invalid := ParseWSProxyConfig(WSProxyConfigInput{
		TrustProxyRaw:     "maybe",
		OriginPolicyRaw:   "locked-down",
		DefaultTrustProxy: true,
		DefaultProxies:    defaultProxies,
		DefaultPolicy:     WSOriginPolicyPermissive,
	})

	if !cfg.TrustProxy {
		t.Fatal("trust proxy = false, want default true")
	}
	if cfg.OriginPolicy != WSOriginPolicyPermissive {
		t.Fatalf("origin policy = %q, want permissive", cfg.OriginPolicy)
	}
	if !invalid.TrustProxy || !invalid.OriginPolicy {
		t.Fatalf("invalid = %#v, want trust proxy and origin policy flags", invalid)
	}
}

func TestParseWSRuntimeLimits(t *testing.T) {
	cfg, invalid := ParseWSRuntimeLimits(WSRuntimeLimitsInput{
		MaxConnsRaw:          "250",
		MaxConnsPerIPRaw:     "0",
		DefaultMaxConns:      100,
		DefaultMaxConnsPerIP: 20,
	})

	if cfg.MaxConns != 250 || cfg.MaxConnsPerIP != 0 {
		t.Fatalf("limits = %#v, want max=250 perIP=0", cfg)
	}
	if invalid.MaxConns || invalid.MaxConnsPerIP {
		t.Fatalf("invalid = %#v, want none", invalid)
	}
}

func TestParseWSRuntimeLimitsInvalidKeepsDefaults(t *testing.T) {
	cfg, invalid := ParseWSRuntimeLimits(WSRuntimeLimitsInput{
		MaxConnsRaw:          "0",
		MaxConnsPerIPRaw:     "-1",
		DefaultMaxConns:      100,
		DefaultMaxConnsPerIP: 20,
	})

	if cfg.MaxConns != 100 || cfg.MaxConnsPerIP != 20 {
		t.Fatalf("limits = %#v, want defaults", cfg)
	}
	if !invalid.MaxConns || !invalid.MaxConnsPerIP {
		t.Fatalf("invalid = %#v, want both flags", invalid)
	}
}

func TestParseWSBoolFlag(t *testing.T) {
	if got, ok := ParseWSBoolFlag(" true "); !ok || !got {
		t.Fatalf("bool flag = %v ok=%v, want true true", got, ok)
	}
	if got, ok := ParseWSBoolFlag("0"); !ok || got {
		t.Fatalf("bool flag = %v ok=%v, want false true", got, ok)
	}
	if _, ok := ParseWSBoolFlag("maybe"); ok {
		t.Fatal("expected invalid bool flag")
	}
}

func TestParseWSLimits(t *testing.T) {
	if got, ok := ParseWSPositiveLimit(" 10 "); !ok || got != 10 {
		t.Fatalf("positive limit = %d ok=%v, want 10 true", got, ok)
	}
	for _, raw := range []string{"0", "-1", "bad"} {
		t.Run("positive_"+raw, func(t *testing.T) {
			if _, ok := ParseWSPositiveLimit(raw); ok {
				t.Fatal("expected invalid positive limit")
			}
		})
	}

	if got, ok := ParseWSNonNegativeLimit("0"); !ok || got != 0 {
		t.Fatalf("non-negative limit = %d ok=%v, want 0 true", got, ok)
	}
	if got, ok := ParseWSNonNegativeLimit("5"); !ok || got != 5 {
		t.Fatalf("non-negative limit = %d ok=%v, want 5 true", got, ok)
	}
	for _, raw := range []string{"-1", "bad"} {
		t.Run("non_negative_"+raw, func(t *testing.T) {
			if _, ok := ParseWSNonNegativeLimit(raw); ok {
				t.Fatal("expected invalid non-negative limit")
			}
		})
	}
}

func TestParseWSTrustedProxyList(t *testing.T) {
	prefixes, invalid := ParseWSTrustedProxyList("127.0.0.1, 10.0.0.0/8, ::1, bad, ")
	if len(invalid) != 1 || invalid[0] != "bad" {
		t.Fatalf("invalid = %#v, want [bad]", invalid)
	}
	if len(prefixes) != 3 {
		t.Fatalf("prefix count = %d, want 3", len(prefixes))
	}
	if got := prefixes[0].String(); got != "127.0.0.1/32" {
		t.Fatalf("prefix[0] = %q, want 127.0.0.1/32", got)
	}
	if got := prefixes[1].String(); got != "10.0.0.0/8" {
		t.Fatalf("prefix[1] = %q, want 10.0.0.0/8", got)
	}
	if got := prefixes[2].String(); got != "::1/128" {
		t.Fatalf("prefix[2] = %q, want ::1/128", got)
	}
}

func TestWSHostWithoutPort(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "host port", raw: "example.com:443", want: "example.com"},
		{name: "ipv6 bracket", raw: "[::1]", want: "::1"},
		{name: "ipv6 bracket port", raw: "[::1]:443", want: "::1"},
		{name: "plain host", raw: " example.com ", want: "example.com"},
		{name: "blank", raw: " ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WSHostWithoutPort(tt.raw); got != tt.want {
				t.Fatalf("WSHostWithoutPort(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseWSForwardedHost(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "left most", raw: " edge.example, app.example ", want: "edge.example", ok: true},
		{name: "single", raw: " app.example ", want: "app.example", ok: true},
		{name: "blank", raw: " ", ok: false},
		{name: "blank left most", raw: " , app.example", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseWSForwardedHost(tt.raw)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("host = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseWSPeerAddr(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "host port", raw: "127.0.0.1:1234", want: "127.0.0.1", ok: true},
		{name: "ipv6 host port", raw: "[::1]:1234", want: "::1", ok: true},
		{name: "plain addr", raw: " 10.0.0.1 ", want: "10.0.0.1", ok: true},
		{name: "bracket addr", raw: "[::1]", want: "::1", ok: true},
		{name: "invalid", raw: "bad", ok: false},
		{name: "blank", raw: " ", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseWSPeerAddr(tt.raw)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if ok && got.String() != tt.want {
				t.Fatalf("addr = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestParseWSForwardedForClientIP(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "left most", raw: " 203.0.113.1, 10.0.0.1 ", want: "203.0.113.1", ok: true},
		{name: "ipv6 bracket", raw: "[::1], 10.0.0.1", want: "::1", ok: true},
		{name: "mapped ipv4", raw: "::ffff:192.0.2.1", want: "192.0.2.1", ok: true},
		{name: "invalid", raw: "bad, 10.0.0.1", ok: false},
		{name: "blank", raw: " ", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseWSForwardedForClientIP(tt.raw)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("client ip = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsWSOriginAllowed(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		host    string
		policy  string
		allowed bool
	}{
		{name: "missing origin permissive", policy: WSOriginPolicyPermissive, allowed: true},
		{name: "missing origin empty policy", allowed: true},
		{name: "missing origin strict", policy: WSOriginPolicyStrict, allowed: false},
		{name: "same host", origin: "https://Example.com", host: "example.com:443", allowed: true},
		{name: "different host", origin: "https://other.example", host: "example.com", allowed: false},
		{name: "invalid origin", origin: "://bad", host: "example.com", allowed: false},
		{name: "blank request host", origin: "https://example.com", host: " ", allowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWSOriginAllowed(tt.origin, tt.host, tt.policy); got != tt.allowed {
				t.Fatalf("IsWSOriginAllowed() = %v, want %v", got, tt.allowed)
			}
		})
	}
}

func TestWSAddrInTrustedProxies(t *testing.T) {
	prefixes, invalid := ParseWSTrustedProxyList("10.0.0.0/8")
	if len(invalid) > 0 {
		t.Fatalf("invalid trusted proxies = %#v", invalid)
	}

	if !WSAddrInTrustedProxies(netip.MustParseAddr("10.1.2.3"), prefixes) {
		t.Fatal("expected address in trusted proxies")
	}
	if WSAddrInTrustedProxies(netip.MustParseAddr("192.168.1.1"), prefixes) {
		t.Fatal("expected address outside trusted proxies")
	}
	if WSAddrInTrustedProxies(netip.Addr{}, prefixes) {
		t.Fatal("expected invalid address outside trusted proxies")
	}
}
