package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Aias00/cloudbase/internal/pkg/logger"
	"github.com/gin-gonic/gin"
)

func (s *AntigravityGatewayService) upstreamErrorBodyReadLimit() int64 {
	limit := gatewayUpstreamErrorBodyReadLimit
	if s != nil && s.settingService != nil && s.settingService.cfg != nil && s.settingService.cfg.Gateway.LogUpstreamErrorBody && s.settingService.cfg.Gateway.LogUpstreamErrorBodyMaxBytes > int(limit) {
		limit = int64(s.settingService.cfg.Gateway.LogUpstreamErrorBodyMaxBytes)
	}
	return limit
}

func (s *AntigravityGatewayService) readUpstreamErrorBody(resp *http.Response) []byte {
	if resp == nil || resp.Body == nil {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, s.upstreamErrorBodyReadLimit()))
	return body
}

// getUpstreamErrorDetail 获取上游错误详情（用于日志记录）
func (s *AntigravityGatewayService) getUpstreamErrorDetail(body []byte) string {
	logBody, maxBytes := s.getLogConfig()
	if !logBody {
		return ""
	}
	return truncateString(string(body), maxBytes)
}

// checkErrorPolicy nil 安全的包装
func (s *AntigravityGatewayService) checkErrorPolicy(ctx context.Context, account *Account, statusCode int, body []byte) ErrorPolicyResult {
	if s.rateLimitService == nil {
		return ErrorPolicyNone
	}
	return s.rateLimitService.CheckErrorPolicy(ctx, account, statusCode, body)
}

// applyErrorPolicy 应用错误策略结果，返回是否应终止当前循环及应返回的状态码。
// ErrorPolicySkipped 时 outStatus 为 500（前端约定：未命中的错误返回 500）。
func (s *AntigravityGatewayService) applyErrorPolicy(p antigravityRetryLoopParams, statusCode int, headers http.Header, respBody []byte) (handled bool, outStatus int, retErr error) {
	switch s.checkErrorPolicy(p.ctx, p.account, statusCode, respBody) {
	case ErrorPolicySkipped:
		if s.handleAntigravityModelRateLimitBeforePolicy(p, statusCode, headers, respBody) {
			return true, statusCode, nil
		}
		return true, http.StatusInternalServerError, nil
	case ErrorPolicyMatched:
		if s.handleAntigravityModelRateLimitBeforePolicy(p, statusCode, headers, respBody) {
			return true, statusCode, nil
		}
		_ = p.handleError(p.ctx, p.prefix, p.account, statusCode, headers, respBody,
			p.requestedModel, p.groupID, p.sessionHash, p.isStickySession)
		return true, statusCode, nil
	case ErrorPolicyTempUnscheduled:
		slog.Info("temp_unschedulable_matched",
			"prefix", p.prefix, "status_code", statusCode, "account_id", p.account.ID)
		return true, statusCode, &AntigravityAccountSwitchError{OriginalAccountID: p.account.ID, RateLimitedModel: p.requestedModel, IsStickySession: p.isStickySession}
	}
	return false, statusCode, nil
}

func (s *AntigravityGatewayService) handleAntigravityModelRateLimitBeforePolicy(p antigravityRetryLoopParams, statusCode int, headers http.Header, respBody []byte) bool {
	if statusCode != http.StatusTooManyRequests && statusCode != http.StatusServiceUnavailable {
		return false
	}
	if p.account == nil || p.account.Platform != PlatformAntigravity {
		return false
	}
	_, shouldRateLimitModel, waitDuration, modelName, isModelCapacityExhausted := shouldTriggerAntigravitySmartRetry(p.account, respBody)
	if isModelCapacityExhausted || !shouldRateLimitModel || strings.TrimSpace(modelName) == "" {
		return false
	}
	rateLimitDuration := waitDuration
	if rateLimitDuration <= 0 {
		rateLimitDuration = antigravityDefaultRateLimitDuration
	}
	resetAt := time.Now().Add(rateLimitDuration)
	if !s.setAntigravityModelRateLimits(p.ctx, p.accountRepo, p.account, modelName, p.prefix, statusCode, resetAt, false) {
		return false
	}
	s.clearStickySession(p.ctx, p.groupID, p.sessionHash)
	logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limited_before_error_policy model=%s account=%d reset_in=%v",
		p.prefix, statusCode, modelName, p.account.ID, rateLimitDuration)
	return true
}

