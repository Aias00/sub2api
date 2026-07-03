//go:build unit

package payment

import "testing"

func TestUnionFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		agg         float64
		limited     bool
		val         float64
		wantMin     bool
		wantAgg     float64
		wantLimited bool
	}{
		{"first non-zero value", 0, true, 5, true, 5, true},
		{"lower min replaces", 10, true, 3, true, 3, true},
		{"higher max replaces", 10, true, 20, false, 20, true},
		{"zero value makes unlimited", 5, true, 0, true, 5, false},
		{"already unlimited stays unlimited", 5, false, 10, true, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotAgg, gotLimited := UnionFloat(tt.agg, tt.limited, tt.val, tt.wantMin)
			if gotAgg != tt.wantAgg || gotLimited != tt.wantLimited {
				t.Fatalf("UnionFloat(%v, %v, %v, %v) = (%v, %v), want (%v, %v)",
					tt.agg, tt.limited, tt.val, tt.wantMin,
					gotAgg, gotLimited, tt.wantAgg, tt.wantLimited)
			}
		})
	}
}

func TestAggregateMethodLimits(t *testing.T) {
	t.Parallel()

	t.Run("two instances union takes widest range", func(t *testing.T) {
		t.Parallel()
		ml := AggregateMethodLimits(TypeAlipay, []ProviderLimitSource{
			{ID: 1, ProviderKey: TypeEasyPay, SupportedTypes: TypeAlipay, Limits: `{"alipay":{"singleMin":5,"singleMax":100}}`},
			{ID: 2, ProviderKey: TypeEasyPay, SupportedTypes: TypeAlipay, Limits: `{"alipay":{"singleMin":2,"singleMax":200}}`},
		})
		if ml.SingleMin != 2 || ml.SingleMax != 200 {
			t.Fatalf("limits = min:%v max:%v, want min:2 max:200", ml.SingleMin, ml.SingleMax)
		}
	})

	t.Run("one instance unlimited makes aggregate unlimited", func(t *testing.T) {
		t.Parallel()
		ml := AggregateMethodLimits(TypeWxpay, []ProviderLimitSource{
			{ID: 1, ProviderKey: TypeEasyPay, SupportedTypes: TypeWxpay, Limits: `{"wxpay":{"singleMin":3,"singleMax":10}}`},
			{ID: 2, ProviderKey: TypeEasyPay, SupportedTypes: TypeWxpay, Limits: ""},
		})
		if ml.SingleMin != 0 || ml.SingleMax != 0 {
			t.Fatalf("limits = min:%v max:%v, want unlimited zeros", ml.SingleMin, ml.SingleMax)
		}
	})

	t.Run("daily limit aggregation", func(t *testing.T) {
		t.Parallel()
		ml := AggregateMethodLimits(TypeAlipay, []ProviderLimitSource{
			{ID: 1, ProviderKey: TypeEasyPay, SupportedTypes: TypeAlipay, Limits: `{"alipay":{"singleMin":1,"singleMax":100,"dailyLimit":500}}`},
			{ID: 2, ProviderKey: TypeEasyPay, SupportedTypes: TypeAlipay, Limits: `{"alipay":{"singleMin":2,"singleMax":200,"dailyLimit":1000}}`},
		})
		if ml.SingleMin != 1 || ml.SingleMax != 200 || ml.DailyLimit != 1000 {
			t.Fatalf("limits = %+v, want min:1 max:200 daily:1000", ml)
		}
	})
}

func TestGroupProviderLimitSourcesByPaymentType(t *testing.T) {
	t.Parallel()

	groups := GroupProviderLimitSourcesByPaymentType([]ProviderLimitSource{
		{ID: 1, ProviderKey: TypeStripe, SupportedTypes: "card,alipay,link,wxpay"},
		{ID: 2, ProviderKey: TypeEasyPay, SupportedTypes: "alipay,wxpay"},
	})

	if len(groups[TypeStripe]) != 1 || groups[TypeStripe][0].ID != 1 {
		t.Fatalf("stripe group = %+v, want only stripe instance", groups[TypeStripe])
	}
	if len(groups[TypeAlipay]) != 1 || groups[TypeAlipay][0].ID != 2 {
		t.Fatalf("alipay group = %+v, want only easypay instance", groups[TypeAlipay])
	}
	if len(groups[TypeWxpay]) != 1 || groups[TypeWxpay][0].ID != 2 {
		t.Fatalf("wxpay group = %+v, want only easypay instance", groups[TypeWxpay])
	}
}

func TestInstanceTypeLimits(t *testing.T) {
	t.Parallel()

	cl, ok := InstanceTypeLimits(`{"alipay":{"singleMin":2,"singleMax":14,"dailyLimit":500}}`, TypeAlipay)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if cl.SingleMin != 2 || cl.SingleMax != 14 || cl.DailyLimit != 500 {
		t.Fatalf("limits = %+v, want min:2 max:14 daily:500", cl)
	}
	if _, ok := InstanceTypeLimits(`{bad json}`, TypeAlipay); ok {
		t.Fatal("invalid JSON should return ok=false")
	}
}

func TestComputeGlobalRange(t *testing.T) {
	t.Parallel()

	gMin, gMax := ComputeGlobalRange(map[string]MethodLimits{
		TypeAlipay: {SingleMin: 2, SingleMax: 14},
		TypeWxpay:  {SingleMin: 1, SingleMax: 12},
		TypeStripe: {SingleMin: 5, SingleMax: 100},
	})
	if gMin != 1 || gMax != 100 {
		t.Fatalf("global range = (%v, %v), want (1, 100)", gMin, gMax)
	}

	gMin, gMax = ComputeGlobalRange(map[string]MethodLimits{
		TypeAlipay: {SingleMin: 2, SingleMax: 14},
		TypeStripe: {SingleMin: 0, SingleMax: 0},
	})
	if gMin != 0 || gMax != 0 {
		t.Fatalf("global range with unlimited = (%v, %v), want (0, 0)", gMin, gMax)
	}
}
