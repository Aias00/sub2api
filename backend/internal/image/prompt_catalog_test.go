package image

import (
	"context"
	"testing"
	"time"

	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type promptCatalogRepoStub struct {
	listParams  pagination.PaginationParams
	listFilters PromptCatalogListFilters
	listItems   []PromptCatalogCase
	listPage    *pagination.PaginationResult
	listErr     error
	summary     *PromptCatalogSummary
	summaryErr  error
	getItem     *PromptCatalogCase
	getErr      error
	getID       string
	upsertItem  *PromptCatalogCase
	upsertErr   error
}

func (s *promptCatalogRepoStub) ListCases(ctx context.Context, params pagination.PaginationParams, filters PromptCatalogListFilters) ([]PromptCatalogCase, *pagination.PaginationResult, error) {
	s.listParams = params
	s.listFilters = filters
	if s.listPage == nil {
		s.listPage = &pagination.PaginationResult{Total: int64(len(s.listItems)), Page: params.Page, PageSize: params.PageSize, Pages: 1}
	}
	return s.listItems, s.listPage, s.listErr
}

func (s *promptCatalogRepoStub) GetCaseSummary(ctx context.Context, filters PromptCatalogListFilters) (*PromptCatalogSummary, error) {
	s.listFilters = filters
	if s.summary != nil || s.summaryErr != nil {
		return s.summary, s.summaryErr
	}
	return &PromptCatalogSummary{}, nil
}

func (s *promptCatalogRepoStub) GetCaseByID(ctx context.Context, id string) (*PromptCatalogCase, error) {
	s.getID = id
	return s.getItem, s.getErr
}

func (s *promptCatalogRepoStub) UpsertCase(ctx context.Context, item *PromptCatalogCase) error {
	s.upsertItem = item
	return s.upsertErr
}

func TestPromptCatalogServiceListCasesNormalizesInput(t *testing.T) {
	repo := &promptCatalogRepoStub{}
	svc := NewPromptCatalogService(repo)

	_, _, err := svc.ListCases(context.Background(), pagination.PaginationParams{Page: -1, PageSize: 500}, PromptCatalogListFilters{
		SourceType:    " case ",
		SourceProject: " Twitter Imports ",
		Category:      " portrait ",
		Search:        " toy ",
	})

	require.NoError(t, err)
	require.Equal(t, 1, repo.listParams.Page)
	require.Equal(t, 100, repo.listParams.PageSize)
	require.Equal(t, "imported_at", repo.listParams.SortBy)
	require.Equal(t, "desc", repo.listParams.SortOrder)
	require.Equal(t, "case", repo.listFilters.SourceType)
	require.Equal(t, "Twitter Imports", repo.listFilters.SourceProject)
	require.Equal(t, "portrait", repo.listFilters.Category)
	require.Equal(t, "toy", repo.listFilters.Search)
}

func TestPromptCatalogServiceListCasesNormalizesSort(t *testing.T) {
	repo := &promptCatalogRepoStub{}
	svc := NewPromptCatalogService(repo)

	_, _, err := svc.ListCases(context.Background(), pagination.PaginationParams{
		Page:      1,
		PageSize:  20,
		SortBy:    " title ",
		SortOrder: " ASC ",
	}, PromptCatalogListFilters{})
	require.NoError(t, err)
	require.Equal(t, "title", repo.listParams.SortBy)
	require.Equal(t, "asc", repo.listParams.SortOrder)

	_, _, err = svc.ListCases(context.Background(), pagination.PaginationParams{
		Page:      1,
		PageSize:  20,
		SortBy:    "raw_sql",
		SortOrder: "sideways",
	}, PromptCatalogListFilters{})
	require.NoError(t, err)
	require.Equal(t, "imported_at", repo.listParams.SortBy)
	require.Equal(t, "desc", repo.listParams.SortOrder)
}

func TestPromptCatalogServiceListCasesInfersModelTags(t *testing.T) {
	repo := &promptCatalogRepoStub{
		listItems: []PromptCatalogCase{
			{
				ID:            "case-1",
				Title:         "Product box",
				Prompt:        "prompt body",
				Category:      "Product",
				Tags:          []string{"gpt-image-2-skill"},
				SourceProject: "GPT-Image2-Skill",
			},
			{
				ID:            "case-2",
				Title:         "Custom tagged",
				Prompt:        "prompt body",
				Category:      "Portrait",
				ModelTags:     []string{"custom-model"},
				SourceProject: "awesome-nano-banana-pro-prompts",
			},
		},
	}
	svc := NewPromptCatalogService(repo)

	items, _, err := svc.ListCases(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 20}, PromptCatalogListFilters{})

	require.NoError(t, err)
	require.Equal(t, []string{"openai-image", "gpt-image-2"}, items[0].ModelTags)
	require.Equal(t, []string{"custom-model"}, items[1].ModelTags)
}

