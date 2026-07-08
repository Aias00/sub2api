package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Aias00/cloudbase/internal/config"
	"github.com/Aias00/cloudbase/internal/pkg/antigravity"
	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/openai"
	"golang.org/x/sync/singleflight"
)

var (
	ErrRegistrationDisabled   = infraerrors.Forbidden("REGISTRATION_DISABLED", "registration is currently disabled")
	ErrSettingNotFound        = infraerrors.NotFound("SETTING_NOT_FOUND", "setting not found")
	ErrDefaultSubGroupInvalid = infraerrors.BadRequest(
		"DEFAULT_SUBSCRIPTION_GROUP_INVALID",
		"default subscription group must exist and be subscription type",
	)
	ErrDefaultSubGroupDuplicate = infraerrors.BadRequest(
		"DEFAULT_SUBSCRIPTION_GROUP_DUPLICATE",
		"default subscription group cannot be duplicated",
	)
)

type SettingRepository interface {
	Get(ctx context.Context, key string) (*Setting, error)
	GetValue(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	GetMultiple(ctx context.Context, keys []string) (map[string]string, error)
	SetMultiple(ctx context.Context, settings map[string]string) error
	GetAll(ctx context.Context) (map[string]string, error)
	Delete(ctx context.Context, key string) error
}

type SiteLogoImage struct {
	ContentType string
	Data        []byte
	ETag        string
}

type parsedSiteLogoDataURL struct {
	ContentType string
	Data        []byte
	ETag        string
}

// cachedVersionBounds 缓存 Claude Code 版本号上下限（进程内缓存，60s TTL）
type cachedVersionBounds struct {
	min       string // 空字符串 = 不检查
	max       string // 空字符串 = 不检查
	expiresAt int64  // unix nano
}

// versionBoundsCache 版本号上下限进程内缓存
var versionBoundsCache atomic.Value // *cachedVersionBounds

// versionBoundsSF 防止缓存过期时 thundering herd
var versionBoundsSF singleflight.Group

// versionBoundsCacheTTL 缓存有效期
const versionBoundsCacheTTL = 60 * time.Second

// versionBoundsErrorTTL DB 错误时的短缓存，快速重试
const versionBoundsErrorTTL = 5 * time.Second

// versionBoundsDBTimeout singleflight 内 DB 查询超时，独立于请求 context
const versionBoundsDBTimeout = 5 * time.Second

// cachedBackendMode Backend Mode cache (in-process, 60s TTL)
type cachedBackendMode struct {
	value     bool
	expiresAt int64 // unix nano
}

var backendModeCache atomic.Value // *cachedBackendMode
var backendModeSF singleflight.Group

const backendModeCacheTTL = 60 * time.Second
const backendModeErrorTTL = 5 * time.Second
const backendModeDBTimeout = 5 * time.Second

// cachedGatewayForwardingSettings 缓存网关转发行为设置（进程内缓存，60s TTL）
type cachedGatewayForwardingSettings struct {
	fingerprintUnification           bool
	metadataPassthrough              bool
	cchSigning                       bool
	claudeOAuthSystemPromptInjection bool
	claudeOAuthSystemPrompt          string
	claudeOAuthSystemPromptBlocks    string
	anthropicCacheTTL1hInjection     bool
	rewriteMessageCacheControl       bool
	clientDatelineNormalization      bool
	expiresAt                        int64 // unix nano
}

var gatewayForwardingCache atomic.Value // *cachedGatewayForwardingSettings
var gatewayForwardingSF singleflight.Group

const gatewayForwardingCacheTTL = 60 * time.Second
const gatewayForwardingErrorTTL = 5 * time.Second
const gatewayForwardingDBTimeout = 5 * time.Second

// cachedAntigravityUserAgentVersion 缓存 Antigravity UA 版本号（进程内缓存，60s TTL）
type cachedAntigravityUserAgentVersion struct {
	version   string
	expiresAt int64 // unix nano
}

const antigravityUserAgentVersionCacheTTL = 60 * time.Second
const antigravityUserAgentVersionErrorTTL = 5 * time.Second
const antigravityUserAgentVersionDBTimeout = 5 * time.Second

// DefaultOpenAICodexUserAgent OpenAI Codex 默认 User-Agent（用于规避 Cloudflare 对浏览器 UA 的质询）
const DefaultOpenAICodexUserAgent = "codex-tui/0.125.0 (Ubuntu 22.4.0; x86_64) xterm-256color (codex-tui; 0.125.0)"

// cachedOpenAICodexUserAgent 缓存 OpenAI Codex UA（进程内缓存，60s TTL）
type cachedOpenAICodexUserAgent struct {
	value     string
	expiresAt int64 // unix nano
}

type cachedOpenAIQuotaAutoPauseSettings struct {
	settings  OpsOpenAIAccountQuotaAutoPauseSettings
	expiresAt int64
}

const openAICodexUserAgentCacheTTL = 60 * time.Second
const openAICodexUserAgentErrorTTL = 5 * time.Second
const openAICodexUserAgentDBTimeout = 5 * time.Second

const codexRestrictionPolicyCacheTTL = 60 * time.Second
const codexRestrictionPolicyDBTimeout = 5 * time.Second

// cachedCodexRestrictionPolicy codex_cli_only 全局加固策略缓存（进程内，60s TTL）。
// GetCodexRestrictionPolicy 在每个 codex_cli_only 账号的网关请求热路径上被调用，避免每次访问 DB。
type cachedCodexRestrictionPolicy struct {
	value     CodexRestrictionPolicy
	expiresAt int64 // unix nano
}

// cachedCyberSessionBlockRuntime cyber 会话屏蔽开关+TTL 进程内缓存（60s TTL）。
// GetCyberSessionBlockRuntime 在网关请求热路径上被调用，避免每次访问 DB。
type cachedCyberSessionBlockRuntime struct {
	enabled   bool
	ttl       time.Duration
	expiresAt int64 // unix nano
}

const cyberSessionBlockRuntimeCacheTTL = 60 * time.Second
const cyberSessionBlockRuntimeErrorTTL = 5 * time.Second
const cyberSessionBlockRuntimeDBTimeout = 5 * time.Second

const openAIQuotaAutoPauseSettingsCacheTTL = 60 * time.Second
const openAIQuotaAutoPauseSettingsErrorTTL = 5 * time.Second
const openAIQuotaAutoPauseSettingsDBTimeout = 5 * time.Second

const openAIQuotaAutoPauseSettingsRefreshKey = "openai_quota_auto_pause_settings"

// DefaultSubscriptionGroupReader validates group references used by default subscriptions.
type DefaultSubscriptionGroupReader interface {
	GetByID(ctx context.Context, id int64) (*Group, error)
}

// WebSearchManagerBuilder creates a websearch.Manager from config (injected by infra layer).
// proxyURLs maps proxy ID to resolved URL for provider-level proxy support.
type WebSearchManagerBuilder func(cfg *WebSearchEmulationConfig, proxyURLs map[int64]string)

// SettingService 系统设置服务
type SettingService struct {
	settingRepo                 SettingRepository
	defaultSubGroupReader       DefaultSubscriptionGroupReader
	proxyRepo                   ProxyRepository // for resolving websearch provider proxy URLs
	cfg                         *config.Config
	onUpdate                    func() // Callback when settings are updated (for cache invalidation)
	version                     string // Application version
	webSearchManagerBuilder     WebSearchManagerBuilder
	antigravityUAVersionCache   atomic.Value // *cachedAntigravityUserAgentVersion
	antigravityUAVersionSF      singleflight.Group
	openAICodexUACache          atomic.Value // *cachedOpenAICodexUserAgent
	openAICodexUASF             singleflight.Group
	codexRestrictionPolicyCache atomic.Value // *cachedCodexRestrictionPolicy
	codexRestrictionPolicySF    singleflight.Group

	cyberSessionBlockRuntimeCache atomic.Value // *cachedCyberSessionBlockRuntime
	cyberSessionBlockRuntimeSF    singleflight.Group

	// openAIQuotaAutoPauseSettingsCache holds the most recently observed quota auto-pause
	// settings. GetOpenAIQuotaAutoPauseSettings reads this atomic.Value on the request hot
	// path without ever blocking on the DB; when the cached entry expires, a background
	// goroutine refreshes it via openAIQuotaAutoPauseSettingsSF (stale-while-revalidate).
	// This per-service field also gives tests natural isolation — each SettingService
	// instance owns its own cache, no shared package-level state.
	openAIQuotaAutoPauseSettingsCache atomic.Value // *cachedOpenAIQuotaAutoPauseSettings
	openAIQuotaAutoPauseSettingsSF    singleflight.Group
}

// DefaultPlatformQuotaSetting 单 platform 三档限额（nil = 沿用上层；0 = 显式禁用；>0 = 上限）
type DefaultPlatformQuotaSetting struct {
	DailyLimitUSD   *float64 `json:"daily"`
	WeeklyLimitUSD  *float64 `json:"weekly"`
	MonthlyLimitUSD *float64 `json:"monthly"`
}

type ProviderDefaultGrantSettings struct {
	Balance          float64
	Concurrency      int
	Subscriptions    []DefaultSubscriptionSetting
	GrantOnSignup    bool
	GrantOnFirstBind bool
	PlatformQuotas   map[string]*DefaultPlatformQuotaSetting // key = platform name
}

type AuthSourceDefaultSettings struct {
	Email                        ProviderDefaultGrantSettings
	GitHub                       ProviderDefaultGrantSettings
	Google                       ProviderDefaultGrantSettings
	ForceEmailOnThirdPartySignup bool
}

type authSourceDefaultKeySet struct {
	// source 是 auth source 标识（如 "email"、"github"），仅用于 parse 时
	// slog.Warn 诊断输出，不再参与 key 拼接（platformQuotas 字段已存完整 key）。
	source           string
	balance          string
	concurrency      string
	subscriptions    string
	grantOnSignup    string
	grantOnFirstBind string
	platformQuotas   string // SettingKeyAuthSourcePlatformQuotas(source)
}

var (
	emailAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "email",
		balance:          SettingKeyAuthSourceDefaultEmailBalance,
		concurrency:      SettingKeyAuthSourceDefaultEmailConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultEmailSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultEmailGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultEmailGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("email"),
	}
	gitHubAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "github",
		balance:          SettingKeyAuthSourceDefaultGitHubBalance,
		concurrency:      SettingKeyAuthSourceDefaultGitHubConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultGitHubSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultGitHubGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultGitHubGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("github"),
	}
	googleAuthSourceDefaultKeys = authSourceDefaultKeySet{
		source:           "google",
		balance:          SettingKeyAuthSourceDefaultGoogleBalance,
		concurrency:      SettingKeyAuthSourceDefaultGoogleConcurrency,
		subscriptions:    SettingKeyAuthSourceDefaultGoogleSubscriptions,
		grantOnSignup:    SettingKeyAuthSourceDefaultGoogleGrantOnSignup,
		grantOnFirstBind: SettingKeyAuthSourceDefaultGoogleGrantOnFirstBind,
		platformQuotas:   SettingKeyAuthSourcePlatformQuotas("google"),
	}
)

const (
	defaultAuthSourceBalance        = 0
	defaultAuthSourceConcurrency    = 5
	defaultWeChatConnectMode        = "open"
	defaultWeChatConnectScopes      = "snsapi_login"
	defaultWeChatConnectFrontend    = "/auth/wechat/callback"
	defaultGitHubOAuthAuthorize     = "https://github.com/login/oauth/authorize"
	defaultGitHubOAuthToken         = "https://github.com/login/oauth/access_token"
	defaultGitHubOAuthUserInfo      = "https://api.github.com/user"
	defaultGitHubOAuthEmails        = "https://api.github.com/user/emails"
	defaultGitHubOAuthScopes        = "read:user user:email"
	defaultGitHubOAuthFrontend      = "/auth/oauth/callback"
	defaultGoogleOAuthAuthorize     = "https://accounts.google.com/o/oauth2/v2/auth"
	defaultGoogleOAuthToken         = "https://oauth2.googleapis.com/token"
	defaultGoogleOAuthUserInfo      = "https://openidconnect.googleapis.com/v1/userinfo"
	defaultGoogleOAuthScopes        = "openid email profile"
	defaultGoogleOAuthFrontend      = "/auth/oauth/callback"
	defaultLoginAgreementMode       = "modal"
	defaultLoginAgreementDate       = "2026-03-31"
	defaultWebAffonsoCookieDuration = "30"
)

// NewSettingService 创建系统设置服务实例
func NewSettingService(settingRepo SettingRepository, cfg *config.Config) *SettingService {
	return &SettingService{
		settingRepo: settingRepo,
		cfg:         cfg,
	}
}

// SetDefaultSubscriptionGroupReader injects an optional group reader for default subscription validation.
func (s *SettingService) SetDefaultSubscriptionGroupReader(reader DefaultSubscriptionGroupReader) {
	s.defaultSubGroupReader = reader
}

// SetProxyRepository injects a proxy repo for resolving websearch provider proxy URLs.
func (s *SettingService) SetProxyRepository(repo ProxyRepository) {
	s.proxyRepo = repo
}

// GetAllSettings 获取所有系统设置
func (s *SettingService) GetAllSettings(ctx context.Context) (*SystemSettings, error) {
	settings, err := s.settingRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}

	return s.parseSettings(settings), nil
}

const defaultWorkspaceShellConfig = `{"zh":{"catalogLabel":"提示词案例","eyebrow":"生图工作台","title":"AI 生图工作台","heroDescription":"从案例库带入提示词，选择模型和参数后直接创建 Cloudbase 生图任务。","draftImported":"已导入「{title}」","draftImportedDescription":"提示词已填入工作台，可以继续调整参数后生成。","promptLabel":"提示词","promptPlaceholder":"输入或从案例库导入提示词","promptTooLong":"提示词过长","clearLabel":"清空","copyPromptLabel":"复制提示词","copySuccessMessage":"提示词已复制","copyEmptyError":"请先输入提示词","workspaceTitle":"任务与产物状态","workspaceDescription":"模型配置、任务历史和产物存储由 Cloudbase 生图工作台统一管理。","workspaceStatus":"登录后可创建真实生图任务，worker 会调用配置的上游模型并回写图片产物。","backToCatalogLabel":"返回案例库"},"en":{"catalogLabel":"Prompt catalog","eyebrow":"Image Workspace","title":"AI Image Workspace","heroDescription":"Bring prompts from the catalog, choose a model and parameters, then create a native Cloudbase image task.","draftImported":"Imported \"{title}\"","draftImportedDescription":"The prompt is ready in the workspace. Adjust parameters before generating.","promptLabel":"Prompt","promptPlaceholder":"Enter a prompt or import one from the catalog","promptTooLong":"Prompt is too long","clearLabel":"Clear","copyPromptLabel":"Copy prompt","copySuccessMessage":"Prompt copied","copyEmptyError":"Enter a prompt first","workspaceTitle":"Task and artifact status","workspaceDescription":"Model config, task history, and artifact storage are managed by the Cloudbase image workspace.","workspaceStatus":"After login, users can create real image tasks; the worker calls the configured upstream model and writes image artifacts back.","backToCatalogLabel":"Back to catalog"}}`

const defaultPromptCatalogShellConfig = `{"zh":{"defaults":{"sourceType":"case","hasImage":true,"pageSize":24,"sortBy":"imported_at","sortOrder":"desc","generatorPath":"/image-generator","generatorDraftSource":"cloudbase-vue-prompt-catalog"},"labels":{"accountActionAuthenticated":"进入控制台","accountActionAnonymous":"登录","eyebrow":"提示词画廊","title":"提示词案例库","description":"直接浏览 Cloudbase 中的提示词案例。筛选和分页由共享 Prompt API 提供。","caseTitle":"提示词案例库","caseDescription":"直接浏览 Cloudbase 中的提示词案例。筛选和分页由共享 Prompt API 提供。","templateTitle":"提示词模板库","templateDescription":"直接浏览 Cloudbase 中的提示词模板。筛选和分页由共享 Prompt API 提供。","total":"总数","sources":"来源","cases":"案例","templates":"模板","search":"搜索","searchPlaceholder":"搜索标题、提示词、标签或来源","caseOnly":"案例","templateOnly":"模板","allTypes":"全部类型","allSources":"全部来源","allCategories":"全部分类","hasImage":"只看有图","resultPrefix":"结果","page":"页","previous":"上一页","next":"下一页","emptyTitle":"没有匹配的案例","emptyDescription":"换一个关键词、来源或分类再试。","noImage":"暂无图片","source":"查看来源","details":"查看","prompt":"提示词","charUnit":"字符","copyPrompt":"复制提示词","promptCopied":"提示词已复制","generate":"去生图","importTitle":"从链接导入案例","importDescription":"管理员可直接导入 X/Twitter 帖子，图片会通过 Cloudbase/R2 同步后进入案例库。","importProviderX":"X / Twitter","importPlaceholder":"粘贴 X/Twitter 帖子链接","importAction":"导入","importing":"导入中...","importSuccess":"已导入案例","importWarnings":"导入提示","loadError":"加载提示词案例失败"}},"en":{"defaults":{"sourceType":"case","hasImage":true,"pageSize":24,"sortBy":"imported_at","sortOrder":"desc","generatorPath":"/image-generator","generatorDraftSource":"cloudbase-vue-prompt-catalog"},"labels":{"accountActionAuthenticated":"Dashboard","accountActionAnonymous":"Log in","eyebrow":"Prompt Catalog","title":"Prompt Catalog","description":"Browse prompt cases directly from Cloudbase. Filtering and pagination are served by the shared prompt API.","caseTitle":"Prompt Catalog","caseDescription":"Browse prompt cases directly from Cloudbase. Filtering and pagination are served by the shared prompt API.","templateTitle":"Prompt Templates","templateDescription":"Browse prompt templates directly from Cloudbase. Filtering and pagination are served by the shared prompt API.","total":"Total","sources":"Sources","cases":"Cases","templates":"Templates","search":"Search","searchPlaceholder":"Search titles, prompts, tags, or sources","caseOnly":"Cases","templateOnly":"Templates","allTypes":"All types","allSources":"All sources","allCategories":"All categories","hasImage":"Images only","resultPrefix":"Results","page":"Page","previous":"Previous","next":"Next","emptyTitle":"No matching prompts","emptyDescription":"Try another keyword, source, or category.","noImage":"No image","source":"View source","details":"Details","prompt":"Prompt","charUnit":"chars","copyPrompt":"Copy prompt","promptCopied":"Prompt copied","generate":"Use in generator","importTitle":"Import from link","importDescription":"Admins can import X/Twitter posts directly. Images are synced through Cloudbase/R2 before entering the catalog.","importProviderX":"X / Twitter","importPlaceholder":"Paste an X/Twitter post URL","importAction":"Import","importing":"Importing...","importSuccess":"Imported prompt case","importWarnings":"Import warnings","loadError":"Failed to load prompt cases"}}}`

const defaultDashboardShellConfig = `{"zh":{"labels":{"balance":"余额","available":"可用","apiKeys":"API Keys","active":"活跃","todayRequests":"今日请求","total":"总计","todayCost":"今日成本","actual":"实际","standard":"标准","todayTokens":"今日 Token","totalTokens":"总 Token","input":"输入","output":"输出","cacheWrite":"缓存写入","cacheRead":"缓存读取","performance":"性能","avgResponse":"平均响应","averageTime":"平均耗时","platformBreakdown":"平台拆分","platformCount":"{count} 个平台","platformOther":"其他","requests":"请求","tokens":"Token","platformQuotaTitle":"额度","platformQuotaDaily":"每日","platformQuotaWeekly":"每周","platformQuotaMonthly":"每月","platformQuotaDisabled":"已禁用","platformQuotaResetsAt":"重置于 {time}","recentUsage":"最近使用","last7Days":"最近 7 天","noUsageRecords":"暂无使用记录","startUsingApi":"开始使用 API 后会在这里显示最近请求。","viewAllUsage":"查看全部用量","timeRange":"时间范围","refresh":"刷新","granularity":"粒度","day":"天","hour":"小时","modelDistribution":"模型分布","noDataAvailable":"暂无数据","model":"模型","quickActions":"快捷操作","createApiKey":"创建 API Key","generateNewKey":"生成新的访问密钥","viewUsage":"查看用量","checkDetailedLogs":"查看详细请求日志","redeemCode":"兑换码","addBalanceWithCode":"使用兑换码增加余额"}},"en":{"labels":{"balance":"Balance","available":"Available","apiKeys":"API Keys","active":"active","todayRequests":"Today requests","total":"Total","todayCost":"Today cost","actual":"actual","standard":"standard","todayTokens":"Today tokens","totalTokens":"Total tokens","input":"Input","output":"Output","cacheWrite":"Cache write","cacheRead":"Cache read","performance":"Performance","avgResponse":"Average response","averageTime":"Average time","platformBreakdown":"Platform breakdown","platformCount":"{count} platforms","platformOther":"Other","requests":"Requests","tokens":"Tokens","platformQuotaTitle":"Quota","platformQuotaDaily":"Daily","platformQuotaWeekly":"Weekly","platformQuotaMonthly":"Monthly","platformQuotaDisabled":"Disabled","platformQuotaResetsAt":"Resets at {time}","recentUsage":"Recent usage","last7Days":"Last 7 days","noUsageRecords":"No usage records","startUsingApi":"Recent requests will appear here after you start using the API.","viewAllUsage":"View all usage","timeRange":"Time range","refresh":"Refresh","granularity":"Granularity","day":"Day","hour":"Hour","modelDistribution":"Model distribution","noDataAvailable":"No data available","model":"Model","quickActions":"Quick actions","createApiKey":"Create API key","generateNewKey":"Generate a new access key","viewUsage":"View usage","checkDetailedLogs":"Check detailed request logs","redeemCode":"Redeem code","addBalanceWithCode":"Add balance with a code"}}}`

const defaultPricingShellConfig = `{"zh":{"button":{"title":"去购买"},"groups":[{"name":"one-time","title":"充值包"},{"name":"subscription","title":"订阅包"}],"labels":{"prompts":"提示词案例","eyebrow":"价格","title":"价格与套餐","description":"浏览由 Cloudbase 统一配置的充值包和订阅套餐，选择后进入统一支付流程。","catalogStatus":"目录状态","rechargeProducts":"充值包","subscriptionPlans":"订阅包","recharge":"充值包","subscription":"订阅包","buy":"去购买","rechargeCta":"购买充值包","subscriptionCta":"购买订阅包","loadFailed":"价格目录加载失败，请稍后重试。","emptyRecharge":"暂未配置充值包。","emptyPlans":"暂未配置订阅包。","recommended":"推荐","creditedBalance":"到账余额","rate":"倍率","quota":"额度","unlimited":"不限","day":"天","days":"天","month":"月"}},"en":{"button":{"title":"Buy"},"groups":[{"name":"one-time","title":"Recharge"},{"name":"subscription","title":"Subscription"}],"labels":{"prompts":"Prompt cases","eyebrow":"Pricing","title":"Pricing","description":"Browse recharge products and subscription plans configured by Cloudbase, then continue to the unified checkout flow.","catalogStatus":"Catalog status","rechargeProducts":"Recharge products","subscriptionPlans":"Subscription plans","recharge":"Recharge","subscription":"Subscription","buy":"Buy","rechargeCta":"Buy balance","subscriptionCta":"Buy subscription","loadFailed":"Failed to load the pricing catalog. Please try again later.","emptyRecharge":"No recharge products are configured yet.","emptyPlans":"No subscription plans are configured yet.","recommended":"Recommended","creditedBalance":"Credited balance","rate":"Rate","quota":"Quota","unlimited":"Unlimited","day":"day","days":"days","month":"month"}}}`

