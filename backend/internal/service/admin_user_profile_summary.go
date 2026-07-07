package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	dbent "github.com/Aias00/cloudbase/ent"
)

type UserProfileSummary struct {
	User           UserProfileSummaryUser           `json:"user"`
	Classification UserProfileClassification        `json:"classification"`
	Registration   UserProfileRegistrationSummary   `json:"registration"`
	AuthIdentities []UserProfileAuthIdentitySummary `json:"auth_identities"`
	Activity       UserProfileActivitySummary       `json:"activity"`
	APIKeys        UserProfileAPIKeySummary         `json:"api_keys"`
	Payments       UserProfilePaymentSummary        `json:"payments"`
	Balance        UserProfileBalanceSummary        `json:"balance"`
	Business       UserProfileBusinessSummary       `json:"business"`
	RiskTags       []UserProfileRiskTag             `json:"risk_tags"`
}

type UserProfileSummaryUser struct {
	ID             int64      `json:"id"`
	Email          string     `json:"email"`
	Username       string     `json:"username"`
	Notes          string     `json:"notes,omitempty"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	SignupSource   string     `json:"signup_source"`
	Balance        float64    `json:"balance"`
	PaidBalance    float64    `json:"paid_balance"`
	GiftBalance    float64    `json:"gift_balance"`
	TotalRecharged float64    `json:"total_recharged"`
	Concurrency    int        `json:"concurrency"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	LastActiveAt   *time.Time `json:"last_active_at,omitempty"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type UserProfileClassification struct {
	Category   string   `json:"category"`
	Label      string   `json:"label"`
	Confidence string   `json:"confidence"`
	Reasons    []string `json:"reasons"`
}

type UserProfileRegistrationSummary struct {
	RegisteredVia         string            `json:"registered_via"`
	RegistrationIP        string            `json:"registration_ip,omitempty"`
	UserAgent             string            `json:"user_agent,omitempty"`
	AcceptLanguage        string            `json:"accept_language,omitempty"`
	DeviceFingerprint     string            `json:"device_fingerprint,omitempty"`
	HeaderSnapshot        map[string]string `json:"header_snapshot,omitempty"`
	NearbyAuthEvent       string            `json:"nearby_auth_event,omitempty"`
	NearbyAuthStatus      string            `json:"nearby_auth_status,omitempty"`
	NearbyAuthAt          *time.Time        `json:"nearby_auth_at,omitempty"`
	SameIPSignupCount24h  int               `json:"same_ip_signup_count_24h"`
	SameDomainSignupCount int               `json:"same_domain_signup_count"`
	EmailDomain           string            `json:"email_domain"`
	DisposableEmail       bool              `json:"disposable_email"`
}

type UserProfileAuthIdentitySummary struct {
	ProviderType    string     `json:"provider_type"`
	ProviderKey     string     `json:"provider_key"`
	ProviderSubject string     `json:"provider_subject"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type UserProfileActivitySummary struct {
	APIUsageCount   int64      `json:"api_usage_count"`
	APIActualCost   float64    `json:"api_actual_cost"`
	FirstAPIUsageAt *time.Time `json:"first_api_usage_at,omitempty"`
	LastAPIUsageAt  *time.Time `json:"last_api_usage_at,omitempty"`
	LastHTTPAt      *time.Time `json:"last_http_at,omitempty"`
}

type UserProfileAPIKeySummary struct {
	TotalCount     int64      `json:"total_count"`
	ActiveCount    int64      `json:"active_count"`
	FirstCreatedAt *time.Time `json:"first_created_at,omitempty"`
	LastCreatedAt  *time.Time `json:"last_created_at,omitempty"`
}

type UserProfilePaymentSummary struct {
	OrderCount     int64      `json:"order_count"`
	PaidOrderCount int64      `json:"paid_order_count"`
	PaidAmount     float64    `json:"paid_amount"`
	RefundAmount   float64    `json:"refund_amount"`
	LastOrderAt    *time.Time `json:"last_order_at,omitempty"`
}

type UserProfileBalanceSummary struct {
	LedgerCount          int64   `json:"ledger_count"`
	PositiveLedgerAmount float64 `json:"positive_ledger_amount"`
	NetLedgerAmount      float64 `json:"net_ledger_amount"`
	RedeemCount          int64   `json:"redeem_count"`
	RedeemBalanceAmount  float64 `json:"redeem_balance_amount"`
}

type UserProfileBusinessSummary struct {
	ImageTaskCount    int64      `json:"image_task_count"`
	ImageSuccessCount int64      `json:"image_success_count"`
	ImageActualCost   float64    `json:"image_actual_cost"`
	FirstImageTaskAt  *time.Time `json:"first_image_task_at,omitempty"`
	LastImageTaskAt   *time.Time `json:"last_image_task_at,omitempty"`
	WechatTaskCount   int64      `json:"wechat_task_count"`
	WechatActualCost  float64    `json:"wechat_actual_cost"`
	FirstWechatTaskAt *time.Time `json:"first_wechat_task_at,omitempty"`
	LastWechatTaskAt  *time.Time `json:"last_wechat_task_at,omitempty"`
}

type UserProfileRiskTag struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Severity string `json:"severity"`
	Detail   string `json:"detail"`
}

