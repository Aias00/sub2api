package repository

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"github.com/Aias00/cloudbase/internal/billing"
	"github.com/Aias00/cloudbase/internal/identity"
)

type userBalanceReservation struct {
	BalanceSnapshot float64
	Paid            float64
	Gift            float64
}

func reserveUserBalanceWithComponents(ctx context.Context, q sqlExecutor, userID int64, amount float64, insufficientErr error) (userBalanceReservation, error) {
	if amount <= 0 {
		return userBalanceReservation{}, nil
	}

	var reservation userBalanceReservation
	if err := scanSingleRow(ctx, q, `
		WITH current_balance AS (
			SELECT balance, paid_balance, gift_balance
			FROM users
			WHERE id = $2 AND deleted_at IS NULL AND balance >= $1
			FOR UPDATE
		),
		updated_balance AS (
			UPDATE users
			SET balance = balance - $1,
				gift_balance = LEAST(gift_balance, GREATEST(balance - $1, 0)),
				paid_balance = (balance - $1) - LEAST(gift_balance, GREATEST(balance - $1, 0)),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND deleted_at IS NULL AND balance >= $1
			RETURNING balance, paid_balance, gift_balance
		)
		SELECT updated_balance.balance,
			GREATEST(current_balance.paid_balance - updated_balance.paid_balance, 0),
			GREATEST(current_balance.gift_balance - updated_balance.gift_balance, 0)
		FROM current_balance, updated_balance
	`, []any{amount, userID}, &reservation.BalanceSnapshot, &reservation.Paid, &reservation.Gift); err == nil {
		return reservation, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return userBalanceReservation{}, err
	}

	var exists bool
	if err := scanSingleRow(ctx, q, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)
	`, []any{userID}, &exists); err != nil {
		return userBalanceReservation{}, err
	}
	if !exists {
		return userBalanceReservation{}, identity.ErrUserNotFound
	}
	if insufficientErr != nil {
		return userBalanceReservation{}, insufficientErr
	}
	return userBalanceReservation{}, billing.ErrInsufficientBalance
}

func creditBalanceComponents(ctx context.Context, q sqlExecutor, userID int64, paidAmount, giftAmount float64) (float64, error) {
	paidAmount = math.Max(paidAmount, 0)
	giftAmount = math.Max(giftAmount, 0)
	if paidAmount == 0 && giftAmount == 0 {
		return 0, nil
	}

	var balanceSnapshot float64
	if err := scanSingleRow(ctx, q, `
		UPDATE users
		SET balance = balance + $1 + $2,
			paid_balance = paid_balance + $1,
			gift_balance = gift_balance + $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING balance
	`, []any{paidAmount, giftAmount, userID}, &balanceSnapshot); err != nil {
		return 0, err
	}
	return balanceSnapshot, nil
}

func refundBalanceReservation(ctx context.Context, q sqlExecutor, userID int64, amount float64, reservedPaid float64, reservedGift float64) (userBalanceReservation, error) {
	if amount <= 0 {
		return userBalanceReservation{}, nil
	}
	giftRefund := math.Min(amount, math.Max(reservedGift, 0))
	paidRefund := math.Min(math.Max(amount-giftRefund, 0), math.Max(reservedPaid, 0))
	balanceSnapshot, err := creditBalanceComponents(ctx, q, userID, paidRefund, giftRefund)
	if err != nil {
		return userBalanceReservation{}, err
	}
	return userBalanceReservation{
		BalanceSnapshot: balanceSnapshot,
		Paid:            paidRefund,
		Gift:            giftRefund,
	}, nil
}

func refundFullBalanceReservation(ctx context.Context, q sqlExecutor, userID int64, reservedPaid float64, reservedGift float64) (userBalanceReservation, error) {
	if reservedPaid <= 0 && reservedGift <= 0 {
		return userBalanceReservation{}, nil
	}
	balanceSnapshot, err := creditBalanceComponents(ctx, q, userID, reservedPaid, reservedGift)
	if err != nil {
		return userBalanceReservation{}, err
	}
	return userBalanceReservation{
		BalanceSnapshot: balanceSnapshot,
		Paid:            reservedPaid,
		Gift:            reservedGift,
	}, nil
}

func mergeBalanceReservation(base userBalanceReservation, delta userBalanceReservation) userBalanceReservation {
	base.BalanceSnapshot = delta.BalanceSnapshot
	base.Paid += delta.Paid
	base.Gift += delta.Gift
	return base
}

func reduceBalanceReservation(base userBalanceReservation, refunded userBalanceReservation) userBalanceReservation {
	base.BalanceSnapshot = refunded.BalanceSnapshot
	base.Paid = math.Max(base.Paid-refunded.Paid, 0)
	base.Gift = math.Max(base.Gift-refunded.Gift, 0)
	return base
}
