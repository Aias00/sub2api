package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/cloudbase/internal/payment"
	infraerrors "github.com/Wei-Shaw/cloudbase/internal/pkg/errors"
	"github.com/Wei-Shaw/cloudbase/internal/pkg/oauth"
	"github.com/Wei-Shaw/cloudbase/internal/pkg/response"
	"github.com/Wei-Shaw/cloudbase/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	wechatPaymentOAuthCookiePath  = "/api/v1/auth/oauth/wechat/payment"
	wechatPaymentOAuthStateName   = "wechat_payment_oauth_state"
	wechatPaymentOAuthRedirect    = "wechat_payment_oauth_redirect"
	wechatPaymentOAuthContextName = "wechat_payment_oauth_context"
	wechatPaymentOAuthScope       = "wechat_payment_oauth_scope"
	wechatPaymentOAuthDefaultTo   = "/purchase"
	wechatPaymentOAuthFrontendCB  = "/auth/wechat/payment/callback"
	wechatOAuthCookieMaxAgeSec    = 10 * 60
)

var (
	wechatOAuthAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
)

type wechatOAuthConfig struct {
	mode         string
	appID        string
	appSecret    string
	authorizeURL string
	scope        string
	redirectURI  string
	openEnabled  bool
	mpEnabled    bool
}

type wechatOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int64  `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

type wechatPaymentOAuthContext struct {
	PaymentType string `json:"payment_type"`
	Amount      string `json:"amount,omitempty"`
	OrderType   string `json:"order_type,omitempty"`
	PlanID      int64  `json:"plan_id,omitempty"`
}

// WeChatPaymentOAuthStart starts the WeChat payment OAuth flow.
// GET /api/v1/auth/oauth/wechat/payment/start?payment_type=wxpay&redirect=/purchase
func (h *AuthHandler) WeChatPaymentOAuthStart(c *gin.Context) {
	cfg, err := h.getWeChatOAuthConfig(c.Request.Context(), "mp", c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	paymentType := normalizeWeChatPaymentType(c.Query("payment_type"))
	if paymentType == "" {
		response.BadRequest(c, "Invalid payment type")
		return
	}

	state, err := oauth.GenerateState()
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_STATE_GEN_FAILED", "failed to generate oauth state").WithCause(err))
		return
	}

	redirectTo := normalizeWeChatPaymentRedirectPath(sanitizeFrontendRedirectPath(c.Query("redirect")))
	if redirectTo == "" {
		redirectTo = wechatPaymentOAuthDefaultTo
	}
	rawContext, err := encodeWeChatPaymentOAuthContext(wechatPaymentOAuthContext{
		PaymentType: paymentType,
		Amount:      strings.TrimSpace(c.Query("amount")),
		OrderType:   strings.TrimSpace(c.Query("order_type")),
		PlanID:      parseWeChatPaymentPlanID(c.Query("plan_id")),
	})
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_CONTEXT_ENCODE_FAILED", "failed to encode oauth context").WithCause(err))
		return
	}

	scope := normalizeWeChatPaymentScope(c.Query("scope"))
	secureCookie := isRequestHTTPS(c)
	wechatPaymentSetCookie(c, wechatPaymentOAuthStateName, encodeCookieValue(state), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatPaymentSetCookie(c, wechatPaymentOAuthRedirect, encodeCookieValue(redirectTo), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatPaymentSetCookie(c, wechatPaymentOAuthContextName, encodeCookieValue(rawContext), wechatOAuthCookieMaxAgeSec, secureCookie)
	wechatPaymentSetCookie(c, wechatPaymentOAuthScope, encodeCookieValue(scope), wechatOAuthCookieMaxAgeSec, secureCookie)

	cfg.redirectURI = h.resolveWeChatPaymentOAuthCallbackURL(c.Request.Context(), c)
	cfg.scope = scope
	authURL, err := buildWeChatAuthorizeURL(cfg, state)
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_BUILD_URL_FAILED", "failed to build oauth authorization url").WithCause(err))
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// WeChatPaymentOAuthCallback exchanges a payment OAuth code for an OpenID and
// forwards the browser back to the frontend callback route.
func (h *AuthHandler) WeChatPaymentOAuthCallback(c *gin.Context) {
	frontendCallback := wechatPaymentOAuthFrontendCB

	if providerErr := strings.TrimSpace(c.Query("error")); providerErr != "" {
		redirectOAuthError(c, frontendCallback, "provider_error", providerErr, c.Query("error_description"))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		redirectOAuthError(c, frontendCallback, "missing_params", "missing code/state", "")
		return
	}

	secureCookie := isRequestHTTPS(c)
	defer func() {
		wechatPaymentClearCookie(c, wechatPaymentOAuthStateName, secureCookie)
		wechatPaymentClearCookie(c, wechatPaymentOAuthRedirect, secureCookie)
		wechatPaymentClearCookie(c, wechatPaymentOAuthContextName, secureCookie)
		wechatPaymentClearCookie(c, wechatPaymentOAuthScope, secureCookie)
	}()

	expectedState, err := readCookieDecoded(c, wechatPaymentOAuthStateName)
	if err != nil || expectedState == "" || state != expectedState {
		redirectOAuthError(c, frontendCallback, "invalid_state", "invalid oauth state", "")
		return
	}

	redirectTo, _ := readCookieDecoded(c, wechatPaymentOAuthRedirect)
	redirectTo = normalizeWeChatPaymentRedirectPath(sanitizeFrontendRedirectPath(redirectTo))
	if redirectTo == "" {
		redirectTo = wechatPaymentOAuthDefaultTo
	}

	rawContext, _ := readCookieDecoded(c, wechatPaymentOAuthContextName)
	paymentContext, err := decodeWeChatPaymentOAuthContext(rawContext)
	if err != nil {
		redirectOAuthError(c, frontendCallback, "invalid_context", "invalid oauth context", "")
		return
	}
	if paymentContext.PaymentType == "" {
		paymentContext.PaymentType = payment.TypeWxpay
	}

	scope, _ := readCookieDecoded(c, wechatPaymentOAuthScope)
	scope = normalizeWeChatPaymentScope(scope)

	cfg, err := h.getWeChatOAuthConfig(c.Request.Context(), "mp", c)
	if err != nil {
		redirectOAuthError(c, frontendCallback, "provider_error", infraerrors.Reason(err), infraerrors.Message(err))
		return
	}
	cfg.redirectURI = h.resolveWeChatPaymentOAuthCallbackURL(c.Request.Context(), c)
	tokenResp, err := exchangeWeChatOAuthCode(c.Request.Context(), cfg, code)
	if err != nil {
		redirectOAuthError(c, frontendCallback, "token_exchange_failed", "failed to exchange oauth code", err.Error())
		return
	}

	openid := strings.TrimSpace(tokenResp.OpenID)
	if openid == "" {
		redirectOAuthError(c, frontendCallback, "missing_openid", "missing openid", "")
		return
	}
	if strings.TrimSpace(tokenResp.Scope) != "" {
		scope = strings.TrimSpace(tokenResp.Scope)
	}

	resumeToken, err := h.wechatPaymentResumeService().CreateWeChatPaymentResumeToken(service.WeChatPaymentResumeClaims{
		OpenID:      openid,
		PaymentType: paymentContext.PaymentType,
		Amount:      paymentContext.Amount,
		OrderType:   paymentContext.OrderType,
		PlanID:      paymentContext.PlanID,
		RedirectTo:  redirectTo,
		Scope:       scope,
	})
	if err != nil {
		redirectOAuthError(c, frontendCallback, "invalid_context", "failed to encode payment resume context", "")
		return
	}

	fragment := url.Values{}
	fragment.Set("wechat_resume_token", resumeToken)
	fragment.Set("redirect", redirectTo)
	redirectWithFragment(c, frontendCallback, fragment)
}

func (h *AuthHandler) wechatPaymentResumeService() *service.PaymentResumeService {
	return service.NewPaymentResumeServiceFromEnvWithLegacyKeys(
		service.ParsePaymentResumeSigningKey(h.cfg.Totp.EncryptionKey),
	)
}

func (h *AuthHandler) getWeChatOAuthConfig(ctx context.Context, rawMode string, c *gin.Context) (wechatOAuthConfig, error) {
	mode, err := resolveWeChatOAuthMode(rawMode, c)
	if err != nil {
		return wechatOAuthConfig{}, err
	}

	if h == nil || h.settingSvc == nil {
		return wechatOAuthConfig{}, infraerrors.ServiceUnavailable("CONFIG_NOT_READY", "wechat oauth settings service not ready")
	}

	apiBaseURL := ""
	if h != nil && h.settingSvc != nil {
		settings, err := h.settingSvc.GetAllSettings(ctx)
		if err == nil && settings != nil {
			apiBaseURL = strings.TrimSpace(settings.APIBaseURL)
		}
	}

	effective, err := h.settingSvc.GetWeChatConnectOAuthConfig(ctx)
	if err != nil {
		return wechatOAuthConfig{}, err
	}
	if !effective.SupportsMode(mode) {
		return wechatOAuthConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "wechat oauth is disabled")
	}

	cfg := wechatOAuthConfig{
		mode:        mode,
		appID:       strings.TrimSpace(effective.AppIDForMode(mode)),
		appSecret:   strings.TrimSpace(effective.AppSecretForMode(mode)),
		redirectURI: firstNonEmpty(strings.TrimSpace(effective.RedirectURL), resolveWeChatOAuthAbsoluteURL(apiBaseURL, c, "/api/v1/auth/oauth/wechat/payment/callback")),
		scope:       effective.ScopeForMode(mode),
		openEnabled: effective.OpenEnabled,
		mpEnabled:   effective.MPEnabled,
	}

	switch mode {
	case "mp":
		cfg.authorizeURL = "https://open.weixin.qq.com/connect/oauth2/authorize"
	default:
		cfg.authorizeURL = "https://open.weixin.qq.com/connect/qrconnect"
	}
	if strings.TrimSpace(cfg.redirectURI) == "" {
		return wechatOAuthConfig{}, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "wechat oauth redirect url not configured")
	}

	return cfg, nil
}

func resolveWeChatOAuthMode(rawMode string, c *gin.Context) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(rawMode))
	if mode == "" {
		if isWeChatBrowserRequest(c) {
			return "mp", nil
		}
		return "open", nil
	}
	if mode != "open" && mode != "mp" {
		return "", infraerrors.BadRequest("INVALID_MODE", "wechat oauth mode must be open or mp")
	}
	return mode, nil
}

func isWeChatBrowserRequest(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(c.GetHeader("User-Agent"))), "micromessenger")
}

func buildWeChatAuthorizeURL(cfg wechatOAuthConfig, state string) (string, error) {
	u, err := url.Parse(cfg.authorizeURL)
	if err != nil {
		return "", fmt.Errorf("parse authorize url: %w", err)
	}
	query := u.Query()
	query.Set("appid", cfg.appID)
	query.Set("redirect_uri", cfg.redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", cfg.scope)
	query.Set("state", state)
	u.RawQuery = query.Encode()
	u.Fragment = "wechat_redirect"
	return u.String(), nil
}

func resolveWeChatOAuthAbsoluteURL(apiBaseURL string, c *gin.Context, callbackPath string) string {
	callbackPath = strings.TrimSpace(callbackPath)
	if callbackPath == "" {
		return ""
	}

	if raw := strings.TrimSpace(apiBaseURL); raw != "" {
		if parsed, err := url.Parse(raw); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			basePath := strings.TrimRight(parsed.EscapedPath(), "/")
			targetPath := callbackPath
			if basePath != "" && strings.HasSuffix(basePath, "/api/v1") && strings.HasPrefix(callbackPath, "/api/v1") {
				targetPath = basePath + strings.TrimPrefix(callbackPath, "/api/v1")
			} else if basePath != "" {
				targetPath = basePath + callbackPath
			}
			return parsed.Scheme + "://" + parsed.Host + targetPath
		}
	}

	if c == nil || c.Request == nil {
		return ""
	}
	scheme := "http"
	if isRequestHTTPS(c) {
		scheme = "https"
	}
	host := strings.TrimSpace(c.Request.Host)
	if forwardedHost := strings.TrimSpace(c.GetHeader("X-Forwarded-Host")); forwardedHost != "" {
		host = forwardedHost
	}
	if host == "" {
		return ""
	}
	return scheme + "://" + host + callbackPath
}

func exchangeWeChatOAuthCode(ctx context.Context, cfg wechatOAuthConfig, code string) (*wechatOAuthTokenResponse, error) {
	endpoint, err := url.Parse(wechatOAuthAccessTokenURL)
	if err != nil {
		return nil, fmt.Errorf("parse wechat access token url: %w", err)
	}

	query := endpoint.Query()
	query.Set("appid", cfg.appID)
	query.Set("secret", cfg.appSecret)
	query.Set("code", strings.TrimSpace(code))
	query.Set("grant_type", "authorization_code")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build wechat access token request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request wechat access token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read wechat access token response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("wechat access token status=%d", resp.StatusCode)
	}

	var tokenResp wechatOAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("decode wechat access token response: %w", err)
	}
	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("wechat access token error=%d %s", tokenResp.ErrCode, strings.TrimSpace(tokenResp.ErrMsg))
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return nil, fmt.Errorf("wechat access token missing access_token")
	}
	return &tokenResp, nil
}

func normalizeWeChatPaymentType(raw string) string {
	switch strings.TrimSpace(raw) {
	case payment.TypeWxpay, payment.TypeWxpayDirect:
		return strings.TrimSpace(raw)
	default:
		return ""
	}
}

func normalizeWeChatPaymentScope(raw string) string {
	for _, part := range strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	}) {
		switch strings.TrimSpace(part) {
		case "snsapi_userinfo":
			return "snsapi_userinfo"
		case "snsapi_base":
			return "snsapi_base"
		}
	}
	return "snsapi_base"
}

func normalizeWeChatPaymentRedirectPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return wechatPaymentOAuthDefaultTo
	}
	if path == "/payment" {
		return "/purchase"
	}
	if strings.HasPrefix(path, "/payment?") {
		return "/purchase" + strings.TrimPrefix(path, "/payment")
	}
	return path
}

func (h *AuthHandler) resolveWeChatPaymentOAuthCallbackURL(ctx context.Context, c *gin.Context) string {
	apiBaseURL := ""
	if h != nil && h.settingSvc != nil {
		if settings, err := h.settingSvc.GetAllSettings(ctx); err == nil && settings != nil {
			apiBaseURL = strings.TrimSpace(settings.APIBaseURL)
		}
	}
	return resolveWeChatOAuthAbsoluteURL(apiBaseURL, c, "/api/v1/auth/oauth/wechat/payment/callback")
}

func encodeWeChatPaymentOAuthContext(ctx wechatPaymentOAuthContext) (string, error) {
	data, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeWeChatPaymentOAuthContext(raw string) (wechatPaymentOAuthContext, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return wechatPaymentOAuthContext{}, nil
	}
	var ctx wechatPaymentOAuthContext
	if err := json.Unmarshal([]byte(raw), &ctx); err != nil {
		return wechatPaymentOAuthContext{}, err
	}
	return ctx, nil
}

func parseWeChatPaymentPlanID(raw string) int64 {
	id, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return id
}

func wechatPaymentSetCookie(c *gin.Context, name string, value string, maxAgeSec int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     wechatPaymentOAuthCookiePath,
		MaxAge:   maxAgeSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func wechatPaymentClearCookie(c *gin.Context, name string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     wechatPaymentOAuthCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
