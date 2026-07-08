package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Aias00/cloudbase/internal/pkg/antigravity"
	"github.com/Aias00/cloudbase/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// wrapV1InternalRequest 包装请求为 v1internal 格式
func (s *AntigravityGatewayService) wrapV1InternalRequest(projectID, model string, originalBody []byte) ([]byte, error) {
	var request any
	if err := json.Unmarshal(originalBody, &request); err != nil {
		return nil, fmt.Errorf("解析请求体失败: %w", err)
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, errAntigravityProjectIDRequired
	}

	wrapped := map[string]any{
		"project":     projectID,
		"requestId":   "agent-" + uuid.New().String(),
		"userAgent":   "antigravity", // 固定值，与官方客户端一致
		"requestType": "agent",
		"model":       model,
		"request":     request,
	}

	return json.Marshal(wrapped)
}

// unwrapV1InternalResponse 解包 v1internal 响应
// 使用 gjson 零拷贝提取 response 字段，避免 Unmarshal+Marshal 双重开销
func (s *AntigravityGatewayService) unwrapV1InternalResponse(body []byte) ([]byte, error) {
	result := gjson.GetBytes(body, "response")
	if result.Exists() {
		return []byte(result.Raw), nil
	}
	return body, nil
}

// Forward 转发 Claude 协议请求（Claude → Gemini 转换）
//
// 限流处理流程:
//
//	请求 → antigravityRetryLoop → 预检查(remaining>0? → 切换账号) → 发送上游
//	  ├─ 成功 → 正常返回
//	  └─ 429/503 → handleSmartRetry
//	      ├─ retryDelay >= 7s → 设置模型限流 + 清除粘性绑定 → 切换账号
//	      └─ retryDelay <  7s → 等待后重试 1 次
//	          ├─ 成功 → 正常返回
//	          └─ 失败 → 设置模型限流 + 清除粘性绑定 → 切换账号
func (s *AntigravityGatewayService) Forward(ctx context.Context, c *gin.Context, account *Account, body []byte, isStickySession bool) (*ForwardResult, error) {
	// 上游透传账号直接转发，不走 OAuth token 刷新
	if account.Type == AccountTypeUpstream {
		return s.ForwardUpstream(ctx, c, account, body)
	}

	startTime := time.Now()

	sessionID := getSessionID(c)
	prefix := logPrefix(sessionID, account.Name)

	// 解析 Claude 请求
	var claudeReq antigravity.ClaudeRequest
	if err := json.Unmarshal(body, &claudeReq); err != nil {
		return nil, s.writeClaudeError(c, http.StatusBadRequest, "invalid_request_error", "Invalid request body")
	}
	if strings.TrimSpace(claudeReq.Model) == "" {
		return nil, s.writeClaudeError(c, http.StatusBadRequest, "invalid_request_error", "Missing model")
	}

	originalModel := claudeReq.Model
	mappedModel := s.getMappedModel(account, claudeReq.Model)
	if mappedModel == "" {
		MarkOpsClientBusinessLimited(c, OpsClientBusinessLimitedReasonLocalFeatureGate)
		return nil, s.writeClaudeError(c, http.StatusForbidden, "permission_error", fmt.Sprintf("model %s not in whitelist", claudeReq.Model))
	}
	// 应用 thinking 模式自动后缀：如果 thinking 开启且目标是 claude-sonnet-4-5，自动改为 thinking 版本
	thinkingEnabled := claudeReq.Thinking != nil && (claudeReq.Thinking.Type == "enabled" || claudeReq.Thinking.Type == "adaptive")
	mappedModel = applyThinkingModelSuffix(mappedModel, thinkingEnabled)
	billingModel := mappedModel

	// 获取 access_token
	if s.tokenProvider == nil {
		return nil, s.writeClaudeError(c, http.StatusBadGateway, "api_error", "Antigravity token provider not configured")
	}
	accessToken, err := s.tokenProvider.GetAccessToken(ctx, account)
	if err != nil {
		return nil, &UpstreamFailoverError{
			StatusCode:   http.StatusBadGateway,
			ResponseBody: []byte(`{"error":{"type":"authentication_error","message":"Failed to get upstream access token"},"type":"error"}`),
		}
	}

	projectID, err := resolveAntigravityProjectID(account)
	if err != nil {
		_ = s.writeClaudeError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return nil, err
	}

	// 代理 URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	// 获取转换选项
	// Antigravity 上游要求必须包含身份提示词，否则会返回 429
	transformOpts := s.getClaudeTransformOptions(ctx)
	transformOpts.EnableIdentityPatch = true // 强制启用，Antigravity 上游必需

	// 转换 Claude 请求为 Gemini 格式
	geminiBody, err := antigravity.TransformClaudeToGeminiWithOptions(&claudeReq, projectID, mappedModel, transformOpts)
	if err != nil {
		return nil, s.writeClaudeError(c, http.StatusBadRequest, "invalid_request_error", "Invalid request")
	}

	// Antigravity 上游只支持流式请求，统一使用 streamGenerateContent
	// 如果客户端请求非流式，在响应处理阶段会收集完整流式响应后转换返回
	action := "streamGenerateContent"

	// 执行带重试的请求
	result, err := s.antigravityRetryLoop(antigravityRetryLoopParams{
		ctx:             ctx,
		prefix:          prefix,
		account:         account,
		proxyURL:        proxyURL,
		accessToken:     accessToken,
		action:          action,
		body:            geminiBody,
		c:               c,
		httpUpstream:    s.httpUpstream,
		settingService:  s.settingService,
		accountRepo:     s.accountRepo,
		handleError:     s.handleUpstreamError,
		requestedModel:  originalModel,
		isStickySession: isStickySession, // Forward 由上层判断粘性会话
		groupID:         0,               // Forward 方法没有 groupID，由上层处理粘性会话清除
		sessionHash:     "",              // Forward 方法没有 sessionHash，由上层处理粘性会话清除
	})
	if err != nil {
		// 检查是否是账号切换信号，转换为 UpstreamFailoverError 让 Handler 切换账号
		if switchErr, ok := IsAntigravityAccountSwitchError(err); ok {
			return nil, &UpstreamFailoverError{
				StatusCode:        http.StatusServiceUnavailable,
				ForceCacheBilling: switchErr.IsStickySession,
			}
		}
		// 区分客户端取消和真正的上游失败，返回更准确的错误消息
		if c.Request.Context().Err() != nil {
			return nil, s.writeClaudeError(c, http.StatusBadGateway, "client_disconnected", "Client disconnected before upstream response")
		}
		return nil, s.writeClaudeError(c, http.StatusBadGateway, "upstream_error", "Upstream request failed after retries")
	}
	resp := result.resp
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody := s.readUpstreamErrorBody(resp)

		// 优先检测 thinking block 的 signature 相关错误（400）并重试一次：
		// Antigravity /v1internal 链路在部分场景会对 thought/thinking signature 做严格校验，
		// 当历史消息携带的 signature 不合法时会直接 400；去除 thinking 后可继续完成请求。
		if resp.StatusCode == http.StatusBadRequest && isSignatureRelatedError(respBody) && s.settingService.IsSignatureRectifierEnabled(ctx) {
			upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
			upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
			logBody, maxBytes := s.getLogConfig()
			upstreamDetail := s.getUpstreamErrorDetail(respBody)
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				Kind:               "signature_error",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})

			// Conservative two-stage fallback:
			// 1) Disable top-level thinking + thinking->text
			// 2) Only if still signature-related 400: also downgrade tool_use/tool_result to text.

			retryStages := []struct {
				name  string
				strip func(*antigravity.ClaudeRequest) (bool, error)
			}{
				{name: "thinking-only", strip: stripThinkingFromClaudeRequest},
				{name: "thinking+tools", strip: stripSignatureSensitiveBlocksFromClaudeRequest},
			}

			for _, stage := range retryStages {
				retryClaudeReq := claudeReq
				retryClaudeReq.Messages = append([]antigravity.ClaudeMessage(nil), claudeReq.Messages...)

				stripped, stripErr := stage.strip(&retryClaudeReq)
				if stripErr != nil || !stripped {
					continue
				}

				logger.LegacyPrintf("service.antigravity_gateway", "Antigravity account %d: detected signature-related 400, retrying once (%s)", account.ID, stage.name)

				retryGeminiBody, txErr := antigravity.TransformClaudeToGeminiWithOptions(&retryClaudeReq, projectID, mappedModel, s.getClaudeTransformOptions(ctx))
				if txErr != nil {
					continue
				}
				retryResult, retryErr := s.antigravityRetryLoop(antigravityRetryLoopParams{
					ctx:             ctx,
					prefix:          prefix,
					account:         account,
					proxyURL:        proxyURL,
					accessToken:     accessToken,
					action:          action,
					body:            retryGeminiBody,
					c:               c,
					httpUpstream:    s.httpUpstream,
					settingService:  s.settingService,
					accountRepo:     s.accountRepo,
					handleError:     s.handleUpstreamError,
					requestedModel:  originalModel,
					isStickySession: isStickySession,
					groupID:         0,  // Forward 方法没有 groupID，由上层处理粘性会话清除
					sessionHash:     "", // Forward 方法没有 sessionHash，由上层处理粘性会话清除
				})
				if retryErr != nil {
					appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
						Platform:           account.Platform,
						AccountID:          account.ID,
						AccountName:        account.Name,
						UpstreamStatusCode: 0,
						Kind:               "signature_retry_request_error",
						Message:            sanitizeUpstreamErrorMessage(retryErr.Error()),
					})
					logger.LegacyPrintf("service.antigravity_gateway", "Antigravity account %d: signature retry request failed (%s): %v", account.ID, stage.name, retryErr)
					continue
				}

				retryResp := retryResult.resp
				if retryResp.StatusCode < 400 {
					_ = resp.Body.Close()
					resp = retryResp
					respBody = nil
					break
				}

				retryBody, _ := io.ReadAll(io.LimitReader(retryResp.Body, 8<<10))
				_ = retryResp.Body.Close()
				if retryResp.StatusCode == http.StatusTooManyRequests {
					retryBaseURL := ""
					if retryResp.Request != nil && retryResp.Request.URL != nil {
						retryBaseURL = retryResp.Request.URL.Scheme + "://" + retryResp.Request.URL.Host
					}
					logger.LegacyPrintf("service.antigravity_gateway", "%s status=429 rate_limited base_url=%s retry_stage=%s body=%s", prefix, retryBaseURL, stage.name, truncateForLog(retryBody, 200))
				}
				kind := "signature_retry"
				if strings.TrimSpace(stage.name) != "" {
					kind = "signature_retry_" + strings.ReplaceAll(stage.name, "+", "_")
				}
				retryUpstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(retryBody))
				retryUpstreamMsg = sanitizeUpstreamErrorMessage(retryUpstreamMsg)
				retryUpstreamDetail := ""
				if logBody {
					retryUpstreamDetail = truncateString(string(retryBody), maxBytes)
				}
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: retryResp.StatusCode,
					UpstreamRequestID:  retryResp.Header.Get("x-request-id"),
					Kind:               kind,
					Message:            retryUpstreamMsg,
					Detail:             retryUpstreamDetail,
				})

				// If this stage fixed the signature issue, we stop; otherwise we may try the next stage.
				if retryResp.StatusCode != http.StatusBadRequest || !isSignatureRelatedError(retryBody) {
					respBody = retryBody
					resp = &http.Response{
						StatusCode: retryResp.StatusCode,
						Header:     retryResp.Header.Clone(),
						Body:       io.NopCloser(bytes.NewReader(retryBody)),
					}
					break
				}

				// Still signature-related; capture context and allow next stage.
				respBody = retryBody
				resp = &http.Response{
					StatusCode: retryResp.StatusCode,
					Header:     retryResp.Header.Clone(),
					Body:       io.NopCloser(bytes.NewReader(retryBody)),
				}
			}
		}

		// Budget 整流：检测 budget_tokens 约束错误并自动修正重试
		if resp.StatusCode == http.StatusBadRequest && respBody != nil && !isSignatureRelatedError(respBody) {
			errMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
			if isThinkingBudgetConstraintError(errMsg) && s.settingService.IsBudgetRectifierEnabled(ctx) {
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: resp.StatusCode,
					UpstreamRequestID:  resp.Header.Get("x-request-id"),
					Kind:               "budget_constraint_error",
					Message:            errMsg,
					Detail:             s.getUpstreamErrorDetail(respBody),
				})

				// 修正 claudeReq 的 thinking 参数（adaptive 模式不修正）
				if claudeReq.Thinking == nil || claudeReq.Thinking.Type != "adaptive" {
					retryClaudeReq := claudeReq
					retryClaudeReq.Messages = append([]antigravity.ClaudeMessage(nil), claudeReq.Messages...)
					// 创建新的 ThinkingConfig 避免修改原始 claudeReq.Thinking 指针
					retryClaudeReq.Thinking = &antigravity.ThinkingConfig{
						Type:         "enabled",
						BudgetTokens: BudgetRectifyBudgetTokens,
					}
					if retryClaudeReq.MaxTokens < BudgetRectifyMinMaxTokens {
						retryClaudeReq.MaxTokens = BudgetRectifyMaxTokens
					}

					logger.LegacyPrintf("service.antigravity_gateway", "Antigravity account %d: detected budget_tokens constraint error, retrying with rectified budget (budget_tokens=%d, max_tokens=%d)", account.ID, BudgetRectifyBudgetTokens, BudgetRectifyMaxTokens)

					retryGeminiBody, txErr := antigravity.TransformClaudeToGeminiWithOptions(&retryClaudeReq, projectID, mappedModel, transformOpts)
					if txErr == nil {
						retryResult, retryErr := s.antigravityRetryLoop(antigravityRetryLoopParams{
							ctx:             ctx,
							prefix:          prefix,
							account:         account,
							proxyURL:        proxyURL,
							accessToken:     accessToken,
							action:          action,
							body:            retryGeminiBody,
							c:               c,
							httpUpstream:    s.httpUpstream,
							settingService:  s.settingService,
							accountRepo:     s.accountRepo,
							handleError:     s.handleUpstreamError,
							requestedModel:  originalModel,
							isStickySession: isStickySession,
							groupID:         0,
							sessionHash:     "",
						})
						if retryErr == nil {
							retryResp := retryResult.resp
							if retryResp.StatusCode < 400 {
								_ = resp.Body.Close()
								resp = retryResp
								respBody = nil
							} else {
								retryBody := s.readUpstreamErrorBody(retryResp)
								_ = retryResp.Body.Close()
								respBody = retryBody
								resp = &http.Response{
									StatusCode: retryResp.StatusCode,
									Header:     retryResp.Header.Clone(),
									Body:       io.NopCloser(bytes.NewReader(retryBody)),
								}
							}
						} else {
							logger.LegacyPrintf("service.antigravity_gateway", "Antigravity account %d: budget rectifier retry failed: %v", account.ID, retryErr)
						}
					}
				}
			}
		}

		// 处理错误响应（重试后仍失败或不触发重试）
		if resp.StatusCode >= 400 {
			// 检测 prompt too long 错误，返回特殊错误类型供上层 fallback
			if resp.StatusCode == http.StatusBadRequest && isPromptTooLongError(respBody) {
				upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
				upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
				upstreamDetail := s.getUpstreamErrorDetail(respBody)
				logBody, maxBytes := s.getLogConfig()
				if logBody {
					logger.LegacyPrintf("service.antigravity_gateway", "%s status=400 prompt_too_long=true upstream_message=%q request_id=%s body=%s", prefix, upstreamMsg, resp.Header.Get("x-request-id"), truncateForLog(respBody, maxBytes))
				}
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: resp.StatusCode,
					UpstreamRequestID:  resp.Header.Get("x-request-id"),
					Kind:               "prompt_too_long",
					Message:            upstreamMsg,
					Detail:             upstreamDetail,
				})
				return nil, &PromptTooLongError{
					StatusCode: resp.StatusCode,
					RequestID:  resp.Header.Get("x-request-id"),
					Body:       respBody,
				}
			}

			s.handleUpstreamError(ctx, prefix, account, resp.StatusCode, resp.Header, respBody, originalModel, 0, "", isStickySession)

			// 精确匹配服务端配置类 400 错误，触发同账号重试 + failover
			if resp.StatusCode == http.StatusBadRequest {
				msg := strings.ToLower(strings.TrimSpace(extractAntigravityErrorMessage(respBody)))
				if isGoogleProjectConfigError(msg) {
					upstreamMsg := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractAntigravityErrorMessage(respBody)))
					upstreamDetail := s.getUpstreamErrorDetail(respBody)
					log.Printf("%s status=400 google_config_error failover=true upstream_message=%q account=%d", prefix, upstreamMsg, account.ID)
					appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
						Platform:           account.Platform,
						AccountID:          account.ID,
						AccountName:        account.Name,
						UpstreamStatusCode: resp.StatusCode,
						UpstreamRequestID:  resp.Header.Get("x-request-id"),
						Kind:               "failover",
						Message:            upstreamMsg,
						Detail:             upstreamDetail,
					})
					return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: respBody, RetryableOnSameAccount: true}
				}
			}

			if s.shouldFailoverUpstreamError(resp.StatusCode) {
				upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(respBody))
				upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
				upstreamDetail := s.getUpstreamErrorDetail(respBody)
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform:           account.Platform,
					AccountID:          account.ID,
					AccountName:        account.Name,
					UpstreamStatusCode: resp.StatusCode,
					UpstreamRequestID:  resp.Header.Get("x-request-id"),
					Kind:               "failover",
					Message:            upstreamMsg,
					Detail:             upstreamDetail,
				})
				return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: respBody}
			}

			return nil, s.writeMappedClaudeError(c, account, resp.StatusCode, resp.Header.Get("x-request-id"), respBody)
		}
	}

	requestID := resp.Header.Get("x-request-id")
	if requestID != "" {
		c.Header("x-request-id", requestID)
	}

	var usage *ClaudeUsage
	var firstTokenMs *int
	var clientDisconnect bool
	if claudeReq.Stream {
		// 客户端要求流式，直接透传转换
		streamRes, err := s.handleClaudeStreamingResponse(c, resp, startTime, originalModel)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=stream_error error=%v", prefix, err)
			return nil, err
		}
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
		clientDisconnect = streamRes.clientDisconnect
	} else {
		// 客户端要求非流式，收集流式响应后转换返回
		streamRes, err := s.handleClaudeStreamToNonStreaming(c, resp, startTime, originalModel)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=stream_collect_error error=%v", prefix, err)
			return nil, err
		}
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
	}

	return &ForwardResult{
		RequestID:        requestID,
		Usage:            *usage,
		Model:            originalModel,
		UpstreamModel:    billingModel,
		Stream:           claudeReq.Stream,
		Duration:         time.Since(startTime),
		FirstTokenMs:     firstTokenMs,
		ClientDisconnect: clientDisconnect,
	}, nil
}

