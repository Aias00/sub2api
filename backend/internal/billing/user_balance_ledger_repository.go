package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/Aias00/cloudbase/ent"
)

type userBalanceLedgerRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewUserBalanceLedgerRepository 创建余额流水仓储
func NewUserBalanceLedgerRepository(client *ent.Client, db *sql.DB) UserBalanceLedgerRepository {
	return &userBalanceLedgerRepository{
		client: client,
		db:     db,
	}
}

// Create 写入流水记录
func (r *userBalanceLedgerRepository) Create(ctx context.Context, entry *UserBalanceLedgerEntry) error {
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
	entry *UserBalanceLedgerEntry,
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
	filter BalanceLedgerFilter,
) ([]UserBalanceLedgerEntry, int64, error) {
	// 构建 WHERE 条件
	whereClause := "WHERE user_id = $1"
	args := []any{userID}
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
		return []UserBalanceLedgerEntry{}, 0, nil
	}

	// 计算分页
	offset := (filter.Page - 1) * filter.PageSize
	// 参数顺序必须与下方 SQL 的 "LIMIT $argIndex OFFSET $argIndex+1" 对应：
	// 先 append PageSize(绑定 LIMIT)，再 append offset(绑定 OFFSET)。
	args = append(args, filter.PageSize, offset)

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
	defer func() { _ = rows.Close() }()

	var entries []UserBalanceLedgerEntry
	for rows.Next() {
		var entry UserBalanceLedgerEntry
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
	sourceType BalanceLedgerSourceType,
	sourceID int64,
) (*UserBalanceLedgerEntry, error) {
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

	var entry UserBalanceLedgerEntry
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

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func txAwareSQLExecutor(ctx context.Context, fallback sqlExecutor, client *ent.Client) sqlExecutor {
	if tx := ent.TxFromContext(ctx); tx != nil {
		if exec := sqlExecutorFromEntClient(tx.Client()); exec != nil {
			return exec
		}
	}
	if fallback != nil {
		return fallback
	}
	return sqlExecutorFromEntClient(client)
}

func sqlExecutorFromEntClient(client *ent.Client) sqlExecutor {
	if client == nil {
		return nil
	}
	clientValue := reflect.ValueOf(client).Elem()
	configValue := clientValue.FieldByName("config")
	driverValue := configValue.FieldByName("driver")
	if !driverValue.IsValid() {
		return nil
	}
	driver := reflect.NewAt(driverValue.Type(), unsafe.Pointer(driverValue.UnsafeAddr())).Elem().Interface()
	exec, ok := driver.(sqlExecutor)
	if !ok {
		return nil
	}
	return exec
}

// ApplyBalanceChange atomically adjusts users.balance and records a ledger row
// in a single transaction:
//  1. SELECT ... FOR UPDATE locks the user row (serializes concurrent changes).
//  2. When SourceID is set, an existing (source_type, source_id) ledger row makes
//     this a no-op (idempotent) returning the prior result.
//  3. Debits that would overdraft are rejected unless AllowNegative is set.
//  4. balance_before/balance_after are captured within the same tx so the ledger
//     is a continuous, reconcilable audit trail.
func (r *userBalanceLedgerRepository) ApplyBalanceChange(ctx context.Context, cmd BalanceChangeCommand) (*BalanceChangeResult, error) {
	if cmd.UserID <= 0 {
		return nil, errors.New("balance change requires a valid user_id")
	}
	if cmd.Delta == 0 {
		return nil, errors.New("balance change delta must be non-zero")
	}
	if r.db == nil {
		return nil, errors.New("balance change requires a *sql.DB executor")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin balance change tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// 1. Lock the user row and read the current balance.
	var before float64
	err = tx.QueryRowContext(ctx,
		`SELECT balance FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`,
		cmd.UserID,
	).Scan(&before)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrLedgerUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock user for balance change: %w", err)
	}

	// 2. Idempotency: a prior ledger row for the same source short-circuits.
	if cmd.SourceID != nil {
		var (
			existingID  int64
			existBefore sql.NullFloat64
			existAfter  sql.NullFloat64
		)
		e := tx.QueryRowContext(ctx,
			`SELECT id, balance_before, balance_after FROM user_balance_ledger
			 WHERE source_type = $1 AND source_id = $2 LIMIT 1`,
			cmd.SourceType, *cmd.SourceID,
		).Scan(&existingID, &existBefore, &existAfter)
		if e == nil {
			if err = tx.Commit(); err != nil {
				return nil, fmt.Errorf("commit idempotent balance change: %w", err)
			}
			committed = true
			return &BalanceChangeResult{
				Applied:       false,
				BalanceBefore: existBefore.Float64,
				BalanceAfter:  existAfter.Float64,
				LedgerID:      existingID,
			}, nil
		}
		if !errors.Is(e, sql.ErrNoRows) {
			return nil, fmt.Errorf("check ledger idempotency: %w", e)
		}
	}

	after := before + cmd.Delta
	if after < 0 && !cmd.AllowNegative {
		return nil, ErrInsufficientBalance
	}

	// 3. Update the balance to the computed value (and gift_balance delta if any).
	// Maintains balance invariant: balance = paid_balance + gift_balance
	// paid_balance derives from (Delta - GiftDelta) to preserve the invariant.
	if _, err = tx.ExecContext(ctx,
		`UPDATE users
		 SET balance = $1,
		     gift_balance = COALESCE(gift_balance, 0) + $2,
		     paid_balance = COALESCE(paid_balance, 0) + ($4 - $2),
		     total_recharged = CASE WHEN $5 AND $4 > 0 THEN total_recharged + $4 ELSE total_recharged END,
		     updated_at = NOW()
		 WHERE id = $3`,
		after, cmd.GiftDelta, cmd.UserID, cmd.Delta, cmd.UpdateRecharged,
	); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	// 4. Record the ledger row with the atomic before/after snapshot.
	metadata := cmd.MetadataJSON
	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}
	var ledgerID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO user_balance_ledger (
			user_id, entry_type, amount,
			balance_before, balance_after,
			source_type, source_id,
			description, metadata_json, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		cmd.UserID,
		cmd.EntryType,
		cmd.Delta,
		before,
		after,
		cmd.SourceType,
		cmd.SourceID,
		cmd.Description,
		[]byte(metadata),
		time.Now().UTC(),
	).Scan(&ledgerID)
	if err != nil {
		return nil, fmt.Errorf("insert balance ledger: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit balance change: %w", err)
	}
	committed = true

	return &BalanceChangeResult{
		Applied:       true,
		BalanceBefore: before,
		BalanceAfter:  after,
		LedgerID:      ledgerID,
	}, nil
}

