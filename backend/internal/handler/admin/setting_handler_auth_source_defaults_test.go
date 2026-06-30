package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type settingHandlerRepoStub struct {
	values      map[string]string
	lastUpdates map[string]string
}

func (s *settingHandlerRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *settingHandlerRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.values != nil {
		if value, ok := s.values[key]; ok {
			return value, nil
		}
	}
	return "", nil
}

func (s *settingHandlerRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingHandlerRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *settingHandlerRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	s.lastUpdates = make(map[string]string, len(settings))
	for key, value := range settings {
		s.lastUpdates[key] = value
		if s.values == nil {
			s.values = map[string]string{}
		}
		s.values[key] = value
	}
	return nil
}

func (s *settingHandlerRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.values))
	for key, value := range s.values {
		out[key] = value
	}
	return out, nil
}

func (s *settingHandlerRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

type failingAuthSourceSettingsRepoStub struct {
	values map[string]string
	err    error
}

func (s *failingAuthSourceSettingsRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *failingAuthSourceSettingsRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *failingAuthSourceSettingsRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *failingAuthSourceSettingsRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *failingAuthSourceSettingsRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	if _, ok := settings[service.SettingKeyAuthSourceDefaultEmailBalance]; ok {
		return s.err
	}
	for key, value := range settings {
		if s.values == nil {
			s.values = map[string]string{}
		}
		s.values[key] = value
	}
	return nil
}

func (s *failingAuthSourceSettingsRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.values))
	for key, value := range s.values {
		out[key] = value
	}
	return out, nil
}

func (s *failingAuthSourceSettingsRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingHandler_GetSettings_InjectsAuthSourceDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyRegistrationEnabled:                 "true",
			service.SettingKeyPromoCodeEnabled:                    "true",
			service.SettingKeyAuthSourceDefaultEmailBalance:       "9.5",
			service.SettingKeyAuthSourceDefaultEmailConcurrency:   "8",
			service.SettingKeyAuthSourceDefaultEmailSubscriptions: `[{"group_id":31,"validity_days":15}]`,
			service.SettingKeyForceEmailOnThirdPartySignup:        "true",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)

	handler.GetSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, 9.5, data["auth_source_default_email_balance"])
	require.Equal(t, float64(8), data["auth_source_default_email_concurrency"])
	require.Equal(t, true, data["force_email_on_third_party_signup"])

	subscriptions, ok := data["auth_source_default_email_subscriptions"].([]any)
	require.True(t, ok)
	require.Len(t, subscriptions, 1)
}

func TestSettingHandler_UpdateSettings_PreservesOmittedAuthSourceDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyRegistrationEnabled:                    "false",
			service.SettingKeyPromoCodeEnabled:                       "true",
			service.SettingKeyAuthSourceDefaultEmailBalance:          "9.5",
			service.SettingKeyAuthSourceDefaultEmailConcurrency:      "8",
			service.SettingKeyAuthSourceDefaultEmailSubscriptions:    `[{"group_id":31,"validity_days":15}]`,
			service.SettingKeyAuthSourceDefaultEmailGrantOnSignup:    "true",
			service.SettingKeyAuthSourceDefaultEmailGrantOnFirstBind: "false",
			service.SettingKeyForceEmailOnThirdPartySignup:           "true",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"registration_enabled":              true,
		"promo_code_enabled":                true,
		"auth_source_default_email_balance": 12.75,
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "12.75000000", repo.values[service.SettingKeyAuthSourceDefaultEmailBalance])
	require.Equal(t, "8", repo.values[service.SettingKeyAuthSourceDefaultEmailConcurrency])
	require.Equal(t, `[{"group_id":31,"validity_days":15}]`, repo.values[service.SettingKeyAuthSourceDefaultEmailSubscriptions])
	require.Equal(t, "true", repo.values[service.SettingKeyForceEmailOnThirdPartySignup])

	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, 12.75, data["auth_source_default_email_balance"])
	require.Equal(t, float64(8), data["auth_source_default_email_concurrency"])
	require.Equal(t, true, data["force_email_on_third_party_signup"])
}

