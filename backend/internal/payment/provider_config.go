package payment

import (
	"fmt"
	"strings"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
)

// providerSensitiveConfigFields is the authoritative list of config keys that
// are treated as secrets per provider. Must stay in sync with the frontend
// definition at frontend/src/components/payment/providerConfig.ts
// (PROVIDER_CONFIG_FIELDS, fields with sensitive: true).
//
// Key matching is case-insensitive. Non-listed keys (e.g. appId, notifyUrl,
// stripe publishableKey) are returned in plaintext by the admin GET API.
var providerSensitiveConfigFields = map[string]map[string]struct{}{
	TypeEasyPay:   {"pkey": {}},
	TypeAlipay:    {"privatekey": {}, "publickey": {}, "alipaypublickey": {}},
	TypeWxpay:     {"privatekey": {}, "apiv3key": {}, "publickey": {}},
	TypeStripe:    {"secretkey": {}, "webhooksecret": {}},
	TypeCreem:     {"apikey": {}, "webhooksecret": {}},
	TypeWaffo:     {"apikey": {}, "privatekey": {}, "waffopublickey": {}},
	TypeAirwallex: {"apikey": {}, "webhooksecret": {}},
}

// providerPendingOrderProtectedConfigFields lists config keys that cannot be
// changed while the instance has in-progress orders. This includes secrets plus
// all provider identity fields that are snapshotted into orders or used by
// webhook/refund verification.
var providerPendingOrderProtectedConfigFields = map[string]map[string]struct{}{
	TypeEasyPay:   {"pkey": {}, "pid": {}},
	TypeAlipay:    {"privatekey": {}, "publickey": {}, "alipaypublickey": {}, "appid": {}},
	TypeWxpay:     {"privatekey": {}, "apiv3key": {}, "publickey": {}, "appid": {}, "mpappid": {}, "mchid": {}, "publickeyid": {}, "certserial": {}},
	TypeStripe:    {"secretkey": {}, "webhooksecret": {}, "currency": {}},
	TypeCreem:     {"apikey": {}, "webhooksecret": {}},
	TypeWaffo:     {"apikey": {}, "privatekey": {}, "waffopublickey": {}, "merchantid": {}},
	TypeAirwallex: {"apikey": {}, "webhooksecret": {}, "clientid": {}, "accountid": {}, "currency": {}},
}

var validProviderKeys = map[string]bool{
	TypeEasyPay:   true,
	TypeAlipay:    true,
	TypeWxpay:     true,
	TypeStripe:    true,
	TypeCreem:     true,
	TypeWaffo:     true,
	TypeAirwallex: true,
}

func IsSensitiveProviderConfigField(providerKey, fieldName string) bool {
	fields, ok := providerSensitiveConfigFields[providerKey]
	if !ok {
		return false
	}
	_, found := fields[strings.ToLower(fieldName)]
	return found
}

func HasPendingOrderProtectedConfigChange(providerKey string, currentConfig, nextConfig map[string]string) bool {
	fields, ok := providerPendingOrderProtectedConfigFields[providerKey]
	if !ok {
		return false
	}
	for fieldName := range fields {
		if ProviderConfigFieldValue(currentConfig, fieldName) != ProviderConfigFieldValue(nextConfig, fieldName) {
			return true
		}
	}
	return false
}

func ProviderConfigFieldValue(config map[string]string, fieldName string) string {
	for key, value := range config {
		if strings.EqualFold(key, fieldName) {
			return value
		}
	}
	return ""
}

func IsValidProviderKey(providerKey string) bool {
	return validProviderKeys[providerKey]
}

func ValidateProviderRequest(providerKey, name string) error {
	if strings.TrimSpace(name) == "" {
		return infraerrors.BadRequest("VALIDATION_ERROR", "provider name is required")
	}
	if !IsValidProviderKey(providerKey) {
		return infraerrors.BadRequest("VALIDATION_ERROR", fmt.Sprintf("invalid provider key: %s", providerKey))
	}
	return nil
}
