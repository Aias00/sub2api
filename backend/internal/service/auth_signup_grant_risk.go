package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	signupGrantRiskDecisionAllowed = "allowed"
	signupGrantRiskDecisionBlocked = "blocked"
	signupGrantRiskWindow          = 24 * time.Hour
)

type signupGrantRiskContextKey struct{}

type SignupGrantRiskInput struct {
	RemoteIP        string
	UserAgent       string
	ProviderType    string
	ProviderSubject string
}

type signupGrantRiskClaim struct {
	ID       int64
	Blocked  bool
	Reason   string
	Decision string
}

type signupGrantRiskConfig struct {
	Enabled     bool
	EmailLimit  int
	IPLimit     int
	DomainLimit int
}

func WithSignupGrantRiskInput(ctx context.Context, input SignupGrantRiskInput) context.Context {
	return context.WithValue(ctx, signupGrantRiskContextKey{}, input)
}

func signupGrantRiskInputFromContext(ctx context.Context) SignupGrantRiskInput {
	if ctx == nil {
		return SignupGrantRiskInput{}
	}
	if input, ok := ctx.Value(signupGrantRiskContextKey{}).(SignupGrantRiskInput); ok {
		return input
	}
	return SignupGrantRiskInput{}
}

func mergeSignupGrantRiskInput(base SignupGrantRiskInput, patch SignupGrantRiskInput) SignupGrantRiskInput {
	if strings.TrimSpace(patch.RemoteIP) != "" {
		base.RemoteIP = patch.RemoteIP
	}
	if strings.TrimSpace(patch.UserAgent) != "" {
		base.UserAgent = patch.UserAgent
	}
	if strings.TrimSpace(patch.ProviderType) != "" {
		base.ProviderType = patch.ProviderType
	}
	if strings.TrimSpace(patch.ProviderSubject) != "" {
		base.ProviderSubject = patch.ProviderSubject
	}
	return base
}

func (s *AuthService) applySignupGrantRiskControl(ctx context.Context, email, signupSource string, plan signupGrantPlan) (signupGrantPlan, *signupGrantRiskClaim) {
	if s == nil || s.settingService == nil || !signupGrantRiskAppliesToSource(signupSource) || !signupGrantPlanHasBonus(plan) {
		return plan, nil
	}
	cfg := s.settingService.signupGrantRiskConfig(ctx)
	if !cfg.Enabled {
		return plan, nil
	}

	db := s.signupGrantRiskDB()
	if db == nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] enabled but database unavailable; stripping signup grant email=%s source=%s", anonymizeEmailForLog(email), signupSource)
		return stripSignupGrantBonus(plan), &signupGrantRiskClaim{
			Blocked:  true,
			Reason:   "risk_check_unavailable",
			Decision: signupGrantRiskDecisionBlocked,
		}
	}

	input := signupGrantRiskInputFromContext(ctx)
	normalizedEmail, domain := normalizeSignupGrantRiskEmail(email)
	remoteIP := normalizeSignupGrantRiskIP(input.RemoteIP)
	now := time.Now().UTC()
	hashSalt := s.signupGrantRiskSalt()
	emailHash := signupGrantRiskHash(hashSalt, normalizedEmail)
	domainHash := signupGrantRiskHash(hashSalt, domain)
	ipHash := signupGrantRiskHash(hashSalt, remoteIP)
	userAgentHash := signupGrantRiskHash(hashSalt, strings.TrimSpace(input.UserAgent))
	providerSubjectHash := signupGrantRiskHash(hashSalt, strings.TrimSpace(input.ProviderSubject))

	reason := ""
	if cfg.EmailLimit > 0 && emailHash != "" {
		count, err := countSignupGrantClaims(ctx, db, "email_hash = $1", []any{emailHash})
		if err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] email count failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
		if count >= cfg.EmailLimit {
			reason = "email_limit"
		}
	}
	if reason == "" && cfg.IPLimit > 0 && ipHash != "" {
		count, err := countSignupGrantClaims(ctx, db, "ip_hash = $1 AND created_at >= $2", []any{ipHash, now.Add(-signupGrantRiskWindow)})
		if err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] ip count failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
		if count >= cfg.IPLimit {
			reason = "ip_daily_limit"
		}
	}
	if reason == "" && cfg.DomainLimit > 0 && domainHash != "" {
		count, err := countSignupGrantClaims(ctx, db, "email_domain_hash = $1 AND created_at >= $2", []any{domainHash, now.Add(-signupGrantRiskWindow)})
		if err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] domain count failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
		if count >= cfg.DomainLimit {
			reason = "email_domain_daily_limit"
		}
	}

	decision := signupGrantRiskDecisionAllowed
	filteredPlan := plan
	if reason != "" {
		decision = signupGrantRiskDecisionBlocked
		filteredPlan = stripSignupGrantBonus(plan)
	}
	claimID, err := insertSignupGrantClaim(ctx, db, signupGrantClaimInsert{
		EmailHash:           emailHash,
		EmailDomainHash:     domainHash,
		IPHash:              ipHash,
		UserAgentHash:       userAgentHash,
		SignupSource:        normalizeSignupGrantRiskLabel(signupSource),
		ProviderType:        normalizeSignupGrantRiskLabel(input.ProviderType),
		ProviderSubjectHash: providerSubjectHash,
		Decision:            decision,
		Reason:              reason,
		GrantBalance:        plan.Balance,
		GrantMetadataJSON:   signupGrantMetadataJSON(plan),
	})
	if err != nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] insert claim failed: %v", err)
		return stripSignupGrantBonus(plan), nil
	}
	if reason != "" {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] stripped signup grant email=%s source=%s reason=%s", anonymizeEmailForLog(email), signupSource, reason)
	}
	return filteredPlan, &signupGrantRiskClaim{
		ID:       claimID,
		Blocked:  reason != "",
		Reason:   reason,
		Decision: decision,
	}
}