func TestSettingHandler_UpdateSettings_PreservesOmittedSecuritySettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyRegistrationEnabled:         "true",
			service.SettingKeyEmailVerifyEnabled:          "true",
			service.SettingKeyPromoCodeEnabled:            "true",
			service.SettingKeyInvitationCodeEnabled:       "true",
			service.SettingKeyPasswordResetEnabled:        "true",
			service.SettingKeyPasswordMinLength:           "12",
			service.SettingKeyTotpEnabled:                 "true",
			service.SettingKeyTurnstileEnabled:            "true",
			service.SettingKeyTurnstileSiteKey:            "site-key",
			service.SettingKeyTurnstileSecretKey:          "secret-key",
			service.SettingKeyWeChatConnectEnabled:        "true",
			service.SettingKeyWeChatConnectOpenEnabled:    "true",
			service.SettingKeyWeChatConnectOpenAppID:      "wx-open-app",
			service.SettingKeyWeChatConnectOpenAppSecret:  "wx-open-secret",
			service.SettingKeyWeChatConnectMode:           "open",
			service.SettingKeyWeChatConnectScopes:         "snsapi_login",
			service.SettingKeyWeChatConnectRedirectURL:    "https://api.example.com/wechat/callback",
			service.SettingKeyGitHubOAuthEnabled:          "true",
			service.SettingKeyGitHubOAuthClientID:         "github-client",
			service.SettingKeyGitHubOAuthClientSecret:     "github-secret",
			service.SettingKeyDefaultConcurrency:          "7",
			service.SettingKeyDefaultBalance:              "18.25000000",
			service.SettingKeyDefaultUserRPMLimit:         "90",
			service.SettingKeyDefaultSubscriptions:        `[{"group_id":31,"validity_days":15}]`,
			service.SettingKeyEnableModelFallback:         "true",
			service.SettingKeyFallbackModelOpenAI:         "gpt-4.1",
			service.SettingKeyAllowUngroupedKeyScheduling: "true",
			service.SettingKeyBackendModeEnabled:          "true",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"payment_recharge_products": []map[string]any{
			{"id": "credits-100", "name": "100 credits", "amount": 10},
		},
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "true", repo.values[service.SettingKeyRegistrationEnabled])
	require.Equal(t, "true", repo.values[service.SettingKeyEmailVerifyEnabled])
	require.Equal(t, "true", repo.values[service.SettingKeyTurnstileEnabled])
	require.Equal(t, "site-key", repo.values[service.SettingKeyTurnstileSiteKey])
	require.Equal(t, "secret-key", repo.values[service.SettingKeyTurnstileSecretKey])
	require.Equal(t, "true", repo.values[service.SettingKeyWeChatConnectEnabled])
	require.Equal(t, "wx-open-app", repo.values[service.SettingKeyWeChatConnectOpenAppID])
	require.Equal(t, "wx-open-secret", repo.values[service.SettingKeyWeChatConnectOpenAppSecret])
	require.Equal(t, "true", repo.values[service.SettingKeyGitHubOAuthEnabled])
	require.Equal(t, "github-client", repo.values[service.SettingKeyGitHubOAuthClientID])
	require.Equal(t, "github-secret", repo.values[service.SettingKeyGitHubOAuthClientSecret])
	require.Equal(t, "7", repo.values[service.SettingKeyDefaultConcurrency])
	require.Equal(t, "18.25000000", repo.values[service.SettingKeyDefaultBalance])
	require.Equal(t, "90", repo.values[service.SettingKeyDefaultUserRPMLimit])
	require.Equal(t, `[{"group_id":31,"validity_days":15}]`, repo.values[service.SettingKeyDefaultSubscriptions])
	require.Equal(t, "true", repo.values[service.SettingKeyEnableModelFallback])
	require.Equal(t, "gpt-4.1", repo.values[service.SettingKeyFallbackModelOpenAI])
	require.Equal(t, "true", repo.values[service.SettingKeyAllowUngroupedKeyScheduling])
	require.Equal(t, "true", repo.values[service.SettingKeyBackendModeEnabled])
}

func TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyWebAppName:                "Web Old",
			service.SettingKeyPromptCasesTitle:          "Old Cases",
			service.SettingKeyPricingTitle:              "Old Pricing",
			service.SettingKeyWebEmailAuthVisible:       "false",
			service.SettingKeyWebGoogleAnalyticsID:      "G-OLD",
			service.SettingKeyWebCrispEnabled:           "false",
			service.SettingKeyWebCrispWebsiteID:         "crisp-old",
			service.SettingKeyWebVercelAnalyticsEnabled: "false",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"app_name":                 " Web New ",
		"prompt_cases_title":       " New Cases ",
		"pricing_title":            " New Pricing ",
		"pricing_currency_symbol":  " $ ",
		"credits_per_balance":      " 12 ",
		"email_auth_visible":       true,
		"google_analytics_id":      "  G-NEW  ",
		"crisp_enabled":            true,
		"crisp_website_id":         " crisp-new ",
		"vercel_analytics_enabled": true,
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "Web New", repo.values[service.SettingKeyWebAppName])
	require.Equal(t, "New Cases", repo.values[service.SettingKeyPromptCasesTitle])
	require.Equal(t, "New Pricing", repo.values[service.SettingKeyPricingTitle])
	require.Equal(t, "$", repo.values[service.SettingKeyPricingCurrencySymbol])
	require.Equal(t, "12", repo.values[service.SettingKeyCreditsPerBalance])
	require.Equal(t, "true", repo.values[service.SettingKeyWebEmailAuthVisible])
	require.Equal(t, "G-NEW", repo.values[service.SettingKeyWebGoogleAnalyticsID])
	require.Equal(t, "true", repo.values[service.SettingKeyWebCrispEnabled])
	require.Equal(t, "crisp-new", repo.values[service.SettingKeyWebCrispWebsiteID])
	require.Equal(t, "true", repo.values[service.SettingKeyWebVercelAnalyticsEnabled])

	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Web New", data["app_name"])
	require.Equal(t, "New Cases", data["prompt_cases_title"])
	require.Equal(t, "New Pricing", data["pricing_title"])
	require.Equal(t, "$", data["pricing_currency_symbol"])
	require.Equal(t, "12", data["credits_per_balance"])
	require.Equal(t, true, data["email_auth_visible"])
	require.Equal(t, "G-NEW", data["google_analytics_id"])
	require.Equal(t, true, data["crisp_enabled"])
	require.Equal(t, "crisp-new", data["crisp_website_id"])
	require.Equal(t, true, data["vercel_analytics_enabled"])
	require.NotContains(t, data, "touch_app_name")
	require.NotContains(t, data, "touch_email_auth_visible")
	require.NotContains(t, data, "touch_google_analytics_id")
	require.NotContains(t, data, "touch_crisp_enabled")
}

func TestSettingHandler_UpdateSettings_PersistsPaymentVisibleMethodsAndAdvancedScheduler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyPromoCodeEnabled: "true",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"promo_code_enabled":                    true,
		"payment_visible_method_alipay_source":  "easypay",
		"payment_visible_method_wxpay_source":   "wxpay",
		"payment_visible_method_alipay_enabled": true,
		"payment_visible_method_wxpay_enabled":  false,
		"openai_advanced_scheduler_enabled":     true,
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, service.VisibleMethodSourceEasyPayAlipay, repo.values[service.SettingPaymentVisibleMethodAlipaySource])
	require.Equal(t, service.VisibleMethodSourceOfficialWechat, repo.values[service.SettingPaymentVisibleMethodWxpaySource])
	require.Equal(t, "true", repo.values[service.SettingPaymentVisibleMethodAlipayEnabled])
	require.Equal(t, "false", repo.values[service.SettingPaymentVisibleMethodWxpayEnabled])
	require.Equal(t, "true", repo.values["openai_advanced_scheduler_enabled"])

	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, service.VisibleMethodSourceEasyPayAlipay, data["payment_visible_method_alipay_source"])
	require.Equal(t, service.VisibleMethodSourceOfficialWechat, data["payment_visible_method_wxpay_source"])
	require.Equal(t, true, data["payment_visible_method_alipay_enabled"])
	require.Equal(t, false, data["payment_visible_method_wxpay_enabled"])
	require.Equal(t, true, data["openai_advanced_scheduler_enabled"])
}

