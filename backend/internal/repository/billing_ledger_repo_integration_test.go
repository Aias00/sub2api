//go:build integration

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/internal/billing"
	"github.com/stretchr/testify/require"
)

// These tests exercise the real user_balance_ledger SQL against a throwaway
// Postgres container (testcontainers). They cover the previously-untested
// financial ledger repository: Create, ListByUser (filter/paging/ordering),
// and GetBySource (idempotency dedup lookup).

func ptrFloat(v float64) *float64 { return &v }
func ptrInt64(v int64) *int64     { return &v }

func newLedgerRepoForTest() billing.UserBalanceLedgerRepository {
	return billing.NewUserBalanceLedgerRepository(integrationEntClient, integrationDB)
}

// createLedgerTestUser inserts a real users row (user_balance_ledger has an FK
// to users) and returns its id.
func createLedgerTestUser(t *testing.T) int64 {
	t.Helper()
	u, err := integrationEntClient.User.Create().
		SetEmail(fmt.Sprintf("ledger-%d@test.local", time.Now().UnixNano())).
		SetPasswordHash("test-password-hash").
		SetStatus("active").
		SetRole("user").
		Save(context.Background())
	require.NoError(t, err, "create test user")
	return u.ID
}

func TestUserBalanceLedgerRepository_CreateAndListByUser(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)

	base := time.Now().UTC().Truncate(time.Second)
	entries := []*billing.UserBalanceLedgerEntry{
		{
			UserID:        userID,
			EntryType:     billing.EntryTypeRecharge,
			Amount:        100,
			BalanceBefore: ptrFloat(0),
			BalanceAfter:  ptrFloat(100),
			SourceType:    billing.SourceTypePaymentOrder,
			SourceID:      ptrInt64(5001),
			Description:   "recharge",
			MetadataJSON:  json.RawMessage(`{"k":"v"}`),
			CreatedAt:     base.Add(1 * time.Second),
		},
		{
			UserID:        userID,
			EntryType:     billing.EntryTypeAPIUsage,
			Amount:        -3.5,
			BalanceBefore: ptrFloat(100),
			BalanceAfter:  ptrFloat(96.5),
			SourceType:    billing.SourceTypeUsageLog,
			SourceID:      ptrInt64(5002),
			Description:   "api usage",
			MetadataJSON:  json.RawMessage(`{}`),
			CreatedAt:     base.Add(2 * time.Second),
		},
		{
			UserID:       userID,
			EntryType:    billing.EntryTypeRedeem,
			Amount:       20,
			SourceType:   billing.SourceTypeRedeemCode,
			SourceID:     ptrInt64(5003),
			Description:  "redeem",
			MetadataJSON: json.RawMessage(`{}`),
			CreatedAt:    base.Add(3 * time.Second),
		},
	}
	for _, e := range entries {
		require.NoError(t, repo.Create(ctx, e))
	}

	// List all: expect 3, ordered by created_at DESC (redeem newest first).
	got, total, err := repo.ListByUser(ctx, userID, billing.BalanceLedgerFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, got, 3)
	require.Equal(t, billing.EntryTypeRedeem, got[0].EntryType, "newest first (DESC)")
	require.Equal(t, billing.EntryTypeRecharge, got[2].EntryType, "oldest last")

	// Round-trip fidelity of a signed amount + nullable balances + metadata.
	require.Equal(t, -3.5, got[1].Amount)
	require.NotNil(t, got[1].BalanceAfter)
	require.Equal(t, 96.5, *got[1].BalanceAfter)

	// Pagination: page size 2 → first page 2 items, total still 3.
	page1, total, err := repo.ListByUser(ctx, userID, billing.BalanceLedgerFilter{Page: 1, PageSize: 2})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, page1, 2)
	page2, _, err := repo.ListByUser(ctx, userID, billing.BalanceLedgerFilter{Page: 2, PageSize: 2})
	require.NoError(t, err)
	require.Len(t, page2, 1)

	// Entry-type filter: only api_usage.
	filtered, total, err := repo.ListByUser(ctx, userID, billing.BalanceLedgerFilter{
		Page: 1, PageSize: 10,
		EntryTypes: []billing.BalanceLedgerEntryType{billing.EntryTypeAPIUsage},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, filtered, 1)
	require.Equal(t, billing.EntryTypeAPIUsage, filtered[0].EntryType)
}

func TestUserBalanceLedgerRepository_ListByUser_Empty(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()

	got, total, err := repo.ListByUser(ctx, createLedgerTestUser(t), billing.BalanceLedgerFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, got)
}

