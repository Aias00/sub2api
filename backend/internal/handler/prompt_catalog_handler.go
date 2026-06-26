package handler

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type PromptCatalogHandler struct {
	service *service.PromptCatalogService
}

func NewPromptCatalogHandler(service *service.PromptCatalogService) *PromptCatalogHandler {
	return &PromptCatalogHandler{service: service}
}

type promptCatalogCaseDTO struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	Prompt             string    `json:"prompt"`
	PromptPreview      string    `json:"prompt_preview"`
	Category           string    `json:"category"`
	Tags               []string  `json:"tags"`
	DisplayTags        []string  `json:"display_tags"`
	ModelTags          []string  `json:"model_tags"`
	AllTags            []string  `json:"all_tags"`
	VisibleTags        []string  `json:"visible_tags"`
	SourceURL          string    `json:"source_url,omitempty"`
	ImageURL           string    `json:"image_url,omitempty"`
	PrimaryImageURL    string    `json:"primary_image_url"`
	ImageURLs          []string  `json:"image_urls"`
	ImageOriginalURL   string    `json:"image_original_url,omitempty"`
	ImagePreviewURL    string    `json:"image_preview_url,omitempty"`
	ImageThumbURL      string    `json:"image_thumb_url,omitempty"`
	SourceProject      string    `json:"source_project"`
	SourceType         string    `json:"source_type"`
	SourceLabel        string    `json:"source_label,omitempty"`
	SourceDisplayLabel string    `json:"source_display_label"`
	GitHubURL          string    `json:"github_url,omitempty"`
	PromptCharCount    int       `json:"prompt_char_count"`
	Featured           bool      `json:"featured"`
	Styles             []string  `json:"styles"`
	Scenes             []string  `json:"scenes"`
	ImportSource       string    `json:"import_source"`
	Status             string    `json:"status"`
	ImportedAt         *string   `json:"imported_at"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type promptCatalogFacetCountDTO struct {
	Value        string `json:"value"`
	Label        string `json:"label,omitempty"`
	Count        int64  `json:"count"`
	DisplayLabel string `json:"display_label"`
}

type promptCatalogSummaryDTO struct {
	Total          int64                        `json:"total"`
	CaseCount      int64                        `json:"case_count"`
	TemplateCount  int64                        `json:"template_count"`
	SourceCount    int                          `json:"source_count"`
	CategoryCount  int                          `json:"category_count"`
	Sources        []promptCatalogFacetCountDTO `json:"sources"`
	Categories     []promptCatalogFacetCountDTO `json:"categories"`
	TemplateGroups []promptCatalogFacetCountDTO `json:"template_groups"`
}

type promptCatalogCaseListResponse struct {
	Items    []promptCatalogCaseDTO  `json:"items"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
	Pages    int                     `json:"pages"`
	Summary  promptCatalogSummaryDTO `json:"summary"`
}

type promptCatalogCaseUpsertRequest struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Prompt           string          `json:"prompt"`
	PromptPreview    string          `json:"prompt_preview"`
	Category         string          `json:"category"`
	Tags             []string        `json:"tags"`
	ModelTags        []string        `json:"model_tags"`
	SourceURL        string          `json:"source_url"`
	ImageURL         string          `json:"image_url"`
	ImageURLs        []string        `json:"image_urls"`
	ImageOriginalURL string          `json:"image_original_url"`
	ImagePreviewURL  string          `json:"image_preview_url"`
	ImageThumbURL    string          `json:"image_thumb_url"`
	SourceProject    string          `json:"source_project"`
	SourceType       string          `json:"source_type"`
	SourceLabel      string          `json:"source_label"`
	GitHubURL        string          `json:"github_url"`
	Featured         bool            `json:"featured"`
	Styles           []string        `json:"styles"`
	Scenes           []string        `json:"scenes"`
	ImportSource     string          `json:"import_source"`
	RawJSON          json.RawMessage `json:"raw_json"`
	Status           string          `json:"status"`
	ImportedAt       *time.Time      `json:"imported_at"`
}

