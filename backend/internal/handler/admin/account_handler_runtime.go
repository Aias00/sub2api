// Package admin provides HTTP handlers for administrative operations.
package admin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Aias00/cloudbase/internal/domain"
	"github.com/Aias00/cloudbase/internal/pkg/antigravity"
	"github.com/Aias00/cloudbase/internal/pkg/claude"
	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/geminicli"
	"github.com/Aias00/cloudbase/internal/pkg/openai"
	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/pkg/timezone"
	"github.com/Aias00/cloudbase/internal/pkg/xai"
	"github.com/Aias00/cloudbase/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

// scheduleOpenAIResponsesProbe 异步触发 OpenAI APIKey 账号的 Responses API 能力探测。
//
// 仅对 platform=openai && type=apikey 账号生效；其他账号无操作。
// 探测本身在 goroutine 中执行（会发一次 HTTP 请求到上游），不会阻塞
// 当前请求。探测错误仅记录日志，不向上下文传播：探测失败时标记保持缺失，
// 网关会按"现状即证据"默认走 Responses。
func (h *AccountHandler) scheduleOpenAIResponsesProbe(account *service.Account) {
	if account == nil || account.Platform != service.PlatformOpenAI || account.Type != service.AccountTypeAPIKey {
		return
	}
	if h.accountTestService == nil {
		return
	}
	accountID := account.ID
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("openai_responses_probe_panic", "account_id", accountID, "recover", r)
			}
		}()
		h.accountTestService.ProbeOpenAIAPIKeyResponsesSupport(context.Background(), accountID)
	}()
}

// Test handles testing account connectivity with SSE streaming
// POST /api/v1/admin/accounts/:id/test
func (h *AccountHandler) Test(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	var req TestAccountRequest
	// Allow empty body, model_id is optional
	_ = c.ShouldBindJSON(&req)

	// Use AccountTestService to test the account with SSE streaming
	if err := h.accountTestService.TestAccountConnection(c, accountID, req.ModelID, req.Prompt, req.Mode); err != nil {
		// Error already sent via SSE, just log
		return
	}

	if h.rateLimitService != nil {
		if _, err := h.rateLimitService.RecoverAccountAfterSuccessfulTest(c.Request.Context(), accountID); err != nil {
			_ = c.Error(err)
		}
	}
}

// RecoverState handles unified recovery of recoverable account runtime state.
// POST /api/v1/admin/accounts/:id/recover-state
func (h *AccountHandler) RecoverState(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	if h.rateLimitService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Rate limit service unavailable")
		return
	}

	if _, err := h.rateLimitService.RecoverAccountState(c.Request.Context(), accountID, service.AccountRecoveryOptions{
		InvalidateToken: true,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, h.buildAccountResponseWithRuntime(c.Request.Context(), account))
}

