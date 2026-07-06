package gateway

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAIPassthroughAllowedRequestHeader(t *testing.T) {
	require.True(t, IsOpenAIPassthroughAllowedRequestHeader("content-type", false))
	require.True(t, IsOpenAIPassthroughAllowedRequestHeader("openai-beta", false))
	require.False(t, IsOpenAIPassthroughAllowedRequestHeader("authorization", false))
	require.False(t, IsOpenAIPassthroughAllowedRequestHeader("", false))

	require.False(t, IsOpenAIPassthroughAllowedRequestHeader("x-stainless-timeout", false))
	require.True(t, IsOpenAIPassthroughAllowedRequestHeader("x-stainless-timeout", true))
}

func TestCollectOpenAIPassthroughTimeoutHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Add("X-Stainless-Timeout", "10")
	headers.Add("Grpc-Timeout", "5S")
	headers.Add("Grpc-Timeout", "6S")
	headers.Add("Content-Type", "application/json")

	require.Equal(t, []string{
		"grpc-timeout=5S|6S",
		"x-stainless-timeout=10",
	}, CollectOpenAIPassthroughTimeoutHeaders(headers))
	require.Nil(t, CollectOpenAIPassthroughTimeoutHeaders(nil))
}
