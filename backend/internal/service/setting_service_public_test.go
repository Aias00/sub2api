//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type settingPublicRepoStub struct {
	values map[string]string
}

func (s *settingPublicRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingPublicRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *settingPublicRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingPublicRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *settingPublicRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *settingPublicRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingPublicRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_GetPublicSettings_ExposesRegistrationEmailSuffixWhitelist(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyRegistrationEnabled:              "true",
			SettingKeyEmailVerifyEnabled:               "true",
			SettingKeyRegistrationEmailSuffixWhitelist: `["@EXAMPLE.com"," @foo.bar ","*.EDU.CN","@invalid_domain",""]`,
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"@example.com", "@foo.bar", "*.edu.cn"}, settings.RegistrationEmailSuffixWhitelist)
}

func TestSettingService_GetPublicSettings_ExposesTablePreferences(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyTableDefaultPageSize: "50",
			SettingKeyTablePageSizeOptions: "[20,50,100]",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, 50, settings.TableDefaultPageSize)
	require.Equal(t, []int{20, 50, 100}, settings.TablePageSizeOptions)
}

func TestSettingService_GetPublicSettings_ExposesForceEmailOnThirdPartySignup(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyForceEmailOnThirdPartySignup: "true",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.ForceEmailOnThirdPartySignup)
}

func TestSettingService_GetPublicSettings_ExposesPasswordMinLength(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyPasswordMinLength: "12",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, 12, settings.PasswordMinLength)
}

func TestSettingService_GetPublicSettings_ExposesWeChatOAuthModeCapabilities(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyWeChatConnectEnabled:             "true",
			SettingKeyWeChatConnectMPAppID:             "wx-mp-app",
			SettingKeyWeChatConnectMPAppSecret:         "wx-mp-secret",
			SettingKeyWeChatConnectMode:                "mp",
			SettingKeyWeChatConnectScopes:              "snsapi_base",
			SettingKeyWeChatConnectOpenEnabled:         "true",
			SettingKeyWeChatConnectMPEnabled:           "true",
			SettingKeyWeChatConnectRedirectURL:         "https://api.example.com/api/v1/auth/oauth/wechat/callback",
			SettingKeyWeChatConnectFrontendRedirectURL: "/auth/wechat/callback",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.WeChatOAuthEnabled)
	require.True(t, settings.WeChatOAuthOpenEnabled)
	require.True(t, settings.WeChatOAuthMPEnabled)
}

func TestSettingService_GetPublicSettings_DoesNotExposeMobileOnlyWeChatAsWebOAuthAvailable(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyWeChatConnectEnabled:             "true",
			SettingKeyWeChatConnectMobileEnabled:       "true",
			SettingKeyWeChatConnectMode:                "mobile",
			SettingKeyWeChatConnectMobileAppID:         "wx-mobile-app",
			SettingKeyWeChatConnectMobileAppSecret:     "wx-mobile-secret",
			SettingKeyWeChatConnectFrontendRedirectURL: "/auth/wechat/callback",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.False(t, settings.WeChatOAuthEnabled)
	require.False(t, settings.WeChatOAuthOpenEnabled)
	require.False(t, settings.WeChatOAuthMPEnabled)
	require.True(t, settings.WeChatOAuthMobileEnabled)
}

func TestSettingService_GetPublicSettings_FallsBackToConfigForWeChatOAuthCapabilities(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{
		WeChat: config.WeChatConnectConfig{
			Enabled:             true,
			OpenEnabled:         true,
			OpenAppID:           "wx-open-config",
			OpenAppSecret:       "wx-open-secret",
			FrontendRedirectURL: "/auth/wechat/config-callback",
		},
	})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.WeChatOAuthEnabled)
	require.True(t, settings.WeChatOAuthOpenEnabled)
	require.False(t, settings.WeChatOAuthMPEnabled)
	require.False(t, settings.WeChatOAuthMobileEnabled)
}

func TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.WorkspaceShellConfig)
	require.Equal(t, settings.WorkspaceShellConfig, settings.WebWorkspaceShellConfig)
	require.True(t, json.Valid([]byte(settings.WorkspaceShellConfig)))

	var payload map[string]map[string]string
	require.NoError(t, json.Unmarshal([]byte(settings.WorkspaceShellConfig), &payload))
	require.Equal(t, "AI 生图工作区", payload["zh"]["title"])
	require.Equal(t, "AI Image Workspace", payload["en"]["title"])
	require.Equal(t, "Copy prompt", payload["en"]["copyPromptLabel"])
}

func TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.PromptCatalogShellConfig)
	require.True(t, json.Valid([]byte(settings.PromptCatalogShellConfig)))

	var payload map[string]struct {
		Defaults struct {
			SourceType           string `json:"sourceType"`
			HasImage             bool   `json:"hasImage"`
			PageSize             int    `json:"pageSize"`
			SortBy               string `json:"sortBy"`
			SortOrder            string `json:"sortOrder"`
			GeneratorPath        string `json:"generatorPath"`
			GeneratorDraftSource string `json:"generatorDraftSource"`
		} `json:"defaults"`
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.PromptCatalogShellConfig), &payload))
	require.Equal(t, "case", payload["zh"].Defaults.SourceType)
	require.True(t, payload["zh"].Defaults.HasImage)
	require.Equal(t, 24, payload["zh"].Defaults.PageSize)
	require.Equal(t, "imported_at", payload["zh"].Defaults.SortBy)
	require.Equal(t, "desc", payload["zh"].Defaults.SortOrder)
	require.Equal(t, "/image-generator", payload["zh"].Defaults.GeneratorPath)
	require.Equal(t, "sub2api-vue-prompt-catalog", payload["zh"].Defaults.GeneratorDraftSource)
	require.Equal(t, "case", payload["en"].Defaults.SourceType)
	require.True(t, payload["en"].Defaults.HasImage)
	require.Equal(t, 24, payload["en"].Defaults.PageSize)
	require.Equal(t, "imported_at", payload["en"].Defaults.SortBy)
	require.Equal(t, "desc", payload["en"].Defaults.SortOrder)
	require.Equal(t, "/image-generator", payload["en"].Defaults.GeneratorPath)
	require.Equal(t, "sub2api-vue-prompt-catalog", payload["en"].Defaults.GeneratorDraftSource)
	require.Equal(t, "提示词案例库", payload["zh"].Labels["title"])
	require.Equal(t, "提示词案例库", payload["zh"].Labels["caseTitle"])
	require.Equal(t, "提示词模板库", payload["zh"].Labels["templateTitle"])
	require.Equal(t, "Prompt Catalog", payload["en"].Labels["title"])
	require.Equal(t, "Prompt Catalog", payload["en"].Labels["caseTitle"])
	require.Equal(t, "Prompt Templates", payload["en"].Labels["templateTitle"])
	require.Equal(t, "Dashboard", payload["en"].Labels["accountActionAuthenticated"])
	require.Equal(t, "Log in", payload["en"].Labels["accountActionAnonymous"])
	require.Equal(t, "X / Twitter", payload["zh"].Labels["importProviderX"])
	require.Equal(t, "X / Twitter", payload["en"].Labels["importProviderX"])
}

