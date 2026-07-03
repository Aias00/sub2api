package payment

import "testing"

func TestOrderQueryReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		providerKey         string
		snapshotProviderKey string
		orderProviderKey    string
		orderPaymentType    string
		outTradeNo          string
		paymentTradeNo      string
		want                string
	}{
		{
			name:             "official wxpay uses out trade no",
			providerKey:      TypeWxpay,
			orderPaymentType: TypeWxpay,
			outTradeNo:       "sub2_out_trade_no",
			paymentTradeNo:   "wx-transaction-id",
			want:             "sub2_out_trade_no",
		},
		{
			name:             "official alipay direct uses out trade no",
			providerKey:      TypeAlipayDirect,
			orderPaymentType: TypeAlipayDirect,
			outTradeNo:       "sub2_out_trade_no",
			paymentTradeNo:   "alipay-trade-no",
			want:             "sub2_out_trade_no",
		},
		{
			name:             "stripe uses payment trade no when present",
			providerKey:      TypeStripe,
			orderPaymentType: TypeStripe,
			outTradeNo:       "sub2_out_trade_no",
			paymentTradeNo:   "pi_123",
			want:             "pi_123",
		},
		{
			name:             "stripe falls back to out trade no",
			providerKey:      TypeStripe,
			orderPaymentType: TypeStripe,
			outTradeNo:       "sub2_out_trade_no",
			want:             "sub2_out_trade_no",
		},
		{
			name:                "snapshot provider key is used when runtime provider is missing",
			snapshotProviderKey: TypeEasyPay,
			orderPaymentType:    TypeAlipay,
			outTradeNo:          "sub2_out_trade_no",
			paymentTradeNo:      "upstream-trade-no",
			want:                "sub2_out_trade_no",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := OrderQueryReference(
				tt.providerKey,
				tt.snapshotProviderKey,
				tt.orderProviderKey,
				tt.orderPaymentType,
				tt.outTradeNo,
				tt.paymentTradeNo,
			)
			if got != tt.want {
				t.Fatalf("OrderQueryReference() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShouldPersistUpstreamTradeNo(t *testing.T) {
	t.Parallel()

	if !ShouldPersistUpstreamTradeNo("sub2_out_trade_no", "upstream-trade-no", "") {
		t.Fatal("new upstream trade no should be persisted")
	}
	if ShouldPersistUpstreamTradeNo("sub2_out_trade_no", "", "") {
		t.Fatal("blank upstream trade no should not be persisted")
	}
	if ShouldPersistUpstreamTradeNo("sub2_out_trade_no", "existing-trade-no", "existing-trade-no") {
		t.Fatal("existing trade no should not be persisted again")
	}
	if ShouldPersistUpstreamTradeNo("sub2_out_trade_no", "sub2_out_trade_no", "") {
		t.Fatal("query reference should not be persisted as upstream trade no")
	}
}

func TestAllowsRegistryFallback(t *testing.T) {
	t.Parallel()

	if !AllowsRegistryFallback(false, "", "") {
		t.Fatal("legacy order without pinned provider state should allow fallback")
	}
	if AllowsRegistryFallback(true, "", "") {
		t.Fatal("snapshot order should not allow registry fallback")
	}
	if AllowsRegistryFallback(false, "12", "") {
		t.Fatal("provider instance order should not allow registry fallback")
	}
	if AllowsRegistryFallback(false, "", TypeStripe) {
		t.Fatal("provider key order should not allow registry fallback")
	}
}

func TestFallbackProviderKey(t *testing.T) {
	t.Parallel()

	if got := FallbackProviderKey(TypeEasyPay, TypeAlipay); got != TypeEasyPay {
		t.Fatalf("FallbackProviderKey registry = %q, want %q", got, TypeEasyPay)
	}
	if got := FallbackProviderKey("", TypeAlipayDirect); got != TypeAlipay {
		t.Fatalf("FallbackProviderKey base type = %q, want %q", got, TypeAlipay)
	}
}

func TestIsRefundStatus(t *testing.T) {
	t.Parallel()

	if !IsRefundStatus(OrderStatusRefundPending) {
		t.Fatal("refund pending should be a refund status")
	}
	if IsRefundStatus(OrderStatusPaid) {
		t.Fatal("paid should not be a refund status")
	}
}
