package gateway

import (
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

const defaultOpenAIUpstreamErrorDetailMaxBytes = 2048

type OpenAIUpstreamFailureInput struct {
	StatusCode               int
	Message                  string
	Body                     []byte
	LogUpstreamErrorBody     bool
	LogUpstreamErrorMaxBytes int
	AccountPoolMode          bool
	AccountRetryableStatus   bool
}

type OpenAIUpstreamFailureDecision struct {
	Failover               bool
	Detail                 string
	RetryableOnSameAccount bool
}

func EvaluateOpenAIUpstreamFailure(input OpenAIUpstreamFailureInput) OpenAIUpstreamFailureDecision {
	if !ShouldFailoverOpenAIUpstreamResponse(input.StatusCode, input.Message, input.Body) {
		return OpenAIUpstreamFailureDecision{}
	}

	return OpenAIUpstreamFailureDecision{
		Failover:               true,
		Detail:                 openAIUpstreamErrorDetail(input.Body, input.LogUpstreamErrorBody, input.LogUpstreamErrorMaxBytes),
		RetryableOnSameAccount: input.AccountPoolMode && (input.AccountRetryableStatus || IsOpenAITransientProcessingError(input.StatusCode, input.Message, input.Body)),
	}
}

func ShouldFailoverOpenAIUpstreamStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusPaymentRequired, http.StatusForbidden, http.StatusTooManyRequests, 529:
		return true
	default:
		return statusCode >= http.StatusInternalServerError
	}
}

func ShouldFailoverOpenAIUpstreamResponse(statusCode int, upstreamMsg string, upstreamBody []byte) bool {
	if IsOpenAIContextWindowError(upstreamMsg, upstreamBody) {
		return false
	}
	if ShouldFailoverOpenAIUpstreamStatus(statusCode) {
		return true
	}
	return IsOpenAITransientProcessingError(statusCode, upstreamMsg, upstreamBody)
}

func IsOpenAITransientProcessingError(upstreamStatusCode int, upstreamMsg string, upstreamBody []byte) bool {
	if upstreamStatusCode != http.StatusBadRequest && upstreamStatusCode != http.StatusServiceUnavailable {
		return false
	}

	if len(upstreamBody) > 0 && hasOpenAIServerOverloadedCode(upstreamBody) {
		return true
	}
	if upstreamStatusCode != http.StatusBadRequest {
		return false
	}

	if matchesOpenAITransientProcessingText(upstreamMsg) {
		return true
	}
	if len(upstreamBody) == 0 {
		return false
	}
	if matchesOpenAITransientProcessingText(gjson.GetBytes(upstreamBody, "error.message").String()) {
		return true
	}
	return matchesOpenAITransientProcessingText(string(upstreamBody))
}

func IsOpenAIContextWindowError(upstreamMsg string, upstreamBody []byte) bool {
	if matchesOpenAIContextWindowText(upstreamMsg) {
		return true
	}
	if len(upstreamBody) == 0 {
		return false
	}
	for _, path := range []string{
		"error.message",
		"response.error.message",
		"message",
		"error.code",
		"response.error.code",
		"code",
	} {
		if matchesOpenAIContextWindowText(gjson.GetBytes(upstreamBody, path).String()) {
			return true
		}
	}
	return matchesOpenAIContextWindowText(string(upstreamBody))
}

func ShouldFailoverOpenAIStreamFailedEvent(payload []byte, message string) bool {
	if IsOpenAIContextWindowError(message, payload) {
		return false
	}
	if IsOpenAITransientProcessingError(http.StatusBadRequest, message, payload) {
		return true
	}
	code := strings.ToLower(strings.TrimSpace(gjson.GetBytes(payload, "response.error.code").String()))
	if code == "" {
		code = strings.ToLower(strings.TrimSpace(gjson.GetBytes(payload, "error.code").String()))
	}
	errType := strings.ToLower(strings.TrimSpace(gjson.GetBytes(payload, "response.error.type").String()))
	if errType == "" {
		errType = strings.ToLower(strings.TrimSpace(gjson.GetBytes(payload, "error.type").String()))
	}
	combined := strings.ToLower(strings.TrimSpace(message + " " + code + " " + errType))
	if combined == "" {
		return true
	}
	for _, marker := range []string{
		"invalid_request",
		"content_policy",
		"policy",
		"safety",
		"high-risk cyber",
		"not allowed",
		"violat",
	} {
		if strings.Contains(combined, marker) {
			return false
		}
	}
	return true
}

func hasOpenAIServerOverloadedCode(payload []byte) bool {
	code := strings.ToLower(strings.TrimSpace(gjson.GetBytes(payload, "error.code").String()))
	if code == "" {
		code = strings.ToLower(strings.TrimSpace(gjson.GetBytes(payload, "response.error.code").String()))
	}
	return code == "server_is_overloaded" || code == "slow_down"
}

func matchesOpenAITransientProcessingText(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "an error occurred while processing your request") {
		return true
	}
	if strings.Contains(lower, "selected model is at capacity") {
		return true
	}
	return strings.Contains(lower, "you can retry your request") &&
		strings.Contains(lower, "help.openai.com") &&
		strings.Contains(lower, "request id")
}

func matchesOpenAIContextWindowText(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "context_too_large") || strings.Contains(lower, "context_length_exceeded") {
		return true
	}
	if strings.Contains(lower, "maximum context length") || strings.Contains(lower, "max context length") {
		return true
	}
	hasExceeded := strings.Contains(lower, "exceed") || strings.Contains(lower, "too large") || strings.Contains(lower, "too long")
	if strings.Contains(lower, "context window") && hasExceeded {
		return true
	}
	if strings.Contains(lower, "context length") && hasExceeded {
		return true
	}
	return strings.Contains(lower, "token limit") &&
		strings.Contains(lower, "context") &&
		hasExceeded
}

func openAIUpstreamErrorDetail(body []byte, enabled bool, maxBytes int) string {
	if !enabled || len(body) == 0 {
		return ""
	}
	if maxBytes <= 0 {
		maxBytes = defaultOpenAIUpstreamErrorDetailMaxBytes
	}
	text := string(body)
	if len(text) <= maxBytes {
		return text
	}
	return text[:maxBytes]
}
