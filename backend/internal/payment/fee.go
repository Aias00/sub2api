package payment

import (
	"fmt"
	"math"
	"strconv"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

func CalculatePayAmount(rechargeAmount float64, feeRate float64) string {
	return CalculatePayAmountForCurrency(rechargeAmount, feeRate, DefaultPaymentCurrency)
}

// CalculatePayAmountForCurrency 按币种精度计算应付金额，手续费向上取整到该币种最小支付单位。
func CalculatePayAmountForCurrency(rechargeAmount float64, feeRate float64, currency string) string {
	fractionDigits := int32(CurrencyMaxFractionDigits(currency))
	amount := decimal.NewFromFloat(rechargeAmount)
	if feeRate <= 0 {
		return amount.StringFixed(fractionDigits)
	}
	rate := decimal.NewFromFloat(feeRate)
	fee := amount.Mul(rate).Div(decimal.NewFromInt(100)).RoundUp(fractionDigits)
	return amount.Add(fee).StringFixed(fractionDigits)
}

func ValidateBalanceRechargeAmount(amount, minAmount, maxAmount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return infraerrors.BadRequest("INVALID_AMOUNT", "amount must be a positive number")
	}
	if (minAmount > 0 && amount < minAmount) || (maxAmount > 0 && amount > maxAmount) {
		return infraerrors.BadRequest("INVALID_AMOUNT", "amount out of range").
			WithMetadata(map[string]string{"min": fmt.Sprintf("%.2f", minAmount), "max": fmt.Sprintf("%.2f", maxAmount)})
	}
	return nil
}

func ValidateAmountCurrency(amount float64, currency string) error {
	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)
	return ValidatePayAmountCurrency(amountStr, currency)
}

func ValidatePayAmountCurrency(payAmount, currency string) error {
	if _, err := AmountToMinorUnit(payAmount, currency); err != nil {
		return infraerrors.BadRequest("INVALID_AMOUNT", err.Error()).
			WithMetadata(map[string]string{"currency": currency})
	}
	return nil
}

func CalculateCreateOrderPayAmount(limitAmount, feeRate float64, currency string) (string, float64, error) {
	if err := ValidateAmountCurrency(limitAmount, currency); err != nil {
		return "", 0, err
	}
	payAmountStr := CalculatePayAmountForCurrency(limitAmount, feeRate, currency)
	if err := ValidatePayAmountCurrency(payAmountStr, currency); err != nil {
		return "", 0, err
	}
	payAmount, err := strconv.ParseFloat(payAmountStr, 64)
	if err != nil {
		return "", 0, infraerrors.BadRequest("INVALID_AMOUNT", "invalid payment amount").
			WithMetadata(map[string]string{"currency": currency})
	}
	return payAmountStr, payAmount, nil
}
