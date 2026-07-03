package payment

import (
	"math"

	"github.com/shopspring/decimal"
)

const DefaultBalanceRechargeMultiplier = 1.0

func NormalizeBalanceRechargeMultiplier(multiplier float64) float64 {
	if math.IsNaN(multiplier) || math.IsInf(multiplier, 0) || multiplier <= 0 {
		return DefaultBalanceRechargeMultiplier
	}
	return multiplier
}

func CalculateCreditedBalance(paymentAmount, multiplier float64) float64 {
	return decimal.NewFromFloat(paymentAmount).
		Mul(decimal.NewFromFloat(NormalizeBalanceRechargeMultiplier(multiplier))).
		Round(2).
		InexactFloat64()
}

func PaymentAmountToleranceForCurrency(currency string, defaultTolerance float64) float64 {
	minorUnit := CurrencyMinorUnit(currency)
	if minorUnit <= 2 {
		return defaultTolerance
	}
	return math.Pow10(-minorUnit) / 2
}

func CalculateGatewayRefundAmount(orderAmount, payAmount, refundAmount float64, currency string, fullRefundTolerance float64) float64 {
	if orderAmount <= 0 || payAmount <= 0 || refundAmount <= 0 {
		return 0
	}
	fractionDigits := int32(CurrencyMaxFractionDigits(currency))
	if math.Abs(refundAmount-orderAmount) <= fullRefundTolerance {
		return decimal.NewFromFloat(payAmount).Round(fractionDigits).InexactFloat64()
	}
	return decimal.NewFromFloat(payAmount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(orderAmount)).
		Round(fractionDigits).
		InexactFloat64()
}

func IsValidProviderAmount(amount float64) bool {
	return amount > 0 && !math.IsNaN(amount) && !math.IsInf(amount, 0)
}
