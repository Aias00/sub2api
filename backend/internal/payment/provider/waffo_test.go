package provider

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestNewWaffoValidatesKeys(t *testing.T) {
	_, err := NewWaffo("waffo-test", map[string]string{})
	require.ErrorContains(t, err, "apiKey")
}

func TestWaffoCreateQueryAndWebhook(t *testing.T) {
	privateKey, publicKey := generateWaffoTestKeys(t)
	var createHeaders http.Header
	var createBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes := mustReadBody(t, r)
		switch r.URL.Path {
		case "/api/v1/order/create":
			createHeaders = r.Header.Clone()
			require.NoError(t, json.Unmarshal(bodyBytes, &createBody))
			response := `{"code":"0","msg":"ok","data":{"paymentRequestId":"pay_req_1","merchantOrderId":"sub2_123","orderStatus":"AUTHORIZATION_REQUIRED","orderAction":"{\"actionType\":\"WEB\",\"webUrl\":\"https://waffo.test/pay\"}"}}`
			signature, err := signWaffo(response, privateKey)
			require.NoError(t, err)
			w.Header().Set(waffoHeaderSignature, signature)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(response))
		case "/api/v1/order/inquiry":
			response := `{"code":"0","msg":"ok","data":{"paymentRequestId":"pay_req_1","merchantOrderId":"sub2_123","orderStatus":"PAY_SUCCESS","finalDealAmount":"12.88","orderCurrency":"USD"}}`
			signature, err := signWaffo(response, privateKey)
			require.NoError(t, err)
			w.Header().Set(waffoHeaderSignature, signature)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(response))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider, err := NewWaffo("waffo-test", map[string]string{
		"apiKey":         "waffo_api",
		"privateKey":     privateKey,
		"waffoPublicKey": publicKey,
		"apiBase":        server.URL + "/api/v1",
		"sandbox":        "false",
	})
	require.NoError(t, err)
	provider.httpClient = server.Client()

	resp, err := provider.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:       "sub2_123",
		Amount:        "12.88",
		Subject:       "Balance recharge",
		CustomerEmail: "user@example.com",
		CustomerName:  "User",
		NotifyURL:     "https://merchant.example/api/v1/payment/webhook/waffo",
		ReturnURL:     "https://merchant.example/payment/result",
		PaymentType:   payment.TypeWaffo,
	})
	require.NoError(t, err)
	require.Equal(t, "waffo_api", createHeaders.Get(waffoHeaderAPIKey))
	require.NotEmpty(t, createHeaders.Get(waffoHeaderSignature))
	require.Equal(t, "sub2_123", createBody["merchantOrderId"])
	require.Equal(t, "12.88", createBody["orderAmount"])
	require.Equal(t, "user@example.com", createBody["userInfo"].(map[string]any)["userEmail"])
	require.Equal(t, "pay_req_1", resp.TradeNo)
	require.Equal(t, "https://waffo.test/pay", resp.PayURL)

	qr, err := provider.QueryOrder(context.Background(), "pay_req_1")
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusPaid, qr.Status)
	require.InEpsilon(t, 12.88, qr.Amount, 0.001)

	webhookBody := `{"eventType":"PAYMENT_NOTIFICATION","result":{"paymentRequestId":"pay_req_1","merchantOrderId":"sub2_123","orderStatus":"PAY_SUCCESS","orderAmount":"12.88","orderCurrency":"USD","merchantInfo":{"merchantId":"m_1"}}}`
	webhookSig, err := signWaffo(webhookBody, privateKey)
	require.NoError(t, err)
	notification, err := provider.VerifyNotification(context.Background(), webhookBody, map[string]string{"x-signature": webhookSig})
	require.NoError(t, err)
	require.Equal(t, "sub2_123", notification.OrderID)
	require.Equal(t, payment.NotificationStatusSuccess, notification.Status)

	status, body, headers, contentType := provider.BuildWebhookSuccessResponse()
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "application/json", contentType)
	require.Contains(t, body, `"success"`)
	require.NotEmpty(t, headers[waffoHeaderSignature])
}

func TestFormatWaffoAmountHonorsZeroDecimalCurrencies(t *testing.T) {
	t.Parallel()

	require.Equal(t, "13", formatWaffoAmount("12.88", "JPY"))
	require.Equal(t, "12.88", formatWaffoAmount("12.88", "USD"))
}

func mustReadBody(t *testing.T, r *http.Request) []byte {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	return body
}

func generateWaffoTestKeys(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	_ = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})
	_ = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	return base64.StdEncoding.EncodeToString(privDER), base64.StdEncoding.EncodeToString(pubDER)
}
