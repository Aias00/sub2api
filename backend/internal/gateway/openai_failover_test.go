package gateway

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAITransientProcessingError(t *testing.T) {
	require.True(t, IsOpenAITransientProcessingError(
		http.StatusBadRequest,
		"An error occurred while processing your request.",
		nil,
	))

	require.True(t, IsOpenAITransientProcessingError(
		http.StatusBadRequest,
		"Selected model is at capacity. Please try a different model.",
		[]byte(`{"error":{"message":"Selected model is at capacity. Please try a different model.","type":"invalid_request_error"}}`),
	))

	require.True(t, IsOpenAITransientProcessingError(
		http.StatusBadRequest,
		"",
		[]byte(`{"error":{"code":"server_is_overloaded","message":"Please retry later.","type":"invalid_request_error"}}`),
	))

	require.True(t, IsOpenAITransientProcessingError(
		http.StatusServiceUnavailable,
		"",
		[]byte(`{"error":{"code":"slow_down","message":"Please retry later."}}`),
	))

	require.True(t, IsOpenAITransientProcessingError(
		http.StatusBadRequest,
		"",
		[]byte(`{"error":{"message":"An error occurred while processing your request. You can retry your request, or contact us through our help center at help.openai.com if the error persists. Please include the request ID req_123 in your message."}}`),
	))

	require.False(t, IsOpenAITransientProcessingError(
		http.StatusBadRequest,
		"Missing required parameter: 'instructions'",
		[]byte(`{"error":{"message":"Missing required parameter: 'instructions'"}}`),
	))
}

func TestIsOpenAIContextWindowError(t *testing.T) {
	require.True(t, IsOpenAIContextWindowError(
		"",
		[]byte(`{"error":{"message":"Your input exceeds the context window of this model. Please adjust your input and try again.","type":"upstream_error","code":null}}`),
	))
	require.True(t, IsOpenAIContextWindowError(
		"maximum context length exceeded",
		nil,
	))
	require.False(t, IsOpenAIContextWindowError(
		"context canceled",
		nil,
	))
}

func TestShouldFailoverOpenAIUpstreamResponse(t *testing.T) {
	contextWindowBody := []byte(`{"error":{"message":"Your input exceeds the context window of this model. Please adjust your input and try again.","type":"upstream_error","code":null}}`)
	require.False(t, ShouldFailoverOpenAIUpstreamResponse(http.StatusBadGateway, "", contextWindowBody))
	require.True(t, ShouldFailoverOpenAIUpstreamResponse(http.StatusBadGateway, "temporary upstream outage", []byte(`{"error":{"message":"temporary upstream outage"}}`)))
	require.True(t, ShouldFailoverOpenAIUpstreamResponse(http.StatusTooManyRequests, "", nil))
	require.True(t, ShouldFailoverOpenAIUpstreamResponse(529, "", nil))
	require.False(t, ShouldFailoverOpenAIUpstreamResponse(http.StatusBadRequest, "Missing required parameter: 'instructions'", nil))
}

func TestEvaluateOpenAIUpstreamFailure(t *testing.T) {
	body := []byte(`{"error":{"code":"server_is_overloaded","message":"retry later"}}`)
	got := EvaluateOpenAIUpstreamFailure(OpenAIUpstreamFailureInput{
		StatusCode:               http.StatusBadRequest,
		Body:                     body,
		LogUpstreamErrorBody:     true,
		LogUpstreamErrorMaxBytes: 12,
		AccountPoolMode:          true,
		AccountRetryableStatus:   false,
	})
	require.True(t, got.Failover)
	require.Equal(t, string(body[:12]), got.Detail)
	require.True(t, got.RetryableOnSameAccount)

	got = EvaluateOpenAIUpstreamFailure(OpenAIUpstreamFailureInput{
		StatusCode:      http.StatusBadRequest,
		Message:         "maximum context length exceeded",
		Body:            []byte(`{"error":{"message":"maximum context length exceeded"}}`),
		AccountPoolMode: true,
	})
	require.False(t, got.Failover)
	require.Empty(t, got.Detail)
	require.False(t, got.RetryableOnSameAccount)
}

func TestShouldFailoverOpenAIStreamFailedEvent(t *testing.T) {
	require.True(t, ShouldFailoverOpenAIStreamFailedEvent(nil, ""))
	require.True(t, ShouldFailoverOpenAIStreamFailedEvent(
		[]byte(`{"type":"response.failed","response":{"error":{"code":"server_is_overloaded","message":"retry"}}}`),
		"",
	))
	require.False(t, ShouldFailoverOpenAIStreamFailedEvent(
		[]byte(`{"type":"response.failed","response":{"error":{"code":"invalid_request","type":"invalid_request_error","message":"bad request"}}}`),
		"bad request",
	))
	require.False(t, ShouldFailoverOpenAIStreamFailedEvent(
		[]byte(`{"type":"response.failed","response":{"error":{"message":"context length exceeded"}}}`),
		"",
	))
}
