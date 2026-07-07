package service

import (
	"math"

	"github.com/Aias00/cloudbase/internal/payment"
)

const defaultBalanceRechargeMultiplier = payment.DefaultBalanceRechargeMultiplier

func normalizeBalanceRechargeMultiplier(multiplier float64) float64 {
	return payment.NormalizeBalanceRechargeMultiplier(multiplier)
}

// normalizeSubscriptionUSDToCNYRate 将非法值归一为 0（换算关闭）。
// 与余额倍率不同，0 是合法状态：表示订阅保持 price 直付的存量行为。
func normalizeSubscriptionUSDToCNYRate(rate float64) float64 {
	if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 {
		return 0
	}
	return rate
}

func calculateCreditedBalance(paymentAmount, multiplier float64) float64 {
	return payment.CalculateCreditedBalance(paymentAmount, multiplier)
}

func calculateGatewayRefundAmount(orderAmount, payAmount, refundAmount float64, currency string) float64 {
	return payment.CalculateGatewayRefundAmount(orderAmount, payAmount, refundAmount, currency, paymentAmountToleranceForCurrency(currency))
}