// testConnectionHandleError 是 TestConnection 使用的轻量 handleError 回调。
// 仅记录日志，不做 ops 错误追踪或粘性会话清除。
func testConnectionHandleError(
	_ context.Context, prefix string, account *Account,
	statusCode int, _ http.Header, body []byte,
	requestedModel string, _ int64, _ string, _ bool,
) *handleModelRateLimitResult {
	logger.LegacyPrintf("service.antigravity_gateway",
		"%s test_handle_error status=%d model=%s account=%d body=%s",
		prefix, statusCode, requestedModel, account.ID, truncateForLog(body, 200))
	return nil
}

func isSignatureRelatedError(respBody []byte) bool {
	msg := strings.ToLower(strings.TrimSpace(extractAntigravityErrorMessage(respBody)))
	if msg == "" {
		// Fallback: best-effort scan of the raw payload.
		msg = strings.ToLower(string(respBody))
	}

	// Keep this intentionally broad: different upstreams may use "signature" or "thought_signature".
	if strings.Contains(msg, "thought_signature") || strings.Contains(msg, "signature") {
		return true
	}

	// Also detect thinking block structural errors:
	// "Expected `thinking` or `redacted_thinking`, but found `text`"
	if strings.Contains(msg, "expected") && (strings.Contains(msg, "thinking") || strings.Contains(msg, "redacted_thinking")) {
		return true
	}

	return false
}

// isPromptTooLongError 检测是否为 prompt too long 错误
func isPromptTooLongError(respBody []byte) bool {
	msg := strings.ToLower(strings.TrimSpace(extractAntigravityErrorMessage(respBody)))
	if msg == "" {
		msg = strings.ToLower(string(respBody))
	}
	return strings.Contains(msg, "prompt is too long") ||
		strings.Contains(msg, "request is too long") ||
		strings.Contains(msg, "context length exceeded") ||
		strings.Contains(msg, "max_tokens")
}

