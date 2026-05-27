package provider

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
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	waffoAPIBaseProduction = "https://api.waffo.com/api/v1"
	waffoAPIBaseSandbox    = "https://api-sandbox.waffo.com/api/v1"
	waffoHeaderAPIKey      = "X-API-KEY"
	waffoHeaderSignature   = "X-SIGNATURE"
	waffoHeaderAPIVersion  = "X-API-VERSION"
	waffoHeaderSDKVersion  = "X-SDK-VERSION"
	waffoAPIVersion        = "1.0.0"
	waffoSDKVersion        = "sub2api/waffo"
	waffoPaySuccess        = "PAY_SUCCESS"
	waffoOrderClosed       = "ORDER_CLOSE"
	waffoOrderAuthRequired = "AUTHORIZATION_REQUIRED"
	waffoOrderPending      = "PAY_IN_PROGRESS"
)

var waffoZeroDecimalCurrencies = map[string]struct{}{
	"IDR": {},
	"JPY": {},
	"KRW": {},
	"VND": {},
}

type Waffo struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
}

func NewWaffo(instanceID string, config map[string]string) (*Waffo, error) {
	for _, key := range []string{"apiKey", "privateKey", "waffoPublicKey"} {
		if strings.TrimSpace(config[key]) == "" {
			return nil, fmt.Errorf("waffo config missing required key: %s", key)
		}
	}
	if err := validateWaffoPrivateKey(config["privateKey"]); err != nil {
		return nil, err
	}
	if err := validateWaffoPublicKey(config["waffoPublicKey"]); err != nil {
		return nil, err
	}
	return &Waffo{
		instanceID: instanceID,
		config:     cloneConfig(config),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (w *Waffo) Name() string        { return "Waffo" }
func (w *Waffo) ProviderKey() string { return payment.TypeWaffo }
func (w *Waffo) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeWaffo}
}

func (w *Waffo) MerchantIdentityMetadata() map[string]string {
	merchantID := strings.TrimSpace(w.config["merchantId"])
	if merchantID == "" {
		return nil
	}
	return map[string]string{"merchant_id": merchantID}
}

func (w *Waffo) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	if strings.TrimSpace(req.CustomerEmail) == "" {
		return nil, fmt.Errorf("waffo create payment: missing customer email")
	}
	requestedAt := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	payload := map[string]any{
		"paymentRequestId":   req.OrderID,
		"merchantOrderId":    req.OrderID,
		"orderCurrency":      w.currency(),
		"orderAmount":        formatWaffoAmount(req.Amount, w.currency()),
		"orderDescription":   req.Subject,
		"orderRequestedAt":   requestedAt,
		"notifyUrl":          w.resolveConfigURL("notifyUrl", req.NotifyURL),
		"successRedirectUrl": w.resolveConfigURL("returnUrl", req.ReturnURL),
		"failedRedirectUrl":  w.resolveConfigURL("returnUrl", req.ReturnURL),
		"merchantInfo": map[string]any{
			"merchantId": strings.TrimSpace(w.config["merchantId"]),
		},
		"userInfo": map[string]any{
			"userId":       req.OrderID,
			"userEmail":    strings.TrimSpace(req.CustomerEmail),
			"userTerminal": "WEB",
		},
		"paymentInfo": map[string]any{
			"productName": "ONE_TIME_PAYMENT",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("waffo create payment: marshal body: %w", err)
	}
	respBody, err := w.doSignedJSON(ctx, http.MethodPost, "/order/create", body)
	if err != nil {
		return nil, fmt.Errorf("waffo create payment: %w", err)
	}

	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			PaymentRequestID string `json:"paymentRequestId"`
			MerchantOrderID  string `json:"merchantOrderId"`
			OrderStatus      string `json:"orderStatus"`
			OrderAction      string `json:"orderAction"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("waffo create payment: parse response: %w", err)
	}
	if strings.TrimSpace(result.Code) != "0" {
		return nil, fmt.Errorf("waffo create payment: %s", strings.TrimSpace(result.Msg))
	}
	return &payment.CreatePaymentResponse{
		TradeNo: result.Data.PaymentRequestID,
		PayURL:  parseWaffoOrderActionURL(result.Data.OrderAction),
	}, nil
}

func (w *Waffo) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	payload := map[string]any{
		"paymentRequestId": strings.TrimSpace(tradeNo),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("waffo query order: marshal body: %w", err)
	}
	respBody, err := w.doSignedJSON(ctx, http.MethodPost, "/order/inquiry", body)
	if err != nil {
		return nil, fmt.Errorf("waffo query order: %w", err)
	}

	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			PaymentRequestID string `json:"paymentRequestId"`
			MerchantOrderID  string `json:"merchantOrderId"`
			OrderStatus      string `json:"orderStatus"`
			OrderAmount      string `json:"orderAmount"`
			FinalDealAmount  string `json:"finalDealAmount"`
			OrderCurrency    string `json:"orderCurrency"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("waffo query order: parse response: %w", err)
	}
	if strings.TrimSpace(result.Code) != "0" {
		return nil, fmt.Errorf("waffo query order: %s", strings.TrimSpace(result.Msg))
	}

	status := payment.ProviderStatusPending
	switch strings.TrimSpace(result.Data.OrderStatus) {
	case waffoPaySuccess:
		status = payment.ProviderStatusPaid
	case waffoOrderClosed:
		status = payment.ProviderStatusFailed
	case waffoOrderAuthRequired, waffoOrderPending:
		status = payment.ProviderStatusPending
	}
	amount := parseWaffoAmount(result.Data.FinalDealAmount)
	if amount <= 0 {
		amount = parseWaffoAmount(result.Data.OrderAmount)
	}

	return &payment.QueryOrderResponse{
		TradeNo: result.Data.PaymentRequestID,
		Status:  status,
		Amount:  amount,
		Metadata: map[string]string{
			"merchant_id": strings.TrimSpace(w.config["merchantId"]),
			"currency":    strings.TrimSpace(result.Data.OrderCurrency),
		},
	}, nil
}

