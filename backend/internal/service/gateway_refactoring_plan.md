# Gateway Service 重构方案

## 当前状态
- **文件**: `gateway_service.go`
- **行数**: 10,932 行
- **函数/方法**: 284个
- **问题**: 单一文件过大，职责混杂，维护困难

## 重构目标
将 `gateway_service.go` 拆分为 7 个功能域独立文件，每个文件职责单一、边界清晰。

---

## 拆分方案

### 1. **gateway_session_manager.go** (会话管理)
**职责**: 粘性会话管理、会话哈希生成、会话绑定

**包含方法**:
- `GenerateSessionHash` - 生成会话哈希
- `BindStickySession` - 绑定粘性会话
- `GetCachedSessionAccountID` - 获取缓存的账号ID
- `FindGeminiSession` / `SaveGeminiSession` - Gemini会话
- `FindAnthropicSession` / `SaveAnthropicSession` - Anthropic会话
- `extractCacheableContent` - 提取可缓存内容
- `hashContent` - 内容哈希
- 相关辅助函数: `parseRawJSONView`, `extractTextFromSystemRaw`, `extractTextFromContentRaw`, `appendMessageTextsFromRaw`, `appendResponsesSessionAnchorFromRaw`, `appendResponsesContentText`, `extractCacheableTextFromSystemRaw`, `extractCacheableTextFromMessagesRaw`, `generateSessionUUID`, `buildStableSessionSeed`, `sessionContextDiscriminator`
- `shouldClearStickySession` - 判断是否清理会话
- `buildOAuthMetadataUserID` 系列方法

**估计行数**: ~1200行

---

### 2. **gateway_account_scheduler.go** (账号调度)
**职责**: 账号选择、负载均衡、可调度性检查、账号过滤

**包含方法**:
- `SelectAccount` - 选择账号
- `SelectAccountForModel` - 为模型选择账号
- `SelectAccountForModelWithExclusions` - 带排除的账号选择
- `SelectAccountWithLoadAwareness` - 负载感知的账号选择
- `selectAccountWithMixedScheduling` - 混合调度策略
- `selectAccountForModelWithPlatform` - 平台级账号选择
- `listSchedulableAccounts` - 列出可调度账号
- `isAccountSchedulableForSelection` - 账号可调度性检查
- `isAccountSchedulableForModelSelection` - 模型级可调度性
- `isAccountSchedulableForRPM` - RPM限制检查
- `isAccountSchedulableForQuota` - 配额检查
- `isAccountSchedulableForWindowCost` - 窗口成本检查
- `isAccountAllowedForPlatform` - 平台许可检查
- `isAccountInGroup` - 账号分组检查
- `tryAcquireAccountSlot` - 尝试获取账号槽位
- `IncrementAccountRPM` - 增加账号RPM
- `checkAndRegisterSession` - 检查并注册会话
- 账号排序/过滤方法: `sortAccountsByPriorityOnly`, `sortAccountsByPriorityAndLastUsed`, `sortCandidatesForFallback`, `shuffleWithinPriority`, `shuffleWithinPriorityAndLastUsed`, `shuffleWithinSortGroups`, `selectByLRU`, `filterByMinPriority`, `filterByMinLoadRate`, `filterBySoonestReset`
- `hydrateSelectedAccount`, `getSchedulableAccount`, `newSelectionResult`
- `tryAcquireByLegacyOrder` - 传统顺序获取
- 诊断方法: `logDetailedSelectionFailure`, `collectSelectionFailureStats`, `diagnoseSelectionFailure`
- RPM/窗口成本预取: `withRPMPrefetch`, `rpmFromPrefetchContext`, `withWindowCostPrefetch`, `windowCostFromPrefetchContext`

**估计行数**: ~2800行

---

### 3. **gateway_billing_tracker.go** (计费追踪)
**职责**: 用量记录、成本计算、计费扣款、配额管理

