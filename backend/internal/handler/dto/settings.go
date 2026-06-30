package dto

import (
	"encoding/json"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// CustomMenuItem represents a user-configured custom menu entry.
type CustomMenuItem struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	IconSVG    string `json:"icon_svg"`
	URL        string `json:"url"`
	PageSlug   string `json:"page_slug,omitempty"`
	Visibility string `json:"visibility"` // "user" or "admin"
	SortOrder  int    `json:"sort_order"`
}

// CustomEndpoint represents an admin-configured API endpoint for quick copy.
type CustomEndpoint struct {
	Name        string `json:"name"`
	Endpoint    string `json:"endpoint"`
	Description string `json:"description"`
}

type ModelPlazaItem struct {
	ID              string   `json:"id"`
	Provider        string   `json:"provider"`
	Title           string   `json:"title"`
	Badge           string   `json:"badge"`
	Description     string   `json:"description"`
	CapabilityTags  []string `json:"capability_tags"`
	ModelIDs        []string `json:"model_ids"`
	InputPrice      string   `json:"input_price"`
	OutputPrice     string   `json:"output_price"`
	CacheReadPrice  string   `json:"cache_read_price"`
	CacheWritePrice string   `json:"cache_write_price"`
	BillingBadge    string   `json:"billing_badge"`
	Visible         bool     `json:"visible"`
	SortOrder       int      `json:"sort_order"`
}

