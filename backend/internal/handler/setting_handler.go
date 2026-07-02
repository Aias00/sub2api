package handler

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler 公开设置处理器（无需认证）
type SettingHandler struct {
	settingService           *service.SettingService
	notificationEmailService *service.NotificationEmailService
	version                  string
}

// NewSettingHandler 创建公开设置处理器
func NewSettingHandler(settingService *service.SettingService, version string) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
		version:        version,
	}
}

// SetNotificationEmailService attaches the public notification email service without
// changing the constructor signature used by existing tests.
func (h *SettingHandler) SetNotificationEmailService(notificationEmailService *service.NotificationEmailService) {
	h.notificationEmailService = notificationEmailService
}

// GetPublicSettings 获取公开设置
// GET /api/v1/settings/public
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PublicSettings{
		RegistrationEnabled:              settings.RegistrationEnabled,
		EmailVerifyEnabled:               settings.EmailVerifyEnabled,
		ForceEmailOnThirdPartySignup:     settings.ForceEmailOnThirdPartySignup,
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
		LoginAgreementDocuments:          publicLoginAgreementDocumentsToDTO(settings.LoginAgreementDocuments),
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
		ModelPlazaItems:                  dto.ParseModelPlazaItems(settings.ModelPlazaItems),
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
		CustomMenuItems:                  dto.ParseUserVisibleMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                  dto.ParseCustomEndpoints(settings.CustomEndpoints),
		WeChatOAuthEnabled:               settings.WeChatOAuthEnabled,
		WeChatOAuthOpenEnabled:           settings.WeChatOAuthOpenEnabled,
		WeChatOAuthMPEnabled:             settings.WeChatOAuthMPEnabled,
		WeChatOAuthMobileEnabled:         settings.WeChatOAuthMobileEnabled,
		GitHubOAuthEnabled:               settings.GitHubOAuthEnabled,
		GoogleOAuthEnabled:               settings.GoogleOAuthEnabled,
		BackendModeEnabled:               settings.BackendModeEnabled,
		PaymentEnabled:                   settings.PaymentEnabled,
		Version:                          h.version,
		DefaultLocale:                    settings.WebDefaultLocale,
		ServerTimezone:                   timezone.Name(),
		ServerUTCOffset:                  timezone.UTCOffset(),
		BalanceLowNotifyEnabled:          settings.BalanceLowNotifyEnabled,
		AccountQuotaNotifyEnabled:        settings.AccountQuotaNotifyEnabled,
		BalanceLowNotifyThreshold:        settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:      settings.BalanceLowNotifyRechargeURL,

		ChannelMonitorEnabled:                settings.ChannelMonitorEnabled,
		ChannelMonitorDefaultIntervalSeconds: settings.ChannelMonitorDefaultIntervalSeconds,

		AvailableChannelsEnabled: settings.AvailableChannelsEnabled,

		AffiliateEnabled: settings.AffiliateEnabled,

		RiskControlEnabled: settings.RiskControlEnabled,

		PromptCasesTitle:           settings.PromptCasesTitle,
		PromptCasesDescription:     settings.PromptCasesDescription,
		PromptTemplatesTitle:       settings.PromptTemplatesTitle,
		PromptTemplatesDescription: settings.PromptTemplatesDescription,
		PromptCatalogShellConfig:   settings.PromptCatalogShellConfig,
		WorkspaceShellConfig:       settings.WorkspaceShellConfig,
		PricingTitle:               settings.PricingTitle,
		PricingDescription:         settings.PricingDescription,
		PricingShellConfig:         settings.PricingShellConfig,
		PaymentShellConfig:         settings.PaymentShellConfig,
		PricingCurrencySymbol:      settings.PricingCurrencySymbol,
		CreditsTitle:               settings.CreditsTitle,
		CreditsDescription:         settings.CreditsDescription,
		CreditsPurchaseLabel:       settings.CreditsPurchaseLabel,
		CreditsBalanceLabel:        settings.CreditsBalanceLabel,
		CreditsPerBalance:          settings.CreditsPerBalance,
		CreditsShellConfig:         settings.CreditsShellConfig,
		GoogleAnalyticsID:          settings.GoogleAnalyticsID,
		AffonsoEnabled:             settings.AffonsoEnabled,
		AffonsoID:                  settings.AffonsoID,
		AffonsoCookieDuration:      settings.AffonsoCookieDuration,
		PromoteKitEnabled:          settings.PromoteKitEnabled,
		PromoteKitID:               settings.PromoteKitID,
		CrispEnabled:               settings.CrispEnabled,
		CrispWebsiteID:             settings.CrispWebsiteID,

		AllowUserViewErrorRequests: settings.AllowUserViewErrorRequests,
	})
}

