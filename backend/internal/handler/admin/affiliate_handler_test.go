//go:build unit

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aias00/cloudbase/internal/config"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type affiliateHandlerRepoStub struct {
	overview *service.AffiliateAdminOverview
}

func (s *affiliateHandlerRepoStub) EnsureUserAffiliate(context.Context, int64) (*service.AffiliateSummary, error) {
	return &service.AffiliateSummary{UserID: 1, AffCode: "AFFCODE1"}, nil
}
func (s *affiliateHandlerRepoStub) GetAffiliateByCode(context.Context, string) (*service.AffiliateSummary, error) {
	return nil, service.ErrAffiliateProfileNotFound
}
func (s *affiliateHandlerRepoStub) BindInviter(context.Context, int64, int64) (bool, error) {
	return false, nil
}
func (s *affiliateHandlerRepoStub) AccrueQuota(context.Context, int64, int64, float64, int, *int64) (bool, error) {
	return false, nil
}
func (s *affiliateHandlerRepoStub) GetAccruedRebateFromInvitee(context.Context, int64, int64) (float64, error) {
	return 0, nil
}
func (s *affiliateHandlerRepoStub) ThawFrozenQuota(context.Context, int64) (float64, error) {
	return 0, nil
}
func (s *affiliateHandlerRepoStub) TransferQuotaToBalance(context.Context, int64) (float64, float64, error) {
	return 0, 0, nil
}
func (s *affiliateHandlerRepoStub) ListInvitees(context.Context, int64, int) ([]service.AffiliateInvitee, error) {
	return nil, nil
}
func (s *affiliateHandlerRepoStub) ListUserRebateRecords(context.Context, int64, service.AffiliateRecordFilter) ([]service.AffiliateRebateRecord, int64, error) {
	return nil, 0, nil
}
func (s *affiliateHandlerRepoStub) ListUserTransferRecords(context.Context, int64, service.AffiliateRecordFilter) ([]service.AffiliateTransferRecord, int64, error) {
	return nil, 0, nil
}
func (s *affiliateHandlerRepoStub) GetAffiliateOverview(context.Context) (*service.AffiliateAdminOverview, error) {
	return s.overview, nil
}
func (s *affiliateHandlerRepoStub) UpdateUserAffCode(context.Context, int64, string) error {
	return nil
}
func (s *affiliateHandlerRepoStub) ResetUserAffCode(context.Context, int64) (string, error) {
	return "", nil
}
func (s *affiliateHandlerRepoStub) SetUserRebateRate(context.Context, int64, *float64) error {
	return nil
}
func (s *affiliateHandlerRepoStub) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	return nil
}
func (s *affiliateHandlerRepoStub) ListUsersWithCustomSettings(context.Context, service.AffiliateAdminFilter) ([]service.AffiliateAdminEntry, int64, error) {
	return nil, 0, nil
}
func (s *affiliateHandlerRepoStub) ListAffiliateInviteRecords(context.Context, service.AffiliateRecordFilter) ([]service.AffiliateInviteRecord, int64, error) {
	return nil, 0, nil
}
func (s *affiliateHandlerRepoStub) ListAffiliateRebateRecords(context.Context, service.AffiliateRecordFilter) ([]service.AffiliateRebateRecord, int64, error) {
	return nil, 0, nil
}
func (s *affiliateHandlerRepoStub) ListAffiliateTransferRecords(context.Context, service.AffiliateRecordFilter) ([]service.AffiliateTransferRecord, int64, error) {
	return nil, 0, nil
}
func (s *affiliateHandlerRepoStub) GetAffiliateUserOverview(context.Context, int64) (*service.AffiliateUserOverview, error) {
	return nil, nil
}

type affiliateHandlerSettingRepoStub struct {
	values map[string]string
}

func (s *affiliateHandlerSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	return nil, service.ErrSettingNotFound
}
func (s *affiliateHandlerSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", service.ErrSettingNotFound
}
func (s *affiliateHandlerSettingRepoStub) Set(_ context.Context, key, value string) error {
	s.values[key] = value
	return nil
}
func (s *affiliateHandlerSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}
func (s *affiliateHandlerSettingRepoStub) SetMultiple(_ context.Context, settings map[string]string) error {
	for key, value := range settings {
		s.values[key] = value
	}
	return nil
}
func (s *affiliateHandlerSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.values))
	for key, value := range s.values {
		out[key] = value
	}
	return out, nil
}
func (s *affiliateHandlerSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(s.values, key)
	return nil
}

