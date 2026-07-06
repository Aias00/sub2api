package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIRequestViewExtractsRawScalars(t *testing.T) {
	view := NewOpenAIRequestView([]byte(`{"model":" gpt-5 ","stream":true,"prompt_cache_key":" ses-1 ","previous_response_id":" resp-1 ","service_tier":" fast ","reasoning":{"effort":" medium "}}`))

	require.Equal(t, "gpt-5", view.Model)
	require.True(t, view.Stream)
	require.Equal(t, "ses-1", view.PromptCacheKey)
	require.Equal(t, "resp-1", view.PreviousResponseID)
	require.Equal(t, "fast", view.ServiceTier)
	require.Equal(t, "medium", view.ReasoningEffort)
}

func TestOpenAIRequestViewApplyPatches(t *testing.T) {
	view := NewOpenAIRequestView([]byte(`{"model":"gpt-5","previous_response_id":"resp_1","reasoning":{"effort":"minimal"},"input":[{"type":"message","content":"hi"}]}`))
	view.MarkPatchSet("model", "gpt-5.1")
	view.MarkPatchDelete("previous_response_id")
	view.MarkPatchSet("reasoning.effort", "none")

	patched, err := view.ApplyPatches()
	require.NoError(t, err)
	require.JSONEq(t, `{"model":"gpt-5.1","reasoning":{"effort":"none"},"input":[{"type":"message","content":"hi"}]}`, string(patched))
}

func TestOpenAIRequestViewRejectsEscapedPatchPath(t *testing.T) {
	view := NewOpenAIRequestView([]byte(`{"metadata":{"user.id":"old"}}`))
	view.MarkPatchSet(`metadata.user\.id`, "new")

	require.False(t, view.HasPatches())
	_, err := view.ApplyPatches()
	require.Error(t, err)
}

func TestOpenAIRequestViewApplyPatchesDisabled(t *testing.T) {
	view := NewOpenAIRequestView([]byte(`{"model":"gpt-5"}`))
	view.MarkPatchSet("model", "gpt-5.1")
	view.DisablePatches()

	_, err := view.ApplyPatches()
	require.Error(t, err)
}

func TestOpenAIRequestViewHasPatches(t *testing.T) {
	view := NewOpenAIRequestView([]byte(`{"model":"gpt-5"}`))
	require.False(t, view.HasPatches())

	view.MarkPatchSet("model", "gpt-5.1")
	require.True(t, view.HasPatches())

	view.DisablePatches()
	require.False(t, view.HasPatches())
}

func TestOpenAIRequestMapPathHelpers(t *testing.T) {
	body := map[string]any{"model": "gpt-5"}

	SetOpenAIRequestMapPath(body, "reasoning.effort", "none")
	require.Equal(t, map[string]any{"effort": "none"}, body["reasoning"])

	DeleteOpenAIRequestMapPath(body, "reasoning.effort")
	require.Equal(t, map[string]any{}, body["reasoning"])
}
