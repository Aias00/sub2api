//go:build unit

package repository

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryUpdateBalanceNegativeRebalancesGiftBucket(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &userRepository{sql: db}
	mock.ExpectExec(`(?s)UPDATE users\s+SET balance = balance \+ \$1,\s+gift_balance = LEAST\(gift_balance, GREATEST\(balance \+ \$1, 0\)\),\s+paid_balance = \(balance \+ \$1\) - LEAST\(gift_balance, GREATEST\(balance \+ \$1, 0\)\),\s+updated_at = NOW\(\)\s+WHERE id = \$2 AND deleted_at IS NULL`).
		WithArgs(-5.0, int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpdateBalance(context.Background(), 42, -5))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryDeductBalanceRebalancesGiftBucket(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &userRepository{sql: db}
	mock.ExpectExec(`(?s)UPDATE users\s+SET balance = balance \+ \$1,\s+gift_balance = LEAST\(gift_balance, GREATEST\(balance \+ \$1, 0\)\),\s+paid_balance = \(balance \+ \$1\) - LEAST\(gift_balance, GREATEST\(balance \+ \$1, 0\)\),\s+updated_at = NOW\(\)\s+WHERE id = \$2 AND deleted_at IS NULL`).
		WithArgs(-5.0, int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.DeductBalance(context.Background(), 42, 5))
	require.NoError(t, mock.ExpectationsWereMet())
}
