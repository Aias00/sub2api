//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/cloudbase/internal/server/middleware"
	"github.com/Wei-Shaw/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type userAffiliateRepoStub struct {
	summary   *service.AffiliateSummary
	invitees  []service.AffiliateInvitee
	rebates   []service.AffiliateRebateRecord
	transfers []service.AffiliateTransferRecord
}

func (s *userAffiliateRepoStub) EnsureUserAffiliate(context.Context, int64) (*service.AffiliateSummary, error) {
	return s.summary, nil
}
func (s *userAffiliateRepoStub) GetAffiliateByCode(context.Context, string) (*service.AffiliateSummary, error) {
	return nil, service.ErrAffiliateProfileNotFound
}
func (s *userAffiliateRepoStub) BindInviter(context.Context, int64, int64) (bool, error) {
	return false, nil
}
func (s *userAffiliateRepoStub) AccrueQuota(context.Context, int64, int64, float64, int, *int64) (bool, error) {
	return false, nil
}
func (s *userAffiliateRepoStub) GetAccruedRebateFromInvitee(context.Context, int64, int64) (float64, error) {
	return 0, nil
}
func (s *userAffiliateRepoStub) ThawFrozenQuota(context.Context, int64) (float64, error) {
	return 0, nil
}
func (s *userAffiliateRepoStub) TransferQuotaToBalance(context.Context, int64) (float64, float64, error) {
	return 0, 0, nil
}
func (s *userAffiliateRepoStub) ListInvitees(context.Context, int64, int) ([]service.AffiliateInvitee, error) {
	return append([]service.AffiliateInvitee(nil), s.invitees...), nil
}
func (s *userAffiliateRepoStub) ListUserRebateRecords(context.Context, int64, service.AffiliateRecordFilter) ([]service.AffiliateRebateRecord, int64, error) {
	return append([]service.AffiliateRebateRecord(nil), s.rebates...), int64(len(s.rebates)), nil
}
func (s *userAffiliateRepoStub) ListUserTransferRecords(context.Context, int64, service.AffiliateRecordFilter) ([]service.AffiliateTransferRecord, int64, error) {
	return append([]service.AffiliateTransferRecord(nil), s.transfers...), int64(len(s.transfers)), nil
}
func (s *userAffiliateRepoStub) GetAffiliateOverview(context.Context) (*service.AffiliateAdminOverview, error) {
	return nil, nil
}
func (s *userAffiliateRepoStub) UpdateUserAffCode(context.Context, int64, string) error { return nil }
func (s *userAffiliateRepoStub) ResetUserAffCode(context.Context, int64) (string, error) {
	return "", nil
}
func (s *userAffiliateRepoStub) SetUserRebateRate(context.Context, int64, *float64) error { return nil }
func (s *userAffiliateRepoStub) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	return nil
}
func (s *userAffiliateRepoStub) ListUsersWithCustomSettings(context.Context, service.AffiliateAdminFilter) ([]service.AffiliateAdminEntry, int64, error) {
	return nil, 0, nil
}
func (s *userAffiliateRepoStub) ListAffiliateInviteRecords(context.Context, service.AffiliateRecordFilter) ([]service.AffiliateInviteRecord, int64, error) {
	return nil, 0, nil
}
func (s *userAffiliateRepoStub) ListAffiliateRebateRecords(context.Context, service.AffiliateRecordFilter) ([]service.AffiliateRebateRecord, int64, error) {
	return nil, 0, nil
}
func (s *userAffiliateRepoStub) ListAffiliateTransferRecords(context.Context, service.AffiliateRecordFilter) ([]service.AffiliateTransferRecord, int64, error) {
	return nil, 0, nil
}
func (s *userAffiliateRepoStub) GetAffiliateUserOverview(context.Context, int64) (*service.AffiliateUserOverview, error) {
	return nil, nil
}

func TestUserHandlerAffiliateRebatesEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createdAt := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	repo := &userAffiliateRepoStub{
		summary: &service.AffiliateSummary{UserID: 11, AffCode: "AFFCODE1"},
		rebates: []service.AffiliateRebateRecord{
			{
				OrderID:         101,
				InviteeID:       12,
				InviteeEmail:    "invitee@example.com",
				InviteeUsername: "invitee",
				OrderAmount:     99,
				PayAmount:       99,
				RebateAmount:    9.9,
				PaymentType:     "stripe",
				OrderStatus:     "paid",
				CreatedAt:       createdAt,
			},
		},
	}
	affiliateService := service.NewAffiliateService(repo, nil, nil, nil)
	handler := NewUserHandler(nil, nil, nil, nil, affiliateService, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/user/aff/rebates?page=1&page_size=20", nil)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 11})

	handler.GetAffiliateRebates(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				OrderID      int64   `json:"order_id"`
				InviteeEmail string  `json:"invitee_email"`
				RebateAmount float64 `json:"rebate_amount"`
			} `json:"items"`
			Total int64 `json:"total"`
			Page  int   `json:"page"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data.Items, 1)
	require.Equal(t, int64(101), resp.Data.Items[0].OrderID)
	require.Equal(t, "invitee@example.com", resp.Data.Items[0].InviteeEmail)
	require.Equal(t, 9.9, resp.Data.Items[0].RebateAmount)
	require.Equal(t, int64(1), resp.Data.Total)
	require.Equal(t, 1, resp.Data.Page)
}

func TestUserHandlerAffiliateTransfersEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createdAt := time.Date(2026, 5, 26, 11, 0, 0, 0, time.UTC)
	balanceAfter := 120.5
	repo := &userAffiliateRepoStub{
		summary: &service.AffiliateSummary{UserID: 11, AffCode: "AFFCODE1"},
		transfers: []service.AffiliateTransferRecord{
			{
				LedgerID:          201,
				UserID:            11,
				Amount:            20.5,
				BalanceAfter:      &balanceAfter,
				SnapshotAvailable: true,
				CreatedAt:         createdAt,
			},
		},
	}
	affiliateService := service.NewAffiliateService(repo, nil, nil, nil)
	handler := NewUserHandler(nil, nil, nil, nil, affiliateService, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/user/aff/transfers?page=1&page_size=20", nil)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 11})

	handler.GetAffiliateTransfers(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				LedgerID     int64    `json:"ledger_id"`
				Amount       float64  `json:"amount"`
				BalanceAfter *float64 `json:"balance_after"`
			} `json:"items"`
			Total int64 `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data.Items, 1)
	require.Equal(t, int64(201), resp.Data.Items[0].LedgerID)
	require.Equal(t, 20.5, resp.Data.Items[0].Amount)
	require.NotNil(t, resp.Data.Items[0].BalanceAfter)
	require.Equal(t, 120.5, *resp.Data.Items[0].BalanceAfter)
	require.Equal(t, int64(1), resp.Data.Total)
}
