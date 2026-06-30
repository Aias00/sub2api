package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/imroc/req/v3"
	"golang.org/x/sync/singleflight"
)

// CoerceDingTalkCorpPolicyForWrite 是 coerceDeprecatedDingTalkCorpPolicy 的导出版本，
// 用于 admin handler 在写入路径上对客户端直传的入参做防御性 coerce（前端 UI 虽已无 whitelist 选项，
// 但 API 可被直接调用）。
func CoerceDingTalkCorpPolicyForWrite(policy string) string {
	return coerceDeprecatedDingTalkCorpPolicy(policy)
}

// coerceDeprecatedDingTalkCorpPolicy 把已废弃的 corp_restriction_policy 值替换成安全的等价值。
// 升级前残留在 DB 中的 "whitelist" 会导致 callback 链路在 default case 静默 fail-closed
// （所有钉钉登录被拒）。这里统一退化为 "none" 让服务保持可用，并 warn 日志提醒 admin 重新保存设置。
func coerceDeprecatedDingTalkCorpPolicy(policy string) string {
	if policy == "whitelist" {
		slog.Warn("dingtalk: corp_restriction_policy=whitelist is deprecated and unsupported, coercing to none",
			"hint", "re-save DingTalk settings in admin UI to clear this warning")
		return "none"
	}
	return policy
}

var (
	ErrRegistrationDisabled   = infraerrors.Forbidden("REGISTRATION_DISABLED", "registration is currently disabled")
	ErrSettingNotFound        = infraerrors.NotFound("SETTING_NOT_FOUND", "setting not found")
	ErrDefaultSubGroupInvalid = infraerrors.BadRequest(
		"DEFAULT_SUBSCRIPTION_GROUP_INVALID",
		"default subscription group must exist and be subscription type",
	)
	ErrDefaultSubGroupDuplicate = infraerrors.BadRequest(
		"DEFAULT_SUBSCRIPTION_GROUP_DUPLICATE",
		"default subscription group cannot be duplicated",
	)
)

type SettingRepository interface {
	Get(ctx context.Context, key string) (*Setting, error)
	GetValue(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	GetMultiple(ctx context.Context, keys []string) (map[string]string, error)
	SetMultiple(ctx context.Context, settings map[string]string) error
	GetAll(ctx context.Context) (map[string]string, error)
	Delete(ctx context.Context, key string) error
}

type SiteLogoImage struct {
	ContentType string
	Data        []byte
	ETag        string
}

// cachedVersionBounds 缓存 Claude Code 版本号上下限（进程内缓存，60s TTL）
type cachedVersionBounds struct {
	min       string // 空字符串 = 不检查
	max       string // 空字符串 = 不检查
	expiresAt int64  // unix nano
}

// versionBoundsCache 版本号上下限进程内缓存
var versionBoundsCache atomic.Value // *cachedVersionBounds

// versionBoundsSF 防止缓存过期时 thundering herd
var versionBoundsSF singleflight.Group

// versionBoundsCacheTTL 缓存有效期
const versionBoundsCacheTTL = 60 * time.Second

// versionBoundsErrorTTL DB 错误时的短缓存，快速重试
const versionBoundsErrorTTL = 5 * time.Second

// versionBoundsDBTimeout singleflight 内 DB 查询超时，独立于请求 context
const versionBoundsDBTimeout = 5 * time.Second

// cachedBackendMode Backend Mode cache (in-process, 60s TTL)
type cachedBackendMode struct {
	value     bool
	expiresAt int64 // unix nano
}

var backendModeCache atomic.Value // *cachedBackendMode
var backendModeSF singleflight.Group

const backendModeCacheTTL = 60 * time.Second
const backendModeErrorTTL = 5 * time.Second
const backendModeDBTimeout = 5 * time.Second

// cachedGatewayForwardingSettings 缓存网关转发行为设置（进程内缓存，60s TTL）
type cachedGatewayForwardingSettings struct {
	fingerprintUnification       bool
	metadataPassthrough          bool
	cchSigning                   bool
	anthropicCacheTTL1hInjection bool
	rewriteMessageCacheControl   bool
	expiresAt                    int64 // unix nano
}

var gatewayForwardingCache atomic.Value // *cachedGatewayForwardingSettings
var gatewayForwardingSF singleflight.Group

const gatewayForwardingCacheTTL = 60 * time.Second
const gatewayForwardingErrorTTL = 5 * time.Second
const gatewayForwardingDBTimeout = 5 * time.Second

// cachedAntigravityUserAgentVersion 缓存 Antigravity UA 版本号（进程内缓存，60s TTL）
type cachedAntigravityUserAgentVersion struct {
	version   string
	expiresAt int64 // unix nano
}

const antigravityUserAgentVersionCacheTTL = 60 * time.Second
const antigravityUserAgentVersionErrorTTL = 5 * time.Second
const antigravityUserAgentVersionDBTimeout = 5 * time.Second

// DefaultOpenAICodexUserAgent OpenAI Codex 默认 User-Agent（用于规避 Cloudflare 对浏览器 UA 的质询）
const DefaultOpenAICodexUserAgent = "codex-tui/0.125.0 (Ubuntu 22.4.0; x86_64) xterm-256color (codex-tui; 0.125.0)"

// cachedOpenAICodexUserAgent 缓存 OpenAI Codex UA（进程内缓存，60s TTL）
type cachedOpenAICodexUserAgent struct {
	value     string
	expiresAt int64 // unix nano
}

type cachedOpenAIQuotaAutoPauseSettings struct {
	settings  OpsOpenAIAccountQuotaAutoPauseSettings
	expiresAt int64
}

const openAICodexUserAgentCacheTTL = 60 * time.Second
const openAICodexUserAgentErrorTTL = 5 * time.Second
const openAICodexUserAgentDBTimeout = 5 * time.Second

// cachedOpenAIAllowCodexPlugin Codex 插件放行开关缓存（进程内缓存，60s TTL）。
// IsOpenAIAllowClaudeCodeCodexPluginEnabled 在每个 codex_cli_only 账号的网关请求热路径上被调用，避免每次访问 DB。
type cachedOpenAIAllowCodexPlugin struct {
	value     bool
	expiresAt int64 // unix nano
}

const openAIAllowCodexPluginCacheTTL = 60 * time.Second
const openAIAllowCodexPluginErrorTTL = 5 * time.Second
const openAIAllowCodexPluginDBTimeout = 5 * time.Second

const openAIQuotaAutoPauseSettingsCacheTTL = 60 * time.Second
const openAIQuotaAutoPauseSettingsErrorTTL = 5 * time.Second
const openAIQuotaAutoPauseSettingsDBTimeout = 5 * time.Second

const openAIQuotaAutoPauseSettingsRefreshKey = "openai_quota_auto_pause_settings"

// DefaultSubscriptionGroupReader validates group references used by default subscriptions.
type DefaultSubscriptionGroupReader interface {
	GetByID(ctx context.Context, id int64) (*Group, error)
}

// WebSearchManagerBuilder creates a websearch.Manager from config (injected by infra layer).
// proxyURLs maps proxy ID to resolved URL for provider-level proxy support.
type WebSearchManagerBuilder func(cfg *WebSearchEmulationConfig, proxyURLs map[int64]string)

// SettingService 系统设置服务
type SettingService struct {
	settingRepo                 SettingRepository
	defaultSubGroupReader       DefaultSubscriptionGroupReader
	proxyRepo                   ProxyRepository // for resolving websearch provider proxy URLs
	cfg                         *config.Config
	onUpdate                    func() // Callback when settings are updated (for cache invalidation)
	version                     string // Application version
	webSearchManagerBuilder     WebSearchManagerBuilder
	antigravityUAVersionCache   atomic.Value // *cachedAntigravityUserAgentVersion
	antigravityUAVersionSF      singleflight.Group
	openAICodexUACache          atomic.Value // *cachedOpenAICodexUserAgent
	openAICodexUASF             singleflight.Group
	openAIAllowCodexPluginCache atomic.Value // *cachedOpenAIAllowCodexPlugin
	openAIAllowCodexPluginSF    singleflight.Group

	// openAIQuotaAutoPauseSettingsCache holds the most recently observed quota auto-pause
	// settings. GetOpenAIQuotaAutoPauseSettings reads this atomic.Value on the request hot
	// path without ever blocking on the DB; when the cached entry expires, a background
	// goroutine refreshes it via openAIQuotaAutoPauseSettingsSF (stale-while-revalidate).
	// This per-service field also gives tests natural isolation — each SettingService
	// instance owns its own cache, no shared package-level state.
	openAIQuotaAutoPauseSettingsCache atomic.Value // *cachedOpenAIQuotaAutoPauseSettings
	openAIQuotaAutoPauseSettingsSF    singleflight.Group
}

// DefaultPlatformQuotaSetting 单 platform 三档限额（nil = 沿用上层；0 = 显式禁用；>0 = 上限）
type DefaultPlatformQuotaSetting struct {
	DailyLimitUSD   *float64 `json:"daily"`
	WeeklyLimitUSD  *float64 `json:"weekly"`
	MonthlyLimitUSD *float64 `json:"monthly"`
}

type ProviderDefaultGrantSettings struct {
	Balance          float64
	Concurrency      int
	Subscriptions    []DefaultSubscriptionSetting
	GrantOnSignup    bool
	GrantOnFirstBind bool
	PlatformQuotas   map[string]*DefaultPlatformQuotaSetting // key = platform name
}

type AuthSourceDefaultSettings struct {
	Email                        ProviderDefaultGrantSettings
	LinuxDo                      ProviderDefaultGrantSettings
	OIDC                         ProviderDefaultGrantSettings
	WeChat                       ProviderDefaultGrantSettings
	GitHub                       ProviderDefaultGrantSettings
	Google                       ProviderDefaultGrantSettings
	DingTalk                     ProviderDefaultGrantSettings
	ForceEmailOnThirdPartySignup bool
}

type authSourceDefaultKeySet struct {
	// source 是 auth source 标识（如 "email"、"github"），仅用于 parse 时
	// slog.Warn 诊断输出，不再参与 key 拼接（platformQuotas 字段已存完整 key）。
	source           string
	balance          string
	concurrency      string
	subscriptions    string
	grantOnSignup    string
	grantOnFirstBind string
	platformQuotas   string // SettingKeyAuthSourcePlatformQuotas(source)
}

var (
	emailAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "email",
		balance:          SettingKeyAuthSourceDefaultEmailBalance,
		concurrency:      SettingKeyAuthSourceDefaultEmailConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultEmailSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultEmailGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultEmailGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("email"),
	}
	linuxDoAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "linuxdo",
		balance:          SettingKeyAuthSourceDefaultLinuxDoBalance,
		concurrency:      SettingKeyAuthSourceDefaultLinuxDoConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultLinuxDoSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultLinuxDoGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultLinuxDoGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("linuxdo"),
	}
	oidcAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "oidc",
		balance:          SettingKeyAuthSourceDefaultOIDCBalance,
		concurrency:      SettingKeyAuthSourceDefaultOIDCConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultOIDCSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultOIDCGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultOIDCGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("oidc"),
	}
	weChatAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "wechat",
		balance:          SettingKeyAuthSourceDefaultWeChatBalance,
		concurrency:      SettingKeyAuthSourceDefaultWeChatConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultWeChatSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultWeChatGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultWeChatGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("wechat"),
	}
	gitHubAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "github",
		balance:          SettingKeyAuthSourceDefaultGitHubBalance,
		concurrency:      SettingKeyAuthSourceDefaultGitHubConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultGitHubSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultGitHubGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultGitHubGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("github"),
	}
	googleAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "google",
		balance:          SettingKeyAuthSourceDefaultGoogleBalance,
		concurrency:      SettingKeyAuthSourceDefaultGoogleConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultGoogleSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultGoogleGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultGoogleGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("google"),
	}
	dingTalkAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "dingtalk",
		balance:          SettingKeyAuthSourceDefaultDingTalkBalance,
		concurrency:      SettingKeyAuthSourceDefaultDingTalkConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultDingTalkSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultDingTalkGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultDingTalkGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("dingtalk"),
	}
)

const (
	defaultAuthSourceBalance        = 0
	defaultAuthSourceConcurrency    = 5
	defaultWeChatConnectMode        = "open"
	defaultWeChatConnectScopes      = "snsapi_login"
	defaultWeChatConnectFrontend    = "/auth/wechat/callback"
	defaultGitHubOAuthAuthorize     = "https://github.com/login/oauth/authorize"
	defaultGitHubOAuthToken         = "https://github.com/login/oauth/access_token"
	defaultGitHubOAuthUserInfo      = "https://api.github.com/user"
	defaultGitHubOAuthEmails        = "https://api.github.com/user/emails"
	defaultGitHubOAuthScopes        = "read:user user:email"
	defaultGitHubOAuthFrontend      = "/auth/oauth/callback"
	defaultGoogleOAuthAuthorize     = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultGoogleOAuthToken         = "https://oauth2.googleapis.com/token"
	defaultGoogleOAuthUserInfo      = "https://openidconnect.googleapis.com/v1/userinfo"
	defaultGoogleOAuthScopes        = "openid email profile"
	defaultGoogleOAuthFrontend      = "/auth/oauth/callback"
	defaultLoginAgreementMode       = "modal"
	defaultLoginAgreementDate       = "2026-03-31"
	defaultWebAffonsoCookieDuration = "30"
)

func normalizeLoginAgreementMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "checkbox":
		return "checkbox"
	default:
		return defaultLoginAgreementMode
	}
}

func defaultLoginAgreementDocuments() []LoginAgreementDocument {
	return []LoginAgreementDocument{
		{
			ID:        "terms",
			Title:     "服务条款",
			ContentMD: "",
		},
		{
			ID:        "usage-policy",
			Title:     "使用政策",
			ContentMD: "",
		},
		{
			ID:        "supported-regions",
			Title:     "支持的国家和地区",
			ContentMD: "",
		},
		{
			ID:        "service-specific-terms",
			Title:     "服务特定条款",
			ContentMD: "",
		},
	}
}

func normalizeLoginAgreementDocumentID(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	var b strings.Builder
	lastSeparator := false
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			_, _ = b.WriteRune(r)
			lastSeparator = false
			continue
		}
		if r == '-' || r == '_' || r == ' ' || r == '.' || r == '/' {
			if !lastSeparator && b.Len() > 0 {
				if r == '_' {
					_, _ = b.WriteRune('_')
				} else {
					_, _ = b.WriteRune('-')
				}
				lastSeparator = true
			}
		}
	}
	return strings.Trim(b.String(), "-_")
}

func normalizeLoginAgreementDocuments(docs []LoginAgreementDocument) []LoginAgreementDocument {
	normalized := make([]LoginAgreementDocument, 0, len(docs))
	seen := make(map[string]int, len(docs))
	for i, doc := range docs {
		title := strings.TrimSpace(doc.Title)
		content := strings.TrimSpace(doc.ContentMD)
		if title == "" && content == "" {
			continue
		}
		id := normalizeLoginAgreementDocumentID(doc.ID)
		if id == "" {
			sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", i, title, content)))
			id = hex.EncodeToString(sum[:])[:12]
		}
		baseID := id
		for suffix := 2; seen[id] > 0; suffix++ {
			id = fmt.Sprintf("%s-%d", baseID, suffix)
		}
		seen[id]++
		normalized = append(normalized, LoginAgreementDocument{
			ID:        id,
			Title:     title,
			ContentMD: content,
		})
	}
	return normalized
}

func parseLoginAgreementDocuments(raw string) []LoginAgreementDocument {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultLoginAgreementDocuments()
	}
	var docs []LoginAgreementDocument
	if err := json.Unmarshal([]byte(raw), &docs); err != nil {
		return defaultLoginAgreementDocuments()
	}
	docs = normalizeLoginAgreementDocuments(docs)
	if len(docs) == 0 {
		return defaultLoginAgreementDocuments()
	}
	return docs
}

func marshalLoginAgreementDocuments(docs []LoginAgreementDocument) (string, error) {
	normalized := normalizeLoginAgreementDocuments(docs)
	if len(normalized) == 0 {
		normalized = defaultLoginAgreementDocuments()
	}
	b, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("marshal login agreement documents: %w", err)
	}
	return string(b), nil
}

func buildLoginAgreementRevision(updatedAt string, docs []LoginAgreementDocument) string {
	normalized := normalizeLoginAgreementDocuments(docs)
	payload, err := json.Marshal(struct {
		UpdatedAt string                   `json:"updated_at"`
		Documents []LoginAgreementDocument `json:"documents"`
	}{
		UpdatedAt: strings.TrimSpace(updatedAt),
		Documents: normalized,
	})
	if err != nil {
		payload = []byte(strings.TrimSpace(updatedAt))
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])[:16]
}

func (s *SettingService) GetCurrentLoginAgreementRequirement(ctx context.Context) (bool, string, error) {
	if s == nil {
		return false, "", nil
	}

	settings, err := s.GetAllSettings(ctx)
	if err != nil {
		return false, "", err
	}

	docs := normalizeLoginAgreementDocuments(settings.LoginAgreementDocuments)
	if !settings.LoginAgreementEnabled || len(docs) == 0 {
		return false, "", nil
	}

	updatedAt := strings.TrimSpace(settings.LoginAgreementUpdatedAt)
	if updatedAt == "" {
		updatedAt = defaultLoginAgreementDate
	}

	return true, buildLoginAgreementRevision(updatedAt, docs), nil
}

func normalizeWeChatConnectModeSetting(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "mp":
		return "mp"
	case "mobile":
		return "mobile"
	default:
		return "open"
	}
}

func defaultWeChatConnectScopeForMode(mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		return "snsapi_userinfo"
	case "mobile":
		return ""
	}
	return defaultWeChatConnectScopes
}

func normalizeWeChatConnectScopeSetting(raw, mode string) string {
	switch normalizeWeChatConnectModeSetting(mode) {
	case "mp":
		switch strings.TrimSpace(raw) {
		case "snsapi_base":
			return "snsapi_base"
		case "snsapi_userinfo":
			return "snsapi_userinfo"
		default:
			return defaultWeChatConnectScopeForMode(mode)
		}
	case "mobile":
		return ""
	default:
		return defaultWeChatConnectScopes
	}
}

func parseWeChatConnectCapabilitySettings(settings map[string]string, enabled bool, mode string) (bool, bool, bool) {
	mode = normalizeWeChatConnectModeSetting(mode)
	rawOpen, hasOpen := settings[SettingKeyWeChatConnectOpenEnabled]
	rawMP, hasMP := settings[SettingKeyWeChatConnectMPEnabled]
	rawMobile, hasMobile := settings[SettingKeyWeChatConnectMobileEnabled]
	openConfigured := hasOpen && strings.TrimSpace(rawOpen) != ""
	mpConfigured := hasMP && strings.TrimSpace(rawMP) != ""
	mobileConfigured := hasMobile && strings.TrimSpace(rawMobile) != ""

	if openConfigured || mpConfigured || mobileConfigured {
		openEnabled := strings.TrimSpace(rawOpen) == "true"
		mpEnabled := strings.TrimSpace(rawMP) == "true"
		mobileEnabled := strings.TrimSpace(rawMobile) == "true"
		return openEnabled, mpEnabled, mobileEnabled
	}

	if !enabled {
		return false, false, false
	}
	if mode == "mp" {
		return false, true, false
	}
	if mode == "mobile" {
		return false, false, true
	}
	return true, false, false
}

func normalizeWeChatConnectStoredMode(openEnabled, mpEnabled, mobileEnabled bool, mode string) string {
	mode = normalizeWeChatConnectModeSetting(mode)
	switch mode {
	case "open":
		if openEnabled {
			return "open"
		}
	case "mp":
		if mpEnabled {
			return "mp"
		}
	case "mobile":
		if mobileEnabled {
			return "mobile"
		}
	}
	switch {
	case openEnabled:
		return "open"
	case mpEnabled:
		return "mp"
	case mobileEnabled:
		return "mobile"
	default:
		return mode
	}
}

func mergeWeChatConnectCapabilitySettings(settings map[string]string, base config.WeChatConnectConfig, enabled bool, mode string) (bool, bool, bool) {
	mode = normalizeWeChatConnectModeSetting(firstNonEmpty(mode, base.Mode))
	rawOpen, hasOpen := settings[SettingKeyWeChatConnectOpenEnabled]
	rawMP, hasMP := settings[SettingKeyWeChatConnectMPEnabled]
	rawMobile, hasMobile := settings[SettingKeyWeChatConnectMobileEnabled]
	openConfigured := hasOpen && strings.TrimSpace(rawOpen) != ""
	mpConfigured := hasMP && strings.TrimSpace(rawMP) != ""
	mobileConfigured := hasMobile && strings.TrimSpace(rawMobile) != ""

	if openConfigured || mpConfigured || mobileConfigured {
		openEnabled := strings.TrimSpace(rawOpen) == "true"
		mpEnabled := strings.TrimSpace(rawMP) == "true"
		mobileEnabled := strings.TrimSpace(rawMobile) == "true"
		_, enabledConfigured := settings[SettingKeyWeChatConnectEnabled]
		if !enabledConfigured &&
			enabled &&
			!openEnabled &&
			!mpEnabled &&
			!mobileEnabled &&
			(base.OpenEnabled || base.MPEnabled || base.MobileEnabled) {
			return base.OpenEnabled, base.MPEnabled, base.MobileEnabled
		}
		return openEnabled, mpEnabled, mobileEnabled
	}
	if !enabled {
		return false, false, false
	}
	if base.OpenEnabled || base.MPEnabled || base.MobileEnabled {
		return base.OpenEnabled, base.MPEnabled, base.MobileEnabled
	}
	return parseWeChatConnectCapabilitySettings(settings, enabled, mode)
}

func (s *SettingService) effectiveWeChatConnectOAuthConfig(settings map[string]string) WeChatConnectOAuthConfig {
	base := config.WeChatConnectConfig{}
	if s != nil && s.cfg != nil {
		base = s.cfg.WeChat
	}

	enabled := base.Enabled
	if raw, ok := settings[SettingKeyWeChatConnectEnabled]; ok {
		enabled = strings.TrimSpace(raw) == "true"
	}

	openAppID := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectOpenAppID], base.OpenAppID))
	openAppSecret := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectOpenAppSecret], base.OpenAppSecret))
	mpAppID := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMPAppID], base.MPAppID))
	mpAppSecret := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMPAppSecret], base.MPAppSecret))
	mobileAppID := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMobileAppID], base.MobileAppID))
	mobileAppSecret := strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectMobileAppSecret], base.MobileAppSecret))

	modeRaw := firstNonEmpty(settings[SettingKeyWeChatConnectMode], base.Mode)
	openEnabled, mpEnabled, mobileEnabled := mergeWeChatConnectCapabilitySettings(settings, base, enabled, modeRaw)
	mode := normalizeWeChatConnectStoredMode(openEnabled, mpEnabled, mobileEnabled, modeRaw)

	return WeChatConnectOAuthConfig{
		Enabled:             enabled,
		OpenAppID:           openAppID,
		OpenAppSecret:       openAppSecret,
		MPAppID:             mpAppID,
		MPAppSecret:         mpAppSecret,
		MobileAppID:         mobileAppID,
		MobileAppSecret:     mobileAppSecret,
		OpenEnabled:         openEnabled,
		MPEnabled:           mpEnabled,
		MobileEnabled:       mobileEnabled,
		Mode:                mode,
		Scopes:              normalizeWeChatConnectScopeSetting(firstNonEmpty(settings[SettingKeyWeChatConnectScopes], base.Scopes), mode),
		RedirectURL:         strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectRedirectURL], base.RedirectURL)),
		FrontendRedirectURL: strings.TrimSpace(firstNonEmpty(settings[SettingKeyWeChatConnectFrontendRedirectURL], base.FrontendRedirectURL, defaultWeChatConnectFrontend)),
	}
}

// NewSettingService 创建系统设置服务实例
func NewSettingService(settingRepo SettingRepository, cfg *config.Config) *SettingService {
	return &SettingService{
		settingRepo: settingRepo,
		cfg:         cfg,
	}
}

// SetDefaultSubscriptionGroupReader injects an optional group reader for default subscription validation.
func (s *SettingService) SetDefaultSubscriptionGroupReader(reader DefaultSubscriptionGroupReader) {
	s.defaultSubGroupReader = reader
}

// SetProxyRepository injects a proxy repo for resolving websearch provider proxy URLs.
func (s *SettingService) SetProxyRepository(repo ProxyRepository) {
	s.proxyRepo = repo
}

func (s *SettingService) LoadAPIKeyACLTrustForwardedIPSetting(ctx context.Context) error {
	if s == nil || s.cfg == nil || s.settingRepo == nil {
		return nil
	}
	value, err := s.settingRepo.GetValue(ctx, SettingKeyAPIKeyACLTrustForwardedIP)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			s.cfg.SetTrustForwardedIPForAPIKeyACL(s.cfg.Security.TrustForwardedIPForAPIKeyACL)
			return nil
		}
		return fmt.Errorf("get api key acl forwarded ip setting: %w", err)
	}
	enabled := value == "true"
	s.cfg.SetTrustForwardedIPForAPIKeyACL(enabled)
	return nil
}

// GetAllSettings 获取所有系统设置
func (s *SettingService) GetAllSettings(ctx context.Context) (*SystemSettings, error) {
	settings, err := s.settingRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}

	return s.parseSettings(settings), nil
}

// GetFrontendURL 获取前端基础URL（数据库优先，fallback 到配置文件）
func (s *SettingService) GetFrontendURL(ctx context.Context) string {
	val, err := s.settingRepo.GetValue(ctx, SettingKeyFrontendURL)
	if err == nil && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	if s.cfg == nil {
		return ""
	}
	return s.cfg.Server.FrontendURL
}

// GetEmailLogoURL resolves the absolute logo URL used by HTML email templates.
func (s *SettingService) GetEmailLogoURL(ctx context.Context) string {
	if s == nil || s.settingRepo == nil {
		return ""
	}
	settings, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeySiteLogo,
		SettingKeyFrontendURL,
		SettingKeyAPIBaseURL,
	})
	if err != nil {
		return ""
	}
	baseURL := firstNonEmpty(settings[SettingKeyFrontendURL], settings[SettingKeyAPIBaseURL], s.GetFrontendURL(ctx))
	if logo, ok := parseSiteLogoDataURL(settings[SettingKeySiteLogo]); ok {
		if endpoint := emailSiteLogoEndpointURL(baseURL, logo.ETag); endpoint != "" {
			return endpoint
		}
	}
	if logoURL := normalizeEmailImageURL(settings[SettingKeySiteLogo], baseURL); logoURL != "" {
		return logoURL
	}
	return emailDefaultLogoURL(baseURL)
}

func (s *SettingService) GetSiteLogoImage(ctx context.Context) (*SiteLogoImage, error) {
	if s == nil || s.settingRepo == nil {
		return nil, nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeySiteLogo)
	if err != nil {
		return nil, fmt.Errorf("get site logo: %w", err)
	}
	logo, ok := parseSiteLogoDataURL(raw)
	if !ok {
		return nil, nil
	}
	return logo, nil
}

func parseSiteLogoDataURL(raw string) (*SiteLogoImage, bool) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "data:") {
		return nil, false
	}
	header, payload, ok := strings.Cut(raw, ",")
	if !ok {
		return nil, false
	}
	mediaType := strings.TrimSpace(strings.TrimPrefix(header, "data:"))
	parts := strings.Split(mediaType, ";")
	if len(parts) < 2 {
		return nil, false
	}
	contentType := strings.ToLower(strings.TrimSpace(parts[0]))
	if !isAllowedSiteLogoContentType(contentType) {
		return nil, false
	}
	base64Encoded := false
	for _, part := range parts[1:] {
		if strings.EqualFold(strings.TrimSpace(part), "base64") {
			base64Encoded = true
			break
		}
	}
	if !base64Encoded {
		return nil, false
	}

	cleanPayload := strings.NewReplacer("\n", "", "\r", "", "\t", "", " ", "").Replace(payload)
	data, err := base64.StdEncoding.DecodeString(cleanPayload)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	sum := sha256.Sum256(data)
	return &SiteLogoImage{
		ContentType: contentType,
		Data:        data,
		ETag:        `"` + hex.EncodeToString(sum[:]) + `"`,
	}, true
}

func isAllowedSiteLogoContentType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/png", "image/jpeg", "image/jpg", "image/webp", "image/gif", "image/x-icon", "image/vnd.microsoft.icon":
		return true
	default:
		return false
	}
}

// GetPublicSettings 获取公开设置（无需登录）
func (s *SettingService) GetPublicSettings(ctx context.Context) (*PublicSettings, error) {
	keys := []string{
		SettingKeyRegistrationEnabled,
		SettingKeyEmailVerifyEnabled,
		SettingKeyForceEmailOnThirdPartySignup,
		SettingKeyRegistrationEmailSuffixWhitelist,
		SettingKeyPromoCodeEnabled,
		SettingKeyPasswordResetEnabled,
		SettingKeyPasswordMinLength,
		SettingKeyInvitationCodeEnabled,
		SettingKeyTotpEnabled,
		SettingKeyLoginAgreementEnabled,
		SettingKeyLoginAgreementMode,
		SettingKeyLoginAgreementUpdatedAt,
		SettingKeyLoginAgreementDocuments,
		SettingKeyTurnstileEnabled,
		SettingKeyTurnstileSiteKey,
		SettingKeyAPIKeyACLTrustForwardedIP,
		SettingKeySiteName,
		SettingKeySiteLogo,
		SettingKeySiteSubtitle,
		SettingKeyAPIBaseURL,
		SettingKeyContactInfo,
		SettingKeyDocURL,
		SettingKeyDocsContentBasePath,
		SettingKeyHomeContent,
		SettingKeyHomeShellConfig,
		SettingKeyHomeBusinessShellConfig,
		SettingKeyModelPlazaItems,
		SettingKeyImageWorkspaceModelConfig,
		SettingKeyModelPlazaShellConfig,
		SettingKeyDocsShellConfig,
		SettingKeyLegalDocumentShellConfig,
		SettingKeyAPIKeysShellConfig,
		SettingKeyKeyUsageShellConfig,
		SettingKeyDashboardShellConfig,
		SettingKeyUsageShellConfig,
		SettingKeyAPIGuideShellConfig,
		SettingKeyAPITestShellConfig,
		SettingKeyAvailableGroupsShellConfig,
		SettingKeyRedeemShellConfig,
		SettingKeyAffiliateShellConfig,
		SettingKeyAvailableChannelsShellConfig,
		SettingKeyChannelStatusShellConfig,
		SettingKeyCustomPageShellConfig,
		SettingKeyProfileShellConfig,
		SettingKeyAuthShellConfig,
		SettingKeyHideCcsImportButton,
		SettingKeyPurchaseSubscriptionEnabled,
		SettingKeyPurchaseSubscriptionURL,
		SettingKeyTableDefaultPageSize,
		SettingKeyTablePageSizeOptions,
		SettingKeyCustomMenuItems,
		SettingKeyCustomEndpoints,
		SettingKeyLinuxDoConnectEnabled,
		SettingKeyDingTalkConnectEnabled,
		SettingKeyWeChatConnectEnabled,
		SettingKeyWeChatConnectOpenAppID,
		SettingKeyWeChatConnectOpenAppSecret,
		SettingKeyWeChatConnectMPAppID,
		SettingKeyWeChatConnectMPAppSecret,
		SettingKeyWeChatConnectMobileAppID,
		SettingKeyWeChatConnectMobileAppSecret,
		SettingKeyWeChatConnectOpenEnabled,
		SettingKeyWeChatConnectMPEnabled,
		SettingKeyWeChatConnectMobileEnabled,
		SettingKeyWeChatConnectMode,
		SettingKeyWeChatConnectScopes,
		SettingKeyWeChatConnectRedirectURL,
		SettingKeyWeChatConnectFrontendRedirectURL,
		SettingKeyBackendModeEnabled,
		SettingPaymentEnabled,
		SettingKeyOIDCConnectEnabled,
		SettingKeyOIDCConnectProviderName,
		SettingKeyGitHubOAuthEnabled,
		SettingKeyGitHubOAuthClientID,
		SettingKeyGitHubOAuthClientSecret,
		SettingKeyGoogleOAuthEnabled,
		SettingKeyGoogleOAuthClientID,
		SettingKeyGoogleOAuthClientSecret,
		SettingKeyBalanceLowNotifyEnabled,
		SettingKeyBalanceLowNotifyThreshold,
		SettingKeyBalanceLowNotifyRechargeURL,
		SettingKeyAccountQuotaNotifyEnabled,
		SettingKeyChannelMonitorEnabled,
		SettingKeyChannelMonitorDefaultIntervalSeconds,
		SettingKeyAvailableChannelsEnabled,
		SettingKeyAffiliateEnabled,
		SettingKeyRiskControlEnabled,
		SettingKeyWebAppURL,
		SettingKeyWebAppName,
		SettingKeyWebAppDescription,
		SettingKeyWebAppLogo,
		SettingKeyWebAppFavicon,
		SettingKeyWebAppPreviewImage,
		SettingKeyWebTheme,
		SettingKeyWebAppearance,
		SettingKeyWebDefaultLocale,
		SettingKeyPromptCasesTitle,
		SettingKeyPromptCasesDescription,
		SettingKeyPromptTemplatesTitle,
		SettingKeyPromptTemplatesDescription,
		SettingKeyPromptCatalogShellConfig,
		SettingKeyWorkspaceShellConfig,
		SettingKeyPricingTitle,
		SettingKeyPricingDescription,
		SettingKeyPricingShellConfig,
		SettingKeyPaymentShellConfig,
		SettingKeyPricingCurrencySymbol,
		SettingKeyCreditsTitle,
		SettingKeyCreditsDescription,
		SettingKeyCreditsPurchaseLabel,
		SettingKeyCreditsBalanceLabel,
		SettingKeyCreditsPerBalance,
		SettingKeyCreditsShellConfig,
		SettingKeyWebLocaleDetectEnabled,
		SettingKeyWebEmailAuthVisible,
		SettingKeyWebGoogleAuthVisible,
		SettingKeyWebGitHubAuthVisible,
		SettingKeyWebGoogleAnalyticsID,
		SettingKeyWebClarityID,
		SettingKeyWebPlausibleDomain,
		SettingKeyWebPlausibleSrc,
		SettingKeyWebOpenPanelClientID,
		SettingKeyWebPublicIntegrationsEnabled,
		SettingKeyWebVercelAnalyticsEnabled,
		SettingKeyWebAdsenseCode,
		SettingKeyWebAffonsoEnabled,
		SettingKeyWebAffonsoID,
		SettingKeyWebAffonsoCookieDuration,
		SettingKeyWebPromoteKitEnabled,
		SettingKeyWebPromoteKitID,
		SettingKeyWebCrispEnabled,
		SettingKeyWebCrispWebsiteID,
		SettingKeyWebTawkEnabled,
		SettingKeyWebTawkPropertyID,
		SettingKeyWebTawkWidgetID,
	}

	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get public settings: %w", err)
	}

	linuxDoEnabled := false
	if raw, ok := settings[SettingKeyLinuxDoConnectEnabled]; ok {
		linuxDoEnabled = raw == "true"
	} else {
		linuxDoEnabled = s.cfg != nil && s.cfg.LinuxDo.Enabled
	}
	dingTalkEnabled := false
	if raw, ok := settings[SettingKeyDingTalkConnectEnabled]; ok {
		dingTalkEnabled = raw == "true"
	} else {
		dingTalkEnabled = s.cfg != nil && s.cfg.DingTalk.Enabled
	}
	oidcEnabled := false
	if raw, ok := settings[SettingKeyOIDCConnectEnabled]; ok {
		oidcEnabled = raw == "true"
	} else {
		oidcEnabled = s.cfg != nil && s.cfg.OIDC.Enabled
	}
	oidcProviderName := strings.TrimSpace(settings[SettingKeyOIDCConnectProviderName])
	if oidcProviderName == "" && s.cfg != nil {
		oidcProviderName = strings.TrimSpace(s.cfg.OIDC.ProviderName)
	}
	if oidcProviderName == "" {
		oidcProviderName = "OIDC"
	}
	gitHubEnabled := s.emailOAuthPublicEnabled(settings, "github")
	googleEnabled := s.emailOAuthPublicEnabled(settings, "google")
	weChatEnabled, weChatOpenEnabled, weChatMPEnabled, weChatMobileEnabled := s.weChatOAuthCapabilitiesFromSettings(settings)
	siteName := s.getStringOrDefault(settings, SettingKeySiteName, "Sub2API")
	siteSubtitle := s.getStringOrDefault(settings, SettingKeySiteSubtitle, "Subscription to API Conversion Platform")
	webEmailVisible := parseBoolSettingWithDefault(settings[SettingKeyWebEmailAuthVisible], settings[SettingKeyRegistrationEnabled] != "false")
	webGoogleVisible := parseBoolSettingWithDefault(settings[SettingKeyWebGoogleAuthVisible], googleEnabled)
	webGitHubVisible := parseBoolSettingWithDefault(settings[SettingKeyWebGitHubAuthVisible], gitHubEnabled)

	// Password reset requires email verification to be enabled
	emailVerifyEnabled := settings[SettingKeyEmailVerifyEnabled] == "true"
	passwordResetEnabled := emailVerifyEnabled && settings[SettingKeyPasswordResetEnabled] == "true"
	registrationEmailSuffixWhitelist := ParseRegistrationEmailSuffixWhitelist(
		settings[SettingKeyRegistrationEmailSuffixWhitelist],
	)
	tableDefaultPageSize, tablePageSizeOptions := parseTablePreferences(
		settings[SettingKeyTableDefaultPageSize],
		settings[SettingKeyTablePageSizeOptions],
	)
	loginAgreementDocuments := parseLoginAgreementDocuments(settings[SettingKeyLoginAgreementDocuments])
	loginAgreementUpdatedAt := strings.TrimSpace(settings[SettingKeyLoginAgreementUpdatedAt])
	if loginAgreementUpdatedAt == "" {
		loginAgreementUpdatedAt = defaultLoginAgreementDate
	}

	var balanceLowNotifyThreshold float64
	if v, err := strconv.ParseFloat(settings[SettingKeyBalanceLowNotifyThreshold], 64); err == nil && v >= 0 {
		balanceLowNotifyThreshold = v
	}

	return &PublicSettings{
		RegistrationEnabled:              settings[SettingKeyRegistrationEnabled] == "true",
		EmailVerifyEnabled:               emailVerifyEnabled,
		ForceEmailOnThirdPartySignup:     settings[SettingKeyForceEmailOnThirdPartySignup] == "true",
		RegistrationEmailSuffixWhitelist: registrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 settings[SettingKeyPromoCodeEnabled] != "false", // 默认启用
		PasswordResetEnabled:             passwordResetEnabled,
		PasswordMinLength:                parsePasswordMinLength(settings[SettingKeyPasswordMinLength]),
		InvitationCodeEnabled:            settings[SettingKeyInvitationCodeEnabled] == "true",
		TotpEnabled:                      settings[SettingKeyTotpEnabled] == "true",
		LoginAgreementEnabled:            settings[SettingKeyLoginAgreementEnabled] == "true" && len(loginAgreementDocuments) > 0,
		LoginAgreementMode:               normalizeLoginAgreementMode(settings[SettingKeyLoginAgreementMode]),
		LoginAgreementUpdatedAt:          loginAgreementUpdatedAt,
		LoginAgreementRevision:           buildLoginAgreementRevision(loginAgreementUpdatedAt, loginAgreementDocuments),
		LoginAgreementDocuments:          loginAgreementDocuments,
		TurnstileEnabled:                 settings[SettingKeyTurnstileEnabled] == "true",
		TurnstileSiteKey:                 settings[SettingKeyTurnstileSiteKey],
		SiteName:                         siteName,
		SiteLogo:                         settings[SettingKeySiteLogo],
		SiteSubtitle:                     siteSubtitle,
		APIBaseURL:                       settings[SettingKeyAPIBaseURL],
		ContactInfo:                      settings[SettingKeyContactInfo],
		DocURL:                           settings[SettingKeyDocURL],
		DocsContentBasePath:              docsContentBasePathSetting(settings[SettingKeyDocsContentBasePath]),
		HomeContent:                      settings[SettingKeyHomeContent],
		HomeShellConfig:                  homeShellConfigSetting(settings[SettingKeyHomeShellConfig]),
		HomeBusinessShellConfig:          homeBusinessShellConfigSetting(settings[SettingKeyHomeBusinessShellConfig]),
		ModelPlazaItems:                  settings[SettingKeyModelPlazaItems],
		ImageWorkspaceModelConfig:        imageWorkspaceModelConfigSetting(settings[SettingKeyImageWorkspaceModelConfig]),
		ModelPlazaShellConfig:            modelPlazaShellConfigSetting(settings[SettingKeyModelPlazaShellConfig]),
		DocsShellConfig:                  docsShellConfigSetting(settings[SettingKeyDocsShellConfig]),
		LegalDocumentShellConfig:         legalDocumentShellConfigSetting(settings[SettingKeyLegalDocumentShellConfig]),
		APIKeysShellConfig:               apiKeysShellConfigSetting(settings[SettingKeyAPIKeysShellConfig]),
		KeyUsageShellConfig:              keyUsageShellConfigSetting(settings[SettingKeyKeyUsageShellConfig]),
		DashboardShellConfig:             dashboardShellConfigSetting(settings[SettingKeyDashboardShellConfig]),
		UsageShellConfig:                 usageShellConfigSetting(settings[SettingKeyUsageShellConfig]),
		APIGuideShellConfig:              apiGuideShellConfigSetting(settings[SettingKeyAPIGuideShellConfig]),
		APITestShellConfig:               apiTestShellConfigSetting(settings[SettingKeyAPITestShellConfig]),
		AvailableGroupsShellConfig:       availableGroupsShellConfigSetting(settings[SettingKeyAvailableGroupsShellConfig]),
		RedeemShellConfig:                redeemShellConfigSetting(settings[SettingKeyRedeemShellConfig]),
		AffiliateShellConfig:             affiliateShellConfigSetting(settings[SettingKeyAffiliateShellConfig]),
		AvailableChannelsShellConfig:     availableChannelsShellConfigSetting(settings[SettingKeyAvailableChannelsShellConfig]),
		ChannelStatusShellConfig:         channelStatusShellConfigSetting(settings[SettingKeyChannelStatusShellConfig]),
		CustomPageShellConfig:            customPageShellConfigSetting(settings[SettingKeyCustomPageShellConfig]),
		ProfileShellConfig:               profileShellConfigSetting(settings[SettingKeyProfileShellConfig]),
		AuthShellConfig:                  authShellConfigSetting(settings[SettingKeyAuthShellConfig]),
		HideCcsImportButton:              settings[SettingKeyHideCcsImportButton] == "true",
		PurchaseSubscriptionEnabled:      settings[SettingKeyPurchaseSubscriptionEnabled] == "true",
		PurchaseSubscriptionURL:          strings.TrimSpace(settings[SettingKeyPurchaseSubscriptionURL]),
		TableDefaultPageSize:             tableDefaultPageSize,
		TablePageSizeOptions:             tablePageSizeOptions,
		CustomMenuItems:                  settings[SettingKeyCustomMenuItems],
		CustomEndpoints:                  settings[SettingKeyCustomEndpoints],
		LinuxDoOAuthEnabled:              linuxDoEnabled,
		DingTalkOAuthEnabled:             dingTalkEnabled,
		WeChatOAuthEnabled:               weChatEnabled,
		WeChatOAuthOpenEnabled:           weChatOpenEnabled,
		WeChatOAuthMPEnabled:             weChatMPEnabled,
		WeChatOAuthMobileEnabled:         weChatMobileEnabled,
		BackendModeEnabled:               settings[SettingKeyBackendModeEnabled] == "true",
		PaymentEnabled:                   settings[SettingPaymentEnabled] == "true",
		OIDCOAuthEnabled:                 oidcEnabled,
		OIDCOAuthProviderName:            oidcProviderName,
		GitHubOAuthEnabled:               gitHubEnabled,
		GoogleOAuthEnabled:               googleEnabled,
		BalanceLowNotifyEnabled:          settings[SettingKeyBalanceLowNotifyEnabled] == "true",
		AccountQuotaNotifyEnabled:        settings[SettingKeyAccountQuotaNotifyEnabled] == "true",
		BalanceLowNotifyThreshold:        balanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:      settings[SettingKeyBalanceLowNotifyRechargeURL],

		ChannelMonitorEnabled:                !isFalseSettingValue(settings[SettingKeyChannelMonitorEnabled]),
		ChannelMonitorDefaultIntervalSeconds: parseChannelMonitorInterval(settings[SettingKeyChannelMonitorDefaultIntervalSeconds]),

		AvailableChannelsEnabled: settings[SettingKeyAvailableChannelsEnabled] == "true",

		AffiliateEnabled: settings[SettingKeyAffiliateEnabled] == "true",

		RiskControlEnabled: settings[SettingKeyRiskControlEnabled] == "true",

		WebAppURL:                     strings.TrimSpace(settings[SettingKeyWebAppURL]),
		WebAppName:                    firstNonEmpty(settings[SettingKeyWebAppName], siteName),
		WebAppDescription:             firstNonEmpty(settings[SettingKeyWebAppDescription], siteSubtitle),
		WebAppLogo:                    firstNonEmpty(settings[SettingKeyWebAppLogo], settings[SettingKeySiteLogo]),
		WebAppFavicon:                 strings.TrimSpace(settings[SettingKeyWebAppFavicon]),
		WebAppPreviewImage:            strings.TrimSpace(settings[SettingKeyWebAppPreviewImage]),
		WebTheme:                      strings.TrimSpace(settings[SettingKeyWebTheme]),
		WebAppearance:                 strings.TrimSpace(settings[SettingKeyWebAppearance]),
		WebDefaultLocale:              strings.TrimSpace(settings[SettingKeyWebDefaultLocale]),
		WebPromptCasesTitle:           strings.TrimSpace(settings[SettingKeyPromptCasesTitle]),
		WebPromptCasesDescription:     strings.TrimSpace(settings[SettingKeyPromptCasesDescription]),
		WebPromptTemplatesTitle:       strings.TrimSpace(settings[SettingKeyPromptTemplatesTitle]),
		WebPromptTemplatesDescription: strings.TrimSpace(settings[SettingKeyPromptTemplatesDescription]),
		PromptCasesTitle:              strings.TrimSpace(settings[SettingKeyPromptCasesTitle]),
		PromptCasesDescription:        strings.TrimSpace(settings[SettingKeyPromptCasesDescription]),
		PromptTemplatesTitle:          strings.TrimSpace(settings[SettingKeyPromptTemplatesTitle]),
		PromptTemplatesDescription:    strings.TrimSpace(settings[SettingKeyPromptTemplatesDescription]),
		PromptCatalogShellConfig:      promptCatalogShellConfigSetting(settings[SettingKeyPromptCatalogShellConfig]),
		WorkspaceShellConfig:          workspaceShellConfigSetting(settings[SettingKeyWorkspaceShellConfig]),
		PricingTitle:                  strings.TrimSpace(settings[SettingKeyPricingTitle]),
		PricingDescription:            strings.TrimSpace(settings[SettingKeyPricingDescription]),
		PricingShellConfig:            pricingShellConfigSetting(settings[SettingKeyPricingShellConfig]),
		PaymentShellConfig:            paymentShellConfigSetting(settings[SettingKeyPaymentShellConfig]),
		PricingCurrencySymbol:         pricingCurrencySymbolSetting(settings[SettingKeyPricingCurrencySymbol]),
		CreditsTitle:                  strings.TrimSpace(settings[SettingKeyCreditsTitle]),
		CreditsDescription:            strings.TrimSpace(settings[SettingKeyCreditsDescription]),
		CreditsPurchaseLabel:          strings.TrimSpace(settings[SettingKeyCreditsPurchaseLabel]),
		CreditsBalanceLabel:           strings.TrimSpace(settings[SettingKeyCreditsBalanceLabel]),
		CreditsPerBalance:             creditsPerBalanceSetting(settings[SettingKeyCreditsPerBalance]),
		CreditsShellConfig:            creditsShellConfigSetting(settings[SettingKeyCreditsShellConfig]),
		GoogleAnalyticsID:             strings.TrimSpace(settings[SettingKeyWebGoogleAnalyticsID]),
		ClarityID:                     strings.TrimSpace(settings[SettingKeyWebClarityID]),
		PlausibleDomain:               strings.TrimSpace(settings[SettingKeyWebPlausibleDomain]),
		PlausibleSrc:                  strings.TrimSpace(settings[SettingKeyWebPlausibleSrc]),
		OpenPanelClientID:             strings.TrimSpace(settings[SettingKeyWebOpenPanelClientID]),
		PublicIntegrationsEnabled:     !isFalseSettingValue(settings[SettingKeyWebPublicIntegrationsEnabled]),
		VercelAnalyticsEnabled:        settings[SettingKeyWebVercelAnalyticsEnabled] == "true",
		AdsenseCode:                   strings.TrimSpace(settings[SettingKeyWebAdsenseCode]),
		AffonsoEnabled:                settings[SettingKeyWebAffonsoEnabled] == "true",
		AffonsoID:                     strings.TrimSpace(settings[SettingKeyWebAffonsoID]),
		AffonsoCookieDuration:         webAffonsoCookieDurationSetting(settings[SettingKeyWebAffonsoCookieDuration]),
		PromoteKitEnabled:             settings[SettingKeyWebPromoteKitEnabled] == "true",
		PromoteKitID:                  strings.TrimSpace(settings[SettingKeyWebPromoteKitID]),
		CrispEnabled:                  settings[SettingKeyWebCrispEnabled] == "true",
		CrispWebsiteID:                strings.TrimSpace(settings[SettingKeyWebCrispWebsiteID]),
		TawkEnabled:                   settings[SettingKeyWebTawkEnabled] == "true",
		TawkPropertyID:                strings.TrimSpace(settings[SettingKeyWebTawkPropertyID]),
		TawkWidgetID:                  strings.TrimSpace(settings[SettingKeyWebTawkWidgetID]),
		WebWorkspaceShellConfig:       workspaceShellConfigSetting(settings[SettingKeyWorkspaceShellConfig]),
		WebImagePromptFilterConfig:    strings.TrimSpace(settings[SettingKeyImagePromptFilterConfig]),
		WebPricingTitle:               strings.TrimSpace(settings[SettingKeyPricingTitle]),
		WebPricingDescription:         strings.TrimSpace(settings[SettingKeyPricingDescription]),
		WebPricingShellConfig:         pricingShellConfigSetting(settings[SettingKeyPricingShellConfig]),
		WebPaymentShellConfig:         paymentShellConfigSetting(settings[SettingKeyPaymentShellConfig]),
		WebPricingCurrencySymbol:      pricingCurrencySymbolSetting(settings[SettingKeyPricingCurrencySymbol]),
		WebCreditsTitle:               strings.TrimSpace(settings[SettingKeyCreditsTitle]),
		WebCreditsDescription:         strings.TrimSpace(settings[SettingKeyCreditsDescription]),
		WebCreditsPurchaseLabel:       strings.TrimSpace(settings[SettingKeyCreditsPurchaseLabel]),
		WebCreditsBalanceLabel:        strings.TrimSpace(settings[SettingKeyCreditsBalanceLabel]),
		WebCreditsPerBalance:          creditsPerBalanceSetting(settings[SettingKeyCreditsPerBalance]),
		WebLocaleDetectEnabled:        settings[SettingKeyWebLocaleDetectEnabled] == "true",
		WebEmailAuthVisible:           webEmailVisible,
		WebGoogleAuthVisible:          webGoogleVisible,
		WebGitHubAuthVisible:          webGitHubVisible,
		WebGoogleAnalyticsID:          strings.TrimSpace(settings[SettingKeyWebGoogleAnalyticsID]),
		WebClarityID:                  strings.TrimSpace(settings[SettingKeyWebClarityID]),
		WebPlausibleDomain:            strings.TrimSpace(settings[SettingKeyWebPlausibleDomain]),
		WebPlausibleSrc:               strings.TrimSpace(settings[SettingKeyWebPlausibleSrc]),
		WebOpenPanelClientID:          strings.TrimSpace(settings[SettingKeyWebOpenPanelClientID]),
		WebPublicIntegrationsEnabled:  !isFalseSettingValue(settings[SettingKeyWebPublicIntegrationsEnabled]),
		WebVercelAnalyticsEnabled:     settings[SettingKeyWebVercelAnalyticsEnabled] == "true",
		WebAdsenseCode:                strings.TrimSpace(settings[SettingKeyWebAdsenseCode]),
		WebAffonsoEnabled:             settings[SettingKeyWebAffonsoEnabled] == "true",
		WebAffonsoID:                  strings.TrimSpace(settings[SettingKeyWebAffonsoID]),
		WebAffonsoCookieDuration:      webAffonsoCookieDurationSetting(settings[SettingKeyWebAffonsoCookieDuration]),
		WebPromoteKitEnabled:          settings[SettingKeyWebPromoteKitEnabled] == "true",
		WebPromoteKitID:               strings.TrimSpace(settings[SettingKeyWebPromoteKitID]),
		WebCrispEnabled:               settings[SettingKeyWebCrispEnabled] == "true",
		WebCrispWebsiteID:             strings.TrimSpace(settings[SettingKeyWebCrispWebsiteID]),
		WebTawkEnabled:                settings[SettingKeyWebTawkEnabled] == "true",
		WebTawkPropertyID:             strings.TrimSpace(settings[SettingKeyWebTawkPropertyID]),
		WebTawkWidgetID:               strings.TrimSpace(settings[SettingKeyWebTawkWidgetID]),
	}, nil
}

func creditsPerBalanceSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || value == "10" {
		return "1"
	}
	return value
}

func pricingCurrencySymbolSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "¥"
	}
	return value
}

func webAffonsoCookieDurationSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultWebAffonsoCookieDuration
	}
	return value
}

const defaultWorkspaceShellConfig = `{"zh":{"catalogLabel":"提示词案例","eyebrow":"生图工作台","title":"AI 生图工作台","heroDescription":"从案例库带入提示词，选择模型和参数后直接创建 Sub2API 生图任务。","draftImported":"已导入「{title}」","draftImportedDescription":"提示词已填入工作台，可以继续调整参数后生成。","promptLabel":"提示词","promptPlaceholder":"输入或从案例库导入提示词","promptTooLong":"提示词过长","clearLabel":"清空","copyPromptLabel":"复制提示词","copySuccessMessage":"提示词已复制","copyEmptyError":"请先输入提示词","workspaceTitle":"任务与产物状态","workspaceDescription":"模型配置、任务历史和产物存储由 Sub2API 生图工作台统一管理。","workspaceStatus":"登录后可创建真实生图任务，worker 会调用配置的上游模型并回写图片产物。","backToCatalogLabel":"返回案例库"},"en":{"catalogLabel":"Prompt catalog","eyebrow":"Image Workspace","title":"AI Image Workspace","heroDescription":"Bring prompts from the catalog, choose a model and parameters, then create a native Sub2API image task.","draftImported":"Imported \"{title}\"","draftImportedDescription":"The prompt is ready in the workspace. Adjust parameters before generating.","promptLabel":"Prompt","promptPlaceholder":"Enter a prompt or import one from the catalog","promptTooLong":"Prompt is too long","clearLabel":"Clear","copyPromptLabel":"Copy prompt","copySuccessMessage":"Prompt copied","copyEmptyError":"Enter a prompt first","workspaceTitle":"Task and artifact status","workspaceDescription":"Model config, task history, and artifact storage are managed by the Sub2API image workspace.","workspaceStatus":"After login, users can create real image tasks; the worker calls the configured upstream model and writes image artifacts back.","backToCatalogLabel":"Back to catalog"}}`

func workspaceShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultWorkspaceShellConfig
	}
	return value
}

const defaultPromptCatalogShellConfig = `{"zh":{"defaults":{"sourceType":"case","hasImage":true,"pageSize":24,"sortBy":"imported_at","sortOrder":"desc","generatorPath":"/image-generator","generatorDraftSource":"sub2api-vue-prompt-catalog"},"labels":{"accountActionAuthenticated":"进入控制台","accountActionAnonymous":"登录","eyebrow":"提示词画廊","title":"提示词案例库","description":"直接浏览 Sub2API 中的提示词案例。筛选和分页由共享 Prompt API 提供。","caseTitle":"提示词案例库","caseDescription":"直接浏览 Sub2API 中的提示词案例。筛选和分页由共享 Prompt API 提供。","templateTitle":"提示词模板库","templateDescription":"直接浏览 Sub2API 中的提示词模板。筛选和分页由共享 Prompt API 提供。","total":"总数","sources":"来源","cases":"案例","templates":"模板","search":"搜索","searchPlaceholder":"搜索标题、提示词、标签或来源","caseOnly":"案例","templateOnly":"模板","allTypes":"全部类型","allSources":"全部来源","allCategories":"全部分类","hasImage":"只看有图","resultPrefix":"结果","page":"页","previous":"上一页","next":"下一页","emptyTitle":"没有匹配的案例","emptyDescription":"换一个关键词、来源或分类再试。","noImage":"暂无图片","source":"查看来源","details":"查看","prompt":"提示词","charUnit":"字符","copyPrompt":"复制提示词","promptCopied":"提示词已复制","generate":"去生图","importTitle":"从链接导入案例","importDescription":"管理员可直接导入 X/Twitter 帖子，图片会通过 Sub2API/R2 同步后进入案例库。","importProviderX":"X / Twitter","importPlaceholder":"粘贴 X/Twitter 帖子链接","importAction":"导入","importing":"导入中...","importSuccess":"已导入案例","importWarnings":"导入提示","loadError":"加载提示词案例失败"}},"en":{"defaults":{"sourceType":"case","hasImage":true,"pageSize":24,"sortBy":"imported_at","sortOrder":"desc","generatorPath":"/image-generator","generatorDraftSource":"sub2api-vue-prompt-catalog"},"labels":{"accountActionAuthenticated":"Dashboard","accountActionAnonymous":"Log in","eyebrow":"Prompt Catalog","title":"Prompt Catalog","description":"Browse prompt cases directly from Sub2API. Filtering and pagination are served by the shared prompt API.","caseTitle":"Prompt Catalog","caseDescription":"Browse prompt cases directly from Sub2API. Filtering and pagination are served by the shared prompt API.","templateTitle":"Prompt Templates","templateDescription":"Browse prompt templates directly from Sub2API. Filtering and pagination are served by the shared prompt API.","total":"Total","sources":"Sources","cases":"Cases","templates":"Templates","search":"Search","searchPlaceholder":"Search titles, prompts, tags, or sources","caseOnly":"Cases","templateOnly":"Templates","allTypes":"All types","allSources":"All sources","allCategories":"All categories","hasImage":"Images only","resultPrefix":"Results","page":"Page","previous":"Previous","next":"Next","emptyTitle":"No matching prompts","emptyDescription":"Try another keyword, source, or category.","noImage":"No image","source":"View source","details":"Details","prompt":"Prompt","charUnit":"chars","copyPrompt":"Copy prompt","promptCopied":"Prompt copied","generate":"Use in generator","importTitle":"Import from link","importDescription":"Admins can import X/Twitter posts directly. Images are synced through Sub2API/R2 before entering the catalog.","importProviderX":"X / Twitter","importPlaceholder":"Paste an X/Twitter post URL","importAction":"Import","importing":"Importing...","importSuccess":"Imported prompt case","importWarnings":"Import warnings","loadError":"Failed to load prompt cases"}}}`

func promptCatalogShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultPromptCatalogShellConfig
	}
	if !json.Valid([]byte(value)) {
		return defaultPromptCatalogShellConfig
	}
	return value
}

const defaultDashboardShellConfig = `{"zh":{"labels":{"balance":"余额","available":"可用","apiKeys":"API Keys","active":"活跃","todayRequests":"今日请求","total":"总计","todayCost":"今日成本","actual":"实际","standard":"标准","todayTokens":"今日 Token","totalTokens":"总 Token","input":"输入","output":"输出","cacheWrite":"缓存写入","cacheRead":"缓存读取","performance":"性能","avgResponse":"平均响应","averageTime":"平均耗时","platformBreakdown":"平台拆分","platformCount":"{count} 个平台","platformOther":"其他","requests":"请求","tokens":"Token","platformQuotaTitle":"额度","platformQuotaDaily":"每日","platformQuotaWeekly":"每周","platformQuotaMonthly":"每月","platformQuotaDisabled":"已禁用","platformQuotaResetsAt":"重置于 {time}","recentUsage":"最近使用","last7Days":"最近 7 天","noUsageRecords":"暂无使用记录","startUsingApi":"开始使用 API 后会在这里显示最近请求。","viewAllUsage":"查看全部用量","timeRange":"时间范围","refresh":"刷新","granularity":"粒度","day":"天","hour":"小时","modelDistribution":"模型分布","noDataAvailable":"暂无数据","model":"模型","quickActions":"快捷操作","createApiKey":"创建 API Key","generateNewKey":"生成新的访问密钥","viewUsage":"查看用量","checkDetailedLogs":"查看详细请求日志","redeemCode":"兑换码","addBalanceWithCode":"使用兑换码增加余额"}},"en":{"labels":{"balance":"Balance","available":"Available","apiKeys":"API Keys","active":"active","todayRequests":"Today requests","total":"Total","todayCost":"Today cost","actual":"actual","standard":"standard","todayTokens":"Today tokens","totalTokens":"Total tokens","input":"Input","output":"Output","cacheWrite":"Cache write","cacheRead":"Cache read","performance":"Performance","avgResponse":"Average response","averageTime":"Average time","platformBreakdown":"Platform breakdown","platformCount":"{count} platforms","platformOther":"Other","requests":"Requests","tokens":"Tokens","platformQuotaTitle":"Quota","platformQuotaDaily":"Daily","platformQuotaWeekly":"Weekly","platformQuotaMonthly":"Monthly","platformQuotaDisabled":"Disabled","platformQuotaResetsAt":"Resets at {time}","recentUsage":"Recent usage","last7Days":"Last 7 days","noUsageRecords":"No usage records","startUsingApi":"Recent requests will appear here after you start using the API.","viewAllUsage":"View all usage","timeRange":"Time range","refresh":"Refresh","granularity":"Granularity","day":"Day","hour":"Hour","modelDistribution":"Model distribution","noDataAvailable":"No data available","model":"Model","quickActions":"Quick actions","createApiKey":"Create API key","generateNewKey":"Generate a new access key","viewUsage":"View usage","checkDetailedLogs":"Check detailed request logs","redeemCode":"Redeem code","addBalanceWithCode":"Add balance with a code"}}}`

func dashboardShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultDashboardShellConfig
	}
	return value
}

const defaultPricingShellConfig = `{"zh":{"button":{"title":"去购买"},"groups":[{"name":"one-time","title":"充值包"},{"name":"subscription","title":"订阅包"}],"labels":{"prompts":"提示词案例","eyebrow":"价格","title":"价格与套餐","description":"浏览由 Sub2API 统一配置的充值包和订阅套餐，选择后进入统一支付流程。","catalogStatus":"目录状态","rechargeProducts":"充值包","subscriptionPlans":"订阅包","recharge":"充值包","subscription":"订阅包","buy":"去购买","rechargeCta":"购买充值包","subscriptionCta":"购买订阅包","loadFailed":"价格目录加载失败，请稍后重试。","emptyRecharge":"暂未配置充值包。","emptyPlans":"暂未配置订阅包。","recommended":"推荐","creditedBalance":"到账余额","rate":"倍率","quota":"额度","unlimited":"不限","day":"天","days":"天","month":"月"}},"en":{"button":{"title":"Buy"},"groups":[{"name":"one-time","title":"Recharge"},{"name":"subscription","title":"Subscription"}],"labels":{"prompts":"Prompt cases","eyebrow":"Pricing","title":"Pricing","description":"Browse recharge products and subscription plans configured by Sub2API, then continue to the unified checkout flow.","catalogStatus":"Catalog status","rechargeProducts":"Recharge products","subscriptionPlans":"Subscription plans","recharge":"Recharge","subscription":"Subscription","buy":"Buy","rechargeCta":"Buy balance","subscriptionCta":"Buy subscription","loadFailed":"Failed to load the pricing catalog. Please try again later.","emptyRecharge":"No recharge products are configured yet.","emptyPlans":"No subscription plans are configured yet.","recommended":"Recommended","creditedBalance":"Credited balance","rate":"Rate","quota":"Quota","unlimited":"Unlimited","day":"day","days":"days","month":"month"}}}`

func pricingShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultPricingShellConfig
	}
	return value
}

const defaultPaymentShellConfig = `{"zh":{"labels":{"tabTopUp":"充值","tabSubscribe":"订阅","rechargeAccount":"充值账户","currentBalance":"当前余额","notAvailable":"支付暂不可用","noRechargeProducts":"暂未配置充值商品","rechargeProductRecommended":"推荐","rechargeProductCreditLine":"到账 ${amount} 余额","rechargeProductCta":"选择此充值包","paymentMethod":"支付方式","methodAlipay":"支付宝","methodWxpay":"微信支付","methodStripe":"Stripe","methodAirwallex":"Airwallex","success":"支付成功","subscriptionSuccess":"订阅成功","orderId":"订单 ID","orderNo":"订单编号","amount":"金额","payAmount":"实付","confirm":"确认","cancelled":"订单已取消","cancelledDesc":"您已取消本次支付","expired":"订单已过期","expiredDesc":"订单已超时，请重新创建订单","scanAlipay":"支付宝扫码支付","scanAlipayHint":"请使用手机打开支付宝，扫描二维码完成支付","scanWxpay":"微信扫码支付","scanWxpayHint":"请使用手机打开微信，扫描二维码完成支付","scanToPay":"请扫码支付","openPayWindow":"重新打开支付页面","expiresIn":"剩余支付时间","waitingPayment":"等待支付...","cancelOrder":"取消订单","payInNewWindowHint":"支付页面已在新窗口打开，请在新窗口中完成支付后返回此页面","paymentAmount":"支付金额","fee":"手续费","actualPay":"实付","creditedBalance":"到账余额","rechargeRatePreview":"充值汇率：1 CNY = {usd} USD 余额","processing":"处理中...","createOrder":"创建订单","cancel":"取消","selectAmountFirst":"请选择充值商品","amountNoMethod":"当前充值商品没有可用支付方式","amountTooLow":"金额不能低于 {min}","amountTooHigh":"金额不能高于 {max}","amountLabel":"金额","noPlans":"暂无订阅套餐","activeSubscription":"当前订阅","selectPlan":"选择套餐","groupFallback":"分组 #{id}","daysRemaining":"剩余 {days} 天","noExpiration":"永久有效","activeStatus":"生效中","rate":"倍率","dailyLimit":"日限额","weeklyLimit":"周限额","monthlyLimit":"月限额","quota":"额度","unlimited":"不限","models":"模型","subscribeNow":"立即开通","renewNow":"续费","perMonth":"月","perYear":"年","days":"天","baseAmount":"充值金额","creditedAmount":"到账金额","status":"状态","failed":"支付失败","processingHint":"支付结果仍在确认中，页面会自动刷新。","backToRecharge":"返回充值","viewOrders":"查看订单","stripeLoadFailed":"支付组件加载失败，请刷新页面重试","stripeMissingParams":"缺少订单ID或支付密钥","stripeNotConfigured":"Stripe 未配置","stripePay":"立即支付","stripeSuccessProcessing":"支付成功，正在处理订单...","airwallexLoadFailed":"Airwallex 支付组件加载失败","airwallexMissingParams":"缺少 Airwallex 支付参数","close":"关闭","stripePopupRedirecting":"正在跳转到支付页面...","stripePopupLoadingQr":"正在获取微信支付二维码...","stripePopupTimeout":"等待支付凭证超时，请重试","payInNewWindow":"支付页面已在新窗口打开","wechatPaymentCallbackTitle":"正在恢复微信支付","wechatPaymentCallbackProcessing":"正在恢复微信支付...","wechatPaymentCallbackBackToPayment":"返回支付页","wechatPaymentCallbackMissingResumeToken":"微信支付回调缺少恢复令牌。","tooManyPending":"待支付订单过多，请完成或取消后再试（最多 {max} 个）","cancelRateLimited":"取消订单过于频繁，请稍后再试","mobilePaymentFallbackToQr":"当前环境无法直接唤起支付，已切换为扫码支付","refresh":"刷新","all":"全部","pending":"待支付","completed":"已完成","refunded":"已退款","statusPending":"待支付","statusPaid":"已支付","statusRecharging":"充值中","statusCompleted":"已完成","statusExpired":"已过期","statusCancelled":"已取消","statusFailed":"失败","statusRefundRequested":"已申请退款","statusRefunding":"退款中","statusRefunded":"已退款","statusPartiallyRefunded":"部分退款","statusRefundFailed":"退款失败","actions":"操作","requestRefund":"申请退款","confirmCancel":"确定要取消这个订单吗？","refundReason":"退款原因","refundReasonPlaceholder":"请填写退款原因","cancelSuccess":"订单已取消","refundSuccess":"退款申请已提交","errorFallback":"操作失败","createdAt":"创建时间","subscriptionNoActive":"暂无有效订阅","subscriptionNoActiveDesc":"您没有任何有效订阅。请联系管理员获取订阅。","subscriptionExpires":"到期时间","subscriptionNoExpiration":"无到期时间","subscriptionStatusActive":"有效","subscriptionStatusExpired":"已过期","subscriptionStatusRevoked":"已撤销","subscriptionDaily":"每日","subscriptionWeekly":"每周","subscriptionMonthly":"每月","subscriptionUnlimited":"无限制","subscriptionUnlimitedDesc":"该订阅无用量限制","subscriptionDaysRemaining":"剩余 {days} 天","subscriptionResetIn":"{time} 后重置","subscriptionQuotaEndsIn":"额度将在 {time} 后重置","subscriptionWindowNotActive":"等待首次使用","subscriptionToday":"今天","subscriptionTomorrow":"明天","subscriptionFailedToLoad":"加载订阅失败"}},"en":{"labels":{"tabTopUp":"Top Up","tabSubscribe":"Subscribe","rechargeAccount":"Recharge Account","currentBalance":"Current Balance","notAvailable":"Payment Not Available","noRechargeProducts":"No recharge products configured","rechargeProductRecommended":"Recommended","rechargeProductCreditLine":"Credited ${amount} balance","rechargeProductCta":"Select this package","paymentMethod":"Payment Method","methodAlipay":"Alipay","methodWxpay":"WeChat Pay","methodStripe":"Stripe","methodAirwallex":"Airwallex","success":"Payment Successful","subscriptionSuccess":"Subscription Successful","orderId":"Order ID","orderNo":"Order No.","amount":"Amount","payAmount":"Paid","confirm":"Confirm","cancelled":"Order Cancelled","cancelledDesc":"You have cancelled this payment.","expired":"Order Expired","expiredDesc":"This order has expired. Please create a new one.","scanAlipay":"Alipay QR Payment","scanAlipayHint":"Open Alipay on your phone and scan the QR code to pay","scanWxpay":"WeChat QR Payment","scanWxpayHint":"Open WeChat on your phone and scan the QR code to pay","scanToPay":"Scan to Pay","openPayWindow":"Reopen Payment Page","expiresIn":"Expires in","waitingPayment":"Waiting for payment...","cancelOrder":"Cancel Order","payInNewWindowHint":"The payment page has opened in a new window. Please complete the payment there and return to this page.","paymentAmount":"Payment Amount","fee":"Fee","actualPay":"Actual Payment","creditedBalance":"Credited Balance","rechargeRatePreview":"Recharge rate: 1 CNY = {usd} USD balance","processing":"Processing...","createOrder":"Create Order","cancel":"Cancel","selectAmountFirst":"Select a recharge product","amountNoMethod":"No payment method is available for this recharge product","amountTooLow":"Amount cannot be lower than {min}","amountTooHigh":"Amount cannot be higher than {max}","amountLabel":"Amount","noPlans":"No plans available","activeSubscription":"Active Subscription","selectPlan":"Select Plan","groupFallback":"Group #{id}","daysRemaining":"{days} days remaining","noExpiration":"No expiration","activeStatus":"Active","rate":"Rate","dailyLimit":"Daily Limit","weeklyLimit":"Weekly Limit","monthlyLimit":"Monthly Limit","quota":"Quota","unlimited":"Unlimited","models":"Models","subscribeNow":"Subscribe Now","renewNow":"Renew","perMonth":"month","perYear":"year","days":"days","baseAmount":"Base Amount","creditedAmount":"Credited Amount","status":"Status","failed":"Payment Failed","processingHint":"Payment confirmation is still pending. This page will refresh automatically.","backToRecharge":"Back to Recharge","viewOrders":"View Orders","stripeLoadFailed":"Failed to load payment component. Please refresh and try again.","stripeMissingParams":"Missing order ID or client secret","stripeNotConfigured":"Stripe is not configured","stripePay":"Pay Now","stripeSuccessProcessing":"Payment successful, processing your order...","airwallexLoadFailed":"Failed to load Airwallex checkout","airwallexMissingParams":"Missing Airwallex payment parameters","close":"Close","stripePopupRedirecting":"Redirecting to payment page...","stripePopupLoadingQr":"Loading WeChat Pay QR code...","stripePopupTimeout":"Timed out waiting for payment credentials, please retry","payInNewWindow":"Payment page opened in a new window","wechatPaymentCallbackTitle":"Resuming WeChat payment","wechatPaymentCallbackProcessing":"Resuming WeChat payment...","wechatPaymentCallbackBackToPayment":"Back to payment","wechatPaymentCallbackMissingResumeToken":"WeChat payment callback is missing a resume token.","tooManyPending":"Too many pending orders. Complete or cancel one first (max {max}).","cancelRateLimited":"Order cancellation is rate limited. Please try again later.","mobilePaymentFallbackToQr":"This environment cannot open the payment sheet directly, so QR payment is shown instead.","refresh":"Refresh","all":"All","pending":"Pending","completed":"Completed","refunded":"Refunded","statusPending":"Pending","statusPaid":"Paid","statusRecharging":"Recharging","statusCompleted":"Completed","statusExpired":"Expired","statusCancelled":"Cancelled","statusFailed":"Failed","statusRefundRequested":"Refund requested","statusRefunding":"Refunding","statusRefunded":"Refunded","statusPartiallyRefunded":"Partially refunded","statusRefundFailed":"Refund failed","actions":"Actions","requestRefund":"Request Refund","confirmCancel":"Are you sure you want to cancel this order?","refundReason":"Refund Reason","refundReasonPlaceholder":"Please enter the refund reason","cancelSuccess":"Order cancelled","refundSuccess":"Refund request submitted","errorFallback":"Operation failed","createdAt":"Created At","subscriptionNoActive":"No Active Subscriptions","subscriptionNoActiveDesc":"You don't have any active subscriptions. Contact administrator to get one.","subscriptionExpires":"Expires","subscriptionNoExpiration":"No expiration","subscriptionStatusActive":"Active","subscriptionStatusExpired":"Expired","subscriptionStatusRevoked":"Revoked","subscriptionDaily":"Daily","subscriptionWeekly":"Weekly","subscriptionMonthly":"Monthly","subscriptionUnlimited":"Unlimited","subscriptionUnlimitedDesc":"No usage limits on this subscription","subscriptionDaysRemaining":"{days} days remaining","subscriptionResetIn":"Resets in {time}","subscriptionQuotaEndsIn":"Quota resets in {time}","subscriptionWindowNotActive":"Awaiting first use","subscriptionToday":"Today","subscriptionTomorrow":"Tomorrow","subscriptionFailedToLoad":"Failed to load subscriptions"}}}`

func paymentShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultPaymentShellConfig
	}
	return value
}

const defaultCreditsShellConfig = `{"zh":{"labels":{"eyebrow":"余额","title":"余额","description":"前端和后端统一使用余额口径；后端 balance 字段仍是唯一账本字段。","purchase":"充值余额","orders":"订单记录","credits":"余额","sub2apiBalance":"账本余额","conversion":"统一口径：1 余额单位 = 1 balance 账本单位。","balanceLabel":"账本余额：{balance}","actionsTitle":"余额操作","actionsDescription":"充值、订单、退款等流程均进入 Sub2API 统一支付体系，最终写入同一份 balance 账本。","recharge":"去充值","viewOrders":"查看订单"},"actions":{"title":"余额操作","description":"充值、订单、退款等流程均进入 Sub2API 统一支付体系，最终写入同一份 balance 账本。"},"buttons":{"recharge":"去充值","orders":"查看订单"},"conversion":"统一口径：1 余额单位 = 1 balance 账本单位。"},"en":{"labels":{"eyebrow":"Balance","title":"Balance","description":"The frontend and backend use the same balance terminology; the backend balance field remains the only ledger field.","purchase":"Recharge balance","orders":"Orders","credits":"Balance","sub2apiBalance":"Ledger balance","conversion":"Unified unit: 1 balance unit = 1 ledger unit.","balanceLabel":"Ledger balance: {balance}","actionsTitle":"Balance actions","actionsDescription":"Recharge, orders, and refunds go through the unified Sub2API payment flow and write to the same balance ledger.","recharge":"Recharge","viewOrders":"View orders"},"actions":{"title":"Balance actions","description":"Recharge, orders, and refunds go through the unified Sub2API payment flow and write to the same balance ledger."},"buttons":{"recharge":"Recharge","orders":"View orders"},"conversion":"Unified unit: 1 balance unit = 1 ledger unit."}}`

func creditsShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultCreditsShellConfig
	}
	return value
}

const defaultHomeShellConfig = `{"zh":{"labels":{"viewDocs":"文档","dashboard":"控制台","login":"登录","primaryCta":"立即开始","secondaryCta":"浏览模型","heroBadge":"开发者首选","heroTitle":"AI 编码工作台","heroDescription":"无需管理多个订阅账号，一站式接入 Claude、GPT 等主流 AI 服务。","modelMatrixKicker":"模型矩阵","modelMatrixTitle":"一个工作台连接 Claude 与 GPT","modelMatrixDescription":"从后台目录读取模型族和能力标签，保持公开页面和实际售卖能力一致。","modelMatrixEmptyCard":"配置模型后会自动出现在这里。","modelMatrixEmptyPill":"即将上线","experienceKicker":"体验","experienceTitle":"更清晰的模型访问流程","experienceDescription":"把模型访问、支付、文档和案例目录统一在一个平台里。","whyChooseKicker":"为什么选择","whyChooseTitle":"面向日常 AI 工作","whyChooseDescription":"更克制的产品形态、更清晰的价格和更贴近日常编码的工作流。","footerDescription":"更简单的模型访问入口，提供清晰价格和日常 AI 辅助编码体验。","allRightsReserved":"保留所有权利。","termsLink":"服务条款","privacyLink":"隐私政策","navHome":"首页","navDocs":"文档","navModels":"模型","navExperience":"体验","footerProduct":"产品","footerCatalog":"目录","footerSupport":"支持","familyClaudeBadge":"Claude","familyGptBadge":"GPT","familyClaudeTagline":"偏重推理的编码模型","familyGptTagline":"快速迭代和智能体","familyClaudeDescription":"适合深度推理、架构设计和代码审查。","familyGptDescription":"适合功能开发、快速迭代和智能体工作流。","familyClaudeReasoning":"深度推理","familyClaudeArchitecture":"架构设计","familyClaudeReview":"代码审查","familyGptCoding":"代码生成","familyGptIteration":"快速迭代","familyGptAgents":"智能体"},"experienceCards":[{"key":"unified","icon":"server","iconClass":"bg-gradient-to-br from-sky-500 to-blue-600","title":"一个密钥统一接入","description":"统一域名和密钥格式，减少在不同模型和工具之间来回切换。"},{"key":"setup","icon":"key","iconClass":"bg-gradient-to-br from-indigo-500 to-violet-600","title":"配置更轻","description":"更贴近 CLI、IDE 与日常开发习惯，不把大量时间花在环境变量和接线细节上。"},{"key":"stability","icon":"sparkles","iconClass":"bg-gradient-to-br from-emerald-500 to-teal-600","title":"链路更稳","description":"通过账号池与路由能力降低单点限制带来的中断，让高频编码更连续。"},{"key":"billing","icon":"chart","iconClass":"bg-gradient-to-br from-fuchsia-500 to-purple-600","title":"计费更透明","description":"充值、订阅和后续用量都公开可见，个人和小团队更容易控成本。"}],"whyChooseCards":[{"key":"lowFriction","title":"少折腾配置","description":"把分散在多个模型入口和订阅账号里的接入复杂度压缩成统一体验。"},{"key":"transparent","title":"模型一眼看清","description":"首页直接展示主力模型家族，开发者在登录前就能判断是否适合自己的工作流。"},{"key":"routing","title":"更适合高频编码","description":"强调链路稳定性与编码工作流，而不是堆叠泛化功能。"},{"key":"team","title":"适配个人与小团队","description":"既适合独立开发者快速上手，也方便小团队统一入口和管理预算。"}]},"en":{"labels":{"viewDocs":"Docs","dashboard":"Dashboard","login":"Log in","primaryCta":"Start now","secondaryCta":"Browse models","heroBadge":"Developer First","heroTitle":"AI Coding Workspace","heroDescription":"Access Claude, GPT, and other core AI services in one place without managing multiple subscriptions.","modelMatrixKicker":"Model Matrix","modelMatrixTitle":"One workspace for Claude and GPT","modelMatrixDescription":"Browse configured model families and capabilities from the backend catalog.","modelMatrixEmptyCard":"Configured models will appear here automatically.","modelMatrixEmptyPill":"Coming soon","experienceKicker":"Experience","experienceTitle":"A cleaner model access flow","experienceDescription":"Keep model access, payments, docs, and catalog discovery in one platform.","whyChooseKicker":"Why choose us","whyChooseTitle":"Built for daily AI work","whyChooseDescription":"A more restrained product shape, clearer pricing, and workflows closer to day-to-day coding.","footerDescription":"A simpler entry point for model access, visible pricing, and day-to-day AI-assisted coding.","allRightsReserved":"All rights reserved.","termsLink":"Terms","privacyLink":"Privacy","navHome":"Home","navDocs":"Docs","navModels":"Models","navExperience":"Experience","footerProduct":"Product","footerCatalog":"Catalog","footerSupport":"Support","familyClaudeBadge":"Claude","familyGptBadge":"GPT","familyClaudeTagline":"Reasoning-first coding","familyGptTagline":"Fast iteration and agents","familyClaudeDescription":"Use Claude models for deep reasoning, architecture, and review-heavy work.","familyGptDescription":"Use GPT models for coding, iteration, and agentic workflows.","familyClaudeReasoning":"Deep reasoning","familyClaudeArchitecture":"Architecture","familyClaudeReview":"Code review","familyGptCoding":"Coding","familyGptIteration":"Iteration","familyGptAgents":"Agents"},"experienceCards":[{"key":"unified","icon":"server","iconClass":"bg-gradient-to-br from-sky-500 to-blue-600","title":"One key, unified access","description":"Use one consistent domain and key format instead of juggling multiple providers and setup flows."},{"key":"setup","icon":"key","iconClass":"bg-gradient-to-br from-indigo-500 to-violet-600","title":"Lower setup friction","description":"Designed to fit better with CLI tools, IDE plugins, and the development habits people already have."},{"key":"stability","icon":"sparkles","iconClass":"bg-gradient-to-br from-emerald-500 to-teal-600","title":"More stable routing","description":"Account-pool and routing capabilities help reduce interruptions caused by single-path limits."},{"key":"billing","icon":"chart","iconClass":"bg-gradient-to-br from-fuchsia-500 to-purple-600","title":"More transparent billing","description":"Recharge products, plans, and usage stay visible so developers can control spend."}],"whyChooseCards":[{"key":"lowFriction","title":"Less setup overhead","description":"Compress scattered model and provider setup into a more unified experience built for repeat coding use."},{"key":"transparent","title":"Models visible at a glance","description":"The homepage surfaces the core model families directly so developers can judge the fit before signing in."},{"key":"routing","title":"Focused on coding throughput","description":"The product emphasizes coding workflows and routing stability instead of loading the homepage with unrelated platform features."},{"key":"team","title":"Fits solo devs and small teams","description":"Simple enough for individual developers to adopt quickly, while still giving small teams a cleaner shared entry point."}]}}`
const defaultHomeBusinessShellConfig = `{"zh":{"labels":{"viewDocs":"文档","dashboard":"控制台","login":"登录","primaryCta":"进入能力中台","secondaryCta":"查看图片提示词","heroBadge":"业务能力首页","heroTitle":"面向业务场景的 AI 能力工作台","heroDescription":"Sub2API 以后沉淀用户、订单、套餐、支付等中台能力；首页重点展示微信导出、热点、图片提示词和生图工作台等可直接理解的业务能力。","modelMatrixKicker":"业务能力","modelMatrixTitle":"把高频业务能力摆到首页","modelMatrixDescription":"围绕内容采集、提示词沉淀与图像生产流程，先把可落地的能力入口讲清楚，再由中台承接账户、订单和套餐等底层能力。","modelMatrixEmptyCard":"业务能力即将上线","modelMatrixEmptyPill":"建设中","experienceKicker":"中台定位","experienceTitle":"业务能力前台，平台能力落到中台","experienceDescription":"用户、订单、套餐、支付与账户治理逐步统一收口到 Sub2API 中台，让前台页面更多表达业务价值而不是底层接线细节。","whyChooseKicker":"能力组织方式","whyChooseTitle":"先讲用户能完成什么，再讲平台怎么支撑","whyChooseDescription":"首页围绕业务工作流编排；底层代理、模型路由和结算能力继续由中台承接。","footerDescription":"聚焦业务能力表达，由中台统一承接用户、订单、套餐和支付能力。","allRightsReserved":"保留所有权利。","termsLink":"服务条款","privacyLink":"隐私政策","navHome":"首页","navDocs":"文档","navModels":"提示词","navExperience":"能力","footerProduct":"首页入口","footerCatalog":"业务能力","footerSupport":"支持","familyClaudeBadge":"","familyGptBadge":"","familyClaudeTagline":"","familyGptTagline":"","familyClaudeDescription":"","familyGptDescription":"","familyClaudeReasoning":"","familyClaudeArchitecture":"","familyClaudeReview":"","familyGptCoding":"","familyGptIteration":"","familyGptAgents":""},"businessCards":[{"key":"wechat-export","badge":"Workflow","title":"微信导出","description":"沉淀公众号内容导出与整理能力，适合把文章资产回收到统一工作流里。","capabilityTags":["内容导出","素材整理","资产回收"],"path":"/wechat","pathLabel":"进入微信导出"},{"key":"hot-topics","badge":"Signal","title":"热点追踪","description":"围绕热点发现、筛选和后续处理，把高频内容观察任务做成稳定入口。当前真实页面仍在建设中。","capabilityTags":["热点收集","线索筛选","建设中"]},{"key":"prompt-catalog","badge":"Library","title":"图片提示词","description":"把沉淀下来的图片提示词案例放到统一目录里，便于检索、复用和二次加工。","capabilityTags":["案例目录","检索复用","图像提示词"],"path":"/prompts","pathLabel":"进入提示词库"},{"key":"image-workspace","badge":"Workspace","title":"生图工作台","description":"以提示词工作流为中心组织图片生成前的整理、复制和后续生产衔接。","capabilityTags":["Prompt 工作流","生图准备","工作台"],"path":"/image-generator","pathLabel":"进入工作台"}],"experienceCards":[{"key":"platform","icon":"server","iconClass":"bg-gradient-to-br from-sky-500 to-blue-600","title":"中台统一承接用户与订单","description":"前台页面聚焦业务表达，用户、订单、支付和套餐配置逐步收口到统一能力中台。"},{"key":"catalog","icon":"key","iconClass":"bg-gradient-to-br from-indigo-500 to-violet-600","title":"内容能力先产品化","description":"优先把微信导出、热点、提示词和生图工作流做成稳定能力，再让底层平台持续支撑它们。"},{"key":"ops","icon":"sparkles","iconClass":"bg-gradient-to-br from-emerald-500 to-teal-600","title":"前后台职责更清晰","description":"首页讲业务价值，后台负责配置、数据和运行时控制，减少首页同时承担两种叙事。"}],"whyChooseCards":[{"key":"business-first","title":"先围绕业务入口组织","description":"把用户真正会点开的业务能力放在首页，而不是先暴露中台实现细节。"},{"key":"platform-backbone","title":"中台继续做能力骨架","description":"账户、订单、套餐与支付能力继续沉到 Sub2API 中台，不需要在首页重复解释。"},{"key":"reuse","title":"提示词与内容资产可复用","description":"把图片提示词、导出内容和热点线索组织成可持续复用的业务资产。"},{"key":"workflow","title":"形成工作流闭环","description":"从内容导出、热点发现到提示词沉淀、生图准备，首页直接表达完整业务链路。"}]},"en":{"labels":{"viewDocs":"Docs","dashboard":"Dashboard","login":"Log in","primaryCta":"Open the platform","secondaryCta":"Browse prompt cases","heroBadge":"Business capability home","heroTitle":"An AI workspace organized around business capabilities","heroDescription":"Sub2API will keep consolidating users, orders, plans, and payment into the capability platform while the homepage highlights concrete workflows such as WeChat export, hot-topic tracking, prompt cases, and the image workspace.","modelMatrixKicker":"Capabilities","modelMatrixTitle":"Put business workflows on the homepage","modelMatrixDescription":"Lead with content export, discovery, prompt reuse, and image production workflows while the platform layer continues to own accounts, plans, and billing.","modelMatrixEmptyCard":"Business capability coming soon","modelMatrixEmptyPill":"In progress","experienceKicker":"Platform direction","experienceTitle":"Business-facing home, platform-backed operations","experienceDescription":"Users, orders, plans, payments, and account management continue moving into the Sub2API platform so public pages can focus on user-facing workflows.","whyChooseKicker":"Information architecture","whyChooseTitle":"Explain what users can do before how the platform works","whyChooseDescription":"The homepage should foreground business workflows while the platform continues to power routing, account management, and settlement behind the scenes.","footerDescription":"Homepage messaging focused on business capabilities, backed by a unified platform for users, plans, orders, and payments.","allRightsReserved":"All rights reserved.","termsLink":"Terms","privacyLink":"Privacy","navHome":"Home","navDocs":"Docs","navModels":"Prompts","navExperience":"Capabilities","footerProduct":"Entry points","footerCatalog":"Workflows","footerSupport":"Support","familyClaudeBadge":"","familyGptBadge":"","familyClaudeTagline":"","familyGptTagline":"","familyClaudeDescription":"","familyGptDescription":"","familyClaudeReasoning":"","familyClaudeArchitecture":"","familyClaudeReview":"","familyGptCoding":"","familyGptIteration":"","familyGptAgents":""},"businessCards":[{"key":"wechat-export","badge":"Workflow","title":"WeChat Export","description":"Turn WeChat export and article recovery into a stable workflow entry instead of an ad hoc operation.","capabilityTags":["Content export","Asset recovery","Workflow"],"path":"/wechat","pathLabel":"Open WeChat export"},{"key":"hot-topics","badge":"Signal","title":"Hot Topic Tracking","description":"Package hot-topic discovery and follow-up processing into a clearer product surface. The real page is still in progress.","capabilityTags":["Signal collection","Trend tracking","In progress"]},{"key":"prompt-catalog","badge":"Library","title":"Image Prompt Cases","description":"Keep image prompt cases in a searchable catalog so teams can reuse and refine proven material.","capabilityTags":["Prompt library","Search","Reuse"],"path":"/prompts","pathLabel":"Open prompt catalog"},{"key":"image-workspace","badge":"Workspace","title":"Image Workspace","description":"Center the image workflow around prompt preparation and handoff instead of exposing only the platform plumbing.","capabilityTags":["Prompt workflow","Image prep","Workspace"],"path":"/image-generator","pathLabel":"Open workspace"}],"experienceCards":[{"key":"platform","icon":"server","iconClass":"bg-gradient-to-br from-sky-500 to-blue-600","title":"Platform-owned users and orders","description":"The public home can focus on business workflows while user, order, payment, and plan capabilities consolidate behind the platform."},{"key":"catalog","icon":"key","iconClass":"bg-gradient-to-br from-indigo-500 to-violet-600","title":"Productize content workflows first","description":"Lead with WeChat export, hot topics, prompt cases, and image preparation instead of putting infrastructure copy first."},{"key":"ops","icon":"sparkles","iconClass":"bg-gradient-to-br from-emerald-500 to-teal-600","title":"Clearer split between home and platform","description":"The homepage explains user-facing workflows; the platform continues to own runtime controls, routing, and settlement."}],"whyChooseCards":[{"key":"business-first","title":"Organize around workflows users recognize","description":"Put the workflows people actually want to enter from the homepage ahead of the supporting platform internals."},{"key":"platform-backbone","title":"Keep the platform as the backbone","description":"Users, orders, plans, and payment keep consolidating into Sub2API without forcing every homepage section to explain the machinery."},{"key":"reuse","title":"Make prompts and content reusable assets","description":"Turn prompt cases, exported content, and hot-topic findings into assets that can be searched, refined, and reused."},{"key":"workflow","title":"Show a complete workflow story","description":"Move from export and topic discovery into prompt curation and image preparation with a clearer end-to-end capability narrative."}]}}`

func homeShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultHomeShellConfig
	}
	return value
}

func homeBusinessShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = defaultHomeBusinessShellConfig
	}
	return normalizeHomeBusinessShellConfig(value)
}

func normalizeHomeBusinessShellConfig(raw string) string {
	var root map[string]any
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return raw
	}

	normalizeLocale := func(locale string, value any) {
		localized, ok := value.(map[string]any)
		if !ok {
			return
		}
		cards, ok := localized["businessCards"].([]any)
		if !ok {
			return
		}
		for _, item := range cards {
			card, ok := item.(map[string]any)
			if !ok {
				continue
			}
			key, _ := card["key"].(string)
			status, hasStatus := card["status"].(string)
			if !hasStatus || status == "" {
				status = "available"
				card["status"] = status
			}
			if _, ok := card["statusLabel"].(string); !ok {
				card["statusLabel"] = homeBusinessStatusLabel(locale, status)
			}
			if _, ok := card["disabled"].(bool); !ok {
				card["disabled"] = false
			}
			if _, ok := card["visible"].(bool); !ok {
				card["visible"] = true
			}
			if key == "hot-topics" {
				normalizeHotTopicsBusinessCard(locale, card, hasStatus)
			}
		}
	}

	for _, locale := range []string{"zh", "en"} {
		if localized, ok := root[locale]; ok {
			normalizeLocale(locale, localized)
		}
	}
	if _, hasLocalized := root["zh"]; !hasLocalized {
		normalizeLocale("zh", root)
	}

	normalized, err := json.Marshal(root)
	if err != nil {
		return raw
	}
	return string(normalized)
}

func normalizeHotTopicsBusinessCard(locale string, card map[string]any, hadExplicitStatus bool) {
	if _, ok := card["path"].(string); !ok && !hadExplicitStatus {
		card["path"] = "/hot"
		card["pathLabel"] = localizedHomeBusinessText(locale, "进入热点追踪", "Open hot topics")
	}
	if description, ok := card["description"].(string); ok {
		if strings.Contains(description, "建设中") || strings.Contains(description, "still in progress") {
			card["description"] = localizedHomeBusinessText(locale,
				"围绕热点发现、筛选和后续处理，把高频内容观察任务做成稳定入口。",
				"Package hot-topic discovery and follow-up processing into a clearer product surface.",
			)
		}
	}
	if tags, ok := card["capabilityTags"].([]any); ok {
		for i, tag := range tags {
			text, _ := tag.(string)
			if text == "建设中" {
				tags[i] = "内容采集"
			}
			if text == "In progress" {
				tags[i] = "Content collection"
			}
		}
	}
}

func homeBusinessStatusLabel(locale, status string) string {
	switch status {
	case "in_progress":
		return localizedHomeBusinessText(locale, "建设中", "In progress")
	case "disabled":
		return localizedHomeBusinessText(locale, "暂不可用", "Disabled")
	case "hidden":
		return ""
	default:
		return localizedHomeBusinessText(locale, "可用", "Available")
	}
}

func localizedHomeBusinessText(locale, zh, en string) string {
	if locale == "en" {
		return en
	}
	return zh
}

func imageWorkspaceModelConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || !json.Valid([]byte(value)) {
		return defaultImageWorkspaceModelConfig
	}
	return value
}

const defaultModelPlazaShellConfig = `{"zh":{"labels":{"viewDocs":"文档","dashboard":"控制台","login":"登录","badge":"模型广场","title":"公开模型目录","description":"从后台直接配置并公开展示可售模型卡片。适合做模型能力说明、价格展示和统一入口。","emptyTitle":"模型广场暂未配置","emptyDescription":"管理员完成模型广场配置后，这里会展示公开模型卡片。","quickFind":"快速查找","searchLabel":"搜索模型广场","searchPlaceholder":"搜索模型、能力或标签","groupsTitle":"平台分组","currentSearch":"当前搜索：{query}","browseHint":"按平台分组浏览公开模型卡片。","results":"结果","emptyFilteredTitle":"没有匹配的模型卡片","emptyFilteredDescription":"试试切换分组，或者换一个更宽松的关键词搜索。","copyModelIds":"复制模型 ID","modelIdsCopied":"模型 ID 已复制","inputPrice":"输入价格","outputPrice":"输出价格","cacheReadPrice":"缓存读取价格","cacheWritePrice":"缓存创建价格","modelIdsConfigured":"已配置模型 ID","groupAll":"全部模型","groupOther":"其他"}},"en":{"labels":{"viewDocs":"Docs","dashboard":"Dashboard","login":"Log in","badge":"Model Plaza","title":"Public Model Catalog","description":"Configure and publish model cards directly from the admin backend for capability overviews, pricing communication, and a unified entry point.","emptyTitle":"Model plaza is not configured yet","emptyDescription":"Once the admin configures model plaza items, public model cards will appear here.","quickFind":"Quick find","searchLabel":"Search model plaza","searchPlaceholder":"Search models, capabilities, or tags","groupsTitle":"Groups","currentSearch":"Current search: {query}","browseHint":"Browse public model cards by provider group.","results":"Results","emptyFilteredTitle":"No matching model cards","emptyFilteredDescription":"Try another group or broaden the search terms.","copyModelIds":"Copy model IDs","modelIdsCopied":"Model IDs copied","inputPrice":"Input price","outputPrice":"Output price","cacheReadPrice":"Cache read price","cacheWritePrice":"Cache write price","modelIdsConfigured":"Model IDs configured","groupAll":"All models","groupOther":"Other"}}}`

func modelPlazaShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultModelPlazaShellConfig
	}
	return value
}

const defaultDocsShellConfig = `{"zh":{"labels":{"title":"文档","dashboard":"控制台","login":"登录","searchPlaceholder":"搜索文档","noData":"没有结果"}},"en":{"labels":{"title":"Docs","dashboard":"Dashboard","login":"Log in","searchPlaceholder":"Search docs","noData":"No results"}}}`

func docsShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultDocsShellConfig
	}
	return value
}

const defaultDocsContentBasePath = `{"zh":"/docs-content/","en":"/docs-content/en/"}`

func docsContentBasePathSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultDocsContentBasePath
	}
	return value
}

const defaultLegalDocumentShellConfig = `{"zh":{"labels":{"login":"登录","agreementLabel":"登录条款","loadFailedTitle":"文档加载失败","loadFailedDescription":"请稍后刷新页面重试。","missingTitle":"文档不存在","missingDescription":"当前条款文档不存在或已被管理员移除。","updatedAt":"更新日期：{date}","emptyContent":"暂无正文内容"}},"en":{"labels":{"login":"Log in","agreementLabel":"Login agreement","loadFailedTitle":"Failed to load document","loadFailedDescription":"Please refresh and try again later.","missingTitle":"Document not found","missingDescription":"This agreement document does not exist or has been removed by an administrator.","updatedAt":"Updated: {date}","emptyContent":"No document content yet"}}}`

func legalDocumentShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultLegalDocumentShellConfig
	}
	return value
}

const defaultKeyUsageShellConfig = `{"zh":{"labels":{"apply":"应用","allRightsReserved":"保留所有权利。","avgDuration":"平均耗时","cacheCreationTokens":"缓存创建","cacheWriteTokens":"缓存写入","cacheReadTokens":"缓存读取","cost":"费用","dailyDetail":"每日明细","date":"日期","dateRange":"统计范围:","dateRange30d":"30 天","dateRange7d":"7 天","dateRange90d":"90 天","dateRangeCustom":"自定义","dateRangeToday":"今日","daysLeft":"({days} 天)","detailInfo":"详细信息","docs":"文档","enterApiKey":"请输入 API Key","expiresAt":"过期时间","inputTokens":"输入 Tokens","limit5h":"5 小时限额","limit7d":"7 天限额","limitDaily":"日限额","limitMonthly":"月限额","limitWeekly":"周限额","model":"模型","modelStats":"模型用量统计","noDailyUsage":"当前筛选范围内没有每日明细数据","outputTokens":"输出 Tokens","placeholder":"sk-ant-mirror-xxxxxxxxxxxx","privacyNote":"您的 Key 仅在浏览器本地处理，不会被存储","query":"查询","queryFailed":"查询失败","queryFailedRetry":"查询失败，请稍后重试","querySuccess":"查询成功","querying":"查询中...","quotaMode":"Key 限额模式","remainingQuota":"剩余额度","requests":"请求数","resetNow":"即将重置","rpmTpm":"RPM / TPM","subscriptionExpires":"订阅到期","subscriptionType":"订阅类型","subtitle":"输入您的 API Key 以查看实时消费金额与使用状态","title":"API Key 用量查询","todayCacheCreation":"今日缓存创建","todayCacheRead":"今日缓存读取","todayCost":"今日费用","todayExpires":"(今日到期)","todayInputTokens":"今日输入","todayOutputTokens":"今日输出","todayRequests":"今日请求","todayTokens":"今日 Tokens","tokenStats":"Token 统计","totalCacheCreation":"累计缓存创建","totalCacheRead":"累计缓存读取","totalCost":"累计费用","totalInputTokens":"累计输入","totalOutputTokens":"累计输出","totalQuota":"总额度","totalRequests":"累计请求","totalTokens":"总 Tokens","totalTokensLabel":"累计 Tokens","used":"已使用","usedQuota":"已用额度","walletBalance":"钱包余额"}},"en":{"labels":{"apply":"Apply","allRightsReserved":"All rights reserved.","avgDuration":"Avg Duration","cacheCreationTokens":"Cache Creation","cacheWriteTokens":"Cache Write","cacheReadTokens":"Cache Read","cost":"Cost","dailyDetail":"Daily Detail","date":"Date","dateRange":"Date Range:","dateRange30d":"30 Days","dateRange7d":"7 Days","dateRange90d":"90 Days","dateRangeCustom":"Custom","dateRangeToday":"Today","daysLeft":"({days} days)","detailInfo":"Detail Information","docs":"Docs","enterApiKey":"Please enter an API Key","expiresAt":"Expires At","inputTokens":"Input Tokens","limit5h":"5-Hour Limit","limit7d":"7-Day Limit","limitDaily":"Daily Limit","limitMonthly":"Monthly Limit","limitWeekly":"Weekly Limit","model":"Model","modelStats":"Model Usage Statistics","noDailyUsage":"No daily usage details in the current range","outputTokens":"Output Tokens","placeholder":"sk-ant-mirror-xxxxxxxxxxxx","privacyNote":"Your Key is processed locally in the browser and will not be stored","query":"Query","queryFailed":"Query failed","queryFailedRetry":"Query failed, please try again later","querySuccess":"Query successful","querying":"Querying...","quotaMode":"Key Quota Mode","remainingQuota":"Remaining Quota","requests":"Requests","resetNow":"Resetting soon","rpmTpm":"RPM / TPM","subscriptionExpires":"Subscription Expires","subscriptionType":"Subscription Type","subtitle":"Enter your API Key to view real-time spending and usage status","title":"API Key Usage","todayCacheCreation":"Today Cache Creation","todayCacheRead":"Today Cache Read","todayCost":"Today Cost","todayExpires":"(expires today)","todayInputTokens":"Today Input","todayOutputTokens":"Today Output","todayRequests":"Today Requests","todayTokens":"Today Tokens","tokenStats":"Token Statistics","totalCacheCreation":"Total Cache Creation","totalCacheRead":"Total Cache Read","totalCost":"Total Cost","totalInputTokens":"Total Input","totalOutputTokens":"Total Output","totalQuota":"Total Quota","totalRequests":"Total Requests","totalTokens":"Total Tokens","totalTokensLabel":"Total Tokens","used":"Used","usedQuota":"Used Quota","walletBalance":"Wallet Balance"}}}`

func keyUsageShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultKeyUsageShellConfig
	}
	return value
}

const defaultUsageShellConfig = `{"zh":{"labels":{"totalRequests":"总请求数","inSelectedRange":"选中范围内","totalTokens":"总 Tokens","in":"输入","out":"输出","totalCost":"总费用","actualCost":"实际费用","standardCost":"标准费用","avgDuration":"平均耗时","perRequest":"每次请求","apiKeyFilter":"API Key","allApiKeys":"全部密钥","timeRange":"时间范围","refresh":"刷新","reset":"重置","exportCsv":"导出 CSV","exporting":"导出中...","model":"模型","reasoningEffort":"推理强度","endpoint":"端点","type":"类型","billingMode":"计费模式","tokens":"Tokens","cost":"费用","firstToken":"首 Token","duration":"耗时","time":"时间","userAgent":"User Agent","noRecords":"暂无使用记录","rate":"倍率","original":"原始","billed":"计费","failedToLoad":"加载使用记录失败","noDataToExport":"没有可导出的数据","preparingExport":"正在准备导出...","exportSuccess":"导出成功","exportFailed":"导出失败"}},"en":{"labels":{"totalRequests":"Total Requests","inSelectedRange":"In selected range","totalTokens":"Total Tokens","in":"In","out":"Out","totalCost":"Total Cost","actualCost":"Actual Cost","standardCost":"Standard Cost","avgDuration":"Avg Duration","perRequest":"Per request","apiKeyFilter":"API Key","allApiKeys":"All API Keys","timeRange":"Time Range","refresh":"Refresh","reset":"Reset","exportCsv":"Export CSV","exporting":"Exporting...","model":"Model","reasoningEffort":"Reasoning Effort","endpoint":"Endpoint","type":"Type","billingMode":"Billing Mode","tokens":"Tokens","cost":"Cost","firstToken":"First Token","duration":"Duration","time":"Time","userAgent":"User Agent","noRecords":"No usage records","rate":"Rate","original":"Original","billed":"Billed","failedToLoad":"Failed to load usage records","noDataToExport":"No data to export","preparingExport":"Preparing export...","exportSuccess":"Export successful","exportFailed":"Export failed"}}}`

func usageShellConfigDefault() string {
	return defaultUsageShellConfig
}

func usageShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return usageShellConfigDefault()
	}
	return value
}

const defaultAPIGuideShellConfig = `{"zh":{"labels":{"badge":"API 调用","title":"网关调用说明","description":"查看当前 API Key 可用的协议、端点、鉴权方式和可复制的 curl 示例。","openTester":"打开在线测试","manageKeys":"管理 API Keys","baseUrl":"Base URL","currentKey":"当前密钥","noSelection":"未选择","selectKeyHint":"请选择一个 API Key","supportedEndpoints":"可用端点","noGroupAssigned":"未分配分组","noKeysTitle":"暂无 API Key","noKeysDescription":"创建 API Key 后即可查看可用网关端点和调用示例。","keySelector":"选择 API Key","keySelectorHint":"选择一个密钥后，将按其分组能力展示可用端点。","unassignedTitle":"该密钥未分配分组","unassignedDescription":"未分配分组的密钥无法确定可用协议和模型，请先在 API Keys 页面绑定分组。","keySummary":"密钥信息","groupName":"分组名称","platform":"平台","status":"状态","authHeaderTitle":"鉴权头","authHeaderDescription":"OpenAI/Anthropic 兼容端点使用 Bearer Token；Google 兼容端点使用 x-goog-api-key。","noEndpointVariants":"当前密钥没有可用端点。","stream":"开启流式输出","testThisVariant":"测试此端点","endpoint":"端点","protocol":"协议","defaultModel":"默认模型","headerMode":"鉴权方式","curlExample":"curl 示例","copyCurl":"复制 curl","copyCurlSuccess":"curl 已复制","defaultPrompt":"用一句话介绍 Sub2API。","loadKeysFailed":"API Keys 加载失败"}},"en":{"labels":{"badge":"API Guide","title":"Gateway API Guide","description":"Review the protocols, endpoints, auth headers, and copy-ready curl examples available to the selected API key.","openTester":"Open Tester","manageKeys":"Manage API Keys","baseUrl":"Base URL","currentKey":"Current Key","noSelection":"No selection","selectKeyHint":"Select an API key","supportedEndpoints":"Supported Endpoints","noGroupAssigned":"No group assigned","noKeysTitle":"No API Keys","noKeysDescription":"Create an API key to view available gateway endpoints and examples.","keySelector":"API Key","keySelectorHint":"Choose a key to show endpoints enabled by its group.","unassignedTitle":"This key has no group","unassignedDescription":"Keys without a group cannot resolve available protocols or models. Assign a group from the API Keys page first.","keySummary":"Key Summary","groupName":"Group Name","platform":"Platform","status":"Status","authHeaderTitle":"Auth Header","authHeaderDescription":"OpenAI/Anthropic compatible endpoints use Bearer Token; Google compatible endpoints use x-goog-api-key.","noEndpointVariants":"No endpoint variants are available for this key.","stream":"Streaming","testThisVariant":"Test this endpoint","endpoint":"Endpoint","protocol":"Protocol","defaultModel":"Default Model","headerMode":"Header Mode","curlExample":"curl Example","copyCurl":"Copy curl","copyCurlSuccess":"curl copied","defaultPrompt":"Introduce Sub2API in one sentence.","loadKeysFailed":"Failed to load API keys"}}}`

func apiGuideShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAPIGuideShellConfig
	}
	return value
}

const defaultAPITestShellConfig = `{"zh":{"labels":{"badge":"Live Request","title":"调用测试","description":"直接在当前页面用你的 API Key 向网关发请求，方便确认路由、模型名、权限和上游响应是否正常。","openGuide":"查看调用说明","send":"发送测试请求","sending":"请求发送中...","keySelector":"选择 API Key","noSelection":"请选择一个 API Key","noGroupAssigned":"未分配分组","protocol":"调用协议","model":"模型名","loading":"加载中...","noOptionsFound":"没有可选项","stream":"开启流式输出","requestMeta":"请求信息","noKeysTitle":"还没有可用的 API Key","noKeysDescription":"先创建一个 API Key 并分配分组，才能在这里直接发起测试调用。","manageKeys":"管理 API 密钥","modelPlaceholder":"输入模型名","modelSearchPlaceholder":"搜索模型","modelHint":"默认会填入一个常用模型，你也可以手动改成自己的目标模型。","customModel":"自定义模型名","customModelHint":"这里会直接使用你输入的精确模型名发起请求。","customModelOption":"手动输入模型名","customModelOptionHint":"如果下拉里没有你要的模型，可以切换到手动输入。","prompt":"测试提示词","promptHint":"这里会直接作为请求体发送到网关，用来快速验证链路是否通畅。","promptPlaceholder":"输入你想发给模型的内容","streamHint":"开启后，请求会按 SSE 文本返回，原始响应区域会显示完整事件流。","unassignedTitle":"这个 API Key 不能直接测试","unassignedDescription":"因为它还没有分组。未分组 Key 会被网关拒绝，请先回到“API 密钥”页完成分配。","liveBillingTitle":"这里发出的是真实请求","liveBillingDescription":"调用测试不会走 mock，也不会免计费。请求成功后会按正常网关链路记录用量并参与余额、订阅或限额统计。","copyCurl":"复制 curl","platform":"分组平台","headerMode":"鉴权头","requestPreview":"请求体预览","copyRequest":"复制请求体","responsePreview":"响应结果","statusCode":"HTTP 状态","duration":"耗时","copyResponse":"复制响应","responseSummary":"响应摘要","usageRecordTitle":"用量记录同步","openUsage":"查看用量记录","rawResponse":"原始响应","responsePending":"点击“发送测试请求”后，这里会显示网关返回的原始响应和摘要。","notReady":"未就绪","copyCurlSuccess":"curl 命令已复制","copyRequestSuccess":"请求体已复制","copyResponseSuccess":"响应内容已复制","usageRecordSyncing":"请求已成功返回，正在同步对应的用量记录...","usageRecordFound":"已写入用量统计：{time} · ${cost} · {tokens} Tokens","usageRecordPending":"请求已经成功返回，但用量记录采用异步写入。如果你已经打开“用量统计”或仪表盘，请刷新页面后查看。","usageRecordIdle":"测试请求成功后，这里会提示它是否已经进入“用量统计”。","loadKeysFailed":"API Keys 加载失败","unknownError":"未知错误"}},"en":{"labels":{"badge":"Live Request","title":"API Test","description":"Send a real request through the gateway from this page to verify routing, model names, permissions, and upstream responses.","openGuide":"Open API Guide","send":"Send Test Request","sending":"Sending...","keySelector":"Select API Key","noSelection":"Select an API key","noGroupAssigned":"No group assigned","protocol":"Protocol","model":"Model","loading":"Loading...","noOptionsFound":"No options found","stream":"Enable streaming","requestMeta":"Request Details","noKeysTitle":"No API key available yet","noKeysDescription":"Create an API key and assign a group before running live gateway tests here.","manageKeys":"Manage API Keys","modelPlaceholder":"Enter a model name","modelSearchPlaceholder":"Search models","modelHint":"A common default model is prefilled, but you can replace it with the exact model you want to test.","customModel":"Custom Model","customModelHint":"The exact model name entered here will be sent to the gateway as-is.","customModelOption":"Enter model manually","customModelOptionHint":"Switch to manual input when the dropdown does not include the model you want.","prompt":"Prompt","promptHint":"This text is sent directly to the gateway so you can validate the full request path quickly.","promptPlaceholder":"Enter the content you want to send","streamHint":"When enabled, the raw response panel will show the SSE event stream instead of a compact JSON payload.","unassignedTitle":"This API key cannot be tested yet","unassignedDescription":"It has no group assignment. Ungrouped keys are rejected by the gateway until you assign one on the API Keys page.","liveBillingTitle":"Requests here are real","liveBillingDescription":"The API tester does not use mock responses and is not billing-free. Successful requests are recorded through the normal gateway path and count toward balance, subscription, and limit statistics.","copyCurl":"Copy curl","platform":"Group Platform","headerMode":"Auth Header","requestPreview":"Request Preview","copyRequest":"Copy Request Body","responsePreview":"Response","statusCode":"HTTP Status","duration":"Duration","copyResponse":"Copy Response","responseSummary":"Response Summary","usageRecordTitle":"Usage Sync","openUsage":"Open Usage Records","rawResponse":"Raw Response","responsePending":"Run a test request and the raw gateway response will appear here.","notReady":"Not ready","copyCurlSuccess":"curl command copied","copyRequestSuccess":"request body copied","copyResponseSuccess":"response copied","usageRecordSyncing":"The request has completed successfully. Checking for the corresponding usage record...","usageRecordFound":"Recorded in usage statistics: {time} · ${cost} · {tokens} tokens","usageRecordPending":"The request succeeded, but usage records are written asynchronously. If the Usage or Dashboard page is already open, refresh it to see the latest record.","usageRecordIdle":"After a successful test, this panel shows whether the request has appeared in usage statistics.","loadKeysFailed":"Failed to load API keys","unknownError":"Unknown error"}}}`

func apiTestShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAPITestShellConfig
	}
	return value
}

const defaultAvailableGroupsShellConfig = `{"zh":{"labels":{"title":"可用分组","description":"查看当前账号可见的模型分组、倍率、额度和订阅访问要求。","total":"总分组","public":"公开分组","memberOnly":"会员专属","searchPlaceholder":"搜索分组名称、描述、平台或订阅类型","emptyTitle":"没有可用分组","emptyDescription":"当前还没有可展示的分组。","emptyFilteredDescription":"没有匹配当前搜索条件的分组。","publicTitle":"公开分组","publicDescription":"这些分组对当前账号可直接使用。","memberTitle":"会员或专属分组","memberDescription":"这些分组需要订阅、权限或专属配置。","publicBadge":"公开","subscriptionBadge":"订阅","exclusiveBadge":"专属","standardBadge":"标准","imageEnabledBadge":"支持生图","rate":"倍率","quota":"额度","dailyLimit":"每日 ${amount}","weeklyLimit":"每周 ${amount}","monthlyLimit":"每月 ${amount}","unlimited":"不限"}},"en":{"labels":{"title":"Available Groups","description":"Review model groups visible to your account, including rates, quotas, and subscription access requirements.","total":"Total Groups","public":"Public Groups","memberOnly":"Member Only","searchPlaceholder":"Search group name, description, platform, or subscription type","emptyTitle":"No available groups","emptyDescription":"There are no groups to display yet.","emptyFilteredDescription":"No groups match the current search.","publicTitle":"Public Groups","publicDescription":"These groups are directly available to the current account.","memberTitle":"Member or Exclusive Groups","memberDescription":"These groups require a subscription, permission, or exclusive configuration.","publicBadge":"Public","subscriptionBadge":"Subscription","exclusiveBadge":"Exclusive","standardBadge":"Standard","imageEnabledBadge":"Image enabled","rate":"Rate","quota":"Quota","dailyLimit":"Daily ${amount}","weeklyLimit":"Weekly ${amount}","monthlyLimit":"Monthly ${amount}","unlimited":"Unlimited"}}}`

func availableGroupsShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAvailableGroupsShellConfig
	}
	return value
}

const defaultRedeemShellConfig = `{"zh":{"labels":{"currentBalance":"当前余额","concurrency":"并发数","requests":"请求","redeemCodeLabel":"兑换码","redeemCodePlaceholder":"请输入兑换码","redeemCodeHint":"兑换码支持大写字母和数字，可直接粘贴输入","redeemButton":"兑换","redeeming":"兑换中...","redeemSuccess":"兑换成功！","redeemFailed":"兑换失败","added":"已添加","concurrentRequests":"并发请求","subscriptionAssigned":"订阅已分配","subscriptionDays":"{days} 天","newBalance":"新余额","newConcurrency":"新并发数","aboutCodes":"关于兑换码","codeRule1":"每个兑换码只能使用一次","codeRule2":"兑换码可以增加余额、并发数或试用权限","codeRule3":"如有兑换问题，请联系客服","codeRule4":"余额和并发数即时更新","recentActivity":"最近活动","historyWillAppear":"您的兑换历史将显示在这里","adminAdjustment":"管理员调整","balanceAddedRedeem":"余额充值（兑换）","balanceAddedAffiliate":"余额充值（返利转入）","balanceAddedAdmin":"余额充值（管理员）","balanceDeductedAdmin":"余额扣除（管理员）","concurrencyAddedRedeem":"并发增加（兑换）","concurrencyAddedAdmin":"并发增加（管理员）","concurrencyReducedAdmin":"并发减少（管理员）","days":"天","pleaseEnterCode":"请输入兑换码","subscriptionRefreshFailed":"兑换成功，但订阅状态刷新失败。","codeRedeemSuccess":"兑换成功！","failedToRedeem":"兑换失败，请检查兑换码后重试。","unknown":"未知"}},"en":{"labels":{"currentBalance":"Current Balance","concurrency":"Concurrency","requests":"requests","redeemCodeLabel":"Redeem Code","redeemCodePlaceholder":"Enter your redeem code","redeemCodeHint":"Redeem codes use uppercase letters and numbers and can be pasted directly","redeemButton":"Redeem Code","redeeming":"Redeeming...","redeemSuccess":"Code Redeemed Successfully!","redeemFailed":"Redemption Failed","added":"Added","concurrentRequests":"concurrent requests","subscriptionAssigned":"Subscription Assigned","subscriptionDays":"{days} days","newBalance":"New Balance","newConcurrency":"New Concurrency","aboutCodes":"About Redeem Codes","codeRule1":"Each code can only be used once","codeRule2":"Codes may add balance, increase concurrency, or grant trial access","codeRule3":"Contact support if you have issues redeeming a code","codeRule4":"Balance and concurrency updates are immediate","recentActivity":"Recent Activity","historyWillAppear":"Your redemption history will appear here","adminAdjustment":"Admin Adjustment","balanceAddedRedeem":"Balance Added (Redeem)","balanceAddedAffiliate":"Balance Added (Affiliate Transfer)","balanceAddedAdmin":"Balance Added (Admin)","balanceDeductedAdmin":"Balance Deducted (Admin)","concurrencyAddedRedeem":"Concurrency Added (Redeem)","concurrencyAddedAdmin":"Concurrency Added (Admin)","concurrencyReducedAdmin":"Concurrency Reduced (Admin)","days":" days","pleaseEnterCode":"Please enter a redeem code","subscriptionRefreshFailed":"Redeemed successfully, but failed to refresh subscription status.","codeRedeemSuccess":"Code redeemed successfully!","failedToRedeem":"Failed to redeem code. Please check the code and try again.","unknown":"Unknown"}}}`

func redeemShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultRedeemShellConfig
	}
	return value
}

const defaultAffiliateShellConfig = `{"zh":{"labels":{"rebateRate":"我的返利比例","rebateRateHint":"被邀请用户每次充值后你可获得的返利比例","invitedUsers":"邀请人数","availableQuota":"可转返利额度","totalQuota":"历史返利额度","frozenQuota":"冻结中","title":"邀请中心","description":"统一管理邀请码、邀请记录、返利流水与返利转余额。","yourCode":"我的邀请码","copyCode":"复制邀请码","inviteLink":"邀请链接","copyLink":"复制链接","tipsTitle":"使用说明","tipShare":"将邀请码或邀请链接分享给新用户。","tipRebate":"被邀请用户充值后，你可获得 {rate} 的返利额度。","tipTransfer":"返利额度可随时转入账户余额。","tipFreeze":"新产生的返利需要经过冻结期后才能提现。","transferTitle":"返利额度转余额","transferDescription":"将当前可用返利额度一键转入账户余额","transferButton":"转入余额","transferring":"转入中...","transferEmpty":"当前没有可转入额度","transferSuccess":"已转入余额：{amount}","inviteesTitle":"已邀请用户","inviteesEmpty":"暂无邀请记录","emailColumn":"邮箱","usernameColumn":"用户名","rebateColumn":"返利明细","joinedAtColumn":"注册时间","rebatesTitle":"返利记录","rebatesEmpty":"暂无返利记录","inviteeColumn":"被邀请用户","orderAmountColumn":"充值金额","payAmountColumn":"支付金额","rebateAmountColumn":"返利金额","paymentTypeColumn":"支付方式","orderStatusColumn":"订单状态","createdAtColumn":"返利时间","transfersTitle":"转余额记录","transfersEmpty":"暂无转余额记录","amountColumn":"转入金额","balanceAfterColumn":"转入后余额","availableQuotaAfterColumn":"转入后可提返利","frozenQuotaAfterColumn":"转入后冻结返利","historyQuotaAfterColumn":"转入后历史返利","transferredAtColumn":"转入时间","codeCopied":"邀请码已复制","linkCopied":"邀请链接已复制","loadFailed":"加载邀请返利数据失败","transferFailed":"转入余额失败"}},"en":{"labels":{"rebateRate":"My Rebate Rate","rebateRateHint":"What you earn each time an invitee recharges","invitedUsers":"Invited Users","availableQuota":"Available Rebate Quota","totalQuota":"Historical Rebate Quota","frozenQuota":"Frozen","title":"Invite Center","description":"Manage invite codes, invitees, rebate records, and rebate transfers in one place.","yourCode":"Your Affiliate Code","copyCode":"Copy Code","inviteLink":"Invite Link","copyLink":"Copy Link","tipsTitle":"How It Works","tipShare":"Share your affiliate code or invite link with new users.","tipRebate":"When invitees recharge, you receive {rate} of the recharge as rebate quota.","tipTransfer":"Transfer rebate quota to balance at any time.","tipFreeze":"Newly earned rebates may have a waiting period before they can be transferred.","transferTitle":"Transfer Rebate Quota","transferDescription":"Move available rebate quota into your account balance","transferButton":"Transfer to Balance","transferring":"Transferring...","transferEmpty":"No available rebate quota","transferSuccess":"{amount} has been transferred to your balance","inviteesTitle":"Invited Users","inviteesEmpty":"No invited users yet","emailColumn":"Email","usernameColumn":"Username","rebateColumn":"Rebate","joinedAtColumn":"Joined At","rebatesTitle":"Rebate Records","rebatesEmpty":"No rebate records yet","inviteeColumn":"Invitee","orderAmountColumn":"Top-up Amount","payAmountColumn":"Paid Amount","rebateAmountColumn":"Rebate Amount","paymentTypeColumn":"Payment Method","orderStatusColumn":"Order Status","createdAtColumn":"Rebated At","transfersTitle":"Transfer Records","transfersEmpty":"No transfer records yet","amountColumn":"Transferred","balanceAfterColumn":"Balance After","availableQuotaAfterColumn":"Available Quota After","frozenQuotaAfterColumn":"Frozen Quota After","historyQuotaAfterColumn":"Historical Rebate After","transferredAtColumn":"Transferred At","codeCopied":"Affiliate code copied","linkCopied":"Invite link copied","loadFailed":"Failed to load affiliate data","transferFailed":"Failed to transfer affiliate quota"}}}`

func affiliateShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAffiliateShellConfig
	}
	return value
}

const defaultAvailableChannelsShellConfig = `{"zh":{"labels":{"searchPlaceholder":"搜索渠道或模型...","refreshTitle":"刷新","noPricing":"未配置定价","noModels":"未配置模型","empty":"暂无可用渠道","loadError":"加载可用渠道失败","exclusive":"专属","exclusiveTooltip":"管理员授权给你的专属分组","public":"公开","publicTooltip":"对所有用户公开的分组","columns":{"name":"渠道名","description":"描述","platform":"平台","groups":"我可访问的分组","supportedModels":"支持模型"},"pricing":{"billingMode":"计费模式","billingModeImage":"按图片","billingModePerRequest":"按次","billingModeToken":"按 Token","cacheReadPrice":"缓存读取","cacheWritePrice":"缓存写入","imageOutputPrice":"图片输出","inputPrice":"输入","intervals":"阶梯定价","outputPrice":"输出","perRequestPrice":"每次请求","unitPerMillion":"/ 1M token","unitPerRequest":"/ 次"}}},"en":{"labels":{"searchPlaceholder":"Search channels or models...","refreshTitle":"Refresh","noPricing":"Pricing not configured","noModels":"No models configured","empty":"No available channels","loadError":"Failed to load available channels","exclusive":"Exclusive","exclusiveTooltip":"Exclusive groups granted to you by an admin","public":"Public","publicTooltip":"Groups open to all users","columns":{"name":"Channel","description":"Description","platform":"Platform","groups":"Your Accessible Groups","supportedModels":"Supported Models"},"pricing":{"billingMode":"Billing Mode","billingModeImage":"Per Image","billingModePerRequest":"Per Request","billingModeToken":"Per Token","cacheReadPrice":"Cache Read","cacheWritePrice":"Cache Write","imageOutputPrice":"Image Output","inputPrice":"Input","intervals":"Tiered Pricing","outputPrice":"Output","perRequestPrice":"Per Request","unitPerMillion":"/ 1M tokens","unitPerRequest":"/ request"}}}}`

func availableChannelsShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAvailableChannelsShellConfig
	}
	return value
}

const defaultChannelStatusShellConfig = `{"zh":{"labels":{"refreshTitle":"刷新","detailTitle":"渠道详情","loadError":"加载渠道状态失败","detailLoadError":"加载渠道详情失败","latency":"延迟","ping":"端点 Ping","availabilityPrefix":"可用率","extraModelsCount":"+{n} 个模型","emptyTitle":"暂无可显示的渠道","emptyDescription":"管理员尚未配置可监控的渠道。","closeDetail":"关闭","windowTab":{"7d":"7 天","15d":"15 天","30d":"30 天"},"overall":{"operational":"OPERATIONAL","degraded":"DEGRADED"},"detailColumns":{"model":"模型","latestStatus":"最新状态","latestLatency":"最新延迟 (ms)","availability7d":"7 天可用率","availability15d":"15 天可用率","availability30d":"30 天可用率","avgLatency7d":"7 天平均延迟 (ms)"}}},"en":{"labels":{"refreshTitle":"Refresh","detailTitle":"Channel Detail","loadError":"Failed to load channel status","detailLoadError":"Failed to load channel detail","latency":"Latency","ping":"Endpoint Ping","availabilityPrefix":"Availability","extraModelsCount":"+{n} models","emptyTitle":"No channels available","emptyDescription":"No monitored channels have been configured yet.","closeDetail":"Close","windowTab":{"7d":"7 days","15d":"15 days","30d":"30 days"},"overall":{"operational":"OPERATIONAL","degraded":"DEGRADED"},"detailColumns":{"model":"Model","latestStatus":"Latest Status","latestLatency":"Latest Latency (ms)","availability7d":"7d Availability","availability15d":"15d Availability","availability30d":"30d Availability","avgLatency7d":"7d Avg Latency (ms)"}}}}`

func channelStatusShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultChannelStatusShellConfig
	}
	return value
}

const defaultCustomPageShellConfig = `{"zh":{"labels":{"tocTitle":"目录","tocToggle":"目录","notFoundTitle":"页面不存在","notFoundDesc":"该自定义页面不存在或已被删除。","notConfiguredTitle":"页面链接未配置","notConfiguredDesc":"该自定义页面的 URL 未正确配置。","openInNewTab":"新窗口打开","markdownNotFound":"页面不存在","markdownLoadFailed":"页面加载失败","copyCode":"复制","copyCodeSuccess":"已复制 ✓","copyCodeFailed":"失败"}},"en":{"labels":{"tocTitle":"Table of Contents","tocToggle":"Contents","notFoundTitle":"Page not found","notFoundDesc":"This custom page does not exist or has been removed.","notConfiguredTitle":"Page URL not configured","notConfiguredDesc":"The URL for this custom page has not been properly configured.","openInNewTab":"Open in new tab","markdownNotFound":"Page not found","markdownLoadFailed":"Failed to load page","copyCode":"Copy","copyCodeSuccess":"Copied ✓","copyCodeFailed":"Failed"}}}`

func customPageShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultCustomPageShellConfig
	}
	return value
}

const defaultProfileShellConfig = `{"zh":{"labels":{"user":"用户","administrator":"管理员","accountBalance":"账户余额","concurrencyLimit":"并发额度","memberSince":"加入时间","basicsTitle":"基础资料","basicsDescription":"管理头像、昵称以及当前账号展示信息。","linkedProfileSources":"资料来源","linkedProfileSourcesDescription":"部分资料会从绑定的第三方登录方式同步。","contactSupport":"联系客服","changePassword":"修改密码","currentPassword":"当前密码","newPassword":"新密码","confirmNewPassword":"确认新密码","passwordHint":"密码至少需要 {count} 位字符","changingPassword":"修改中...","changePasswordButton":"修改密码","passwordsNotMatch":"两次输入的新密码不一致","passwordTooShort":"密码至少需要 {count} 位字符","passwordChangeSuccess":"密码修改成功","passwordChangeFailed":"密码修改失败","balanceNotifyTitle":"余额不足提醒","balanceNotifyDescription":"当账户余额低于阈值时发送邮件提醒","balanceNotifyEnabled":"启用余额不足提醒","balanceNotifyThreshold":"自定义提醒阈值","balanceNotifyThresholdHint":"留空使用系统默认值","balanceNotifySystemDefault":"系统默认值","balanceNotifyThresholdPlaceholder":"输入金额","balanceNotifyExtraEmails":"通知邮箱","balanceNotifyExtraEmailsHint":"必须添加并验证邮箱后，余额不足时才能收到提醒邮件","balanceNotifyCodePlaceholder":"6位验证码","balanceNotifyVerify":"验证","balanceNotifyResend":"重发","balanceNotifyUnverified":"未验证","balanceNotifyVerified":"已验证","balanceNotifyRemoveEmail":"移除","balanceNotifySendCode":"发送验证码","balanceNotifyEmailPlaceholder":"输入邮箱地址","balanceNotifyMaxEmailsReached":"已达到通知邮箱数量上限","balanceNotifyEmailDuplicate":"该邮箱已存在","balanceNotifyCodeSent":"验证码已发送","balanceNotifyVerifySuccess":"邮箱添加成功","balanceNotifyRemoveSuccess":"邮箱已移除","balanceNotifySaving":"保存中...","balanceNotifySave":"保存","balanceNotifyCancel":"取消","balanceNotifyAdd":"添加","balanceNotifySaved":"已保存","balanceNotifyError":"操作失败","avatarTitle":"资料头像","avatarDescription":"仅支持上传头像图片；静态图片会自动压缩到 20KB 以内后再保存。","avatarUploadHint":"上传图片时会自动压缩静态图片到 20KB 以内，GIF 需自行控制在 20KB 以内","avatarUploadAction":"上传图片","avatarUploadRequired":"请先上传头像图片","avatarReadFailed":"读取所选图片失败","avatarCompressFailed":"压缩所选图片失败","avatarCompressTooLarge":"无法将图片压缩到 20KB 以内，请换一张更小的图片","avatarInvalidType":"请选择图片文件","avatarGifTooLarge":"GIF 头像必须在 20KB 以内","avatarSaveSuccess":"头像已更新","avatarEmptyDeleteHint":"当前没有可删除的头像","avatarDeleteSuccess":"头像已删除","totpTitle":"两步验证","totpDescription":"使用身份验证器应用为账号增加额外保护","totpFeatureDisabled":"两步验证当前未启用","totpFeatureDisabledHint":"管理员尚未开启此功能","totpEnabled":"两步验证已启用","totpEnabledAt":"启用时间","totpDisable":"停用","totpNotEnabled":"两步验证未启用","totpNotEnabledHint":"启用后，登录时需要输入身份验证器中的动态验证码","totpEnable":"启用","providers":{"email":"邮箱","linuxdo":"LinuxDo","dingtalk":"钉钉","oidc":"{providerName}","wechat":"微信","github":"GitHub","google":"Google"},"sourceAvatar":"头像当前来自 {providerName}","sourceUsername":"昵称当前来自 {providerName}"}},"en":{"labels":{"user":"User","administrator":"Administrator","accountBalance":"Account Balance","concurrencyLimit":"Concurrency Limit","memberSince":"Member Since","basicsTitle":"Basic Profile","basicsDescription":"Manage avatar, nickname, and account display information.","linkedProfileSources":"Profile Sources","linkedProfileSourcesDescription":"Some profile fields can be synced from connected sign-in providers.","contactSupport":"Contact Support","changePassword":"Change Password","currentPassword":"Current Password","newPassword":"New Password","confirmNewPassword":"Confirm New Password","passwordHint":"Password must be at least {count} characters long","changingPassword":"Changing...","changePasswordButton":"Change Password","passwordsNotMatch":"New passwords do not match","passwordTooShort":"Password must be at least {count} characters long","passwordChangeSuccess":"Password changed successfully","passwordChangeFailed":"Failed to change password","balanceNotifyTitle":"Balance Low Notification","balanceNotifyDescription":"Send email alert when account balance falls below threshold","balanceNotifyEnabled":"Enable Balance Low Notification","balanceNotifyThreshold":"Custom Threshold","balanceNotifyThresholdHint":"Leave empty to use system default","balanceNotifySystemDefault":"System Default","balanceNotifyThresholdPlaceholder":"Enter amount","balanceNotifyExtraEmails":"Notification Emails","balanceNotifyExtraEmailsHint":"You must add and verify an email address to receive low balance alerts","balanceNotifyCodePlaceholder":"6-digit code","balanceNotifyVerify":"Verify","balanceNotifyResend":"Resend","balanceNotifyUnverified":"Unverified","balanceNotifyVerified":"Verified","balanceNotifyRemoveEmail":"Remove","balanceNotifySendCode":"Send Code","balanceNotifyEmailPlaceholder":"Enter email address","balanceNotifyMaxEmailsReached":"Maximum number of notification emails reached","balanceNotifyEmailDuplicate":"This email already exists","balanceNotifyCodeSent":"Verification code sent","balanceNotifyVerifySuccess":"Email added successfully","balanceNotifyRemoveSuccess":"Email removed","balanceNotifySaving":"Saving...","balanceNotifySave":"Save","balanceNotifyCancel":"Cancel","balanceNotifyAdd":"Add","balanceNotifySaved":"Saved","balanceNotifyError":"Operation failed","avatarTitle":"Profile Avatar","avatarDescription":"Upload an avatar image. Static uploads are compressed to 20KB before saving.","avatarUploadHint":"Static uploads are compressed to 20KB when possible. GIF uploads must already be within 20KB.","avatarUploadAction":"Upload image","avatarUploadRequired":"Upload an avatar image first","avatarReadFailed":"Failed to read the selected image.","avatarCompressFailed":"Failed to compress the selected image.","avatarCompressTooLarge":"Unable to compress this image below 20KB. Try a smaller image.","avatarInvalidType":"Please choose an image file","avatarGifTooLarge":"GIF avatars must already be 20KB or smaller","avatarSaveSuccess":"Avatar updated","avatarEmptyDeleteHint":"Avatar is already empty","avatarDeleteSuccess":"Avatar removed","totpTitle":"Two-Factor Authentication","totpDescription":"Use an authenticator app to add extra protection to your account","totpFeatureDisabled":"Two-factor authentication is unavailable","totpFeatureDisabledHint":"This feature has not been enabled by an administrator","totpEnabled":"Two-factor authentication enabled","totpEnabledAt":"Enabled at","totpDisable":"Disable","totpNotEnabled":"Two-factor authentication is not enabled","totpNotEnabledHint":"After enabling it, sign-in requires a dynamic code from your authenticator app","totpEnable":"Enable","providers":{"email":"Email","linuxdo":"LinuxDo","dingtalk":"DingTalk","oidc":"{providerName}","wechat":"WeChat","github":"GitHub","google":"Google"},"sourceAvatar":"Avatar is currently synced from {providerName}","sourceUsername":"Nickname is currently synced from {providerName}"}}}`

func profileShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return profileShellConfigDefault()
	}
	return value
}

func profileShellConfigDefault() string {
	zhProfileEditLabels := `"profileStatusActive":"启用","profileStatusDisabled":"禁用","profileEditTitle":"编辑个人资料","profileUsername":"用户名","profileUsernamePlaceholder":"请输入用户名","profileUpdating":"更新中...","profileUpdateAction":"更新资料","profileUsernameRequired":"用户名不能为空","profileUpdateSuccess":"资料更新成功","profileUpdateFailed":"资料更新失败",`
	enProfileEditLabels := `"profileStatusActive":"Active","profileStatusDisabled":"Disabled","profileEditTitle":"Edit Profile","profileUsername":"Username","profileUsernamePlaceholder":"Enter username","profileUpdating":"Updating...","profileUpdateAction":"Update Profile","profileUsernameRequired":"Username is required","profileUpdateSuccess":"Profile updated successfully","profileUpdateFailed":"Failed to update profile",`
	zhAvatarActionLabels := `"avatarSave":"保存","avatarDelete":"删除","avatarError":"操作失败",`
	enAvatarActionLabels := `"avatarSave":"Save","avatarDelete":"Delete","avatarError":"Operation failed",`
	zhAuthBindingLabels := `"totpSetupTitle":"设置双因素认证","totpVerifyEmailFirst":"请先验证您的邮箱","totpVerifyPasswordFirst":"请先验证您的身份","totpSetupStep1":"使用认证器应用扫描下方二维码","totpSetupStep2":"输入应用显示的 6 位验证码","totpEmailCode":"邮箱验证码","totpEnterEmailCode":"请输入 6 位验证码","totpSendCode":"发送验证码","totpSending":"发送中...","totpEnterPassword":"请输入当前密码确认","totpManualEntry":"无法扫码？手动输入密钥：","totpEnterCode":"输入 6 位验证码","totpVerify":"验证","totpCancel":"取消","totpNext":"下一步","totpBack":"返回","totpLoading":"加载中...","totpVerifying":"验证中...","totpCopied":"已复制","totpCopyFailed":"复制失败","totpCodeSent":"验证码已发送到您的邮箱","totpSendCodeFailed":"发送验证码失败","totpSetupFailed":"获取设置信息失败","totpEnableSuccess":"双因素认证已启用","totpVerifyFailed":"验证码错误，请重试","totpDisableTitle":"禁用双因素认证","totpDisableWarning":"禁用后，登录时将不再需要验证码。这可能会降低您的账户安全性。","totpConfirmDisable":"确认禁用","totpProcessing":"处理中...","totpDisableSuccess":"双因素认证已禁用","totpDisableFailed":"禁用失败，请检查密码是否正确","totpError":"操作失败","authBindingsTitle":"登录方式绑定","authBindingsDescription":"查看当前绑定状态，并将更多第三方登录方式关联到这个账号。","authBindingsStatusBound":"已绑定","authBindingsStatusNotBound":"未绑定","authBindingsStatusPasswordNotSet":"未设置密码","authBindingsBindAction":"绑定 {providerName}","authBindingsEmailPlaceholder":"输入邮箱地址","authBindingsCodePlaceholder":"输入验证码","authBindingsPasswordPlaceholder":"设置登录密码","authBindingsReplaceEmailPasswordPlaceholder":"输入当前密码","authBindingsSendCodeAction":"发送验证码","authBindingsUnbindAction":"解绑","authBindingsManageEmailAction":"管理邮箱","authBindingsHideEmailFormAction":"收起邮箱表单","authBindingsConfirmEmailBindAction":"绑定邮箱","authBindingsConfirmEmailReplaceAction":"更换主邮箱","authBindingsBoundCount":"已关联 {count} 条记录","authBindingsUnbindSuccess":"{providerName} 已解绑","authBindingsCodeSentTo":"验证码已发送到 {email}","authBindingsBindSuccess":"账号绑定成功","authBindingsReplaceSuccess":"主邮箱已更新","authBindingsLoading":"加载中...","authBindingsTryAgain":"请稍后重试","authBindingsEmailRequired":"请输入邮箱","authBindingsInvalidEmail":"请输入有效的邮箱地址","authBindingsCodeRequired":"请输入验证码","authBindingsPasswordRequired":"请输入密码","authBindingsPasswordMinLength":"密码至少需要 {count} 位字符","authBindingsSendCodeFailed":"发送验证码失败","authBindingsNoteEmailManagedFromProfile":"主邮箱在资料表单中管理","authBindingsNoteCanUnbind":"你可以解绑这个登录方式。","authBindingsNoteBindAnotherBeforeUnbind":"请先绑定其他登录方式，再解除当前绑定。",`
	enAuthBindingLabels := `"totpSetupTitle":"Set Up Two-Factor Authentication","totpVerifyEmailFirst":"Please verify your email first","totpVerifyPasswordFirst":"Please verify your identity first","totpSetupStep1":"Scan the QR code below with your authenticator app","totpSetupStep2":"Enter the 6-digit code from your app","totpEmailCode":"Email Verification Code","totpEnterEmailCode":"Enter 6-digit code","totpSendCode":"Send Code","totpSending":"Sending...","totpEnterPassword":"Enter your current password to confirm","totpManualEntry":"Can't scan? Enter the key manually:","totpEnterCode":"Enter 6-digit code","totpVerify":"Verify","totpCancel":"Cancel","totpNext":"Next","totpBack":"Back","totpLoading":"Loading...","totpVerifying":"Verifying...","totpCopied":"Copied","totpCopyFailed":"Copy failed","totpCodeSent":"Verification code sent to your email","totpSendCodeFailed":"Failed to send verification code","totpSetupFailed":"Failed to get setup information","totpEnableSuccess":"Two-factor authentication enabled","totpVerifyFailed":"Invalid code, please try again","totpDisableTitle":"Disable Two-Factor Authentication","totpDisableWarning":"After disabling, you will no longer need a verification code to log in. This may reduce your account security.","totpConfirmDisable":"Confirm Disable","totpProcessing":"Processing...","totpDisableSuccess":"Two-factor authentication disabled","totpDisableFailed":"Failed to disable, please check your password","totpError":"Operation failed","authBindingsTitle":"Connected Sign-In Methods","authBindingsDescription":"View current bindings and connect another provider to this account.","authBindingsStatusBound":"Bound","authBindingsStatusNotBound":"Not bound","authBindingsStatusPasswordNotSet":"Password not set","authBindingsBindAction":"Bind {providerName}","authBindingsEmailPlaceholder":"Enter email address","authBindingsCodePlaceholder":"Enter verification code","authBindingsPasswordPlaceholder":"Set a login password","authBindingsReplaceEmailPasswordPlaceholder":"Enter current password","authBindingsSendCodeAction":"Send code","authBindingsUnbindAction":"Unbind","authBindingsManageEmailAction":"Manage email","authBindingsHideEmailFormAction":"Hide email form","authBindingsConfirmEmailBindAction":"Bind email","authBindingsConfirmEmailReplaceAction":"Replace primary email","authBindingsBoundCount":"{count} linked records","authBindingsUnbindSuccess":"{providerName} unbound","authBindingsCodeSentTo":"Code sent to {email}","authBindingsBindSuccess":"Account linked successfully","authBindingsReplaceSuccess":"Primary email updated","authBindingsLoading":"Loading...","authBindingsTryAgain":"Please try again later","authBindingsEmailRequired":"Email is required","authBindingsInvalidEmail":"Enter a valid email address","authBindingsCodeRequired":"Verification code is required","authBindingsPasswordRequired":"Password is required","authBindingsPasswordMinLength":"Password must be at least {count} characters long","authBindingsSendCodeFailed":"Failed to send verification code","authBindingsNoteEmailManagedFromProfile":"Primary email is managed in the profile form","authBindingsNoteCanUnbind":"You can unbind this sign-in method","authBindingsNoteBindAnotherBeforeUnbind":"Bind another sign-in method before unbinding",`
	value := strings.Replace(defaultProfileShellConfig, `"linkedProfileSourcesDescription":"部分资料会从绑定的第三方登录方式同步。","contactSupport"`, `"linkedProfileSourcesDescription":"部分资料会从绑定的第三方登录方式同步。",`+zhProfileEditLabels+`"contactSupport"`, 1)
	value = strings.Replace(value, `"linkedProfileSourcesDescription":"Some profile fields can be synced from connected sign-in providers.","contactSupport"`, `"linkedProfileSourcesDescription":"Some profile fields can be synced from connected sign-in providers.",`+enProfileEditLabels+`"contactSupport"`, 1)
	value = strings.Replace(value, `"avatarDeleteSuccess":"头像已删除","totpTitle"`, `"avatarDeleteSuccess":"头像已删除",`+zhAvatarActionLabels+`"totpTitle"`, 1)
	value = strings.Replace(value, `"avatarDeleteSuccess":"Avatar removed","totpTitle"`, `"avatarDeleteSuccess":"Avatar removed",`+enAvatarActionLabels+`"totpTitle"`, 1)
	value = strings.Replace(value, `"totpEnable":"启用","providers"`, `"totpEnable":"启用",`+zhAuthBindingLabels+`"providers"`, 1)
	value = strings.Replace(value, `"totpEnable":"Enable","providers"`, `"totpEnable":"Enable",`+enAuthBindingLabels+`"providers"`, 1)
	return value
}

const defaultAuthShellConfig = `{
  "zh": {
    "labels": {
      "welcomeBack": "欢迎回来",
      "signInToAccount": "登录您的账户",
      "emailLabel": "邮箱",
      "emailPlaceholder": "请输入邮箱",
      "emailVerifyBackToRegistration": "返回注册",
      "emailVerifyClickToResend": "点击重新发送验证码",
      "emailVerifyCodeHint": "请输入发送到您邮箱的6位验证码",
      "emailVerifyCodeLabel": "验证码",
      "emailVerifyCodeSentSuccess": "验证码已发送！请查收您的邮箱。",
      "emailVerifyDescriptionPrefix": "我们将发送验证码到",
      "emailVerifyResendCode": "重新发送验证码",
      "emailVerifyResendCountdown": "{countdown}秒后可重新发送",
      "emailVerifySessionExpiredDescription": "请返回注册页面重新开始。",
      "emailVerifySessionExpiredTitle": "会话已过期",
      "emailVerifySubmit": "验证并创建账户",
      "emailVerifyTitle": "验证您的邮箱",
      "emailVerifyVerifying": "验证中...",
      "passwordLabel": "密码",
      "passwordPlaceholder": "请输入密码",
      "forgotPassword": "忘记密码？",
      "forgotPasswordTitle": "重置密码",
      "forgotPasswordHint": "输入您的邮箱地址，我们将向您发送密码重置链接。",
      "resetPasswordTitle": "设置新密码",
      "resetPasswordHint": "请在下方输入您的新密码。",
      "resetPassword": "重置密码",
      "resettingPassword": "重置中...",
      "invalidResetLink": "无效的重置链接",
      "invalidResetLinkHint": "此密码重置链接无效或已过期。请重新请求一个新链接。",
      "requestNewResetLink": "请求新的重置链接",
      "signingIn": "登录中...",
      "signIn": "登录",
      "signInWithProvider": "使用 {providerName} 登录",
      "oauthAlternativeMethods": "或使用以下方式继续",
      "oauthCallbackCode": "授权码",
      "oauthCallbackHint": "如果页面未自动跳转，请返回登录页重试。",
      "oauthCallbackFullUrl": "完整URL",
      "oauthCallbackInvalidHint": "当前页面缺少有效的授权结果，请返回登录页重新发起快捷登录。",
      "oauthCallbackInvalidTitle": "无效的登录回调",
      "oauthCallbackPasswordOptionalHint": "{providerName} 登录可留空，稍后可在个人资料中补充设置密码。",
      "oauthCallbackRegistrationHint": "完成注册",
      "oauthCallbackRegistrationInvitationRequired": "该 {providerName} 账号尚未注册，站点已开启邀请码注册，请输入邀请码以完成注册。",
      "oauthCallbackState": "状态",
      "oauthCallbackSubmitRegistration": "完成注册",
      "oauthCallbackTitle": "OAuth 回调",
      "oauthFlowAvatarAlt": "{providerName} 头像",
      "oauthFlowBindCurrentAccount": "绑定当前账户",
      "oauthFlowBindCurrentAccountDescription": "将此次 {providerName} 登录绑定到当前浏览器已登录的账户。",
      "oauthFlowBindCurrentAccountTitle": "绑定当前账户",
      "oauthFlowBindExistingAccount": "绑定已有账户",
      "oauthFlowBindLoginHint": "登录一个已有账户以绑定此次 {providerName} 登录。",
      "oauthFlowBindSignInToExistingAccount": "将此次 {providerName} 登录绑定到已有账户。",
      "oauthFlowChooseAccountActionHint": "请选择绑定已有账户，或创建一个新账户。",
      "oauthFlowChooseHowToContinue": "选择后续操作",
      "oauthFlowCreateAccountHint": "请输入邮箱地址以创建账户并继续。",
      "oauthFlowCreateAccountTitle": "完成 {providerName} 账户注册",
      "oauthFlowCreateNewAccount": "创建新账户",
      "oauthFlowLogInAndBind": "登录并绑定",
      "oauthFlowProfileDetailsDescription": "选择是否将 {providerName} 的昵称或头像应用到当前账户。",
      "oauthFlowProfileDetailsTitle": "使用 {providerName} 资料",
      "oauthFlowReviewProfileBeforeContinue": "请先确认 {providerName} 资料后再继续。",
      "oauthFlowSignInThenBindDescription": "请先登录已有账户，再将此次 {providerName} 登录绑定到该账户。",
      "oauthFlowSuggestedEmail": "建议邮箱：{email}",
      "oauthFlowTotpHint": "请输入 {account} 的 6 位验证码，以完成此次 {providerName} 登录绑定。",
      "oauthFlowUseAvatar": "使用头像",
      "oauthFlowUseDifferentEmail": "使用其他邮箱",
      "oauthFlowUseDisplayName": "使用昵称",
      "oauthFlowVerifyAndContinue": "验证并继续",
      "oauthFlowYourAccount": "当前账户",
      "wechatAvailabilityUnknown": "暂时无法确认微信登录可用性，请刷新后重试。",
      "wechatBrowserOnly": "当前微信登录流程仅支持在微信内置浏览器中继续。",
      "wechatNativeAppOnly": "当前仅配置微信移动应用登录，需要在原生 App 中通过微信 SDK 发起授权。",
      "wechatNotConfigured": "微信登录尚未配置。",
      "wechatProviderName": "微信",
      "wechatSystemBrowserOnly": "当前微信登录流程仅支持在系统浏览器中继续。",
      "dontHaveAccount": "还没有账户？",
      "signUp": "注册",
      "createAccount": "创建账户",
      "signUpToStart": "注册以开始使用 {siteName}",
      "registrationDisabled": "注册功能暂时关闭，请联系管理员。",
      "createPasswordPlaceholder": "创建一个安全的密码",
      "newPassword": "新密码",
      "newPasswordPlaceholder": "输入新密码",
      "confirmPassword": "确认密码",
      "confirmPasswordPlaceholder": "再次输入新密码",
      "dingtalkProviderName": "钉钉",
      "passwordHint": "至少 {count} 个字符",
      "invitationCodeLabel": "邀请码",
      "invitationCodePlaceholder": "请输入邀请码",
      "invitationCodeValid": "邀请码有效",
      "affiliateInvitationDetected": "已检测到邀请链接，可直接继续注册。",
      "sendCode": "发送验证码",
      "sendingCode": "发送中...",
      "sendResetLink": "发送重置链接",
      "sendingResetLink": "发送中...",
      "resendCountdown": "{countdown}s",
      "codeSentSuccess": "验证码已发送",
      "verificationCodeHint": "请输入邮箱收到的 6 位验证码",
      "promoCodeLabel": "优惠码",
      "optional": "可选",
      "promoCodePlaceholder": "请输入优惠码",
      "promoCodeValid": "优惠码有效，注册后将获得 {amount} 余额",
      "processing": "处理中...",
      "continue": "继续",
      "totpLoginTitle": "两步验证",
      "totpLoginHint": "请输入身份验证器应用中的 6 位验证码",
      "totpVerifying": "验证中...",
      "totpCancel": "取消",
      "resetEmailSent": "重置链接已发送",
      "resetEmailSentHint": "如果该邮箱已注册，您将很快收到密码重置链接。请检查您的收件箱和垃圾邮件文件夹。",
      "passwordResetSuccess": "密码重置成功",
      "passwordResetSuccessHint": "您的密码已重置。现在可以使用新密码登录。",
      "providerCallbackHint": "如果页面未自动跳转，请返回登录页重试。",
      "providerCallbackProcessing": "正在验证 {providerName} 登录信息，请稍候...",
      "providerCallbackTitle": "正在完成 {providerName} 登录",
      "providerCompleteRegistration": "完成注册",
      "providerCompletingRegistration": "正在完成注册...",
      "providerInvitationRequired": "该 {providerName} 账号尚未注册，站点已开启邀请码注册，请输入邀请码以完成注册。",
      "agreementAcceptAndContinue": "同意并继续",
      "agreementAcceptedDescription": "您已同意当前版本条款，可随时重新查看相关文档。",
      "agreementAcceptedTitle": "登录条款入口",
      "agreementCheckboxPrefix": "我已阅读并同意",
      "agreementRecent": "近期",
      "agreementReject": "拒绝",
      "agreementRelevantDocuments": "相关文档",
      "agreementReviewDescription": "您可以先输入账号信息；如果当前账号尚未确认最新条款，我们会在提交登录时提示确认。",
      "agreementReviewTitle": "继续登录前可能需要确认最新条款。",
      "agreementTermsUpdateTitle": "条款更新通知",
      "agreementUpdatedAt": "我们的服务条款已于 {date} 更新。在继续使用服务之前，请仔细阅读并同意以下条款。",
      "agreementViewAndAccept": "查看并同意",
      "agreementViewTerms": "查看条款",
      "alreadyHaveAccount": "已有账户？",
      "rememberedPassword": "想起密码了？",
      "backToLogin": "返回登录",
      "allRightsReserved": "保留所有权利。"
    }
  },
  "en": {
    "labels": {
      "welcomeBack": "Welcome Back",
      "signInToAccount": "Sign in to your account",
      "emailLabel": "Email",
      "emailPlaceholder": "Enter your email",
      "emailVerifyBackToRegistration": "Back to registration",
      "emailVerifyClickToResend": "Click to resend code",
      "emailVerifyCodeHint": "Enter the 6-digit code sent to your email",
      "emailVerifyCodeLabel": "Verification Code",
      "emailVerifyCodeSentSuccess": "Verification code sent! Please check your inbox.",
      "emailVerifyDescriptionPrefix": "We'll send a verification code to",
      "emailVerifyResendCode": "Resend verification code",
      "emailVerifyResendCountdown": "Resend code in {countdown}s",
      "emailVerifySessionExpiredDescription": "Please go back to the registration page and start again.",
      "emailVerifySessionExpiredTitle": "Session expired",
      "emailVerifySubmit": "Verify & Create Account",
      "emailVerifyTitle": "Verify Your Email",
      "emailVerifyVerifying": "Verifying...",
      "passwordLabel": "Password",
      "passwordPlaceholder": "Enter your password",
      "forgotPassword": "Forgot password?",
      "forgotPasswordTitle": "Reset Your Password",
      "forgotPasswordHint": "Enter your email address and we will send you a link to reset your password.",
      "resetPasswordTitle": "Set New Password",
      "resetPasswordHint": "Enter your new password below.",
      "resetPassword": "Reset Password",
      "resettingPassword": "Resetting...",
      "invalidResetLink": "Invalid Reset Link",
      "invalidResetLinkHint": "This password reset link is invalid or has expired. Please request a new one.",
      "requestNewResetLink": "Request New Reset Link",
      "signingIn": "Signing in...",
      "signIn": "Sign In",
      "signInWithProvider": "Sign in with {providerName}",
      "oauthAlternativeMethods": "or continue with one of the following methods",
      "oauthCallbackCode": "Code",
      "oauthCallbackHint": "If you are not redirected automatically, go back to the login page and try again.",
      "oauthCallbackFullUrl": "Full URL",
      "oauthCallbackInvalidHint": "This page does not contain a valid authorization result. Return to the login page and start quick sign-in again.",
      "oauthCallbackInvalidTitle": "Invalid sign-in callback",
      "oauthCallbackPasswordOptionalHint": "You can leave this blank for {providerName} sign-in and set a password later from your profile.",
      "oauthCallbackRegistrationHint": "Complete Registration",
      "oauthCallbackRegistrationInvitationRequired": "This {providerName} account is not yet registered. The site requires an invitation code — please enter one to complete registration.",
      "oauthCallbackState": "State",
      "oauthCallbackSubmitRegistration": "Complete Registration",
      "oauthCallbackTitle": "OAuth Callback",
      "oauthFlowAvatarAlt": "{providerName} avatar",
      "oauthFlowBindCurrentAccount": "Bind current account",
      "oauthFlowBindCurrentAccountDescription": "Bind this {providerName} sign-in to the account currently signed in on this browser.",
      "oauthFlowBindCurrentAccountTitle": "Bind the current account",
      "oauthFlowBindExistingAccount": "Bind existing account",
      "oauthFlowBindLoginHint": "Log in to an existing account to bind this {providerName} sign-in.",
      "oauthFlowBindSignInToExistingAccount": "Bind this {providerName} sign-in to an existing account.",
      "oauthFlowChooseAccountActionHint": "Choose whether to bind an existing account or create a new one.",
      "oauthFlowChooseHowToContinue": "Choose how to continue",
      "oauthFlowCreateAccountHint": "Enter an email address to create your account and continue.",
      "oauthFlowCreateAccountTitle": "Complete your {providerName} account setup",
      "oauthFlowCreateNewAccount": "Create new account",
      "oauthFlowLogInAndBind": "Log in and bind",
      "oauthFlowProfileDetailsDescription": "Choose whether to apply the nickname or avatar from {providerName} to this account.",
      "oauthFlowProfileDetailsTitle": "Use {providerName} profile details",
      "oauthFlowReviewProfileBeforeContinue": "Review the {providerName} profile details before continuing.",
      "oauthFlowSignInThenBindDescription": "Sign in to an existing account, then bind this {providerName} sign-in to it.",
      "oauthFlowSuggestedEmail": "Suggested email: {email}",
      "oauthFlowTotpHint": "Enter the 6-digit verification code for {account} to finish binding this {providerName} sign-in.",
      "oauthFlowUseAvatar": "Use avatar",
      "oauthFlowUseDifferentEmail": "Use a different email",
      "oauthFlowUseDisplayName": "Use display name",
      "oauthFlowVerifyAndContinue": "Verify and continue",
      "oauthFlowYourAccount": "your account",
      "wechatAvailabilityUnknown": "WeChat sign-in availability could not be confirmed. Refresh and retry.",
      "wechatBrowserOnly": "This WeChat sign-in flow is only available inside the WeChat browser.",
      "wechatNativeAppOnly": "This site only has WeChat mobile app login configured. Continue from the native app through the WeChat SDK.",
      "wechatNotConfigured": "WeChat sign-in is not configured yet.",
      "wechatProviderName": "WeChat",
      "wechatSystemBrowserOnly": "This WeChat sign-in flow is only available in your system browser.",
      "dontHaveAccount": "Don't have an account?",
      "signUp": "Sign Up",
      "createAccount": "Create Account",
      "signUpToStart": "Sign up to start using {siteName}",
      "registrationDisabled": "Registration is currently disabled. Please contact the administrator.",
      "createPasswordPlaceholder": "Create a strong password",
      "newPassword": "New Password",
      "newPasswordPlaceholder": "Enter your new password",
      "confirmPassword": "Confirm Password",
      "confirmPasswordPlaceholder": "Confirm your new password",
      "dingtalkProviderName": "DingTalk",
      "passwordHint": "At least {count} characters",
      "invitationCodeLabel": "Invitation Code",
      "invitationCodePlaceholder": "Enter invitation code",
      "invitationCodeValid": "Invitation code is valid",
      "affiliateInvitationDetected": "Invitation link detected. You can continue registration directly.",
      "sendCode": "Send Code",
      "sendingCode": "Sending...",
      "sendResetLink": "Send Reset Link",
      "sendingResetLink": "Sending...",
      "resendCountdown": "Resend in {countdown}s",
      "codeSentSuccess": "Code sent successfully",
      "verificationCodeHint": "Enter the 6-digit code sent to your email",
      "promoCodeLabel": "Promo Code",
      "optional": "Optional",
      "promoCodePlaceholder": "Enter promo code",
      "promoCodeValid": "Promo code valid. You will receive {amount} balance after registration",
      "processing": "Processing...",
      "continue": "Continue",
      "totpLoginTitle": "Two-Factor Authentication",
      "totpLoginHint": "Enter the 6-digit code from your authenticator app",
      "totpVerifying": "Verifying...",
      "totpCancel": "Cancel",
      "resetEmailSent": "Reset Link Sent",
      "resetEmailSentHint": "If an account exists with this email, you will receive a password reset link shortly. Please check your inbox and spam folder.",
      "passwordResetSuccess": "Password Reset Successful",
      "passwordResetSuccessHint": "Your password has been reset. You can now sign in with your new password.",
      "providerCallbackHint": "If you are not redirected automatically, go back to the login page and try again.",
      "providerCallbackProcessing": "Completing login with {providerName}, please wait...",
      "providerCallbackTitle": "Signing you in with {providerName}",
      "providerCompleteRegistration": "Complete Registration",
      "providerCompletingRegistration": "Completing registration…",
      "providerInvitationRequired": "This {providerName} account is not yet registered. The site requires an invitation code — please enter one to complete registration.",
      "agreementAcceptAndContinue": "Accept and continue",
      "agreementAcceptedDescription": "You have already accepted the current agreement version and can review the documents again at any time.",
      "agreementAcceptedTitle": "Agreement entry",
      "agreementCheckboxPrefix": "I have read and agree to",
      "agreementRecent": "recently",
      "agreementReject": "Reject",
      "agreementRelevantDocuments": "Related documents",
      "agreementReviewDescription": "You can enter your account details first. If this account has not accepted the latest agreement yet, we will prompt for confirmation when you submit sign-in.",
      "agreementReviewTitle": "You may need to confirm the latest agreement before signing in.",
      "agreementTermsUpdateTitle": "Terms update notice",
      "agreementUpdatedAt": "Our terms of service were updated on {date}. Please review and accept the following documents before continuing.",
      "agreementViewAndAccept": "View and accept",
      "agreementViewTerms": "View terms",
      "alreadyHaveAccount": "Already have an account?",
      "rememberedPassword": "Remembered your password?",
      "backToLogin": "Back to Login",
      "allRightsReserved": "All rights reserved."
    }
  }
}`

func authShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAuthShellConfig
	}
	return value
}

const defaultAPIKeysShellConfig = `{"zh":{"labels":{"actions":"操作","active":"启用","allGroups":"所有分组","allStatus":"所有状态","apiKey":"API Key","cancel":"取消","ccsClientSelectClaudeCode":"Claude Code","ccsClientSelectClaudeCodeDesc":"导入到 Claude Code","ccsClientSelectDescription":"选择要导入的客户端。","ccsClientSelectGeminiCli":"Gemini CLI","ccsClientSelectGeminiCliDesc":"导入到 Gemini CLI","ccsClientSelectTitle":"选择 CCS 客户端","ccSwitchNotInstalled":"未检测到 CC Switch 客户端","clickToChangeGroup":"点击切换分组","copyToClipboard":"复制到剪贴板","copied":"已复制","create":"创建","createFirstKey":"创建第一个 API Key 后即可开始调用。","created":"创建时间","createKey":"创建 Key","currentExpiration":"当前过期时间","customDate":"自定义日期","customKeyHint":"留空则自动生成安全 API Key。","customKeyInvalidChars":"自定义 Key 只能包含字母、数字、下划线和连字符","customKeyLabel":"自定义 Key","customKeyPlaceholder":"输入自定义 API Key","customKeyRequired":"请输入自定义 Key","customKeyTooShort":"自定义 Key 至少需要 16 个字符","delete":"删除","deleteConfirmMessage":"确定要删除 {name} 吗？此操作不可恢复。","deleteKey":"删除 API Key","disable":"禁用","edit":"编辑","editKey":"编辑 Key","enable":"启用","expiration":"过期时间","expirationDate":"过期日期","expirationDateHint":"到达该时间后，API Key 将自动失效。","expiresInDays":"{days} 天后过期","expiresAt":"过期时间","extendDays":"延长 {days} 天","failedToChangeGroup":"切换分组失败","failedToDelete":"删除失败","failedToLoad":"加载 API Keys 失败","failedToResetQuota":"重置额度用量失败","failedToResetRateLimit":"重置频率限制用量失败","failedToSave":"保存失败","failedToUpdateStatus":"更新状态失败","group":"分组","groupChangedSuccess":"分组已切换","groupLabel":"分组","groupRequired":"请选择分组","importToCcSwitch":"导入 CCS","inactive":"禁用","ipBlacklist":"IP 黑名单","ipBlacklistHint":"每行一个 IP 或 CIDR。","ipBlacklistPlaceholder":"例如：192.168.1.1","ipRestriction":"IP 访问限制","ipRestrictionEnabled":"已启用 IP 访问限制","ipWhitelist":"IP 白名单","ipWhitelistHint":"每行一个 IP 或 CIDR；留空表示不限制白名单。","ipWhitelistPlaceholder":"例如：192.168.1.1","keyCreatedSuccess":"API Key 已创建","keyDeletedSuccess":"API Key 已删除","keyDisabledSuccess":"API Key 已禁用","keyEnabledSuccess":"API Key 已启用","keyUpdatedSuccess":"API Key 已更新","lastUsedAt":"最后使用","name":"名称","nameLabel":"名称","namePlaceholder":"输入 Key 名称","noExpiration":"永不过期","noGroup":"未分组","noGroupFound":"没有匹配的分组","noKeysYet":"还没有 API Key","quota":"额度","quotaAmountHint":"填 0 或留空表示不限额。","quotaAmountPlaceholder":"0 表示不限额","quotaLimit":"额度限制","quotaResetSuccess":"额度用量已重置","quotaUsed":"已用额度","rateLimitColumn":"频率限制","rateLimitHint":"设置 5 小时、1 天或 7 天窗口内的消费上限。","rateLimit1d":"1 天限制","rateLimit5h":"5 小时限制","rateLimit7d":"7 天限制","rateLimitResetSuccess":"频率限制用量已重置","rateLimitSection":"频率限制","refresh":"刷新","reset":"重置","resetQuotaConfirmMessage":"确定要将 {name} 的已用额度从 ${used} 重置为 0 吗？","resetQuotaTitle":"重置已用额度","resetQuotaUsed":"重置已用额度","resetRateLimitConfirmMessage":"确定要重置 {name} 的频率限制用量吗？","resetRateLimitTitle":"重置频率限制用量","resetRateLimitUsage":"重置频率限制用量","resetNow":"即将重置","resetUsage":"重置用量","saving":"保存中...","searchGroup":"搜索分组","searchPlaceholder":"搜索 API Key","selectGroup":"选择分组","selectStatus":"选择状态","status":"状态","statusActive":"启用","statusExpired":"已过期","statusInactive":"禁用","statusLabel":"状态","statusQuotaExhausted":"额度耗尽","today":"今日","total":"累计","update":"更新","useKey":"使用 Key","usage":"用量"}},"en":{"labels":{"actions":"Actions","active":"Active","allGroups":"All groups","allStatus":"All status","apiKey":"API Key","cancel":"Cancel","ccsClientSelectClaudeCode":"Claude Code","ccsClientSelectClaudeCodeDesc":"Import to Claude Code","ccsClientSelectDescription":"Choose the client to import into.","ccsClientSelectGeminiCli":"Gemini CLI","ccsClientSelectGeminiCliDesc":"Import to Gemini CLI","ccsClientSelectTitle":"Select CCS Client","ccSwitchNotInstalled":"CC Switch client was not detected","clickToChangeGroup":"Click to change group","copyToClipboard":"Copy to clipboard","copied":"Copied","create":"Create","createFirstKey":"Create your first API key to start making requests.","created":"Created","createKey":"Create Key","currentExpiration":"Current expiration","customDate":"Custom date","customKeyHint":"Leave empty to generate a secure API key automatically.","customKeyInvalidChars":"Custom keys can only contain letters, numbers, underscores, and hyphens","customKeyLabel":"Custom Key","customKeyPlaceholder":"Enter a custom API key","customKeyRequired":"Please enter a custom key","customKeyTooShort":"Custom key must be at least 16 characters","delete":"Delete","deleteConfirmMessage":"Delete {name}? This action cannot be undone.","deleteKey":"Delete API Key","disable":"Disable","edit":"Edit","editKey":"Edit Key","enable":"Enable","expiration":"Expiration","expirationDate":"Expiration date","expirationDateHint":"The API key will stop working after this time.","expiresInDays":"Expires in {days} days","expiresAt":"Expires At","extendDays":"Extend {days} days","failedToChangeGroup":"Failed to change group","failedToDelete":"Failed to delete key","failedToLoad":"Failed to load API keys","failedToResetQuota":"Failed to reset quota usage","failedToResetRateLimit":"Failed to reset rate limit usage","failedToSave":"Failed to save API key","failedToUpdateStatus":"Failed to update status","group":"Group","groupChangedSuccess":"Group changed","groupLabel":"Group","groupRequired":"Please select a group","importToCcSwitch":"Import CCS","inactive":"Inactive","ipBlacklist":"IP blacklist","ipBlacklistHint":"One IP or CIDR per line.","ipBlacklistPlaceholder":"Example: 192.168.1.1","ipRestriction":"IP access restriction","ipRestrictionEnabled":"IP access restriction enabled","ipWhitelist":"IP whitelist","ipWhitelistHint":"One IP or CIDR per line. Leave empty to avoid whitelist restrictions.","ipWhitelistPlaceholder":"Example: 192.168.1.1","keyCreatedSuccess":"API key created","keyDeletedSuccess":"API key deleted","keyDisabledSuccess":"API key disabled","keyEnabledSuccess":"API key enabled","keyUpdatedSuccess":"API key updated","lastUsedAt":"Last Used","name":"Name","nameLabel":"Name","namePlaceholder":"Enter key name","noExpiration":"Never expires","noGroup":"No group","noGroupFound":"No matching groups","noKeysYet":"No API keys yet","quota":"Quota","quotaAmountHint":"Use 0 or leave empty for unlimited quota.","quotaAmountPlaceholder":"0 means unlimited","quotaLimit":"Quota limit","quotaResetSuccess":"Quota usage reset","quotaUsed":"Quota used","rateLimitColumn":"Rate Limit","rateLimitHint":"Set spend limits for 5-hour, 1-day, or 7-day windows.","rateLimit1d":"1-day limit","rateLimit5h":"5-hour limit","rateLimit7d":"7-day limit","rateLimitResetSuccess":"Rate limit usage reset","rateLimitSection":"Rate limit","refresh":"Refresh","reset":"Reset","resetQuotaConfirmMessage":"Reset used quota for {name} from ${used} to 0?","resetQuotaTitle":"Reset used quota","resetQuotaUsed":"Reset used quota","resetRateLimitConfirmMessage":"Reset rate limit usage for {name}?","resetRateLimitTitle":"Reset rate limit usage","resetRateLimitUsage":"Reset rate limit usage","resetNow":"Resetting soon","resetUsage":"Reset usage","saving":"Saving...","searchGroup":"Search groups","searchPlaceholder":"Search API keys","selectGroup":"Select group","selectStatus":"Select status","status":"Status","statusActive":"Active","statusExpired":"Expired","statusInactive":"Inactive","statusLabel":"Status","statusQuotaExhausted":"Quota exhausted","today":"Today","total":"Total","update":"Update","useKey":"Use Key","usage":"Usage"}}}`

const zhUseKeyModalLabels = `"useKeyModalAntigravityClaudeNote":"这些环境变量将在当前终端会话中生效。如需永久配置，请将其添加到 ~/.bashrc、~/.zshrc 或相应的配置文件中。","useKeyModalAntigravityDescription":"为 Antigravity 分组配置 API 访问。请根据您使用的客户端选择对应的配置方式。","useKeyModalAntigravityGeminiNote":"这些环境变量将在当前终端会话中生效。如需永久配置，请将其添加到 ~/.bashrc、~/.zshrc 或相应的配置文件中。","useKeyModalCliClaudeCode":"Claude Code","useKeyModalCliCodexCli":"Codex CLI","useKeyModalCliCodexCliWs":"Codex CLI (WebSocket)","useKeyModalCliGeminiCli":"Gemini CLI","useKeyModalCliOpencode":"OpenCode","useKeyModalCopied":"已复制","useKeyModalCopy":"复制","useKeyModalDescription":"将以下环境变量添加到您的终端配置文件或直接在终端中运行。","useKeyModalGeminiDescription":"将以下环境变量添加到您的终端配置文件或直接在终端中运行，以配置 Gemini CLI 访问。","useKeyModalGeminiModelComment":"如果你有 Gemini 3 权限可以填：gemini-3-pro-preview","useKeyModalGeminiNote":"这些环境变量将在当前终端会话中生效。如需永久配置，请将其添加到 ~/.bashrc、~/.zshrc 或相应的配置文件中。","useKeyModalNoGroupDescription":"此 API 密钥尚未分配分组，请先在密钥列表中点击分组列进行分配，然后才能查看使用配置。","useKeyModalNoGroupTitle":"请先分配分组","useKeyModalNote":"这些环境变量将在当前终端会话中生效。如需永久配置，请将其添加到 ~/.bashrc、~/.zshrc 或相应的配置文件中。","useKeyModalOpenAIConfigTomlHint":"请确保以下内容位于 config.toml 文件的开头部分","useKeyModalOpenAIDescription":"将以下配置文件添加到 Codex CLI 配置目录中。","useKeyModalOpenAINote":"请确保配置目录存在。macOS/Linux 用户可运行 mkdir -p ~/.codex 创建目录。","useKeyModalOpenAINoteWindows":"按 Win+R，输入 %userprofile%\\.codex 打开配置目录。如目录不存在，请先手动创建。","useKeyModalOpencodeHint":"配置文件路径：~/.config/opencode/opencode.json（或 opencode.jsonc），不存在需手动创建。可使用默认 provider（openai/anthropic/google）或自定义 provider_id。API Key 支持直接配置或通过客户端 /connect 命令配置。示例仅供参考，模型与选项可按需调整。","useKeyModalTitle":"使用 API 密钥",`

const enUseKeyModalLabels = `"useKeyModalAntigravityClaudeNote":"These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.","useKeyModalAntigravityDescription":"Configure API access for Antigravity group. Select the configuration method based on your client.","useKeyModalAntigravityGeminiNote":"These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.","useKeyModalCliClaudeCode":"Claude Code","useKeyModalCliCodexCli":"Codex CLI","useKeyModalCliCodexCliWs":"Codex CLI (WebSocket)","useKeyModalCliGeminiCli":"Gemini CLI","useKeyModalCliOpencode":"OpenCode","useKeyModalCopied":"Copied","useKeyModalCopy":"Copy","useKeyModalDescription":"Add the following environment variables to your terminal profile or run directly in terminal to configure API access.","useKeyModalGeminiDescription":"Add the following environment variables to your terminal profile or run directly in terminal to configure Gemini CLI access.","useKeyModalGeminiModelComment":"If you have Gemini 3 access, you can use: gemini-3-pro-preview","useKeyModalGeminiNote":"These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.","useKeyModalNoGroupDescription":"This API key has not been assigned to a group. Please click the group column in the key list to assign one before viewing the configuration.","useKeyModalNoGroupTitle":"Please assign a group first","useKeyModalNote":"These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.","useKeyModalOpenAIConfigTomlHint":"Make sure the following content is at the beginning of the config.toml file","useKeyModalOpenAIDescription":"Add the following configuration files to your Codex CLI config directory.","useKeyModalOpenAINote":"Make sure the config directory exists. macOS/Linux users can run mkdir -p ~/.codex to create it.","useKeyModalOpenAINoteWindows":"Press Win+R and enter %userprofile%\\.codex to open the config directory. Create it manually if it does not exist.","useKeyModalOpencodeHint":"Config path: ~/.config/opencode/opencode.json (or opencode.jsonc), create if not exists. Use default providers (openai/anthropic/google) or custom provider_id. API Key can be configured directly or via /connect command. This is an example, adjust models and options as needed.","useKeyModalTitle":"Use API Key",`

const zhAPIKeysEndpointLabels = `"endpointClickToCopy":"点击可复制此端点","endpointCopied":"已复制","endpointCopiedHint":"已复制到剪贴板","endpointDefault":"默认","endpointSpeedTest":"测速","endpointTitle":"API 端点",`

const enAPIKeysEndpointLabels = `"endpointClickToCopy":"Click to copy this endpoint","endpointCopied":"Copied","endpointCopiedHint":"Copied to clipboard","endpointDefault":"Default","endpointSpeedTest":"Speed Test","endpointTitle":"API Endpoints",`

func apiKeysShellConfigDefault() string {
	value := strings.Replace(defaultAPIKeysShellConfig, `"usage":"用量"`, zhAPIKeysEndpointLabels+zhUseKeyModalLabels+`"usage":"用量"`, 1)
	value = strings.Replace(value, `"usage":"Usage"`, enAPIKeysEndpointLabels+enUseKeyModalLabels+`"usage":"Usage"`, 1)
	return value
}

func apiKeysShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return apiKeysShellConfigDefault()
	}
	return value
}

func parseBoolSettingWithDefault(raw string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true":
		return true
	case "false":
		return false
	default:
		return fallback
	}
}

// channelMonitorIntervalMin / channelMonitorIntervalMax bound the default interval
// (mirrors the monitor-level constraint but lives here so setting_service stays decoupled).
const (
	channelMonitorIntervalMin      = 15
	channelMonitorIntervalMax      = 3600
	channelMonitorIntervalFallback = 60
)

// parseChannelMonitorInterval parses the stored string and clamps to [15, 3600].
// Empty / invalid input falls back to channelMonitorIntervalFallback.
func parseChannelMonitorInterval(raw string) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return channelMonitorIntervalFallback
	}
	return clampChannelMonitorInterval(v)
}

// clampChannelMonitorInterval clamps v to the allowed range. 0 means "not provided".
func clampChannelMonitorInterval(v int) int {
	if v <= 0 {
		return 0
	}
	if v < channelMonitorIntervalMin {
		return channelMonitorIntervalMin
	}
	if v > channelMonitorIntervalMax {
		return channelMonitorIntervalMax
	}
	return v
}

// ChannelMonitorRuntime is the lightweight view of the channel monitor feature
// consumed by the runner and user-facing handlers.
type ChannelMonitorRuntime struct {
	Enabled                bool
	DefaultIntervalSeconds int
}

// GetChannelMonitorRuntime reads the channel monitor feature flags directly from
// the settings store. Fail-open: on error returns Enabled=true with the default interval.
func (s *SettingService) GetChannelMonitorRuntime(ctx context.Context) ChannelMonitorRuntime {
	vals, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyChannelMonitorEnabled,
		SettingKeyChannelMonitorDefaultIntervalSeconds,
	})
	if err != nil {
		return ChannelMonitorRuntime{Enabled: true, DefaultIntervalSeconds: channelMonitorIntervalFallback}
	}
	return ChannelMonitorRuntime{
		Enabled:                !isFalseSettingValue(vals[SettingKeyChannelMonitorEnabled]),
		DefaultIntervalSeconds: parseChannelMonitorInterval(vals[SettingKeyChannelMonitorDefaultIntervalSeconds]),
	}
}

// AvailableChannelsRuntime is the lightweight view of the available-channels feature
// switch consumed by the user-facing handler.
type AvailableChannelsRuntime struct {
	Enabled bool
}

// GetAvailableChannelsRuntime reads the available-channels feature switch directly
// from the settings store. Fail-closed: on error returns Enabled=false, matching
// the opt-in default (unknown ↔ disabled).
func (s *SettingService) GetAvailableChannelsRuntime(ctx context.Context) AvailableChannelsRuntime {
	vals, err := s.settingRepo.GetMultiple(ctx, []string{SettingKeyAvailableChannelsEnabled})
	if err != nil {
		return AvailableChannelsRuntime{Enabled: false}
	}
	return AvailableChannelsRuntime{
		Enabled: vals[SettingKeyAvailableChannelsEnabled] == "true",
	}
}

// GetAntigravityUserAgentVersion 返回 Antigravity 上游请求使用的版本号。
// 后台设置优先；为空、缺失或非法时回退到 ANTIGRAVITY_USER_AGENT_VERSION / 内置默认值。
func (s *SettingService) GetAntigravityUserAgentVersion(ctx context.Context) string {
	fallback := antigravity.GetDefaultUserAgentVersion()
	if s == nil || s.settingRepo == nil {
		return fallback
	}
	if cached, ok := s.antigravityUAVersionCache.Load().(*cachedAntigravityUserAgentVersion); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.version
		}
	}

	result, _, _ := s.antigravityUAVersionSF.Do("antigravity_user_agent_version", func() (any, error) {
		if cached, ok := s.antigravityUAVersionCache.Load().(*cachedAntigravityUserAgentVersion); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return cached.version, nil
			}
		}
		if ctx == nil {
			ctx = context.Background()
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), antigravityUserAgentVersionDBTimeout)
		defer cancel()
		value, err := s.settingRepo.GetValue(dbCtx, SettingKeyAntigravityUserAgentVersion)
		if err != nil && !errors.Is(err, ErrSettingNotFound) {
			slog.Warn("failed to get antigravity user agent version setting", "error", err)
			s.antigravityUAVersionCache.Store(&cachedAntigravityUserAgentVersion{
				version:   fallback,
				expiresAt: time.Now().Add(antigravityUserAgentVersionErrorTTL).UnixNano(),
			})
			return fallback, nil
		}
		version := antigravity.NormalizeUserAgentVersion(value)
		if version == "" {
			version = fallback
		}
		s.antigravityUAVersionCache.Store(&cachedAntigravityUserAgentVersion{
			version:   version,
			expiresAt: time.Now().Add(antigravityUserAgentVersionCacheTTL).UnixNano(),
		})
		return version, nil
	})
	if version, ok := result.(string); ok && version != "" {
		return version
	}
	return fallback
}

// GetOpenAICodexUserAgent 返回 OpenAI Codex 上游请求使用的 User-Agent。
// 后台设置优先；为空时回退到内置默认值。
func (s *SettingService) GetOpenAICodexUserAgent(ctx context.Context) string {
	fallback := DefaultOpenAICodexUserAgent
	if s == nil || s.settingRepo == nil {
		return fallback
	}
	if cached, ok := s.openAICodexUACache.Load().(*cachedOpenAICodexUserAgent); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.value
		}
	}

	result, _, _ := s.openAICodexUASF.Do("openai_codex_user_agent", func() (any, error) {
		if cached, ok := s.openAICodexUACache.Load().(*cachedOpenAICodexUserAgent); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return cached.value, nil
			}
		}
		if ctx == nil {
			ctx = context.Background()
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), openAICodexUserAgentDBTimeout)
		defer cancel()
		value, err := s.settingRepo.GetValue(dbCtx, SettingKeyOpenAICodexUserAgent)
		if err != nil && !errors.Is(err, ErrSettingNotFound) {
			slog.Warn("failed to get openai codex user agent setting", "error", err)
			s.openAICodexUACache.Store(&cachedOpenAICodexUserAgent{
				value:     fallback,
				expiresAt: time.Now().Add(openAICodexUserAgentErrorTTL).UnixNano(),
			})
			return fallback, nil
		}
		ua := strings.TrimSpace(value)
		if ua == "" {
			ua = fallback
		}
		s.openAICodexUACache.Store(&cachedOpenAICodexUserAgent{
			value:     ua,
			expiresAt: time.Now().Add(openAICodexUserAgentCacheTTL).UnixNano(),
		})
		return ua, nil
	})
	if ua, ok := result.(string); ok && ua != "" {
		return ua
	}
	return fallback
}

// IsOpenAIAllowClaudeCodeCodexPluginEnabled 全局开关：是否额外放行 Claude Code 的 Codex 插件（默认关闭）。
// 仅在调用方已确认账号 codex_cli_only 开启时读取，避免对非受限账号产生无谓查询。
// 使用进程内 atomic.Value 缓存（60s TTL），避免在每个网关请求热路径上访问 DB。
func (s *SettingService) IsOpenAIAllowClaudeCodeCodexPluginEnabled(ctx context.Context) bool {
	if cached, ok := s.openAIAllowCodexPluginCache.Load().(*cachedOpenAIAllowCodexPlugin); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.value
		}
	}
	result, _, _ := s.openAIAllowCodexPluginSF.Do("openai_allow_codex_plugin_enabled", func() (any, error) {
		if cached, ok := s.openAIAllowCodexPluginCache.Load().(*cachedOpenAIAllowCodexPlugin); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return cached.value, nil
			}
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), openAIAllowCodexPluginDBTimeout)
		defer cancel()
		value, err := s.settingRepo.GetValue(dbCtx, SettingKeyOpenAIAllowClaudeCodeCodexPlugin)
		if err != nil {
			if errors.Is(err, ErrSettingNotFound) {
				// 设置不存在 → 默认关闭，正常 TTL 缓存
				s.openAIAllowCodexPluginCache.Store(&cachedOpenAIAllowCodexPlugin{
					value:     false,
					expiresAt: time.Now().Add(openAIAllowCodexPluginCacheTTL).UnixNano(),
				})
				return false, nil
			}
			slog.Warn("failed to get openai_allow_claude_code_codex_plugin setting", "error", err)
			// DB 错误 → 安全默认关闭，短 TTL 快速重试
			s.openAIAllowCodexPluginCache.Store(&cachedOpenAIAllowCodexPlugin{
				value:     false,
				expiresAt: time.Now().Add(openAIAllowCodexPluginErrorTTL).UnixNano(),
			})
			return false, nil
		}
		enabled := value == "true"
		s.openAIAllowCodexPluginCache.Store(&cachedOpenAIAllowCodexPlugin{
			value:     enabled,
			expiresAt: time.Now().Add(openAIAllowCodexPluginCacheTTL).UnixNano(),
		})
		return enabled, nil
	})
	if val, ok := result.(bool); ok {
		return val
	}
	return false
}

// SetOnUpdateCallback sets a callback function to be called when settings are updated
// This is used for cache invalidation (e.g., HTML cache in frontend server)
func (s *SettingService) SetOnUpdateCallback(callback func()) {
	s.onUpdate = callback
}

// SetVersion sets the application version for injection into public settings
func (s *SettingService) SetVersion(version string) {
	s.version = version
}

// PublicSettingsInjectionPayload is the JSON shape embedded into HTML as
// `window.__APP_CONFIG__` so the frontend can hydrate feature flags & site
// config before the first XHR finishes.
//
// INVARIANT: every `json` tag here MUST also exist on handler/dto.PublicSettings.
// If you forget a feature-flag field here, the frontend's
// `cachedPublicSettings.xxx_enabled` will be `undefined` on refresh until the
// async `/api/v1/settings/public` call returns — which causes opt-in menus
// (strict `=== true`) to flicker off/on. See
// frontend/src/utils/featureFlags.ts for the matching registry.
//
// A unit test diffs this struct's JSON keys against dto.PublicSettings to catch
// drift automatically (see setting_service_injection_test.go).
type PublicSettingsInjectionPayload struct {
	RegistrationEnabled              bool                     `json:"registration_enabled"`
	EmailVerifyEnabled               bool                     `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist []string                 `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool                     `json:"promo_code_enabled"`
	PasswordResetEnabled             bool                     `json:"password_reset_enabled"`
	PasswordMinLength                int                      `json:"password_min_length"`
	InvitationCodeEnabled            bool                     `json:"invitation_code_enabled"`
	TotpEnabled                      bool                     `json:"totp_enabled"`
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
	ModelPlazaItems                  json.RawMessage          `json:"model_plaza_items"`
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
	CustomMenuItems                  json.RawMessage          `json:"custom_menu_items"`
	CustomEndpoints                  json.RawMessage          `json:"custom_endpoints"`
	LinuxDoOAuthEnabled              bool                     `json:"linuxdo_oauth_enabled"`
	DingTalkOAuthEnabled             bool                     `json:"dingtalk_oauth_enabled"`
	WeChatOAuthEnabled               bool                     `json:"wechat_oauth_enabled"`
	WeChatOAuthOpenEnabled           bool                     `json:"wechat_oauth_open_enabled"`
	WeChatOAuthMPEnabled             bool                     `json:"wechat_oauth_mp_enabled"`
	WeChatOAuthMobileEnabled         bool                     `json:"wechat_oauth_mobile_enabled"`
	OIDCOAuthEnabled                 bool                     `json:"oidc_oauth_enabled"`
	OIDCOAuthProviderName            string                   `json:"oidc_oauth_provider_name"`
	GitHubOAuthEnabled               bool                     `json:"github_oauth_enabled"`
	GoogleOAuthEnabled               bool                     `json:"google_oauth_enabled"`
	BackendModeEnabled               bool                     `json:"backend_mode_enabled"`
	PaymentEnabled                   bool                     `json:"payment_enabled"`
	Version                          string                   `json:"version"`
	DefaultLocale                    string                   `json:"default_locale"`
	BalanceLowNotifyEnabled          bool                     `json:"balance_low_notify_enabled"`
	AccountQuotaNotifyEnabled        bool                     `json:"account_quota_notify_enabled"`
	BalanceLowNotifyThreshold        float64                  `json:"balance_low_notify_threshold"`
	BalanceLowNotifyRechargeURL      string                   `json:"balance_low_notify_recharge_url"`

	// Feature flags — MUST match the opt-in/opt-out registry in
	// frontend/src/utils/featureFlags.ts. Missing a field here is the bug
	// that hid the "可用渠道" menu on page refresh.
	ChannelMonitorEnabled                bool `json:"channel_monitor_enabled"`
	ChannelMonitorDefaultIntervalSeconds int  `json:"channel_monitor_default_interval_seconds"`
	AvailableChannelsEnabled             bool `json:"available_channels_enabled"`
	AffiliateEnabled                     bool `json:"affiliate_enabled"`
	RiskControlEnabled                   bool `json:"risk_control_enabled"`

	PromptCasesTitle           string `json:"prompt_cases_title"`
	PromptCasesDescription     string `json:"prompt_cases_description"`
	PromptTemplatesTitle       string `json:"prompt_templates_title"`
	PromptTemplatesDescription string `json:"prompt_templates_description"`
	PromptCatalogShellConfig   string `json:"prompt_catalog_shell_config"`
	WorkspaceShellConfig       string `json:"workspace_shell_config"`
	ImagePromptFilterConfig    string `json:"image_prompt_filter_config"`
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

// GetPublicSettingsForInjection returns public settings in a format suitable for HTML injection.
// This implements the web.PublicSettingsProvider interface.
func (s *SettingService) GetPublicSettingsForInjection(ctx context.Context) (any, error) {
	settings, err := s.GetPublicSettings(ctx)
	if err != nil {
		return nil, err
	}

	return &PublicSettingsInjectionPayload{
		RegistrationEnabled:              settings.RegistrationEnabled,
		EmailVerifyEnabled:               settings.EmailVerifyEnabled,
		RegistrationEmailSuffixWhitelist: settings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 settings.PromoCodeEnabled,
		PasswordResetEnabled:             settings.PasswordResetEnabled,
		PasswordMinLength:                settings.PasswordMinLength,
		InvitationCodeEnabled:            settings.InvitationCodeEnabled,
		TotpEnabled:                      settings.TotpEnabled,
		LoginAgreementEnabled:            settings.LoginAgreementEnabled,
		LoginAgreementMode:               settings.LoginAgreementMode,
		LoginAgreementUpdatedAt:          settings.LoginAgreementUpdatedAt,
		LoginAgreementRevision:           settings.LoginAgreementRevision,
		LoginAgreementDocuments:          settings.LoginAgreementDocuments,
		TurnstileEnabled:                 settings.TurnstileEnabled,
		TurnstileSiteKey:                 settings.TurnstileSiteKey,
		SiteName:                         settings.SiteName,
		SiteLogo:                         settings.SiteLogo,
		SiteSubtitle:                     settings.SiteSubtitle,
		APIBaseURL:                       settings.APIBaseURL,
		ContactInfo:                      settings.ContactInfo,
		DocURL:                           settings.DocURL,
		DocsContentBasePath:              settings.DocsContentBasePath,
		HomeContent:                      settings.HomeContent,
		HomeShellConfig:                  settings.HomeShellConfig,
		HomeBusinessShellConfig:          settings.HomeBusinessShellConfig,
		ModelPlazaItems:                  safeRawJSONArray(settings.ModelPlazaItems),
		ImageWorkspaceModelConfig:        settings.ImageWorkspaceModelConfig,
		ModelPlazaShellConfig:            settings.ModelPlazaShellConfig,
		DocsShellConfig:                  settings.DocsShellConfig,
		LegalDocumentShellConfig:         settings.LegalDocumentShellConfig,
		APIKeysShellConfig:               settings.APIKeysShellConfig,
		KeyUsageShellConfig:              settings.KeyUsageShellConfig,
		DashboardShellConfig:             settings.DashboardShellConfig,
		UsageShellConfig:                 settings.UsageShellConfig,
		APIGuideShellConfig:              settings.APIGuideShellConfig,
		APITestShellConfig:               settings.APITestShellConfig,
		AvailableGroupsShellConfig:       settings.AvailableGroupsShellConfig,
		RedeemShellConfig:                settings.RedeemShellConfig,
		AffiliateShellConfig:             settings.AffiliateShellConfig,
		AvailableChannelsShellConfig:     settings.AvailableChannelsShellConfig,
		ChannelStatusShellConfig:         settings.ChannelStatusShellConfig,
		CustomPageShellConfig:            settings.CustomPageShellConfig,
		ProfileShellConfig:               settings.ProfileShellConfig,
		AuthShellConfig:                  settings.AuthShellConfig,
		HideCcsImportButton:              settings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:      settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:          settings.PurchaseSubscriptionURL,
		TableDefaultPageSize:             settings.TableDefaultPageSize,
		TablePageSizeOptions:             settings.TablePageSizeOptions,
		CustomMenuItems:                  filterUserVisibleMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                  safeRawJSONArray(settings.CustomEndpoints),
		LinuxDoOAuthEnabled:              settings.LinuxDoOAuthEnabled,
		DingTalkOAuthEnabled:             settings.DingTalkOAuthEnabled,
		WeChatOAuthEnabled:               settings.WeChatOAuthEnabled,
		WeChatOAuthOpenEnabled:           settings.WeChatOAuthOpenEnabled,
		WeChatOAuthMPEnabled:             settings.WeChatOAuthMPEnabled,
		WeChatOAuthMobileEnabled:         settings.WeChatOAuthMobileEnabled,
		OIDCOAuthEnabled:                 settings.OIDCOAuthEnabled,
		OIDCOAuthProviderName:            settings.OIDCOAuthProviderName,
		GitHubOAuthEnabled:               settings.GitHubOAuthEnabled,
		GoogleOAuthEnabled:               settings.GoogleOAuthEnabled,
		BackendModeEnabled:               settings.BackendModeEnabled,
		PaymentEnabled:                   settings.PaymentEnabled,
		Version:                          s.version,
		DefaultLocale:                    settings.WebDefaultLocale,
		BalanceLowNotifyEnabled:          settings.BalanceLowNotifyEnabled,
		AccountQuotaNotifyEnabled:        settings.AccountQuotaNotifyEnabled,
		BalanceLowNotifyThreshold:        settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:      settings.BalanceLowNotifyRechargeURL,

		ChannelMonitorEnabled:                settings.ChannelMonitorEnabled,
		ChannelMonitorDefaultIntervalSeconds: settings.ChannelMonitorDefaultIntervalSeconds,
		AvailableChannelsEnabled:             settings.AvailableChannelsEnabled,
		AffiliateEnabled:                     settings.AffiliateEnabled,
		RiskControlEnabled:                   settings.RiskControlEnabled,
		PromptCasesTitle:                     settings.PromptCasesTitle,
		PromptCasesDescription:               settings.PromptCasesDescription,
		PromptTemplatesTitle:                 settings.PromptTemplatesTitle,
		PromptTemplatesDescription:           settings.PromptTemplatesDescription,
		PromptCatalogShellConfig:             settings.PromptCatalogShellConfig,
		WorkspaceShellConfig:                 settings.WorkspaceShellConfig,
		ImagePromptFilterConfig:              settings.ImagePromptFilterConfig,
		PricingTitle:                         settings.PricingTitle,
		PricingDescription:                   settings.PricingDescription,
		PricingShellConfig:                   settings.PricingShellConfig,
		PaymentShellConfig:                   settings.PaymentShellConfig,
		PricingCurrencySymbol:                settings.PricingCurrencySymbol,
		CreditsTitle:                         settings.CreditsTitle,
		CreditsDescription:                   settings.CreditsDescription,
		CreditsPurchaseLabel:                 settings.CreditsPurchaseLabel,
		CreditsBalanceLabel:                  settings.CreditsBalanceLabel,
		CreditsPerBalance:                    settings.CreditsPerBalance,
		CreditsShellConfig:                   settings.CreditsShellConfig,
		GoogleAnalyticsID:                    settings.GoogleAnalyticsID,
		ClarityID:                            settings.ClarityID,
		PlausibleDomain:                      settings.PlausibleDomain,
		PlausibleSrc:                         settings.PlausibleSrc,
		OpenPanelClientID:                    settings.OpenPanelClientID,
		PublicIntegrationsEnabled:            settings.PublicIntegrationsEnabled,
		VercelAnalyticsEnabled:               settings.VercelAnalyticsEnabled,
		AdsenseCode:                          settings.AdsenseCode,
		AffonsoEnabled:                       settings.AffonsoEnabled,
		AffonsoID:                            settings.AffonsoID,
		AffonsoCookieDuration:                settings.AffonsoCookieDuration,
		PromoteKitEnabled:                    settings.PromoteKitEnabled,
		PromoteKitID:                         settings.PromoteKitID,
		CrispEnabled:                         settings.CrispEnabled,
		CrispWebsiteID:                       settings.CrispWebsiteID,
		TawkEnabled:                          settings.TawkEnabled,
		TawkPropertyID:                       settings.TawkPropertyID,
		TawkWidgetID:                         settings.TawkWidgetID,
	}, nil
}

func DefaultWeChatConnectScopesForMode(mode string) string {
	return defaultWeChatConnectScopeForMode(mode)
}

func (s *SettingService) parseWeChatConnectOAuthConfig(settings map[string]string) (WeChatConnectOAuthConfig, error) {
	cfg := s.effectiveWeChatConnectOAuthConfig(settings)

	if !cfg.Enabled || (!cfg.OpenEnabled && !cfg.MPEnabled) {
		return WeChatConnectOAuthConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "wechat oauth is disabled")
	}
	if cfg.OpenEnabled {
		if cfg.AppIDForMode("open") == "" {
			return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth pc app id not configured")
		}
		if cfg.AppSecretForMode("open") == "" {
			return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth pc app secret not configured")
		}
	}
	if cfg.MPEnabled {
		if cfg.AppIDForMode("mp") == "" {
			return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth official account app id not configured")
		}
		if cfg.AppSecretForMode("mp") == "" {
			return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth official account app secret not configured")
		}
	}
	if cfg.MobileEnabled {
		if cfg.AppIDForMode("mobile") == "" {
			return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth mobile app id not configured")
		}
		if cfg.AppSecretForMode("mobile") == "" {
			return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth mobile app secret not configured")
		}
	}
	if v := strings.TrimSpace(cfg.RedirectURL); v != "" {
		if err := config.ValidateAbsoluteHTTPURL(v); err != nil {
			return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth redirect url invalid")
		}
	}
	if err := config.ValidateFrontendRedirectURL(cfg.FrontendRedirectURL); err != nil {
		return WeChatConnectOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth frontend redirect url invalid")
	}
	return cfg, nil
}

func (s *SettingService) weChatOAuthCapabilitiesFromSettings(settings map[string]string) (bool, bool, bool, bool) {
	cfg := s.effectiveWeChatConnectOAuthConfig(settings)
	if !cfg.Enabled {
		return false, false, false, false
	}

	openReady := cfg.OpenEnabled && cfg.AppIDForMode("open") != "" && cfg.AppSecretForMode("open") != ""
	mpReady := cfg.MPEnabled && cfg.AppIDForMode("mp") != "" && cfg.AppSecretForMode("mp") != ""
	mobileReady := cfg.MobileEnabled && cfg.AppIDForMode("mobile") != "" && cfg.AppSecretForMode("mobile") != ""

	return openReady || mpReady, openReady, mpReady, mobileReady
}

func (s *SettingService) emailOAuthBaseConfig(provider string) config.EmailOAuthProviderConfig {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "github":
		cfg := config.EmailOAuthProviderConfig{
			AuthorizeURL:        defaultGitHubOAuthAuthorize,
			TokenURL:            defaultGitHubOAuthToken,
			UserInfoURL:         defaultGitHubOAuthUserInfo,
			EmailsURL:           defaultGitHubOAuthEmails,
			Scopes:              defaultGitHubOAuthScopes,
			FrontendRedirectURL: defaultGitHubOAuthFrontend,
		}
		if s != nil && s.cfg != nil {
			cfg = mergeEmailOAuthBaseConfig(cfg, s.cfg.GitHubOAuth)
		}
		return cfg
	case "google":
		cfg := config.EmailOAuthProviderConfig{
			AuthorizeURL:        defaultGoogleOAuthAuthorize,
			TokenURL:            defaultGoogleOAuthToken,
			UserInfoURL:         defaultGoogleOAuthUserInfo,
			Scopes:              defaultGoogleOAuthScopes,
			FrontendRedirectURL: defaultGoogleOAuthFrontend,
		}
		if s != nil && s.cfg != nil {
			cfg = mergeEmailOAuthBaseConfig(cfg, s.cfg.GoogleOAuth)
		}
		return cfg
	default:
		return config.EmailOAuthProviderConfig{}
	}
}

func mergeEmailOAuthBaseConfig(base, override config.EmailOAuthProviderConfig) config.EmailOAuthProviderConfig {
	base.Enabled = override.Enabled
	if strings.TrimSpace(override.ClientID) != "" {
		base.ClientID = strings.TrimSpace(override.ClientID)
	}
	if strings.TrimSpace(override.ClientSecret) != "" {
		base.ClientSecret = strings.TrimSpace(override.ClientSecret)
	}
	if strings.TrimSpace(override.AuthorizeURL) != "" {
		base.AuthorizeURL = strings.TrimSpace(override.AuthorizeURL)
	}
	if strings.TrimSpace(override.TokenURL) != "" {
		base.TokenURL = strings.TrimSpace(override.TokenURL)
	}
	if strings.TrimSpace(override.UserInfoURL) != "" {
		base.UserInfoURL = strings.TrimSpace(override.UserInfoURL)
	}
	if strings.TrimSpace(override.EmailsURL) != "" {
		base.EmailsURL = strings.TrimSpace(override.EmailsURL)
	}
	if strings.TrimSpace(override.Scopes) != "" {
		base.Scopes = strings.TrimSpace(override.Scopes)
	}
	if strings.TrimSpace(override.RedirectURL) != "" {
		base.RedirectURL = strings.TrimSpace(override.RedirectURL)
	}
	if strings.TrimSpace(override.FrontendRedirectURL) != "" {
		base.FrontendRedirectURL = strings.TrimSpace(override.FrontendRedirectURL)
	}
	return base
}

func (s *SettingService) emailOAuthPublicEnabled(settings map[string]string, provider string) bool {
	cfg := s.effectiveEmailOAuthConfig(settings, provider)
	return cfg.Enabled && strings.TrimSpace(cfg.ClientID) != "" && strings.TrimSpace(cfg.ClientSecret) != ""
}

func (s *SettingService) effectiveEmailOAuthConfig(settings map[string]string, provider string) config.EmailOAuthProviderConfig {
	cfg := s.emailOAuthBaseConfig(provider)
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "github":
		if raw, ok := settings[SettingKeyGitHubOAuthEnabled]; ok {
			cfg.Enabled = raw == "true"
		}
		cfg.ClientID = firstNonEmpty(settings[SettingKeyGitHubOAuthClientID], cfg.ClientID)
		cfg.ClientSecret = firstNonEmpty(settings[SettingKeyGitHubOAuthClientSecret], cfg.ClientSecret)
		cfg.RedirectURL = firstNonEmpty(settings[SettingKeyGitHubOAuthRedirectURL], cfg.RedirectURL)
		cfg.FrontendRedirectURL = firstNonEmpty(settings[SettingKeyGitHubOAuthFrontendRedirectURL], cfg.FrontendRedirectURL, defaultGitHubOAuthFrontend)
	case "google":
		if raw, ok := settings[SettingKeyGoogleOAuthEnabled]; ok {
			cfg.Enabled = raw == "true"
		}
		cfg.ClientID = firstNonEmpty(settings[SettingKeyGoogleOAuthClientID], cfg.ClientID)
		cfg.ClientSecret = firstNonEmpty(settings[SettingKeyGoogleOAuthClientSecret], cfg.ClientSecret)
		cfg.RedirectURL = firstNonEmpty(settings[SettingKeyGoogleOAuthRedirectURL], cfg.RedirectURL)
		cfg.FrontendRedirectURL = firstNonEmpty(settings[SettingKeyGoogleOAuthFrontendRedirectURL], cfg.FrontendRedirectURL, defaultGoogleOAuthFrontend)
	}
	return cfg
}

// filterUserVisibleMenuItems filters out admin-only menu items from a raw JSON
// array string, returning only items with visibility != "admin".
func filterUserVisibleMenuItems(raw string) json.RawMessage {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return json.RawMessage("[]")
	}
	var items []struct {
		Visibility string `json:"visibility"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return json.RawMessage("[]")
	}

	// Parse full items to preserve all fields
	var fullItems []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fullItems); err != nil {
		return json.RawMessage("[]")
	}

	var filtered []json.RawMessage
	for i, item := range items {
		if item.Visibility != "admin" {
			filtered = append(filtered, fullItems[i])
		}
	}
	if len(filtered) == 0 {
		return json.RawMessage("[]")
	}
	result, err := json.Marshal(filtered)
	if err != nil {
		return json.RawMessage("[]")
	}
	return result
}

// safeRawJSONArray returns raw as json.RawMessage if it's valid JSON, otherwise "[]".
func safeRawJSONArray(raw string) json.RawMessage {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return json.RawMessage("[]")
	}
	if json.Valid([]byte(raw)) {
		return json.RawMessage(raw)
	}
	return json.RawMessage("[]")
}

// GetFrameSrcOrigins returns deduplicated http(s) origins from home_content URL,
// purchase_subscription_url, and all custom_menu_items URLs. Used by the router layer for CSP frame-src injection.
func (s *SettingService) GetFrameSrcOrigins(ctx context.Context) ([]string, error) {
	settings, err := s.GetPublicSettings(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var origins []string

	addOrigin := func(rawURL string) {
		if origin := extractOriginFromURL(rawURL); origin != "" {
			if _, ok := seen[origin]; !ok {
				seen[origin] = struct{}{}
				origins = append(origins, origin)
			}
		}
	}

	// home content URL (when home_content is set to a URL for iframe embedding)
	addOrigin(settings.HomeContent)

	// purchase subscription URL
	if settings.PurchaseSubscriptionEnabled {
		addOrigin(settings.PurchaseSubscriptionURL)
	}

	// all custom menu items (including admin-only, since CSP must allow all iframes)
	for _, item := range parseCustomMenuItemURLs(settings.CustomMenuItems) {
		addOrigin(item)
	}

	return origins, nil
}

// extractOriginFromURL returns the scheme+host origin from rawURL.
// Only http and https schemes are accepted.
func extractOriginFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// parseCustomMenuItemURLs extracts URLs from a raw JSON array of custom menu items.
func parseCustomMenuItemURLs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var items []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	urls := make([]string, 0, len(items))
	for _, item := range items {
		if item.URL != "" {
			urls = append(urls, item.URL)
		}
	}
	return urls
}

func oidcUsePKCECompatibilityDefault(base config.OIDCConnectConfig) bool {
	if base.UsePKCEExplicit {
		return base.UsePKCE
	}
	return true
}

func oidcValidateIDTokenCompatibilityDefault(base config.OIDCConnectConfig) bool {
	if base.ValidateIDTokenExplicit {
		return base.ValidateIDToken
	}
	return true
}

func oidcCompatibilityWriteDefault(base config.OIDCConnectConfig, configured bool, raw string, explicit bool, explicitValue bool) bool {
	if configured {
		return strings.TrimSpace(raw) == "true"
	}
	if explicit {
		return explicitValue
	}
	return false
}

// UpdateSettings 更新系统设置
func (s *SettingService) UpdateSettings(ctx context.Context, settings *SystemSettings) error {
	updates, err := s.buildSystemSettingsUpdates(ctx, settings)
	if err != nil {
		return err
	}

	err = s.settingRepo.SetMultiple(ctx, updates)
	if err == nil {
		s.refreshCachedSettings(settings)
	}
	return err
}

func (s *SettingService) OIDCSecurityWriteDefaults(ctx context.Context) (bool, bool, error) {
	rawSettings, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyOIDCConnectUsePKCE,
		SettingKeyOIDCConnectValidateIDToken,
	})
	if err != nil {
		return false, false, fmt.Errorf("get oidc security write defaults: %w", err)
	}

	base := config.OIDCConnectConfig{}
	if s != nil && s.cfg != nil {
		base = s.cfg.OIDC
	}

	rawUsePKCE, hasUsePKCE := rawSettings[SettingKeyOIDCConnectUsePKCE]
	rawValidateIDToken, hasValidateIDToken := rawSettings[SettingKeyOIDCConnectValidateIDToken]

	return oidcCompatibilityWriteDefault(base, hasUsePKCE, rawUsePKCE, base.UsePKCEExplicit, base.UsePKCE),
		oidcCompatibilityWriteDefault(base, hasValidateIDToken, rawValidateIDToken, base.ValidateIDTokenExplicit, base.ValidateIDToken),
		nil
}

// UpdateSettingsWithAuthSourceDefaults persists system settings and auth-source defaults in a single write.
func (s *SettingService) UpdateSettingsWithAuthSourceDefaults(ctx context.Context, settings *SystemSettings, authDefaults *AuthSourceDefaultSettings) error {
	updates, err := s.buildSystemSettingsUpdates(ctx, settings)
	if err != nil {
		return err
	}

	authSourceUpdates, err := s.buildAuthSourceDefaultUpdates(ctx, authDefaults)
	if err != nil {
		return err
	}
	for key, value := range authSourceUpdates {
		updates[key] = value
	}

	err = s.settingRepo.SetMultiple(ctx, updates)
	if err == nil {
		s.refreshCachedSettings(settings)
	}
	return err
}

func (s *SettingService) buildSystemSettingsUpdates(ctx context.Context, settings *SystemSettings) (map[string]string, error) {
	if err := s.validateDefaultSubscriptionGroups(ctx, settings.DefaultSubscriptions); err != nil {
		return nil, err
	}
	normalizedWhitelist, err := NormalizeRegistrationEmailSuffixWhitelist(settings.RegistrationEmailSuffixWhitelist)
	if err != nil {
		return nil, infraerrors.BadRequest("INVALID_REGISTRATION_EMAIL_SUFFIX_WHITELIST", err.Error())
	}
	if normalizedWhitelist == nil {
		normalizedWhitelist = []string{}
	}
	settings.RegistrationEmailSuffixWhitelist = normalizedWhitelist
	alipaySource, err := normalizeVisibleMethodSettingSource("alipay", settings.PaymentVisibleMethodAlipaySource, settings.PaymentVisibleMethodAlipayEnabled)
	if err != nil {
		return nil, err
	}
	wxpaySource, err := normalizeVisibleMethodSettingSource("wxpay", settings.PaymentVisibleMethodWxpaySource, settings.PaymentVisibleMethodWxpayEnabled)
	if err != nil {
		return nil, err
	}
	registrationNotifyProvider, err := normalizeRegistrationNotifyProvider(settings.RegistrationNotifyProvider)
	if err != nil {
		return nil, err
	}
	registrationNotifyWebhookURL, err := normalizeRegistrationNotifyWebhookURL(settings.RegistrationNotifyWebhookURL, settings.RegistrationNotifyEnabled)
	if err != nil {
		return nil, err
	}
	if settings.RegistrationNotifyEnabled && registrationNotifyProvider == "" {
		return nil, infraerrors.BadRequest("REGISTRATION_NOTIFY_PROVIDER_REQUIRED", "registration notification provider is required")
	}
	settings.PaymentVisibleMethodAlipaySource = alipaySource
	settings.PaymentVisibleMethodWxpaySource = wxpaySource
	settings.RegistrationNotifyProvider = registrationNotifyProvider
	settings.RegistrationNotifyWebhookURL = registrationNotifyWebhookURL
	settings.RegistrationNotifySecret = strings.TrimSpace(settings.RegistrationNotifySecret)
	settings.WeChatConnectOpenAppID = strings.TrimSpace(settings.WeChatConnectOpenAppID)
	settings.WeChatConnectOpenAppSecret = strings.TrimSpace(settings.WeChatConnectOpenAppSecret)
	settings.WeChatConnectMPAppID = strings.TrimSpace(settings.WeChatConnectMPAppID)
	settings.WeChatConnectMPAppSecret = strings.TrimSpace(settings.WeChatConnectMPAppSecret)
	settings.WeChatConnectMobileAppID = strings.TrimSpace(settings.WeChatConnectMobileAppID)
	settings.WeChatConnectMobileAppSecret = strings.TrimSpace(settings.WeChatConnectMobileAppSecret)
	settings.WeChatConnectMode = normalizeWeChatConnectStoredMode(
		settings.WeChatConnectOpenEnabled,
		settings.WeChatConnectMPEnabled,
		settings.WeChatConnectMobileEnabled,
		settings.WeChatConnectMode,
	)
	settings.WeChatConnectScopes = normalizeWeChatConnectScopeSetting(settings.WeChatConnectScopes, settings.WeChatConnectMode)
	settings.WeChatConnectRedirectURL = strings.TrimSpace(settings.WeChatConnectRedirectURL)
	settings.WeChatConnectFrontendRedirectURL = strings.TrimSpace(settings.WeChatConnectFrontendRedirectURL)
	if settings.WeChatConnectFrontendRedirectURL == "" {
		settings.WeChatConnectFrontendRedirectURL = defaultWeChatConnectFrontend
	}
	settings.GitHubOAuthRedirectURL = strings.TrimSpace(settings.GitHubOAuthRedirectURL)
	settings.GitHubOAuthFrontendRedirectURL = strings.TrimSpace(settings.GitHubOAuthFrontendRedirectURL)
	if settings.GitHubOAuthFrontendRedirectURL == "" {
		settings.GitHubOAuthFrontendRedirectURL = defaultGitHubOAuthFrontend
	}
	settings.GoogleOAuthRedirectURL = strings.TrimSpace(settings.GoogleOAuthRedirectURL)
	settings.GoogleOAuthFrontendRedirectURL = strings.TrimSpace(settings.GoogleOAuthFrontendRedirectURL)
	if settings.GoogleOAuthFrontendRedirectURL == "" {
		settings.GoogleOAuthFrontendRedirectURL = defaultGoogleOAuthFrontend
	}
	settings.PasswordMinLength = normalizePasswordMinLength(settings.PasswordMinLength)

	updates := make(map[string]string)

	// 注册设置
	updates[SettingKeyRegistrationEnabled] = strconv.FormatBool(settings.RegistrationEnabled)
	updates[SettingKeyEmailVerifyEnabled] = strconv.FormatBool(settings.EmailVerifyEnabled)
	registrationEmailSuffixWhitelistJSON, err := json.Marshal(settings.RegistrationEmailSuffixWhitelist)
	if err != nil {
		return nil, fmt.Errorf("marshal registration email suffix whitelist: %w", err)
	}
	updates[SettingKeyRegistrationEmailSuffixWhitelist] = string(registrationEmailSuffixWhitelistJSON)
	updates[SettingKeyPromoCodeEnabled] = strconv.FormatBool(settings.PromoCodeEnabled)
	updates[SettingKeyPasswordResetEnabled] = strconv.FormatBool(settings.PasswordResetEnabled)
	updates[SettingKeyPasswordMinLength] = strconv.Itoa(settings.PasswordMinLength)
	updates[SettingKeyFrontendURL] = settings.FrontendURL
	updates[SettingKeyInvitationCodeEnabled] = strconv.FormatBool(settings.InvitationCodeEnabled)
	updates[SettingKeyTotpEnabled] = strconv.FormatBool(settings.TotpEnabled)
	settings.LoginAgreementMode = normalizeLoginAgreementMode(settings.LoginAgreementMode)
	settings.LoginAgreementUpdatedAt = strings.TrimSpace(settings.LoginAgreementUpdatedAt)
	if settings.LoginAgreementUpdatedAt == "" {
		settings.LoginAgreementUpdatedAt = defaultLoginAgreementDate
	}
	loginAgreementDocumentsJSON, err := marshalLoginAgreementDocuments(settings.LoginAgreementDocuments)
	if err != nil {
		return nil, err
	}
	updates[SettingKeyLoginAgreementEnabled] = strconv.FormatBool(settings.LoginAgreementEnabled)
	updates[SettingKeyLoginAgreementMode] = settings.LoginAgreementMode
	updates[SettingKeyLoginAgreementUpdatedAt] = settings.LoginAgreementUpdatedAt
	updates[SettingKeyLoginAgreementDocuments] = loginAgreementDocumentsJSON

	// 邮件服务设置（只有非空才更新密码）
	updates[SettingKeySMTPHost] = settings.SMTPHost
	updates[SettingKeySMTPPort] = strconv.Itoa(settings.SMTPPort)
	updates[SettingKeySMTPUsername] = settings.SMTPUsername
	if settings.SMTPPassword != "" {
		updates[SettingKeySMTPPassword] = settings.SMTPPassword
	}
	updates[SettingKeySMTPFrom] = settings.SMTPFrom
	updates[SettingKeySMTPFromName] = settings.SMTPFromName
	updates[SettingKeySMTPUseTLS] = strconv.FormatBool(settings.SMTPUseTLS)
	updates[SettingKeySMTPDailyLimit] = strconv.Itoa(normalizeSMTPDailyLimit(settings.SMTPDailyLimit))
	smtpChannelsJSON, err := marshalSMTPChannels(settings.SMTPChannels)
	if err != nil {
		return nil, fmt.Errorf("marshal smtp channels: %w", err)
	}
	updates[SettingKeySMTPChannels] = smtpChannelsJSON

	// Cloudflare Turnstile 设置（只有非空才更新密钥）
	updates[SettingKeyTurnstileEnabled] = strconv.FormatBool(settings.TurnstileEnabled)
	updates[SettingKeyTurnstileSiteKey] = settings.TurnstileSiteKey
	if settings.TurnstileSecretKey != "" {
		updates[SettingKeyTurnstileSecretKey] = settings.TurnstileSecretKey
	}
	updates[SettingKeyAPIKeyACLTrustForwardedIP] = strconv.FormatBool(settings.APIKeyACLTrustForwardedIP)

	// LinuxDo Connect OAuth 登录
	updates[SettingKeyLinuxDoConnectEnabled] = strconv.FormatBool(settings.LinuxDoConnectEnabled)
	updates[SettingKeyLinuxDoConnectClientID] = settings.LinuxDoConnectClientID
	updates[SettingKeyLinuxDoConnectRedirectURL] = settings.LinuxDoConnectRedirectURL
	if settings.LinuxDoConnectClientSecret != "" {
		updates[SettingKeyLinuxDoConnectClientSecret] = settings.LinuxDoConnectClientSecret
	}

	// DingTalk Connect OAuth 登录
	updates[SettingKeyDingTalkConnectEnabled] = strconv.FormatBool(settings.DingTalkConnectEnabled)
	updates[SettingKeyDingTalkConnectClientID] = settings.DingTalkConnectClientID
	updates[SettingKeyDingTalkConnectRedirectURL] = settings.DingTalkConnectRedirectURL
	if settings.DingTalkConnectClientSecret != "" {
		updates[SettingKeyDingTalkConnectClientSecret] = settings.DingTalkConnectClientSecret
	}
	updates[SettingKeyDingTalkConnectCorpRestrictionPolicy] = settings.DingTalkConnectCorpRestrictionPolicy
	updates[SettingKeyDingTalkConnectInternalCorpID] = settings.DingTalkConnectInternalCorpID
	updates[SettingKeyDingTalkConnectBypassRegistration] = strconv.FormatBool(settings.DingTalkConnectBypassRegistration)
	updates[SettingKeyDingTalkConnectSyncCorpEmail] = strconv.FormatBool(settings.DingTalkConnectSyncCorpEmail)
	updates[SettingKeyDingTalkConnectSyncDisplayName] = strconv.FormatBool(settings.DingTalkConnectSyncDisplayName)
	updates[SettingKeyDingTalkConnectSyncDept] = strconv.FormatBool(settings.DingTalkConnectSyncDept)
	updates[SettingKeyDingTalkConnectSyncCorpEmailAttrKey] = settings.DingTalkConnectSyncCorpEmailAttrKey
	updates[SettingKeyDingTalkConnectSyncDisplayNameAttrKey] = settings.DingTalkConnectSyncDisplayNameAttrKey
	updates[SettingKeyDingTalkConnectSyncDeptAttrKey] = settings.DingTalkConnectSyncDeptAttrKey
	updates[SettingKeyDingTalkConnectSyncCorpEmailAttrName] = settings.DingTalkConnectSyncCorpEmailAttrName
	updates[SettingKeyDingTalkConnectSyncDisplayNameAttrName] = settings.DingTalkConnectSyncDisplayNameAttrName
	updates[SettingKeyDingTalkConnectSyncDeptAttrName] = settings.DingTalkConnectSyncDeptAttrName

	// Generic OIDC OAuth 登录
	updates[SettingKeyOIDCConnectEnabled] = strconv.FormatBool(settings.OIDCConnectEnabled)
	updates[SettingKeyOIDCConnectProviderName] = settings.OIDCConnectProviderName
	updates[SettingKeyOIDCConnectClientID] = settings.OIDCConnectClientID
	updates[SettingKeyOIDCConnectIssuerURL] = settings.OIDCConnectIssuerURL
	updates[SettingKeyOIDCConnectDiscoveryURL] = settings.OIDCConnectDiscoveryURL
	updates[SettingKeyOIDCConnectAuthorizeURL] = settings.OIDCConnectAuthorizeURL
	updates[SettingKeyOIDCConnectTokenURL] = settings.OIDCConnectTokenURL
	updates[SettingKeyOIDCConnectUserInfoURL] = settings.OIDCConnectUserInfoURL
	updates[SettingKeyOIDCConnectJWKSURL] = settings.OIDCConnectJWKSURL
	updates[SettingKeyOIDCConnectScopes] = settings.OIDCConnectScopes
	updates[SettingKeyOIDCConnectRedirectURL] = settings.OIDCConnectRedirectURL
	updates[SettingKeyOIDCConnectFrontendRedirectURL] = settings.OIDCConnectFrontendRedirectURL
	updates[SettingKeyOIDCConnectTokenAuthMethod] = settings.OIDCConnectTokenAuthMethod
	updates[SettingKeyOIDCConnectUsePKCE] = strconv.FormatBool(settings.OIDCConnectUsePKCE)
	updates[SettingKeyOIDCConnectValidateIDToken] = strconv.FormatBool(settings.OIDCConnectValidateIDToken)
	updates[SettingKeyOIDCConnectAllowedSigningAlgs] = settings.OIDCConnectAllowedSigningAlgs
	updates[SettingKeyOIDCConnectClockSkewSeconds] = strconv.Itoa(settings.OIDCConnectClockSkewSeconds)
	updates[SettingKeyOIDCConnectRequireEmailVerified] = strconv.FormatBool(settings.OIDCConnectRequireEmailVerified)
	updates[SettingKeyOIDCConnectUserInfoEmailPath] = settings.OIDCConnectUserInfoEmailPath
	updates[SettingKeyOIDCConnectUserInfoIDPath] = settings.OIDCConnectUserInfoIDPath
	updates[SettingKeyOIDCConnectUserInfoUsernamePath] = settings.OIDCConnectUserInfoUsernamePath
	if settings.OIDCConnectClientSecret != "" {
		updates[SettingKeyOIDCConnectClientSecret] = settings.OIDCConnectClientSecret
	}

	// GitHub / Google 邮箱快捷登录
	updates[SettingKeyGitHubOAuthEnabled] = strconv.FormatBool(settings.GitHubOAuthEnabled)
	updates[SettingKeyGitHubOAuthClientID] = strings.TrimSpace(settings.GitHubOAuthClientID)
	updates[SettingKeyGitHubOAuthRedirectURL] = settings.GitHubOAuthRedirectURL
	updates[SettingKeyGitHubOAuthFrontendRedirectURL] = settings.GitHubOAuthFrontendRedirectURL
	if settings.GitHubOAuthClientSecret != "" {
		updates[SettingKeyGitHubOAuthClientSecret] = strings.TrimSpace(settings.GitHubOAuthClientSecret)
	}
	updates[SettingKeyGoogleOAuthEnabled] = strconv.FormatBool(settings.GoogleOAuthEnabled)
	updates[SettingKeyGoogleOAuthClientID] = strings.TrimSpace(settings.GoogleOAuthClientID)
	updates[SettingKeyGoogleOAuthRedirectURL] = settings.GoogleOAuthRedirectURL
	updates[SettingKeyGoogleOAuthFrontendRedirectURL] = settings.GoogleOAuthFrontendRedirectURL
	if settings.GoogleOAuthClientSecret != "" {
		updates[SettingKeyGoogleOAuthClientSecret] = strings.TrimSpace(settings.GoogleOAuthClientSecret)
	}

	// WeChat Connect OAuth 登录
	updates[SettingKeyWeChatConnectEnabled] = strconv.FormatBool(settings.WeChatConnectEnabled)
	updates[SettingKeyWeChatConnectOpenAppID] = settings.WeChatConnectOpenAppID
	updates[SettingKeyWeChatConnectMPAppID] = settings.WeChatConnectMPAppID
	updates[SettingKeyWeChatConnectMobileAppID] = settings.WeChatConnectMobileAppID
	updates[SettingKeyWeChatConnectOpenEnabled] = strconv.FormatBool(settings.WeChatConnectOpenEnabled)
	updates[SettingKeyWeChatConnectMPEnabled] = strconv.FormatBool(settings.WeChatConnectMPEnabled)
	updates[SettingKeyWeChatConnectMobileEnabled] = strconv.FormatBool(settings.WeChatConnectMobileEnabled)
	updates[SettingKeyWeChatConnectMode] = settings.WeChatConnectMode
	updates[SettingKeyWeChatConnectScopes] = settings.WeChatConnectScopes
	updates[SettingKeyWeChatConnectRedirectURL] = settings.WeChatConnectRedirectURL
	updates[SettingKeyWeChatConnectFrontendRedirectURL] = settings.WeChatConnectFrontendRedirectURL
	if settings.WeChatConnectOpenAppSecret != "" {
		updates[SettingKeyWeChatConnectOpenAppSecret] = settings.WeChatConnectOpenAppSecret
	}
	if settings.WeChatConnectMPAppSecret != "" {
		updates[SettingKeyWeChatConnectMPAppSecret] = settings.WeChatConnectMPAppSecret
	}
	if settings.WeChatConnectMobileAppSecret != "" {
		updates[SettingKeyWeChatConnectMobileAppSecret] = settings.WeChatConnectMobileAppSecret
	}

	// OEM设置
	updates[SettingKeySiteName] = settings.SiteName
	updates[SettingKeySiteLogo] = settings.SiteLogo
	updates[SettingKeySiteSubtitle] = settings.SiteSubtitle
	updates[SettingKeyAPIBaseURL] = settings.APIBaseURL
	updates[SettingKeyContactInfo] = settings.ContactInfo
	updates[SettingKeyDocURL] = settings.DocURL
	updates[SettingKeyDocsContentBasePath] = strings.TrimSpace(settings.DocsContentBasePath)
	updates[SettingKeyHomeContent] = settings.HomeContent
	updates[SettingKeyHomeShellConfig] = strings.TrimSpace(settings.HomeShellConfig)
	updates[SettingKeyHomeBusinessShellConfig] = strings.TrimSpace(settings.HomeBusinessShellConfig)
	updates[SettingKeyModelPlazaItems] = settings.ModelPlazaItems
	updates[SettingKeyImageWorkspaceModelConfig] = strings.TrimSpace(settings.ImageWorkspaceModelConfig)
	updates[SettingKeyModelPlazaShellConfig] = strings.TrimSpace(settings.ModelPlazaShellConfig)
	updates[SettingKeyDocsShellConfig] = strings.TrimSpace(settings.DocsShellConfig)
	updates[SettingKeyLegalDocumentShellConfig] = strings.TrimSpace(settings.LegalDocumentShellConfig)
	updates[SettingKeyAPIKeysShellConfig] = strings.TrimSpace(settings.APIKeysShellConfig)
	updates[SettingKeyKeyUsageShellConfig] = strings.TrimSpace(settings.KeyUsageShellConfig)
	updates[SettingKeyDashboardShellConfig] = strings.TrimSpace(settings.DashboardShellConfig)
	updates[SettingKeyUsageShellConfig] = strings.TrimSpace(settings.UsageShellConfig)
	updates[SettingKeyAPIGuideShellConfig] = strings.TrimSpace(settings.APIGuideShellConfig)
	updates[SettingKeyAPITestShellConfig] = strings.TrimSpace(settings.APITestShellConfig)
	updates[SettingKeyAvailableGroupsShellConfig] = strings.TrimSpace(settings.AvailableGroupsShellConfig)
	updates[SettingKeyRedeemShellConfig] = strings.TrimSpace(settings.RedeemShellConfig)
	updates[SettingKeyAffiliateShellConfig] = strings.TrimSpace(settings.AffiliateShellConfig)
	updates[SettingKeyAvailableChannelsShellConfig] = strings.TrimSpace(settings.AvailableChannelsShellConfig)
	updates[SettingKeyChannelStatusShellConfig] = strings.TrimSpace(settings.ChannelStatusShellConfig)
	updates[SettingKeyCustomPageShellConfig] = strings.TrimSpace(settings.CustomPageShellConfig)
	updates[SettingKeyProfileShellConfig] = strings.TrimSpace(settings.ProfileShellConfig)
	updates[SettingKeyAuthShellConfig] = strings.TrimSpace(settings.AuthShellConfig)
	updates[SettingKeyHideCcsImportButton] = strconv.FormatBool(settings.HideCcsImportButton)
	updates[SettingKeyPurchaseSubscriptionEnabled] = strconv.FormatBool(settings.PurchaseSubscriptionEnabled)
	updates[SettingKeyPurchaseSubscriptionURL] = strings.TrimSpace(settings.PurchaseSubscriptionURL)
	tableDefaultPageSize, tablePageSizeOptions := normalizeTablePreferences(
		settings.TableDefaultPageSize,
		settings.TablePageSizeOptions,
	)
	updates[SettingKeyTableDefaultPageSize] = strconv.Itoa(tableDefaultPageSize)
	tablePageSizeOptionsJSON, err := json.Marshal(tablePageSizeOptions)
	if err != nil {
		return nil, fmt.Errorf("marshal table page size options: %w", err)
	}
	updates[SettingKeyTablePageSizeOptions] = string(tablePageSizeOptionsJSON)
	updates[SettingKeyCustomMenuItems] = settings.CustomMenuItems
	updates[SettingKeyCustomEndpoints] = settings.CustomEndpoints
	updates[SettingKeyWebAppURL] = strings.TrimSpace(settings.WebAppURL)
	updates[SettingKeyWebAppName] = strings.TrimSpace(settings.WebAppName)
	updates[SettingKeyWebAppDescription] = strings.TrimSpace(settings.WebAppDescription)
	updates[SettingKeyWebAppLogo] = strings.TrimSpace(settings.WebAppLogo)
	updates[SettingKeyWebAppFavicon] = strings.TrimSpace(settings.WebAppFavicon)
	updates[SettingKeyWebAppPreviewImage] = strings.TrimSpace(settings.WebAppPreviewImage)
	updates[SettingKeyWebTheme] = strings.TrimSpace(settings.WebTheme)
	updates[SettingKeyWebAppearance] = strings.TrimSpace(settings.WebAppearance)
	updates[SettingKeyWebDefaultLocale] = strings.TrimSpace(settings.WebDefaultLocale)
	updates[SettingKeyPromptCasesTitle] = strings.TrimSpace(settings.WebPromptCasesTitle)
	updates[SettingKeyPromptCasesDescription] = strings.TrimSpace(settings.WebPromptCasesDescription)
	updates[SettingKeyPromptTemplatesTitle] = strings.TrimSpace(settings.WebPromptTemplatesTitle)
	updates[SettingKeyPromptTemplatesDescription] = strings.TrimSpace(settings.WebPromptTemplatesDescription)
	updates[SettingKeyPromptCatalogShellConfig] = strings.TrimSpace(settings.PromptCatalogShellConfig)
	updates[SettingKeyWorkspaceShellConfig] = strings.TrimSpace(settings.WebWorkspaceShellConfig)
	updates[SettingKeyImagePromptFilterConfig] = strings.TrimSpace(settings.WebImagePromptFilterConfig)
	updates[SettingKeyPricingTitle] = strings.TrimSpace(settings.WebPricingTitle)
	updates[SettingKeyPricingDescription] = strings.TrimSpace(settings.WebPricingDescription)
	updates[SettingKeyPricingShellConfig] = strings.TrimSpace(settings.WebPricingShellConfig)
	updates[SettingKeyPaymentShellConfig] = strings.TrimSpace(settings.WebPaymentShellConfig)
	updates[SettingKeyPricingCurrencySymbol] = pricingCurrencySymbolSetting(settings.WebPricingCurrencySymbol)
	updates[SettingKeyCreditsTitle] = strings.TrimSpace(settings.WebCreditsTitle)
	updates[SettingKeyCreditsDescription] = strings.TrimSpace(settings.WebCreditsDescription)
	updates[SettingKeyCreditsPurchaseLabel] = strings.TrimSpace(settings.WebCreditsPurchaseLabel)
	updates[SettingKeyCreditsBalanceLabel] = strings.TrimSpace(settings.WebCreditsBalanceLabel)
	updates[SettingKeyCreditsPerBalance] = creditsPerBalanceSetting(settings.WebCreditsPerBalance)
	updates[SettingKeyCreditsShellConfig] = strings.TrimSpace(settings.CreditsShellConfig)
	updates[SettingKeyWebLocaleDetectEnabled] = strconv.FormatBool(settings.WebLocaleDetectEnabled)
	updates[SettingKeyWebEmailAuthVisible] = strconv.FormatBool(settings.WebEmailAuthVisible)
	updates[SettingKeyWebGoogleAuthVisible] = strconv.FormatBool(settings.WebGoogleAuthVisible)
	updates[SettingKeyWebGitHubAuthVisible] = strconv.FormatBool(settings.WebGitHubAuthVisible)
	updates[SettingKeyWebGoogleAnalyticsID] = strings.TrimSpace(settings.WebGoogleAnalyticsID)
	updates[SettingKeyWebClarityID] = strings.TrimSpace(settings.WebClarityID)
	updates[SettingKeyWebPlausibleDomain] = strings.TrimSpace(settings.WebPlausibleDomain)
	updates[SettingKeyWebPlausibleSrc] = strings.TrimSpace(settings.WebPlausibleSrc)
	updates[SettingKeyWebOpenPanelClientID] = strings.TrimSpace(settings.WebOpenPanelClientID)
	updates[SettingKeyWebPublicIntegrationsEnabled] = strconv.FormatBool(settings.WebPublicIntegrationsEnabled)
	updates[SettingKeyWebVercelAnalyticsEnabled] = strconv.FormatBool(settings.WebVercelAnalyticsEnabled)
	updates[SettingKeyWebAdsenseCode] = strings.TrimSpace(settings.WebAdsenseCode)
	updates[SettingKeyWebAffonsoEnabled] = strconv.FormatBool(settings.WebAffonsoEnabled)
	updates[SettingKeyWebAffonsoID] = strings.TrimSpace(settings.WebAffonsoID)
	updates[SettingKeyWebAffonsoCookieDuration] = strings.TrimSpace(settings.WebAffonsoCookieDuration)
	updates[SettingKeyWebPromoteKitEnabled] = strconv.FormatBool(settings.WebPromoteKitEnabled)
	updates[SettingKeyWebPromoteKitID] = strings.TrimSpace(settings.WebPromoteKitID)
	updates[SettingKeyWebCrispEnabled] = strconv.FormatBool(settings.WebCrispEnabled)
	updates[SettingKeyWebCrispWebsiteID] = strings.TrimSpace(settings.WebCrispWebsiteID)
	updates[SettingKeyWebTawkEnabled] = strconv.FormatBool(settings.WebTawkEnabled)
	updates[SettingKeyWebTawkPropertyID] = strings.TrimSpace(settings.WebTawkPropertyID)
	updates[SettingKeyWebTawkWidgetID] = strings.TrimSpace(settings.WebTawkWidgetID)

	// 默认配置
	updates[SettingKeyDefaultConcurrency] = strconv.Itoa(settings.DefaultConcurrency)
	updates[SettingKeyDefaultBalance] = strconv.FormatFloat(settings.DefaultBalance, 'f', 8, 64)
	settings.AffiliateRebateRate = clampAffiliateRebateRate(settings.AffiliateRebateRate)
	updates[SettingKeyAffiliateRebateRate] = strconv.FormatFloat(settings.AffiliateRebateRate, 'f', 8, 64)
	if settings.AffiliateRebateFreezeHours < 0 {
		settings.AffiliateRebateFreezeHours = AffiliateRebateFreezeHoursDefault
	}
	if settings.AffiliateRebateFreezeHours > AffiliateRebateFreezeHoursMax {
		settings.AffiliateRebateFreezeHours = AffiliateRebateFreezeHoursMax
	}
	updates[SettingKeyAffiliateRebateFreezeHours] = strconv.Itoa(settings.AffiliateRebateFreezeHours)
	if settings.AffiliateRebateDurationDays < 0 {
		settings.AffiliateRebateDurationDays = AffiliateRebateDurationDaysDefault
	}
	if settings.AffiliateRebateDurationDays > AffiliateRebateDurationDaysMax {
		settings.AffiliateRebateDurationDays = AffiliateRebateDurationDaysMax
	}
	updates[SettingKeyAffiliateRebateDurationDays] = strconv.Itoa(settings.AffiliateRebateDurationDays)
	if settings.AffiliateRebatePerInviteeCap < 0 {
		settings.AffiliateRebatePerInviteeCap = AffiliateRebatePerInviteeCapDefault
	}
	updates[SettingKeyAffiliateRebatePerInviteeCap] = strconv.FormatFloat(settings.AffiliateRebatePerInviteeCap, 'f', 8, 64)
	updates[SettingKeyDefaultUserRPMLimit] = strconv.Itoa(settings.DefaultUserRPMLimit)
	defaultSubsJSON, err := json.Marshal(settings.DefaultSubscriptions)
	if err != nil {
		return nil, fmt.Errorf("marshal default subscriptions: %w", err)
	}
	updates[SettingKeyDefaultSubscriptions] = string(defaultSubsJSON)

	// Model fallback configuration
	updates[SettingKeyEnableModelFallback] = strconv.FormatBool(settings.EnableModelFallback)
	updates[SettingKeyFallbackModelAnthropic] = settings.FallbackModelAnthropic
	updates[SettingKeyFallbackModelOpenAI] = settings.FallbackModelOpenAI
	updates[SettingKeyFallbackModelGemini] = settings.FallbackModelGemini
	updates[SettingKeyFallbackModelAntigravity] = settings.FallbackModelAntigravity

	// Identity patch configuration (Claude -> Gemini)
	updates[SettingKeyEnableIdentityPatch] = strconv.FormatBool(settings.EnableIdentityPatch)
	updates[SettingKeyIdentityPatchPrompt] = settings.IdentityPatchPrompt

	// Ops monitoring (vNext)
	updates[SettingKeyOpsMonitoringEnabled] = strconv.FormatBool(settings.OpsMonitoringEnabled)
	updates[SettingKeyOpsRealtimeMonitoringEnabled] = strconv.FormatBool(settings.OpsRealtimeMonitoringEnabled)
	updates[SettingKeyOpsQueryModeDefault] = string(ParseOpsQueryMode(settings.OpsQueryModeDefault))
	if settings.OpsMetricsIntervalSeconds > 0 {
		updates[SettingKeyOpsMetricsIntervalSeconds] = strconv.Itoa(settings.OpsMetricsIntervalSeconds)
	}

	// Channel monitor feature switch
	updates[SettingKeyChannelMonitorEnabled] = strconv.FormatBool(settings.ChannelMonitorEnabled)
	if v := clampChannelMonitorInterval(settings.ChannelMonitorDefaultIntervalSeconds); v > 0 {
		updates[SettingKeyChannelMonitorDefaultIntervalSeconds] = strconv.Itoa(v)
	}

	// Available channels feature switch
	updates[SettingKeyAvailableChannelsEnabled] = strconv.FormatBool(settings.AvailableChannelsEnabled)

	// Affiliate (邀请返利) feature switch
	updates[SettingKeyAffiliateEnabled] = strconv.FormatBool(settings.AffiliateEnabled)

	// 风控中心功能开关
	updates[SettingKeyRiskControlEnabled] = strconv.FormatBool(settings.RiskControlEnabled)

	// Claude Code version check
	updates[SettingKeyMinClaudeCodeVersion] = settings.MinClaudeCodeVersion
	updates[SettingKeyMaxClaudeCodeVersion] = settings.MaxClaudeCodeVersion

	// 分组隔离
	updates[SettingKeyAllowUngroupedKeyScheduling] = strconv.FormatBool(settings.AllowUngroupedKeyScheduling)

	// Backend Mode
	updates[SettingKeyBackendModeEnabled] = strconv.FormatBool(settings.BackendModeEnabled)

	// Gateway forwarding behavior
	updates[SettingKeyEnableFingerprintUnification] = strconv.FormatBool(settings.EnableFingerprintUnification)
	updates[SettingKeyEnableMetadataPassthrough] = strconv.FormatBool(settings.EnableMetadataPassthrough)
	updates[SettingKeyEnableCCHSigning] = strconv.FormatBool(settings.EnableCCHSigning)
	updates[SettingKeyEnableAnthropicCacheTTL1hInjection] = strconv.FormatBool(settings.EnableAnthropicCacheTTL1hInjection)
	updates[SettingKeyRewriteMessageCacheControl] = strconv.FormatBool(settings.RewriteMessageCacheControl)
	updates[SettingKeyAntigravityUserAgentVersion] = antigravity.NormalizeUserAgentVersion(settings.AntigravityUserAgentVersion)
	updates[SettingKeyOpenAICodexUserAgent] = strings.TrimSpace(settings.OpenAICodexUserAgent)
	updates[SettingKeyOpenAIAllowClaudeCodeCodexPlugin] = strconv.FormatBool(settings.OpenAIAllowClaudeCodeCodexPlugin)
	updates[SettingPaymentVisibleMethodAlipaySource] = settings.PaymentVisibleMethodAlipaySource
	updates[SettingPaymentVisibleMethodWxpaySource] = settings.PaymentVisibleMethodWxpaySource
	updates[SettingPaymentVisibleMethodAlipayEnabled] = strconv.FormatBool(settings.PaymentVisibleMethodAlipayEnabled)
	updates[SettingPaymentVisibleMethodWxpayEnabled] = strconv.FormatBool(settings.PaymentVisibleMethodWxpayEnabled)
	updates[openAIAdvancedSchedulerSettingKey] = strconv.FormatBool(settings.OpenAIAdvancedSchedulerEnabled)

	// 余额、订阅到期与账号限额通知
	updates[SettingKeyBalanceLowNotifyEnabled] = strconv.FormatBool(settings.BalanceLowNotifyEnabled)
	updates[SettingKeyBalanceLowNotifyThreshold] = strconv.FormatFloat(settings.BalanceLowNotifyThreshold, 'f', 8, 64)
	updates[SettingKeyBalanceLowNotifyRechargeURL] = settings.BalanceLowNotifyRechargeURL
	updates[SettingKeySubscriptionExpiryNotifyEnabled] = strconv.FormatBool(settings.SubscriptionExpiryNotifyEnabled)
	updates[SettingKeyAccountQuotaNotifyEnabled] = strconv.FormatBool(settings.AccountQuotaNotifyEnabled)
	updates[SettingKeyAccountQuotaNotifyEmails] = MarshalNotifyEmails(settings.AccountQuotaNotifyEmails)
	updates[SettingKeyRegistrationNotifyEnabled] = strconv.FormatBool(settings.RegistrationNotifyEnabled)
	updates[SettingKeyRegistrationNotifyProvider] = settings.RegistrationNotifyProvider
	updates[SettingKeyRegistrationNotifyWebhookURL] = settings.RegistrationNotifyWebhookURL
	if settings.RegistrationNotifySecret != "" {
		updates[SettingKeyRegistrationNotifySecret] = settings.RegistrationNotifySecret
	}

	// 系统全局 platform quota：整体替换语义（null/缺省 = 不限制）。
	if settings.DefaultPlatformQuotas != nil {
		if err := validateDefaultPlatformQuotaMap(settings.DefaultPlatformQuotas); err != nil {
			return nil, err
		}
		blob, err := json.Marshal(settings.DefaultPlatformQuotas)
		if err != nil {
			return nil, fmt.Errorf("marshal default platform quotas: %w", err)
		}
		updates[SettingKeyDefaultPlatformQuotas] = string(blob)
	}

	return updates, nil
}

// validateDefaultPlatformQuotaMap 校验 platform quota map 的合法性：
// 平台名须在 AllowedQuotaPlatforms 白名单内，每个非 nil 上限须 finite 且 >= 0。
// 系统层和 auth-source 层共用此 helper。
func validateDefaultPlatformQuotaMap(m map[string]*DefaultPlatformQuotaSetting) error {
	for platform, pq := range m {
		if !IsAllowedQuotaPlatform(platform) {
			return infraerrors.BadRequest("INVALID_DEFAULT_PLATFORM_QUOTA", fmt.Sprintf("unknown platform %q", platform))
		}
		if pq == nil {
			continue
		}
		for _, v := range []*float64{pq.DailyLimitUSD, pq.WeeklyLimitUSD, pq.MonthlyLimitUSD} {
			if v != nil && (*v < 0 || math.IsNaN(*v) || math.IsInf(*v, 0)) {
				return infraerrors.BadRequest("INVALID_DEFAULT_PLATFORM_QUOTA", "platform quota limit must be a finite non-negative number")
			}
		}
	}
	return nil
}

func (s *SettingService) buildAuthSourceDefaultUpdates(ctx context.Context, settings *AuthSourceDefaultSettings) (map[string]string, error) {
	if settings == nil {
		return nil, nil
	}

	for _, subscriptions := range [][]DefaultSubscriptionSetting{
		settings.Email.Subscriptions,
		settings.LinuxDo.Subscriptions,
		settings.OIDC.Subscriptions,
		settings.WeChat.Subscriptions,
		settings.GitHub.Subscriptions,
		settings.Google.Subscriptions,
		settings.DingTalk.Subscriptions,
	} {
		if err := s.validateDefaultSubscriptionGroups(ctx, subscriptions); err != nil {
			return nil, err
		}
	}

	// 校验各 auth source 的 platform quota map（改动 C：对等系统层校验）
	for _, pgs := range []struct {
		name string
		pq   map[string]*DefaultPlatformQuotaSetting
	}{
		{"email", settings.Email.PlatformQuotas},
		{"linuxdo", settings.LinuxDo.PlatformQuotas},
		{"oidc", settings.OIDC.PlatformQuotas},
		{"wechat", settings.WeChat.PlatformQuotas},
		{"github", settings.GitHub.PlatformQuotas},
		{"google", settings.Google.PlatformQuotas},
		{"dingtalk", settings.DingTalk.PlatformQuotas},
	} {
		if pgs.pq != nil {
			if err := validateDefaultPlatformQuotaMap(pgs.pq); err != nil {
				return nil, err
			}
		}
	}

	updates := make(map[string]string, 36)
	writeProviderDefaultGrantUpdates(updates, emailAuthSourceDefaultKeys, settings.Email)
	writeProviderDefaultGrantUpdates(updates, linuxDoAuthSourceDefaultKeys, settings.LinuxDo)
	writeProviderDefaultGrantUpdates(updates, oidcAuthSourceDefaultKeys, settings.OIDC)
	writeProviderDefaultGrantUpdates(updates, weChatAuthSourceDefaultKeys, settings.WeChat)
	writeProviderDefaultGrantUpdates(updates, gitHubAuthSourceDefaultKeys, settings.GitHub)
	writeProviderDefaultGrantUpdates(updates, googleAuthSourceDefaultKeys, settings.Google)
	writeProviderDefaultGrantUpdates(updates, dingTalkAuthSourceDefaultKeys, settings.DingTalk)
	updates[SettingKeyForceEmailOnThirdPartySignup] = strconv.FormatBool(settings.ForceEmailOnThirdPartySignup)
	return updates, nil
}

func (s *SettingService) refreshCachedSettings(settings *SystemSettings) {
	if settings == nil {
		return
	}

	// 先使 inflight singleflight 失效，再刷新缓存，缩小旧值覆盖新值的竞态窗口
	versionBoundsSF.Forget("version_bounds")
	versionBoundsCache.Store(&cachedVersionBounds{
		min:       settings.MinClaudeCodeVersion,
		max:       settings.MaxClaudeCodeVersion,
		expiresAt: time.Now().Add(versionBoundsCacheTTL).UnixNano(),
	})
	backendModeSF.Forget("backend_mode")
	backendModeCache.Store(&cachedBackendMode{
		value:     settings.BackendModeEnabled,
		expiresAt: time.Now().Add(backendModeCacheTTL).UnixNano(),
	})
	gatewayForwardingSF.Forget("gateway_forwarding")
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{
		fingerprintUnification:       settings.EnableFingerprintUnification,
		metadataPassthrough:          settings.EnableMetadataPassthrough,
		cchSigning:                   settings.EnableCCHSigning,
		anthropicCacheTTL1hInjection: settings.EnableAnthropicCacheTTL1hInjection,
		rewriteMessageCacheControl:   settings.RewriteMessageCacheControl,
		expiresAt:                    time.Now().Add(gatewayForwardingCacheTTL).UnixNano(),
	})
	s.antigravityUAVersionSF.Forget("antigravity_user_agent_version")
	antigravityUserAgentVersion := antigravity.NormalizeUserAgentVersion(settings.AntigravityUserAgentVersion)
	if antigravityUserAgentVersion == "" {
		antigravityUserAgentVersion = antigravity.GetDefaultUserAgentVersion()
	}
	s.antigravityUAVersionCache.Store(&cachedAntigravityUserAgentVersion{
		version:   antigravityUserAgentVersion,
		expiresAt: time.Now().Add(antigravityUserAgentVersionCacheTTL).UnixNano(),
	})
	s.openAICodexUASF.Forget("openai_codex_user_agent")
	codexUA := strings.TrimSpace(settings.OpenAICodexUserAgent)
	if codexUA == "" {
		codexUA = DefaultOpenAICodexUserAgent
	}
	s.openAICodexUACache.Store(&cachedOpenAICodexUserAgent{
		value:     codexUA,
		expiresAt: time.Now().Add(openAICodexUserAgentCacheTTL).UnixNano(),
	})
	openAIAdvancedSchedulerSettingSF.Forget(openAIAdvancedSchedulerSettingKey)
	openAIAdvancedSchedulerSettingCache.Store(&cachedOpenAIAdvancedSchedulerSetting{
		enabled:   settings.OpenAIAdvancedSchedulerEnabled,
		expiresAt: time.Now().Add(openAIAdvancedSchedulerSettingCacheTTL).UnixNano(),
	})
	// Invalidate the quota auto-pause cache and let the next read trigger a fresh load.
	// We can't know from here whether ops_advanced_settings was also touched, so be
	// defensive: store an expired entry — GetOpenAIQuotaAutoPauseSettings will serve
	// stale and kick off an async refresh, never blocking the request that follows.
	s.openAIQuotaAutoPauseSettingsSF.Forget(openAIQuotaAutoPauseSettingsRefreshKey)
	if cached, _ := s.openAIQuotaAutoPauseSettingsCache.Load().(*cachedOpenAIQuotaAutoPauseSettings); cached != nil {
		s.openAIQuotaAutoPauseSettingsCache.Store(&cachedOpenAIQuotaAutoPauseSettings{
			settings:  cached.settings,
			expiresAt: 0,
		})
	}
	if s.cfg != nil {
		s.cfg.SetTrustForwardedIPForAPIKeyACL(settings.APIKeyACLTrustForwardedIP)
	}
	s.openAIAllowCodexPluginSF.Forget("openai_allow_codex_plugin_enabled")
	s.openAIAllowCodexPluginCache.Store(&cachedOpenAIAllowCodexPlugin{
		value:     settings.OpenAIAllowClaudeCodeCodexPlugin,
		expiresAt: time.Now().Add(openAIAllowCodexPluginCacheTTL).UnixNano(),
	})
	if s.onUpdate != nil {
		s.onUpdate() // Invalidate cache after settings update
	}
}

func (s *SettingService) defaultRewriteMessageCacheControl() bool {
	return false
}

func (s *SettingService) validateDefaultSubscriptionGroups(ctx context.Context, items []DefaultSubscriptionSetting) error {
	if len(items) == 0 {
		return nil
	}

	checked := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if item.GroupID <= 0 {
			continue
		}
		if _, ok := checked[item.GroupID]; ok {
			return ErrDefaultSubGroupDuplicate.WithMetadata(map[string]string{
				"group_id": strconv.FormatInt(item.GroupID, 10),
			})
		}
		checked[item.GroupID] = struct{}{}
		if s.defaultSubGroupReader == nil {
			continue
		}

		group, err := s.defaultSubGroupReader.GetByID(ctx, item.GroupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				return ErrDefaultSubGroupInvalid.WithMetadata(map[string]string{
					"group_id": strconv.FormatInt(item.GroupID, 10),
				})
			}
			return fmt.Errorf("get default subscription group %d: %w", item.GroupID, err)
		}
		if !group.IsSubscriptionType() {
			return ErrDefaultSubGroupInvalid.WithMetadata(map[string]string{
				"group_id": strconv.FormatInt(item.GroupID, 10),
			})
		}
	}

	return nil
}

func (s *SettingService) GetEmailOAuthProviderConfig(ctx context.Context, provider string) (config.EmailOAuthProviderConfig, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "github" && provider != "google" {
		return config.EmailOAuthProviderConfig{}, infraerrors.NotFound("OAUTH_PROVIDER_NOT_FOUND", "oauth provider not found")
	}
	keys := []string{
		SettingKeyGitHubOAuthEnabled,
		SettingKeyGitHubOAuthClientID,
		SettingKeyGitHubOAuthClientSecret,
		SettingKeyGitHubOAuthRedirectURL,
		SettingKeyGitHubOAuthFrontendRedirectURL,
		SettingKeyGoogleOAuthEnabled,
		SettingKeyGoogleOAuthClientID,
		SettingKeyGoogleOAuthClientSecret,
		SettingKeyGoogleOAuthRedirectURL,
		SettingKeyGoogleOAuthFrontendRedirectURL,
	}
	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return config.EmailOAuthProviderConfig{}, fmt.Errorf("get email oauth settings: %w", err)
	}
	cfg := s.effectiveEmailOAuthConfig(settings, provider)
	if !cfg.Enabled {
		return config.EmailOAuthProviderConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "oauth login is disabled")
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return config.EmailOAuthProviderConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth client id not configured")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return config.EmailOAuthProviderConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth client secret not configured")
	}
	for label, rawURL := range map[string]string{
		"authorize": cfg.AuthorizeURL,
		"token":     cfg.TokenURL,
		"userinfo":  cfg.UserInfoURL,
		"redirect":  cfg.RedirectURL,
	} {
		if strings.TrimSpace(rawURL) == "" {
			return config.EmailOAuthProviderConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth "+label+" url not configured")
		}
		if err := config.ValidateAbsoluteHTTPURL(rawURL); err != nil {
			return config.EmailOAuthProviderConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth "+label+" url invalid")
		}
	}
	if strings.TrimSpace(cfg.EmailsURL) != "" {
		if err := config.ValidateAbsoluteHTTPURL(cfg.EmailsURL); err != nil {
			return config.EmailOAuthProviderConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth emails url invalid")
		}
	}
	if err := config.ValidateFrontendRedirectURL(cfg.FrontendRedirectURL); err != nil {
		return config.EmailOAuthProviderConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth frontend redirect url invalid")
	}
	return cfg, nil
}

// IsRegistrationEnabled 检查是否开放注册
func (s *SettingService) IsRegistrationEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyRegistrationEnabled)
	if err != nil {
		// 安全默认：如果设置不存在或查询出错，默认关闭注册
		return false
	}
	return value == "true"
}

// IsBackendModeEnabled checks if backend mode is enabled
// Uses in-process atomic.Value cache with 60s TTL, zero-lock hot path
func (s *SettingService) IsBackendModeEnabled(ctx context.Context) bool {
	if cached, ok := backendModeCache.Load().(*cachedBackendMode); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.value
		}
	}
	result, _, _ := backendModeSF.Do("backend_mode", func() (any, error) {
		if cached, ok := backendModeCache.Load().(*cachedBackendMode); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return cached.value, nil
			}
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), backendModeDBTimeout)
		defer cancel()
		value, err := s.settingRepo.GetValue(dbCtx, SettingKeyBackendModeEnabled)
		if err != nil {
			if errors.Is(err, ErrSettingNotFound) {
				// Setting not yet created (fresh install) - default to disabled with full TTL
				backendModeCache.Store(&cachedBackendMode{
					value:     false,
					expiresAt: time.Now().Add(backendModeCacheTTL).UnixNano(),
				})
				return false, nil
			}
			slog.Warn("failed to get backend_mode_enabled setting", "error", err)
			backendModeCache.Store(&cachedBackendMode{
				value:     false,
				expiresAt: time.Now().Add(backendModeErrorTTL).UnixNano(),
			})
			return false, nil
		}
		enabled := value == "true"
		backendModeCache.Store(&cachedBackendMode{
			value:     enabled,
			expiresAt: time.Now().Add(backendModeCacheTTL).UnixNano(),
		})
		return enabled, nil
	})
	if val, ok := result.(bool); ok {
		return val
	}
	return false
}

type gatewayForwardingSettingsResult struct {
	fp, mp, cch, cacheTTL1h, rewriteMessageCacheControl bool
}

func (s *SettingService) getGatewayForwardingSettingsCached(ctx context.Context) gatewayForwardingSettingsResult {
	if cached, ok := gatewayForwardingCache.Load().(*cachedGatewayForwardingSettings); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return gatewayForwardingSettingsResult{
				fp:                         cached.fingerprintUnification,
				mp:                         cached.metadataPassthrough,
				cch:                        cached.cchSigning,
				cacheTTL1h:                 cached.anthropicCacheTTL1hInjection,
				rewriteMessageCacheControl: cached.rewriteMessageCacheControl,
			}
		}
	}
	val, _, _ := gatewayForwardingSF.Do("gateway_forwarding", func() (any, error) {
		if cached, ok := gatewayForwardingCache.Load().(*cachedGatewayForwardingSettings); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return gatewayForwardingSettingsResult{
					fp:                         cached.fingerprintUnification,
					mp:                         cached.metadataPassthrough,
					cch:                        cached.cchSigning,
					cacheTTL1h:                 cached.anthropicCacheTTL1hInjection,
					rewriteMessageCacheControl: cached.rewriteMessageCacheControl,
				}, nil
			}
		}
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), gatewayForwardingDBTimeout)
		defer cancel()
		values, err := s.settingRepo.GetMultiple(dbCtx, []string{
			SettingKeyEnableFingerprintUnification,
			SettingKeyEnableMetadataPassthrough,
			SettingKeyEnableCCHSigning,
			SettingKeyEnableAnthropicCacheTTL1hInjection,
			SettingKeyRewriteMessageCacheControl,
		})
		if err != nil {
			slog.Warn("failed to get gateway forwarding settings", "error", err)
			gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{
				fingerprintUnification:       true,
				metadataPassthrough:          false,
				cchSigning:                   false,
				anthropicCacheTTL1hInjection: false,
				rewriteMessageCacheControl:   s.defaultRewriteMessageCacheControl(),
				expiresAt:                    time.Now().Add(gatewayForwardingErrorTTL).UnixNano(),
			})
			return gatewayForwardingSettingsResult{fp: true, rewriteMessageCacheControl: s.defaultRewriteMessageCacheControl()}, nil
		}
		fp := true
		if v, ok := values[SettingKeyEnableFingerprintUnification]; ok && v != "" {
			fp = v == "true"
		}
		mp := values[SettingKeyEnableMetadataPassthrough] == "true"
		cch := values[SettingKeyEnableCCHSigning] == "true"
		cacheTTL1h := values[SettingKeyEnableAnthropicCacheTTL1hInjection] == "true"
		rewriteMessageCacheControl := s.defaultRewriteMessageCacheControl()
		if v, ok := values[SettingKeyRewriteMessageCacheControl]; ok && v != "" {
			rewriteMessageCacheControl = v == "true"
		}
		gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{
			fingerprintUnification:       fp,
			metadataPassthrough:          mp,
			cchSigning:                   cch,
			anthropicCacheTTL1hInjection: cacheTTL1h,
			rewriteMessageCacheControl:   rewriteMessageCacheControl,
			expiresAt:                    time.Now().Add(gatewayForwardingCacheTTL).UnixNano(),
		})
		return gatewayForwardingSettingsResult{
			fp:                         fp,
			mp:                         mp,
			cch:                        cch,
			cacheTTL1h:                 cacheTTL1h,
			rewriteMessageCacheControl: rewriteMessageCacheControl,
		}, nil
	})
	if r, ok := val.(gatewayForwardingSettingsResult); ok {
		return r
	}
	return gatewayForwardingSettingsResult{fp: true}
}

// GetGatewayForwardingSettings returns cached gateway forwarding settings.
// Uses in-process atomic.Value cache with 60s TTL, zero-lock hot path.
// Returns (fingerprintUnification, metadataPassthrough, cchSigning).
func (s *SettingService) GetGatewayForwardingSettings(ctx context.Context) (fingerprintUnification, metadataPassthrough, cchSigning bool) {
	result := s.getGatewayForwardingSettingsCached(ctx)
	return result.fp, result.mp, result.cch
}

// IsAnthropicCacheTTL1hInjectionEnabled 检查是否对 Anthropic OAuth/SetupToken 请求体注入 1h cache_control ttl。
func (s *SettingService) IsAnthropicCacheTTL1hInjectionEnabled(ctx context.Context) bool {
	return s.getGatewayForwardingSettingsCached(ctx).cacheTTL1h
}

// IsRewriteMessageCacheControlEnabled 检查是否启用 messages cache_control 改写。
func (s *SettingService) IsRewriteMessageCacheControlEnabled(ctx context.Context) bool {
	return s.getGatewayForwardingSettingsCached(ctx).rewriteMessageCacheControl
}

// IsEmailVerifyEnabled 检查是否开启邮件验证
func (s *SettingService) IsEmailVerifyEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyEmailVerifyEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

// GetRegistrationEmailSuffixWhitelist returns normalized registration email suffix whitelist.
func (s *SettingService) GetRegistrationEmailSuffixWhitelist(ctx context.Context) []string {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyRegistrationEmailSuffixWhitelist)
	if err != nil {
		return []string{}
	}
	return ParseRegistrationEmailSuffixWhitelist(value)
}

// IsPromoCodeEnabled 检查是否启用优惠码功能
func (s *SettingService) IsPromoCodeEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyPromoCodeEnabled)
	if err != nil {
		return true // 默认启用
	}
	return value != "false"
}

// IsInvitationCodeEnabled 检查是否启用邀请码注册功能
func (s *SettingService) IsInvitationCodeEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyInvitationCodeEnabled)
	if err != nil {
		return false // 默认关闭
	}
	return value == "true"
}

// GetCustomMenuItemsRaw returns the raw JSON string of custom_menu_items setting.
func (s *SettingService) GetCustomMenuItemsRaw(ctx context.Context) string {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyCustomMenuItems)
	if err != nil {
		return "[]"
	}
	return value
}

// IsAffiliateEnabled 检查是否启用邀请返利功能（总开关）
func (s *SettingService) IsAffiliateEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyAffiliateEnabled)
	if err != nil {
		return false // 默认关闭
	}
	return value == "true"
}

// GetAffiliateRebateRatePercent 读取并 clamp 全局返利比例。
// 解析失败、缺失或越界都回退到 AffiliateRebateRateDefault — 该比例从不抛错，
// 调用方只关心一个可用的数值。
func (s *SettingService) GetAffiliateRebateRatePercent(ctx context.Context) float64 {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyAffiliateRebateRate)
	if err != nil {
		return AffiliateRebateRateDefault
	}
	rate, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(rate) || math.IsInf(rate, 0) {
		return AffiliateRebateRateDefault
	}
	return clampAffiliateRebateRate(rate)
}

// GetAffiliateRebateFreezeHours 返回返利冻结期（小时）。
// 返回 0 表示不冻结（向后兼容）。
func (s *SettingService) GetAffiliateRebateFreezeHours(ctx context.Context) int {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyAffiliateRebateFreezeHours)
	if err != nil {
		return AffiliateRebateFreezeHoursDefault
	}
	hours, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || hours < 0 {
		return AffiliateRebateFreezeHoursDefault
	}
	if hours > AffiliateRebateFreezeHoursMax {
		return AffiliateRebateFreezeHoursMax
	}
	return hours
}

// GetAffiliateRebateDurationDays 返回返利有效期（天）。
// 返回 0 表示永久有效。
func (s *SettingService) GetAffiliateRebateDurationDays(ctx context.Context) int {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyAffiliateRebateDurationDays)
	if err != nil {
		return AffiliateRebateDurationDaysDefault
	}
	days, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || days < 0 {
		return AffiliateRebateDurationDaysDefault
	}
	if days > AffiliateRebateDurationDaysMax {
		return AffiliateRebateDurationDaysMax
	}
	return days
}

// GetAffiliateRebatePerInviteeCap 返回单人返利上限。
// 返回 0 表示无上限。
func (s *SettingService) GetAffiliateRebatePerInviteeCap(ctx context.Context) float64 {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyAffiliateRebatePerInviteeCap)
	if err != nil {
		return AffiliateRebatePerInviteeCapDefault
	}
	cap, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || cap < 0 || math.IsNaN(cap) || math.IsInf(cap, 0) {
		return AffiliateRebatePerInviteeCapDefault
	}
	return cap
}

// IsPasswordResetEnabled 检查是否启用密码重置功能
// 要求：必须同时开启邮件验证
func (s *SettingService) IsPasswordResetEnabled(ctx context.Context) bool {
	// Password reset requires email verification to be enabled
	if !s.IsEmailVerifyEnabled(ctx) {
		return false
	}
	value, err := s.settingRepo.GetValue(ctx, SettingKeyPasswordResetEnabled)
	if err != nil {
		return false // 默认关闭
	}
	return value == "true"
}

// IsTotpEnabled 检查是否启用 TOTP 双因素认证功能
func (s *SettingService) IsTotpEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyTotpEnabled)
	if err != nil {
		return false // 默认关闭
	}
	return value == "true"
}

// IsTotpEncryptionKeyConfigured 检查 TOTP 加密密钥是否已手动配置
// 只有手动配置了密钥才允许在管理后台启用 TOTP 功能
func (s *SettingService) IsTotpEncryptionKeyConfigured() bool {
	return s.cfg.Totp.EncryptionKeyConfigured
}

// GetSiteName 获取网站名称
func (s *SettingService) GetSiteName(ctx context.Context) string {
	value, err := s.settingRepo.GetValue(ctx, SettingKeySiteName)
	if err != nil || value == "" {
		return "Sub2API"
	}
	return value
}

// GetDefaultConcurrency 获取默认并发量
func (s *SettingService) GetDefaultConcurrency(ctx context.Context) int {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyDefaultConcurrency)
	if err != nil {
		return s.cfg.Default.UserConcurrency
	}
	if v, err := strconv.Atoi(value); err == nil && v > 0 {
		return v
	}
	return s.cfg.Default.UserConcurrency
}

// GetDefaultBalance 获取默认余额
func (s *SettingService) GetDefaultBalance(ctx context.Context) float64 {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyDefaultBalance)
	if err != nil {
		return s.cfg.Default.UserBalance
	}
	if v, err := strconv.ParseFloat(value, 64); err == nil && v >= 0 {
		return v
	}
	return s.cfg.Default.UserBalance
}

// GetDefaultUserRPMLimit 获取新用户默认 RPM 限制（0 = 不限制）。未配置则返回 0。
func (s *SettingService) GetDefaultUserRPMLimit(ctx context.Context) int {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyDefaultUserRPMLimit)
	if err != nil || value == "" {
		return 0
	}
	if v, err := strconv.Atoi(value); err == nil && v >= 0 {
		return v
	}
	return 0
}

// GetDefaultSubscriptions 获取新用户默认订阅配置列表。
func (s *SettingService) GetDefaultSubscriptions(ctx context.Context) []DefaultSubscriptionSetting {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyDefaultSubscriptions)
	if err != nil {
		return nil
	}
	return parseDefaultSubscriptions(value)
}

func (s *SettingService) GetAuthSourceDefaultSettings(ctx context.Context) (*AuthSourceDefaultSettings, error) {
	keys := []string{
		SettingKeyAuthSourceDefaultEmailBalance,
		SettingKeyAuthSourceDefaultEmailConcurrency,
		SettingKeyAuthSourceDefaultEmailSubscriptions,
		SettingKeyAuthSourceDefaultEmailGrantOnSignup,
		SettingKeyAuthSourceDefaultEmailGrantOnFirstBind,
		SettingKeyAuthSourceDefaultLinuxDoBalance,
		SettingKeyAuthSourceDefaultLinuxDoConcurrency,
		SettingKeyAuthSourceDefaultLinuxDoSubscriptions,
		SettingKeyAuthSourceDefaultLinuxDoGrantOnSignup,
		SettingKeyAuthSourceDefaultLinuxDoGrantOnFirstBind,
		SettingKeyAuthSourceDefaultOIDCBalance,
		SettingKeyAuthSourceDefaultOIDCConcurrency,
		SettingKeyAuthSourceDefaultOIDCSubscriptions,
		SettingKeyAuthSourceDefaultOIDCGrantOnSignup,
		SettingKeyAuthSourceDefaultOIDCGrantOnFirstBind,
		SettingKeyAuthSourceDefaultWeChatBalance,
		SettingKeyAuthSourceDefaultWeChatConcurrency,
		SettingKeyAuthSourceDefaultWeChatSubscriptions,
		SettingKeyAuthSourceDefaultWeChatGrantOnSignup,
		SettingKeyAuthSourceDefaultWeChatGrantOnFirstBind,
		SettingKeyAuthSourceDefaultGitHubBalance,
		SettingKeyAuthSourceDefaultGitHubConcurrency,
		SettingKeyAuthSourceDefaultGitHubSubscriptions,
		SettingKeyAuthSourceDefaultGitHubGrantOnSignup,
		SettingKeyAuthSourceDefaultGitHubGrantOnFirstBind,
		SettingKeyAuthSourceDefaultGoogleBalance,
		SettingKeyAuthSourceDefaultGoogleConcurrency,
		SettingKeyAuthSourceDefaultGoogleSubscriptions,
		SettingKeyAuthSourceDefaultGoogleGrantOnSignup,
		SettingKeyAuthSourceDefaultGoogleGrantOnFirstBind,
		SettingKeyAuthSourceDefaultDingTalkBalance,
		SettingKeyAuthSourceDefaultDingTalkConcurrency,
		SettingKeyAuthSourceDefaultDingTalkSubscriptions,
		SettingKeyAuthSourceDefaultDingTalkGrantOnSignup,
		SettingKeyAuthSourceDefaultDingTalkGrantOnFirstBind,
		SettingKeyAuthSourcePlatformQuotas("email"),
		SettingKeyAuthSourcePlatformQuotas("linuxdo"),
		SettingKeyAuthSourcePlatformQuotas("oidc"),
		SettingKeyAuthSourcePlatformQuotas("wechat"),
		SettingKeyAuthSourcePlatformQuotas("github"),
		SettingKeyAuthSourcePlatformQuotas("google"),
		SettingKeyAuthSourcePlatformQuotas("dingtalk"),
		SettingKeyForceEmailOnThirdPartySignup,
	}

	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get auth source default settings: %w", err)
	}

	return &AuthSourceDefaultSettings{
		Email:                        parseProviderDefaultGrantSettings(settings, emailAuthSourceDefaultKeys),
		LinuxDo:                      parseProviderDefaultGrantSettings(settings, linuxDoAuthSourceDefaultKeys),
		OIDC:                         parseProviderDefaultGrantSettings(settings, oidcAuthSourceDefaultKeys),
		WeChat:                       parseProviderDefaultGrantSettings(settings, weChatAuthSourceDefaultKeys),
		GitHub:                       parseProviderDefaultGrantSettings(settings, gitHubAuthSourceDefaultKeys),
		Google:                       parseProviderDefaultGrantSettings(settings, googleAuthSourceDefaultKeys),
		DingTalk:                     parseProviderDefaultGrantSettings(settings, dingTalkAuthSourceDefaultKeys),
		ForceEmailOnThirdPartySignup: settings[SettingKeyForceEmailOnThirdPartySignup] == "true",
	}, nil
}

func (s *SettingService) ResolveAuthSourceGrantSettings(ctx context.Context, signupSource string, firstBind bool) (ProviderDefaultGrantSettings, bool, error) {
	result := ProviderDefaultGrantSettings{
		Balance:       s.GetDefaultBalance(ctx),
		Concurrency:   s.GetDefaultConcurrency(ctx),
		Subscriptions: s.GetDefaultSubscriptions(ctx),
	}

	defaults, err := s.GetAuthSourceDefaultSettings(ctx)
	if err != nil {
		return result, false, err
	}

	providerDefaults, ok := authSourceSignupSettings(defaults, signupSource)
	if !ok {
		return result, false, nil
	}

	enabled := providerDefaults.GrantOnSignup
	if firstBind {
		enabled = providerDefaults.GrantOnFirstBind
	}
	if !enabled {
		return result, false, nil
	}

	return mergeProviderDefaultGrantSettings(result, providerDefaults), true, nil
}

func (s *SettingService) UpdateAuthSourceDefaultSettings(ctx context.Context, settings *AuthSourceDefaultSettings) error {
	updates, err := s.buildAuthSourceDefaultUpdates(ctx, settings)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}

	if err := s.settingRepo.SetMultiple(ctx, updates); err != nil {
		return fmt.Errorf("update auth source default settings: %w", err)
	}
	return nil
}

// InitializeDefaultSettings 初始化默认设置
func (s *SettingService) InitializeDefaultSettings(ctx context.Context) error {
	// 检查是否已有设置
	_, err := s.settingRepo.GetValue(ctx, SettingKeyRegistrationEnabled)
	if err == nil {
		// 已有设置，不需要初始化
		return nil
	}
	if !errors.Is(err, ErrSettingNotFound) {
		return fmt.Errorf("check existing settings: %w", err)
	}

	oidcUsePKCEDefault := true
	oidcValidateIDTokenDefault := true
	if s != nil && s.cfg != nil {
		if s.cfg.OIDC.UsePKCEExplicit {
			oidcUsePKCEDefault = s.cfg.OIDC.UsePKCE
		}
		if s.cfg.OIDC.ValidateIDTokenExplicit {
			oidcValidateIDTokenDefault = s.cfg.OIDC.ValidateIDToken
		}
	}
	loginAgreementDocumentsJSON, err := marshalLoginAgreementDocuments(defaultLoginAgreementDocuments())
	if err != nil {
		return err
	}

	// 初始化默认设置
	defaults := map[string]string{
		SettingKeyRegistrationEnabled:                      "true",
		SettingKeyEmailVerifyEnabled:                       "false",
		SettingKeyRegistrationEmailSuffixWhitelist:         "[]",
		SettingKeyPromoCodeEnabled:                         "true", // 默认启用优惠码功能
		SettingKeyPasswordMinLength:                        strconv.Itoa(DefaultPasswordMinLength),
		SettingKeyLoginAgreementEnabled:                    "false",
		SettingKeyLoginAgreementMode:                       defaultLoginAgreementMode,
		SettingKeyLoginAgreementUpdatedAt:                  defaultLoginAgreementDate,
		SettingKeyLoginAgreementDocuments:                  loginAgreementDocumentsJSON,
		SettingKeySiteName:                                 "Sub2API",
		SettingKeySiteLogo:                                 "",
		SettingKeyDocsContentBasePath:                      defaultDocsContentBasePath,
		SettingKeyHomeShellConfig:                          defaultHomeShellConfig,
		SettingKeyHomeBusinessShellConfig:                  defaultHomeBusinessShellConfig,
		SettingKeyModelPlazaItems:                          "[]",
		SettingKeyImageWorkspaceModelConfig:                defaultImageWorkspaceModelConfig,
		SettingKeyModelPlazaShellConfig:                    defaultModelPlazaShellConfig,
		SettingKeyDocsShellConfig:                          defaultDocsShellConfig,
		SettingKeyLegalDocumentShellConfig:                 defaultLegalDocumentShellConfig,
		SettingKeyAPIKeysShellConfig:                       apiKeysShellConfigDefault(),
		SettingKeyKeyUsageShellConfig:                      defaultKeyUsageShellConfig,
		SettingKeyDashboardShellConfig:                     defaultDashboardShellConfig,
		SettingKeyUsageShellConfig:                         usageShellConfigDefault(),
		SettingKeyAPIGuideShellConfig:                      defaultAPIGuideShellConfig,
		SettingKeyAPITestShellConfig:                       defaultAPITestShellConfig,
		SettingKeyAvailableGroupsShellConfig:               defaultAvailableGroupsShellConfig,
		SettingKeyRedeemShellConfig:                        defaultRedeemShellConfig,
		SettingKeyAffiliateShellConfig:                     defaultAffiliateShellConfig,
		SettingKeyAvailableChannelsShellConfig:             defaultAvailableChannelsShellConfig,
		SettingKeyChannelStatusShellConfig:                 defaultChannelStatusShellConfig,
		SettingKeyCustomPageShellConfig:                    defaultCustomPageShellConfig,
		SettingKeyProfileShellConfig:                       profileShellConfigDefault(),
		SettingKeyAuthShellConfig:                          defaultAuthShellConfig,
		SettingKeyPurchaseSubscriptionEnabled:              "false",
		SettingKeyPurchaseSubscriptionURL:                  "",
		SettingKeyTableDefaultPageSize:                     "20",
		SettingKeyTablePageSizeOptions:                     "[10,20,50,100]",
		SettingKeyCustomMenuItems:                          "[]",
		SettingKeyCustomEndpoints:                          "[]",
		SettingKeyWebAppURL:                                "",
		SettingKeyWebAppName:                               "",
		SettingKeyWebAppDescription:                        "",
		SettingKeyWebAppLogo:                               "",
		SettingKeyWebAppFavicon:                            "",
		SettingKeyWebAppPreviewImage:                       "",
		SettingKeyWebTheme:                                 "",
		SettingKeyWebAppearance:                            "",
		SettingKeyWebDefaultLocale:                         "",
		SettingKeyPromptCasesTitle:                         "",
		SettingKeyPromptCasesDescription:                   "",
		SettingKeyPromptTemplatesTitle:                     "",
		SettingKeyPromptTemplatesDescription:               "",
		SettingKeyPromptCatalogShellConfig:                 defaultPromptCatalogShellConfig,
		SettingKeyWorkspaceShellConfig:                     defaultWorkspaceShellConfig,
		SettingKeyPricingTitle:                             "",
		SettingKeyPricingDescription:                       "",
		SettingKeyPricingShellConfig:                       defaultPricingShellConfig,
		SettingKeyPaymentShellConfig:                       defaultPaymentShellConfig,
		SettingKeyPricingCurrencySymbol:                    "¥",
		SettingKeyCreditsTitle:                             "",
		SettingKeyCreditsDescription:                       "",
		SettingKeyCreditsPurchaseLabel:                     "",
		SettingKeyCreditsBalanceLabel:                      "",
		SettingKeyCreditsPerBalance:                        "1",
		SettingKeyCreditsShellConfig:                       defaultCreditsShellConfig,
		SettingKeyWebLocaleDetectEnabled:                   "false",
		SettingKeyWebEmailAuthVisible:                      "true",
		SettingKeyWebGoogleAuthVisible:                     "false",
		SettingKeyWebGitHubAuthVisible:                     "false",
		SettingKeyWebGoogleAnalyticsID:                     "",
		SettingKeyWebClarityID:                             "",
		SettingKeyWebPlausibleDomain:                       "",
		SettingKeyWebPlausibleSrc:                          "",
		SettingKeyWebOpenPanelClientID:                     "",
		SettingKeyWebPublicIntegrationsEnabled:             "true",
		SettingKeyWebVercelAnalyticsEnabled:                "false",
		SettingKeyWebAdsenseCode:                           "",
		SettingKeyWebAffonsoEnabled:                        "false",
		SettingKeyWebAffonsoID:                             "",
		SettingKeyWebAffonsoCookieDuration:                 defaultWebAffonsoCookieDuration,
		SettingKeyWebPromoteKitEnabled:                     "false",
		SettingKeyWebPromoteKitID:                          "",
		SettingKeyWebCrispEnabled:                          "false",
		SettingKeyWebCrispWebsiteID:                        "",
		SettingKeyWebTawkEnabled:                           "false",
		SettingKeyWebTawkPropertyID:                        "",
		SettingKeyWebTawkWidgetID:                          "",
		SettingKeyWeChatConnectEnabled:                     "false",
		SettingKeyWeChatConnectOpenAppID:                   "",
		SettingKeyWeChatConnectOpenAppSecret:               "",
		SettingKeyWeChatConnectMPAppID:                     "",
		SettingKeyWeChatConnectMPAppSecret:                 "",
		SettingKeyWeChatConnectMobileAppID:                 "",
		SettingKeyWeChatConnectMobileAppSecret:             "",
		SettingKeyWeChatConnectOpenEnabled:                 "false",
		SettingKeyWeChatConnectMPEnabled:                   "false",
		SettingKeyWeChatConnectMobileEnabled:               "false",
		SettingKeyWeChatConnectMode:                        "open",
		SettingKeyWeChatConnectScopes:                      "snsapi_login",
		SettingKeyWeChatConnectRedirectURL:                 "",
		SettingKeyWeChatConnectFrontendRedirectURL:         defaultWeChatConnectFrontend,
		SettingKeyGitHubOAuthEnabled:                       "false",
		SettingKeyGitHubOAuthClientID:                      "",
		SettingKeyGitHubOAuthClientSecret:                  "",
		SettingKeyGitHubOAuthRedirectURL:                   "",
		SettingKeyGitHubOAuthFrontendRedirectURL:           defaultGitHubOAuthFrontend,
		SettingKeyGoogleOAuthEnabled:                       "false",
		SettingKeyGoogleOAuthClientID:                      "",
		SettingKeyGoogleOAuthClientSecret:                  "",
		SettingKeyGoogleOAuthRedirectURL:                   "",
		SettingKeyGoogleOAuthFrontendRedirectURL:           defaultGoogleOAuthFrontend,
		SettingKeyOIDCConnectEnabled:                       "false",
		SettingKeyOIDCConnectProviderName:                  "OIDC",
		SettingKeyOIDCConnectClientID:                      "",
		SettingKeyOIDCConnectClientSecret:                  "",
		SettingKeyOIDCConnectIssuerURL:                     "",
		SettingKeyOIDCConnectDiscoveryURL:                  "",
		SettingKeyOIDCConnectAuthorizeURL:                  "",
		SettingKeyOIDCConnectTokenURL:                      "",
		SettingKeyOIDCConnectUserInfoURL:                   "",
		SettingKeyOIDCConnectJWKSURL:                       "",
		SettingKeyOIDCConnectScopes:                        "openid email profile",
		SettingKeyOIDCConnectRedirectURL:                   "",
		SettingKeyOIDCConnectFrontendRedirectURL:           "/auth/oidc/callback",
		SettingKeyOIDCConnectTokenAuthMethod:               "client_secret_post",
		SettingKeyOIDCConnectUsePKCE:                       strconv.FormatBool(oidcUsePKCEDefault),
		SettingKeyOIDCConnectValidateIDToken:               strconv.FormatBool(oidcValidateIDTokenDefault),
		SettingKeyOIDCConnectAllowedSigningAlgs:            "RS256,ES256,PS256",
		SettingKeyOIDCConnectClockSkewSeconds:              "120",
		SettingKeyOIDCConnectRequireEmailVerified:          "false",
		SettingKeyOIDCConnectUserInfoEmailPath:             "",
		SettingKeyOIDCConnectUserInfoIDPath:                "",
		SettingKeyOIDCConnectUserInfoUsernamePath:          "",
		SettingKeyDefaultConcurrency:                       strconv.Itoa(s.cfg.Default.UserConcurrency),
		SettingKeyDefaultBalance:                           strconv.FormatFloat(s.cfg.Default.UserBalance, 'f', 8, 64),
		SettingKeyAffiliateRebateRate:                      strconv.FormatFloat(AffiliateRebateRateDefault, 'f', 8, 64),
		SettingKeyAffiliateRebateFreezeHours:               strconv.Itoa(AffiliateRebateFreezeHoursDefault),
		SettingKeyAffiliateRebateDurationDays:              strconv.Itoa(AffiliateRebateDurationDaysDefault),
		SettingKeyAffiliateRebatePerInviteeCap:             strconv.FormatFloat(AffiliateRebatePerInviteeCapDefault, 'f', 2, 64),
		SettingKeyDefaultUserRPMLimit:                      "0",
		SettingKeyDefaultSubscriptions:                     "[]",
		SettingKeyAuthSourceDefaultEmailBalance:            "0",
		SettingKeyAuthSourceDefaultEmailConcurrency:        "5",
		SettingKeyAuthSourceDefaultEmailSubscriptions:      "[]",
		SettingKeyAuthSourceDefaultEmailGrantOnSignup:      "false",
		SettingKeyAuthSourceDefaultEmailGrantOnFirstBind:   "false",
		SettingKeyAuthSourceDefaultLinuxDoBalance:          "0",
		SettingKeyAuthSourceDefaultLinuxDoConcurrency:      "5",
		SettingKeyAuthSourceDefaultLinuxDoSubscriptions:    "[]",
		SettingKeyAuthSourceDefaultLinuxDoGrantOnSignup:    "false",
		SettingKeyAuthSourceDefaultLinuxDoGrantOnFirstBind: "false",
		SettingKeyAuthSourceDefaultOIDCBalance:             "0",
		SettingKeyAuthSourceDefaultOIDCConcurrency:         "5",
		SettingKeyAuthSourceDefaultOIDCSubscriptions:       "[]",
		SettingKeyAuthSourceDefaultOIDCGrantOnSignup:       "false",
		SettingKeyAuthSourceDefaultOIDCGrantOnFirstBind:    "false",
		SettingKeyAuthSourceDefaultWeChatBalance:           "0",
		SettingKeyAuthSourceDefaultWeChatConcurrency:       "5",
		SettingKeyAuthSourceDefaultWeChatSubscriptions:     "[]",
		SettingKeyAuthSourceDefaultWeChatGrantOnSignup:     "false",
		SettingKeyAuthSourceDefaultWeChatGrantOnFirstBind:  "false",
		SettingKeyAuthSourceDefaultGitHubBalance:           "0",
		SettingKeyAuthSourceDefaultGitHubConcurrency:       "5",
		SettingKeyAuthSourceDefaultGitHubSubscriptions:     "[]",
		SettingKeyAuthSourceDefaultGitHubGrantOnSignup:     "false",
		SettingKeyAuthSourceDefaultGitHubGrantOnFirstBind:  "false",
		SettingKeyAuthSourceDefaultGoogleBalance:           "0",
		SettingKeyAuthSourceDefaultGoogleConcurrency:       "5",
		SettingKeyAuthSourceDefaultGoogleSubscriptions:     "[]",
		SettingKeyAuthSourceDefaultGoogleGrantOnSignup:     "false",
		SettingKeyAuthSourceDefaultGoogleGrantOnFirstBind:  "false",
		SettingKeyForceEmailOnThirdPartySignup:             "false",
		SettingKeySMTPPort:                                 "587",
		SettingKeySMTPUseTLS:                               "false",
		SettingKeySMTPDailyLimit:                           "0",
		SettingKeySMTPChannels:                             "[]",
		// Model fallback defaults
		SettingKeyEnableModelFallback:      "false",
		SettingKeyFallbackModelAnthropic:   "claude-3-5-sonnet-20241022",
		SettingKeyFallbackModelOpenAI:      "gpt-4o",
		SettingKeyFallbackModelGemini:      "gemini-2.5-pro",
		SettingKeyFallbackModelAntigravity: "gemini-2.5-pro",
		// Identity patch defaults
		SettingKeyEnableIdentityPatch: "true",
		SettingKeyIdentityPatchPrompt: "",

		// Ops monitoring defaults (vNext)
		SettingKeyOpsMonitoringEnabled:         "true",
		SettingKeyOpsRealtimeMonitoringEnabled: "true",
		SettingKeyOpsQueryModeDefault:          "auto",
		SettingKeyOpsMetricsIntervalSeconds:    "60",

		// Channel monitor defaults (enabled, 60s)
		SettingKeyChannelMonitorEnabled:                "true",
		SettingKeyChannelMonitorDefaultIntervalSeconds: "60",

		// Available channels feature (default disabled; opt-in)
		SettingKeyAvailableChannelsEnabled: "false",

		// Affiliate (邀请返利) feature (default disabled; opt-in)
		SettingKeyAffiliateEnabled: "false",

		// 风控中心功能（默认关闭，显式启用）
		SettingKeyRiskControlEnabled: "false",

		// Claude Code version check (default: empty = disabled)
		SettingKeyMinClaudeCodeVersion: "",
		SettingKeyMaxClaudeCodeVersion: "",

		// 分组隔离（默认不允许未分组 Key 调度）
		SettingKeyAllowUngroupedKeyScheduling:        "false",
		SettingKeyEnableAnthropicCacheTTL1hInjection: "false",
		SettingKeyRewriteMessageCacheControl:         strconv.FormatBool(s.defaultRewriteMessageCacheControl()),
		SettingKeyAntigravityUserAgentVersion:        "",
		SettingKeyOpenAICodexUserAgent:               "",
		SettingPaymentVisibleMethodAlipaySource:      "",
		SettingPaymentVisibleMethodWxpaySource:       "",
		SettingPaymentVisibleMethodAlipayEnabled:     "false",
		SettingPaymentVisibleMethodWxpayEnabled:      "false",
		openAIAdvancedSchedulerSettingKey:            "false",
		SettingKeyRegistrationNotifyEnabled:          "false",
		SettingKeyRegistrationNotifyProvider:         "",
		SettingKeyRegistrationNotifyWebhookURL:       "",
		SettingKeyRegistrationNotifySecret:           "",
	}

	return s.settingRepo.SetMultiple(ctx, defaults)
}

// parseSettings 解析设置到结构体
func (s *SettingService) parseSettings(settings map[string]string) *SystemSettings {
	emailVerifyEnabled := settings[SettingKeyEmailVerifyEnabled] == "true"
	loginAgreementDocuments := parseLoginAgreementDocuments(settings[SettingKeyLoginAgreementDocuments])
	loginAgreementUpdatedAt := strings.TrimSpace(settings[SettingKeyLoginAgreementUpdatedAt])
	if loginAgreementUpdatedAt == "" {
		loginAgreementUpdatedAt = defaultLoginAgreementDate
	}
	apiKeyACLTrustForwardedIP := false
	if value, ok := settings[SettingKeyAPIKeyACLTrustForwardedIP]; ok {
		apiKeyACLTrustForwardedIP = value == "true"
	} else if s != nil && s.cfg != nil {
		apiKeyACLTrustForwardedIP = s.cfg.Security.TrustForwardedIPForAPIKeyACL
	}
	result := &SystemSettings{
		RegistrationEnabled:              settings[SettingKeyRegistrationEnabled] == "true",
		EmailVerifyEnabled:               emailVerifyEnabled,
		RegistrationEmailSuffixWhitelist: ParseRegistrationEmailSuffixWhitelist(settings[SettingKeyRegistrationEmailSuffixWhitelist]),
		PromoCodeEnabled:                 settings[SettingKeyPromoCodeEnabled] != "false", // 默认启用
		PasswordResetEnabled:             emailVerifyEnabled && settings[SettingKeyPasswordResetEnabled] == "true",
		PasswordMinLength:                parsePasswordMinLength(settings[SettingKeyPasswordMinLength]),
		FrontendURL:                      settings[SettingKeyFrontendURL],
		InvitationCodeEnabled:            settings[SettingKeyInvitationCodeEnabled] == "true",
		TotpEnabled:                      settings[SettingKeyTotpEnabled] == "true",
		LoginAgreementEnabled:            settings[SettingKeyLoginAgreementEnabled] == "true",
		LoginAgreementMode:               normalizeLoginAgreementMode(settings[SettingKeyLoginAgreementMode]),
		LoginAgreementUpdatedAt:          loginAgreementUpdatedAt,
		LoginAgreementDocuments:          loginAgreementDocuments,
		SMTPHost:                         settings[SettingKeySMTPHost],
		SMTPUsername:                     settings[SettingKeySMTPUsername],
		SMTPFrom:                         settings[SettingKeySMTPFrom],
		SMTPFromName:                     settings[SettingKeySMTPFromName],
		SMTPUseTLS:                       settings[SettingKeySMTPUseTLS] == "true",
		SMTPDailyLimit:                   parseSMTPDailyLimit(settings[SettingKeySMTPDailyLimit]),
		SMTPChannels:                     parseSMTPChannels(settings[SettingKeySMTPChannels]),
		SMTPPasswordConfigured:           settings[SettingKeySMTPPassword] != "",
		TurnstileEnabled:                 settings[SettingKeyTurnstileEnabled] == "true",
		TurnstileSiteKey:                 settings[SettingKeyTurnstileSiteKey],
		TurnstileSecretKeyConfigured:     settings[SettingKeyTurnstileSecretKey] != "",
		APIKeyACLTrustForwardedIP:        apiKeyACLTrustForwardedIP,
		SiteName:                         s.getStringOrDefault(settings, SettingKeySiteName, "Sub2API"),
		SiteLogo:                         settings[SettingKeySiteLogo],
		SiteSubtitle:                     s.getStringOrDefault(settings, SettingKeySiteSubtitle, "Subscription to API Conversion Platform"),
		APIBaseURL:                       settings[SettingKeyAPIBaseURL],
		ContactInfo:                      settings[SettingKeyContactInfo],
		DocURL:                           settings[SettingKeyDocURL],
		DocsContentBasePath:              docsContentBasePathSetting(settings[SettingKeyDocsContentBasePath]),
		HomeContent:                      settings[SettingKeyHomeContent],
		HomeShellConfig:                  homeShellConfigSetting(settings[SettingKeyHomeShellConfig]),
		HomeBusinessShellConfig:          homeBusinessShellConfigSetting(settings[SettingKeyHomeBusinessShellConfig]),
		ModelPlazaItems:                  settings[SettingKeyModelPlazaItems],
		ImageWorkspaceModelConfig:        imageWorkspaceModelConfigSetting(settings[SettingKeyImageWorkspaceModelConfig]),
		ModelPlazaShellConfig:            modelPlazaShellConfigSetting(settings[SettingKeyModelPlazaShellConfig]),
		DocsShellConfig:                  docsShellConfigSetting(settings[SettingKeyDocsShellConfig]),
		LegalDocumentShellConfig:         legalDocumentShellConfigSetting(settings[SettingKeyLegalDocumentShellConfig]),
		APIKeysShellConfig:               apiKeysShellConfigSetting(settings[SettingKeyAPIKeysShellConfig]),
		KeyUsageShellConfig:              keyUsageShellConfigSetting(settings[SettingKeyKeyUsageShellConfig]),
		DashboardShellConfig:             dashboardShellConfigSetting(settings[SettingKeyDashboardShellConfig]),
		UsageShellConfig:                 usageShellConfigSetting(settings[SettingKeyUsageShellConfig]),
		APIGuideShellConfig:              apiGuideShellConfigSetting(settings[SettingKeyAPIGuideShellConfig]),
		APITestShellConfig:               apiTestShellConfigSetting(settings[SettingKeyAPITestShellConfig]),
		AvailableGroupsShellConfig:       availableGroupsShellConfigSetting(settings[SettingKeyAvailableGroupsShellConfig]),
		RedeemShellConfig:                redeemShellConfigSetting(settings[SettingKeyRedeemShellConfig]),
		AffiliateShellConfig:             affiliateShellConfigSetting(settings[SettingKeyAffiliateShellConfig]),
		AvailableChannelsShellConfig:     availableChannelsShellConfigSetting(settings[SettingKeyAvailableChannelsShellConfig]),
		ChannelStatusShellConfig:         channelStatusShellConfigSetting(settings[SettingKeyChannelStatusShellConfig]),
		CustomPageShellConfig:            customPageShellConfigSetting(settings[SettingKeyCustomPageShellConfig]),
		ProfileShellConfig:               profileShellConfigSetting(settings[SettingKeyProfileShellConfig]),
		AuthShellConfig:                  authShellConfigSetting(settings[SettingKeyAuthShellConfig]),
		HideCcsImportButton:              settings[SettingKeyHideCcsImportButton] == "true",
		PurchaseSubscriptionEnabled:      settings[SettingKeyPurchaseSubscriptionEnabled] == "true",
		PurchaseSubscriptionURL:          strings.TrimSpace(settings[SettingKeyPurchaseSubscriptionURL]),
		CustomMenuItems:                  settings[SettingKeyCustomMenuItems],
		CustomEndpoints:                  settings[SettingKeyCustomEndpoints],
		WebAppURL:                        strings.TrimSpace(settings[SettingKeyWebAppURL]),
		WebAppName:                       strings.TrimSpace(settings[SettingKeyWebAppName]),
		WebAppDescription:                strings.TrimSpace(settings[SettingKeyWebAppDescription]),
		WebAppLogo:                       strings.TrimSpace(settings[SettingKeyWebAppLogo]),
		WebAppFavicon:                    strings.TrimSpace(settings[SettingKeyWebAppFavicon]),
		WebAppPreviewImage:               strings.TrimSpace(settings[SettingKeyWebAppPreviewImage]),
		WebTheme:                         strings.TrimSpace(settings[SettingKeyWebTheme]),
		WebAppearance:                    strings.TrimSpace(settings[SettingKeyWebAppearance]),
		WebDefaultLocale:                 strings.TrimSpace(settings[SettingKeyWebDefaultLocale]),
		WebPromptCasesTitle:              strings.TrimSpace(settings[SettingKeyPromptCasesTitle]),
		WebPromptCasesDescription:        strings.TrimSpace(settings[SettingKeyPromptCasesDescription]),
		WebPromptTemplatesTitle:          strings.TrimSpace(settings[SettingKeyPromptTemplatesTitle]),
		WebPromptTemplatesDescription:    strings.TrimSpace(settings[SettingKeyPromptTemplatesDescription]),
		PromptCatalogShellConfig:         promptCatalogShellConfigSetting(settings[SettingKeyPromptCatalogShellConfig]),
		WebWorkspaceShellConfig:          workspaceShellConfigSetting(settings[SettingKeyWorkspaceShellConfig]),
		ImagePromptFilterConfig:          strings.TrimSpace(settings[SettingKeyImagePromptFilterConfig]),
		WebPricingTitle:                  strings.TrimSpace(settings[SettingKeyPricingTitle]),
		WebPricingDescription:            strings.TrimSpace(settings[SettingKeyPricingDescription]),
		WebPricingShellConfig:            pricingShellConfigSetting(settings[SettingKeyPricingShellConfig]),
		WebPaymentShellConfig:            paymentShellConfigSetting(settings[SettingKeyPaymentShellConfig]),
		WebPricingCurrencySymbol:         pricingCurrencySymbolSetting(settings[SettingKeyPricingCurrencySymbol]),
		WebCreditsTitle:                  strings.TrimSpace(settings[SettingKeyCreditsTitle]),
		WebCreditsDescription:            strings.TrimSpace(settings[SettingKeyCreditsDescription]),
		WebCreditsPurchaseLabel:          strings.TrimSpace(settings[SettingKeyCreditsPurchaseLabel]),
		WebCreditsBalanceLabel:           strings.TrimSpace(settings[SettingKeyCreditsBalanceLabel]),
		WebCreditsPerBalance:             creditsPerBalanceSetting(settings[SettingKeyCreditsPerBalance]),
		CreditsShellConfig:               creditsShellConfigSetting(settings[SettingKeyCreditsShellConfig]),
		WebLocaleDetectEnabled:           settings[SettingKeyWebLocaleDetectEnabled] == "true",
		WebEmailAuthVisible:              parseBoolSettingWithDefault(settings[SettingKeyWebEmailAuthVisible], true),
		WebGoogleAuthVisible:             settings[SettingKeyWebGoogleAuthVisible] == "true",
		WebGitHubAuthVisible:             settings[SettingKeyWebGitHubAuthVisible] == "true",
		WebGoogleAnalyticsID:             strings.TrimSpace(settings[SettingKeyWebGoogleAnalyticsID]),
		WebClarityID:                     strings.TrimSpace(settings[SettingKeyWebClarityID]),
		WebPlausibleDomain:               strings.TrimSpace(settings[SettingKeyWebPlausibleDomain]),
		WebPlausibleSrc:                  strings.TrimSpace(settings[SettingKeyWebPlausibleSrc]),
		WebOpenPanelClientID:             strings.TrimSpace(settings[SettingKeyWebOpenPanelClientID]),
		WebPublicIntegrationsEnabled:     !isFalseSettingValue(settings[SettingKeyWebPublicIntegrationsEnabled]),
		WebVercelAnalyticsEnabled:        settings[SettingKeyWebVercelAnalyticsEnabled] == "true",
		WebAdsenseCode:                   strings.TrimSpace(settings[SettingKeyWebAdsenseCode]),
		WebAffonsoEnabled:                settings[SettingKeyWebAffonsoEnabled] == "true",
		WebAffonsoID:                     strings.TrimSpace(settings[SettingKeyWebAffonsoID]),
		WebAffonsoCookieDuration:         webAffonsoCookieDurationSetting(settings[SettingKeyWebAffonsoCookieDuration]),
		WebPromoteKitEnabled:             settings[SettingKeyWebPromoteKitEnabled] == "true",
		WebPromoteKitID:                  strings.TrimSpace(settings[SettingKeyWebPromoteKitID]),
		WebCrispEnabled:                  settings[SettingKeyWebCrispEnabled] == "true",
		WebCrispWebsiteID:                strings.TrimSpace(settings[SettingKeyWebCrispWebsiteID]),
		WebTawkEnabled:                   settings[SettingKeyWebTawkEnabled] == "true",
		WebTawkPropertyID:                strings.TrimSpace(settings[SettingKeyWebTawkPropertyID]),
		WebTawkWidgetID:                  strings.TrimSpace(settings[SettingKeyWebTawkWidgetID]),
		BackendModeEnabled:               settings[SettingKeyBackendModeEnabled] == "true",
	}
	result.TableDefaultPageSize, result.TablePageSizeOptions = parseTablePreferences(
		settings[SettingKeyTableDefaultPageSize],
		settings[SettingKeyTablePageSizeOptions],
	)

	// 解析整数类型
	if port, err := strconv.Atoi(settings[SettingKeySMTPPort]); err == nil {
		result.SMTPPort = port
	} else {
		result.SMTPPort = 587
	}

	if concurrency, err := strconv.Atoi(settings[SettingKeyDefaultConcurrency]); err == nil {
		result.DefaultConcurrency = concurrency
	} else {
		result.DefaultConcurrency = s.cfg.Default.UserConcurrency
	}

	if rpm, err := strconv.Atoi(settings[SettingKeyDefaultUserRPMLimit]); err == nil && rpm >= 0 {
		result.DefaultUserRPMLimit = rpm
	}

	// 解析浮点数类型
	if balance, err := strconv.ParseFloat(settings[SettingKeyDefaultBalance], 64); err == nil {
		result.DefaultBalance = balance
	} else {
		result.DefaultBalance = s.cfg.Default.UserBalance
	}
	if rebateRate, err := strconv.ParseFloat(settings[SettingKeyAffiliateRebateRate], 64); err == nil {
		result.AffiliateRebateRate = clampAffiliateRebateRate(rebateRate)
	} else {
		result.AffiliateRebateRate = AffiliateRebateRateDefault
	}
	if freezeHours, err := strconv.Atoi(settings[SettingKeyAffiliateRebateFreezeHours]); err == nil && freezeHours >= 0 {
		if freezeHours > AffiliateRebateFreezeHoursMax {
			freezeHours = AffiliateRebateFreezeHoursMax
		}
		result.AffiliateRebateFreezeHours = freezeHours
	}
	if durationDays, err := strconv.Atoi(settings[SettingKeyAffiliateRebateDurationDays]); err == nil && durationDays >= 0 {
		if durationDays > AffiliateRebateDurationDaysMax {
			durationDays = AffiliateRebateDurationDaysMax
		}
		result.AffiliateRebateDurationDays = durationDays
	}
	if perInviteeCap, err := strconv.ParseFloat(settings[SettingKeyAffiliateRebatePerInviteeCap], 64); err == nil && perInviteeCap >= 0 {
		result.AffiliateRebatePerInviteeCap = perInviteeCap
	}
	result.DefaultSubscriptions = parseDefaultSubscriptions(settings[SettingKeyDefaultSubscriptions])

	// 敏感信息直接返回，方便测试连接时使用
	result.SMTPPassword = settings[SettingKeySMTPPassword]
	result.TurnstileSecretKey = settings[SettingKeyTurnstileSecretKey]

	// LinuxDo Connect 设置：
	// - 兼容 config.yaml/env（避免老部署因为未迁移到数据库设置而被意外关闭）
	// - 支持在后台“系统设置”中覆盖并持久化（存储于 DB）
	linuxDoBase := config.LinuxDoConnectConfig{}
	if s.cfg != nil {
		linuxDoBase = s.cfg.LinuxDo
	}

	if raw, ok := settings[SettingKeyLinuxDoConnectEnabled]; ok {
		result.LinuxDoConnectEnabled = raw == "true"
	} else {
		result.LinuxDoConnectEnabled = linuxDoBase.Enabled
	}

	if v, ok := settings[SettingKeyLinuxDoConnectClientID]; ok && strings.TrimSpace(v) != "" {
		result.LinuxDoConnectClientID = strings.TrimSpace(v)
	} else {
		result.LinuxDoConnectClientID = linuxDoBase.ClientID
	}

	if v, ok := settings[SettingKeyLinuxDoConnectRedirectURL]; ok && strings.TrimSpace(v) != "" {
		result.LinuxDoConnectRedirectURL = strings.TrimSpace(v)
	} else {
		result.LinuxDoConnectRedirectURL = linuxDoBase.RedirectURL
	}

	result.LinuxDoConnectClientSecret = strings.TrimSpace(settings[SettingKeyLinuxDoConnectClientSecret])
	if result.LinuxDoConnectClientSecret == "" {
		result.LinuxDoConnectClientSecret = strings.TrimSpace(linuxDoBase.ClientSecret)
	}
	result.LinuxDoConnectClientSecretConfigured = result.LinuxDoConnectClientSecret != ""

	// DingTalk Connect 设置：
	// - 兼容 config.yaml/env
	// - 支持后台系统设置覆盖并持久化（存储于 DB）
	dingTalkBase := config.DingTalkConnectConfig{}
	if s.cfg != nil {
		dingTalkBase = s.cfg.DingTalk
	}

	if raw, ok := settings[SettingKeyDingTalkConnectEnabled]; ok {
		result.DingTalkConnectEnabled = raw == "true"
	} else {
		result.DingTalkConnectEnabled = dingTalkBase.Enabled
	}

	if v, ok := settings[SettingKeyDingTalkConnectClientID]; ok && strings.TrimSpace(v) != "" {
		result.DingTalkConnectClientID = strings.TrimSpace(v)
	} else {
		result.DingTalkConnectClientID = dingTalkBase.ClientID
	}

	if v, ok := settings[SettingKeyDingTalkConnectRedirectURL]; ok && strings.TrimSpace(v) != "" {
		result.DingTalkConnectRedirectURL = strings.TrimSpace(v)
	} else {
		result.DingTalkConnectRedirectURL = dingTalkBase.RedirectURL
	}

	result.DingTalkConnectClientSecret = strings.TrimSpace(settings[SettingKeyDingTalkConnectClientSecret])
	if result.DingTalkConnectClientSecret == "" {
		result.DingTalkConnectClientSecret = strings.TrimSpace(dingTalkBase.ClientSecret)
	}
	result.DingTalkConnectClientSecretConfigured = result.DingTalkConnectClientSecret != ""

	if v, ok := settings[SettingKeyDingTalkConnectCorpRestrictionPolicy]; ok && strings.TrimSpace(v) != "" {
		result.DingTalkConnectCorpRestrictionPolicy = strings.TrimSpace(v)
	} else {
		result.DingTalkConnectCorpRestrictionPolicy = dingTalkBase.CorpRestrictionPolicy
	}
	result.DingTalkConnectCorpRestrictionPolicy = coerceDeprecatedDingTalkCorpPolicy(result.DingTalkConnectCorpRestrictionPolicy)

	if v, ok := settings[SettingKeyDingTalkConnectInternalCorpID]; ok && strings.TrimSpace(v) != "" {
		result.DingTalkConnectInternalCorpID = strings.TrimSpace(v)
	} else {
		result.DingTalkConnectInternalCorpID = dingTalkBase.InternalCorpID
	}

	if v, ok := settings[SettingKeyDingTalkConnectBypassRegistration]; ok && strings.TrimSpace(v) != "" {
		result.DingTalkConnectBypassRegistration = strings.EqualFold(strings.TrimSpace(v), "true")
	} else {
		result.DingTalkConnectBypassRegistration = dingTalkBase.BypassRegistration
	}
	// bypass_registration 仅在 internal_only 模式下有意义；其它策略下强制 false，
	// 以保证加载出的 effective config 永远是一致状态。
	if result.DingTalkConnectCorpRestrictionPolicy != "internal_only" {
		result.DingTalkConnectBypassRegistration = false
	}

	if v, ok := settings[SettingKeyDingTalkConnectSyncCorpEmail]; ok && strings.TrimSpace(v) != "" {
		result.DingTalkConnectSyncCorpEmail = strings.EqualFold(strings.TrimSpace(v), "true")
	} else {
		result.DingTalkConnectSyncCorpEmail = dingTalkBase.SyncCorpEmail
	}
	if v, ok := settings[SettingKeyDingTalkConnectSyncDisplayName]; ok && strings.TrimSpace(v) != "" {
		result.DingTalkConnectSyncDisplayName = strings.EqualFold(strings.TrimSpace(v), "true")
	} else {
		result.DingTalkConnectSyncDisplayName = dingTalkBase.SyncDisplayName
	}
	if v, ok := settings[SettingKeyDingTalkConnectSyncDept]; ok && strings.TrimSpace(v) != "" {
		result.DingTalkConnectSyncDept = strings.EqualFold(strings.TrimSpace(v), "true")
	} else {
		result.DingTalkConnectSyncDept = dingTalkBase.SyncDept
	}
	// 身份同步三开关仅在 internal_only 模式下有意义；其它策略强制 false。
	if result.DingTalkConnectCorpRestrictionPolicy != "internal_only" {
		result.DingTalkConnectSyncCorpEmail = false
		result.DingTalkConnectSyncDisplayName = false
		result.DingTalkConnectSyncDept = false
	}

	// 身份同步目标 attr key（DB 空 → fallback 默认值）
	result.DingTalkConnectSyncCorpEmailAttrKey = strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncCorpEmailAttrKey])
	if result.DingTalkConnectSyncCorpEmailAttrKey == "" {
		if v := strings.TrimSpace(dingTalkBase.SyncCorpEmailAttrKey); v != "" {
			result.DingTalkConnectSyncCorpEmailAttrKey = v
		} else {
			result.DingTalkConnectSyncCorpEmailAttrKey = "dingtalk_email"
		}
	}
	result.DingTalkConnectSyncDisplayNameAttrKey = strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncDisplayNameAttrKey])
	if result.DingTalkConnectSyncDisplayNameAttrKey == "" {
		if v := strings.TrimSpace(dingTalkBase.SyncDisplayNameAttrKey); v != "" {
			result.DingTalkConnectSyncDisplayNameAttrKey = v
		} else {
			result.DingTalkConnectSyncDisplayNameAttrKey = "dingtalk_name"
		}
	}
	result.DingTalkConnectSyncDeptAttrKey = strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncDeptAttrKey])
	if result.DingTalkConnectSyncDeptAttrKey == "" {
		if v := strings.TrimSpace(dingTalkBase.SyncDeptAttrKey); v != "" {
			result.DingTalkConnectSyncDeptAttrKey = v
		} else {
			result.DingTalkConnectSyncDeptAttrKey = "dingtalk_department"
		}
	}

	// 身份同步目标 attr 显示名称（DB 空 → fallback 默认中文）
	result.DingTalkConnectSyncCorpEmailAttrName = strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncCorpEmailAttrName])
	if result.DingTalkConnectSyncCorpEmailAttrName == "" {
		if v := strings.TrimSpace(dingTalkBase.SyncCorpEmailAttrName); v != "" {
			result.DingTalkConnectSyncCorpEmailAttrName = v
		} else {
			result.DingTalkConnectSyncCorpEmailAttrName = "钉钉企业邮箱"
		}
	}
	result.DingTalkConnectSyncDisplayNameAttrName = strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncDisplayNameAttrName])
	if result.DingTalkConnectSyncDisplayNameAttrName == "" {
		if v := strings.TrimSpace(dingTalkBase.SyncDisplayNameAttrName); v != "" {
			result.DingTalkConnectSyncDisplayNameAttrName = v
		} else {
			result.DingTalkConnectSyncDisplayNameAttrName = "钉钉姓名"
		}
	}
	result.DingTalkConnectSyncDeptAttrName = strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncDeptAttrName])
	if result.DingTalkConnectSyncDeptAttrName == "" {
		if v := strings.TrimSpace(dingTalkBase.SyncDeptAttrName); v != "" {
			result.DingTalkConnectSyncDeptAttrName = v
		} else {
			result.DingTalkConnectSyncDeptAttrName = "钉钉部门"
		}
	}

	// Generic OIDC 设置：
	// - 兼容 config.yaml/env
	// - 支持后台系统设置覆盖并持久化（存储于 DB）
	oidcBase := config.OIDCConnectConfig{}
	if s.cfg != nil {
		oidcBase = s.cfg.OIDC
	}

	if raw, ok := settings[SettingKeyOIDCConnectEnabled]; ok {
		result.OIDCConnectEnabled = raw == "true"
	} else {
		result.OIDCConnectEnabled = oidcBase.Enabled
	}

	if v, ok := settings[SettingKeyOIDCConnectProviderName]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectProviderName = strings.TrimSpace(v)
	} else {
		result.OIDCConnectProviderName = strings.TrimSpace(oidcBase.ProviderName)
	}
	if result.OIDCConnectProviderName == "" {
		result.OIDCConnectProviderName = "OIDC"
	}

	if v, ok := settings[SettingKeyOIDCConnectClientID]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectClientID = strings.TrimSpace(v)
	} else {
		result.OIDCConnectClientID = strings.TrimSpace(oidcBase.ClientID)
	}
	if v, ok := settings[SettingKeyOIDCConnectIssuerURL]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectIssuerURL = strings.TrimSpace(v)
	} else {
		result.OIDCConnectIssuerURL = strings.TrimSpace(oidcBase.IssuerURL)
	}
	if v, ok := settings[SettingKeyOIDCConnectDiscoveryURL]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectDiscoveryURL = strings.TrimSpace(v)
	} else {
		result.OIDCConnectDiscoveryURL = strings.TrimSpace(oidcBase.DiscoveryURL)
	}
	if v, ok := settings[SettingKeyOIDCConnectAuthorizeURL]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectAuthorizeURL = strings.TrimSpace(v)
	} else {
		result.OIDCConnectAuthorizeURL = strings.TrimSpace(oidcBase.AuthorizeURL)
	}
	if v, ok := settings[SettingKeyOIDCConnectTokenURL]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectTokenURL = strings.TrimSpace(v)
	} else {
		result.OIDCConnectTokenURL = strings.TrimSpace(oidcBase.TokenURL)
	}
	if v, ok := settings[SettingKeyOIDCConnectUserInfoURL]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectUserInfoURL = strings.TrimSpace(v)
	} else {
		result.OIDCConnectUserInfoURL = strings.TrimSpace(oidcBase.UserInfoURL)
	}
	if v, ok := settings[SettingKeyOIDCConnectJWKSURL]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectJWKSURL = strings.TrimSpace(v)
	} else {
		result.OIDCConnectJWKSURL = strings.TrimSpace(oidcBase.JWKSURL)
	}
	if v, ok := settings[SettingKeyOIDCConnectScopes]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectScopes = strings.TrimSpace(v)
	} else {
		result.OIDCConnectScopes = strings.TrimSpace(oidcBase.Scopes)
	}
	if v, ok := settings[SettingKeyOIDCConnectRedirectURL]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectRedirectURL = strings.TrimSpace(v)
	} else {
		result.OIDCConnectRedirectURL = strings.TrimSpace(oidcBase.RedirectURL)
	}
	if v, ok := settings[SettingKeyOIDCConnectFrontendRedirectURL]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectFrontendRedirectURL = strings.TrimSpace(v)
	} else {
		result.OIDCConnectFrontendRedirectURL = strings.TrimSpace(oidcBase.FrontendRedirectURL)
	}
	if v, ok := settings[SettingKeyOIDCConnectTokenAuthMethod]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectTokenAuthMethod = strings.ToLower(strings.TrimSpace(v))
	} else {
		result.OIDCConnectTokenAuthMethod = strings.ToLower(strings.TrimSpace(oidcBase.TokenAuthMethod))
	}
	if raw, ok := settings[SettingKeyOIDCConnectUsePKCE]; ok {
		result.OIDCConnectUsePKCE = raw == "true"
	} else {
		result.OIDCConnectUsePKCE = oidcUsePKCECompatibilityDefault(oidcBase)
	}
	if raw, ok := settings[SettingKeyOIDCConnectValidateIDToken]; ok {
		result.OIDCConnectValidateIDToken = raw == "true"
	} else {
		result.OIDCConnectValidateIDToken = oidcValidateIDTokenCompatibilityDefault(oidcBase)
	}
	if v, ok := settings[SettingKeyOIDCConnectAllowedSigningAlgs]; ok && strings.TrimSpace(v) != "" {
		result.OIDCConnectAllowedSigningAlgs = strings.TrimSpace(v)
	} else {
		result.OIDCConnectAllowedSigningAlgs = strings.TrimSpace(oidcBase.AllowedSigningAlgs)
	}
	clockSkewSet := false
	if raw, ok := settings[SettingKeyOIDCConnectClockSkewSeconds]; ok && strings.TrimSpace(raw) != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
			result.OIDCConnectClockSkewSeconds = parsed
			clockSkewSet = true
		}
	}
	if !clockSkewSet {
		result.OIDCConnectClockSkewSeconds = oidcBase.ClockSkewSeconds
	}
	if !clockSkewSet && result.OIDCConnectClockSkewSeconds == 0 {
		result.OIDCConnectClockSkewSeconds = 120
	}
	if raw, ok := settings[SettingKeyOIDCConnectRequireEmailVerified]; ok {
		result.OIDCConnectRequireEmailVerified = raw == "true"
	} else {
		result.OIDCConnectRequireEmailVerified = oidcBase.RequireEmailVerified
	}
	if v, ok := settings[SettingKeyOIDCConnectUserInfoEmailPath]; ok {
		result.OIDCConnectUserInfoEmailPath = strings.TrimSpace(v)
	} else {
		result.OIDCConnectUserInfoEmailPath = strings.TrimSpace(oidcBase.UserInfoEmailPath)
	}
	if v, ok := settings[SettingKeyOIDCConnectUserInfoIDPath]; ok {
		result.OIDCConnectUserInfoIDPath = strings.TrimSpace(v)
	} else {
		result.OIDCConnectUserInfoIDPath = strings.TrimSpace(oidcBase.UserInfoIDPath)
	}
	if v, ok := settings[SettingKeyOIDCConnectUserInfoUsernamePath]; ok {
		result.OIDCConnectUserInfoUsernamePath = strings.TrimSpace(v)
	} else {
		result.OIDCConnectUserInfoUsernamePath = strings.TrimSpace(oidcBase.UserInfoUsernamePath)
	}
	result.OIDCConnectClientSecret = strings.TrimSpace(settings[SettingKeyOIDCConnectClientSecret])
	if result.OIDCConnectClientSecret == "" {
		result.OIDCConnectClientSecret = strings.TrimSpace(oidcBase.ClientSecret)
	}
	result.OIDCConnectClientSecretConfigured = result.OIDCConnectClientSecret != ""

	gitHubEffective := s.effectiveEmailOAuthConfig(settings, "github")
	result.GitHubOAuthEnabled = gitHubEffective.Enabled
	result.GitHubOAuthClientID = strings.TrimSpace(gitHubEffective.ClientID)
	result.GitHubOAuthClientSecret = strings.TrimSpace(gitHubEffective.ClientSecret)
	result.GitHubOAuthClientSecretConfigured = result.GitHubOAuthClientSecret != ""
	result.GitHubOAuthRedirectURL = strings.TrimSpace(gitHubEffective.RedirectURL)
	result.GitHubOAuthFrontendRedirectURL = strings.TrimSpace(gitHubEffective.FrontendRedirectURL)

	googleEffective := s.effectiveEmailOAuthConfig(settings, "google")
	result.GoogleOAuthEnabled = googleEffective.Enabled
	result.GoogleOAuthClientID = strings.TrimSpace(googleEffective.ClientID)
	result.GoogleOAuthClientSecret = strings.TrimSpace(googleEffective.ClientSecret)
	result.GoogleOAuthClientSecretConfigured = result.GoogleOAuthClientSecret != ""
	result.GoogleOAuthRedirectURL = strings.TrimSpace(googleEffective.RedirectURL)
	result.GoogleOAuthFrontendRedirectURL = strings.TrimSpace(googleEffective.FrontendRedirectURL)

	// WeChat Connect 设置：
	// - 优先读取 DB 系统设置
	// - 缺失时回退到 config/env，保持升级兼容
	weChatEffective := s.effectiveWeChatConnectOAuthConfig(settings)
	result.WeChatConnectEnabled = weChatEffective.Enabled
	result.WeChatConnectOpenAppID = weChatEffective.OpenAppID
	result.WeChatConnectOpenAppSecret = weChatEffective.OpenAppSecret
	result.WeChatConnectOpenAppSecretConfigured = weChatEffective.OpenAppSecret != ""
	result.WeChatConnectMPAppID = weChatEffective.MPAppID
	result.WeChatConnectMPAppSecret = weChatEffective.MPAppSecret
	result.WeChatConnectMPAppSecretConfigured = weChatEffective.MPAppSecret != ""
	result.WeChatConnectMobileAppID = weChatEffective.MobileAppID
	result.WeChatConnectMobileAppSecret = weChatEffective.MobileAppSecret
	result.WeChatConnectMobileAppSecretConfigured = weChatEffective.MobileAppSecret != ""
	result.WeChatConnectOpenEnabled = weChatEffective.OpenEnabled
	result.WeChatConnectMPEnabled = weChatEffective.MPEnabled
	result.WeChatConnectMobileEnabled = weChatEffective.MobileEnabled
	result.WeChatConnectMode = weChatEffective.Mode
	result.WeChatConnectScopes = weChatEffective.Scopes
	result.WeChatConnectRedirectURL = weChatEffective.RedirectURL
	result.WeChatConnectFrontendRedirectURL = weChatEffective.FrontendRedirectURL

	// Model fallback settings
	result.EnableModelFallback = settings[SettingKeyEnableModelFallback] == "true"
	result.FallbackModelAnthropic = s.getStringOrDefault(settings, SettingKeyFallbackModelAnthropic, "claude-3-5-sonnet-20241022")
	result.FallbackModelOpenAI = s.getStringOrDefault(settings, SettingKeyFallbackModelOpenAI, "gpt-4o")
	result.FallbackModelGemini = s.getStringOrDefault(settings, SettingKeyFallbackModelGemini, "gemini-2.5-pro")
	result.FallbackModelAntigravity = s.getStringOrDefault(settings, SettingKeyFallbackModelAntigravity, "gemini-2.5-pro")

	// Identity patch settings (default: enabled, to preserve existing behavior)
	if v, ok := settings[SettingKeyEnableIdentityPatch]; ok && v != "" {
		result.EnableIdentityPatch = v == "true"
	} else {
		result.EnableIdentityPatch = true
	}
	result.IdentityPatchPrompt = settings[SettingKeyIdentityPatchPrompt]

	// Ops monitoring settings (default: enabled, fail-open)
	result.OpsMonitoringEnabled = !isFalseSettingValue(settings[SettingKeyOpsMonitoringEnabled])
	result.OpsRealtimeMonitoringEnabled = !isFalseSettingValue(settings[SettingKeyOpsRealtimeMonitoringEnabled])
	result.OpsQueryModeDefault = string(ParseOpsQueryMode(settings[SettingKeyOpsQueryModeDefault]))
	result.OpsMetricsIntervalSeconds = 60
	if raw := strings.TrimSpace(settings[SettingKeyOpsMetricsIntervalSeconds]); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			if v < 60 {
				v = 60
			}
			if v > 3600 {
				v = 3600
			}
			result.OpsMetricsIntervalSeconds = v
		}
	}

	// Channel monitor feature (default: enabled, 60s)
	result.ChannelMonitorEnabled = !isFalseSettingValue(settings[SettingKeyChannelMonitorEnabled])
	result.ChannelMonitorDefaultIntervalSeconds = parseChannelMonitorInterval(
		settings[SettingKeyChannelMonitorDefaultIntervalSeconds],
	)

	// Available channels feature (default: disabled; strict true)
	result.AvailableChannelsEnabled = settings[SettingKeyAvailableChannelsEnabled] == "true"

	// Affiliate (邀请返利) feature (default: disabled; strict true)
	result.AffiliateEnabled = settings[SettingKeyAffiliateEnabled] == "true"

	// 风控中心功能（默认关闭，严格 true 才启用）
	result.RiskControlEnabled = settings[SettingKeyRiskControlEnabled] == "true"

	// Claude Code version check
	result.MinClaudeCodeVersion = settings[SettingKeyMinClaudeCodeVersion]
	result.MaxClaudeCodeVersion = settings[SettingKeyMaxClaudeCodeVersion]

	// 分组隔离
	result.AllowUngroupedKeyScheduling = settings[SettingKeyAllowUngroupedKeyScheduling] == "true"

	// Gateway forwarding behavior (defaults: fingerprint=true, metadata_passthrough=false, cch_signing=false)
	if v, ok := settings[SettingKeyEnableFingerprintUnification]; ok && v != "" {
		result.EnableFingerprintUnification = v == "true"
	} else {
		result.EnableFingerprintUnification = true // default: enabled (current behavior)
	}
	result.EnableMetadataPassthrough = settings[SettingKeyEnableMetadataPassthrough] == "true"
	result.EnableCCHSigning = settings[SettingKeyEnableCCHSigning] == "true"
	result.EnableAnthropicCacheTTL1hInjection = settings[SettingKeyEnableAnthropicCacheTTL1hInjection] == "true"
	if v, ok := settings[SettingKeyRewriteMessageCacheControl]; ok && v != "" {
		result.RewriteMessageCacheControl = v == "true"
	} else {
		result.RewriteMessageCacheControl = s.defaultRewriteMessageCacheControl()
	}
	result.AntigravityUserAgentVersion = antigravity.NormalizeUserAgentVersion(settings[SettingKeyAntigravityUserAgentVersion])
	result.OpenAICodexUserAgent = strings.TrimSpace(settings[SettingKeyOpenAICodexUserAgent])
	result.OpenAIAllowClaudeCodeCodexPlugin = settings[SettingKeyOpenAIAllowClaudeCodeCodexPlugin] == "true"

	// Web search emulation: quick enabled check from the JSON config
	if raw := settings[SettingKeyWebSearchEmulationConfig]; raw != "" {
		var wsCfg WebSearchEmulationConfig
		if err := json.Unmarshal([]byte(raw), &wsCfg); err == nil {
			result.WebSearchEmulationEnabled = wsCfg.Enabled && len(wsCfg.Providers) > 0
		}
	}
	result.PaymentVisibleMethodAlipaySource = NormalizeVisibleMethodSource("alipay", settings[SettingPaymentVisibleMethodAlipaySource])
	result.PaymentVisibleMethodWxpaySource = NormalizeVisibleMethodSource("wxpay", settings[SettingPaymentVisibleMethodWxpaySource])
	result.PaymentVisibleMethodAlipayEnabled = settings[SettingPaymentVisibleMethodAlipayEnabled] == "true"
	result.PaymentVisibleMethodWxpayEnabled = settings[SettingPaymentVisibleMethodWxpayEnabled] == "true"
	result.OpenAIAdvancedSchedulerEnabled = settings[openAIAdvancedSchedulerSettingKey] == "true"

	// 余额、订阅到期与账号限额通知
	result.BalanceLowNotifyEnabled = settings[SettingKeyBalanceLowNotifyEnabled] == "true"
	if v, err := strconv.ParseFloat(settings[SettingKeyBalanceLowNotifyThreshold], 64); err == nil && v >= 0 {
		result.BalanceLowNotifyThreshold = v
	}
	result.BalanceLowNotifyRechargeURL = settings[SettingKeyBalanceLowNotifyRechargeURL]
	result.SubscriptionExpiryNotifyEnabled = !isFalseSettingValue(settings[SettingKeySubscriptionExpiryNotifyEnabled])

	// 账号限额通知
	result.AccountQuotaNotifyEnabled = settings[SettingKeyAccountQuotaNotifyEnabled] == "true"
	if raw := strings.TrimSpace(settings[SettingKeyAccountQuotaNotifyEmails]); raw != "" {
		result.AccountQuotaNotifyEmails = ParseNotifyEmails(raw)
	}
	if result.AccountQuotaNotifyEmails == nil {
		result.AccountQuotaNotifyEmails = []NotifyEmailEntry{}
	}

	// Registration notification
	result.RegistrationNotifyEnabled = settings[SettingKeyRegistrationNotifyEnabled] == "true"
	result.RegistrationNotifyProvider = normalizeStoredRegistrationNotifyProvider(settings[SettingKeyRegistrationNotifyProvider])
	result.RegistrationNotifyWebhookURL = strings.TrimSpace(settings[SettingKeyRegistrationNotifyWebhookURL])
	result.RegistrationNotifySecret = strings.TrimSpace(settings[SettingKeyRegistrationNotifySecret])
	result.RegistrationNotifySecretConfigured = result.RegistrationNotifySecret != ""

	return result
}

func clampAffiliateRebateRate(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return AffiliateRebateRateDefault
	}
	if value < AffiliateRebateRateMin {
		return AffiliateRebateRateMin
	}
	if value > AffiliateRebateRateMax {
		return AffiliateRebateRateMax
	}
	return value
}

func isFalseSettingValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "false", "0", "off", "disabled":
		return true
	default:
		return false
	}
}

func normalizeVisibleMethodSettingSource(method, source string, enabled bool) (string, error) {
	_ = enabled
	source = strings.TrimSpace(source)
	if source == "" {
		return "", nil
	}

	normalized := NormalizeVisibleMethodSource(method, source)
	if normalized == "" {
		return "", infraerrors.BadRequest(
			"INVALID_PAYMENT_VISIBLE_METHOD_SOURCE",
			fmt.Sprintf("%s source must be one of the supported payment providers", method),
		)
	}
	return normalized, nil
}

func parseDefaultSubscriptions(raw string) []DefaultSubscriptionSetting {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var items []DefaultSubscriptionSetting
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}

	normalized := make([]DefaultSubscriptionSetting, 0, len(items))
	for _, item := range items {
		if item.GroupID <= 0 || item.ValidityDays <= 0 {
			continue
		}
		if item.ValidityDays > MaxValidityDays {
			item.ValidityDays = MaxValidityDays
		}
		normalized = append(normalized, item)
	}

	return normalized
}

func parseProviderDefaultGrantSettings(settings map[string]string, keys authSourceDefaultKeySet) ProviderDefaultGrantSettings {
	result := ProviderDefaultGrantSettings{
		Balance:          defaultAuthSourceBalance,
		Concurrency:      defaultAuthSourceConcurrency,
		Subscriptions:    []DefaultSubscriptionSetting{},
		GrantOnSignup:    false,
		GrantOnFirstBind: false,
	}

	if v, err := strconv.ParseFloat(strings.TrimSpace(settings[keys.balance]), 64); err == nil {
		result.Balance = v
	}
	if v, err := strconv.Atoi(strings.TrimSpace(settings[keys.concurrency])); err == nil {
		result.Concurrency = v
	}
	if items := parseDefaultSubscriptions(settings[keys.subscriptions]); items != nil {
		result.Subscriptions = items
	}
	if raw, ok := settings[keys.grantOnSignup]; ok {
		result.GrantOnSignup = raw == "true"
	}
	if raw, ok := settings[keys.grantOnFirstBind]; ok {
		result.GrantOnFirstBind = raw == "true"
	}

	if raw := settings[keys.platformQuotas]; raw != "" {
		parsed := map[string]*DefaultPlatformQuotaSetting{}
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			slog.Warn("[Setting] parseProviderDefaultGrantSettings: unmarshal auth source platform quotas failed", "source", keys.source, "error", err)
		} else {
			result.PlatformQuotas = parsed
		}
	}

	return result
}

func writeProviderDefaultGrantUpdates(updates map[string]string, keys authSourceDefaultKeySet, settings ProviderDefaultGrantSettings) {
	updates[keys.balance] = strconv.FormatFloat(settings.Balance, 'f', 8, 64)
	updates[keys.concurrency] = strconv.Itoa(settings.Concurrency)

	subscriptions := settings.Subscriptions
	if subscriptions == nil {
		subscriptions = []DefaultSubscriptionSetting{}
	}
	raw, err := json.Marshal(subscriptions)
	if err != nil {
		raw = []byte("[]")
	}
	updates[keys.subscriptions] = string(raw)
	updates[keys.grantOnSignup] = strconv.FormatBool(settings.GrantOnSignup)
	updates[keys.grantOnFirstBind] = strconv.FormatBool(settings.GrantOnFirstBind)

	// auth source platform quota：整体替换语义。
	// nil = 请求未携带该字段，跳过写入以保留既有配置（与系统层 buildSystemSettingsUpdates 的
	// DefaultPlatformQuotas nil 守卫一致）；非 nil（含空 map）才整体替换。二者语义不可混同。
	if keys.platformQuotas != "" && settings.PlatformQuotas != nil {
		blob, err := json.Marshal(settings.PlatformQuotas)
		if err != nil {
			blob = []byte("{}")
		}
		updates[keys.platformQuotas] = string(blob)
	}
}

func mergeProviderDefaultGrantSettings(globalDefaults ProviderDefaultGrantSettings, providerDefaults ProviderDefaultGrantSettings) ProviderDefaultGrantSettings {
	result := ProviderDefaultGrantSettings{
		Balance:          globalDefaults.Balance,
		Concurrency:      globalDefaults.Concurrency,
		Subscriptions:    append([]DefaultSubscriptionSetting(nil), globalDefaults.Subscriptions...),
		GrantOnSignup:    providerDefaults.GrantOnSignup,
		GrantOnFirstBind: providerDefaults.GrantOnFirstBind,
	}

	// 注意：不能把 parse 默认值 (defaultAuthSourceBalance / defaultAuthSourceConcurrency)
	// 当作"未配置"哨兵——admin 完全有权显式设成相同的值，那时仍应覆盖 globalDefaults。
	// 旧实现的 `!= defaultAuthSourceConcurrency` 会把 admin 设的 5 与 fallback 5 混淆，
	// 导致渠道发放退回到全局默认（如 1），表现为"管理员设 5、新用户实际拿 1"。
	if providerDefaults.Balance >= 0 {
		result.Balance = providerDefaults.Balance
	}
	if providerDefaults.Concurrency > 0 {
		result.Concurrency = providerDefaults.Concurrency
	}
	if len(providerDefaults.Subscriptions) > 0 {
		result.Subscriptions = append([]DefaultSubscriptionSetting(nil), providerDefaults.Subscriptions...)
	}

	return result
}

func parseTablePreferences(defaultPageSizeRaw, optionsRaw string) (int, []int) {
	defaultPageSize := 20
	if v, err := strconv.Atoi(strings.TrimSpace(defaultPageSizeRaw)); err == nil {
		defaultPageSize = v
	}

	var options []int
	if strings.TrimSpace(optionsRaw) != "" {
		_ = json.Unmarshal([]byte(optionsRaw), &options)
	}

	return normalizeTablePreferences(defaultPageSize, options)
}

func normalizeTablePreferences(defaultPageSize int, options []int) (int, []int) {
	const minPageSize = 5
	const maxPageSize = 1000
	const fallbackPageSize = 20

	seen := make(map[int]struct{}, len(options))
	normalizedOptions := make([]int, 0, len(options))
	for _, option := range options {
		if option < minPageSize || option > maxPageSize {
			continue
		}
		if _, ok := seen[option]; ok {
			continue
		}
		seen[option] = struct{}{}
		normalizedOptions = append(normalizedOptions, option)
	}
	sort.Ints(normalizedOptions)

	if defaultPageSize < minPageSize || defaultPageSize > maxPageSize {
		defaultPageSize = fallbackPageSize
	}

	if len(normalizedOptions) == 0 {
		normalizedOptions = []int{10, 20, 50}
	}

	return defaultPageSize, normalizedOptions
}

// getStringOrDefault 获取字符串值或默认值
func (s *SettingService) getStringOrDefault(settings map[string]string, key, defaultValue string) string {
	if value, ok := settings[key]; ok && value != "" {
		return value
	}
	return defaultValue
}

// IsTurnstileEnabled 检查是否启用 Turnstile 验证
func (s *SettingService) IsTurnstileEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyTurnstileEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

// GetTurnstileSecretKey 获取 Turnstile Secret Key
func (s *SettingService) GetTurnstileSecretKey(ctx context.Context) string {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyTurnstileSecretKey)
	if err != nil {
		return ""
	}
	return value
}

// IsIdentityPatchEnabled 检查是否启用身份补丁（Claude -> Gemini systemInstruction 注入）
func (s *SettingService) IsIdentityPatchEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyEnableIdentityPatch)
	if err != nil {
		// 默认开启，保持兼容
		return true
	}
	return value == "true"
}

// GetIdentityPatchPrompt 获取自定义身份补丁提示词（为空表示使用内置默认模板）
func (s *SettingService) GetIdentityPatchPrompt(ctx context.Context) string {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyIdentityPatchPrompt)
	if err != nil {
		return ""
	}
	return value
}

// GenerateAdminAPIKey 生成新的管理员 API Key
func (s *SettingService) GenerateAdminAPIKey(ctx context.Context) (string, error) {
	// 生成 32 字节随机数 = 64 位十六进制字符
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	key := AdminAPIKeyPrefix + hex.EncodeToString(bytes)

	// 存储到 settings 表
	if err := s.settingRepo.Set(ctx, SettingKeyAdminAPIKey, key); err != nil {
		return "", fmt.Errorf("save admin api key: %w", err)
	}

	return key, nil
}

// GetAdminAPIKeyStatus 获取管理员 API Key 状态
// 返回脱敏的 key、是否存在、错误
func (s *SettingService) GetAdminAPIKeyStatus(ctx context.Context) (maskedKey string, exists bool, err error) {
	key, err := s.settingRepo.GetValue(ctx, SettingKeyAdminAPIKey)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	if key == "" {
		return "", false, nil
	}

	// 脱敏：显示前 10 位和后 4 位
	if len(key) > 14 {
		maskedKey = key[:10] + "..." + key[len(key)-4:]
	} else {
		maskedKey = key
	}

	return maskedKey, true, nil
}

// GetAdminAPIKey 获取完整的管理员 API Key（仅供内部验证使用）
// 如果未配置返回空字符串和 nil 错误，只有数据库错误时才返回 error
func (s *SettingService) GetAdminAPIKey(ctx context.Context) (string, error) {
	key, err := s.settingRepo.GetValue(ctx, SettingKeyAdminAPIKey)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return "", nil // 未配置，返回空字符串
		}
		return "", err // 数据库错误
	}
	return key, nil
}

// DeleteAdminAPIKey 删除管理员 API Key
func (s *SettingService) DeleteAdminAPIKey(ctx context.Context) error {
	return s.settingRepo.Delete(ctx, SettingKeyAdminAPIKey)
}

// IsModelFallbackEnabled 检查是否启用模型兜底机制
func (s *SettingService) IsModelFallbackEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyEnableModelFallback)
	if err != nil {
		return false // Default: disabled
	}
	return value == "true"
}

// GetFallbackModel 获取指定平台的兜底模型
func (s *SettingService) GetFallbackModel(ctx context.Context, platform string) string {
	var key string
	var defaultModel string

	switch platform {
	case PlatformAnthropic:
		key = SettingKeyFallbackModelAnthropic
		defaultModel = "claude-3-5-sonnet-20241022"
	case PlatformOpenAI:
		key = SettingKeyFallbackModelOpenAI
		defaultModel = "gpt-4o"
	case PlatformGemini:
		key = SettingKeyFallbackModelGemini
		defaultModel = "gemini-2.5-pro"
	case PlatformAntigravity:
		key = SettingKeyFallbackModelAntigravity
		defaultModel = "gemini-2.5-pro"
	default:
		return ""
	}

	value, err := s.settingRepo.GetValue(ctx, key)
	if err != nil || value == "" {
		return defaultModel
	}
	return value
}

// GetLinuxDoConnectOAuthConfig 返回用于登录的"最终生效" LinuxDo Connect 配置。
//
// 优先级：
// - 若对应系统设置键存在，则覆盖 config.yaml/env 的值
// - 否则回退到 config.yaml/env 的值
func (s *SettingService) GetLinuxDoConnectOAuthConfig(ctx context.Context) (config.LinuxDoConnectConfig, error) {
	if s == nil || s.cfg == nil {
		return config.LinuxDoConnectConfig{}, infraerrors.ServiceUnavailable("CONFIG_NOT_READY", "config not loaded")
	}

	effective := s.cfg.LinuxDo

	keys := []string{
		SettingKeyLinuxDoConnectEnabled,
		SettingKeyLinuxDoConnectClientID,
		SettingKeyLinuxDoConnectClientSecret,
		SettingKeyLinuxDoConnectRedirectURL,
	}
	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return config.LinuxDoConnectConfig{}, fmt.Errorf("get linuxdo connect settings: %w", err)
	}

	if raw, ok := settings[SettingKeyLinuxDoConnectEnabled]; ok {
		effective.Enabled = raw == "true"
	}
	if v, ok := settings[SettingKeyLinuxDoConnectClientID]; ok && strings.TrimSpace(v) != "" {
		effective.ClientID = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyLinuxDoConnectClientSecret]; ok && strings.TrimSpace(v) != "" {
		effective.ClientSecret = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyLinuxDoConnectRedirectURL]; ok && strings.TrimSpace(v) != "" {
		effective.RedirectURL = strings.TrimSpace(v)
	}
	if !effective.Enabled {
		return config.LinuxDoConnectConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "oauth login is disabled")
	}

	// 基础健壮性校验（避免把用户重定向到一个必然失败或不安全的 OAuth 流程里）。
	if strings.TrimSpace(effective.ClientID) == "" {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth client id not configured")
	}
	if strings.TrimSpace(effective.AuthorizeURL) == "" {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth authorize url not configured")
	}
	if strings.TrimSpace(effective.TokenURL) == "" {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth token url not configured")
	}
	if strings.TrimSpace(effective.UserInfoURL) == "" {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth userinfo url not configured")
	}
	if strings.TrimSpace(effective.RedirectURL) == "" {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth redirect url not configured")
	}
	if strings.TrimSpace(effective.FrontendRedirectURL) == "" {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth frontend redirect url not configured")
	}

	if err := config.ValidateAbsoluteHTTPURL(effective.AuthorizeURL); err != nil {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth authorize url invalid")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.TokenURL); err != nil {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth token url invalid")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.UserInfoURL); err != nil {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth userinfo url invalid")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.RedirectURL); err != nil {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth redirect url invalid")
	}
	if err := config.ValidateFrontendRedirectURL(effective.FrontendRedirectURL); err != nil {
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth frontend redirect url invalid")
	}

	method := strings.ToLower(strings.TrimSpace(effective.TokenAuthMethod))
	switch method {
	case "", "client_secret_post", "client_secret_basic":
		if strings.TrimSpace(effective.ClientSecret) == "" {
			return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth client secret not configured")
		}
	case "none":
	default:
		return config.LinuxDoConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth token_auth_method invalid")
	}

	return effective, nil
}

// GetDingTalkConnectOAuthConfig 返回用于登录的"最终生效" DingTalk Connect 配置。
//
// 优先级：
// - 若对应系统设置键存在，则覆盖 config.yaml/env 的值
// - 否则回退到 config.yaml/env 的值
func (s *SettingService) GetDingTalkConnectOAuthConfig(ctx context.Context) (config.DingTalkConnectConfig, error) {
	if s == nil || s.cfg == nil {
		return config.DingTalkConnectConfig{}, infraerrors.ServiceUnavailable("CONFIG_NOT_READY", "config not loaded")
	}

	effective := s.cfg.DingTalk

	keys := []string{
		SettingKeyDingTalkConnectEnabled,
		SettingKeyDingTalkConnectClientID,
		SettingKeyDingTalkConnectClientSecret,
		SettingKeyDingTalkConnectRedirectURL,
		SettingKeyDingTalkConnectCorpRestrictionPolicy,
		SettingKeyDingTalkConnectInternalCorpID,
		SettingKeyDingTalkConnectBypassRegistration,
		SettingKeyDingTalkConnectSyncCorpEmail,
		SettingKeyDingTalkConnectSyncDisplayName,
		SettingKeyDingTalkConnectSyncDept,
		SettingKeyDingTalkConnectSyncCorpEmailAttrKey,
		SettingKeyDingTalkConnectSyncDisplayNameAttrKey,
		SettingKeyDingTalkConnectSyncDeptAttrKey,
	}
	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return config.DingTalkConnectConfig{}, fmt.Errorf("get dingtalk connect settings: %w", err)
	}

	if raw, ok := settings[SettingKeyDingTalkConnectEnabled]; ok {
		effective.Enabled = raw == "true"
	}
	if v, ok := settings[SettingKeyDingTalkConnectClientID]; ok && strings.TrimSpace(v) != "" {
		effective.ClientID = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyDingTalkConnectClientSecret]; ok && strings.TrimSpace(v) != "" {
		effective.ClientSecret = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyDingTalkConnectRedirectURL]; ok && strings.TrimSpace(v) != "" {
		effective.RedirectURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyDingTalkConnectCorpRestrictionPolicy]; ok && strings.TrimSpace(v) != "" {
		effective.CorpRestrictionPolicy = strings.TrimSpace(v)
	}
	effective.CorpRestrictionPolicy = coerceDeprecatedDingTalkCorpPolicy(effective.CorpRestrictionPolicy)
	if v, ok := settings[SettingKeyDingTalkConnectInternalCorpID]; ok && strings.TrimSpace(v) != "" {
		effective.InternalCorpID = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyDingTalkConnectBypassRegistration]; ok && strings.TrimSpace(v) != "" {
		effective.BypassRegistration = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	// bypass_registration 仅在 internal_only 模式下有意义；其它策略下强制 false，
	// 以保证 OAuth callback 看到的 effective config 永远是一致状态。
	if effective.CorpRestrictionPolicy != "internal_only" {
		effective.BypassRegistration = false
	}

	if v, ok := settings[SettingKeyDingTalkConnectSyncCorpEmail]; ok && strings.TrimSpace(v) != "" {
		effective.SyncCorpEmail = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	if v, ok := settings[SettingKeyDingTalkConnectSyncDisplayName]; ok && strings.TrimSpace(v) != "" {
		effective.SyncDisplayName = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	if v, ok := settings[SettingKeyDingTalkConnectSyncDept]; ok && strings.TrimSpace(v) != "" {
		effective.SyncDept = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	// 身份同步三开关仅在 internal_only 模式下有意义；其它策略强制 false。
	if effective.CorpRestrictionPolicy != "internal_only" {
		effective.SyncCorpEmail = false
		effective.SyncDisplayName = false
		effective.SyncDept = false
	}

	// 身份同步目标 attr key（DB 空 → fallback 默认值）
	if v := strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncCorpEmailAttrKey]); v != "" {
		effective.SyncCorpEmailAttrKey = v
	}
	if effective.SyncCorpEmailAttrKey == "" {
		effective.SyncCorpEmailAttrKey = "dingtalk_email"
	}
	if v := strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncDisplayNameAttrKey]); v != "" {
		effective.SyncDisplayNameAttrKey = v
	}
	if effective.SyncDisplayNameAttrKey == "" {
		effective.SyncDisplayNameAttrKey = "dingtalk_name"
	}
	if v := strings.TrimSpace(settings[SettingKeyDingTalkConnectSyncDeptAttrKey]); v != "" {
		effective.SyncDeptAttrKey = v
	}
	if effective.SyncDeptAttrKey == "" {
		effective.SyncDeptAttrKey = "dingtalk_department"
	}

	if !effective.Enabled {
		return config.DingTalkConnectConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "dingtalk oauth login is disabled")
	}

	// 基础健壮性校验（避免把用户重定向到一个必然失败或不安全的 OAuth 流程里）。
	if strings.TrimSpace(effective.ClientID) == "" {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth client id not configured")
	}
	if strings.TrimSpace(effective.AuthorizeURL) == "" {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth authorize url not configured")
	}
	if strings.TrimSpace(effective.TokenURL) == "" {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth token url not configured")
	}
	if strings.TrimSpace(effective.UserInfoURL) == "" {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth userinfo url not configured")
	}
	if strings.TrimSpace(effective.RedirectURL) == "" {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth redirect url not configured")
	}
	if strings.TrimSpace(effective.FrontendRedirectURL) == "" {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth frontend redirect url not configured")
	}

	if err := config.ValidateAbsoluteHTTPURL(effective.AuthorizeURL); err != nil {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth authorize url invalid")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.TokenURL); err != nil {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth token url invalid")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.UserInfoURL); err != nil {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth userinfo url invalid")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.RedirectURL); err != nil {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth redirect url invalid")
	}
	if err := config.ValidateFrontendRedirectURL(effective.FrontendRedirectURL); err != nil {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth frontend redirect url invalid")
	}
	if strings.TrimSpace(effective.ClientSecret) == "" {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "dingtalk oauth client secret not configured")
	}

	// 镜像 admin handler 行为：internal_only policy 隐式要求 AppType=internal
	if effective.CorpRestrictionPolicy == "internal_only" {
		effective.AppType = "internal"
	}

	if err := config.ValidateDingTalkConfig(effective); err != nil {
		return config.DingTalkConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", err.Error())
	}

	return effective, nil
}

// GetWeChatConnectOAuthConfig 返回用于登录的最终生效 WeChat Connect 配置。
//
// WeChat Connect 已回归 DB 系统设置模型，不再回退到 config/env。
func (s *SettingService) GetWeChatConnectOAuthConfig(ctx context.Context) (WeChatConnectOAuthConfig, error) {
	keys := []string{
		SettingKeyWeChatConnectEnabled,
		SettingKeyWeChatConnectOpenAppID,
		SettingKeyWeChatConnectOpenAppSecret,
		SettingKeyWeChatConnectMPAppID,
		SettingKeyWeChatConnectMPAppSecret,
		SettingKeyWeChatConnectMobileAppID,
		SettingKeyWeChatConnectMobileAppSecret,
		SettingKeyWeChatConnectOpenEnabled,
		SettingKeyWeChatConnectMPEnabled,
		SettingKeyWeChatConnectMobileEnabled,
		SettingKeyWeChatConnectMode,
		SettingKeyWeChatConnectScopes,
		SettingKeyWeChatConnectRedirectURL,
		SettingKeyWeChatConnectFrontendRedirectURL,
	}
	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return WeChatConnectOAuthConfig{}, fmt.Errorf("get wechat connect settings: %w", err)
	}
	return s.parseWeChatConnectOAuthConfig(settings)
}

// GetOverloadCooldownSettings 获取529过载冷却配置
func (s *SettingService) GetOverloadCooldownSettings(ctx context.Context) (*OverloadCooldownSettings, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyOverloadCooldownSettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultOverloadCooldownSettings(), nil
		}
		return nil, fmt.Errorf("get overload cooldown settings: %w", err)
	}
	if value == "" {
		return DefaultOverloadCooldownSettings(), nil
	}

	var settings OverloadCooldownSettings
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		return DefaultOverloadCooldownSettings(), nil
	}

	// 修正配置值范围
	if settings.CooldownMinutes < 1 {
		settings.CooldownMinutes = 1
	}
	if settings.CooldownMinutes > 120 {
		settings.CooldownMinutes = 120
	}

	return &settings, nil
}

// SetOverloadCooldownSettings 设置529过载冷却配置
func (s *SettingService) SetOverloadCooldownSettings(ctx context.Context, settings *OverloadCooldownSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// 禁用时修正为合法值即可，不拒绝请求
	if settings.CooldownMinutes < 1 || settings.CooldownMinutes > 120 {
		if settings.Enabled {
			return fmt.Errorf("cooldown_minutes must be between 1-120")
		}
		settings.CooldownMinutes = 10 // 禁用状态下归一化为默认值
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal overload cooldown settings: %w", err)
	}

	return s.settingRepo.Set(ctx, SettingKeyOverloadCooldownSettings, string(data))
}

// GetImagePromptFilterConfig 获取图片提示词过滤配置
func (s *SettingService) GetImagePromptFilterConfig(ctx context.Context) (*ImagePromptFilterConfig, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyImagePromptFilterConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultImagePromptFilterConfig(), nil
		}
		return nil, fmt.Errorf("get image prompt filter config: %w", err)
	}
	if value == "" {
		return DefaultImagePromptFilterConfig(), nil
	}

	var config ImagePromptFilterConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return DefaultImagePromptFilterConfig(), nil
	}

	if config.ExplicitKeywords == nil {
		config.ExplicitKeywords = DefaultImagePromptFilterConfig().ExplicitKeywords
	}
	if config.YouthContextKeywords == nil {
		config.YouthContextKeywords = DefaultImagePromptFilterConfig().YouthContextKeywords
	}
	if config.WarningMessage == "" {
		config.WarningMessage = DefaultImagePromptFilterConfig().WarningMessage
	}
	if config.YouthWarningMessage == "" {
		config.YouthWarningMessage = DefaultImagePromptFilterConfig().YouthWarningMessage
	}

	return &config, nil
}

// SetImagePromptFilterConfig 设置图片提示词过滤配置
func (s *SettingService) SetImagePromptFilterConfig(ctx context.Context, config *ImagePromptFilterConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.ExplicitKeywords == nil {
		config.ExplicitKeywords = []string{}
	}
	if config.YouthContextKeywords == nil {
		config.YouthContextKeywords = []string{}
	}

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal image prompt filter config: %w", err)
	}

	return s.settingRepo.Set(ctx, SettingKeyImagePromptFilterConfig, string(data))
}

// GetRateLimit429CooldownSettings 获取429默认回避配置
func (s *SettingService) GetRateLimit429CooldownSettings(ctx context.Context) (*RateLimit429CooldownSettings, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyRateLimit429CooldownSettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultRateLimit429CooldownSettings(), nil
		}
		return nil, fmt.Errorf("get 429 cooldown settings: %w", err)
	}
	if value == "" {
		return DefaultRateLimit429CooldownSettings(), nil
	}

	var settings RateLimit429CooldownSettings
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		return DefaultRateLimit429CooldownSettings(), nil
	}

	if settings.CooldownSeconds < 1 {
		settings.CooldownSeconds = 1
	}
	if settings.CooldownSeconds > 7200 {
		settings.CooldownSeconds = 7200
	}

	return &settings, nil
}

// SetRateLimit429CooldownSettings 设置429默认回避配置
func (s *SettingService) SetRateLimit429CooldownSettings(ctx context.Context, settings *RateLimit429CooldownSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	if settings.CooldownSeconds < 1 || settings.CooldownSeconds > 7200 {
		if settings.Enabled {
			return fmt.Errorf("cooldown_seconds must be between 1-7200")
		}
		settings.CooldownSeconds = 5
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal 429 cooldown settings: %w", err)
	}

	return s.settingRepo.Set(ctx, SettingKeyRateLimit429CooldownSettings, string(data))
}

// GetOIDCConnectOAuthConfig 返回用于登录的“最终生效” OIDC 配置。
//
// 优先级：
// - 若对应系统设置键存在，则覆盖 config.yaml/env 的值
// - 否则回退到 config.yaml/env 的值
func (s *SettingService) GetOIDCConnectOAuthConfig(ctx context.Context) (config.OIDCConnectConfig, error) {
	if s == nil || s.cfg == nil {
		return config.OIDCConnectConfig{}, infraerrors.ServiceUnavailable("CONFIG_NOT_READY", "config not loaded")
	}

	effective := s.cfg.OIDC

	keys := []string{
		SettingKeyOIDCConnectEnabled,
		SettingKeyOIDCConnectProviderName,
		SettingKeyOIDCConnectClientID,
		SettingKeyOIDCConnectClientSecret,
		SettingKeyOIDCConnectIssuerURL,
		SettingKeyOIDCConnectDiscoveryURL,
		SettingKeyOIDCConnectAuthorizeURL,
		SettingKeyOIDCConnectTokenURL,
		SettingKeyOIDCConnectUserInfoURL,
		SettingKeyOIDCConnectJWKSURL,
		SettingKeyOIDCConnectScopes,
		SettingKeyOIDCConnectRedirectURL,
		SettingKeyOIDCConnectFrontendRedirectURL,
		SettingKeyOIDCConnectTokenAuthMethod,
		SettingKeyOIDCConnectUsePKCE,
		SettingKeyOIDCConnectValidateIDToken,
		SettingKeyOIDCConnectAllowedSigningAlgs,
		SettingKeyOIDCConnectClockSkewSeconds,
		SettingKeyOIDCConnectRequireEmailVerified,
		SettingKeyOIDCConnectUserInfoEmailPath,
		SettingKeyOIDCConnectUserInfoIDPath,
		SettingKeyOIDCConnectUserInfoUsernamePath,
	}
	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return config.OIDCConnectConfig{}, fmt.Errorf("get oidc connect settings: %w", err)
	}

	if raw, ok := settings[SettingKeyOIDCConnectEnabled]; ok {
		effective.Enabled = raw == "true"
	}
	if v, ok := settings[SettingKeyOIDCConnectProviderName]; ok && strings.TrimSpace(v) != "" {
		effective.ProviderName = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectClientID]; ok && strings.TrimSpace(v) != "" {
		effective.ClientID = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectClientSecret]; ok && strings.TrimSpace(v) != "" {
		effective.ClientSecret = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectIssuerURL]; ok && strings.TrimSpace(v) != "" {
		effective.IssuerURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectDiscoveryURL]; ok && strings.TrimSpace(v) != "" {
		effective.DiscoveryURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectAuthorizeURL]; ok && strings.TrimSpace(v) != "" {
		effective.AuthorizeURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectTokenURL]; ok && strings.TrimSpace(v) != "" {
		effective.TokenURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectUserInfoURL]; ok && strings.TrimSpace(v) != "" {
		effective.UserInfoURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectJWKSURL]; ok && strings.TrimSpace(v) != "" {
		effective.JWKSURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectScopes]; ok && strings.TrimSpace(v) != "" {
		effective.Scopes = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectRedirectURL]; ok && strings.TrimSpace(v) != "" {
		effective.RedirectURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectFrontendRedirectURL]; ok && strings.TrimSpace(v) != "" {
		effective.FrontendRedirectURL = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectTokenAuthMethod]; ok && strings.TrimSpace(v) != "" {
		effective.TokenAuthMethod = strings.ToLower(strings.TrimSpace(v))
	}
	if raw, ok := settings[SettingKeyOIDCConnectUsePKCE]; ok {
		effective.UsePKCE = raw == "true"
	} else {
		effective.UsePKCE = oidcUsePKCECompatibilityDefault(effective)
	}
	if raw, ok := settings[SettingKeyOIDCConnectValidateIDToken]; ok {
		effective.ValidateIDToken = raw == "true"
	} else {
		effective.ValidateIDToken = oidcValidateIDTokenCompatibilityDefault(effective)
	}
	if v, ok := settings[SettingKeyOIDCConnectAllowedSigningAlgs]; ok && strings.TrimSpace(v) != "" {
		effective.AllowedSigningAlgs = strings.TrimSpace(v)
	}
	if raw, ok := settings[SettingKeyOIDCConnectClockSkewSeconds]; ok && strings.TrimSpace(raw) != "" {
		if parsed, parseErr := strconv.Atoi(strings.TrimSpace(raw)); parseErr == nil {
			effective.ClockSkewSeconds = parsed
		}
	}
	if raw, ok := settings[SettingKeyOIDCConnectRequireEmailVerified]; ok {
		effective.RequireEmailVerified = raw == "true"
	}
	if v, ok := settings[SettingKeyOIDCConnectUserInfoEmailPath]; ok {
		effective.UserInfoEmailPath = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectUserInfoIDPath]; ok {
		effective.UserInfoIDPath = strings.TrimSpace(v)
	}
	if v, ok := settings[SettingKeyOIDCConnectUserInfoUsernamePath]; ok {
		effective.UserInfoUsernamePath = strings.TrimSpace(v)
	}

	if !effective.Enabled {
		return config.OIDCConnectConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "oauth login is disabled")
	}
	if strings.TrimSpace(effective.ProviderName) == "" {
		effective.ProviderName = "OIDC"
	}
	if strings.TrimSpace(effective.ClientID) == "" {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth client id not configured")
	}
	if strings.TrimSpace(effective.IssuerURL) == "" {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth issuer url not configured")
	}
	if strings.TrimSpace(effective.RedirectURL) == "" {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth redirect url not configured")
	}
	if strings.TrimSpace(effective.FrontendRedirectURL) == "" {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth frontend redirect url not configured")
	}
	if !scopesContainOpenID(effective.Scopes) {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth scopes must contain openid")
	}
	if effective.ClockSkewSeconds < 0 || effective.ClockSkewSeconds > 600 {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth clock skew must be between 0 and 600")
	}

	if err := config.ValidateAbsoluteHTTPURL(effective.IssuerURL); err != nil {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth issuer url invalid")
	}

	discoveryURL := strings.TrimSpace(effective.DiscoveryURL)
	if discoveryURL == "" {
		discoveryURL = oidcDefaultDiscoveryURL(effective.IssuerURL)
		effective.DiscoveryURL = discoveryURL
	}
	if discoveryURL != "" {
		if err := config.ValidateAbsoluteHTTPURL(discoveryURL); err != nil {
			return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth discovery url invalid")
		}
	}

	needsDiscovery := strings.TrimSpace(effective.AuthorizeURL) == "" ||
		strings.TrimSpace(effective.TokenURL) == "" ||
		(effective.ValidateIDToken && strings.TrimSpace(effective.JWKSURL) == "")
	if needsDiscovery && discoveryURL != "" {
		metadata, resolveErr := oidcResolveProviderMetadata(ctx, discoveryURL)
		if resolveErr != nil {
			return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth discovery resolve failed").WithCause(resolveErr)
		}
		if strings.TrimSpace(effective.AuthorizeURL) == "" {
			effective.AuthorizeURL = strings.TrimSpace(metadata.AuthorizationEndpoint)
		}
		if strings.TrimSpace(effective.TokenURL) == "" {
			effective.TokenURL = strings.TrimSpace(metadata.TokenEndpoint)
		}
		if strings.TrimSpace(effective.UserInfoURL) == "" {
			effective.UserInfoURL = strings.TrimSpace(metadata.UserInfoEndpoint)
		}
		if strings.TrimSpace(effective.JWKSURL) == "" {
			effective.JWKSURL = strings.TrimSpace(metadata.JWKSURI)
		}
	}

	if strings.TrimSpace(effective.AuthorizeURL) == "" {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth authorize url not configured")
	}
	if strings.TrimSpace(effective.TokenURL) == "" {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth token url not configured")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.AuthorizeURL); err != nil {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth authorize url invalid")
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.TokenURL); err != nil {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth token url invalid")
	}
	if v := strings.TrimSpace(effective.UserInfoURL); v != "" {
		if err := config.ValidateAbsoluteHTTPURL(v); err != nil {
			return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth userinfo url invalid")
		}
	}
	if effective.ValidateIDToken {
		if strings.TrimSpace(effective.JWKSURL) == "" {
			return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth jwks url not configured")
		}
		if strings.TrimSpace(effective.AllowedSigningAlgs) == "" {
			return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth signing algs not configured")
		}
	}
	if v := strings.TrimSpace(effective.JWKSURL); v != "" {
		if err := config.ValidateAbsoluteHTTPURL(v); err != nil {
			return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth jwks url invalid")
		}
	}
	if err := config.ValidateAbsoluteHTTPURL(effective.RedirectURL); err != nil {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth redirect url invalid")
	}
	if err := config.ValidateFrontendRedirectURL(effective.FrontendRedirectURL); err != nil {
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth frontend redirect url invalid")
	}

	method := strings.ToLower(strings.TrimSpace(effective.TokenAuthMethod))
	switch method {
	case "", "client_secret_post", "client_secret_basic":
		if strings.TrimSpace(effective.ClientSecret) == "" {
			return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth client secret not configured")
		}
	case "none":
	default:
		return config.OIDCConnectConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth token_auth_method invalid")
	}

	return effective, nil
}

func scopesContainOpenID(scopes string) bool {
	for _, scope := range strings.Fields(strings.ToLower(strings.TrimSpace(scopes))) {
		if scope == "openid" {
			return true
		}
	}
	return false
}

type oidcProviderMetadata struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

func oidcDefaultDiscoveryURL(issuerURL string) string {
	issuerURL = strings.TrimSpace(issuerURL)
	if issuerURL == "" {
		return ""
	}
	return strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"
}

func oidcResolveProviderMetadata(ctx context.Context, discoveryURL string) (*oidcProviderMetadata, error) {
	discoveryURL = strings.TrimSpace(discoveryURL)
	if discoveryURL == "" {
		return nil, fmt.Errorf("discovery url is empty")
	}

	resp, err := req.C().
		SetTimeout(15*time.Second).
		R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("request discovery document: %w", err)
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("discovery request failed: status=%d", resp.StatusCode)
	}

	metadata := &oidcProviderMetadata{}
	if err := json.Unmarshal(resp.Bytes(), metadata); err != nil {
		return nil, fmt.Errorf("parse discovery document: %w", err)
	}
	return metadata, nil
}

// GetStreamTimeoutSettings 获取流超时处理配置
func (s *SettingService) GetStreamTimeoutSettings(ctx context.Context) (*StreamTimeoutSettings, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyStreamTimeoutSettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultStreamTimeoutSettings(), nil
		}
		return nil, fmt.Errorf("get stream timeout settings: %w", err)
	}
	if value == "" {
		return DefaultStreamTimeoutSettings(), nil
	}

	var settings StreamTimeoutSettings
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		return DefaultStreamTimeoutSettings(), nil
	}

	// 验证并修正配置值
	if settings.TempUnschedMinutes < 1 {
		settings.TempUnschedMinutes = 1
	}
	if settings.TempUnschedMinutes > 60 {
		settings.TempUnschedMinutes = 60
	}
	if settings.ThresholdCount < 1 {
		settings.ThresholdCount = 1
	}
	if settings.ThresholdCount > 10 {
		settings.ThresholdCount = 10
	}
	if settings.ThresholdWindowMinutes < 1 {
		settings.ThresholdWindowMinutes = 1
	}
	if settings.ThresholdWindowMinutes > 60 {
		settings.ThresholdWindowMinutes = 60
	}

	// 验证 action
	switch settings.Action {
	case StreamTimeoutActionTempUnsched, StreamTimeoutActionError, StreamTimeoutActionNone:
		// valid
	default:
		settings.Action = StreamTimeoutActionTempUnsched
	}

	return &settings, nil
}

// IsUngroupedKeySchedulingAllowed 查询是否允许未分组 Key 调度
func (s *SettingService) IsUngroupedKeySchedulingAllowed(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyAllowUngroupedKeyScheduling)
	if err != nil {
		return false // fail-closed: 查询失败时默认不允许
	}
	return value == "true"
}

// GetClaudeCodeVersionBounds 获取 Claude Code 版本号上下限要求
// 使用进程内 atomic.Value 缓存，60 秒 TTL，热路径零锁开销
// singleflight 防止缓存过期时 thundering herd
// 返回空字符串表示不做对应方向的版本检查
func (s *SettingService) GetClaudeCodeVersionBounds(ctx context.Context) (min, max string) {
	if cached, ok := versionBoundsCache.Load().(*cachedVersionBounds); ok {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.min, cached.max
		}
	}
	// singleflight: 同一时刻只有一个 goroutine 查询 DB，其余复用结果
	type bounds struct{ min, max string }
	result, err, _ := versionBoundsSF.Do("version_bounds", func() (any, error) {
		// 二次检查，避免排队的 goroutine 重复查询
		if cached, ok := versionBoundsCache.Load().(*cachedVersionBounds); ok {
			if time.Now().UnixNano() < cached.expiresAt {
				return bounds{cached.min, cached.max}, nil
			}
		}
		// 使用独立 context：断开请求取消链，避免客户端断连导致空值被长期缓存
		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), versionBoundsDBTimeout)
		defer cancel()
		values, err := s.settingRepo.GetMultiple(dbCtx, []string{
			SettingKeyMinClaudeCodeVersion,
			SettingKeyMaxClaudeCodeVersion,
		})
		if err != nil {
			// fail-open: DB 错误时不阻塞请求，但记录日志并使用短 TTL 快速重试
			slog.Warn("failed to get claude code version bounds setting, skipping version check", "error", err)
			versionBoundsCache.Store(&cachedVersionBounds{
				min:       "",
				max:       "",
				expiresAt: time.Now().Add(versionBoundsErrorTTL).UnixNano(),
			})
			return bounds{"", ""}, nil
		}
		b := bounds{
			min: values[SettingKeyMinClaudeCodeVersion],
			max: values[SettingKeyMaxClaudeCodeVersion],
		}
		versionBoundsCache.Store(&cachedVersionBounds{
			min:       b.min,
			max:       b.max,
			expiresAt: time.Now().Add(versionBoundsCacheTTL).UnixNano(),
		})
		return b, nil
	})
	if err != nil {
		return "", ""
	}
	b, ok := result.(bounds)
	if !ok {
		return "", ""
	}
	return b.min, b.max
}

// GetOpenAIQuotaAutoPauseSettings returns the current global default quota auto-pause
// settings. It is invoked on the OpenAI scheduling hot path (once per request) and is
// therefore designed to never block on the DB:
//
//   - Fresh cached value → returned immediately.
//   - Stale or empty cache → the last known value is returned, and a background
//     goroutine refreshes the cache via singleflight (stale-while-revalidate).
//   - First call with no cache yet → zero defaults are returned and the same async
//     refresh is kicked off; the next call gets the freshly populated value.
//
// Callers that need the freshly persisted value synchronously (tests, post-update
// confirmation, optional startup warm-up) should call WarmOpenAIQuotaAutoPauseSettings.
func (s *SettingService) GetOpenAIQuotaAutoPauseSettings(ctx context.Context) OpsOpenAIAccountQuotaAutoPauseSettings {
	if s == nil {
		return OpsOpenAIAccountQuotaAutoPauseSettings{}
	}
	cached, _ := s.openAIQuotaAutoPauseSettingsCache.Load().(*cachedOpenAIQuotaAutoPauseSettings)
	now := time.Now().UnixNano()
	if cached != nil && now < cached.expiresAt {
		return cached.settings
	}
	// Stale or unset: trigger background refresh without blocking this request.
	// singleflight.DoChan dedupes concurrent refreshes; we deliberately ignore the
	// returned channel — the result is observable via the atomic cache.
	s.openAIQuotaAutoPauseSettingsSF.DoChan(openAIQuotaAutoPauseSettingsRefreshKey, func() (any, error) {
		s.refreshOpenAIQuotaAutoPauseSettings(context.Background())
		return nil, nil
	})
	if cached != nil {
		return cached.settings // serve stale value while revalidating
	}
	return OpsOpenAIAccountQuotaAutoPauseSettings{}
}

// WarmOpenAIQuotaAutoPauseSettings synchronously loads the quota auto-pause settings
// into the in-memory cache. Useful for application startup (so the first request hits
// a warm cache) and for tests that need deterministic reads immediately after
// constructing the service.
func (s *SettingService) WarmOpenAIQuotaAutoPauseSettings(ctx context.Context) OpsOpenAIAccountQuotaAutoPauseSettings {
	if s == nil {
		return OpsOpenAIAccountQuotaAutoPauseSettings{}
	}
	s.refreshOpenAIQuotaAutoPauseSettings(ctx)
	cached, _ := s.openAIQuotaAutoPauseSettingsCache.Load().(*cachedOpenAIQuotaAutoPauseSettings)
	if cached == nil {
		return OpsOpenAIAccountQuotaAutoPauseSettings{}
	}
	return cached.settings
}

// refreshOpenAIQuotaAutoPauseSettings reads the latest settings from the DB and stores
// them into the in-memory cache. On error it stores the prior value (or zero defaults
// if nothing is cached yet) with the shorter error TTL so the next refresh comes
// sooner. Always uses its own timeout-bounded context to keep refresh latency
// predictable regardless of the caller.
func (s *SettingService) refreshOpenAIQuotaAutoPauseSettings(ctx context.Context) {
	if s == nil || s.settingRepo == nil {
		return
	}
	dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), openAIQuotaAutoPauseSettingsDBTimeout)
	defer cancel()

	settings := OpsOpenAIAccountQuotaAutoPauseSettings{}
	ttl := openAIQuotaAutoPauseSettingsCacheTTL
	raw, err := s.settingRepo.GetValue(dbCtx, SettingKeyOpsAdvancedSettings)
	if err == nil {
		cfg := defaultOpsAdvancedSettings()
		if strings.TrimSpace(raw) != "" {
			if jsonErr := json.Unmarshal([]byte(raw), cfg); jsonErr == nil {
				normalizeOpsAdvancedSettings(cfg)
			}
		}
		settings = cfg.OpenAIAccountQuotaAutoPause
	} else if !errors.Is(err, ErrSettingNotFound) {
		// Real error: keep serving prior value but refresh sooner.
		if prior, _ := s.openAIQuotaAutoPauseSettingsCache.Load().(*cachedOpenAIQuotaAutoPauseSettings); prior != nil {
			settings = prior.settings
		}
		ttl = openAIQuotaAutoPauseSettingsErrorTTL
	}

	s.openAIQuotaAutoPauseSettingsCache.Store(&cachedOpenAIQuotaAutoPauseSettings{
		settings:  settings,
		expiresAt: time.Now().Add(ttl).UnixNano(),
	})
}

// SetOpenAIQuotaAutoPauseSettings writes the given settings directly into the in-memory
// cache. Called from settings-write code paths so that the next read reflects the new
// value immediately, without waiting for the background refresh.
func (s *SettingService) SetOpenAIQuotaAutoPauseSettings(settings OpsOpenAIAccountQuotaAutoPauseSettings) {
	if s == nil {
		return
	}
	s.openAIQuotaAutoPauseSettingsCache.Store(&cachedOpenAIQuotaAutoPauseSettings{
		settings:  settings,
		expiresAt: time.Now().Add(openAIQuotaAutoPauseSettingsCacheTTL).UnixNano(),
	})
}

// GetRectifierSettings 获取请求整流器配置
func (s *SettingService) GetRectifierSettings(ctx context.Context) (*RectifierSettings, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyRectifierSettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultRectifierSettings(), nil
		}
		return nil, fmt.Errorf("get rectifier settings: %w", err)
	}
	if value == "" {
		return DefaultRectifierSettings(), nil
	}

	var settings RectifierSettings
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		return DefaultRectifierSettings(), nil
	}

	return &settings, nil
}

// SetRectifierSettings 设置请求整流器配置
func (s *SettingService) SetRectifierSettings(ctx context.Context, settings *RectifierSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal rectifier settings: %w", err)
	}

	return s.settingRepo.Set(ctx, SettingKeyRectifierSettings, string(data))
}

// IsSignatureRectifierEnabled 判断签名整流是否启用（总开关 && 签名子开关）
func (s *SettingService) IsSignatureRectifierEnabled(ctx context.Context) bool {
	settings, err := s.GetRectifierSettings(ctx)
	if err != nil {
		return true // fail-open: 查询失败时默认启用
	}
	return settings.Enabled && settings.ThinkingSignatureEnabled
}

// IsBudgetRectifierEnabled 判断 Budget 整流是否启用（总开关 && Budget 子开关）
func (s *SettingService) IsBudgetRectifierEnabled(ctx context.Context) bool {
	settings, err := s.GetRectifierSettings(ctx)
	if err != nil {
		return true // fail-open: 查询失败时默认启用
	}
	return settings.Enabled && settings.ThinkingBudgetEnabled
}

// GetBetaPolicySettings 获取 Beta 策略配置
func (s *SettingService) GetBetaPolicySettings(ctx context.Context) (*BetaPolicySettings, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyBetaPolicySettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultBetaPolicySettings(), nil
		}
		return nil, fmt.Errorf("get beta policy settings: %w", err)
	}
	if value == "" {
		return DefaultBetaPolicySettings(), nil
	}

	var settings BetaPolicySettings
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		return DefaultBetaPolicySettings(), nil
	}

	return &settings, nil
}

// SetBetaPolicySettings 设置 Beta 策略配置
func (s *SettingService) SetBetaPolicySettings(ctx context.Context, settings *BetaPolicySettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	validActions := map[string]bool{
		BetaPolicyActionPass: true, BetaPolicyActionFilter: true, BetaPolicyActionBlock: true,
	}
	validScopes := map[string]bool{
		BetaPolicyScopeAll: true, BetaPolicyScopeOAuth: true, BetaPolicyScopeAPIKey: true, BetaPolicyScopeBedrock: true,
	}

	for i, rule := range settings.Rules {
		if rule.BetaToken == "" {
			return fmt.Errorf("rule[%d]: beta_token cannot be empty", i)
		}
		if !validActions[rule.Action] {
			return fmt.Errorf("rule[%d]: invalid action %q", i, rule.Action)
		}
		if !validScopes[rule.Scope] {
			return fmt.Errorf("rule[%d]: invalid scope %q", i, rule.Scope)
		}
		// Validate model_whitelist patterns
		for j, pattern := range rule.ModelWhitelist {
			trimmed := strings.TrimSpace(pattern)
			if trimmed == "" {
				return fmt.Errorf("rule[%d]: model_whitelist[%d] cannot be empty", i, j)
			}
			settings.Rules[i].ModelWhitelist[j] = trimmed
		}
		// Validate fallback_action
		if rule.FallbackAction != "" && !validActions[rule.FallbackAction] {
			return fmt.Errorf("rule[%d]: invalid fallback_action %q", i, rule.FallbackAction)
		}
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal beta policy settings: %w", err)
	}

	return s.settingRepo.Set(ctx, SettingKeyBetaPolicySettings, string(data))
}

// GetOpenAIFastPolicySettings 获取 OpenAI fast 策略配置
func (s *SettingService) GetOpenAIFastPolicySettings(ctx context.Context) (*OpenAIFastPolicySettings, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyOpenAIFastPolicySettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultOpenAIFastPolicySettings(), nil
		}
		return nil, fmt.Errorf("get openai fast policy settings: %w", err)
	}
	if value == "" {
		return DefaultOpenAIFastPolicySettings(), nil
	}

	var settings OpenAIFastPolicySettings
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		// JSON 损坏时静默 fallback 到默认配置会让策略意外失效（管理员配
		// 置的 block/filter 规则被忽略）。记录 Warn 让运维能在出现异常
		// 行为时定位到 settings 表里的脏数据。
		slog.Warn("failed to unmarshal openai fast policy settings, falling back to defaults",
			"error", err,
			"key", SettingKeyOpenAIFastPolicySettings)
		return DefaultOpenAIFastPolicySettings(), nil
	}

	return &settings, nil
}

// SetOpenAIFastPolicySettings 设置 OpenAI fast 策略配置
func (s *SettingService) SetOpenAIFastPolicySettings(ctx context.Context, settings *OpenAIFastPolicySettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	validActions := map[string]bool{
		BetaPolicyActionPass: true, BetaPolicyActionFilter: true, BetaPolicyActionBlock: true,
	}
	validScopes := map[string]bool{
		BetaPolicyScopeAll: true, BetaPolicyScopeOAuth: true, BetaPolicyScopeAPIKey: true, BetaPolicyScopeBedrock: true,
	}
	validTiers := map[string]bool{
		OpenAIFastTierAny: true, OpenAIFastTierPriority: true, OpenAIFastTierFlex: true,
	}

	for i, rule := range settings.Rules {
		tier := strings.ToLower(strings.TrimSpace(rule.ServiceTier))
		if tier == "" {
			tier = OpenAIFastTierAny
		}
		if !validTiers[tier] {
			return fmt.Errorf("rule[%d]: invalid service_tier %q", i, rule.ServiceTier)
		}
		settings.Rules[i].ServiceTier = tier
		if !validActions[rule.Action] {
			return fmt.Errorf("rule[%d]: invalid action %q", i, rule.Action)
		}
		if !validScopes[rule.Scope] {
			return fmt.Errorf("rule[%d]: invalid scope %q", i, rule.Scope)
		}
		for j, pattern := range rule.ModelWhitelist {
			trimmed := strings.TrimSpace(pattern)
			if trimmed == "" {
				return fmt.Errorf("rule[%d]: model_whitelist[%d] cannot be empty", i, j)
			}
			settings.Rules[i].ModelWhitelist[j] = trimmed
		}
		if rule.FallbackAction != "" && !validActions[rule.FallbackAction] {
			return fmt.Errorf("rule[%d]: invalid fallback_action %q", i, rule.FallbackAction)
		}
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal openai fast policy settings: %w", err)
	}

	return s.settingRepo.Set(ctx, SettingKeyOpenAIFastPolicySettings, string(data))
}

// SetStreamTimeoutSettings 设置流超时处理配置
func (s *SettingService) SetStreamTimeoutSettings(ctx context.Context, settings *StreamTimeoutSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// 验证配置值
	if settings.TempUnschedMinutes < 1 || settings.TempUnschedMinutes > 60 {
		return fmt.Errorf("temp_unsched_minutes must be between 1-60")
	}
	if settings.ThresholdCount < 1 || settings.ThresholdCount > 10 {
		return fmt.Errorf("threshold_count must be between 1-10")
	}
	if settings.ThresholdWindowMinutes < 1 || settings.ThresholdWindowMinutes > 60 {
		return fmt.Errorf("threshold_window_minutes must be between 1-60")
	}

	switch settings.Action {
	case StreamTimeoutActionTempUnsched, StreamTimeoutActionError, StreamTimeoutActionNone:
		// valid
	default:
		return fmt.Errorf("invalid action: %s", settings.Action)
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal stream timeout settings: %w", err)
	}

	return s.settingRepo.Set(ctx, SettingKeyStreamTimeoutSettings, string(data))
}

// GetDefaultPlatformQuotas 读取系统全局 platform quota JSON key，返回 4 platform x 3 window 的设置。
// 永远返回包含全部 4 platform key 的 map（值可能为零值/nil 字段，表示"上层未配置 = 不限制"）。
//
// 使用单个 JSON key（default_platform_quotas），一次 DB roundtrip，消除旧 12-KV 格式的 N+1 问题。
// 容错语义：取值失败或 unmarshal 失败 → 返回补齐 4 key 的空 map（fail-open，注册不被阻断）。
func (s *SettingService) GetDefaultPlatformQuotas(ctx context.Context) (map[string]*DefaultPlatformQuotaSetting, error) {
	out := map[string]*DefaultPlatformQuotaSetting{
		"anthropic":   {},
		"openai":      {},
		"gemini":      {},
		"antigravity": {},
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyDefaultPlatformQuotas)
	if err != nil || raw == "" {
		return out, nil // 无配置 = 全部不限制
	}
	parsed := map[string]*DefaultPlatformQuotaSetting{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		slog.Warn("[Setting] unmarshal default_platform_quotas failed (fail-open)", "error", err)
		return out, nil
	}
	for _, platform := range AllowedQuotaPlatforms {
		if v := parsed[platform]; v != nil {
			out[platform] = v
		}
	}
	return out, nil // 补齐 4 platform key，保持与旧实现一致的下游契约
}

// GetAuthSourcePlatformQuotas 读取指定 auth source 的 platform quota 覆盖（仅返回有配置的平台，override 语义）。
func (s *SettingService) GetAuthSourcePlatformQuotas(ctx context.Context, source string) map[string]*DefaultPlatformQuotaSetting {
	out := map[string]*DefaultPlatformQuotaSetting{}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyAuthSourcePlatformQuotas(source))
	if err != nil || raw == "" {
		return out // 无 override
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		slog.Warn("[Setting] unmarshal auth source platform quotas failed (fail-open)", "source", source, "error", err)
		return map[string]*DefaultPlatformQuotaSetting{}
	}
	return out // 仅含已配置平台，保持 override 语义
}

// mergePlatformQuotaDefaults 按字段级 patch：src 中非 nil 字段覆盖 dst。
// 区分 nil（"未配置"，保留 dst）vs &0.0（"显式禁用"，覆盖 dst 为 0）
func mergePlatformQuotaDefaults(dst, src *DefaultPlatformQuotaSetting) {
	if src == nil || dst == nil {
		return
	}
	if src.DailyLimitUSD != nil {
		dst.DailyLimitUSD = src.DailyLimitUSD
	}
	if src.WeeklyLimitUSD != nil {
		dst.WeeklyLimitUSD = src.WeeklyLimitUSD
	}
	if src.MonthlyLimitUSD != nil {
		dst.MonthlyLimitUSD = src.MonthlyLimitUSD
	}
}

// GetWeChatExportCostPerArticle 返回微信导出每篇文章的成本
// 实现 CostPerArticleGetter 接口
func (s *SettingService) GetWeChatExportCostPerArticle() float64 {
	ctx := context.Background()
	val, err := s.settingRepo.GetValue(ctx, "wechat_export_cost_per_article")
	if err != nil || val == "" {
		return 0 // 默认值为 0（不计费）
	}
	cost, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return cost
}