// refreshSingleAccount refreshes credentials for a single OAuth account.
// Returns (updatedAccount, warning, error) where warning is used for Antigravity ProjectIDMissing scenario.
func (h *AccountHandler) refreshSingleAccount(ctx context.Context, account *service.Account) (*service.Account, string, error) {
	if !account.IsOAuth() {
		return nil, "", infraerrors.BadRequest("NOT_OAUTH", "cannot refresh non-OAuth account")
	}
	// spark 影子凭据由母账号管理、自身恒空,刷新无意义且会先打上游;在调用上游前早拒
	// (覆盖单账号与批量两入口;批量侧将其计为 failed 并附说明)(外审第6轮)。
	if account.IsCredentialShadow() {
		return nil, "", infraerrors.BadRequest("SPARK_SHADOW_NO_REFRESH",
			"cannot refresh spark shadow account; its credentials are managed by the parent account")
	}

	var newCredentials map[string]any

	if account.IsOpenAI() {
		tokenInfo, err := h.openaiOAuthService.RefreshAccountToken(ctx, account)
		if err != nil {
			// 刷新失败但 access_token 可能仍有效，尝试设置隐私
			h.adminService.EnsureOpenAIPrivacy(ctx, account)
			return nil, "", err
		}

		newCredentials = h.openaiOAuthService.BuildAccountCredentials(tokenInfo)
		for k, v := range account.Credentials {
			if _, exists := newCredentials[k]; !exists {
				newCredentials[k] = v
			}
		}
		newCredentials = service.NormalizeOpenAIPersonalAccessTokenCredentials(account, tokenInfo, newCredentials)
	} else if account.Platform == service.PlatformGemini {
		tokenInfo, err := h.geminiOAuthService.RefreshAccountToken(ctx, account)
		if err != nil {
			return nil, "", fmt.Errorf("failed to refresh credentials: %w", err)
		}

		newCredentials = h.geminiOAuthService.BuildAccountCredentials(tokenInfo)
		for k, v := range account.Credentials {
			if _, exists := newCredentials[k]; !exists {
				newCredentials[k] = v
			}
		}
	} else if account.Platform == service.PlatformAntigravity {
		tokenInfo, err := h.antigravityOAuthService.RefreshAccountToken(ctx, account)
		if err != nil {
			return nil, "", err
		}

		newCredentials = h.antigravityOAuthService.BuildAccountCredentials(tokenInfo)
		for k, v := range account.Credentials {
			if _, exists := newCredentials[k]; !exists {
				newCredentials[k] = v
			}
		}

		// 特殊处理 project_id：如果新值为空但旧值非空，保留旧值
		// 这确保了即使 LoadCodeAssist 失败，project_id 也不会丢失
		if newProjectID, _ := newCredentials["project_id"].(string); newProjectID == "" {
			if oldProjectID := strings.TrimSpace(account.GetCredential("project_id")); oldProjectID != "" {
				newCredentials["project_id"] = oldProjectID
			}
		}

		// 如果 project_id 获取失败，更新凭证但不标记为 error
		if tokenInfo.ProjectIDMissing {
			updatedAccount, updateErr := h.adminService.UpdateAccount(ctx, account.ID, &service.UpdateAccountInput{
				Credentials: newCredentials,
			})
			if updateErr != nil {
				return nil, "", fmt.Errorf("failed to update credentials: %w", updateErr)
			}
			h.adminService.EnsureAntigravityPrivacy(ctx, updatedAccount)
			return updatedAccount, "missing_project_id_temporary", nil
		}

		// 成功获取到 project_id，如果之前是 missing_project_id 错误则清除
		if account.Status == service.StatusError && strings.Contains(account.ErrorMessage, "missing_project_id:") {
			if _, clearErr := h.adminService.ClearAccountError(ctx, account.ID); clearErr != nil {
				return nil, "", fmt.Errorf("failed to clear account error: %w", clearErr)
			}
		}
	} else {
		// Use Anthropic/Claude OAuth service to refresh token
		tokenInfo, err := h.oauthService.RefreshAccountToken(ctx, account)
		if err != nil {
			return nil, "", err
		}

		// Copy existing credentials to preserve non-token settings (e.g., intercept_warmup_requests)
		newCredentials = make(map[string]any)
		for k, v := range account.Credentials {
			newCredentials[k] = v
		}

		// Update token-related fields
		newCredentials["access_token"] = tokenInfo.AccessToken
		newCredentials["token_type"] = tokenInfo.TokenType
		newCredentials["expires_in"] = strconv.FormatInt(tokenInfo.ExpiresIn, 10)
		newCredentials["expires_at"] = strconv.FormatInt(tokenInfo.ExpiresAt, 10)
		if strings.TrimSpace(tokenInfo.RefreshToken) != "" {
			newCredentials["refresh_token"] = tokenInfo.RefreshToken
		}
		if strings.TrimSpace(tokenInfo.Scope) != "" {
			newCredentials["scope"] = tokenInfo.Scope
		}
	}

	updatedAccount, err := h.adminService.UpdateAccount(ctx, account.ID, &service.UpdateAccountInput{
		Credentials: newCredentials,
	})
	if err != nil {
		return nil, "", err
	}

	// 刷新成功后，清除 token 缓存，确保下次请求使用新 token
	if h.tokenCacheInvalidator != nil {
		if invalidateErr := h.tokenCacheInvalidator.InvalidateToken(ctx, updatedAccount); invalidateErr != nil {
			log.Printf("[WARN] Failed to invalidate token cache for account %d: %v", updatedAccount.ID, invalidateErr)
		}
	}

	// OpenAI OAuth: 刷新成功后检查并设置 privacy_mode
	h.adminService.EnsureOpenAIPrivacy(ctx, updatedAccount)
	// Antigravity OAuth: 刷新成功后检查并设置 privacy_mode
	h.adminService.EnsureAntigravityPrivacy(ctx, updatedAccount)

	return updatedAccount, "", nil
}