func TestPromptCatalogServiceListCasesInfersImportedAt(t *testing.T) {
	existingImportedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	repo := &promptCatalogRepoStub{
		listItems: []PromptCatalogCase{
			{
				ID:        "tw-2047884126343032995",
				Title:     "Twitter prompt",
				Prompt:    "prompt body",
				SourceURL: "https://x.com/hmontilla_/status/2047884126343032995",
			},
			{
				ID:         "tw-2047884126343032995",
				Title:      "Existing imported at",
				Prompt:     "prompt body",
				ImportedAt: &existingImportedAt,
			},
		},
	}
	svc := NewPromptCatalogService(repo)

	items, _, err := svc.ListCases(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 20}, PromptCatalogListFilters{})

	require.NoError(t, err)
	require.NotNil(t, items[0].ImportedAt)
	require.Equal(t, "2026-04-25T03:43:18Z", items[0].ImportedAt.Format(time.RFC3339))
	require.Equal(t, existingImportedAt, *items[1].ImportedAt)
}

func TestPromptCatalogServiceListCasesDerivesDisplayTags(t *testing.T) {
	repo := &promptCatalogRepoStub{
		listItems: []PromptCatalogCase{{
			ID:            "case-1",
			Title:         "Display tags",
			Prompt:        "prompt body",
			Category:      "portrait",
			Tags:          []string{"portrait", "Twitter Imports", "@author", "ui", "soft light", "openai-image", "source_project_slug", "soft light"},
			Styles:        []string{"editorial"},
			Scenes:        []string{"studio"},
			ModelTags:     []string{"openai-image"},
			SourceProject: "Twitter Imports",
			SourceLabel:   "Twitter Imports",
		}},
	}
	svc := NewPromptCatalogService(repo)

	items, _, err := svc.ListCases(context.Background(), pagination.PaginationParams{Page: 1, PageSize: 20}, PromptCatalogListFilters{})

	require.NoError(t, err)
	require.Equal(t, []string{"soft light", "editorial", "studio"}, items[0].DisplayTags)
}

func TestPromptCatalogServiceGetCaseSummaryNormalizesFilters(t *testing.T) {
	repo := &promptCatalogRepoStub{
		summary: &PromptCatalogSummary{Total: 2},
	}
	svc := NewPromptCatalogService(repo)

	summary, err := svc.GetCaseSummary(context.Background(), PromptCatalogListFilters{
		SourceType:    " case ",
		SourceProject: " Twitter Imports ",
		Category:      " portrait ",
		Search:        " toy ",
	})

	require.NoError(t, err)
	require.Equal(t, int64(2), summary.Total)
	require.Equal(t, "case", repo.listFilters.SourceType)
	require.Equal(t, "Twitter Imports", repo.listFilters.SourceProject)
	require.Equal(t, "portrait", repo.listFilters.Category)
	require.Equal(t, "toy", repo.listFilters.Search)
}

func TestPromptCatalogServiceGetCaseByIDTrimsAndRejectsEmpty(t *testing.T) {
	repo := &promptCatalogRepoStub{getItem: &PromptCatalogCase{ID: "case-1"}}
	svc := NewPromptCatalogService(repo)

	item, err := svc.GetCaseByID(context.Background(), " case-1 ")
	require.NoError(t, err)
	require.Equal(t, "case-1", item.ID)
	require.Equal(t, "case-1", repo.getID)

	_, err = svc.GetCaseByID(context.Background(), " ")
	require.ErrorIs(t, err, ErrPromptCatalogNotFound)
}

