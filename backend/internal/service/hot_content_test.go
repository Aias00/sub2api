//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

// hotContentRepoStub implements HotContentRepository for testing.
type hotContentRepoStub struct {
	sources    []HotSource
	items      []HotItem
	itemsTotal int64
	runEvents  []HotRunEvent
	runTotal   int64
	err        error
}

func (s *hotContentRepoStub) ListSources(ctx context.Context) ([]HotSource, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.sources, nil
}

func (s *hotContentRepoStub) ListItems(ctx context.Context, params pagination.PaginationParams, filters HotContentListFilters) ([]HotItem, *pagination.PaginationResult, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.items, &pagination.PaginationResult{Total: s.itemsTotal, Page: params.Page, PageSize: params.PageSize}, nil
}

func (s *hotContentRepoStub) ListRunEvents(ctx context.Context, runID string, params pagination.PaginationParams) ([]HotRunEvent, *pagination.PaginationResult, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.runEvents, &pagination.PaginationResult{Total: s.runTotal, Page: params.Page, PageSize: params.PageSize}, nil
}

func TestHotContentService_Health_NilService(t *testing.T) {
	var s *HotContentService
	require.ErrorIs(t, s.Health(), ErrHotContentInvalidInput)
}

func TestHotContentService_Health_NilRepo(t *testing.T) {
	s := &HotContentService{}
	require.ErrorIs(t, s.Health(), ErrHotContentInvalidInput)
}

func TestHotContentService_Health_Valid(t *testing.T) {
	s := NewHotContentService(&hotContentRepoStub{})
	require.NoError(t, s.Health())
}

func TestHotContentService_ListSources(t *testing.T) {
	repo := &hotContentRepoStub{
		sources: []HotSource{{ID: 1, SourceID: "rss-1", Title: "Test Source"}},
	}
	svc := NewHotContentService(repo)
	result, err := svc.ListSources(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "rss-1", result[0].SourceID)
}

func TestHotContentService_ListSources_Error(t *testing.T) {
	repo := &hotContentRepoStub{err: errors.New("db error")}
	svc := NewHotContentService(repo)
	_, err := svc.ListSources(context.Background())
	require.Error(t, err)
}

func TestHotContentService_ListItems_DefaultStatus(t *testing.T) {
	repo := &hotContentRepoStub{
		items:      []HotItem{{ID: 1, Title: "Test Item"}},
		itemsTotal: 1,
	}
	svc := NewHotContentService(repo)
	result, pag, err := svc.ListItems(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 20}, HotContentListFilters{})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, int64(1), pag.Total)
}

func TestHotContentService_ListItems_NormalizesPagination(t *testing.T) {
	repo := &hotContentRepoStub{}
	svc := NewHotContentService(repo)

	// Page < 1 should be clamped to 1
	_, _, err := svc.ListItems(context.Background(), pagination.PaginationParams{Page: -1, PageSize: 20}, HotContentListFilters{})
	require.NoError(t, err)

	// PageSize < 1 should be clamped to 20
	_, _, err = svc.ListItems(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 0}, HotContentListFilters{})
	require.NoError(t, err)

	// PageSize > 100 should be clamped to 100
	_, _, err = svc.ListItems(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 200}, HotContentListFilters{})
	require.NoError(t, err)
}

func TestHotContentService_ListItems_TrimsFilter(t *testing.T) {
	repo := &hotContentRepoStub{
		items:      []HotItem{},
		itemsTotal: 0,
	}
	svc := NewHotContentService(repo)
	_, _, err := svc.ListItems(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 20}, HotContentListFilters{
		SourceID: "  rss-1  ",
		Query:    "  test  ",
		Status:   "  published  ",
	})
	require.NoError(t, err)
}

func TestHotContentService_ListRunEvents_EmptyRunID(t *testing.T) {
	svc := NewHotContentService(&hotContentRepoStub{})
	_, _, err := svc.ListRunEvents(context.Background(), "  ", pagination.PaginationParams{Page: 1, PageSize: 20})
	require.ErrorIs(t, err, ErrHotContentInvalidInput)
}

func TestHotContentService_ListRunEvents_Success(t *testing.T) {
	repo := &hotContentRepoStub{
		runEvents: []HotRunEvent{{ID: 1, RunID: "run-1", Message: "completed"}},
		runTotal:  1,
	}
	svc := NewHotContentService(repo)
	result, pag, err := svc.ListRunEvents(context.Background(), "run-1", pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, int64(1), pag.Total)
}
