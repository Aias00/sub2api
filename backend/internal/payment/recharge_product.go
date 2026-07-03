package payment

import (
	"encoding/json"
	"math"
	"sort"
	"strings"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
)

type RechargeProduct struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Amount         float64  `json:"amount"`
	CreditedAmount float64  `json:"credited_amount"`
	CreemProductID string   `json:"creem_product_id,omitempty"`
	Badge          string   `json:"badge"`
	Recommended    bool     `json:"recommended"`
	Features       []string `json:"features"`
	SortOrder      int      `json:"sort_order"`
}

func ParseRechargeProducts(raw string, multiplier float64) []RechargeProduct {
	if strings.TrimSpace(raw) == "" {
		return []RechargeProduct{}
	}
	var products []RechargeProduct
	if err := json.Unmarshal([]byte(raw), &products); err != nil {
		return []RechargeProduct{}
	}
	normalized, err := NormalizeRechargeProducts(products, &multiplier)
	if err != nil {
		return []RechargeProduct{}
	}
	return normalized
}

func NormalizeRechargeProducts(products []RechargeProduct, multiplier *float64) ([]RechargeProduct, error) {
	if len(products) == 0 {
		return []RechargeProduct{}, nil
	}
	resolvedMultiplier := DefaultBalanceRechargeMultiplier
	if multiplier != nil {
		resolvedMultiplier = NormalizeBalanceRechargeMultiplier(*multiplier)
	}
	seenIDs := make(map[string]struct{}, len(products))
	normalized := make([]RechargeProduct, 0, len(products))
	for _, product := range products {
		product.ID = strings.TrimSpace(product.ID)
		product.Name = strings.TrimSpace(product.Name)
		product.Description = strings.TrimSpace(product.Description)
		product.Badge = strings.TrimSpace(product.Badge)
		if product.ID == "" {
			return nil, infraerrors.BadRequest("INVALID_RECHARGE_PRODUCT_ID", "recharge product id is required")
		}
		if _, exists := seenIDs[product.ID]; exists {
			return nil, infraerrors.BadRequest("DUPLICATE_RECHARGE_PRODUCT_ID", "recharge product ids must be unique")
		}
		if product.Name == "" {
			return nil, infraerrors.BadRequest("INVALID_RECHARGE_PRODUCT_NAME", "recharge product name is required")
		}
		if math.IsNaN(product.Amount) || math.IsInf(product.Amount, 0) || product.Amount <= 0 {
			return nil, infraerrors.BadRequest("INVALID_RECHARGE_PRODUCT_AMOUNT", "recharge product amount must be greater than 0")
		}
		product.Amount = math.Round(product.Amount*100) / 100
		product.CreditedAmount = CalculateCreditedBalance(product.Amount, resolvedMultiplier)
		cleanFeatures := make([]string, 0, len(product.Features))
		for _, feature := range product.Features {
			feature = strings.TrimSpace(feature)
			if feature != "" {
				cleanFeatures = append(cleanFeatures, feature)
			}
		}
		product.Features = cleanFeatures
		seenIDs[product.ID] = struct{}{}
		normalized = append(normalized, product)
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].SortOrder == normalized[j].SortOrder {
			return normalized[i].Name < normalized[j].Name
		}
		return normalized[i].SortOrder < normalized[j].SortOrder
	})
	return normalized, nil
}

func HasProductNameAffix(prefix, suffix string) bool {
	return strings.TrimSpace(prefix) != "" || strings.TrimSpace(suffix) != ""
}

func ApplyProductNameAffix(productName, prefix, suffix string) string {
	if !HasProductNameAffix(prefix, suffix) {
		return productName
	}
	return strings.TrimSpace(strings.TrimSpace(prefix) + " " + productName + " " + strings.TrimSpace(suffix))
}
