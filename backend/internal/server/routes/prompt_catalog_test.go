package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type promptCatalogRouteRepoStub struct {
	items   []service.PromptCatalogCase
	page    *pagination.PaginationResult
	summary *service.PromptCatalogSummary
	filters service.PromptCatalogListFilters
	params  pagination.PaginationParams
	upsert  *service.PromptCatalogCase
}

func (s *promptCatalogRouteRepoStub) ListCases(ctx context.Context, params pagination.PaginationParams, filters service.PromptCatalogListFilters) ([]service.PromptCatalogCase, *pagination.PaginationResult, error) {
	s.params = params
	s.filters = filters
	return s.items, s.page, nil
}

func (s *promptCatalogRouteRepoStub) GetCaseSummary(ctx context.Context, filters service.PromptCatalogListFilters) (*service.PromptCatalogSummary, error) {
	if s.summary != nil {
		return s.summary, nil
	}
	return &service.PromptCatalogSummary{}, nil
}

func (s *promptCatalogRouteRepoStub) GetCaseByID(ctx context.Context, id string) (*service.PromptCatalogCase, error) {
	for _, item := range s.items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, service.ErrPromptCatalogNotFound
}

func (s *promptCatalogRouteRepoStub) UpsertCase(ctx context.Context, item *service.PromptCatalogCase) error {
	s.upsert = item
	return nil
}

func TestPromptCasesPublicRouteFiltersAndSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	importedAt := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	repo := &promptCatalogRouteRepoStub{
		items: []service.PromptCatalogCase{{
			ID:            "case-1",
			Title:         "Toy Portrait",
			Prompt:        "prompt body",
			PromptPreview: "prompt body",
			Category:      "portrait",
			Tags:          []string{"twitter", "cinematic", "editorial", "studio", "portrait", "extra", "ignored"},
			DisplayTags:   []string{"cinematic", "editorial", "studio"},
			ModelTags:     []string{"OpenAI Image"},
			ImageThumbURL: "https://static.example/case-1-thumb.jpg",
			ImageURLs:     []string{"https://static.example/case-1.jpg"},
			SourceProject: "Twitter Imports",
			SourceLabel:   "X Imports",
			SourceType:    "case",
			ImportSource:  "twitter",
			Status:        service.PromptCatalogStatusPublished,
			ImportedAt:    &importedAt,
			CreatedAt:     importedAt,
			UpdatedAt:     importedAt,
		}},
		page: &pagination.PaginationResult{Total: 1, Page: 1, PageSize: 20, Pages: 1},
		summary: &service.PromptCatalogSummary{
			Total:         1,
			CaseCount:     1,
			TemplateCount: 0,
			SourceCount:   1,
			CategoryCount: 1,
			Sources:       []service.PromptCatalogFacetCount{{Value: "Twitter Imports", Label: "X Imports", Count: 1}},
			Categories:    []service.PromptCatalogFacetCount{{Value: "portrait", Count: 1}},
		},
	}
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterPromptRoutes(v1, &handler.Handlers{
		PromptCatalog: handler.NewPromptCatalogHandler(service.NewPromptCatalogService(repo)),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/prompts/cases?source_type=case&source_project=Twitter+Imports&category=portrait&featured=true&has_image=true&sort_by=title&sort_order=asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "case", repo.filters.SourceType)
	require.Equal(t, "Twitter Imports", repo.filters.SourceProject)
	require.Equal(t, "portrait", repo.filters.Category)
	require.NotNil(t, repo.filters.Featured)
	require.True(t, *repo.filters.Featured)
	require.NotNil(t, repo.filters.HasImage)
	require.True(t, *repo.filters.HasImage)
	require.Equal(t, "title", repo.params.SortBy)
	require.Equal(t, "asc", repo.params.SortOrder)

	var body struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				ID                 string   `json:"id"`
				PrimaryImageURL    string   `json:"primary_image_url"`
				ImageURLs          []string `json:"image_urls"`
				DisplayTags        []string `json:"display_tags"`
				ModelTags          []string `json:"model_tags"`
				AllTags            []string `json:"all_tags"`
				VisibleTags        []string `json:"visible_tags"`
				SourceDisplayLabel string   `json:"source_display_label"`
				PromptCharCount    int      `json:"prompt_char_count"`
				ImportedAt         *string  `json:"imported_at"`
			} `json:"items"`
			Total   int64 `json:"total"`
			Summary struct {
				Total       int64 `json:"total"`
				CaseCount   int64 `json:"case_count"`
				SourceCount int   `json:"source_count"`
				Sources     []struct {
					Value        string `json:"value"`
					Label        string `json:"label"`
					Count        int64  `json:"count"`
					DisplayLabel string `json:"display_label"`
				} `json:"sources"`
				Categories []struct {
					Value        string `json:"value"`
					Count        int64  `json:"count"`
					DisplayLabel string `json:"display_label"`
				} `json:"categories"`
				TemplateGroups []struct {
					Value        string `json:"value"`
					Count        int64  `json:"count"`
					DisplayLabel string `json:"display_label"`
				} `json:"template_groups"`
			} `json:"summary"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, int64(1), body.Data.Total)
	require.Len(t, body.Data.Items, 1)
	require.Equal(t, "case-1", body.Data.Items[0].ID)
	require.Equal(t, "https://static.example/case-1-thumb.jpg", body.Data.Items[0].PrimaryImageURL)
	require.Equal(t, []string{"https://static.example/case-1.jpg"}, body.Data.Items[0].ImageURLs)
	require.Equal(t, []string{"cinematic", "editorial", "studio"}, body.Data.Items[0].DisplayTags)
	require.Equal(t, []string{"OpenAI Image"}, body.Data.Items[0].ModelTags)
	require.Equal(t, []string{"OpenAI Image", "cinematic", "editorial", "studio", "twitter", "portrait", "extra", "ignored"}, body.Data.Items[0].AllTags)
	require.Equal(t, []string{"OpenAI Image", "cinematic", "editorial", "studio", "twitter", "portrait"}, body.Data.Items[0].VisibleTags)
	require.Equal(t, "X Imports", body.Data.Items[0].SourceDisplayLabel)
	require.Equal(t, len([]rune("prompt body")), body.Data.Items[0].PromptCharCount)
	require.NotNil(t, body.Data.Items[0].ImportedAt)
	require.Equal(t, "2026-06-09T10:00:00Z", *body.Data.Items[0].ImportedAt)
	require.Equal(t, int64(1), body.Data.Summary.Total)
	require.Equal(t, int64(1), body.Data.Summary.CaseCount)
	require.Equal(t, 1, body.Data.Summary.SourceCount)
	require.Len(t, body.Data.Summary.Sources, 1)
	require.Equal(t, "Twitter Imports", body.Data.Summary.Sources[0].Value)
	require.Equal(t, "X Imports", body.Data.Summary.Sources[0].Label)
	require.Equal(t, "X Imports (1)", body.Data.Summary.Sources[0].DisplayLabel)
	require.Len(t, body.Data.Summary.Categories, 1)
	require.Equal(t, "portrait (1)", body.Data.Summary.Categories[0].DisplayLabel)
	require.Empty(t, body.Data.Summary.TemplateGroups)
}

func TestPromptCatalogPublicAliasIsNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterPromptRoutes(v1, &handler.Handlers{
		PromptCatalog: handler.NewPromptCatalogHandler(service.NewPromptCatalogService(&promptCatalogRouteRepoStub{})),
	}, nil)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}
	require.True(t, registered[http.MethodGet+" /api/v1/prompts/cases"])

	req := httptest.NewRequest(http.MethodGet, "/api/v1/touch/prompts/cases", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestPromptCasesPublicRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &promptCatalogRouteRepoStub{
		items: []service.PromptCatalogCase{{
			ID:            "case-public",
			Title:         "Public Case",
			Prompt:        "prompt body",
			PromptPreview: "prompt body",
			Category:      "portrait",
			SourceProject: "Twitter Imports",
			SourceType:    "case",
			ImportSource:  "twitter",
			Status:        service.PromptCatalogStatusPublished,
		}},
		page: &pagination.PaginationResult{Total: 1, Page: 1, PageSize: 20, Pages: 1},
		summary: &service.PromptCatalogSummary{
			Total:          1,
			CaseCount:      1,
			SourceCount:    1,
			Sources:        []service.PromptCatalogFacetCount{{Value: "Twitter Imports", Count: 1}},
			TemplateGroups: []service.PromptCatalogFacetCount{{Value: "portrait", Count: 1}},
		},
	}
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterPromptRoutes(v1, &handler.Handlers{
		PromptCatalog: handler.NewPromptCatalogHandler(service.NewPromptCatalogService(repo)),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/prompts/cases?source_type=case&source_project=Twitter+Imports", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "case", repo.filters.SourceType)
	require.Equal(t, "Twitter Imports", repo.filters.SourceProject)

	var body struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Summary struct {
				Total          int64 `json:"total"`
				TemplateGroups []struct {
					Value string `json:"value"`
					Count int64  `json:"count"`
				} `json:"template_groups"`
			} `json:"summary"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Len(t, body.Data.Items, 1)
	require.Equal(t, "case-public", body.Data.Items[0].ID)
	require.Equal(t, int64(1), body.Data.Summary.Total)
	require.Len(t, body.Data.Summary.TemplateGroups, 1)
	require.Equal(t, "portrait", body.Data.Summary.TemplateGroups[0].Value)
}

func TestPromptCatalogAdminAliasIsNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &promptCatalogRouteRepoStub{}
	router := gin.New()
	v1 := router.Group("/api/v1")
	adminAuth := middleware.AdminAuthMiddleware(func(c *gin.Context) {
		c.Next()
	})
	RegisterPromptRoutes(v1, &handler.Handlers{
		PromptCatalog: handler.NewPromptCatalogHandler(service.NewPromptCatalogService(repo)),
	}, adminAuth)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}
	require.True(t, registered[http.MethodPost+" /api/v1/admin/prompts/cases"])

	req := httptest.NewRequest(http.MethodPost, "/api/v1/touch/admin/prompts/cases", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Nil(t, repo.upsert)
}

func TestPromptAdminUpsertRouteRequiresAdminAuthAndWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &promptCatalogRouteRepoStub{}
	router := gin.New()
	v1 := router.Group("/api/v1")
	adminAuth := middleware.AdminAuthMiddleware(func(c *gin.Context) {
		c.Next()
	})
	RegisterPromptRoutes(v1, &handler.Handlers{
		PromptCatalog: handler.NewPromptCatalogHandler(service.NewPromptCatalogService(repo)),
	}, adminAuth)

	body := []byte(`{
		"id": "case-admin",
		"title": "Admin Imported Case",
		"prompt": "prompt body",
		"image_urls": ["https://static.example/case-admin.jpg"],
		"model_tags": ["GPT Image 2"],
		"raw_json": {"source":"admin"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prompts/cases", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, repo.upsert)
	require.Equal(t, "case-admin", repo.upsert.ID)
	require.Equal(t, []string{"https://static.example/case-admin.jpg"}, repo.upsert.ImageURLs)
	require.Equal(t, []string{"GPT Image 2"}, repo.upsert.ModelTags)
	require.JSONEq(t, `{"source":"admin"}`, repo.upsert.RawJSON)
}