// isPassthroughErrorMessage 检查错误消息是否在透传白名单中
func isPassthroughErrorMessage(msg string) bool {
	lower := strings.ToLower(msg)
	for _, pattern := range antigravityPassthroughErrorMessages {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// getPassthroughOrDefault 若消息在白名单内则返回原始消息，否则返回默认消息
func getPassthroughOrDefault(upstreamMsg, defaultMsg string) string {
	if isPassthroughErrorMessage(upstreamMsg) {
		return upstreamMsg
	}
	return defaultMsg
}

func extractAntigravityErrorMessage(body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	// Google-style: {"error": {"message": "..."}}
	if errObj, ok := payload["error"].(map[string]any); ok {
		if msg, ok := errObj["message"].(string); ok && strings.TrimSpace(msg) != "" {
			return msg
		}
	}

	// Fallback: top-level message
	if msg, ok := payload["message"].(string); ok && strings.TrimSpace(msg) != "" {
		return msg
	}

	return ""
}

func (s *AntigravityGatewayService) shouldFailoverUpstreamError(statusCode int) bool {
	switch statusCode {
	case 401, 403, 429, 529:
		return true
	default:
		return statusCode >= 500
	}
}

// isGoogleProjectConfigError 判断（已提取的小写）错误消息是否属于 Google 服务端配置类问题。
// 只精确匹配已知的服务端侧错误，避免对客户端请求错误做无意义重试。
// 适用于所有走 Google 后端的平台（Antigravity、Gemini）。
func isGoogleProjectConfigError(lowerMsg string) bool {
	// Google 间歇性 Bug：Project ID 有效但被临时识别失败
	return strings.Contains(lowerMsg, "invalid project resource name")
}

// tempUnscheduleGoogleConfigError 对服务端配置类 400 错误触发临时封禁，
// 避免短时间内反复调度到同一个有问题的账号。
func tempUnscheduleGoogleConfigError(ctx context.Context, repo AccountRepository, accountID int64, logPrefix string) {
	until := time.Now().Add(googleConfigErrorCooldown)
	reason := "400: invalid project resource name (auto temp-unschedule 1m)"
	if err := repo.SetTempUnschedulable(ctx, accountID, until, reason); err != nil {
		log.Printf("%s temp_unschedule_failed account=%d error=%v", logPrefix, accountID, err)
	} else {
		log.Printf("%s temp_unscheduled account=%d until=%v reason=%q", logPrefix, accountID, until.Format("15:04:05"), reason)
	}
}

// tempUnscheduleEmptyResponse 对空流式响应触发临时封禁，
// 避免短时间内反复调度到同一个返回空响应的账号。
func tempUnscheduleEmptyResponse(ctx context.Context, repo AccountRepository, accountID int64, logPrefix string) {
	until := time.Now().Add(emptyResponseCooldown)
	reason := "empty stream response (auto temp-unschedule 1m)"
	if err := repo.SetTempUnschedulable(ctx, accountID, until, reason); err != nil {
		log.Printf("%s temp_unschedule_failed account=%d error=%v", logPrefix, accountID, err)
	} else {
		log.Printf("%s temp_unscheduled account=%d until=%v reason=%q", logPrefix, accountID, until.Format("15:04:05"), reason)
	}
}

// setModelRateLimitByModelName 使用官方模型 ID 设置模型级限流
// 直接使用上游返回的模型 ID（如 claude-sonnet-4-5）作为限流 key
// 返回是否已成功设置（若模型名为空或 repo 为 nil 将返回 false）
func setModelRateLimitByModelName(ctx context.Context, repo AccountRepository, accountID int64, modelName, prefix string, statusCode int, resetAt time.Time, afterSmartRetry bool) bool {
	if repo == nil || modelName == "" {
		return false
	}
	// 直接使用官方模型 ID 作为 key，不再转换为 scope
	if err := repo.SetModelRateLimit(ctx, accountID, modelName, resetAt); err != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limit_failed model=%s error=%v", prefix, statusCode, modelName, err)
		return false
	}
	if afterSmartRetry {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limited_after_smart_retry model=%s account=%d reset_in=%v", prefix, statusCode, modelName, accountID, time.Until(resetAt).Truncate(time.Second))
	} else {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limited model=%s account=%d reset_in=%v", prefix, statusCode, modelName, accountID, time.Until(resetAt).Truncate(time.Second))
	}
	return true
}

func (s *AntigravityGatewayService) setAntigravityModelRateLimits(ctx context.Context, repo AccountRepository, account *Account, modelName, prefix string, statusCode int, resetAt time.Time, afterSmartRetry bool) bool {
	if account == nil || repo == nil {
		return false
	}
	keys := antigravityModelRateLimitKeys(modelName)
	if len(keys) == 0 {
		return false
	}

	success := false
	for _, key := range keys {
		if setModelRateLimitByModelName(ctx, repo, account.ID, key, prefix, statusCode, resetAt, afterSmartRetry) {
			s.updateAccountModelRateLimitInCache(ctx, account, key, resetAt)
			success = true
		}
	}
	return success
}

func (s *AntigravityGatewayService) clearStickySession(ctx context.Context, groupID int64, sessionHash string) {
	if s == nil || s.cache == nil || strings.TrimSpace(sessionHash) == "" {
		return
	}
	if err := s.cache.DeleteSessionAccountID(ctx, groupID, sessionHash); err != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] sticky_session_clear_failed group_id=%d session=%s err=%v", groupID, shortSessionHash(sessionHash), err)
	}
}