func (s *adminServiceImpl) GetUserProfileSummary(ctx context.Context, userID int64) (*UserProfileSummary, error) {
	if s.entClient == nil {
		return nil, ErrServiceUnavailable
	}

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := &UserProfileSummary{
		User: UserProfileSummaryUser{
			ID:             user.ID,
			Email:          user.Email,
			Username:       user.Username,
			Notes:          user.Notes,
			Role:           user.Role,
			Status:         user.Status,
			SignupSource:   user.SignupSource,
			Balance:        user.Balance,
			PaidBalance:    user.PaidBalance,
			GiftBalance:    user.GiftBalance,
			TotalRecharged: user.TotalRecharged,
			Concurrency:    user.Concurrency,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
			LastLoginAt:    user.LastLoginAt,
			LastActiveAt:   user.LastActiveAt,
			LastUsedAt:     user.LastUsedAt,
			DeletedAt:      user.DeletedAt,
		},
		Registration: UserProfileRegistrationSummary{
			RegisteredVia: user.SignupSource,
			EmailDomain:   userProfileEmailDomain(user.Email),
		},
		RiskTags: []UserProfileRiskTag{},
	}
	if summary.Registration.RegisteredVia == "" {
		summary.Registration.RegisteredVia = "unknown"
	}
	summary.Registration.DisposableEmail = isDisposableUserProfileEmailDomain(summary.Registration.EmailDomain)

	s.loadUserProfileRegistration(ctx, summary, user.CreatedAt)
	s.loadUserProfileAuthIdentities(ctx, summary, userID)
	s.loadUserProfileAPIKeys(ctx, summary, userID)
	s.loadUserProfileActivity(ctx, summary, userID)
	s.loadUserProfilePayments(ctx, summary, userID)
	s.loadUserProfileBalance(ctx, summary, userID)
	s.loadUserProfileBusiness(ctx, summary, userID)
	summary.Classification = classifyUserProfile(summary)

	return summary, nil
}

