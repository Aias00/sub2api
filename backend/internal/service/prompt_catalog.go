package service

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	PromptCatalogStatusPublished = "published"
	PromptCatalogStatusDraft     = "draft"
)

var promptCatalogValidStatuses = map[string]bool{
	PromptCatalogStatusPublished: true,
	PromptCatalogStatusDraft:     true,
}

var promptCatalogValidSourceTypes = map[string]bool{
	"case":     true,
	"template": true,
}

const twitterSnowflakeEpochMs int64 = 1288834974657

var ErrPromptCatalogNotFound = infraerrors.NotFound("PROMPT_CATALOG_NOT_FOUND", "prompt catalog case not found")

var (
	ErrPromptCatalogInvalidInput  = infraerrors.BadRequest("PROMPT_CATALOG_INVALID_INPUT", "prompt catalog case input is invalid")
	ErrPromptCatalogIDRequired    = infraerrors.BadRequest("PROMPT_CATALOG_ID_REQUIRED", "prompt catalog case id is required")
	ErrPromptCatalogTitleRequired = infraerrors.BadRequest("PROMPT_CATALOG_TITLE_REQUIRED", "prompt catalog case title is required")
	ErrPromptCatalogBodyRequired  = infraerrors.BadRequest("PROMPT_CATALOG_BODY_REQUIRED", "prompt catalog case prompt is required")
)

type PromptCatalogCase struct {
	ID               string
	Title            string
	Prompt           string
	PromptPreview    string
	Category         string
	Tags             []string
	DisplayTags      []string
	ModelTags        []string
	SourceURL        string
	ImageURL         string
	ImageURLs        []string
	ImageOriginalURL string
	ImagePreviewURL  string
	ImageThumbURL    string
	SourceProject    string
	SourceType       string
	SourceLabel      string
	GitHubURL        string
	Featured         bool
	Styles           []string
	Scenes           []string
	ImportSource     string
	RawJSON          string
	Status           string
	ImportedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PromptCatalogListFilters struct {
	SourceType    string
	SourceProject string
	Category      string
	Search        string
	Featured      *bool
	HasImage      *bool
}

type PromptCatalogFacetCount struct {
	Value string
	Label string
	Count int64
}

type PromptCatalogSummary struct {
	Total          int64
	CaseCount      int64
	TemplateCount  int64
	SourceCount    int
	CategoryCount  int
	Sources        []PromptCatalogFacetCount
	Categories     []PromptCatalogFacetCount
	TemplateGroups []PromptCatalogFacetCount
}

type PromptCatalogRepository interface {
	ListCases(ctx context.Context, params pagination.PaginationParams, filters PromptCatalogListFilters) ([]PromptCatalogCase, *pagination.PaginationResult, error)
	GetCaseSummary(ctx context.Context, filters PromptCatalogListFilters) (*PromptCatalogSummary, error)
	GetCaseByID(ctx context.Context, id string) (*PromptCatalogCase, error)
	UpsertCase(ctx context.Context, item *PromptCatalogCase) error
}

type PromptCatalogService struct {
	repo PromptCatalogRepository
}

func NewPromptCatalogService(repo PromptCatalogRepository) *PromptCatalogService {
	return &PromptCatalogService{repo: repo}
}

func (s *PromptCatalogService) ListCases(ctx context.Context, params pagination.PaginationParams, filters PromptCatalogListFilters) ([]PromptCatalogCase, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, infraerrors.InternalServer("PROMPT_CATALOG_REPOSITORY_MISSING", "prompt catalog repository is not configured")
	}
	params = normalizePromptCatalogPagination(params)
	filters = normalizePromptCatalogFilters(filters)
	items, pageResult, err := s.repo.ListCases(ctx, params, filters)
	if err != nil {
		return nil, nil, err
	}
	for idx := range items {
		normalizePromptCatalogCase(&items[idx])
	}
	return items, pageResult, nil
}

func (s *PromptCatalogService) GetCaseSummary(ctx context.Context, filters PromptCatalogListFilters) (*PromptCatalogSummary, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.InternalServer("PROMPT_CATALOG_REPOSITORY_MISSING", "prompt catalog repository is not configured")
	}
	return s.repo.GetCaseSummary(ctx, normalizePromptCatalogFilters(filters))
}

func (s *PromptCatalogService) GetCaseByID(ctx context.Context, id string) (*PromptCatalogCase, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.InternalServer("PROMPT_CATALOG_REPOSITORY_MISSING", "prompt catalog repository is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrPromptCatalogNotFound
	}
	item, err := s.repo.GetCaseByID(ctx, id)
	if err != nil {
		return nil, err
	}
	normalizePromptCatalogCase(item)
	return item, nil
}