// GetSiteLogo serves the configured uploaded site logo as a normal image URL for
// clients that do not support data URI images, especially email clients.
func (h *SettingHandler) GetSiteLogo(c *gin.Context) {
	logo, err := h.settingService.GetSiteLogoImage(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if logo == nil {
		response.NotFound(c, "site logo not configured")
		return
	}
	if siteLogoETagMatches(c.GetHeader("If-None-Match"), logo.ETag) {
		c.Status(http.StatusNotModified)
		c.Writer.WriteHeaderNow()
		return
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.Header("ETag", logo.ETag)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, logo.ContentType, logo.Data)
}

// GetAdsTxt serves the public web marketing ads.txt content from public settings.
// GET /ads.txt
func (h *SettingHandler) GetAdsTxt(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	adsenseCode := strings.TrimSpace(settings.WebAdsenseCode)
	if adsenseCode == "" {
		response.NotFound(c, "ads.txt is not configured")
		return
	}

	adsenseCode = strings.TrimPrefix(adsenseCode, "ca-")
	c.Header("Cache-Control", "public, max-age=300")
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(http.StatusOK, "google.com, %s, DIRECT, f08c47fec0942fa0", adsenseCode)
}

// GetFaviconICO redirects the legacy browser favicon request to the configured
// public web favicon or the embedded Vue static fallback.
// GET /favicon.ico
func (h *SettingHandler) GetFaviconICO(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Redirect(http.StatusPermanentRedirect, publicFaviconRedirectTarget(settings.WebAppFavicon))
}

// GetRobotsTxt serves robots.txt for the public frontend shell.
// GET /robots.txt
func (h *SettingHandler) GetRobotsTxt(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	baseURL := publicBaseURLFromSettingsOrRequest(settings.WebAppURL, c)
	c.Header("Cache-Control", "public, max-age=300")
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(http.StatusOK, strings.Join([]string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /*?*q=",
		"Disallow: /settings/*",
		"Disallow: /admin/*",
		"Disallow: /api/*",
		"Sitemap: " + baseURL + "/sitemap.xml",
		"",
	}, "\n"))
}

// GetSitemapXML serves the public sitemap for the Vue web routes.
// GET /sitemap.xml
func (h *SettingHandler) GetSitemapXML(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	baseURL := publicBaseURLFromSettingsOrRequest(settings.WebAppURL, c)
	entries := webSitemapURLs(baseURL, settings.WebDefaultLocale)
	now := time.Now().UTC().Format(time.RFC3339)

	var buf bytes.Buffer
	_, _ = buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	_, _ = buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, entry := range entries {
		_, _ = buf.WriteString("  <url>\n")
		_, _ = buf.WriteString("    <loc>")
		_ = xml.EscapeText(&buf, []byte(entry.loc))
		_, _ = buf.WriteString("</loc>\n")
		_, _ = buf.WriteString("    <lastmod>")
		_ = xml.EscapeText(&buf, []byte(now))
		_, _ = buf.WriteString("</lastmod>\n")
		_, _ = buf.WriteString("    <changefreq>")
		_ = xml.EscapeText(&buf, []byte(entry.changefreq))
		_, _ = buf.WriteString("</changefreq>\n")
		_, _ = buf.WriteString("    <priority>")
		_ = xml.EscapeText(&buf, []byte(entry.priority))
		_, _ = buf.WriteString("</priority>\n")
		_, _ = buf.WriteString("  </url>\n")
	}
	_, _ = buf.WriteString("</urlset>\n")

	c.Header("Cache-Control", "public, max-age=300")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", buf.Bytes())
}

type sitemapEntry struct {
	loc        string
	changefreq string
	priority   string
}

func webSitemapURLs(baseURL, defaultLocale string) []sitemapEntry {
	if defaultLocale == "" {
		defaultLocale = "en"
	}
	locales := []string{"en", "zh"}
	unlocalized := []string{"/docs", "/pricing", "/prompts", "/image-generator"}
	seen := make(map[string]bool, len(locales)+len(unlocalized))
	entries := make([]sitemapEntry, 0, len(locales)+len(unlocalized))

	for _, locale := range locales {
		path := "/"
		if locale != defaultLocale {
			path = "/" + locale
		}
		loc := baseURL + path
		if !seen[loc] {
			entries = append(entries, sitemapEntry{loc: loc, changefreq: "weekly", priority: "1"})
			seen[loc] = true
		}
	}

	for _, path := range unlocalized {
		loc := baseURL + path
		if !seen[loc] {
			entries = append(entries, sitemapEntry{loc: loc, changefreq: "monthly", priority: "0.7"})
			seen[loc] = true
		}
	}

	return entries
}

func publicBaseURLFromSettingsOrRequest(configured string, c *gin.Context) string {
	if baseURL := strings.TrimRight(strings.TrimSpace(configured), "/"); baseURL != "" {
		return baseURL
	}

	proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if proto == "" {
		proto = "http"
		if c.Request != nil && c.Request.TLS != nil {
			proto = "https"
		}
	}
	if comma := strings.Index(proto, ","); comma >= 0 {
		proto = strings.TrimSpace(proto[:comma])
	}
	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" && c.Request != nil {
		host = c.Request.Host
	}
	if comma := strings.Index(host, ","); comma >= 0 {
		host = strings.TrimSpace(host[:comma])
	}
	if host == "" {
		return ""
	}
	return strings.TrimRight(proto+"://"+host, "/")
}

func publicFaviconRedirectTarget(configured string) string {
	target := strings.TrimSpace(configured)
	if target == "" || target == "/favicon.ico" {
		return "/favicon.svg"
	}
	return target
}

func siteLogoETagMatches(ifNoneMatch, etag string) bool {
	ifNoneMatch = strings.TrimSpace(ifNoneMatch)
	etag = strings.TrimSpace(etag)
	if ifNoneMatch == "" || etag == "" {
		return false
	}
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || candidate == etag {
			return true
		}
	}
	return false
}

func publicLoginAgreementDocumentsToDTO(items []service.LoginAgreementDocument) []dto.LoginAgreementDocument {
	result := make([]dto.LoginAgreementDocument, 0, len(items))
	for _, item := range items {
		result = append(result, dto.LoginAgreementDocument{
			ID:        item.ID,
			Title:     item.Title,
			ContentMD: item.ContentMD,
		})
	}
	return result
}
