package payment

import "strings"

type ProviderInstanceSource struct {
	ProviderKey    string
	SupportedTypes string
}

func LegacyOrderMatchesProviderInstance(orderPaymentType string, inst ProviderInstanceSource) bool {
	baseType := GetBasePaymentType(strings.TrimSpace(orderPaymentType))
	instanceProviderKey := strings.TrimSpace(inst.ProviderKey)
	if baseType == "" {
		return false
	}

	if baseType == TypeStripe {
		return instanceProviderKey == TypeStripe
	}
	if instanceProviderKey == TypeStripe {
		return false
	}
	if instanceProviderKey == baseType {
		return true
	}
	return InstanceSupportsType(inst.SupportedTypes, baseType)
}

func LegacyOrderMatchesProviderInstanceFields(orderPaymentType, providerKey, supportedTypes string) bool {
	return LegacyOrderMatchesProviderInstance(orderPaymentType, ProviderInstanceSource{
		ProviderKey:    providerKey,
		SupportedTypes: supportedTypes,
	})
}
