//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type settingUpdateRepoStub struct {
	updates map[string]string
}

func (s *settingUpdateRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingUpdateRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *settingUpdateRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingUpdateRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *settingUpdateRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	s.updates = make(map[string]string, len(settings))
	for k, v := range settings {
		s.updates[k] = v
	}
	return nil
}

func (s *settingUpdateRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingUpdateRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

type settingAntigravityUARepoStub struct {
	values map[string]string
}

func (s *settingAntigravityUARepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingAntigravityUARepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *settingAntigravityUARepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingAntigravityUARepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *settingAntigravityUARepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *settingAntigravityUARepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingAntigravityUARepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

type defaultSubGroupReaderStub struct {
	byID  map[int64]*Group
	errBy map[int64]error
	calls []int64
}

func (s *defaultSubGroupReaderStub) GetByID(ctx context.Context, id int64) (*Group, error) {
	s.calls = append(s.calls, id)
	if err, ok := s.errBy[id]; ok {
		return nil, err
	}
	if g, ok := s.byID[id]; ok {
		return g, nil
	}
	return nil, ErrGroupNotFound
}

func TestSettingService_UpdateSettings_DefaultSubscriptions_ValidGroup(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	groupReader := &defaultSubGroupReaderStub{
		byID: map[int64]*Group{
			11: {ID: 11, SubscriptionType: SubscriptionTypeSubscription},
		},
	}
	svc := NewSettingService(repo, &config.Config{})
	svc.SetDefaultSubscriptionGroupReader(groupReader)

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		DefaultSubscriptions: []DefaultSubscriptionSetting{
			{GroupID: 11, ValidityDays: 30},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []int64{11}, groupReader.calls)

	raw, ok := repo.updates[SettingKeyDefaultSubscriptions]
	require.True(t, ok)

	var got []DefaultSubscriptionSetting
	require.NoError(t, json.Unmarshal([]byte(raw), &got))
	require.Equal(t, []DefaultSubscriptionSetting{
		{GroupID: 11, ValidityDays: 30},
	}, got)
}

func TestSettingService_UpdateSettings_SignupGrantRiskControl(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		SignupGrantRiskControlEnabled:              true,
		SignupGrantRiskControlEmailLimit:           1,
		SignupGrantRiskControlIPLimit:              3,
		SignupGrantRiskControlDomainLimit:          10,
		SignupGrantRiskControlOAuthIdentityEnabled: true,
		SignupGrantRiskControlDeviceEnabled:        true,
		SignupGrantRiskControlDeviceLimit:          2,
		SignupGrantRiskControlFreeDomainLimit:      5,
		SignupGrantRiskControlBlockedDomains:       " @TempMail.COM\n*.Trash.test ",
		SignupGrantRiskControlFreeDomains:          "gmail.com, QQ.com",
		SignupGrantRiskControlTrustedDomains:       "example.com; corp.test",
	})
	require.NoError(t, err)
	require.Equal(t, "true", repo.updates[SettingKeySignupGrantRiskControlEnabled])
	require.Equal(t, "1", repo.updates[SettingKeySignupGrantRiskControlEmailLimit])
	require.Equal(t, "3", repo.updates[SettingKeySignupGrantRiskControlIPLimit])
	require.Equal(t, "10", repo.updates[SettingKeySignupGrantRiskControlDomainLimit])
	require.Equal(t, "true", repo.updates[SettingKeySignupGrantRiskControlOAuthIdentityEnabled])
	require.Equal(t, "true", repo.updates[SettingKeySignupGrantRiskControlDeviceEnabled])
	require.Equal(t, "2", repo.updates[SettingKeySignupGrantRiskControlDeviceLimit])
	require.Equal(t, "5", repo.updates[SettingKeySignupGrantRiskControlFreeDomainLimit])
	require.Equal(t, "tempmail.com,trash.test", repo.updates[SettingKeySignupGrantRiskControlBlockedDomains])
	require.Equal(t, "gmail.com,qq.com", repo.updates[SettingKeySignupGrantRiskControlFreeDomains])
	require.Equal(t, "example.com,corp.test", repo.updates[SettingKeySignupGrantRiskControlTrustedDomains])
}

func TestSettingService_UpdateSettings_DefaultSubscriptions_RejectsNonSubscriptionGroup(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	groupReader := &defaultSubGroupReaderStub{
		byID: map[int64]*Group{
			12: {ID: 12, SubscriptionType: SubscriptionTypeStandard},
		},
	}
	svc := NewSettingService(repo, &config.Config{})
	svc.SetDefaultSubscriptionGroupReader(groupReader)

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		DefaultSubscriptions: []DefaultSubscriptionSetting{
			{GroupID: 12, ValidityDays: 7},
		},
	})
	require.Error(t, err)
	require.Equal(t, "DEFAULT_SUBSCRIPTION_GROUP_INVALID", infraerrors.Reason(err))
	require.Nil(t, repo.updates)
}

