//go:build unit

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dbent "github.com/Wei-Shaw/cloudbase/ent"
	"github.com/Wei-Shaw/cloudbase/ent/enttest"
	"github.com/Wei-Shaw/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type paymentPublicCatalogSettingRepoStub struct {
	values map[string]string
}

func (s *paymentPublicCatalogSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	return nil, service.ErrSettingNotFound
}

func (s *paymentPublicCatalogSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", service.ErrSettingNotFound
}

func (s *paymentPublicCatalogSettingRepoStub) Set(context.Context, string, string) error { return nil }

func (s *paymentPublicCatalogSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *paymentPublicCatalogSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	return nil
}
func (s *paymentPublicCatalogSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return s.values, nil
}
func (s *paymentPublicCatalogSettingRepoStub) Delete(context.Context, string) error { return nil }

func TestPaymentHandlerGetPublicCatalog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := sql.Open("sqlite", "file:payment_public_catalog?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	groupAnthropic, err := client.Group.Create().
		SetName("Claude").
		SetPlatform("anthropic").
		SetSubscriptionType("standard").
		SetDailyLimitUsd(25).
		SetMonthlyLimitUsd(120).
		SetSupportedModelScopes([]string{"Claude Opus 4.6", "Claude Sonnet 4.6"}).
		Save(context.Background())
	require.NoError(t, err)

	groupOpenAI, err := client.Group.Create().
		SetName("GPT").
		SetPlatform("openai").
		SetSubscriptionType("standard").
		SetSupportedModelScopes([]string{"GPT-5.4", "GPT-5.3 Codex"}).
		Save(context.Background())
	require.NoError(t, err)

	_, err = client.SubscriptionPlan.Create().
		SetGroupID(groupAnthropic.ID).
		SetName("Claude 开发包").
		SetDescription("复杂推理").
		SetPrice(59).
		SetValidityDays(30).
		SetValidityUnit("day").
		SetFeatures("priority").
		SetProductName("Claude 开发包").
		SetForSale(true).
		SetSortOrder(1).
		Save(context.Background())
	require.NoError(t, err)

	_, err = client.SubscriptionPlan.Create().
		SetGroupID(groupOpenAI.ID).
		SetName("GPT 开发包").
		SetDescription("代码生成").
		SetPrice(49).
		SetValidityDays(30).
		SetValidityUnit("day").
		SetFeatures("daily coding").
		SetProductName("GPT 开发包").
		SetForSale(true).
		SetSortOrder(2).
		Save(context.Background())
	require.NoError(t, err)

	rechargeProductsJSON := `[{"id":"starter","name":"入门包","description":"适合快速体验","amount":30,"credited_amount":45,"badge":"推荐","recommended":true,"features":["45 credits"],"sort_order":1}]`
	repo := &paymentPublicCatalogSettingRepoStub{values: map[string]string{
		service.SettingPaymentEnabled:      "true",
		service.SettingRechargeProducts:    rechargeProductsJSON,
		service.SettingBalanceRechargeMult: "1",
		service.SettingEnabledPaymentTypes: "stripe,alipay",
		service.SettingRechargeFeeRate:     "2.5",
	}}

	configSvc := service.NewPaymentConfigService(client, repo, nil)
	handler := NewPaymentHandler(nil, configSvc, nil)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/public/catalog", nil)

	handler.GetPublicCatalog(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			RechargeProducts []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"recharge_products"`
			Plans []struct {
				GroupPlatform        string   `json:"group_platform"`
				GroupDisplayLabel    string   `json:"group_display_label"`
				QuotaLabel           string   `json:"quota_label"`
				Name                 string   `json:"name"`
				SupportedModelScopes []string `json:"supported_model_scopes"`
			} `json:"plans"`
			EnabledPaymentTypes []string `json:"enabled_payment_types"`
			BalanceDisabled     bool     `json:"balance_disabled"`
			RechargeFeeRate     float64  `json:"recharge_fee_rate"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data.RechargeProducts, 1)
	require.Equal(t, "starter", resp.Data.RechargeProducts[0].ID)
	require.Len(t, resp.Data.Plans, 2)
	require.Equal(t, "anthropic", resp.Data.Plans[0].GroupPlatform)
	require.Equal(t, "Claude", resp.Data.Plans[0].GroupDisplayLabel)
	require.Equal(t, "$120", resp.Data.Plans[0].QuotaLabel)
	require.Contains(t, resp.Data.Plans[0].SupportedModelScopes, "Claude Opus 4.6")
	require.Equal(t, "openai", resp.Data.Plans[1].GroupPlatform)
	require.Equal(t, "OpenAI", resp.Data.Plans[1].GroupDisplayLabel)
	require.Equal(t, "", resp.Data.Plans[1].QuotaLabel)
	require.Equal(t, []string{"stripe", "alipay"}, resp.Data.EnabledPaymentTypes)
	require.False(t, resp.Data.BalanceDisabled)
	require.Equal(t, 2.5, resp.Data.RechargeFeeRate)
}