func (s *AuthService) attachSignupGrantClaim(ctx context.Context, claim *signupGrantRiskClaim, userID int64) {
	if s == nil || claim == nil || claim.ID <= 0 || userID <= 0 {
		return
	}
	db := s.signupGrantRiskDB()
	if db == nil {
		return
	}
	if _, err := db.ExecContext(ctx, `UPDATE signup_grant_claims SET user_id = $1 WHERE id = $2`, userID, claim.ID); err != nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] attach claim failed: claim_id=%d user_id=%d err=%v", claim.ID, userID, err)
	}
}

func (s *AuthService) resolveSignupGrantPlanForFinalizedUser(ctx context.Context, userID int64, email, signupSource string) (signupGrantPlan, *signupGrantRiskClaim) {
	plan := s.resolveSignupGrantPlan(ctx, signupSource)
	if s == nil || userID <= 0 || s.settingService == nil || !signupGrantPlanHasBonus(plan) {
		return plan, nil
	}
	cfg := s.settingService.signupGrantRiskConfig(ctx)
	if !cfg.Enabled {
		return plan, nil
	}
	db := s.signupGrantRiskDB()
	if db == nil {
		return stripSignupGrantBonus(plan), nil
	}
	claim, ok := latestSignupGrantClaimForUser(ctx, db, userID)
	if ok {
		if claim.Decision == signupGrantRiskDecisionBlocked {
			return stripSignupGrantBonus(plan), claim
		}
		return plan, claim
	}
	return s.applySignupGrantRiskControl(ctx, email, signupSource, plan)
}

func (s *AuthService) signupGrantRiskDB() *sql.DB {
	if s == nil || s.entClient == nil {
		return nil
	}
	driver, ok := s.entClient.Driver().(*entsql.Driver)
	if !ok || driver == nil {
		return nil
	}
	return driver.DB()
}

func latestSignupGrantClaimForUser(ctx context.Context, db *sql.DB, userID int64) (*signupGrantRiskClaim, bool) {
	claim := signupGrantRiskClaim{}
	err := db.QueryRowContext(ctx, `
SELECT id, decision, reason
FROM signup_grant_claims
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1
`, userID).Scan(&claim.ID, &claim.Decision, &claim.Reason)
	if err != nil {
		return nil, false
	}
	claim.Blocked = claim.Decision == signupGrantRiskDecisionBlocked
	return &claim, true
}

