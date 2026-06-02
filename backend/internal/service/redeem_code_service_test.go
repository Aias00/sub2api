package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type redeemLookupStub struct {
	codes map[string]*RedeemCode
}

func (s *redeemLookupStub) Create(context.Context, *RedeemCode) error       { return nil }
func (s *redeemLookupStub) CreateBatch(context.Context, []RedeemCode) error { return nil }
func (s *redeemLookupStub) GetByID(context.Context, int64) (*RedeemCode, error) {
	return nil, ErrRedeemCodeNotFound
}
func (s *redeemLookupStub) GetByCode(_ context.Context, code string) (*RedeemCode, error) {
	if c, ok := s.codes[code]; ok {
		return c, nil
	}
	return nil, ErrRedeemCodeNotFound
}
func (s *redeemLookupStub) Update(context.Context, *RedeemCode) error { return nil }
func (s *redeemLookupStub) Delete(context.Context, int64) error       { return nil }
func (s *redeemLookupStub) Use(context.Context, int64, int64) error   { return nil }
func (s *redeemLookupStub) List(context.Context, pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *redeemLookupStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *redeemLookupStub) ListByUser(context.Context, int64, int) ([]RedeemCode, error) {
	return nil, nil
}
func (s *redeemLookupStub) ListByUserPaginated(context.Context, int64, pagination.PaginationParams, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *redeemLookupStub) SumPositiveBalanceByUser(context.Context, int64) (float64, error) {
	return 0, nil
}
func (s *redeemLookupStub) BatchUpdate(context.Context, []int64, RedeemCodeBatchUpdateFields) (int64, error) {
	return 0, nil
}

func TestGenerateRedeemCodeFormat(t *testing.T) {
	code, err := GenerateRedeemCode()
	require.NoError(t, err)
	require.Len(t, code, 8)
	require.Regexp(t, `^[A-Z0-9]{8}$`, code)
}

func TestRedeemServiceFindRedeemCodeByInputFallsBackToUppercase(t *testing.T) {
	svc := &RedeemService{redeemRepo: &redeemLookupStub{codes: map[string]*RedeemCode{
		"ABC12345":   {Code: "ABC12345", Type: RedeemTypeBalance, Status: StatusUnused},
		"legacyabcd": {Code: "legacyabcd", Type: RedeemTypeBalance, Status: StatusUnused},
	}}}

	modern, err := svc.findRedeemCodeByInput(context.Background(), "abc12345")
	require.NoError(t, err)
	require.Equal(t, "ABC12345", modern.Code)

	legacy, err := svc.findRedeemCodeByInput(context.Background(), "legacyabcd")
	require.NoError(t, err)
	require.Equal(t, "legacyabcd", legacy.Code)
}
