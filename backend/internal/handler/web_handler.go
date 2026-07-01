package handler

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	webAccessTokenCookie  = "sub2api_web_access_token"
	webRefreshTokenCookie = "sub2api_web_refresh_token"
	webPaymentSource      = "sub2api_web"
)

type WebHandler struct {
	authHandler     *AuthHandler
	authService     *service.AuthService
	userService     *service.UserService
	paymentService  webPaymentOrderCreator
	configService   webPaymentConfigReader
	settingService  *service.SettingService
	twitterImporter *service.TwitterImportService
}

type webPaymentOrderCreator interface {
	CreateOrder(ctx context.Context, req service.CreateOrderRequest) (*service.CreateOrderResponse, error)
}

type webPaymentConfigReader interface {
	GetPaymentConfig(ctx context.Context) (*service.PaymentConfig, error)
}

func NewWebHandler(authHandler *AuthHandler, userService *service.UserService, paymentService *service.PaymentService, configService *service.PaymentConfigService, settingService *service.SettingService, twitterImporter *service.TwitterImportService) *WebHandler {
	return &WebHandler{
		authHandler:     authHandler,
		authService:     authHandler.authService,
		userService:     userService,
		paymentService:  paymentService,
		configService:   configService,
		settingService:  settingService,
		twitterImporter: twitterImporter,
	}
}

type webLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type webRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

type webCheckoutRequest struct {
	ProductID     string            `json:"product_id"`
	PaymentType   string            `json:"payment_type"`
	PaymentSource string            `json:"payment_source"`
	ReturnURL     string            `json:"return_url"`
	OpenID        string            `json:"openid"`
	IsMobile      *bool             `json:"is_mobile"`
	Metadata      map[string]string `json:"metadata"`
}

type webImportRequest struct {
	Provider  string   `json:"provider"`
	URL       string   `json:"url" binding:"required"`
	Prompt    string   `json:"prompt"`
	Title     string   `json:"title"`
	Category  string   `json:"category"`
	ImageURLs []string `json:"image_urls"`
}

type webOAuthSessionRequest struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

func (h *WebHandler) OAuthStart(c *gin.Context) {
	if h.authHandler == nil {
		response.InternalError(c, "Web OAuth is not configured")
		return
	}
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	if provider != "google" && provider != "github" {
		response.BadRequest(c, "unsupported OAuth provider")
		return
	}

	query := c.Request.URL.Query()
	query.Set("source", webAuthSource)
	c.Request.URL.RawQuery = query.Encode()
	c.Set(webAuthSourceTrustedContextName, true)

	switch provider {
	case "google":
		h.authHandler.GoogleOAuthStart(c)
	case "github":
		h.authHandler.GitHubOAuthStart(c)
	default:
		response.BadRequest(c, "unsupported OAuth provider")
	}
}

func (h *WebHandler) Login(c *gin.Context) {
	var req webLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	user, err := h.loginWebUser(c, req.Email, req.Password)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.respondWithWebSession(c, user)
}