func TestAffiliateHandlerOverviewAndRulesEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	settingRepo := &affiliateHandlerSettingRepoStub{
		values: map[string]string{
			service.SettingKeyAffiliateEnabled:             "true",
			service.SettingKeyInvitationCodeEnabled:        "true",
			service.SettingKeyAffiliateRebateRate:          "15",
			service.SettingKeyAffiliateRebateFreezeHours:   "48",
			service.SettingKeyAffiliateRebateDurationDays:  "90",
			service.SettingKeyAffiliateRebatePerInviteeCap: "100",
		},
	}
	settingService := service.NewSettingService(settingRepo, &config.Config{
		Default: config.DefaultConfig{
			UserConcurrency: 5,
		},
	})
	repo := &affiliateHandlerRepoStub{
		overview: &service.AffiliateAdminOverview{
			InvitedUserCount:        12,
			RebatedInviteeCount:     7,
			AvailableQuotaTotal:     88.5,
			FrozenQuotaTotal:        9,
			HistoryQuotaTotal:       120,
			RecentRebateRecordCount: 3,
		},
	}
	affiliateService := service.NewAffiliateService(repo, settingService, nil, nil)
	handler := NewAffiliateHandler(affiliateService, nil)

	router := gin.New()
	router.GET("/api/v1/admin/affiliates/overview", handler.GetOverview)
	router.GET("/api/v1/admin/affiliates/rules", handler.GetRules)
	router.PUT("/api/v1/admin/affiliates/rules", handler.UpdateRules)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/affiliates/overview", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var overviewResp struct {
		Code int `json:"code"`
		Data struct {
			AffiliateEnabled        bool    `json:"affiliate_enabled"`
			InvitationCodeEnabled   bool    `json:"invitation_code_enabled"`
			AffiliateRebateRate     float64 `json:"affiliate_rebate_rate"`
			InvitedUserCount        int64   `json:"invited_user_count"`
			RebatedInviteeCount     int64   `json:"rebated_invitee_count"`
			AvailableQuotaTotal     float64 `json:"available_quota_total"`
			FrozenQuotaTotal        float64 `json:"frozen_quota_total"`
			HistoryQuotaTotal       float64 `json:"history_quota_total"`
			RecentRebateRecordCount int64   `json:"recent_rebate_record_count"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &overviewResp))
	require.Equal(t, 0, overviewResp.Code)
	require.True(t, overviewResp.Data.AffiliateEnabled)
	require.True(t, overviewResp.Data.InvitationCodeEnabled)
	require.Equal(t, 15.0, overviewResp.Data.AffiliateRebateRate)
	require.Equal(t, int64(12), overviewResp.Data.InvitedUserCount)
	require.Equal(t, int64(3), overviewResp.Data.RecentRebateRecordCount)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/affiliates/rules", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var rulesResp struct {
		Code int `json:"code"`
		Data struct {
			AffiliateEnabled            bool    `json:"affiliate_enabled"`
			InvitationCodeEnabled       bool    `json:"invitation_code_enabled"`
			AffiliateRebateRate         float64 `json:"affiliate_rebate_rate"`
			AffiliateRebateFreezeHours  int     `json:"affiliate_rebate_freeze_hours"`
			AffiliateRebateDurationDays int     `json:"affiliate_rebate_duration_days"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rulesResp))
	require.Equal(t, 0, rulesResp.Code)
	require.True(t, rulesResp.Data.AffiliateEnabled)
	require.Equal(t, 48, rulesResp.Data.AffiliateRebateFreezeHours)

	body, err := json.Marshal(map[string]any{
		"affiliate_enabled":                false,
		"invitation_code_enabled":          false,
		"affiliate_rebate_rate":            20,
		"affiliate_rebate_freeze_hours":    24,
		"affiliate_rebate_duration_days":   60,
		"affiliate_rebate_per_invitee_cap": 88.8,
	})
	require.NoError(t, err)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/affiliates/rules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", settingRepo.values[service.SettingKeyAffiliateEnabled])
	require.Equal(t, "20.00000000", settingRepo.values[service.SettingKeyAffiliateRebateRate])
	require.Equal(t, "24", settingRepo.values[service.SettingKeyAffiliateRebateFreezeHours])
}
