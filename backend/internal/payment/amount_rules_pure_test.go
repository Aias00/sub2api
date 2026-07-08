package payment

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Amount conversion (CNY: 1 yuan = 100 fen). Fund-safety critical.
// ---------------------------------------------------------------------------

func TestYuanToFen(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"1.00", 100, false},
		{"0.01", 1, false},
		{"1.23", 123, false},
		{"100", 10000, false},
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, tc := range cases {
		got, err := YuanToFen(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("YuanToFen(%q) expected error, got %d", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("YuanToFen(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("YuanToFen(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestFenToYuan(t *testing.T) {
	if got := FenToYuan(100); got != 1.0 {
		t.Errorf("FenToYuan(100) = %v, want 1.0", got)
	}
	if got := FenToYuan(1); got != 0.01 {
		t.Errorf("FenToYuan(1) = %v, want 0.01", got)
	}
	if got := FenToYuan(123); got != 1.23 {
		t.Errorf("FenToYuan(123) = %v, want 1.23", got)
	}
}

func TestYuanFenRoundTrip(t *testing.T) {
	for _, s := range []string{"1.00", "0.01", "9.99", "1234.56"} {
		fen, err := YuanToFen(s)
		if err != nil {
			t.Fatalf("YuanToFen(%q): %v", s, err)
		}
		if back := FenToYuan(fen); FenToYuan(fen) != back {
			t.Fatalf("round trip unstable for %q", s)
		}
	}
}

// ---------------------------------------------------------------------------
// Balance recharge multiplier / credited balance
// ---------------------------------------------------------------------------

func TestNormalizeBalanceRechargeMultiplier(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{1.5, 1.5},
		{2, 2},
		{0, DefaultBalanceRechargeMultiplier},
		{-1, DefaultBalanceRechargeMultiplier},
		{math.NaN(), DefaultBalanceRechargeMultiplier},
		{math.Inf(1), DefaultBalanceRechargeMultiplier},
	}
	for _, tc := range cases {
		if got := NormalizeBalanceRechargeMultiplier(tc.in); got != tc.want {
			t.Errorf("NormalizeBalanceRechargeMultiplier(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestCalculateCreditedBalance(t *testing.T) {
	if got := CalculateCreditedBalance(100, 1.5); got != 150 {
		t.Errorf("100 * 1.5 = %v, want 150", got)
	}
	// Invalid multiplier falls back to 1.0 → credited == paid.
	if got := CalculateCreditedBalance(100, 0); got != 100 {
		t.Errorf("invalid multiplier must credit paid amount, got %v", got)
	}
	if got := CalculateCreditedBalance(100, -5); got != 100 {
		t.Errorf("negative multiplier must credit paid amount, got %v", got)
	}
	// Rounds to 2 decimals.
	if got := CalculateCreditedBalance(10, 1.005); got != 10.05 {
		t.Errorf("expected rounding to 2 decimals, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// IsValidProviderAmount
// ---------------------------------------------------------------------------

func TestIsValidProviderAmount(t *testing.T) {
	valid := []float64{0.01, 1, 999999}
	for _, v := range valid {
		if !IsValidProviderAmount(v) {
			t.Errorf("IsValidProviderAmount(%v) = false, want true", v)
		}
	}
	invalid := []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, v := range invalid {
		if IsValidProviderAmount(v) {
			t.Errorf("IsValidProviderAmount(%v) = true, want false", v)
		}
	}
}

// ---------------------------------------------------------------------------
// Refund amount (gateway) — full vs proportional partial refund
// ---------------------------------------------------------------------------

func TestCalculateGatewayRefundAmount(t *testing.T) {
	const tol = 0.01

	// Guard: non-positive inputs yield 0 (never refund garbage).
	if got := CalculateGatewayRefundAmount(0, 95, 100, "CNY", tol); got != 0 {
		t.Errorf("zero order amount must yield 0, got %v", got)
	}
	if got := CalculateGatewayRefundAmount(100, 0, 100, "CNY", tol); got != 0 {
		t.Errorf("zero pay amount must yield 0, got %v", got)
	}
	if got := CalculateGatewayRefundAmount(100, 95, 0, "CNY", tol); got != 0 {
		t.Errorf("zero refund amount must yield 0, got %v", got)
	}

	// Full refund (refund ≈ order within tolerance) → return the full paid amount.
	if got := CalculateGatewayRefundAmount(100, 95, 100, "CNY", tol); got != 95 {
		t.Errorf("full refund must return full pay amount, got %v", got)
	}

	// Partial refund → proportional: payAmount * refund/order.
	if got := CalculateGatewayRefundAmount(100, 95, 50, "CNY", tol); got != 47.5 {
		t.Errorf("partial refund = 95*50/100 = 47.5, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Currency tolerance
// ---------------------------------------------------------------------------

func TestPaymentAmountToleranceForCurrency(t *testing.T) {
	// CNY has minor unit 2 → uses default tolerance.
	if got := PaymentAmountToleranceForCurrency("CNY", 0.01); got != 0.01 {
		t.Errorf("CNY tolerance should be default 0.01, got %v", got)
	}
	// Unknown currency defaults to two-decimal unit → default tolerance.
	if got := PaymentAmountToleranceForCurrency("XYZ", 0.02); got != 0.02 {
		t.Errorf("unknown currency should use default tolerance, got %v", got)
	}
}