const defaultPaymentShellConfig = `{"zh":{"labels":{"tabTopUp":"充值","tabSubscribe":"订阅","rechargeAccount":"充值账户","currentBalance":"当前余额","notAvailable":"支付暂不可用","noRechargeProducts":"暂未配置充值商品","rechargeProductRecommended":"推荐","rechargeProductCreditLine":"到账 ${amount} 余额","rechargeProductCta":"选择此充值包","paymentMethod":"支付方式","methodAlipay":"支付宝","methodWxpay":"微信支付","methodStripe":"Stripe","methodAirwallex":"Airwallex","success":"支付成功","subscriptionSuccess":"订阅成功","orderId":"订单 ID","orderNo":"订单编号","amount":"金额","payAmount":"实付","confirm":"确认","cancelled":"订单已取消","cancelledDesc":"您已取消本次支付","expired":"订单已过期","expiredDesc":"订单已超时，请重新创建订单","scanAlipay":"支付宝扫码支付","scanAlipayHint":"请使用手机打开支付宝，扫描二维码完成支付","scanWxpay":"微信扫码支付","scanWxpayHint":"请使用手机打开微信，扫描二维码完成支付","scanToPay":"请扫码支付","openPayWindow":"重新打开支付页面","expiresIn":"剩余支付时间","waitingPayment":"等待支付...","cancelOrder":"取消订单","payInNewWindowHint":"支付页面已在新窗口打开，请在新窗口中完成支付后返回此页面","paymentAmount":"支付金额","fee":"手续费","actualPay":"实付","creditedBalance":"到账余额","rechargeRatePreview":"充值汇率：1 CNY = {usd} USD 余额","processing":"处理中...","createOrder":"创建订单","cancel":"取消","selectAmountFirst":"请选择充值商品","amountNoMethod":"当前充值商品没有可用支付方式","amountTooLow":"金额不能低于 {min}","amountTooHigh":"金额不能高于 {max}","amountLabel":"金额","noPlans":"暂无订阅套餐","activeSubscription":"当前订阅","selectPlan":"选择套餐","groupFallback":"分组 #{id}","daysRemaining":"剩余 {days} 天","noExpiration":"永久有效","activeStatus":"生效中","rate":"倍率","dailyLimit":"日限额","weeklyLimit":"周限额","monthlyLimit":"月限额","quota":"额度","unlimited":"不限","models":"模型","subscribeNow":"立即开通","renewNow":"续费","perMonth":"月","perYear":"年","days":"天","baseAmount":"充值金额","creditedAmount":"到账金额","status":"状态","failed":"支付失败","processingHint":"支付结果仍在确认中，页面会自动刷新。","backToRecharge":"返回充值","viewOrders":"查看订单","stripeLoadFailed":"支付组件加载失败，请刷新页面重试","stripeMissingParams":"缺少订单ID或支付密钥","stripeNotConfigured":"Stripe 未配置","stripePay":"立即支付","stripeSuccessProcessing":"支付成功，正在处理订单...","airwallexLoadFailed":"Airwallex 支付组件加载失败","airwallexMissingParams":"缺少 Airwallex 支付参数","close":"关闭","stripePopupRedirecting":"正在跳转到支付页面...","stripePopupLoadingQr":"正在获取微信支付二维码...","stripePopupTimeout":"等待支付凭证超时，请重试","payInNewWindow":"支付页面已在新窗口打开","wechatPaymentCallbackTitle":"正在恢复微信支付","wechatPaymentCallbackProcessing":"正在恢复微信支付...","wechatPaymentCallbackBackToPayment":"返回支付页","wechatPaymentCallbackMissingResumeToken":"微信支付回调缺少恢复令牌。","tooManyPending":"待支付订单过多，请完成或取消后再试（最多 {max} 个）","cancelRateLimited":"取消订单过于频繁，请稍后再试","mobilePaymentFallbackToQr":"当前环境无法直接唤起支付，已切换为扫码支付","refresh":"刷新","all":"全部","pending":"待支付","completed":"已完成","refunded":"已退款","statusPending":"待支付","statusPaid":"已支付","statusRecharging":"充值中","statusCompleted":"已完成","statusExpired":"已过期","statusCancelled":"已取消","statusFailed":"失败","statusRefundRequested":"已申请退款","statusRefunding":"退款中","statusRefunded":"已退款","statusPartiallyRefunded":"部分退款","statusRefundFailed":"退款失败","actions":"操作","requestRefund":"申请退款","confirmCancel":"确定要取消这个订单吗？","refundReason":"退款原因","refundReasonPlaceholder":"请填写退款原因","cancelSuccess":"订单已取消","refundSuccess":"退款申请已提交","errorFallback":"操作失败","createdAt":"创建时间","subscriptionNoActive":"暂无有效订阅","subscriptionNoActiveDesc":"您没有任何有效订阅。请联系管理员获取订阅。","subscriptionExpires":"到期时间","subscriptionNoExpiration":"无到期时间","subscriptionStatusActive":"有效","subscriptionStatusExpired":"已过期","subscriptionStatusRevoked":"已撤销","subscriptionDaily":"每日","subscriptionWeekly":"每周","subscriptionMonthly":"每月","subscriptionUnlimited":"无限制","subscriptionUnlimitedDesc":"该订阅无用量限制","subscriptionDaysRemaining":"剩余 {days} 天","subscriptionResetIn":"{time} 后重置","subscriptionQuotaEndsIn":"额度将在 {time} 后重置","subscriptionWindowNotActive":"等待首次使用","subscriptionToday":"今天","subscriptionTomorrow":"明天","subscriptionFailedToLoad":"加载订阅失败"}},"en":{"labels":{"tabTopUp":"Top Up","tabSubscribe":"Subscribe","rechargeAccount":"Recharge Account","currentBalance":"Current Balance","notAvailable":"Payment Not Available","noRechargeProducts":"No recharge products configured","rechargeProductRecommended":"Recommended","rechargeProductCreditLine":"Credited ${amount} balance","rechargeProductCta":"Select this package","paymentMethod":"Payment Method","methodAlipay":"Alipay","methodWxpay":"WeChat Pay","methodStripe":"Stripe","methodAirwallex":"Airwallex","success":"Payment Successful","subscriptionSuccess":"Subscription Successful","orderId":"Order ID","orderNo":"Order No.","amount":"Amount","payAmount":"Paid","confirm":"Confirm","cancelled":"Order Cancelled","cancelledDesc":"You have cancelled this payment.","expired":"Order Expired","expiredDesc":"This order has expired. Please create a new one.","scanAlipay":"Alipay QR Payment","scanAlipayHint":"Open Alipay on your phone and scan the QR code to pay","scanWxpay":"WeChat QR Payment","scanWxpayHint":"Open WeChat on your phone and scan the QR code to pay","scanToPay":"Scan to Pay","openPayWindow":"Reopen Payment Page","expiresIn":"Expires in","waitingPayment":"Waiting for payment...","cancelOrder":"Cancel Order","payInNewWindowHint":"The payment page has opened in a new window. Please complete the payment there and return to this page.","paymentAmount":"Payment Amount","fee":"Fee","actualPay":"Actual Payment","creditedBalance":"Credited Balance","rechargeRatePreview":"Recharge rate: 1 CNY = {usd} USD balance","processing":"Processing...","createOrder":"Create Order","cancel":"Cancel","selectAmountFirst":"Select a recharge product","amountNoMethod":"No payment method is available for this recharge product","amountTooLow":"Amount cannot be lower than {min}","amountTooHigh":"Amount cannot be higher than {max}","amountLabel":"Amount","noPlans":"No plans available","activeSubscription":"Active Subscription","selectPlan":"Select Plan","groupFallback":"Group #{id}","daysRemaining":"{days} days remaining","noExpiration":"No expiration","activeStatus":"Active","rate":"Rate","dailyLimit":"Daily Limit","weeklyLimit":"Weekly Limit","monthlyLimit":"Monthly Limit","quota":"Quota","unlimited":"Unlimited","models":"Models","subscribeNow":"Subscribe Now","renewNow":"Renew","perMonth":"month","perYear":"year","days":"days","baseAmount":"Base Amount","creditedAmount":"Credited Amount","status":"Status","failed":"Payment Failed","processingHint":"Payment confirmation is still pending. This page will refresh automatically.","backToRecharge":"Back to Recharge","viewOrders":"View Orders","stripeLoadFailed":"Failed to load payment component. Please refresh and try again.","stripeMissingParams":"Missing order ID or client secret","stripeNotConfigured":"Stripe is not configured","stripePay":"Pay Now","stripeSuccessProcessing":"Payment successful, processing your order...","airwallexLoadFailed":"Failed to load Airwallex checkout","airwallexMissingParams":"Missing Airwallex payment parameters","close":"Close","stripePopupRedirecting":"Redirecting to payment page...","stripePopupLoadingQr":"Loading WeChat Pay QR code...","stripePopupTimeout":"Timed out waiting for payment credentials, please retry","payInNewWindow":"Payment page opened in a new window","wechatPaymentCallbackTitle":"Resuming WeChat payment","wechatPaymentCallbackProcessing":"Resuming WeChat payment...","wechatPaymentCallbackBackToPayment":"Back to payment","wechatPaymentCallbackMissingResumeToken":"WeChat payment callback is missing a resume token.","tooManyPending":"Too many pending orders. Complete or cancel one first (max {max}).","cancelRateLimited":"Order cancellation is rate limited. Please try again later.","mobilePaymentFallbackToQr":"This environment cannot open the payment sheet directly, so QR payment is shown instead.","refresh":"Refresh","all":"All","pending":"Pending","completed":"Completed","refunded":"Refunded","statusPending":"Pending","statusPaid":"Paid","statusRecharging":"Recharging","statusCompleted":"Completed","statusExpired":"Expired","statusCancelled":"Cancelled","statusFailed":"Failed","statusRefundRequested":"Refund requested","statusRefunding":"Refunding","statusRefunded":"Refunded","statusPartiallyRefunded":"Partially refunded","statusRefundFailed":"Refund failed","actions":"Actions","requestRefund":"Request Refund","confirmCancel":"Are you sure you want to cancel this order?","refundReason":"Refund Reason","refundReasonPlaceholder":"Please enter the refund reason","cancelSuccess":"Order cancelled","refundSuccess":"Refund request submitted","errorFallback":"Operation failed","createdAt":"Created At","subscriptionNoActive":"No Active Subscriptions","subscriptionNoActiveDesc":"You don't have any active subscriptions. Contact administrator to get one.","subscriptionExpires":"Expires","subscriptionNoExpiration":"No expiration","subscriptionStatusActive":"Active","subscriptionStatusExpired":"Expired","subscriptionStatusRevoked":"Revoked","subscriptionDaily":"Daily","subscriptionWeekly":"Weekly","subscriptionMonthly":"Monthly","subscriptionUnlimited":"Unlimited","subscriptionUnlimitedDesc":"No usage limits on this subscription","subscriptionDaysRemaining":"{days} days remaining","subscriptionResetIn":"Resets in {time}","subscriptionQuotaEndsIn":"Quota resets in {time}","subscriptionWindowNotActive":"Awaiting first use","subscriptionToday":"Today","subscriptionTomorrow":"Tomorrow","subscriptionFailedToLoad":"Failed to load subscriptions"}}}`

const defaultCreditsShellConfig = `{"zh":{"labels":{"eyebrow":"余额","title":"余额","description":"前端和后端统一使用余额口径；后端 balance 字段仍是唯一账本字段。","purchase":"充值余额","orders":"订单记录","credits":"余额","cloudbaseBalance":"账本余额","conversion":"统一口径：1 余额单位 = 1 balance 账本单位。","balanceLabel":"账本余额：{balance}","actionsTitle":"余额操作","actionsDescription":"充值、订单、退款等流程均进入 Cloudbase 统一支付体系，最终写入同一份 balance 账本。","recharge":"去充值","viewOrders":"查看订单"},"actions":{"title":"余额操作","description":"充值、订单、退款等流程均进入 Cloudbase 统一支付体系，最终写入同一份 balance 账本。"},"buttons":{"recharge":"去充值","orders":"查看订单"},"conversion":"统一口径：1 余额单位 = 1 balance 账本单位。"},"en":{"labels":{"eyebrow":"Balance","title":"Balance","description":"The frontend and backend use the same balance terminology; the backend balance field remains the only ledger field.","purchase":"Recharge balance","orders":"Orders","credits":"Balance","cloudbaseBalance":"Ledger balance","conversion":"Unified unit: 1 balance unit = 1 ledger unit.","balanceLabel":"Ledger balance: {balance}","actionsTitle":"Balance actions","actionsDescription":"Recharge, orders, and refunds go through the unified Cloudbase payment flow and write to the same balance ledger.","recharge":"Recharge","viewOrders":"View orders"},"actions":{"title":"Balance actions","description":"Recharge, orders, and refunds go through the unified Cloudbase payment flow and write to the same balance ledger."},"buttons":{"recharge":"Recharge","orders":"View orders"},"conversion":"Unified unit: 1 balance unit = 1 ledger unit."}}`

const defaultHomeShellConfig = `{"zh":{"labels":{"viewDocs":"文档","dashboard":"控制台","login":"登录","primaryCta":"立即开始","secondaryCta":"浏览模型","heroBadge":"开发者首选","heroTitle":"AI 编码工作台","heroDescription":"无需管理多个订阅账号，一站式接入 Claude、GPT 等主流 AI 服务。","modelMatrixKicker":"模型矩阵","modelMatrixTitle":"一个工作台连接 Claude 与 GPT","modelMatrixDescription":"从后台目录读取模型族和能力标签，保持公开页面和实际售卖能力一致。","modelMatrixEmptyCard":"配置模型后会自动出现在这里。","modelMatrixEmptyPill":"即将上线","experienceKicker":"体验","experienceTitle":"更清晰的模型访问流程","experienceDescription":"把模型访问、支付、文档和案例目录统一在一个平台里。","whyChooseKicker":"为什么选择","whyChooseTitle":"面向日常 AI 工作","whyChooseDescription":"更克制的产品形态、更清晰的价格和更贴近日常编码的工作流。","footerDescription":"更简单的模型访问入口，提供清晰价格和日常 AI 辅助编码体验。","allRightsReserved":"保留所有权利。","termsLink":"服务条款","privacyLink":"隐私政策","navHome":"首页","navDocs":"文档","navModels":"模型","navExperience":"体验","footerProduct":"产品","footerCatalog":"目录","footerSupport":"支持","familyClaudeBadge":"Claude","familyGptBadge":"GPT","familyClaudeTagline":"偏重推理的编码模型","familyGptTagline":"快速迭代和智能体","familyClaudeDescription":"适合深度推理、架构设计和代码审查。","familyGptDescription":"适合功能开发、快速迭代和智能体工作流。","familyClaudeReasoning":"深度推理","familyClaudeArchitecture":"架构设计","familyClaudeReview":"代码审查","familyGptCoding":"代码生成","familyGptIteration":"快速迭代","familyGptAgents":"智能体"},"experienceCards":[{"key":"unified","icon":"server","iconClass":"bg-gradient-to-br from-sky-500 to-blue-600","title":"一个密钥统一接入","description":"统一域名和密钥格式，减少在不同模型和工具之间来回切换。"},{"key":"setup","icon":"key","iconClass":"bg-gradient-to-br from-indigo-500 to-violet-600","title":"配置更轻","description":"更贴近 CLI、IDE 与日常开发习惯，不把大量时间花在环境变量和接线细节上。"},{"key":"stability","icon":"sparkles","iconClass":"bg-gradient-to-br from-emerald-500 to-teal-600","title":"链路更稳","description":"通过账号池与路由能力降低单点限制带来的中断，让高频编码更连续。"},{"key":"billing","icon":"chart","iconClass":"bg-gradient-to-br from-fuchsia-500 to-purple-600","title":"计费更透明","description":"充值、订阅和后续用量都公开可见，个人和小团队更容易控成本。"}],"whyChooseCards":[{"key":"lowFriction","title":"少折腾配置","description":"把分散在多个模型入口和订阅账号里的接入复杂度压缩成统一体验。"},{"key":"transparent","title":"模型一眼看清","description":"首页直接展示主力模型家族，开发者在登录前就能判断是否适合自己的工作流。"},{"key":"routing","title":"更适合高频编码","description":"强调链路稳定性与编码工作流，而不是堆叠泛化功能。"},{"key":"team","title":"适配个人与小团队","description":"既适合独立开发者快速上手，也方便小团队统一入口和管理预算。"}]},"en":{"labels":{"viewDocs":"Docs","dashboard":"Dashboard","login":"Log in","primaryCta":"Start now","secondaryCta":"Browse models","heroBadge":"Developer First","heroTitle":"AI Coding Workspace","heroDescription":"Access Claude, GPT, and other core AI services in one place without managing multiple subscriptions.","modelMatrixKicker":"Model Matrix","modelMatrixTitle":"One workspace for Claude and GPT","modelMatrixDescription":"Browse configured model families and capabilities from the backend catalog.","modelMatrixEmptyCard":"Configured models will appear here automatically.","modelMatrixEmptyPill":"Coming soon","experienceKicker":"Experience","experienceTitle":"A cleaner model access flow","experienceDescription":"Keep model access, payments, docs, and catalog discovery in one platform.","whyChooseKicker":"Why choose us","whyChooseTitle":"Built for daily AI work","whyChooseDescription":"A more restrained product shape, clearer pricing, and workflows closer to day-to-day coding.","footerDescription":"A simpler entry point for model access, visible pricing, and day-to-day AI-assisted coding.","allRightsReserved":"All rights reserved.","termsLink":"Terms","privacyLink":"Privacy","navHome":"Home","navDocs":"Docs","navModels":"Models","navExperience":"Experience","footerProduct":"Product","footerCatalog":"Catalog","footerSupport":"Support","familyClaudeBadge":"Claude","familyGptBadge":"GPT","familyClaudeTagline":"Reasoning-first coding","familyGptTagline":"Fast iteration and agents","familyClaudeDescription":"Use Claude models for deep reasoning, architecture, and review-heavy work.","familyGptDescription":"Use GPT models for coding, iteration, and agentic workflows.","familyClaudeReasoning":"Deep reasoning","familyClaudeArchitecture":"Architecture","familyClaudeReview":"Code review","familyGptCoding":"Coding","familyGptIteration":"Iteration","familyGptAgents":"Agents"},"experienceCards":[{"key":"unified","icon":"server","iconClass":"bg-gradient-to-br from-sky-500 to-blue-600","title":"One key, unified access","description":"Use one consistent domain and key format instead of juggling multiple providers and setup flows."},{"key":"setup","icon":"key","iconClass":"bg-gradient-to-br from-indigo-500 to-violet-600","title":"Lower setup friction","description":"Designed to fit better with CLI tools, IDE plugins, and the development habits people already have."},{"key":"stability","icon":"sparkles","iconClass":"bg-gradient-to-br from-emerald-500 to-teal-600","title":"More stable routing","description":"Account-pool and routing capabilities help reduce interruptions caused by single-path limits."},{"key":"billing","icon":"chart","iconClass":"bg-gradient-to-br from-fuchsia-500 to-purple-600","title":"More transparent billing","description":"Recharge products, plans, and usage stay visible so developers can control spend."}],"whyChooseCards":[{"key":"lowFriction","title":"Less setup overhead","description":"Compress scattered model and provider setup into a more unified experience built for repeat coding use."},{"key":"transparent","title":"Models visible at a glance","description":"The homepage surfaces the core model families directly so developers can judge the fit before signing in."},{"key":"routing","title":"Focused on coding throughput","description":"The product emphasizes coding workflows and routing stability instead of loading the homepage with unrelated platform features."},{"key":"team","title":"Fits solo devs and small teams","description":"Simple enough for individual developers to adopt quickly, while still giving small teams a cleaner shared entry point."}]}}`
const defaultHomeBusinessShellConfig = `{"zh":{"labels":{"viewDocs":"文档","dashboard":"控制台","login":"登录","primaryCta":"进入能力中台","secondaryCta":"查看图片提示词","heroBadge":"业务能力首页","heroTitle":"面向业务场景的 AI 能力工作台","heroDescription":"可直接上手的业务能力，包括微信内容导出、热点内容采集、图片提示词管理和生图工作台，让用户快速找到并使用核心功能。","modelMatrixKicker":"业务能力","modelMatrixTitle":"把高频业务能力摆到首页","modelMatrixDescription":"围绕微信内容导出、热点内容采集、图片提示词管理和生图工作台展示高频入口。","modelMatrixEmptyCard":"配置业务能力后会自动出现在这里。","modelMatrixEmptyPill":"按配置启用","experienceKicker":"业务能力","experienceTitle":"核心能力直接可用","experienceDescription":"首页聚焦内容导出、热点采集、提示词管理和图片生成，让用户从入口直接进入工作流。","whyChooseKicker":"能力组织方式","whyChooseTitle":"围绕用户要完成的任务组织入口","whyChooseDescription":"把微信内容导出、热点内容采集、图片提示词管理和生图工作台放到清晰位置。","footerDescription":"聚焦微信内容导出、热点内容采集、图片提示词管理和生图工作台。","allRightsReserved":"保留所有权利。","termsLink":"服务条款","privacyLink":"隐私政策","navHome":"首页","navDocs":"文档","navModels":"提示词","navExperience":"能力","footerProduct":"首页入口","footerCatalog":"业务能力","footerSupport":"支持","familyClaudeBadge":"","familyGptBadge":"","familyClaudeTagline":"","familyGptTagline":"","familyClaudeDescription":"","familyGptDescription":"","familyClaudeReasoning":"","familyClaudeArchitecture":"","familyClaudeReview":"","familyGptCoding":"","familyGptIteration":"","familyGptAgents":""},"businessCards":[{"key":"wechat-export","badge":"Workflow","title":"微信导出","description":"沉淀公众号内容导出与整理能力，适合把文章资产回收到统一工作流里。","capabilityTags":["内容导出","素材整理","资产回收"],"path":"/wechat","pathLabel":"进入微信导出"},{"key":"hot-topics","badge":"Signal","title":"热点追踪","description":"围绕热点发现、筛选和后续处理，把高频内容观察任务做成稳定入口。","capabilityTags":["热点收集","线索筛选","内容处理"],"path":"/hot","pathLabel":"进入热点追踪"},{"key":"prompt-catalog","badge":"Library","title":"图片提示词","description":"把沉淀下来的图片提示词案例放到统一目录里，便于检索、复用和二次加工。","capabilityTags":["案例目录","检索复用","图像提示词"],"path":"/prompts","pathLabel":"进入提示词库"},{"key":"image-workspace","badge":"Workspace","title":"生图工作台","description":"以提示词工作流为中心组织图片生成前的整理、复制和后续生产衔接。","capabilityTags":["Prompt 工作流","生图准备","工作台"],"path":"/image-generator","pathLabel":"进入工作台"}],"experienceCards":[{"key":"wechat-export","icon":"server","iconClass":"bg-gradient-to-br from-sky-500 to-blue-600","title":"微信内容导出","description":"把公众号文章导出、整理和后续复用做成稳定入口。"},{"key":"catalog","icon":"key","iconClass":"bg-gradient-to-br from-indigo-500 to-violet-600","title":"图片提示词管理","description":"围绕图片提示词案例组织检索、复用和二次加工。"},{"key":"image-workspace","icon":"sparkles","iconClass":"bg-gradient-to-br from-emerald-500 to-teal-600","title":"生图工作台","description":"从提示词准备到生图任务创建，形成更直接的工作流。"}],"whyChooseCards":[{"key":"business-first","title":"业务入口更直接","description":"用户进入首页后可以快速找到微信导出、热点、提示词和生图能力。"},{"key":"hot-content","title":"热点内容可持续采集","description":"热点线索进入统一入口，便于筛选、跟进和转化为内容选题。"},{"key":"reuse","title":"提示词与内容资产可复用","description":"把图片提示词、导出内容和热点线索组织成可持续复用的业务资产。"},{"key":"workflow","title":"形成工作流闭环","description":"从内容导出、热点发现到提示词管理、生图准备，首页直接表达完整业务链路。"}]},"en":{"labels":{"viewDocs":"Docs","dashboard":"Dashboard","login":"Log in","primaryCta":"Open capabilities","secondaryCta":"Browse prompt cases","heroBadge":"Business capability home","heroTitle":"An AI workspace organized around business capabilities","heroDescription":"Ready-to-use business capabilities, including WeChat content export, hot content collection, image prompt management, and an image generation workspace, so users can quickly find and use the core features.","modelMatrixKicker":"Capabilities","modelMatrixTitle":"Put core workflows on the homepage","modelMatrixDescription":"Show high-frequency entry points for WeChat content export, hot content collection, image prompt management, and image generation.","modelMatrixEmptyCard":"Configured business capabilities will appear here automatically.","modelMatrixEmptyPill":"Configurable","experienceKicker":"Business capabilities","experienceTitle":"Core workflows ready to use","experienceDescription":"The homepage focuses on content export, hot content collection, prompt management, and image generation so users can enter the workflow directly.","whyChooseKicker":"Information architecture","whyChooseTitle":"Organized around tasks users want to complete","whyChooseDescription":"Put WeChat content export, hot content collection, image prompt management, and the image generation workspace in clear, direct entry points.","footerDescription":"Focused on WeChat content export, hot content collection, image prompt management, and the image generation workspace.","allRightsReserved":"All rights reserved.","termsLink":"Terms","privacyLink":"Privacy","navHome":"Home","navDocs":"Docs","navModels":"Prompts","navExperience":"Capabilities","footerProduct":"Entry points","footerCatalog":"Workflows","footerSupport":"Support","familyClaudeBadge":"","familyGptBadge":"","familyClaudeTagline":"","familyGptTagline":"","familyClaudeDescription":"","familyGptDescription":"","familyClaudeReasoning":"","familyClaudeArchitecture":"","familyClaudeReview":"","familyGptCoding":"","familyGptIteration":"","familyGptAgents":""},"businessCards":[{"key":"wechat-export","badge":"Workflow","title":"WeChat Export","description":"Turn WeChat export and article recovery into a stable workflow entry instead of an ad hoc operation.","capabilityTags":["Content export","Asset recovery","Workflow"],"path":"/wechat","pathLabel":"Open WeChat export"},{"key":"hot-topics","badge":"Signal","title":"Hot Topic Tracking","description":"Package hot-topic discovery, filtering, and follow-up processing into a stable workflow entry.","capabilityTags":["Signal collection","Trend tracking","Content workflow"],"path":"/hot","pathLabel":"Open hot topics"},{"key":"prompt-catalog","badge":"Library","title":"Image Prompt Cases","description":"Keep image prompt cases in a searchable catalog so teams can reuse and refine proven material.","capabilityTags":["Prompt library","Search","Reuse"],"path":"/prompts","pathLabel":"Open prompt catalog"},{"key":"image-workspace","badge":"Workspace","title":"Image Workspace","description":"Center the image workflow around prompt preparation and handoff instead of exposing only the platform plumbing.","capabilityTags":["Prompt workflow","Image prep","Workspace"],"path":"/image-generator","pathLabel":"Open workspace"}],"experienceCards":[{"key":"wechat-export","icon":"server","iconClass":"bg-gradient-to-br from-sky-500 to-blue-600","title":"WeChat content export","description":"Make article export, organization, and reuse a stable entry point."},{"key":"catalog","icon":"key","iconClass":"bg-gradient-to-br from-indigo-500 to-violet-600","title":"Image prompt management","description":"Organize image prompt cases for search, reuse, and refinement."},{"key":"image-workspace","icon":"sparkles","iconClass":"bg-gradient-to-br from-emerald-500 to-teal-600","title":"Image generation workspace","description":"Move directly from prompt preparation to image task creation."}],"whyChooseCards":[{"key":"business-first","title":"More direct capability entry points","description":"Users can quickly find WeChat export, hot content, prompts, and image generation from the homepage."},{"key":"hot-content","title":"Continuous hot content collection","description":"Hot signals enter one workflow for filtering, follow-up, and topic development."},{"key":"reuse","title":"Reusable prompts and content assets","description":"Turn prompt cases, exported content, and hot-topic findings into assets that can be searched, refined, and reused."},{"key":"workflow","title":"A clearer workflow loop","description":"Move from content export and topic discovery into prompt management and image preparation with a direct capability narrative."}]}}`

const defaultModelPlazaShellConfig = `{"zh":{"labels":{"viewDocs":"文档","dashboard":"控制台","login":"登录","badge":"模型广场","title":"公开模型目录","description":"从后台直接配置并公开展示可售模型卡片。适合做模型能力说明、价格展示和统一入口。","emptyTitle":"模型广场暂未配置","emptyDescription":"管理员完成模型广场配置后，这里会展示公开模型卡片。","quickFind":"快速查找","searchLabel":"搜索模型广场","searchPlaceholder":"搜索模型、能力或标签","groupsTitle":"平台分组","currentSearch":"当前搜索：{query}","browseHint":"按平台分组浏览公开模型卡片。","results":"结果","emptyFilteredTitle":"没有匹配的模型卡片","emptyFilteredDescription":"试试切换分组，或者换一个更宽松的关键词搜索。","copyModelIds":"复制模型 ID","modelIdsCopied":"模型 ID 已复制","inputPrice":"输入价格","outputPrice":"输出价格","cacheReadPrice":"缓存读取价格","cacheWritePrice":"缓存创建价格","modelIdsConfigured":"已配置模型 ID","groupAll":"全部模型","groupOther":"其他"}},"en":{"labels":{"viewDocs":"Docs","dashboard":"Dashboard","login":"Log in","badge":"Model Plaza","title":"Public Model Catalog","description":"Configure and publish model cards directly from the admin backend for capability overviews, pricing communication, and a unified entry point.","emptyTitle":"Model plaza is not configured yet","emptyDescription":"Once the admin configures model plaza items, public model cards will appear here.","quickFind":"Quick find","searchLabel":"Search model plaza","searchPlaceholder":"Search models, capabilities, or tags","groupsTitle":"Groups","currentSearch":"Current search: {query}","browseHint":"Browse public model cards by provider group.","results":"Results","emptyFilteredTitle":"No matching model cards","emptyFilteredDescription":"Try another group or broaden the search terms.","copyModelIds":"Copy model IDs","modelIdsCopied":"Model IDs copied","inputPrice":"Input price","outputPrice":"Output price","cacheReadPrice":"Cache read price","cacheWritePrice":"Cache write price","modelIdsConfigured":"Model IDs configured","groupAll":"All models","groupOther":"Other"}}}`

const defaultDocsShellConfig = `{"zh":{"labels":{"title":"文档","dashboard":"控制台","login":"登录","searchPlaceholder":"搜索文档","noData":"没有结果"}},"en":{"labels":{"title":"Docs","dashboard":"Dashboard","login":"Log in","searchPlaceholder":"Search docs","noData":"No results"}}}`

const defaultDocsContentBasePath = `{"zh":"/docs-content/","en":"/docs-content/en/"}`

const defaultLegalDocumentShellConfig = `{"zh":{"labels":{"login":"登录","agreementLabel":"登录条款","loadFailedTitle":"文档加载失败","loadFailedDescription":"请稍后刷新页面重试。","missingTitle":"文档不存在","missingDescription":"当前条款文档不存在或已被管理员移除。","updatedAt":"更新日期：{date}","emptyContent":"暂无正文内容"}},"en":{"labels":{"login":"Log in","agreementLabel":"Login agreement","loadFailedTitle":"Failed to load document","loadFailedDescription":"Please refresh and try again later.","missingTitle":"Document not found","missingDescription":"This agreement document does not exist or has been removed by an administrator.","updatedAt":"Updated: {date}","emptyContent":"No document content yet"}}}`

const defaultKeyUsageShellConfig = `{"zh":{"labels":{"apply":"应用","allRightsReserved":"保留所有权利。","avgDuration":"平均耗时","cacheCreationTokens":"缓存创建","cacheWriteTokens":"缓存写入","cacheReadTokens":"缓存读取","cost":"费用","dailyDetail":"每日明细","date":"日期","dateRange":"统计范围:","dateRange30d":"30 天","dateRange7d":"7 天","dateRange90d":"90 天","dateRangeCustom":"自定义","dateRangeToday":"今日","daysLeft":"({days} 天)","detailInfo":"详细信息","docs":"文档","enterApiKey":"请输入 API Key","expiresAt":"过期时间","inputTokens":"输入 Tokens","limit5h":"5 小时限额","limit7d":"7 天限额","limitDaily":"日限额","limitMonthly":"月限额","limitWeekly":"周限额","model":"模型","modelStats":"模型用量统计","noDailyUsage":"当前筛选范围内没有每日明细数据","outputTokens":"输出 Tokens","placeholder":"sk-ant-mirror-xxxxxxxxxxxx","privacyNote":"您的 Key 仅在浏览器本地处理，不会被存储","query":"查询","queryFailed":"查询失败","queryFailedRetry":"查询失败，请稍后重试","querySuccess":"查询成功","querying":"查询中...","quotaMode":"Key 限额模式","remainingQuota":"剩余额度","requests":"请求数","resetNow":"即将重置","rpmTpm":"RPM / TPM","subscriptionExpires":"订阅到期","subscriptionType":"订阅类型","subtitle":"输入您的 API Key 以查看实时消费金额与使用状态","title":"API Key 用量查询","todayCacheCreation":"今日缓存创建","todayCacheRead":"今日缓存读取","todayCost":"今日费用","todayExpires":"(今日到期)","todayInputTokens":"今日输入","todayOutputTokens":"今日输出","todayRequests":"今日请求","todayTokens":"今日 Tokens","tokenStats":"Token 统计","totalCacheCreation":"累计缓存创建","totalCacheRead":"累计缓存读取","totalCost":"累计费用","totalInputTokens":"累计输入","totalOutputTokens":"累计输出","totalQuota":"总额度","totalRequests":"累计请求","totalTokens":"总 Tokens","totalTokensLabel":"累计 Tokens","used":"已使用","usedQuota":"已用额度","walletBalance":"钱包余额"}},"en":{"labels":{"apply":"Apply","allRightsReserved":"All rights reserved.","avgDuration":"Avg Duration","cacheCreationTokens":"Cache Creation","cacheWriteTokens":"Cache Write","cacheReadTokens":"Cache Read","cost":"Cost","dailyDetail":"Daily Detail","date":"Date","dateRange":"Date Range:","dateRange30d":"30 Days","dateRange7d":"7 Days","dateRange90d":"90 Days","dateRangeCustom":"Custom","dateRangeToday":"Today","daysLeft":"({days} days)","detailInfo":"Detail Information","docs":"Docs","enterApiKey":"Please enter an API Key","expiresAt":"Expires At","inputTokens":"Input Tokens","limit5h":"5-Hour Limit","limit7d":"7-Day Limit","limitDaily":"Daily Limit","limitMonthly":"Monthly Limit","limitWeekly":"Weekly Limit","model":"Model","modelStats":"Model Usage Statistics","noDailyUsage":"No daily usage details in the current range","outputTokens":"Output Tokens","placeholder":"sk-ant-mirror-xxxxxxxxxxxx","privacyNote":"Your Key is processed locally in the browser and will not be stored","query":"Query","queryFailed":"Query failed","queryFailedRetry":"Query failed, please try again later","querySuccess":"Query successful","querying":"Querying...","quotaMode":"Key Quota Mode","remainingQuota":"Remaining Quota","requests":"Requests","resetNow":"Resetting soon","rpmTpm":"RPM / TPM","subscriptionExpires":"Subscription Expires","subscriptionType":"Subscription Type","subtitle":"Enter your API Key to view real-time spending and usage status","title":"API Key Usage","todayCacheCreation":"Today Cache Creation","todayCacheRead":"Today Cache Read","todayCost":"Today Cost","todayExpires":"(expires today)","todayInputTokens":"Today Input","todayOutputTokens":"Today Output","todayRequests":"Today Requests","todayTokens":"Today Tokens","tokenStats":"Token Statistics","totalCacheCreation":"Total Cache Creation","totalCacheRead":"Total Cache Read","totalCost":"Total Cost","totalInputTokens":"Total Input","totalOutputTokens":"Total Output","totalQuota":"Total Quota","totalRequests":"Total Requests","totalTokens":"Total Tokens","totalTokensLabel":"Total Tokens","used":"Used","usedQuota":"Used Quota","walletBalance":"Wallet Balance"}}}`

