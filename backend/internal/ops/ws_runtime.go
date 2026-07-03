package ops

import (
	"math"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

const (
	WSOriginPolicyStrict     = "strict"
	WSOriginPolicyPermissive = "permissive"
)

type WSProxyConfig struct {
	TrustProxy     bool
	TrustedProxies []netip.Prefix
	OriginPolicy   string
}

type WSProxyConfigInput struct {
	TrustProxyRaw     string
	TrustedProxiesRaw string
	OriginPolicyRaw   string
	DefaultTrustProxy bool
	DefaultProxies    []netip.Prefix
	DefaultPolicy     string
}

type WSProxyConfigInvalid struct {
	TrustProxy     bool
	TrustedProxies []string
	OriginPolicy   bool
}

type WSRuntimeLimits struct {
	MaxConns      int32
	MaxConnsPerIP int32
}

type WSRuntimeLimitsInput struct {
	MaxConnsRaw          string
	MaxConnsPerIPRaw     string
	DefaultMaxConns      int32
	DefaultMaxConnsPerIP int32
}

type WSRuntimeLimitsInvalid struct {
	MaxConns      bool
	MaxConnsPerIP bool
}

func RoundTo1DP(v float64) float64 {
	return math.Round(v*10) / 10
}

func WSClientIPSlotKey(raw string) (string, bool) {
	key := strings.TrimSpace(raw)
	if key == "" {
		return "", false
	}
	return key, true
}

func ParseWSProxyConfig(input WSProxyConfigInput) (WSProxyConfig, WSProxyConfigInvalid) {
	cfg := WSProxyConfig{
		TrustProxy:     input.DefaultTrustProxy,
		TrustedProxies: input.DefaultProxies,
		OriginPolicy:   input.DefaultPolicy,
	}
	var invalid WSProxyConfigInvalid

	if raw := strings.TrimSpace(input.TrustProxyRaw); raw != "" {
		if parsed, ok := ParseWSBoolFlag(raw); ok {
			cfg.TrustProxy = parsed
		} else {
			invalid.TrustProxy = true
		}
	}

	if raw := strings.TrimSpace(input.TrustedProxiesRaw); raw != "" {
		prefixes, invalidEntries := ParseWSTrustedProxyList(raw)
		cfg.TrustedProxies = prefixes
		invalid.TrustedProxies = invalidEntries
	}

	if raw := strings.TrimSpace(input.OriginPolicyRaw); raw != "" {
		if parsed, ok := ParseWSOriginPolicy(raw); ok {
			cfg.OriginPolicy = parsed
		} else {
			invalid.OriginPolicy = true
		}
	}

	return cfg, invalid
}

func ParseWSRuntimeLimits(input WSRuntimeLimitsInput) (WSRuntimeLimits, WSRuntimeLimitsInvalid) {
	cfg := WSRuntimeLimits{
		MaxConns:      input.DefaultMaxConns,
		MaxConnsPerIP: input.DefaultMaxConnsPerIP,
	}
	var invalid WSRuntimeLimitsInvalid

	if raw := strings.TrimSpace(input.MaxConnsRaw); raw != "" {
		if parsed, ok := ParseWSPositiveLimit(raw); ok {
			cfg.MaxConns = parsed
		} else {
			invalid.MaxConns = true
		}
	}

	if raw := strings.TrimSpace(input.MaxConnsPerIPRaw); raw != "" {
		if parsed, ok := ParseWSNonNegativeLimit(raw); ok {
			cfg.MaxConnsPerIP = parsed
		} else {
			invalid.MaxConnsPerIP = true
		}
	}

	return cfg, invalid
}

func ParseWSOriginPolicy(raw string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case WSOriginPolicyStrict, WSOriginPolicyPermissive:
		return normalized, true
	default:
		return "", false
	}
}

func ParseWSBoolFlag(raw string) (bool, bool) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return false, false
	}
	return parsed, true
}

func ParseWSPositiveLimit(raw string) (int32, bool) {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return int32(parsed), true
}

func ParseWSNonNegativeLimit(raw string) (int32, bool) {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed < 0 {
		return 0, false
	}
	return int32(parsed), true
}

func ParseWSTrustedProxyList(raw string) (prefixes []netip.Prefix, invalid []string) {
	for _, token := range strings.Split(raw, ",") {
		item := strings.TrimSpace(token)
		if item == "" {
			continue
		}

		var (
			p   netip.Prefix
			err error
		)
		if strings.Contains(item, "/") {
			p, err = netip.ParsePrefix(item)
		} else {
			var addr netip.Addr
			addr, err = netip.ParseAddr(item)
			if err == nil {
				addr = addr.Unmap()
				bits := 128
				if addr.Is4() {
					bits = 32
				}
				p = netip.PrefixFrom(addr, bits)
			}
		}

		if err != nil || !p.IsValid() {
			invalid = append(invalid, item)
			continue
		}

		prefixes = append(prefixes, p.Masked())
	}
	return prefixes, invalid
}

func WSHostWithoutPort(hostport string) string {
	hostport = strings.TrimSpace(hostport)
	if hostport == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return host
	}
	if strings.HasPrefix(hostport, "[") && strings.HasSuffix(hostport, "]") {
		return strings.Trim(hostport, "[]")
	}
	parts := strings.Split(hostport, ":")
	return parts[0]
}

func ParseWSPeerAddr(remoteAddr string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		host = strings.TrimSpace(remoteAddr)
	}
	addr, ok := parseWSAddrToken(host)
	if !ok {
		return netip.Addr{}, false
	}
	return addr, true
}

func ParseWSForwardedForClientIP(raw string) (string, bool) {
	xff := strings.TrimSpace(raw)
	if xff == "" {
		return "", false
	}
	xff = strings.TrimSpace(strings.Split(xff, ",")[0])
	addr, ok := parseWSAddrToken(xff)
	if !ok {
		return "", false
	}
	return addr.String(), true
}

func ParseWSForwardedHost(raw string) (string, bool) {
	host := strings.TrimSpace(raw)
	if host == "" {
		return "", false
	}
	host = strings.TrimSpace(strings.Split(host, ",")[0])
	if host == "" {
		return "", false
	}
	return host, true
}

func parseWSAddrToken(raw string) (netip.Addr, bool) {
	host := strings.TrimSpace(raw)
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	if host == "" {
		return netip.Addr{}, false
	}
	addr, err := netip.ParseAddr(host)
	if err != nil || !addr.IsValid() {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

func IsWSOriginAllowed(originRaw string, requestHostRaw string, policy string) bool {
	origin := strings.TrimSpace(originRaw)
	if origin == "" {
		normalized, ok := ParseWSOriginPolicy(policy)
		if ok && normalized == WSOriginPolicyStrict {
			return false
		}
		return true
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Hostname() == "" {
		return false
	}

	requestHost := strings.ToLower(WSHostWithoutPort(requestHostRaw))
	if requestHost == "" {
		return false
	}
	return strings.ToLower(parsed.Hostname()) == requestHost
}

func WSAddrInTrustedProxies(addr netip.Addr, trusted []netip.Prefix) bool {
	if !addr.IsValid() {
		return false
	}
	for _, p := range trusted {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}
