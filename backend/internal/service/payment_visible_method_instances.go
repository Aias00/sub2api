package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/ent/paymentproviderinstance"
	"github.com/Aias00/cloudbase/internal/payment"
	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
)

func filterEnabledVisibleMethodInstances(instances []*dbent.PaymentProviderInstance, method string) []*dbent.PaymentProviderInstance {
	sources := payment.FilterEnabledVisibleMethodSources(visibleMethodSourcesFromEnt(instances), method)
	return visibleMethodInstancesFromSources(instances, sources)
}

func filterVisibleMethodInstancesByProviderKey(instances []*dbent.PaymentProviderInstance, method string, providerKey string) []*dbent.PaymentProviderInstance {
	sources := payment.FilterVisibleMethodSourcesByProviderKey(visibleMethodSourcesFromEnt(instances), method, providerKey)
	return visibleMethodInstancesFromSources(instances, sources)
}

func distinctVisibleMethodProviderKeys(instances []*dbent.PaymentProviderInstance) []string {
	return payment.DistinctVisibleMethodProviderKeys(visibleMethodSourcesFromEnt(instances))
}

func selectVisibleMethodInstanceByProviderKey(instances []*dbent.PaymentProviderInstance, providerKey string) *dbent.PaymentProviderInstance {
	source, ok := payment.SelectVisibleMethodSourceByProviderKey(visibleMethodSourcesFromEnt(instances), providerKey)
	if !ok {
		return nil
	}
	for _, inst := range instances {
		if inst != nil && int64(inst.ID) == source.ID {
			return inst
		}
	}
	return nil
}

func visibleMethodSourceFromEnt(inst *dbent.PaymentProviderInstance) (payment.VisibleMethodProviderSource, bool) {
	if inst == nil {
		return payment.VisibleMethodProviderSource{}, false
	}
	return payment.VisibleMethodProviderSource{
		ID:             int64(inst.ID),
		ProviderKey:    inst.ProviderKey,
		SupportedTypes: inst.SupportedTypes,
		Enabled:        inst.Enabled,
	}, true
}

func visibleMethodSourcesFromEnt(instances []*dbent.PaymentProviderInstance) []payment.VisibleMethodProviderSource {
	sources := make([]payment.VisibleMethodProviderSource, 0, len(instances))
	for _, inst := range instances {
		if source, ok := visibleMethodSourceFromEnt(inst); ok {
			sources = append(sources, source)
		}
	}
	return sources
}

func visibleMethodInstancesFromSources(instances []*dbent.PaymentProviderInstance, sources []payment.VisibleMethodProviderSource) []*dbent.PaymentProviderInstance {
	sourceIDs := make(map[int64]struct{}, len(sources))
	for _, source := range sources {
		sourceIDs[source.ID] = struct{}{}
	}
	filtered := make([]*dbent.PaymentProviderInstance, 0, len(sources))
	seen := make(map[int64]struct{}, len(instances))
	for _, inst := range instances {
		if inst == nil {
			continue
		}
		id := int64(inst.ID)
		if _, ok := sourceIDs[id]; !ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		filtered = append(filtered, inst)
	}
	return filtered
}

func (s *PaymentConfigService) validateVisibleMethodEnablementConflicts(
	ctx context.Context,
	excludeID int64,
	providerKey string,
	supportedTypes string,
	enabled bool,
) error {
	// Visible methods are selected by configured source (official/easypay),
	// so multiple enabled providers can intentionally claim the same user-facing
	// method. Order creation and limits will route through the configured source.
	_, _, _, _, _ = ctx, excludeID, providerKey, supportedTypes, enabled
	return nil
}

func (s *PaymentConfigService) resolveVisibleMethodSourceProviderKey(ctx context.Context, method string) (string, error) {
	method = NormalizeVisibleMethod(method)
	sourceKey := visibleMethodSourceSettingKey(method)
	rawSource := ""
	if s != nil && s.settingRepo != nil && sourceKey != "" {
		value, err := s.settingRepo.GetValue(ctx, sourceKey)
		if err != nil {
			if !errors.Is(err, ErrSettingNotFound) {
				return "", fmt.Errorf("get %s: %w", sourceKey, err)
			}
		} else {
			rawSource = value
		}
	}

	normalizedSource, err := normalizeVisibleMethodSettingSource(method, rawSource, true)
	if err != nil {
		return "", err
	}
	if normalizedSource == "" {
		return "", nil
	}
	providerKey, ok := VisibleMethodProviderKeyForSource(method, normalizedSource)
	if !ok {
		return "", infraerrors.BadRequest(
			"INVALID_PAYMENT_VISIBLE_METHOD_SOURCE",
			fmt.Sprintf("%s source must be one of the supported payment providers", method),
		)
	}
	return providerKey, nil
}

func (s *PaymentConfigService) resolveVisibleMethodProviderKey(
	ctx context.Context,
	method string,
	matching []*dbent.PaymentProviderInstance,
) (string, error) {
	switch providerKeys := distinctVisibleMethodProviderKeys(matching); len(providerKeys) {
	case 0:
		return "", nil
	case 1:
		return strings.TrimSpace(providerKeys[0]), nil
	default:
		providerKey, err := s.resolveVisibleMethodSourceProviderKey(ctx, method)
		if err != nil {
			return "", err
		}
		if providerKey == "" {
			return "", nil
		}
		selected := selectVisibleMethodInstanceByProviderKey(matching, providerKey)
		if selected == nil {
			return "", infraerrors.BadRequest(
				"INVALID_PAYMENT_VISIBLE_METHOD_SOURCE",
				fmt.Sprintf("%s source has no enabled provider instance", method),
			)
		}
		return strings.TrimSpace(selected.ProviderKey), nil
	}
}

func (s *PaymentConfigService) resolveEnabledVisibleMethodInstance(
	ctx context.Context,
	method string,
) (*dbent.PaymentProviderInstance, error) {
	if s == nil || s.entClient == nil {
		return nil, nil
	}

	method = NormalizeVisibleMethod(method)
	if method != payment.TypeAlipay && method != payment.TypeWxpay {
		return nil, nil
	}

	instances, err := s.entClient.PaymentProviderInstance.Query().
		Where(paymentproviderinstance.EnabledEQ(true)).
		Order(paymentproviderinstance.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query enabled payment providers: %w", err)
	}

	matching := filterEnabledVisibleMethodInstances(instances, method)
	providerKey, err := s.resolveVisibleMethodProviderKey(ctx, method, matching)
	if err != nil {
		return nil, err
	}
	if providerKey == "" {
		if len(matching) == 0 {
			return nil, nil
		}
		return &dbent.PaymentProviderInstance{ProviderKey: ""}, nil
	}
	return selectVisibleMethodInstanceByProviderKey(matching, providerKey), nil
}