func TestSettingService_UpdateSettings_DefaultSubscriptions_RejectsNotFoundGroup(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	groupReader := &defaultSubGroupReaderStub{
		errBy: map[int64]error{
			13: ErrGroupNotFound,
		},
	}
	svc := NewSettingService(repo, &config.Config{})
	svc.SetDefaultSubscriptionGroupReader(groupReader)

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		DefaultSubscriptions: []DefaultSubscriptionSetting{
			{GroupID: 13, ValidityDays: 7},
		},
	})
	require.Error(t, err)
	require.Equal(t, "DEFAULT_SUBSCRIPTION_GROUP_INVALID", infraerrors.Reason(err))
	require.Equal(t, "13", infraerrors.FromError(err).Metadata["group_id"])
	require.Nil(t, repo.updates)
}

func TestSettingService_UpdateSettings_DefaultSubscriptions_RejectsDuplicateGroup(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	groupReader := &defaultSubGroupReaderStub{
		byID: map[int64]*Group{
			11: {ID: 11, SubscriptionType: SubscriptionTypeSubscription},
		},
	}
	svc := NewSettingService(repo, &config.Config{})
	svc.SetDefaultSubscriptionGroupReader(groupReader)

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		DefaultSubscriptions: []DefaultSubscriptionSetting{
			{GroupID: 11, ValidityDays: 30},
			{GroupID: 11, ValidityDays: 60},
		},
	})
	require.Error(t, err)
	require.Equal(t, "DEFAULT_SUBSCRIPTION_GROUP_DUPLICATE", infraerrors.Reason(err))
	require.Equal(t, "11", infraerrors.FromError(err).Metadata["group_id"])
	require.Nil(t, repo.updates)
}

func TestSettingService_UpdateSettings_DefaultSubscriptions_RejectsDuplicateGroupWithoutGroupReader(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		DefaultSubscriptions: []DefaultSubscriptionSetting{
			{GroupID: 11, ValidityDays: 30},
			{GroupID: 11, ValidityDays: 60},
		},
	})
	require.Error(t, err)
	require.Equal(t, "DEFAULT_SUBSCRIPTION_GROUP_DUPLICATE", infraerrors.Reason(err))
	require.Equal(t, "11", infraerrors.FromError(err).Metadata["group_id"])
	require.Nil(t, repo.updates)
}

func TestSettingService_UpdateSettings_RegistrationEmailSuffixWhitelist_Normalized(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		RegistrationEmailSuffixWhitelist: []string{"example.com", "@EXAMPLE.com", " @foo.bar ", "*.EDU.CN"},
	})
	require.NoError(t, err)
	require.Equal(t, `["@example.com","@foo.bar","*.edu.cn"]`, repo.updates[SettingKeyRegistrationEmailSuffixWhitelist])
}

func TestSettingService_UpdateSettings_RegistrationEmailSuffixWhitelist_Invalid(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		RegistrationEmailSuffixWhitelist: []string{"@invalid_domain"},
	})
	require.Error(t, err)
	require.Equal(t, "INVALID_REGISTRATION_EMAIL_SUFFIX_WHITELIST", infraerrors.Reason(err))
}