func TestPromptCatalogServiceUpsertCaseNormalizesAndValidates(t *testing.T) {
	repo := &promptCatalogRepoStub{}
	svc := NewPromptCatalogService(repo)

	item := &PromptCatalogCase{
		ID:         " case-1 ",
		Title:      " Toy Portrait ",
		Prompt:     " prompt body ",
		Tags:       []string{" twitter ", ""},
		ModelTags:  []string{" OpenAI Image "},
		ImageURLs:  nil,
		RawJSON:    "",
		SourceType: "",
	}
	err := svc.UpsertCase(context.Background(), item)

	require.NoError(t, err)
	require.Equal(t, "case-1", repo.upsertItem.ID)
	require.Equal(t, "Toy Portrait", repo.upsertItem.Title)
	require.Equal(t, "prompt body", repo.upsertItem.Prompt)
	require.Equal(t, "prompt body", repo.upsertItem.PromptPreview)
	require.Equal(t, "general", repo.upsertItem.Category)
	require.Equal(t, "case", repo.upsertItem.SourceType)
	require.Equal(t, "catalog", repo.upsertItem.ImportSource)
	require.Equal(t, PromptCatalogStatusPublished, repo.upsertItem.Status)
	require.Equal(t, "{}", repo.upsertItem.RawJSON)
	require.Equal(t, []string{"twitter"}, repo.upsertItem.Tags)
	require.Equal(t, []string{}, repo.upsertItem.ImageURLs)

	err = svc.UpsertCase(context.Background(), &PromptCatalogCase{Title: "x", Prompt: "y"})
	require.ErrorIs(t, err, ErrPromptCatalogIDRequired)
}

func TestPromptCatalogServiceUpsertCaseInfersModelTags(t *testing.T) {
	repo := &promptCatalogRepoStub{}
	svc := NewPromptCatalogService(repo)

	err := svc.UpsertCase(context.Background(), &PromptCatalogCase{
		ID:            "case-1",
		Title:         "3D character",
		Prompt:        "prompt body",
		Category:      "Nano Banana Pro",
		Tags:          []string{"nano-banana-pro"},
		SourceProject: "awesome-nano-banana-pro-prompts",
	})

	require.NoError(t, err)
	require.Equal(t, []string{"nano-banana-pro"}, repo.upsertItem.ModelTags)
}

func TestPromptCatalogServiceUpsertCaseRejectsInvalidStatus(t *testing.T) {
	repo := &promptCatalogRepoStub{}
	svc := NewPromptCatalogService(repo)

	err := svc.UpsertCase(context.Background(), &PromptCatalogCase{
		ID:     "case-1",
		Title:  "Title",
		Prompt: "Body",
		Status: "archived",
	})
	require.ErrorIs(t, err, ErrPromptCatalogInvalidInput)
}

func TestPromptCatalogServiceUpsertCaseRejectsInvalidSourceType(t *testing.T) {
	repo := &promptCatalogRepoStub{}
	svc := NewPromptCatalogService(repo)

	err := svc.UpsertCase(context.Background(), &PromptCatalogCase{
		ID:         "case-1",
		Title:      "Title",
		Prompt:     "Body",
		SourceType: "unknown",
	})
	require.ErrorIs(t, err, ErrPromptCatalogInvalidInput)
}

func TestPromptCatalogServiceUpsertCaseAcceptsDraftStatus(t *testing.T) {
	repo := &promptCatalogRepoStub{}
	svc := NewPromptCatalogService(repo)

	err := svc.UpsertCase(context.Background(), &PromptCatalogCase{
		ID:     "case-1",
		Title:  "Title",
		Prompt: "Body",
		Status: PromptCatalogStatusDraft,
	})
	require.NoError(t, err)
	require.Equal(t, PromptCatalogStatusDraft, repo.upsertItem.Status)
}

func TestPromptCatalogServiceUpsertCaseAcceptsTemplateSourceType(t *testing.T) {
	repo := &promptCatalogRepoStub{}
	svc := NewPromptCatalogService(repo)

	err := svc.UpsertCase(context.Background(), &PromptCatalogCase{
		ID:         "case-1",
		Title:      "Title",
		Prompt:     "Body",
		SourceType: "template",
	})
	require.NoError(t, err)
	require.Equal(t, "template", repo.upsertItem.SourceType)
}