func TestSettingHandler_UpdateSettings_PreservesLegacyBlankPaymentVisibleMethodSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyPromoCodeEnabled:               "true",
			service.SettingPaymentVisibleMethodAlipayEnabled: "true",
			service.SettingPaymentVisibleMethodAlipaySource:  "",
			service.SettingPaymentVisibleMethodWxpayEnabled:  "false",
			service.SettingPaymentVisibleMethodWxpaySource:   "",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"promo_code_enabled": false,
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "", repo.values[service.SettingPaymentVisibleMethodAlipaySource])
	require.Equal(t, "true", repo.values[service.SettingPaymentVisibleMethodAlipayEnabled])
}

func TestSettingHandler_UpdateSettings_PersistsExplicitFalseOIDCCompatibilityFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyPromoCodeEnabled:               "true",
			service.SettingKeyOIDCConnectEnabled:             "true",
			service.SettingKeyOIDCConnectProviderName:        "OIDC",
			service.SettingKeyOIDCConnectClientID:            "oidc-client",
			service.SettingKeyOIDCConnectClientSecret:        "oidc-secret",
			service.SettingKeyOIDCConnectIssuerURL:           "https://issuer.example.com",
			service.SettingKeyOIDCConnectAuthorizeURL:        "https://issuer.example.com/auth",
			service.SettingKeyOIDCConnectTokenURL:            "https://issuer.example.com/token",
			service.SettingKeyOIDCConnectUserInfoURL:         "https://issuer.example.com/userinfo",
			service.SettingKeyOIDCConnectJWKSURL:             "https://issuer.example.com/jwks",
			service.SettingKeyOIDCConnectScopes:              "openid email profile",
			service.SettingKeyOIDCConnectRedirectURL:         "https://example.com/api/v1/auth/oauth/oidc/callback",
			service.SettingKeyOIDCConnectFrontendRedirectURL: "/auth/oidc/callback",
			service.SettingKeyOIDCConnectTokenAuthMethod:     "client_secret_post",
			service.SettingKeyOIDCConnectUsePKCE:             "true",
			service.SettingKeyOIDCConnectValidateIDToken:     "true",
			service.SettingKeyOIDCConnectAllowedSigningAlgs:  "RS256",
			service.SettingKeyOIDCConnectClockSkewSeconds:    "120",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"promo_code_enabled":                true,
		"oidc_connect_enabled":              true,
		"oidc_connect_use_pkce":             false,
		"oidc_connect_validate_id_token":    false,
		"oidc_connect_allowed_signing_algs": "",
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", repo.values[service.SettingKeyOIDCConnectUsePKCE])
	require.Equal(t, "false", repo.values[service.SettingKeyOIDCConnectValidateIDToken])

	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, false, data["oidc_connect_use_pkce"])
	require.Equal(t, false, data["oidc_connect_validate_id_token"])
}

func TestSettingHandler_UpdateSettings_DoesNotSolidifyImplicitOIDCSecurityDefaultsOnLegacyUpgrade(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyPromoCodeEnabled:                "true",
			service.SettingKeyOIDCConnectEnabled:              "true",
			service.SettingKeyOIDCConnectProviderName:         "OIDC",
			service.SettingKeyOIDCConnectClientID:             "oidc-client",
			service.SettingKeyOIDCConnectClientSecret:         "oidc-secret",
			service.SettingKeyOIDCConnectIssuerURL:            "https://issuer.example.com",
			service.SettingKeyOIDCConnectAuthorizeURL:         "https://issuer.example.com/auth",
			service.SettingKeyOIDCConnectTokenURL:             "https://issuer.example.com/token",
			service.SettingKeyOIDCConnectUserInfoURL:          "https://issuer.example.com/userinfo",
			service.SettingKeyOIDCConnectJWKSURL:              "https://issuer.example.com/jwks",
			service.SettingKeyOIDCConnectScopes:               "openid email profile",
			service.SettingKeyOIDCConnectRedirectURL:          "https://example.com/api/v1/auth/oauth/oidc/callback",
			service.SettingKeyOIDCConnectFrontendRedirectURL:  "/auth/oidc/callback",
			service.SettingKeyOIDCConnectTokenAuthMethod:      "client_secret_post",
			service.SettingKeyOIDCConnectAllowedSigningAlgs:   "RS256",
			service.SettingKeyOIDCConnectClockSkewSeconds:     "120",
			service.SettingKeyOIDCConnectRequireEmailVerified: "false",
			service.SettingKeyOIDCConnectUserInfoEmailPath:    "",
			service.SettingKeyOIDCConnectUserInfoIDPath:       "",
			service.SettingKeyOIDCConnectUserInfoUsernamePath: "",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{
		Default: config.DefaultConfig{UserConcurrency: 5},
		OIDC: config.OIDCConnectConfig{
			Enabled:             true,
			ProviderName:        "OIDC",
			ClientID:            "oidc-client",
			ClientSecret:        "oidc-secret",
			IssuerURL:           "https://issuer.example.com",
			AuthorizeURL:        "https://issuer.example.com/auth",
			TokenURL:            "https://issuer.example.com/token",
			UserInfoURL:         "https://issuer.example.com/userinfo",
			JWKSURL:             "https://issuer.example.com/jwks",
			Scopes:              "openid email profile",
			RedirectURL:         "https://example.com/api/v1/auth/oauth/oidc/callback",
			FrontendRedirectURL: "/auth/oidc/callback",
			TokenAuthMethod:     "client_secret_post",
			UsePKCE:             true,
			ValidateIDToken:     true,
			AllowedSigningAlgs:  "RS256",
			ClockSkewSeconds:    120,
		},
	})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"promo_code_enabled":   true,
		"oidc_connect_enabled": true,
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", repo.values[service.SettingKeyOIDCConnectUsePKCE])
	require.Equal(t, "false", repo.values[service.SettingKeyOIDCConnectValidateIDToken])
}