// handleModelRateLimit 处理模型级限流（在原有逻辑之前调用）
// 仅处理 429/503，解析模型名和 retryDelay
// - MODEL_CAPACITY_EXHAUSTED: 返回 Handled=true（实际重试由 handleSmartRetry 处理）
// - RATE_LIMIT_EXCEEDED + retryDelay < 阈值: 返回 ShouldRetry=true，由调用方等待后重试
// - RATE_LIMIT_EXCEEDED + retryDelay >= 阈值: 设置模型限流 + 清除粘性会话 + 返回 SwitchError
func (s *AntigravityGatewayService) handleModelRateLimit(p *handleModelRateLimitParams) *handleModelRateLimitResult {
	if p.statusCode != 429 && p.statusCode != 503 {
		return &handleModelRateLimitResult{Handled: false}
	}

	info := parseAntigravitySmartRetryInfo(p.body)
	if info == nil || info.ModelName == "" {
		return &handleModelRateLimitResult{Handled: false}
	}

	// MODEL_CAPACITY_EXHAUSTED：模型容量不足，所有账号共享同一容量池
	// 切换账号无意义，不设置模型限流（实际重试由 handleSmartRetry 处理）
	if info.IsModelCapacityExhausted {
		log.Printf("%s status=%d model_capacity_exhausted model=%s (not switching account, retry handled by smart retry)",
			p.prefix, p.statusCode, info.ModelName)
		return &handleModelRateLimitResult{
			Handled: true,
		}
	}

	// RATE_LIMIT_EXCEEDED: < antigravityRateLimitThreshold: 等待后重试
	if info.RetryDelay < antigravityRateLimitThreshold {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limit_wait model=%s wait=%v",
			p.prefix, p.statusCode, info.ModelName, info.RetryDelay)
		return &handleModelRateLimitResult{
			Handled:      true,
			ShouldRetry:  true,
			WaitDuration: info.RetryDelay,
		}
	}

	// RATE_LIMIT_EXCEEDED: >= antigravityRateLimitThreshold: 设置限流 + 清除粘性会话 + 切换账号
	s.setModelRateLimitAndClearSession(p, info)

	return &handleModelRateLimitResult{
		Handled: true,
		SwitchError: &AntigravityAccountSwitchError{
			OriginalAccountID: p.account.ID,
			RateLimitedModel:  info.ModelName,
			IsStickySession:   p.isStickySession,
		},
	}
}

// setModelRateLimitAndClearSession 设置模型限流并清除粘性会话
func (s *AntigravityGatewayService) setModelRateLimitAndClearSession(p *handleModelRateLimitParams, info *antigravitySmartRetryInfo) {
	resetAt := time.Now().Add(info.RetryDelay)
	logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d model_rate_limited model=%s account=%d reset_in=%v",
		p.prefix, p.statusCode, info.ModelName, p.account.ID, info.RetryDelay)

	s.setAntigravityModelRateLimits(p.ctx, s.accountRepo, p.account, info.ModelName, p.prefix, p.statusCode, resetAt, false)

	// 清除粘性会话绑定
	if p.cache != nil && p.sessionHash != "" {
		_ = p.cache.DeleteSessionAccountID(p.ctx, p.groupID, p.sessionHash)
	}
}

// updateAccountModelRateLimitInCache 立即更新 Redis 中账号的模型限流状态
func (s *AntigravityGatewayService) updateAccountModelRateLimitInCache(ctx context.Context, account *Account, modelKey string, resetAt time.Time) {
	if s.schedulerSnapshot == nil || account == nil || modelKey == "" {
		return
	}

	// 更新账号对象的 Extra 字段
	if account.Extra == nil {
		account.Extra = make(map[string]any)
	}

	limits, _ := account.Extra["model_rate_limits"].(map[string]any)
	if limits == nil {
		limits = make(map[string]any)
		account.Extra["model_rate_limits"] = limits
	}

	limits[modelKey] = map[string]any{
		"rate_limited_at":     time.Now().UTC().Format(time.RFC3339),
		"rate_limit_reset_at": resetAt.UTC().Format(time.RFC3339),
	}

	// 更新 Redis 快照
	if err := s.schedulerSnapshot.UpdateAccountInCache(ctx, account); err != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] cache_update_failed account=%d model=%s err=%v", account.ID, modelKey, err)
	}
}

