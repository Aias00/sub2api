package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/cloudbase/internal/payment"
	"github.com/Wei-Shaw/cloudbase/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentWebhookHandler handles payment provider webhook callbacks.
type PaymentWebhookHandler struct {
	paymentService *service.PaymentService
	registry       *payment.Registry
}

// maxWebhookBodySize is the maximum allowed webhook request body size (1 MB).
const maxWebhookBodySize = 1 << 20

// webhookLogTruncateLen is the maximum length of raw body logged on verify failure.
const webhookLogTruncateLen = 200

// NewPaymentWebhookHandler creates a new PaymentWebhookHandler.
func NewPaymentWebhookHandler(paymentService *service.PaymentService, registry *payment.Registry) *PaymentWebhookHandler {
	return &PaymentWebhookHandler{
		paymentService: paymentService,
		registry:       registry,
	}
}

// EasyPayNotify handles EasyPay payment notifications.
// POST /api/v1/payment/webhook/easypay
func (h *PaymentWebhookHandler) EasyPayNotify(c *gin.Context) {
	h.handleNotify(c, payment.TypeEasyPay)
}

// AlipayNotify handles Alipay payment notifications.
// POST /api/v1/payment/webhook/alipay
func (h *PaymentWebhookHandler) AlipayNotify(c *gin.Context) {
	h.handleNotify(c, payment.TypeAlipay)
}

// WxpayNotify handles WeChat Pay payment notifications.
// POST /api/v1/payment/webhook/wxpay
func (h *PaymentWebhookHandler) WxpayNotify(c *gin.Context) {
	h.handleNotify(c, payment.TypeWxpay)
}

// StripeWebhook handles Stripe webhook events.
// POST /api/v1/payment/webhook/stripe
func (h *PaymentWebhookHandler) StripeWebhook(c *gin.Context) {
	h.handleNotify(c, payment.TypeStripe)
}

// CreemWebhook handles Creem webhook events.
// POST /api/v1/payment/webhook/creem
func (h *PaymentWebhookHandler) CreemWebhook(c *gin.Context) {
	h.handleNotify(c, payment.TypeCreem)
}

// WaffoWebhook handles Waffo webhook events.
// POST /api/v1/payment/webhook/waffo
func (h *PaymentWebhookHandler) WaffoWebhook(c *gin.Context) {
	h.handleNotify(c, payment.TypeWaffo)
}

// handleNotify is the shared logic for all provider webhook handlers.
func (h *PaymentWebhookHandler) handleNotify(c *gin.Context, providerKey string) {
	var rawBody string
	if c.Request.Method == http.MethodGet {
		// GET callbacks (e.g. EasyPay) pass params as URL query string
		rawBody = c.Request.URL.RawQuery
	} else {
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBodySize))
		if err != nil {
			slog.Error("[Payment Webhook] failed to read body", "provider", providerKey, "error", err)
			c.String(http.StatusBadRequest, "failed to read body")
			return
		}
		rawBody = string(body)
	}

	// Extract out_trade_no to look up the order's specific provider instance.
	// This is needed when multiple instances of the same provider exist (e.g. multiple EasyPay accounts).
	outTradeNo := extractOutTradeNo(rawBody, providerKey)

	providers, err := h.paymentService.GetWebhookProviders(c.Request.Context(), providerKey, outTradeNo)
	if err != nil {
		slog.Warn("[Payment Webhook] provider not found", "provider", providerKey, "outTradeNo", outTradeNo, "error", err)
		if providerKey == payment.TypeWxpay {
			c.String(http.StatusBadRequest, "verify failed")
			return
		}
		writeSuccessResponse(c, providerKey, nil)
		return
	}

	headers := make(map[string]string)
	for k := range c.Request.Header {
		headers[strings.ToLower(k)] = c.GetHeader(k)
	}

	resolvedProviderKey, verifiedProvider, notification, err := verifyNotificationWithProviders(c.Request.Context(), providers, rawBody, headers)
	if err != nil {
		truncatedBody := rawBody
		if len(truncatedBody) > webhookLogTruncateLen {
			truncatedBody = truncatedBody[:webhookLogTruncateLen] + "...(truncated)"
		}
		slog.Error("[Payment Webhook] verify failed", "provider", providerKey, "error", err, "method", c.Request.Method, "bodyLen", len(rawBody))
		slog.Debug("[Payment Webhook] verify failed body", "provider", providerKey, "rawBody", truncatedBody)
		c.String(http.StatusBadRequest, "verify failed")
		return
	}

	// nil notification means irrelevant event (e.g. Stripe non-payment event); return success.
	if notification == nil {
		writeSuccessResponse(c, resolvedProviderKey, verifiedProvider)
		return
	}

	if err := h.paymentService.HandlePaymentNotification(c.Request.Context(), notification, resolvedProviderKey); err != nil {
		// Unknown order: ack with 2xx so the provider stops retrying. This
		// guards against foreign environments whose webhook endpoints are
		// (mis)configured to point at us — without a 2xx, the provider will
		// retry for days and spam our error logs. We still emit a WARN so the
		// event is discoverable in logs.
		if errors.Is(err, service.ErrOrderNotFound) {
			slog.Warn("[Payment Webhook] unknown order, acking to stop retries",
				"provider", resolvedProviderKey,
				"outTradeNo", notification.OrderID,
				"tradeNo", notification.TradeNo,
			)
			writeSuccessResponse(c, resolvedProviderKey, verifiedProvider)
			return
		}
		slog.Error("[Payment Webhook] handle notification failed", "provider", resolvedProviderKey, "error", err)
		c.String(http.StatusInternalServerError, "handle failed")
		return
	}

	writeSuccessResponse(c, resolvedProviderKey, verifiedProvider)
}

