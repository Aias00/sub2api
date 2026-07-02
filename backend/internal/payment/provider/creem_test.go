package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/cloudbase/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestNewCreemRequiresKeys(t *testing.T) {
	_, err := NewCreem("creem-test", map[string]string{})
	require.ErrorContains(t, err, "apiKey")

	_, err = NewCreem("creem-test", map[string]string{"apiKey": "ck_test"})
	require.ErrorContains(t, err, "webhookSecret")
}

func TestCreemCreatePaymentAndQueryOrder(t *testing.T) {
	var createAuth string
	var createBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/checkouts":
			createAuth = r.Header.Get("x-api-key")
			require.NoError(t, json.NewDecoder(r.Body).Decode(&createBody))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"chk_123","checkout_url":"https://checkout.creem.test/session"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/checkouts":
			if r.URL.Query().Get("checkout_id") != "chk_123" {
				t.Fatalf("checkout_id = %q", r.URL.Query().Get("checkout_id"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"chk_123","status":"completed","object":{"request_id":"sub2_123","order":{"id":"ord_1","status":"paid","amount_paid":1288,"currency":"USD"}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider, err := NewCreem("creem-test", map[string]string{
		"apiKey":        "creem_api",
		"webhookSecret": "whsec_123",
		"apiBase":       server.URL,
	})
	require.NoError(t, err)

	resp, err := provider.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:           "sub2_123",
		ProviderProductID: "prod_abc",
		CustomerEmail:     "user@example.com",
	})
	require.NoError(t, err)
	require.Equal(t, "creem_api", createAuth)
	require.Equal(t, "prod_abc", createBody["product_id"])
	require.Equal(t, "sub2_123", createBody["request_id"])
	customer, ok := createBody["customer"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "user@example.com", customer["email"])
	require.Equal(t, "chk_123", resp.TradeNo)
	require.Equal(t, "https://checkout.creem.test/session", resp.PayURL)

	status, err := provider.QueryOrder(context.Background(), "chk_123")
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusPaid, status.Status)
	require.InEpsilon(t, 12.88, status.Amount, 0.001)
}

func TestCreemVerifyNotification(t *testing.T) {
	provider, err := NewCreem("creem-test", map[string]string{
		"apiKey":        "creem_api",
		"webhookSecret": "whsec_123",
	})
	require.NoError(t, err)

	raw := `{"eventType":"checkout.completed","object":{"id":"chk_123","request_id":"sub2_123","order":{"id":"ord_1","type":"onetime","status":"paid","amount_paid":1288,"currency":"USD"}}}`
	mac := hmac.New(sha256.New, []byte("whsec_123"))
	_, _ = mac.Write([]byte(raw))
	signature := hex.EncodeToString(mac.Sum(nil))

	notification, err := provider.VerifyNotification(context.Background(), raw, map[string]string{"creem-signature": signature})
	require.NoError(t, err)
	require.Equal(t, "chk_123", notification.TradeNo)
	require.Equal(t, "sub2_123", notification.OrderID)
	require.Equal(t, payment.NotificationStatusSuccess, notification.Status)
	require.InEpsilon(t, 12.88, notification.Amount, 0.001)
}
