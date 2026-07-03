package payment

import (
	"fmt"
	"strings"
)

func FormatGatewayRefundAmount(amount float64, currency string) string {
	return FormatAmountForCurrency(amount, currency)
}

func ValidateRefundProviderResponse(resp *RefundResponse) error {
	if resp == nil {
		return fmt.Errorf("payment refund response missing")
	}
	status := strings.TrimSpace(resp.Status)
	switch status {
	case ProviderStatusSuccess, ProviderStatusRefunded, ProviderStatusPending:
		return nil
	case ProviderStatusFailed:
		return fmt.Errorf("payment refund failed: status %s", status)
	default:
		return fmt.Errorf("payment refund returned unknown status: %s", status)
	}
}

func RefundResponseID(resp *RefundResponse) string {
	if resp == nil {
		return ""
	}
	return strings.TrimSpace(resp.RefundID)
}
