package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/internal/identity"
	"github.com/Aias00/cloudbase/internal/pkg/logger"
)

const (
	signupGrantRiskDecisionAllowed = "allowed"
	signupGrantRiskDecisionBlocked = "blocked"
	signupGrantRiskWindow          = 24 * time.Hour
)

type signupGrantRiskContextKey struct{}

type SignupGrantRiskInput struct {
	RemoteIP          string
	UserAgent         string
	AcceptLanguage    string
	DeviceFingerprint string
	ProviderType      string
	ProviderSubject   string
	HeaderSnapshot    map[string]string
	// EmailVerified 指示本次注册/绑定的邮箱是否已验证。用指针以区分"未设置"与"false"，
	// 便于 mergeSignupGrantRiskInput 不丢失 false。nil/false 在 RequireVerifiedEmail 规则下视为未验证。
	EmailVerified *bool
}

type signupGrantRiskClaim struct {
	ID           int64
	Blocked      bool
	Reason       string
	Decision     string
	GrantBalance float64
}

type signupGrantRiskConfig struct {
	Enabled              bool
	EmailLimit           int
	IPLimit              int
	DomainLimit          int
	OAuthIdentityEnabled bool
	DeviceEnabled        bool
	DeviceLimit          int
	FreeDomainLimit      int
	BlockedDomains       map[string]struct{}
	FreeDomains          map[string]struct{}
	TrustedDomains       map[string]struct{}
	RequireVerifiedEmail bool
}

type SignupGrantRiskClaimRecord struct {
	ID                  int64     `json:"id"`
	UserID              *int64    `json:"user_id,omitempty"`
	UserPublicID        string    `json:"user_public_id,omitempty"`
	Email               string    `json:"email"`
	EmailDomain         string    `json:"email_domain"`
	IPAddress           string    `json:"ip_address"`
	EmailHash           string    `json:"email_hash"`
	EmailDomainHash     string    `json:"email_domain_hash"`
	IPHash              string    `json:"ip_hash"`
	UserAgentHash       string    `json:"device_hash"`
	SignupSource        string    `json:"signup_source"`
	ProviderType        string    `json:"provider_type"`
	ProviderSubject     string    `json:"provider_subject"`
	ProviderSubjectHash string    `json:"provider_subject_hash"`
	Decision            string    `json:"decision"`
	Reason              string    `json:"reason"`
	GrantBalance        float64   `json:"grant_balance"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type SignupGrantRiskClaimFilter struct {
	Decision     string
	UserID       int64
	SubjectType  string
	SubjectHash  string
	SubjectQuery string
	Reason       string
}

type SignupGrantRiskUserSummary struct {
	UserID       int64      `json:"user_id"`
	HasClaim     bool       `json:"has_claim"`
	Decision     string     `json:"decision,omitempty"`
	Reason       string     `json:"reason,omitempty"`
	GrantBalance float64    `json:"grant_balance,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type SignupGrantRiskOverrideInput struct {
	SubjectType  string `json:"subject_type"`
	Subject      string `json:"subject"`
	SubjectValue string `json:"subject_value,omitempty"`
	Action       string `json:"action"`
	Reason       string `json:"reason"`
	CreatedBy    int64  `json:"created_by,omitempty"`
}

type SignupGrantRiskOverrideRecord struct {
	ID           int64      `json:"id"`
	SubjectType  string     `json:"subject_type"`
	SubjectValue string     `json:"subject_value"`
	SubjectHash  string     `json:"subject_hash"`
	Action       string     `json:"action"`
	Reason       string     `json:"reason"`
	CreatedBy    *int64     `json:"created_by,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SignupGrantRiskOverrideFilter struct {
	SubjectType  string
	Action       string
	SubjectHash  string
	SubjectQuery string
}

type SignupGrantAdminAuditLog struct {
	ID                 int64          `json:"id"`
	Operation          string         `json:"operation"`
	TargetUserID       *int64         `json:"target_user_id,omitempty"`
	TargetUserPublicID string         `json:"target_user_public_id,omitempty"`
	SubjectType        string         `json:"subject_type"`
	SubjectValue       string         `json:"subject_value"`
	SubjectHash        string         `json:"subject_hash"`
	Action             string         `json:"action"`
	Amount             float64        `json:"amount"`
	Reason             string         `json:"reason"`
	AdminID            *int64         `json:"admin_id,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
}

