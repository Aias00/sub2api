package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractOpenAIReasoningEffortFromBody(t *testing.T) {
	tests := []struct {
		name      string
		body      []byte
		model     string
		wantNil   bool
		wantValue string
	}{
		{name: "nested effort wins", body: []byte(`{"reasoning":{"effort":"medium"}}`), model: "gpt-5-high", wantValue: "medium"},
		{name: "flat effort", body: []byte(`{"reasoning_effort":"x-high"}`), wantValue: "xhigh"},
		{name: "max aliases xhigh", body: []byte(`{"reasoning_effort":"max"}`), model: "deepseek-v4-pro", wantValue: "xhigh"},
		{name: "minimal empty", body: []byte(`{"reasoning":{"effort":"minimal"}}`), model: "gpt-5-high", wantNil: true},
		{name: "model suffix fallback", body: []byte(`{"input":"hi"}`), model: "gpt-5-high", wantValue: "high"},
		{name: "unknown suffix", body: []byte(`{"input":"hi"}`), model: "gpt-5-unknown", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractOpenAIReasoningEffortFromBody(tt.body, tt.model)
			if tt.wantNil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, tt.wantValue, *got)
		})
	}
}

func TestExtractOpenAIReasoningEffort(t *testing.T) {
	got := ExtractOpenAIReasoningEffort(map[string]any{
		"reasoning": map[string]any{"effort": "extra_high"},
	}, "gpt-5-low")
	require.NotNil(t, got)
	require.Equal(t, "xhigh", *got)

	got = ExtractOpenAIReasoningEffort(map[string]any{"reasoning_effort": "low"}, "")
	require.NotNil(t, got)
	require.Equal(t, "low", *got)

	got = ExtractOpenAIReasoningEffort(nil, "provider/gpt-5-medium")
	require.NotNil(t, got)
	require.Equal(t, "medium", *got)
}

func TestNormalizeOpenAIReasoningEffort(t *testing.T) {
	require.Equal(t, "", NormalizeOpenAIReasoningEffort("minimal"))
	require.Equal(t, "low", NormalizeOpenAIReasoningEffort(" low "))
	require.Equal(t, "xhigh", NormalizeOpenAIReasoningEffort("x-high"))
	require.Equal(t, "xhigh", NormalizeOpenAIReasoningEffort("extra high"))
	require.Equal(t, "", NormalizeOpenAIReasoningEffort("unknown"))
}

func TestExtractOpenAIServiceTier(t *testing.T) {
	require.Equal(t, "priority", *ExtractOpenAIServiceTier(map[string]any{"service_tier": "fast"}))
	require.Equal(t, "flex", *ExtractOpenAIServiceTier(map[string]any{"service_tier": "flex"}))
	require.Equal(t, "auto", *ExtractOpenAIServiceTier(map[string]any{"service_tier": "auto"}))
	require.Equal(t, "default", *ExtractOpenAIServiceTier(map[string]any{"service_tier": "default"}))
	require.Equal(t, "scale", *ExtractOpenAIServiceTier(map[string]any{"service_tier": "scale"}))
	require.Nil(t, ExtractOpenAIServiceTier(map[string]any{"service_tier": 1}))
	require.Nil(t, ExtractOpenAIServiceTier(nil))
}

func TestExtractOpenAIServiceTierFromBody(t *testing.T) {
	require.Equal(t, "priority", *ExtractOpenAIServiceTierFromBody([]byte(`{"service_tier":"fast"}`)))
	require.Nil(t, ExtractOpenAIServiceTierFromBody(nil))
	require.Nil(t, ExtractOpenAIServiceTierFromBody([]byte(`{"service_tier":"unknown"}`)))
}

func TestDetectOpenAIPassthroughInstructionsRejectReason(t *testing.T) {
	tests := []struct {
		name  string
		model string
		body  []byte
		want  string
	}{
		{
			name:  "non codex model ignored",
			model: "gpt-5.4",
			body:  []byte(`{}`),
		},
		{
			name:  "codex missing instructions",
			model: "gpt-5.4-codex",
			body:  []byte(`{"input":"hi"}`),
			want:  "instructions_missing",
		},
		{
			name:  "codex non string instructions",
			model: "gpt-5.4-codex",
			body:  []byte(`{"instructions":123}`),
			want:  "instructions_not_string",
		},
		{
			name:  "codex empty instructions",
			model: "gpt-5.4-codex",
			body:  []byte(`{"instructions":"  "}`),
			want:  "instructions_empty",
		},
		{
			name:  "codex valid instructions",
			model: " GPT-5.4-CODEX ",
			body:  []byte(`{"instructions":"be concise"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, DetectOpenAIPassthroughInstructionsRejectReason(tt.model, tt.body))
		})
	}
}