func WithForwardGeminiSession(groupID int64, sessionHash string) ForwardGeminiOption {
	return func(opts *forwardGeminiOptions) {
		opts.groupID = groupID
		opts.sessionHash = sessionHash
	}
}

func (s *AntigravityGatewayService) ForwardGemini(ctx context.Context, c *gin.Context, account *Account, originalModel string, action string, stream bool, body []byte, isStickySession bool, options ...ForwardGeminiOption) (*ForwardResult, error) {
	startTime := time.Now()
	forwardOpts := forwardGeminiOptions{}
	for _, apply := range options {
		if apply != nil {
			apply(&forwardOpts)
		}
	}

	sessionID := getSessionID(c)
	prefix := logPrefix(sessionID, account.Name)

	if strings.TrimSpace(originalModel) == "" {
		return nil, s.writeGoogleError(c, http.StatusBadRequest, "Missing model in URL")
	}
	if strings.TrimSpace(action) == "" {
		return nil, s.writeGoogleError(c, http.StatusBadRequest, "Missing action in URL")
	}
	if len(body) == 0 {
		return nil, s.writeGoogleError(c, http.StatusBadRequest, "Request body is empty")
	}

	// 解析请求以获取 image_size（用于图片计费）
	imageInputSize := s.extractImageInputSize(body)
	imageSize := normalizeOpenAIImageSizeTier(imageInputSize)

	switch action {
	case "generateContent", "streamGenerateContent":
		// ok
	case "countTokens":
		// 直接返回空值，不透传上游
		c.JSON(http.StatusOK, map[string]any{"totalTokens": 0})
		return &ForwardResult{
			RequestID:    "",
			Usage:        ClaudeUsage{},
			Model:        originalModel,
			Stream:       false,
			Duration:     time.Since(startTime),
			FirstTokenMs: nil,
		}, nil
	default:
		return nil, s.writeGoogleError(c, http.StatusNotFound, "Unsupported action: "+action)
	}

	mappedModel := s.getMappedModel(account, originalModel)
	if mappedModel == "" {
		MarkOpsClientBusinessLimited(c, OpsClientBusinessLimitedReasonLocalFeatureGate)
		return nil, s.writeGoogleError(c, http.StatusForbidden, fmt.Sprintf("model %s not in whitelist", originalModel))
	}
	billingModel := mappedModel

	// 获取 access_token
	if s.tokenProvider == nil {
		return nil, s.writeGoogleError(c, http.StatusBadGateway, "Antigravity token provider not configured")
	}
	accessToken, err := s.tokenProvider.GetAccessToken(ctx, account)
	if err != nil {
		return nil, &UpstreamFailoverError{
			StatusCode:   http.StatusBadGateway,
			ResponseBody: []byte(`{"error":{"message":"Failed to get upstream access token","status":"UNAVAILABLE"}}`),
		}
	}

	projectID, err := resolveAntigravityProjectID(account)
	if err != nil {
		_ = s.writeGoogleError(c, http.StatusBadRequest, err.Error())
		return nil, err
	}

	// 代理 URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	// Antigravity 上游要求必须包含身份提示词，注入到请求中
	injectedBody, err := injectIdentityPatchToGeminiRequest(body)
	if err != nil {
		return nil, s.writeGoogleError(c, http.StatusBadRequest, "Invalid request body")
	}

	// 清理 Schema
	if cleanedBody, err := cleanGeminiRequest(injectedBody); err == nil {
		injectedBody = cleanedBody
		logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] Cleaned request schema in forwarded request for account %s", account.Name)
	} else {
		logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] Failed to clean schema: %v", err)
	}

	// 包装请求
	wrappedBody, err := s.wrapV1InternalRequest(projectID, mappedModel, injectedBody)
	if err != nil {
		return nil, s.writeGoogleError(c, http.StatusInternalServerError, "Failed to build upstream request")
	}

	// Antigravity 上游只支持流式请求，统一使用 streamGenerateContent
	// 如果客户端请求非流式，在响应处理阶段会收集完整流式响应后返回
	upstreamAction := "streamGenerateContent"

	// 执行带重试的请求
	result, err := s.antigravityRetryLoop(antigravityRetryLoopParams{
		ctx:             ctx,
		prefix:          prefix,
		account:         account,
		proxyURL:        proxyURL,
		accessToken:     accessToken,
		action:          upstreamAction,
		body:            wrappedBody,
		c:               c,
		httpUpstream:    s.httpUpstream,
		settingService:  s.settingService,
		accountRepo:     s.accountRepo,
		handleError:     s.handleUpstreamError,
		requestedModel:  originalModel,
		isStickySession: isStickySession, // ForwardGemini 由上层判断粘性会话
		groupID:         forwardOpts.groupID,
		sessionHash:     forwardOpts.sessionHash,
	})
	if err != nil {
		// 检查是否是账号切换信号，转换为 UpstreamFailoverError 让 Handler 切换账号
		if switchErr, ok := IsAntigravityAccountSwitchError(err); ok {
			return nil, &UpstreamFailoverError{
				StatusCode:        http.StatusServiceUnavailable,
				ForceCacheBilling: switchErr.IsStickySession,
			}
		}
		// 区分客户端取消和真正的上游失败，返回更准确的错误消息
		if c.Request.Context().Err() != nil {
			return nil, s.writeGoogleError(c, http.StatusBadGateway, "Client disconnected before upstream response")
		}
		return nil, s.writeGoogleError(c, http.StatusBadGateway, "Upstream request failed after retries")
	}
	resp := result.resp
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	// 处理错误响应
	if resp.StatusCode >= 400 {
		respBody := s.readUpstreamErrorBody(resp)
		contentType := resp.Header.Get("Content-Type")
		// 尽早关闭原始响应体，释放连接；后续逻辑仍可能需要读取 body，因此用内存副本重新包装。
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		// 模型兜底：模型不存在且开启 fallback 时，自动用 fallback 模型重试一次
		if s.settingService != nil && s.settingService.IsModelFallbackEnabled(ctx) &&
			isModelNotFoundError(resp.StatusCode, respBody) {
			fallbackModel := s.settingService.GetFallbackModel(ctx, PlatformAntigravity)
			if fallbackModel != "" && fallbackModel != mappedModel {
				logger.LegacyPrintf("service.antigravity_gateway", "[Antigravity] Model not found (%s), retrying with fallback model %s (account: %s)", mappedModel, fallbackModel, account.Name)

				fallbackWrapped, err := s.wrapV1InternalRequest(projectID, fallbackModel, injectedBody)
				if err == nil {
					fallbackReq, err := antigravity.NewAPIRequest(ctx, upstreamAction, accessToken, fallbackWrapped)
					if err == nil {
						fallbackResp, err := s.httpUpstream.Do(fallbackReq, proxyURL, account.ID, account.Concurrency)
						if err == nil && fallbackResp.StatusCode < 400 {
							_ = resp.Body.Close()
							resp = fallbackResp
						} else if fallbackResp != nil {
							_ = fallbackResp.Body.Close()
						}
					}
				}
			}
		}

		// Gemini 原生请求中的 thoughtSignature 可能来自旧上下文/旧账号，触发上游严格校验后返回
		// "Corrupted thought signature."。检测到此类 400 时，将 thoughtSignature 清理为 dummy 值后重试一次。
		signatureCheckBody := respBody
		if unwrapped, unwrapErr := s.unwrapV1InternalResponse(respBody); unwrapErr == nil && len(unwrapped) > 0 {
			signatureCheckBody = unwrapped
		}
		if resp.StatusCode == http.StatusBadRequest &&
			s.settingService != nil &&
			s.settingService.IsSignatureRectifierEnabled(ctx) &&
			isSignatureRelatedError(signatureCheckBody) &&
			bytes.Contains(injectedBody, []byte(`"thoughtSignature"`)) {
			upstreamMsg := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractAntigravityErrorMessage(signatureCheckBody)))
			upstreamDetail := s.getUpstreamErrorDetail(signatureCheckBody)
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				Kind:               "signature_error",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})

			logger.LegacyPrintf("service.antigravity_gateway", "Antigravity Gemini account %d: detected signature-related 400, retrying with cleaned thought signatures", account.ID)

			cleanedInjectedBody := CleanGeminiNativeThoughtSignatures(injectedBody)
			retryWrappedBody, wrapErr := s.wrapV1InternalRequest(projectID, mappedModel, cleanedInjectedBody)
			if wrapErr == nil {
				retryResult, retryErr := s.antigravityRetryLoop(antigravityRetryLoopParams{
					ctx:             ctx,
					prefix:          prefix,
					account:         account,
					proxyURL:        proxyURL,
					accessToken:     accessToken,
					action:          upstreamAction,
					body:            retryWrappedBody,
					c:               c,
					httpUpstream:    s.httpUpstream,
					settingService:  s.settingService,
					accountRepo:     s.accountRepo,
					handleError:     s.handleUpstreamError,
					requestedModel:  originalModel,
					isStickySession: isStickySession,
					groupID:         forwardOpts.groupID,
					sessionHash:     forwardOpts.sessionHash,
				})
				if retryErr == nil {
					retryResp := retryResult.resp
					if retryResp.StatusCode < 400 {
						resp = retryResp
					} else {
						retryRespBody := s.readUpstreamErrorBody(retryResp)
						_ = retryResp.Body.Close()
						retryOpsBody := retryRespBody
						if retryUnwrapped, unwrapErr := s.unwrapV1InternalResponse(retryRespBody); unwrapErr == nil && len(retryUnwrapped) > 0 {
							retryOpsBody = retryUnwrapped
						}
						appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
							Platform:           account.Platform,
							AccountID:          account.ID,
							AccountName:        account.Name,
							UpstreamStatusCode: retryResp.StatusCode,
							UpstreamRequestID:  retryResp.Header.Get("x-request-id"),
							Kind:               "signature_retry",
							Message:            sanitizeUpstreamErrorMessage(strings.TrimSpace(extractAntigravityErrorMessage(retryOpsBody))),
							Detail:             s.getUpstreamErrorDetail(retryOpsBody),
						})
						respBody = retryRespBody
						resp = &http.Response{
							StatusCode: retryResp.StatusCode,
							Header:     retryResp.Header.Clone(),
							Body:       io.NopCloser(bytes.NewReader(retryRespBody)),
						}
						contentType = resp.Header.Get("Content-Type")
					}
				} else {
					if switchErr, ok := IsAntigravityAccountSwitchError(retryErr); ok {
						appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
							Platform:           account.Platform,
							AccountID:          account.ID,
							AccountName:        account.Name,
							UpstreamStatusCode: http.StatusServiceUnavailable,
							Kind:               "failover",
							Message:            sanitizeUpstreamErrorMessage(retryErr.Error()),
						})
						return nil, &UpstreamFailoverError{
							StatusCode:        http.StatusServiceUnavailable,
							ForceCacheBilling: switchErr.IsStickySession,
						}
					}
					appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
						Platform:           account.Platform,
						AccountID:          account.ID,
						AccountName:        account.Name,
						UpstreamStatusCode: 0,
						Kind:               "signature_retry_request_error",
						Message:            sanitizeUpstreamErrorMessage(retryErr.Error()),
					})
					logger.LegacyPrintf("service.antigravity_gateway", "Antigravity Gemini account %d: signature retry request failed: %v", account.ID, retryErr)
				}
			} else {
				logger.LegacyPrintf("service.antigravity_gateway", "Antigravity Gemini account %d: signature retry wrap failed: %v", account.ID, wrapErr)
			}
		}

		// fallback 成功：继续按正常响应处理
		if resp.StatusCode < 400 {
			goto handleSuccess
		}

		requestID := resp.Header.Get("x-request-id")
		if requestID != "" {
			c.Header("x-request-id", requestID)
		}

		unwrapped, unwrapErr := s.unwrapV1InternalResponse(respBody)
		unwrappedForOps := unwrapped
		if unwrapErr != nil || len(unwrappedForOps) == 0 {
			unwrappedForOps = respBody
		}
		s.handleUpstreamError(ctx, prefix, account, resp.StatusCode, resp.Header, respBody, originalModel, forwardOpts.groupID, forwardOpts.sessionHash, isStickySession)
		upstreamMsg := strings.TrimSpace(extractAntigravityErrorMessage(unwrappedForOps))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
		upstreamDetail := s.getUpstreamErrorDetail(unwrappedForOps)

		// Always record upstream context for Ops error logs, even when we will failover.
		setOpsUpstreamError(c, resp.StatusCode, upstreamMsg, upstreamDetail)

		// 精确匹配服务端配置类 400 错误，触发同账号重试 + failover
		if resp.StatusCode == http.StatusBadRequest && isGoogleProjectConfigError(strings.ToLower(upstreamMsg)) {
			log.Printf("%s status=400 google_config_error failover=true upstream_message=%q account=%d", prefix, upstreamMsg, account.ID)
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  requestID,
				Kind:               "failover",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})
			return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: unwrappedForOps, RetryableOnSameAccount: true}
		}

		if s.shouldFailoverUpstreamError(resp.StatusCode) {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  requestID,
				Kind:               "failover",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})
			return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode, ResponseBody: unwrappedForOps}
		}
		if contentType == "" {
			contentType = "application/json"
		}
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: resp.StatusCode,
			UpstreamRequestID:  requestID,
			Kind:               "http_error",
			Message:            upstreamMsg,
			Detail:             upstreamDetail,
		})
		logger.LegacyPrintf("service.antigravity_gateway", "[antigravity-Forward] upstream error status=%d body=%s", resp.StatusCode, truncateForLog(unwrappedForOps, 500))
		MarkResponseCommitted(c)
		c.Data(resp.StatusCode, contentType, unwrappedForOps)
		return nil, fmt.Errorf("antigravity upstream error: %d", resp.StatusCode)
	}

