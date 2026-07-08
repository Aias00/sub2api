package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Aias00/cloudbase/internal/pkg/timezone"
)

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
		SettingKeyImagePromptFilterConfig,
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
		SettingKeyAllowUserViewErrorRequests,
	}

	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get public settings: %w", err)
	}

	gitHubEnabled := s.emailOAuthPublicEnabled(settings, "github")
	googleEnabled := s.emailOAuthPublicEnabled(settings, "google")
	siteName := s.getStringOrDefault(settings, SettingKeySiteName, "Cloudbase")
	siteSubtitle := s.getStringOrDefault(settings, SettingKeySiteSubtitle, "AI Gateway and Business Operations Platform")
	webAppName := strings.TrimSpace(settings[SettingKeyWebAppName])
	if webAppName == "" {
		webAppName = siteName
	}
	webAppDescription := strings.TrimSpace(settings[SettingKeyWebAppDescription])
	if webAppDescription == "" {
		webAppDescription = siteSubtitle
	}
	webAppLogo := strings.TrimSpace(settings[SettingKeyWebAppLogo])
	if webAppLogo == "" {
		webAppLogo = strings.TrimSpace(settings[SettingKeySiteLogo])
	}
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
		WebAppURL:                        strings.TrimSpace(settings[SettingKeyWebAppURL]),
		WebAppName:                       webAppName,
		WebAppDescription:                webAppDescription,
		WebAppLogo:                       webAppLogo,
		WebAppFavicon:                    strings.TrimSpace(settings[SettingKeyWebAppFavicon]),
		WebAppPreviewImage:               strings.TrimSpace(settings[SettingKeyWebAppPreviewImage]),
		WebTheme:                         strings.TrimSpace(settings[SettingKeyWebTheme]),
		WebAppearance:                    strings.TrimSpace(settings[SettingKeyWebAppearance]),
		WebDefaultLocale:                 strings.TrimSpace(settings[SettingKeyWebDefaultLocale]),
		PromptCasesTitle:                 strings.TrimSpace(settings[SettingKeyPromptCasesTitle]),
		WebPromptCasesTitle:              strings.TrimSpace(settings[SettingKeyPromptCasesTitle]),
		PromptCasesDescription:           strings.TrimSpace(settings[SettingKeyPromptCasesDescription]),
		WebPromptCasesDescription:        strings.TrimSpace(settings[SettingKeyPromptCasesDescription]),
		PromptTemplatesTitle:             strings.TrimSpace(settings[SettingKeyPromptTemplatesTitle]),
		WebPromptTemplatesTitle:          strings.TrimSpace(settings[SettingKeyPromptTemplatesTitle]),
		PromptTemplatesDescription:       strings.TrimSpace(settings[SettingKeyPromptTemplatesDescription]),
		WebPromptTemplatesDescription:    strings.TrimSpace(settings[SettingKeyPromptTemplatesDescription]),
		PromptCatalogShellConfig:         promptCatalogShellConfigSetting(settings[SettingKeyPromptCatalogShellConfig]),
		WorkspaceShellConfig:             workspaceShellConfigSetting(settings[SettingKeyWorkspaceShellConfig]),
		WebWorkspaceShellConfig:          workspaceShellConfigSetting(settings[SettingKeyWorkspaceShellConfig]),
		ImagePromptFilterConfig:          strings.TrimSpace(settings[SettingKeyImagePromptFilterConfig]),
		WebImagePromptFilterConfig:       strings.TrimSpace(settings[SettingKeyImagePromptFilterConfig]),
		PricingTitle:                     strings.TrimSpace(settings[SettingKeyPricingTitle]),
		WebPricingTitle:                  strings.TrimSpace(settings[SettingKeyPricingTitle]),
		PricingDescription:               strings.TrimSpace(settings[SettingKeyPricingDescription]),
		WebPricingDescription:            strings.TrimSpace(settings[SettingKeyPricingDescription]),
		PricingShellConfig:               pricingShellConfigSetting(settings[SettingKeyPricingShellConfig]),
		WebPricingShellConfig:            pricingShellConfigSetting(settings[SettingKeyPricingShellConfig]),
		PaymentShellConfig:               paymentShellConfigSetting(settings[SettingKeyPaymentShellConfig]),
		WebPaymentShellConfig:            paymentShellConfigSetting(settings[SettingKeyPaymentShellConfig]),
		PricingCurrencySymbol:            pricingCurrencySymbolSetting(settings[SettingKeyPricingCurrencySymbol]),
		WebPricingCurrencySymbol:         pricingCurrencySymbolSetting(settings[SettingKeyPricingCurrencySymbol]),
		CreditsTitle:                     strings.TrimSpace(settings[SettingKeyCreditsTitle]),
		WebCreditsTitle:                  strings.TrimSpace(settings[SettingKeyCreditsTitle]),
		CreditsDescription:               strings.TrimSpace(settings[SettingKeyCreditsDescription]),
		WebCreditsDescription:            strings.TrimSpace(settings[SettingKeyCreditsDescription]),
		CreditsPurchaseLabel:             strings.TrimSpace(settings[SettingKeyCreditsPurchaseLabel]),
		WebCreditsPurchaseLabel:          strings.TrimSpace(settings[SettingKeyCreditsPurchaseLabel]),
		CreditsBalanceLabel:              strings.TrimSpace(settings[SettingKeyCreditsBalanceLabel]),
		WebCreditsBalanceLabel:           strings.TrimSpace(settings[SettingKeyCreditsBalanceLabel]),
		WebCreditsPerBalance:             creditsPerBalanceSetting(settings[SettingKeyCreditsPerBalance]),
		CreditsPerBalance:                creditsPerBalanceSetting(settings[SettingKeyCreditsPerBalance]),
		CreditsShellConfig:               creditsShellConfigSetting(settings[SettingKeyCreditsShellConfig]),
		WebLocaleDetectEnabled:           settings[SettingKeyWebLocaleDetectEnabled] == "true",
		WebEmailAuthVisible:              webEmailVisible,
		WebGoogleAuthVisible:             webGoogleVisible,
		WebGitHubAuthVisible:             webGitHubVisible,
		GoogleAnalyticsID:                strings.TrimSpace(settings[SettingKeyWebGoogleAnalyticsID]),
		WebGoogleAnalyticsID:             strings.TrimSpace(settings[SettingKeyWebGoogleAnalyticsID]),
		WebClarityID:                     strings.TrimSpace(settings[SettingKeyWebClarityID]),
		WebPlausibleDomain:               strings.TrimSpace(settings[SettingKeyWebPlausibleDomain]),
		WebPlausibleSrc:                  strings.TrimSpace(settings[SettingKeyWebPlausibleSrc]),
		WebOpenPanelClientID:             strings.TrimSpace(settings[SettingKeyWebOpenPanelClientID]),
		PublicIntegrationsEnabled:        !isFalseSettingValue(settings[SettingKeyWebPublicIntegrationsEnabled]),
		WebPublicIntegrationsEnabled:     !isFalseSettingValue(settings[SettingKeyWebPublicIntegrationsEnabled]),
		WebVercelAnalyticsEnabled:        settings[SettingKeyWebVercelAnalyticsEnabled] == "true",
		WebAdsenseCode:                   strings.TrimSpace(settings[SettingKeyWebAdsenseCode]),
		AffonsoEnabled:                   settings[SettingKeyWebAffonsoEnabled] == "true",
		WebAffonsoEnabled:                settings[SettingKeyWebAffonsoEnabled] == "true",
		AffonsoID:                        strings.TrimSpace(settings[SettingKeyWebAffonsoID]),
		WebAffonsoID:                     strings.TrimSpace(settings[SettingKeyWebAffonsoID]),
		AffonsoCookieDuration:            webAffonsoCookieDurationSetting(settings[SettingKeyWebAffonsoCookieDuration]),
		WebAffonsoCookieDuration:         webAffonsoCookieDurationSetting(settings[SettingKeyWebAffonsoCookieDuration]),
		PromoteKitEnabled:                settings[SettingKeyWebPromoteKitEnabled] == "true",
		WebPromoteKitEnabled:             settings[SettingKeyWebPromoteKitEnabled] == "true",
		PromoteKitID:                     strings.TrimSpace(settings[SettingKeyWebPromoteKitID]),
		WebPromoteKitID:                  strings.TrimSpace(settings[SettingKeyWebPromoteKitID]),
		CrispEnabled:                     settings[SettingKeyWebCrispEnabled] == "true",
		WebCrispEnabled:                  settings[SettingKeyWebCrispEnabled] == "true",
		CrispWebsiteID:                   strings.TrimSpace(settings[SettingKeyWebCrispWebsiteID]),
		WebCrispWebsiteID:                strings.TrimSpace(settings[SettingKeyWebCrispWebsiteID]),
		WebTawkEnabled:                   settings[SettingKeyWebTawkEnabled] == "true",
		WebTawkPropertyID:                strings.TrimSpace(settings[SettingKeyWebTawkPropertyID]),
		WebTawkWidgetID:                  strings.TrimSpace(settings[SettingKeyWebTawkWidgetID]),
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
		WeChatOAuthEnabled:               false,
		WeChatOAuthOpenEnabled:           false,
		WeChatOAuthMPEnabled:             false,
		WeChatOAuthMobileEnabled:         false,
		BackendModeEnabled:               settings[SettingKeyBackendModeEnabled] == "true",
		PaymentEnabled:                   settings[SettingPaymentEnabled] == "true",
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

		AllowUserViewErrorRequests: settings[SettingKeyAllowUserViewErrorRequests] == "true",
	}, nil
}

