package payment

import "strings"

type VisibleMethodProviderSource struct {
	ID             int64
	ProviderKey    string
	SupportedTypes string
	Enabled        bool
}

func EnabledVisibleMethodsForProvider(providerKey, supportedTypes string) []string {
	methodSet := make(map[string]struct{}, 2)
	methodOrder := make([]string, 0, 2)
	addMethod := func(method string) {
		method = NormalizeVisibleMethod(method)
		if method == "" {
			return
		}
		if _, ok := methodSet[method]; ok {
			return
		}
		methodSet[method] = struct{}{}
		methodOrder = append(methodOrder, method)
	}

	switch strings.TrimSpace(providerKey) {
	case TypeAlipay:
		if strings.TrimSpace(supportedTypes) == "" {
			addMethod(TypeAlipay)
			break
		}
		for _, supportedType := range SplitTypes(supportedTypes) {
			if NormalizeVisibleMethod(supportedType) == TypeAlipay {
				addMethod(TypeAlipay)
				break
			}
		}
	case TypeWxpay:
		if strings.TrimSpace(supportedTypes) == "" {
			addMethod(TypeWxpay)
			break
		}
		for _, supportedType := range SplitTypes(supportedTypes) {
			if NormalizeVisibleMethod(supportedType) == TypeWxpay {
				addMethod(TypeWxpay)
				break
			}
		}
	case TypeEasyPay:
		for _, supportedType := range SplitTypes(supportedTypes) {
			addMethod(supportedType)
		}
	}

	methods := make([]string, 0, len(methodSet))
	for _, method := range []string{TypeAlipay, TypeWxpay} {
		if _, ok := methodSet[method]; ok {
			methods = append(methods, method)
		}
	}
	for _, method := range methodOrder {
		if method == TypeAlipay || method == TypeWxpay {
			continue
		}
		methods = append(methods, method)
	}
	return methods
}

func ProviderSupportsVisibleMethod(source VisibleMethodProviderSource, method string) bool {
	if !source.Enabled {
		return false
	}
	method = NormalizeVisibleMethod(method)
	for _, candidate := range EnabledVisibleMethodsForProvider(source.ProviderKey, source.SupportedTypes) {
		if candidate == method {
			return true
		}
	}
	return false
}

func FilterEnabledVisibleMethodSources(sources []VisibleMethodProviderSource, method string) []VisibleMethodProviderSource {
	filtered := make([]VisibleMethodProviderSource, 0, len(sources))
	for _, source := range sources {
		if ProviderSupportsVisibleMethod(source, method) {
			filtered = append(filtered, source)
		}
	}
	return filtered
}

func FilterVisibleMethodSourcesByProviderKey(sources []VisibleMethodProviderSource, method string, providerKey string) []VisibleMethodProviderSource {
	filtered := make([]VisibleMethodProviderSource, 0, len(sources))
	for _, source := range sources {
		if !ProviderSupportsVisibleMethod(source, method) {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(source.ProviderKey), strings.TrimSpace(providerKey)) {
			continue
		}
		filtered = append(filtered, source)
	}
	return filtered
}

func DistinctVisibleMethodProviderKeys(sources []VisibleMethodProviderSource) []string {
	seen := make(map[string]struct{}, len(sources))
	keys := make([]string, 0, len(sources))
	for _, source := range sources {
		key := strings.TrimSpace(source.ProviderKey)
		if key == "" {
			continue
		}
		normalized := strings.ToLower(key)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}

func SelectVisibleMethodSourceByProviderKey(sources []VisibleMethodProviderSource, providerKey string) (VisibleMethodProviderSource, bool) {
	providerKey = strings.TrimSpace(providerKey)
	if providerKey == "" {
		return VisibleMethodProviderSource{}, false
	}
	for _, source := range sources {
		if strings.EqualFold(strings.TrimSpace(source.ProviderKey), providerKey) {
			return source, true
		}
	}
	return VisibleMethodProviderSource{}, false
}

func BuildVisibleMethodSourceAvailability(sources []VisibleMethodProviderSource) map[string]bool {
	available := make(map[string]bool, 4)
	for _, source := range sources {
		switch source.ProviderKey {
		case TypeAlipay:
			if source.SupportedTypes == "" || InstanceSupportsType(source.SupportedTypes, TypeAlipay) || InstanceSupportsType(source.SupportedTypes, TypeAlipayDirect) {
				available[VisibleMethodSourceOfficialAlipay] = true
			}
		case TypeWxpay:
			if source.SupportedTypes == "" || InstanceSupportsType(source.SupportedTypes, TypeWxpay) || InstanceSupportsType(source.SupportedTypes, TypeWxpayDirect) {
				available[VisibleMethodSourceOfficialWechat] = true
			}
		case TypeEasyPay:
			for _, supportedType := range SplitTypes(source.SupportedTypes) {
				switch NormalizeVisibleMethod(supportedType) {
				case TypeAlipay:
					available[VisibleMethodSourceEasyPayAlipay] = true
				case TypeWxpay:
					available[VisibleMethodSourceEasyPayWechat] = true
				}
			}
		}
	}
	return available
}

func ApplyVisibleMethodRoutingToEnabledTypes(base []string, vals map[string]string, available map[string]bool) []string {
	shouldExpose := map[string]bool{
		TypeAlipay: VisibleMethodShouldBeExposed(TypeAlipay, vals, available),
		TypeWxpay:  VisibleMethodShouldBeExposed(TypeWxpay, vals, available),
	}

	seen := make(map[string]struct{}, len(base)+2)
	out := make([]string, 0, len(base)+2)
	appendType := func(paymentType string) {
		paymentType = NormalizeVisibleMethod(paymentType)
		if paymentType == "" {
			return
		}
		if _, ok := seen[paymentType]; ok {
			return
		}
		seen[paymentType] = struct{}{}
		out = append(out, paymentType)
	}

	for _, paymentType := range base {
		visibleMethod := NormalizeVisibleMethod(paymentType)
		switch visibleMethod {
		case TypeAlipay, TypeWxpay:
			if shouldExpose[visibleMethod] {
				appendType(visibleMethod)
			}
		default:
			appendType(visibleMethod)
		}
	}

	for _, visibleMethod := range []string{TypeAlipay, TypeWxpay} {
		if shouldExpose[visibleMethod] {
			appendType(visibleMethod)
		}
	}
	return out
}

func VisibleMethodShouldBeExposed(method string, vals map[string]string, available map[string]bool) bool {
	enabledKey := VisibleMethodEnabledSettingKey(method)
	sourceKey := VisibleMethodSourceSettingKey(method)
	if enabledKey == "" || sourceKey == "" || vals[enabledKey] != "true" {
		return false
	}
	source := NormalizeVisibleMethodSource(method, vals[sourceKey])
	return source != "" && available[source]
}
