package service

import (
	"bytes"

	"github.com/Aias00/cloudbase/internal/gateway"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func isOpenAICompatMessagesBridgeBody(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	if bytes.Contains(body, []byte(openAICompatClaudeCodeTodoGuardMarker)) {
		return true
	}
	return isOpenAICompatMessagesBridgePromptCacheKey(gjson.GetBytes(body, "prompt_cache_key").String())
}

func isOpenAICompatMessagesBridgeRequestBody(reqBody map[string]any) bool {
	if reqBody == nil {
		return false
	}
	if input, ok := reqBody["input"].([]any); ok && inputContainsText(input, openAICompatClaudeCodeTodoGuardMarker) {
		return true
	}
	return isOpenAICompatMessagesBridgePromptCacheKey(firstNonEmptyString(reqBody["prompt_cache_key"]))
}

func isOpenAICompatMessagesBridgePromptCacheKey(key string) bool {
	return gateway.IsOpenAICompatMessagesBridgePromptCacheKey(key)
}

func setOpenAICompatMessagesBridgeContext(c *gin.Context, enabled bool) {
	gateway.SetOpenAICompatMessagesBridgeContext(c, enabled)
}

func isOpenAICompatMessagesBridgeContext(c *gin.Context) bool {
	return gateway.IsOpenAICompatMessagesBridgeContext(c)
}
