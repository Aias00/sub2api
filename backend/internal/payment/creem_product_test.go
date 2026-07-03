//go:build unit

package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveCreemRechargeProduct(t *testing.T) {
	t.Parallel()

	product, err := ResolveCreemRechargeProduct(TypeCreem, OrderTypeBalance, "starter", []RechargeProduct{
		{ID: "starter", Amount: 30, CreditedAmount: 45, CreemProductID: "prod_123"},
	})
	require.NoError(t, err)
	require.NotNil(t, product)
	require.Equal(t, 30.0, product.Amount)
	require.Equal(t, 45.0, product.CreditedAmount)
	require.Equal(t, "prod_123", product.CreemProductID)
}

func TestResolveCreemRechargeProductSkipsNonCreemPayments(t *testing.T) {
	t.Parallel()

	product, err := ResolveCreemRechargeProduct(TypeStripe, OrderTypeBalance, "starter", nil)
	require.NoError(t, err)
	require.Nil(t, product)

	product, err = ResolveCreemRechargeProduct(TypeCreem, OrderTypeSubscription, "starter", nil)
	require.NoError(t, err)
	require.Nil(t, product)
}

func TestResolveCreemRechargeProductRejectsInvalidSelection(t *testing.T) {
	t.Parallel()

	_, err := ResolveCreemRechargeProduct(TypeCreem, OrderTypeBalance, "", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CREEM_PRODUCT_REQUIRED")

	_, err = ResolveCreemRechargeProduct(TypeCreem, OrderTypeBalance, "starter", []RechargeProduct{{ID: "starter"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "CREEM_PRODUCT_NOT_CONFIGURED")

	_, err = ResolveCreemRechargeProduct(TypeCreem, OrderTypeBalance, "missing", []RechargeProduct{{ID: "starter", CreemProductID: "prod_123"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "CREEM_PRODUCT_NOT_FOUND")
}

func TestResolveCreemProviderProductID(t *testing.T) {
	t.Parallel()

	productID, err := ResolveCreemProviderProductID(TypeCreem, "plan_prod_456", true, nil)
	require.NoError(t, err)
	require.Equal(t, "plan_prod_456", productID)

	productID, err = ResolveCreemProviderProductID(TypeCreem, "", false, &RechargeProduct{CreemProductID: "prod_123"})
	require.NoError(t, err)
	require.Equal(t, "prod_123", productID)

	productID, err = ResolveCreemProviderProductID(TypeStripe, "", false, nil)
	require.NoError(t, err)
	require.Empty(t, productID)
}

func TestResolveCreemProviderProductIDRejectsMissingProductID(t *testing.T) {
	t.Parallel()

	_, err := ResolveCreemProviderProductID(TypeCreem, "", true, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CREEM_PRODUCT_NOT_CONFIGURED")

	_, err = ResolveCreemProviderProductID(TypeCreem, "", false, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CREEM_PRODUCT_REQUIRED")
}