func TestUserBalanceLedgerRepository_GetBySource(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)

	// A distinct source id used for idempotency-dedup lookups.
	sourceID := time.Now().UnixNano()%1_000_000_000 + 700_000_000
	require.NoError(t, repo.Create(ctx, &billing.UserBalanceLedgerEntry{
		UserID:       userID,
		EntryType:    billing.EntryTypeRefund,
		Amount:       -10,
		SourceType:   billing.SourceTypeRefund,
		SourceID:     ptrInt64(sourceID),
		Description:  "refund dedup",
		MetadataJSON: json.RawMessage(`{}`),
		CreatedAt:    time.Now().UTC(),
	}))

	// Found: matches by (source_type, source_id).
	found, err := repo.GetBySource(ctx, billing.SourceTypeRefund, sourceID)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, userID, found.UserID)
	require.Equal(t, billing.EntryTypeRefund, found.EntryType)

	// Not found: unknown source id returns (nil, nil) — the dedup "no prior entry" signal.
	missing, err := repo.GetBySource(ctx, billing.SourceTypeRefund, sourceID+1)
	require.NoError(t, err)
	require.Nil(t, missing)

	// Not found: right id but wrong source type.
	wrongType, err := repo.GetBySource(ctx, billing.SourceTypePaymentOrder, sourceID)
	require.NoError(t, err)
	require.Nil(t, wrongType)
}

// readUserBalance reads the live users.balance for assertions.
func readUserBalance(t *testing.T, userID int64) float64 {
	t.Helper()
	var b float64
	require.NoError(t, integrationDB.QueryRowContext(context.Background(),
		"SELECT balance FROM users WHERE id = $1", userID).Scan(&b))
	return b
}

func TestApplyBalanceChange_CreditThenDebit_AtomicWithLedger(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)
	start := readUserBalance(t, userID)

	// Credit 100.
	credit, err := repo.ApplyBalanceChange(ctx, billing.BalanceChangeCommand{
		UserID: userID, Delta: 100, EntryType: billing.EntryTypeRecharge,
		SourceType: billing.SourceTypePaymentOrder, SourceID: ptrInt64(time.Now().UnixNano() % 1_000_000_000),
		Description: "recharge",
	})
	require.NoError(t, err)
	require.True(t, credit.Applied)
	require.Equal(t, start, credit.BalanceBefore)
	require.Equal(t, start+100, credit.BalanceAfter)
	require.Equal(t, start+100, readUserBalance(t, userID), "balance must reflect credit")

	// Debit 30.
	debit, err := repo.ApplyBalanceChange(ctx, billing.BalanceChangeCommand{
		UserID: userID, Delta: -30, EntryType: billing.EntryTypeAPIUsage,
		SourceType: billing.SourceTypeUsageLog, SourceID: ptrInt64(time.Now().UnixNano()%1_000_000_000 + 1),
		Description: "api usage",
	})
	require.NoError(t, err)
	require.True(t, debit.Applied)
	require.Equal(t, start+100, debit.BalanceBefore)
	require.Equal(t, start+70, debit.BalanceAfter)
	require.Equal(t, start+70, readUserBalance(t, userID))

	// Ledger must contain both rows with a continuous before/after chain.
	entries, total, err := repo.ListByUser(ctx, userID, billing.BalanceLedgerFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, entries, 2)
	// Newest first: the debit. Its before must equal the credit's after (continuity).
	require.NotNil(t, entries[0].BalanceBefore)
	require.NotNil(t, entries[1].BalanceAfter)
	require.Equal(t, *entries[1].BalanceAfter, *entries[0].BalanceBefore, "ledger before/after chain must be continuous")
}

func TestApplyBalanceChange_Idempotent(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)
	start := readUserBalance(t, userID)
	src := ptrInt64(time.Now().UnixNano()%1_000_000_000 + 2)

	cmd := billing.BalanceChangeCommand{
		UserID: userID, Delta: 50, EntryType: billing.EntryTypeRecharge,
		SourceType: billing.SourceTypePaymentOrder, SourceID: src, Description: "dup",
	}
	first, err := repo.ApplyBalanceChange(ctx, cmd)
	require.NoError(t, err)
	require.True(t, first.Applied)

	// Second call with the same source is a no-op.
	second, err := repo.ApplyBalanceChange(ctx, cmd)
	require.NoError(t, err)
	require.False(t, second.Applied, "duplicate source must not re-apply")
	require.Equal(t, first.LedgerID, second.LedgerID)

	// Balance credited exactly once; exactly one ledger row.
	require.Equal(t, start+50, readUserBalance(t, userID))
	_, total, err := repo.ListByUser(ctx, userID, billing.BalanceLedgerFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
}

func TestApplyBalanceChange_InsufficientBalance(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)
	start := readUserBalance(t, userID)

	// Debit more than available with AllowNegative=false → rejected, no mutation.
	_, err := repo.ApplyBalanceChange(ctx, billing.BalanceChangeCommand{
		UserID: userID, Delta: -(start + 1), EntryType: billing.EntryTypeAPIUsage,
		SourceType: billing.SourceTypeUsageLog, Description: "overdraft",
	})
	require.Error(t, err)
	require.Equal(t, start, readUserBalance(t, userID), "rejected debit must not change balance")

	// No ledger row written on rejection.
	_, total, err := repo.ListByUser(ctx, userID, billing.BalanceLedgerFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
}