handleSuccess:
	requestID := resp.Header.Get("x-request-id")
	if requestID != "" {
		c.Header("x-request-id", requestID)
	}

	var usage *ClaudeUsage
	var firstTokenMs *int
	var clientDisconnect bool

	if stream {
		// 客户端要求流式，直接透传
		streamRes, err := s.handleGeminiStreamingResponse(c, resp, startTime)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=stream_error error=%v", prefix, err)
			return nil, err
		}
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
		clientDisconnect = streamRes.clientDisconnect
	} else {
		// 客户端要求非流式，收集流式响应后返回
		streamRes, err := s.handleGeminiStreamToNonStreaming(c, resp, startTime)
		if err != nil {
			logger.LegacyPrintf("service.antigravity_gateway", "%s status=stream_collect_error error=%v", prefix, err)
			return nil, err
		}
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
	}

	if usage == nil {
		usage = &ClaudeUsage{}
	}

	// 判断是否为图片生成模型
	imageCount := 0
	if isImageGenerationModel(mappedModel) {
		// Gemini 图片生成 API 每次请求只生成一张图片（API 限制）
		imageCount = 1
	}

	return &ForwardResult{
		RequestID:        requestID,
		Usage:            *usage,
		Model:            originalModel,
		UpstreamModel:    billingModel,
		Stream:           stream,
		Duration:         time.Since(startTime),
		FirstTokenMs:     firstTokenMs,
		ClientDisconnect: clientDisconnect,
		ImageCount:       imageCount,
		ImageSize:        imageSize,
		ImageInputSize:   imageInputSize,
	}, nil
}