// Refresh handles refreshing account credentials
// POST /api/v1/admin/accounts/:id/refresh
func (h *AccountHandler) Refresh(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	// Get account
	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.NotFound(c, "Account not found")
		return
	}

	updatedAccount, warning, err := h.refreshSingleAccount(c.Request.Context(), account)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if warning == "missing_project_id_temporary" {
		response.Success(c, gin.H{
			"message": "Token refreshed successfully, but project_id could not be retrieved (will retry automatically)",
			"warning": "missing_project_id_temporary",
		})
		return
	}

	response.Success(c, h.buildAccountResponseWithRuntime(c.Request.Context(), updatedAccount))
}

// GetStats handles getting account statistics
// GET /api/v1/admin/accounts/:id/stats
func (h *AccountHandler) GetStats(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	// Parse days parameter (default 30)
	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 90 {
			days = d
		}
	}

	// Calculate time range
	now := timezone.Now()
	endTime := timezone.StartOfDay(now.AddDate(0, 0, 1))
	startTime := timezone.StartOfDay(now.AddDate(0, 0, -days+1))

	stats, err := h.accountUsageService.GetAccountUsageStats(c.Request.Context(), accountID, startTime, endTime)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, stats)
}

// ClearError handles clearing account error
// POST /api/v1/admin/accounts/:id/clear-error
func (h *AccountHandler) ClearError(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	account, err := h.adminService.ClearAccountError(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 清除错误后，同时清除 token 缓存，确保下次请求会获取最新的 token（触发刷新或从 DB 读取）
	// 这解决了管理员重置账号状态后，旧的失效 token 仍在缓存中导致立即再次 401 的问题
	if h.tokenCacheInvalidator != nil && account.IsOAuth() {
		if invalidateErr := h.tokenCacheInvalidator.InvalidateToken(c.Request.Context(), account); invalidateErr != nil {
			log.Printf("[WARN] Failed to invalidate token cache for account %d: %v", accountID, invalidateErr)
		}
	}

	response.Success(c, h.buildAccountResponseWithRuntime(c.Request.Context(), account))
}

// RevertProxyFallback handles reverting account proxy to original before fallback.
// POST /api/v1/admin/accounts/:id/revert-proxy-fallback
func (h *AccountHandler) RevertProxyFallback(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	if err := h.adminService.RevertAccountProxyFallback(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "reverted"})
}

// GetUsage handles getting account usage information
// GET /api/v1/admin/accounts/:id/usage?source=passive|active&force=true
func (h *AccountHandler) GetUsage(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	source := c.DefaultQuery("source", "active")
	force := c.Query("force") == "true"

	var usage *service.UsageInfo
	if source == "passive" {
		usage, err = h.accountUsageService.GetPassiveUsage(c.Request.Context(), accountID)
	} else {
		usage, err = h.accountUsageService.GetUsage(c.Request.Context(), accountID, force)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, usage)
}

// ClearRateLimit handles clearing account rate limit status
// POST /api/v1/admin/accounts/:id/clear-rate-limit
func (h *AccountHandler) ClearRateLimit(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	err = h.rateLimitService.ClearRateLimit(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, h.buildAccountResponseWithRuntime(c.Request.Context(), account))
}

