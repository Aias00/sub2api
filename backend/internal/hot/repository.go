package hot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Aias00/cloudbase/internal/pkg/pagination"
)

type sqlExecutor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type repository struct {
	db  *sql.DB
	sql sqlExecutor
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db, sql: db}
}

func (r *repository) ListSources(ctx context.Context) ([]Source, error) {
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
	return scanSourceRows(rows)
}

func (r *repository) ListItems(ctx context.Context, params pagination.PaginationParams, filters ListFilters) ([]Item, *pagination.PaginationResult, error) {
	where, args := buildItemWhere(filters)
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM hot_items "+where, args, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []Item{}, paginationResultFromTotal(0, params), nil
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
	items, err := scanItemRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *repository) ListRunEvents(ctx context.Context, runID string, params pagination.PaginationParams) ([]RunEvent, *pagination.PaginationResult, error) {
	args := []any{runID}
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM hot_run_events WHERE run_id = $1", args, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []RunEvent{}, paginationResultFromTotal(0, params), nil
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
	items, err := scanRunEventRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func buildItemWhere(filters ListFilters) (string, []any) {
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

func scanSourceRows(rows *sql.Rows) ([]Source, error) {
	items := make([]Source, 0)
	for rows.Next() {
		var item Source
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

func scanItemRows(rows *sql.Rows) ([]Item, error) {
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
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

func scanRunEventRows(rows *sql.Rows) ([]RunEvent, error) {
	items := make([]RunEvent, 0)
	for rows.Next() {
		var item RunEvent
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

func scanSingleRow(ctx context.Context, q sqlExecutor, query string, args []any, dest ...any) (err error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	if err = rows.Scan(dest...); err != nil {
		return err
	}
	return rows.Err()
}

func paginationResultFromTotal(total int64, params pagination.PaginationParams) *pagination.PaginationResult {
	pages := int(total) / params.Limit()
	if int(total)%params.Limit() > 0 {
		pages++
	}
	return &pagination.PaginationResult{
		Total:    total,
		Page:     params.Page,
		PageSize: params.Limit(),
		Pages:    pages,
	}
}
