package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Aias00/cloudbase/internal/config"
	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
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
