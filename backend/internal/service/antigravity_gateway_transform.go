package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Aias00/cloudbase/internal/pkg/antigravity"
)

// mapAntigravityModel 获取映射后的模型名
// 完全依赖映射配置：账户映射（通配符）→ 默认映射兜底（DefaultAntigravityModelMapping）
// 注意：返回空字符串表示模型不被支持，调度时会过滤掉该账号
func mapAntigravityModel(account *Account, requestedModel string) string {
	if account == nil {
		return ""
	}
	requestedModel = strings.TrimPrefix(requestedModel, "models/")

	// 获取映射表（未配置时自动使用 DefaultAntigravityModelMapping）
	mapping := account.GetModelMapping()
	if len(mapping) == 0 {
		return "" // 无映射配置（非 Antigravity 平台）
	}

	// 通过映射表查询（支持精确匹配 + 通配符）
	mapped := account.GetMappedModel(requestedModel)

	// 判断是否映射成功（mapped != requestedModel 说明找到了映射规则）
	if mapped != requestedModel {
		return mapped
	}

	// 如果 mapped == requestedModel，检查是否在映射表中配置（精确或通配符）
	// 这区分两种情况：
	// 1. 映射表中有 "model-a": "model-a"（显式透传）→ 返回 model-a
	// 2. 通配符匹配 "claude-*": "claude-sonnet-4-5" 恰好目标等于请求名 → 返回 model-a
	// 3. 映射表中没有 model-a 的配置 → 返回空（不支持）
	if account.IsModelSupported(requestedModel) {
		return requestedModel
	}

	// 未在映射表中配置的模型，返回空字符串（不支持）
	return ""
}

// getMappedModel 获取映射后的模型名
// 完全依赖映射配置：账户映射（通配符）→ 默认映射兜底
func (s *AntigravityGatewayService) getMappedModel(account *Account, requestedModel string) string {
	return mapAntigravityModel(account, requestedModel)
}

func resolveAntigravityProjectID(account *Account) (string, error) {
	if account == nil {
		return "", errAntigravityProjectIDRequired
	}
	if projectID := strings.TrimSpace(account.GetCredential("project_id")); projectID != "" {
		return projectID, nil
	}
	if projectID := strings.TrimSpace(account.GetCredential(antigravityProjectIDFallbackCredentialKey)); projectID != "" {
		return projectID, nil
	}
	if projectID := strings.TrimSpace(account.GetExtraString(antigravityProjectIDFallbackCredentialKey)); projectID != "" {
		return projectID, nil
	}
	return "", errAntigravityProjectIDRequired
}

// applyThinkingModelSuffix 根据 thinking 配置调整模型名
// 当映射结果是 claude-sonnet-4-5 且请求开启了 thinking 时，改为 claude-sonnet-4-5-thinking
func applyThinkingModelSuffix(mappedModel string, thinkingEnabled bool) string {
	if !thinkingEnabled {
		return mappedModel
	}
	if mappedModel == "claude-sonnet-4-5" {
		return "claude-sonnet-4-5-thinking"
	}
	return mappedModel
}

// IsModelSupported 检查模型是否被支持
// 所有 claude- 和 gemini- 前缀的模型都能通过映射或透传支持
func (s *AntigravityGatewayService) IsModelSupported(requestedModel string) bool {
	return strings.HasPrefix(requestedModel, "claude-") ||
		strings.HasPrefix(requestedModel, "gemini-")
}

// TestConnection 测试 Antigravity 账号连接。
// 复用 antigravityRetryLoop 的完整重试 / credits overages / 智能重试逻辑，
// 与真实调度行为一致。差异：不做账号切换（测试指定账号）、不记录 ops 错误。
func (s *AntigravityGatewayService) TestConnection(ctx context.Context, account *Account, modelID string) (*TestConnectionResult, error) {

	// 获取 token
	if s.tokenProvider == nil {
		return nil, errors.New("antigravity token provider not configured")
	}
	accessToken, err := s.tokenProvider.GetAccessToken(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("获取 access_token 失败: %w", err)
	}

	projectID, err := resolveAntigravityProjectID(account)
	if err != nil {
		return nil, err
	}

	// 模型映射
	mappedModel := s.getMappedModel(account, modelID)
	if mappedModel == "" {
		return nil, fmt.Errorf("model %s not in whitelist", modelID)
	}

	// 构建请求体
	var requestBody []byte
	if strings.HasPrefix(modelID, "gemini-") {
		requestBody, err = s.buildGeminiTestRequest(projectID, mappedModel)
	} else {
		requestBody, err = s.buildClaudeTestRequest(projectID, mappedModel)
	}
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	// 代理 URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	// 复用 antigravityRetryLoop：完整的重试 / credits overages / 智能重试
	prefix := fmt.Sprintf("[antigravity-Test] account=%d(%s)", account.ID, account.Name)
	p := antigravityRetryLoopParams{
		ctx:            ctx,
		prefix:         prefix,
		account:        account,
		proxyURL:       proxyURL,
		accessToken:    accessToken,
		action:         "streamGenerateContent",
		body:           requestBody,
		c:              nil, // 无 gin.Context → 跳过 ops 追踪
		httpUpstream:   s.httpUpstream,
		settingService: s.settingService,
		accountRepo:    s.accountRepo,
		requestedModel: modelID,
		handleError:    testConnectionHandleError,
	}

	result, err := s.antigravityRetryLoop(p)
	if err != nil {
		// AccountSwitchError → 测试时不切换账号，返回友好提示
		var switchErr *AntigravityAccountSwitchError
		if errors.As(err, &switchErr) {
			return nil, fmt.Errorf("该账号模型 %s 当前限流中，请稍后重试", switchErr.RateLimitedModel)
		}
		return nil, err
	}

	if result == nil || result.resp == nil {
		return nil, errors.New("upstream returned empty response")
	}
	defer func() { _ = result.resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(result.resp.Body, s.upstreamErrorBodyReadLimit()))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if result.resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API 返回 %d: %s", result.resp.StatusCode, string(respBody))
	}

	text := extractTextFromSSEResponse(respBody)
	return &TestConnectionResult{Text: text, MappedModel: mappedModel}, nil
}

