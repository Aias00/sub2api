//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Aias00/cloudbase/internal/billing"
)

// seedBalanceComponents is intentionally inlined per-test via integrationDB.

func TestReserveAndRefund_ImageWorkspace_WriteLedger(t *testing.T) {
	ctx := context.Background()
	userID := createLedgerTestUser(t)

	// Seed consistent components: balance = paid + gift.
	_, err := integrationDB.ExecContext(ctx,
		`UPDATE users SET balance = 100, paid_balance = 100, gift_balance = 0 WHERE id = $1`, userID)
	require.NoError(t, err)

	tx := testTx(t)

	// Reserve 30 (image workspace) → deduction ledger row.
	reservation, err := reserveImageWorkspaceBalance(ctx, tx, userID, 30)
	require.NoError(t, err)
	require.InDelta(t, 70, reservation.BalanceSnapshot, 0.0001)
	require.InDelta(t, 30, reservation.Paid, 0.0001)
	require.InDelta(t, 0, reservation.Gift, 0.0001)

	var count int
	var amount, before, after float64
	require.NoError(t, tx.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(amount),0), COALESCE(MAX(balance_before),0), COALESCE(MAX(balance_after),0)
		 FROM user_balance_ledger WHERE user_id = $1 AND entry_type = $2`,
		userID, string(billing.EntryTypeImageWorkspace)).Scan(&count, &amount, &before, &after))
	require.Equal(t, 1, count)
	require.InDelta(t, -30, amount, 0.0001)
	require.InDelta(t, 100, before, 0.0001)
	require.InDelta(t, 70, after, 0.0001)

	// Full refund → positive reversal ledger row, balance restored.
	refunded, err := refundFullBalanceReservation(ctx, tx, userID, reservation.Paid, reservation.Gift, &imageWorkspaceBalanceLedger)
	require.NoError(t, err)
	require.InDelta(t, 100, refunded.BalanceSnapshot, 0.0001)

	// Two rows total; net movement reconciles to zero after full refund.
	var total int
	var net float64
	require.NoError(t, tx.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(amount),0)
		 FROM user_balance_ledger WHERE user_id = $1 AND entry_type = $2`,
		userID, string(billing.EntryTypeImageWorkspace)).Scan(&total, &net))
	require.Equal(t, 2, total)
	require.InDelta(t, 0, net, 0.0001, "reserve + full refund must net to zero")
}

func TestReserveWeChatExport_WritesLedger(t *testing.T) {
	ctx := context.Background()
	userID := createLedgerTestUser(t)

	_, err := integrationDB.ExecContext(ctx,
		`UPDATE users SET balance = 50, paid_balance = 20, gift_balance = 30 WHERE id = $1`, userID)
	require.NoError(t, err)

	tx := testTx(t)

	// Reserve 40: gift-first consumption (30 gift + 10 paid).
	reservation, err := reserveWeChatExportBalance(ctx, tx, userID, 40)
	require.NoError(t, err)
	require.InDelta(t, 10, reservation.BalanceSnapshot, 0.0001)

	var count int
	var amount, after float64
	require.NoError(t, tx.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(amount),0), COALESCE(MAX(balance_after),0)
		 FROM user_balance_ledger WHERE user_id = $1 AND entry_type = $2`,
		userID, string(billing.EntryTypeWechatExport)).Scan(&count, &amount, &after))
	require.Equal(t, 1, count)
	require.InDelta(t, -40, amount, 0.0001)
	require.InDelta(t, 10, after, 0.0001)
}