func (h *PromptCatalogHandler) ListCases(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	filters := service.PromptCatalogListFilters{
		SourceType:    c.Query("source_type"),
		SourceProject: c.Query("source_project"),
		Category:      c.Query("category"),
		Search:        c.Query("search"),
	}
	if featuredRaw := c.Query("featured"); featuredRaw != "" {
		featured := featuredRaw == "true" || featuredRaw == "1"
		filters.Featured = &featured
	}
	if hasImageRaw := c.Query("has_image"); hasImageRaw != "" {
		hasImage := hasImageRaw == "true" || hasImageRaw == "1"
		filters.HasImage = &hasImage
	}

	items, pageResult, err := h.service.ListCases(c.Request.Context(), pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "imported_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	summary, err := h.service.GetCaseSummary(c.Request.Context(), filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, promptCatalogCaseListResponse{
		Items:    promptCatalogCaseDTOs(items),
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
		Pages:    pageResult.Pages,
		Summary:  promptCatalogSummaryDTOFromService(summary),
	})
}

func (h *PromptCatalogHandler) GetCase(c *gin.Context) {
	item, err := h.service.GetCaseByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, promptCatalogCaseDTOFromService(*item))
}

func (h *PromptCatalogHandler) UpsertCase(c *gin.Context) {
	var req promptCatalogCaseUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	item := promptCatalogCaseFromUpsertRequest(req)
	if err := h.service.UpsertCase(c.Request.Context(), &item); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, promptCatalogCaseDTOFromService(item))
}

func promptCatalogCaseDTOs(items []service.PromptCatalogCase) []promptCatalogCaseDTO {
	out := make([]promptCatalogCaseDTO, 0, len(items))
	for _, item := range items {
		out = append(out, promptCatalogCaseDTOFromService(item))
	}
	return out
}

func promptCatalogSummaryDTOFromService(summary *service.PromptCatalogSummary) promptCatalogSummaryDTO {
	if summary == nil {
		return promptCatalogSummaryDTO{
			Sources:        []promptCatalogFacetCountDTO{},
			Categories:     []promptCatalogFacetCountDTO{},
			TemplateGroups: []promptCatalogFacetCountDTO{},
		}
	}
	return promptCatalogSummaryDTO{
		Total:          summary.Total,
		CaseCount:      summary.CaseCount,
		TemplateCount:  summary.TemplateCount,
		SourceCount:    summary.SourceCount,
		CategoryCount:  summary.CategoryCount,
		Sources:        promptCatalogFacetCountDTOs(summary.Sources),
		Categories:     promptCatalogFacetCountDTOs(summary.Categories),
		TemplateGroups: promptCatalogFacetCountDTOs(summary.TemplateGroups),
	}
}

func promptCatalogFacetCountDTOs(items []service.PromptCatalogFacetCount) []promptCatalogFacetCountDTO {
	out := make([]promptCatalogFacetCountDTO, 0, len(items))
	for _, item := range items {
		out = append(out, promptCatalogFacetCountDTO{
			Value:        item.Value,
			Label:        item.Label,
			Count:        item.Count,
			DisplayLabel: promptCatalogFacetDisplayLabel(item),
		})
	}
	return out
}

func promptCatalogFacetDisplayLabel(item service.PromptCatalogFacetCount) string {
	label := strings.TrimSpace(item.Label)
	if label == "" {
		label = strings.TrimSpace(item.Value)
	}
	if label == "" {
		return ""
	}
	return label + " (" + strconv.FormatInt(item.Count, 10) + ")"
}