// buildGeminiTestRequest 构建 Gemini 格式测试请求
// 使用最小 token 消耗：输入 "." + maxOutputTokens: 1
func (s *AntigravityGatewayService) buildGeminiTestRequest(projectID, model string) ([]byte, error) {
	payload := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]any{
					{"text": "."},
				},
			},
		},
		// Antigravity 上游要求必须包含身份提示词
		"systemInstruction": map[string]any{
			"parts": []map[string]any{
				{"text": antigravity.GetDefaultIdentityPatch()},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": 1,
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	return s.wrapV1InternalRequest(projectID, model, payloadBytes)
}

// buildClaudeTestRequest 构建 Claude 格式测试请求并转换为 Gemini 格式
// 使用最小 token 消耗：输入 "." + MaxTokens: 1
func (s *AntigravityGatewayService) buildClaudeTestRequest(projectID, mappedModel string) ([]byte, error) {
	claudeReq := &antigravity.ClaudeRequest{
		Model: mappedModel,
		Messages: []antigravity.ClaudeMessage{
			{
				Role:    "user",
				Content: json.RawMessage(`"."`),
			},
		},
		MaxTokens: 1,
		Stream:    false,
	}
	return antigravity.TransformClaudeToGemini(claudeReq, projectID, mappedModel)
}

func (s *AntigravityGatewayService) getClaudeTransformOptions(ctx context.Context) antigravity.TransformOptions {
	opts := antigravity.DefaultTransformOptions()
	if s.settingService == nil {
		return opts
	}
	opts.EnableIdentityPatch = s.settingService.IsIdentityPatchEnabled(ctx)
	opts.IdentityPatch = s.settingService.GetIdentityPatchPrompt(ctx)
	return opts
}

// injectIdentityPatchToGeminiRequest 为 Gemini 格式请求注入身份提示词
// 如果请求中已包含 "You are Antigravity" 则不重复注入
func injectIdentityPatchToGeminiRequest(body []byte) ([]byte, error) {
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, fmt.Errorf("解析 Gemini 请求失败: %w", err)
	}

	// Google 的 v1internal 端点在存在 systemInstruction 时会更严格校验 contents.role。
	// 原生 Gemini 请求允许省略 role，此时默认语义等价于 user；这里显式补齐，
	// 保持代理注入身份提示词后的兼容性。
	if contents, ok := request["contents"].([]any); ok {
		for _, content := range contents {
			contentMap, ok := content.(map[string]any)
			if !ok {
				continue
			}
			if role, ok := contentMap["role"].(string); ok && strings.TrimSpace(role) != "" {
				continue
			}
			if _, hasParts := contentMap["parts"]; hasParts {
				contentMap["role"] = "user"
			}
		}
	}

	// 检查现有 systemInstruction 是否已包含身份提示词
	if sysInst, ok := request["systemInstruction"].(map[string]any); ok {
		if parts, ok := sysInst["parts"].([]any); ok {
			for _, part := range parts {
				if partMap, ok := part.(map[string]any); ok {
					if text, ok := partMap["text"].(string); ok {
						if strings.Contains(text, "You are Antigravity") {
							// 已包含身份提示词，直接返回原始请求
							return body, nil
						}
					}
				}
			}
		}
	}

	// 获取默认身份提示词
	identityPatch := antigravity.GetDefaultIdentityPatch()

	// 构建新的 systemInstruction
	newPart := map[string]any{"text": identityPatch}

	if existing, ok := request["systemInstruction"].(map[string]any); ok {
		// 已有 systemInstruction，在开头插入身份提示词
		if parts, ok := existing["parts"].([]any); ok {
			existing["parts"] = append([]any{newPart}, parts...)
		} else {
			existing["parts"] = []any{newPart}
		}
	} else {
		// 没有 systemInstruction，创建新的
		request["systemInstruction"] = map[string]any{
			"parts": []any{newPart},
		}
	}

	return json.Marshal(request)
}