func TestSettingService_GetPublicSettings_DefaultsPricingShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.PricingShellConfig)
	require.Equal(t, settings.PricingShellConfig, settings.WebPricingShellConfig)
	require.True(t, json.Valid([]byte(settings.PricingShellConfig)))

	var payload map[string]struct {
		Button struct {
			Title string `json:"title"`
		} `json:"button"`
		Groups []struct {
			Name  string `json:"name"`
			Title string `json:"title"`
		} `json:"groups"`
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.PricingShellConfig), &payload))
	require.Equal(t, "价格与套餐", payload["zh"].Labels["title"])
	require.Equal(t, "Pricing", payload["en"].Labels["title"])
	require.Equal(t, "Buy", payload["en"].Button.Title)
	require.Len(t, payload["en"].Groups, 2)
	require.Equal(t, "one-time", payload["en"].Groups[0].Name)
	require.Equal(t, "Recharge", payload["en"].Groups[0].Title)
}

func TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.PaymentShellConfig)
	require.Equal(t, settings.PaymentShellConfig, settings.WebPaymentShellConfig)
	require.True(t, json.Valid([]byte(settings.PaymentShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.PaymentShellConfig), &payload))
	require.Equal(t, "充值", payload["zh"].Labels["tabTopUp"])
	require.Equal(t, "支付方式", payload["zh"].Labels["paymentMethod"])
	require.Equal(t, "到账 ${amount} 余额", payload["zh"].Labels["rechargeProductCreditLine"])
	require.Equal(t, "订单已过期", payload["zh"].Labels["expired"])
	require.Equal(t, "重新打开支付页面", payload["zh"].Labels["openPayWindow"])
	require.Equal(t, "充值金额", payload["zh"].Labels["baseAmount"])
	require.Equal(t, "当前充值商品没有可用支付方式", payload["zh"].Labels["amountNoMethod"])
	require.Equal(t, "取消订单过于频繁，请稍后再试", payload["zh"].Labels["cancelRateLimited"])
	require.Equal(t, "返回充值", payload["zh"].Labels["backToRecharge"])
	require.Equal(t, "支付组件加载失败，请刷新页面重试", payload["zh"].Labels["stripeLoadFailed"])
	require.Equal(t, "缺少订单ID或支付密钥", payload["zh"].Labels["stripeMissingParams"])
	require.Equal(t, "立即支付", payload["zh"].Labels["stripePay"])
	require.Equal(t, "Airwallex 支付组件加载失败", payload["zh"].Labels["airwallexLoadFailed"])
	require.Equal(t, "缺少 Airwallex 支付参数", payload["zh"].Labels["airwallexMissingParams"])
	require.Equal(t, "正在跳转到支付页面...", payload["zh"].Labels["stripePopupRedirecting"])
	require.Equal(t, "等待支付凭证超时，请重试", payload["zh"].Labels["stripePopupTimeout"])
	require.Equal(t, "支付页面已在新窗口打开", payload["zh"].Labels["payInNewWindow"])
	require.Equal(t, "正在恢复微信支付", payload["zh"].Labels["wechatPaymentCallbackTitle"])
	require.Equal(t, "微信支付回调缺少恢复令牌。", payload["zh"].Labels["wechatPaymentCallbackMissingResumeToken"])
	require.Equal(t, "申请退款", payload["zh"].Labels["requestRefund"])
	require.Equal(t, "订单已取消", payload["zh"].Labels["cancelSuccess"])
	require.Equal(t, "操作失败", payload["zh"].Labels["errorFallback"])
	require.Equal(t, "创建时间", payload["zh"].Labels["createdAt"])
	require.Equal(t, "已支付", payload["zh"].Labels["statusPaid"])
	require.Equal(t, "部分退款", payload["zh"].Labels["statusPartiallyRefunded"])
	require.Equal(t, "暂无有效订阅", payload["zh"].Labels["subscriptionNoActive"])
	require.Equal(t, "额度将在 {time} 后重置", payload["zh"].Labels["subscriptionQuotaEndsIn"])
	require.Equal(t, "Subscribe", payload["en"].Labels["tabSubscribe"])
	require.Equal(t, "Create Order", payload["en"].Labels["createOrder"])
	require.Equal(t, "Subscribe Now", payload["en"].Labels["subscribeNow"])
	require.Equal(t, "Models", payload["en"].Labels["models"])
	require.Equal(t, "Payment Successful", payload["en"].Labels["success"])
	require.Equal(t, "Waiting for payment...", payload["en"].Labels["waitingPayment"])
	require.Equal(t, "Payment Failed", payload["en"].Labels["failed"])
	require.Equal(t, "Resuming WeChat payment", payload["en"].Labels["wechatPaymentCallbackTitle"])
	require.Equal(t, "WeChat payment callback is missing a resume token.", payload["en"].Labels["wechatPaymentCallbackMissingResumeToken"])
	require.Equal(t, "No payment method is available for this recharge product", payload["en"].Labels["amountNoMethod"])
	require.Equal(t, "Too many pending orders. Complete or cancel one first (max {max}).", payload["en"].Labels["tooManyPending"])
	require.Equal(t, "This environment cannot open the payment sheet directly, so QR payment is shown instead.", payload["en"].Labels["mobilePaymentFallbackToQr"])
	require.Equal(t, "View Orders", payload["en"].Labels["viewOrders"])
	require.Equal(t, "Pay Now", payload["en"].Labels["stripePay"])
	require.Equal(t, "Failed to load Airwallex checkout", payload["en"].Labels["airwallexLoadFailed"])
	require.Equal(t, "Missing Airwallex payment parameters", payload["en"].Labels["airwallexMissingParams"])
	require.Equal(t, "Redirecting to payment page...", payload["en"].Labels["stripePopupRedirecting"])
	require.Equal(t, "Payment page opened in a new window", payload["en"].Labels["payInNewWindow"])
	require.Equal(t, "Request Refund", payload["en"].Labels["requestRefund"])
	require.Equal(t, "Order cancelled", payload["en"].Labels["cancelSuccess"])
	require.Equal(t, "Refund request submitted", payload["en"].Labels["refundSuccess"])
	require.Equal(t, "Paid", payload["en"].Labels["statusPaid"])
	require.Equal(t, "Partially refunded", payload["en"].Labels["statusPartiallyRefunded"])
	require.Equal(t, "No Active Subscriptions", payload["en"].Labels["subscriptionNoActive"])
	require.Equal(t, "Quota resets in {time}", payload["en"].Labels["subscriptionQuotaEndsIn"])
	require.Equal(t, "Recharge rate: 1 CNY = {usd} USD balance", payload["en"].Labels["rechargeRatePreview"])
}

