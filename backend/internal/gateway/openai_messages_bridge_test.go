package gateway

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestIsOpenAICompatMessagesBridgePromptCacheKey(t *testing.T) {
	require.True(t, IsOpenAICompatMessagesBridgePromptCacheKey("anthropic-metadata-session-1"))
	require.True(t, IsOpenAICompatMessagesBridgePromptCacheKey(" anthropic-cache-session-1 "))
	require.True(t, IsOpenAICompatMessagesBridgePromptCacheKey("anthropic-digest-abc"))
	require.False(t, IsOpenAICompatMessagesBridgePromptCacheKey("cache-session-1"))
	require.False(t, IsOpenAICompatMessagesBridgePromptCacheKey(""))
}

func TestOpenAICompatMessagesBridgeContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	require.False(t, IsOpenAICompatMessagesBridgeContext(c))
	SetOpenAICompatMessagesBridgeContext(c, false)
	require.False(t, IsOpenAICompatMessagesBridgeContext(c))
	SetOpenAICompatMessagesBridgeContext(c, true)
	require.True(t, IsOpenAICompatMessagesBridgeContext(c))
	require.False(t, IsOpenAICompatMessagesBridgeContext(nil))
}