// stripThinkingFromClaudeRequest converts thinking blocks to text blocks in a Claude Messages request.
// This preserves the thinking content while avoiding signature validation errors.
// Note: redacted_thinking blocks are removed because they cannot be converted to text.
// It also disables top-level `thinking` to avoid upstream structural constraints for thinking mode.
func stripThinkingFromClaudeRequest(req *antigravity.ClaudeRequest) (bool, error) {
	if req == nil {
		return false, nil
	}

	changed := false
	if req.Thinking != nil {
		req.Thinking = nil
		changed = true
	}

	for i := range req.Messages {
		raw := req.Messages[i].Content
		if len(raw) == 0 {
			continue
		}

		// If content is a string, nothing to strip.
		var str string
		if json.Unmarshal(raw, &str) == nil {
			continue
		}

		// Otherwise treat as an array of blocks and convert thinking blocks to text.
		var blocks []map[string]any
		if err := json.Unmarshal(raw, &blocks); err != nil {
			continue
		}

		filtered := make([]map[string]any, 0, len(blocks))
		modifiedAny := false
		for _, block := range blocks {
			t, _ := block["type"].(string)
			switch t {
			case "thinking":
				thinkingText, _ := block["thinking"].(string)
				if thinkingText != "" {
					filtered = append(filtered, map[string]any{
						"type": "text",
						"text": thinkingText,
					})
				}
				modifiedAny = true
			case "redacted_thinking":
				modifiedAny = true
			case "":
				if thinkingText, hasThinking := block["thinking"].(string); hasThinking {
					if thinkingText != "" {
						filtered = append(filtered, map[string]any{
							"type": "text",
							"text": thinkingText,
						})
					}
					modifiedAny = true
				} else {
					filtered = append(filtered, block)
				}
			default:
				filtered = append(filtered, block)
			}
		}

		if !modifiedAny {
			continue
		}

		if len(filtered) == 0 {
			filtered = append(filtered, map[string]any{
				"type": "text",
				"text": "(content removed)",
			})
		}

		newRaw, err := json.Marshal(filtered)
		if err != nil {
			return changed, err
		}
		req.Messages[i].Content = newRaw
		changed = true
	}

	return changed, nil
}

// stripSignatureSensitiveBlocksFromClaudeRequest is a stronger retry degradation that additionally converts
// tool blocks to plain text. Use this only after a thinking-only retry still fails with signature errors.
func stripSignatureSensitiveBlocksFromClaudeRequest(req *antigravity.ClaudeRequest) (bool, error) {
	if req == nil {
		return false, nil
	}

	changed := false
	if req.Thinking != nil {
		req.Thinking = nil
		changed = true
	}

	for i := range req.Messages {
		raw := req.Messages[i].Content
		if len(raw) == 0 {
			continue
		}

		// If content is a string, nothing to strip.
		var str string
		if json.Unmarshal(raw, &str) == nil {
			continue
		}

		// Otherwise treat as an array of blocks and convert signature-sensitive blocks to text.
		var blocks []map[string]any
		if err := json.Unmarshal(raw, &blocks); err != nil {
			continue
		}

		filtered := make([]map[string]any, 0, len(blocks))
		modifiedAny := false
		for _, block := range blocks {
			t, _ := block["type"].(string)
			switch t {
			case "thinking":
				// Convert thinking to text, skip if empty
				thinkingText, _ := block["thinking"].(string)
				if thinkingText != "" {
					filtered = append(filtered, map[string]any{
						"type": "text",
						"text": thinkingText,
					})
				}
				modifiedAny = true
			case "redacted_thinking":
				// Remove redacted_thinking (cannot convert encrypted content)
				modifiedAny = true
			case "tool_use":
				// Convert tool_use to text to avoid upstream signature/thought_signature validation errors.
				// This is a retry-only degradation path, so we prioritise request validity over tool semantics.
				name, _ := block["name"].(string)
				id, _ := block["id"].(string)
				input := block["input"]
				inputJSON, _ := json.Marshal(input)
				text := "(tool_use)"
				if name != "" {
					text += " name=" + name
				}
				if id != "" {
					text += " id=" + id
				}
				if len(inputJSON) > 0 && string(inputJSON) != "null" {
					text += " input=" + string(inputJSON)
				}
				filtered = append(filtered, map[string]any{
					"type": "text",
					"text": text,
				})
				modifiedAny = true
			case "tool_result":
				// Convert tool_result to text so it stays consistent when tool_use is downgraded.
				toolUseID, _ := block["tool_use_id"].(string)
				isError, _ := block["is_error"].(bool)
				content := block["content"]
				contentJSON, _ := json.Marshal(content)
				text := "(tool_result)"
				if toolUseID != "" {
					text += " tool_use_id=" + toolUseID
				}
				if isError {
					text += " is_error=true"
				}
				if len(contentJSON) > 0 && string(contentJSON) != "null" {
					text += "\n" + string(contentJSON)
				}
				filtered = append(filtered, map[string]any{
					"type": "text",
					"text": text,
				})
				modifiedAny = true
			case "":
				// Handle untyped block with "thinking" field
				if thinkingText, hasThinking := block["thinking"].(string); hasThinking {
					if thinkingText != "" {
						filtered = append(filtered, map[string]any{
							"type": "text",
							"text": thinkingText,
						})
					}
					modifiedAny = true
				} else {
					filtered = append(filtered, block)
				}
			default:
				filtered = append(filtered, block)
			}
		}

		if !modifiedAny {
			continue
		}

		if len(filtered) == 0 {
			// Keep request valid: upstream rejects empty content arrays.
			filtered = append(filtered, map[string]any{
				"type": "text",
				"text": "(content removed)",
			})
		}

		newRaw, err := json.Marshal(filtered)
		if err != nil {
			return changed, err
		}
		req.Messages[i].Content = newRaw
		changed = true
	}

	return changed, nil
}

