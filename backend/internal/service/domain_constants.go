package service

import (
	"fmt"

	"github.com/Wei-Shaw/cloudbase/internal/domain"
)

// Status constants
const (
	StatusActive   = domain.StatusActive
	StatusDisabled = domain.StatusDisabled
	StatusError    = domain.StatusError
	StatusUnused   = domain.StatusUnused
	StatusUsed     = domain.StatusUsed
	StatusExpired  = domain.StatusExpired
)

// Role constants
const (
	RoleAdmin = domain.RoleAdmin
	RoleUser  = domain.RoleUser
)

// Affiliate rebate settings
const (
	AffiliateRebateRateDefault          = 20.0
	AffiliateRebateRateMin              = 0.0
	AffiliateRebateRateMax              = 100.0
	AffiliateEnabledDefault             = false // 邀请返利总开关默认关闭
	AffiliateRebateFreezeHoursDefault   = 0     // 0 = 不冻结（向后兼容）
	AffiliateRebateFreezeHoursMax       = 720   // 最大 30 天
	AffiliateRebateDurationDaysDefault  = 0     // 0 = 永久有效
	AffiliateRebateDurationDaysMax      = 3650  // ~10 年
	AffiliateRebatePerInviteeCapDefault = 0.0   // 0 = 无上限
)

// Platform constants
const (
	PlatformAnthropic   = domain.PlatformAnthropic
	PlatformOpenAI      = domain.PlatformOpenAI
	PlatformGemini      = domain.PlatformGemini
	PlatformAntigravity = domain.PlatformAntigravity
	PlatformGrok        = domain.PlatformGrok
)

// AllowedQuotaPlatforms 是允许设置 user × platform quota 的平台列表（单一权威来源）。
// ent/schema/user_platform_quota.go 的 Validate 函数独立维护（构建期约束），
// 若新增平台需同步修改该 schema。
var AllowedQuotaPlatforms = []string{
	PlatformAnthropic,
	PlatformOpenAI,
	PlatformGemini,
	PlatformAntigravity,
	PlatformGrok,
}

// IsAllowedQuotaPlatform 报告 s 是否为合法的 quota platform 标识。
func IsAllowedQuotaPlatform(s string) bool {
	for _, p := range AllowedQuotaPlatforms {
		if p == s {
			return true
		}
	}
	return false
}

// Account type constants
const (
	AccountTypeOAuth          = domain.AccountTypeOAuth          // OAuth类型账号（full scope: profile + inference）
	AccountTypeSetupToken     = domain.AccountTypeSetupToken     // Setup Token类型账号（inference only scope）
	AccountTypeAPIKey         = domain.AccountTypeAPIKey         // API Key类型账号
	AccountTypeUpstream       = domain.AccountTypeUpstream       // 上游透传类型账号（通过 Base URL + API Key 连接上游）
	AccountTypeGeminiWeb      = domain.AccountTypeGeminiWeb      // Gemini Web 登录态网关账号（通过 Base URL + API Key 连接 Gemini Web 代理）
	AccountTypeBedrock        = domain.AccountTypeBedrock        // AWS Bedrock 类型账号（通过 SigV4 签名或 API Key 连接 Bedrock，由 credentials.auth_mode 区分）
	AccountTypeServiceAccount = domain.AccountTypeServiceAccount // Google Service Account 类型账号（用于 Vertex AI）
)

// Redeem type constants
const (
	RedeemTypeBalance          = domain.RedeemTypeBalance
	RedeemTypeConcurrency      = domain.RedeemTypeConcurrency
	RedeemTypeSubscription     = domain.RedeemTypeSubscription
	RedeemTypeInvitation       = domain.RedeemTypeInvitation
	RedeemTypeAffiliateBalance = "affiliate_balance"
)

// PromoCode status constants
const (
	PromoCodeStatusActive   = domain.PromoCodeStatusActive
	PromoCodeStatusDisabled = domain.PromoCodeStatusDisabled
)

// Admin adjustment type constants
const (
	AdjustmentTypeAdminBalance     = domain.AdjustmentTypeAdminBalance     // 管理员调整余额
	AdjustmentTypeAdminConcurrency = domain.AdjustmentTypeAdminConcurrency // 管理员调整并发数
)

// Group subscription type constants
const (
	SubscriptionTypeStandard     = domain.SubscriptionTypeStandard     // 标准计费模式（按余额扣费）
	SubscriptionTypeSubscription = domain.SubscriptionTypeSubscription // 订阅模式（按限额控制）
)

// Subscription status constants
const (
	SubscriptionStatusActive    = domain.SubscriptionStatusActive
	SubscriptionStatusExpired   = domain.SubscriptionStatusExpired
	SubscriptionStatusSuspended = domain.SubscriptionStatusSuspended
	// SubscriptionStatusRevoked 是 soft-deleted 订阅的 API 展示态，不写入 status 字段。
	SubscriptionStatusRevoked = "revoked"
)

