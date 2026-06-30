package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type promptCatalogRepository struct {
	db *sql.DB
}

const promptCatalogItemsTable = "prompt_catalog_items"

func NewPromptCatalogRepository(db *sql.DB) service.PromptCatalogRepository {
	return &promptCatalogRepository{db: db}
}

func (r *promptCatalogRepository) ListCases(
	ctx context.Context,
	params pagination.PaginationParams,
	filters service.PromptCatalogListFilters,
) ([]service.PromptCatalogCase, *pagination.PaginationResult, error) {
	whereSQL, args := buildPromptCatalogWhere(filters)

	var total int64
	countSQL := "SELECT COUNT(*) FROM " + promptCatalogItemsTable + whereSQL
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, nil, err
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, params.Limit(), params.Offset())
	listSQL := promptCatalogSelectSQL() + whereSQL + `
ORDER BY ` + promptCatalogOrderBy(params) + `
LIMIT $` + fmt.Sprint(len(queryArgs)-1) + ` OFFSET $` + fmt.Sprint(len(queryArgs))

	rows, err := r.db.QueryContext(ctx, listSQL, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	items := make([]service.PromptCatalogCase, 0, params.Limit())
	for rows.Next() {
		item, err := scanPromptCatalogCase(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return items, paginationResultFromTotal(total, params), nil
}

func (r *promptCatalogRepository) GetCaseSummary(ctx context.Context, filters service.PromptCatalogListFilters) (*service.PromptCatalogSummary, error) {
	whereSQL, args := buildPromptCatalogWhere(filters)
	rows, err := r.db.QueryContext(ctx, `
SELECT
	source_type,
	source_project,
	category,
	COALESCE(MAX(NULLIF(source_label, '')), '') AS source_label,
	COUNT(*) AS item_count
FROM `+promptCatalogItemsTable+whereSQL+`
GROUP BY source_type, source_project, category`, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	summary := &service.PromptCatalogSummary{}
	sourceCounts := map[string]int64{}
	sourceLabels := map[string]string{}
	categoryCounts := map[string]int64{}
	templateGroupCounts := map[string]int64{}

	for rows.Next() {
		var sourceType, sourceProject, category, sourceLabel string
		var count int64
		if err := rows.Scan(&sourceType, &sourceProject, &category, &sourceLabel, &count); err != nil {
			return nil, err
		}
		sourceType = strings.TrimSpace(sourceType)
		sourceProject = strings.TrimSpace(sourceProject)
		category = strings.TrimSpace(category)
		sourceLabel = strings.TrimSpace(sourceLabel)
		if sourceProject == "" {
			sourceProject = "manual"
		}
		if category == "" {
			category = "general"
		}

		summary.Total += count
		if sourceType == "template" {
			summary.TemplateCount += count
			templateGroupCounts[category] += count
		} else {
			summary.CaseCount += count
		}
		sourceCounts[sourceProject] += count
		if sourceLabel != "" {
			sourceLabels[sourceProject] = sourceLabel
		}
		categoryCounts[category] += count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	summary.Sources = promptCatalogFacetCounts(sourceCounts, sourceLabels, false)
	summary.Categories = promptCatalogFacetCounts(categoryCounts, nil, true)
	summary.TemplateGroups = promptCatalogFacetCounts(templateGroupCounts, nil, true)
	summary.SourceCount = len(summary.Sources)
	summary.CategoryCount = len(summary.Categories)
	return summary, nil
}

func (r *promptCatalogRepository) GetCaseByID(ctx context.Context, id string) (*service.PromptCatalogCase, error) {
	query := promptCatalogSelectSQL() + `
WHERE id = $1 AND status = $2 AND deleted_at IS NULL
LIMIT 1`
	item, err := scanPromptCatalogCase(r.db.QueryRowContext(ctx, query, id, service.PromptCatalogStatusPublished))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrPromptCatalogNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *promptCatalogRepository) UpsertCase(ctx context.Context, item *service.PromptCatalogCase) error {
	tagsJSON, err := encodePromptCatalogStringArray(item.Tags)
	if err != nil {
		return err
	}
	modelTagsJSON, err := encodePromptCatalogStringArray(item.ModelTags)
	if err != nil {
		return err
	}
	imageURLsJSON, err := encodePromptCatalogStringArray(item.ImageURLs)
	if err != nil {
		return err
	}
	stylesJSON, err := encodePromptCatalogStringArray(item.Styles)
	if err != nil {
		return err
	}
	scenesJSON, err := encodePromptCatalogStringArray(item.Scenes)
	if err != nil {
		return err
	}
	rawJSON := strings.TrimSpace(item.RawJSON)
	if rawJSON == "" {
		rawJSON = "{}"
	}

	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, `
INSERT INTO `+promptCatalogItemsTable+` (
	id,
	title,
	prompt,
	prompt_preview,
	category,
	tags,
	model_tags,
	source_url,
	image_url,
	image_urls,
	source_project,
	source_type,
	source_label,
	github_url,
	featured,
	styles,
	scenes,
	image_original_url,
	image_preview_url,
	image_thumb_url,
	import_source,
	raw_json,
	status,
	imported_at,
	created_at,
	updated_at
) VALUES (
	$1, $2, $3, $4, $5,
	$6::jsonb, $7::jsonb,
	NULLIF($8, ''),
	NULLIF($9, ''),
	$10::jsonb,
	$11, $12,
	NULLIF($13, ''),
	NULLIF($14, ''),
	$15,
	$16::jsonb, $17::jsonb,
	NULLIF($18, ''),
	NULLIF($19, ''),
	NULLIF($20, ''),
	$21,
	$22::jsonb,
	$23,
	$24,
	$25,
	$26
)
ON CONFLICT (id) DO UPDATE SET
	title = EXCLUDED.title,
	prompt = EXCLUDED.prompt,
	prompt_preview = EXCLUDED.prompt_preview,
	category = EXCLUDED.category,
	tags = EXCLUDED.tags,
	model_tags = EXCLUDED.model_tags,
	source_url = EXCLUDED.source_url,
	image_url = EXCLUDED.image_url,
	image_urls = EXCLUDED.image_urls,
	source_project = EXCLUDED.source_project,
	source_type = EXCLUDED.source_type,
	source_label = EXCLUDED.source_label,
	github_url = EXCLUDED.github_url,
	featured = EXCLUDED.featured,
	styles = EXCLUDED.styles,
	scenes = EXCLUDED.scenes,
	image_original_url = EXCLUDED.image_original_url,
	image_preview_url = EXCLUDED.image_preview_url,
	image_thumb_url = EXCLUDED.image_thumb_url,
	import_source = EXCLUDED.import_source,
	raw_json = EXCLUDED.raw_json,
	status = EXCLUDED.status,
	imported_at = EXCLUDED.imported_at,
	updated_at = EXCLUDED.updated_at,
	deleted_at = NULL
`, item.ID,
		item.Title,
		item.Prompt,
		item.PromptPreview,
		item.Category,
		tagsJSON,
		modelTagsJSON,
		item.SourceURL,
		item.ImageURL,
		imageURLsJSON,
		item.SourceProject,
		item.SourceType,
		item.SourceLabel,
		item.GitHubURL,
		item.Featured,
		stylesJSON,
		scenesJSON,
		item.ImageOriginalURL,
		item.ImagePreviewURL,
		item.ImageThumbURL,
		item.ImportSource,
		rawJSON,
		item.Status,
		item.ImportedAt,
		item.CreatedAt,
		item.UpdatedAt,
	)
	return err
}

func buildPromptCatalogWhere(filters service.PromptCatalogListFilters) (string, []any) {
	clauses := []string{"status = $1", "deleted_at IS NULL"}
	args := []any{service.PromptCatalogStatusPublished}

	addClause := func(sql string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(sql, len(args)))
	}

	if filters.SourceType != "" {
		addClause("source_type = $%d", filters.SourceType)
	}
	if filters.SourceProject != "" {
		addClause("source_project = $%d", filters.SourceProject)
	}
	if filters.Category != "" {
		addClause("category = $%d", filters.Category)
	}
	if filters.Featured != nil {
		addClause("featured = $%d", *filters.Featured)
	}
	if filters.HasImage != nil {
		imageClause := `(NULLIF(image_url, '') IS NOT NULL OR NULLIF(image_thumb_url, '') IS NOT NULL OR NULLIF(image_preview_url, '') IS NOT NULL OR NULLIF(image_original_url, '') IS NOT NULL OR COALESCE(jsonb_array_length(image_urls), 0) > 0)`
		if *filters.HasImage {
			clauses = append(clauses, imageClause)
		} else {
			clauses = append(clauses, "NOT "+imageClause)
		}
	}
	if filters.Search != "" {
		args = append(args, "%"+filters.Search+"%")
		placeholder := "$" + fmt.Sprint(len(args))
		clauses = append(clauses, "(title ILIKE "+placeholder+" OR prompt ILIKE "+placeholder+" OR category ILIKE "+placeholder+" OR source_project ILIKE "+placeholder+" OR source_label ILIKE "+placeholder+")")
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func promptCatalogSelectSQL() string {
	return `
SELECT
	id,
	title,
	prompt,
	prompt_preview,
	category,
	tags,
	model_tags,
	COALESCE(source_url, ''),
	COALESCE(image_url, ''),
	image_urls,
	COALESCE(image_original_url, ''),
	COALESCE(image_preview_url, ''),
	COALESCE(image_thumb_url, ''),
	source_project,
	source_type,
	COALESCE(source_label, ''),
	COALESCE(github_url, ''),
	featured,
	styles,
	scenes,
	import_source,
	COALESCE(raw_json::text, '{}'),
	status,
	imported_at,
	created_at,
	updated_at
FROM ` + promptCatalogItemsTable
}

func promptCatalogOrderBy(params pagination.PaginationParams) string {
	sortOrder := strings.ToUpper(params.NormalizedSortOrder(pagination.SortOrderDesc))
	switch strings.ToLower(strings.TrimSpace(params.SortBy)) {
	case "title":
		return "title " + sortOrder + ", imported_at DESC NULLS LAST, id ASC"
	case "created_at":
		return "created_at " + sortOrder + " NULLS LAST, title ASC, id ASC"
	case "updated_at":
		return "updated_at " + sortOrder + " NULLS LAST, title ASC, id ASC"
	default:
		return "imported_at " + sortOrder + " NULLS LAST, title ASC, id ASC"
	}
}

type promptCatalogScanner interface {
	Scan(dest ...any) error
}

func scanPromptCatalogCase(row promptCatalogScanner) (service.PromptCatalogCase, error) {
	var item service.PromptCatalogCase
	var tagsJSON, modelTagsJSON, imageURLsJSON, stylesJSON, scenesJSON []byte
	var importedAt sql.NullTime

	err := row.Scan(
		&item.ID,
		&item.Title,
		&item.Prompt,
		&item.PromptPreview,
		&item.Category,
		&tagsJSON,
		&modelTagsJSON,
		&item.SourceURL,
		&item.ImageURL,
		&imageURLsJSON,
		&item.ImageOriginalURL,
		&item.ImagePreviewURL,
		&item.ImageThumbURL,
		&item.SourceProject,
		&item.SourceType,
		&item.SourceLabel,
		&item.GitHubURL,
		&item.Featured,
		&stylesJSON,
		&scenesJSON,
		&item.ImportSource,
		&item.RawJSON,
		&item.Status,
		&importedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return item, err
	}
	item.Tags = decodePromptCatalogStringArray(tagsJSON)
	item.ModelTags = decodePromptCatalogStringArray(modelTagsJSON)
	item.ImageURLs = decodePromptCatalogStringArray(imageURLsJSON)
	item.Styles = decodePromptCatalogStringArray(stylesJSON)
	item.Scenes = decodePromptCatalogStringArray(scenesJSON)
	if importedAt.Valid {
		t := importedAt.Time
		item.ImportedAt = &t
	}
	return item, nil
}

func decodePromptCatalogStringArray(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return []string{}
	}
	if values == nil {
		return []string{}
	}
	return values
}

func encodePromptCatalogStringArray(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func promptCatalogFacetCounts(counts map[string]int64, labels map[string]string, countFirst bool) []service.PromptCatalogFacetCount {
	facets := make([]service.PromptCatalogFacetCount, 0, len(counts))
	for value, count := range counts {
		facets = append(facets, service.PromptCatalogFacetCount{Value: value, Label: labels[value], Count: count})
	}
	sort.Slice(facets, func(i, j int) bool {
		if countFirst && facets[i].Count != facets[j].Count {
			return facets[i].Count > facets[j].Count
		}
		return facets[i].Value < facets[j].Value
	})
	return facets
}
