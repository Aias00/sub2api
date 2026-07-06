package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAIChatUsageOnlyStreamChunk(t *testing.T) {
	require.True(t, IsOpenAIChatUsageOnlyStreamChunk(`{"choices":[],"usage":{"prompt_tokens":1,"completion_tokens":2}}`))
	require.False(t, IsOpenAIChatUsageOnlyStreamChunk(`{"choices":[{"index":0}],"usage":{"prompt_tokens":1,"completion_tokens":2}}`))
	require.False(t, IsOpenAIChatUsageOnlyStreamChunk(`{"choices":[]}`))
	require.False(t, IsOpenAIChatUsageOnlyStreamChunk(``))
}