// ForwardUpstream 使用 base_url + /v1/messages + 双 header 认证透传上游 Claude 请求
func (s *AntigravityGatewayService) ForwardUpstream(ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
	startTime := time.Now()
	sessionID := getSessionID(c)
	prefix := logPrefix(sessionID, account.Name)

	// 获取上游配置
	baseURL := strings.TrimSpace(account.GetCredential("base_url"))
	apiKey := strings.TrimSpace(account.GetCredential("api_key"))
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("upstream account missing base_url or api_key")
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	// 解析请求获取模型信息
	var claudeReq antigravity.ClaudeRequest
	if err := json.Unmarshal(body, &claudeReq); err != nil {
		return nil, fmt.Errorf("parse claude request: %w", err)
	}
	if strings.TrimSpace(claudeReq.Model) == "" {
		return nil, fmt.Errorf("missing model")
	}
	originalModel := claudeReq.Model

	// 构建上游请求 URL
	upstreamURL := baseURL + "/v1/messages"

	// 能力维度 sanitize：Anthropic-compatible 上游透传路径也需要保证 body↔beta header
	// 对称。客户端 anthropic-beta header 不含 context-management-2025-06-27 但 body 带
	// context_management 时 strip，与 Anthropic 直连 / Bedrock / Vertex 路径保持一致。
	clientBeta := c.GetHeader("anthropic-beta")
	if sanitized, changed := sanitizeAnthropicBodyForBetaTokens(body, clientBeta); changed {
		body = sanitized
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create upstream request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("x-api-key", apiKey) // Claude API 兼容

	// 透传 Claude 相关 headers
	if v := c.GetHeader("anthropic-version"); v != "" {
		req.Header.Set("anthropic-version", v)
	}
	if v := clientBeta; v != "" {
		req.Header.Set("anthropic-beta", v)
	}

	// 代理 URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	// 发送请求
	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		logger.LegacyPrintf("service.antigravity_gateway", "%s upstream request failed: %v", prefix, err)
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 处理错误响应
	if resp.StatusCode >= 400 {
		respBody := s.readUpstreamErrorBody(resp)

		// 429 错误时标记账号限流
		if resp.StatusCode == http.StatusTooManyRequests {
			s.handleUpstreamError(ctx, prefix, account, resp.StatusCode, resp.Header, respBody, originalModel, 0, "", false)
		}

		// 透传上游错误
		c.Header("Content-Type", resp.Header.Get("Content-Type"))
		c.Status(resp.StatusCode)
		_, _ = c.Writer.Write(respBody)

		return &ForwardResult{
			Model: originalModel,
		}, nil
	}

	// 处理成功响应（流式/非流式）
	var usage *ClaudeUsage
	var firstTokenMs *int
	var clientDisconnect bool

	if claudeReq.Stream {
		// 流式响应：透传
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")
		c.Status(http.StatusOK)

		streamRes := s.streamUpstreamResponse(c, resp, startTime)
		usage = streamRes.usage
		firstTokenMs = streamRes.firstTokenMs
		clientDisconnect = streamRes.clientDisconnect
	} else {
		// 非流式响应：直接透传
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read upstream response: %w", err)
		}

		// 提取 usage
		usage = s.extractClaudeUsage(respBody)

		c.Header("Content-Type", resp.Header.Get("Content-Type"))
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(respBody)
	}

	// 构建计费结果
	duration := time.Since(startTime)
	logger.LegacyPrintf("service.antigravity_gateway", "%s status=success duration_ms=%d", prefix, duration.Milliseconds())

	return &ForwardResult{
		Model:            originalModel,
		Stream:           claudeReq.Stream,
		Duration:         duration,
		FirstTokenMs:     firstTokenMs,
		ClientDisconnect: clientDisconnect,
		Usage: ClaudeUsage{
			InputTokens:              usage.InputTokens,
			OutputTokens:             usage.OutputTokens,
			CacheReadInputTokens:     usage.CacheReadInputTokens,
			CacheCreationInputTokens: usage.CacheCreationInputTokens,
		},
	}, nil
}