func (s *AuthService) signupGrantRiskSalt() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.JWT.Secret) != "" {
		return s.cfg.JWT.Secret
	}
	return "sub2api-signup-grant-risk"
}

func signupGrantPlanHasBonus(plan signupGrantPlan) bool {
	return plan.Balance > 0 || len(plan.Subscriptions) > 0 || len(plan.PlatformQuotas) > 0
}

func signupGrantRiskAppliesToSource(signupSource string) bool {
	switch strings.ToLower(strings.TrimSpace(signupSource)) {
	case authSignupSourceEmail, authSignupSourceTouch, "github", "google":
		return true
	default:
		return false
	}
}

func stripSignupGrantBonus(plan signupGrantPlan) signupGrantPlan {
	return signupGrantPlan{Concurrency: plan.Concurrency}
}

func signupGrantRiskHash(salt string, value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(salt + "\x00" + value))
	return hex.EncodeToString(sum[:])
}

func normalizeSignupGrantRiskEmail(raw string) (string, string) {
	local, domain, ok := splitEmailForPolicy(raw)
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

func normalizeSignupGrantRiskIP(raw string) string {
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

func normalizeSignupGrantRiskLabel(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

type signupGrantClaimInsert struct {
	EmailHash           string
	EmailDomainHash     string
	IPHash              string
	UserAgentHash       string
	SignupSource        string
	ProviderType        string
	ProviderSubjectHash string
	Decision            string
	Reason              string
	GrantBalance        float64
	GrantMetadataJSON   string
}

func countSignupGrantClaims(ctx context.Context, db *sql.DB, predicate string, args []any) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM signup_grant_claims WHERE decision = 'allowed' AND %s`, predicate)
	var count int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func insertSignupGrantClaim(ctx context.Context, db *sql.DB, input signupGrantClaimInsert) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, `
INSERT INTO signup_grant_claims (
    email_hash, email_domain_hash, ip_hash, user_agent_hash,
    signup_source, provider_type, provider_subject_hash,
    decision, reason, grant_balance, grant_metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
RETURNING id
`,
		input.EmailHash,
		input.EmailDomainHash,
		input.IPHash,
		input.UserAgentHash,
		input.SignupSource,
		input.ProviderType,
		input.ProviderSubjectHash,
		input.Decision,
		input.Reason,
		input.GrantBalance,
		input.GrantMetadataJSON,
	).Scan(&id)
	return id, err
}

func signupGrantMetadataJSON(plan signupGrantPlan) string {
	quotaKeys := make([]string, 0, len(plan.PlatformQuotas))
	for key := range plan.PlatformQuotas {
		quotaKeys = append(quotaKeys, key)
	}
	payload := map[string]any{
		"subscriptions":   len(plan.Subscriptions),
		"platform_quotas": quotaKeys,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(buf)
}

func (s *SettingService) signupGrantRiskConfig(ctx context.Context) signupGrantRiskConfig {
	if s == nil || s.settingRepo == nil {
		return signupGrantRiskConfig{}
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeySignupGrantRiskControlEnabled,
		SettingKeySignupGrantRiskControlEmailLimit,
		SettingKeySignupGrantRiskControlIPLimit,
		SettingKeySignupGrantRiskControlDomainLimit,
	})
	if err != nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] load config failed: %v", err)
		return signupGrantRiskConfig{}
	}
	return signupGrantRiskConfig{
		Enabled:     values[SettingKeySignupGrantRiskControlEnabled] == "true",
		EmailLimit:  parseNonNegativeIntSetting(values[SettingKeySignupGrantRiskControlEmailLimit], 1),
		IPLimit:     parseNonNegativeIntSetting(values[SettingKeySignupGrantRiskControlIPLimit], 3),
		DomainLimit: parseNonNegativeIntSetting(values[SettingKeySignupGrantRiskControlDomainLimit], 10),
	}
}

func parseNonNegativeIntSetting(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func nonNegativeInt(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func anonymizeEmailForLog(email string) string {
	local, domain, ok := splitEmailForPolicy(email)
	if !ok {
		return "<invalid>"
	}
	if len(local) <= 2 {
		return "**@" + domain
	}
	return local[:1] + "***" + local[len(local)-1:] + "@" + domain
}
