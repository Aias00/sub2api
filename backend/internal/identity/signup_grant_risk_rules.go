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

// defaultSignupGrantRiskIPv6PrefixBits 是 IPv6 注册赠金风控聚合的默认前缀长度。
// 现代终端默认开启 IPv6 隐私扩展（RFC 4941），接口标识符（后 64 位）会定期轮换，
// 用完整 128 位做精确等值匹配会被同一台设备的轮换地址绕过。截断到 /64 前缀后，
// 同一条宽带/同一家庭网段内仍能聚合，对反滥用更有意义。
const defaultSignupGrantRiskIPv6PrefixBits = 64

// NormalizeSignupGrantRiskIPForHash 按 ipv6PrefixBits 将 IPv6 截断到前缀后规范化，
// 供注册赠金风控的 IP 哈希（限额 / advisory lock 锁键 / override）使用。
//
//   - IPv4 透传，不截断（地址空间小，全量匹配是合理的）。
//   - IPv6 按 ipv6PrefixBits 掩码到前缀；前缀外位清零后以 ip.String() 输出，
//     零段会被压缩（如 2408:8215:5413:4700:cd91:7fd7:23c8:c71 → 2408:8215:5413:4700::）。
//   - ipv6PrefixBits <= 0 视为禁用截断，保留全量（等价于 NormalizeSignupGrantRiskIP）。
//   - ipv6PrefixBits 超出 [0,128] 时回退到默认 /64。
//   - 非法 IP 返回空字符串。
//
// 注意：此函数仅供风控哈希聚合使用。审计落库（user_registration_events.ip_address）
// 应保留全量 IP，使用 NormalizeSignupGrantRiskIP。
func NormalizeSignupGrantRiskIPForHash(raw string, ipv6PrefixBits int) string {
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
	// IPv4（含 IPv4-mapped IPv6 的 4in6 表示）透传，不截断。
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	bits := ipv6PrefixBits
	if bits < 0 || bits > 128 {
		bits = defaultSignupGrantRiskIPv6PrefixBits
	}
	if bits == 0 {
		return ip.String()
	}
	mask := net.CIDRMask(bits, 128)
	return ip.Mask(mask).String()
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