func (s *SettingService) GetSiteLogoImage(ctx context.Context) (*SiteLogoImage, error) {
	if s == nil || s.settingRepo == nil {
		return nil, ErrServiceUnavailable
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeySiteLogo)
	if err != nil {
		return nil, fmt.Errorf("get site logo: %w", err)
	}
	logo, ok := parseSiteLogoDataURL(raw)
	if !ok {
		return nil, nil
	}
	return &SiteLogoImage{
		ContentType: logo.ContentType,
		Data:        logo.Data,
		ETag:        logo.ETag,
	}, nil
}

func (s *SettingService) GetEmailLogoURL(ctx context.Context) string {
	if s == nil || s.settingRepo == nil {
		return ""
	}
	return resolveEmailLogoURL(ctx, s.settingRepo)
}

func parseSiteLogoDataURL(raw string) (parsedSiteLogoDataURL, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(strings.ToLower(raw), "data:image/") {
		return parsedSiteLogoDataURL{}, false
	}
	meta, encoded, ok := strings.Cut(raw, ",")
	if !ok || !strings.Contains(strings.ToLower(meta), ";base64") {
		return parsedSiteLogoDataURL{}, false
	}
	contentType := strings.TrimPrefix(strings.TrimSpace(strings.Split(meta, ";")[0]), "data:")
	if contentType == "" {
		return parsedSiteLogoDataURL{}, false
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(data) == 0 {
		return parsedSiteLogoDataURL{}, false
	}
	sum := sha256.Sum256(data)
	return parsedSiteLogoDataURL{
		ContentType: contentType,
		Data:        data,
		ETag:        `"` + hex.EncodeToString(sum[:]) + `"`,
	}, true
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

func workspaceShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultWorkspaceShellConfig
	}
	return value
}

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

func dashboardShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultDashboardShellConfig
	}
	return value
}

func pricingShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultPricingShellConfig
	}
	return value
}

func paymentShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultPaymentShellConfig
	}
	return value
}

func creditsShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultCreditsShellConfig
	}
	return value
}

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

func modelPlazaShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultModelPlazaShellConfig
	}
	return value
}

func docsShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultDocsShellConfig
	}
	return value
}

func docsContentBasePathSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultDocsContentBasePath
	}
	return value
}

func legalDocumentShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultLegalDocumentShellConfig
	}
	return value
}

func keyUsageShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultKeyUsageShellConfig
	}
	return value
}

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

func apiGuideShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAPIGuideShellConfig
	}
	return value
}

func apiTestShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAPITestShellConfig
	}
	return value
}

func availableGroupsShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAvailableGroupsShellConfig
	}
	return value
}

func redeemShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultRedeemShellConfig
	}
	return value
}

func affiliateShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAffiliateShellConfig
	}
	return value
}

func availableChannelsShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAvailableChannelsShellConfig
	}
	return value
}

func channelStatusShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultChannelStatusShellConfig
	}
	return value
}

func customPageShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultCustomPageShellConfig
	}
	return value
}

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