// SystemSettings represents the admin settings API response payload.
type SystemSettings struct {
	RegistrationEnabled              bool                     `json:"registration_enabled"`
	EmailVerifyEnabled               bool                     `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist []string                 `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool                     `json:"promo_code_enabled"`
	PasswordResetEnabled             bool                     `json:"password_reset_enabled"`
	PasswordMinLength                int                      `json:"password_min_length"`
	FrontendURL                      string                   `json:"frontend_url"`
	InvitationCodeEnabled            bool                     `json:"invitation_code_enabled"`
	TotpEnabled                      bool                     `json:"totp_enabled"`                   // TOTP 双因素认证
	TotpEncryptionKeyConfigured      bool                     `json:"totp_encryption_key_configured"` // TOTP 加密密钥是否已配置
	LoginAgreementEnabled            bool                     `json:"login_agreement_enabled"`
	LoginAgreementMode               string                   `json:"login_agreement_mode"`
	LoginAgreementUpdatedAt          string                   `json:"login_agreement_updated_at"`
	LoginAgreementDocuments          []LoginAgreementDocument `json:"login_agreement_documents"`

	SMTPHost               string              `json:"smtp_host"`
	SMTPPort               int                 `json:"smtp_port"`
	SMTPUsername           string              `json:"smtp_username"`
	SMTPPasswordConfigured bool                `json:"smtp_password_configured"`
	SMTPFrom               string              `json:"smtp_from_email"`
	SMTPFromName           string              `json:"smtp_from_name"`
	SMTPUseTLS             bool                `json:"smtp_use_tls"`
	SMTPDailyLimit         int                 `json:"smtp_daily_limit"`
	SMTPChannels           []SMTPChannelConfig `json:"smtp_channels"`

	TurnstileEnabled             bool   `json:"turnstile_enabled"`
	TurnstileSiteKey             string `json:"turnstile_site_key"`
	TurnstileSecretKeyConfigured bool   `json:"turnstile_secret_key_configured"`
	APIKeyACLTrustForwardedIP    bool   `json:"api_key_acl_trust_forwarded_ip"`

	LinuxDoConnectEnabled                bool   `json:"linuxdo_connect_enabled"`
	LinuxDoConnectClientID               string `json:"linuxdo_connect_client_id"`
	LinuxDoConnectClientSecretConfigured bool   `json:"linuxdo_connect_client_secret_configured"`
	LinuxDoConnectRedirectURL            string `json:"linuxdo_connect_redirect_url"`

	DingTalkConnectEnabled                 bool   `json:"dingtalk_connect_enabled"`
	DingTalkConnectClientID                string `json:"dingtalk_connect_client_id"`
	DingTalkConnectClientSecretConfigured  bool   `json:"dingtalk_connect_client_secret_configured"`
	DingTalkConnectRedirectURL             string `json:"dingtalk_connect_redirect_url"`
	DingTalkConnectCorpRestrictionPolicy   string `json:"dingtalk_connect_corp_restriction_policy"`
	DingTalkConnectInternalCorpID          string `json:"dingtalk_connect_internal_corp_id"`
	DingTalkConnectBypassRegistration      bool   `json:"dingtalk_connect_bypass_registration"`
	DingTalkConnectSyncCorpEmail           bool   `json:"dingtalk_connect_sync_corp_email"`
	DingTalkConnectSyncDisplayName         bool   `json:"dingtalk_connect_sync_display_name"`
	DingTalkConnectSyncDept                bool   `json:"dingtalk_connect_sync_dept"`
	DingTalkConnectSyncCorpEmailAttrKey    string `json:"dingtalk_connect_sync_corp_email_attr_key"`
	DingTalkConnectSyncDisplayNameAttrKey  string `json:"dingtalk_connect_sync_display_name_attr_key"`
	DingTalkConnectSyncDeptAttrKey         string `json:"dingtalk_connect_sync_dept_attr_key"`
	DingTalkConnectSyncCorpEmailAttrName   string `json:"dingtalk_connect_sync_corp_email_attr_name"`
	DingTalkConnectSyncDisplayNameAttrName string `json:"dingtalk_connect_sync_display_name_attr_name"`
	DingTalkConnectSyncDeptAttrName        string `json:"dingtalk_connect_sync_dept_attr_name"`

	WeChatConnectEnabled                   bool   `json:"wechat_connect_enabled"`
	WeChatConnectOpenAppID                 string `json:"wechat_connect_open_app_id"`
	WeChatConnectOpenAppSecretConfigured   bool   `json:"wechat_connect_open_app_secret_configured"`
	WeChatConnectMPAppID                   string `json:"wechat_connect_mp_app_id"`
	WeChatConnectMPAppSecretConfigured     bool   `json:"wechat_connect_mp_app_secret_configured"`
	WeChatConnectMobileAppID               string `json:"wechat_connect_mobile_app_id"`
	WeChatConnectMobileAppSecretConfigured bool   `json:"wechat_connect_mobile_app_secret_configured"`
	WeChatConnectOpenEnabled               bool   `json:"wechat_connect_open_enabled"`
	WeChatConnectMPEnabled                 bool   `json:"wechat_connect_mp_enabled"`
	WeChatConnectMobileEnabled             bool   `json:"wechat_connect_mobile_enabled"`
	WeChatConnectMode                      string `json:"wechat_connect_mode"`
	WeChatConnectScopes                    string `json:"wechat_connect_scopes"`
	WeChatConnectRedirectURL               string `json:"wechat_connect_redirect_url"`
	WeChatConnectFrontendRedirectURL       string `json:"wechat_connect_frontend_redirect_url"`

	OIDCConnectEnabled                bool   `json:"oidc_connect_enabled"`
	OIDCConnectProviderName           string `json:"oidc_connect_provider_name"`
	OIDCConnectClientID               string `json:"oidc_connect_client_id"`
	OIDCConnectClientSecretConfigured bool   `json:"oidc_connect_client_secret_configured"`
	OIDCConnectIssuerURL              string `json:"oidc_connect_issuer_url"`
	OIDCConnectDiscoveryURL           string `json:"oidc_connect_discovery_url"`
	OIDCConnectAuthorizeURL           string `json:"oidc_connect_authorize_url"`
	OIDCConnectTokenURL               string `json:"oidc_connect_token_url"`
	OIDCConnectUserInfoURL            string `json:"oidc_connect_userinfo_url"`
	OIDCConnectJWKSURL                string `json:"oidc_connect_jwks_url"`
	OIDCConnectScopes                 string `json:"oidc_connect_scopes"`
	OIDCConnectRedirectURL            string `json:"oidc_connect_redirect_url"`
	OIDCConnectFrontendRedirectURL    string `json:"oidc_connect_frontend_redirect_url"`
	OIDCConnectTokenAuthMethod        string `json:"oidc_connect_token_auth_method"`
	OIDCConnectUsePKCE                bool   `json:"oidc_connect_use_pkce"`
	OIDCConnectValidateIDToken        bool   `json:"oidc_connect_validate_id_token"`
	OIDCConnectAllowedSigningAlgs     string `json:"oidc_connect_allowed_signing_algs"`
	OIDCConnectClockSkewSeconds       int    `json:"oidc_connect_clock_skew_seconds"`
	OIDCConnectRequireEmailVerified   bool   `json:"oidc_connect_require_email_verified"`
	OIDCConnectUserInfoEmailPath      string `json:"oidc_connect_userinfo_email_path"`
	OIDCConnectUserInfoIDPath         string `json:"oidc_connect_userinfo_id_path"`
	OIDCConnectUserInfoUsernamePath   string `json:"oidc_connect_userinfo_username_path"`

	GitHubOAuthEnabled                bool   `json:"github_oauth_enabled"`
	GitHubOAuthClientID               string `json:"github_oauth_client_id"`
	GitHubOAuthClientSecretConfigured bool   `json:"github_oauth_client_secret_configured"`
	GitHubOAuthRedirectURL            string `json:"github_oauth_redirect_url"`
	GitHubOAuthFrontendRedirectURL    string `json:"github_oauth_frontend_redirect_url"`
	GoogleOAuthEnabled                bool   `json:"google_oauth_enabled"`
	GoogleOAuthClientID               string `json:"google_oauth_client_id"`
	GoogleOAuthClientSecretConfigured bool   `json:"google_oauth_client_secret_configured"`
	GoogleOAuthRedirectURL            string `json:"google_oauth_redirect_url"`
	GoogleOAuthFrontendRedirectURL    string `json:"google_oauth_frontend_redirect_url"`

	SiteName                     string           `json:"site_name"`
	SiteLogo                     string           `json:"site_logo"`
	SiteSubtitle                 string           `json:"site_subtitle"`
	APIBaseURL                   string           `json:"api_base_url"`
	ContactInfo                  string           `json:"contact_info"`
	DocURL                       string           `json:"doc_url"`
	DocsContentBasePath          string           `json:"docs_content_base_path"`
	HomeContent                  string           `json:"home_content"`
	HomeShellConfig              string           `json:"home_shell_config"`
	HomeBusinessShellConfig      string           `json:"home_business_shell_config"`
	ModelPlazaItems              []ModelPlazaItem `json:"model_plaza_items"`
	ImageWorkspaceModelConfig    string           `json:"image_workspace_model_config"`
	ModelPlazaShellConfig        string           `json:"model_plaza_shell_config"`
	DocsShellConfig              string           `json:"docs_shell_config"`
	LegalDocumentShellConfig     string           `json:"legal_document_shell_config"`
	APIKeysShellConfig           string           `json:"api_keys_shell_config"`
	KeyUsageShellConfig          string           `json:"key_usage_shell_config"`
	DashboardShellConfig         string           `json:"dashboard_shell_config"`
	UsageShellConfig             string           `json:"usage_shell_config"`
	APIGuideShellConfig          string           `json:"api_guide_shell_config"`
	APITestShellConfig           string           `json:"api_test_shell_config"`
	AvailableGroupsShellConfig   string           `json:"available_groups_shell_config"`
	RedeemShellConfig            string           `json:"redeem_shell_config"`
	AffiliateShellConfig         string           `json:"affiliate_shell_config"`
	AvailableChannelsShellConfig string           `json:"available_channels_shell_config"`
	ChannelStatusShellConfig     string           `json:"channel_status_shell_config"`
	CustomPageShellConfig        string           `json:"custom_page_shell_config"`
	ProfileShellConfig           string           `json:"profile_shell_config"`
	AuthShellConfig              string           `json:"auth_shell_config"`
	HideCcsImportButton          bool             `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled  bool             `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL      string           `json:"purchase_subscription_url"`
	TableDefaultPageSize         int              `json:"table_default_page_size"`
	TablePageSizeOptions         []int            `json:"table_page_size_options"`
	CustomMenuItems              []CustomMenuItem `json:"custom_menu_items"`
	CustomEndpoints              []CustomEndpoint `json:"custom_endpoints"`

	AppURL                     string `json:"app_url"`
	AppName                    string `json:"app_name"`
	AppDescription             string `json:"app_description"`
	AppLogo                    string `json:"app_logo"`
	AppFavicon                 string `json:"app_favicon"`
	AppPreviewImage            string `json:"app_preview_image"`
	Theme                      string `json:"theme"`
	Appearance                 string `json:"appearance"`
	DefaultLocale              string `json:"default_locale"`
	PromptCasesTitle           string `json:"prompt_cases_title"`
	PromptCasesDescription     string `json:"prompt_cases_description"`
	PromptTemplatesTitle       string `json:"prompt_templates_title"`
	PromptTemplatesDescription string `json:"prompt_templates_description"`
	WorkspaceShellConfig       string `json:"workspace_shell_config"`
	ImagePromptFilterConfig   string `json:"image_prompt_filter_config"`
	PricingTitle               string `json:"pricing_title"`
	PricingDescription         string `json:"pricing_description"`
	PricingShellConfig         string `json:"pricing_shell_config"`
	PaymentShellConfig         string `json:"payment_shell_config"`
	PricingCurrencySymbol      string `json:"pricing_currency_symbol"`
	CreditsTitle               string `json:"credits_title"`
	CreditsDescription         string `json:"credits_description"`
	CreditsPurchaseLabel       string `json:"credits_purchase_label"`
	CreditsBalanceLabel        string `json:"credits_balance_label"`
	CreditsPerBalance          string `json:"credits_per_balance"`
	LocaleDetectEnabled        bool   `json:"locale_detect_enabled"`
	EmailAuthVisible           bool   `json:"email_auth_visible"`
	GoogleAuthVisible          bool   `json:"google_auth_visible"`
	GitHubAuthVisible          bool   `json:"github_auth_visible"`

	PromptCatalogShellConfig  string `json:"prompt_catalog_shell_config"`
	CreditsShellConfig        string `json:"credits_shell_config"`
	GoogleAnalyticsID         string `json:"google_analytics_id"`
	ClarityID                 string `json:"clarity_id"`
	PlausibleDomain           string `json:"plausible_domain"`
	PlausibleSrc              string `json:"plausible_src"`
	OpenPanelClientID         string `json:"openpanel_client_id"`
	PublicIntegrationsEnabled bool   `json:"public_integrations_enabled"`
	VercelAnalyticsEnabled    bool   `json:"vercel_analytics_enabled"`
	AdsenseCode               string `json:"adsense_code"`
	AffonsoEnabled            bool   `json:"affonso_enabled"`
	AffonsoID                 string `json:"affonso_id"`
	AffonsoCookieDuration     string `json:"affonso_cookie_duration"`
	PromoteKitEnabled         bool   `json:"promotekit_enabled"`
	PromoteKitID              string `json:"promotekit_id"`
	CrispEnabled              bool   `json:"crisp_enabled"`
	CrispWebsiteID            string `json:"crisp_website_id"`
	TawkEnabled               bool   `json:"tawk_enabled"`
	TawkPropertyID            string `json:"tawk_property_id"`
	TawkWidgetID              string `json:"tawk_widget_id"`

	DefaultConcurrency           int                          `json:"default_concurrency"`
	DefaultBalance               float64                      `json:"default_balance"`
	AffiliateRebateRate          float64                      `json:"affiliate_rebate_rate"`
	AffiliateRebateFreezeHours   int                          `json:"affiliate_rebate_freeze_hours"`
	AffiliateRebateDurationDays  int                          `json:"affiliate_rebate_duration_days"`
	AffiliateRebatePerInviteeCap float64                      `json:"affiliate_rebate_per_invitee_cap"`
	DefaultUserRPMLimit          int                          `json:"default_user_rpm_limit"`
	DefaultSubscriptions         []DefaultSubscriptionSetting `json:"default_subscriptions"`

	// Model fallback configuration
	EnableModelFallback      bool   `json:"enable_model_fallback"`
	FallbackModelAnthropic   string `json:"fallback_model_anthropic"`
	FallbackModelOpenAI      string `json:"fallback_model_openai"`
	FallbackModelGemini      string `json:"fallback_model_gemini"`
	FallbackModelAntigravity string `json:"fallback_model_antigravity"`

	// Identity patch configuration (Claude -> Gemini)
	EnableIdentityPatch bool   `json:"enable_identity_patch"`
	IdentityPatchPrompt string `json:"identity_patch_prompt"`

	// Ops monitoring (vNext)
	OpsMonitoringEnabled         bool   `json:"ops_monitoring_enabled"`
	OpsRealtimeMonitoringEnabled bool   `json:"ops_realtime_monitoring_enabled"`
	OpsQueryModeDefault          string `json:"ops_query_mode_default"`
	OpsMetricsIntervalSeconds    int    `json:"ops_metrics_interval_seconds"`

	MinClaudeCodeVersion string `json:"min_claude_code_version"`
	MaxClaudeCodeVersion string `json:"max_claude_code_version"`

	// 分组隔离
	AllowUngroupedKeyScheduling bool `json:"allow_ungrouped_key_scheduling"`

	// Backend Mode
	BackendModeEnabled bool `json:"backend_mode_enabled"`

	// Gateway forwarding behavior
	EnableFingerprintUnification       bool   `json:"enable_fingerprint_unification"`
	EnableMetadataPassthrough          bool   `json:"enable_metadata_passthrough"`
	EnableCCHSigning                   bool   `json:"enable_cch_signing"`
	EnableAnthropicCacheTTL1hInjection bool   `json:"enable_anthropic_cache_ttl_1h_injection"`
	RewriteMessageCacheControl         bool   `json:"rewrite_message_cache_control"`
	AntigravityUserAgentVersion        string `json:"antigravity_user_agent_version"`
	OpenAICodexUserAgent               string `json:"openai_codex_user_agent"`
	OpenAIAllowClaudeCodeCodexPlugin   bool   `json:"openai_allow_claude_code_codex_plugin"`

	// Web Search Emulation
	WebSearchEmulationEnabled bool `json:"web_search_emulation_enabled"`

	// Payment visible method routing
	PaymentVisibleMethodAlipaySource  string `json:"payment_visible_method_alipay_source"`
	PaymentVisibleMethodWxpaySource   string `json:"payment_visible_method_wxpay_source"`
	PaymentVisibleMethodAlipayEnabled bool   `json:"payment_visible_method_alipay_enabled"`
	PaymentVisibleMethodWxpayEnabled  bool   `json:"payment_visible_method_wxpay_enabled"`

	// OpenAI account scheduling
	OpenAIAdvancedSchedulerEnabled bool `json:"openai_advanced_scheduler_enabled"`

	// Payment configuration
	PaymentEnabled                   bool                    `json:"payment_enabled"`
	PaymentMinAmount                 float64                 `json:"payment_min_amount"`
	PaymentMaxAmount                 float64                 `json:"payment_max_amount"`
	PaymentDailyLimit                float64                 `json:"payment_daily_limit"`
	PaymentOrderTimeoutMin           int                     `json:"payment_order_timeout_minutes"`
	PaymentMaxPendingOrders          int                     `json:"payment_max_pending_orders"`
	PaymentEnabledTypes              []string                `json:"payment_enabled_types"`
	PaymentBalanceDisabled           bool                    `json:"payment_balance_disabled"`
	PaymentBalanceRechargeMultiplier float64                 `json:"payment_balance_recharge_multiplier"`
	PaymentRechargeFeeRate           float64                 `json:"payment_recharge_fee_rate"`
	PaymentRechargeProducts          []RechargeProductConfig `json:"payment_recharge_products"`
	PaymentLoadBalanceStrat          string                  `json:"payment_load_balance_strategy"`
	PaymentProductNamePrefix         string                  `json:"payment_product_name_prefix"`
	PaymentProductNameSuffix         string                  `json:"payment_product_name_suffix"`
	PaymentHelpImageURL              string                  `json:"payment_help_image_url"`
	PaymentHelpText                  string                  `json:"payment_help_text"`

	// Cancel rate limit
	PaymentCancelRateLimitEnabled bool   `json:"payment_cancel_rate_limit_enabled"`
	PaymentCancelRateLimitMax     int    `json:"payment_cancel_rate_limit_max"`
	PaymentCancelRateLimitWindow  int    `json:"payment_cancel_rate_limit_window"`
	PaymentCancelRateLimitUnit    string `json:"payment_cancel_rate_limit_unit"`
	PaymentCancelRateLimitMode    string `json:"payment_cancel_rate_limit_window_mode"`

	// Force Alipay mobile clients to use QR code payment instead of mobile redirect
	PaymentAlipayForceQRCode bool `json:"payment_alipay_force_qrcode"`

	// 余额、订阅到期与账号限额通知
	BalanceLowNotifyEnabled         bool               `json:"balance_low_notify_enabled"`
	BalanceLowNotifyThreshold       float64            `json:"balance_low_notify_threshold"`
	BalanceLowNotifyRechargeURL     string             `json:"balance_low_notify_recharge_url"`
	SubscriptionExpiryNotifyEnabled bool               `json:"subscription_expiry_notify_enabled"`
	AccountQuotaNotifyEnabled       bool               `json:"account_quota_notify_enabled"`
	AccountQuotaNotifyEmails        []NotifyEmailEntry `json:"account_quota_notify_emails"`

	// Registration notification
	RegistrationNotifyEnabled          bool   `json:"registration_notify_enabled"`
	RegistrationNotifyProvider         string `json:"registration_notify_provider"`
	RegistrationNotifyWebhookURL       string `json:"registration_notify_webhook_url"`
	RegistrationNotifySecretConfigured bool   `json:"registration_notify_secret_configured"`

	// Channel Monitor feature switch
	ChannelMonitorEnabled                bool `json:"channel_monitor_enabled"`
	ChannelMonitorDefaultIntervalSeconds int  `json:"channel_monitor_default_interval_seconds"`

	// Available Channels feature switch (user-facing aggregate view)
	AvailableChannelsEnabled bool `json:"available_channels_enabled"`

	// 风控中心功能开关
	RiskControlEnabled bool `json:"risk_control_enabled"`

	// Affiliate (邀请返利) feature switch
	AffiliateEnabled bool `json:"affiliate_enabled"`

	// OpenAI fast/flex policy
	OpenAIFastPolicySettings *OpenAIFastPolicySettings `json:"openai_fast_policy_settings,omitempty"`

	// 系统全局默认平台配额（key = platform，nil/缺省 = 不限制）
	DefaultPlatformQuotas map[string]*service.DefaultPlatformQuotaSetting `json:"default_platform_quotas,omitempty"`
}

