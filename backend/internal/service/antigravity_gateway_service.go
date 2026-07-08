package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	antigravityStickySessionTTL = time.Hour
	antigravityMaxRetries       = 3
	antigravityRetryBaseDelay   = 1 * time.Second
	antigravityRetryMaxDelay    = 16 * time.Second

	// 限流相关常量
	// antigravityRateLimitThreshold 限流等待/切换阈值
	// - 智能重试：retryDelay < 此阈值时等待后重试，>= 此阈值时直接限流模型
	// - 预检查：剩余限流时间 < 此阈值时等待，>= 此阈值时切换账号
	antigravityRateLimitThreshold       = 7 * time.Second
	antigravitySmartRetryMinWait        = 1 * time.Second  // 智能重试最小等待时间
	antigravitySmartRetryMaxAttempts    = 1                // 智能重试最大次数（仅重试 1 次，防止重复限流/长期等待）
	antigravityDefaultRateLimitDuration = 30 * time.Second // 默认限流时间（无 retryDelay 时使用）

	// MODEL_CAPACITY_EXHAUSTED 专用重试参数
	// 模型容量不足时，所有账号共享同一容量池，切换账号无意义
	// 使用固定 1s 间隔重试，最多重试 60 次
	antigravityModelCapacityRetryMaxAttempts = 60
	antigravityModelCapacityRetryWait        = 1 * time.Second

	// Google RPC 状态和类型常量
	googleRPCStatusResourceExhausted      = "RESOURCE_EXHAUSTED"
	googleRPCStatusUnavailable            = "UNAVAILABLE"
	googleRPCTypeRetryInfo                = "type.googleapis.com/google.rpc.RetryInfo"
	googleRPCTypeErrorInfo                = "type.googleapis.com/google.rpc.ErrorInfo"
	googleRPCReasonModelCapacityExhausted = "MODEL_CAPACITY_EXHAUSTED"
	googleRPCReasonRateLimitExceeded      = "RATE_LIMIT_EXCEEDED"

	// 单账号 503 退避重试：Service 层原地重试的最大次数
	// 在 handleSmartRetry 中，对于 shouldRateLimitModel（长延迟 ≥ 7s）的情况，
	// 多账号模式下会设限流+切换账号；但单账号模式下改为原地等待+重试。
	antigravitySingleAccountSmartRetryMaxAttempts = 3

	// 单账号 503 退避重试：原地重试时单次最大等待时间
	// 防止上游返回过长的 retryDelay 导致请求卡住太久
	antigravitySingleAccountSmartRetryMaxWait = 15 * time.Second

	// 单账号 503 退避重试：原地重试的总累计等待时间上限
	// 超过此上限将不再重试，直接返回 503
	antigravitySingleAccountSmartRetryTotalMaxWait = 30 * time.Second

	// MODEL_CAPACITY_EXHAUSTED 全局去重：重试全部失败后的 cooldown 时间
	antigravityModelCapacityCooldown = 10 * time.Second
)

// antigravityPassthroughErrorMessages 透传给客户端的错误消息白名单（小写）
// 匹配时使用 strings.Contains，无需完全匹配
var antigravityPassthroughErrorMessages = []string{
	"prompt is too long",
}

// MODEL_CAPACITY_EXHAUSTED 全局去重：避免多个并发请求同时对同一模型进行容量耗尽重试
var (
	modelCapacityExhaustedMu    sync.RWMutex
	modelCapacityExhaustedUntil = make(map[string]time.Time) // modelName -> cooldown until
)

const (
	antigravityForwardBaseURLEnv  = "GATEWAY_ANTIGRAVITY_FORWARD_BASE_URL"
	antigravityFallbackSecondsEnv = "GATEWAY_ANTIGRAVITY_FALLBACK_COOLDOWN_SECONDS"
)

const antigravityProjectIDFallbackCredentialKey = "antigravity_project_id"

var errAntigravityProjectIDRequired = errors.New("该 standard-tier Antigravity 账号需配置 project_id")

// AntigravityAccountSwitchError 账号切换信号
// 当账号限流时间超过阈值时，通知上层切换账号
type AntigravityAccountSwitchError struct {
	OriginalAccountID int64
	RateLimitedModel  string
	IsStickySession   bool // 是否为粘性会话切换（决定是否缓存计费）
}

func (e *AntigravityAccountSwitchError) Error() string {
	return fmt.Sprintf("account %d model %s rate limited, need switch",
		e.OriginalAccountID, e.RateLimitedModel)
}

// PromptTooLongError 表示上游明确返回 prompt too long
type PromptTooLongError struct {
	StatusCode int
	RequestID  string
	Body       []byte
}

func (e *PromptTooLongError) Error() string {
	return fmt.Sprintf("prompt too long: status=%d", e.StatusCode)
}

// antigravityRetryLoopParams 重试循环的参数
type antigravityRetryLoopParams struct {
	ctx             context.Context
	prefix          string
	account         *Account
	proxyURL        string
	accessToken     string
	action          string
	body            []byte
	c               *gin.Context
	httpUpstream    HTTPUpstream
	settingService  *SettingService
	accountRepo     AccountRepository // 用于智能重试的模型级别限流
	handleError     func(ctx context.Context, prefix string, account *Account, statusCode int, headers http.Header, body []byte, requestedModel string, groupID int64, sessionHash string, isStickySession bool) *handleModelRateLimitResult
	requestedModel  string // 用于限流检查的原始请求模型
	isStickySession bool   // 是否为粘性会话（用于账号切换时的缓存计费判断）
	groupID         int64  // 用于模型级限流时清除粘性会话
	sessionHash     string // 用于模型级限流时清除粘性会话
}

