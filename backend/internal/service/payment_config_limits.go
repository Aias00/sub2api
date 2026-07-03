package service

import (
	"context"
	"fmt"

	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/ent/paymentproviderinstance"
	"github.com/Aias00/cloudbase/internal/payment"
	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
)

// GetAvailableMethodLimits collects all payment types from enabled provider
// instances and returns limits for each, plus the global widest range.
// Stripe sub-types (card, link) are aggregated under "stripe".
func (s *PaymentConfigService) GetAvailableMethodLimits(ctx context.Context) (*MethodLimitsResponse, error) {
	instances, err := s.entClient.PaymentProviderInstance.Query().
		Where(paymentproviderinstance.EnabledEQ(true)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query provider instances: %w", err)
	}
	typeInstances := pcGroupByPaymentType(instances)
	typeInstances = s.pcApplyEnabledVisibleMethodInstances(ctx, typeInstances, instances)
	resp := &MethodLimitsResponse{
		Methods: make(map[string]MethodLimits, len(typeInstances)),
	}
	for pt, insts := range typeInstances {
		currency, ok := s.pcAggregateMethodCurrency(insts)
		if !ok {
			continue
		}
		ml := pcAggregateMethodLimits(pt, insts)
		ml.Currency = currency
		resp.Methods[ml.PaymentType] = ml
	}
	resp.GlobalMin, resp.GlobalMax = pcComputeGlobalRange(resp.Methods)
	return resp, nil
}

func (s *PaymentConfigService) pcApplyEnabledVisibleMethodInstances(ctx context.Context, typeInstances map[string][]*dbent.PaymentProviderInstance, instances []*dbent.PaymentProviderInstance) map[string][]*dbent.PaymentProviderInstance {
	if len(typeInstances) == 0 {
		return typeInstances
	}

	filtered := make(map[string][]*dbent.PaymentProviderInstance, len(typeInstances))
	for paymentType, groupedInstances := range typeInstances {
		filtered[paymentType] = groupedInstances
	}

	for _, method := range []string{payment.TypeAlipay, payment.TypeWxpay} {
		matching := filterEnabledVisibleMethodInstances(instances, method)
		providerKey, err := s.resolveVisibleMethodProviderKey(ctx, method, matching)
		if err != nil {
			delete(filtered, method)
			continue
		}
		if providerKey == "" {
			if len(matching) == 0 {
				delete(filtered, method)
				continue
			}
			filtered[method] = matching
			continue
		}
		selectedInstances := filterVisibleMethodInstancesByProviderKey(instances, method, providerKey)
		if len(selectedInstances) == 0 {
			delete(filtered, method)
			continue
		}
		filtered[method] = selectedInstances
	}
	return filtered
}

// GetMethodLimits returns per-payment-type limits from enabled provider instances.
func (s *PaymentConfigService) GetMethodLimits(ctx context.Context, types []string) ([]MethodLimits, error) {
	instances, err := s.entClient.PaymentProviderInstance.Query().
		Where(paymentproviderinstance.EnabledEQ(true)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query provider instances: %w", err)
	}
	result := make([]MethodLimits, 0, len(types))
	for _, pt := range types {
		var matching []*dbent.PaymentProviderInstance
		for _, inst := range instances {
			if payment.InstanceSupportsType(inst.SupportedTypes, pt) {
				matching = append(matching, inst)
			}
		}
		currency, ok := s.pcAggregateMethodCurrency(matching)
		if !ok {
			continue
		}
		ml := pcAggregateMethodLimits(pt, matching)
		ml.Currency = currency
		result = append(result, ml)
	}
	return result, nil
}

func (s *PaymentConfigService) ValidateMethodCurrencyConsistency(ctx context.Context, paymentType string) (string, error) {
	method := NormalizeVisibleMethod(paymentType)
	if method == "" || s == nil || s.entClient == nil {
		return payment.DefaultPaymentCurrency, nil
	}

	instances, err := s.entClient.PaymentProviderInstance.Query().
		Where(paymentproviderinstance.EnabledEQ(true)).All(ctx)
	if err != nil {
		return "", fmt.Errorf("query provider instances: %w", err)
	}

	typeInstances := pcGroupByPaymentType(instances)
	typeInstances = s.pcApplyEnabledVisibleMethodInstances(ctx, typeInstances, instances)
	matching := typeInstances[method]
	if len(matching) == 0 {
		return payment.DefaultPaymentCurrency, nil
	}

	currency, ok := s.pcAggregateMethodCurrency(matching)
	if !ok {
		return "", infraerrors.ServiceUnavailable(
			"PAYMENT_METHOD_CURRENCY_CONFLICT",
			"payment method has enabled provider instances with mixed currencies",
		).WithMetadata(map[string]string{"payment_type": method})
	}
	return currency, nil
}