// isImageGenerationModel 判断模型是否为图片生成模型
// 支持的模型：gemini-3.1-flash-image, gemini-3-pro-image, gemini-2.5-flash-image 等
func isImageGenerationModel(model string) bool {
	modelLower := strings.ToLower(model)
	// 移除 models/ 前缀
	modelLower = strings.TrimPrefix(modelLower, "models/")

	// 精确匹配或前缀匹配
	return modelLower == "gemini-3.1-flash-image" ||
		modelLower == "gemini-3.1-flash-image-preview" ||
		strings.HasPrefix(modelLower, "gemini-3.1-flash-image-") ||
		modelLower == "gemini-3-pro-image" ||
		modelLower == "gemini-3-pro-image-preview" ||
		strings.HasPrefix(modelLower, "gemini-3-pro-image-") ||
		modelLower == "gemini-2.5-flash-image" ||
		modelLower == "gemini-2.5-flash-image-preview" ||
		strings.HasPrefix(modelLower, "gemini-2.5-flash-image-")
}

// cleanGeminiRequest 清理 Gemini 请求体中的 Schema
func cleanGeminiRequest(body []byte) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	modified := false

	// 1. 清理 Tools
	if tools, ok := payload["tools"].([]any); ok && len(tools) > 0 {
		for _, t := range tools {
			toolMap, ok := t.(map[string]any)
			if !ok {
				continue
			}

			// function_declarations (snake_case) or functionDeclarations (camelCase)
			var funcs []any
			if f, ok := toolMap["functionDeclarations"].([]any); ok {
				funcs = f
			} else if f, ok := toolMap["function_declarations"].([]any); ok {
				funcs = f
			}

			if len(funcs) == 0 {
				continue
			}

			for _, f := range funcs {
				funcMap, ok := f.(map[string]any)
				if !ok {
					continue
				}

				if params, ok := funcMap["parameters"].(map[string]any); ok {
					antigravity.DeepCleanUndefined(params)
					cleaned := antigravity.CleanJSONSchema(params)
					funcMap["parameters"] = cleaned
					modified = true
				}
			}
		}
	}

	if !modified {
		return body, nil
	}

	return json.Marshal(payload)
}

// filterEmptyPartsFromGeminiRequest 过滤掉 parts 为空的消息
// Gemini API 不接受空 parts，需要在请求前过滤
func filterEmptyPartsFromGeminiRequest(body []byte) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	contents, ok := payload["contents"].([]any)
	if !ok || len(contents) == 0 {
		return body, nil
	}

	filtered := make([]any, 0, len(contents))
	modified := false

	for _, c := range contents {
		contentMap, ok := c.(map[string]any)
		if !ok {
			filtered = append(filtered, c)
			continue
		}

		parts, hasParts := contentMap["parts"]
		if !hasParts {
			filtered = append(filtered, c)
			continue
		}

		partsSlice, ok := parts.([]any)
		if !ok {
			filtered = append(filtered, c)
			continue
		}

		// 跳过 parts 为空数组的消息
		if len(partsSlice) == 0 {
			modified = true
			continue
		}

		filtered = append(filtered, c)
	}

	if !modified {
		return body, nil
	}

	payload["contents"] = filtered
	return json.Marshal(payload)
}
