package payment

import (
	"encoding/json"
	"strings"
)

// MethodLimits holds per-payment-type limits.
type MethodLimits struct {
	PaymentType string  `json:"payment_type"`
	Currency    string  `json:"currency"`
	FeeRate     float64 `json:"fee_rate"`
	DailyLimit  float64 `json:"daily_limit"`
	SingleMin   float64 `json:"single_min"`
	SingleMax   float64 `json:"single_max"`
}

// MethodLimitsResponse is the full response for the user-facing /limits API.
// It includes per-method limits and the global widest range (union of all methods).
type MethodLimitsResponse struct {
	Methods   map[string]MethodLimits `json:"methods"`
	GlobalMin float64                 `json:"global_min"` // 0 = no minimum
	GlobalMax float64                 `json:"global_max"` // 0 = no maximum
}

type ProviderLimitSource struct {
	ID             int64
	ProviderKey    string
	SupportedTypes string
	Limits         string
}

// GroupProviderLimitSourcesByPaymentType groups instances by user-facing payment type.
// For Stripe providers, all sub-types map to "stripe" because the user sees a
// single Stripe button, not individual sub-methods.
func GroupProviderLimitSourcesByPaymentType(instances []ProviderLimitSource) map[string][]ProviderLimitSource {
	typeInstances := make(map[string][]ProviderLimitSource)
	seen := make(map[string]map[int64]bool)
	add := func(key string, inst ProviderLimitSource) {
		if seen[key] == nil {
			seen[key] = make(map[int64]bool)
		}
		if !seen[key][inst.ID] {
			seen[key][inst.ID] = true
			typeInstances[key] = append(typeInstances[key], inst)
		}
	}
	for _, inst := range instances {
		if inst.ProviderKey == TypeStripe {
			add(TypeStripe, inst)
			continue
		}
		for _, t := range SplitTypes(inst.SupportedTypes) {
			add(t, inst)
		}
	}
	return typeInstances
}

func SplitTypes(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func InstanceTypeLimits(limitsJSON string, paymentType string) (ChannelLimits, bool) {
	if limitsJSON == "" {
		return ChannelLimits{}, false
	}
	var limits InstanceLimits
	if err := json.Unmarshal([]byte(limitsJSON), &limits); err != nil {
		return ChannelLimits{}, false
	}
	cl, ok := limits[paymentType]
	return cl, ok
}

// UnionFloat merges a single limit value into the aggregate using UNION semantics.
// For min fields it keeps the lowest non-zero value; for max/cap fields it keeps
// the highest non-zero value. Any zero value makes that dimension unlimited.
func UnionFloat(agg float64, limited bool, val float64, wantMin bool) (float64, bool) {
	if val == 0 {
		return agg, false
	}
	if !limited {
		return agg, false
	}
	if agg == 0 {
		return val, true
	}
	if wantMin && val < agg {
		return val, true
	}
	if !wantMin && val > agg {
		return val, true
	}
	return agg, true
}

func AggregateMethodLimits(paymentType string, instances []ProviderLimitSource) MethodLimits {
	ml := MethodLimits{PaymentType: paymentType}
	minLimited, maxLimited, dailyLimited := true, true, true

	for _, inst := range instances {
		cl, hasLimits := InstanceTypeLimits(inst.Limits, paymentType)
		if !hasLimits {
			return MethodLimits{PaymentType: paymentType}
		}
		ml.SingleMin, minLimited = UnionFloat(ml.SingleMin, minLimited, cl.SingleMin, true)
		ml.SingleMax, maxLimited = UnionFloat(ml.SingleMax, maxLimited, cl.SingleMax, false)
		ml.DailyLimit, dailyLimited = UnionFloat(ml.DailyLimit, dailyLimited, cl.DailyLimit, false)
	}

	if !minLimited {
		ml.SingleMin = 0
	}
	if !maxLimited {
		ml.SingleMax = 0
	}
	if !dailyLimited {
		ml.DailyLimit = 0
	}
	return ml
}

func ComputeGlobalRange(methods map[string]MethodLimits) (globalMin, globalMax float64) {
	minLimited, maxLimited := true, true
	for _, ml := range methods {
		globalMin, minLimited = UnionFloat(globalMin, minLimited, ml.SingleMin, true)
		globalMax, maxLimited = UnionFloat(globalMax, maxLimited, ml.SingleMax, false)
	}
	if !minLimited {
		globalMin = 0
	}
	if !maxLimited {
		globalMax = 0
	}
	return globalMin, globalMax
}