func (s *adminServiceImpl) loadUserProfileRegistration(ctx context.Context, summary *UserProfileSummary, createdAt time.Time) {
	s.loadPersistedUserRegistrationEvent(ctx, summary)
	var ip, event, status sql.NullString
	var eventAt sql.NullTime
	if summary.Registration.RegistrationIP == "" {
		err := scanAdminProfileRow(ctx, s.entClient, `
SELECT
  COALESCE(extra->>'client_ip', '') AS client_ip,
  COALESCE(extra->>'path', '') AS path,
  COALESCE(extra->>'status_code', '') AS status_code,
  created_at
FROM ops_system_logs
WHERE component = 'http.access'
  AND created_at BETWEEN $1::timestamptz - INTERVAL '2 minutes' AND $1::timestamptz + INTERVAL '2 minutes'
  AND COALESCE(extra->>'path', '') IN ('/api/v1/auth/register', '/api/v1/auth/login', '/api/v1/auth/oauth/google/callback')
ORDER BY ABS(EXTRACT(EPOCH FROM (created_at - $1::timestamptz))) ASC
LIMIT 1`, []any{createdAt}, &ip, &event, &status, &eventAt)
		if err == nil {
			summary.Registration.RegistrationIP = ip.String
			summary.Registration.NearbyAuthEvent = event.String
			summary.Registration.NearbyAuthStatus = status.String
			summary.Registration.NearbyAuthAt = nullableTimePtr(eventAt)
		}
	}

	if summary.Registration.EmailDomain != "" {
		var sameDomain int
		if err := scanAdminProfileRow(ctx, s.entClient, `
SELECT COUNT(*)::int
FROM users
WHERE deleted_at IS NULL
  AND lower(split_part(email, '@', 2)) = $1`, []any{summary.Registration.EmailDomain}, &sameDomain); err == nil {
			summary.Registration.SameDomainSignupCount = sameDomain
		}
	}

	if summary.Registration.RegistrationIP == "" {
		return
	}
	var sameIP int
	if err := scanAdminProfileRow(ctx, s.entClient, `
SELECT COUNT(*)::int
FROM user_registration_events
WHERE ip_address = $1
  AND created_at BETWEEN $2::timestamptz - INTERVAL '24 hours' AND $2::timestamptz + INTERVAL '24 hours'`, []any{summary.Registration.RegistrationIP, createdAt}, &sameIP); err == nil {
		summary.Registration.SameIPSignupCount24h = sameIP
		return
	}
	if err := scanAdminProfileRow(ctx, s.entClient, `
WITH recent_user_ips AS (
  SELECT u.id,
    (
      SELECT COALESCE(l.extra->>'client_ip', '')
      FROM ops_system_logs l
      WHERE l.component = 'http.access'
        AND l.created_at BETWEEN u.created_at - INTERVAL '2 minutes' AND u.created_at + INTERVAL '2 minutes'
        AND COALESCE(l.extra->>'path', '') IN ('/api/v1/auth/register', '/api/v1/auth/login', '/api/v1/auth/oauth/google/callback')
      ORDER BY ABS(EXTRACT(EPOCH FROM (l.created_at - u.created_at))) ASC
      LIMIT 1
    ) AS signup_ip
  FROM users u
  WHERE u.deleted_at IS NULL
    AND u.created_at BETWEEN $2::timestamptz - INTERVAL '24 hours' AND $2::timestamptz + INTERVAL '24 hours'
)
SELECT COUNT(*)::int FROM recent_user_ips WHERE signup_ip = $1`, []any{summary.Registration.RegistrationIP, createdAt}, &sameIP); err == nil {
		summary.Registration.SameIPSignupCount24h = sameIP
	}
}

func (s *adminServiceImpl) loadPersistedUserRegistrationEvent(ctx context.Context, summary *UserProfileSummary) {
	var ip, userAgent, acceptLanguage, deviceFingerprint, headersRaw sql.NullString
	var eventAt sql.NullTime
	err := scanAdminProfileRow(ctx, s.entClient, `
SELECT
  COALESCE(ip_address, ''),
  COALESCE(user_agent, ''),
  COALESCE(accept_language, ''),
  COALESCE(device_fingerprint, ''),
  COALESCE(headers_json::text, '{}'),
  created_at
FROM user_registration_events
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT 1`, []any{summary.User.ID}, &ip, &userAgent, &acceptLanguage, &deviceFingerprint, &headersRaw, &eventAt)
	if err != nil {
		return
	}
	if ip.String != "" {
		summary.Registration.RegistrationIP = ip.String
	}
	summary.Registration.UserAgent = userAgent.String
	summary.Registration.AcceptLanguage = acceptLanguage.String
	summary.Registration.DeviceFingerprint = deviceFingerprint.String
	summary.Registration.HeaderSnapshot = parseRegistrationHeaderSnapshot(headersRaw.String)
	summary.Registration.NearbyAuthEvent = "user_registration_events"
	summary.Registration.NearbyAuthStatus = "recorded"
	summary.Registration.NearbyAuthAt = nullableTimePtr(eventAt)
}

