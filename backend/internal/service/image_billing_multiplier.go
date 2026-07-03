package service

import billingctx "github.com/Aias00/cloudbase/internal/billing"

func resolveImageRateMultiplier(apiKey *APIKey, effectiveGroupMultiplier float64) float64 {
	if apiKey != nil && apiKey.Group != nil && apiKey.Group.ImageRateIndependent {
		return billingctx.ResolveImageRateMultiplier(true, apiKey.Group.ImageRateMultiplier, effectiveGroupMultiplier)
	}
	return billingctx.ResolveImageRateMultiplier(false, 0, effectiveGroupMultiplier)
}