**包含方法**:
- `RecordUsage` - 记录用量
- `RecordUsageWithLongContext` - 长上下文用量记录
- `recordUsageCore` - 核心用量记录
- `calculateRecordUsageCost` - 计算用量成本
- `calculateTokenCost` - 计算token成本
- `calculateImageCost` - 计算图片成本
- `resolveChannelPricing` - 解析通道定价
- `buildRecordUsageLog` - 构建用量日志
- 计费相关: `postUsageBilling`, `applyUsageBilling`, `buildUsageBillingCommand`, `finalizePostUsageBilling`
- `resolveBillingMode`, `optionalSubscriptionID`
- 辅助方法: `resolveUsageBillingRequestID`, `resolveUsageBillingPayloadFingerprint`, `syncBalanceCacheAfterDeduction`
- 通知方法: `notifyBalanceLow`, `notifyAccountQuota`, `resolveOldBalance`
- `shouldDeductAPIKeyQuota`, `shouldUpdateRateLimits`, `shouldUpdateAccountQuota`
- `billingDeps` - 获取计费依赖
- 上下文方法: `detachedBillingContext`, `detachStreamUpstreamContext`, `detachUpstreamContext`
- `writeUsageLogBestEffort`
- `QuotaPlatform`, `PlatformFromAPIKey`
- `getUserGroupRateMultiplier`

**估计行数**: ~1500行

---

### 4. **gateway_request_handler.go** (请求处理)
**职责**: 上游请求构建、Beta策略、模型映射、请求规范化

**包含方法**:
- `buildUpstreamRequest` - 构建上游请求
- `buildUpstreamRequestAnthropicAPIKeyPassthrough` - Anthropic API Key直通
- `buildUpstreamRequestAnthropicVertex` - Anthropic Vertex
- `buildUpstreamRequestBedrock` - Bedrock请求
- `buildUpstreamRequestBedrockAPIKey` - Bedrock API Key
- Beta策略相关: `computeFinalAnthropicBeta`, `computeFinalCountTokensAnthropicBeta`, `stripBetaTokens`, `stripBetaTokensWithSet`, `evaluateBetaPolicy`, `getBetaPolicyFilterSet`, `mergeDropSets`, `checkBetaPolicyBlockForTokens`, `resolveBedrockBetaTokensForRequest`, `filterBetaTokens`, `containsBetaToken`, `droppedBetaSet`, `buildBetaTokenSet`, `resolveRuleAction`, `matchModelWhitelist`, `betaPolicyScopeMatches`, `filterVertexBetaTokens`, `mergeAnthropicBeta`, `mergeAnthropicBetaDropping`
- 模型映射: `ReplaceModelInBody`, `replaceModelInBody`, `replaceModelInResponseBody`
- 通道映射: `ResolveChannelMapping`, `ResolveChannelMappingAndRestrict`, `IsModelRestricted`, `checkChannelPricingRestriction`, `isUpstreamModelRestrictedByChannel`, `needsUpstreamChannelRestrictionCheck`, `isStickyAccountUpstreamRestricted`, `billingModelForRestriction`, `resolveAccountUpstreamModel`
- 请求规范化: Claude OAuth相关的 `normalizeClaudeOAuthRequestBody`, `normalizeClaudeOAuthSystemBody`, `ensureClaudeOAuthMetadataUserID`, `applyClaudeCodeOAuthMimicryToBody`, `buildOAuthMetadataUserIDFromBody`
- Claude Code系统提示词: `injectClaudeCodePrompt`, `rewriteSystemForNonClaudeCode`, `rewriteSystemForNonClaudeCodeWithPrompt`, `rewriteSystemForNonClaudeCodeWithPromptBlocks`, `buildClaudeOAuthSystemPromptBlocksJSON`, `parseClaudeOAuthSystemPromptBlocksConfig`, `defaultClaudeOAuthSystemPromptBlockConfig`, `expandClaudeOAuthSystemPromptTextTemplate`, `decodeClaudeOAuthSystemPromptCacheControl`, `ValidateClaudeOAuthSystemPromptBlocksConfig`
- Cache TTL相关: `applyCacheTTLOverride`, `rewriteCacheCreationJSON`, `resolveCacheTTLUsageOverrideTarget`, `shouldInjectAnthropicCacheTTL1h`, `injectAnthropicCacheControlTTL1h`, `forceEphemeralCacheControlTTL`, `enforceCacheControlLimit`, `collectCacheControlPaths`
- 系统提示词: `normalizeSystemParam`, `systemIncludesClaudeCodePrompt`, `hasClaudeCodePrefix`, `sanitizeSystemText`
- JSON操作辅助: `marshalAnthropicSystemTextBlock`, `marshalAnthropicSystemTextBlockWithCacheControl`, `marshalAnthropicMetadata`, `buildJSONArrayRaw`, `setJSONValueBytes`, `setJSONRawBytes`, `deleteJSONPathBytes`
- Claude Code mimic: `applyClaudeCodeMimicHeaders`, `applyClaudeOAuthHeaderDefaults`, `getBetaHeader`, `requestNeedsBetaFeatures`, `defaultAPIKeyBetaHeader`
- Bedrock: `ApplyBedrockCCCompat`, `isBedrockCCCompatEnabled`
- URL相关: `buildCustomRelayURL`, `validateUpstreamBaseURL`
- CountTokens: `buildCountTokensRequest`, `buildCountTokensRequestAnthropicAPIKeyPassthrough`, `sanitizeCountTokensRequestBody`

