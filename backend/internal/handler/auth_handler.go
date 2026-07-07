package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/Aias00/cloudbase/internal/config"
	"github.com/Aias00/cloudbase/internal/handler/dto"
	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/ip"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	middleware2 "github.com/Aias00/cloudbase/internal/server/middleware"
	"github.com/Aias00/cloudbase/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	cfg                  *config.Config
	authService          *service.AuthService
	userService          *service.UserService
	settingSvc           *service.SettingService
	promoService         *service.PromoService
	redeemService        *service.RedeemService
	totpService          *service.TotpService
	userAttributeService *service.UserAttributeService
}

const (
	webAuthSource                   = "web"
	webAccountSignupSource          = "email"
	webBridgeTokenQueryName         = "cloudbase_web_bridge_token"
	webAuthSourceTrustedContextName = "web_auth_source_trusted"
)

type webOAuthBridgeTokenPayload struct {
	Source   string `json:"source"`
	Provider string `json:"provider"`
	Expires  int64  `json:"exp"`
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(cfg *config.Config, authService *service.AuthService, userService *service.UserService, settingService *service.SettingService, promoService *service.PromoService, redeemService *service.RedeemService, totpService *service.TotpService, userAttributeService *service.UserAttributeService) *AuthHandler {
	return &AuthHandler{
		cfg:                  cfg,
		authService:          authService,
		userService:          userService,
		settingSvc:           settingService,
		promoService:         promoService,
		redeemService:        redeemService,
		totpService:          totpService,
		userAttributeService: userAttributeService,
	}
}

func isWebAuthSource(source string) bool {
	return strings.EqualFold(strings.TrimSpace(source), webAuthSource)
}

func (h *AuthHandler) signupGrantRiskContext(c *gin.Context, providerType, providerSubject string) context.Context {
	ctx := context.Background()
	if c != nil && c.Request != nil {
		ctx = c.Request.Context()
	}
	input := service.SignupGrantRiskInput{
		ProviderType:    providerType,
		ProviderSubject: providerSubject,
	}
	if c != nil {
		input.RemoteIP = ip.GetClientIP(c)
		if c.Request != nil {
			input.UserAgent = c.Request.UserAgent()
			input.AcceptLanguage = c.GetHeader("Accept-Language")
			input.DeviceFingerprint = c.GetHeader("X-Device-Fingerprint")
			input.HeaderSnapshot = registrationHeaderSnapshot(c)
		}
	}
	return service.WithSignupGrantRiskInput(ctx, input)
}

func registrationHeaderSnapshot(c *gin.Context) map[string]string {
	if c == nil || c.Request == nil {
		return nil
	}
	keys := []string{
		"User-Agent",
		"Accept-Language",
		"X-Device-Fingerprint",
		"X-Forwarded-For",
		"X-Real-IP",
		"CF-Connecting-IP",
		"True-Client-IP",
		"Forwarded",
		"Origin",
		"Referer",
		"Sec-CH-UA",
		"Sec-CH-UA-Mobile",
		"Sec-CH-UA-Platform",
	}
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(c.GetHeader(key))
		if value == "" {
			continue
		}
		if len(value) > 1024 {
			value = value[:1024]
		}
		out[key] = value
	}
	return out
}

func (h *AuthHandler) ensureWebAuthSourceAllowed(c *gin.Context, source string, expectedProvider ...string) error {
	if !isWebAuthSource(source) {
		return nil
	}
	if trusted, ok := c.Get(webAuthSourceTrustedContextName); ok && trusted == true {
		return nil
	}
	if h.settingSvc == nil {
		return infraerrors.Unauthorized("WEB_SOURCE_UNAUTHORIZED", "web auth source requires bridge authorization")
	}

	inboundKey := strings.TrimSpace(c.GetHeader("x-api-key"))
	storedKey, err := h.settingSvc.GetAdminAPIKey(c.Request.Context())
	if err != nil {
		return infraerrors.InternalServer("WEB_SOURCE_AUTH_CHECK_FAILED", "failed to verify web auth source")
	}
	if storedKey == "" {
		return infraerrors.Unauthorized("WEB_SOURCE_UNAUTHORIZED", "web auth source requires bridge authorization")
	}
	if inboundKey != "" && subtle.ConstantTimeCompare([]byte(inboundKey), []byte(storedKey)) == 1 {
		return nil
	}

	if len(expectedProvider) > 0 && strings.TrimSpace(expectedProvider[0]) != "" && h.verifyWebOAuthBridgeToken(c, storedKey, expectedProvider...) {
		return nil
	}
	return infraerrors.Unauthorized("WEB_SOURCE_UNAUTHORIZED", "web auth source requires bridge authorization")
}

