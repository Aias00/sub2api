package payment

import (
	"strings"
	"testing"
)

func TestFormatGatewayRefundAmount(t *testing.T) {
	t.Parallel()

	if got := FormatGatewayRefundAmount(12.345, "KWD"); got != "12.345" {
		t.Fatalf("FormatGatewayRefundAmount KWD = %q, want 12.345", got)
	}
	if got := FormatGatewayRefundAmount(52, "JPY"); got != "52" {
		t.Fatalf("FormatGatewayRefundAmount JPY = %q, want 52", got)
	}
}

func TestValidateRefundProviderResponse(t *testing.T) {
	t.Parallel()

	for _, status := range []string{ProviderStatusSuccess, ProviderStatusRefunded, ProviderStatusPending} {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()

			if err := ValidateRefundProviderResponse(&RefundResponse{Status: status}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	tests := []struct {
		name    string
		resp    *RefundResponse
		wantErr string
	}{
		{name: "nil response", wantErr: "missing"},
		{name: "failed response", resp: &RefundResponse{Status: ProviderStatusFailed}, wantErr: "failed"},
		{name: "unknown response", resp: &RefundResponse{Status: "processing"}, wantErr: "unknown status"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateRefundProviderResponse(tt.resp)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRefundResponseID(t *testing.T) {
	t.Parallel()

	if got := RefundResponseID(&RefundResponse{RefundID: " rf_123 "}); got != "rf_123" {
		t.Fatalf("RefundResponseID = %q, want rf_123", got)
	}
	if got := RefundResponseID(nil); got != "" {
		t.Fatalf("RefundResponseID(nil) = %q, want empty", got)
	}
}