const defaultUsageShellConfig = `{"zh":{"labels":{"totalRequests":"总请求数","inSelectedRange":"选中范围内","totalTokens":"总 Tokens","in":"输入","out":"输出","totalCost":"总费用","actualCost":"实际费用","standardCost":"标准费用","avgDuration":"平均耗时","perRequest":"每次请求","apiKeyFilter":"API Key","allApiKeys":"全部密钥","timeRange":"时间范围","refresh":"刷新","reset":"重置","exportCsv":"导出 CSV","exporting":"导出中...","model":"模型","reasoningEffort":"推理强度","endpoint":"端点","type":"类型","billingMode":"计费模式","tokens":"Tokens","cost":"费用","firstToken":"首 Token","duration":"耗时","time":"时间","userAgent":"User Agent","noRecords":"暂无使用记录","rate":"倍率","original":"原始","billed":"计费","failedToLoad":"加载使用记录失败","noDataToExport":"没有可导出的数据","preparingExport":"正在准备导出...","exportSuccess":"导出成功","exportFailed":"导出失败"}},"en":{"labels":{"totalRequests":"Total Requests","inSelectedRange":"In selected range","totalTokens":"Total Tokens","in":"In","out":"Out","totalCost":"Total Cost","actualCost":"Actual Cost","standardCost":"Standard Cost","avgDuration":"Avg Duration","perRequest":"Per request","apiKeyFilter":"API Key","allApiKeys":"All API Keys","timeRange":"Time Range","refresh":"Refresh","reset":"Reset","exportCsv":"Export CSV","exporting":"Exporting...","model":"Model","reasoningEffort":"Reasoning Effort","endpoint":"Endpoint","type":"Type","billingMode":"Billing Mode","tokens":"Tokens","cost":"Cost","firstToken":"First Token","duration":"Duration","time":"Time","userAgent":"User Agent","noRecords":"No usage records","rate":"Rate","original":"Original","billed":"Billed","failedToLoad":"Failed to load usage records","noDataToExport":"No data to export","preparingExport":"Preparing export...","exportSuccess":"Export successful","exportFailed":"Export failed"}}}`

const defaultAPIGuideShellConfig = `{"zh":{"labels":{"badge":"API 调用","title":"网关调用说明","description":"查看当前 API Key 可用的协议、端点、鉴权方式和可复制的 curl 示例。","openTester":"打开在线测试","manageKeys":"管理 API Keys","baseUrl":"Base URL","currentKey":"当前密钥","noSelection":"未选择","selectKeyHint":"请选择一个 API Key","supportedEndpoints":"可用端点","noGroupAssigned":"未分配分组","noKeysTitle":"暂无 API Key","noKeysDescription":"创建 API Key 后即可查看可用网关端点和调用示例。","keySelector":"选择 API Key","keySelectorHint":"选择一个密钥后，将按其分组能力展示可用端点。","unassignedTitle":"该密钥未分配分组","unassignedDescription":"未分配分组的密钥无法确定可用协议和模型，请先在 API Keys 页面绑定分组。","keySummary":"密钥信息","groupName":"分组名称","platform":"平台","status":"状态","authHeaderTitle":"鉴权头","authHeaderDescription":"OpenAI/Anthropic 兼容端点使用 Bearer Token；Google 兼容端点使用 x-goog-api-key。","noEndpointVariants":"当前密钥没有可用端点。","stream":"开启流式输出","testThisVariant":"测试此端点","endpoint":"端点","protocol":"协议","defaultModel":"默认模型","headerMode":"鉴权方式","curlExample":"curl 示例","copyCurl":"复制 curl","copyCurlSuccess":"curl 已复制","defaultPrompt":"用一句话介绍 Cloudbase。","loadKeysFailed":"API Keys 加载失败"}},"en":{"labels":{"badge":"API Guide","title":"Gateway API Guide","description":"Review the protocols, endpoints, auth headers, and copy-ready curl examples available to the selected API key.","openTester":"Open Tester","manageKeys":"Manage API Keys","baseUrl":"Base URL","currentKey":"Current Key","noSelection":"No selection","selectKeyHint":"Select an API key","supportedEndpoints":"Supported Endpoints","noGroupAssigned":"No group assigned","noKeysTitle":"No API Keys","noKeysDescription":"Create an API key to view available gateway endpoints and examples.","keySelector":"API Key","keySelectorHint":"Choose a key to show endpoints enabled by its group.","unassignedTitle":"This key has no group","unassignedDescription":"Keys without a group cannot resolve available protocols or models. Assign a group from the API Keys page first.","keySummary":"Key Summary","groupName":"Group Name","platform":"Platform","status":"Status","authHeaderTitle":"Auth Header","authHeaderDescription":"OpenAI/Anthropic compatible endpoints use Bearer Token; Google compatible endpoints use x-goog-api-key.","noEndpointVariants":"No endpoint variants are available for this key.","stream":"Streaming","testThisVariant":"Test this endpoint","endpoint":"Endpoint","protocol":"Protocol","defaultModel":"Default Model","headerMode":"Header Mode","curlExample":"curl Example","copyCurl":"Copy curl","copyCurlSuccess":"curl copied","defaultPrompt":"Introduce Cloudbase in one sentence.","loadKeysFailed":"Failed to load API keys"}}}`

const defaultAPITestShellConfig = `{"zh":{"labels":{"badge":"Live Request","title":"调用测试","description":"直接在当前页面用你的 API Key 向网关发请求，方便确认路由、模型名、权限和上游响应是否正常。","openGuide":"查看调用说明","send":"发送测试请求","sending":"请求发送中...","keySelector":"选择 API Key","noSelection":"请选择一个 API Key","noGroupAssigned":"未分配分组","protocol":"调用协议","model":"模型名","loading":"加载中...","noOptionsFound":"没有可选项","stream":"开启流式输出","requestMeta":"请求信息","noKeysTitle":"还没有可用的 API Key","noKeysDescription":"先创建一个 API Key 并分配分组，才能在这里直接发起测试调用。","manageKeys":"管理 API 密钥","modelPlaceholder":"输入模型名","modelSearchPlaceholder":"搜索模型","modelHint":"默认会填入一个常用模型，你也可以手动改成自己的目标模型。","customModel":"自定义模型名","customModelHint":"这里会直接使用你输入的精确模型名发起请求。","customModelOption":"手动输入模型名","customModelOptionHint":"如果下拉里没有你要的模型，可以切换到手动输入。","prompt":"测试提示词","promptHint":"这里会直接作为请求体发送到网关，用来快速验证链路是否通畅。","promptPlaceholder":"输入你想发给模型的内容","streamHint":"开启后，请求会按 SSE 文本返回，原始响应区域会显示完整事件流。","unassignedTitle":"这个 API Key 不能直接测试","unassignedDescription":"因为它还没有分组。未分组 Key 会被网关拒绝，请先回到“API 密钥”页完成分配。","liveBillingTitle":"这里发出的是真实请求","liveBillingDescription":"调用测试不会走 mock，也不会免计费。请求成功后会按正常网关链路记录用量并参与余额、订阅或限额统计。","copyCurl":"复制 curl","platform":"分组平台","headerMode":"鉴权头","requestPreview":"请求体预览","copyRequest":"复制请求体","responsePreview":"响应结果","statusCode":"HTTP 状态","duration":"耗时","copyResponse":"复制响应","responseSummary":"响应摘要","usageRecordTitle":"用量记录同步","openUsage":"查看用量记录","rawResponse":"原始响应","responsePending":"点击“发送测试请求”后，这里会显示网关返回的原始响应和摘要。","notReady":"未就绪","copyCurlSuccess":"curl 命令已复制","copyRequestSuccess":"请求体已复制","copyResponseSuccess":"响应内容已复制","usageRecordSyncing":"请求已成功返回，正在同步对应的用量记录...","usageRecordFound":"已写入用量统计：{time} · ${cost} · {tokens} Tokens","usageRecordPending":"请求已经成功返回，但用量记录采用异步写入。如果你已经打开“用量统计”或仪表盘，请刷新页面后查看。","usageRecordIdle":"测试请求成功后，这里会提示它是否已经进入“用量统计”。","loadKeysFailed":"API Keys 加载失败","unknownError":"未知错误"}},"en":{"labels":{"badge":"Live Request","title":"API Test","description":"Send a real request through the gateway from this page to verify routing, model names, permissions, and upstream responses.","openGuide":"Open API Guide","send":"Send Test Request","sending":"Sending...","keySelector":"Select API Key","noSelection":"Select an API key","noGroupAssigned":"No group assigned","protocol":"Protocol","model":"Model","loading":"Loading...","noOptionsFound":"No options found","stream":"Enable streaming","requestMeta":"Request Details","noKeysTitle":"No API key available yet","noKeysDescription":"Create an API key and assign a group before running live gateway tests here.","manageKeys":"Manage API Keys","modelPlaceholder":"Enter a model name","modelSearchPlaceholder":"Search models","modelHint":"A common default model is prefilled, but you can replace it with the exact model you want to test.","customModel":"Custom Model","customModelHint":"The exact model name entered here will be sent to the gateway as-is.","customModelOption":"Enter model manually","customModelOptionHint":"Switch to manual input when the dropdown does not include the model you want.","prompt":"Prompt","promptHint":"This text is sent directly to the gateway so you can validate the full request path quickly.","promptPlaceholder":"Enter the content you want to send","streamHint":"When enabled, the raw response panel will show the SSE event stream instead of a compact JSON payload.","unassignedTitle":"This API key cannot be tested yet","unassignedDescription":"It has no group assignment. Ungrouped keys are rejected by the gateway until you assign one on the API Keys page.","liveBillingTitle":"Requests here are real","liveBillingDescription":"The API tester does not use mock responses and is not billing-free. Successful requests are recorded through the normal gateway path and count toward balance, subscription, and limit statistics.","copyCurl":"Copy curl","platform":"Group Platform","headerMode":"Auth Header","requestPreview":"Request Preview","copyRequest":"Copy Request Body","responsePreview":"Response","statusCode":"HTTP Status","duration":"Duration","copyResponse":"Copy Response","responseSummary":"Response Summary","usageRecordTitle":"Usage Sync","openUsage":"Open Usage Records","rawResponse":"Raw Response","responsePending":"Run a test request and the raw gateway response will appear here.","notReady":"Not ready","copyCurlSuccess":"curl command copied","copyRequestSuccess":"request body copied","copyResponseSuccess":"response copied","usageRecordSyncing":"The request has completed successfully. Checking for the corresponding usage record...","usageRecordFound":"Recorded in usage statistics: {time} · ${cost} · {tokens} tokens","usageRecordPending":"The request succeeded, but usage records are written asynchronously. If the Usage or Dashboard page is already open, refresh it to see the latest record.","usageRecordIdle":"After a successful test, this panel shows whether the request has appeared in usage statistics.","loadKeysFailed":"Failed to load API keys","unknownError":"Unknown error"}}}`

const defaultAvailableGroupsShellConfig = `{"zh":{"labels":{"title":"可用分组","description":"查看当前账号可见的模型分组、倍率、额度和订阅访问要求。","total":"总分组","public":"公开分组","memberOnly":"会员专属","searchPlaceholder":"搜索分组名称、描述、平台或订阅类型","emptyTitle":"没有可用分组","emptyDescription":"当前还没有可展示的分组。","emptyFilteredDescription":"没有匹配当前搜索条件的分组。","publicTitle":"公开分组","publicDescription":"这些分组对当前账号可直接使用。","memberTitle":"会员或专属分组","memberDescription":"这些分组需要订阅、权限或专属配置。","publicBadge":"公开","subscriptionBadge":"订阅","exclusiveBadge":"专属","standardBadge":"标准","imageEnabledBadge":"支持生图","rate":"倍率","quota":"额度","dailyLimit":"每日 ${amount}","weeklyLimit":"每周 ${amount}","monthlyLimit":"每月 ${amount}","unlimited":"不限"}},"en":{"labels":{"title":"Available Groups","description":"Review model groups visible to your account, including rates, quotas, and subscription access requirements.","total":"Total Groups","public":"Public Groups","memberOnly":"Member Only","searchPlaceholder":"Search group name, description, platform, or subscription type","emptyTitle":"No available groups","emptyDescription":"There are no groups to display yet.","emptyFilteredDescription":"No groups match the current search.","publicTitle":"Public Groups","publicDescription":"These groups are directly available to the current account.","memberTitle":"Member or Exclusive Groups","memberDescription":"These groups require a subscription, permission, or exclusive configuration.","publicBadge":"Public","subscriptionBadge":"Subscription","exclusiveBadge":"Exclusive","standardBadge":"Standard","imageEnabledBadge":"Image enabled","rate":"Rate","quota":"Quota","dailyLimit":"Daily ${amount}","weeklyLimit":"Weekly ${amount}","monthlyLimit":"Monthly ${amount}","unlimited":"Unlimited"}}}`

const defaultRedeemShellConfig = `{"zh":{"labels":{"currentBalance":"当前余额","concurrency":"并发数","requests":"请求","redeemCodeLabel":"兑换码","redeemCodePlaceholder":"请输入兑换码","redeemCodeHint":"兑换码支持大写字母和数字，可直接粘贴输入","redeemButton":"兑换","redeeming":"兑换中...","redeemSuccess":"兑换成功！","redeemFailed":"兑换失败","added":"已添加","concurrentRequests":"并发请求","subscriptionAssigned":"订阅已分配","subscriptionDays":"{days} 天","newBalance":"新余额","newConcurrency":"新并发数","aboutCodes":"关于兑换码","codeRule1":"每个兑换码只能使用一次","codeRule2":"兑换码可以增加余额、并发数或试用权限","codeRule3":"如有兑换问题，请联系客服","codeRule4":"余额和并发数即时更新","recentActivity":"最近活动","historyWillAppear":"您的兑换历史将显示在这里","adminAdjustment":"管理员调整","balanceAddedRedeem":"余额充值（兑换）","balanceAddedAffiliate":"余额充值（返利转入）","balanceAddedAdmin":"余额充值（管理员）","balanceDeductedAdmin":"余额扣除（管理员）","concurrencyAddedRedeem":"并发增加（兑换）","concurrencyAddedAdmin":"并发增加（管理员）","concurrencyReducedAdmin":"并发减少（管理员）","days":"天","pleaseEnterCode":"请输入兑换码","subscriptionRefreshFailed":"兑换成功，但订阅状态刷新失败。","codeRedeemSuccess":"兑换成功！","failedToRedeem":"兑换失败，请检查兑换码后重试。","unknown":"未知"}},"en":{"labels":{"currentBalance":"Current Balance","concurrency":"Concurrency","requests":"requests","redeemCodeLabel":"Redeem Code","redeemCodePlaceholder":"Enter your redeem code","redeemCodeHint":"Redeem codes use uppercase letters and numbers and can be pasted directly","redeemButton":"Redeem Code","redeeming":"Redeeming...","redeemSuccess":"Code Redeemed Successfully!","redeemFailed":"Redemption Failed","added":"Added","concurrentRequests":"concurrent requests","subscriptionAssigned":"Subscription Assigned","subscriptionDays":"{days} days","newBalance":"New Balance","newConcurrency":"New Concurrency","aboutCodes":"About Redeem Codes","codeRule1":"Each code can only be used once","codeRule2":"Codes may add balance, increase concurrency, or grant trial access","codeRule3":"Contact support if you have issues redeeming a code","codeRule4":"Balance and concurrency updates are immediate","recentActivity":"Recent Activity","historyWillAppear":"Your redemption history will appear here","adminAdjustment":"Admin Adjustment","balanceAddedRedeem":"Balance Added (Redeem)","balanceAddedAffiliate":"Balance Added (Affiliate Transfer)","balanceAddedAdmin":"Balance Added (Admin)","balanceDeductedAdmin":"Balance Deducted (Admin)","concurrencyAddedRedeem":"Concurrency Added (Redeem)","concurrencyAddedAdmin":"Concurrency Added (Admin)","concurrencyReducedAdmin":"Concurrency Reduced (Admin)","days":" days","pleaseEnterCode":"Please enter a redeem code","subscriptionRefreshFailed":"Redeemed successfully, but failed to refresh subscription status.","codeRedeemSuccess":"Code redeemed successfully!","failedToRedeem":"Failed to redeem code. Please check the code and try again.","unknown":"Unknown"}}}`

const defaultAffiliateShellConfig = `{"zh":{"labels":{"rebateRate":"我的返利比例","rebateRateHint":"被邀请用户每次充值后你可获得的返利比例","invitedUsers":"邀请人数","availableQuota":"可转返利额度","totalQuota":"历史返利额度","frozenQuota":"冻结中","title":"邀请中心","description":"统一管理邀请码、邀请记录、返利流水与返利转余额。","yourCode":"我的邀请码","copyCode":"复制邀请码","inviteLink":"邀请链接","copyLink":"复制链接","tipsTitle":"使用说明","tipShare":"将邀请码或邀请链接分享给新用户。","tipRebate":"被邀请用户充值后，你可获得 {rate} 的返利额度。","tipTransfer":"返利额度可随时转入账户余额。","tipFreeze":"新产生的返利需要经过冻结期后才能提现。","transferTitle":"返利额度转余额","transferDescription":"将当前可用返利额度一键转入账户余额","transferButton":"转入余额","transferring":"转入中...","transferEmpty":"当前没有可转入额度","transferSuccess":"已转入余额：{amount}","inviteesTitle":"已邀请用户","inviteesEmpty":"暂无邀请记录","emailColumn":"邮箱","usernameColumn":"用户名","rebateColumn":"返利明细","joinedAtColumn":"注册时间","rebatesTitle":"返利记录","rebatesEmpty":"暂无返利记录","inviteeColumn":"被邀请用户","orderAmountColumn":"充值金额","payAmountColumn":"支付金额","rebateAmountColumn":"返利金额","paymentTypeColumn":"支付方式","orderStatusColumn":"订单状态","createdAtColumn":"返利时间","transfersTitle":"转余额记录","transfersEmpty":"暂无转余额记录","amountColumn":"转入金额","balanceAfterColumn":"转入后余额","availableQuotaAfterColumn":"转入后可提返利","frozenQuotaAfterColumn":"转入后冻结返利","historyQuotaAfterColumn":"转入后历史返利","transferredAtColumn":"转入时间","codeCopied":"邀请码已复制","linkCopied":"邀请链接已复制","loadFailed":"加载邀请返利数据失败","transferFailed":"转入余额失败"}},"en":{"labels":{"rebateRate":"My Rebate Rate","rebateRateHint":"What you earn each time an invitee recharges","invitedUsers":"Invited Users","availableQuota":"Available Rebate Quota","totalQuota":"Historical Rebate Quota","frozenQuota":"Frozen","title":"Invite Center","description":"Manage invite codes, invitees, rebate records, and rebate transfers in one place.","yourCode":"Your Affiliate Code","copyCode":"Copy Code","inviteLink":"Invite Link","copyLink":"Copy Link","tipsTitle":"How It Works","tipShare":"Share your affiliate code or invite link with new users.","tipRebate":"When invitees recharge, you receive {rate} of the recharge as rebate quota.","tipTransfer":"Transfer rebate quota to balance at any time.","tipFreeze":"Newly earned rebates may have a waiting period before they can be transferred.","transferTitle":"Transfer Rebate Quota","transferDescription":"Move available rebate quota into your account balance","transferButton":"Transfer to Balance","transferring":"Transferring...","transferEmpty":"No available rebate quota","transferSuccess":"{amount} has been transferred to your balance","inviteesTitle":"Invited Users","inviteesEmpty":"No invited users yet","emailColumn":"Email","usernameColumn":"Username","rebateColumn":"Rebate","joinedAtColumn":"Joined At","rebatesTitle":"Rebate Records","rebatesEmpty":"No rebate records yet","inviteeColumn":"Invitee","orderAmountColumn":"Top-up Amount","payAmountColumn":"Paid Amount","rebateAmountColumn":"Rebate Amount","paymentTypeColumn":"Payment Method","orderStatusColumn":"Order Status","createdAtColumn":"Rebated At","transfersTitle":"Transfer Records","transfersEmpty":"No transfer records yet","amountColumn":"Transferred","balanceAfterColumn":"Balance After","availableQuotaAfterColumn":"Available Quota After","frozenQuotaAfterColumn":"Frozen Quota After","historyQuotaAfterColumn":"Historical Rebate After","transferredAtColumn":"Transferred At","codeCopied":"Affiliate code copied","linkCopied":"Invite link copied","loadFailed":"Failed to load affiliate data","transferFailed":"Failed to transfer affiliate quota"}}}`

const defaultAvailableChannelsShellConfig = `{"zh":{"labels":{"searchPlaceholder":"搜索渠道或模型...","refreshTitle":"刷新","noPricing":"未配置定价","noModels":"未配置模型","empty":"暂无可用渠道","loadError":"加载可用渠道失败","exclusive":"专属","exclusiveTooltip":"管理员授权给你的专属分组","public":"公开","publicTooltip":"对所有用户公开的分组","columns":{"name":"渠道名","description":"描述","platform":"平台","groups":"我可访问的分组","supportedModels":"支持模型"},"pricing":{"billingMode":"计费模式","billingModeImage":"按图片","billingModePerRequest":"按次","billingModeToken":"按 Token","cacheReadPrice":"缓存读取","cacheWritePrice":"缓存写入","imageOutputPrice":"图片输出","inputPrice":"输入","intervals":"阶梯定价","outputPrice":"输出","perRequestPrice":"每次请求","unitPerMillion":"/ 1M token","unitPerRequest":"/ 次"}}},"en":{"labels":{"searchPlaceholder":"Search channels or models...","refreshTitle":"Refresh","noPricing":"Pricing not configured","noModels":"No models configured","empty":"No available channels","loadError":"Failed to load available channels","exclusive":"Exclusive","exclusiveTooltip":"Exclusive groups granted to you by an admin","public":"Public","publicTooltip":"Groups open to all users","columns":{"name":"Channel","description":"Description","platform":"Platform","groups":"Your Accessible Groups","supportedModels":"Supported Models"},"pricing":{"billingMode":"Billing Mode","billingModeImage":"Per Image","billingModePerRequest":"Per Request","billingModeToken":"Per Token","cacheReadPrice":"Cache Read","cacheWritePrice":"Cache Write","imageOutputPrice":"Image Output","inputPrice":"Input","intervals":"Tiered Pricing","outputPrice":"Output","perRequestPrice":"Per Request","unitPerMillion":"/ 1M tokens","unitPerRequest":"/ request"}}}}`

const defaultChannelStatusShellConfig = `{"zh":{"labels":{"refreshTitle":"刷新","detailTitle":"渠道详情","loadError":"加载渠道状态失败","detailLoadError":"加载渠道详情失败","latency":"延迟","ping":"端点 Ping","availabilityPrefix":"可用率","extraModelsCount":"+{n} 个模型","emptyTitle":"暂无可显示的渠道","emptyDescription":"管理员尚未配置可监控的渠道。","closeDetail":"关闭","windowTab":{"7d":"7 天","15d":"15 天","30d":"30 天"},"overall":{"operational":"OPERATIONAL","degraded":"DEGRADED"},"detailColumns":{"model":"模型","latestStatus":"最新状态","latestLatency":"最新延迟 (ms)","availability7d":"7 天可用率","availability15d":"15 天可用率","availability30d":"30 天可用率","avgLatency7d":"7 天平均延迟 (ms)"}}},"en":{"labels":{"refreshTitle":"Refresh","detailTitle":"Channel Detail","loadError":"Failed to load channel status","detailLoadError":"Failed to load channel detail","latency":"Latency","ping":"Endpoint Ping","availabilityPrefix":"Availability","extraModelsCount":"+{n} models","emptyTitle":"No channels available","emptyDescription":"No monitored channels have been configured yet.","closeDetail":"Close","windowTab":{"7d":"7 days","15d":"15 days","30d":"30 days"},"overall":{"operational":"OPERATIONAL","degraded":"DEGRADED"},"detailColumns":{"model":"Model","latestStatus":"Latest Status","latestLatency":"Latest Latency (ms)","availability7d":"7d Availability","availability15d":"15d Availability","availability30d":"30d Availability","avgLatency7d":"7d Avg Latency (ms)"}}}}`

const defaultCustomPageShellConfig = `{"zh":{"labels":{"tocTitle":"目录","tocToggle":"目录","notFoundTitle":"页面不存在","notFoundDesc":"该自定义页面不存在或已被删除。","notConfiguredTitle":"页面链接未配置","notConfiguredDesc":"该自定义页面的 URL 未正确配置。","openInNewTab":"新窗口打开","markdownNotFound":"页面不存在","markdownLoadFailed":"页面加载失败","copyCode":"复制","copyCodeSuccess":"已复制 ✓","copyCodeFailed":"失败"}},"en":{"labels":{"tocTitle":"Table of Contents","tocToggle":"Contents","notFoundTitle":"Page not found","notFoundDesc":"This custom page does not exist or has been removed.","notConfiguredTitle":"Page URL not configured","notConfiguredDesc":"The URL for this custom page has not been properly configured.","openInNewTab":"Open in new tab","markdownNotFound":"Page not found","markdownLoadFailed":"Failed to load page","copyCode":"Copy","copyCodeSuccess":"Copied ✓","copyCodeFailed":"Failed"}}}`