**估计行数**: ~2500行

---

### 5. **gateway_response_processor.go** (响应处理)
**职责**: 流式/非流式响应处理、SSE解析、错误处理

**包含方法**:
- `handleStreamingResponse` - 处理流式响应
- `handleNonStreamingResponse` - 处理非流式响应
- `handleStreamingResponseAnthropicAPIKeyPassthrough` - Anthropic流式直通
- `handleNonStreamingResponseAnthropicAPIKeyPassthrough` - Anthropic非流式直通
- `handleBedrockNonStreamingResponse` - Bedrock非流式响应
- SSE解析: `parseSSEUsage`, `parseSSEUsagePassthrough`, `extractSSEUsagePatch`, `mergeSSEUsagePatch`, `parseSSEUsageInt`, `parseSSEUsageFromBody`
- SSE数据提取: `extractAnthropicSSEDataLine`, `extractOpenAISSEDataLine`, `extractOpenAISSEEventLine`, `extractOpenAISSETerminalEvent`, `extractOpenAISSEErrorMessage`
- 错误处理: `handleErrorResponse`, `handleFailoverSideEffects`, `handleRetryExhaustedSideEffects`, `handleRetryExhaustedError`
- 错误读取: `readUpstreamErrorBody`, `ExtractUpstreamErrorMessage`, `extractUpstreamErrorMessage`, `extractUpstreamErrorCode`, `isCountTokensUnsupported404`
- Failover判断: `shouldFailoverOn400`, `shouldFailoverUpstreamError`, `shouldRetryUpstreamError`
- 签名错误: `isThinkingBlockSignatureError`, `shouldRectifySignatureError`, `isSignatureErrorPattern`, `matchSignaturePatterns`
- 上游错误处理: `handleBedrockUpstreamErrors`, `executeBedrockUpstream`
- 响应头处理: `writeAnthropicPassthroughResponseHeaders`
- Failover错误: `invalidNonStreamingJSONFailoverError`
- Claude Code Noop Delta: `shouldUseClaudeCodeNoopDeltaKeepalive`, `claudeCodeKeepaliveDeltaTypeForContentBlock`, `claudeCodeKeepaliveFieldForDeltaType`, `buildClaudeCodeNoopDeltaKeepalive`, `sseEventIndex`
- 辅助判断: `isClaudeCodeClient`, `isEventStreamResponse`, `bodyHasSSEFraming`, `openAIStreamEventIsTerminal`, `anthropicStreamEventIsTerminal`
- 响应体处理: `parseClaudeUsageFromResponseBody`, `normalizeResponsesStreamingTerminalOutput`, `responsesStreamEventMayContributeToOutput`, `reconstructResponseOutputFromSSE`, `buildResponsesOutputJSON`, `extractImageGenerationOutputFromSSEData`, `replaceModelInSSELine`, `normalizeOpenAIResponsesFunctionCallArguments`, `correctToolCallsInResponseBody`, `extractOpenAIResponseIDFromJSONBytes`, `extractOpenAIUsageFromJSONBytes`, `parseSSEUsageBytes`, `bindHTTPResponseAccount`, `openAIUsageFromGJSON`
- SSE到JSON: `handleSSEToJSON`, `writeOpenAINonStreamingProtocolError`
- 错误清理: `sanitizeStreamError`, `sanitizeOpenAIResponseFailedEventForClient`
- Sleep: `sleepWithContext`, `retryBackoffDelay`

