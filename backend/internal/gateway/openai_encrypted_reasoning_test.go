package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrimOpenAIEncryptedReasoningItemsFiltersArrayInput(t *testing.T) {
	body := map[string]any{
		"input": []any{
			map[string]any{"type": "message", "content": "hi"},
			map[string]any{"type": "reasoning", "encrypted_content": "sealed"},
			map[string]any{"type": "reasoning", "id": "rs_1", "encrypted_content": "sealed"},
		},
	}

	require.True(t, TrimOpenAIEncryptedReasoningItems(body))
	require.Equal(t, []any{
		map[string]any{"type": "message", "content": "hi"},
		map[string]any{"type": "reasoning", "id": "rs_1"},
	}, body["input"])
}

func TestTrimOpenAIEncryptedReasoningItemsDeletesOnlyEncryptedReasoningInput(t *testing.T) {
	body := map[string]any{
		"input": map[string]any{"type": "reasoning", "encrypted_content": "sealed"},
	}

	require.True(t, TrimOpenAIEncryptedReasoningItems(body))
	require.NotContains(t, body, "input")
}

func TestTrimOpenAIEncryptedReasoningItemsKeepsSingleReasoningWithOtherFields(t *testing.T) {
	body := map[string]any{
		"input": map[string]any{"type": " reasoning ", "id": "rs_1", "encrypted_content": "sealed"},
	}

	require.True(t, TrimOpenAIEncryptedReasoningItems(body))
	require.Equal(t, map[string]any{"type": " reasoning ", "id": "rs_1"}, body["input"])
}

func TestTrimOpenAIEncryptedReasoningItemsNoopWithoutEncryptedReasoning(t *testing.T) {
	body := map[string]any{
		"input": []map[string]any{
			{"type": "message", "content": "hi"},
			{"type": "reasoning", "summary": "ok"},
		},
	}

	require.False(t, TrimOpenAIEncryptedReasoningItems(body))
	require.Equal(t, []map[string]any{
		{"type": "message", "content": "hi"},
		{"type": "reasoning", "summary": "ok"},
	}, body["input"])
}

func TestSanitizeOpenAIEncryptedReasoningInputItemIgnoresNonMap(t *testing.T) {
	next, changed, keep := SanitizeOpenAIEncryptedReasoningInputItem("plain")
	require.Equal(t, "plain", next)
	require.False(t, changed)
	require.True(t, keep)
}
