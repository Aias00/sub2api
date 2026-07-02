//go:build unit

package service

import (
	"testing"

	dbent "github.com/Wei-Shaw/cloudbase/ent"
	"github.com/Wei-Shaw/cloudbase/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestResolveProviderProductIDForCreemRecharge(t *testing.T) {
	t.Parallel()

	productID, err := resolveProviderProductID(
		CreateOrderRequest{
			PaymentType: payment.TypeCreem,
			ProductID:   "starter",
		},
		nil,
		&PaymentConfig{
			RechargeProducts: []RechargeProduct{
				{ID: "starter", CreemProductID: "prod_123"},
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, "prod_123", productID)
}

func TestResolveCreemRechargeProductUsesConfiguredCatalogValues(t *testing.T) {
	t.Parallel()

	product, err := resolveCreemRechargeProduct(
		CreateOrderRequest{
			PaymentType: payment.TypeCreem,
			ProductID:   "starter",
			Amount:      999,
		},
		&PaymentConfig{
			RechargeProducts: []RechargeProduct{
				{ID: "starter", Amount: 30, CreditedAmount: 45, CreemProductID: "prod_123"},
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, product)
	require.Equal(t, 30.0, product.Amount)
	require.Equal(t, 45.0, product.CreditedAmount)
	require.Equal(t, "prod_123", product.CreemProductID)
}

func TestResolveProviderProductIDForCreemSubscription(t *testing.T) {
	t.Parallel()

	productID, err := resolveProviderProductID(
		CreateOrderRequest{PaymentType: payment.TypeCreem},
		&dbent.SubscriptionPlan{CreemProductID: "plan_prod_456"},
		&PaymentConfig{},
	)
	require.NoError(t, err)
	require.Equal(t, "plan_prod_456", productID)
}

func TestResolveProviderProductIDForCreemRequiresConfiguredProduct(t *testing.T) {
	t.Parallel()

	_, err := resolveProviderProductID(
		CreateOrderRequest{
			PaymentType: payment.TypeCreem,
			ProductID:   "starter",
		},
		nil,
		&PaymentConfig{
			RechargeProducts: []RechargeProduct{
				{ID: "starter"},
			},
		},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CREEM_PRODUCT_NOT_CONFIGURED")
}

func TestResolveProviderProductIDSkipsNonCreemPayments(t *testing.T) {
	t.Parallel()

	productID, err := resolveProviderProductID(
		CreateOrderRequest{
			PaymentType: payment.TypeStripe,
			ProductID:   "starter",
		},
		nil,
		&PaymentConfig{},
	)
	require.NoError(t, err)
	require.Empty(t, productID)
}

func TestResolveCreemRechargeProductSkipsNonCreemPayments(t *testing.T) {
	t.Parallel()

	product, err := resolveCreemRechargeProduct(
		CreateOrderRequest{
			PaymentType: payment.TypeStripe,
			ProductID:   "starter",
		},
		&PaymentConfig{},
	)
	require.NoError(t, err)
	require.Nil(t, product)
}
