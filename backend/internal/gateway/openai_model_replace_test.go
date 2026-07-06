package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceOpenAIModelInSSELine(t *testing.T) {
	require.Equal(t,
		`data: {"model":"alias"}`,
		ReplaceOpenAIModelInSSELine(`data: {"model":"gpt-4o"}`, "gpt-4o", "alias"),
	)
	require.Equal(t,
		`data: {"response":{"model":"alias"}}`,
		ReplaceOpenAIModelInSSELine(`data: {"response":{"model":"gpt-4o"}}`, "gpt-4o", "alias"),
	)
	require.Equal(t, `data: [DONE]`, ReplaceOpenAIModelInSSELine(`data: [DONE]`, "gpt-4o", "alias"))
	require.Equal(t, `event: message`, ReplaceOpenAIModelInSSELine(`event: message`, "gpt-4o", "alias"))
}

func TestReplaceOpenAIModelInSSEBody(t *testing.T) {
	body := "event: message\ndata: {\"model\":\"gpt-4o\"}\n\ndata: [DONE]\n"

	got := ReplaceOpenAIModelInSSEBody(body, "gpt-4o", "alias")
	require.Equal(t, "event: message\ndata: {\"model\":\"alias\"}\n\ndata: [DONE]\n", got)
}

func TestReplaceOpenAIModelInResponseBody(t *testing.T) {
	require.JSONEq(t,
		`{"id":"chatcmpl-123","model":"alias","choices":[]}`,
		string(ReplaceOpenAIModelInResponseBody([]byte(`{"id":"chatcmpl-123","model":"gpt-4o","choices":[]}`), "gpt-4o", "alias")),
	)
	require.Equal(t,
		`{"id":"chatcmpl-123","model":"gpt-3.5-turbo","choices":[]}`,
		string(ReplaceOpenAIModelInResponseBody([]byte(`{"id":"chatcmpl-123","model":"gpt-3.5-turbo","choices":[]}`), "gpt-4o", "alias")),
	)
	require.Equal(t, `not json`, string(ReplaceOpenAIModelInResponseBody([]byte(`not json`), "gpt-4o", "alias")))
}