func (s *adminServiceImpl) loadUserProfileAuthIdentities(ctx context.Context, summary *UserProfileSummary, userID int64) {
	rows, err := s.entClient.QueryContext(ctx, `
SELECT provider_type, provider_key, provider_subject, verified_at, created_at
FROM auth_identities
WHERE user_id = $1
ORDER BY created_at ASC, id ASC`, userID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item UserProfileAuthIdentitySummary
		var verifiedAt sql.NullTime
		if err := rows.Scan(&item.ProviderType, &item.ProviderKey, &item.ProviderSubject, &verifiedAt, &item.CreatedAt); err != nil {
			return
		}
		item.VerifiedAt = nullableTimePtr(verifiedAt)
		summary.AuthIdentities = append(summary.AuthIdentities, item)
	}
}

func (s *adminServiceImpl) loadUserProfileAPIKeys(ctx context.Context, summary *UserProfileSummary, userID int64) {
	var firstCreated, lastCreated sql.NullTime
	if err := scanAdminProfileRow(ctx, s.entClient, `
SELECT
  COUNT(*)::bigint,
  COUNT(*) FILTER (WHERE deleted_at IS NULL AND status = 'active')::bigint,
  MIN(created_at),
  MAX(created_at)
FROM api_keys
WHERE user_id = $1`, []any{userID}, &summary.APIKeys.TotalCount, &summary.APIKeys.ActiveCount, &firstCreated, &lastCreated); err != nil {
		return
	}
	summary.APIKeys.FirstCreatedAt = nullableTimePtr(firstCreated)
	summary.APIKeys.LastCreatedAt = nullableTimePtr(lastCreated)
}

func (s *adminServiceImpl) loadUserProfileActivity(ctx context.Context, summary *UserProfileSummary, userID int64) {
	var firstUsage, lastUsage, lastHTTP sql.NullTime
	if err := scanAdminProfileRow(ctx, s.entClient, `
SELECT COUNT(*)::bigint, COALESCE(SUM(actual_cost), 0)::double precision, MIN(created_at), MAX(created_at)
FROM usage_logs
WHERE user_id = $1`, []any{userID}, &summary.Activity.APIUsageCount, &summary.Activity.APIActualCost, &firstUsage, &lastUsage); err == nil {
		summary.Activity.FirstAPIUsageAt = nullableTimePtr(firstUsage)
		summary.Activity.LastAPIUsageAt = nullableTimePtr(lastUsage)
	}
	if err := scanAdminProfileRow(ctx, s.entClient, `
SELECT MAX(created_at)
FROM ops_system_logs
WHERE user_id = $1`, []any{userID}, &lastHTTP); err == nil {
		summary.Activity.LastHTTPAt = nullableTimePtr(lastHTTP)
	}
}

func (s *adminServiceImpl) loadUserProfilePayments(ctx context.Context, summary *UserProfileSummary, userID int64) {
	var lastOrder sql.NullTime
	if err := scanAdminProfileRow(ctx, s.entClient, `
SELECT
  COUNT(*)::bigint,
  COUNT(*) FILTER (WHERE status IN ('PAID', 'RECHARGING', 'COMPLETED'))::bigint,
  COALESCE(SUM(pay_amount) FILTER (WHERE status IN ('PAID', 'RECHARGING', 'COMPLETED')), 0)::double precision,
  COALESCE(SUM(refund_amount), 0)::double precision,
  MAX(created_at)
FROM payment_orders
WHERE user_id = $1`, []any{userID}, &summary.Payments.OrderCount, &summary.Payments.PaidOrderCount, &summary.Payments.PaidAmount, &summary.Payments.RefundAmount, &lastOrder); err != nil {
		return
	}
	summary.Payments.LastOrderAt = nullableTimePtr(lastOrder)
}