func TestApplyBalanceChange_UserNotFound(t *testing.T) {
	_, err := newLedgerRepoForTest().ApplyBalanceChange(context.Background(), billing.BalanceChangeCommand{
		UserID: 999_999_999, Delta: 10, EntryType: billing.EntryTypeRecharge,
		SourceType: billing.SourceTypePaymentOrder,
	})
	require.ErrorIs(t, err, billing.ErrLedgerUserNotFound)
}

func TestApplyBalanceChangeTx_CreditDebitWithinCallerTx(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)
	start := readUserBalance(t, userID)

	tx := testTx(t) // participates as the caller's tx; rolled back after test

	credit, err := repo.ApplyBalanceChangeTx(ctx, tx, billing.BalanceChangeCommand{
		UserID: userID, Delta: 100, EntryType: billing.EntryTypeRecharge,
		SourceType: billing.SourceTypePaymentOrder, SourceID: ptrInt64(time.Now().UnixNano() % 1_000_000_000),
		Description: "recharge in tx",
	})
	require.NoError(t, err)
	require.True(t, credit.Applied)
	require.Equal(t, start+100, credit.BalanceAfter)

	debit, err := repo.ApplyBalanceChangeTx(ctx, tx, billing.BalanceChangeCommand{
		UserID: userID, Delta: -30, EntryType: billing.EntryTypeAPIUsage,
		SourceType: billing.SourceTypeUsageLog, SourceID: ptrInt64(time.Now().UnixNano()%1_000_000_000 + 1),
		Description: "usage in tx",
	})
	require.NoError(t, err)
	require.Equal(t, start+100, debit.BalanceBefore, "chain continuity within tx")
	require.Equal(t, start+70, debit.BalanceAfter)

	// Visible within the same tx.
	var bal float64
	require.NoError(t, tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1", userID).Scan(&bal))
	require.Equal(t, start+70, bal)

	var ledgerCount int
	require.NoError(t, tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_balance_ledger WHERE user_id = $1", userID).Scan(&ledgerCount))
	require.Equal(t, 2, ledgerCount)
}

func TestApplyBalanceChangeTx_IdempotentWithinTx(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)
	start := readUserBalance(t, userID)
	tx := testTx(t)

	cmd := billing.BalanceChangeCommand{
		UserID: userID, Delta: 40, EntryType: billing.EntryTypeRecharge,
		SourceType: billing.SourceTypePaymentOrder, SourceID: ptrInt64(time.Now().UnixNano()%1_000_000_000 + 3),
		Description: "dup in tx",
	}
	first, err := repo.ApplyBalanceChangeTx(ctx, tx, cmd)
	require.NoError(t, err)
	require.True(t, first.Applied)

	second, err := repo.ApplyBalanceChangeTx(ctx, tx, cmd)
	require.NoError(t, err)
	require.False(t, second.Applied, "same source within tx must be a no-op")
	require.Equal(t, first.LedgerID, second.LedgerID)

	var bal float64
	require.NoError(t, tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1", userID).Scan(&bal))
	require.Equal(t, start+40, bal, "credited exactly once")
}

func TestApplyBalanceChangeTx_InsufficientBalance(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)
	start := readUserBalance(t, userID)
	tx := testTx(t)

	_, err := repo.ApplyBalanceChangeTx(ctx, tx, billing.BalanceChangeCommand{
		UserID: userID, Delta: -(start + 1), EntryType: billing.EntryTypeAPIUsage,
		SourceType: billing.SourceTypeUsageLog, Description: "overdraft in tx",
	})
	require.ErrorIs(t, err, billing.ErrInsufficientBalance)

	var bal float64
	require.NoError(t, tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1", userID).Scan(&bal))
	require.Equal(t, start, bal, "rejected debit must not change balance")
}

func TestApplyBalanceChangeTx_UserNotFound(t *testing.T) {
	tx := testTx(t)
	_, err := newLedgerRepoForTest().ApplyBalanceChangeTx(context.Background(), tx, billing.BalanceChangeCommand{
		UserID: 999_999_999, Delta: 10, EntryType: billing.EntryTypeRecharge,
		SourceType: billing.SourceTypePaymentOrder,
	})
	require.ErrorIs(t, err, billing.ErrLedgerUserNotFound)
}

func TestApplyBalanceChange_GiftDelta(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)

	var startBal, startGift float64
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		"SELECT balance, COALESCE(gift_balance, 0) FROM users WHERE id = $1", userID).Scan(&startBal, &startGift))

	// OAuth-first-bind style grant: credit both balance and gift_balance atomically.
	res, err := repo.ApplyBalanceChange(ctx, billing.BalanceChangeCommand{
		UserID: userID, Delta: 50, GiftDelta: 50,
		EntryType: billing.EntryTypeOAuthBindBonus, SourceType: billing.SourceTypeOAuthBinding,
		SourceID: ptrInt64(time.Now().UnixNano()%1_000_000_000 + 7), Description: "first bind grant",
	})
	require.NoError(t, err)
	require.True(t, res.Applied)
	require.Equal(t, startBal+50, res.BalanceAfter)

	var bal, gift float64
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		"SELECT balance, COALESCE(gift_balance, 0) FROM users WHERE id = $1", userID).Scan(&bal, &gift))
	require.Equal(t, startBal+50, bal, "balance credited")
	require.Equal(t, startGift+50, gift, "gift_balance credited atomically")
}

