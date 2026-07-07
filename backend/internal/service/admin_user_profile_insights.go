package service

import (
	"context"
	"database/sql"
	"math"
	"strings"
	"time"
)

type UserProfileInsights struct {
	GeneratedAt     time.Time               `json:"generated_at"`
	Classification  []UserInsightCount      `json:"classification"`
	SignupSources   []UserInsightCount      `json:"signup_sources"`
	RegistrationIPs []UserInsightDimension  `json:"registration_ips"`
	UserAgents      []UserInsightDimension  `json:"user_agents"`
	Languages       []UserInsightDimension  `json:"languages"`
	Funnel          []UserInsightFunnelStep `json:"funnel"`
	RiskSamples     []UserInsightRiskSample `json:"risk_samples"`
}

type UserInsightCount struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type UserInsightDimension struct {
	Value    string     `json:"value"`
	Count    int64      `json:"count"`
	LastSeen *time.Time `json:"last_seen,omitempty"`
}

type UserInsightFunnelStep struct {
	Key        string  `json:"key"`
	Label      string  `json:"label"`
	Count      int64   `json:"count"`
	Conversion float64 `json:"conversion"`
}

type UserInsightRiskSample struct {
	UserID         int64      `json:"user_id"`
	Email          string     `json:"email"`
	Username       string     `json:"username"`
	Label          string     `json:"label"`
	Reason         string     `json:"reason"`
	Severity       string     `json:"severity"`
	RegistrationIP string     `json:"registration_ip,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	LastActiveAt   *time.Time `json:"last_active_at,omitempty"`
}

func (s *adminServiceImpl) GetUserProfileInsights(ctx context.Context, limit int) (*UserProfileInsights, error) {
	if s.entClient == nil {
		return nil, ErrServiceUnavailable
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	out := &UserProfileInsights{GeneratedAt: time.Now().UTC()}
	s.loadUserInsightClassification(ctx, out)
	s.loadUserInsightSignupSources(ctx, out)
	s.loadUserInsightDimensions(ctx, out)
	s.loadUserInsightFunnel(ctx, out)
	s.loadUserInsightRiskSamples(ctx, out, limit)
	return out, nil
}

func (s *adminServiceImpl) loadUserInsightClassification(ctx context.Context, out *UserProfileInsights) {
	var total, admins, systemAccounts, smokeTests, needsReview, inactive int64
	err := scanAdminProfileRow(ctx, s.entClient, `
WITH active_users AS (
  SELECT u.*
  FROM users u
  WHERE u.deleted_at IS NULL
),
review_flags AS (
  SELECT u.id,
    EXISTS (
      SELECT 1 FROM auth_identities ai WHERE ai.user_id = u.id
    ) AS has_identity,
    EXISTS (
      SELECT 1
      FROM user_registration_events e
      WHERE e.user_id = u.id
        AND e.ip_address <> ''
        AND (
          SELECT COUNT(*)
          FROM user_registration_events e2
          WHERE e2.ip_address = e.ip_address
            AND e2.created_at BETWEEN e.created_at - INTERVAL '24 hours' AND e.created_at + INTERVAL '24 hours'
        ) >= 3
    ) AS same_ip_burst
  FROM active_users u
),
activity_flags AS (
  SELECT u.id,
    EXISTS (SELECT 1 FROM api_keys ak WHERE ak.user_id = u.id AND ak.deleted_at IS NULL) AS has_key,
    EXISTS (SELECT 1 FROM usage_logs ul WHERE ul.user_id = u.id) AS has_usage,
    EXISTS (SELECT 1 FROM payment_orders po WHERE po.user_id = u.id) AS has_order,
    EXISTS (SELECT 1 FROM image_workspace_tasks iwt WHERE iwt.user_id = u.id)
      OR EXISTS (SELECT 1 FROM wechat_export_tasks wet WHERE wet.user_id = u.id) AS has_business
  FROM active_users u
)
SELECT
  COUNT(*)::bigint,
  COUNT(*) FILTER (WHERE role = 'admin')::bigint,
  COUNT(*) FILTER (WHERE lower(email) LIKE '%@sub2api.local')::bigint,
  COUNT(*) FILTER (WHERE lower(email) LIKE '%example.test%' OR lower(email) LIKE '%smoke%' OR lower(username) LIKE '%smoke%')::bigint,
  COUNT(*) FILTER (WHERE role <> 'admin' AND (NOT rf.has_identity OR rf.same_ip_burst))::bigint,
  COUNT(*) FILTER (WHERE NOT af.has_key AND NOT af.has_usage AND NOT af.has_order AND NOT af.has_business)::bigint
FROM active_users u
JOIN review_flags rf ON rf.id = u.id
JOIN activity_flags af ON af.id = u.id`, []any{}, &total, &admins, &systemAccounts, &smokeTests, &needsReview, &inactive)
	if err != nil {
		return
	}
	registered := total - admins - systemAccounts - smokeTests - needsReview - inactive
	if registered < 0 {
		registered = 0
	}
	out.Classification = []UserInsightCount{
		{Key: "registered", Label: "注册用户", Count: registered},
		{Key: "inactive_registered", Label: "未活跃注册用户", Count: inactive},
		{Key: "needs_review", Label: "需关注账号", Count: needsReview},
		{Key: "smoke_test", Label: "测试账号", Count: smokeTests},
		{Key: "system", Label: "系统账号", Count: systemAccounts},
		{Key: "admin", Label: "管理员", Count: admins},
	}
}

func (s *adminServiceImpl) loadUserInsightSignupSources(ctx context.Context, out *UserProfileInsights) {
	rows, err := s.entClient.QueryContext(ctx, `
SELECT COALESCE(NULLIF(signup_source, ''), 'unknown') AS source, COUNT(*)::bigint
FROM users
WHERE deleted_at IS NULL
GROUP BY source
ORDER BY COUNT(*) DESC, source ASC`)
	if err != nil {
		return
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var item UserInsightCount
		if err := rows.Scan(&item.Key, &item.Count); err != nil {
			return
		}
		item.Label = userInsightSourceLabel(item.Key)
		out.SignupSources = append(out.SignupSources, item)
	}
}

func (s *adminServiceImpl) loadUserInsightDimensions(ctx context.Context, out *UserProfileInsights) {
	out.RegistrationIPs = s.queryUserInsightDimensions(ctx, `
SELECT ip_address, COUNT(*)::bigint, MAX(created_at)
FROM user_registration_events
WHERE ip_address <> ''
GROUP BY ip_address
ORDER BY COUNT(*) DESC, MAX(created_at) DESC
LIMIT 10`)
	out.UserAgents = s.queryUserInsightDimensions(ctx, `
SELECT user_agent, COUNT(*)::bigint, MAX(created_at)
FROM user_registration_events
WHERE user_agent <> ''
GROUP BY user_agent
ORDER BY COUNT(*) DESC, MAX(created_at) DESC
LIMIT 10`)
	out.Languages = s.queryUserInsightDimensions(ctx, `
SELECT accept_language, COUNT(*)::bigint, MAX(created_at)
FROM user_registration_events
WHERE accept_language <> ''
GROUP BY accept_language
ORDER BY COUNT(*) DESC, MAX(created_at) DESC
LIMIT 10`)
}

func (s *adminServiceImpl) queryUserInsightDimensions(ctx context.Context, query string) []UserInsightDimension {
	rows, err := s.entClient.QueryContext(ctx, query)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()
	out := []UserInsightDimension{}
	for rows.Next() {
		var item UserInsightDimension
		var lastSeen sql.NullTime
		if err := rows.Scan(&item.Value, &item.Count, &lastSeen); err != nil {
			return out
		}
		item.LastSeen = nullableTimePtr(lastSeen)
		out = append(out, item)
	}
	return out
}

func (s *adminServiceImpl) loadUserInsightFunnel(ctx context.Context, out *UserProfileInsights) {
	var total, withKey, withUsage, withPaidOrder, withBusiness int64
	err := scanAdminProfileRow(ctx, s.entClient, `
WITH active_users AS (
  SELECT id FROM users WHERE deleted_at IS NULL
)
SELECT
  COUNT(*)::bigint,
  COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM api_keys ak WHERE ak.user_id = active_users.id AND ak.deleted_at IS NULL))::bigint,
  COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM usage_logs ul WHERE ul.user_id = active_users.id))::bigint,
  COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM payment_orders po WHERE po.user_id = active_users.id AND po.status IN ('PAID', 'RECHARGING', 'COMPLETED')))::bigint,
  COUNT(*) FILTER (
    WHERE EXISTS (SELECT 1 FROM image_workspace_tasks iwt WHERE iwt.user_id = active_users.id)
       OR EXISTS (SELECT 1 FROM wechat_export_tasks wet WHERE wet.user_id = active_users.id)
  )::bigint