const defaultProfileShellConfig = `{"zh":{"labels":{"user":"用户","administrator":"管理员","accountBalance":"账户余额","concurrencyLimit":"并发额度","memberSince":"加入时间","basicsTitle":"基础资料","basicsDescription":"管理头像、昵称以及当前账号展示信息。","linkedProfileSources":"资料来源","linkedProfileSourcesDescription":"部分资料会从绑定的第三方登录方式同步。","contactSupport":"联系客服","changePassword":"修改密码","currentPassword":"当前密码","newPassword":"新密码","confirmNewPassword":"确认新密码","passwordHint":"密码至少需要 {count} 位字符","changingPassword":"修改中...","changePasswordButton":"修改密码","passwordsNotMatch":"两次输入的新密码不一致","passwordTooShort":"密码至少需要 {count} 位字符","passwordChangeSuccess":"密码修改成功","passwordChangeFailed":"密码修改失败","balanceNotifyTitle":"余额不足提醒","balanceNotifyDescription":"当账户余额低于阈值时发送邮件提醒","balanceNotifyEnabled":"启用余额不足提醒","balanceNotifyThreshold":"自定义提醒阈值","balanceNotifyThresholdHint":"留空使用系统默认值","balanceNotifySystemDefault":"系统默认值","balanceNotifyThresholdPlaceholder":"输入金额","balanceNotifyExtraEmails":"通知邮箱","balanceNotifyExtraEmailsHint":"必须添加并验证邮箱后，余额不足时才能收到提醒邮件","balanceNotifyCodePlaceholder":"6位验证码","balanceNotifyVerify":"验证","balanceNotifyResend":"重发","balanceNotifyUnverified":"未验证","balanceNotifyVerified":"已验证","balanceNotifyRemoveEmail":"移除","balanceNotifySendCode":"发送验证码","balanceNotifyEmailPlaceholder":"输入邮箱地址","balanceNotifyMaxEmailsReached":"已达到通知邮箱数量上限","balanceNotifyEmailDuplicate":"该邮箱已存在","balanceNotifyCodeSent":"验证码已发送","balanceNotifyVerifySuccess":"邮箱添加成功","balanceNotifyRemoveSuccess":"邮箱已移除","balanceNotifySaving":"保存中...","balanceNotifySave":"保存","balanceNotifyCancel":"取消","balanceNotifyAdd":"添加","balanceNotifySaved":"已保存","balanceNotifyError":"操作失败","avatarTitle":"资料头像","avatarDescription":"仅支持上传头像图片；静态图片会自动压缩到 20KB 以内后再保存。","avatarUploadHint":"上传图片时会自动压缩静态图片到 20KB 以内，GIF 需自行控制在 20KB 以内","avatarUploadAction":"上传图片","avatarUploadRequired":"请先上传头像图片","avatarReadFailed":"读取所选图片失败","avatarCompressFailed":"压缩所选图片失败","avatarCompressTooLarge":"无法将图片压缩到 20KB 以内，请换一张更小的图片","avatarInvalidType":"请选择图片文件","avatarGifTooLarge":"GIF 头像必须在 20KB 以内","avatarSaveSuccess":"头像已更新","avatarEmptyDeleteHint":"当前没有可删除的头像","avatarDeleteSuccess":"头像已删除","totpTitle":"两步验证","totpDescription":"使用身份验证器应用为账号增加额外保护","totpFeatureDisabled":"两步验证当前未启用","totpFeatureDisabledHint":"管理员尚未开启此功能","totpEnabled":"两步验证已启用","totpEnabledAt":"启用时间","totpDisable":"停用","totpNotEnabled":"两步验证未启用","totpNotEnabledHint":"启用后，登录时需要输入身份验证器中的动态验证码","totpEnable":"启用","providers":{"email":"邮箱","github":"GitHub","google":"Google"},"sourceAvatar":"头像当前来自 {providerName}","sourceUsername":"昵称当前来自 {providerName}"}},"en":{"labels":{"user":"User","administrator":"Administrator","accountBalance":"Account Balance","concurrencyLimit":"Concurrency Limit","memberSince":"Member Since","basicsTitle":"Basic Profile","basicsDescription":"Manage avatar, nickname, and account display information.","linkedProfileSources":"Profile Sources","linkedProfileSourcesDescription":"Some profile fields can be synced from connected sign-in providers.","contactSupport":"Contact Support","changePassword":"Change Password","currentPassword":"Current Password","newPassword":"New Password","confirmNewPassword":"Confirm New Password","passwordHint":"Password must be at least {count} characters long","changingPassword":"Changing...","changePasswordButton":"Change Password","passwordsNotMatch":"New passwords do not match","passwordTooShort":"Password must be at least {count} characters long","passwordChangeSuccess":"Password changed successfully","passwordChangeFailed":"Failed to change password","balanceNotifyTitle":"Balance Low Notification","balanceNotifyDescription":"Send email alert when account balance falls below threshold","balanceNotifyEnabled":"Enable Balance Low Notification","balanceNotifyThreshold":"Custom Threshold","balanceNotifyThresholdHint":"Leave empty to use system default","balanceNotifySystemDefault":"System Default","balanceNotifyThresholdPlaceholder":"Enter amount","balanceNotifyExtraEmails":"Notification Emails","balanceNotifyExtraEmailsHint":"You must add and verify an email address to receive low balance alerts","balanceNotifyCodePlaceholder":"6-digit code","balanceNotifyVerify":"Verify","balanceNotifyResend":"Resend","balanceNotifyUnverified":"Unverified","balanceNotifyVerified":"Verified","balanceNotifyRemoveEmail":"Remove","balanceNotifySendCode":"Send Code","balanceNotifyEmailPlaceholder":"Enter email address","balanceNotifyMaxEmailsReached":"Maximum number of notification emails reached","balanceNotifyEmailDuplicate":"This email already exists","balanceNotifyCodeSent":"Verification code sent","balanceNotifyVerifySuccess":"Email added successfully","balanceNotifyRemoveSuccess":"Email removed","balanceNotifySaving":"Saving...","balanceNotifySave":"Save","balanceNotifyCancel":"Cancel","balanceNotifyAdd":"Add","balanceNotifySaved":"Saved","balanceNotifyError":"Operation failed","avatarTitle":"Profile Avatar","avatarDescription":"Upload an avatar image. Static uploads are compressed to 20KB before saving.","avatarUploadHint":"Static uploads are compressed to 20KB when possible. GIF uploads must already be within 20KB.","avatarUploadAction":"Upload image","avatarUploadRequired":"Upload an avatar image first","avatarReadFailed":"Failed to read the selected image.","avatarCompressFailed":"Failed to compress the selected image.","avatarCompressTooLarge":"Unable to compress this image below 20KB. Try a smaller image.","avatarInvalidType":"Please choose an image file","avatarGifTooLarge":"GIF avatars must already be 20KB or smaller","avatarSaveSuccess":"Avatar updated","avatarEmptyDeleteHint":"Avatar is already empty","avatarDeleteSuccess":"Avatar removed","totpTitle":"Two-Factor Authentication","totpDescription":"Use an authenticator app to add extra protection to your account","totpFeatureDisabled":"Two-factor authentication is unavailable","totpFeatureDisabledHint":"This feature has not been enabled by an administrator","totpEnabled":"Two-factor authentication enabled","totpEnabledAt":"Enabled at","totpDisable":"Disable","totpNotEnabled":"Two-factor authentication is not enabled","totpNotEnabledHint":"After enabling it, sign-in requires a dynamic code from your authenticator app","totpEnable":"Enable","providers":{"email":"Email","github":"GitHub","google":"Google"},"sourceAvatar":"Avatar is currently synced from {providerName}","sourceUsername":"Nickname is currently synced from {providerName}"}}}`

const defaultAuthShellConfig = `{
  "zh": {
    "labels": {
      "welcomeBack": "欢迎回来",
      "signInToAccount": "登录您的账户",
      "emailLabel": "邮箱",
      "emailPlaceholder": "请输入邮箱",
      "emailVerifyBackToRegistration": "返回注册",
      "emailVerifyClickToResend": "点击重新发送验证码",
      "emailVerifyCodeHint": "请输入发送到您邮箱的6位验证码",
      "emailVerifyCodeLabel": "验证码",
      "emailVerifyCodeSentSuccess": "验证码已发送！请查收您的邮箱。",
      "emailVerifyDescriptionPrefix": "我们将发送验证码到",
      "emailVerifyResendCode": "重新发送验证码",
      "emailVerifyResendCountdown": "{countdown}秒后可重新发送",
      "emailVerifySessionExpiredDescription": "请返回注册页面重新开始。",
      "emailVerifySessionExpiredTitle": "会话已过期",
      "emailVerifySubmit": "验证并创建账户",
      "emailVerifyTitle": "验证您的邮箱",
      "emailVerifyVerifying": "验证中...",
      "passwordLabel": "密码",
      "passwordPlaceholder": "请输入密码",
      "forgotPassword": "忘记密码？",
      "forgotPasswordTitle": "重置密码",
      "forgotPasswordHint": "输入您的邮箱地址，我们将向您发送密码重置链接。",
      "resetPasswordTitle": "设置新密码",
      "resetPasswordHint": "请在下方输入您的新密码。",
      "resetPassword": "重置密码",
      "resettingPassword": "重置中...",
      "invalidResetLink": "无效的重置链接",
      "invalidResetLinkHint": "此密码重置链接无效或已过期。请重新请求一个新链接。",
      "requestNewResetLink": "请求新的重置链接",
      "signingIn": "登录中...",
      "signIn": "登录",
      "signInWithProvider": "使用 {providerName} 登录",
      "oauthAlternativeMethods": "或使用以下方式继续",
      "oauthCallbackCode": "授权码",
      "oauthCallbackHint": "如果页面未自动跳转，请返回登录页重试。",
      "oauthCallbackFullUrl": "完整URL",
      "oauthCallbackInvalidHint": "当前页面缺少有效的授权结果，请返回登录页重新发起快捷登录。",
      "oauthCallbackInvalidTitle": "无效的登录回调",
      "oauthCallbackPasswordOptionalHint": "{providerName} 登录可留空，稍后可在个人资料中补充设置密码。",
      "oauthCallbackRegistrationHint": "完成注册",
      "oauthCallbackRegistrationInvitationRequired": "该 {providerName} 账号尚未注册，站点已开启邀请码注册，请输入邀请码以完成注册。",
      "oauthCallbackState": "状态",
      "oauthCallbackSubmitRegistration": "完成注册",
      "oauthCallbackTitle": "OAuth 回调",
      "oauthFlowAvatarAlt": "{providerName} 头像",
      "oauthFlowBindCurrentAccount": "绑定当前账户",
      "oauthFlowBindCurrentAccountDescription": "将此次 {providerName} 登录绑定到当前浏览器已登录的账户。",
      "oauthFlowBindCurrentAccountTitle": "绑定当前账户",
      "oauthFlowBindExistingAccount": "绑定已有账户",
      "oauthFlowBindLoginHint": "登录一个已有账户以绑定此次 {providerName} 登录。",
      "oauthFlowBindSignInToExistingAccount": "将此次 {providerName} 登录绑定到已有账户。",
      "oauthFlowChooseAccountActionHint": "请选择绑定已有账户，或创建一个新账户。",
      "oauthFlowChooseHowToContinue": "选择后续操作",
      "oauthFlowCreateAccountHint": "请输入邮箱地址以创建账户并继续。",
      "oauthFlowCreateAccountTitle": "完成 {providerName} 账户注册",
      "oauthFlowCreateNewAccount": "创建新账户",
      "oauthFlowLogInAndBind": "登录并绑定",
      "oauthFlowProfileDetailsDescription": "选择是否将 {providerName} 的昵称或头像应用到当前账户。",
      "oauthFlowProfileDetailsTitle": "使用 {providerName} 资料",
      "oauthFlowReviewProfileBeforeContinue": "请先确认 {providerName} 资料后再继续。",
      "oauthFlowSignInThenBindDescription": "请先登录已有账户，再将此次 {providerName} 登录绑定到该账户。",
      "oauthFlowSuggestedEmail": "建议邮箱：{email}",
      "oauthFlowTotpHint": "请输入 {account} 的 6 位验证码，以完成此次 {providerName} 登录绑定。",
      "oauthFlowUseAvatar": "使用头像",
      "oauthFlowUseDifferentEmail": "使用其他邮箱",
      "oauthFlowUseDisplayName": "使用昵称",
      "oauthFlowVerifyAndContinue": "验证并继续",
      "oauthFlowYourAccount": "当前账户",
      "dontHaveAccount": "还没有账户？",
      "signUp": "注册",
      "createAccount": "创建账户",
      "signUpToStart": "注册以开始使用 {siteName}",
      "registrationDisabled": "注册功能暂时关闭，请联系管理员。",
      "createPasswordPlaceholder": "创建一个安全的密码",
      "newPassword": "新密码",
      "newPasswordPlaceholder": "输入新密码",
      "confirmPassword": "确认密码",
      "confirmPasswordPlaceholder": "再次输入新密码",
      "passwordHint": "至少 {count} 个字符",
      "invitationCodeLabel": "邀请码",
      "invitationCodePlaceholder": "请输入邀请码",
      "invitationCodeValid": "邀请码有效",
      "affiliateInvitationDetected": "已检测到邀请链接，可直接继续注册。",
      "sendCode": "发送验证码",
      "sendingCode": "发送中...",
      "sendResetLink": "发送重置链接",
      "sendingResetLink": "发送中...",
      "resendCountdown": "{countdown}s",
      "codeSentSuccess": "验证码已发送",
      "verificationCodeHint": "请输入邮箱收到的 6 位验证码",
      "promoCodeLabel": "优惠码",
      "optional": "可选",
      "promoCodePlaceholder": "请输入优惠码",
      "promoCodeValid": "优惠码有效，注册后将获得 {amount} 余额",
      "processing": "处理中...",
      "continue": "继续",
      "totpLoginTitle": "两步验证",
      "totpLoginHint": "请输入身份验证器应用中的 6 位验证码",
      "totpVerifying": "验证中...",
      "totpCancel": "取消",
      "resetEmailSent": "重置链接已发送",
      "resetEmailSentHint": "如果该邮箱已注册，您将很快收到密码重置链接。请检查您的收件箱和垃圾邮件文件夹。",
      "passwordResetSuccess": "密码重置成功",
      "passwordResetSuccessHint": "您的密码已重置。现在可以使用新密码登录。",
      "providerCallbackHint": "如果页面未自动跳转，请返回登录页重试。",
      "providerCallbackProcessing": "正在验证 {providerName} 登录信息，请稍候...",
      "providerCallbackTitle": "正在完成 {providerName} 登录",
      "providerCompleteRegistration": "完成注册",
      "providerCompletingRegistration": "正在完成注册...",
      "providerInvitationRequired": "该 {providerName} 账号尚未注册，站点已开启邀请码注册，请输入邀请码以完成注册。",
      "agreementAcceptAndContinue": "同意并继续",
      "agreementAcceptedDescription": "您已同意当前版本条款，可随时重新查看相关文档。",
      "agreementAcceptedTitle": "登录条款入口",
      "agreementCheckboxPrefix": "我已阅读并同意",
      "agreementRecent": "近期",
      "agreementReject": "拒绝",
      "agreementRelevantDocuments": "相关文档",
      "agreementReviewDescription": "您可以先输入账号信息；如果当前账号尚未确认最新条款，我们会在提交登录时提示确认。",
      "agreementReviewTitle": "继续登录前可能需要确认最新条款。",
      "agreementTermsUpdateTitle": "条款更新通知",
      "agreementUpdatedAt": "我们的服务条款已于 {date} 更新。在继续使用服务之前，请仔细阅读并同意以下条款。",
      "agreementViewAndAccept": "查看并同意",
      "agreementViewTerms": "查看条款",
      "alreadyHaveAccount": "已有账户？",
      "rememberedPassword": "想起密码了？",
      "backToLogin": "返回登录",
      "allRightsReserved": "保留所有权利。"
    }
  },
  "en": {
    "labels": {
      "welcomeBack": "Welcome Back",
      "signInToAccount": "Sign in to your account",
      "emailLabel": "Email",
      "emailPlaceholder": "Enter your email",
      "emailVerifyBackToRegistration": "Back to registration",
      "emailVerifyClickToResend": "Click to resend code",
      "emailVerifyCodeHint": "Enter the 6-digit code sent to your email",
      "emailVerifyCodeLabel": "Verification Code",
      "emailVerifyCodeSentSuccess": "Verification code sent! Please check your inbox.",
      "emailVerifyDescriptionPrefix": "We'll send a verification code to",
      "emailVerifyResendCode": "Resend verification code",
      "emailVerifyResendCountdown": "Resend code in {countdown}s",
      "emailVerifySessionExpiredDescription": "Please go back to the registration page and start again.",
      "emailVerifySessionExpiredTitle": "Session expired",
      "emailVerifySubmit": "Verify & Create Account",
      "emailVerifyTitle": "Verify Your Email",
      "emailVerifyVerifying": "Verifying...",
      "passwordLabel": "Password",
      "passwordPlaceholder": "Enter your password",
      "forgotPassword": "Forgot password?",
      "forgotPasswordTitle": "Reset Your Password",
      "forgotPasswordHint": "Enter your email address and we will send you a link to reset your password.",
      "resetPasswordTitle": "Set New Password",
      "resetPasswordHint": "Enter your new password below.",
      "resetPassword": "Reset Password",
      "resettingPassword": "Resetting...",
      "invalidResetLink": "Invalid Reset Link",
      "invalidResetLinkHint": "This password reset link is invalid or has expired. Please request a new one.",
      "requestNewResetLink": "Request New Reset Link",
      "signingIn": "Signing in...",
      "signIn": "Sign In",
      "signInWithProvider": "Sign in with {providerName}",
      "oauthAlternativeMethods": "or continue with one of the following methods",
      "oauthCallbackCode": "Code",
      "oauthCallbackHint": "If you are not redirected automatically, go back to the login page and try again.",
      "oauthCallbackFullUrl": "Full URL",
      "oauthCallbackInvalidHint": "This page does not contain a valid authorization result. Return to the login page and start quick sign-in again.",
      "oauthCallbackInvalidTitle": "Invalid sign-in callback",
      "oauthCallbackPasswordOptionalHint": "You can leave this blank for {providerName} sign-in and set a password later from your profile.",
      "oauthCallbackRegistrationHint": "Complete Registration",
      "oauthCallbackRegistrationInvitationRequired": "This {providerName} account is not yet registered. The site requires an invitation code — please enter one to complete registration.",
      "oauthCallbackState": "State",
      "oauthCallbackSubmitRegistration": "Complete Registration",
      "oauthCallbackTitle": "OAuth Callback",
      "oauthFlowAvatarAlt": "{providerName} avatar",
      "oauthFlowBindCurrentAccount": "Bind current account",
      "oauthFlowBindCurrentAccountDescription": "Bind this {providerName} sign-in to the account currently signed in on this browser.",
      "oauthFlowBindCurrentAccountTitle": "Bind the current account",
      "oauthFlowBindExistingAccount": "Bind existing account",
      "oauthFlowBindLoginHint": "Log in to an existing account to bind this {providerName} sign-in.",
      "oauthFlowBindSignInToExistingAccount": "Bind this {providerName} sign-in to an existing account.",
      "oauthFlowChooseAccountActionHint": "Choose whether to bind an existing account or create a new one.",
      "oauthFlowChooseHowToContinue": "Choose how to continue",
      "oauthFlowCreateAccountHint": "Enter an email address to create your account and continue.",
      "oauthFlowCreateAccountTitle": "Complete your {providerName} account setup",
      "oauthFlowCreateNewAccount": "Create new account",
      "oauthFlowLogInAndBind": "Log in and bind",
      "oauthFlowProfileDetailsDescription": "Choose whether to apply the nickname or avatar from {providerName} to this account.",
      "oauthFlowProfileDetailsTitle": "Use {providerName} profile details",
      "oauthFlowReviewProfileBeforeContinue": "Review the {providerName} profile details before continuing.",
      "oauthFlowSignInThenBindDescription": "Sign in to an existing account, then bind this {providerName} sign-in to it.",
      "oauthFlowSuggestedEmail": "Suggested email: {email}",
      "oauthFlowTotpHint": "Enter the 6-digit verification code for {account} to finish binding this {providerName} sign-in.",
      "oauthFlowUseAvatar": "Use avatar",
      "oauthFlowUseDifferentEmail": "Use a different email",
      "oauthFlowUseDisplayName": "Use display name",
      "oauthFlowVerifyAndContinue": "Verify and continue",
      "oauthFlowYourAccount": "your account",
      "dontHaveAccount": "Don't have an account?",
      "signUp": "Sign Up",
      "createAccount": "Create Account",
      "signUpToStart": "Sign up to start using {siteName}",
      "registrationDisabled": "Registration is currently disabled. Please contact the administrator.",
      "createPasswordPlaceholder": "Create a strong password",
      "newPassword": "New Password",
      "newPasswordPlaceholder": "Enter your new password",
      "confirmPassword": "Confirm Password",
      "confirmPasswordPlaceholder": "Confirm your new password",
      "passwordHint": "At least {count} characters",
      "invitationCodeLabel": "Invitation Code",
      "invitationCodePlaceholder": "Enter invitation code",
      "invitationCodeValid": "Invitation code is valid",
      "affiliateInvitationDetected": "Invitation link detected. You can continue registration directly.",
      "sendCode": "Send Code",
      "sendingCode": "Sending...",
      "sendResetLink": "Send Reset Link",
      "sendingResetLink": "Sending...",
      "resendCountdown": "Resend in {countdown}s",
      "codeSentSuccess": "Code sent successfully",
      "verificationCodeHint": "Enter the 6-digit code sent to your email",
      "promoCodeLabel": "Promo Code",
      "optional": "Optional",
      "promoCodePlaceholder": "Enter promo code",
      "promoCodeValid": "Promo code valid. You will receive {amount} balance after registration",
      "processing": "Processing...",
      "continue": "Continue",
      "totpLoginTitle": "Two-Factor Authentication",
      "totpLoginHint": "Enter the 6-digit code from your authenticator app",
      "totpVerifying": "Verifying...",
      "totpCancel": "Cancel",
      "resetEmailSent": "Reset Link Sent",
      "resetEmailSentHint": "If an account exists with this email, you will receive a password reset link shortly. Please check your inbox and spam folder.",
      "passwordResetSuccess": "Password Reset Successful",
      "passwordResetSuccessHint": "Your password has been reset. You can now sign in with your new password.",
      "providerCallbackHint": "If you are not redirected automatically, go back to the login page and try again.",
      "providerCallbackProcessing": "Completing login with {providerName}, please wait...",
      "providerCallbackTitle": "Signing you in with {providerName}",
      "providerCompleteRegistration": "Complete Registration",
      "providerCompletingRegistration": "Completing registration…",
      "providerInvitationRequired": "This {providerName} account is not yet registered. The site requires an invitation code — please enter one to complete registration.",
      "agreementAcceptAndContinue": "Accept and continue",
      "agreementAcceptedDescription": "You have already accepted the current agreement version and can review the documents again at any time.",
      "agreementAcceptedTitle": "Agreement entry",
      "agreementCheckboxPrefix": "I have read and agree to",
      "agreementRecent": "recently",
      "agreementReject": "Reject",
      "agreementRelevantDocuments": "Related documents",
      "agreementReviewDescription": "You can enter your account details first. If this account has not accepted the latest agreement yet, we will prompt for confirmation when you submit sign-in.",
      "agreementReviewTitle": "You may need to confirm the latest agreement before signing in.",
      "agreementTermsUpdateTitle": "Terms update notice",
      "agreementUpdatedAt": "Our terms of service were updated on {date}. Please review and accept the following documents before continuing.",
      "agreementViewAndAccept": "View and accept",
      "agreementViewTerms": "View terms",
      "alreadyHaveAccount": "Already have an account?",
      "rememberedPassword": "Remembered your password?",
      "backToLogin": "Back to Login",
      "allRightsReserved": "All rights reserved."
    }
  }
}`

const defaultAPIKeysShellConfig = `{"zh":{"labels":{"actions":"操作","active":"启用","allGroups":"所有分组","allStatus":"所有状态","apiKey":"API Key","cancel":"取消","ccsClientSelectClaudeCode":"Claude Code","ccsClientSelectClaudeCodeDesc":"导入到 Claude Code","ccsClientSelectDescription":"选择要导入的客户端。","ccsClientSelectGeminiCli":"Gemini CLI","ccsClientSelectGeminiCliDesc":"导入到 Gemini CLI","ccsClientSelectTitle":"选择 CCS 客户端","ccSwitchNotInstalled":"未检测到 CC Switch 客户端","clickToChangeGroup":"点击切换分组","copyToClipboard":"复制到剪贴板","copied":"已复制","create":"创建","createFirstKey":"创建第一个 API Key 后即可开始调用。","created":"创建时间","createKey":"创建 Key","currentExpiration":"当前过期时间","customDate":"自定义日期","customKeyHint":"留空则自动生成安全 API Key。","customKeyInvalidChars":"自定义 Key 只能包含字母、数字、下划线和连字符","customKeyLabel":"自定义 Key","customKeyPlaceholder":"输入自定义 API Key","customKeyRequired":"请输入自定义 Key","customKeyTooShort":"自定义 Key 至少需要 16 个字符","delete":"删除","deleteConfirmMessage":"确定要删除 {name} 吗？此操作不可恢复。","deleteKey":"删除 API Key","disable":"禁用","edit":"编辑","editKey":"编辑 Key","enable":"启用","expiration":"过期时间","expirationDate":"过期日期","expirationDateHint":"到达该时间后，API Key 将自动失效。","expiresInDays":"{days} 天后过期","expiresAt":"过期时间","extendDays":"延长 {days} 天","failedToChangeGroup":"切换分组失败","failedToDelete":"删除失败","failedToLoad":"加载 API Keys 失败","failedToResetQuota":"重置额度用量失败","failedToResetRateLimit":"重置频率限制用量失败","failedToSave":"保存失败","failedToUpdateStatus":"更新状态失败","group":"分组","groupChangedSuccess":"分组已切换","groupLabel":"分组","groupRequired":"请选择分组","importToCcSwitch":"导入 CCS","inactive":"禁用","ipBlacklist":"IP 黑名单","ipBlacklistHint":"每行一个 IP 或 CIDR。","ipBlacklistPlaceholder":"例如：192.168.1.1","ipRestriction":"IP 访问限制","ipRestrictionEnabled":"已启用 IP 访问限制","ipWhitelist":"IP 白名单","ipWhitelistHint":"每行一个 IP 或 CIDR；留空表示不限制白名单。","ipWhitelistPlaceholder":"例如：192.168.1.1","keyCreatedSuccess":"API Key 已创建","keyDeletedSuccess":"API Key 已删除","keyDisabledSuccess":"API Key 已禁用","keyEnabledSuccess":"API Key 已启用","keyUpdatedSuccess":"API Key 已更新","lastUsedAt":"最后使用","name":"名称","nameLabel":"名称","namePlaceholder":"输入 Key 名称","noExpiration":"永不过期","noGroup":"未分组","noGroupFound":"没有匹配的分组","noKeysYet":"还没有 API Key","quota":"额度","quotaAmountHint":"填 0 或留空表示不限额。","quotaAmountPlaceholder":"0 表示不限额","quotaLimit":"额度限制","quotaResetSuccess":"额度用量已重置","quotaUsed":"已用额度","rateLimitColumn":"频率限制","rateLimitHint":"设置 5 小时、1 天或 7 天窗口内的消费上限。","rateLimit1d":"1 天限制","rateLimit5h":"5 小时限制","rateLimit7d":"7 天限制","rateLimitResetSuccess":"频率限制用量已重置","rateLimitSection":"频率限制","refresh":"刷新","reset":"重置","resetQuotaConfirmMessage":"确定要将 {name} 的已用额度从 ${used} 重置为 0 吗？","resetQuotaTitle":"重置已用额度","resetQuotaUsed":"重置已用额度","resetRateLimitConfirmMessage":"确定要重置 {name} 的频率限制用量吗？","resetRateLimitTitle":"重置频率限制用量","resetRateLimitUsage":"重置频率限制用量","resetNow":"即将重置","resetUsage":"重置用量","saving":"保存中...","searchGroup":"搜索分组","searchPlaceholder":"搜索 API Key","selectGroup":"选择分组","selectStatus":"选择状态","status":"状态","statusActive":"启用","statusExpired":"已过期","statusInactive":"禁用","statusLabel":"状态","statusQuotaExhausted":"额度耗尽","today":"今日","total":"累计","update":"更新","useKey":"使用 Key","usage":"用量"}},"en":{"labels":{"actions":"Actions","active":"Active","allGroups":"All groups","allStatus":"All status","apiKey":"API Key","cancel":"Cancel","ccsClientSelectClaudeCode":"Claude Code","ccsClientSelectClaudeCodeDesc":"Import to Claude Code","ccsClientSelectDescription":"Choose the client to import into.","ccsClientSelectGeminiCli":"Gemini CLI","ccsClientSelectGeminiCliDesc":"Import to Gemini CLI","ccsClientSelectTitle":"Select CCS Client","ccSwitchNotInstalled":"CC Switch client was not detected","clickToChangeGroup":"Click to change group","copyToClipboard":"Copy to clipboard","copied":"Copied","create":"Create","createFirstKey":"Create your first API key to start making requests.","created":"Created","createKey":"Create Key","currentExpiration":"Current expiration","customDate":"Custom date","customKeyHint":"Leave empty to generate a secure API key automatically.","customKeyInvalidChars":"Custom keys can only contain letters, numbers, underscores, and hyphens","customKeyLabel":"Custom Key","customKeyPlaceholder":"Enter a custom API key","customKeyRequired":"Please enter a custom key","customKeyTooShort":"Custom key must be at least 16 characters","delete":"Delete","deleteConfirmMessage":"Delete {name}? This action cannot be undone.","deleteKey":"Delete API Key","disable":"Disable","edit":"Edit","editKey":"Edit Key","enable":"Enable","expiration":"Expiration","expirationDate":"Expiration date","expirationDateHint":"The API key will stop working after this time.","expiresInDays":"Expires in {days} days","expiresAt":"Expires At","extendDays":"Extend {days} days","failedToChangeGroup":"Failed to change group","failedToDelete":"Failed to delete key","failedToLoad":"Failed to load API keys","failedToResetQuota":"Failed to reset quota usage","failedToResetRateLimit":"Failed to reset rate limit usage","failedToSave":"Failed to save API key","failedToUpdateStatus":"Failed to update status","group":"Group","groupChangedSuccess":"Group changed","groupLabel":"Group","groupRequired":"Please select a group","importToCcSwitch":"Import CCS","inactive":"Inactive","ipBlacklist":"IP blacklist","ipBlacklistHint":"One IP or CIDR per line.","ipBlacklistPlaceholder":"Example: 192.168.1.1","ipRestriction":"IP access restriction","ipRestrictionEnabled":"IP access restriction enabled","ipWhitelist":"IP whitelist","ipWhitelistHint":"One IP or CIDR per line. Leave empty to avoid whitelist restrictions.","ipWhitelistPlaceholder":"Example: 192.168.1.1","keyCreatedSuccess":"API key created","keyDeletedSuccess":"API key deleted","keyDisabledSuccess":"API key disabled","keyEnabledSuccess":"API key enabled","keyUpdatedSuccess":"API key updated","lastUsedAt":"Last Used","name":"Name","nameLabel":"Name","namePlaceholder":"Enter key name","noExpiration":"Never expires","noGroup":"No group","noGroupFound":"No matching groups","noKeysYet":"No API keys yet","quota":"Quota","quotaAmountHint":"Use 0 or leave empty for unlimited quota.","quotaAmountPlaceholder":"0 means unlimited","quotaLimit":"Quota limit","quotaResetSuccess":"Quota usage reset","quotaUsed":"Quota used","rateLimitColumn":"Rate Limit","rateLimitHint":"Set spend limits for 5-hour, 1-day, or 7-day windows.","rateLimit1d":"1-day limit","rateLimit5h":"5-hour limit","rateLimit7d":"7-day limit","rateLimitResetSuccess":"Rate limit usage reset","rateLimitSection":"Rate limit","refresh":"Refresh","reset":"Reset","resetQuotaConfirmMessage":"Reset used quota for {name} from ${used} to 0?","resetQuotaTitle":"Reset used quota","resetQuotaUsed":"Reset used quota","resetRateLimitConfirmMessage":"Reset rate limit usage for {name}?","resetRateLimitTitle":"Reset rate limit usage","resetRateLimitUsage":"Reset rate limit usage","resetNow":"Resetting soon","resetUsage":"Reset usage","saving":"Saving...","searchGroup":"Search groups","searchPlaceholder":"Search API keys","selectGroup":"Select group","selectStatus":"Select status","status":"Status","statusActive":"Active","statusExpired":"Expired","statusInactive":"Inactive","statusLabel":"Status","statusQuotaExhausted":"Quota exhausted","today":"Today","total":"Total","update":"Update","useKey":"Use Key","usage":"Usage"}}}`

