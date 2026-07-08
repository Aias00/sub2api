package payment

import "testing"

// ---------------------------------------------------------------------------
// SplitTypes
// ---------------------------------------------------------------------------

func TestSplitTypes(t *testing.T) {
	if got := SplitTypes(""); got != nil {
		t.Fatalf("empty string must yield nil, got %v", got)
	}
	got := SplitTypes("alipay, wxpay ,, creem ")
	want := []string{"alipay", "wxpay", "creem"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Currency helpers
// ---------------------------------------------------------------------------

func TestCurrencyMinorUnitAndConversion(t *testing.T) {
	if got := CurrencyMinorUnit("CNY"); got != 2 {
		t.Fatalf("CNY minor unit = %d, want 2", got)
	}
	if got := MinorUnitToAmount(100, "CNY"); got != 1.0 {
		t.Fatalf("MinorUnitToAmount(100,CNY) = %v, want 1.0", got)
	}
	if got := MinorUnitToAmount(1, "CNY"); got != 0.01 {
		t.Fatalf("MinorUnitToAmount(1,CNY) = %v, want 0.01", got)
	}
}

func TestProviderConfigCurrency_Pure(t *testing.T) {
	// Non-stripe/airwallex providers always use the default currency.
	if got := ProviderConfigCurrency(TypeCreem, map[string]string{"currency": "USD"}); got != DefaultPaymentCurrency {
		t.Fatalf("non-stripe provider must use default currency, got %q", got)
	}
	// Stripe honors a valid configured currency.
	if got := ProviderConfigCurrency(TypeStripe, map[string]string{"currency": "USD"}); got != "USD" {
		t.Fatalf("stripe should honor configured currency, got %q", got)
	}
	// Stripe with invalid currency falls back to default.
	if got := ProviderConfigCurrency(TypeStripe, map[string]string{"currency": "US1"}); got != DefaultPaymentCurrency {
		t.Fatalf("invalid stripe currency should fall back to default, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Creem product resolution
// ---------------------------------------------------------------------------

func TestResolveCreemRechargeProduct_Pure(t *testing.T) {
	products := []RechargeProduct{
		{ID: "p1", CreemProductID: "creem_p1"},
		{ID: "p2", CreemProductID: ""}, // not configured for creem
	}

	// Non-creem payment type → no-op (nil, nil).
	if p, err := ResolveCreemRechargeProduct("alipay", OrderTypeBalance, "p1", products); p != nil || err != nil {
		t.Fatalf("non-creem must be no-op, got p=%v err=%v", p, err)
	}

	// Creem + subscription order → no-op.
	if p, err := ResolveCreemRechargeProduct(TypeCreem, OrderTypeSubscription, "p1", products); p != nil || err != nil {
		t.Fatalf("creem subscription must be no-op, got p=%v err=%v", p, err)
	}

	// Creem recharge, empty selection → error.
	if _, err := ResolveCreemRechargeProduct(TypeCreem, OrderTypeBalance, "  ", products); err == nil {
		t.Fatal("empty product selection must error")
	}

	// Creem recharge, valid product → returns it.
	p, err := ResolveCreemRechargeProduct(TypeCreem, OrderTypeBalance, "p1", products)
	if err != nil || p == nil || p.ID != "p1" {
		t.Fatalf("expected p1, got p=%v err=%v", p, err)
	}

	// Creem recharge, product exists but not configured for creem → error.
	if _, err := ResolveCreemRechargeProduct(TypeCreem, OrderTypeBalance, "p2", products); err == nil {
		t.Fatal("product without CreemProductID must error")
	}

	// Creem recharge, unknown product → error.
	if _, err := ResolveCreemRechargeProduct(TypeCreem, OrderTypeBalance, "nope", products); err == nil {
		t.Fatal("unknown product must error")
	}
}

func TestResolveCreemProviderProductID_Pure(t *testing.T) {
	// Non-creem → no-op.
	if id, err := ResolveCreemProviderProductID("alipay", "sub1", true, nil); id != "" || err != nil {
		t.Fatalf("non-creem must be no-op, got id=%q err=%v", id, err)
	}

	// Subscription plan with configured product id.
	if id, err := ResolveCreemProviderProductID(TypeCreem, "sub_prod", true, nil); err != nil || id != "sub_prod" {
		t.Fatalf("expected sub_prod, got id=%q err=%v", id, err)
	}

	// Subscription plan without product id → error.
	if _, err := ResolveCreemProviderProductID(TypeCreem, "  ", true, nil); err == nil {
		t.Fatal("subscription plan without creem product id must error")
	}

	// Recharge with nil product → error.
	if _, err := ResolveCreemProviderProductID(TypeCreem, "", false, nil); err == nil {
		t.Fatal("recharge without product must error")
	}

	// Recharge with configured product.
	rp := &RechargeProduct{ID: "p1", CreemProductID: "creem_p1"}
	if id, err := ResolveCreemProviderProductID(TypeCreem, "", false, rp); err != nil || id != "creem_p1" {
		t.Fatalf("expected creem_p1, got id=%q err=%v", id, err)
	}
}
