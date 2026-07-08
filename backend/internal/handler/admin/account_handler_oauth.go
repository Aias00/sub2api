// Package admin provides HTTP handlers for administrative operations.
package admin

import (
	"log/slog"
	"strconv"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/service"

	"github.com/gin-gonic/gin"
)

// SyncFromCRS handles syncing accounts from claude-relay-service (CRS)
// POST /api/v1/admin/accounts/sync/crs
func (h *AccountHandler) SyncFromCRS(c *gin.Context) {
	var req SyncFromCRSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	// Default to syncing proxies (can be disabled by explicitly setting false)
	syncProxies := true
	if req.SyncProxies != nil {
		syncProxies = *req.SyncProxies
	}

	result, err := h.crsSyncService.SyncFromCRS(c.Request.Context(), service.SyncFromCRSInput{
		BaseURL:            req.BaseURL,
		Username:           req.Username,
		Password:           req.Password,
		SyncProxies:        syncProxies,
		SelectedAccountIDs: req.SelectedAccountIDs,
	})
	if err != nil {
		// Provide detailed error message for CRS sync failures
		response.InternalErrorWithError(c, err)
		return
	}

	response.Success(c, result)
}

// PreviewFromCRS handles previewing accounts from CRS before sync
// POST /api/v1/admin/accounts/sync/crs/preview
func (h *AccountHandler) PreviewFromCRS(c *gin.Context) {
	var req PreviewFromCRSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	result, err := h.crsSyncService.PreviewFromCRS(c.Request.Context(), service.SyncFromCRSInput{
		BaseURL:  req.BaseURL,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.InternalErrorWithError(c, err)
		return
	}

	response.Success(c, result)
}

// ApplyOAuthCredentials 将"重新授权"得到的新凭据原子落库。
// POST /api/v1/admin/accounts/:id/apply-oauth-credentials
//
// 与通用 PUT /:id (Update) 接口的关键区别：
//   - 仅接收 type / credentials / extra 三个字段（不接受 concurrency / rpm / quota_* 等可能误传的字段）
//   - Extra 走 UpdateAccountExtra(JSONB key 级合并)，**绝不**全量覆盖；
//     避免 base_rpm / window_cost_limit / max_sessions / quota_* / privacy_mode
//     等持久化配置在重新授权后丢失
//   - 内置 ClearError + InvalidateToken，避免前端额外两次调用，
//     并修复旧路径未失效 token 缓存导致重新授权后立即 401 的隐性 bug
//
// 与 /refresh 的区别：/refresh 用现有 refresh_token 换 access_token（无用户交互），
// 本接口承接前端完成完整 OAuth 流程后的落库步骤。
func (h *AccountHandler) ApplyOAuthCredentials(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	var req ApplyOAuthCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()

	// 预检查账号存在 + OAuth 类型（与 Refresh handler 语义一致，提供更友好的错误信息）。
	existing, err := h.adminService.GetAccount(ctx, accountID)
	if err != nil {
		response.NotFound(c, "Account not found")
		return
	}
	if !existing.IsOAuth() {
		response.ErrorFrom(c, infraerrors.BadRequest("NOT_OAUTH", "cannot apply oauth credentials to non-OAuth account"))
		return
	}

	updatedAccount, err := h.adminService.UpdateAccount(ctx, accountID, &service.UpdateAccountInput{
		Type:        req.Type,
		Credentials: req.Credentials,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 增量合并 Extra（JSONB key 级 merge，绝不覆盖 base_rpm / window_cost_limit /
	// max_sessions / quota_* / privacy_mode 等持久化键）。
	// best-effort：失败仅记日志；下方 ClearAccountError 会从 DB 重新读取最新 account，
	// 因此响应里的 extra 始终以 DB 为准——这里不需要手动维护内存快照。
	if len(req.Extra) > 0 {
		if extraErr := h.adminService.UpdateAccountExtra(ctx, accountID, req.Extra); extraErr != nil {
			extraKeys := make([]string, 0, len(req.Extra))
			for k := range req.Extra {
				extraKeys = append(extraKeys, k)
			}
			slog.Error("apply_oauth_credentials.update_extra_failed",
				"account_id", accountID,
				"extra_keys", extraKeys,
				"err", extraErr,
			)
		}
	}

	if cleared, clearErr := h.adminService.ClearAccountError(ctx, accountID); clearErr != nil {
		slog.Warn("apply_oauth_credentials.clear_error_failed",
			"account_id", accountID,
			"err", clearErr,
		)
	} else if cleared != nil {
		updatedAccount = cleared
	}

	if h.tokenCacheInvalidator != nil && updatedAccount.IsOAuth() {
		if invalidateErr := h.tokenCacheInvalidator.InvalidateToken(ctx, updatedAccount); invalidateErr != nil {
			slog.Warn("apply_oauth_credentials.invalidate_token_failed",
				"account_id", accountID,
				"err", invalidateErr,
			)
		}
	}

	response.Success(c, h.buildAccountResponseWithRuntime(ctx, updatedAccount))
}

// GenerateAuthURL generates OAuth authorization URL with full scope
// POST /api/v1/admin/accounts/generate-auth-url
func (h *OAuthHandler) GenerateAuthURL(c *gin.Context) {
	var req GenerateAuthURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = GenerateAuthURLRequest{}
	}

	result, err := h.oauthService.GenerateAuthURL(c.Request.Context(), req.ProxyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

// GenerateSetupTokenURL generates OAuth authorization URL for setup token (inference only)
// POST /api/v1/admin/accounts/generate-setup-token-url
func (h *OAuthHandler) GenerateSetupTokenURL(c *gin.Context) {
	var req GenerateAuthURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = GenerateAuthURLRequest{}
	}

	result, err := h.oauthService.GenerateSetupTokenURL(c.Request.Context(), req.ProxyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

// ExchangeCode exchanges authorization code for tokens
// POST /api/v1/admin/accounts/exchange-code
func (h *OAuthHandler) ExchangeCode(c *gin.Context) {
	var req ExchangeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	tokenInfo, err := h.oauthService.ExchangeCode(c.Request.Context(), &service.ExchangeCodeInput{
		SessionID: req.SessionID,
		Code:      req.Code,
		ProxyID:   req.ProxyID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, tokenInfo)
}

// ExchangeSetupTokenCode exchanges authorization code for setup token
// POST /api/v1/admin/accounts/exchange-setup-token-code
func (h *OAuthHandler) ExchangeSetupTokenCode(c *gin.Context) {
	var req ExchangeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	tokenInfo, err := h.oauthService.ExchangeCode(c.Request.Context(), &service.ExchangeCodeInput{
		SessionID: req.SessionID,
		Code:      req.Code,
		ProxyID:   req.ProxyID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, tokenInfo)
}

// CookieAuth performs OAuth using sessionKey (cookie-based auto-auth)
// POST /api/v1/admin/accounts/cookie-auth
func (h *OAuthHandler) CookieAuth(c *gin.Context) {
	var req CookieAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	tokenInfo, err := h.oauthService.CookieAuth(c.Request.Context(), &service.CookieAuthInput{
		SessionKey: req.SessionKey,
		ProxyID:    req.ProxyID,
		Scope:      "full",
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, tokenInfo)
}

// SetupTokenCookieAuth performs OAuth using sessionKey for setup token (inference only)
// POST /api/v1/admin/accounts/setup-token-cookie-auth
func (h *OAuthHandler) SetupTokenCookieAuth(c *gin.Context) {
	var req CookieAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	tokenInfo, err := h.oauthService.CookieAuth(c.Request.Context(), &service.CookieAuthInput{
		SessionKey: req.SessionKey,
		ProxyID:    req.ProxyID,
		Scope:      "inference",
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, tokenInfo)
}
