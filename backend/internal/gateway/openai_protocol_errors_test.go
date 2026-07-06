package gateway

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAIInstructionsRequiredError(t *testing.T) {
	require.True(t, IsOpenAIInstructionsRequiredError(
		http.StatusBadRequest,
		"Missing required parameter: 'instructions'",
		nil,
	))
	require.True(t, IsOpenAIInstructionsRequiredError(
		http.StatusBadRequest,
		"",
		[]byte(`{"error":{"message":"Missing required parameter: 'instructions'","type":"invalid_request_error"}}`),
	))
	require.True(t, IsOpenAIInstructionsRequiredError(
		http.StatusBadRequest,
		"",
		[]byte(`{"error":{"message":"missing required parameter","param":"instructions"}}`),
	))
	require.False(t, IsOpenAIInstructionsRequiredError(
		http.StatusForbidden,
		"Missing required parameter: 'instructions'",
		nil,
	))
	require.False(t, IsOpenAIInstructionsRequiredError(
		http.StatusBadRequest,
		"forbidden",
		[]byte(`{"error":{"message":"forbidden"}}`),
	))
}

func TestIsOpenAIInputTokensUnsupported(t *testing.T) {
	require.True(t, IsOpenAIInputTokensUnsupported(
		http.StatusNotFound,
		[]byte(`{"error":{"message":"input_tokens endpoint not found"}}`),
	))
	require.True(t, IsOpenAIInputTokensUnsupported(
		http.StatusNotFound,
		[]byte(`{"detail":"responses/input_tokens not found"}`),
	))
	require.False(t, IsOpenAIInputTokensUnsupported(
		http.StatusBadRequest,
		[]byte(`{"error":{"message":"input_tokens endpoint not found"}}`),
	))
}

func TestIsOpenAIOAuthInputTokensUnsupported(t *testing.T) {
	require.True(t, IsOpenAIOAuthInputTokensUnsupported(
		http.StatusForbidden,
		[]byte(`{"error":{"code":"missing_scope","message":"missing scope api.responses.write"}}`),
	))
	require.True(t, IsOpenAIOAuthInputTokensUnsupported(
		http.StatusForbidden,
		[]byte(`{"error":{"message":"{\"error\":{\"code\":\"missing_scope\",\"message\":\"missing scopes\"}}"}}`),
	))
	require.True(t, IsOpenAIOAuthInputTokensUnsupported(
		http.StatusNotFound,
		[]byte(`{"error":{"message":"input_tokens is not supported"}}`),
	))
	require.False(t, IsOpenAIOAuthInputTokensUnsupported(
		http.StatusBadRequest,
		[]byte(`{"error":{"message":"input_tokens is not supported"}}`),
	))
}
