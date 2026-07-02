//go:build unit

package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/cloudbase/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCreateOrderWithWaffoProviderInstance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	privateKey, publicKey := providerTestWaffoKeys(t)

	var createBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes := mustReadBodyForWaffo(t, r)
		switch r.URL.Path {
		case "/api/v1/order/create":
			require.NoError(t, json.Unmarshal(bodyBytes, &createBody))
			response := `{"code":"0","msg":"ok","data":{"paymentRequestId":"pay_req_e2e_1","merchantOrderId":"sub2_test","orderStatus":"AUTHORIZATION_REQUIRED","orderAction":"{\"actionType\":\"WEB\",\"webUrl\":\"https://waffo.test/pay/e2e\"}"}}`
			signature, err := signWaffoForServiceTest(response, privateKey)
			require.NoError(t, err)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-SIGNATURE", signature)
			_, _ = w.Write([]byte(response))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfgSvc := &PaymentConfigService{
		entClient: client,
		settingRepo: &paymentConfigSettingRepoStub{values: map[string]string{
			SettingPaymentEnabled:      "true",
			SettingEnabledPaymentTypes: payment.TypeWaffo,
			SettingMinRechargeAmount:   "1",
			SettingMaxRechargeAmount:   "0",
			SettingOrderTimeoutMinutes: "30",
			SettingMaxPendingOrders:    "3",
			SettingLoadBalanceStrategy: string(payment.DefaultLoadBalanceStrategy),
			SettingBalancePayDisabled:  "false",
			SettingDailyRechargeLimit:  "0",
			SettingBalanceRechargeMult: "1",
			SettingRechargeFeeRate:     "0",
			SettingRechargeProducts:    "[]",
		}},
		encryptionKey: []byte(webhookProviderTestEncryptionKey),
	}

	user, err := client.User.Create().
		SetEmail("waffo-e2e@example.com").
		SetPasswordHash("hash").
		SetUsername("waffo-e2e").
		SetRole(RoleUser).
		SetStatus(StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = cfgSvc.CreateProviderInstance(ctx, CreateProviderInstanceRequest{
		ProviderKey:    payment.TypeWaffo,
		Name:           "waffo-local-test",
		SupportedTypes: []string{payment.TypeWaffo},
		Config: map[string]string{
			"apiKey":         "waffo_api",
			"privateKey":     privateKey,
			"waffoPublicKey": publicKey,
			"apiBase":        server.URL + "/api/v1",
			"merchantId":     "merchant_local_1",
			"currency":       "USD",
			"notifyUrl":      "https://merchant.example/api/v1/payment/webhook/waffo",
			"returnUrl":      "https://merchant.example/payment/result",
		},
		Enabled:     true,
		PaymentMode: "redirect",
	})
	require.NoError(t, err)

	userRepo := &mockUserRepo{
		getByIDUser: &User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Status:   payment.EntityStatusActive,
		},
	}

	loadBalancer := payment.NewDefaultLoadBalancer(client, []byte(webhookProviderTestEncryptionKey))
	svc := &PaymentService{
		entClient:       client,
		registry:        payment.NewRegistry(),
		loadBalancer:    newVisibleMethodLoadBalancer(loadBalancer, cfgSvc),
		configService:   cfgSvc,
		userRepo:        userRepo,
		providersLoaded: true,
	}

	resp, err := svc.CreateOrder(ctx, CreateOrderRequest{
		UserID:      user.ID,
		Amount:      12.88,
		PaymentType: payment.TypeWaffo,
		OrderType:   payment.OrderTypeBalance,
	})
	require.NoError(t, err)
	require.Equal(t, payment.TypeWaffo, resp.PaymentType)
	require.Equal(t, "https://waffo.test/pay/e2e", resp.PayURL)
	require.NotZero(t, resp.OrderID)
	require.NotEmpty(t, resp.OutTradeNo)
	merchantOrderID, _ := createBody["merchantOrderId"].(string)
	require.True(t, strings.HasPrefix(merchantOrderID, "sub2_"))
	require.Equal(t, resp.OutTradeNo, merchantOrderID)
	require.Equal(t, "12.88", createBody["orderAmount"])
	require.Equal(t, "waffo-e2e@example.com", createBody["userInfo"].(map[string]any)["userEmail"])
}

func providerTestWaffoKeys(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	return base64.StdEncoding.EncodeToString(privDER), base64.StdEncoding.EncodeToString(pubDER)
}

func signWaffoForServiceTest(data string, base64PrivateKey string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		return "", err
	}
	privateKeyAny, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return "", err
	}
	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA")
	}
	hashed := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func mustReadBodyForWaffo(t *testing.T, r *http.Request) []byte {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	return body
}