const zhUseKeyModalLabels = `"useKeyModalAntigravityClaudeNote":"这些环境变量将在当前终端会话中生效。如需永久配置，请将其添加到 ~/.bashrc、~/.zshrc 或相应的配置文件中。","useKeyModalAntigravityDescription":"为 Antigravity 分组配置 API 访问。请根据您使用的客户端选择对应的配置方式。","useKeyModalAntigravityGeminiNote":"这些环境变量将在当前终端会话中生效。如需永久配置，请将其添加到 ~/.bashrc、~/.zshrc 或相应的配置文件中。","useKeyModalCliClaudeCode":"Claude Code","useKeyModalCliCodexCli":"Codex CLI","useKeyModalCliCodexCliWs":"Codex CLI (WebSocket)","useKeyModalCliGeminiCli":"Gemini CLI","useKeyModalCliOpencode":"OpenCode","useKeyModalCopied":"已复制","useKeyModalCopy":"复制","useKeyModalDescription":"将以下环境变量添加到您的终端配置文件或直接在终端中运行。","useKeyModalGeminiDescription":"将以下环境变量添加到您的终端配置文件或直接在终端中运行，以配置 Gemini CLI 访问。","useKeyModalGeminiModelComment":"如果你有 Gemini 3 权限可以填：gemini-3-pro-preview","useKeyModalGeminiNote":"这些环境变量将在当前终端会话中生效。如需永久配置，请将其添加到 ~/.bashrc、~/.zshrc 或相应的配置文件中。","useKeyModalNoGroupDescription":"此 API 密钥尚未分配分组，请先在密钥列表中点击分组列进行分配，然后才能查看使用配置。","useKeyModalNoGroupTitle":"请先分配分组","useKeyModalNote":"这些环境变量将在当前终端会话中生效。如需永久配置，请将其添加到 ~/.bashrc、~/.zshrc 或相应的配置文件中。","useKeyModalOpenAIConfigTomlHint":"请确保以下内容位于 config.toml 文件的开头部分","useKeyModalOpenAIDescription":"将以下配置文件添加到 Codex CLI 配置目录中。","useKeyModalOpenAINote":"请确保配置目录存在。macOS/Linux 用户可运行 mkdir -p ~/.codex 创建目录。","useKeyModalOpenAINoteWindows":"按 Win+R，输入 %userprofile%\\.codex 打开配置目录。如目录不存在，请先手动创建。","useKeyModalOpencodeHint":"配置文件路径：~/.config/opencode/opencode.json（或 opencode.jsonc），不存在需手动创建。可使用默认 provider（openai/anthropic/google）或自定义 provider_id。API Key 支持直接配置或通过客户端 /connect 命令配置。示例仅供参考，模型与选项可按需调整。","useKeyModalTitle":"使用 API 密钥",`

const enUseKeyModalLabels = `"useKeyModalAntigravityClaudeNote":"These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.","useKeyModalAntigravityDescription":"Configure API access for Antigravity group. Select the configuration method based on your client.","useKeyModalAntigravityGeminiNote":"These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.","useKeyModalCliClaudeCode":"Claude Code","useKeyModalCliCodexCli":"Codex CLI","useKeyModalCliCodexCliWs":"Codex CLI (WebSocket)","useKeyModalCliGeminiCli":"Gemini CLI","useKeyModalCliOpencode":"OpenCode","useKeyModalCopied":"Copied","useKeyModalCopy":"Copy","useKeyModalDescription":"Add the following environment variables to your terminal profile or run directly in terminal to configure API access.","useKeyModalGeminiDescription":"Add the following environment variables to your terminal profile or run directly in terminal to configure Gemini CLI access.","useKeyModalGeminiModelComment":"If you have Gemini 3 access, you can use: gemini-3-pro-preview","useKeyModalGeminiNote":"These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.","useKeyModalNoGroupDescription":"This API key has not been assigned to a group. Please click the group column in the key list to assign one before viewing the configuration.","useKeyModalNoGroupTitle":"Please assign a group first","useKeyModalNote":"These environment variables will be active in the current terminal session. For permanent configuration, add them to ~/.bashrc, ~/.zshrc, or the appropriate configuration file.","useKeyModalOpenAIConfigTomlHint":"Make sure the following content is at the beginning of the config.toml file","useKeyModalOpenAIDescription":"Add the following configuration files to your Codex CLI config directory.","useKeyModalOpenAINote":"Make sure the config directory exists. macOS/Linux users can run mkdir -p ~/.codex to create it.","useKeyModalOpenAINoteWindows":"Press Win+R and enter %userprofile%\\.codex to open the config directory. Create it manually if it does not exist.","useKeyModalOpencodeHint":"Config path: ~/.config/opencode/opencode.json (or opencode.jsonc), create if not exists. Use default providers (openai/anthropic/google) or custom provider_id. API Key can be configured directly or via /connect command. This is an example, adjust models and options as needed.","useKeyModalTitle":"Use API Key",`

const zhAPIKeysEndpointLabels = `"endpointClickToCopy":"点击可复制此端点","endpointCopied":"已复制","endpointCopiedHint":"已复制到剪贴板","endpointDefault":"默认","endpointSpeedTest":"测速","endpointTitle":"API 端点",`

const enAPIKeysEndpointLabels = `"endpointClickToCopy":"Click to copy this endpoint","endpointCopied":"Copied","endpointCopiedHint":"Copied to clipboard","endpointDefault":"Default","endpointSpeedTest":"Speed Test","endpointTitle":"API Endpoints",`

func parseBoolSettingWithDefault(raw string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true":
		return true
	case "false":
		return false
	default:
		return fallback
	}
}

func parseBoundedIntSetting(raw string, fallback, minValue, maxValue int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		v = fallback
	}
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

func normalizeJSONObjectSetting(raw string, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return fallback
	}
	normalized, err := json.Marshal(object)
	if err != nil {
		return fallback
	}
	return string(normalized)
}

func validateJSONObjectSetting(raw string, field string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "{}", nil
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return "", infraerrors.BadRequest("INVALID_JSON_SETTING", fmt.Sprintf("%s must be a valid JSON object", field))
	}
	normalized, err := json.Marshal(object)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", field, err)
	}
	return string(normalized), nil
}

// channelMonitorIntervalMin / channelMonitorIntervalMax bound the default interval
// (mirrors the monitor-level constraint but lives here so setting_service stays decoupled).
const (
	channelMonitorIntervalMin      = 15
	channelMonitorIntervalMax      = 3600
	channelMonitorIntervalFallback = 60
)

// ChannelMonitorRuntime is the lightweight view of the channel monitor feature
// consumed by the runner and user-facing handlers.
type ChannelMonitorRuntime struct {
	Enabled                bool
	DefaultIntervalSeconds int
}

// AvailableChannelsRuntime is the lightweight view of the available-channels feature
// switch consumed by the user-facing handler.
type AvailableChannelsRuntime struct {
	Enabled bool
}

var legacyClaudeCodeCodexWhitelistEntry = openai.AllowedClientEntry{
	Originator: "Claude Code",
	UAContains: []string{"Claude Code/"},
}

// SetOnUpdateCallback sets a callback function to be called when settings are updated
// This is used for cache invalidation (e.g., HTML cache in frontend server)
func (s *SettingService) SetOnUpdateCallback(callback func()) {
	s.onUpdate = callback
}

// SetVersion sets the application version for injection into public settings
func (s *SettingService) SetVersion(version string) {
	s.version = version
}

// PublicSettingsInjectionPayload is the JSON shape embedded into HTML as
// `window.__APP_CONFIG__` so the frontend can hydrate feature flags & site
// config before the first XHR finishes.
//
// INVARIANT: every `json` tag here MUST also exist on handler/dto.PublicSettings.
// If you forget a feature-flag field here, the frontend's
// `cachedPublicSettings.xxx_enabled` will be `undefined` on refresh until the
// async `/api/v1/settings/public` call returns — which causes opt-in menus
// (strict `=== true`) to flicker off/on. See
// frontend/src/utils/featureFlags.ts for the matching registry.
//
// A unit test diffs this struct's JSON keys against dto.PublicSettings to catch
// drift automatically (see setting_service_injection_test.go).
type PublicSettingsInjectionPayload struct {
	RegistrationEnabled              bool                     `json:"registration_enabled"`
	EmailVerifyEnabled               bool                     `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist []string                 `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool                     `json:"promo_code_enabled"`
	PasswordResetEnabled             bool                     `json:"password_reset_enabled"`
	PasswordMinLength                int                      `json:"password_min_length"`
	InvitationCodeEnabled            bool                     `json:"invitation_code_enabled"`
	TotpEnabled                      bool                     `json:"totp_enabled"`
	LoginAgreementEnabled            bool                     `json:"login_agreement_enabled"`
	LoginAgreementMode               string                   `json:"login_agreement_mode"`
	LoginAgreementUpdatedAt          string                   `json:"login_agreement_updated_at"`
	LoginAgreementRevision           string                   `json:"login_agreement_revision"`
	LoginAgreementDocuments          []LoginAgreementDocument `json:"login_agreement_documents"`
	TurnstileEnabled                 bool                     `json:"turnstile_enabled"`
	TurnstileSiteKey                 string                   `json:"turnstile_site_key"`
	SiteName                         string                   `json:"site_name"`
	SiteLogo                         string                   `json:"site_logo"`
	SiteSubtitle                     string                   `json:"site_subtitle"`
	APIBaseURL                       string                   `json:"api_base_url"`
	ContactInfo                      string                   `json:"contact_info"`
	DocURL                           string                   `json:"doc_url"`
	DocsContentBasePath              string                   `json:"docs_content_base_path"`
	HomeContent                      string                   `json:"home_content"`
	HomeShellConfig                  string                   `json:"home_shell_config"`
	HomeBusinessShellConfig          string                   `json:"home_business_shell_config"`
	ModelPlazaItems                  json.RawMessage          `json:"model_plaza_items"`
	ImageWorkspaceModelConfig        string                   `json:"image_workspace_model_config"`
	ModelPlazaShellConfig            string                   `json:"model_plaza_shell_config"`
	DocsShellConfig                  string                   `json:"docs_shell_config"`
	LegalDocumentShellConfig         string                   `json:"legal_document_shell_config"`
	APIKeysShellConfig               string                   `json:"api_keys_shell_config"`
	KeyUsageShellConfig              string                   `json:"key_usage_shell_config"`
	DashboardShellConfig             string                   `json:"dashboard_shell_config"`
	UsageShellConfig                 string                   `json:"usage_shell_config"`
	APIGuideShellConfig              string                   `json:"api_guide_shell_config"`
	APITestShellConfig               string                   `json:"api_test_shell_config"`
	AvailableGroupsShellConfig       string                   `json:"available_groups_shell_config"`
	RedeemShellConfig                string                   `json:"redeem_shell_config"`
	AffiliateShellConfig             string                   `json:"affiliate_shell_config"`
	AvailableChannelsShellConfig     string                   `json:"available_channels_shell_config"`
	ChannelStatusShellConfig         string                   `json:"channel_status_shell_config"`
	CustomPageShellConfig            string                   `json:"custom_page_shell_config"`
	ProfileShellConfig               string                   `json:"profile_shell_config"`
	AuthShellConfig                  string                   `json:"auth_shell_config"`
	HideCcsImportButton              bool                     `json:"hide_ccs_import_button"`
	PurchaseSubscriptionEnabled      bool                     `json:"purchase_subscription_enabled"`
	PurchaseSubscriptionURL          string                   `json:"purchase_subscription_url"`
	TableDefaultPageSize             int                      `json:"table_default_page_size"`
	TablePageSizeOptions             []int                    `json:"table_page_size_options"`
	CustomMenuItems                  json.RawMessage          `json:"custom_menu_items"`
	CustomEndpoints                  json.RawMessage          `json:"custom_endpoints"`
	WeChatOAuthEnabled               bool                     `json:"wechat_oauth_enabled"`
	WeChatOAuthOpenEnabled           bool                     `json:"wechat_oauth_open_enabled"`
	WeChatOAuthMPEnabled             bool                     `json:"wechat_oauth_mp_enabled"`
	WeChatOAuthMobileEnabled         bool                     `json:"wechat_oauth_mobile_enabled"`
	GitHubOAuthEnabled               bool                     `json:"github_oauth_enabled"`
	GoogleOAuthEnabled               bool                     `json:"google_oauth_enabled"`
	BackendModeEnabled               bool                     `json:"backend_mode_enabled"`
	PaymentEnabled                   bool                     `json:"payment_enabled"`
	Version                          string                   `json:"version"`
	DefaultLocale                    string                   `json:"default_locale"`
	// 服务器全局时区（IANA 名称与当前 UTC 偏移），高峰时段等服务端本地时间窗口的展示标注用
	ServerTimezone              string  `json:"server_timezone"`
	ServerUTCOffset             string  `json:"server_utc_offset"`
	BalanceLowNotifyEnabled     bool    `json:"balance_low_notify_enabled"`
	AccountQuotaNotifyEnabled   bool    `json:"account_quota_notify_enabled"`
	BalanceLowNotifyThreshold   float64 `json:"balance_low_notify_threshold"`
	BalanceLowNotifyRechargeURL string  `json:"balance_low_notify_recharge_url"`

	// Feature flags — MUST match the opt-in/opt-out registry in
	// frontend/src/utils/featureFlags.ts. Missing a field here is the bug
	// that hid the "可用渠道" menu on page refresh.
	ChannelMonitorEnabled                bool `json:"channel_monitor_enabled"`
	ChannelMonitorDefaultIntervalSeconds int  `json:"channel_monitor_default_interval_seconds"`
	AvailableChannelsEnabled             bool `json:"available_channels_enabled"`
	AffiliateEnabled                     bool `json:"affiliate_enabled"`
	RiskControlEnabled                   bool `json:"risk_control_enabled"`
	AllowUserViewErrorRequests           bool `json:"allow_user_view_error_requests"`

	PromptCasesTitle           string `json:"prompt_cases_title"`
	PromptCasesDescription     string `json:"prompt_cases_description"`
	PromptTemplatesTitle       string `json:"prompt_templates_title"`
	PromptTemplatesDescription string `json:"prompt_templates_description"`
	PromptCatalogShellConfig   string `json:"prompt_catalog_shell_config"`
	WorkspaceShellConfig       string `json:"workspace_shell_config"`
	PricingTitle               string `json:"pricing_title"`
	PricingDescription         string `json:"pricing_description"`
	PricingShellConfig         string `json:"pricing_shell_config"`
	PaymentShellConfig         string `json:"payment_shell_config"`
	PricingCurrencySymbol      string `json:"pricing_currency_symbol"`
	CreditsTitle               string `json:"credits_title"`
	CreditsDescription         string `json:"credits_description"`
	CreditsPurchaseLabel       string `json:"credits_purchase_label"`
	CreditsBalanceLabel        string `json:"credits_balance_label"`
	CreditsPerBalance          string `json:"credits_per_balance"`
	CreditsShellConfig         string `json:"credits_shell_config"`
	GoogleAnalyticsID          string `json:"google_analytics_id"`
	AffonsoEnabled             bool   `json:"affonso_enabled"`
	AffonsoID                  string `json:"affonso_id"`
	AffonsoCookieDuration      string `json:"affonso_cookie_duration"`
	PromoteKitEnabled          bool   `json:"promotekit_enabled"`
	PromoteKitID               string `json:"promotekit_id"`
	CrispEnabled               bool   `json:"crisp_enabled"`
	CrispWebsiteID             string `json:"crisp_website_id"`
}

// safeRawJSONArray returns raw as json.RawMessage if it's valid JSON, otherwise "[]".
func safeRawJSONArray(raw string) json.RawMessage {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return json.RawMessage("[]")
	}
	if json.Valid([]byte(raw)) {
		return json.RawMessage(raw)
	}
	return json.RawMessage("[]")
}

// UpdateSettings 更新系统设置
func (s *SettingService) UpdateSettings(ctx context.Context, settings *SystemSettings) error {
	updates, err := s.buildSystemSettingsUpdates(ctx, settings)
	if err != nil {
		return err
	}

	err = s.settingRepo.SetMultiple(ctx, updates)
	if err == nil {
		s.refreshCachedSettings(settings)
	}
	return err
}

// UpdateSettingsWithAuthSourceDefaults persists system settings and auth-source defaults in a single write.
func (s *SettingService) UpdateSettingsWithAuthSourceDefaults(ctx context.Context, settings *SystemSettings, authDefaults *AuthSourceDefaultSettings) error {
	updates, err := s.buildSystemSettingsUpdates(ctx, settings)
	if err != nil {
		return err
	}

	authSourceUpdates, err := s.buildAuthSourceDefaultUpdates(ctx, authDefaults)
	if err != nil {
		return err
	}
	for key, value := range authSourceUpdates {
		updates[key] = value
	}

	err = s.settingRepo.SetMultiple(ctx, updates)
	if err == nil {
		s.refreshCachedSettings(settings)
	}
	return err
}

// normalizeSettingsForPersist validates and normalizes (trims / defaults) the
// settings struct in place prior to building the persisted key/value updates.
// Extracted from buildSystemSettingsUpdates to keep that function focused on
// mapping normalized fields to setting keys.
func (s *SettingService) normalizeSettingsForPersist(ctx context.Context, settings *SystemSettings) error {
	if err := s.validateDefaultSubscriptionGroups(ctx, settings.DefaultSubscriptions); err != nil {
		return err
	}
	normalizedWhitelist, err := NormalizeRegistrationEmailSuffixWhitelist(settings.RegistrationEmailSuffixWhitelist)
	if err != nil {
		return infraerrors.BadRequest("INVALID_REGISTRATION_EMAIL_SUFFIX_WHITELIST", err.Error())
	}
	if normalizedWhitelist == nil {
		normalizedWhitelist = []string{}
	}
	settings.RegistrationEmailSuffixWhitelist = normalizedWhitelist
	alipaySource, err := normalizeVisibleMethodSettingSource("alipay", settings.PaymentVisibleMethodAlipaySource, settings.PaymentVisibleMethodAlipayEnabled)
	if err != nil {
		return err
	}
	wxpaySource, err := normalizeVisibleMethodSettingSource("wxpay", settings.PaymentVisibleMethodWxpaySource, settings.PaymentVisibleMethodWxpayEnabled)
	if err != nil {
		return err
	}
	registrationNotifyProvider, err := normalizeRegistrationNotifyProvider(settings.RegistrationNotifyProvider)
	if err != nil {
		return err
	}
	registrationNotifyWebhookURL, err := normalizeRegistrationNotifyWebhookURL(settings.RegistrationNotifyWebhookURL, settings.RegistrationNotifyEnabled)
	if err != nil {
		return err
	}
	if settings.RegistrationNotifyEnabled && registrationNotifyProvider == "" {
		return infraerrors.BadRequest("REGISTRATION_NOTIFY_PROVIDER_REQUIRED", "registration notification provider is required")
	}
	if err := s.normalizeOpenAIAdvancedSchedulerOverrides(settings); err != nil {
		return err
	}
	settings.PaymentVisibleMethodAlipaySource = alipaySource
	settings.PaymentVisibleMethodWxpaySource = wxpaySource
	settings.RegistrationNotifyProvider = registrationNotifyProvider
	settings.RegistrationNotifyWebhookURL = registrationNotifyWebhookURL
	settings.RegistrationNotifySecret = strings.TrimSpace(settings.RegistrationNotifySecret)
	settings.WeChatConnectOpenAppID = strings.TrimSpace(settings.WeChatConnectOpenAppID)
	settings.WeChatConnectOpenAppSecret = strings.TrimSpace(settings.WeChatConnectOpenAppSecret)
	settings.WeChatConnectMPAppID = strings.TrimSpace(settings.WeChatConnectMPAppID)
	settings.WeChatConnectMPAppSecret = strings.TrimSpace(settings.WeChatConnectMPAppSecret)
	settings.WeChatConnectMobileAppID = strings.TrimSpace(settings.WeChatConnectMobileAppID)
	settings.WeChatConnectMobileAppSecret = strings.TrimSpace(settings.WeChatConnectMobileAppSecret)
	settings.WeChatConnectMode = normalizeWeChatConnectStoredMode(
		settings.WeChatConnectOpenEnabled,
		settings.WeChatConnectMPEnabled,
		settings.WeChatConnectMobileEnabled,
		settings.WeChatConnectMode,
	)
	settings.WeChatConnectScopes = normalizeWeChatConnectScopeSetting(settings.WeChatConnectScopes, settings.WeChatConnectMode)
	settings.WeChatConnectRedirectURL = strings.TrimSpace(settings.WeChatConnectRedirectURL)
	settings.WeChatConnectFrontendRedirectURL = strings.TrimSpace(settings.WeChatConnectFrontendRedirectURL)
	if settings.WeChatConnectFrontendRedirectURL == "" {
		settings.WeChatConnectFrontendRedirectURL = defaultWeChatConnectFrontend
	}
	settings.GitHubOAuthRedirectURL = strings.TrimSpace(settings.GitHubOAuthRedirectURL)
	settings.GitHubOAuthFrontendRedirectURL = strings.TrimSpace(settings.GitHubOAuthFrontendRedirectURL)
	if settings.GitHubOAuthFrontendRedirectURL == "" {
		settings.GitHubOAuthFrontendRedirectURL = defaultGitHubOAuthFrontend
	}
	settings.GoogleOAuthRedirectURL = strings.TrimSpace(settings.GoogleOAuthRedirectURL)
	settings.GoogleOAuthFrontendRedirectURL = strings.TrimSpace(settings.GoogleOAuthFrontendRedirectURL)
	if settings.GoogleOAuthFrontendRedirectURL == "" {
		settings.GoogleOAuthFrontendRedirectURL = defaultGoogleOAuthFrontend
	}
	settings.PasswordMinLength = normalizePasswordMinLength(settings.PasswordMinLength)
	return nil
}