func TestSettingService_GetPublicSettings_DefaultsCreditsShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.CreditsShellConfig)
	require.True(t, json.Valid([]byte(settings.CreditsShellConfig)))

	var payload map[string]struct {
		Labels     map[string]string `json:"labels"`
		Conversion string            `json:"conversion"`
		Actions    struct {
			Title string `json:"title"`
		} `json:"actions"`
		Buttons struct {
			Recharge string `json:"recharge"`
			Orders   string `json:"orders"`
		} `json:"buttons"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.CreditsShellConfig), &payload))
	require.Equal(t, "积分", payload["zh"].Labels["eyebrow"])
	require.Equal(t, "积分余额", payload["zh"].Labels["title"])
	require.Equal(t, "Credits", payload["en"].Labels["eyebrow"])
	require.Equal(t, "Credit Balance", payload["en"].Labels["title"])
	require.Equal(t, "Conversion: {creditsPerBalance} credits = 1 Sub2API balance.", payload["en"].Conversion)
	require.Equal(t, "Balance actions", payload["en"].Actions.Title)
	require.Equal(t, "Recharge", payload["en"].Buttons.Recharge)
	require.Equal(t, "View orders", payload["en"].Buttons.Orders)
}

func TestSettingService_GetPublicSettings_DefaultsHomeShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.HomeShellConfig)
	require.True(t, json.Valid([]byte(settings.HomeShellConfig)))

	var payload map[string]struct {
		Labels          map[string]string `json:"labels"`
		ExperienceCards []struct {
			Key   string `json:"key"`
			Title string `json:"title"`
		} `json:"experienceCards"`
		WhyChooseCards []struct {
			Key   string `json:"key"`
			Title string `json:"title"`
		} `json:"whyChooseCards"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.HomeShellConfig), &payload))
	require.Equal(t, "AI 编码工作台", payload["zh"].Labels["heroTitle"])
	require.Equal(t, "文档", payload["zh"].Labels["navDocs"])
	require.Equal(t, "AI Coding Workspace", payload["en"].Labels["heroTitle"])
	require.Equal(t, "Docs", payload["en"].Labels["navDocs"])
	require.Equal(t, "Agents", payload["en"].Labels["familyGptAgents"])
	require.Equal(t, "Support", payload["en"].Labels["footerSupport"])
	require.Len(t, payload["en"].ExperienceCards, 4)
	require.Equal(t, "unified", payload["en"].ExperienceCards[0].Key)
	require.Equal(t, "One key, unified access", payload["en"].ExperienceCards[0].Title)
	require.Len(t, payload["zh"].WhyChooseCards, 4)
	require.Equal(t, "lowFriction", payload["zh"].WhyChooseCards[0].Key)
	require.Equal(t, "少折腾配置", payload["zh"].WhyChooseCards[0].Title)
}