// Setting keys
const (
	// 注册设置
	SettingKeyRegistrationEnabled                        = "registration_enabled"                // 是否开放注册
	SettingKeyEmailVerifyEnabled                         = "email_verify_enabled"                // 是否开启邮件验证
	SettingKeyRegistrationEmailSuffixWhitelist           = "registration_email_suffix_whitelist" // 注册邮箱后缀白名单（JSON 数组）
	SettingKeyPromoCodeEnabled                           = "promo_code_enabled"                  // 是否启用优惠码功能
	SettingKeyPasswordResetEnabled                       = "password_reset_enabled"              // 是否启用忘记密码功能（需要先开启邮件验证）
	SettingKeyPasswordMinLength                          = "password_min_length"                 // 密码最小长度（下限 8）
	SettingKeyFrontendURL                                = "frontend_url"                        // 前端基础URL，用于生成邮件中的重置密码链接
	SettingKeyInvitationCodeEnabled                      = "invitation_code_enabled"             // 是否启用邀请码注册
	SettingKeyAffiliateEnabled                           = "affiliate_enabled"                   // 邀请返利功能总开关
	SettingKeyAffiliateRebateRate                        = "affiliate_rebate_rate"               // 邀请返利比例（百分比，0-100）
	SettingKeyAffiliateRebateFreezeHours                 = "affiliate_rebate_freeze_hours"       // 返利冻结期（小时，0=不冻结）
	SettingKeyAffiliateRebateDurationDays                = "affiliate_rebate_duration_days"      // 返利有效期（天，0=永久）
	SettingKeyAffiliateRebatePerInviteeCap               = "affiliate_rebate_per_invitee_cap"    // 单人返利上限（0=无上限）
	SettingKeyRiskControlEnabled                         = "risk_control_enabled"                // 是否启用风控中心入口与审计链路
	SettingKeySignupGrantRiskControlEnabled              = "signup_grant_risk_control_enabled"   // 是否启用注册赠送风控
	SettingKeySignupGrantRiskControlEmailLimit           = "signup_grant_risk_control_email_limit"
	SettingKeySignupGrantRiskControlIPLimit              = "signup_grant_risk_control_ip_daily_limit"
	SettingKeySignupGrantRiskControlDomainLimit          = "signup_grant_risk_control_domain_daily_limit"
	SettingKeySignupGrantRiskControlOAuthIdentityEnabled = "signup_grant_risk_control_oauth_identity_enabled"
	SettingKeySignupGrantRiskControlDeviceEnabled        = "signup_grant_risk_control_device_enabled"
	SettingKeySignupGrantRiskControlDeviceLimit          = "signup_grant_risk_control_device_daily_limit"
	SettingKeySignupGrantRiskControlFreeDomainLimit      = "signup_grant_risk_control_free_domain_daily_limit"
	SettingKeySignupGrantRiskControlBlockedDomains       = "signup_grant_risk_control_blocked_email_domains"
	SettingKeySignupGrantRiskControlFreeDomains          = "signup_grant_risk_control_free_email_domains"
	SettingKeySignupGrantRiskControlTrustedDomains       = "signup_grant_risk_control_trusted_email_domains"
	SettingKeyContentModerationConfig                    = "content_moderation_config"       // 内容审计配置（JSON）
	SettingKeyCyberSessionBlockEnabled                   = "cyber_session_block_enabled"     // cyber 命中后会话级自动屏蔽总开关(默认关)
	SettingKeyCyberSessionBlockTTLSeconds                = "cyber_session_block_ttl_seconds" // 会话屏蔽 TTL 秒数(默认 3600)
	SettingKeyLoginAgreementEnabled                      = "login_agreement_enabled"         // 登录前是否要求同意条款
	SettingKeyLoginAgreementMode                         = "login_agreement_mode"            // 条款确认展示模式：modal / checkbox
	SettingKeyLoginAgreementUpdatedAt                    = "login_agreement_updated_at"      // 条款更新日期（展示用）
	SettingKeyLoginAgreementDocuments                    = "login_agreement_documents"       // 条款文档列表（JSON，Markdown 内容）

	// 邮件服务设置
	SettingKeySMTPHost       = "smtp_host"        // SMTP服务器地址
	SettingKeySMTPPort       = "smtp_port"        // SMTP端口
	SettingKeySMTPUsername   = "smtp_username"    // SMTP用户名
	SettingKeySMTPPassword   = "smtp_password"    // SMTP密码（加密存储）
	SettingKeySMTPFrom       = "smtp_from"        // 发件人地址
	SettingKeySMTPFromName   = "smtp_from_name"   // 发件人名称
	SettingKeySMTPUseTLS     = "smtp_use_tls"     // 是否使用TLS
	SettingKeySMTPDailyLimit = "smtp_daily_limit" // 主SMTP通道日发送限额，0 表示不限
	SettingKeySMTPChannels   = "smtp_channels"    // 备用SMTP通道配置（JSON）

	// Cloudflare Turnstile 设置
	SettingKeyTurnstileEnabled   = "turnstile_enabled"    // 是否启用 Turnstile 验证
	SettingKeyTurnstileSiteKey   = "turnstile_site_key"   // Turnstile Site Key
	SettingKeyTurnstileSecretKey = "turnstile_secret_key" // Turnstile Secret Key

	// API Key IP 访问控制设置
	SettingKeyAPIKeyACLTrustForwardedIP = "api_key_acl_trust_forwarded_ip" // API Key IP 白/黑名单是否信任转发 IP

	// TOTP 双因素认证设置
	SettingKeyTotpEnabled = "totp_enabled" // 是否启用 TOTP 2FA 功能

	// WeChat Connect OAuth 登录设置
	SettingKeyWeChatConnectEnabled             = "wechat_connect_enabled"
	SettingKeyWeChatConnectOpenAppID           = "wechat_connect_open_app_id"
	SettingKeyWeChatConnectOpenAppSecret       = "wechat_connect_open_app_secret"
	SettingKeyWeChatConnectMPAppID             = "wechat_connect_mp_app_id"
	SettingKeyWeChatConnectMPAppSecret         = "wechat_connect_mp_app_secret"
	SettingKeyWeChatConnectMobileAppID         = "wechat_connect_mobile_app_id"
	SettingKeyWeChatConnectMobileAppSecret     = "wechat_connect_mobile_app_secret"
	SettingKeyWeChatConnectOpenEnabled         = "wechat_connect_open_enabled"
	SettingKeyWeChatConnectMPEnabled           = "wechat_connect_mp_enabled"
	SettingKeyWeChatConnectMobileEnabled       = "wechat_connect_mobile_enabled"
	SettingKeyWeChatConnectMode                = "wechat_connect_mode"
	SettingKeyWeChatConnectScopes              = "wechat_connect_scopes"
	SettingKeyWeChatConnectRedirectURL         = "wechat_connect_redirect_url"
	SettingKeyWeChatConnectFrontendRedirectURL = "wechat_connect_frontend_redirect_url"

	// GitHub / Google 邮箱快捷登录设置
	SettingKeyGitHubOAuthEnabled             = "github_oauth_enabled"
	SettingKeyGitHubOAuthClientID            = "github_oauth_client_id"
	SettingKeyGitHubOAuthClientSecret        = "github_oauth_client_secret"
	SettingKeyGitHubOAuthRedirectURL         = "github_oauth_redirect_url"
	SettingKeyGitHubOAuthFrontendRedirectURL = "github_oauth_frontend_redirect_url"
	SettingKeyGoogleOAuthEnabled             = "google_oauth_enabled"
	SettingKeyGoogleOAuthClientID            = "google_oauth_client_id"
	SettingKeyGoogleOAuthClientSecret        = "google_oauth_client_secret"
	SettingKeyGoogleOAuthRedirectURL         = "google_oauth_redirect_url"
	SettingKeyGoogleOAuthFrontendRedirectURL = "google_oauth_frontend_redirect_url"

	// OEM设置
	SettingKeySiteName                     = "site_name"                       // 网站名称
	SettingKeySiteLogo                     = "site_logo"                       // 网站Logo (base64)
	SettingKeySiteSubtitle                 = "site_subtitle"                   // 网站副标题
	SettingKeyAPIBaseURL                   = "api_base_url"                    // API端点地址（用于客户端配置和导入）
	SettingKeyContactInfo                  = "contact_info"                    // 客服联系方式
	SettingKeyDocURL                       = "doc_url"                         // 文档链接
	SettingKeyDocsContentBasePath          = "docs_content_base_path"          // Docsify 文档内容根路径配置（JSON 或路径）
	SettingKeyHomeContent                  = "home_content"                    // 首页内容（支持 Markdown/HTML，或 URL 作为 iframe src）
	SettingKeyHomeShellConfig              = "home_shell_config"               // 首页默认页面壳配置（JSON）
	SettingKeyHomeBusinessShellConfig      = "home_business_shell_config"      // 业务能力首页页面壳配置（JSON）
	SettingKeyModelPlazaItems              = "model_plaza_items"               // 模型广场配置（JSON 数组）
	SettingKeyImageWorkspaceModelConfig    = "image_workspace_model_config"    // 生图工作台模型配置（JSON）
	SettingKeyModelPlazaShellConfig        = "model_plaza_shell_config"        // 模型广场页面壳配置（JSON）
	SettingKeyDocsShellConfig              = "docs_shell_config"               // 文档页页面壳配置（JSON）
	SettingKeyLegalDocumentShellConfig     = "legal_document_shell_config"     // 法务文档页页面壳配置（JSON）
	SettingKeyAPIKeysShellConfig           = "api_keys_shell_config"           // API Keys 管理页页面壳配置（JSON）
	SettingKeyKeyUsageShellConfig          = "key_usage_shell_config"          // API Key 用量页页面壳配置（JSON）
	SettingKeyDashboardShellConfig         = "dashboard_shell_config"          // 用户仪表盘页面壳配置（JSON）
	SettingKeyUsageShellConfig             = "usage_shell_config"              // 用户用量历史页页面壳配置（JSON）
	SettingKeyAPIGuideShellConfig          = "api_guide_shell_config"          // API 调用说明页页面壳配置（JSON）
	SettingKeyAPITestShellConfig           = "api_test_shell_config"           // API 调用测试页页面壳配置（JSON）
	SettingKeyAvailableGroupsShellConfig   = "available_groups_shell_config"   // 可用分组页页面壳配置（JSON）
	SettingKeyRedeemShellConfig            = "redeem_shell_config"             // 兑换码页页面壳配置（JSON）
	SettingKeyAffiliateShellConfig         = "affiliate_shell_config"          // 邀请中心页页面壳配置（JSON）
	SettingKeyAvailableChannelsShellConfig = "available_channels_shell_config" // 可用渠道页页面壳配置（JSON）
	SettingKeyChannelStatusShellConfig     = "channel_status_shell_config"     // 渠道状态页页面壳配置（JSON）
	SettingKeyCustomPageShellConfig        = "custom_page_shell_config"        // 自定义页面壳配置（JSON）
	SettingKeyProfileShellConfig           = "profile_shell_config"            // 用户资料页页面壳配置（JSON）
	SettingKeyAuthShellConfig              = "auth_shell_config"               // 登录注册页页面壳配置（JSON）
	SettingKeyHideCcsImportButton          = "hide_ccs_import_button"          // 是否隐藏 API Keys 页面的导入 CCS 按钮
	SettingKeyPurchaseSubscriptionEnabled  = "purchase_subscription_enabled"   // 是否展示"购买订阅"页面入口
	SettingKeyPurchaseSubscriptionURL      = "purchase_subscription_url"       // "购买订阅"页面 URL（作为 iframe src）
	SettingKeyTableDefaultPageSize         = "table_default_page_size"         // 表格默认每页条数
	SettingKeyTablePageSizeOptions         = "table_page_size_options"         // 表格可选每页条数（JSON 数组）
	SettingKeyCustomMenuItems              = "custom_menu_items"               // 自定义菜单项（JSON 数组）
	SettingKeyCustomEndpoints              = "custom_endpoints"                // 自定义端点列表（JSON 数组）

	// Web frontend public runtime settings
	SettingKeyWebAppURL                    = "web_app_url"
	SettingKeyWebAppName                   = "web_app_name"
	SettingKeyWebAppDescription            = "web_app_description"
	SettingKeyWebAppLogo                   = "web_app_logo"
	SettingKeyWebAppFavicon                = "web_app_favicon"
	SettingKeyWebAppPreviewImage           = "web_app_preview_image"
	SettingKeyWebTheme                     = "web_theme"
	SettingKeyWebAppearance                = "web_appearance"
	SettingKeyWebDefaultLocale             = "web_default_locale"
	SettingKeyPromptCasesTitle             = "prompt_cases_title"
	SettingKeyPromptCasesDescription       = "prompt_cases_description"
	SettingKeyPromptTemplatesTitle         = "prompt_templates_title"
	SettingKeyPromptTemplatesDescription   = "prompt_templates_description"
	SettingKeyPromptCatalogShellConfig     = "prompt_catalog_shell_config"
	SettingKeyWorkspaceShellConfig         = "workspace_shell_config"
	SettingKeyImagePromptFilterConfig      = "image_prompt_filter_config"
	SettingKeyPricingTitle                 = "pricing_title"
	SettingKeyPricingDescription           = "pricing_description"
	SettingKeyPricingShellConfig           = "pricing_shell_config"
	SettingKeyPaymentShellConfig           = "payment_shell_config"
	SettingKeyPricingCurrencySymbol        = "pricing_currency_symbol"
	SettingKeyCreditsTitle                 = "credits_title"
	SettingKeyCreditsDescription           = "credits_description"
	SettingKeyCreditsPurchaseLabel         = "credits_purchase_label"
	SettingKeyCreditsBalanceLabel          = "credits_balance_label"
	SettingKeyCreditsPerBalance            = "credits_per_balance"
	SettingKeyCreditsShellConfig           = "credits_shell_config"
	SettingKeyWebLocaleDetectEnabled       = "web_locale_detect_enabled"
	SettingKeyWebEmailAuthVisible          = "web_email_auth_visible"
	SettingKeyWebGoogleAuthVisible         = "web_google_auth_visible"
	SettingKeyWebGitHubAuthVisible         = "web_github_auth_visible"
	SettingKeyWebGoogleAnalyticsID         = "web_google_analytics_id"
	SettingKeyWebClarityID                 = "web_clarity_id"
	SettingKeyWebPlausibleDomain           = "web_plausible_domain"
	SettingKeyWebPlausibleSrc              = "web_plausible_src"
	SettingKeyWebOpenPanelClientID         = "web_openpanel_client_id"
	SettingKeyWebPublicIntegrationsEnabled = "web_public_integrations_enabled"
	SettingKeyWebVercelAnalyticsEnabled    = "web_vercel_analytics_enabled"
	SettingKeyWebAdsenseCode               = "web_adsense_code"
	SettingKeyWebAffonsoEnabled            = "web_affonso_enabled"
	SettingKeyWebAffonsoID                 = "web_affonso_id"
	SettingKeyWebAffonsoCookieDuration     = "web_affonso_cookie_duration"
	SettingKeyWebPromoteKitEnabled         = "web_promotekit_enabled"
	SettingKeyWebPromoteKitID              = "web_promotekit_id"

	// Business worker runtime settings. Secrets and deployment paths stay in env.
	SettingKeyWeChatExportFetchRetries                 = "wechat_export_fetch_retries"
	SettingKeyWeChatExportFetchTimeoutMS               = "wechat_export_fetch_timeout_ms"
	SettingKeyWeChatExportWorkerConcurrency            = "wechat_export_worker_concurrency"
	SettingKeyWeChatExportWorkerIntervalMS             = "wechat_export_worker_interval_ms"
	SettingKeyWeChatExportWorkerLeaseSeconds           = "wechat_export_worker_lease_seconds"
	SettingKeyWeChatExportWorkerMaxBackoffMS           = "wechat_export_worker_max_backoff_ms"
	SettingKeyImageWorkspaceUpstreamURL                = "image_workspace_upstream_url"
	SettingKeyImageWorkspaceGenerationTimeoutMS        = "image_workspace_generation_timeout_ms"
	SettingKeyImageWorkspaceCompletionCost             = "image_workspace_completion_cost"
	SettingKeyImageWorkspaceCompletionCostMapJSON      = "image_workspace_completion_cost_map_json"
	SettingKeyImageWorkspacePromptSafetyEnabled        = "image_workspace_prompt_safety_enabled"
	SettingKeyImageWorkspaceAssumeWorkerReady          = "image_workspace_assume_worker_ready"
	SettingKeyImageWorkspaceObjectStorageEnabled       = "image_workspace_object_storage_enabled"
	SettingKeyImageWorkspaceObjectStorageProvider      = "image_workspace_object_storage_provider"
	SettingKeyImageWorkspaceObjectStorageBucket        = "image_workspace_object_storage_bucket"
	SettingKeyImageWorkspaceObjectStorageRegion        = "image_workspace_object_storage_region"
	SettingKeyImageWorkspaceObjectStoragePrefix        = "image_workspace_object_storage_prefix"
	SettingKeyImageWorkspaceObjectStoragePublicBaseURL = "image_workspace_object_storage_public_base_url"
	SettingKeyMediaCDNBaseURL                          = "media_cdn_base_url"
	SettingKeyWebCrispEnabled                          = "web_crisp_enabled"
	SettingKeyWebCrispWebsiteID                        = "web_crisp_website_id"
	SettingKeyWebTawkEnabled                           = "web_tawk_enabled"
	SettingKeyWebTawkPropertyID                        = "web_tawk_property_id"
	SettingKeyWebTawkWidgetID                          = "web_tawk_widget_id"

	// 默认配置
	SettingKeyDefaultConcurrency   = "default_concurrency"    // 新用户默认并发量
	SettingKeyDefaultBalance       = "default_balance"        // 新用户默认余额
	SettingKeyDefaultSubscriptions = "default_subscriptions"  // 新用户默认订阅列表（JSON）
	SettingKeyDefaultUserRPMLimit  = "default_user_rpm_limit" // 新用户默认 RPM 限制（0 = 不限制）

	// 第三方认证来源默认授予配置
	SettingKeyAuthSourceDefaultEmailBalance           = "auth_source_default_email_balance"
	SettingKeyAuthSourceDefaultEmailConcurrency       = "auth_source_default_email_concurrency"
	SettingKeyAuthSourceDefaultEmailSubscriptions     = "auth_source_default_email_subscriptions"
	SettingKeyAuthSourceDefaultEmailGrantOnSignup     = "auth_source_default_email_grant_on_signup"
	SettingKeyAuthSourceDefaultEmailGrantOnFirstBind  = "auth_source_default_email_grant_on_first_bind"
	SettingKeyAuthSourceDefaultGitHubBalance          = "auth_source_default_github_balance"
	SettingKeyAuthSourceDefaultGitHubConcurrency      = "auth_source_default_github_concurrency"
	SettingKeyAuthSourceDefaultGitHubSubscriptions    = "auth_source_default_github_subscriptions"
	SettingKeyAuthSourceDefaultGitHubGrantOnSignup    = "auth_source_default_github_grant_on_signup"
	SettingKeyAuthSourceDefaultGitHubGrantOnFirstBind = "auth_source_default_github_grant_on_first_bind"
	SettingKeyAuthSourceDefaultGoogleBalance          = "auth_source_default_google_balance"
	SettingKeyAuthSourceDefaultGoogleConcurrency      = "auth_source_default_google_concurrency"
	SettingKeyAuthSourceDefaultGoogleSubscriptions    = "auth_source_default_google_subscriptions"
	SettingKeyAuthSourceDefaultGoogleGrantOnSignup    = "auth_source_default_google_grant_on_signup"
	SettingKeyAuthSourceDefaultGoogleGrantOnFirstBind = "auth_source_default_google_grant_on_first_bind"
	SettingKeyForceEmailOnThirdPartySignup            = "force_email_on_third_party_signup"

	// 管理员 API Key
	SettingKeyAdminAPIKey = "admin_api_key" // 全局管理员 API Key（用于外部系统集成）

	// Gemini 配额策略（JSON）
	SettingKeyGeminiQuotaPolicy = "gemini_quota_policy"

	// Model fallback settings
	SettingKeyEnableModelFallback      = "enable_model_fallback"
	SettingKeyFallbackModelAnthropic   = "fallback_model_anthropic"
	SettingKeyFallbackModelOpenAI      = "fallback_model_openai"
	SettingKeyFallbackModelGemini      = "fallback_model_gemini"
	SettingKeyFallbackModelAntigravity = "fallback_model_antigravity"

	// Request identity patch (Claude -> Gemini systemInstruction injection)
	SettingKeyEnableIdentityPatch = "enable_identity_patch"
	SettingKeyIdentityPatchPrompt = "identity_patch_prompt"

	// =========================
	// Ops Monitoring (vNext)
	// =========================

	// SettingKeyOpsMonitoringEnabled is a DB-backed soft switch to enable/disable ops module at runtime.
	SettingKeyOpsMonitoringEnabled = "ops_monitoring_enabled"

	// SettingKeyOpsRealtimeMonitoringEnabled controls realtime features (e.g. WS/QPS push).
	SettingKeyOpsRealtimeMonitoringEnabled = "ops_realtime_monitoring_enabled"

	// SettingKeyOpsQueryModeDefault controls the default query mode for ops dashboard (auto/raw/preagg).
	SettingKeyOpsQueryModeDefault = "ops_query_mode_default"

	// SettingKeyOpsEmailNotificationConfig stores JSON config for ops email notifications.
	SettingKeyOpsEmailNotificationConfig = "ops_email_notification_config"

	// SettingKeyOpsAlertRuntimeSettings stores JSON config for ops alert evaluator runtime settings.
	SettingKeyOpsAlertRuntimeSettings = "ops_alert_runtime_settings"

	// SettingKeyOpsMetricsIntervalSeconds controls the ops metrics collector interval (>=60).
	SettingKeyOpsMetricsIntervalSeconds = "ops_metrics_interval_seconds"

	// SettingKeyOpsAdvancedSettings stores JSON config for ops advanced settings (data retention, aggregation).
	SettingKeyOpsAdvancedSettings = "ops_advanced_settings"

	// SettingKeyOpsRuntimeLogConfig stores JSON config for runtime log settings.
	SettingKeyOpsRuntimeLogConfig = "ops_runtime_log_config"

	// =========================
	// Channel Monitor (渠道监控)
	// =========================

	// SettingKeyChannelMonitorEnabled is a DB-backed soft switch for the channel monitor feature.
	// When false: runner skips scheduling and user-facing endpoints return an empty list.
	SettingKeyChannelMonitorEnabled = "channel_monitor_enabled"

	// SettingKeyChannelMonitorDefaultIntervalSeconds controls the default interval (seconds)
	// pre-filled when creating a new channel monitor from the admin UI. Range: [15, 3600].
	SettingKeyChannelMonitorDefaultIntervalSeconds = "channel_monitor_default_interval_seconds"

	// SettingKeyAvailableChannelsEnabled is a DB-backed soft switch for the "Available Channels"
	// user-facing aggregate view. When false: user endpoint returns an empty list and the
	// sidebar entry is hidden. Defaults to false (opt-in feature).
	SettingKeyAvailableChannelsEnabled = "available_channels_enabled"

	// =========================
	// Overload Cooldown (529)
	// =========================

	// SettingKeyOverloadCooldownSettings stores JSON config for 529 overload cooldown handling.
	SettingKeyOverloadCooldownSettings = "overload_cooldown_settings"

	// SettingKeyRateLimit429CooldownSettings stores JSON config for 429 fallback cooldown handling.
	SettingKeyRateLimit429CooldownSettings = "rate_limit_429_cooldown_settings"

	// =========================
	// Stream Timeout Handling
	// =========================

	// SettingKeyStreamTimeoutSettings stores JSON config for stream timeout handling.
	SettingKeyStreamTimeoutSettings = "stream_timeout_settings"

	// =========================
	// Request Rectifier (请求整流器)
	// =========================

	// SettingKeyRectifierSettings stores JSON config for rectifier settings (thinking signature + budget).
	SettingKeyRectifierSettings = "rectifier_settings"

	// =========================
	// Beta Policy Settings
	// =========================

	// SettingKeyBetaPolicySettings stores JSON config for beta policy rules.
	SettingKeyBetaPolicySettings = "beta_policy_settings"

	// SettingKeyOpenAIFastPolicySettings stores JSON config for OpenAI
	// service_tier (fast/flex) policy rules. Mirrors BetaPolicySettings but
	// targets OpenAI's body-level service_tier field instead of Claude's
	// anthropic-beta header.
	SettingKeyOpenAIFastPolicySettings = "openai_fast_policy_settings"

	// =========================
	// Claude Code Version Check
	// =========================

	// SettingKeyMinClaudeCodeVersion 最低 Claude Code 版本号要求 (semver, 如 "2.1.0"，空值=不检查)
	SettingKeyMinClaudeCodeVersion = "min_claude_code_version"
	// SettingKeyMinCodexVersion 最低 Codex 引擎版本要求 (semver, 如 "0.141.0"，空值=不检查)
	SettingKeyMinCodexVersion = "min_codex_version"
	// SettingKeyMaxCodexVersion 最高 Codex 引擎版本限制 (semver, 如 "0.200.0"，空值=不检查)
	SettingKeyMaxCodexVersion = "max_codex_version"
	// SettingKeyCodexCLIOnlyBlacklist codex_cli_only 全局黑名单（[]AllowedClientEntry JSON，OR deny）。
	SettingKeyCodexCLIOnlyBlacklist = "codex_cli_only_blacklist"
	// SettingKeyCodexCLIOnlyWhitelist codex_cli_only 全局白名单（[]AllowedClientEntry JSON，双因子 AND allow）。
	SettingKeyCodexCLIOnlyWhitelist = "codex_cli_only_whitelist"
	// SettingKeyCodexCLIOnlyAllowAppServerClients App Server 开关：对未列名客户端开闸（默认 false；仅显式 "true" 开）。
	SettingKeyCodexCLIOnlyAllowAppServerClients = "codex_cli_only_allow_app_server_clients"
	// SettingKeyCodexCLIOnlyAllowBodyEngineFingerprint 引擎门 body 通道开关：接受 client_metadata 引擎指纹（默认 false；仅显式 "true" 开）。(已废弃，迁移并入信号列表)
	SettingKeyCodexCLIOnlyAllowBodyEngineFingerprint = "codex_cli_only_allow_body_engine_fingerprint"
	// SettingKeyCodexCLIOnlyEngineFingerprintSignals codex_cli_only 引擎指纹门信号列表（[]EngineFingerprintSignal JSON）。
	// 勾选(required)信号之间 AND;每条 match 变体行内 OR;缺失/空/非法 → 默认种子(只勾 x-codex-)。
	SettingKeyCodexCLIOnlyEngineFingerprintSignals = "codex_cli_only_engine_fingerprint_signals"

	// SettingKeyMaxClaudeCodeVersion 最高 Claude Code 版本号限制 (semver, 如 "3.0.0"，空值=不检查)
	SettingKeyMaxClaudeCodeVersion = "max_claude_code_version"

	// SettingKeyAllowUngroupedKeyScheduling 允许未分组 API Key 调度（默认 false：未分组 Key 返回 403）
	SettingKeyAllowUngroupedKeyScheduling = "allow_ungrouped_key_scheduling"

	// SettingKeyBackendModeEnabled Backend 模式：禁用用户注册和自助服务，仅管理员可登录
	SettingKeyBackendModeEnabled = "backend_mode_enabled"

	// Gateway Forwarding Behavior
	// SettingKeyEnableFingerprintUnification 是否统一 OAuth 账号的 X-Stainless-* 指纹头（默认 true）
	SettingKeyEnableFingerprintUnification = "enable_fingerprint_unification"
	// SettingKeyEnableMetadataPassthrough 是否透传客户端原始 metadata.user_id（默认 false）
	SettingKeyEnableMetadataPassthrough = "enable_metadata_passthrough"
	// SettingKeyEnableCCHSigning 已废弃（no-op）：新版 Claude Code CLI 已取消 cch 签名字段，
	// 网关随之不再注入/签名 cch（见 buildBillingAttributionText）。保留该 key 仅为向后兼容，
	// 开关不再产生任何效果。
	SettingKeyEnableCCHSigning = "enable_cch_signing"
	// SettingKeyEnableClaudeOAuthSystemPromptInjection 是否对 Claude OAuth mimic 路径注入 Claude Code system blocks（默认 true）
	SettingKeyEnableClaudeOAuthSystemPromptInjection = "enable_claude_oauth_system_prompt_injection"
	// SettingKeyClaudeOAuthSystemPrompt Claude OAuth mimic 路径注入的通用扩展 system prompt（空值使用内置默认）
	SettingKeyClaudeOAuthSystemPrompt = "claude_oauth_system_prompt"
	// SettingKeyClaudeOAuthSystemPromptBlocks Claude OAuth mimic 路径注入的 system blocks JSON 配置（空值使用内置默认）
	SettingKeyClaudeOAuthSystemPromptBlocks = "claude_oauth_system_prompt_blocks"
	// SettingKeyEnableAnthropicCacheTTL1hInjection 是否对 Anthropic OAuth/SetupToken 请求体注入 1h cache_control ttl（默认 false）
	SettingKeyEnableAnthropicCacheTTL1hInjection = "enable_anthropic_cache_ttl_1h_injection"
	// SettingKeyEnableClientDatelineNormalization 是否对 Anthropic OAuth/SetupToken 账号
	// 的 /v1/messages 请求体做客户端 dateline 归一化（默认 true）。
	// 归一化把 system prompt / <system-reminder> 块中 "Today's date is …" 语句里的
	// 非 ASCII 撇号与 "/" 日期分隔符还原为 ASCII 撇号 + "-" 分隔符，抹除某些客户端
	// 在检测到非官方 base URL 时注入的 3 bit 隐写指纹。仅适用于 Anthropic OAuth/SetupToken
	// 账号；API Key 账号不受影响。
	SettingKeyEnableClientDatelineNormalization = "enable_client_dateline_normalization"
	// SettingKeyRewriteMessageCacheControl 是否改写 messages[*].content[*].cache_control（默认 false）
	SettingKeyRewriteMessageCacheControl = "rewrite_message_cache_control"
	// SettingKeyAntigravityUserAgentVersion Antigravity 上游 User-Agent 版本号（空值使用环境变量/默认值）
	SettingKeyAntigravityUserAgentVersion = "antigravity_user_agent_version"
	// SettingKeyOpenAICodexUserAgent OpenAI Codex 完整 User-Agent（空值使用内置默认）
	// 当客户端 UA 被识别为浏览器（Chrome/Firefox/Safari/Edge 等）时，转发给 OpenAI 上游前会替换为此值，
	// 用于避免 Cloudflare 对浏览器型 UA 的质询拦截。
	SettingKeyOpenAICodexUserAgent = "openai_codex_user_agent"
	// SettingKeyOpenAIAllowClaudeCodeCodexPlugin 已废弃：历史全局开关只作为升级迁移输入读取。
	// 迁移后等价规则写入 SettingKeyCodexCLIOnlyWhitelist，不再参与运行时判定。
	SettingKeyOpenAIAllowClaudeCodeCodexPlugin = "openai_allow_claude_code_codex_plugin"

	// 余额不足提醒
	SettingKeyBalanceLowNotifyEnabled     = "balance_low_notify_enabled"      // 全局开关
	SettingKeyBalanceLowNotifyThreshold   = "balance_low_notify_threshold"    // 默认阈值（USD）
	SettingKeyBalanceLowNotifyRechargeURL = "balance_low_notify_recharge_url" // 充值页面 URL

	// 订阅到期提醒
	SettingKeySubscriptionExpiryNotifyEnabled = "subscription_expiry_notify_enabled" // 订阅到期提醒全局开关，默认开启

	// 账号限额通知
	SettingKeyAccountQuotaNotifyEnabled = "account_quota_notify_enabled" // 全局开关
	SettingKeyAccountQuotaNotifyEmails  = "account_quota_notify_emails"  // 管理员通知邮箱列表（JSON 数组）

	// Registration Notification
	SettingKeyRegistrationNotifyEnabled    = "registration_notify_enabled"     // 全局开关
	SettingKeyRegistrationNotifyProvider   = "registration_notify_provider"    // dingtalk | feishu
	SettingKeyRegistrationNotifyWebhookURL = "registration_notify_webhook_url" // 机器人 Webhook 地址
	SettingKeyRegistrationNotifySecret     = "registration_notify_secret"      // 机器人签名密钥

	// Web Search Emulation
	SettingKeyWebSearchEmulationConfig = "web_search_emulation_config" // JSON 配置
)

// SettingKeyDefaultPlatformQuotas —— 系统全局：每用户 × 平台日/周/月 USD 上限（JSON）。
// 值为 map[platform]{daily,weekly,monthly}，null/缺省 = 不限制；0 = 禁用；>0 = USD 上限。
const SettingKeyDefaultPlatformQuotas = "default_platform_quotas"

// SettingKeyAuthSourcePlatformQuotas 返回某 auth source 的 platform quota JSON key。
// 形如 auth_source_default_{source}_platform_quotas
func SettingKeyAuthSourcePlatformQuotas(source string) string {
	return fmt.Sprintf("auth_source_default_%s_platform_quotas", source)
}

// QuotaDimension constants for spark shadow accounts.
const (
	QuotaDimensionGlobal = "global"
	QuotaDimensionSpark  = "spark"
)

// AdminAPIKeyPrefix is the prefix for admin API keys (distinct from user "sk-" keys).
const AdminAPIKeyPrefix = "admin-"

// SettingKeyAllowUserViewErrorRequests controls whether end users can view
// their own failed requests on the usage page. Default false (opt-in).
const SettingKeyAllowUserViewErrorRequests = "allow_user_view_error_requests"