func (s *adminServiceImpl) loadUserProfileBalance(ctx context.Context, summary *UserProfileSummary, userID int64) {
	_ = scanAdminProfileRow(ctx, s.entClient, `
SELECT
  COUNT(*)::bigint,
  COALESCE(SUM(amount) FILTER (WHERE amount > 0), 0)::double precision,
  COALESCE(SUM(amount), 0)::double precision
FROM user_balance_ledger
WHERE user_id = $1`, []any{userID}, &summary.Balance.LedgerCount, &summary.Balance.PositiveLedgerAmount, &summary.Balance.NetLedgerAmount)

	_ = scanAdminProfileRow(ctx, s.entClient, `
SELECT COUNT(*)::bigint, COALESCE(SUM(value), 0)::double precision
FROM redeem_codes
WHERE used_by = $1`, []any{userID}, &summary.Balance.RedeemCount, &summary.Balance.RedeemBalanceAmount)
}

func (s *adminServiceImpl) loadUserProfileBusiness(ctx context.Context, summary *UserProfileSummary, userID int64) {
	var firstImage, lastImage, firstWechat, lastWechat sql.NullTime
	if err := scanAdminProfileRow(ctx, s.entClient, `
SELECT COUNT(*)::bigint, COUNT(*) FILTER (WHERE status = 'succeeded')::bigint, MIN(created_at), MAX(updated_at)
FROM image_workspace_tasks
WHERE user_id = $1`, []any{userID}, &summary.Business.ImageTaskCount, &summary.Business.ImageSuccessCount, &firstImage, &lastImage); err == nil {
		summary.Business.FirstImageTaskAt = nullableTimePtr(firstImage)
		summary.Business.LastImageTaskAt = nullableTimePtr(lastImage)
	}
	_ = scanAdminProfileRow(ctx, s.entClient, `
SELECT COALESCE(SUM(actual_cost), 0)::double precision
FROM image_workspace_usage_records
WHERE user_id = $1`, []any{userID}, &summary.Business.ImageActualCost)

	if err := scanAdminProfileRow(ctx, s.entClient, `
SELECT COUNT(*)::bigint, MIN(created_at), MAX(updated_at)
FROM wechat_export_tasks
WHERE user_id = $1`, []any{userID}, &summary.Business.WechatTaskCount, &firstWechat, &lastWechat); err == nil {
		summary.Business.FirstWechatTaskAt = nullableTimePtr(firstWechat)
		summary.Business.LastWechatTaskAt = nullableTimePtr(lastWechat)
	}
	_ = scanAdminProfileRow(ctx, s.entClient, `
SELECT COALESCE(SUM(actual_cost), 0)::double precision
FROM wechat_export_usage_records
WHERE user_id = $1`, []any{userID}, &summary.Business.WechatActualCost)
}