**估计行数**: ~2200行

---

### 6. **gateway_orchestrator.go** (核心编排器)
**职责**: 协调各模块、主要转发流程、分组解析、OAuth令牌管理

**包含方法**:
- `Forward` - 主转发方法 (核心编排)
- `forwardAnthropicAPIKeyPassthrough` - Anthropic直通转发
- `forwardAnthropicAPIKeyPassthroughWithInput` - 带输入的Anthropic直通
- `forwardBedrock` - Bedrock转发
- `ForwardCountTokens` - CountTokens转发
- `forwardCountTokensAnthropicAPIKeyPassthrough` - CountTokens Anthropic直通
- 分组解析: `resolveGatewayGroup`, `resolveGroupByID`, `ResolveGroupByID`, `groupFromContext`, `withGroupContext`, `routingAccountIDsForRequest`
- Claude Code限制: `checkClaudeCodeRestriction`
- 平台解析: `resolvePlatform`
- OAuth: `GetAccessToken`, `getOAuthToken`
- 模型支持: `isModelSupportedByAccount`, `isModelSupportedByAccountWithContext`
- 可用模型: `GetAvailableModels`, `InvalidateAvailableModelsCache`
- 临时不可调度: `TempUnscheduleRetryableError`
- 调度配置: `schedulingConfig`
- 特殊分组: `IsSingleAntigravityAccountGroup`
- 客户端归一化: `shouldNormalizeClientDateline`, `normalizeClientDatelineIfEnabled`, `claudeOAuthSystemPromptInjectionSettings`
- 上下文辅助: `withGroupContext`, `groupFromContext`, `prefetchedStickyGroupIDFromContext`, `prefetchedStickyAccountIDFromContext`, `shouldClearStickySession`

**估计行数**: ~800行

---

### 7. **gateway_service_core.go** (核心定义和工具)
**职责**: 类型定义、常量、辅助函数、缓存接口、统计指标

**包含内容**:
- 结构体定义: `GatewayService`, `ForwardResult`, `ClaudeUsage`, `UpstreamFailoverError`, `sseStreamErrorEventError`, `accountWithLoad`
- 常量: `claudeAPIURL`, `claudeAPICountTokensURL`, `stickySessionTTL`, `claudeCodeSystemPrompt`, 等
- 构造函数: `NewGatewayService`
- 缓存接口: `GatewayCache`
- 统计函数: `GatewayWindowCostPrefetchStats`, `GatewayUserGroupRateCacheStats`, `GatewayModelsListCacheStats`, `GatewayUserPlatformQuotaIncrStats`, `GatewayUserPlatformQuotaFlusherStats`
- 辅助函数: `derefGroupID`, `resolveUserGroupRateCacheTTL`, `resolveModelsListCacheTTL`, `modelsListCacheKey`
- Debug函数: `debugModelRoutingEnabled`, `debugClaudeMimicEnabled`, `parseDebugEnvBool`, `logClaudeMimicDebug`, `buildClaudeMimicDebugLine`, `extractSystemPreviewFromBody`, `safeHeaderValueForLog`, `redactAuthHeaderValue`, `shortSessionHash`, `debugLogGatewaySnapshot`, `initDebugGatewayBodyFile`
- 上下文函数: `IsForceCacheBilling`, `WithForceCacheBilling`
- 错误判断: `isClaudeCodeCredentialScopeError`, `truncateForLog`
- 工具函数: `cloneStringSlice`, `reconcileCachedTokens`