func (s *AntigravityGatewayService) handleUpstreamError(
	ctx context.Context, prefix string, account *Account,
	statusCode int, headers http.Header, body []byte,
	requestedModel string,
	groupID int64, sessionHash string, isStickySession bool,
) *handleModelRateLimitResult {
	// 遵守自定义错误码策略：未命中则跳过所有限流处理
	if !account.ShouldHandleErrorCode(statusCode) {
		return nil
	}
	// 模型级限流处理（优先）
	result := s.handleModelRateLimit(&handleModelRateLimitParams{
		ctx:             ctx,
		prefix:          prefix,
		account:         account,
		statusCode:      statusCode,
		body:            body,
		cache:           s.cache,
		groupID:         groupID,
		sessionHash:     sessionHash,
		isStickySession: isStickySession,
	})
	if result.Handled {
		return result
	}

	// 503 仅处理模型限流（MODEL_CAPACITY_EXHAUSTED），非模型限流不做额外处理
	// 避免将普通的 503 错误误判为账号问题
	if statusCode == 503 {
		return nil
	}

	// 429：尝试解析模型级限流，解析失败时兜底为账号级限流
	if statusCode == 429 {
		if logBody, maxBytes := s.getLogConfig(); logBody {
			logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity-Debug] 429 response body: %s", truncateString(string(body), maxBytes))
		}

		resetAt := ParseGeminiRateLimitResetTime(body)
		defaultDur := s.getDefaultRateLimitDuration()

		// 尝试解析模型 key 并设置模型级限流
		//
		// 注意：requestedModel 可能是"映射前"的请求模型名（例如 claude-opus-4-6），
		// 调度与限流判定使用的是 Antigravity 最终模型名（包含映射与 thinking 后缀）。
		// 因此这里必须写入最终模型 key，确保后续调度能正确避开已限流模型。
		modelKey := resolveFinalAntigravityModelKey(ctx, account, requestedModel)
		if strings.TrimSpace(modelKey) == "" {
			// 极少数情况下无法映射（理论上不应发生：能转发成功说明映射已通过），
			// 保持旧行为作为兜底，避免完全丢失模型级限流记录。
			modelKey = resolveAntigravityModelKey(requestedModel)
		}
		if modelKey != "" {
			ra := s.resolveResetTime(resetAt, defaultDur)
			if !s.setAntigravityModelRateLimits(ctx, s.accountRepo, account, modelKey, prefix, statusCode, ra, false) {
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 model_rate_limit_set_failed model=%s", prefix, modelKey)
			} else {
				logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 model_rate_limited model=%s account=%d reset_at=%v reset_in=%v",
					prefix, modelKey, account.ID, ra.Format("15:04:05"), time.Until(ra).Truncate(time.Second))
			}
			return nil
		}

		// 无法解析模型 key，兜底为账号级限流
		ra := s.resolveResetTime(resetAt, defaultDur)
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 rate_limited account=%d reset_at=%v reset_in=%v (fallback)",
			prefix, account.ID, ra.Format("15:04:05"), time.Until(ra).Truncate(time.Second))
		if err := s.accountRepo.SetRateLimited(ctx, account.ID, ra); err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 rate_limit_set_failed account=%d error=%v", prefix, account.ID, err)
		}
		return nil
	}
	// 其他错误码继续使用 rateLimitService
	if s.rateLimitService == nil {
		return nil
	}
	shouldDisable := s.rateLimitService.HandleUpstreamError(ctx, account, statusCode, headers, body)
	if shouldDisable {
		logger.LegacyPrintf("service.antigravity_gateway", "%s status=%d marked_error", prefix, statusCode)
	}
	return nil
}

// getDefaultRateLimitDuration 获取默认限流时间
func (s *AntigravityGatewayService) getDefaultRateLimitDuration() time.Duration {
	defaultDur := antigravityDefaultRateLimitDuration
	if s.settingService != nil && s.settingService.cfg != nil && s.settingService.cfg.Gateway.AntigravityFallbackCooldownMinutes > 0 {
		defaultDur = time.Duration(s.settingService.cfg.Gateway.AntigravityFallbackCooldownMinutes) * time.Minute
	}
	if override, ok := antigravityFallbackCooldownSeconds(); ok {
		defaultDur = override
	}
	return defaultDur
}

// resolveResetTime 根据解析的重置时间或默认时长计算重置时间点
func (s *AntigravityGatewayService) resolveResetTime(resetAt *int64, defaultDur time.Duration) time.Time {
	if resetAt != nil {
		return time.Unix(*resetAt, 0)
	}
	return time.Now().Add(defaultDur)
}