func (h *AuthHandler) verifyWebOAuthBridgeToken(c *gin.Context, secret string, expectedProvider ...string) bool {
	if c == nil || secret == "" {
		return false
	}
	token := strings.TrimSpace(c.Query(webBridgeTokenQueryName))
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return false
	}
	var payload webOAuthBridgeTokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return false
	}
	if !isWebAuthSource(payload.Source) || payload.Expires < time.Now().Unix() {
		return false
	}
	if len(expectedProvider) > 0 {
		provider := strings.ToLower(strings.TrimSpace(expectedProvider[0]))
		if provider != "" && !strings.EqualFold(strings.TrimSpace(payload.Provider), provider) {
			return false
		}
	}
	return true
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username          string `json:"username"`
	Email             string `json:"email" binding:"required,email"`
	Password          string `json:"password" binding:"required"`
	VerifyCode        string `json:"verify_code"`
	TurnstileToken    string `json:"turnstile_token"`
	PromoCode         string `json:"promo_code"`         // 注册优惠码
	InvitationCode    string `json:"invitation_code"`    // 邀请码
	AffCode           string `json:"aff_code"`           // 邀请返利码
	AgreementAccepted bool   `json:"agreement_accepted"` // 登录条款确认
	AgreementRevision string `json:"agreement_revision"`
	Source            string `json:"source"`
}

// SendVerifyCodeRequest 发送验证码请求
type SendVerifyCodeRequest struct {
	Email          string `json:"email" binding:"required,email"`
	TurnstileToken string `json:"turnstile_token"`
}

// SendVerifyCodeResponse 发送验证码响应
type SendVerifyCodeResponse struct {
	Message   string `json:"message"`
	Countdown int    `json:"countdown"` // 倒计时秒数
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email             string `json:"email" binding:"required,email"`
	Password          string `json:"password" binding:"required"`
	TurnstileToken    string `json:"turnstile_token"`
	AgreementAccepted bool   `json:"agreement_accepted"`
	AgreementRevision string `json:"agreement_revision"`
	Source            string `json:"source"`
}

// AuthResponse 认证响应格式（匹配前端期望）
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"` // 新增：Refresh Token
	ExpiresIn    int       `json:"expires_in,omitempty"`    // 新增：Access Token有效期（秒）
	TokenType    string    `json:"token_type"`
	User         *dto.User `json:"user"`
}

func ensureLoginUserActive(user *service.User) error {
	if user == nil {
		return infraerrors.Unauthorized("INVALID_USER", "user not found")
	}
	if !user.IsActive() {
		return service.ErrUserNotActive
	}
	return nil
}

// respondWithTokenPair 生成 Token 对并返回认证响应
// 如果 Token 对生成失败，回退到只返回 Access Token（向后兼容）
func (h *AuthHandler) respondWithTokenPair(c *gin.Context, user *service.User) {
	if err := ensureLoginUserActive(user); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user, "")
	if err != nil {
		slog.Error("failed to generate token pair", "error", err, "user_id", user.ID)
		// 回退到只返回Access Token
		token, tokenErr := h.authService.GenerateToken(user)
		if tokenErr != nil {
			response.InternalError(c, "Failed to generate token")
			return
		}
		response.Success(c, AuthResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			User:        dto.UserFromService(user),
		})
		return
	}
	response.Success(c, AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         dto.UserFromService(user),
	})
}