type DefaultSubscriptionSetting struct {
	GroupID      int64 `json:"group_id"`
	ValidityDays int   `json:"validity_days"`
}

type SMTPChannelConfig struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Enabled            bool   `json:"enabled"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Username           string `json:"username"`
	PasswordConfigured bool   `json:"password_configured"`
	From               string `json:"from_email"`
	FromName           string `json:"from_name"`
	UseTLS             bool   `json:"use_tls"`
	DailyLimit         int    `json:"daily_limit"`
	SortOrder          int    `json:"sort_order"`
}

type RechargeProductConfig struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Amount         float64  `json:"amount"`
	CreditedAmount float64  `json:"credited_amount"`
	CreemProductID string   `json:"creem_product_id,omitempty"`
	Badge          string   `json:"badge"`
	Recommended    bool     `json:"recommended"`
	Features       []string `json:"features"`
	SortOrder      int      `json:"sort_order"`
}

type PublicSettings struct {
	RegistrationEnabled              bool                     `json:"registration_enabled"`
	EmailVerifyEnabled               bool                     `json:"email_verify_enabled"`
	ForceEmailOnThirdPartySignup     bool                     `json:"force_email_on_third_party_signup"`
	RegistrationEmailSuffixWhitelist []string                 `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool                     `json:"promo_code_enabled"`
	PasswordResetEnabled             bool                     `json:"password_reset_enabled"`
	PasswordMinLength                int                      `json:"password_min_length"`
	InvitationCodeEnabled            bool                     `json:"invitation_code_enabled"`
	TotpEnabled                      bool                     `json:"totp_enabled"` // TOTP 双因素认证
	LoginAgreementEnabled            bool                     `json:"login_agreement_enabled"`
	LoginAgreementMode               string                   `json:"login_agreement_mode"`
	LoginAgreementUpdatedAt          string                   `json:"login_agreement_updated_at"`
	LoginAgreementRevision           string                   `json:"login_agreement_revision"`
	LoginAgreementDocuments          []LoginAgreementDocument `json:"login_agreement_documents"`
	TurnstileEnabled                 bool                     `json:"turnstile_enabled"`
	TurnstileSiteKey                 string                   `json:"turnstile_site_key"`
	SiteName                         string                   `json:"site_name"`
	SiteLogo                         string                   `json:"site_logo"`
	SiteSubtitle                     string                   `json:"site_subtitle"`
	APIBaseURL                       string                   `json:"api_base_url"`
	ContactInfo                      string                   `json:"contact_info"`
	DocURL                           string                   `json:"doc_url"`
	DocsContentBasePath              string                   `json:"docs_content_base_path"`
	HomeContent                      string                   `json:"home_content"`
	HomeShellConfig                  string                   `json:"home_shell_config"`
	HomeBusinessShellConfig          string                   `json:"home_business_shell_config"`
	ModelPlazaItems                  []ModelPlazaItem         `json:"model_plaza_items"`
	ImageWorkspaceModelConfig        string                   `json:"image_workspace_model_config"`
	ModelPlazaShellConfig            string                   `json:"model_plaza_shell_config"`
	DocsShellConfig                  string                   `json:"docs_shell_config"`
	LegalDocumentShellConfig         string                   `json:"legal_document_shell_config"`
	APIKeysShellConfig               string                   `json:"api_keys_shell_config"`
	KeyUsageShellConfig              string                   `json:"key_usage_shell_config"`
	DashboardShellConfig             string                   `json:"dashboard_shell_config"`
	UsageShellConfig                 string                   `json:"usage_shell_config"`
	APIGuideShellConfig              string                   `json:"api_guide_shell_config"`
	APITestShellConfig               string                   `json:"api_test_shell_config"`
	AvailableGroupsShellConfig       string                   `json:"available_groups_shell_config"`
	RedeemShellConfig                string                   `json:"redeem_shell_config"`
	AffiliateShellConfig             string                   `json:"affiliate_shell_config"`
	AvailableChannelsShellConfig     string                   `json:"available_channels_shell_config"`
	ChannelStatusShellConfig         string                   `json:"channel_status_shell_config"`
	CustomPageShellConfig            string                   `json:"custom_page_shell_config"`
	ProfileShellConfig               string                   `json:"profile_shell_config"`
	AuthShellConfig                  string                   `json:"auth_shell_config"`
	HideCcsImportButton              bool                     `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled      bool                     `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL          string                   `json:"purchase_subscription_url"`
	TableDefaultPageSize             int                      `json:"table_default_page_size"`
	TablePageSizeOptions             []int                    `json:"table_page_size_options"`
	CustomMenuItems                  []CustomMenuItem         `json:"custom_menu_items"`
	CustomEndpoints                  []CustomEndpoint         `json:"custom_endpoints"`
	DingTalkOAuthEnabled             bool                     `json:"dingtalk_oauth_enabled"`
	LinuxDoOAuthEnabled              bool                     `json:"linuxdo_oauth_enabled"`
	WeChatOAuthEnabled               bool                     `json:"wechat_oauth_enabled"`
	WeChatOAuthOpenEnabled           bool                     `json:"wechat_oauth_open_enabled"`
	WeChatOAuthMPEnabled             bool                     `json:"wechat_oauth_mp_enabled"`
	WeChatOAuthMobileEnabled         bool                     `json:"wechat_oauth_mobile_enabled"`
	OIDCOAuthEnabled                 bool                     `json:"oidc_oauth_enabled"`
	OIDCOAuthProviderName            string                   `json:"oidc_oauth_provider_name"`
	GitHubOAuthEnabled               bool                     `json:"github_oauth_enabled"`
	GoogleOAuthEnabled               bool                     `json:"google_oauth_enabled"`
	SoraClientEnabled                bool                     `json:"sora_client_enabled"`
	BackendModeEnabled               bool                     `json:"backend_mode_enabled"`
	PaymentEnabled                   bool                     `json:"payment_enabled"`
	Version                          string                   `json:"version"`
	DefaultLocale                    string                   `json:"default_locale"`
	BalanceLowNotifyEnabled          bool                     `json:"balance_low_notify_enabled"`
	AccountQuotaNotifyEnabled        bool                     `json:"account_quota_notify_enabled"`
	BalanceLowNotifyThreshold        float64                  `json:"balance_low_notify_threshold"`
	BalanceLowNotifyRechargeURL      string                   `json:"balance_low_notify_recharge_url"`

	ChannelMonitorEnabled                bool `json:"channel_monitor_enabled"`
	ChannelMonitorDefaultIntervalSeconds int  `json:"channel_monitor_default_interval_seconds"`

	AvailableChannelsEnabled bool `json:"available_channels_enabled"`

	AffiliateEnabled bool `json:"affiliate_enabled"`

	RiskControlEnabled bool `json:"risk_control_enabled"`

	PromptCasesTitle           string `json:"prompt_cases_title"`
	PromptCasesDescription     string `json:"prompt_cases_description"`
	PromptTemplatesTitle       string `json:"prompt_templates_title"`
	PromptTemplatesDescription string `json:"prompt_templates_description"`
	PromptCatalogShellConfig   string `json:"prompt_catalog_shell_config"`
	WorkspaceShellConfig       string `json:"workspace_shell_config"`
	ImagePromptFilterConfig   string `json:"image_prompt_filter_config"`
	PricingTitle               string `json:"pricing_title"`
	PricingDescription         string `json:"pricing_description"`
	PricingShellConfig         string `json:"pricing_shell_config"`
	PaymentShellConfig         string `json:"payment_shell_config"`
	PricingCurrencySymbol      string `json:"pricing_currency_symbol"`
	CreditsTitle               string `json:"credits_title"`
	CreditsDescription         string `json:"credits_description"`
	CreditsPurchaseLabel       string `json:"credits_purchase_label"`
	CreditsBalanceLabel        string `json:"credits_balance_label"`
	CreditsPerBalance          string `json:"credits_per_balance"`
	CreditsShellConfig         string `json:"credits_shell_config"`
	GoogleAnalyticsID          string `json:"google_analytics_id"`
	ClarityID                  string `json:"clarity_id"`
	PlausibleDomain            string `json:"plausible_domain"`
	PlausibleSrc               string `json:"plausible_src"`
	OpenPanelClientID          string `json:"openpanel_client_id"`
	PublicIntegrationsEnabled  bool   `json:"public_integrations_enabled"`
	VercelAnalyticsEnabled     bool   `json:"vercel_analytics_enabled"`
	AdsenseCode                string `json:"adsense_code"`
	AffonsoEnabled             bool   `json:"affonso_enabled"`
	AffonsoID                  string `json:"affonso_id"`
	AffonsoCookieDuration      string `json:"affonso_cookie_duration"`
	PromoteKitEnabled          bool   `json:"promotekit_enabled"`
	PromoteKitID               string `json:"promotekit_id"`
	CrispEnabled               bool   `json:"crisp_enabled"`
	CrispWebsiteID             string `json:"crisp_website_id"`
	TawkEnabled                bool   `json:"tawk_enabled"`
	TawkPropertyID             string `json:"tawk_property_id"`
	TawkWidgetID               string `json:"tawk_widget_id"`
}