FROM active_users`, []any{}, &total, &withKey, &withUsage, &withPaidOrder, &withBusiness)
	if err != nil {
		return
	}
	out.Funnel = []UserInsightFunnelStep{
		{Key: "registered", Label: "注册", Count: total, Conversion: 1},
		{Key: "api_key_created", Label: "创建 API Key", Count: withKey, Conversion: ratio(withKey, total)},
		{Key: "first_api_usage", Label: "产生调用", Count: withUsage, Conversion: ratio(withUsage, total)},
		{Key: "paid", Label: "成功支付", Count: withPaidOrder, Conversion: ratio(withPaidOrder, total)},
		{Key: "content_business", Label: "使用内容业务", Count: withBusiness, Conversion: ratio(withBusiness, total)},
	}
}

func (s *adminServiceImpl) loadUserInsightRiskSamples(ctx context.Context, out *UserProfileInsights, limit int) {
	rows, err := s.entClient.QueryContext(ctx, `
WITH sample_rows AS (
  SELECT
    u.id, u.email, COALESCE(u.username, '') AS username,
    '同 IP 注册偏多' AS label,
    '24 小时内同 IP 注册账号数偏高' AS reason,
    'warning' AS severity,
    COALESCE(e.ip_address, '') AS registration_ip,
    u.created_at,
    u.last_active_at
  FROM users u
  JOIN user_registration_events e ON e.user_id = u.id
  WHERE u.deleted_at IS NULL
    AND e.ip_address <> ''
    AND (
      SELECT COUNT(*)
      FROM user_registration_events e2
      WHERE e2.ip_address = e.ip_address
        AND e2.created_at BETWEEN e.created_at - INTERVAL '24 hours' AND e.created_at + INTERVAL '24 hours'
    ) >= 3

  UNION ALL

  SELECT
    u.id, u.email, COALESCE(u.username, '') AS username,
    '缺少身份绑定' AS label,
    '未找到 auth_identities 记录，可能是历史数据或人工创建' AS reason,
    'warning' AS severity,
    COALESCE(e.ip_address, '') AS registration_ip,
    u.created_at,
    u.last_active_at
  FROM users u
  LEFT JOIN user_registration_events e ON e.user_id = u.id
  WHERE u.deleted_at IS NULL
    AND u.role <> 'admin'
    AND NOT EXISTS (SELECT 1 FROM auth_identities ai WHERE ai.user_id = u.id)

  UNION ALL

  SELECT
    u.id, u.email, COALESCE(u.username, '') AS username,
    'Key 未产生调用' AS label,
    '存在活跃 API Key，但暂无调用记录' AS reason,
    'info' AS severity,
    COALESCE(e.ip_address, '') AS registration_ip,
    u.created_at,
    u.last_active_at
  FROM users u
  LEFT JOIN user_registration_events e ON e.user_id = u.id
  WHERE u.deleted_at IS NULL
    AND EXISTS (SELECT 1 FROM api_keys ak WHERE ak.user_id = u.id AND ak.deleted_at IS NULL AND ak.status = 'active')
    AND NOT EXISTS (SELECT 1 FROM usage_logs ul WHERE ul.user_id = u.id)

  UNION ALL

  SELECT
    u.id, u.email, COALESCE(u.username, '') AS username,
    '测试账号' AS label,
    '邮箱或用户名符合 smoke/test 模式' AS reason,
    'info' AS severity,
    COALESCE(e.ip_address, '') AS registration_ip,
    u.created_at,
    u.last_active_at
  FROM users u
  LEFT JOIN user_registration_events e ON e.user_id = u.id
  WHERE u.deleted_at IS NULL
    AND (lower(u.email) LIKE '%example.test%' OR lower(u.email) LIKE '%smoke%' OR lower(u.username) LIKE '%smoke%')
)
SELECT id, email, username, label, reason, severity, registration_ip, created_at, last_active_at
FROM sample_rows
ORDER BY CASE severity WHEN 'warning' THEN 0 ELSE 1 END, created_at DESC
LIMIT $1`, limit)
	if err != nil {
		return
	}
	defer func() { _ = rows.Close() }()
	seen := map[int64]struct{}{}
	for rows.Next() {
		var item UserInsightRiskSample
		var lastActive sql.NullTime
		if err := rows.Scan(&item.UserID, &item.Email, &item.Username, &item.Label, &item.Reason, &item.Severity, &item.RegistrationIP, &item.CreatedAt, &lastActive); err != nil {
			return
		}
		if _, ok := seen[item.UserID]; ok {
			continue
		}
		seen[item.UserID] = struct{}{}
		item.LastActiveAt = nullableTimePtr(lastActive)
		out.RiskSamples = append(out.RiskSamples, item)
	}
}

func ratio(value, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round((float64(value)/float64(total))*10000) / 10000
}

func userInsightSourceLabel(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "unknown"
	}
	return source
}