// extractOutTradeNo parses the webhook body to find the out_trade_no.
// This allows looking up the correct provider instance before verification.
func extractOutTradeNo(rawBody, providerKey string) string {
	switch providerKey {
	case payment.TypeEasyPay, payment.TypeAlipay:
		values, err := url.ParseQuery(rawBody)
		if err == nil {
			return values.Get("out_trade_no")
		}
	case payment.TypeCreem:
		var payload struct {
			Object struct {
				RequestID string `json:"request_id"`
				Metadata  struct {
					OrderID string `json:"order_id"`
				} `json:"metadata"`
			} `json:"object"`
		}
		if err := json.Unmarshal([]byte(rawBody), &payload); err == nil {
			if requestID := strings.TrimSpace(payload.Object.RequestID); requestID != "" {
				return requestID
			}
			return strings.TrimSpace(payload.Object.Metadata.OrderID)
		}
	case payment.TypeWaffo:
		var payload struct {
			Result struct {
				MerchantOrderID  string `json:"merchantOrderId"`
				PaymentRequestID string `json:"paymentRequestId"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(rawBody), &payload); err == nil {
			if merchantOrderID := strings.TrimSpace(payload.Result.MerchantOrderID); merchantOrderID != "" {
				return merchantOrderID
			}
			return strings.TrimSpace(payload.Result.PaymentRequestID)
		}
	}
	// For other providers (Stripe, Alipay direct, WxPay direct), the registry
	// typically has only one instance, so no instance lookup is needed.
	return ""
}

func verifyNotificationWithProviders(ctx context.Context, providers []payment.Provider, rawBody string, headers map[string]string) (string, payment.Provider, *payment.PaymentNotification, error) {
	var lastErr error
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		notification, err := provider.VerifyNotification(ctx, rawBody, headers)
		if err != nil {
			lastErr = err
			continue
		}
		return provider.ProviderKey(), provider, notification, nil
	}
	if lastErr != nil {
		return "", nil, nil, lastErr
	}
	return "", nil, nil, fmt.Errorf("no webhook provider could verify notification")
}

// wxpaySuccessResponse is the JSON response expected by WeChat Pay webhook.
type wxpaySuccessResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WeChat Pay webhook success response constants.
const (
	wxpaySuccessCode    = "SUCCESS"
	wxpaySuccessMessage = "成功"
)

// writeSuccessResponse sends the provider-specific success response.
// WeChat Pay requires JSON {"code":"SUCCESS","message":"成功"};
// Stripe expects an empty 200; others accept plain text "success".
func writeSuccessResponse(c *gin.Context, providerKey string, prov payment.Provider) {
	if responder, ok := prov.(payment.WebhookResponseProvider); ok {
		status, body, headers, contentType := responder.BuildWebhookSuccessResponse()
		for key, value := range headers {
			c.Header(key, value)
		}
		c.Data(status, contentType, []byte(body))
		return
	}
	switch providerKey {
	case payment.TypeWxpay:
		c.JSON(http.StatusOK, wxpaySuccessResponse{Code: wxpaySuccessCode, Message: wxpaySuccessMessage})
	case payment.TypeStripe, payment.TypeAirwallex:
		c.String(http.StatusOK, "")
	default:
		c.String(http.StatusOK, "success")
	}
}
