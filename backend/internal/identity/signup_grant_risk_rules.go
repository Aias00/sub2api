package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strings"
)

type SignupGrantRiskInput struct {
	RemoteIP          string
	UserAgent         string
	AcceptLanguage    string
	DeviceFingerprint string
	ProviderType      string
	ProviderSubject   string
}

type SignupGrantRiskConfig struct {
	DomainLimit     int
	FreeDomainLimit int
	FreeDomains     map[string]struct{}
	TrustedDomains  map[string]struct{}
}

func SignupGrantRiskAppliesToSource(signupSource string) bool {
	switch strings.ToLower(strings.TrimSpace(signupSource)) {
	case "email", "github", "google":
		return true
	default:
		return false
	}
}

func SignupGrantRiskHash(salt string, value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(salt + "\x00" + value))
	return hex.EncodeToString(sum[:])
}

func NormalizeSignupGrantRiskEmail(raw string) (string, string) {
	local, domain, ok := SplitEmailForPolicy(raw)
	if !ok {
		return strings.ToLower(strings.TrimSpace(raw)), ""
	}
	if beforePlus, _, found := strings.Cut(local, "+"); found {
		local = beforePlus
	}
	if domain == "googlemail.com" {
		domain = "gmail.com"
	}
	if domain == "gmail.com" {
		local = strings.ReplaceAll(local, ".", "")
	}
	if local == "" {
		return "", domain
	}
	return local + "@" + domain, domain
}

func NormalizeSignupGrantRiskIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(raw); err == nil {
		raw = host
	}
	ip := net.ParseIP(raw)
	if ip == nil {
		return ""
	}
	return ip.String()
}

func NormalizeSignupGrantRiskLabel(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func SignupGrantDeviceSignal(input SignupGrantRiskInput) string {
	if fp := strings.TrimSpace(input.DeviceFingerprint); fp != "" {
		return fp
	}
	return strings.Join([]string{
		strings.TrimSpace(input.UserAgent),
		strings.TrimSpace(input.AcceptLanguage),
	}, "\x00")
}

func (cfg SignupGrantRiskConfig) DomainLimitFor(domain string) int {
	domain = NormalizeSignupGrantRiskDomain(domain)
	if domain == "" {
		return cfg.DomainLimit
	}
	if _, trusted := cfg.TrustedDomains[domain]; trusted {
		return 0
	}
	if _, free := cfg.FreeDomains[domain]; free && cfg.FreeDomainLimit > 0 {
		return cfg.FreeDomainLimit
	}
	return cfg.DomainLimit
}

func NormalizeSignupGrantRiskDomain(raw string) string {
	domain := strings.ToLower(strings.TrimSpace(raw))
	domain = strings.TrimPrefix(domain, "@")
	domain = strings.TrimPrefix(domain, "*.")
	return domain
}

func ParseSignupGrantRiskDomainSet(raw string) map[string]struct{} {
	normalized := NormalizeDomainListSetting(raw)
	out := map[string]struct{}{}
	for _, domain := range strings.Split(normalized, ",") {
		domain = NormalizeSignupGrantRiskDomain(domain)
		if domain != "" {
			out[domain] = struct{}{}
		}
	}
	return out
}

func NormalizeDomainListSetting(raw string) string {
	replacer := strings.NewReplacer("\n", ",", "\r", ",", "\t", ",", " ", ",", ";", ",", "；", ",", "，", ",")
	parts := strings.Split(replacer.Replace(strings.TrimSpace(raw)), ",")
	seen := map[string]struct{}{}
	domains := make([]string, 0, len(parts))
	for _, part := range parts {
		domain := NormalizeSignupGrantRiskDomain(part)
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		domains = append(domains, domain)
	}
	return strings.Join(domains, ",")
}

func SplitEmailForPolicy(raw string) (local string, domain string, ok bool) {
	email := strings.ToLower(strings.TrimSpace(raw))
	local, domain, found := strings.Cut(email, "@")
	if !found || local == "" || domain == "" || strings.Contains(domain, "@") {
		return "", "", false
	}
	return local, domain, true
}
