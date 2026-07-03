//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aias00/cloudbase/internal/hot"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// stubHotContentRepo implements hot.Repository for handler tests.
type stubHotContentRepo struct {
	sources    []hot.Source
	items      []hot.Item
	itemsTotal int64
	runEvents  []hot.RunEvent
	runTotal   int64
}

func (r *stubHotContentRepo) ListSources(_ context.Context) ([]hot.Source, error) {
	return r.sources, nil
}

func (r *stubHotContentRepo) ListItems(_ context.Context, params pagination.PaginationParams, filters hot.ListFilters) ([]hot.Item, *pagination.PaginationResult, error) {
	return r.items, &pagination.PaginationResult{Total: r.itemsTotal, Page: params.Page, PageSize: params.PageSize}, nil
}

func (r *stubHotContentRepo) ListRunEvents(_ context.Context, _ string, params pagination.PaginationParams) ([]hot.RunEvent, *pagination.PaginationResult, error) {
	return r.runEvents, &pagination.PaginationResult{Total: r.runTotal, Page: params.Page, PageSize: params.PageSize}, nil
}

func newTestHotContentHandler(repo hot.Repository) *HotContentHandler {
	return NewHotContentHandler(hot.NewService(repo))
}

func TestHotContentHandler_ListSources(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHotContentHandler(&stubHotContentRepo{
		sources: []hot.Source{{ID: 1, SourceID: "rss-test", Title: "Test RSS", Enabled: true}},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/hot/sources", nil)

	handler.ListSources(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	items, ok := data["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)
}

func TestHotContentHandler_ListItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHotContentHandler(&stubHotContentRepo{
		items:      []hot.Item{{ID: 1, Title: "Hot Item 1", Status: "published"}},
		itemsTotal: 1,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/hot/items?page=1&page_size=10", nil)
	c.Request.URL.RawQuery = "page=1&page_size=10"

	handler.ListItems(c)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestHotContentHandler_ListRunEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newTestHotContentHandler(&stubHotContentRepo{
		runEvents: []hot.RunEvent{{ID: 1, RunID: "run-1", Message: "completed"}},
		runTotal:  1,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/hot/run-events?run_id=run-1&page=1&page_size=10", nil)
	c.Request.URL.RawQuery = "run_id=run-1&page=1&page_size=10"

	handler.ListRunEvents(c)

	require.Equal(t, http.StatusOK, w.Code)
}
