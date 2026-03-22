package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/gemini"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiWebHTTPUpstreamStub struct{}

func (s *geminiWebHTTPUpstreamStub) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func (s *geminiWebHTTPUpstreamStub) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ bool) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func TestGeminiMessagesCompatService_Forward_GeminiWeb(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/messages", r.URL.Path)
		require.Equal(t, "Bearer gateway-key", r.Header.Get("Authorization"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-request-id", "req-gw-1")
		_, _ = w.Write([]byte(`{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"text","text":"hello"}],"model":"gemini-web","usage":{"input_tokens":12,"output_tokens":34}}`))
	}))
	defer upstream.Close()

	svc := &GeminiMessagesCompatService{
		httpUpstream: &geminiWebHTTPUpstreamStub{},
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{
					Enabled:           false,
					AllowInsecureHTTP: true,
				},
			},
		},
	}

	account := &Account{
		ID:          1001,
		Name:        "gw",
		Platform:    PlatformGemini,
		Type:        AccountTypeGeminiWeb,
		Concurrency: 1,
		Credentials: map[string]any{
			"base_url": upstream.URL,
			"api_key":  "gateway-key",
		},
	}

	payload := map[string]any{
		"model": "gemini-2.5-pro",
		"messages": []any{
			map[string]any{"role": "user", "content": "hi"},
		},
		"stream": false,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))

	result, err := svc.Forward(context.Background(), c, account, body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "req-gw-1", result.RequestID)
	require.Equal(t, 12, result.Usage.InputTokens)
	require.Equal(t, 34, result.Usage.OutputTokens)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "hello")
}

func TestGeminiMessagesCompatService_ForwardNative_GeminiWeb(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer gateway-key", r.Header.Get("Authorization"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		require.Equal(t, "gemini-2.5-pro", payload["model"])
		require.Equal(t, false, payload["stream"])

		messages, ok := payload["messages"].([]any)
		require.True(t, ok)
		require.Len(t, messages, 2)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-request-id", "req-gw-native-1")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_1","object":"chat.completion","model":"gemini-2.5-pro","choices":[{"index":0,"message":{"role":"assistant","content":"hello from gemini web"},"finish_reason":"stop"}],"usage":{"prompt_tokens":-1,"completion_tokens":-1,"total_tokens":-1}}`))
	}))
	defer upstream.Close()

	svc := &GeminiMessagesCompatService{
		httpUpstream: &geminiWebHTTPUpstreamStub{},
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{
					Enabled:           false,
					AllowInsecureHTTP: true,
				},
			},
		},
	}

	account := &Account{
		ID:          1002,
		Name:        "gw-native",
		Platform:    PlatformGemini,
		Type:        AccountTypeGeminiWeb,
		Concurrency: 1,
		Credentials: map[string]any{
			"base_url": upstream.URL,
			"api_key":  "gateway-key",
		},
	}

	requestBody := []byte(`{
		"systemInstruction":{"parts":[{"text":"You are helpful."}]},
		"contents":[
			{"role":"user","parts":[{"text":"Say hello"}]}
		]
	}`)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-2.5-pro:generateContent", bytes.NewReader(requestBody))

	result, err := svc.ForwardNative(context.Background(), c, account, "gemini-2.5-pro", "generateContent", false, requestBody)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "req-gw-native-1", result.RequestID)
	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "gemini-2.5-pro", response["modelVersion"])

	candidates, ok := response["candidates"].([]any)
	require.True(t, ok)
	require.Len(t, candidates, 1)

	candidate, ok := candidates[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "STOP", candidate["finishReason"])

	content, ok := candidate["content"].(map[string]any)
	require.True(t, ok)
	parts, ok := content["parts"].([]any)
	require.True(t, ok)
	part, ok := parts[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "hello from gemini web", part["text"])
}

func TestGeminiMessagesCompatService_ForwardAIStudioGET_GeminiWebFallback(t *testing.T) {
	svc := &GeminiMessagesCompatService{}

	account := &Account{
		ID:       1003,
		Name:     "gw-models",
		Platform: PlatformGemini,
		Type:     AccountTypeGeminiWeb,
	}

	res, err := svc.ForwardAIStudioGET(context.Background(), account, "/v1beta/models")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode)

	var response gemini.ModelsListResponse
	require.NoError(t, json.Unmarshal(res.Body, &response))
	require.NotEmpty(t, response.Models)
}