func TestSettingHandler_UpdateSettings_RejectsInvalidPaymentVisibleMethodSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyPromoCodeEnabled: "true",
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"promo_code_enabled":                   true,
		"payment_visible_method_alipay_source": "bogus",
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.NotContains(t, repo.values, service.SettingPaymentVisibleMethodAlipaySource)
}

func TestSettingHandler_UpdateSettings_DoesNotPersistPartialSystemSettingsWhenAuthSourceDefaultsFail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &failingAuthSourceSettingsRepoStub{
		values: map[string]string{
			service.SettingKeyRegistrationEnabled:                 "false",
			service.SettingKeyPromoCodeEnabled:                    "true",
			service.SettingKeyAuthSourceDefaultEmailBalance:       "9.5",
			service.SettingKeyAuthSourceDefaultEmailConcurrency:   "8",
			service.SettingKeyAuthSourceDefaultEmailSubscriptions: `[{"group_id":31,"validity_days":15}]`,
		},
		err: errors.New("write auth source defaults failed"),
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := map[string]any{
		"registration_enabled":              true,
		"promo_code_enabled":                true,
		"auth_source_default_email_balance": 12.75,
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "false", repo.values[service.SettingKeyRegistrationEnabled])
	require.Equal(t, "9.5", repo.values[service.SettingKeyAuthSourceDefaultEmailBalance])
}

func TestDiffSettings_IncludesAuthSourceDefaultsAndForceEmail(t *testing.T) {
	changed := diffSettings(
		&service.SystemSettings{},
		&service.SystemSettings{},
		&service.AuthSourceDefaultSettings{
			Email: service.ProviderDefaultGrantSettings{
				Balance:          0,
				Concurrency:      5,
				Subscriptions:    nil,
				GrantOnSignup:    true,
				GrantOnFirstBind: false,
			},
			ForceEmailOnThirdPartySignup: false,
		},
		&service.AuthSourceDefaultSettings{
			Email: service.ProviderDefaultGrantSettings{
				Balance:          12.5,
				Concurrency:      7,
				Subscriptions:    []service.DefaultSubscriptionSetting{{GroupID: 21, ValidityDays: 30}},
				GrantOnSignup:    false,
				GrantOnFirstBind: true,
			},
			ForceEmailOnThirdPartySignup: true,
		},
		UpdateSettingsRequest{},
	)

	require.Contains(t, changed, "auth_source_default_email_balance")
	require.Contains(t, changed, "auth_source_default_email_concurrency")
	require.Contains(t, changed, "auth_source_default_email_subscriptions")
	require.Contains(t, changed, "auth_source_default_email_grant_on_signup")
	require.Contains(t, changed, "auth_source_default_email_grant_on_first_bind")
	require.Contains(t, changed, "force_email_on_third_party_signup")
}
