package payment

import "testing"

func TestLegacyOrderMatchesProviderInstance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		orderPaymentType string
		source           ProviderInstanceSource
		want             bool
	}{
		{
			name:             "stripe order matches stripe provider",
			orderPaymentType: TypeCard,
			source:           ProviderInstanceSource{ProviderKey: TypeStripe, SupportedTypes: "card,link"},
			want:             true,
		},
		{
			name:             "stripe provider does not match alipay order",
			orderPaymentType: TypeAlipay,
			source:           ProviderInstanceSource{ProviderKey: TypeStripe, SupportedTypes: "alipay"},
			want:             false,
		},
		{
			name:             "same provider key matches",
			orderPaymentType: TypeWxpayDirect,
			source:           ProviderInstanceSource{ProviderKey: TypeWxpay},
			want:             true,
		},
		{
			name:             "aggregator supported type matches",
			orderPaymentType: TypeAlipayDirect,
			source:           ProviderInstanceSource{ProviderKey: TypeEasyPay, SupportedTypes: "wxpay,alipay"},
			want:             true,
		},
		{
			name:             "aggregator unsupported type does not match",
			orderPaymentType: TypeWxpay,
			source:           ProviderInstanceSource{ProviderKey: TypeEasyPay, SupportedTypes: TypeAlipay},
			want:             false,
		},
		{
			name:             "blank order payment type does not match",
			orderPaymentType: "",
			source:           ProviderInstanceSource{ProviderKey: TypeEasyPay, SupportedTypes: TypeAlipay},
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := LegacyOrderMatchesProviderInstance(tt.orderPaymentType, tt.source); got != tt.want {
				t.Fatalf("LegacyOrderMatchesProviderInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}
