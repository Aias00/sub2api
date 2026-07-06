package gateway

import (
	"strings"

	"github.com/Aias00/cloudbase/internal/domain"
)

func IsGroupActive(status string) bool {
	return status == domain.StatusActive
}

func IsGroupSubscriptionType(subscriptionType string) bool {
	return subscriptionType == domain.SubscriptionTypeSubscription
}

func HasPositiveLimit(limit *float64) bool {
	return limit != nil && *limit > 0
}

func SelectGroupImagePrice(imageSize string, imagePrice1K, imagePrice2K, imagePrice4K *float64) *float64 {
	switch imageSize {
	case "1K":
		return imagePrice1K
	case "2K":
		return imagePrice2K
	case "4K":
		return imagePrice4K
	default:
		return imagePrice2K
	}
}

type GroupContextValidityInput struct {
	ID       int64
	Hydrated bool
	Platform string
	Status   string
}

func IsGroupContextValid(input GroupContextValidityInput) bool {
	return input.ID > 0 && input.Hydrated && input.Platform != "" && input.Status != ""
}

func RoutingAccountIDs(modelRoutingEnabled bool, modelRouting map[string][]int64, requestedModel string) []int64 {
	if !modelRoutingEnabled || len(modelRouting) == 0 || requestedModel == "" {
		return nil
	}
	if accountIDs, ok := modelRouting[requestedModel]; ok && len(accountIDs) > 0 {
		return accountIDs
	}
	for pattern, accountIDs := range modelRouting {
		if MatchModelPattern(pattern, requestedModel) && len(accountIDs) > 0 {
			return accountIDs
		}
	}
	return nil
}

func MatchModelPattern(pattern, model string) bool {
	if pattern == model {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(model, strings.TrimSuffix(pattern, "*"))
	}
	return false
}
