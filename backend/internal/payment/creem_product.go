package payment

import (
	"strings"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
)

func ResolveCreemRechargeProduct(paymentType, orderType, selectedProductID string, products []RechargeProduct) (*RechargeProduct, error) {
	if GetBasePaymentType(paymentType) != TypeCreem || orderType == OrderTypeSubscription {
		return nil, nil
	}

	selectedProductID = strings.TrimSpace(selectedProductID)
	if selectedProductID == "" {
		return nil, infraerrors.BadRequest("CREEM_PRODUCT_REQUIRED", "creem recharge requires selecting a configured recharge product")
	}
	for idx := range products {
		product := &products[idx]
		if strings.TrimSpace(product.ID) != selectedProductID {
			continue
		}
		if strings.TrimSpace(product.CreemProductID) == "" {
			return nil, infraerrors.BadRequest("CREEM_PRODUCT_NOT_CONFIGURED", "selected recharge product is not configured for Creem")
		}
		return product, nil
	}
	return nil, infraerrors.BadRequest("CREEM_PRODUCT_NOT_FOUND", "selected recharge product does not exist")
}

func ResolveCreemProviderProductID(paymentType string, subscriptionProductID string, hasSubscriptionPlan bool, rechargeProduct *RechargeProduct) (string, error) {
	if GetBasePaymentType(paymentType) != TypeCreem {
		return "", nil
	}

	if hasSubscriptionPlan {
		productID := strings.TrimSpace(subscriptionProductID)
		if productID == "" {
			return "", infraerrors.BadRequest("CREEM_PRODUCT_NOT_CONFIGURED", "selected subscription plan is not configured for Creem")
		}
		return productID, nil
	}

	if rechargeProduct == nil {
		return "", infraerrors.BadRequest("CREEM_PRODUCT_REQUIRED", "creem recharge requires selecting a configured recharge product")
	}
	productID := strings.TrimSpace(rechargeProduct.CreemProductID)
	if productID == "" {
		return "", infraerrors.BadRequest("CREEM_PRODUCT_NOT_CONFIGURED", "selected recharge product is not configured for Creem")
	}
	return productID, nil
}
