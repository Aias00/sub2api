//go:build unit

package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRechargeProducts(t *testing.T) {
	t.Parallel()

	products := ParseRechargeProducts(
		`[{"id":"starter","name":"  Starter ","description":" Desc ","amount":30,"badge":" Hot ","features":[" one ",""," two "],"sort_order":10}]`,
		1.5,
	)
	require.Len(t, products, 1)
	product := products[0]
	require.Equal(t, "starter", product.ID)
	require.Equal(t, "Starter", product.Name)
	require.Equal(t, "Desc", product.Description)
	require.Equal(t, "Hot", product.Badge)
	require.Equal(t, 30.0, product.Amount)
	require.Equal(t, 45.0, product.CreditedAmount)
	require.Equal(t, []string{"one", "two"}, product.Features)
}

func TestNormalizeRechargeProductsRejectsInvalidProducts(t *testing.T) {
	t.Parallel()

	_, err := NormalizeRechargeProducts([]RechargeProduct{{Name: "Missing ID", Amount: 1}}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "INVALID_RECHARGE_PRODUCT_ID")

	_, err = NormalizeRechargeProducts([]RechargeProduct{
		{ID: "dup", Name: "A", Amount: 1},
		{ID: "dup", Name: "B", Amount: 2},
	}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "DUPLICATE_RECHARGE_PRODUCT_ID")

	_, err = NormalizeRechargeProducts([]RechargeProduct{{ID: "bad", Name: "Bad", Amount: 0}}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "INVALID_RECHARGE_PRODUCT_AMOUNT")
}

func TestNormalizeRechargeProductsSortsBySortOrderThenName(t *testing.T) {
	t.Parallel()

	products, err := NormalizeRechargeProducts([]RechargeProduct{
		{ID: "b", Name: "Beta", Amount: 2, SortOrder: 10},
		{ID: "a", Name: "Alpha", Amount: 1, SortOrder: 10},
		{ID: "c", Name: "Core", Amount: 3, SortOrder: 1},
	}, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"c", "a", "b"}, []string{products[0].ID, products[1].ID, products[2].ID})
}

func TestApplyProductNameAffix(t *testing.T) {
	t.Parallel()

	require.False(t, HasProductNameAffix("", " "))
	require.Equal(t, "Cloudbase", ApplyProductNameAffix("Cloudbase", "", ""))
	require.Equal(t, "PRE Cloudbase SUF", ApplyProductNameAffix("Cloudbase", " PRE ", " SUF "))
}