**估计行数**: ~800行

---

## 重构步骤

### 阶段1: 准备工作
1. ✅ 创建重构计划文档
2. 备份当前 `gateway_service.go`
3. 为每个新文件创建空骨架

### 阶段2: 代码迁移
按以下顺序逐个迁移:
1. `gateway_service_core.go` - 基础定义
2. `gateway_session_manager.go` - 会话管理
3. `gateway_billing_tracker.go` - 计费追踪
4. `gateway_account_scheduler.go` - 账号调度
5. `gateway_request_handler.go` - 请求处理
6. `gateway_response_processor.go` - 响应处理
7. `gateway_orchestrator.go` - 核心编排

### 阶段3: 集成和验证
1. 更新 `wire.go` 依赖注入
2. 运行 `go generate ./...`
3. 编译测试: `go build ./...`
4. 为每个模块添加单元测试
5. 运行集成测试

---

## 接口设计原则

### 内部接口 (跨文件调用)
所有方法保持为 `GatewayService` 的方法,通过结构体共享状态:
```go
type GatewayService struct {
    // 共享依赖
    cache GatewayCache
    repo  repository.Repository
    // ...
}
```

### 方法可见性
- 公开方法 (大写开头): 供外部Handler调用
- 私有方法 (小写开头): 模块内部使用

---

## 测试策略

### 单元测试覆盖率目标
- `gateway_service_core.go`: 80%+
- `gateway_session_manager.go`: 75%+
- `gateway_billing_tracker.go`: 85%+ (核心计费逻辑)
- `gateway_account_scheduler.go`: 80%+ (关键调度算法)
- `gateway_request_handler.go`: 70%+
- `gateway_response_processor.go`: 75%+
- `gateway_orchestrator.go`: 85%+ (核心流程)

### 测试文件命名
对应文件名 + `_test.go` 后缀

---

## 风险评估

### 高风险区域
1. **账号调度逻辑** - 复杂的负载均衡和过滤规则
2. **计费追踪** - 不能丢失用量数据
3. **会话管理** - 粘性会话绑定关系

### 缓解措施
1. 重构前补充现有单元测试
2. 每个模块迁移后立即编译验证
3. 保留原文件副本用于对比
4. 增量提交,每个模块独立PR

---

## 文件大小预估

| 文件 | 预估行数 | 当前占比 |
|------|---------|---------|
| gateway_service_core.go | ~800 | 7% |
| gateway_session_manager.go | ~1200 | 11% |
| gateway_billing_tracker.go | ~1500 | 14% |
| gateway_account_scheduler.go | ~2800 | 26% |
| gateway_request_handler.go | ~2500 | 23% |
| gateway_response_processor.go | ~2200 | 20% |
| gateway_orchestrator.go | ~800 | 7% |
| **总计** | **~11800** | **108%** (重复统计) |

实际总行数约 10,932行，预估略有重叠属正常。

---

## 成功标准

### 代码质量
- ✅ 单个文件不超过 3000 行
- ✅ 单个方法不超过 150 行
- ✅ 每个文件职责单一清晰
- ✅ 无编译错误
- ✅ 所有现有测试通过

### 可维护性
- ✅ 新开发者能快速定位功能代码
- ✅ 修改某个功能域只需改动对应文件
- ✅ 模块间依赖关系清晰

---

## 时间估算
- 阶段1 准备: 0.5天 (已完成)
- 阶段2 迁移: 3-4天
- 阶段3 测试: 1-2天
- **总计**: 4.5-6.5天

---

*重构计划创建时间: 2026-07-08*
*计划负责人: AI Code Reviewer*