// TestApplyBalanceChange_RedeemStyleInvariant locks in the dc7fe1d97 regression fix:
// a redeem-code-style change (Delta>0, GiftDelta=0, UpdateRecharged=true) must credit
// paid_balance (not gift_balance), preserve the balance = paid + gift invariant, and
// bump total_recharged. Previously the primitive touched only balance + gift_balance,
// leaving paid_balance frozen and total_recharged lost — breaking the invariant and
// under-reporting lifetime recharges.
func TestApplyBalanceChange_RedeemStyleInvariant(t *testing.T) {
	ctx := context.Background()
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)

	var startBal, startPaid, startGift, startRecharged float64
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		`SELECT balance, COALESCE(paid_balance,0), COALESCE(gift_balance,0), COALESCE(total_recharged,0) FROM users WHERE id = $1`, userID).
		Scan(&startBal, &startPaid, &startGift, &startRecharged))

	const amount = 30.0
	res, err := repo.ApplyBalanceChange(ctx, billing.BalanceChangeCommand{
		UserID:          userID,
		Delta:           amount,
		GiftDelta:       0, // redeem codes go to paid_balance, not gift_balance
		EntryType:       billing.EntryTypeRedeem,
		SourceType:      billing.SourceTypeRedeemCode,
		SourceID:        ptrInt64(time.Now().UnixNano()%1_000_000_000 + 91),
		Description:     "redeem code",
		UpdateRecharged: true,
	})
	require.NoError(t, err)
	require.True(t, res.Applied)
	require.Equal(t, startBal+amount, res.BalanceAfter)

	var bal, paid, gift, recharged float64
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		`SELECT balance, COALESCE(paid_balance,0), COALESCE(gift_balance,0), COALESCE(total_recharged,0) FROM users WHERE id = $1`, userID).
		Scan(&bal, &paid, &gift, &recharged))

	// paid_balance absorbs the full delta; gift_balance untouched.
	require.Equal(t, startPaid+amount, paid, "paid_balance must be credited for GiftDelta=0")
	require.Equal(t, startGift, gift, "gift_balance must be untouched for GiftDelta=0")
	// Invariant: balance == paid_balance + gift_balance.
	require.Equal(t, paid+gift, bal, "balance invariant balance = paid_balance + gift_balance must hold")
	// total_recharged bumps for positive redeem amounts.
	require.Equal(t, startRecharged+amount, recharged, "total_recharged must bump for UpdateRecharged positive delta")
}

func TestApplyBalanceChangeCtx_ParticipatesInEntTx(t *testing.T) {
	repo := newLedgerRepoForTest()
	userID := createLedgerTestUser(t)
	start := readUserBalance(t, userID)

	// Open an ent tx and place it in ctx (mirrors the OAuth first-bind flow).
	tx, err := integrationEntClient.Tx(context.Background())
	require.NoError(t, err)
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	ctx := dbent.NewTxContext(context.Background(), tx)

	res, err := repo.ApplyBalanceChangeCtx(ctx, billing.BalanceChangeCommand{
		UserID: userID, Delta: 25, EntryType: billing.EntryTypeOAuthBindBonus,
		SourceType: billing.SourceTypeOAuthBinding, Description: "ctx-dispatch grant",
	})
	require.NoError(t, err, "should derive an executor from the ent tx and apply")
	require.True(t, res.Applied)
	require.Equal(t, start+25, res.BalanceAfter)

	// Not yet committed: the committed balance is unchanged outside the tx.
	require.Equal(t, start, readUserBalance(t, userID), "change must be scoped to the uncommitted ent tx")

	require.NoError(t, tx.Commit())
	committed = true

	// After commit, both balance and ledger row are visible.
	require.Equal(t, start+25, readUserBalance(t, userID))
	_, total, err := repo.ListByUser(context.Background(), userID, billing.BalanceLedgerFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
}