func classifyUserProfile(summary *UserProfileSummary) UserProfileClassification {
	reasons := []string{}
	addTag := func(key, label, severity, detail string) {
		summary.RiskTags = append(summary.RiskTags, UserProfileRiskTag{
			Key: key, Label: label, Severity: severity, Detail: detail,
		})
	}

	email := strings.ToLower(strings.TrimSpace(summary.User.Email))
	username := strings.ToLower(strings.TrimSpace(summary.User.Username))
	switch {
	case summary.User.Role == RoleAdmin:
		reasons = append(reasons, "用户角色为管理员")
		addTag("admin", "管理员账号", "info", "该账号拥有管理员权限")
		return UserProfileClassification{Category: "admin", Label: "管理员", Confidence: "high", Reasons: reasons}
	case strings.HasSuffix(email, "@sub2api.local"):
		reasons = append(reasons, "系统保留邮箱域")
		addTag("system_account", "系统账号", "info", "邮箱域为系统保留域")
		return UserProfileClassification{Category: "system", Label: "系统账号", Confidence: "high", Reasons: reasons}
	case strings.Contains(email, "example.test") || strings.Contains(email, "smoke") || strings.Contains(username, "smoke"):
		reasons = append(reasons, "邮箱或用户名符合 smoke/test 模式")
		addTag("smoke_test", "测试账号", "info", "疑似由 smoke test 或本地验证创建")
		return UserProfileClassification{Category: "smoke_test", Label: "测试账号", Confidence: "high", Reasons: reasons}
	}

	if summary.Registration.DisposableEmail {
		reasons = append(reasons, "邮箱域疑似临时邮箱")
		addTag("disposable_email", "临时邮箱", "warning", summary.Registration.EmailDomain)
	}
	if summary.Registration.SameIPSignupCount24h >= 3 {
		reasons = append(reasons, "24 小时内同 IP 注册数偏高")
		addTag("same_ip_signup_burst", "同 IP 注册偏多", "warning", "24 小时内同 IP 注册账号数偏高")
	}
	if summary.Registration.SameDomainSignupCount >= 5 && !commonConsumerEmailDomain(summary.Registration.EmailDomain) {
		addTag("same_domain_many", "同域账号较多", "info", "该邮箱域已有多个账号")
	}
	if len(summary.AuthIdentities) == 0 && summary.User.Role != RoleAdmin {
		addTag("missing_auth_identity", "缺少身份绑定", "warning", "未找到 auth_identities 记录，可能是历史数据或人工创建")
	}
	if summary.APIKeys.ActiveCount > 0 && summary.Activity.APIUsageCount == 0 {
		addTag("api_key_without_usage", "Key 未产生调用", "info", "存在活跃 API Key，但暂无调用记录")
	}
	if summary.Payments.PaidOrderCount == 0 {
		addTag("no_payment", "无成功支付", "info", "未找到已支付或已完成订单")
	}

	if len(reasons) > 0 {
		return UserProfileClassification{Category: "needs_review", Label: "需关注", Confidence: "medium", Reasons: reasons}
	}
	if summary.APIKeys.TotalCount == 0 && summary.Activity.APIUsageCount == 0 && summary.Payments.OrderCount == 0 && summary.Business.ImageTaskCount == 0 && summary.Business.WechatTaskCount == 0 {
		return UserProfileClassification{
			Category:   "inactive_registered",
			Label:      "未活跃注册用户",
			Confidence: "medium",
			Reasons:    []string{"暂无 Key、调用、订单或业务任务记录"},
		}
	}
	return UserProfileClassification{
		Category:   "registered",
		Label:      "注册用户",
		Confidence: "medium",
		Reasons:    []string{"存在正常用户记录"},
	}
}

func scanAdminProfileRow(ctx context.Context, client *dbent.Client, query string, args []any, dest ...any) error {
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return sql.ErrNoRows
	}
	return rows.Scan(dest...)
}

func nullableTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func parseRegistrationHeaderSnapshot(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil
	}
	var out map[string]string
	if err := json.Unmarshal([]byte(raw), &out); err != nil || len(out) == 0 {
		return nil
	}
	return out
}

func userProfileEmailDomain(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return ""
	}
	return email[at+1:]
}

func isDisposableUserProfileEmailDomain(domain string) bool {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return false
	}
	disposableDomains := map[string]struct{}{
		"10minutemail.com": {},
		"1secmail.com":     {},
		"mailinator.com":   {},
		"tempmail.com":     {},
		"temp-mail.org":    {},
		"yopmail.com":      {},
	}
	if _, ok := disposableDomains[domain]; ok {
		return true
	}
	return strings.Contains(domain, "tempmail") || strings.Contains(domain, "temporary-mail")
}

func commonConsumerEmailDomain(domain string) bool {
	switch strings.ToLower(strings.TrimSpace(domain)) {
	case "gmail.com", "googlemail.com", "outlook.com", "hotmail.com", "live.com", "qq.com", "163.com", "126.com", "icloud.com", "yahoo.com":
		return true
	default:
		return false
	}
}