func (s *SettingService) buildSystemSettingsUpdates(ctx context.Context, settings *SystemSettings) (map[string]string, error) {
	if err := s.normalizeSettingsForPersist(ctx, settings); err != nil {
		return nil, err
	}

	updates := make(map[string]string)

	// 注册设置
	updates[SettingKeyRegistrationEnabled] = strconv.FormatBool(settings.RegistrationEnabled)
	updates[SettingKeyEmailVerifyEnabled] = strconv.FormatBool(settings.EmailVerifyEnabled)
	registrationEmailSuffixWhitelistJSON, err := json.Marshal(settings.RegistrationEmailSuffixWhitelist)
	if err != nil {
		return nil, fmt.Errorf("marshal registration email suffix whitelist: %w", err)
	}
	updates[SettingKeyRegistrationEmailSuffixWhitelist] = string(registrationEmailSuffixWhitelistJSON)
	updates[SettingKeyPromoCodeEnabled] = strconv.FormatBool(settings.PromoCodeEnabled)
	updates[SettingKeyPasswordResetEnabled] = strconv.FormatBool(settings.PasswordResetEnabled)
	updates[SettingKeyPasswordMinLength] = strconv.Itoa(settings.PasswordMinLength)
	updates[SettingKeyFrontendURL] = settings.FrontendURL
	updates[SettingKeyInvitationCodeEnabled] = strconv.FormatBool(settings.InvitationCodeEnabled)
	updates[SettingKeyTotpEnabled] = strconv.FormatBool(settings.TotpEnabled)
	settings.LoginAgreementMode = normalizeLoginAgreementMode(settings.LoginAgreementMode)
	settings.LoginAgreementUpdatedAt = strings.TrimSpace(settings.LoginAgreementUpdatedAt)
	if settings.LoginAgreementUpdatedAt == "" {
		settings.LoginAgreementUpdatedAt = defaultLoginAgreementDate
	}
	loginAgreementDocumentsJSON, err := marshalLoginAgreementDocuments(settings.LoginAgreementDocuments)
	if err != nil {
		return nil, err
	}
	updates[SettingKeyLoginAgreementEnabled] = strconv.FormatBool(settings.LoginAgreementEnabled)
	updates[SettingKeyLoginAgreementMode] = settings.LoginAgreementMode
	updates[SettingKeyLoginAgreementUpdatedAt] = settings.LoginAgreementUpdatedAt
	updates[SettingKeyLoginAgreementDocuments] = loginAgreementDocumentsJSON

	// 邮件服务设置（只有非空才更新密码）
	updates[SettingKeySMTPHost] = settings.SMTPHost
	updates[SettingKeySMTPPort] = strconv.Itoa(settings.SMTPPort)
	updates[SettingKeySMTPUsername] = settings.SMTPUsername
	if settings.SMTPPassword != "" {
		updates[SettingKeySMTPPassword] = settings.SMTPPassword
	}
	updates[SettingKeySMTPFrom] = settings.SMTPFrom
	updates[SettingKeySMTPFromName] = settings.SMTPFromName
	updates[SettingKeySMTPUseTLS] = strconv.FormatBool(settings.SMTPUseTLS)
	updates[SettingKeySMTPDailyLimit] = strconv.Itoa(normalizeSMTPDailyLimit(settings.SMTPDailyLimit))
	smtpChannelsJSON, err := marshalSMTPChannels(settings.SMTPChannels)
	if err != nil {
		return nil, fmt.Errorf("marshal smtp channels: %w", err)
	}
	updates[SettingKeySMTPChannels] = smtpChannelsJSON

	// Cloudflare Turnstile 设置（只有非空才更新密钥）
	updates[SettingKeyTurnstileEnabled] = strconv.FormatBool(settings.TurnstileEnabled)
	updates[SettingKeyTurnstileSiteKey] = settings.TurnstileSiteKey
	if settings.TurnstileSecretKey != "" {
		updates[SettingKeyTurnstileSecretKey] = settings.TurnstileSecretKey
	}
	updates[SettingKeyAPIKeyACLTrustForwardedIP] = strconv.FormatBool(settings.APIKeyACLTrustForwardedIP)

	// GitHub / Google 邮箱快捷登录
	updates[SettingKeyGitHubOAuthEnabled] = strconv.FormatBool(settings.GitHubOAuthEnabled)
	updates[SettingKeyGitHubOAuthClientID] = strings.TrimSpace(settings.GitHubOAuthClientID)
	updates[SettingKeyGitHubOAuthRedirectURL] = settings.GitHubOAuthRedirectURL
	updates[SettingKeyGitHubOAuthFrontendRedirectURL] = settings.GitHubOAuthFrontendRedirectURL
	if settings.GitHubOAuthClientSecret != "" {
		updates[SettingKeyGitHubOAuthClientSecret] = strings.TrimSpace(settings.GitHubOAuthClientSecret)
	}
	updates[SettingKeyGoogleOAuthEnabled] = strconv.FormatBool(settings.GoogleOAuthEnabled)
	updates[SettingKeyGoogleOAuthClientID] = strings.TrimSpace(settings.GoogleOAuthClientID)
	updates[SettingKeyGoogleOAuthRedirectURL] = settings.GoogleOAuthRedirectURL
	updates[SettingKeyGoogleOAuthFrontendRedirectURL] = settings.GoogleOAuthFrontendRedirectURL
	if settings.GoogleOAuthClientSecret != "" {
		updates[SettingKeyGoogleOAuthClientSecret] = strings.TrimSpace(settings.GoogleOAuthClientSecret)
	}

	// WeChat Connect OAuth 登录
	updates[SettingKeyWeChatConnectEnabled] = strconv.FormatBool(settings.WeChatConnectEnabled)
	updates[SettingKeyWeChatConnectOpenAppID] = settings.WeChatConnectOpenAppID
	updates[SettingKeyWeChatConnectMPAppID] = settings.WeChatConnectMPAppID
	updates[SettingKeyWeChatConnectMobileAppID] = settings.WeChatConnectMobileAppID
	updates[SettingKeyWeChatConnectOpenEnabled] = strconv.FormatBool(settings.WeChatConnectOpenEnabled)
	updates[SettingKeyWeChatConnectMPEnabled] = strconv.FormatBool(settings.WeChatConnectMPEnabled)
	updates[SettingKeyWeChatConnectMobileEnabled] = strconv.FormatBool(settings.WeChatConnectMobileEnabled)
	updates[SettingKeyWeChatConnectMode] = settings.WeChatConnectMode
	updates[SettingKeyWeChatConnectScopes] = settings.WeChatConnectScopes
	updates[SettingKeyWeChatConnectRedirectURL] = settings.WeChatConnectRedirectURL
	updates[SettingKeyWeChatConnectFrontendRedirectURL] = settings.WeChatConnectFrontendRedirectURL
	if settings.WeChatConnectOpenAppSecret != "" {
		updates[SettingKeyWeChatConnectOpenAppSecret] = settings.WeChatConnectOpenAppSecret
	}
	if settings.WeChatConnectMPAppSecret != "" {
		updates[SettingKeyWeChatConnectMPAppSecret] = settings.WeChatConnectMPAppSecret
	}
	if settings.WeChatConnectMobileAppSecret != "" {
		updates[SettingKeyWeChatConnectMobileAppSecret] = settings.WeChatConnectMobileAppSecret
	}

	// OEM设置
	updates[SettingKeySiteName] = settings.SiteName
	updates[SettingKeySiteLogo] = settings.SiteLogo
	updates[SettingKeySiteSubtitle] = settings.SiteSubtitle
	updates[SettingKeyAPIBaseURL] = settings.APIBaseURL
	updates[SettingKeyContactInfo] = settings.ContactInfo
	updates[SettingKeyDocURL] = settings.DocURL
	updates[SettingKeyDocsContentBasePath] = strings.TrimSpace(settings.DocsContentBasePath)
	updates[SettingKeyHomeContent] = settings.HomeContent
	updates[SettingKeyHomeShellConfig] = strings.TrimSpace(settings.HomeShellConfig)
	updates[SettingKeyHomeBusinessShellConfig] = strings.TrimSpace(settings.HomeBusinessShellConfig)
	updates[SettingKeyModelPlazaItems] = settings.ModelPlazaItems
	updates[SettingKeyImageWorkspaceModelConfig] = strings.TrimSpace(settings.ImageWorkspaceModelConfig)
	updates[SettingKeyModelPlazaShellConfig] = strings.TrimSpace(settings.ModelPlazaShellConfig)
	updates[SettingKeyDocsShellConfig] = strings.TrimSpace(settings.DocsShellConfig)
	updates[SettingKeyLegalDocumentShellConfig] = strings.TrimSpace(settings.LegalDocumentShellConfig)
	updates[SettingKeyAPIKeysShellConfig] = strings.TrimSpace(settings.APIKeysShellConfig)
	updates[SettingKeyKeyUsageShellConfig] = strings.TrimSpace(settings.KeyUsageShellConfig)
	updates[SettingKeyDashboardShellConfig] = strings.TrimSpace(settings.DashboardShellConfig)
	updates[SettingKeyUsageShellConfig] = strings.TrimSpace(settings.UsageShellConfig)
	updates[SettingKeyAPIGuideShellConfig] = strings.TrimSpace(settings.APIGuideShellConfig)
	updates[SettingKeyAPITestShellConfig] = strings.TrimSpace(settings.APITestShellConfig)
	updates[SettingKeyAvailableGroupsShellConfig] = strings.TrimSpace(settings.AvailableGroupsShellConfig)
	updates[SettingKeyRedeemShellConfig] = strings.TrimSpace(settings.RedeemShellConfig)
	updates[SettingKeyAffiliateShellConfig] = strings.TrimSpace(settings.AffiliateShellConfig)
	updates[SettingKeyAvailableChannelsShellConfig] = strings.TrimSpace(settings.AvailableChannelsShellConfig)
	updates[SettingKeyChannelStatusShellConfig] = strings.TrimSpace(settings.ChannelStatusShellConfig)
	updates[SettingKeyCustomPageShellConfig] = strings.TrimSpace(settings.CustomPageShellConfig)
	updates[SettingKeyProfileShellConfig] = strings.TrimSpace(settings.ProfileShellConfig)
	updates[SettingKeyAuthShellConfig] = strings.TrimSpace(settings.AuthShellConfig)
	updates[SettingKeyHideCcsImportButton] = strconv.FormatBool(settings.HideCcsImportButton)
	updates[SettingKeyPurchaseSubscriptionEnabled] = strconv.FormatBool(settings.PurchaseSubscriptionEnabled)
	updates[SettingKeyPurchaseSubscriptionURL] = strings.TrimSpace(settings.PurchaseSubscriptionURL)
	tableDefaultPageSize, tablePageSizeOptions := normalizeTablePreferences(
		settings.TableDefaultPageSize,
		settings.TablePageSizeOptions,
	)
	updates[SettingKeyTableDefaultPageSize] = strconv.Itoa(tableDefaultPageSize)
	tablePageSizeOptionsJSON, err := json.Marshal(tablePageSizeOptions)
	if err != nil {
		return nil, fmt.Errorf("marshal table page size options: %w", err)
	}
	updates[SettingKeyTablePageSizeOptions] = string(tablePageSizeOptionsJSON)
	updates[SettingKeyCustomMenuItems] = settings.CustomMenuItems
	updates[SettingKeyCustomEndpoints] = settings.CustomEndpoints
	updates[SettingKeyWebAppURL] = strings.TrimSpace(settings.WebAppURL)
	updates[SettingKeyWebAppName] = strings.TrimSpace(settings.WebAppName)
	updates[SettingKeyWebAppDescription] = strings.TrimSpace(settings.WebAppDescription)
	updates[SettingKeyWebAppLogo] = strings.TrimSpace(settings.WebAppLogo)
	updates[SettingKeyWebAppFavicon] = strings.TrimSpace(settings.WebAppFavicon)
	updates[SettingKeyWebAppPreviewImage] = strings.TrimSpace(settings.WebAppPreviewImage)
	updates[SettingKeyWebTheme] = strings.TrimSpace(settings.WebTheme)
	updates[SettingKeyWebAppearance] = strings.TrimSpace(settings.WebAppearance)
	updates[SettingKeyWebDefaultLocale] = strings.TrimSpace(settings.WebDefaultLocale)
	updates[SettingKeyPromptCasesTitle] = strings.TrimSpace(settings.WebPromptCasesTitle)
	updates[SettingKeyPromptCasesDescription] = strings.TrimSpace(settings.WebPromptCasesDescription)
	updates[SettingKeyPromptTemplatesTitle] = strings.TrimSpace(settings.WebPromptTemplatesTitle)
	updates[SettingKeyPromptTemplatesDescription] = strings.TrimSpace(settings.WebPromptTemplatesDescription)
	updates[SettingKeyPromptCatalogShellConfig] = strings.TrimSpace(settings.PromptCatalogShellConfig)
	updates[SettingKeyWorkspaceShellConfig] = strings.TrimSpace(settings.WebWorkspaceShellConfig)
	updates[SettingKeyImagePromptFilterConfig] = strings.TrimSpace(settings.WebImagePromptFilterConfig)
	updates[SettingKeyPricingTitle] = strings.TrimSpace(settings.WebPricingTitle)
	updates[SettingKeyPricingDescription] = strings.TrimSpace(settings.WebPricingDescription)
	updates[SettingKeyPricingShellConfig] = strings.TrimSpace(settings.WebPricingShellConfig)
	updates[SettingKeyPaymentShellConfig] = strings.TrimSpace(settings.WebPaymentShellConfig)
	updates[SettingKeyPricingCurrencySymbol] = pricingCurrencySymbolSetting(settings.WebPricingCurrencySymbol)
	updates[SettingKeyCreditsTitle] = strings.TrimSpace(settings.WebCreditsTitle)
	updates[SettingKeyCreditsDescription] = strings.TrimSpace(settings.WebCreditsDescription)
	updates[SettingKeyCreditsPurchaseLabel] = strings.TrimSpace(settings.WebCreditsPurchaseLabel)
	updates[SettingKeyCreditsBalanceLabel] = strings.TrimSpace(settings.WebCreditsBalanceLabel)
	updates[SettingKeyCreditsPerBalance] = creditsPerBalanceSetting(settings.WebCreditsPerBalance)
	updates[SettingKeyCreditsShellConfig] = strings.TrimSpace(settings.CreditsShellConfig)
	updates[SettingKeyWebLocaleDetectEnabled] = strconv.FormatBool(settings.WebLocaleDetectEnabled)
	updates[SettingKeyWebEmailAuthVisible] = strconv.FormatBool(settings.WebEmailAuthVisible)
	updates[SettingKeyWebGoogleAuthVisible] = strconv.FormatBool(settings.WebGoogleAuthVisible)
	updates[SettingKeyWebGitHubAuthVisible] = strconv.FormatBool(settings.WebGitHubAuthVisible)
	updates[SettingKeyWebGoogleAnalyticsID] = strings.TrimSpace(settings.WebGoogleAnalyticsID)
	updates[SettingKeyWebClarityID] = strings.TrimSpace(settings.WebClarityID)
	updates[SettingKeyWebPlausibleDomain] = strings.TrimSpace(settings.WebPlausibleDomain)
	updates[SettingKeyWebPlausibleSrc] = strings.TrimSpace(settings.WebPlausibleSrc)
	updates[SettingKeyWebOpenPanelClientID] = strings.TrimSpace(settings.WebOpenPanelClientID)
	updates[SettingKeyWebPublicIntegrationsEnabled] = strconv.FormatBool(settings.WebPublicIntegrationsEnabled)
	updates[SettingKeyWebVercelAnalyticsEnabled] = strconv.FormatBool(settings.WebVercelAnalyticsEnabled)
	updates[SettingKeyWebAdsenseCode] = strings.TrimSpace(settings.WebAdsenseCode)
	updates[SettingKeyWebAffonsoEnabled] = strconv.FormatBool(settings.WebAffonsoEnabled)
	updates[SettingKeyWebAffonsoID] = strings.TrimSpace(settings.WebAffonsoID)
	updates[SettingKeyWebAffonsoCookieDuration] = strings.TrimSpace(settings.WebAffonsoCookieDuration)
	updates[SettingKeyWebPromoteKitEnabled] = strconv.FormatBool(settings.WebPromoteKitEnabled)
	updates[SettingKeyWebPromoteKitID] = strings.TrimSpace(settings.WebPromoteKitID)
	updates[SettingKeyWebCrispEnabled] = strconv.FormatBool(settings.WebCrispEnabled)
	updates[SettingKeyWebCrispWebsiteID] = strings.TrimSpace(settings.WebCrispWebsiteID)
	updates[SettingKeyWebTawkEnabled] = strconv.FormatBool(settings.WebTawkEnabled)
	updates[SettingKeyWebTawkPropertyID] = strings.TrimSpace(settings.WebTawkPropertyID)
	updates[SettingKeyWebTawkWidgetID] = strings.TrimSpace(settings.WebTawkWidgetID)
	wechatWorkerIntervalMS := parseBoundedIntSetting(strconv.Itoa(settings.WeChatExportWorkerIntervalMS), 5000, 1000, 300000)
	wechatWorkerMaxBackoffMS := parseBoundedIntSetting(strconv.Itoa(settings.WeChatExportWorkerMaxBackoffMS), 60000, wechatWorkerIntervalMS, 300000)
	updates[SettingKeyWeChatExportFetchRetries] = strconv.Itoa(parseBoundedIntSetting(strconv.Itoa(settings.WeChatExportFetchRetries), 2, 0, 5))
	updates[SettingKeyWeChatExportFetchTimeoutMS] = strconv.Itoa(parseBoundedIntSetting(strconv.Itoa(settings.WeChatExportFetchTimeoutMS), 20000, 1000, 120000))
	updates[SettingKeyWeChatExportWorkerConcurrency] = strconv.Itoa(parseBoundedIntSetting(strconv.Itoa(settings.WeChatExportWorkerConcurrency), 1, 1, 8))
	updates[SettingKeyWeChatExportWorkerIntervalMS] = strconv.Itoa(wechatWorkerIntervalMS)
	updates[SettingKeyWeChatExportWorkerLeaseSeconds] = strconv.Itoa(parseBoundedIntSetting(strconv.Itoa(settings.WeChatExportWorkerLeaseSeconds), 300, 60, 3600))
	updates[SettingKeyWeChatExportWorkerMaxBackoffMS] = strconv.Itoa(wechatWorkerMaxBackoffMS)
	updates[SettingKeyImageWorkspaceUpstreamURL] = strings.TrimSpace(settings.ImageWorkspaceUpstreamURL)
	updates[SettingKeyImageWorkspaceGenerationTimeoutMS] = strconv.Itoa(parseBoundedIntSetting(strconv.Itoa(settings.ImageWorkspaceGenerationTimeoutMS), 420000, 1000, 900000))
	updates[SettingKeyImageWorkspaceCompletionCost] = strings.TrimSpace(settings.ImageWorkspaceCompletionCost)
	imageWorkspaceCompletionCostMapJSON, err := validateJSONObjectSetting(settings.ImageWorkspaceCompletionCostMapJSON, "image_workspace_completion_cost_map_json")
	if err != nil {
		return nil, err
	}
	updates[SettingKeyImageWorkspaceCompletionCostMapJSON] = imageWorkspaceCompletionCostMapJSON
	updates[SettingKeyImageWorkspacePromptSafetyEnabled] = strconv.FormatBool(settings.ImageWorkspacePromptSafetyEnabled)
	updates[SettingKeyImageWorkspaceAssumeWorkerReady] = strconv.FormatBool(settings.ImageWorkspaceAssumeWorkerReady)
	updates[SettingKeyImageWorkspaceObjectStorageEnabled] = strconv.FormatBool(settings.ImageWorkspaceObjectStorageEnabled)
	updates[SettingKeyImageWorkspaceObjectStorageProvider] = strings.TrimSpace(settings.ImageWorkspaceObjectStorageProvider)
	updates[SettingKeyImageWorkspaceObjectStorageBucket] = strings.TrimSpace(settings.ImageWorkspaceObjectStorageBucket)
	updates[SettingKeyImageWorkspaceObjectStorageRegion] = strings.TrimSpace(settings.ImageWorkspaceObjectStorageRegion)
	updates[SettingKeyImageWorkspaceObjectStoragePrefix] = strings.TrimSpace(settings.ImageWorkspaceObjectStoragePrefix)
	updates[SettingKeyImageWorkspaceObjectStoragePublicBaseURL] = strings.TrimSpace(settings.ImageWorkspaceObjectStoragePublicBaseURL)
	updates[SettingKeyMediaCDNBaseURL] = strings.TrimSpace(settings.MediaCDNBaseURL)

	// 默认配置
	updates[SettingKeyDefaultConcurrency] = strconv.Itoa(settings.DefaultConcurrency)
	updates[SettingKeyDefaultBalance] = strconv.FormatFloat(settings.DefaultBalance, 'f', 8, 64)
	settings.AffiliateRebateRate = clampAffiliateRebateRate(settings.AffiliateRebateRate)
	updates[SettingKeyAffiliateRebateRate] = strconv.FormatFloat(settings.AffiliateRebateRate, 'f', 8, 64)
	if settings.AffiliateRebateFreezeHours < 0 {
		settings.AffiliateRebateFreezeHours = AffiliateRebateFreezeHoursDefault
	}
	if settings.AffiliateRebateFreezeHours > AffiliateRebateFreezeHoursMax {
		settings.AffiliateRebateFreezeHours = AffiliateRebateFreezeHoursMax
	}
	updates[SettingKeyAffiliateRebateFreezeHours] = strconv.Itoa(settings.AffiliateRebateFreezeHours)
	if settings.AffiliateRebateDurationDays < 0 {
		settings.AffiliateRebateDurationDays = AffiliateRebateDurationDaysDefault
	}
	if settings.AffiliateRebateDurationDays > AffiliateRebateDurationDaysMax {
		settings.AffiliateRebateDurationDays = AffiliateRebateDurationDaysMax
	}
	updates[SettingKeyAffiliateRebateDurationDays] = strconv.Itoa(settings.AffiliateRebateDurationDays)
	if settings.AffiliateRebatePerInviteeCap < 0 {
		settings.AffiliateRebatePerInviteeCap = AffiliateRebatePerInviteeCapDefault
	}
	updates[SettingKeyAffiliateRebatePerInviteeCap] = strconv.FormatFloat(settings.AffiliateRebatePerInviteeCap, 'f', 8, 64)
	updates[SettingKeyDefaultUserRPMLimit] = strconv.Itoa(settings.DefaultUserRPMLimit)
	defaultSubsJSON, err := json.Marshal(settings.DefaultSubscriptions)
	if err != nil {
		return nil, fmt.Errorf("marshal default subscriptions: %w", err)
	}
	updates[SettingKeyDefaultSubscriptions] = string(defaultSubsJSON)

	// Model fallback configuration
	updates[SettingKeyEnableModelFallback] = strconv.FormatBool(settings.EnableModelFallback)
	updates[SettingKeyFallbackModelAnthropic] = settings.FallbackModelAnthropic
	updates[SettingKeyFallbackModelOpenAI] = settings.FallbackModelOpenAI
	updates[SettingKeyFallbackModelGemini] = settings.FallbackModelGemini
	updates[SettingKeyFallbackModelAntigravity] = settings.FallbackModelAntigravity

	// Identity patch configuration (Claude -> Gemini)
	updates[SettingKeyEnableIdentityPatch] = strconv.FormatBool(settings.EnableIdentityPatch)
	updates[SettingKeyIdentityPatchPrompt] = settings.IdentityPatchPrompt

	// Ops monitoring (vNext)
	updates[SettingKeyOpsMonitoringEnabled] = strconv.FormatBool(settings.OpsMonitoringEnabled)
	updates[SettingKeyOpsRealtimeMonitoringEnabled] = strconv.FormatBool(settings.OpsRealtimeMonitoringEnabled)
	updates[SettingKeyOpsQueryModeDefault] = string(ParseOpsQueryMode(settings.OpsQueryModeDefault))
	if settings.OpsMetricsIntervalSeconds > 0 {
		updates[SettingKeyOpsMetricsIntervalSeconds] = strconv.Itoa(settings.OpsMetricsIntervalSeconds)
	}

	// Channel monitor feature switch
	updates[SettingKeyChannelMonitorEnabled] = strconv.FormatBool(settings.ChannelMonitorEnabled)
	if v := clampChannelMonitorInterval(settings.ChannelMonitorDefaultIntervalSeconds); v > 0 {
		updates[SettingKeyChannelMonitorDefaultIntervalSeconds] = strconv.Itoa(v)
	}

	// Available channels feature switch
	updates[SettingKeyAvailableChannelsEnabled] = strconv.FormatBool(settings.AvailableChannelsEnabled)

	// Affiliate (邀请返利) feature switch
	updates[SettingKeyAffiliateEnabled] = strconv.FormatBool(settings.AffiliateEnabled)

	// 风控中心功能开关
	updates[SettingKeyRiskControlEnabled] = strconv.FormatBool(settings.RiskControlEnabled)
	updates[SettingKeySignupGrantRiskControlEnabled] = strconv.FormatBool(settings.SignupGrantRiskControlEnabled)
	updates[SettingKeySignupGrantRiskControlEmailLimit] = strconv.Itoa(nonNegativeInt(settings.SignupGrantRiskControlEmailLimit))
	updates[SettingKeySignupGrantRiskControlIPLimit] = strconv.Itoa(nonNegativeInt(settings.SignupGrantRiskControlIPLimit))
	updates[SettingKeySignupGrantRiskControlDomainLimit] = strconv.Itoa(nonNegativeInt(settings.SignupGrantRiskControlDomainLimit))
	updates[SettingKeySignupGrantRiskControlOAuthIdentityEnabled] = strconv.FormatBool(settings.SignupGrantRiskControlOAuthIdentityEnabled)
	updates[SettingKeySignupGrantRiskControlDeviceEnabled] = strconv.FormatBool(settings.SignupGrantRiskControlDeviceEnabled)
	updates[SettingKeySignupGrantRiskControlDeviceLimit] = strconv.Itoa(nonNegativeInt(settings.SignupGrantRiskControlDeviceLimit))
	updates[SettingKeySignupGrantRiskControlFreeDomainLimit] = strconv.Itoa(nonNegativeInt(settings.SignupGrantRiskControlFreeDomainLimit))
	updates[SettingKeySignupGrantRiskControlBlockedDomains] = normalizeDomainListSetting(settings.SignupGrantRiskControlBlockedDomains)
	updates[SettingKeySignupGrantRiskControlFreeDomains] = normalizeDomainListSetting(settings.SignupGrantRiskControlFreeDomains)
	updates[SettingKeySignupGrantRiskControlTrustedDomains] = normalizeDomainListSetting(settings.SignupGrantRiskControlTrustedDomains)

	// cyber 会话屏蔽开关 + TTL
	updates[SettingKeyCyberSessionBlockEnabled] = strconv.FormatBool(settings.CyberSessionBlockEnabled)
	if settings.CyberSessionBlockTTLSeconds > 0 {
		updates[SettingKeyCyberSessionBlockTTLSeconds] = strconv.Itoa(settings.CyberSessionBlockTTLSeconds)
	}

	// Claude Code version check
	updates[SettingKeyMinClaudeCodeVersion] = settings.MinClaudeCodeVersion
	updates[SettingKeyMaxClaudeCodeVersion] = settings.MaxClaudeCodeVersion

	// 分组隔离
	updates[SettingKeyAllowUngroupedKeyScheduling] = strconv.FormatBool(settings.AllowUngroupedKeyScheduling)

	// Backend Mode
	updates[SettingKeyBackendModeEnabled] = strconv.FormatBool(settings.BackendModeEnabled)

	// Gateway forwarding behavior
	updates[SettingKeyEnableFingerprintUnification] = strconv.FormatBool(settings.EnableFingerprintUnification)
	updates[SettingKeyEnableMetadataPassthrough] = strconv.FormatBool(settings.EnableMetadataPassthrough)
	updates[SettingKeyEnableCCHSigning] = strconv.FormatBool(settings.EnableCCHSigning)
	updates[SettingKeyEnableClaudeOAuthSystemPromptInjection] = strconv.FormatBool(settings.EnableClaudeOAuthSystemPromptInjection)
	updates[SettingKeyClaudeOAuthSystemPrompt] = settings.ClaudeOAuthSystemPrompt
	if err := ValidateClaudeOAuthSystemPromptBlocksConfig(settings.ClaudeOAuthSystemPromptBlocks); err != nil {
		return nil, err
	}
	updates[SettingKeyClaudeOAuthSystemPromptBlocks] = settings.ClaudeOAuthSystemPromptBlocks
	updates[SettingKeyEnableAnthropicCacheTTL1hInjection] = strconv.FormatBool(settings.EnableAnthropicCacheTTL1hInjection)
	updates[SettingKeyRewriteMessageCacheControl] = strconv.FormatBool(settings.RewriteMessageCacheControl)
	updates[SettingKeyEnableClientDatelineNormalization] = strconv.FormatBool(settings.EnableClientDatelineNormalization)
	updates[SettingKeyAntigravityUserAgentVersion] = antigravity.NormalizeUserAgentVersion(settings.AntigravityUserAgentVersion)
	updates[SettingKeyOpenAICodexUserAgent] = strings.TrimSpace(settings.OpenAICodexUserAgent)
	// codex_cli_only 加固
	updates[SettingKeyMinCodexVersion] = strings.TrimSpace(settings.MinCodexVersion)
	updates[SettingKeyMaxCodexVersion] = strings.TrimSpace(settings.MaxCodexVersion)
	updates[SettingKeyCodexCLIOnlyBlacklist] = strings.TrimSpace(settings.CodexCLIOnlyBlacklist)
	updates[SettingKeyCodexCLIOnlyWhitelist] = strings.TrimSpace(settings.CodexCLIOnlyWhitelist)
	updates[SettingKeyCodexCLIOnlyAllowAppServerClients] = strconv.FormatBool(settings.CodexCLIOnlyAllowAppServerClients)
	updates[SettingKeyCodexCLIOnlyEngineFingerprintSignals] = strings.TrimSpace(settings.CodexCLIOnlyEngineFingerprintSignals)
	updates[SettingPaymentVisibleMethodAlipaySource] = settings.PaymentVisibleMethodAlipaySource
	updates[SettingPaymentVisibleMethodWxpaySource] = settings.PaymentVisibleMethodWxpaySource
	updates[SettingPaymentVisibleMethodAlipayEnabled] = strconv.FormatBool(settings.PaymentVisibleMethodAlipayEnabled)
	updates[SettingPaymentVisibleMethodWxpayEnabled] = strconv.FormatBool(settings.PaymentVisibleMethodWxpayEnabled)
	updates[openAIAdvancedSchedulerSettingKey] = strconv.FormatBool(settings.OpenAIAdvancedSchedulerEnabled)
	updates[SettingKeyOpenAIAdvancedSchedulerStickyWeightedEnabled] = strconv.FormatBool(settings.OpenAIAdvancedSchedulerStickyWeightedEnabled)
	updates[SettingKeyOpenAIAdvancedSchedulerSubscriptionPriorityEnabled] = strconv.FormatBool(settings.OpenAIAdvancedSchedulerSubscriptionPriorityEnabled)
	updates[SettingKeyOpenAIAdvancedSchedulerLBTopK] = settings.OpenAIAdvancedSchedulerLBTopK
	updates[SettingKeyOpenAIAdvancedSchedulerWeightPriority] = settings.OpenAIAdvancedSchedulerWeightPriority
	updates[SettingKeyOpenAIAdvancedSchedulerWeightLoad] = settings.OpenAIAdvancedSchedulerWeightLoad
	updates[SettingKeyOpenAIAdvancedSchedulerWeightQueue] = settings.OpenAIAdvancedSchedulerWeightQueue
	updates[SettingKeyOpenAIAdvancedSchedulerWeightErrorRate] = settings.OpenAIAdvancedSchedulerWeightErrorRate
	updates[SettingKeyOpenAIAdvancedSchedulerWeightTTFT] = settings.OpenAIAdvancedSchedulerWeightTTFT
	updates[SettingKeyOpenAIAdvancedSchedulerWeightReset] = settings.OpenAIAdvancedSchedulerWeightReset
	updates[SettingKeyOpenAIAdvancedSchedulerWeightQuotaHeadroom] = settings.OpenAIAdvancedSchedulerWeightQuotaHeadroom
	updates[SettingKeyOpenAIAdvancedSchedulerWeightPreviousResponse] = settings.OpenAIAdvancedSchedulerWeightPreviousResponse
	updates[SettingKeyOpenAIAdvancedSchedulerWeightSessionSticky] = settings.OpenAIAdvancedSchedulerWeightSessionSticky

	// 余额、订阅到期与账号限额通知
	updates[SettingKeyBalanceLowNotifyEnabled] = strconv.FormatBool(settings.BalanceLowNotifyEnabled)
	updates[SettingKeyBalanceLowNotifyThreshold] = strconv.FormatFloat(settings.BalanceLowNotifyThreshold, 'f', 8, 64)
	updates[SettingKeyBalanceLowNotifyRechargeURL] = settings.BalanceLowNotifyRechargeURL
	updates[SettingKeySubscriptionExpiryNotifyEnabled] = strconv.FormatBool(settings.SubscriptionExpiryNotifyEnabled)
	updates[SettingKeyAccountQuotaNotifyEnabled] = strconv.FormatBool(settings.AccountQuotaNotifyEnabled)
	updates[SettingKeyAccountQuotaNotifyEmails] = MarshalNotifyEmails(settings.AccountQuotaNotifyEmails)
	updates[SettingKeyRegistrationNotifyEnabled] = strconv.FormatBool(settings.RegistrationNotifyEnabled)
	updates[SettingKeyRegistrationNotifyProvider] = settings.RegistrationNotifyProvider
	updates[SettingKeyRegistrationNotifyWebhookURL] = settings.RegistrationNotifyWebhookURL
	if settings.RegistrationNotifySecret != "" {
		updates[SettingKeyRegistrationNotifySecret] = settings.RegistrationNotifySecret
	}

	// 系统全局 platform quota：整体替换语义（null/缺省 = 不限制）。
	if settings.DefaultPlatformQuotas != nil {
		if err := validateDefaultPlatformQuotaMap(settings.DefaultPlatformQuotas); err != nil {
			return nil, err
		}
		blob, err := json.Marshal(settings.DefaultPlatformQuotas)
		if err != nil {
			return nil, fmt.Errorf("marshal default platform quotas: %w", err)
		}
		updates[SettingKeyDefaultPlatformQuotas] = string(blob)
	}

	updates[SettingKeyAllowUserViewErrorRequests] = strconv.FormatBool(settings.AllowUserViewErrorRequests)

	return updates, nil
}

