package image

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsGenerationIntent(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		model    string
		body     []byte
		want     bool
	}{
		{
			name:     "dedicated images endpoint",
			endpoint: "https://api.openai.com/v1/images/generations?ignored=true",
			body:     []byte(`{"model":"gpt-image-2"}`),
			want:     true,
		},
		{
			name:     "image model",
			endpoint: "/v1/responses",
			model:    "gpt-image-2",
			body:     []byte(`{"model":"gpt-image-2"}`),
			want:     true,
		},
		{
			name:     "image tool",
			endpoint: "/v1/responses",
			model:    "gpt-5.4",
			body:     []byte(`{"model":"gpt-5.4","tools":[{"type":"image_generation"}]}`),
			want:     true,
		},
		{
			name:     "tool choice object",
			endpoint: "/v1/responses",
			model:    "gpt-5.4",
			body:     []byte(`{"model":"gpt-5.4","tool_choice":{"tool":{"type":"image_generation"}}}`),
			want:     true,
		},
		{
			name:     "required tool choice alone is text",
			endpoint: "/v1/responses",
			model:    "gpt-5.4",
			body:     []byte(`{"model":"gpt-5.4","tool_choice":"required"}`),
			want:     false,
		},
		{
			name:     "text only",
			endpoint: "/v1/responses",
			model:    "gpt-5.4",
			body:     []byte(`{"model":"gpt-5.4","input":"write code"}`),
			want:     false,
		},
		{
			name:     "invalid body",
			endpoint: "/v1/responses",
			model:    "gpt-5.4",
			body:     []byte(`{`),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsGenerationIntent(tt.endpoint, tt.model, tt.body))
		})
	}
}

func TestIsGenerationIntentMap(t *testing.T) {
	require.True(t, IsGenerationIntentMap("/v1/responses", "gpt-5.4", map[string]any{
		"tools": []any{map[string]any{"type": "image_generation"}},
	}))
	require.True(t, IsGenerationIntentMap("/v1/responses", "gpt-5.4", map[string]any{
		"tool_choice": map[string]any{"function": map[string]any{"name": "image_generation"}},
	}))
	require.False(t, IsGenerationIntentMap("/v1/responses", "gpt-5.4", map[string]any{
		"tool_choice": "required",
	}))
}

func TestRequestBodyImageGenerationToolNeedsNormalization(t *testing.T) {
	require.True(t, RequestBodyImageGenerationToolNeedsNormalization([]byte(`{"tools":[{"type":"image_generation","format":"png"}]}`)))
	require.True(t, RequestBodyImageGenerationToolNeedsNormalization([]byte(`{"tools":[{"type":"image_generation","compression":80}]}`)))
	require.False(t, RequestBodyImageGenerationToolNeedsNormalization([]byte(`{"tools":[{"type":"image_generation"}]}`)))
	require.False(t, RequestBodyImageGenerationToolNeedsNormalization([]byte(`{"tools":[{"type":"web_search"}]}`)))
}

func TestCodexBridgeOverrides(t *testing.T) {
	require.Nil(t, BoolOverrideFromMap(nil, FeatureKeyCodexImageGenerationBridge))
	require.Equal(t, true, *BoolOverrideFromMap(map[string]any{FeatureKeyCodexImageGenerationBridge: true}, FeatureKeyCodexImageGenerationBridge))
	require.Equal(t, false, *PlatformBoolOverride(map[string]any{
		FeatureKeyCodexImageGenerationBridge: map[string]any{"openai": false},
	}, FeatureKeyCodexImageGenerationBridge, "openai"))
}
