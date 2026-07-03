//go:build unit

package payment

import (
	"math"
	"testing"
)

func TestNormalizeBalanceRechargeMultiplier(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{name: "positive", in: 1.2, want: 1.2},
		{name: "zero defaults", in: 0, want: DefaultBalanceRechargeMultiplier},
		{name: "negative defaults", in: -1, want: DefaultBalanceRechargeMultiplier},
		{name: "nan defaults", in: math.NaN(), want: DefaultBalanceRechargeMultiplier},
		{name: "inf defaults", in: math.Inf(1), want: DefaultBalanceRechargeMultiplier},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeBalanceRechargeMultiplier(tt.in); got != tt.want {
				t.Fatalf("NormalizeBalanceRechargeMultiplier(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCalculateCreditedBalance(t *testing.T) {
	if got := CalculateCreditedBalance(10, 0.14); got != 1.4 {
		t.Fatalf("CalculateCreditedBalance = %v, want 1.4", got)
	}
	if got := CalculateCreditedBalance(5, 10); got != 50 {
		t.Fatalf("CalculateCreditedBalance = %v, want 50", got)
	}
	if got := CalculateCreditedBalance(5, 0); got != 5 {
		t.Fatalf("CalculateCreditedBalance default multiplier = %v, want 5", got)
	}
}

func TestPaymentAmountToleranceForCurrency(t *testing.T) {
	const defaultTolerance = 0.01
	if got := PaymentAmountToleranceForCurrency("CNY", defaultTolerance); got != defaultTolerance {
		t.Fatalf("CNY tolerance = %v, want %v", got, defaultTolerance)
	}
	if got := PaymentAmountToleranceForCurrency("JPY", defaultTolerance); got != defaultTolerance {
		t.Fatalf("JPY tolerance = %v, want %v", got, defaultTolerance)
	}
	if got := PaymentAmountToleranceForCurrency("KWD", defaultTolerance); math.Abs(got-0.0005) > 1e-12 {
		t.Fatalf("KWD tolerance = %v, want 0.0005", got)
	}
}

func TestCalculateGatewayRefundAmount(t *testing.T) {
	const defaultTolerance = 0.01
	cases := []struct {
		name         string
		orderAmount  float64
		payAmount    float64
		refundAmount float64
		currency     string
		want         float64
	}{
		{name: "three decimal partial", orderAmount: 100, payAmount: 12.345, refundAmount: 50, currency: "KWD", want: 6.173},
		{name: "three decimal full", orderAmount: 100, payAmount: 12.345, refundAmount: 100, currency: "KWD", want: 12.345},
		{name: "zero decimal partial", orderAmount: 100, payAmount: 103, refundAmount: 50, currency: "JPY", want: 52},
		{name: "invalid order", orderAmount: 0, payAmount: 103, refundAmount: 50, currency: "JPY", want: 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateGatewayRefundAmount(tt.orderAmount, tt.payAmount, tt.refundAmount, tt.currency, PaymentAmountToleranceForCurrency(tt.currency, defaultTolerance))
			if math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("CalculateGatewayRefundAmount = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidProviderAmount(t *testing.T) {
	t.Parallel()

	if !IsValidProviderAmount(0.01) {
		t.Fatal("positive amount should be valid")
	}
	for _, amount := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		amount := amount
		t.Run("invalid", func(t *testing.T) {
			t.Parallel()

			if IsValidProviderAmount(amount) {
				t.Fatalf("IsValidProviderAmount(%v) = true, want false", amount)
			}
		})
	}
}