func (s *SettingService) refreshCachedSettings(settings *SystemSettings) {
	if settings == nil {
		return
	}

	// 先使 inflight singleflight 失效，再刷新缓存，缩小旧值覆盖新值的竞态窗口
	versionBoundsSF.Forget("version_bounds")
	versionBoundsCache.Store(&cachedVersionBounds{
		min:       settings.MinClaudeCodeVersion,
		max:       settings.MaxClaudeCodeVersion,
		expiresAt: time.Now().Add(versionBoundsCacheTTL).UnixNano(),
	})
	backendModeSF.Forget("backend_mode")
	backendModeCache.Store(&cachedBackendMode{
		value:     settings.BackendModeEnabled,
		expiresAt: time.Now().Add(backendModeCacheTTL).UnixNano(),
	})
	gatewayForwardingSF.Forget("gateway_forwarding")
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{
		fingerprintUnification:           settings.EnableFingerprintUnification,
		metadataPassthrough:              settings.EnableMetadataPassthrough,
		cchSigning:                       settings.EnableCCHSigning,
		claudeOAuthSystemPromptInjection: settings.EnableClaudeOAuthSystemPromptInjection,
		claudeOAuthSystemPrompt:          settings.ClaudeOAuthSystemPrompt,
		claudeOAuthSystemPromptBlocks:    settings.ClaudeOAuthSystemPromptBlocks,
		anthropicCacheTTL1hInjection:     settings.EnableAnthropicCacheTTL1hInjection,
		rewriteMessageCacheControl:       settings.RewriteMessageCacheControl,
		clientDatelineNormalization:      settings.EnableClientDatelineNormalization,
		expiresAt:                        time.Now().Add(gatewayForwardingCacheTTL).UnixNano(),
	})
	s.antigravityUAVersionSF.Forget("antigravity_user_agent_version")
	antigravityUserAgentVersion := antigravity.NormalizeUserAgentVersion(settings.AntigravityUserAgentVersion)
	if antigravityUserAgentVersion == "" {
		antigravityUserAgentVersion = antigravity.GetDefaultUserAgentVersion()
	}
	s.antigravityUAVersionCache.Store(&cachedAntigravityUserAgentVersion{
		version:   antigravityUserAgentVersion,
		expiresAt: time.Now().Add(antigravityUserAgentVersionCacheTTL).UnixNano(),
	})
	s.openAICodexUASF.Forget("openai_codex_user_agent")
	codexUA := strings.TrimSpace(settings.OpenAICodexUserAgent)
	if codexUA == "" {
		codexUA = DefaultOpenAICodexUserAgent
	}
	s.openAICodexUACache.Store(&cachedOpenAICodexUserAgent{
		value:     codexUA,
		expiresAt: time.Now().Add(openAICodexUserAgentCacheTTL).UnixNano(),
	})
	openAIAdvancedSchedulerSettingSF.Forget(openAIAdvancedSchedulerSettingKey)
	openAIAdvancedSchedulerSettingCache.Store(&cachedOpenAIAdvancedSchedulerSetting{
		enabled:                     settings.OpenAIAdvancedSchedulerEnabled,
		stickyWeightedEnabled:       settings.OpenAIAdvancedSchedulerStickyWeightedEnabled,
		subscriptionPriorityEnabled: settings.OpenAIAdvancedSchedulerSubscriptionPriorityEnabled,
		lbTopKOverride:              parsePositiveIntOverride(settings.OpenAIAdvancedSchedulerLBTopK),
		weightOverrides: parseOpenAIAdvancedSchedulerWeightOverrides(map[string]string{
			SettingKeyOpenAIAdvancedSchedulerWeightPriority:         settings.OpenAIAdvancedSchedulerWeightPriority,
			SettingKeyOpenAIAdvancedSchedulerWeightLoad:             settings.OpenAIAdvancedSchedulerWeightLoad,
			SettingKeyOpenAIAdvancedSchedulerWeightQueue:            settings.OpenAIAdvancedSchedulerWeightQueue,
			SettingKeyOpenAIAdvancedSchedulerWeightErrorRate:        settings.OpenAIAdvancedSchedulerWeightErrorRate,
			SettingKeyOpenAIAdvancedSchedulerWeightTTFT:             settings.OpenAIAdvancedSchedulerWeightTTFT,
			SettingKeyOpenAIAdvancedSchedulerWeightReset:            settings.OpenAIAdvancedSchedulerWeightReset,
			SettingKeyOpenAIAdvancedSchedulerWeightQuotaHeadroom:    settings.OpenAIAdvancedSchedulerWeightQuotaHeadroom,
			SettingKeyOpenAIAdvancedSchedulerWeightPreviousResponse: settings.OpenAIAdvancedSchedulerWeightPreviousResponse,
			SettingKeyOpenAIAdvancedSchedulerWeightSessionSticky:    settings.OpenAIAdvancedSchedulerWeightSessionSticky,
		}),
		expiresAt: time.Now().Add(openAIAdvancedSchedulerSettingCacheTTL).UnixNano(),
	})
	// Invalidate the quota auto-pause cache and let the next read trigger a fresh load.
	// We can't know from here whether ops_advanced_settings was also touched, so be
	// defensive: store an expired entry — GetOpenAIQuotaAutoPauseSettings will serve
	// stale and kick off an async refresh, never blocking the request that follows.
	s.openAIQuotaAutoPauseSettingsSF.Forget(openAIQuotaAutoPauseSettingsRefreshKey)
	if cached, _ := s.openAIQuotaAutoPauseSettingsCache.Load().(*cachedOpenAIQuotaAutoPauseSettings); cached != nil {
		s.openAIQuotaAutoPauseSettingsCache.Store(&cachedOpenAIQuotaAutoPauseSettings{
			settings:  cached.settings,
			expiresAt: 0,
		})
	}
	if s.cfg != nil {
		s.cfg.SetTrustForwardedIPForAPIKeyACL(settings.APIKeyACLTrustForwardedIP)
	}
	// codex_cli_only 加固策略缓存：设置更新后强制下次重载（涉及 4 个键 + JSON 解析，直接置过期）。
	s.codexRestrictionPolicySF.Forget("codex_restriction_policy")
	s.codexRestrictionPolicyCache.Store(&cachedCodexRestrictionPolicy{expiresAt: 0})
	if s.onUpdate != nil {
		s.onUpdate() // Invalidate cache after settings update
	}
}

type gatewayForwardingSettingsResult struct {
	fp, mp, cch, claudeOAuthSystemPromptInjection, cacheTTL1h, rewriteMessageCacheControl bool
	clientDatelineNormalization                                                           bool
	claudeOAuthSystemPrompt, claudeOAuthSystemPromptBlocks                                string
}

// InitializeDefaultSettings 初始化默认设置
func (s *SettingService) InitializeDefaultSettings(ctx context.Context) error {
	// 检查是否已有设置
	_, err := s.settingRepo.GetValue(ctx, SettingKeyRegistrationEnabled)
	if err == nil {
		// 已有设置，不需要初始化
		return nil
	}
	if !errors.Is(err, ErrSettingNotFound) {
		return fmt.Errorf("check existing settings: %w", err)
	}
	loginAgreementDocumentsJSON, err := marshalLoginAgreementDocuments(defaultLoginAgreementDocuments())
	if err != nil {
		return err
	}

	// 初始化默认设置
	defaults := map[string]string{
		SettingKeyRegistrationEnabled:                      "true",
		SettingKeyEmailVerifyEnabled:                       "false",
		SettingKeyRegistrationEmailSuffixWhitelist:         "[]",
		SettingKeyPromoCodeEnabled:                         "true", // 默认启用优惠码功能
		SettingKeyPasswordMinLength:                        strconv.Itoa(DefaultPasswordMinLength),
		SettingKeyLoginAgreementEnabled:                    "false",
		SettingKeyLoginAgreementMode:                       defaultLoginAgreementMode,
		SettingKeyLoginAgreementUpdatedAt:                  defaultLoginAgreementDate,
		SettingKeyLoginAgreementDocuments:                  loginAgreementDocumentsJSON,
		SettingKeySiteName:                                 "Cloudbase",
		SettingKeySiteLogo:                                 "",
		SettingKeyDocsContentBasePath:                      defaultDocsContentBasePath,
		SettingKeyHomeShellConfig:                          defaultHomeShellConfig,
		SettingKeyHomeBusinessShellConfig:                  defaultHomeBusinessShellConfig,
		SettingKeyModelPlazaItems:                          "[]",
		SettingKeyImageWorkspaceModelConfig:                defaultImageWorkspaceModelConfig,
		SettingKeyModelPlazaShellConfig:                    defaultModelPlazaShellConfig,
		SettingKeyDocsShellConfig:                          defaultDocsShellConfig,
		SettingKeyLegalDocumentShellConfig:                 defaultLegalDocumentShellConfig,
		SettingKeyAPIKeysShellConfig:                       apiKeysShellConfigDefault(),
		SettingKeyKeyUsageShellConfig:                      defaultKeyUsageShellConfig,
		SettingKeyDashboardShellConfig:                     defaultDashboardShellConfig,
		SettingKeyUsageShellConfig:                         usageShellConfigDefault(),
		SettingKeyAPIGuideShellConfig:                      defaultAPIGuideShellConfig,
		SettingKeyAPITestShellConfig:                       defaultAPITestShellConfig,
		SettingKeyAvailableGroupsShellConfig:               defaultAvailableGroupsShellConfig,
		SettingKeyRedeemShellConfig:                        defaultRedeemShellConfig,
		SettingKeyAffiliateShellConfig:                     defaultAffiliateShellConfig,
		SettingKeyAvailableChannelsShellConfig:             defaultAvailableChannelsShellConfig,
		SettingKeyChannelStatusShellConfig:                 defaultChannelStatusShellConfig,
		SettingKeyCustomPageShellConfig:                    defaultCustomPageShellConfig,
		SettingKeyProfileShellConfig:                       profileShellConfigDefault(),
		SettingKeyAuthShellConfig:                          defaultAuthShellConfig,
		SettingKeyPurchaseSubscriptionEnabled:              "false",
		SettingKeyPurchaseSubscriptionURL:                  "",
		SettingKeyTableDefaultPageSize:                     "20",
		SettingKeyTablePageSizeOptions:                     "[10,20,50,100]",
		SettingKeyCustomMenuItems:                          "[]",
		SettingKeyCustomEndpoints:                          "[]",
		SettingKeyWebAppURL:                                "",
		SettingKeyWebAppName:                               "",
		SettingKeyWebAppDescription:                        "",
		SettingKeyWebAppLogo:                               "",
		SettingKeyWebAppFavicon:                            "",
		SettingKeyWebAppPreviewImage:                       "",
		SettingKeyWebTheme:                                 "",
		SettingKeyWebAppearance:                            "",
		SettingKeyWebDefaultLocale:                         "",
		SettingKeyPromptCasesTitle:                         "",
		SettingKeyPromptCasesDescription:                   "",
		SettingKeyPromptTemplatesTitle:                     "",
		SettingKeyPromptTemplatesDescription:               "",
		SettingKeyPromptCatalogShellConfig:                 defaultPromptCatalogShellConfig,
		SettingKeyWorkspaceShellConfig:                     defaultWorkspaceShellConfig,
		SettingKeyPricingTitle:                             "",
		SettingKeyPricingDescription:                       "",
		SettingKeyPricingShellConfig:                       defaultPricingShellConfig,
		SettingKeyPaymentShellConfig:                       defaultPaymentShellConfig,
		SettingKeyPricingCurrencySymbol:                    "¥",
		SettingKeyCreditsTitle:                             "",
		SettingKeyCreditsDescription:                       "",
		SettingKeyCreditsPurchaseLabel:                     "",
		SettingKeyCreditsBalanceLabel:                      "",
		SettingKeyCreditsPerBalance:                        "1",
		SettingKeyCreditsShellConfig:                       defaultCreditsShellConfig,
		SettingKeyWebLocaleDetectEnabled:                   "false",
		SettingKeyWebEmailAuthVisible:                      "true",
		SettingKeyWebGoogleAuthVisible:                     "false",
		SettingKeyWebGitHubAuthVisible:                     "false",
		SettingKeyWebGoogleAnalyticsID:                     "",
		SettingKeyWebClarityID:                             "",
		SettingKeyWebPlausibleDomain:                       "",
		SettingKeyWebPlausibleSrc:                          "",
		SettingKeyWebOpenPanelClientID:                     "",
		SettingKeyWebPublicIntegrationsEnabled:             "true",
		SettingKeyWebVercelAnalyticsEnabled:                "false",
		SettingKeyWebAdsenseCode:                           "",
		SettingKeyWebAffonsoEnabled:                        "false",
		SettingKeyWebAffonsoID:                             "",
		SettingKeyWebAffonsoCookieDuration:                 defaultWebAffonsoCookieDuration,
		SettingKeyWebPromoteKitEnabled:                     "false",
		SettingKeyWebPromoteKitID:                          "",
		SettingKeyWebCrispEnabled:                          "false",
		SettingKeyWebCrispWebsiteID:                        "",
		SettingKeyWebTawkEnabled:                           "false",
		SettingKeyWebTawkPropertyID:                        "",
		SettingKeyWebTawkWidgetID:                          "",
		SettingKeyWeChatExportFetchRetries:                 "2",
		SettingKeyWeChatExportFetchTimeoutMS:               "20000",
		SettingKeyWeChatExportWorkerConcurrency:            "1",
		SettingKeyWeChatExportWorkerIntervalMS:             "5000",
		SettingKeyWeChatExportWorkerLeaseSeconds:           "300",
		SettingKeyWeChatExportWorkerMaxBackoffMS:           "60000",
		SettingKeyImageWorkspaceUpstreamURL:                "https://api.openai.com/v1/images/generations",
		SettingKeyImageWorkspaceGenerationTimeoutMS:        "420000",
		SettingKeyImageWorkspaceCompletionCost:             "0",
		SettingKeyImageWorkspaceCompletionCostMapJSON:      "{}",
		SettingKeyImageWorkspacePromptSafetyEnabled:        "true",
		SettingKeyImageWorkspaceAssumeWorkerReady:          "false",
		SettingKeyImageWorkspaceObjectStorageEnabled:       "false",
		SettingKeyImageWorkspaceObjectStorageProvider:      "r2",
		SettingKeyImageWorkspaceObjectStorageBucket:        "",
		SettingKeyImageWorkspaceObjectStorageRegion:        "auto",
		SettingKeyImageWorkspaceObjectStoragePrefix:        "image-workspace",
		SettingKeyImageWorkspaceObjectStoragePublicBaseURL: "",
		SettingKeyMediaCDNBaseURL:                          "",
		SettingKeyWeChatConnectEnabled:                     "false",
		SettingKeyWeChatConnectOpenAppID:                   "",
		SettingKeyWeChatConnectOpenAppSecret:               "",
		SettingKeyWeChatConnectMPAppID:                     "",
		SettingKeyWeChatConnectMPAppSecret:                 "",
		SettingKeyWeChatConnectMobileAppID:                 "",
		SettingKeyWeChatConnectMobileAppSecret:             "",
		SettingKeyWeChatConnectOpenEnabled:                 "false",
		SettingKeyWeChatConnectMPEnabled:                   "false",
		SettingKeyWeChatConnectMobileEnabled:               "false",
		SettingKeyWeChatConnectMode:                        "open",
		SettingKeyWeChatConnectScopes:                      "snsapi_login",
		SettingKeyWeChatConnectRedirectURL:                 "",
		SettingKeyWeChatConnectFrontendRedirectURL:         defaultWeChatConnectFrontend,
		SettingKeyGitHubOAuthEnabled:                       "false",
		SettingKeyGitHubOAuthClientID:                      "",
		SettingKeyGitHubOAuthClientSecret:                  "",
		SettingKeyGitHubOAuthRedirectURL:                   "",
		SettingKeyGitHubOAuthFrontendRedirectURL:           defaultGitHubOAuthFrontend,
		SettingKeyGoogleOAuthEnabled:                       "false",
		SettingKeyGoogleOAuthClientID:                      "",
		SettingKeyGoogleOAuthClientSecret:                  "",
		SettingKeyGoogleOAuthRedirectURL:                   "",
		SettingKeyGoogleOAuthFrontendRedirectURL:           defaultGoogleOAuthFrontend,
		SettingKeyDefaultConcurrency:                       strconv.Itoa(s.cfg.Default.UserConcurrency),
		SettingKeyDefaultBalance:                           strconv.FormatFloat(s.cfg.Default.UserBalance, 'f', 8, 64),
		SettingKeyAffiliateRebateRate:                      strconv.FormatFloat(AffiliateRebateRateDefault, 'f', 8, 64),
		SettingKeyAffiliateRebateFreezeHours:               strconv.Itoa(AffiliateRebateFreezeHoursDefault),
		SettingKeyAffiliateRebateDurationDays:              strconv.Itoa(AffiliateRebateDurationDaysDefault),
		SettingKeyAffiliateRebatePerInviteeCap:             strconv.FormatFloat(AffiliateRebatePerInviteeCapDefault, 'f', 2, 64),
		SettingKeyDefaultUserRPMLimit:                      "0",
		SettingKeyDefaultSubscriptions:                     "[]",
		SettingKeyAuthSourceDefaultEmailBalance:            "0",
		SettingKeyAuthSourceDefaultEmailConcurrency:        "5",
		SettingKeyAuthSourceDefaultEmailSubscriptions:      "[]",
		SettingKeyAuthSourceDefaultEmailGrantOnSignup:      "false",
		SettingKeyAuthSourceDefaultEmailGrantOnFirstBind:   "false",
		SettingKeyAuthSourceDefaultGitHubBalance:           "0",
		SettingKeyAuthSourceDefaultGitHubConcurrency:       "5",
		SettingKeyAuthSourceDefaultGitHubSubscriptions:     "[]",
		SettingKeyAuthSourceDefaultGitHubGrantOnSignup:     "false",
		SettingKeyAuthSourceDefaultGitHubGrantOnFirstBind:  "false",
		SettingKeyAuthSourceDefaultGoogleBalance:           "0",
		SettingKeyAuthSourceDefaultGoogleConcurrency:       "5",
		SettingKeyAuthSourceDefaultGoogleSubscriptions:     "[]",
		SettingKeyAuthSourceDefaultGoogleGrantOnSignup:     "false",
		SettingKeyAuthSourceDefaultGoogleGrantOnFirstBind:  "false",
		SettingKeyForceEmailOnThirdPartySignup:             "false",
		SettingKeySMTPPort:                                 "587",
		SettingKeySMTPUseTLS:                               "false",
		SettingKeySMTPDailyLimit:                           "0",
		SettingKeySMTPChannels:                             "[]",
		// Model fallback defaults
		SettingKeyEnableModelFallback:      "false",
		SettingKeyFallbackModelAnthropic:   "claude-3-5-sonnet-20241022",
		SettingKeyFallbackModelOpenAI:      "gpt-4o",
		SettingKeyFallbackModelGemini:      "gemini-2.5-pro",
		SettingKeyFallbackModelAntigravity: "gemini-2.5-pro",
		// Identity patch defaults
		SettingKeyEnableIdentityPatch: "true",
		SettingKeyIdentityPatchPrompt: "",

		// Ops monitoring defaults (vNext)
		SettingKeyOpsMonitoringEnabled:         "true",
		SettingKeyOpsRealtimeMonitoringEnabled: "true",
		SettingKeyOpsQueryModeDefault:          "auto",
		SettingKeyOpsMetricsIntervalSeconds:    "60",

		// Channel monitor defaults (enabled, 60s)
		SettingKeyChannelMonitorEnabled:                "true",
		SettingKeyChannelMonitorDefaultIntervalSeconds: "60",

		// Available channels feature (default disabled; opt-in)
		SettingKeyAvailableChannelsEnabled: "false",

		// Affiliate (邀请返利) feature (default disabled; opt-in)
		SettingKeyAffiliateEnabled: "false",

		// 风控中心功能（默认关闭，显式启用）
		SettingKeyRiskControlEnabled:                         "false",
		SettingKeySignupGrantRiskControlEnabled:              "false",
		SettingKeySignupGrantRiskControlEmailLimit:           "1",
		SettingKeySignupGrantRiskControlIPLimit:              "3",
		SettingKeySignupGrantRiskControlDomainLimit:          "10",
		SettingKeySignupGrantRiskControlOAuthIdentityEnabled: "true",
		SettingKeySignupGrantRiskControlDeviceEnabled:        "true",
		SettingKeySignupGrantRiskControlDeviceLimit:          "2",
		SettingKeySignupGrantRiskControlFreeDomainLimit:      "5",
		SettingKeySignupGrantRiskControlBlockedDomains:       "",
		SettingKeySignupGrantRiskControlFreeDomains:          defaultSignupGrantRiskFreeDomains,
		SettingKeySignupGrantRiskControlTrustedDomains:       "",

		// cyber 会话屏蔽（默认关闭，TTL 默认 3600s）
		SettingKeyCyberSessionBlockEnabled:    "false",
		SettingKeyCyberSessionBlockTTLSeconds: "3600",

		// Claude Code version check (default: empty = disabled)
		SettingKeyMinClaudeCodeVersion: "",
		SettingKeyMaxClaudeCodeVersion: "",

		// codex_cli_only 加固（默认：版本不检查、名单空、默认种子指纹信号）
		SettingKeyMinCodexVersion:                      "",
		SettingKeyMaxCodexVersion:                      "",
		SettingKeyCodexCLIOnlyBlacklist:                "",
		SettingKeyCodexCLIOnlyWhitelist:                "",
		SettingKeyCodexCLIOnlyAllowAppServerClients:    "false",
		SettingKeyCodexCLIOnlyEngineFingerprintSignals: openai.DefaultEngineFingerprintSignalsJSON(),

		// 分组隔离（默认不允许未分组 Key 调度）
		SettingKeyAllowUngroupedKeyScheduling:                        "false",
		SettingKeyEnableAnthropicCacheTTL1hInjection:                 "false",
		SettingKeyRewriteMessageCacheControl:                         strconv.FormatBool(s.defaultRewriteMessageCacheControl()),
		SettingKeyEnableClientDatelineNormalization:                  "true",
		SettingKeyAntigravityUserAgentVersion:                        "",
		SettingKeyOpenAICodexUserAgent:                               "",
		SettingPaymentVisibleMethodAlipaySource:                      "",
		SettingPaymentVisibleMethodWxpaySource:                       "",
		SettingPaymentVisibleMethodAlipayEnabled:                     "false",
		SettingPaymentVisibleMethodWxpayEnabled:                      "false",
		openAIAdvancedSchedulerSettingKey:                            "false",
		SettingKeyOpenAIAdvancedSchedulerStickyWeightedEnabled:       "false",
		SettingKeyOpenAIAdvancedSchedulerSubscriptionPriorityEnabled: "false",
		SettingKeyOpenAIAdvancedSchedulerLBTopK:                      "",
		SettingKeyOpenAIAdvancedSchedulerWeightPriority:              "",
		SettingKeyOpenAIAdvancedSchedulerWeightLoad:                  "",
		SettingKeyOpenAIAdvancedSchedulerWeightQueue:                 "",
		SettingKeyOpenAIAdvancedSchedulerWeightErrorRate:             "",
		SettingKeyOpenAIAdvancedSchedulerWeightTTFT:                  "",
		SettingKeyOpenAIAdvancedSchedulerWeightReset:                 "",
		SettingKeyOpenAIAdvancedSchedulerWeightQuotaHeadroom:         "",
		SettingKeyOpenAIAdvancedSchedulerWeightPreviousResponse:      "",
		SettingKeyOpenAIAdvancedSchedulerWeightSessionSticky:         "",

		SettingKeyAllowUserViewErrorRequests: "false",
	}

	return s.settingRepo.SetMultiple(ctx, defaults)
}