func (h *WebHandler) Register(c *gin.Context) {
	var req webRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	username := firstNonEmptyString(req.Username, req.Name)
	riskCtx := h.authHandler.signupGrantRiskContext(c, "", "")
	_, user, err := h.authService.RegisterWithVerificationSourceAndUsername(riskCtx, req.Email, username, req.Password, "", "", "", "", webAuthSource)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if err := h.authHandler.ensureUserAcceptedCurrentLoginAgreement(riskCtx, user, agreementAcceptanceInput{
		Accepted: true,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.respondWithWebSession(c, user)
}

func (h *WebHandler) Refresh(c *gin.Context) {
	refreshToken, err := readWebSessionCookie(c, webRefreshTokenCookie)
	if err != nil || strings.TrimSpace(refreshToken) == "" {
		response.Unauthorized(c, "Sub2API refresh token is missing")
		return
	}

	result, err := h.authService.RefreshTokenPair(c.Request.Context(), refreshToken)
	if err != nil {
		h.clearWebSessionCookies(c)
		response.ErrorFrom(c, err)
		return
	}
	if h.authHandler.settingSvc.IsBackendModeEnabled(c.Request.Context()) && result.UserRole != service.RoleAdmin {
		h.clearWebSessionCookies(c)
		response.Forbidden(c, "Backend mode is active. Only admin login is allowed.")
		return
	}

	h.setWebSessionCookies(c, result.AccessToken, result.RefreshToken, result.ExpiresIn)
	response.Success(c, gin.H{"ok": true})
}

func (h *WebHandler) OAuthSession(c *gin.Context) {
	var req webOAuthSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	accessToken := strings.TrimSpace(req.AccessToken)
	if accessToken == "" {
		response.BadRequest(c, "accessToken is required")
		return
	}
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		response.BadRequest(c, "refreshToken is required")
		return
	}
	claims, err := h.authService.ValidateToken(accessToken)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if user.SignupSource != webAuthSource {
		response.Forbidden(c, "Sub2API token is not valid for this web session")
		return
	}
	if !user.IsActive() {
		response.ErrorFrom(c, service.ErrUserNotActive)
		return
	}
	if err := h.authService.ValidateRefreshTokenForUser(c.Request.Context(), refreshToken, user); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.setWebSessionCookies(c, accessToken, refreshToken, req.ExpiresIn)
	response.Success(c, gin.H{
		"user": h.webUserPayload(c.Request.Context(), user),
	})
}

func (h *WebHandler) Logout(c *gin.Context) {
	if refreshToken, err := readWebSessionCookie(c, webRefreshTokenCookie); err == nil && strings.TrimSpace(refreshToken) != "" {
		_ = h.authService.RevokeRefreshToken(c.Request.Context(), refreshToken)
	}
	h.clearWebSessionCookies(c)
	response.Success(c, gin.H{"ok": true})
}

func (h *WebHandler) Me(c *gin.Context) {
	user, err := h.currentWebUser(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, h.webUserPayload(c.Request.Context(), user))
}

func (h *WebHandler) Credits(c *gin.Context) {
	user, err := h.currentWebUser(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, h.webCreditsPayload(c.Request.Context(), user))
}

func (h *WebHandler) Checkout(c *gin.Context) {
	user, err := h.currentWebUser(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var req webCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	productID := strings.TrimSpace(req.ProductID)
	if productID == "" {
		response.BadRequest(c, "product_id is required")
		return
	}
	paymentType := strings.TrimSpace(req.PaymentType)
	if paymentType == "" {
		response.BadRequest(c, "payment_type required")
		return
	}
	amount, ok, err := h.resolveRechargeProductAmount(c.Request.Context(), productID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if !ok || amount <= 0 {
		response.BadRequest(c, "recharge product not found")
		return
	}

	isMobile := false
	if req.IsMobile != nil {
		isMobile = *req.IsMobile
	}
	result, err := h.paymentService.CreateOrder(c.Request.Context(), service.CreateOrderRequest{
		UserID:          user.ID,
		Amount:          amount,
		ProductID:       productID,
		PaymentType:     paymentType,
		OpenID:          req.OpenID,
		ClientIP:        c.ClientIP(),
		IsMobile:        isMobile,
		IsWeChatBrowser: strings.Contains(strings.ToLower(c.GetHeader("User-Agent")), "micromessenger"),
		SrcHost:         c.Request.Host,
		SrcURL:          c.Request.Referer(),
		ReturnURL:       strings.TrimSpace(req.ReturnURL),
		PaymentSource:   webCheckoutPaymentSource(req.PaymentSource),
		OrderType:       payment.OrderTypeBalance,
		Locale:          c.GetHeader("Accept-Language"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

func (h *WebHandler) ImportTwitter(c *gin.Context) {
	if h.twitterImporter == nil {
		response.InternalError(c, "Twitter importer is not configured")
		return
	}
	user, err := h.currentWebUser(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if !user.IsAdmin() {
		response.Forbidden(c, "no permission")
		return
	}

	var req webImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider != "" && provider != "x" && provider != "twitter" {
		response.BadRequest(c, "unsupported provider")
		return
	}
	result, err := h.twitterImporter.Import(c.Request.Context(), service.TwitterImportInput{
		URL:       req.URL,
		Prompt:    req.Prompt,
		Title:     req.Title,
		Category:  req.Category,
		ImageURLs: req.ImageURLs,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, twitterImportResponse{
		Item:         promptCatalogCaseDTOFromService(result.Item),
		ImageURLs:    nonNilStrings(result.ImageURLs),
		UploadedURLs: nonNilStrings(result.UploadedURLs),
		Warnings:     nonNilStrings(result.Warnings),
	})
}

func (h *WebHandler) loginWebUser(c *gin.Context, email, password string) (*service.User, error) {
	_, user, err := h.authService.LoginWithSource(c.Request.Context(), email, password, webAuthSource)
	if err != nil {
		return nil, err
	}
	if err := h.authHandler.ensureBackendModeAllowsUser(c.Request.Context(), user); err != nil {
		return nil, err
	}
	if err := h.authHandler.ensureUserAcceptedCurrentLoginAgreement(c.Request.Context(), user, agreementAcceptanceInput{
		Accepted: true,
	}); err != nil {
		return nil, err
	}
	h.authService.RecordSuccessfulLogin(c.Request.Context(), user.ID)
	return user, nil
}

func (h *WebHandler) respondWithWebSession(c *gin.Context, user *service.User) {
	if err := ensureLoginUserActive(user); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user, "")
	if err != nil {
		response.InternalError(c, "Failed to generate token")
		return
	}
	h.setWebSessionCookies(c, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)
	response.Success(c, gin.H{
		"user": h.webUserPayload(c.Request.Context(), user),
	})
}

func (h *WebHandler) currentWebUser(c *gin.Context) (*service.User, error) {
	accessToken, err := readWebSessionCookie(c, webAccessTokenCookie)
	if err != nil || strings.TrimSpace(accessToken) == "" {
		return nil, service.ErrInvalidToken
	}
	claims, err := h.authService.ValidateToken(accessToken)
	if err != nil {
		return nil, err
	}
	user, err := h.userService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.SignupSource != webAuthSource {
		return nil, service.ErrInvalidCredentials
	}
	if !user.IsActive() {
		return nil, service.ErrUserNotActive
	}
	return user, nil
}

func (h *WebHandler) webUserPayload(ctx context.Context, user *service.User) gin.H {
	isAdmin := false
	if user != nil {
		isAdmin = user.IsAdmin()
	}
	payload := gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"username":      user.Username,
		"name":          firstNonEmptyString(user.Username, user.Email),
		"role":          user.Role,
		"status":        user.Status,
		"signup_source": user.SignupSource,
		"balance":       user.Balance,
		"isAdmin":       isAdmin,
		"is_admin":      isAdmin,
		"credits":       h.webCreditsPayload(ctx, user),
	}
	if identities, err := h.userService.GetProfileIdentitySummaries(ctx, user.ID, user); err == nil {
		payload["identities"] = identities
	}
	return payload
}

func (h *WebHandler) webCreditsPayload(ctx context.Context, user *service.User) gin.H {
	balance := 0.0
	var userID int64
	if user != nil {
		balance = user.Balance
		userID = user.ID
	}
	creditsPerBalance := h.creditsPerBalance(ctx)
	return gin.H{
		"remainingCredits": int(balance * creditsPerBalance),
		"sub2apiBalance":   balance,
		"sub2apiUserId":    userID,
	}
}

func (h *WebHandler) creditsPerBalance(ctx context.Context) float64 {
	if h == nil || h.settingService == nil {
		return 10
	}
	settings, err := h.settingService.GetPublicSettings(ctx)
	if err != nil {
		return 10
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(settings.CreditsPerBalance), 64)
	if err != nil || value <= 0 {
		return 10
	}
	return value
}

func (h *WebHandler) resolveRechargeProductAmount(ctx context.Context, productID string) (float64, bool, error) {
	if h.configService == nil {
		return 0, false, nil
	}
	cfg, err := h.configService.GetPaymentConfig(ctx)
	if err != nil {
		return 0, false, err
	}
	if cfg == nil || cfg.BalanceDisabled {
		return 0, false, nil
	}
	for _, product := range cfg.RechargeProducts {
		if strings.TrimSpace(product.ID) == productID {
			return product.Amount, true, nil
		}
	}
	return 0, false, nil
}

func (h *WebHandler) setWebSessionCookies(c *gin.Context, accessToken, refreshToken string, expiresIn int) {
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	h.setWebCookie(c, webAccessTokenCookie, accessToken, expiresIn)
	h.setWebCookie(c, webRefreshTokenCookie, refreshToken, 60*60*24*30)
}

func (h *WebHandler) clearWebSessionCookies(c *gin.Context) {
	h.setWebCookie(c, webAccessTokenCookie, "", -1)
	h.setWebCookie(c, webRefreshTokenCookie, "", -1)
}

func readWebSessionCookie(c *gin.Context, name string) (string, error) {
	return c.Cookie(name)
}

func webCheckoutPaymentSource(source string) string {
	return firstNonEmptyString(source, webPaymentSource)
}

func (h *WebHandler) setWebCookie(c *gin.Context, name, value string, maxAge int) {
	sameSite := http.SameSiteLaxMode
	if shouldUseSameSiteNoneForWebCookie(c) {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isRequestHTTPS(c) || sameSite == http.SameSiteNoneMode,
		SameSite: sameSite,
	})
}

func shouldUseSameSiteNoneForWebCookie(c *gin.Context) bool {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin == "" {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(c.Writer.Header().Get("Access-Control-Allow-Credentials")), "true") {
		return false
	}
	if strings.TrimSpace(c.Writer.Header().Get("Access-Control-Allow-Origin")) != origin {
		return false
	}
	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		return false
	}
	return !strings.EqualFold(originURL.Host, c.Request.Host)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

var _ webPaymentOrderCreator = (*service.PaymentService)(nil)
var _ webPaymentConfigReader = (*service.PaymentConfigService)(nil)