type SignupGrantAdminAuditLogFilter struct {
	Operation    string
	AdminID      int64
	TargetUserID int64
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
	if strings.TrimSpace(patch.AcceptLanguage) != "" {
		base.AcceptLanguage = patch.AcceptLanguage
	}
	if strings.TrimSpace(patch.DeviceFingerprint) != "" {
		base.DeviceFingerprint = patch.DeviceFingerprint
	}
	if strings.TrimSpace(patch.ProviderType) != "" {
		base.ProviderType = patch.ProviderType
	}
	if strings.TrimSpace(patch.ProviderSubject) != "" {
		base.ProviderSubject = patch.ProviderSubject
	}
	if len(patch.HeaderSnapshot) > 0 {
		base.HeaderSnapshot = make(map[string]string, len(patch.HeaderSnapshot))
		for key, value := range patch.HeaderSnapshot {
			if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
				base.HeaderSnapshot[key] = value
			}
		}
	}
	if patch.EmailVerified != nil {
		base.EmailVerified = patch.EmailVerified
	}
	return base
}

// signupGrantEmailVerified 把 bool 包成 *bool，便于在调用点设置 SignupGrantRiskInput.EmailVerified。
func signupGrantEmailVerified(v bool) *bool { return &v }

func (s *AuthService) applySignupGrantRiskControl(ctx context.Context, email, signupSource string, plan signupGrantPlan) (signupGrantPlan, *signupGrantRiskClaim) {
	return s.applySignupGrantRiskControlEx(ctx, email, signupSource, plan, false)
}

