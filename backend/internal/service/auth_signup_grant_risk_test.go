package service

import "testing"

func TestSignupGrantRiskAppliesToEmailSourcesOnly(t *testing.T) {
	allowed := []string{"email", "touch", "github", "google"}
	for _, source := range allowed {
		if !signupGrantRiskAppliesToSource(source) {
			t.Fatalf("source %q should be eligible for signup grant risk control", source)
		}
	}

	blocked := []string{"", "linuxdo", "wechat", "oidc", "dingtalk"}
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
