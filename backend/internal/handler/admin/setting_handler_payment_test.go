package admin

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestDTORechargeProductsEncodesEmptyFeaturesAsArray(t *testing.T) {
	t.Parallel()

	products := dtoRechargeProducts([]service.RechargeProduct{
		{
			ID:     "starter",
			Name:   "Starter",
			Amount: 30,
		},
	})

	payload, err := json.Marshal(products)
	if err != nil {
		t.Fatalf("marshal products: %v", err)
	}
	if !strings.Contains(string(payload), `"features":[]`) {
		t.Fatalf("features should encode as empty array, got %s", string(payload))
	}
}