func (s *AntigravityGatewayService) writeClaudeError(c *gin.Context, status int, errType, message string) error {
	MarkResponseCommitted(c)
	c.JSON(status, gin.H{
		"type":  "error",
		"error": gin.H{"type": errType, "message": message},
	})
	return fmt.Errorf("%s", message)
}

// WriteMappedClaudeError 导出版本，供 handler 层使用（如 fallback 错误处理）
func (s *AntigravityGatewayService) WriteMappedClaudeError(c *gin.Context, account *Account, upstreamStatus int, upstreamRequestID string, body []byte) error {
	return s.writeMappedClaudeError(c, account, upstreamStatus, upstreamRequestID, body)
}

func (s *AntigravityGatewayService) writeMappedClaudeError(c *gin.Context, account *Account, upstreamStatus int, upstreamRequestID string, body []byte) error {
	MarkResponseCommitted(c)
	upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(body))
	upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
	logBody, maxBytes := s.getLogConfig()
	upstreamDetail := s.getUpstreamErrorDetail(body)
	setOpsUpstreamError(c, upstreamStatus, upstreamMsg, upstreamDetail)
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform:           account.Platform,
		AccountID:          account.ID,
		AccountName:        account.Name,
		UpstreamStatusCode: upstreamStatus,
		UpstreamRequestID:  upstreamRequestID,
		Kind:               "http_error",
		Message:            upstreamMsg,
		Detail:             upstreamDetail,
	})

	// 记录上游错误详情便于排障（可选：由配置控制；不回显到客户端）
	if logBody {
		logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] upstream_error status=%d body=%s", upstreamStatus, truncateForLog(body, maxBytes))
	}

	// 检查错误透传规则
	if ptStatus, ptErrType, ptErrMsg, matched := applyErrorPassthroughRule(
		c, account.Platform, upstreamStatus, body,
		0, "", "",
	); matched {
		c.JSON(ptStatus, gin.H{
			"type":  "error",
			"error": gin.H{"type": ptErrType, "message": ptErrMsg},
		})
		if upstreamMsg == "" {
			return fmt.Errorf("upstream error: %d", upstreamStatus)
		}
		return fmt.Errorf("upstream error: %d message=%s", upstreamStatus, upstreamMsg)
	}

	var statusCode int
	var errType, errMsg string

	switch upstreamStatus {
	case 400:
		statusCode = http.StatusBadRequest
		errType = "invalid_request_error"
		errMsg = getPassthroughOrDefault(upstreamMsg, "Invalid request")
	case 401:
		statusCode = http.StatusBadGateway
		errType = "authentication_error"
		errMsg = "Upstream authentication failed"
	case 403:
		statusCode = http.StatusBadGateway
		errType = "permission_error"
		errMsg = "Upstream access forbidden"
	case 429:
		statusCode = http.StatusTooManyRequests
		errType = "rate_limit_error"
		errMsg = "Upstream rate limit exceeded"
	case 529:
		statusCode = http.StatusServiceUnavailable
		errType = "overloaded_error"
		errMsg = "Upstream service overloaded"
	default:
		statusCode = http.StatusBadGateway
		errType = "upstream_error"
		errMsg = "Upstream request failed"
	}

	c.JSON(statusCode, gin.H{
		"type":  "error",
		"error": gin.H{"type": errType, "message": errMsg},
	})
	if upstreamMsg == "" {
		return fmt.Errorf("upstream error: %d", upstreamStatus)
	}
	return fmt.Errorf("upstream error: %d message=%s", upstreamStatus, upstreamMsg)
}

func (s *AntigravityGatewayService) writeGoogleError(c *gin.Context, status int, message string) error {
	MarkResponseCommitted(c)
	statusStr := "UNKNOWN"
	switch status {
	case 400:
		statusStr = "INVALID_ARGUMENT"
	case 404:
		statusStr = "NOT_FOUND"
	case 429:
		statusStr = "RESOURCE_EXHAUSTED"
	case 500:
		statusStr = "INTERNAL"
	case 502, 503:
		statusStr = "UNAVAILABLE"
	}

	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    status,
			"message": message,
			"status":  statusStr,
		},
	})
	return fmt.Errorf("%s", message)
}
