package gateway

import "github.com/Aias00/cloudbase/internal/domain"

func IsAccountActive(status string) bool {
	return status == domain.StatusActive
}

func AccountBillingRateMultiplier(rateMultiplier *float64) float64 {
	if rateMultiplier == nil || *rateMultiplier < 0 {
		return 1.0
	}
	return *rateMultiplier
}

func AccountEffectiveLoadFactor(loadFactor *int, concurrency int) int {
	if loadFactor != nil && *loadFactor > 0 {
		return *loadFactor
	}
	if concurrency > 0 {
		return concurrency
	}
	return 1
}

func IsAccountSchedulable(status string, schedulable bool) bool {
	return IsAccountActive(status) && schedulable
}