func TestSettingService_GetPublicSettings_DefaultsHomeBusinessShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.HomeBusinessShellConfig)
	require.True(t, json.Valid([]byte(settings.HomeBusinessShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
		BusinessCards []struct {
			Key   string `json:"key"`
			Title string `json:"title"`
		} `json:"businessCards"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.HomeBusinessShellConfig), &payload))
	require.Equal(t, "面向业务场景的 AI 能力工作台", payload["zh"].Labels["heroTitle"])
	require.Equal(t, "An AI workspace organized around business capabilities", payload["en"].Labels["heroTitle"])
	require.Equal(t, "提示词", payload["zh"].Labels["navModels"])
	require.Equal(t, "Prompts", payload["en"].Labels["navModels"])
	require.Len(t, payload["zh"].BusinessCards, 4)
	require.Equal(t, "wechat-export", payload["zh"].BusinessCards[0].Key)
	require.Equal(t, "微信导出", payload["zh"].BusinessCards[0].Title)
}

func TestSettingService_GetPublicSettings_DefaultsModelPlazaShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.ModelPlazaShellConfig)
	require.True(t, json.Valid([]byte(settings.ModelPlazaShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.ModelPlazaShellConfig), &payload))
	require.Equal(t, "公开模型目录", payload["zh"].Labels["title"])
	require.Equal(t, "搜索模型、能力或标签", payload["zh"].Labels["searchPlaceholder"])
	require.Equal(t, "Public Model Catalog", payload["en"].Labels["title"])
	require.Equal(t, "Search models, capabilities, or tags", payload["en"].Labels["searchPlaceholder"])
	require.Equal(t, "All models", payload["en"].Labels["groupAll"])
}

func TestSettingService_GetPublicSettings_DefaultsDocsShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.DocsShellConfig)
	require.True(t, json.Valid([]byte(settings.DocsShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.DocsShellConfig), &payload))
	require.Equal(t, "文档", payload["zh"].Labels["title"])
	require.Equal(t, "搜索文档", payload["zh"].Labels["searchPlaceholder"])
	require.Equal(t, "Docs", payload["en"].Labels["title"])
	require.Equal(t, "Search docs", payload["en"].Labels["searchPlaceholder"])
	require.Equal(t, "No results", payload["en"].Labels["noData"])
}

func TestSettingService_GetPublicSettings_DefaultsDocsContentBasePath(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.JSONEq(t, `{"zh":"/docs-content/","en":"/docs-content/en/"}`, settings.DocsContentBasePath)
}

func TestSettingService_GetPublicSettings_DefaultsLegalDocumentShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.LegalDocumentShellConfig)
	require.True(t, json.Valid([]byte(settings.LegalDocumentShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.LegalDocumentShellConfig), &payload))
	require.Equal(t, "登录条款", payload["zh"].Labels["agreementLabel"])
	require.Equal(t, "文档不存在", payload["zh"].Labels["missingTitle"])
	require.Equal(t, "Login agreement", payload["en"].Labels["agreementLabel"])
	require.Equal(t, "Document not found", payload["en"].Labels["missingTitle"])
	require.Equal(t, "Updated: {date}", payload["en"].Labels["updatedAt"])
}

func TestSettingService_GetPublicSettings_DefaultsKeyUsageShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.KeyUsageShellConfig)
	require.True(t, json.Valid([]byte(settings.KeyUsageShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.KeyUsageShellConfig), &payload))
	require.Equal(t, "API Key 用量查询", payload["zh"].Labels["title"])
	require.Equal(t, "请输入 API Key", payload["zh"].Labels["enterApiKey"])
	require.Equal(t, "文档", payload["zh"].Labels["docs"])
	require.Equal(t, "保留所有权利。", payload["zh"].Labels["allRightsReserved"])
	require.Equal(t, "API Key Usage", payload["en"].Labels["title"])
	require.Equal(t, "Query", payload["en"].Labels["query"])
	require.Equal(t, "Docs", payload["en"].Labels["docs"])
	require.Equal(t, "All rights reserved.", payload["en"].Labels["allRightsReserved"])
	require.Equal(t, "({days} days)", payload["en"].Labels["daysLeft"])
}

func TestSettingService_GetPublicSettings_DefaultsAPIKeysShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.APIKeysShellConfig)
	require.True(t, json.Valid([]byte(settings.APIKeysShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.APIKeysShellConfig), &payload))
	require.Equal(t, "搜索 API Key", payload["zh"].Labels["searchPlaceholder"])
	require.Equal(t, "创建 Key", payload["zh"].Labels["createKey"])
	require.Equal(t, "删除 API Key", payload["zh"].Labels["deleteKey"])
	require.Equal(t, "重置频率限制用量", payload["zh"].Labels["resetRateLimitUsage"])
	require.Equal(t, "API 端点", payload["zh"].Labels["endpointTitle"])
	require.Equal(t, "点击可复制此端点", payload["zh"].Labels["endpointClickToCopy"])
	require.Equal(t, "Search API keys", payload["en"].Labels["searchPlaceholder"])
	require.Equal(t, "Create Key", payload["en"].Labels["createKey"])
	require.Equal(t, "Delete API Key", payload["en"].Labels["deleteKey"])
	require.Equal(t, "API Endpoints", payload["en"].Labels["endpointTitle"])
	require.Equal(t, "Click to copy this endpoint", payload["en"].Labels["endpointClickToCopy"])
	require.Equal(t, "Select CCS Client", payload["en"].Labels["ccsClientSelectTitle"])
	require.Equal(t, "Use API Key", payload["en"].Labels["useKeyModalTitle"])
	require.Equal(t, "Please assign a group first", payload["en"].Labels["useKeyModalNoGroupTitle"])
	require.Equal(t, "Copy", payload["en"].Labels["useKeyModalCopy"])
	require.Equal(t, "Failed to load API keys", payload["en"].Labels["failedToLoad"])
}

func TestSettingService_GetPublicSettings_DefaultsDashboardShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.DashboardShellConfig)
	require.True(t, json.Valid([]byte(settings.DashboardShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.DashboardShellConfig), &payload))
	require.Equal(t, "余额", payload["zh"].Labels["balance"])
	require.Equal(t, "最近使用", payload["zh"].Labels["recentUsage"])
	require.Equal(t, "Balance", payload["en"].Labels["balance"])
	require.Equal(t, "Recent usage", payload["en"].Labels["recentUsage"])
	require.Equal(t, "{count} platforms", payload["en"].Labels["platformCount"])
}

func TestSettingService_GetPublicSettings_DefaultsUsageShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.UsageShellConfig)
	require.True(t, json.Valid([]byte(settings.UsageShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.UsageShellConfig), &payload))
	require.Equal(t, "总请求数", payload["zh"].Labels["totalRequests"])
	require.Equal(t, "导出 CSV", payload["zh"].Labels["exportCsv"])
	require.Equal(t, "Total Requests", payload["en"].Labels["totalRequests"])
	require.Equal(t, "Export CSV", payload["en"].Labels["exportCsv"])
	require.Equal(t, "Billing Mode", payload["en"].Labels["billingMode"])
	require.Equal(t, "No usage records", payload["en"].Labels["noRecords"])
	require.Equal(t, "Failed to load usage records", payload["en"].Labels["failedToLoad"])
	require.Equal(t, "Export successful", payload["en"].Labels["exportSuccess"])
	require.NotContains(t, payload["en"].Labels, "tokenDetails")
	require.NotContains(t, payload["en"].Labels, "inputTokens")
	require.NotContains(t, payload["en"].Labels, "cacheReadTokens")
	require.NotContains(t, payload["en"].Labels, "costDetails")
	require.NotContains(t, payload["en"].Labels, "inputCost")
	require.NotContains(t, payload["en"].Labels, "cacheReadCost")
	require.NotContains(t, payload["en"].Labels, "ws")
	require.NotContains(t, payload["en"].Labels, "stream")
	require.NotContains(t, payload["en"].Labels, "sync")
	require.NotContains(t, payload["en"].Labels, "unknown")
}

func TestSettingService_GetPublicSettings_DefaultsAPIGuideShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.APIGuideShellConfig)
	require.True(t, json.Valid([]byte(settings.APIGuideShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.APIGuideShellConfig), &payload))
	require.Equal(t, "网关调用说明", payload["zh"].Labels["title"])
	require.Equal(t, "状态", payload["zh"].Labels["status"])
	require.Equal(t, "开启流式输出", payload["zh"].Labels["stream"])
	require.Equal(t, "复制 curl", payload["zh"].Labels["copyCurl"])
	require.Equal(t, "Gateway API Guide", payload["en"].Labels["title"])
	require.Equal(t, "Status", payload["en"].Labels["status"])
	require.Equal(t, "Streaming", payload["en"].Labels["stream"])
	require.Equal(t, "Copy curl", payload["en"].Labels["copyCurl"])
	require.Equal(t, "Failed to load API keys", payload["en"].Labels["loadKeysFailed"])
}

func TestSettingService_GetPublicSettings_DefaultsAPITestShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.APITestShellConfig)
	require.True(t, json.Valid([]byte(settings.APITestShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.APITestShellConfig), &payload))
	require.Equal(t, "调用测试", payload["zh"].Labels["title"])
	require.Equal(t, "发送测试请求", payload["zh"].Labels["send"])
	require.Equal(t, "加载中...", payload["zh"].Labels["loading"])
	require.Equal(t, "没有可选项", payload["zh"].Labels["noOptionsFound"])
	require.Equal(t, "未知错误", payload["zh"].Labels["unknownError"])
	require.Equal(t, "API Test", payload["en"].Labels["title"])
	require.Equal(t, "Send Test Request", payload["en"].Labels["send"])
	require.Equal(t, "Loading...", payload["en"].Labels["loading"])
	require.Equal(t, "No options found", payload["en"].Labels["noOptionsFound"])
	require.Equal(t, "Unknown error", payload["en"].Labels["unknownError"])
	require.Equal(t, "Failed to load API keys", payload["en"].Labels["loadKeysFailed"])
}

func TestSettingService_GetPublicSettings_DefaultsAvailableGroupsShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.AvailableGroupsShellConfig)
	require.True(t, json.Valid([]byte(settings.AvailableGroupsShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.AvailableGroupsShellConfig), &payload))
	require.Equal(t, "可用分组", payload["zh"].Labels["title"])
	require.Equal(t, "公开分组", payload["zh"].Labels["publicTitle"])
	require.Equal(t, "Available Groups", payload["en"].Labels["title"])
	require.Equal(t, "Public Groups", payload["en"].Labels["publicTitle"])
	require.Equal(t, "Daily ${amount}", payload["en"].Labels["dailyLimit"])
}

func TestSettingService_GetPublicSettings_DefaultsRedeemShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.RedeemShellConfig)
	require.True(t, json.Valid([]byte(settings.RedeemShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.RedeemShellConfig), &payload))
	require.Equal(t, "当前余额", payload["zh"].Labels["currentBalance"])
	require.Equal(t, "兑换", payload["zh"].Labels["redeemButton"])
	require.Equal(t, "最近活动", payload["zh"].Labels["recentActivity"])
	require.Equal(t, "余额充值（返利转入）", payload["zh"].Labels["balanceAddedAffiliate"])
	require.Equal(t, "Current Balance", payload["en"].Labels["currentBalance"])
	require.Equal(t, "Redeem Code", payload["en"].Labels["redeemButton"])
	require.Equal(t, "Recent Activity", payload["en"].Labels["recentActivity"])
	require.Equal(t, "Balance Added (Affiliate Transfer)", payload["en"].Labels["balanceAddedAffiliate"])
}

func TestSettingService_GetPublicSettings_DefaultsAffiliateShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.AffiliateShellConfig)
	require.True(t, json.Valid([]byte(settings.AffiliateShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.AffiliateShellConfig), &payload))
	require.Equal(t, "邀请中心", payload["zh"].Labels["title"])
	require.Equal(t, "转入余额", payload["zh"].Labels["transferButton"])
	require.Equal(t, "Invite Center", payload["en"].Labels["title"])
	require.Equal(t, "Transfer to Balance", payload["en"].Labels["transferButton"])
	require.Equal(t, "{amount} has been transferred to your balance", payload["en"].Labels["transferSuccess"])
}

func TestSettingService_GetPublicSettings_DefaultsAvailableChannelsShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.AvailableChannelsShellConfig)
	require.True(t, json.Valid([]byte(settings.AvailableChannelsShellConfig)))

	var payload map[string]struct {
		Labels map[string]any `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.AvailableChannelsShellConfig), &payload))
	require.Equal(t, "搜索渠道或模型...", payload["zh"].Labels["searchPlaceholder"])
	require.Equal(t, "暂无可用渠道", payload["zh"].Labels["empty"])
	require.Equal(t, "加载可用渠道失败", payload["zh"].Labels["loadError"])
	require.Equal(t, "对所有用户公开的分组", payload["zh"].Labels["publicTooltip"])
	require.Equal(t, "Search channels or models...", payload["en"].Labels["searchPlaceholder"])
	require.Equal(t, "No available channels", payload["en"].Labels["empty"])
	require.Equal(t, "Failed to load available channels", payload["en"].Labels["loadError"])
	require.Equal(t, "Groups open to all users", payload["en"].Labels["publicTooltip"])
	zhColumns, ok := payload["zh"].Labels["columns"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "渠道名", zhColumns["name"])
	require.Equal(t, "支持模型", zhColumns["supportedModels"])
	zhPricing, ok := payload["zh"].Labels["pricing"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "计费模式", zhPricing["billingMode"])
	require.Equal(t, "/ 1M token", zhPricing["unitPerMillion"])
	enPricing, ok := payload["en"].Labels["pricing"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Billing Mode", enPricing["billingMode"])
	require.Equal(t, "/ 1M tokens", enPricing["unitPerMillion"])
}

