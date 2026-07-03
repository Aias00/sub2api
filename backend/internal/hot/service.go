package hot

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
)

var ErrHotContentInvalidInput = infraerrors.BadRequest("HOT_CONTENT_INVALID_INPUT", "hot content input is invalid")

type Source struct {
	ID           int64     `json:"id"`
	SourceID     string    `json:"source_id"`
	AdapterKind  string    `json:"adapter_kind"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Enabled      bool      `json:"enabled"`
	BaseURL      string    `json:"base_url"`
	SeedURLsJSON string    `json:"seed_urls_json"`
	ConfigJSON   string    `json:"config_json"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Item struct {
	ID           int64      `json:"id"`
	SourceID     string     `json:"source_id"`
	ExternalID   string     `json:"external_id"`
	CanonicalURL string     `json:"canonical_url"`
	Title        string     `json:"title"`
	Summary      string     `json:"summary"`
	Body         string     `json:"body"`
	Quoted       string     `json:"quoted"`
	Reason       string     `json:"reason"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	Author       string     `json:"author"`
	SourceName   string     `json:"source_name"`
	SourceHandle string     `json:"source_handle"`
	Badge        string     `json:"badge"`
	Score        string     `json:"score"`
	ContentType  string     `json:"content_type"`
	TagsJSON     string     `json:"tags_json"`
	MetricsJSON  string     `json:"metrics_json"`
	RawRefJSON   string     `json:"raw_ref_json"`
	ContentHash  string     `json:"content_hash"`
	HasMedia     bool       `json:"has_media"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RunEvent struct {
	ID          int64     `json:"id"`
	LegacyID    int64     `json:"legacy_id"`
	RunID       string    `json:"run_id"`
	Node        string    `json:"node"`
	Message     string    `json:"message"`
	PayloadJSON string    `json:"payload_json"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListFilters struct {
	SourceID string
	Query    string
	Status   string
}

type Repository interface {
	ListSources(ctx context.Context) ([]Source, error)
	ListItems(ctx context.Context, params pagination.PaginationParams, filters ListFilters) ([]Item, *pagination.PaginationResult, error)
	ListRunEvents(ctx context.Context, runID string, params pagination.PaginationParams) ([]RunEvent, *pagination.PaginationResult, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Health() error {
	if s == nil || s.repo == nil {
		return ErrHotContentInvalidInput
	}
	return nil
}

func (s *Service) ListSources(ctx context.Context) ([]Source, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	return s.repo.ListSources(ctx)
}

func (s *Service) ListItems(ctx context.Context, params pagination.PaginationParams, filters ListFilters) ([]Item, *pagination.PaginationResult, error) {
	if err := s.Health(); err != nil {
		return nil, nil, err
	}
	params = normalizeHotContentPagination(params)
	filters.SourceID = strings.TrimSpace(filters.SourceID)
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Status = strings.TrimSpace(filters.Status)
	if filters.Status == "" {
		filters.Status = "published"
	}
	return s.repo.ListItems(ctx, params, filters)
}

func (s *Service) ListRunEvents(ctx context.Context, runID string, params pagination.PaginationParams) ([]RunEvent, *pagination.PaginationResult, error) {
	if err := s.Health(); err != nil {
		return nil, nil, err
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, nil, ErrHotContentInvalidInput
	}
	params = normalizeHotContentPagination(params)
	return s.repo.ListRunEvents(ctx, runID, params)
}

func normalizeHotContentPagination(params pagination.PaginationParams) pagination.PaginationParams {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	return params
}
