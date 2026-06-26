//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type settingHandlerPublicRepoStub struct {
	values map[string]string
}

func (s *settingHandlerPublicRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *settingHandlerPublicRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	return s.values[key], nil
}

func (s *settingHandlerPublicRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingHandlerPublicRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *settingHandlerPublicRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *settingHandlerPublicRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingHandlerPublicRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingHandler_GetPublicSettings_ExposesForceEmailOnThirdPartySignup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyForceEmailOnThirdPartySignup: "true",
		},
	}
	h := NewSettingHandler(service.NewSettingService(repo, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)

	h.GetPublicSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ForceEmailOnThirdPartySignup bool `json:"force_email_on_third_party_signup"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.True(t, resp.Data.ForceEmailOnThirdPartySignup)
}

func TestSettingHandler_GetPublicSettings_ExposesPasswordMinLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyPasswordMinLength: "12",
		},
	}
	h := NewSettingHandler(service.NewSettingService(repo, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)

	h.GetPublicSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			PasswordMinLength int `json:"password_min_length"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, 12, resp.Data.PasswordMinLength)
}

func TestSettingHandler_GetPublicSettings_ExposesGenericRuntimeSettingAliases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyPromptCasesTitle:           "Cases title",
			service.SettingKeyPromptCasesDescription:     "Cases description",
			service.SettingKeyPromptTemplatesTitle:       "Templates title",
			service.SettingKeyPromptTemplatesDescription: "Templates description",
			service.SettingKeyPromptCatalogShellConfig:   `{"zh":{"labels":{"total":"总数"}}}`,
			service.SettingKeyWorkspaceShellConfig:       `{"zh":{"title":"工作台"}}`,
			service.SettingKeyPricingTitle:               "Pricing title",
			service.SettingKeyPricingDescription:         "Pricing description",
			service.SettingKeyPricingShellConfig:         `{"zh":{"button":{"title":"选择"}}}`,
			service.SettingKeyPricingCurrencySymbol:      "$",
			service.SettingKeyCreditsTitle:               "Credits title",
			service.SettingKeyCreditsDescription:         "Credits description",
			service.SettingKeyCreditsPurchaseLabel:       "Buy credits",
			service.SettingKeyCreditsBalanceLabel:        "Balance: {balance}",
			service.SettingKeyCreditsPerBalance:          "12",
			service.SettingKeyCreditsShellConfig:         `{"en":{"actions":{"title":"Balance actions"}}}`,
			service.SettingKeyWebGoogleAnalyticsID:       "G-WEB",
			service.SettingKeyWebAffonsoEnabled:          "true",
			service.SettingKeyWebAffonsoID:               "affonso-public",
			service.SettingKeyWebAffonsoCookieDuration:   "45",
			service.SettingKeyWebPromoteKitEnabled:       "true",
			service.SettingKeyWebPromoteKitID:            "promotekit-public",
			service.SettingKeyWebCrispEnabled:            "true",
			service.SettingKeyWebCrispWebsiteID:          "crisp-public",
		},
	}
	h := NewSettingHandler(service.NewSettingService(repo, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)

	h.GetPublicSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			PromptCasesTitle           string `json:"prompt_cases_title"`
			PromptCasesDescription     string `json:"prompt_cases_description"`
			PromptTemplatesTitle       string `json:"prompt_templates_title"`
			PromptTemplatesDescription string `json:"prompt_templates_description"`
			PromptCatalogShellConfig   string `json:"prompt_catalog_shell_config"`
			WorkspaceShellConfig       string `json:"workspace_shell_config"`
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
			AffonsoEnabled             bool   `json:"affonso_enabled"`
			AffonsoID                  string `json:"affonso_id"`
			AffonsoCookieDuration      string `json:"affonso_cookie_duration"`
			PromoteKitEnabled          bool   `json:"promotekit_enabled"`
			PromoteKitID               string `json:"promotekit_id"`
			CrispEnabled               bool   `json:"crisp_enabled"`
			CrispWebsiteID             string `json:"crisp_website_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, "Cases title", resp.Data.PromptCasesTitle)
	require.Equal(t, "Cases description", resp.Data.PromptCasesDescription)
	require.Equal(t, "Templates title", resp.Data.PromptTemplatesTitle)
	require.Equal(t, "Templates description", resp.Data.PromptTemplatesDescription)
	require.Equal(t, `{"zh":{"labels":{"total":"总数"}}}`, resp.Data.PromptCatalogShellConfig)
	require.Equal(t, `{"zh":{"title":"工作台"}}`, resp.Data.WorkspaceShellConfig)
	require.Equal(t, "Pricing title", resp.Data.PricingTitle)
	require.Equal(t, "Pricing description", resp.Data.PricingDescription)
	require.Equal(t, `{"zh":{"button":{"title":"选择"}}}`, resp.Data.PricingShellConfig)
	require.NotEmpty(t, resp.Data.PaymentShellConfig)
	require.Equal(t, "$", resp.Data.PricingCurrencySymbol)
	require.Equal(t, "Credits title", resp.Data.CreditsTitle)
	require.Equal(t, "Credits description", resp.Data.CreditsDescription)
	require.Equal(t, "Buy credits", resp.Data.CreditsPurchaseLabel)
	require.Equal(t, "Balance: {balance}", resp.Data.CreditsBalanceLabel)
	require.Equal(t, "12", resp.Data.CreditsPerBalance)
	require.Equal(t, `{"en":{"actions":{"title":"Balance actions"}}}`, resp.Data.CreditsShellConfig)
	require.Equal(t, "G-WEB", resp.Data.GoogleAnalyticsID)
	require.True(t, resp.Data.AffonsoEnabled)
	require.Equal(t, "affonso-public", resp.Data.AffonsoID)
	require.Equal(t, "45", resp.Data.AffonsoCookieDuration)
	require.True(t, resp.Data.PromoteKitEnabled)
	require.Equal(t, "promotekit-public", resp.Data.PromoteKitID)
	require.True(t, resp.Data.CrispEnabled)
	require.Equal(t, "crisp-public", resp.Data.CrispWebsiteID)
	require.NotContains(t, recorder.Body.String(), "touch_prompt_cases_title")
	require.NotContains(t, recorder.Body.String(), "touch_workspace_shell_config")
	require.NotContains(t, recorder.Body.String(), "touch_pricing_title")
	require.NotContains(t, recorder.Body.String(), "touch_credits_title")
}

func TestSettingHandler_GetSiteLogo_ServesConfiguredDataLogo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeySiteLogo: "data:image/png;base64,aGVsbG8=",
		},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/site-logo", nil)

	h.GetSiteLogo(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "image/png", recorder.Header().Get("Content-Type"))
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	require.NotEmpty(t, recorder.Header().Get("ETag"))
	require.Equal(t, "hello", recorder.Body.String())
}

func TestSettingHandler_GetSiteLogo_ReturnsNotModifiedForMatchingETag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeySiteLogo: "data:image/png;base64,aGVsbG8=",
		},
	}, &config.Config{}), "test-version")

	first := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(first)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/site-logo", nil)
	h.GetSiteLogo(c)
	etag := first.Header().Get("ETag")
	require.NotEmpty(t, etag)

	second := httptest.NewRecorder()
	c, _ = gin.CreateTestContext(second)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/site-logo", nil)
	c.Request.Header.Set("If-None-Match", etag)
	h.GetSiteLogo(c)

	require.Equal(t, http.StatusNotModified, second.Code)
	require.Empty(t, second.Body.String())
}

func TestSettingHandler_GetAdsTxt_ServesWebAdsenseCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyWebAdsenseCode: " ca-pub-web ",
		},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/ads.txt", nil)

	h.GetAdsTxt(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "google.com, pub-web, DIRECT, f08c47fec0942fa0", recorder.Body.String())
}

func TestSettingHandler_GetAdsTxt_ReturnsNotFoundWhenUnset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/ads.txt", nil)

	h.GetAdsTxt(c)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestSettingHandler_GetFaviconICO_RedirectsToConfiguredWebFavicon(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyWebAppFavicon: " https://static.example.com/favicon.ico ",
		},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)

	h.GetFaviconICO(c)

	require.Equal(t, http.StatusPermanentRedirect, recorder.Code)
	require.Equal(t, "https://static.example.com/favicon.ico", recorder.Header().Get("Location"))
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
}

func TestSettingHandler_GetFaviconICO_FallsBackToEmbeddedFavicon(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)

	h.GetFaviconICO(c)

	require.Equal(t, http.StatusPermanentRedirect, recorder.Code)
	require.Equal(t, "/favicon.svg", recorder.Header().Get("Location"))
}

func TestSettingHandler_GetRobotsTxt_UsesWebAppURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyWebAppURL: " https://web.example.com/ ",
		},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/robots.txt", nil)

	h.GetRobotsTxt(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	body := recorder.Body.String()
	require.Contains(t, body, "User-agent: *")
	require.Contains(t, body, "Allow: /")
	require.Contains(t, body, "Disallow: /api/*")
	require.Contains(t, body, "Sitemap: https://web.example.com/sitemap.xml")
}

func TestSettingHandler_GetRobotsTxt_FallsBackToRequestOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	c.Request.Host = "sub2api.example.com"
	c.Request.Header.Set("X-Forwarded-Proto", "https")

	h.GetRobotsTxt(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Sitemap: https://sub2api.example.com/sitemap.xml")
}

func TestSettingHandler_GetSitemapXML_UsesWebAppURLAndDefaultLocale(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyWebAppURL:        "https://web.example.com/",
			service.SettingKeyWebDefaultLocale: "zh",
		},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)

	h.GetSitemapXML(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "application/xml; charset=utf-8", recorder.Header().Get("Content-Type"))
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	body := recorder.Body.String()
	require.Contains(t, body, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	require.Contains(t, body, "<loc>https://web.example.com/en</loc>")
	require.Contains(t, body, "<loc>https://web.example.com/</loc>")
	require.Contains(t, body, "<loc>https://web.example.com/docs</loc>")
	require.Contains(t, body, "<loc>https://web.example.com/pricing</loc>")
	require.Contains(t, body, "<loc>https://web.example.com/prompts</loc>")
	require.Contains(t, body, "<loc>https://web.example.com/image-generator</loc>")
}

func TestSettingHandler_GetPublicSettings_ExposesWeChatOAuthModeCapabilities(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyWeChatConnectEnabled:             "true",
			service.SettingKeyWeChatConnectMPAppID:             "wx-mp-app",
			service.SettingKeyWeChatConnectMPAppSecret:         "wx-mp-secret",
			service.SettingKeyWeChatConnectMode:                "mp",
			service.SettingKeyWeChatConnectScopes:              "snsapi_base",
			service.SettingKeyWeChatConnectOpenEnabled:         "true",
			service.SettingKeyWeChatConnectMPEnabled:           "true",
			service.SettingKeyWeChatConnectRedirectURL:         "https://api.example.com/api/v1/auth/oauth/wechat/callback",
			service.SettingKeyWeChatConnectFrontendRedirectURL: "/auth/wechat/callback",
		},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)

	h.GetPublicSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			WeChatOAuthEnabled     bool `json:"wechat_oauth_enabled"`
			WeChatOAuthOpenEnabled bool `json:"wechat_oauth_open_enabled"`
			WeChatOAuthMPEnabled   bool `json:"wechat_oauth_mp_enabled"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.True(t, resp.Data.WeChatOAuthEnabled)
	require.True(t, resp.Data.WeChatOAuthOpenEnabled)
	require.True(t, resp.Data.WeChatOAuthMPEnabled)
}

func TestSettingHandler_GetPublicSettings_ExposesModelPlazaItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingHandler(service.NewSettingService(&settingHandlerPublicRepoStub{
		values: map[string]string{
			service.SettingKeyModelPlazaItems: `[{"id":"claude-opus-4-6","provider":"anthropic","title":"Claude Opus 4.6","visible":true,"sort_order":10}]`,
		},
	}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)

	h.GetPublicSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ModelPlazaItems []struct {
				ID      string `json:"id"`
				Title   string `json:"title"`
				Visible bool   `json:"visible"`
			} `json:"model_plaza_items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data.ModelPlazaItems, 1)
	require.Equal(t, "claude-opus-4-6", resp.Data.ModelPlazaItems[0].ID)
	require.Equal(t, "Claude Opus 4.6", resp.Data.ModelPlazaItems[0].Title)
	require.True(t, resp.Data.ModelPlazaItems[0].Visible)
}
