package payment

import "strings"

func OrderQueryReference(providerKey, snapshotProviderKey, orderProviderKey, orderPaymentType, outTradeNo, paymentTradeNo string) string {
	resolvedProviderKey := strings.TrimSpace(providerKey)
	if resolvedProviderKey == "" {
		resolvedProviderKey = strings.TrimSpace(snapshotProviderKey)
	}
	if resolvedProviderKey == "" {
		resolvedProviderKey = strings.TrimSpace(orderProviderKey)
	}
	if resolvedProviderKey == "" {
		resolvedProviderKey = strings.TrimSpace(orderPaymentType)
	}

	switch GetBasePaymentType(resolvedProviderKey) {
	case TypeAlipay, TypeEasyPay, TypeWxpay:
		return strings.TrimSpace(outTradeNo)
	default:
		if tradeNo := strings.TrimSpace(paymentTradeNo); tradeNo != "" {
			return tradeNo
		}
		return strings.TrimSpace(outTradeNo)
	}
}

func ShouldPersistUpstreamTradeNo(queryRef, upstreamTradeNo, currentTradeNo string) bool {
	upstreamTradeNo = strings.TrimSpace(upstreamTradeNo)
	if upstreamTradeNo == "" {
		return false
	}
	if strings.EqualFold(upstreamTradeNo, strings.TrimSpace(currentTradeNo)) {
		return false
	}
	if strings.EqualFold(upstreamTradeNo, strings.TrimSpace(queryRef)) {
		return false
	}
	return true
}

func AllowsRegistryFallback(hasProviderSnapshot bool, providerInstanceID, providerKey string) bool {
	if hasProviderSnapshot {
		return false
	}
	if strings.TrimSpace(providerInstanceID) != "" {
		return false
	}
	if strings.TrimSpace(providerKey) != "" {
		return false
	}
	return true
}

func FallbackProviderKey(registryProviderKey, orderPaymentType string) string {
	if key := strings.TrimSpace(registryProviderKey); key != "" {
		return key
	}
	return strings.TrimSpace(GetBasePaymentType(strings.TrimSpace(orderPaymentType)))
}

func IsRefundStatus(status string) bool {
	switch status {
	case OrderStatusRefundRequested, OrderStatusRefunding, OrderStatusRefundPending, OrderStatusPartiallyRefunded, OrderStatusRefunded, OrderStatusRefundFailed:
		return true
	}
	return false
}