func (h *AuthHandler) ensureBackendModeAllowsUser(ctx context.Context, user *service.User) error {
	if user == nil {
		return infraerrors.Unauthorized("INVALID_USER", "user not found")
	}
	if h == nil || !h.isBackendModeEnabled(ctx) || user.IsAdmin() {
		return nil
	}
	return infraerrors.Forbidden("BACKEND_MODE_ADMIN_ONLY", "Backend mode is active. Only admin login is allowed.")
}

func (h *AuthHandler) ensureBackendModeAllowsNewUserLogin(ctx context.Context) error {
	if h == nil || !h.isBackendModeEnabled(ctx) {
		return nil
	}
	return infraerrors.Forbidden("BACKEND_MODE_ADMIN_ONLY", "Backend mode is active. Only admin login is allowed.")
}

func (h *AuthHandler) isBackendModeEnabled(ctx context.Context) bool {
	if h == nil || h.settingSvc == nil {
		return false
	}
	settings, err := h.settingSvc.GetPublicSettings(ctx)
	if err == nil && settings != nil {
		return settings.BackendModeEnabled
	}
	return h.settingSvc.IsBackendModeEnabled(ctx)
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	currentAgreementRevision, err := h.ensureAnonymousLoginAgreementAccepted(c.Request.Context(), agreementAcceptanceInput{
		Accepted: req.AgreementAccepted,
		Revision: req.AgreementRevision,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Turnstile 验证（邮箱验证码注册场景避免重复校验一次性 token）
	if err := h.authService.VerifyTurnstileForRegister(c.Request.Context(), req.TurnstileToken, ip.GetClientIP(c), req.VerifyCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.ensureWebAuthSourceAllowed(c, req.Source); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	riskCtx := h.signupGrantRiskContext(c, "", "")
	_, user, err := h.authService.RegisterWithVerificationSourceAndUsername(
		riskCtx,
		req.Email,
		req.Username,
		req.Password,
		req.VerifyCode,
		req.PromoCode,
		req.InvitationCode,
		req.AffCode,
		req.Source,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.recordLoginAgreementAcceptance(riskCtx, user, currentAgreementRevision); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.respondWithTokenPair(c, user)
}

// SendVerifyCode 发送邮箱验证码
// POST /api/v1/auth/send-verify-code
func (h *AuthHandler) SendVerifyCode(c *gin.Context) {
	var req SendVerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	// Turnstile 验证
	if err := h.authService.VerifyTurnstile(c.Request.Context(), req.TurnstileToken, ip.GetClientIP(c)); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	result, err := h.authService.SendVerifyCodeAsync(c.Request.Context(), req.Email, c.GetHeader("Accept-Language"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, SendVerifyCodeResponse{
		Message:   "Verification code sent successfully",
		Countdown: result.Countdown,
	})
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}
	if err := h.authService.VerifyTurnstile(c.Request.Context(), req.TurnstileToken, ip.GetClientIP(c)); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.ensureWebAuthSourceAllowed(c, req.Source); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	token, user, err := h.authService.LoginWithSource(c.Request.Context(), req.Email, req.Password, req.Source)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	_ = token // token 由 authService.Login 返回但此处由 respondWithTokenPair 重新生成

	if err := h.ensureBackendModeAllowsUser(c.Request.Context(), user); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.ensureUserAcceptedCurrentLoginAgreement(c.Request.Context(), user, agreementAcceptanceInput{
		Accepted: req.AgreementAccepted,
		Revision: req.AgreementRevision,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Check if TOTP 2FA is enabled for this user
	if h.totpService != nil && h.settingSvc.IsTotpEnabled(c.Request.Context()) && user.TotpEnabled {
		// Create a temporary login session for 2FA
		tempToken, err := h.totpService.CreateLoginSession(c.Request.Context(), user.ID, user.Email)
		if err != nil {
			response.InternalError(c, "Failed to create 2FA session")
			return
		}

		response.Success(c, TotpLoginResponse{
			Requires2FA:     true,
			TempToken:       tempToken,
			UserEmailMasked: service.MaskEmail(user.Email),
		})
		return
	}

	h.authService.RecordSuccessfulLogin(c.Request.Context(), user.ID)

	h.respondWithTokenPair(c, user)
}

// TotpLoginResponse represents the response when 2FA is required
type TotpLoginResponse struct {
	Requires2FA     bool   `json:"requires_2fa"`
	TempToken       string `json:"temp_token,omitempty"`
	UserEmailMasked string `json:"user_email_masked,omitempty"`
}

// Login2FARequest represents the 2FA login request
type Login2FARequest struct {
	TempToken string `json:"temp_token" binding:"required"`
	TotpCode  string `json:"totp_code" binding:"required,len=6"`
}

// Login2FA completes the login with 2FA verification
// POST /api/v1/auth/login/2fa
func (h *AuthHandler) Login2FA(c *gin.Context) {
	var req Login2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	slog.Debug("login_2fa_request",
		"temp_token_len", len(req.TempToken),
		"totp_code_len", len(req.TotpCode))

	// Get the login session
	session, err := h.totpService.GetLoginSession(c.Request.Context(), req.TempToken)
	if err != nil || session == nil {
		tokenPrefix := ""
		if len(req.TempToken) >= 8 {
			tokenPrefix = req.TempToken[:8]
		}
		slog.Debug("login_2fa_session_invalid",
			"temp_token_prefix", tokenPrefix,
			"error", err)
		response.BadRequest(c, "Invalid or expired 2FA session")
		return
	}

	slog.Debug("login_2fa_session_found",
		"user_id", session.UserID,
		"email", session.Email)

	// Verify the TOTP code
	if err := h.totpService.VerifyCode(c.Request.Context(), session.UserID, req.TotpCode); err != nil {
		slog.Debug("login_2fa_verify_failed",
			"user_id", session.UserID,
			"error", err)
		response.ErrorFrom(c, err)
		return
	}

	// Get the user (before session deletion so we can check backend mode)
	user, err := h.userService.GetByID(c.Request.Context(), session.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := ensureLoginUserActive(user); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if err := h.ensureBackendModeAllowsUser(c.Request.Context(), user); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if session.PendingOAuthBind != nil {
		pendingSvc, err := h.pendingIdentityService()
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}

		pendingSession, err := pendingSvc.GetBrowserSession(
			c.Request.Context(),
			session.PendingOAuthBind.PendingSessionToken,
			session.PendingOAuthBind.BrowserSessionKey,
		)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}

		decision, err := h.ensurePendingOAuthAdoptionDecision(c, pendingSession.ID, oauthAdoptionDecisionRequest{})
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if err := applyPendingOAuthBinding(
			c.Request.Context(),
			h.entClient(),
			h.authService,
			h.userService,
			pendingSession,
			decision,
			&user.ID,
			true,
			true,
		); err != nil {
			response.ErrorFrom(c, infraerrors.InternalServer("PENDING_AUTH_BIND_APPLY_FAILED", "failed to bind pending oauth identity").WithCause(err))
			return
		}
		if _, err := pendingSvc.ConsumeBrowserSession(
			c.Request.Context(),
			pendingSession.SessionToken,
			pendingSession.BrowserSessionKey,
		); err != nil {
			response.ErrorFrom(c, err)
			return
		}

		secureCookie := isRequestHTTPS(c)
		clearOAuthPendingSessionCookie(c, secureCookie)
		clearOAuthPendingBrowserCookie(c, secureCookie)
		h.authService.RecordSuccessfulLogin(c.Request.Context(), user.ID)

		user, err = h.userService.GetByID(c.Request.Context(), session.UserID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}

	// Delete the login session (only after all checks pass)
	_ = h.totpService.DeleteLoginSession(c.Request.Context(), req.TempToken)

	if session.PendingOAuthBind == nil {
		h.authService.RecordSuccessfulLogin(c.Request.Context(), user.ID)
	}

	h.respondWithTokenPair(c, user)
}

// GetCurrentUser handles getting current authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	identities, err := h.userService.GetProfileIdentitySummaries(c.Request.Context(), subject.UserID, user)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	type UserResponse struct {
		userProfileResponse
		RunMode string `json:"run_mode"`
	}

	runMode := config.RunModeStandard
	if h.cfg != nil {
		runMode = h.cfg.RunMode
	}

	response.Success(c, UserResponse{
		userProfileResponse: userProfileResponseFromService(user, identities),
		RunMode:             runMode,
	})
}

// ValidatePromoCodeRequest 验证优惠码请求
type ValidatePromoCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// ValidatePromoCodeResponse 验证优惠码响应
type ValidatePromoCodeResponse struct {
	Valid       bool    `json:"valid"`
	BonusAmount float64 `json:"bonus_amount,omitempty"`
	ErrorCode   string  `json:"error_code,omitempty"`
	Message     string  `json:"message,omitempty"`
}

// ValidatePromoCode 验证优惠码（公开接口，注册前调用）
// POST /api/v1/auth/validate-promo-code
func (h *AuthHandler) ValidatePromoCode(c *gin.Context) {
	// 检查优惠码功能是否启用
	if h.settingSvc != nil && !h.settingSvc.IsPromoCodeEnabled(c.Request.Context()) {
		response.Success(c, ValidatePromoCodeResponse{
			Valid:     false,
			ErrorCode: "PROMO_CODE_DISABLED",
		})
		return
	}

	var req ValidatePromoCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	promoCode, err := h.promoService.ValidatePromoCode(c.Request.Context(), req.Code)
	if err != nil {
		// 根据错误类型返回对应的错误码
		errorCode := "PROMO_CODE_INVALID"
		switch err {
		case service.ErrPromoCodeNotFound:
			errorCode = "PROMO_CODE_NOT_FOUND"
		case service.ErrPromoCodeExpired:
			errorCode = "PROMO_CODE_EXPIRED"
		case service.ErrPromoCodeDisabled:
			errorCode = "PROMO_CODE_DISABLED"
		case service.ErrPromoCodeMaxUsed:
			errorCode = "PROMO_CODE_MAX_USED"
		case service.ErrPromoCodeAlreadyUsed:
			errorCode = "PROMO_CODE_ALREADY_USED"
		}

		response.Success(c, ValidatePromoCodeResponse{
			Valid:     false,
			ErrorCode: errorCode,
		})
		return
	}

	if promoCode == nil {
		response.Success(c, ValidatePromoCodeResponse{
			Valid:     false,
			ErrorCode: "PROMO_CODE_INVALID",
		})
		return
	}

	response.Success(c, ValidatePromoCodeResponse{
		Valid:       true,
		BonusAmount: promoCode.BonusAmount,
	})
}

// ValidateInvitationCodeRequest 验证邀请码请求
type ValidateInvitationCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// ValidateInvitationCodeResponse 验证邀请码响应
type ValidateInvitationCodeResponse struct {
	Valid     bool   `json:"valid"`
	ErrorCode string `json:"error_code,omitempty"`
}

// ValidateInvitationCode 验证邀请码（公开接口，注册前调用）
// POST /api/v1/auth/validate-invitation-code
func (h *AuthHandler) ValidateInvitationCode(c *gin.Context) {
	// 检查邀请码功能是否启用
	if h.settingSvc == nil || !h.settingSvc.IsInvitationCodeEnabled(c.Request.Context()) {
		response.Success(c, ValidateInvitationCodeResponse{
			Valid:     false,
			ErrorCode: "INVITATION_CODE_DISABLED",
		})
		return
	}

	var req ValidateInvitationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	// 验证一次性邀请码；如果失败，再尝试好友邀请码（可复用，注册后绑定返利关系）。
	redeemCode, err := h.redeemService.GetByCode(c.Request.Context(), req.Code)
	if err != nil {
		if h.authService != nil && h.authService.CanUseAffiliateCodeAsRegistrationInvite(c.Request.Context(), req.Code) {
			response.Success(c, ValidateInvitationCodeResponse{
				Valid: true,
			})
			return
		}
		response.Success(c, ValidateInvitationCodeResponse{
			Valid:     false,
			ErrorCode: "INVITATION_CODE_NOT_FOUND",
		})
		return
	}

	// 检查类型和状态
	if redeemCode.Type != service.RedeemTypeInvitation {
		if h.authService != nil && h.authService.CanUseAffiliateCodeAsRegistrationInvite(c.Request.Context(), req.Code) {
			response.Success(c, ValidateInvitationCodeResponse{
				Valid: true,
			})
			return
		}
		response.Success(c, ValidateInvitationCodeResponse{
			Valid:     false,
			ErrorCode: "INVITATION_CODE_INVALID",
		})
		return
	}

	if redeemCode.Status != service.StatusUnused {
		if h.authService != nil && h.authService.CanUseAffiliateCodeAsRegistrationInvite(c.Request.Context(), req.Code) {
			response.Success(c, ValidateInvitationCodeResponse{
				Valid: true,
			})
			return
		}
		response.Success(c, ValidateInvitationCodeResponse{
			Valid:     false,
			ErrorCode: "INVITATION_CODE_USED",
		})
		return
	}

	response.Success(c, ValidateInvitationCodeResponse{
		Valid: true,
	})
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email          string `json:"email" binding:"required,email"`
	TurnstileToken string `json:"turnstile_token"`
}

// ForgotPasswordResponse 忘记密码响应
type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// ForgotPassword 请求密码重置
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	// Turnstile 验证
	if err := h.authService.VerifyTurnstile(c.Request.Context(), req.TurnstileToken, ip.GetClientIP(c)); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	frontendBaseURL := strings.TrimSpace(h.settingSvc.GetFrontendURL(c.Request.Context()))
	if frontendBaseURL == "" {
		slog.Error("frontend_url not configured in settings or config; cannot build password reset link")
		response.InternalError(c, "Password reset is not configured")
		return
	}

	// Request password reset (async)
	// Note: This returns success even if email doesn't exist (to prevent enumeration)
	if err := h.authService.RequestPasswordResetAsync(c.Request.Context(), req.Email, frontendBaseURL, c.GetHeader("Accept-Language")); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, ForgotPasswordResponse{
		Message: "If your email is registered, you will receive a password reset link shortly.",
	})
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// ResetPassword 重置密码
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	// Reset password
	if err := h.authService.ResetPassword(c.Request.Context(), req.Email, req.Token, req.NewPassword); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, ResetPasswordResponse{
		Message: "Your password has been reset successfully. You can now log in with your new password.",
	})
}

// ==================== Token Refresh Endpoints ====================

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // Access Token有效期（秒）
	TokenType    string `json:"token_type"`
}