// ApplyBalanceChangeTx runs the same balance+ledger mutation as ApplyBalanceChange
// but participates in the caller's transaction (it does NOT open or commit a tx).
// Instead of SELECT ... FOR UPDATE it uses a single conditional UPDATE ... RETURNING
// as both the overdraft guard and the row lock, so it composes cleanly inside a
// larger transaction (e.g. usage billing, recharge, first-bind grant).
func (r *userBalanceLedgerRepository) ApplyBalanceChangeTx(ctx context.Context, exec BalanceTxExecutor, cmd BalanceChangeCommand) (*BalanceChangeResult, error) {
	if exec == nil {
		return nil, errors.New("balance change tx requires an executor")
	}
	if cmd.UserID <= 0 {
		return nil, errors.New("balance change requires a valid user_id")
	}
	if cmd.Delta == 0 {
		return nil, errors.New("balance change delta must be non-zero")
	}

	// Idempotency: a prior ledger row for the same source short-circuits.
	if cmd.SourceID != nil {
		var (
			existingID  int64
			existBefore sql.NullFloat64
			existAfter  sql.NullFloat64
		)
		found, err := queryOneBalanceRow(ctx, exec,
			`SELECT id, balance_before, balance_after FROM user_balance_ledger
			 WHERE source_type = $1 AND source_id = $2 LIMIT 1`,
			func(rows *sql.Rows) error { return rows.Scan(&existingID, &existBefore, &existAfter) },
			cmd.SourceType, *cmd.SourceID,
		)
		if err != nil {
			return nil, fmt.Errorf("check ledger idempotency: %w", err)
		}
		if found {
			return &BalanceChangeResult{
				Applied:       false,
				BalanceBefore: existBefore.Float64,
				BalanceAfter:  existAfter.Float64,
				LedgerID:      existingID,
			}, nil
		}
	}

	// Conditional atomic update: guards overdraft and row-locks the user in one shot.
	// Maintains balance invariant: balance = paid_balance + gift_balance
	// paid_balance derives from (Delta - GiftDelta) to preserve the invariant.
	var after float64
	updated, err := queryOneBalanceRow(ctx, exec,
		`UPDATE users
		 SET balance = balance + $1,
		     gift_balance = COALESCE(gift_balance, 0) + $2,
		     paid_balance = COALESCE(paid_balance, 0) + ($1 - $2),
		     total_recharged = CASE WHEN $5 AND $1 > 0 THEN total_recharged + $1 ELSE total_recharged END,
		     updated_at = NOW()
		 WHERE id = $3 AND deleted_at IS NULL AND ($4 OR balance + $1 >= 0)
		 RETURNING balance`,
		func(rows *sql.Rows) error { return rows.Scan(&after) },
		cmd.Delta, cmd.GiftDelta, cmd.UserID, cmd.AllowNegative, cmd.UpdateRecharged,
	)
	if err != nil {
		return nil, fmt.Errorf("conditional balance update: %w", err)
	}
	if !updated {
		// No row updated: distinguish "user missing" from "insufficient balance".
		var exists bool
		if _, err := queryOneBalanceRow(ctx, exec,
			`SELECT true FROM users WHERE id = $1 AND deleted_at IS NULL`,
			func(rows *sql.Rows) error { return rows.Scan(&exists) },
			cmd.UserID,
		); err != nil {
			return nil, fmt.Errorf("resolve balance update failure: %w", err)
		}
		if !exists {
			return nil, ErrLedgerUserNotFound
		}
		return nil, ErrInsufficientBalance
	}

	before := after - cmd.Delta

	metadata := cmd.MetadataJSON
	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}
	var ledgerID int64
	if _, err := queryOneBalanceRow(ctx, exec,
		`INSERT INTO user_balance_ledger (
			user_id, entry_type, amount,
			balance_before, balance_after,
			source_type, source_id,
			description, metadata_json, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		func(rows *sql.Rows) error { return rows.Scan(&ledgerID) },
		cmd.UserID, cmd.EntryType, cmd.Delta, before, after,
		cmd.SourceType, cmd.SourceID, cmd.Description, []byte(metadata), time.Now().UTC(),
	); err != nil {
		return nil, fmt.Errorf("insert balance ledger (tx): %w", err)
	}

	return &BalanceChangeResult{
		Applied:       true,
		BalanceBefore: before,
		BalanceAfter:  after,
		LedgerID:      ledgerID,
	}, nil
}

// queryOneBalanceRow runs a query via the (tx) executor and scans at most one row.
// Returns found=false when the query yields no rows.
func queryOneBalanceRow(ctx context.Context, exec BalanceTxExecutor, query string, scan func(*sql.Rows) error, args ...any) (bool, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return false, rows.Err()
	}
	if err := scan(rows); err != nil {
		return false, err
	}
	return true, rows.Err()
}

// ApplyBalanceChangeCtx dispatches to the tx-aware or self-tx variant based on
// whether an ent transaction is present in ctx. Paths that already run inside an
// ent tx (e.g. OAuth first-bind grant) get their balance change + ledger row
// committed atomically with the surrounding work; paths with no ambient tx get a
// self-contained transaction.
func (r *userBalanceLedgerRepository) ApplyBalanceChangeCtx(ctx context.Context, cmd BalanceChangeCommand) (*BalanceChangeResult, error) {
	if tx := ent.TxFromContext(ctx); tx != nil {
		exec := sqlExecutorFromEntClient(tx.Client())
		if exec == nil {
			return nil, errors.New("cannot derive sql executor from ent transaction")
		}
		return r.ApplyBalanceChangeTx(ctx, exec, cmd)
	}
	return r.ApplyBalanceChange(ctx, cmd)
}