// parseSettings 解析设置到结构体
func (s *SettingService) parseSettings(settings map[string]string) *SystemSettings {
	emailVerifyEnabled := settings[SettingKeyEmailVerifyEnabled] == "true"
	loginAgreementDocuments := parseLoginAgreementDocuments(settings[SettingKeyLoginAgreementDocuments])
	loginAgreementUpdatedAt := strings.TrimSpace(settings[SettingKeyLoginAgreementUpdatedAt])
	if loginAgreementUpdatedAt == "" {
		loginAgreementUpdatedAt = defaultLoginAgreementDate
	}
	apiKeyACLTrustForwardedIP := false
	if value, ok := settings[SettingKeyAPIKeyACLTrustForwardedIP]; ok {
		apiKeyACLTrustForwardedIP = value == "true"
	} else if s != nil && s.cfg != nil {
		apiKeyACLTrustForwardedIP = s.cfg.Security.TrustForwardedIPForAPIKeyACL
	}
	result := &SystemSettings{
		RegistrationEnabled:                      settings[SettingKeyRegistrationEnabled] == "true",
		EmailVerifyEnabled:                       emailVerifyEnabled,
		RegistrationEmailSuffixWhitelist:         ParseRegistrationEmailSuffixWhitelist(settings[SettingKeyRegistrationEmailSuffixWhitelist]),
		PromoCodeEnabled:                         settings[SettingKeyPromoCodeEnabled] != "false", // 默认启用
		PasswordResetEnabled:                     emailVerifyEnabled && settings[SettingKeyPasswordResetEnabled] == "true",
		PasswordMinLength:                        parsePasswordMinLength(settings[SettingKeyPasswordMinLength]),
		FrontendURL:                              settings[SettingKeyFrontendURL],
		InvitationCodeEnabled:                    settings[SettingKeyInvitationCodeEnabled] == "true",
		TotpEnabled:                              settings[SettingKeyTotpEnabled] == "true",
		LoginAgreementEnabled:                    settings[SettingKeyLoginAgreementEnabled] == "true",
		LoginAgreementMode:                       normalizeLoginAgreementMode(settings[SettingKeyLoginAgreementMode]),
		LoginAgreementUpdatedAt:                  loginAgreementUpdatedAt,
		LoginAgreementDocuments:                  loginAgreementDocuments,
		SMTPHost:                                 settings[SettingKeySMTPHost],
		SMTPUsername:                             settings[SettingKeySMTPUsername],
		SMTPFrom:                                 settings[SettingKeySMTPFrom],
		SMTPFromName:                             settings[SettingKeySMTPFromName],
		SMTPUseTLS:                               settings[SettingKeySMTPUseTLS] == "true",
		SMTPDailyLimit:                           parseSMTPDailyLimit(settings[SettingKeySMTPDailyLimit]),
		SMTPChannels:                             parseSMTPChannels(settings[SettingKeySMTPChannels]),
		SMTPPasswordConfigured:                   settings[SettingKeySMTPPassword] != "",
		TurnstileEnabled:                         settings[SettingKeyTurnstileEnabled] == "true",
		TurnstileSiteKey:                         settings[SettingKeyTurnstileSiteKey],
		TurnstileSecretKeyConfigured:             settings[SettingKeyTurnstileSecretKey] != "",
		APIKeyACLTrustForwardedIP:                apiKeyACLTrustForwardedIP,
		SiteName:                                 s.getStringOrDefault(settings, SettingKeySiteName, "Cloudbase"),
		SiteLogo:                                 settings[SettingKeySiteLogo],
		SiteSubtitle:                             s.getStringOrDefault(settings, SettingKeySiteSubtitle, "AI Gateway and Business Operations Platform"),
		APIBaseURL:                               settings[SettingKeyAPIBaseURL],
		ContactInfo:                              settings[SettingKeyContactInfo],
		DocURL:                                   settings[SettingKeyDocURL],
		DocsContentBasePath:                      docsContentBasePathSetting(settings[SettingKeyDocsContentBasePath]),
		HomeContent:                              settings[SettingKeyHomeContent],
		HomeShellConfig:                          homeShellConfigSetting(settings[SettingKeyHomeShellConfig]),
		HomeBusinessShellConfig:                  homeBusinessShellConfigSetting(settings[SettingKeyHomeBusinessShellConfig]),
		ModelPlazaItems:                          settings[SettingKeyModelPlazaItems],
		ImageWorkspaceModelConfig:                imageWorkspaceModelConfigSetting(settings[SettingKeyImageWorkspaceModelConfig]),
		ModelPlazaShellConfig:                    modelPlazaShellConfigSetting(settings[SettingKeyModelPlazaShellConfig]),
		DocsShellConfig:                          docsShellConfigSetting(settings[SettingKeyDocsShellConfig]),
		LegalDocumentShellConfig:                 legalDocumentShellConfigSetting(settings[SettingKeyLegalDocumentShellConfig]),
		APIKeysShellConfig:                       apiKeysShellConfigSetting(settings[SettingKeyAPIKeysShellConfig]),
		KeyUsageShellConfig:                      keyUsageShellConfigSetting(settings[SettingKeyKeyUsageShellConfig]),
		DashboardShellConfig:                     dashboardShellConfigSetting(settings[SettingKeyDashboardShellConfig]),
		UsageShellConfig:                         usageShellConfigSetting(settings[SettingKeyUsageShellConfig]),
		APIGuideShellConfig:                      apiGuideShellConfigSetting(settings[SettingKeyAPIGuideShellConfig]),
		APITestShellConfig:                       apiTestShellConfigSetting(settings[SettingKeyAPITestShellConfig]),
		AvailableGroupsShellConfig:               availableGroupsShellConfigSetting(settings[SettingKeyAvailableGroupsShellConfig]),
		RedeemShellConfig:                        redeemShellConfigSetting(settings[SettingKeyRedeemShellConfig]),
		AffiliateShellConfig:                     affiliateShellConfigSetting(settings[SettingKeyAffiliateShellConfig]),
		AvailableChannelsShellConfig:             availableChannelsShellConfigSetting(settings[SettingKeyAvailableChannelsShellConfig]),
		ChannelStatusShellConfig:                 channelStatusShellConfigSetting(settings[SettingKeyChannelStatusShellConfig]),
		CustomPageShellConfig:                    customPageShellConfigSetting(settings[SettingKeyCustomPageShellConfig]),
		ProfileShellConfig:                       profileShellConfigSetting(settings[SettingKeyProfileShellConfig]),
		AuthShellConfig:                          authShellConfigSetting(settings[SettingKeyAuthShellConfig]),
		HideCcsImportButton:                      settings[SettingKeyHideCcsImportButton] == "true",
		PurchaseSubscriptionEnabled:              settings[SettingKeyPurchaseSubscriptionEnabled] == "true",
		PurchaseSubscriptionURL:                  strings.TrimSpace(settings[SettingKeyPurchaseSubscriptionURL]),
		CustomMenuItems:                          settings[SettingKeyCustomMenuItems],
		CustomEndpoints:                          settings[SettingKeyCustomEndpoints],
		WebAppURL:                                strings.TrimSpace(settings[SettingKeyWebAppURL]),
		WebAppName:                               strings.TrimSpace(settings[SettingKeyWebAppName]),
		WebAppDescription:                        strings.TrimSpace(settings[SettingKeyWebAppDescription]),
		WebAppLogo:                               strings.TrimSpace(settings[SettingKeyWebAppLogo]),
		WebAppFavicon:                            strings.TrimSpace(settings[SettingKeyWebAppFavicon]),
		WebAppPreviewImage:                       strings.TrimSpace(settings[SettingKeyWebAppPreviewImage]),
		WebTheme:                                 strings.TrimSpace(settings[SettingKeyWebTheme]),
		WebAppearance:                            strings.TrimSpace(settings[SettingKeyWebAppearance]),
		WebDefaultLocale:                         strings.TrimSpace(settings[SettingKeyWebDefaultLocale]),
		WebPromptCasesTitle:                      strings.TrimSpace(settings[SettingKeyPromptCasesTitle]),
		WebPromptCasesDescription:                strings.TrimSpace(settings[SettingKeyPromptCasesDescription]),
		WebPromptTemplatesTitle:                  strings.TrimSpace(settings[SettingKeyPromptTemplatesTitle]),
		WebPromptTemplatesDescription:            strings.TrimSpace(settings[SettingKeyPromptTemplatesDescription]),
		PromptCatalogShellConfig:                 promptCatalogShellConfigSetting(settings[SettingKeyPromptCatalogShellConfig]),
		WebWorkspaceShellConfig:                  workspaceShellConfigSetting(settings[SettingKeyWorkspaceShellConfig]),
		ImagePromptFilterConfig:                  strings.TrimSpace(settings[SettingKeyImagePromptFilterConfig]),
		WebPricingTitle:                          strings.TrimSpace(settings[SettingKeyPricingTitle]),
		WebPricingDescription:                    strings.TrimSpace(settings[SettingKeyPricingDescription]),
		WebPricingShellConfig:                    pricingShellConfigSetting(settings[SettingKeyPricingShellConfig]),
		WebPaymentShellConfig:                    paymentShellConfigSetting(settings[SettingKeyPaymentShellConfig]),
		WebPricingCurrencySymbol:                 pricingCurrencySymbolSetting(settings[SettingKeyPricingCurrencySymbol]),
		WebCreditsTitle:                          strings.TrimSpace(settings[SettingKeyCreditsTitle]),
		WebCreditsDescription:                    strings.TrimSpace(settings[SettingKeyCreditsDescription]),
		WebCreditsPurchaseLabel:                  strings.TrimSpace(settings[SettingKeyCreditsPurchaseLabel]),
		WebCreditsBalanceLabel:                   strings.TrimSpace(settings[SettingKeyCreditsBalanceLabel]),
		WebCreditsPerBalance:                     creditsPerBalanceSetting(settings[SettingKeyCreditsPerBalance]),
		CreditsShellConfig:                       creditsShellConfigSetting(settings[SettingKeyCreditsShellConfig]),
		WebLocaleDetectEnabled:                   settings[SettingKeyWebLocaleDetectEnabled] == "true",
		WebEmailAuthVisible:                      parseBoolSettingWithDefault(settings[SettingKeyWebEmailAuthVisible], true),
		WebGoogleAuthVisible:                     settings[SettingKeyWebGoogleAuthVisible] == "true",
		WebGitHubAuthVisible:                     settings[SettingKeyWebGitHubAuthVisible] == "true",
		WebGoogleAnalyticsID:                     strings.TrimSpace(settings[SettingKeyWebGoogleAnalyticsID]),
		WebClarityID:                             strings.TrimSpace(settings[SettingKeyWebClarityID]),
		WebPlausibleDomain:                       strings.TrimSpace(settings[SettingKeyWebPlausibleDomain]),
		WebPlausibleSrc:                          strings.TrimSpace(settings[SettingKeyWebPlausibleSrc]),
		WebOpenPanelClientID:                     strings.TrimSpace(settings[SettingKeyWebOpenPanelClientID]),
		WebPublicIntegrationsEnabled:             !isFalseSettingValue(settings[SettingKeyWebPublicIntegrationsEnabled]),
		WebVercelAnalyticsEnabled:                settings[SettingKeyWebVercelAnalyticsEnabled] == "true",
		WebAdsenseCode:                           strings.TrimSpace(settings[SettingKeyWebAdsenseCode]),
		WebAffonsoEnabled:                        settings[SettingKeyWebAffonsoEnabled] == "true",
		WebAffonsoID:                             strings.TrimSpace(settings[SettingKeyWebAffonsoID]),
		WebAffonsoCookieDuration:                 webAffonsoCookieDurationSetting(settings[SettingKeyWebAffonsoCookieDuration]),
		WebPromoteKitEnabled:                     settings[SettingKeyWebPromoteKitEnabled] == "true",
		WebPromoteKitID:                          strings.TrimSpace(settings[SettingKeyWebPromoteKitID]),
		WebCrispEnabled:                          settings[SettingKeyWebCrispEnabled] == "true",
		WebCrispWebsiteID:                        strings.TrimSpace(settings[SettingKeyWebCrispWebsiteID]),
		WebTawkEnabled:                           settings[SettingKeyWebTawkEnabled] == "true",
		WebTawkPropertyID:                        strings.TrimSpace(settings[SettingKeyWebTawkPropertyID]),
		WebTawkWidgetID:                          strings.TrimSpace(settings[SettingKeyWebTawkWidgetID]),
		WeChatExportFetchRetries:                 parseBoundedIntSetting(settings[SettingKeyWeChatExportFetchRetries], 2, 0, 5),
		WeChatExportFetchTimeoutMS:               parseBoundedIntSetting(settings[SettingKeyWeChatExportFetchTimeoutMS], 20000, 1000, 120000),
		WeChatExportWorkerConcurrency:            parseBoundedIntSetting(settings[SettingKeyWeChatExportWorkerConcurrency], 1, 1, 8),
		WeChatExportWorkerIntervalMS:             parseBoundedIntSetting(settings[SettingKeyWeChatExportWorkerIntervalMS], 5000, 1000, 300000),
		WeChatExportWorkerLeaseSeconds:           parseBoundedIntSetting(settings[SettingKeyWeChatExportWorkerLeaseSeconds], 300, 60, 3600),
		WeChatExportWorkerMaxBackoffMS:           parseBoundedIntSetting(settings[SettingKeyWeChatExportWorkerMaxBackoffMS], 60000, 1000, 300000),
		ImageWorkspaceUpstreamURL:                strings.TrimSpace(settings[SettingKeyImageWorkspaceUpstreamURL]),
		ImageWorkspaceGenerationTimeoutMS:        parseBoundedIntSetting(settings[SettingKeyImageWorkspaceGenerationTimeoutMS], 420000, 1000, 900000),
		ImageWorkspaceCompletionCost:             strings.TrimSpace(settings[SettingKeyImageWorkspaceCompletionCost]),
		ImageWorkspaceCompletionCostMapJSON:      normalizeJSONObjectSetting(settings[SettingKeyImageWorkspaceCompletionCostMapJSON], "{}"),
		ImageWorkspacePromptSafetyEnabled:        parseBoolSettingWithDefault(settings[SettingKeyImageWorkspacePromptSafetyEnabled], true),
		ImageWorkspaceAssumeWorkerReady:          settings[SettingKeyImageWorkspaceAssumeWorkerReady] == "true",
		ImageWorkspaceObjectStorageEnabled:       settings[SettingKeyImageWorkspaceObjectStorageEnabled] == "true",
		ImageWorkspaceObjectStorageProvider:      strings.TrimSpace(settings[SettingKeyImageWorkspaceObjectStorageProvider]),
		ImageWorkspaceObjectStorageBucket:        strings.TrimSpace(settings[SettingKeyImageWorkspaceObjectStorageBucket]),
		ImageWorkspaceObjectStorageRegion:        strings.TrimSpace(settings[SettingKeyImageWorkspaceObjectStorageRegion]),
		ImageWorkspaceObjectStoragePrefix:        strings.TrimSpace(settings[SettingKeyImageWorkspaceObjectStoragePrefix]),
		ImageWorkspaceObjectStoragePublicBaseURL: strings.TrimSpace(settings[SettingKeyImageWorkspaceObjectStoragePublicBaseURL]),
		MediaCDNBaseURL:                          strings.TrimSpace(settings[SettingKeyMediaCDNBaseURL]),
		BackendModeEnabled:                       settings[SettingKeyBackendModeEnabled] == "true",
	}
	result.TableDefaultPageSize, result.TablePageSizeOptions = parseTablePreferences(
		settings[SettingKeyTableDefaultPageSize],
		settings[SettingKeyTablePageSizeOptions],
	)

	// 解析整数类型
	if port, err := strconv.Atoi(settings[SettingKeySMTPPort]); err == nil {
		result.SMTPPort = port
	} else {
		result.SMTPPort = 587
	}

	if concurrency, err := strconv.Atoi(settings[SettingKeyDefaultConcurrency]); err == nil {
		result.DefaultConcurrency = concurrency
	} else {
		result.DefaultConcurrency = s.cfg.Default.UserConcurrency
	}

	if rpm, err := strconv.Atoi(settings[SettingKeyDefaultUserRPMLimit]); err == nil && rpm >= 0 {
		result.DefaultUserRPMLimit = rpm
	}

	// 解析浮点数类型
	if balance, err := strconv.ParseFloat(settings[SettingKeyDefaultBalance], 64); err == nil {
		result.DefaultBalance = balance
	} else {
		result.DefaultBalance = s.cfg.Default.UserBalance
	}
	if rebateRate, err := strconv.ParseFloat(settings[SettingKeyAffiliateRebateRate], 64); err == nil {
		result.AffiliateRebateRate = clampAffiliateRebateRate(rebateRate)
	} else {
		result.AffiliateRebateRate = AffiliateRebateRateDefault
	}
	if freezeHours, err := strconv.Atoi(settings[SettingKeyAffiliateRebateFreezeHours]); err == nil && freezeHours >= 0 {
		if freezeHours > AffiliateRebateFreezeHoursMax {
			freezeHours = AffiliateRebateFreezeHoursMax
		}
		result.AffiliateRebateFreezeHours = freezeHours
	}
	if durationDays, err := strconv.Atoi(settings[SettingKeyAffiliateRebateDurationDays]); err == nil && durationDays >= 0 {
		if durationDays > AffiliateRebateDurationDaysMax {
			durationDays = AffiliateRebateDurationDaysMax
		}
		result.AffiliateRebateDurationDays = durationDays
	}
	if perInviteeCap, err := strconv.ParseFloat(settings[SettingKeyAffiliateRebatePerInviteeCap], 64); err == nil && perInviteeCap >= 0 {
		result.AffiliateRebatePerInviteeCap = perInviteeCap
	}
	result.DefaultSubscriptions = parseDefaultSubscriptions(settings[SettingKeyDefaultSubscriptions])

	// 敏感信息直接返回，方便测试连接时使用
	result.SMTPPassword = settings[SettingKeySMTPPassword]
	result.TurnstileSecretKey = settings[SettingKeyTurnstileSecretKey]

	gitHubEffective := s.effectiveEmailOAuthConfig(settings, "github")
	result.GitHubOAuthEnabled = gitHubEffective.Enabled
	result.GitHubOAuthClientID = strings.TrimSpace(gitHubEffective.ClientID)
	result.GitHubOAuthClientSecret = strings.TrimSpace(gitHubEffective.ClientSecret)
	result.GitHubOAuthClientSecretConfigured = result.GitHubOAuthClientSecret != ""
	result.GitHubOAuthRedirectURL = strings.TrimSpace(gitHubEffective.RedirectURL)
	result.GitHubOAuthFrontendRedirectURL = strings.TrimSpace(gitHubEffective.FrontendRedirectURL)

	googleEffective := s.effectiveEmailOAuthConfig(settings, "google")
	result.GoogleOAuthEnabled = googleEffective.Enabled
	result.GoogleOAuthClientID = strings.TrimSpace(googleEffective.ClientID)
	result.GoogleOAuthClientSecret = strings.TrimSpace(googleEffective.ClientSecret)
	result.GoogleOAuthClientSecretConfigured = result.GoogleOAuthClientSecret != ""
	result.GoogleOAuthRedirectURL = strings.TrimSpace(googleEffective.RedirectURL)
	result.GoogleOAuthFrontendRedirectURL = strings.TrimSpace(googleEffective.FrontendRedirectURL)

	// WeChat Connect 设置：
	// - 优先读取 DB 系统设置
	// - 缺失时回退到 config/env，保持升级兼容
	weChatEffective := s.effectiveWeChatConnectOAuthConfig(settings)
	result.WeChatConnectEnabled = weChatEffective.Enabled
	result.WeChatConnectOpenAppID = weChatEffective.OpenAppID
	result.WeChatConnectOpenAppSecret = weChatEffective.OpenAppSecret
	result.WeChatConnectOpenAppSecretConfigured = weChatEffective.OpenAppSecret != ""
	result.WeChatConnectMPAppID = weChatEffective.MPAppID
	result.WeChatConnectMPAppSecret = weChatEffective.MPAppSecret
	result.WeChatConnectMPAppSecretConfigured = weChatEffective.MPAppSecret != ""
	result.WeChatConnectMobileAppID = weChatEffective.MobileAppID
	result.WeChatConnectMobileAppSecret = weChatEffective.MobileAppSecret
	result.WeChatConnectMobileAppSecretConfigured = weChatEffective.MobileAppSecret != ""
	result.WeChatConnectOpenEnabled = weChatEffective.OpenEnabled
	result.WeChatConnectMPEnabled = weChatEffective.MPEnabled
	result.WeChatConnectMobileEnabled = weChatEffective.MobileEnabled
	result.WeChatConnectMode = weChatEffective.Mode
	result.WeChatConnectScopes = weChatEffective.Scopes
	result.WeChatConnectRedirectURL = weChatEffective.RedirectURL
	result.WeChatConnectFrontendRedirectURL = weChatEffective.FrontendRedirectURL

	// Model fallback settings
	result.EnableModelFallback = settings[SettingKeyEnableModelFallback] == "true"
	result.FallbackModelAnthropic = s.getStringOrDefault(settings, SettingKeyFallbackModelAnthropic, "claude-3-5-sonnet-20241022")
	result.FallbackModelOpenAI = s.getStringOrDefault(settings, SettingKeyFallbackModelOpenAI, "gpt-4o")
	result.FallbackModelGemini = s.getStringOrDefault(settings, SettingKeyFallbackModelGemini, "gemini-2.5-pro")
	result.FallbackModelAntigravity = s.getStringOrDefault(settings, SettingKeyFallbackModelAntigravity, "gemini-2.5-pro")

	// Identity patch settings (default: enabled, to preserve existing behavior)
	if v, ok := settings[SettingKeyEnableIdentityPatch]; ok && v != "" {
		result.EnableIdentityPatch = v == "true"
	} else {
		result.EnableIdentityPatch = true
	}
	result.IdentityPatchPrompt = settings[SettingKeyIdentityPatchPrompt]

	// Ops monitoring settings (default: enabled, fail-open)
	result.OpsMonitoringEnabled = !isFalseSettingValue(settings[SettingKeyOpsMonitoringEnabled])
	result.OpsRealtimeMonitoringEnabled = !isFalseSettingValue(settings[SettingKeyOpsRealtimeMonitoringEnabled])
	result.OpsQueryModeDefault = string(ParseOpsQueryMode(settings[SettingKeyOpsQueryModeDefault]))
	result.OpsMetricsIntervalSeconds = 60
	if raw := strings.TrimSpace(settings[SettingKeyOpsMetricsIntervalSeconds]); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			if v < 60 {
				v = 60
			}
			if v > 3600 {
				v = 3600
			}
			result.OpsMetricsIntervalSeconds = v
		}
	}

	// Channel monitor feature (default: enabled, 60s)
	result.ChannelMonitorEnabled = !isFalseSettingValue(settings[SettingKeyChannelMonitorEnabled])
	result.ChannelMonitorDefaultIntervalSeconds = parseChannelMonitorInterval(
		settings[SettingKeyChannelMonitorDefaultIntervalSeconds],
	)

	// Available channels feature (default: disabled; strict true)
	result.AvailableChannelsEnabled = settings[SettingKeyAvailableChannelsEnabled] == "true"

	// Affiliate (邀请返利) feature (default: disabled; strict true)
	result.AffiliateEnabled = settings[SettingKeyAffiliateEnabled] == "true"

	// 风控中心功能（默认关闭，严格 true 才启用）
	result.RiskControlEnabled = settings[SettingKeyRiskControlEnabled] == "true"
	result.SignupGrantRiskControlEnabled = settings[SettingKeySignupGrantRiskControlEnabled] == "true"
	result.SignupGrantRiskControlEmailLimit = parseNonNegativeIntSetting(settings[SettingKeySignupGrantRiskControlEmailLimit], 1)
	result.SignupGrantRiskControlIPLimit = parseNonNegativeIntSetting(settings[SettingKeySignupGrantRiskControlIPLimit], 3)
	result.SignupGrantRiskControlDomainLimit = parseNonNegativeIntSetting(settings[SettingKeySignupGrantRiskControlDomainLimit], 10)
	result.SignupGrantRiskControlOAuthIdentityEnabled = settings[SettingKeySignupGrantRiskControlOAuthIdentityEnabled] != "false"
	result.SignupGrantRiskControlDeviceEnabled = settings[SettingKeySignupGrantRiskControlDeviceEnabled] != "false"
	result.SignupGrantRiskControlDeviceLimit = parseNonNegativeIntSetting(settings[SettingKeySignupGrantRiskControlDeviceLimit], 2)
	result.SignupGrantRiskControlFreeDomainLimit = parseNonNegativeIntSetting(settings[SettingKeySignupGrantRiskControlFreeDomainLimit], 5)
	result.SignupGrantRiskControlBlockedDomains = normalizeDomainListSetting(settings[SettingKeySignupGrantRiskControlBlockedDomains])
	result.SignupGrantRiskControlFreeDomains = normalizeDomainListSetting(settings[SettingKeySignupGrantRiskControlFreeDomains])
	result.SignupGrantRiskControlTrustedDomains = normalizeDomainListSetting(settings[SettingKeySignupGrantRiskControlTrustedDomains])

	// cyber 会话屏蔽（默认关闭，TTL 默认 3600s）
	result.CyberSessionBlockEnabled = settings[SettingKeyCyberSessionBlockEnabled] == "true"
	if v, err := strconv.Atoi(strings.TrimSpace(settings[SettingKeyCyberSessionBlockTTLSeconds])); err == nil && v > 0 {
		result.CyberSessionBlockTTLSeconds = v
	} else {
		result.CyberSessionBlockTTLSeconds = 3600
	}

	// Claude Code version check
	result.MinClaudeCodeVersion = settings[SettingKeyMinClaudeCodeVersion]
	result.MaxClaudeCodeVersion = settings[SettingKeyMaxClaudeCodeVersion]

	// 分组隔离
	result.AllowUngroupedKeyScheduling = settings[SettingKeyAllowUngroupedKeyScheduling] == "true"

	// Gateway forwarding behavior (defaults: fingerprint=true, metadata_passthrough=false,
	// cch_signing=false, claude_oauth_system_prompt_injection=true)
	if v, ok := settings[SettingKeyEnableFingerprintUnification]; ok && v != "" {
		result.EnableFingerprintUnification = v == "true"
	} else {
		result.EnableFingerprintUnification = true // default: enabled (current behavior)
	}
	result.EnableMetadataPassthrough = settings[SettingKeyEnableMetadataPassthrough] == "true"
	result.EnableCCHSigning = settings[SettingKeyEnableCCHSigning] == "true"
	if v, ok := settings[SettingKeyEnableClaudeOAuthSystemPromptInjection]; ok && v != "" {
		result.EnableClaudeOAuthSystemPromptInjection = v == "true"
	} else {
		result.EnableClaudeOAuthSystemPromptInjection = true
	}
	result.ClaudeOAuthSystemPrompt = settings[SettingKeyClaudeOAuthSystemPrompt]
	result.ClaudeOAuthSystemPromptBlocks = settings[SettingKeyClaudeOAuthSystemPromptBlocks]
	result.EnableAnthropicCacheTTL1hInjection = settings[SettingKeyEnableAnthropicCacheTTL1hInjection] == "true"
	if v, ok := settings[SettingKeyRewriteMessageCacheControl]; ok && v != "" {
		result.RewriteMessageCacheControl = v == "true"
	} else {
		result.RewriteMessageCacheControl = s.defaultRewriteMessageCacheControl()
	}
	if v, ok := settings[SettingKeyEnableClientDatelineNormalization]; ok && v != "" {
		result.EnableClientDatelineNormalization = v == "true"
	} else {
		result.EnableClientDatelineNormalization = true
	}
	result.AntigravityUserAgentVersion = antigravity.NormalizeUserAgentVersion(settings[SettingKeyAntigravityUserAgentVersion])
	result.OpenAICodexUserAgent = strings.TrimSpace(settings[SettingKeyOpenAICodexUserAgent])
	// codex_cli_only 加固
	result.MinCodexVersion = settings[SettingKeyMinCodexVersion]
	result.MaxCodexVersion = settings[SettingKeyMaxCodexVersion]
	result.CodexCLIOnlyBlacklist = settings[SettingKeyCodexCLIOnlyBlacklist]
	result.CodexCLIOnlyWhitelist = settings[SettingKeyCodexCLIOnlyWhitelist]
	result.CodexCLIOnlyAllowAppServerClients = settings[SettingKeyCodexCLIOnlyAllowAppServerClients] == "true"
	if raw := strings.TrimSpace(settings[SettingKeyCodexCLIOnlyEngineFingerprintSignals]); raw != "" {
		result.CodexCLIOnlyEngineFingerprintSignals = raw
	} else {
		result.CodexCLIOnlyEngineFingerprintSignals = openai.DefaultEngineFingerprintSignalsJSON() // 缺失/空 → 展示默认种子
	}

	// Web search emulation: quick enabled check from the JSON config
	if raw := settings[SettingKeyWebSearchEmulationConfig]; raw != "" {
		var wsCfg WebSearchEmulationConfig
		if err := json.Unmarshal([]byte(raw), &wsCfg); err == nil {
			result.WebSearchEmulationEnabled = wsCfg.Enabled && len(wsCfg.Providers) > 0
		}
	}
	result.PaymentVisibleMethodAlipaySource = NormalizeVisibleMethodSource("alipay", settings[SettingPaymentVisibleMethodAlipaySource])
	result.PaymentVisibleMethodWxpaySource = NormalizeVisibleMethodSource("wxpay", settings[SettingPaymentVisibleMethodWxpaySource])
	result.PaymentVisibleMethodAlipayEnabled = settings[SettingPaymentVisibleMethodAlipayEnabled] == "true"
	result.PaymentVisibleMethodWxpayEnabled = settings[SettingPaymentVisibleMethodWxpayEnabled] == "true"
	result.OpenAIAdvancedSchedulerEnabled = settings[openAIAdvancedSchedulerSettingKey] == "true"
	result.OpenAIAdvancedSchedulerStickyWeightedEnabled = settings[SettingKeyOpenAIAdvancedSchedulerStickyWeightedEnabled] == "true"
	result.OpenAIAdvancedSchedulerSubscriptionPriorityEnabled = settings[SettingKeyOpenAIAdvancedSchedulerSubscriptionPriorityEnabled] == "true"
	result.OpenAIAdvancedSchedulerLBTopK = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerLBTopK])
	result.OpenAIAdvancedSchedulerWeightPriority = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightPriority])
	result.OpenAIAdvancedSchedulerWeightLoad = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightLoad])
	result.OpenAIAdvancedSchedulerWeightQueue = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightQueue])
	result.OpenAIAdvancedSchedulerWeightErrorRate = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightErrorRate])
	result.OpenAIAdvancedSchedulerWeightTTFT = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightTTFT])
	result.OpenAIAdvancedSchedulerWeightReset = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightReset])
	result.OpenAIAdvancedSchedulerWeightQuotaHeadroom = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightQuotaHeadroom])
	result.OpenAIAdvancedSchedulerWeightPreviousResponse = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightPreviousResponse])
	result.OpenAIAdvancedSchedulerWeightSessionSticky = strings.TrimSpace(settings[SettingKeyOpenAIAdvancedSchedulerWeightSessionSticky])
	result.OpenAIAdvancedSchedulerEffectiveLBTopK = s.openAIAdvancedSchedulerEffectiveLBTopK()
	effectiveWeights := s.openAIAdvancedSchedulerEffectiveWeights()
	result.OpenAIAdvancedSchedulerEffectiveWeightPriority = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.Priority)
	result.OpenAIAdvancedSchedulerEffectiveWeightLoad = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.Load)
	result.OpenAIAdvancedSchedulerEffectiveWeightQueue = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.Queue)
	result.OpenAIAdvancedSchedulerEffectiveWeightErrorRate = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.ErrorRate)
	result.OpenAIAdvancedSchedulerEffectiveWeightTTFT = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.TTFT)
	result.OpenAIAdvancedSchedulerEffectiveWeightReset = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.Reset)
	result.OpenAIAdvancedSchedulerEffectiveWeightQuotaHeadroom = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.QuotaHeadroom)
	result.OpenAIAdvancedSchedulerEffectiveWeightPreviousResponse = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.PreviousResponse)
	result.OpenAIAdvancedSchedulerEffectiveWeightSessionSticky = formatOpenAIAdvancedSchedulerFloat(effectiveWeights.SessionSticky)

	// 余额、订阅到期与账号限额通知
	result.BalanceLowNotifyEnabled = settings[SettingKeyBalanceLowNotifyEnabled] == "true"
	if v, err := strconv.ParseFloat(settings[SettingKeyBalanceLowNotifyThreshold], 64); err == nil && v >= 0 {
		result.BalanceLowNotifyThreshold = v
	}
	result.BalanceLowNotifyRechargeURL = settings[SettingKeyBalanceLowNotifyRechargeURL]
	result.SubscriptionExpiryNotifyEnabled = !isFalseSettingValue(settings[SettingKeySubscriptionExpiryNotifyEnabled])

	// 账号限额通知
	result.AccountQuotaNotifyEnabled = settings[SettingKeyAccountQuotaNotifyEnabled] == "true"
	if raw := strings.TrimSpace(settings[SettingKeyAccountQuotaNotifyEmails]); raw != "" {
		result.AccountQuotaNotifyEmails = ParseNotifyEmails(raw)
	}
	if result.AccountQuotaNotifyEmails == nil {
		result.AccountQuotaNotifyEmails = []NotifyEmailEntry{}
	}

	// Registration notification
	result.RegistrationNotifyEnabled = settings[SettingKeyRegistrationNotifyEnabled] == "true"
	result.RegistrationNotifyProvider = normalizeStoredRegistrationNotifyProvider(settings[SettingKeyRegistrationNotifyProvider])
	result.RegistrationNotifyWebhookURL = strings.TrimSpace(settings[SettingKeyRegistrationNotifyWebhookURL])
	result.RegistrationNotifySecret = strings.TrimSpace(settings[SettingKeyRegistrationNotifySecret])
	result.RegistrationNotifySecretConfigured = result.RegistrationNotifySecret != ""

	result.AllowUserViewErrorRequests = settings[SettingKeyAllowUserViewErrorRequests] == "true" // default false

	return result
}

// getStringOrDefault 获取字符串值或默认值
func (s *SettingService) getStringOrDefault(settings map[string]string, key, defaultValue string) string {
	if value, ok := settings[key]; ok && value != "" {
		return value
	}
	return defaultValue
}