type LoginAgreementDocument struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ContentMD string `json:"content_md"`
}

// OverloadCooldownSettings 529过载冷却配置 DTO
type OverloadCooldownSettings struct {
	Enabled         bool `json:"enabled"`
	CooldownMinutes int  `json:"cooldown_minutes"`
}

// RateLimit429CooldownSettings 429默认回避配置 DTO
type RateLimit429CooldownSettings struct {
	Enabled         bool `json:"enabled"`
	CooldownSeconds int  `json:"cooldown_seconds"`
}

// StreamTimeoutSettings 流超时处理配置 DTO
type StreamTimeoutSettings struct {
	Enabled                bool   `json:"enabled"`
	Action                 string `json:"action"`
	TempUnschedMinutes     int    `json:"temp_unsched_minutes"`
	ThresholdCount         int    `json:"threshold_count"`
	ThresholdWindowMinutes int    `json:"threshold_window_minutes"`
}

// RectifierSettings 请求整流器配置 DTO
type RectifierSettings struct {
	Enabled                  bool     `json:"enabled"`
	ThinkingSignatureEnabled bool     `json:"thinking_signature_enabled"`
	ThinkingBudgetEnabled    bool     `json:"thinking_budget_enabled"`
	APIKeySignatureEnabled   bool     `json:"apikey_signature_enabled"`
	APIKeySignaturePatterns  []string `json:"apikey_signature_patterns"`
}

