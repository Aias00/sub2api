package service

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Aias00/cloudbase/internal/config"
	"github.com/Aias00/cloudbase/internal/pkg/claude"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAccountTestService_TestAccountConnection_ClaudeCodeNativeModeUsesClaudeCodeHeadersAndPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	account := Account{
		ID:          11,
		Name:        "claude-apikey",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "sk-ant-test",
			"base_url": "https://example.com",
		},
	}
	repo := &snapshotUpdateAccountRepo{
		stubOpenAIAccountRepo: stubOpenAIAccountRepo{accounts: []Account{account}},
	}
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader("data: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"Hi! 👋\"}}\n\ndata: [DONE]\n\n")),
	}}
	svc := &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/11/test", bytes.NewReader(nil))

	err := svc.TestAccountConnection(c, account.ID, "claude-opus-4-7", "", AccountTestModeClaudeCode)
	require.NoError(t, err)

	require.Equal(t, "https://example.com/v1/messages?beta=true", upstream.lastReq.URL.String())
	require.Equal(t, "application/json", upstream.lastReq.Header.Get("Accept"))
	require.Equal(t, "stream", upstream.lastReq.Header.Get("X-Stainless-Helper-Method"))
	require.NotEmpty(t, upstream.lastReq.Header.Get("X-Client-Request-Id"))
	require.Equal(t, claude.DefaultHeaders["User-Agent"], upstream.lastReq.Header.Get("User-Agent"))

	betaHeader := upstream.lastReq.Header.Get("Anthropic-Beta")
	require.Contains(t, betaHeader, claude.BetaClaudeCode)
	require.Contains(t, betaHeader, claude.BetaPromptCachingScope)
	require.Contains(t, betaHeader, claude.BetaEffort)
	require.Contains(t, betaHeader, claude.BetaContextManagement)
	require.Contains(t, betaHeader, claude.BetaExtendedCacheTTL)

	require.Equal(t, "claude-opus-4-7", gjson.GetBytes(upstream.lastBody, "model").String())
	require.True(t, gjson.GetBytes(upstream.lastBody, "stream").Bool())
	require.Equal(t, "hi", gjson.GetBytes(upstream.lastBody, "messages.0.content.0.text").String())
	require.Equal(t, claudeCodeSystemPrompt, gjson.GetBytes(upstream.lastBody, "system.0.text").String())
	require.Equal(t, "ephemeral", gjson.GetBytes(upstream.lastBody, "system.0.cache_control.type").String())
	require.NotEmpty(t, gjson.GetBytes(upstream.lastBody, "metadata.user_id").String())

	bodyStr := string(upstream.lastBody)
	require.Less(t, strings.Index(bodyStr, `"metadata"`), strings.Index(bodyStr, `"system"`))
	require.Less(t, strings.Index(bodyStr, `"system"`), strings.Index(bodyStr, `"messages"`))
	require.Contains(t, rec.Body.String(), "[Upstream Response]")
	require.Contains(t, rec.Body.String(), "content_block_delta")
	require.Contains(t, rec.Body.String(), "Hi! 👋")
	require.NotContains(t, rec.Body.String(), "[streamed response body shown below]")
}