// antigravityRetryLoopResult 重试循环的结果
type antigravityRetryLoopResult struct {
	resp *http.Response
}

// smartRetryAction 智能重试的处理结果
type smartRetryAction int

const (
	smartRetryActionContinue      smartRetryAction = iota // 继续默认重试逻辑
	smartRetryActionBreakWithResp                         // 结束循环并返回 resp
	smartRetryActionContinueURL                           // 继续 URL fallback 循环
)

// smartRetryResult 智能重试的结果
type smartRetryResult struct {
	action      smartRetryAction
	resp        *http.Response
	err         error
	switchError *AntigravityAccountSwitchError // 模型限流时返回账号切换信号
}

// getSessionID 从 gin.Context 获取 session_id（用于日志追踪）
func getSessionID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.GetHeader("session_id")
}

// logPrefix 生成统一的日志前缀
func logPrefix(sessionID, accountName string) string {
	if sessionID != "" {
		return fmt.Sprintf("[antigravity-Forward] session=%s account=%s", sessionID, accountName)
	}
	return fmt.Sprintf("[antigravity-Forward] account=%s", accountName)
}

// AntigravityGatewayService 处理 Antigravity 平台的 API 转发
type AntigravityGatewayService struct {
	accountRepo       AccountRepository
	tokenProvider     *AntigravityTokenProvider
	rateLimitService  *RateLimitService
	httpUpstream      HTTPUpstream
	settingService    *SettingService
	cache             GatewayCache // 用于模型级限流时清除粘性会话绑定
	schedulerSnapshot *SchedulerSnapshotService
	internal500Cache  Internal500CounterCache // INTERNAL 500 渐进惩罚计数器
}

func NewAntigravityGatewayService(
	accountRepo AccountRepository,
	cache GatewayCache,
	schedulerSnapshot *SchedulerSnapshotService,
	tokenProvider *AntigravityTokenProvider,
	rateLimitService *RateLimitService,
	httpUpstream HTTPUpstream,
	settingService *SettingService,
	internal500Cache Internal500CounterCache,
) *AntigravityGatewayService {
	return &AntigravityGatewayService{
		accountRepo:       accountRepo,
		tokenProvider:     tokenProvider,
		rateLimitService:  rateLimitService,
		httpUpstream:      httpUpstream,
		settingService:    settingService,
		cache:             cache,
		schedulerSnapshot: schedulerSnapshot,
		internal500Cache:  internal500Cache,
	}
}

// GetTokenProvider 返回 token provider
func (s *AntigravityGatewayService) GetTokenProvider() *AntigravityTokenProvider {
	return s.tokenProvider
}

// getLogConfig 获取上游错误日志配置
// 返回是否记录日志体和最大字节数
func (s *AntigravityGatewayService) getLogConfig() (logBody bool, maxBytes int) {
	maxBytes = 2048 // 默认值
	if s.settingService == nil || s.settingService.cfg == nil {
		return false, maxBytes
	}
	cfg := s.settingService.cfg.Gateway
	if cfg.LogUpstreamErrorBodyMaxBytes > 0 {
		maxBytes = cfg.LogUpstreamErrorBodyMaxBytes
	}
	return cfg.LogUpstreamErrorBody, maxBytes
}

// TestConnectionResult 测试连接结果
type TestConnectionResult struct {
	Text        string // 响应文本
	MappedModel string // 实际使用的模型
}

// ForwardGemini 转发 Gemini 协议请求
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
type ForwardGeminiOption func(*forwardGeminiOptions)

type forwardGeminiOptions struct {
	groupID     int64
	sessionHash string
}

// googleConfigErrorCooldown 服务端配置类 400 错误的临时封禁时长
const googleConfigErrorCooldown = 1 * time.Minute

// emptyResponseCooldown 空流式响应的临时封禁时长
const emptyResponseCooldown = 1 * time.Minute

// antigravitySmartRetryInfo 智能重试所需的信息
type antigravitySmartRetryInfo struct {
	RetryDelay               time.Duration // 重试延迟时间
	ModelName                string        // 限流的模型名称（如 "claude-sonnet-4-5"）
	IsModelCapacityExhausted bool          // 是否为模型容量不足（MODEL_CAPACITY_EXHAUSTED）
}

// handleModelRateLimitParams 模型级限流处理参数
type handleModelRateLimitParams struct {
	ctx             context.Context
	prefix          string
	account         *Account
	statusCode      int
	body            []byte
	cache           GatewayCache
	groupID         int64
	sessionHash     string
	isStickySession bool
}

// handleModelRateLimitResult 模型级限流处理结果
type handleModelRateLimitResult struct {
	Handled      bool                           // 是否已处理
	ShouldRetry  bool                           // 是否等待后重试
	WaitDuration time.Duration                  // 等待时间
	SwitchError  *AntigravityAccountSwitchError // 账号切换错误
}

type antigravityStreamResult struct {
	usage            *ClaudeUsage
	firstTokenMs     *int
	clientDisconnect bool // 客户端是否在流式传输过程中断开
}

// antigravityClientWriter 封装流式响应的客户端写入，自动检测断开并标记。
// 断开后所有写入操作变为 no-op，调用方通过 Disconnected() 判断是否继续 drain 上游。
type antigravityClientWriter struct {
	w            gin.ResponseWriter
	flusher      http.Flusher
	disconnected bool
	prefix       string // 日志前缀，标识来源方法
}