func TestSettingService_GetPublicSettings_DefaultsChannelStatusShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.ChannelStatusShellConfig)
	require.True(t, json.Valid([]byte(settings.ChannelStatusShellConfig)))

	var payload map[string]struct {
		Labels map[string]any `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.ChannelStatusShellConfig), &payload))
	require.Equal(t, "渠道详情", payload["zh"].Labels["detailTitle"])
	require.Equal(t, "Channel Detail", payload["en"].Labels["detailTitle"])
	zhWindowTabs, ok := payload["zh"].Labels["windowTab"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "7 天", zhWindowTabs["7d"])
	enDetailColumns, ok := payload["en"].Labels["detailColumns"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Model", enDetailColumns["model"])
	require.Equal(t, "Latest Status", enDetailColumns["latestStatus"])
}

func TestSettingService_GetPublicSettings_DefaultsCustomPageShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.CustomPageShellConfig)
	require.True(t, json.Valid([]byte(settings.CustomPageShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.CustomPageShellConfig), &payload))
	require.Equal(t, "目录", payload["zh"].Labels["tocTitle"])
	require.Equal(t, "新窗口打开", payload["zh"].Labels["openInNewTab"])
	require.Equal(t, "Table of Contents", payload["en"].Labels["tocTitle"])
	require.Equal(t, "Open in new tab", payload["en"].Labels["openInNewTab"])
	require.Equal(t, "Copied ✓", payload["en"].Labels["copyCodeSuccess"])
}

func TestSettingService_GetPublicSettings_DefaultsProfileShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.ProfileShellConfig)
	require.True(t, json.Valid([]byte(settings.ProfileShellConfig)))

	var payload map[string]struct {
		Labels map[string]any `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.ProfileShellConfig), &payload))
	require.GreaterOrEqual(t, len(payload["zh"].Labels), 124)
	require.GreaterOrEqual(t, len(payload["en"].Labels), 124)
	require.Equal(t, "账户余额", payload["zh"].Labels["accountBalance"])
	require.Equal(t, "联系客服", payload["zh"].Labels["contactSupport"])
	require.Equal(t, "密码至少需要 {count} 位字符", payload["zh"].Labels["passwordHint"])
	require.Equal(t, "密码修改成功", payload["zh"].Labels["passwordChangeSuccess"])
	require.Equal(t, "启用", payload["zh"].Labels["profileStatusActive"])
	require.Equal(t, "禁用", payload["zh"].Labels["profileStatusDisabled"])
	require.Equal(t, "编辑个人资料", payload["zh"].Labels["profileEditTitle"])
	require.Equal(t, "用户名", payload["zh"].Labels["profileUsername"])
	require.Equal(t, "更新中...", payload["zh"].Labels["profileUpdating"])
	require.Equal(t, "更新资料", payload["zh"].Labels["profileUpdateAction"])
	require.Equal(t, "用户名不能为空", payload["zh"].Labels["profileUsernameRequired"])
	require.Equal(t, "资料更新成功", payload["zh"].Labels["profileUpdateSuccess"])
	require.Equal(t, "资料更新失败", payload["zh"].Labels["profileUpdateFailed"])
	require.Equal(t, "余额不足提醒", payload["zh"].Labels["balanceNotifyTitle"])
	require.Equal(t, "通知邮箱", payload["zh"].Labels["balanceNotifyExtraEmails"])
	require.Equal(t, "保存", payload["zh"].Labels["balanceNotifySave"])
	require.Equal(t, "操作失败", payload["zh"].Labels["balanceNotifyError"])
	require.Equal(t, "资料头像", payload["zh"].Labels["avatarTitle"])
	require.Equal(t, "保存", payload["zh"].Labels["avatarSave"])
	require.Equal(t, "删除", payload["zh"].Labels["avatarDelete"])
	require.Equal(t, "头像已删除", payload["zh"].Labels["avatarDeleteSuccess"])
	require.Equal(t, "操作失败", payload["zh"].Labels["avatarError"])
	require.Equal(t, "两步验证", payload["zh"].Labels["totpTitle"])
	require.Equal(t, "启用", payload["zh"].Labels["totpEnable"])
	require.Equal(t, "设置双因素认证", payload["zh"].Labels["totpSetupTitle"])
	require.Equal(t, "禁用双因素认证", payload["zh"].Labels["totpDisableTitle"])
	require.Equal(t, "发送中...", payload["zh"].Labels["totpSending"])
	require.Equal(t, "取消", payload["zh"].Labels["totpCancel"])
	require.Equal(t, "下一步", payload["zh"].Labels["totpNext"])
	require.Equal(t, "返回", payload["zh"].Labels["totpBack"])
	require.Equal(t, "加载中...", payload["zh"].Labels["totpLoading"])
	require.Equal(t, "验证中...", payload["zh"].Labels["totpVerifying"])
	require.Equal(t, "已复制", payload["zh"].Labels["totpCopied"])
	require.Equal(t, "复制失败", payload["zh"].Labels["totpCopyFailed"])
	require.Equal(t, "处理中...", payload["zh"].Labels["totpProcessing"])
	require.Equal(t, "操作失败", payload["zh"].Labels["totpError"])
	require.Equal(t, "登录方式绑定", payload["zh"].Labels["authBindingsTitle"])
	require.Equal(t, "绑定 {providerName}", payload["zh"].Labels["authBindingsBindAction"])
	require.Equal(t, "加载中...", payload["zh"].Labels["authBindingsLoading"])
	require.Equal(t, "请稍后重试", payload["zh"].Labels["authBindingsTryAgain"])
	require.Equal(t, "请输入邮箱", payload["zh"].Labels["authBindingsEmailRequired"])
	require.Equal(t, "请输入有效的邮箱地址", payload["zh"].Labels["authBindingsInvalidEmail"])
	require.Equal(t, "请输入验证码", payload["zh"].Labels["authBindingsCodeRequired"])
	require.Equal(t, "请输入密码", payload["zh"].Labels["authBindingsPasswordRequired"])
	require.Equal(t, "密码至少需要 {count} 位字符", payload["zh"].Labels["authBindingsPasswordMinLength"])
	require.Equal(t, "发送验证码失败", payload["zh"].Labels["authBindingsSendCodeFailed"])
	zhProviders, ok := payload["zh"].Labels["providers"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "邮箱", zhProviders["email"])
	require.Equal(t, "微信", zhProviders["wechat"])
	require.Equal(t, "GitHub", zhProviders["github"])
	require.Equal(t, "Google", zhProviders["google"])
	require.Equal(t, "Account Balance", payload["en"].Labels["accountBalance"])
	require.Equal(t, "Contact Support", payload["en"].Labels["contactSupport"])
	require.Equal(t, "Password must be at least {count} characters long", payload["en"].Labels["passwordHint"])
	require.Equal(t, "Password changed successfully", payload["en"].Labels["passwordChangeSuccess"])
	require.Equal(t, "Active", payload["en"].Labels["profileStatusActive"])
	require.Equal(t, "Disabled", payload["en"].Labels["profileStatusDisabled"])
	require.Equal(t, "Edit Profile", payload["en"].Labels["profileEditTitle"])
	require.Equal(t, "Username", payload["en"].Labels["profileUsername"])
	require.Equal(t, "Updating...", payload["en"].Labels["profileUpdating"])
	require.Equal(t, "Update Profile", payload["en"].Labels["profileUpdateAction"])
	require.Equal(t, "Username is required", payload["en"].Labels["profileUsernameRequired"])
	require.Equal(t, "Profile updated successfully", payload["en"].Labels["profileUpdateSuccess"])
	require.Equal(t, "Failed to update profile", payload["en"].Labels["profileUpdateFailed"])
	require.Equal(t, "Balance Low Notification", payload["en"].Labels["balanceNotifyTitle"])
	require.Equal(t, "Notification Emails", payload["en"].Labels["balanceNotifyExtraEmails"])
	require.Equal(t, "Save", payload["en"].Labels["balanceNotifySave"])
	require.Equal(t, "Operation failed", payload["en"].Labels["balanceNotifyError"])
	require.Equal(t, "Profile Avatar", payload["en"].Labels["avatarTitle"])
	require.Equal(t, "Save", payload["en"].Labels["avatarSave"])
	require.Equal(t, "Delete", payload["en"].Labels["avatarDelete"])
	require.Equal(t, "Avatar removed", payload["en"].Labels["avatarDeleteSuccess"])
	require.Equal(t, "Operation failed", payload["en"].Labels["avatarError"])
	require.Equal(t, "Two-Factor Authentication", payload["en"].Labels["totpTitle"])
	require.Equal(t, "Enable", payload["en"].Labels["totpEnable"])
	require.Equal(t, "Set Up Two-Factor Authentication", payload["en"].Labels["totpSetupTitle"])
	require.Equal(t, "Disable Two-Factor Authentication", payload["en"].Labels["totpDisableTitle"])
	require.Equal(t, "Sending...", payload["en"].Labels["totpSending"])
	require.Equal(t, "Cancel", payload["en"].Labels["totpCancel"])
	require.Equal(t, "Next", payload["en"].Labels["totpNext"])
	require.Equal(t, "Back", payload["en"].Labels["totpBack"])
	require.Equal(t, "Loading...", payload["en"].Labels["totpLoading"])
	require.Equal(t, "Verifying...", payload["en"].Labels["totpVerifying"])
	require.Equal(t, "Copied", payload["en"].Labels["totpCopied"])
	require.Equal(t, "Copy failed", payload["en"].Labels["totpCopyFailed"])
	require.Equal(t, "Processing...", payload["en"].Labels["totpProcessing"])
	require.Equal(t, "Operation failed", payload["en"].Labels["totpError"])
	require.Equal(t, "Connected Sign-In Methods", payload["en"].Labels["authBindingsTitle"])
	require.Equal(t, "Bind {providerName}", payload["en"].Labels["authBindingsBindAction"])
	require.Equal(t, "Loading...", payload["en"].Labels["authBindingsLoading"])
	require.Equal(t, "Please try again later", payload["en"].Labels["authBindingsTryAgain"])
	require.Equal(t, "Email is required", payload["en"].Labels["authBindingsEmailRequired"])
	require.Equal(t, "Enter a valid email address", payload["en"].Labels["authBindingsInvalidEmail"])
	require.Equal(t, "Verification code is required", payload["en"].Labels["authBindingsCodeRequired"])
	require.Equal(t, "Password is required", payload["en"].Labels["authBindingsPasswordRequired"])
	require.Equal(t, "Password must be at least {count} characters long", payload["en"].Labels["authBindingsPasswordMinLength"])
	require.Equal(t, "Failed to send verification code", payload["en"].Labels["authBindingsSendCodeFailed"])
	providers, ok := payload["en"].Labels["providers"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Email", providers["email"])
	require.Equal(t, "WeChat", providers["wechat"])
	require.Equal(t, "GitHub", providers["github"])
	require.Equal(t, "Google", providers["google"])
}