// applySignupGrantRiskControlEx 执行注册赠金风控。force=true 时绕过来源白名单
// （signupGrantRiskAppliesToSource），用于首绑奖励等非白名单来源但仍需复用全部限额逻辑的场景。
func (s *AuthService) applySignupGrantRiskControlEx(ctx context.Context, email, signupSource string, plan signupGrantPlan, force bool) (signupGrantPlan, *signupGrantRiskClaim) {
	if s == nil || s.settingService == nil || !signupGrantPlanHasBonus(plan) {
		return plan, nil
	}
	if !force && !signupGrantRiskAppliesToSource(signupSource) {
		return plan, nil
	}
	cfg := s.settingService.signupGrantRiskConfig(ctx)
	input := signupGrantRiskInputFromContext(ctx)

	// RequireVerifiedEmail 规则独立于风控总开关：即便 cfg.Enabled=false（默认部署）也强制评估，
	// 使运营关闭邮箱验证后本地邮箱赠金无法被刷。位置在 override 查询之上，故无法被管理员 allow 覆盖。
	if cfg.RequireVerifiedEmail && (input.EmailVerified == nil || !*input.EmailVerified) {
		return stripSignupGrantBonus(plan), &signupGrantRiskClaim{
			Blocked:  true,
			Reason:   "email_not_verified",
			Decision: signupGrantRiskDecisionBlocked,
		}
	}
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

	normalizedEmail, domain := normalizeSignupGrantRiskEmail(email)
	remoteIP := normalizeSignupGrantRiskIP(input.RemoteIP)
	now := time.Now().UTC()
	hashSalt := s.signupGrantRiskSalt()
	emailHash := signupGrantRiskHash(hashSalt, normalizedEmail)
	domainHash := signupGrantRiskHash(hashSalt, domain)
	ipHash := signupGrantRiskHash(hashSalt, remoteIP)
	userAgentHash := signupGrantRiskHash(hashSalt, signupGrantDeviceSignal(input))
	providerSubjectHash := signupGrantRiskHash(hashSalt, strings.TrimSpace(input.ProviderSubject))

	// 用事务级 advisory lock 串行化"同一 IP/设备/邮箱/域名"的并发注册评估，消除 count→insert 的 TOCTOU 竞态。
	// 失败一律 fail-closed：剥夺赠金并返回 nil claim（与既有 COUNT/INSERT 错误形态一致，不返回合成 blocked claim，
	// 否则 attachSignupGrantClaim 会去 UPDATE 不存在的行）。
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] begin tx failed: %v", err)
		return stripSignupGrantBonus(plan), nil
	}
	defer func() { _ = tx.Rollback() }()

	// pg_advisory_xact_lock 仅 PostgreSQL 支持；SQLite（单测内存库）等非 PG 方言下跳过加锁，
	// count+insert 仍在同一事务内执行。生产环境恒为 Postgres。
	lockKey := firstNonEmpty(ipHash, userAgentHash, emailHash, domainHash)
	if lockKey != "" && s.signupGrantRiskIsPostgres() {
		if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", signupGrantRiskLockHash(lockKey)); err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] advisory lock failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
	}

	reason := ""
	overrideAction, overrideReason := signupGrantRiskOverrideAction(ctx, tx, emailHash, domainHash, ipHash, providerSubjectHash, userAgentHash)
	if overrideAction == "block" {
		reason = "override_block"
		if overrideReason != "" {
			reason = overrideReason
		}
	}
	if reason == "" && overrideAction != "allow" && domain != "" {
		if _, blocked := cfg.BlockedDomains[domain]; blocked {
			reason = "email_domain_blocked"
		}
	}
	if reason == "" && overrideAction != "allow" && cfg.OAuthIdentityEnabled && providerSubjectHash != "" && normalizeSignupGrantRiskLabel(input.ProviderType) != "" {
		count, err := countSignupGrantClaims(ctx, tx, "provider_type = $1 AND provider_subject_hash = $2", []any{normalizeSignupGrantRiskLabel(input.ProviderType), providerSubjectHash})
		if err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] oauth identity count failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
		if count > 0 {
			reason = "oauth_identity_limit"
		}
	}
	if reason == "" && overrideAction != "allow" && cfg.EmailLimit > 0 && emailHash != "" {
		count, err := countSignupGrantClaims(ctx, tx, "email_hash = $1", []any{emailHash})
		if err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] email count failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
		if count >= cfg.EmailLimit {
			reason = "email_limit"
		}
	}
	if reason == "" && overrideAction != "allow" && cfg.IPLimit > 0 && ipHash != "" {
		count, err := countSignupGrantClaims(ctx, tx, "ip_hash = $1 AND created_at >= $2", []any{ipHash, now.Add(-signupGrantRiskWindow)})
		if err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] ip count failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
		if count >= cfg.IPLimit {
			reason = "ip_daily_limit"
		}
	}
	if reason == "" && overrideAction != "allow" && cfg.DeviceEnabled && cfg.DeviceLimit > 0 && userAgentHash != "" {
		count, err := countSignupGrantClaims(ctx, tx, "user_agent_hash = $1 AND created_at >= $2", []any{userAgentHash, now.Add(-signupGrantRiskWindow)})
		if err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] device count failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
		if count >= cfg.DeviceLimit {
			reason = "device_daily_limit"
		}
	}
	domainLimit := cfg.domainLimitFor(domain)
	if reason == "" && overrideAction != "allow" && domainLimit > 0 && domainHash != "" {
		count, err := countSignupGrantClaims(ctx, tx, "email_domain_hash = $1 AND created_at >= $2", []any{domainHash, now.Add(-signupGrantRiskWindow)})
		if err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] domain count failed: %v", err)
			return stripSignupGrantBonus(plan), nil
		}
		if count >= domainLimit {
			reason = "email_domain_daily_limit"
		}
	}

	decision := signupGrantRiskDecisionAllowed
	filteredPlan := plan
	if reason != "" {
		decision = signupGrantRiskDecisionBlocked
		filteredPlan = stripSignupGrantBonus(plan)
	}
	claimID, err := insertSignupGrantClaim(ctx, tx, signupGrantClaimInsert{
		Email:               normalizedEmail,
		EmailDomain:         domain,
		IPAddress:           remoteIP,
		EmailHash:           emailHash,
		EmailDomainHash:     domainHash,
		IPHash:              ipHash,
		UserAgentHash:       userAgentHash,
		SignupSource:        normalizeSignupGrantRiskLabel(signupSource),
		ProviderType:        normalizeSignupGrantRiskLabel(input.ProviderType),
		ProviderSubject:     strings.TrimSpace(input.ProviderSubject),
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
	if err := tx.Commit(); err != nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] commit tx failed: %v", err)
		return stripSignupGrantBonus(plan), nil
	}
	if reason != "" {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] stripped signup grant email=%s source=%s reason=%s", anonymizeEmailForLog(email), signupSource, reason)
	}
	return filteredPlan, &signupGrantRiskClaim{
		ID:           claimID,
		Blocked:      reason != "",
		Reason:       reason,
		Decision:     decision,
		GrantBalance: plan.Balance,
	}
}