func (w *Waffo) VerifyNotification(_ context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	signature := headerValue(headers, waffoHeaderSignature)
	if signature == "" {
		return nil, fmt.Errorf("waffo verify notification: missing X-SIGNATURE header")
	}
	if !verifyWaffoSignature(rawBody, signature, w.config["waffoPublicKey"]) {
		return nil, fmt.Errorf("waffo verify notification: invalid signature")
	}

	var event struct {
		EventType string `json:"eventType"`
		Result    struct {
			PaymentRequestID string `json:"paymentRequestId"`
			MerchantOrderID  string `json:"merchantOrderId"`
			OrderStatus      string `json:"orderStatus"`
			OrderAmount      string `json:"orderAmount"`
			OrderCurrency    string `json:"orderCurrency"`
			MerchantInfo     struct {
				MerchantID string `json:"merchantId"`
			} `json:"merchantInfo"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(rawBody), &event); err != nil {
		return nil, fmt.Errorf("waffo verify notification: parse body: %w", err)
	}
	if strings.TrimSpace(event.EventType) != "PAYMENT_NOTIFICATION" {
		return nil, nil
	}
	if strings.TrimSpace(event.Result.OrderStatus) != waffoPaySuccess {
		return nil, nil
	}
	return &payment.PaymentNotification{
		TradeNo: event.Result.PaymentRequestID,
		OrderID: event.Result.MerchantOrderID,
		Amount:  parseWaffoAmount(event.Result.OrderAmount),
		Status:  payment.NotificationStatusSuccess,
		RawData: rawBody,
		Metadata: map[string]string{
			"merchant_id": event.Result.MerchantInfo.MerchantID,
			"currency":    event.Result.OrderCurrency,
		},
	}, nil
}

func (w *Waffo) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("waffo refund is not implemented")
}

func (w *Waffo) BuildWebhookSuccessResponse() (int, string, map[string]string, string) {
	body := `{"message":"success"}`
	signature, _ := signWaffo(body, w.config["privateKey"])
	return http.StatusOK, body, map[string]string{waffoHeaderSignature: signature}, "application/json"
}

func (w *Waffo) apiBase() string {
	if apiBase := strings.TrimSpace(w.config["apiBase"]); apiBase != "" {
		return strings.TrimRight(apiBase, "/")
	}
	if strings.EqualFold(strings.TrimSpace(w.config["sandbox"]), "true") {
		return waffoAPIBaseSandbox
	}
	return waffoAPIBaseProduction
}

func (w *Waffo) currency() string {
	if currency := strings.TrimSpace(w.config["currency"]); currency != "" {
		return strings.ToUpper(currency)
	}
	return "USD"
}

func (w *Waffo) resolveConfigURL(configKey, fallback string) string {
	if value := strings.TrimSpace(w.config[configKey]); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func (w *Waffo) doSignedJSON(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(w.apiBase(), "/")+path, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	signature, err := signWaffo(string(body), w.config["privateKey"])
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(waffoHeaderAPIKey, w.config["apiKey"])
	req.Header.Set(waffoHeaderSignature, signature)
	req.Header.Set(waffoHeaderAPIVersion, waffoAPIVersion)
	req.Header.Set(waffoHeaderSDKVersion, waffoSDKVersion)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	respSignature := strings.TrimSpace(resp.Header.Get(waffoHeaderSignature))
	if respSignature != "" && !verifyWaffoSignature(string(respBody), respSignature, w.config["waffoPublicKey"]) {
		return nil, fmt.Errorf("waffo response signature verification failed")
	}
	return respBody, nil
}

func parseWaffoOrderActionURL(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	var action struct {
		ActionType  string `json:"actionType"`
		WebURL      string `json:"webUrl"`
		DeeplinkURL string `json:"deeplinkUrl"`
	}
	if err := json.Unmarshal([]byte(raw), &action); err != nil {
		return ""
	}
	if strings.EqualFold(action.ActionType, "DEEPLINK") && strings.TrimSpace(action.DeeplinkURL) != "" {
		return strings.TrimSpace(action.DeeplinkURL)
	}
	return strings.TrimSpace(action.WebURL)
}

func parseWaffoAmount(raw string) float64 {
	value, _ := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	return value
}

func formatWaffoAmount(raw string, currency string) string {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return strings.TrimSpace(raw)
	}
	if _, ok := waffoZeroDecimalCurrencies[strings.ToUpper(strings.TrimSpace(currency))]; ok {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func signWaffo(data string, base64PrivateKey string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64PrivateKey))
	if err != nil {
		return "", fmt.Errorf("waffo sign: decode private key: %w", err)
	}
	privateKeyAny, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return "", fmt.Errorf("waffo sign: parse private key: %w", err)
	}
	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("waffo sign: private key is not RSA")
	}
	hashed := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("waffo sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func verifyWaffoSignature(data string, base64Signature string, base64PublicKey string) bool {
	keyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64PublicKey))
	if err != nil {
		return false
	}
	publicKeyAny, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return false
	}
	publicKey, ok := publicKeyAny.(*rsa.PublicKey)
	if !ok {
		return false
	}
	signatureBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64Signature))
	if err != nil {
		return false
	}
	hashed := sha256.Sum256([]byte(data))
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signatureBytes) == nil
}

func validateWaffoPrivateKey(base64PrivateKey string) error {
	keyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64PrivateKey))
	if err != nil {
		return fmt.Errorf("waffo private key is invalid base64: %w", err)
	}
	privateKeyAny, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return fmt.Errorf("waffo private key is not valid PKCS#8: %w", err)
	}
	if _, ok := privateKeyAny.(*rsa.PrivateKey); !ok {
		return fmt.Errorf("waffo private key is not RSA")
	}
	return nil
}

func validateWaffoPublicKey(base64PublicKey string) error {
	keyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64PublicKey))
	if err != nil {
		return fmt.Errorf("waffo public key is invalid base64: %w", err)
	}
	publicKeyAny, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return fmt.Errorf("waffo public key is not valid X.509: %w", err)
	}
	if _, ok := publicKeyAny.(*rsa.PublicKey); !ok {
		return fmt.Errorf("waffo public key is not RSA")
	}
	return nil
}
