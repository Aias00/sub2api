package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/internal/pkg/logger"
)

// recordUserRegistrationEvent 把一次注册上下文写入 user_registration_events。
// 邮箱注册与 OAuth 注册路径复用此方法，客户端 IP 由 handler 通过
// signupGrantRiskContext 注入 context，此处从 context 取出。
func (s *AuthService) recordUserRegistrationEvent(ctx context.Context, user *User, signupSource string) {
	if s == nil || user == nil || user.ID <= 0 || s.entClient == nil {
		return
	}
	prefixBits := defaultSignupGrantRiskIPv6PrefixBits
	if s.settingService != nil {
		prefixBits = s.settingService.signupGrantRiskConfig(ctx).IPv6PrefixBits
	}
	insertUserRegistrationEvent(ctx, s.entClient, user, signupSource, signupGrantRiskInputFromContext(ctx), prefixBits)
}

// insertUserRegistrationEvent 把一次注册上下文写入 user_registration_events（ON CONFLICT 覆盖语义）。
// 供所有用户创建路径复用：entClient 为 nil 或取不到底层 *sql.DB 时静默跳过（与原 no-op 语义一致）。
// 后台建号等非自助路径可传入空 SignupGrantRiskInput，IP 留空以避免污染按 IP 聚合的风控信号
// （洞察查询 WHERE ip_address <> ” 会天然排除空 IP 行）。
func insertUserRegistrationEvent(ctx context.Context, entClient *dbent.Client, user *User, signupSource string, input SignupGrantRiskInput, ipv6PrefixBits int) {
	if entClient == nil || user == nil || user.ID <= 0 {
		return
	}
	db := signupGrantRiskDBFromClient(entClient)
	if db == nil {
		return
	}
	headersJSON := registrationHeadersJSON(input.HeaderSnapshot)
	providerType := strings.TrimSpace(input.ProviderType)
	providerSubject := strings.TrimSpace(input.ProviderSubject)
	source := strings.TrimSpace(signupSource)
	if source == "" {
		source = strings.TrimSpace(user.SignupSource)
	}
	// ip_address 保留全量供审计；ip_prefix 按 cfg 的 IPv6 前缀长度截断后落库，
	// 供 admin 聚合查询（同 IP 注册计数）抵抗 IPv6 隐私扩展轮换。两者口径与应用
	// 侧 normalizeSignupGrantRiskIPForHash 一致，与风控 ip_hash 共用同一前缀语义。
	ipAddress := normalizeSignupGrantRiskIP(input.RemoteIP)
	ipPrefix := normalizeSignupGrantRiskIPForHash(input.RemoteIP, ipv6PrefixBits)

	query := `
INSERT INTO user_registration_events (
  user_id, email, signup_source, provider_type, provider_subject,
  ip_address, ip_prefix, user_agent, accept_language, device_fingerprint, headers_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
ON CONFLICT (user_id) DO UPDATE SET
  email = EXCLUDED.email,
  signup_source = EXCLUDED.signup_source,
  provider_type = EXCLUDED.provider_type,
  provider_subject = EXCLUDED.provider_subject,
  ip_address = EXCLUDED.ip_address,
  ip_prefix = EXCLUDED.ip_prefix,
  user_agent = EXCLUDED.user_agent,
  accept_language = EXCLUDED.accept_language,
  device_fingerprint = EXCLUDED.device_fingerprint,
  headers_json = EXCLUDED.headers_json`
	args := []any{
		user.ID,
		user.Email,
		source,
		providerType,
		providerSubject,
		ipAddress,
		ipPrefix,
		trimRegistrationContextValue(input.UserAgent),
		trimRegistrationContextValue(input.AcceptLanguage),
		trimRegistrationContextValue(input.DeviceFingerprint),
		headersJSON,
	}
	if entClientDialect(entClient) != dialect.Postgres {
		query = `
INSERT INTO user_registration_events (
  user_id, email, signup_source, provider_type, provider_subject,
  ip_address, ip_prefix, user_agent, accept_language, device_fingerprint, headers_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (user_id) DO UPDATE SET
  email = excluded.email,
  signup_source = excluded.signup_source,
  provider_type = excluded.provider_type,
  provider_subject = excluded.provider_subject,
  ip_address = excluded.ip_address,
  ip_prefix = excluded.ip_prefix,
  user_agent = excluded.user_agent,
  accept_language = excluded.accept_language,
  device_fingerprint = excluded.device_fingerprint,
  headers_json = excluded.headers_json`
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		logger.LegacyPrintf("service.auth.registration", "[Auth] record registration event failed: user_id=%d err=%v", user.ID, err)
	}
}

func registrationHeadersJSON(headers map[string]string) string {
	if len(headers) == 0 {
		return "{}"
	}
	clean := make(map[string]string, len(headers))
	for key, value := range headers {
		key = strings.TrimSpace(key)
		value = trimRegistrationContextValue(value)
		if key == "" || value == "" {
			continue
		}
		clean[key] = value
	}
	buf, err := json.Marshal(clean)
	if err != nil {
		return "{}"
	}
	return string(buf)
}

func trimRegistrationContextValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 1024 {
		return value[:1024]
	}
	return value
}

// signupGrantRiskDBFromClient 从 ent 客户端取出底层 *sql.DB，供 user_registration_events 直写。
// 与 (*AuthService).signupGrantRiskDB 同逻辑，使 admin 侧无需 AuthService 即可复用注册事件写入。
func signupGrantRiskDBFromClient(entClient *dbent.Client) *sql.DB {
	if entClient == nil {
		return nil
	}
	driver, ok := entClient.Driver().(*entsql.Driver)
	if !ok || driver == nil {
		return nil
	}
	return driver.DB()
}

// entClientDialect 报告 ent 客户端的底层驱动方言，取不到时默认 PostgreSQL。
func entClientDialect(entClient *dbent.Client) string {
	if entClient == nil || entClient.Driver() == nil {
		return dialect.Postgres
	}
	if driver, ok := entClient.Driver().(*entsql.Driver); ok && driver != nil {
		return driver.Dialect()
	}
	return dialect.Postgres
}
