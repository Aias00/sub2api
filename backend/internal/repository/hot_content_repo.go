package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type hotContentRepository struct {
	db  *sql.DB
	sql sqlExecutor
}

func NewHotContentRepository(db *sql.DB) service.HotContentRepository {
	return &hotContentRepository{db: db, sql: db}
}

func (r *hotContentRepository) ListSources(ctx context.Context) ([]service.HotSource, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, source_id, adapter_kind, title, description, enabled, base_url,
			seed_urls_json, config_json, sort_order, created_at, updated_at
		FROM hot_sources
		WHERE enabled = TRUE
		ORDER BY sort_order ASC, title ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanHotSourceRows(rows)
}

func (r *hotContentRepository) ListItems(ctx context.Context, params pagination.PaginationParams, filters service.HotContentListFilters) ([]service.HotItem, *pagination.PaginationResult, error) {
	where, args := buildHotItemWhere(filters)
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM hot_items "+where, args, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []service.HotItem{}, paginationResultFromTotal(0, params), nil
	}
	query := `
		SELECT id, source_id, external_id, canonical_url, title, summary, body, quoted, reason,
			published_at, author, source_name, source_handle, badge, score, content_type,
			tags_json, metrics_json, raw_ref_json, content_hash, has_media, status, created_at, updated_at
		FROM hot_items
		` + where + `
		ORDER BY published_at DESC NULLS LAST, id DESC
		LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, params.Limit(), params.Offset())
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items, err := scanHotItemRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *hotContentRepository) ListRunEvents(ctx context.Context, runID string, params pagination.PaginationParams) ([]service.HotRunEvent, *pagination.PaginationResult, error) {
	args := []any{runID}
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM hot_run_events WHERE run_id = $1", args, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []service.HotRunEvent{}, paginationResultFromTotal(0, params), nil
	}
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, legacy_id, run_id, node, message, payload_json, created_at
		FROM hot_run_events
		WHERE run_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT $2 OFFSET $3
	`, runID, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items, err := scanHotRunEventRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func buildHotItemWhere(filters service.HotContentListFilters) (string, []any) {
	clauses := []string{"status = $1"}
	args := []any{filters.Status}
	if strings.TrimSpace(filters.SourceID) != "" {
		args = append(args, strings.TrimSpace(filters.SourceID))
		clauses = append(clauses, fmt.Sprintf("source_id = $%d", len(args)))
	}
	if strings.TrimSpace(filters.Query) != "" {
		args = append(args, "%"+strings.TrimSpace(filters.Query)+"%")
		clauses = append(clauses, fmt.Sprintf("(title ILIKE $%d OR summary ILIKE $%d OR body ILIKE $%d)", len(args), len(args), len(args)))
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func scanHotSourceRows(rows *sql.Rows) ([]service.HotSource, error) {
	items := make([]service.HotSource, 0)
	for rows.Next() {
		var item service.HotSource
		var seedURLsJSON []byte
		var configJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.SourceID,
			&item.AdapterKind,
			&item.Title,
			&item.Description,
			&item.Enabled,
			&item.BaseURL,
			&seedURLsJSON,
			&configJSON,
			&item.SortOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.SeedURLsJSON = string(seedURLsJSON)
		item.ConfigJSON = string(configJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanHotItemRows(rows *sql.Rows) ([]service.HotItem, error) {
	items := make([]service.HotItem, 0)
	for rows.Next() {
		var item service.HotItem
		var publishedAt sql.NullTime
		var tagsJSON []byte
		var metricsJSON []byte
		var rawRefJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.SourceID,
			&item.ExternalID,
			&item.CanonicalURL,
			&item.Title,
			&item.Summary,
			&item.Body,
			&item.Quoted,
			&item.Reason,
			&publishedAt,
			&item.Author,
			&item.SourceName,
			&item.SourceHandle,
			&item.Badge,
			&item.Score,
			&item.ContentType,
			&tagsJSON,
			&metricsJSON,
			&rawRefJSON,
			&item.ContentHash,
			&item.HasMedia,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if publishedAt.Valid {
			item.PublishedAt = &publishedAt.Time
		}
		item.TagsJSON = string(tagsJSON)
		item.MetricsJSON = string(metricsJSON)
		item.RawRefJSON = string(rawRefJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanHotRunEventRows(rows *sql.Rows) ([]service.HotRunEvent, error) {
	items := make([]service.HotRunEvent, 0)
	for rows.Next() {
		var item service.HotRunEvent
		var payloadJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.LegacyID,
			&item.RunID,
			&item.Node,
			&item.Message,
			&payloadJSON,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.PayloadJSON = string(payloadJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}