// attachSignupGrantClaim 把风控 claim 行关联到用户（写 user_id 审计列）。
// markGift=true 时同步标记 gift 余额组件（SET 语义），用于注册路径——那里 balance 由 userRepo.Create
// 一次性写入，本调用只是把"其中多少是 gift"做幂等标记。
// markGift=false 时仅做审计关联，不碰 gift——用于首绑路径，那里 balance 通过 ApplyBalanceChangeCtx
// （ADD + GiftDelta）追加，已自行处理 gift 组件；若再叠加 SET 标记会导致 gift_balance 双记。
func (s *AuthService) attachSignupGrantClaim(ctx context.Context, claim *signupGrantRiskClaim, userID int64, markGift bool) {
	if s == nil || claim == nil || claim.ID <= 0 || userID <= 0 {
		return
	}
	if markGift && claim.Decision == signupGrantRiskDecisionAllowed && claim.GrantBalance > 0 {
		s.markSignupGrantGiftBalance(ctx, userID, claim.GrantBalance)
	}
	if tx := dbent.TxFromContext(ctx); tx != nil {
		query := `UPDATE signup_grant_claims SET user_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		if tx.Client().Driver().Dialect() == dialect.Postgres {
			query = `UPDATE signup_grant_claims SET user_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
		}
		var result entsql.Result
		if err := tx.Client().Driver().Exec(ctx, query, []any{userID, claim.ID}, &result); err != nil {
			logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] attach claim failed: claim_id=%d user_id=%d err=%v", claim.ID, userID, err)
		}
		return
	}
	db := s.signupGrantRiskDB()
	if db == nil {
		return
	}
	if _, err := db.ExecContext(ctx, `UPDATE signup_grant_claims SET user_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, userID, claim.ID); err != nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] attach claim failed: claim_id=%d user_id=%d err=%v", claim.ID, userID, err)
	}
}

type signupGrantGiftBalanceSetter interface {
	SetGiftBalanceComponent(ctx context.Context, id int64, amount float64) error
}

func (s *AuthService) markSignupGrantGiftBalance(ctx context.Context, userID int64, amount float64) {
	if s == nil || s.userRepo == nil || amount <= 0 {
		return
	}
	setter, ok := s.userRepo.(signupGrantGiftBalanceSetter)
	if !ok {
		return
	}
	if err := setter.SetGiftBalanceComponent(ctx, userID, amount); err != nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] mark gift balance failed: user_id=%d amount=%f err=%v", userID, amount, err)
	}
}

func (s *AuthService) resolveSignupGrantPlanForFinalizedUser(ctx context.Context, userID int64, email, signupSource string) (signupGrantPlan, *signupGrantRiskClaim) {
	plan := s.resolveSignupGrantPlan(ctx, signupSource)
	if s == nil || userID <= 0 || s.settingService == nil || !signupGrantPlanHasBonus(plan) {
		return plan, nil
	}
	cfg := s.settingService.signupGrantRiskConfig(ctx)
	// 与 applySignupGrantRiskControlEx 一致：RequireVerifiedEmail 独立于风控总开关，finalize 路径同样强制。
	if cfg.RequireVerifiedEmail {
		input := signupGrantRiskInputFromContext(ctx)
		if input.EmailVerified == nil || !*input.EmailVerified {
			return stripSignupGrantBonus(plan), &signupGrantRiskClaim{
				Blocked:  true,
				Reason:   "email_not_verified",
				Decision: signupGrantRiskDecisionBlocked,
			}
		}
	}
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

// signupGrantRiskIsPostgres 报告底层驱动方言是否为 PostgreSQL。pg_advisory_xact_lock 仅 PG 支持。
func (s *AuthService) signupGrantRiskIsPostgres() bool {
	if s == nil || s.entClient == nil {
		return false
	}
	return s.entClient.Driver().Dialect() == dialect.Postgres
}

// signupGrantRiskQueryRunner 是 count/insert/override 共需的最小查询接口，*sql.DB 与 *sql.Tx 均满足，
// 使限额评估可在事务（含 advisory lock）内执行以消除 TOCTOU。
type signupGrantRiskQueryRunner interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// signupGrantRiskLockHash 把锁键哈希成 int64，匹配代码库既有的 advisory lock 约定
// （Go 侧 FNV-64a + pg_advisory_xact_lock($1)，见 auth_pending_identity_service.go）。
func signupGrantRiskLockHash(key string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(key))
	return int64(hasher.Sum64())
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
	return "cloudbase-signup-grant-risk"
}

func signupGrantPlanHasBonus(plan signupGrantPlan) bool {
	return plan.Balance > 0 || len(plan.Subscriptions) > 0 || len(plan.PlatformQuotas) > 0
}

func signupGrantRiskAppliesToSource(signupSource string) bool {
	return identity.SignupGrantRiskAppliesToSource(signupSource)
}

func stripSignupGrantBonus(plan signupGrantPlan) signupGrantPlan {
	return signupGrantPlan{Concurrency: plan.Concurrency}
}

func signupGrantRiskHash(salt string, value string) string {
	return identity.SignupGrantRiskHash(salt, value)
}

func normalizeSignupGrantRiskEmail(raw string) (string, string) {
	return identity.NormalizeSignupGrantRiskEmail(raw)
}

func normalizeSignupGrantRiskIP(raw string) string {
	return identity.NormalizeSignupGrantRiskIP(raw)
}

func normalizeSignupGrantRiskLabel(raw string) string {
	return identity.NormalizeSignupGrantRiskLabel(raw)
}

func signupGrantDeviceSignal(input SignupGrantRiskInput) string {
	return identity.SignupGrantDeviceSignal(identity.SignupGrantRiskInput{
		RemoteIP:          input.RemoteIP,
		UserAgent:         input.UserAgent,
		AcceptLanguage:    input.AcceptLanguage,
		DeviceFingerprint: input.DeviceFingerprint,
		ProviderType:      input.ProviderType,
		ProviderSubject:   input.ProviderSubject,
	})
}

func (cfg signupGrantRiskConfig) domainLimitFor(domain string) int {
	return identity.SignupGrantRiskConfig{
		DomainLimit:     cfg.DomainLimit,
		FreeDomainLimit: cfg.FreeDomainLimit,
		FreeDomains:     cfg.FreeDomains,
		TrustedDomains:  cfg.TrustedDomains,
	}.DomainLimitFor(domain)
}

func normalizeSignupGrantRiskDomain(raw string) string {
	return identity.NormalizeSignupGrantRiskDomain(raw)
}

const defaultSignupGrantRiskFreeDomains = "gmail.com,googlemail.com,outlook.com,hotmail.com,live.com,icloud.com,yahoo.com,qq.com,163.com,126.com,foxmail.com"

func parseSignupGrantRiskDomainSet(raw string) map[string]struct{} {
	return identity.ParseSignupGrantRiskDomainSet(raw)
}

func normalizeDomainListSetting(raw string) string {
	return identity.NormalizeDomainListSetting(raw)
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func signupGrantRiskOverrideAction(ctx context.Context, db signupGrantRiskQueryRunner, emailHash, domainHash, ipHash, providerSubjectHash, deviceHash string) (string, string) {
	if db == nil {
		return "", ""
	}
	candidates := []struct {
		typ  string
		hash string
	}{
		{typ: "email", hash: emailHash},
		{typ: "email_domain", hash: domainHash},
		{typ: "ip", hash: ipHash},
		{typ: "oauth_identity", hash: providerSubjectHash},
		{typ: "device", hash: deviceHash},
	}
	for _, action := range []string{"block", "allow"} {
		for _, c := range candidates {
			if c.hash == "" {
				continue
			}
			var reason string
			err := db.QueryRowContext(ctx, `
				SELECT reason
				FROM signup_grant_risk_overrides
				WHERE subject_type = $1
				  AND subject_hash = $2
				  AND action = $3
				  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
				LIMIT 1
			`, c.typ, c.hash, action).Scan(&reason)
			if err == nil {
				return action, strings.TrimSpace(reason)
			}
			if !errors.Is(err, sql.ErrNoRows) {
				logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] override lookup failed: %v", err)
				return "", ""
			}
		}
	}
	return "", ""
}

type signupGrantClaimInsert struct {
	Email               string
	EmailDomain         string
	IPAddress           string
	EmailHash           string
	EmailDomainHash     string
	IPHash              string
	UserAgentHash       string
	SignupSource        string
	ProviderType        string
	ProviderSubject     string
	ProviderSubjectHash string
	Decision            string
	Reason              string
	GrantBalance        float64
	GrantMetadataJSON   string
}

func countSignupGrantClaims(ctx context.Context, db signupGrantRiskQueryRunner, predicate string, args []any) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM signup_grant_claims WHERE decision = 'allowed' AND %s`, predicate)
	var count int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func insertSignupGrantClaim(ctx context.Context, db signupGrantRiskQueryRunner, input signupGrantClaimInsert) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, `
INSERT INTO signup_grant_claims (
    email, email_domain, ip_address,
    email_hash, email_domain_hash, ip_hash, user_agent_hash,
    signup_source, provider_type, provider_subject, provider_subject_hash,
    decision, reason, grant_balance, grant_metadata, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, CURRENT_TIMESTAMP)
RETURNING id
`,
		input.Email,
		input.EmailDomain,
		input.IPAddress,
		input.EmailHash,
		input.EmailDomainHash,
		input.IPHash,
		input.UserAgentHash,
		input.SignupSource,
		input.ProviderType,
		input.ProviderSubject,
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
		SettingKeySignupGrantRiskControlOAuthIdentityEnabled,
		SettingKeySignupGrantRiskControlDeviceEnabled,
		SettingKeySignupGrantRiskControlDeviceLimit,
		SettingKeySignupGrantRiskControlFreeDomainLimit,
		SettingKeySignupGrantRiskControlBlockedDomains,
		SettingKeySignupGrantRiskControlFreeDomains,
		SettingKeySignupGrantRiskControlTrustedDomains,
		SettingKeySignupGrantRiskControlRequireVerifiedEmail,
	})
	if err != nil {
		logger.LegacyPrintf("service.auth.risk", "[SignupGrantRisk] load config failed: %v", err)
		return signupGrantRiskConfig{}
	}
	return signupGrantRiskConfig{
		Enabled:              values[SettingKeySignupGrantRiskControlEnabled] == "true",
		EmailLimit:           parseNonNegativeIntSetting(values[SettingKeySignupGrantRiskControlEmailLimit], 1),
		IPLimit:              parseNonNegativeIntSetting(values[SettingKeySignupGrantRiskControlIPLimit], 3),
		DomainLimit:          parseNonNegativeIntSetting(values[SettingKeySignupGrantRiskControlDomainLimit], 10),
		OAuthIdentityEnabled: values[SettingKeySignupGrantRiskControlOAuthIdentityEnabled] != "false",
		DeviceEnabled:        values[SettingKeySignupGrantRiskControlDeviceEnabled] != "false",
		DeviceLimit:          parseNonNegativeIntSetting(values[SettingKeySignupGrantRiskControlDeviceLimit], 2),
		FreeDomainLimit:      parseNonNegativeIntSetting(values[SettingKeySignupGrantRiskControlFreeDomainLimit], 5),
		BlockedDomains:       parseSignupGrantRiskDomainSet(values[SettingKeySignupGrantRiskControlBlockedDomains]),
		FreeDomains:          parseSignupGrantRiskDomainSet(defaultIfEmpty(values[SettingKeySignupGrantRiskControlFreeDomains], defaultSignupGrantRiskFreeDomains)),
		TrustedDomains:       parseSignupGrantRiskDomainSet(values[SettingKeySignupGrantRiskControlTrustedDomains]),
		RequireVerifiedEmail: values[SettingKeySignupGrantRiskControlRequireVerifiedEmail] != "false",
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