func (s *PromptCatalogService) UpsertCase(ctx context.Context, item *PromptCatalogCase) error {
	if s == nil || s.repo == nil {
		return infraerrors.InternalServer("PROMPT_CATALOG_REPOSITORY_MISSING", "prompt catalog repository is not configured")
	}
	if item == nil {
		return ErrPromptCatalogInvalidInput
	}
	normalizePromptCatalogCase(item)
	if item.ID == "" {
		return ErrPromptCatalogIDRequired
	}
	if item.Title == "" {
		return ErrPromptCatalogTitleRequired
	}
	if item.Prompt == "" {
		return ErrPromptCatalogBodyRequired
	}
	if item.Status != "" && !promptCatalogValidStatuses[item.Status] {
		return ErrPromptCatalogInvalidInput
	}
	if item.SourceType != "" && !promptCatalogValidSourceTypes[item.SourceType] {
		return ErrPromptCatalogInvalidInput
	}
	return s.repo.UpsertCase(ctx, item)
}

func normalizePromptCatalogPagination(params pagination.PaginationParams) pagination.PaginationParams {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	params.SortBy = normalizePromptCatalogSortBy(params.SortBy)
	params.SortOrder = pagination.NormalizeSortOrder(params.SortOrder, pagination.SortOrderDesc)
	return params
}

func normalizePromptCatalogSortBy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "title":
		return "title"
	case "created_at":
		return "created_at"
	case "updated_at":
		return "updated_at"
	default:
		return "imported_at"
	}
}

func normalizePromptCatalogFilters(filters PromptCatalogListFilters) PromptCatalogListFilters {
	filters.SourceType = strings.TrimSpace(filters.SourceType)
	filters.SourceProject = strings.TrimSpace(filters.SourceProject)
	filters.Category = strings.TrimSpace(filters.Category)
	filters.Search = strings.TrimSpace(filters.Search)
	return filters
}

func normalizePromptCatalogCase(item *PromptCatalogCase) {
	item.ID = strings.TrimSpace(item.ID)
	item.Title = strings.TrimSpace(item.Title)
	item.Prompt = strings.TrimSpace(item.Prompt)
	item.PromptPreview = strings.TrimSpace(item.PromptPreview)
	if item.PromptPreview == "" {
		item.PromptPreview = item.Prompt
	}
	item.Category = strings.TrimSpace(item.Category)
	if item.Category == "" {
		item.Category = "general"
	}
	item.SourceProject = strings.TrimSpace(item.SourceProject)
	if item.SourceProject == "" {
		item.SourceProject = "manual"
	}
	item.SourceType = strings.TrimSpace(item.SourceType)
	if item.SourceType == "" {
		item.SourceType = "case"
	}
	item.ImportSource = strings.TrimSpace(item.ImportSource)
	if item.ImportSource == "" {
		item.ImportSource = "catalog"
	}
	item.Status = strings.TrimSpace(item.Status)
	if item.Status == "" {
		item.Status = PromptCatalogStatusPublished
	}
	if item.RawJSON == "" {
		item.RawJSON = "{}"
	}
	item.Tags = normalizePromptCatalogStringSlice(item.Tags)
	item.DisplayTags = normalizePromptCatalogStringSlice(item.DisplayTags)
	item.ModelTags = normalizePromptCatalogStringSlice(item.ModelTags)
	if len(item.ModelTags) == 0 {
		item.ModelTags = inferPromptCatalogModelTags(*item)
	}
	item.ImageURLs = normalizePromptCatalogStringSlice(item.ImageURLs)
	item.Styles = normalizePromptCatalogStringSlice(item.Styles)
	item.Scenes = normalizePromptCatalogStringSlice(item.Scenes)
	if len(item.DisplayTags) == 0 {
		item.DisplayTags = derivePromptCatalogDisplayTags(*item)
	}
	if item.ImportedAt == nil {
		item.ImportedAt = inferPromptCatalogImportedAt(*item)
	}
}

func normalizePromptCatalogStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

var promptCatalogTokenSeparators = regexp.MustCompile(`[^a-z0-9]+`)

func normalizePromptCatalogToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = promptCatalogTokenSeparators.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func promptCatalogMetadataIncludesAny(metadata string, needles ...string) bool {
	normalized := normalizePromptCatalogToken(metadata)
	for _, needle := range needles {
		if strings.Contains(normalized, needle) {
			return true
		}
	}
	return false
}