// streamUpstreamResponse 透传上游 SSE 流并提取 Claude usage
func (s *AntigravityGatewayService) streamUpstreamResponse(c *gin.Context, resp *http.Response, startTime time.Time) *antigravityStreamResult {
	usage := &ClaudeUsage{}
	var firstTokenMs *int

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.settingService.cfg.Gateway.MaxLineSize
	}
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)

	type scanEvent struct {
		line string
		err  error
	}
	events := make(chan scanEvent, 16)
	done := make(chan struct{})
	sendEvent := func(ev scanEvent) bool {
		select {
		case events <- ev:
			return true
		case <-done:
			return false
		}
	}
	var lastReadAt int64
	atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
	go func() {
		defer close(events)
		for scanner.Scan() {
			atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
			if !sendEvent(scanEvent{line: scanner.Text()}) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			_ = sendEvent(scanEvent{err: err})
		}
	}()
	defer close(done)

	streamInterval := time.Duration(0)
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.StreamDataIntervalTimeout > 0 {
		streamInterval = time.Duration(s.settingService.cfg.Gateway.StreamDataIntervalTimeout) * time.Second
	}
	var intervalTicker *time.Ticker
	if streamInterval > 0 {
		intervalTicker = time.NewTicker(streamInterval)
		defer intervalTicker.Stop()
	}
	var intervalCh <-chan time.Time
	if intervalTicker != nil {
		intervalCh = intervalTicker.C
	}

	// 下游 keepalive：防止代理/Cloudflare Tunnel 因连接空闲而断开
	keepaliveInterval := time.Duration(0)
	if s.settingService.cfg != nil && s.settingService.cfg.Gateway.StreamKeepaliveInterval > 0 {
		keepaliveInterval = time.Duration(s.settingService.cfg.Gateway.StreamKeepaliveInterval) * time.Second
	}
	var keepaliveTicker *time.Ticker
	if keepaliveInterval > 0 {
		keepaliveTicker = time.NewTicker(keepaliveInterval)
		defer keepaliveTicker.Stop()
	}
	var keepaliveCh <-chan time.Time
	if keepaliveTicker != nil {
		keepaliveCh = keepaliveTicker.C
	}
	lastDataAt := time.Now()

	flusher, _ := c.Writer.(http.Flusher)
	cw := newAntigravityClientWriter(c.Writer, flusher, "antigravity upstream")

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: cw.Disconnected()}
			}
			if ev.err != nil {
				if disconnect, handled := handleStreamReadError(ev.err, cw.Disconnected(), "antigravity upstream"); handled {
					return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: disconnect}
				}
				logger.LegacyPrintf("service.antigravity_gateway", "Stream read error (antigravity upstream): %v", ev.err)
				return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs}
			}

			lastDataAt = time.Now()

			line := ev.line

			// 记录首 token 时间
			if firstTokenMs == nil && len(line) > 0 {
				ms := int(time.Since(startTime).Milliseconds())
				firstTokenMs = &ms
			}

			// 尝试从 message_delta 或 message_stop 事件提取 usage
			s.extractSSEUsage(line, usage)

			// 透传行
			cw.Fprintf("%s\n", line)

		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			if cw.Disconnected() {
				logger.LegacyPrintf("service.antigravity_gateway", "Upstream timeout after client disconnect (antigravity upstream), returning collected usage")
				return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: true}
			}
			logger.LegacyPrintf("service.antigravity_gateway", "Stream data interval timeout (antigravity upstream)")
			return &antigravityStreamResult{usage: usage, firstTokenMs: firstTokenMs}

		case <-keepaliveCh:
			if cw.Disconnected() {
				continue
			}
			if time.Since(lastDataAt) < keepaliveInterval {
				continue
			}
			// SSE ping 事件：Anthropic 原生格式，客户端会正确处理，
			// 同时保持连接活跃防止 Cloudflare Tunnel 等代理断开
			if !cw.Fprintf("event: ping\ndata: {\"type\": \"ping\"}\n\n") {
				logger.LegacyPrintf("service.antigravity_gateway", "Client disconnected during keepalive ping (antigravity upstream), continuing to drain upstream for billing")
				continue
			}
		}
	}
}