func authShellConfigSetting(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultAuthShellConfig
	}
	return value
}

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
		WeChatOAuthEnabled:               false,
		WeChatOAuthOpenEnabled:           false,
		WeChatOAuthMPEnabled:             false,
		WeChatOAuthMobileEnabled:         false,
		GitHubOAuthEnabled:               settings.GitHubOAuthEnabled,
		GoogleOAuthEnabled:               settings.GoogleOAuthEnabled,
		BackendModeEnabled:               settings.BackendModeEnabled,
		PaymentEnabled:                   settings.PaymentEnabled,
		Version:                          s.version,
		DefaultLocale:                    settings.WebDefaultLocale,
		ServerTimezone:                   timezone.Name(),
		ServerUTCOffset:                  timezone.UTCOffset(),
		BalanceLowNotifyEnabled:          settings.BalanceLowNotifyEnabled,
		AccountQuotaNotifyEnabled:        settings.AccountQuotaNotifyEnabled,
		BalanceLowNotifyThreshold:        settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:      settings.BalanceLowNotifyRechargeURL,

		ChannelMonitorEnabled:                settings.ChannelMonitorEnabled,
		ChannelMonitorDefaultIntervalSeconds: settings.ChannelMonitorDefaultIntervalSeconds,
		AvailableChannelsEnabled:             settings.AvailableChannelsEnabled,
		AffiliateEnabled:                     settings.AffiliateEnabled,
		RiskControlEnabled:                   settings.RiskControlEnabled,
		AllowUserViewErrorRequests:           settings.AllowUserViewErrorRequests,
		PromptCasesTitle:                     settings.PromptCasesTitle,
		PromptCasesDescription:               settings.PromptCasesDescription,
		PromptTemplatesTitle:                 settings.PromptTemplatesTitle,
		PromptTemplatesDescription:           settings.PromptTemplatesDescription,
		PromptCatalogShellConfig:             settings.PromptCatalogShellConfig,
		WorkspaceShellConfig:                 settings.WorkspaceShellConfig,
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
		AffonsoEnabled:                       settings.AffonsoEnabled,
		AffonsoID:                            settings.AffonsoID,
		AffonsoCookieDuration:                settings.AffonsoCookieDuration,
		PromoteKitEnabled:                    settings.PromoteKitEnabled,
		PromoteKitID:                         settings.PromoteKitID,
		CrispEnabled:                         settings.CrispEnabled,
		CrispWebsiteID:                       settings.CrispWebsiteID,
	}, nil
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

// GetCustomMenuItemsRaw returns the raw JSON string of custom_menu_items setting.
func (s *SettingService) GetCustomMenuItemsRaw(ctx context.Context) string {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyCustomMenuItems)
	if err != nil {
		return "[]"
	}
	return value
}
