package gateway

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeOpenAIResponsesFunctionCallArguments_TopLevelDone(t *testing.T) {
	args := `{"cmd":"pwd"}`
	body := []byte(fmt.Sprintf(`{"type":"response.function_call_arguments.done","arguments":%q}`, args+args))

	normalized, changed := NormalizeOpenAIResponsesFunctionCallArguments(body)

	require.True(t, changed)
	require.Equal(t, args, gjson.GetBytes(normalized, "arguments").String())
}

func TestNormalizeOpenAIResponsesFunctionCallArguments_ItemAndOutputs(t *testing.T) {
	argsA := `{"a":1}`
	argsB := `[1,2]`
	body := []byte(fmt.Sprintf(`{
		"item":{"type":"function_call","arguments":%q},
		"response":{"output":[{"type":"custom_tool_call","arguments":%q}]},
		"output":[{"type":"message","arguments":"noop"},{"type":"function_call","arguments":%q}]
	}`, argsA+argsA, argsB+argsB, argsA+argsA))

	normalized, changed := NormalizeOpenAIResponsesFunctionCallArguments(body)

	require.True(t, changed)
	require.Equal(t, argsA, gjson.GetBytes(normalized, "item.arguments").String())
	require.Equal(t, argsB, gjson.GetBytes(normalized, "response.output.0.arguments").String())
	require.Equal(t, argsA, gjson.GetBytes(normalized, "output.1.arguments").String())
	require.Equal(t, "noop", gjson.GetBytes(normalized, "output.0.arguments").String())
}

func TestNormalizeOpenAIResponsesFunctionCallArguments_NoChange(t *testing.T) {
	for _, body := range [][]byte{
		[]byte(``),
		[]byte(`not-json`),
		[]byte(`{"type":"response.function_call_arguments.done","arguments":"{\"cmd\":\"pwd\"}"}`),
		[]byte(`{"type":"response.function_call_arguments.done","arguments":"abcabc"}`),
	} {
		normalized, changed := NormalizeOpenAIResponsesFunctionCallArguments(body)
		require.False(t, changed, string(body))
		require.Equal(t, string(body), string(normalized))
	}
}

func TestDedupeRepeatedJSONArgumentString(t *testing.T) {
	got, ok := DedupeRepeatedJSONArgumentString(`{"a":1}` + `{"a":1}`)
	require.True(t, ok)
	require.Equal(t, `{"a":1}`, got)

	_, ok = DedupeRepeatedJSONArgumentString(`abcabc`)
	require.False(t, ok)
}