func (s *PaymentConfigService) pcAggregateMethodCurrency(instances []*dbent.PaymentProviderInstance) (string, bool) {
	currency := ""
	for _, inst := range instances {
		next := s.pcInstancePaymentCurrency(inst)
		if next == "" {
			continue
		}
		if currency == "" {
			currency = next
			continue
		}
		if currency != next {
			return "", false
		}
	}
	if currency == "" {
		return payment.DefaultPaymentCurrency, true
	}
	return currency, true
}

func (s *PaymentConfigService) pcInstancePaymentCurrency(inst *dbent.PaymentProviderInstance) string {
	if inst == nil {
		return payment.DefaultPaymentCurrency
	}
	cfg := map[string]string{}
	if s != nil {
		decrypted, err := s.decryptConfig(inst.Config)
		if err == nil && decrypted != nil {
			cfg = decrypted
		}
	}
	return paymentProviderConfigCurrency(inst.ProviderKey, cfg)
}

// pcGroupByPaymentType groups instances by user-facing payment type.
// For Stripe providers, ALL sub-types (card, link, alipay, wxpay) map to "stripe"
// because the user sees a single "Stripe" button, not individual sub-methods.
// Uses a seen set to avoid counting one instance twice.
func pcGroupByPaymentType(instances []*dbent.PaymentProviderInstance) map[string][]*dbent.PaymentProviderInstance {
	sourceGroups := payment.GroupProviderLimitSourcesByPaymentType(paymentLimitSourcesFromEnt(instances))
	byID := make(map[int64]*dbent.PaymentProviderInstance, len(instances))
	for _, inst := range instances {
		if inst != nil {
			byID[int64(inst.ID)] = inst
		}
	}
	typeInstances := make(map[string][]*dbent.PaymentProviderInstance, len(sourceGroups))
	for paymentType, sources := range sourceGroups {
		for _, source := range sources {
			if inst := byID[source.ID]; inst != nil {
				typeInstances[paymentType] = append(typeInstances[paymentType], inst)
			}
		}
	}
	return typeInstances
}

// pcInstanceTypeLimits extracts per-type limits from a provider instance.
// Returns (limits, true) if configured; (zero, false) if unlimited.
// For Stripe instances, limits are stored under "stripe" key regardless of sub-types.
func pcInstanceTypeLimits(inst *dbent.PaymentProviderInstance, pt string) (payment.ChannelLimits, bool) {
	if inst == nil {
		return payment.ChannelLimits{}, false
	}
	return payment.InstanceTypeLimits(inst.Limits, pt)
}

// unionFloat merges a single limit value into the aggregate using UNION semantics.
//   - For "min" fields (wantMin=true): keeps the lowest non-zero value
//   - For "max"/"cap" fields (wantMin=false): keeps the highest non-zero value
//   - If any value is 0 (unlimited), the result is unlimited.
//
// Returns (aggregated value, still limited).
func unionFloat(agg float64, limited bool, val float64, wantMin bool) (float64, bool) {
	return payment.UnionFloat(agg, limited, val, wantMin)
}

// pcAggregateMethodLimits computes the UNION (least restrictive) of limits
// across all provider instances for a given payment type.
//
// Since the load balancer can route an order to any available instance,
// the user should see the widest possible range:
//   - SingleMin: lowest floor across instances; 0 if any is unlimited
//   - SingleMax: highest ceiling across instances; 0 if any is unlimited
//   - DailyLimit: highest cap across instances; 0 if any is unlimited
func pcAggregateMethodLimits(pt string, instances []*dbent.PaymentProviderInstance) MethodLimits {
	return payment.AggregateMethodLimits(pt, paymentLimitSourcesFromEnt(instances))
}

// pcComputeGlobalRange computes the widest [min, max] across all methods.
// Uses the same union logic: lowest min, highest max, 0 if any is unlimited.
func pcComputeGlobalRange(methods map[string]MethodLimits) (globalMin, globalMax float64) {
	return payment.ComputeGlobalRange(methods)
}

func paymentLimitSourcesFromEnt(instances []*dbent.PaymentProviderInstance) []payment.ProviderLimitSource {
	sources := make([]payment.ProviderLimitSource, 0, len(instances))
	for _, inst := range instances {
		if inst == nil {
			continue
		}
		sources = append(sources, payment.ProviderLimitSource{
			ID:             int64(inst.ID),
			ProviderKey:    inst.ProviderKey,
			SupportedTypes: inst.SupportedTypes,
			Limits:         inst.Limits,
		})
	}
	return sources
}