func TestSettingService_UpdateSettings_PersistsWebRuntimeSettings(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		WebAppName:                    " Prompt Web ",
		WebAppDescription:             " Managed by Sub2API ",
		WebAppLogo:                    " https://static.example.com/logo.png ",
		WebDefaultLocale:              " zh ",
		WebPromptCasesTitle:           " Cases title ",
		WebPromptCasesDescription:     " Cases description ",
		WebPromptTemplatesTitle:       " Templates title ",
		WebPromptTemplatesDescription: " Templates description ",
		PromptCatalogShellConfig:      ` {"zh":{"labels":{"total":"总数"}}} `,
		WebWorkspaceShellConfig:       ` {"zh":{"title":"工作台"}} `,
		WebPricingTitle:               " Pricing title ",
		WebPricingDescription:         " Pricing description ",
		WebPricingShellConfig:         ` {"zh":{"button":{"title":"选择"}}} `,
		WebPaymentShellConfig:         ` {"zh":{"labels":{"createOrder":"下单"}}} `,
		WebPricingCurrencySymbol:      " $ ",
		WebCreditsTitle:               " Balance title ",
		WebCreditsDescription:         " Balance description ",
		WebCreditsPurchaseLabel:       " Buy balance ",
		WebCreditsBalanceLabel:        " Balance: {balance} ",
		WebCreditsPerBalance:          " 12 ",
		CreditsShellConfig:            ` {"en":{"actions":{"title":"Balance actions"}}} `,
		WebLocaleDetectEnabled:        true,
		WebEmailAuthVisible:           true,
		WebGoogleAuthVisible:          false,
		WebGitHubAuthVisible:          true,
		WebGoogleAnalyticsID:          " G-WEB ",
		WebPublicIntegrationsEnabled:  false,
		WebVercelAnalyticsEnabled:     true,
		WebAdsenseCode:                " ca-pub-web ",
		WebAffonsoEnabled:             true,
		WebAffonsoID:                  " affonso-public ",
		WebPromoteKitEnabled:          true,
		WebPromoteKitID:               " promotekit-public ",
		WebCrispEnabled:               true,
		WebCrispWebsiteID:             " crisp-public ",
	})
	require.NoError(t, err)
	require.Equal(t, "Prompt Web", repo.updates[SettingKeyWebAppName])
	require.Equal(t, "Managed by Sub2API", repo.updates[SettingKeyWebAppDescription])
	require.Equal(t, "https://static.example.com/logo.png", repo.updates[SettingKeyWebAppLogo])
	require.Equal(t, "zh", repo.updates[SettingKeyWebDefaultLocale])
	require.Equal(t, "Cases title", repo.updates[SettingKeyPromptCasesTitle])
	require.Equal(t, "Cases description", repo.updates[SettingKeyPromptCasesDescription])
	require.Equal(t, "Templates title", repo.updates[SettingKeyPromptTemplatesTitle])
	require.Equal(t, "Templates description", repo.updates[SettingKeyPromptTemplatesDescription])
	require.Equal(t, `{"zh":{"labels":{"total":"总数"}}}`, repo.updates[SettingKeyPromptCatalogShellConfig])
	require.Equal(t, `{"zh":{"title":"工作台"}}`, repo.updates[SettingKeyWorkspaceShellConfig])
	require.Equal(t, "Pricing title", repo.updates[SettingKeyPricingTitle])
	require.Equal(t, "Pricing description", repo.updates[SettingKeyPricingDescription])
	require.Equal(t, `{"zh":{"button":{"title":"选择"}}}`, repo.updates[SettingKeyPricingShellConfig])
	require.Equal(t, `{"zh":{"labels":{"createOrder":"下单"}}}`, repo.updates[SettingKeyPaymentShellConfig])
	require.Equal(t, "$", repo.updates[SettingKeyPricingCurrencySymbol])
	require.Equal(t, "Balance title", repo.updates[SettingKeyCreditsTitle])
	require.Equal(t, "Balance description", repo.updates[SettingKeyCreditsDescription])
	require.Equal(t, "Buy balance", repo.updates[SettingKeyCreditsPurchaseLabel])
	require.Equal(t, "Balance: {balance}", repo.updates[SettingKeyCreditsBalanceLabel])
	require.Equal(t, "12", repo.updates[SettingKeyCreditsPerBalance])
	require.Equal(t, `{"en":{"actions":{"title":"Balance actions"}}}`, repo.updates[SettingKeyCreditsShellConfig])
	require.Equal(t, "true", repo.updates[SettingKeyWebLocaleDetectEnabled])
	require.Equal(t, "true", repo.updates[SettingKeyWebEmailAuthVisible])
	require.Equal(t, "false", repo.updates[SettingKeyWebGoogleAuthVisible])
	require.Equal(t, "true", repo.updates[SettingKeyWebGitHubAuthVisible])
	require.Equal(t, "G-WEB", repo.updates[SettingKeyWebGoogleAnalyticsID])
	require.Equal(t, "false", repo.updates[SettingKeyWebPublicIntegrationsEnabled])
	require.Equal(t, "true", repo.updates[SettingKeyWebVercelAnalyticsEnabled])
	require.Equal(t, "ca-pub-web", repo.updates[SettingKeyWebAdsenseCode])
	require.Equal(t, "true", repo.updates[SettingKeyWebAffonsoEnabled])
	require.Equal(t, "affonso-public", repo.updates[SettingKeyWebAffonsoID])
	require.Equal(t, "true", repo.updates[SettingKeyWebPromoteKitEnabled])
	require.Equal(t, "promotekit-public", repo.updates[SettingKeyWebPromoteKitID])
	require.Equal(t, "true", repo.updates[SettingKeyWebCrispEnabled])
	require.Equal(t, "crisp-public", repo.updates[SettingKeyWebCrispWebsiteID])
}

func TestParseDefaultSubscriptions_NormalizesValues(t *testing.T) {
	got := parseDefaultSubscriptions(`[{"group_id":11,"validity_days":30},{"group_id":11,"validity_days":60},{"group_id":0,"validity_days":10},{"group_id":12,"validity_days":99999}]`)
	require.Equal(t, []DefaultSubscriptionSetting{
		{GroupID: 11, ValidityDays: 30},
		{GroupID: 11, ValidityDays: 60},
		{GroupID: 12, ValidityDays: MaxValidityDays},
	}, got)
}