// ResetQuota handles resetting account quota usage
// POST /api/v1/admin/accounts/:id/reset-quota
func (h *AccountHandler) ResetQuota(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	if err := h.adminService.ResetAccountQuota(c.Request.Context(), accountID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, h.buildAccountResponseWithRuntime(c.Request.Context(), account))
}

// GetTempUnschedulable handles getting temporary unschedulable status
// GET /api/v1/admin/accounts/:id/temp-unschedulable
func (h *AccountHandler) GetTempUnschedulable(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	state, err := h.rateLimitService.GetTempUnschedStatus(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if state == nil || state.UntilUnix <= time.Now().Unix() {
		response.Success(c, gin.H{"active": false})
		return
	}

	response.Success(c, gin.H{
		"active": true,
		"state":  state,
	})
}

// ClearTempUnschedulable handles clearing temporary unschedulable status
// DELETE /api/v1/admin/accounts/:id/temp-unschedulable
func (h *AccountHandler) ClearTempUnschedulable(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	if err := h.rateLimitService.ClearTempUnschedulable(c.Request.Context(), accountID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Temp unschedulable cleared successfully"})
}

// GetTodayStats handles getting account today statistics
// GET /api/v1/admin/accounts/:id/today-stats
func (h *AccountHandler) GetTodayStats(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	stats, err := h.accountUsageService.GetTodayStats(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, stats)
}

// GetBatchTodayStats 批量获取多个账号的今日统计。
// POST /api/v1/admin/accounts/today-stats/batch
func (h *AccountHandler) GetBatchTodayStats(c *gin.Context) {
	var req BatchTodayStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	accountIDs := normalizeInt64IDList(req.AccountIDs)
	if len(accountIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	cacheKey := buildAccountTodayStatsBatchCacheKey(accountIDs)
	if cached, ok := accountTodayStatsBatchCache.Get(cacheKey); ok {
		if cached.ETag != "" {
			c.Header("ETag", cached.ETag)
			c.Header("Vary", "If-None-Match")
			if ifNoneMatchMatched(c.GetHeader("If-None-Match"), cached.ETag) {
				c.Status(http.StatusNotModified)
				return
			}
		}
		c.Header("X-Snapshot-Cache", "hit")
		response.Success(c, cached.Payload)
		return
	}

	stats, err := h.accountUsageService.GetTodayStatsBatch(c.Request.Context(), accountIDs)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	payload := gin.H{"stats": stats}
	cached := accountTodayStatsBatchCache.Set(cacheKey, payload)
	if cached.ETag != "" {
		c.Header("ETag", cached.ETag)
		c.Header("Vary", "If-None-Match")
	}
	c.Header("X-Snapshot-Cache", "miss")
	response.Success(c, payload)
}

// SetSchedulable handles toggling account schedulable status
// POST /api/v1/admin/accounts/:id/schedulable
func (h *AccountHandler) SetSchedulable(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	var req SetSchedulableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	account, err := h.adminService.SetAccountSchedulable(c.Request.Context(), accountID, req.Schedulable)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, h.buildAccountResponseWithRuntime(c.Request.Context(), account))
}

// GetAvailableModels handles getting available models for an account
// GET /api/v1/admin/accounts/:id/models
func (h *AccountHandler) GetAvailableModels(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.NotFound(c, "Account not found")
		return
	}

	// Handle OpenAI accounts
	if account.IsOpenAI() {
		// OpenAI 自动透传会绕过常规模型改写，测试/模型列表也应回落到默认模型集。
		if account.IsOpenAIPassthroughEnabled() {
			response.Success(c, openai.DefaultModels)
			return
		}

		mapping := account.GetModelMapping()
		if len(mapping) == 0 {
			response.Success(c, openai.DefaultModels)
			return
		}

		// Return mapped models
		var models []openai.Model
		for requestedModel := range mapping {
			var found bool
			for _, dm := range openai.DefaultModels {
				if dm.ID == requestedModel {
					models = append(models, dm)
					found = true
					break
				}
			}
			if !found {
				models = append(models, openai.Model{
					ID:          requestedModel,
					Object:      "model",
					Type:        "model",
					DisplayName: requestedModel,
				})
			}
		}
		response.Success(c, models)
		return
	}

	// Handle Gemini accounts
	if account.IsGemini() {
		// For OAuth accounts: return default Gemini models
		if account.IsOAuth() {
			response.Success(c, geminicli.DefaultModels)
			return
		}

		// For API Key accounts: return models based on model_mapping
		mapping := account.GetModelMapping()
		if len(mapping) == 0 {
			response.Success(c, geminicli.DefaultModels)
			return
		}

		var models []geminicli.Model
		for requestedModel := range mapping {
			var found bool
			for _, dm := range geminicli.DefaultModels {
				if dm.ID == requestedModel {
					models = append(models, dm)
					found = true
					break
				}
			}
			if !found {
				models = append(models, geminicli.Model{
					ID:          requestedModel,
					Type:        "model",
					DisplayName: requestedModel,
					CreatedAt:   "",
				})
			}
		}
		response.Success(c, models)
		return
	}

	// Handle Antigravity accounts: return Claude + Gemini models
	if account.Platform == service.PlatformAntigravity {
		// 直接复用 antigravity.DefaultModels()，与 /v1/models 端点保持同步
		response.Success(c, antigravity.DefaultModels())
		return
	}

	// Handle Grok accounts
	if account.Platform == service.PlatformGrok {
		defaultModels := xai.DefaultModels()

		hasExplicitMapping := false
		switch rawMapping := account.Credentials["model_mapping"].(type) {
		case map[string]any:
			hasExplicitMapping = len(rawMapping) > 0
		case map[string]string:
			hasExplicitMapping = len(rawMapping) > 0
		}
		if !hasExplicitMapping {
			response.Success(c, defaultModels)
			return
		}

		mapping := account.GetModelMapping()
		if len(mapping) == 0 {
			response.Success(c, defaultModels)
			return
		}

		defaultByID := make(map[string]xai.Model, len(defaultModels))
		for _, model := range defaultModels {
			defaultByID[model.ID] = model
		}

		requestedModels := make([]string, 0, len(mapping))
		for requestedModel := range mapping {
			requestedModels = append(requestedModels, requestedModel)
		}
		sort.Strings(requestedModels)

		var models []xai.Model
		for _, requestedModel := range requestedModels {
			if defaultModel, found := defaultByID[requestedModel]; found {
				models = append(models, defaultModel)
				continue
			}
			models = append(models, xai.Model{
				ID:          requestedModel,
				Object:      "model",
				OwnedBy:     "xai",
				DisplayName: requestedModel,
			})
		}
		response.Success(c, models)
		return
	}

	// Handle Claude/Anthropic accounts
	// For OAuth and Setup-Token accounts: return default models
	if account.IsOAuth() {
		response.Success(c, claude.DefaultModels)
		return
	}

	// For API Key accounts: return models based on model_mapping
	mapping := account.GetModelMapping()
	if len(mapping) == 0 {
		// No mapping configured, return default models
		response.Success(c, claude.DefaultModels)
		return
	}

	// Return mapped models (keys of the mapping are the available model IDs)
	var models []claude.Model
	for requestedModel := range mapping {
		// Try to find display info from default models
		var found bool
		for _, dm := range claude.DefaultModels {
			if dm.ID == requestedModel {
				models = append(models, dm)
				found = true
				break
			}
		}
		// If not found in defaults, create a basic entry
		if !found {
			models = append(models, claude.Model{
				ID:          requestedModel,
				Type:        "model",
				DisplayName: requestedModel,
				CreatedAt:   "",
			})
		}
	}

	response.Success(c, models)
}

// SyncUpstreamModels handles syncing live supported models from an account's upstream.
// POST /api/v1/admin/accounts/:id/models/sync-upstream
func (h *AccountHandler) SyncUpstreamModels(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.NotFound(c, "Account not found")
		return
	}

	if h.accountTestService == nil {
		response.InternalError(c, "Account test service is not configured")
		return
	}

	models, err := h.accountTestService.FetchUpstreamSupportedModels(c.Request.Context(), account)
	if err != nil {
		var syncErr *service.UpstreamModelSyncError
		if errors.As(err, &syncErr) {
			switch syncErr.Kind {
			case service.UpstreamModelSyncErrorConfiguration, service.UpstreamModelSyncErrorUnsupported:
				response.BadRequest(c, syncErr.SafeMessage())
			default:
				slog.Warn("sync_upstream_models_failed", "account_id", accountID, "kind", syncErr.Kind)
				response.Error(c, http.StatusBadGateway, syncErr.SafeMessage())
			}
			return
		}

		slog.Warn("sync_upstream_models_failed", "account_id", accountID)
		response.Error(c, http.StatusBadGateway, "Failed to sync upstream models from upstream")
		return
	}

	response.Success(c, gin.H{"models": models})
}

