//go:build unit

package repository

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestRefundBalanceReservationRestoresGiftTailBeforePaid(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`(?s)UPDATE users\s+SET balance = balance \+ \$1 \+ \$2,\s+paid_balance = paid_balance \+ \$1,\s+gift_balance = gift_balance \+ \$2`).
		WithArgs(1.0, 3.0, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(12.0))

	refunded, err := refundBalanceReservation(context.Background(), db, 42, 4, 2, 3)

	require.NoError(t, err)
	require.Equal(t, 12.0, refunded.BalanceSnapshot)
	require.Equal(t, 1.0, refunded.Paid)
	require.Equal(t, 3.0, refunded.Gift)
	require.NoError(t, mock.ExpectationsWereMet())
}
