package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	creemSignatureHeader = "creem-signature"
	creemDefaultAPIBase  = "https://api.creem.io"
	creemDefaultTestBase = "https://test-api.creem.io"
)

type Creem struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
}

func NewCreem(instanceID string, config map[string]string) (*Creem, error) {
	if strings.TrimSpace(config["apiKey"]) == "" {
		return nil, fmt.Errorf("creem config missing required key: apiKey")
	}
	if strings.TrimSpace(config["webhookSecret"]) == "" {
		return nil, fmt.Errorf("creem config missing required key: webhookSecret")
	}
	return &Creem{
		instanceID: instanceID,
		config:     cloneConfig(config),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Creem) Name() string        { return "Creem" }
func (c *Creem) ProviderKey() string { return payment.TypeCreem }
func (c *Creem) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeCreem}
}

func (c *Creem) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	productID := strings.TrimSpace(req.ProviderProductID)
	if productID == "" {
		return nil, fmt.Errorf("creem create payment: missing provider product id")
	}

	requestBody := map[string]any{
		"product_id": productID,
		"request_id": req.OrderID,
		"customer": map[string]any{
			"email": strings.TrimSpace(req.CustomerEmail),
		},
		"metadata": map[string]string{
			"order_id": req.OrderID,
		},
	}
	if requestBody["customer"].(map[string]any)["email"] == "" {
		return nil, fmt.Errorf("creem create payment: missing customer email")
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("creem create payment: marshal body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.apiBase(), "/")+"/v1/checkouts", strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("creem create payment: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config["apiKey"])

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("creem create payment: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("creem create payment: read response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("creem create payment: http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		ID          string `json:"id"`
		CheckoutURL string `json:"checkout_url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("creem create payment: parse response: %w", err)
	}
	if strings.TrimSpace(result.CheckoutURL) == "" {
		return nil, fmt.Errorf("creem create payment: missing checkout_url")
	}
	return &payment.CreatePaymentResponse{
		TradeNo: result.ID,
		PayURL:  result.CheckoutURL,
	}, nil
}

func (c *Creem) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	if strings.TrimSpace(tradeNo) == "" {
		return nil, fmt.Errorf("creem query order: missing checkout id")
	}
	queryURL := strings.TrimRight(c.apiBase(), "/") + "/v1/checkouts?checkout_id=" + url.QueryEscape(strings.TrimSpace(tradeNo))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creem query order: create request: %w", err)
	}
	httpReq.Header.Set("x-api-key", c.config["apiKey"])

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("creem query order: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("creem query order: read response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("creem query order: http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Object struct {
			RequestID string `json:"request_id"`
			Order     struct {
				ID         string `json:"id"`
				Status     string `json:"status"`
				AmountPaid int64  `json:"amount_paid"`
				Currency   string `json:"currency"`
			} `json:"order"`
		} `json:"object"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("creem query order: parse response: %w", err)
	}

	status := payment.ProviderStatusPending
	switch strings.ToLower(strings.TrimSpace(result.Object.Order.Status)) {
	case "paid":
		status = payment.ProviderStatusPaid
	case "failed", "cancelled", "expired":
		status = payment.ProviderStatusFailed
	}

	return &payment.QueryOrderResponse{
		TradeNo: result.ID,
		Status:  status,
		Amount:  float64(result.Object.Order.AmountPaid) / 100.0,
		Metadata: map[string]string{
			"currency": strings.TrimSpace(result.Object.Order.Currency),
		},
	}, nil
}

func (c *Creem) VerifyNotification(_ context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	signature := headerValue(headers, creemSignatureHeader)
	if signature == "" {
		return nil, fmt.Errorf("creem verify notification: missing creem-signature header")
	}
	if !verifyCreemSignature(rawBody, signature, c.config["webhookSecret"]) {
		return nil, fmt.Errorf("creem verify notification: invalid signature")
	}

	var event struct {
		EventType string `json:"eventType"`
		Object    struct {
			ID        string `json:"id"`
			RequestID string `json:"request_id"`
			Order     struct {
				ID         string `json:"id"`
				Type       string `json:"type"`
				Status     string `json:"status"`
				AmountPaid int64  `json:"amount_paid"`
				Currency   string `json:"currency"`
			} `json:"order"`
		} `json:"object"`
	}
	if err := json.Unmarshal([]byte(rawBody), &event); err != nil {
		return nil, fmt.Errorf("creem verify notification: parse body: %w", err)
	}
	if event.EventType != "checkout.completed" {
		return nil, nil
	}
	if strings.ToLower(strings.TrimSpace(event.Object.Order.Status)) != "paid" {
		return nil, nil
	}
	return &payment.PaymentNotification{
		TradeNo: event.Object.ID,
		OrderID: event.Object.RequestID,
		Amount:  float64(event.Object.Order.AmountPaid) / 100.0,
		Status:  payment.NotificationStatusSuccess,
		RawData: rawBody,
		Metadata: map[string]string{
			"currency": strings.TrimSpace(event.Object.Order.Currency),
		},
	}, nil
}

func (c *Creem) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("creem refund is not implemented")
}

func (c *Creem) apiBase() string {
	if v := strings.TrimSpace(c.config["apiBase"]); v != "" {
		return v
	}
	if strings.EqualFold(strings.TrimSpace(c.config["testMode"]), "true") {
		return creemDefaultTestBase
	}
	return creemDefaultAPIBase
}

func verifyCreemSignature(payload string, signature string, secret string) bool {
	if strings.TrimSpace(secret) == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.TrimSpace(signature)), []byte(expected))
}

func cloneConfig(config map[string]string) map[string]string {
	cp := make(map[string]string, len(config))
	for k, v := range config {
		cp[k] = v
	}
	return cp
}

func headerValue(headers map[string]string, key string) string {
	for k, v := range headers {
		if strings.EqualFold(strings.TrimSpace(k), key) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