// RefreshToken 刷新Token
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	result, err := h.authService.RefreshTokenPair(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Backend mode: block non-admin token refresh
	if h.settingSvc.IsBackendModeEnabled(c.Request.Context()) && result.UserRole != "admin" {
		response.Forbidden(c, "Backend mode is active. Only admin login is allowed.")
		return
	}

	response.Success(c, RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    "Bearer",
	})
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"` // 可选：撤销指定的Refresh Token
}

// LogoutResponse 登出响应
type LogoutResponse struct {
	Message string `json:"message"`
}

// Logout 用户登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	// 允许空请求体（向后兼容）
	_ = c.ShouldBindJSON(&req)

	// 如果提供了Refresh Token，撤销它
	if req.RefreshToken != "" {
		if err := h.authService.RevokeRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
			slog.Debug("failed to revoke refresh token", "error", err)
			// 不影响登出流程
		}
	}
	h.consumePendingOAuthSessionOnLogout(c)
	clearOAuthLogoutCookies(c)

	response.Success(c, LogoutResponse{
		Message: "Logged out successfully",
	})
}

// RevokeAllSessionsResponse 撤销所有会话响应
type RevokeAllSessionsResponse struct {
	Message string `json:"message"`
}

// RevokeAllSessions 撤销当前用户的所有会话
// POST /api/v1/auth/revoke-all-sessions
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if err := h.authService.RevokeAllUserTokens(c.Request.Context(), subject.UserID); err != nil {
		slog.Error("failed to revoke all sessions", "user_id", subject.UserID, "error", err)
		response.InternalError(c, "Failed to revoke sessions")
		return
	}

	response.Success(c, RevokeAllSessionsResponse{
		Message: "All sessions have been revoked. Please log in again.",
	})
}