// SyncUpstreamModelsPreview handles syncing live supported models using provided credentials (no account ID needed).
// POST /api/v1/admin/accounts/models/sync-upstream-preview
func (h *AccountHandler) SyncUpstreamModelsPreview(c *gin.Context) {
	var req struct {
		Platform string `json:"platform" binding:"required"`
		Type     string `json:"type" binding:"required"`
		BaseURL  string `json:"base_url"`
		APIKey   string `json:"api_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	tempAccount := &service.Account{
		Platform: req.Platform,
		Type:     req.Type,
		Credentials: map[string]any{
			"api_key":  req.APIKey,
			"base_url": req.BaseURL,
		},
	}

	if h.accountTestService == nil {
		response.InternalError(c, "Account test service is not configured")
		return
	}

	models, err := h.accountTestService.FetchUpstreamSupportedModels(c.Request.Context(), tempAccount)
	if err != nil {
		var syncErr *service.UpstreamModelSyncError
		if errors.As(err, &syncErr) {
			switch syncErr.Kind {
			case service.UpstreamModelSyncErrorConfiguration, service.UpstreamModelSyncErrorUnsupported:
				response.BadRequest(c, syncErr.SafeMessage())
			default:
				slog.Warn("sync_upstream_models_preview_failed", "platform", req.Platform, "kind", syncErr.Kind)
				response.Error(c, http.StatusBadGateway, syncErr.SafeMessage())
			}
			return
		}

		slog.Warn("sync_upstream_models_preview_failed", "platform", req.Platform)
		response.Error(c, http.StatusBadGateway, "Failed to sync upstream models from upstream")
		return
	}

	response.Success(c, gin.H{"models": models})
}

// SetPrivacy handles setting privacy for a single OpenAI/Antigravity OAuth account
// POST /api/v1/admin/accounts/:id/set-privacy
func (h *AccountHandler) SetPrivacy(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.NotFound(c, "Account not found")
		return
	}
	if account.Type != service.AccountTypeOAuth {
		response.BadRequest(c, "Only OAuth accounts support privacy setting")
		return
	}
	var mode string
	switch account.Platform {
	case service.PlatformOpenAI:
		mode = h.adminService.ForceOpenAIPrivacy(c.Request.Context(), account)
	case service.PlatformAntigravity:
		mode = h.adminService.ForceAntigravityPrivacy(c.Request.Context(), account)
	default:
		response.BadRequest(c, "Only OpenAI and Antigravity OAuth accounts support privacy setting")
		return
	}
	if mode == "" {
		response.BadRequest(c, "Cannot set privacy: missing access_token")
		return
	}
	// 从 DB 重新读取以确保返回最新状态
	updated, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		// 隐私已设置成功但读取失败，回退到内存更新
		if account.Extra == nil {
			account.Extra = make(map[string]any)
		}
		account.Extra["privacy_mode"] = mode
		response.Success(c, h.buildAccountResponseWithRuntime(c.Request.Context(), account))
		return
	}
	response.Success(c, h.buildAccountResponseWithRuntime(c.Request.Context(), updated))
}

// RefreshTier handles refreshing Google One tier for a single account
// POST /api/v1/admin/accounts/:id/refresh-tier
func (h *AccountHandler) RefreshTier(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	ctx := c.Request.Context()
	account, err := h.adminService.GetAccount(ctx, accountID)
	if err != nil {
		response.NotFound(c, "Account not found")
		return
	}

	if account.Platform != service.PlatformGemini || account.Type != service.AccountTypeOAuth {
		response.BadRequest(c, "Only Gemini OAuth accounts support tier refresh")
		return
	}

	oauthType, _ := account.Credentials["oauth_type"].(string)
	if oauthType != "google_one" {
		response.BadRequest(c, "Only google_one OAuth accounts support tier refresh")
		return
	}

	tierID, extra, creds, err := h.geminiOAuthService.RefreshAccountGoogleOneTier(ctx, account)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	_, updateErr := h.adminService.UpdateAccount(ctx, accountID, &service.UpdateAccountInput{
		Credentials: creds,
		Extra:       extra,
	})
	if updateErr != nil {
		response.ErrorFrom(c, updateErr)
		return
	}

	response.Success(c, gin.H{
		"tier_id":             tierID,
		"storage_info":        extra,
		"drive_storage_limit": extra["drive_storage_limit"],
		"drive_storage_usage": extra["drive_storage_usage"],
		"updated_at":          extra["drive_tier_updated_at"],
	})
}

// BatchRefreshTier handles batch refreshing Google One tier
// POST /api/v1/admin/accounts/batch-refresh-tier
func (h *AccountHandler) BatchRefreshTier(c *gin.Context) {
	var req BatchRefreshTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = BatchRefreshTierRequest{}
	}

	ctx := c.Request.Context()
	accounts := make([]*service.Account, 0)

	if len(req.AccountIDs) == 0 {
		allAccounts, _, err := h.adminService.ListAccounts(ctx, 1, 10000, "gemini", "oauth", "", "", 0, "", "name", "asc")
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for i := range allAccounts {
			acc := &allAccounts[i]
			oauthType, _ := acc.Credentials["oauth_type"].(string)
			if oauthType == "google_one" {
				accounts = append(accounts, acc)
			}
		}
	} else {
		fetched, err := h.adminService.GetAccountsByIDs(ctx, req.AccountIDs)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}

		for _, acc := range fetched {
			if acc == nil {
				continue
			}
			if acc.Platform != service.PlatformGemini || acc.Type != service.AccountTypeOAuth {
				continue
			}
			oauthType, _ := acc.Credentials["oauth_type"].(string)
			if oauthType != "google_one" {
				continue
			}
			accounts = append(accounts, acc)
		}
	}

	const maxConcurrency = 10
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrency)

	var mu sync.Mutex
	var successCount, failedCount int
	var errors []gin.H

	for _, account := range accounts {
		acc := account // 闭包捕获
		g.Go(func() error {
			_, extra, creds, err := h.geminiOAuthService.RefreshAccountGoogleOneTier(gctx, acc)
			if err != nil {
				mu.Lock()
				failedCount++
				errors = append(errors, gin.H{
					"account_id": acc.ID,
					"error":      err.Error(),
				})
				mu.Unlock()
				return nil
			}

			_, updateErr := h.adminService.UpdateAccount(gctx, acc.ID, &service.UpdateAccountInput{
				Credentials: creds,
				Extra:       extra,
			})

			mu.Lock()
			if updateErr != nil {
				failedCount++
				errors = append(errors, gin.H{
					"account_id": acc.ID,
					"error":      updateErr.Error(),
				})
			} else {
				successCount++
			}
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	results := gin.H{
		"total":   len(accounts),
		"success": successCount,
		"failed":  failedCount,
		"errors":  errors,
	}

	response.Success(c, results)
}

// GetAntigravityDefaultModelMapping 获取 Antigravity 平台的默认模型映射
// GET /api/v1/admin/accounts/antigravity/default-model-mapping
func (h *AccountHandler) GetAntigravityDefaultModelMapping(c *gin.Context) {
	response.Success(c, domain.DefaultAntigravityModelMapping)
}
