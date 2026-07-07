package service

import (
	"context"
	"encoding/json"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Aias00/cloudbase/internal/pkg/logger"
)

func (s *AuthService) recordUserRegistrationEvent(ctx context.Context, user *User, signupSource string) {
	if s == nil || user == nil || user.ID <= 0 || s.entClient == nil {
		return
	}
	db := s.signupGrantRiskDB()
	if db == nil {
		return
	}
	input := signupGrantRiskInputFromContext(ctx)
	headersJSON := registrationHeadersJSON(input.HeaderSnapshot)
	providerType := strings.TrimSpace(input.ProviderType)
	providerSubject := strings.TrimSpace(input.ProviderSubject)
	source := strings.TrimSpace(signupSource)
	if source == "" {
		source = strings.TrimSpace(user.SignupSource)
	}

	query := `
INSERT INTO user_registration_events (
  user_id, email, signup_source, provider_type, provider_subject,
  ip_address, user_agent, accept_language, device_fingerprint, headers_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
ON CONFLICT (user_id) DO UPDATE SET
  email = EXCLUDED.email,
  signup_source = EXCLUDED.signup_source,
  provider_type = EXCLUDED.provider_type,
  provider_subject = EXCLUDED.provider_subject,
  ip_address = EXCLUDED.ip_address,
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
		normalizeSignupGrantRiskIP(input.RemoteIP),
		trimRegistrationContextValue(input.UserAgent),
		trimRegistrationContextValue(input.AcceptLanguage),
		trimRegistrationContextValue(input.DeviceFingerprint),
		headersJSON,
	}
	if authServiceDialect(s) != dialect.Postgres {
		query = `
INSERT INTO user_registration_events (
  user_id, email, signup_source, provider_type, provider_subject,
  ip_address, user_agent, accept_language, device_fingerprint, headers_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (user_id) DO UPDATE SET
  email = excluded.email,
  signup_source = excluded.signup_source,
  provider_type = excluded.provider_type,
  provider_subject = excluded.provider_subject,
  ip_address = excluded.ip_address,
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

func authServiceDialect(s *AuthService) string {
	if s == nil || s.entClient == nil || s.entClient.Driver() == nil {
		return dialect.Postgres
	}
	if driver, ok := s.entClient.Driver().(*entsql.Driver); ok && driver != nil {
		return driver.Dialect()
	}
	return dialect.Postgres
}
