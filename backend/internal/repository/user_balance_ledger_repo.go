package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type userBalanceLedgerRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewUserBalanceLedgerRepository 创建余额流水仓储
func NewUserBalanceLedgerRepository(client *ent.Client, db *sql.DB) service.UserBalanceLedgerRepository {
	return &userBalanceLedgerRepository{
		client: client,
		db:     db,
	}
}

// Create 写入流水记录
func (r *userBalanceLedgerRepository) Create(ctx context.Context, entry *service.UserBalanceLedgerEntry) error {
	query := `
INSERT INTO user_balance_ledger (
    user_id, entry_type, amount,
    balance_before, balance_after,
    source_type, source_id,
    description, metadata_json,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`
	exec := txAwareSQLExecutor(ctx, r.db, r.client)
	_, err := exec.ExecContext(ctx, query,
		entry.UserID,
		entry.EntryType,
		entry.Amount,
		entry.BalanceBefore,
		entry.BalanceAfter,
		entry.SourceType,
		entry.SourceID,
		entry.Description,
		entry.MetadataJSON,
		entry.CreatedAt,
	)
	return err
}

// CreateTx 在指定事务内写入流水记录
// exec 参数可以是 *sql.Tx 或其他支持 ExecContext 的执行器
func (r *userBalanceLedgerRepository) CreateTx(
	ctx context.Context,
	exec interface {
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	},
	entry *service.UserBalanceLedgerEntry,
) error {
	query := `
INSERT INTO user_balance_ledger (
    user_id, entry_type, amount,
    balance_before, balance_after,
    source_type, source_id,
    description, metadata_json,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`
	_, err := exec.ExecContext(ctx, query,
		entry.UserID,
		entry.EntryType,
		entry.Amount,
		entry.BalanceBefore,
		entry.BalanceAfter,
		entry.SourceType,
		entry.SourceID,
		entry.Description,
		entry.MetadataJSON,
		entry.CreatedAt,
	)
	return err
}

// ListByUser 查询用户流水（分页）
func (r *userBalanceLedgerRepository) ListByUser(
	ctx context.Context,
	userID int64,
	filter service.BalanceLedgerFilter,
) ([]service.UserBalanceLedgerEntry, int64, error) {
	// 构建 WHERE 条件
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if len(filter.EntryTypes) > 0 {
		whereClause += fmt.Sprintf(" AND entry_type IN (%s)", buildInClause(argIndex, len(filter.EntryTypes)))
		for _, t := range filter.EntryTypes {
			args = append(args, t)
			argIndex++
		}
	}

	if filter.StartAt != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartAt)
		argIndex++
	}

	if filter.EndAt != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndAt)
		argIndex++
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM user_balance_ledger %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count balance ledger: %w", err)
	}

	if total == 0 {
		return []service.UserBalanceLedgerEntry{}, 0, nil
	}

	// 计算分页
	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, offset, filter.PageSize)

	// 查询列表
	listQuery := fmt.Sprintf(`
SELECT id, user_id, entry_type, amount,
       balance_before, balance_after,
       source_type, source_id,
       description, metadata_json,
       created_at
FROM user_balance_ledger %s
ORDER BY created_at DESC
LIMIT $%d OFFSET $%d
`, whereClause, argIndex, argIndex+1)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list balance ledger: %w", err)
	}
	defer rows.Close()

	var entries []service.UserBalanceLedgerEntry
	for rows.Next() {
		var entry service.UserBalanceLedgerEntry
		var metadataJSON []byte
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.EntryType,
			&entry.Amount,
			&entry.BalanceBefore,
			&entry.BalanceAfter,
			&entry.SourceType,
			&entry.SourceID,
			&entry.Description,
			&metadataJSON,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan balance ledger entry: %w", err)
		}
		if metadataJSON != nil {
			entry.MetadataJSON = json.RawMessage(metadataJSON)
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate balance ledger: %w", err)
	}

	return entries, total, nil
}

// GetBySource 查询某来源的流水（用于去重检查）
func (r *userBalanceLedgerRepository) GetBySource(
	ctx context.Context,
	sourceType service.BalanceLedgerSourceType,
	sourceID int64,
) (*service.UserBalanceLedgerEntry, error) {
	query := `
SELECT id, user_id, entry_type, amount,
       balance_before, balance_after,
       source_type, source_id,
       description, metadata_json,
       created_at
FROM user_balance_ledger
WHERE source_type = $1 AND source_id = $2
LIMIT 1
`
	row := r.db.QueryRowContext(ctx, query, sourceType, sourceID)

	var entry service.UserBalanceLedgerEntry
	var metadataJSON []byte
	err := row.Scan(
		&entry.ID,
		&entry.UserID,
		&entry.EntryType,
		&entry.Amount,
		&entry.BalanceBefore,
		&entry.BalanceAfter,
		&entry.SourceType,
		&entry.SourceID,
		&entry.Description,
		&metadataJSON,
		&entry.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get balance ledger by source: %w", err)
	}
	if metadataJSON != nil {
		entry.MetadataJSON = json.RawMessage(metadataJSON)
	}
	return &entry, nil
}

// buildInClause 构建 IN 子查询占位符
func buildInClause(startIndex, count int) string {
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", startIndex+i)
	}
	return fmt.Sprintf("(%s)", joinPlaceholders(placeholders))
}

func joinPlaceholders(placeholders []string) string {
	result := ""
	for i, p := range placeholders {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
