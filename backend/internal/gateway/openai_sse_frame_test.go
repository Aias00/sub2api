package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAICompatSSEFrameParserResetsEventTypeAtFrameBoundary(t *testing.T) {
	var parser OpenAICompatSSEFrameParser

	frame, ok := parser.AddLine("event: response.created")
	require.False(t, ok)
	require.Empty(t, frame)

	frame, ok = parser.AddLine(`data: {"response":{"id":"resp_1"}}`)
	require.False(t, ok)
	require.Empty(t, frame)

	frame, ok = parser.AddLine("")
	require.True(t, ok)
	require.Equal(t, "response.created", frame.EventType)
	require.JSONEq(t, `{"response":{"id":"resp_1"}}`, frame.Data)

	frame, ok = parser.AddLine(`data: {"delta":"ok"}`)
	require.False(t, ok)
	require.Empty(t, frame.EventType)

	frame, ok = parser.AddLine("")
	require.True(t, ok)
	require.Empty(t, frame.EventType)
	require.JSONEq(t, `{"delta":"ok"}`, frame.Data)
}

func TestOpenAICompatSSEFrameParserCombinesDataLinesAndIgnoresComments(t *testing.T) {
	var parser OpenAICompatSSEFrameParser

	_, ok := parser.AddLine(": keep-alive")
	require.False(t, ok)
	_, ok = parser.AddLine("event: response.output_text.delta")
	require.False(t, ok)
	_, ok = parser.AddLine("data: {\"part\":1}")
	require.False(t, ok)
	_, ok = parser.AddLine("data: {\"part\":2}")
	require.False(t, ok)

	frame, ok := parser.Finish()
	require.True(t, ok)
	require.Equal(t, "response.output_text.delta", frame.EventType)
	require.Equal(t, "{\"part\":1}\n{\"part\":2}", frame.Data)
}
