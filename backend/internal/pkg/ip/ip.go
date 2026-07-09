// Package ip 提供客户端 IP 地址提取工具。
package ip

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClientIP 从 Gin Context 中提取客户端真实 IP 地址。
// 按以下优先级检查 Header：
// 1. CF-Connecting-IP (Cloudflare)
// 2. X-Real-IP (Nginx)
// 3. X-Forwarded-For (取第一个非私有 IP)
// 4. c.ClientIP() (Gin 内置方法)
func GetClientIP(c *gin.Context) string {
	// 1. Cloudflare
	if ip := c.GetHeader("CF-Connecting-IP"); ip != "" {
		return normalizeIP(ip)
	}

	// 2. Nginx X-Real-IP
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return normalizeIP(ip)
	}

	// 3. X-Forwarded-For (多个 IP 时取第一个公网 IP)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" && !isPrivateIP(ip) {
				return normalizeIP(ip)
			}
		}
		// 如果都是私有 IP，返回第一个
		if len(ips) > 0 {
			return normalizeIP(strings.TrimSpace(ips[0]))
		}
	}

	// 4. Gin 内置方法
	return normalizeIP(c.ClientIP())
}

// GetTrustedClientIP 从 Gin 的可信代理解析链提取客户端 IP。
// 该方法依赖 gin.Engine.SetTrustedProxies 配置，不会优先直接信任原始转发头值。
// 适用于 ACL / 风控等安全敏感场景。
func GetTrustedClientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return normalizeIP(c.ClientIP())
}

// cloudflareNets 是 Cloudflare 公开的回源 CIDR 段（https://www.cloudflare.com/ips/）。
// 仅当解析出的直连跳板落在这些段内时，才认为 CF-Connecting-IP 由 Cloudflare 注入、可信。
var cloudflareNets []*net.IPNet

func init() {
	// Cloudflare IPv4 + IPv6 段（官方公开，变动极少；如有更新需同步此处）。
	for _, cidr := range []string{
		"173.245.48.0/20",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"141.101.64.0/18",
		"108.162.192.0/18",
		"190.93.240.0/20",
		"188.114.96.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"162.158.0.0/15",
		"104.16.0.0/13",
		"104.24.0.0/14",
		"172.64.0.0/13",
		"131.0.72.0/22",
		"2400:cb00::/32",
		"2606:4700::/32",
		"2803:f800::/32",
		"2405:b500::/32",
		"2405:8100::/32",
		"2a06:98c0::/29",
		"2c0f:f248::/32",
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic("invalid Cloudflare CIDR: " + cidr)
		}
		cloudflareNets = append(cloudflareNets, block)
	}
}

// isCloudflareIP 报告 ip 是否落在 Cloudflare 回源段内。
func isCloudflareIP(ipStr string) bool {
	parsed := net.ParseIP(ipStr)
	if parsed == nil {
		return false
	}
	for _, block := range cloudflareNets {
		if block.Contains(parsed) {
			return true
		}
	}
	return false
}

// GetClientIPForRisk 提取用于风控的客户端 IP，抵抗 CF-Connecting-IP 伪造。
//
// 生产部署恒在 Cloudflare 之后（用户 → CF → Caddy → 后端）。Caddy 透传 CF-Connecting-IP
// 但将其 X-Real-IP/X-Forwarded-For 设为 CF 节点 IP，故：
//   - 若解析出的直连跳板（c.ClientIP()，跳过可信反代后）落在 Cloudflare 段 → 请求确经 CF，
//     信任 CF-Connecting-IP 取真实用户 IP。
//   - 否则（直连 Caddy、或攻击者直连后端伪造 CF 头）→ 不信任 CF-Connecting-IP，
//     回退到 c.ClientIP() 的可信链解析值。
//
// 这使得攻击者无法通过伪造 CF-Connecting-IP 绕过 IP 维度限额：非 CF 来源的伪造头被忽略，
// 返回的是其真实直连 IP。依赖 server.trusted_proxies 配置反代段以正确解析 c.ClientIP()。
func GetClientIPForRisk(c *gin.Context) string {
	if c == nil {
		return ""
	}
	resolved := normalizeIP(c.ClientIP())
	if isCloudflareIP(resolved) {
		if cf := normalizeIP(c.GetHeader("CF-Connecting-IP")); cf != "" {
			return cf
		}
	}
	return resolved
}

// normalizeIP 规范化 IP 地址，去除端口号和空格。
func normalizeIP(ip string) string {
	ip = strings.TrimSpace(ip)
	// 移除端口号（如 "192.168.1.1:8080" -> "192.168.1.1"）
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}
	return ip
}

// privateNets 预编译私有 IP CIDR 块，避免每次调用 isPrivateIP 时重复解析
var privateNets []*net.IPNet

