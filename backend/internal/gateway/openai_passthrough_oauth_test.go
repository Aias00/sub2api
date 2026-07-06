package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeOpenAIPassthroughOAuthBodyRemovesUnsupportedFields(t *testing.T) {
	body := []byte(`{"model":"gpt-5.4","input":"hello","user":"user_123","metadata":{"user_id":"user_123"},"stream_options":{"include_usage":true}}`)

	normalized, changed, err := NormalizeOpenAIPassthroughOAuthBody(body, false, []string{"user", "metadata", "stream_options"})
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, gjson.GetBytes(normalized, "user").Exists())
	require.False(t, gjson.GetBytes(normalized, "metadata").Exists())
	require.False(t, gjson.GetBytes(normalized, "stream_options").Exists())
	require.True(t, gjson.GetBytes(normalized, "stream").Bool())
	require.False(t, gjson.GetBytes(normalized, "store").Bool())
}

func TestNormalizeOpenAIPassthroughOAuthBodyCompactRemovesStoreAndStream(t *testing.T) {
	body := []byte(`{"model":"gpt-5.4","input":"hello","stream":true,"store":true}`)

	normalized, changed, err := NormalizeOpenAIPassthroughOAuthBody(body, true, nil)
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, gjson.GetBytes(normalized, "stream").Exists())
	require.False(t, gjson.GetBytes(normalized, "store").Exists())
}

func TestNormalizeOpenAIPassthroughOAuthBodyNoopWhenAlreadyNormalized(t *testing.T) {
	body := []byte(`{"model":"gpt-5.4","input":"hello","stream":true,"store":false}`)

	normalized, changed, err := NormalizeOpenAIPassthroughOAuthBody(body, false, []string{"user"})
	require.NoError(t, err)
	require.False(t, changed)
	require.JSONEq(t, string(body), string(normalized))
}

func TestNormalizeOpenAIPassthroughOAuthBodyEmptyNoop(t *testing.T) {
	normalized, changed, err := NormalizeOpenAIPassthroughOAuthBody(nil, false, []string{"user"})
	require.NoError(t, err)
	require.False(t, changed)
	require.Nil(t, normalized)
}
