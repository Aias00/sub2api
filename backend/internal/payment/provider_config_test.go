//go:build unit

package payment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProviderRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		providerKey string
		provider    string
		wantErr     bool
		errContains string
	}{
		{name: "valid easypay", providerKey: TypeEasyPay, provider: "MyProvider"},
		{name: "valid stripe", providerKey: TypeStripe, provider: "Stripe Provider"},
		{name: "valid airwallex", providerKey: TypeAirwallex, provider: "Airwallex Provider"},
		{name: "invalid provider key", providerKey: "invalid", provider: "Name", wantErr: true, errContains: "invalid provider key"},
		{name: "empty name", providerKey: TypeEasyPay, provider: "", wantErr: true, errContains: "provider name is required"},
		{name: "whitespace-only name", providerKey: TypeEasyPay, provider: "  ", wantErr: true, errContains: "provider name is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateProviderRequest(tc.providerKey, tc.provider)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestIsSensitiveProviderConfigField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		providerKey string
		field       string
		want        bool
	}{
		{TypeStripe, "secretKey", true},
		{TypeStripe, "webhookSecret", true},
		{TypeStripe, "SecretKey", true},
		{TypeStripe, "publishableKey", false},
		{TypeStripe, "currency", false},
		{TypeAlipay, "privateKey", true},
		{TypeAlipay, "appId", false},
		{TypeWxpay, "apiV3Key", true},
		{TypeWxpay, "publicKeyId", false},
		{TypeEasyPay, "pkey", true},
		{TypeEasyPay, "pid", false},
		{TypeAirwallex, "apiKey", true},
		{TypeAirwallex, "accountId", false},
		{"unknown", "secretKey", false},
	}

	for _, tc := range tests {
		t.Run(tc.providerKey+"/"+tc.field, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, IsSensitiveProviderConfigField(tc.providerKey, tc.field))
		})
	}
}

func TestHasPendingOrderProtectedConfigChange(t *testing.T) {
	t.Parallel()

	current := map[string]string{
		"secretKey": "old-secret",
		"currency":  "CNY",
		"note":      "old-note",
	}
	nextSafe := map[string]string{
		"secretKey": "old-secret",
		"currency":  "CNY",
		"note":      "new-note",
	}
	nextProtected := map[string]string{
		"secretKey": "old-secret",
		"currency":  "USD",
		"note":      "old-note",
	}

	if HasPendingOrderProtectedConfigChange(TypeStripe, current, nextSafe) {
		t.Fatal("safe config change should not be protected")
	}
	if !HasPendingOrderProtectedConfigChange(TypeStripe, current, nextProtected) {
		t.Fatal("currency change should be protected for stripe")
	}
}

func TestProviderConfigFieldValueCaseInsensitive(t *testing.T) {
	t.Parallel()

	cfg := map[string]string{"secretKey": "value"}
	if got := ProviderConfigFieldValue(cfg, "SECRETKEY"); got != "value" {
		t.Fatalf("ProviderConfigFieldValue = %q, want value", got)
	}
}