func promptCatalogCaseDTOFromService(item service.PromptCatalogCase) promptCatalogCaseDTO {
	var importedAt *string
	if item.ImportedAt != nil {
		value := item.ImportedAt.UTC().Format(time.RFC3339)
		importedAt = &value
	}
	return promptCatalogCaseDTO{
		ID:                 item.ID,
		Title:              item.Title,
		Prompt:             item.Prompt,
		PromptPreview:      item.PromptPreview,
		Category:           item.Category,
		Tags:               nonNilStrings(item.Tags),
		DisplayTags:        nonNilStrings(item.DisplayTags),
		ModelTags:          nonNilStrings(item.ModelTags),
		AllTags:            promptCatalogAllTags(item),
		VisibleTags:        promptCatalogVisibleTags(item),
		SourceURL:          item.SourceURL,
		ImageURL:           item.ImageURL,
		PrimaryImageURL:    promptCatalogPrimaryImageURL(item),
		ImageURLs:          nonNilStrings(item.ImageURLs),
		ImageOriginalURL:   item.ImageOriginalURL,
		ImagePreviewURL:    item.ImagePreviewURL,
		ImageThumbURL:      item.ImageThumbURL,
		SourceProject:      item.SourceProject,
		SourceType:         item.SourceType,
		SourceLabel:        item.SourceLabel,
		SourceDisplayLabel: promptCatalogSourceDisplayLabel(item),
		GitHubURL:          item.GitHubURL,
		PromptCharCount:    utf8.RuneCountInString(item.Prompt),
		Featured:           item.Featured,
		Styles:             nonNilStrings(item.Styles),
		Scenes:             nonNilStrings(item.Scenes),
		ImportSource:       item.ImportSource,
		Status:             item.Status,
		ImportedAt:         importedAt,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func promptCatalogSourceDisplayLabel(item service.PromptCatalogCase) string {
	if label := strings.TrimSpace(item.SourceLabel); label != "" {
		return label
	}
	return strings.TrimSpace(item.SourceProject)
}

func promptCatalogVisibleTags(item service.PromptCatalogCase) []string {
	tags := promptCatalogAllTags(item)
	if len(tags) > 6 {
		return tags[:6]
	}
	return tags
}

func promptCatalogAllTags(item service.PromptCatalogCase) []string {
	values := make([]string, 0, len(item.ModelTags)+len(item.DisplayTags)+len(item.Tags))
	values = append(values, item.ModelTags...)
	values = append(values, item.DisplayTags...)
	values = append(values, item.Tags...)
	out := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func promptCatalogPrimaryImageURL(item service.PromptCatalogCase) string {
	for _, candidate := range []string{
		item.ImageThumbURL,
		item.ImagePreviewURL,
		item.ImageURL,
	} {
		if candidate != "" {
			return candidate
		}
	}
	for _, candidate := range item.ImageURLs {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func promptCatalogCaseFromUpsertRequest(req promptCatalogCaseUpsertRequest) service.PromptCatalogCase {
	rawJSON := string(req.RawJSON)
	if rawJSON == "" {
		rawJSON = "{}"
	}
	return service.PromptCatalogCase{
		ID:               req.ID,
		Title:            req.Title,
		Prompt:           req.Prompt,
		PromptPreview:    req.PromptPreview,
		Category:         req.Category,
		Tags:             nonNilStrings(req.Tags),
		ModelTags:        nonNilStrings(req.ModelTags),
		SourceURL:        req.SourceURL,
		ImageURL:         req.ImageURL,
		ImageURLs:        nonNilStrings(req.ImageURLs),
		ImageOriginalURL: req.ImageOriginalURL,
		ImagePreviewURL:  req.ImagePreviewURL,
		ImageThumbURL:    req.ImageThumbURL,
		SourceProject:    req.SourceProject,
		SourceType:       req.SourceType,
		SourceLabel:      req.SourceLabel,
		GitHubURL:        req.GitHubURL,
		Featured:         req.Featured,
		Styles:           nonNilStrings(req.Styles),
		Scenes:           nonNilStrings(req.Scenes),
		ImportSource:     req.ImportSource,
		RawJSON:          rawJSON,
		Status:           req.Status,
		ImportedAt:       req.ImportedAt,
	}
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
