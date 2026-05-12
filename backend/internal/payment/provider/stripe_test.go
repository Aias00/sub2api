package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripeVerifyNotificationIgnoresNonPaymentEventsWithoutSignature(t *testing.T) {
	provider, err := NewStripe("stripe-test", map[string]string{
		"secretKey": "sk_test_123",
	})
	require.NoError(t, err)

	notification, err := provider.VerifyNotification(
		context.Background(),
		`{"id":"evt_123","object":"event","type":"capability.updated","data":{"object":{"id":"alipay_payments"}}}`,
		nil,
	)

	require.NoError(t, err)
	require.Nil(t, notification)
}

func TestStripeVerifyNotificationRequiresWebhookSecretForPaymentEvents(t *testing.T) {
	provider, err := NewStripe("stripe-test", map[string]string{
		"secretKey": "sk_test_123",
	})
	require.NoError(t, err)

	_, err = provider.VerifyNotification(
		context.Background(),
		`{"id":"evt_123","object":"event","type":"payment_intent.succeeded","data":{"object":{"id":"pi_123"}}}`,
		map[string]string{"stripe-signature": "t=123,v1=invalid"},
	)

	require.ErrorContains(t, err, "stripe webhookSecret not configured")
}