// CompiledIPRules 表示预编译的 IP 匹配规则。
// PatternCount 记录原始规则数量，用于保留“规则存在但全无效”时的行为语义。
type CompiledIPRules struct {
	CIDRs        []*net.IPNet
	IPs          []net.IP
	PatternCount int
}

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic("invalid CIDR: " + cidr)
		}
		privateNets = append(privateNets, block)
	}
}

// CompileIPRules 将 IP/CIDR 字符串规则预编译为可复用结构。
// 非法规则会被忽略，但 PatternCount 会保留原始规则条数。
func CompileIPRules(patterns []string) *CompiledIPRules {
	compiled := &CompiledIPRules{
		CIDRs:        make([]*net.IPNet, 0, len(patterns)),
		IPs:          make([]net.IP, 0, len(patterns)),
		PatternCount: len(patterns),
	}
	for _, pattern := range patterns {
		normalized := strings.TrimSpace(pattern)
		if normalized == "" {
			continue
		}
		if strings.Contains(normalized, "/") {
			_, cidr, err := net.ParseCIDR(normalized)
			if err != nil || cidr == nil {
				continue
			}
			compiled.CIDRs = append(compiled.CIDRs, cidr)
			continue
		}
		parsedIP := net.ParseIP(normalized)
		if parsedIP == nil {
			continue
		}
		compiled.IPs = append(compiled.IPs, parsedIP)
	}
	return compiled
}

func matchesCompiledRules(parsedIP net.IP, rules *CompiledIPRules) bool {
	if parsedIP == nil || rules == nil {
		return false
	}
	for _, cidr := range rules.CIDRs {
		if cidr.Contains(parsedIP) {
			return true
		}
	}
	for _, ruleIP := range rules.IPs {
		if parsedIP.Equal(ruleIP) {
			return true
		}
	}
	return false
}

// isPrivateIP 检查 IP 是否为私有地址。
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, block := range privateNets {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// MatchesPattern 检查 IP 是否匹配指定的模式（支持单个 IP 或 CIDR）。
// pattern 可以是：
// - 单个 IP: "192.168.1.100"
// - CIDR 范围: "192.168.1.0/24"
func MatchesPattern(clientIP, pattern string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	// 尝试解析为 CIDR
	if strings.Contains(pattern, "/") {
		_, cidr, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		return cidr.Contains(ip)
	}

	// 作为单个 IP 处理
	patternIP := net.ParseIP(pattern)
	if patternIP == nil {
		return false
	}
	return ip.Equal(patternIP)
}

// MatchesAnyPattern 检查 IP 是否匹配任意一个模式。
func MatchesAnyPattern(clientIP string, patterns []string) bool {
	for _, pattern := range patterns {
		if MatchesPattern(clientIP, pattern) {
			return true
		}
	}
	return false
}

// CheckIPRestriction 检查 IP 是否被 API Key 的 IP 限制允许。
// 返回值：(是否允许, 拒绝原因)
// 逻辑：
// 1. 先检查黑名单，如果在黑名单中则直接拒绝
// 2. 如果白名单不为空，IP 必须在白名单中
// 3. 如果白名单为空，允许访问（除非被黑名单拒绝）
func CheckIPRestriction(clientIP string, whitelist, blacklist []string) (bool, string) {
	return CheckIPRestrictionWithCompiledRules(
		clientIP,
		CompileIPRules(whitelist),
		CompileIPRules(blacklist),
	)
}

// CheckIPRestrictionWithCompiledRules 使用预编译规则检查 IP 是否允许访问。
func CheckIPRestrictionWithCompiledRules(clientIP string, whitelist, blacklist *CompiledIPRules) (bool, string) {
	// 规范化 IP
	clientIP = normalizeIP(clientIP)
	if clientIP == "" {
		return false, "access denied"
	}
	parsedIP := net.ParseIP(clientIP)
	if parsedIP == nil {
		return false, "access denied"
	}

	// 1. 检查黑名单
	if blacklist != nil && blacklist.PatternCount > 0 && matchesCompiledRules(parsedIP, blacklist) {
		return false, "access denied"
	}

	// 2. 检查白名单（如果设置了白名单，IP 必须在其中）
	if whitelist != nil && whitelist.PatternCount > 0 && !matchesCompiledRules(parsedIP, whitelist) {
		return false, "access denied"
	}

	return true, ""
}

// ValidateIPPattern 验证 IP 或 CIDR 格式是否有效。
func ValidateIPPattern(pattern string) bool {
	if strings.Contains(pattern, "/") {
		_, _, err := net.ParseCIDR(pattern)
		return err == nil
	}
	return net.ParseIP(pattern) != nil
}

// ValidateIPPatterns 验证多个 IP 或 CIDR 格式。
// 返回无效的模式列表。
func ValidateIPPatterns(patterns []string) []string {
	var invalid []string
	for _, p := range patterns {
		if !ValidateIPPattern(p) {
			invalid = append(invalid, p)
		}
	}
	return invalid
}
