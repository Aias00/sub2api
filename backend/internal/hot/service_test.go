//go:build unit

package hot

import (
	"context"
	"errors"
	"testing"

	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

// hotContentRepoStub implements Repository for testing.
type hotContentRepoStub struct {
	sources    []Source
	items      []Item
	itemsTotal int64
	runEvents  []RunEvent
	runTotal   int64
	err        error
}

func (s *hotContentRepoStub) ListSources(ctx context.Context) ([]Source, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.sources, nil
}

func (s *hotContentRepoStub) ListItems(ctx context.Context, params pagination.PaginationParams, filters ListFilters) ([]Item, *pagination.PaginationResult, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.items, &pagination.PaginationResult{Total: s.itemsTotal, Page: params.Page, PageSize: params.PageSize}, nil
}

func (s *hotContentRepoStub) ListRunEvents(ctx context.Context, runID string, params pagination.PaginationParams) ([]RunEvent, *pagination.PaginationResult, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.runEvents, &pagination.PaginationResult{Total: s.runTotal, Page: params.Page, PageSize: params.PageSize}, nil
}

func TestService_Health_NilService(t *testing.T) {
	var s *Service
	require.ErrorIs(t, s.Health(), ErrHotContentInvalidInput)
}

func TestService_Health_NilRepo(t *testing.T) {
	s := &Service{}
	require.ErrorIs(t, s.Health(), ErrHotContentInvalidInput)
}

func TestService_Health_Valid(t *testing.T) {
	s := NewService(&hotContentRepoStub{})
	require.NoError(t, s.Health())
}

func TestService_ListSources(t *testing.T) {
	repo := &hotContentRepoStub{
		sources: []Source{{ID: 1, SourceID: "rss-1", Title: "Test Source"}},
	}
	svc := NewService(repo)
	result, err := svc.ListSources(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "rss-1", result[0].SourceID)
}

func TestService_ListSources_Error(t *testing.T) {
	repo := &hotContentRepoStub{err: errors.New("db error")}
	svc := NewService(repo)
	_, err := svc.ListSources(context.Background())
	require.Error(t, err)
}

func TestService_ListItems_DefaultStatus(t *testing.T) {
	repo := &hotContentRepoStub{
		items:      []Item{{ID: 1, Title: "Test Item"}},
		itemsTotal: 1,
	}
	svc := NewService(repo)
	result, pag, err := svc.ListItems(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 20}, ListFilters{})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, int64(1), pag.Total)
}

func TestService_ListItems_NormalizesPagination(t *testing.T) {
	repo := &hotContentRepoStub{}
	svc := NewService(repo)

	// Page < 1 should be clamped to 1
	_, _, err := svc.ListItems(context.Background(), pagination.PaginationParams{Page: -1, PageSize: 20}, ListFilters{})
	require.NoError(t, err)

	// PageSize < 1 should be clamped to 20
	_, _, err = svc.ListItems(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 0}, ListFilters{})
	require.NoError(t, err)

	// PageSize > 100 should be clamped to 100
	_, _, err = svc.ListItems(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 200}, ListFilters{})
	require.NoError(t, err)
}

func TestService_ListItems_TrimsFilter(t *testing.T) {
	repo := &hotContentRepoStub{
		items:      []Item{},
		itemsTotal: 0,
	}
	svc := NewService(repo)
	_, _, err := svc.ListItems(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 20}, ListFilters{
		SourceID: "  rss-1  ",
		Query:    "  test  ",
		Status:   "  published  ",
	})
	require.NoError(t, err)
}

func TestService_ListRunEvents_EmptyRunID(t *testing.T) {
	svc := NewService(&hotContentRepoStub{})
	_, _, err := svc.ListRunEvents(context.Background(), "  ", pagination.PaginationParams{Page: 1, PageSize: 20})
	require.ErrorIs(t, err, ErrHotContentInvalidInput)
}

func TestService_ListRunEvents_Success(t *testing.T) {
	repo := &hotContentRepoStub{
		runEvents: []RunEvent{{ID: 1, RunID: "run-1", Message: "completed"}},
		runTotal:  1,
	}
	svc := NewService(repo)
	result, pag, err := svc.ListRunEvents(context.Background(), "run-1", pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, int64(1), pag.Total)
}
