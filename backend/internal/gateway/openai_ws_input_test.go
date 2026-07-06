package gateway

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIWSInputIsPrefixExtended(t *testing.T) {
	tests := []struct {
		name      string
		previous  []byte
		current   []byte
		want      bool
		expectErr bool
	}{
		{
			name:     "both_missing",
			previous: []byte(`{"model":"gpt"}`),
			current:  []byte(`{"model":"gpt"}`),
			want:     true,
		},
		{
			name:     "current_extends_previous",
			previous: []byte(`{"input":[{"b":2,"a":1}]}`),
			current:  []byte(`{"input":[{"a":1,"b":2},{"type":"input_text","text":"next"}]}`),
			want:     true,
		},
		{
			name:     "current_shorter",
			previous: []byte(`{"input":[{"type":"input_text","text":"first"},{"type":"input_text","text":"second"}]}`),
			current:  []byte(`{"input":[{"type":"input_text","text":"first"}]}`),
			want:     false,
		},
		{
			name:     "current_changes_prefix",
			previous: []byte(`{"input":[{"type":"input_text","text":"first"}]}`),
			current:  []byte(`{"input":[{"type":"input_text","text":"changed"}]}`),
			want:     false,
		},
		{
			name:      "invalid_input_json",
			previous:  []byte(`{"input":[}`),
			current:   []byte(`{"input":[]}`),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OpenAIWSInputIsPrefixExtended(tt.previous, tt.current)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeOpenAIWSJSONForCompare(t *testing.T) {
	normalized, err := NormalizeOpenAIWSJSONForCompare([]byte(`{"b":2,"a":1}`))
	require.NoError(t, err)
	require.Equal(t, `{"a":1,"b":2}`, string(normalized))

	_, err = NormalizeOpenAIWSJSONForCompare([]byte("   "))
	require.Error(t, err)

	_, err = NormalizeOpenAIWSJSONForCompare([]byte(`{"a":`))
	require.Error(t, err)
}

func TestNormalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(t *testing.T) {
	normalized, err := NormalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(
		[]byte(`{"model":"gpt-5.1","input":[1],"previous_response_id":"resp_x","metadata":{"b":2,"a":1}}`),
	)
	require.NoError(t, err)
	require.False(t, gjson.GetBytes(normalized, "input").Exists())
	require.False(t, gjson.GetBytes(normalized, "previous_response_id").Exists())
	require.Equal(t, float64(1), gjson.GetBytes(normalized, "metadata.a").Float())

	_, err = NormalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID(nil)
	require.Error(t, err)

	_, err = NormalizeOpenAIWSPayloadWithoutInputAndPreviousResponseID([]byte(`[]`))
	require.Error(t, err)
}

func TestExtractOpenAIWSNormalizedInputSequence(t *testing.T) {
	tests := []struct {
		name       string
		payload    []byte
		wantExists bool
		wantItems  []string
		expectErr  bool
	}{
		{name: "empty_payload", payload: nil},
		{name: "input_missing", payload: []byte(`{"type":"response.create"}`)},
		{name: "input_array", payload: []byte(`{"input":[{"type":"input_text","text":"hello"}]}`), wantExists: true, wantItems: []string{`{"type":"input_text","text":"hello"}`}},
		{name: "input_object", payload: []byte(`{"input":{"type":"input_text","text":"hello"}}`), wantExists: true, wantItems: []string{`{"type":"input_text","text":"hello"}`}},
		{name: "input_string", payload: []byte(`{"input":"hello"}`), wantExists: true, wantItems: []string{`"hello"`}},
		{name: "input_number", payload: []byte(`{"input":42}`), wantExists: true, wantItems: []string{`42`}},
		{name: "input_bool", payload: []byte(`{"input":true}`), wantExists: true, wantItems: []string{`true`}},
		{name: "input_null", payload: []byte(`{"input":null}`), wantExists: true, wantItems: []string{`null`}},
		{name: "input_invalid_array_json", payload: []byte(`{"input":[}`), wantExists: true, expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, exists, err := ExtractOpenAIWSNormalizedInputSequence(tt.payload)
			if tt.expectErr {
				require.Error(t, err)
				require.Equal(t, tt.wantExists, exists)
				require.Nil(t, items)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantExists, exists)
			require.Len(t, items, len(tt.wantItems))
			for idx, want := range tt.wantItems {
				require.Equal(t, want, string(items[idx]))
			}
		})
	}
}

func TestOpenAIWSRawItemsToolCallPredicates(t *testing.T) {
	items := []json.RawMessage{
		json.RawMessage(`{"type":"function_call","call_id":"call_1"}`),
		json.RawMessage(`{"type":"function_call_output","call_id":"call_1"}`),
	}
	require.True(t, OpenAIWSRawItemsHasFunctionCallOutput(items))
	require.True(t, OpenAIWSRawItemsHaveToolCallContextForOutputs(items))

	missingContext := []json.RawMessage{
		json.RawMessage(`{"type":"function_call_output","call_id":"call_1"}`),
	}
	require.True(t, OpenAIWSRawItemsHasFunctionCallOutput(missingContext))
	require.False(t, OpenAIWSRawItemsHaveToolCallContextForOutputs(missingContext))

	missingCallID := []json.RawMessage{
		json.RawMessage(`{"type":"function_call"}`),
		json.RawMessage(`{"type":"function_call_output"}`),
	}
	require.True(t, OpenAIWSRawItemsHasFunctionCallOutput(missingCallID))
	require.False(t, OpenAIWSRawItemsHaveToolCallContextForOutputs(missingCallID))
}

func TestOpenAIWSRawPayloadHasToolCallOutput(t *testing.T) {
	require.True(t, OpenAIWSRawPayloadHasToolCallOutput([]byte(`{"input":[{"type":"function_call_output","call_id":"call_1"}]}`)))
	require.True(t, OpenAIWSRawPayloadHasToolCallOutput([]byte(`{"input":{"type":"function_call_output","call_id":"call_1"}}`)))
	require.False(t, OpenAIWSRawPayloadHasToolCallOutput([]byte(`{"input":[{"type":"input_text","text":"hello"}]}`)))
	require.False(t, OpenAIWSRawPayloadHasToolCallOutput(nil))
}