// BetaPolicyRule Beta 策略规则 DTO
type BetaPolicyRule struct {
	BetaToken            string   `json:"beta_token"`
	Action               string   `json:"action"`
	Scope                string   `json:"scope"`
	ErrorMessage         string   `json:"error_message,omitempty"`
	ModelWhitelist       []string `json:"model_whitelist,omitempty"`
	FallbackAction       string   `json:"fallback_action,omitempty"`
	FallbackErrorMessage string   `json:"fallback_error_message,omitempty"`
}

// BetaPolicySettings Beta 策略配置 DTO
type BetaPolicySettings struct {
	Rules []BetaPolicyRule `json:"rules"`
}

// OpenAIFastPolicyRule OpenAI fast/flex 策略规则 DTO
type OpenAIFastPolicyRule struct {
	ServiceTier          string   `json:"service_tier"`
	Action               string   `json:"action"`
	Scope                string   `json:"scope"`
	ErrorMessage         string   `json:"error_message,omitempty"`
	ModelWhitelist       []string `json:"model_whitelist,omitempty"`
	FallbackAction       string   `json:"fallback_action,omitempty"`
	FallbackErrorMessage string   `json:"fallback_error_message,omitempty"`
}

// OpenAIFastPolicySettings OpenAI fast 策略配置 DTO
type OpenAIFastPolicySettings struct {
	Rules []OpenAIFastPolicyRule `json:"rules"`
}