func TestSettingService_UpdateSettings_TablePreferences(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		TableDefaultPageSize: 50,
		TablePageSizeOptions: []int{20, 50, 100},
	})
	require.NoError(t, err)
	require.Equal(t, "50", repo.updates[SettingKeyTableDefaultPageSize])
	require.Equal(t, "[20,50,100]", repo.updates[SettingKeyTablePageSizeOptions])

	err = svc.UpdateSettings(context.Background(), &SystemSettings{
		TableDefaultPageSize: 1000,
		TablePageSizeOptions: []int{20, 100},
	})
	require.NoError(t, err)
	require.Equal(t, "1000", repo.updates[SettingKeyTableDefaultPageSize])
	require.Equal(t, "[20,100]", repo.updates[SettingKeyTablePageSizeOptions])
}

func TestSettingService_UpdateSettings_PaymentVisibleMethodsAndAdvancedScheduler(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		PaymentVisibleMethodAlipaySource:  "alipay",
		PaymentVisibleMethodWxpaySource:   "easypay",
		PaymentVisibleMethodAlipayEnabled: true,
		PaymentVisibleMethodWxpayEnabled:  false,
		OpenAIAdvancedSchedulerEnabled:    true,
	})
	require.NoError(t, err)
	require.Equal(t, VisibleMethodSourceOfficialAlipay, repo.updates[SettingPaymentVisibleMethodAlipaySource])
	require.Equal(t, VisibleMethodSourceEasyPayWechat, repo.updates[SettingPaymentVisibleMethodWxpaySource])
	require.Equal(t, "true", repo.updates[SettingPaymentVisibleMethodAlipayEnabled])
	require.Equal(t, "false", repo.updates[SettingPaymentVisibleMethodWxpayEnabled])
	require.Equal(t, "true", repo.updates[openAIAdvancedSchedulerSettingKey])
}

func TestSettingService_UpdateSettings_AntigravityUserAgentVersion(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		AntigravityUserAgentVersion: "1.23.2",
	})
	require.NoError(t, err)
	require.Equal(t, "1.23.2", repo.updates[SettingKeyAntigravityUserAgentVersion])
}

func TestSettingService_UpdateSettings_APIKeyACLTrustForwardedIPRefreshesConfig(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	cfg := &config.Config{}
	svc := NewSettingService(repo, cfg)

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		APIKeyACLTrustForwardedIP: true,
	})
	require.NoError(t, err)
	require.Equal(t, "true", repo.updates[SettingKeyAPIKeyACLTrustForwardedIP])
	require.True(t, cfg.Security.TrustForwardedIPForAPIKeyACL)
	require.True(t, cfg.TrustForwardedIPForAPIKeyACL())
}

func TestSettingService_ParseSettings_APIKeyACLTrustForwardedIPFallsBackToConfigWhenMissing(t *testing.T) {
	cfg := &config.Config{}
	cfg.Security.TrustForwardedIPForAPIKeyACL = true
	svc := NewSettingService(&settingUpdateRepoStub{}, cfg)

	got := svc.parseSettings(map[string]string{})

	require.True(t, got.APIKeyACLTrustForwardedIP)
}

func TestSettingService_GetAntigravityUserAgentVersion_Precedence(t *testing.T) {
	t.Run("后台设置优先", func(t *testing.T) {
		svc := NewSettingService(&settingAntigravityUARepoStub{values: map[string]string{
			SettingKeyAntigravityUserAgentVersion: "1.24.0",
		}}, &config.Config{})

		require.Equal(t, "1.24.0", svc.GetAntigravityUserAgentVersion(context.Background()))
	})

	t.Run("空值回退配置默认值", func(t *testing.T) {
		svc := NewSettingService(&settingAntigravityUARepoStub{values: map[string]string{
			SettingKeyAntigravityUserAgentVersion: "",
		}}, &config.Config{})

		require.Equal(t, antigravity.GetDefaultUserAgentVersion(), svc.GetAntigravityUserAgentVersion(context.Background()))
	})

	t.Run("缺失回退配置默认值", func(t *testing.T) {
		svc := NewSettingService(&settingAntigravityUARepoStub{values: map[string]string{}}, &config.Config{})

		require.Equal(t, antigravity.GetDefaultUserAgentVersion(), svc.GetAntigravityUserAgentVersion(context.Background()))
	})
}

func TestSettingService_UpdateSettings_RejectsInvalidPaymentVisibleMethodSource(t *testing.T) {
	repo := &settingUpdateRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		PaymentVisibleMethodAlipaySource: "not-a-provider",
	})
	require.Error(t, err)
	require.Equal(t, "INVALID_PAYMENT_VISIBLE_METHOD_SOURCE", infraerrors.Reason(err))
	require.Nil(t, repo.updates)
}
