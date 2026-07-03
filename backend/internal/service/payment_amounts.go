package service

import "github.com/Aias00/cloudbase/internal/payment"

const defaultBalanceRechargeMultiplier = payment.DefaultBalanceRechargeMultiplier

func normalizeBalanceRechargeMultiplier(multiplier float64) float64 {
	return payment.NormalizeBalanceRechargeMultiplier(multiplier)
}

func calculateCreditedBalance(paymentAmount, multiplier float64) float64 {
	return payment.CalculateCreditedBalance(paymentAmount, multiplier)
}

func calculateGatewayRefundAmount(orderAmount, payAmount, refundAmount float64, currency string) float64 {
	return payment.CalculateGatewayRefundAmount(orderAmount, payAmount, refundAmount, currency, paymentAmountToleranceForCurrency(currency))
}
