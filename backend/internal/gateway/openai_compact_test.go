package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIResponsesRequestPathSuffix(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "exact v1 responses", path: "/v1/responses", want: ""},
		{name: "compact v1 responses", path: "/v1/responses/compact", want: "/compact"},
		{name: "compact alias responses", path: "/responses/compact/", want: "/compact"},
		{name: "nested suffix", path: "/openai/v1/responses/compact/detail", want: "/compact/detail"},
		{name: "unrelated path", path: "/v1/chat/completions", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, OpenAIResponsesRequestPathSuffix(tt.path))
		})
	}
}

func TestIsOpenAIResponsesCompactSuffix(t *testing.T) {
	require.True(t, IsOpenAIResponsesCompactSuffix("/compact"))
	require.True(t, IsOpenAIResponsesCompactSuffix(" /compact/detail "))
	require.False(t, IsOpenAIResponsesCompactSuffix(""))
	require.False(t, IsOpenAIResponsesCompactSuffix("/other"))
}

func TestAppendOpenAIResponsesRequestPathSuffix(t *testing.T) {
	require.Equal(t, "https://api.openai.com/v1/responses/compact", AppendOpenAIResponsesRequestPathSuffix("https://api.openai.com/v1/responses/", "/compact"))
	require.Equal(t, "https://api.openai.com/v1/responses", AppendOpenAIResponsesRequestPathSuffix(" https://api.openai.com/v1/responses/ ", ""))
	require.Equal(t, "", AppendOpenAIResponsesRequestPathSuffix("", "/compact"))
}

func TestNormalizeOpenAICompactRequestBodyPreservesCurrentCodexPayloadFields(t *testing.T) {
	body := []byte(`{"model":"gpt-5.5","input":[{"type":"message","role":"user","content":"compact me"}],"instructions":"compact-test","tools":[{"type":"function","name":"shell"}],"parallel_tool_calls":true,"reasoning":{"effort":"high"},"text":{"verbosity":"low"},"previous_response_id":"resp_123","store":true,"stream":true,"prompt_cache_key":"cache_123"}`)

	normalized, changed, err := NormalizeOpenAICompactRequestBody(body)

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "gpt-5.5", gjson.GetBytes(normalized, "model").String())
	require.True(t, gjson.GetBytes(normalized, "tools").Exists())
	require.True(t, gjson.GetBytes(normalized, "parallel_tool_calls").Bool())
	require.Equal(t, "high", gjson.GetBytes(normalized, "reasoning.effort").String())
	require.Equal(t, "low", gjson.GetBytes(normalized, "text.verbosity").String())
	require.Equal(t, "resp_123", gjson.GetBytes(normalized, "previous_response_id").String())
	require.False(t, gjson.GetBytes(normalized, "store").Exists())
	require.False(t, gjson.GetBytes(normalized, "stream").Exists())
	require.False(t, gjson.GetBytes(normalized, "prompt_cache_key").Exists())
}

func TestNormalizeOpenAICompactRequestBodyUnchanged(t *testing.T) {
	body := []byte(`{"model":"gpt-5","input":"hi"}`)
	normalized, changed, err := NormalizeOpenAICompactRequestBody(body)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, string(body), string(normalized))
}