func inferPromptCatalogModelTags(item PromptCatalogCase) []string {
	metadataParts := []string{
		item.SourceProject,
		item.SourceLabel,
		item.Title,
		item.Category,
	}
	metadataParts = append(metadataParts, item.Tags...)
	metadataParts = append(metadataParts, item.Styles...)
	metadataParts = append(metadataParts, item.Scenes...)
	metadata := strings.Join(metadataParts, " ")
	sourceProject := normalizePromptCatalogToken(item.SourceProject)
	tags := make([]string, 0, 4)
	seen := map[string]bool{}
	addTag := func(tag string) {
		if tag == "" || seen[tag] {
			return
		}
		seen[tag] = true
		tags = append(tags, tag)
	}

	if sourceProject == "awesome-nano-banana-pro-prompts" ||
		promptCatalogMetadataIncludesAny(metadata, "nano-banana", "nanobanana") {
		addTag("nano-banana-pro")
	}

	if sourceProject == "gpt4o-image-prompts" ||
		promptCatalogMetadataIncludesAny(metadata, "gpt4o-image", "gpt-4o-image") {
		addTag("openai-image")
		addTag("gpt-4o-image")
	}

	if sourceProject == "awesome-gpt-image-2" ||
		sourceProject == "awesome-gpt-image-2-api-and-prompts" ||
		sourceProject == "gpt-image2-skill" ||
		promptCatalogMetadataIncludesAny(metadata, "gpt-image-2", "gpt-image2") {
		addTag("openai-image")
		addTag("gpt-image-2")
	}

	if promptCatalogMetadataIncludesAny(metadata, "grok") {
		addTag("grok")
	}

	return tags
}

func derivePromptCatalogDisplayTags(item PromptCatalogCase) []string {
	values := make([]string, 0, len(item.Tags)+len(item.Styles)+len(item.Scenes))
	values = append(values, item.Tags...)
	values = append(values, item.Styles...)
	values = append(values, item.Scenes...)

	category := normalizePromptCatalogComparable(item.Category)
	sourceProject := normalizePromptCatalogComparable(item.SourceProject)
	sourceLabel := normalizePromptCatalogComparable(item.SourceLabel)
	modelTags := map[string]bool{}
	for _, tag := range item.ModelTags {
		modelTags[normalizePromptCatalogComparable(tag)] = true
	}

	seen := map[string]bool{}
	displayTags := make([]string, 0, len(values))
	for _, value := range values {
		tag := strings.TrimSpace(value)
		if tag == "" {
			continue
		}
		comparable := normalizePromptCatalogComparable(tag)
		if comparable == "" ||
			comparable == "ui" ||
			comparable == category ||
			comparable == sourceProject ||
			comparable == sourceLabel ||
			modelTags[comparable] ||
			strings.HasPrefix(strings.TrimSpace(tag), "@") ||
			promptCatalogLooksLikeSourceSlug(tag) {
			continue
		}
		seenKey := strings.ToLower(tag)
		if seen[seenKey] {
			continue
		}
		seen[seenKey] = true
		displayTags = append(displayTags, tag)
	}
	return displayTags
}

func normalizePromptCatalogComparable(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			_, _ = builder.WriteRune(r)
		}
	}
	return builder.String()
}

func promptCatalogLooksLikeSourceSlug(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, " ") || !strings.Contains(value, "_") {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func inferPromptCatalogImportedAt(item PromptCatalogCase) *time.Time {
	tweetID := extractTwitterStatusID(item.SourceURL)
	if tweetID == "" && strings.HasPrefix(item.ID, "tw-") {
		tweetID = strings.TrimPrefix(item.ID, "tw-")
	}
	if tweetID == "" {
		return nil
	}

	var snowflake uint64
	for _, r := range tweetID {
		if r < '0' || r > '9' {
			return nil
		}
		snowflake = snowflake*10 + uint64(r-'0')
	}
	if snowflake == 0 {
		return nil
	}

	timestampMs := int64(snowflake>>22) + twitterSnowflakeEpochMs
	if timestampMs <= 0 {
		return nil
	}
	importedAt := time.UnixMilli(timestampMs).UTC()
	return &importedAt
}

func extractTwitterStatusID(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	hostname := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
	if hostname != "x.com" && hostname != "twitter.com" && hostname != "mobile.twitter.com" {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for idx, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if (part == "status" || part == "statuses") && idx+1 < len(parts) {
			return strings.TrimSpace(parts[idx+1])
		}
	}
	return ""
}