// EmailTemplateEventOption 描述可编辑的通知邮件事件。
type EmailTemplateEventOption struct {
	Value       string `json:"value"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Optional    bool   `json:"optional,omitempty"`
}

// EmailTemplateSummary is shown in the admin email template list.
type EmailTemplateSummary struct {
	Event     string `json:"event"`
	Locale    string `json:"locale"`
	Subject   string `json:"subject"`
	IsCustom  bool   `json:"is_custom,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// EmailTemplateListResponse is returned by GET /admin/settings/email-templates.
type EmailTemplateListResponse struct {
	Events       []EmailTemplateEventOption `json:"events"`
	Locales      []string                   `json:"locales"`
	Templates    []EmailTemplateSummary     `json:"templates,omitempty"`
	Placeholders []string                   `json:"placeholders,omitempty"`
}

// EmailTemplateDetail is returned for a specific event/locale template.
type EmailTemplateDetail struct {
	Event        string   `json:"event"`
	Locale       string   `json:"locale"`
	Subject      string   `json:"subject"`
	HTML         string   `json:"html"`
	IsCustom     bool     `json:"is_custom,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
	Placeholders []string `json:"placeholders,omitempty"`
}

// UpdateEmailTemplateRequest updates a template override.
type UpdateEmailTemplateRequest struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// PreviewEmailTemplateRequest previews a template without saving it.
type PreviewEmailTemplateRequest struct {
	Event     string            `json:"event"`
	Locale    string            `json:"locale"`
	Subject   string            `json:"subject"`
	HTML      string            `json:"html"`
	Variables map[string]string `json:"variables,omitempty"`
}

// EmailTemplatePreviewResponse is the rendered preview payload.
type EmailTemplatePreviewResponse struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// ParseCustomMenuItems parses a JSON string into a slice of CustomMenuItem.
// Returns empty slice on empty/invalid input.
func ParseCustomMenuItems(raw string) []CustomMenuItem {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return []CustomMenuItem{}
	}
	var items []CustomMenuItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []CustomMenuItem{}
	}
	return items
}

// ParseUserVisibleMenuItems parses custom menu items and filters out admin-only entries.
func ParseUserVisibleMenuItems(raw string) []CustomMenuItem {
	items := ParseCustomMenuItems(raw)
	filtered := make([]CustomMenuItem, 0, len(items))
	for _, item := range items {
		if item.Visibility != "admin" {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// ParseCustomEndpoints parses a JSON string into a slice of CustomEndpoint.
// Returns empty slice on empty/invalid input.
func ParseCustomEndpoints(raw string) []CustomEndpoint {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return []CustomEndpoint{}
	}
	var items []CustomEndpoint
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []CustomEndpoint{}
	}
	return items
}

func ParseModelPlazaItems(raw string) []ModelPlazaItem {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return []ModelPlazaItem{}
	}
	var items []ModelPlazaItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []ModelPlazaItem{}
	}
	return NormalizeModelPlazaItems(items)
}

func NormalizeModelPlazaItems(items []ModelPlazaItem) []ModelPlazaItem {
	if len(items) == 0 {
		return []ModelPlazaItem{}
	}
	normalized := make([]ModelPlazaItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item.ID = strings.TrimSpace(item.ID)
		item.Provider = strings.TrimSpace(item.Provider)
		item.Title = strings.TrimSpace(item.Title)
		item.Badge = strings.TrimSpace(item.Badge)
		item.Description = strings.TrimSpace(item.Description)
		item.InputPrice = strings.TrimSpace(item.InputPrice)
		item.OutputPrice = strings.TrimSpace(item.OutputPrice)
		item.CacheReadPrice = strings.TrimSpace(item.CacheReadPrice)
		item.CacheWritePrice = strings.TrimSpace(item.CacheWritePrice)
		item.BillingBadge = strings.TrimSpace(item.BillingBadge)
		item.CapabilityTags = normalizeStringList(item.CapabilityTags)
		item.ModelIDs = normalizeStringList(item.ModelIDs)
		if item.ID == "" || item.Title == "" {
			continue
		}
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		normalized = append(normalized, item)
	}
	return normalized
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

// ImagePromptFilterConfig 图片提示词过滤配置 DTO
type ImagePromptFilterConfig struct {
	Enabled              bool     `json:"enabled"`
	ExplicitKeywords     []string `json:"explicit_keywords"`
	YouthContextKeywords []string `json:"youth_context_keywords"`
	WarningMessage       string   `json:"warning_message"`
	YouthWarningMessage  string   `json:"youth_warning_message"`
}