func TestSettingService_GetPublicSettings_DefaultsAuthShellConfig(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, settings.AuthShellConfig)
	require.True(t, json.Valid([]byte(settings.AuthShellConfig)))

	var payload map[string]struct {
		Labels map[string]string `json:"labels"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings.AuthShellConfig), &payload))
	require.Equal(t, "欢迎回来", payload["zh"].Labels["welcomeBack"])
	require.Equal(t, "注册以开始使用 {siteName}", payload["zh"].Labels["signUpToStart"])
	require.Equal(t, "如果页面未自动跳转，请返回登录页重试。", payload["zh"].Labels["oauthCallbackHint"])
	require.Equal(t, "OAuth 回调", payload["zh"].Labels["oauthCallbackTitle"])
	require.Equal(t, "无效的登录回调", payload["zh"].Labels["oauthCallbackInvalidTitle"])
	require.Equal(t, "完成注册", payload["zh"].Labels["oauthCallbackSubmitRegistration"])
	require.Equal(t, "使用 {providerName} 资料", payload["zh"].Labels["oauthFlowProfileDetailsTitle"])
	require.Equal(t, "绑定当前账户", payload["zh"].Labels["oauthFlowBindCurrentAccount"])
	require.Equal(t, "绑定已有账户", payload["zh"].Labels["oauthFlowBindExistingAccount"])
	require.Equal(t, "完成 {providerName} 账户注册", payload["zh"].Labels["oauthFlowCreateAccountTitle"])
	require.Equal(t, "请输入 {account} 的 6 位验证码，以完成此次 {providerName} 登录绑定。", payload["zh"].Labels["oauthFlowTotpHint"])
	require.Equal(t, "钉钉", payload["zh"].Labels["dingtalkProviderName"])
	require.Equal(t, "微信", payload["zh"].Labels["wechatProviderName"])
	require.Equal(t, "暂时无法确认微信登录可用性，请刷新后重试。", payload["zh"].Labels["wechatAvailabilityUnknown"])
	require.Equal(t, "当前微信登录流程仅支持在系统浏览器中继续。", payload["zh"].Labels["wechatSystemBrowserOnly"])
	require.Equal(t, "重置密码", payload["zh"].Labels["forgotPasswordTitle"])
	require.Equal(t, "验证您的邮箱", payload["zh"].Labels["emailVerifyTitle"])
	require.Equal(t, "{countdown}秒后可重新发送", payload["zh"].Labels["emailVerifyResendCountdown"])
	require.Equal(t, "发送重置链接", payload["zh"].Labels["sendResetLink"])
	require.Equal(t, "重置链接已发送", payload["zh"].Labels["resetEmailSent"])
	require.Equal(t, "设置新密码", payload["zh"].Labels["resetPasswordTitle"])
	require.Equal(t, "重置密码", payload["zh"].Labels["resetPassword"])
	require.Equal(t, "密码重置成功", payload["zh"].Labels["passwordResetSuccess"])
	require.Equal(t, "正在完成 {providerName} 登录", payload["zh"].Labels["providerCallbackTitle"])
	require.Equal(t, "正在验证 {providerName} 登录信息，请稍候...", payload["zh"].Labels["providerCallbackProcessing"])
	require.Equal(t, "该 {providerName} 账号尚未注册，站点已开启邀请码注册，请输入邀请码以完成注册。", payload["zh"].Labels["providerInvitationRequired"])
	require.Equal(t, "输入新密码", payload["zh"].Labels["newPasswordPlaceholder"])
	require.Equal(t, "可选", payload["zh"].Labels["optional"])
	require.Equal(t, "发送验证码", payload["zh"].Labels["sendCode"])
	require.Equal(t, "验证码已发送", payload["zh"].Labels["codeSentSuccess"])
	require.Equal(t, "请输入邮箱收到的 6 位验证码", payload["zh"].Labels["verificationCodeHint"])
	require.Equal(t, "同意并继续", payload["zh"].Labels["agreementAcceptAndContinue"])
	require.Equal(t, "我们的服务条款已于 {date} 更新。在继续使用服务之前，请仔细阅读并同意以下条款。", payload["zh"].Labels["agreementUpdatedAt"])
	require.Equal(t, "保留所有权利。", payload["zh"].Labels["allRightsReserved"])
	require.Equal(t, "Welcome Back", payload["en"].Labels["welcomeBack"])
	require.Equal(t, "Create Account", payload["en"].Labels["createAccount"])
	require.Equal(t, "Sign in with {providerName}", payload["en"].Labels["signInWithProvider"])
	require.Equal(t, "If you are not redirected automatically, go back to the login page and try again.", payload["en"].Labels["oauthCallbackHint"])
	require.Equal(t, "OAuth Callback", payload["en"].Labels["oauthCallbackTitle"])
	require.Equal(t, "Invalid sign-in callback", payload["en"].Labels["oauthCallbackInvalidTitle"])
	require.Equal(t, "Complete Registration", payload["en"].Labels["oauthCallbackSubmitRegistration"])
	require.Equal(t, "Use {providerName} profile details", payload["en"].Labels["oauthFlowProfileDetailsTitle"])
	require.Equal(t, "Bind current account", payload["en"].Labels["oauthFlowBindCurrentAccount"])
	require.Equal(t, "Bind existing account", payload["en"].Labels["oauthFlowBindExistingAccount"])
	require.Equal(t, "Complete your {providerName} account setup", payload["en"].Labels["oauthFlowCreateAccountTitle"])
	require.Equal(t, "Enter the 6-digit verification code for {account} to finish binding this {providerName} sign-in.", payload["en"].Labels["oauthFlowTotpHint"])
	require.Equal(t, "DingTalk", payload["en"].Labels["dingtalkProviderName"])
	require.Equal(t, "WeChat", payload["en"].Labels["wechatProviderName"])
	require.Equal(t, "WeChat sign-in availability could not be confirmed. Refresh and retry.", payload["en"].Labels["wechatAvailabilityUnknown"])
	require.Equal(t, "This WeChat sign-in flow is only available in your system browser.", payload["en"].Labels["wechatSystemBrowserOnly"])
	require.Equal(t, "Reset Your Password", payload["en"].Labels["forgotPasswordTitle"])
	require.Equal(t, "Verify Your Email", payload["en"].Labels["emailVerifyTitle"])
	require.Equal(t, "Resend code in {countdown}s", payload["en"].Labels["emailVerifyResendCountdown"])
	require.Equal(t, "Send Reset Link", payload["en"].Labels["sendResetLink"])
	require.Equal(t, "Reset Link Sent", payload["en"].Labels["resetEmailSent"])
	require.Equal(t, "Set New Password", payload["en"].Labels["resetPasswordTitle"])
	require.Equal(t, "Reset Password", payload["en"].Labels["resetPassword"])
	require.Equal(t, "Password Reset Successful", payload["en"].Labels["passwordResetSuccess"])
	require.Equal(t, "Signing you in with {providerName}", payload["en"].Labels["providerCallbackTitle"])
	require.Equal(t, "Completing login with {providerName}, please wait...", payload["en"].Labels["providerCallbackProcessing"])
	require.Equal(t, "This {providerName} account is not yet registered. The site requires an invitation code — please enter one to complete registration.", payload["en"].Labels["providerInvitationRequired"])
	require.Equal(t, "Enter your new password", payload["en"].Labels["newPasswordPlaceholder"])
	require.Equal(t, "At least {count} characters", payload["en"].Labels["passwordHint"])
	require.Equal(t, "Optional", payload["en"].Labels["optional"])
	require.Equal(t, "Send Code", payload["en"].Labels["sendCode"])
	require.Equal(t, "Resend in {countdown}s", payload["en"].Labels["resendCountdown"])
	require.Equal(t, "Code sent successfully", payload["en"].Labels["codeSentSuccess"])
	require.Equal(t, "Two-Factor Authentication", payload["en"].Labels["totpLoginTitle"])
	require.Equal(t, "Verifying...", payload["en"].Labels["totpVerifying"])
	require.Equal(t, "Cancel", payload["en"].Labels["totpCancel"])
	require.Equal(t, "Accept and continue", payload["en"].Labels["agreementAcceptAndContinue"])
	require.Equal(t, "Our terms of service were updated on {date}. Please review and accept the following documents before continuing.", payload["en"].Labels["agreementUpdatedAt"])
	require.Equal(t, "All rights reserved.", payload["en"].Labels["allRightsReserved"])
}

func TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyRegistrationEnabled:          "true",
			SettingKeySiteName:                     "Sub2API Name",
			SettingKeySiteLogo:                     "https://static.example.com/site-logo.png",
			SettingKeySiteSubtitle:                 "Sub2API subtitle",
			SettingKeyGoogleOAuthEnabled:           "true",
			SettingKeyGoogleOAuthClientID:          "google-client",
			SettingKeyWebAppName:                   "Web Name",
			SettingKeyWebAppDescription:            "Web description",
			SettingKeyWebAppFavicon:                "https://static.example.com/favicon.ico",
			SettingKeyWebDefaultLocale:             "zh",
			SettingKeyPromptCasesTitle:             "Cases title",
			SettingKeyPromptCasesDescription:       "Cases description",
			SettingKeyPromptTemplatesTitle:         "Templates title",
			SettingKeyPromptTemplatesDescription:   "Templates description",
			SettingKeyPromptCatalogShellConfig:     `{"zh":{"labels":{"total":"总数"}}}`,
			SettingKeyWorkspaceShellConfig:         `{"zh":{"title":"工作台"}}`,
			SettingKeyPricingTitle:                 "Pricing title",
			SettingKeyPricingDescription:           "Pricing description",
			SettingKeyPricingShellConfig:           `{"zh":{"button":{"title":"选择"}}}`,
			SettingKeyPaymentShellConfig:           `{"zh":{"labels":{"createOrder":"下单"}}}`,
			SettingKeyPricingCurrencySymbol:        "$",
			SettingKeyCreditsTitle:                 "Credits title",
			SettingKeyCreditsDescription:           "Credits description",
			SettingKeyCreditsPurchaseLabel:         "Buy credits",
			SettingKeyCreditsBalanceLabel:          "Balance: {balance}",
			SettingKeyCreditsPerBalance:            "12",
			SettingKeyCreditsShellConfig:           `{"en":{"actions":{"title":"Balance actions"}}}`,
			SettingKeyWebLocaleDetectEnabled:       "true",
			SettingKeyWebPublicIntegrationsEnabled: "false",
			SettingKeyWebGoogleAnalyticsID:         "G-WEB",
			SettingKeyWebAffonsoEnabled:            "true",
			SettingKeyWebAffonsoID:                 "affonso-public",
			SettingKeyWebAffonsoCookieDuration:     "45",
			SettingKeyWebPromoteKitEnabled:         "true",
			SettingKeyWebPromoteKitID:              "promotekit-public",
			SettingKeyWebCrispEnabled:              "true",
			SettingKeyWebCrispWebsiteID:            "crisp-public",
			SettingKeyWebGoogleAuthVisible:         "false",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "Web Name", settings.WebAppName)
	require.Equal(t, "Web description", settings.WebAppDescription)
	require.Equal(t, "https://static.example.com/site-logo.png", settings.WebAppLogo)
	require.Equal(t, "https://static.example.com/favicon.ico", settings.WebAppFavicon)
	require.Equal(t, "zh", settings.WebDefaultLocale)
	require.Equal(t, "Cases title", settings.WebPromptCasesTitle)
	require.Equal(t, "Cases description", settings.WebPromptCasesDescription)
	require.Equal(t, "Templates title", settings.WebPromptTemplatesTitle)
	require.Equal(t, "Templates description", settings.WebPromptTemplatesDescription)
	require.Equal(t, "Cases title", settings.PromptCasesTitle)
	require.Equal(t, "Cases description", settings.PromptCasesDescription)
	require.Equal(t, "Templates title", settings.PromptTemplatesTitle)
	require.Equal(t, "Templates description", settings.PromptTemplatesDescription)
	require.Equal(t, `{"zh":{"labels":{"total":"总数"}}}`, settings.PromptCatalogShellConfig)
	require.Equal(t, `{"zh":{"title":"工作台"}}`, settings.WorkspaceShellConfig)
	require.Equal(t, "Pricing title", settings.PricingTitle)
	require.Equal(t, "Pricing description", settings.PricingDescription)
	require.Equal(t, `{"zh":{"button":{"title":"选择"}}}`, settings.PricingShellConfig)
	require.Equal(t, `{"zh":{"labels":{"createOrder":"下单"}}}`, settings.PaymentShellConfig)
	require.Equal(t, "$", settings.PricingCurrencySymbol)
	require.Equal(t, "Credits title", settings.CreditsTitle)
	require.Equal(t, "Credits description", settings.CreditsDescription)
	require.Equal(t, "Buy credits", settings.CreditsPurchaseLabel)
	require.Equal(t, "Balance: {balance}", settings.CreditsBalanceLabel)
	require.Equal(t, "12", settings.CreditsPerBalance)
	require.Equal(t, `{"en":{"actions":{"title":"Balance actions"}}}`, settings.CreditsShellConfig)
	require.Equal(t, "G-WEB", settings.GoogleAnalyticsID)
	require.False(t, settings.PublicIntegrationsEnabled)
	require.True(t, settings.AffonsoEnabled)
	require.Equal(t, "affonso-public", settings.AffonsoID)
	require.Equal(t, "45", settings.AffonsoCookieDuration)
	require.True(t, settings.PromoteKitEnabled)
	require.Equal(t, "promotekit-public", settings.PromoteKitID)
	require.True(t, settings.CrispEnabled)
	require.Equal(t, "crisp-public", settings.CrispWebsiteID)
	require.Equal(t, `{"zh":{"title":"工作台"}}`, settings.WebWorkspaceShellConfig)
	require.Equal(t, "Pricing title", settings.WebPricingTitle)
	require.Equal(t, "Pricing description", settings.WebPricingDescription)
	require.Equal(t, `{"zh":{"button":{"title":"选择"}}}`, settings.WebPricingShellConfig)
	require.Equal(t, `{"zh":{"labels":{"createOrder":"下单"}}}`, settings.WebPaymentShellConfig)
	require.Equal(t, "Credits title", settings.WebCreditsTitle)
	require.Equal(t, "Credits description", settings.WebCreditsDescription)
	require.Equal(t, "Buy credits", settings.WebCreditsPurchaseLabel)
	require.Equal(t, "Balance: {balance}", settings.WebCreditsBalanceLabel)
	require.True(t, settings.WebLocaleDetectEnabled)
	require.False(t, settings.WebPublicIntegrationsEnabled)
	require.True(t, settings.WebEmailAuthVisible)
	require.False(t, settings.WebGoogleAuthVisible)
	require.False(t, settings.WebGitHubAuthVisible)
	require.Equal(t, "G-WEB", settings.WebGoogleAnalyticsID)
	require.True(t, settings.WebAffonsoEnabled)
	require.Equal(t, "affonso-public", settings.WebAffonsoID)
	require.Equal(t, "45", settings.WebAffonsoCookieDuration)
	require.True(t, settings.WebPromoteKitEnabled)
	require.Equal(t, "promotekit-public", settings.WebPromoteKitID)
	require.True(t, settings.WebCrispEnabled)
	require.Equal(t, "crisp-public", settings.WebCrispWebsiteID)
}

func TestSettingService_GetPublicSettings_DefaultsAffonsoCookieDuration(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyWebAffonsoEnabled: "true",
			SettingKeyWebAffonsoID:      "affonso-public",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "30", settings.AffonsoCookieDuration)
	require.Equal(t, "30", settings.WebAffonsoCookieDuration)
}

func TestSettingService_GetPublicSettings_IgnoresLegacyTouchRuntimeKeys(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyRegistrationEnabled:        "true",
			SettingKeySiteName:                   "Sub2API Name",
			"touch_app_name":                     "Legacy Touch Name",
			"touch_prompt_cases_title":           "Legacy Cases",
			"touch_prompt_templates_description": "Legacy Templates",
			"touch_workspace_shell_config":       `{"zh":{"title":"旧工作台"}}`,
			"touch_pricing_title":                "Legacy Pricing",
			"touch_credits_balance_label":        "Legacy balance: {balance}",
			"touch_email_auth_visible":           "false",
			"touch_google_analytics_id":          "G-LEGACY",
			"touch_crisp_enabled":                "true",
			"touch_crisp_website_id":             "crisp-legacy",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "Sub2API Name", settings.WebAppName)
	require.Empty(t, settings.PromptCasesTitle)
	require.Empty(t, settings.PromptTemplatesDescription)
	require.NotEqual(t, `{"zh":{"title":"旧工作台"}}`, settings.WorkspaceShellConfig)
	require.Empty(t, settings.PricingTitle)
	require.Empty(t, settings.CreditsBalanceLabel)
	require.True(t, settings.WebEmailAuthVisible)
	require.Empty(t, settings.GoogleAnalyticsID)
	require.False(t, settings.CrispEnabled)
	require.Empty(t, settings.CrispWebsiteID)
}

func TestSettingService_GetPublicSettings_UsesOnlyGenericRuntimeKeys(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyRegistrationEnabled: "true",
			SettingKeySiteName:            "Sub2API Name",
			SettingKeyWebAppName:          "Web Name",
			"touch_app_name":              "Legacy Touch Name",
			SettingKeyPromptCasesTitle:    "Web Cases",
			"touch_prompt_cases_title":    "Legacy Cases",
			SettingKeyWebEmailAuthVisible: "false",
			"touch_email_auth_visible":    "false",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "Web Name", settings.WebAppName)
	require.Equal(t, "Web Cases", settings.PromptCasesTitle)
	require.False(t, settings.WebEmailAuthVisible)
}
