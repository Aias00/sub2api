//go:build unit

package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type signupGrantRiskAdminStoreStub struct {
	*userRepoStub
	overrideInput        SignupGrantRiskOverrideInput
	deletedOverrideID    int64
	deletedOverrideAdmin int64
	addGiftCalls         int
	auditErr             error
}

func (s *signupGrantRiskAdminStoreStub) ListSignupGrantRiskClaims(context.Context, int, int, SignupGrantRiskClaimFilter) ([]SignupGrantRiskClaimRecord, int64, error) {
	return nil, 0, nil
}

func (s *signupGrantRiskAdminStoreStub) GetSignupGrantRiskUserSummary(context.Context, int64) (*SignupGrantRiskUserSummary, error) {
	return nil, nil
}

func (s *signupGrantRiskAdminStoreStub) UpsertSignupGrantRiskOverride(_ context.Context, input SignupGrantRiskOverrideInput) error {
	s.overrideInput = input
	return nil
}

func (s *signupGrantRiskAdminStoreStub) DeleteSignupGrantRiskOverride(_ context.Context, id int64, adminID int64) error {
	s.deletedOverrideID = id
	s.deletedOverrideAdmin = adminID
	return nil
}

func (s *signupGrantRiskAdminStoreStub) ListSignupGrantRiskOverrides(context.Context, int, int, SignupGrantRiskOverrideFilter) ([]SignupGrantRiskOverrideRecord, int64, error) {
	return nil, 0, nil
}

func (s *signupGrantRiskAdminStoreStub) CreateSignupGrantAdminAuditLog(context.Context, SignupGrantAdminAuditLog) error {
	return s.auditErr
}

func (s *signupGrantRiskAdminStoreStub) ListSignupGrantAdminAuditLogs(context.Context, int, int, SignupGrantAdminAuditLogFilter) ([]SignupGrantAdminAuditLog, int64, error) {
	return nil, 0, nil
}

func (s *signupGrantRiskAdminStoreStub) AddGiftBalance(context.Context, int64, float64) error {
	s.addGiftCalls++
	return nil
}

func TestAdminServiceUpsertSignupGrantRiskOverrideRejectsInvalidInputAsBadRequest(t *testing.T) {
	svc := &adminServiceImpl{userRepo: &signupGrantRiskAdminStoreStub{userRepoStub: &userRepoStub{}}}

	err := svc.UpsertSignupGrantRiskOverride(context.Background(), SignupGrantRiskOverrideInput{
		SubjectType: "email",
		Subject:     "",
		Action:      "block",
	})
	require.Equal(t, http.StatusBadRequest, infraerrors.Code(err))

	err = svc.UpsertSignupGrantRiskOverride(context.Background(), SignupGrantRiskOverrideInput{
		SubjectType: "email",
		Subject:     "user@example.com",
		Action:      "delete",
	})
	require.Equal(t, http.StatusBadRequest, infraerrors.Code(err))
}

func TestAdminServiceUpsertSignupGrantRiskOverrideAcceptsExistingHash(t *testing.T) {
	store := &signupGrantRiskAdminStoreStub{userRepoStub: &userRepoStub{}}
	svc := &adminServiceImpl{userRepo: store}
	hash := strings.Repeat("a", 64)

	err := svc.UpsertSignupGrantRiskOverride(context.Background(), SignupGrantRiskOverrideInput{
		SubjectType: "device",
		Subject:     hash,
		Action:      "block",
		Reason:      "manual",
	})

	require.NoError(t, err)
	require.Equal(t, hash, store.overrideInput.Subject)
	require.Equal(t, hash, store.overrideInput.SubjectValue)
	require.Equal(t, "device", store.overrideInput.SubjectType)
}

func TestAdminServiceDeleteSignupGrantRiskOverrideRejectsInvalidID(t *testing.T) {
	svc := &adminServiceImpl{userRepo: &signupGrantRiskAdminStoreStub{userRepoStub: &userRepoStub{}}}

	err := svc.DeleteSignupGrantRiskOverride(context.Background(), 0, 7)

	require.Equal(t, http.StatusBadRequest, infraerrors.Code(err))
}

func TestAdminServiceDeleteSignupGrantRiskOverrideDelegatesToStore(t *testing.T) {
	store := &signupGrantRiskAdminStoreStub{userRepoStub: &userRepoStub{}}
	svc := &adminServiceImpl{userRepo: store}

	err := svc.DeleteSignupGrantRiskOverride(context.Background(), 42, 7)

	require.NoError(t, err)
	require.EqualValues(t, 42, store.deletedOverrideID)
	require.EqualValues(t, 7, store.deletedOverrideAdmin)
}

func TestAdminServiceManualGrantSignupGiftBalanceFailsWhenAuditFails(t *testing.T) {
	store := &signupGrantRiskAdminStoreStub{
		userRepoStub: &userRepoStub{
			user: &User{ID: 11, Balance: 3, Status: StatusActive},
		},
		auditErr: errors.New("audit failed"),
	}
	svc := &adminServiceImpl{userRepo: store}

	user, err := svc.ManualGrantSignupGiftBalance(context.Background(), 11, 5, "manual", 7)

	require.Nil(t, user)
	require.Error(t, err)
	require.Equal(t, 1, store.addGiftCalls)
}
