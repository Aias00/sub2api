package image

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutputCounterDeduplicatesFinalImages(t *testing.T) {
	counter := NewOutputCounter()
	counter.AddSSEData([]byte(`{"type":"response.image_generation_call.partial_image","partial_image_b64":"abc"}`))
	counter.AddSSEData([]byte(`{"type":"response.output_item.done","item":{"id":"ig_1","type":"image_generation_call","result":"final-a","size":"1024x1024"}}`))
	counter.AddSSEData([]byte(`{"type":"response.completed","response":{"output":[{"id":"ig_1","type":"image_generation_call","result":"final-a"},{"id":"ig_2","type":"image_generation_call","result":"final-b","size":"3840x2160"}]}}`))

	require.Equal(t, 2, counter.Count())
	require.Equal(t, []string{"1024x1024", "3840x2160"}, counter.Sizes())
}

func TestOutputCounterCountsImagesAPIStreamShapes(t *testing.T) {
	counter := NewOutputCounter()
	counter.AddSSEData([]byte(`{"type":"image_generation.completed","id":"ig_complete","b64_json":"final-a"}`))
	counter.AddSSEData([]byte(`{"type":"response.output_item.done","item":{"id":"ig_item","type":"image_generation_call","result":"final-b"}}`))
	counter.AddSSEData([]byte(`{"type":"response.completed","response":{"output":[{"id":"ig_done","type":"image_generation_call","result":"final-c"}]}}`))
	require.Equal(t, 3, counter.Count())

	dataCounter := NewOutputCounter()
	dataCounter.AddSSEData([]byte(`{"data":[{"b64_json":"a"},{"b64_json":"b"}]}`))
	dataCounter.AddSSEData([]byte(`{"data":[{"b64_json":"a"},{"b64_json":"b"},{"b64_json":"c"}]}`))
	require.Equal(t, 3, dataCounter.Count())
}

func TestOutputCounterCountsMultilineSSEBodyPayload(t *testing.T) {
	counter := NewOutputCounter()
	counter.AddSSEBody(
		"data: {\"type\":\"image_generation.completed\",\n" +
			"data: \"b64_json\":\"final-a\"}\n\n" +
			"data: [DONE]\n\n",
	)
	require.Equal(t, 1, counter.Count())
}

func TestOutputCounterFallsBackForInvalidMultilineSSEBody(t *testing.T) {
	counter := NewOutputCounter()
	counter.AddSSEBody(
		"data: {\"type\":\"image_generation.completed\",\"b64_json\":\"final-a\"}\n" +
			"data: {\"type\":\"image_generation.completed\",\"b64_json\":\"final-b\"}\n\n",
	)
	require.Equal(t, 2, counter.Count())
}

func TestCollectResponseOutputSizesFromJSONBytes(t *testing.T) {
	body := []byte(`{
		"output": [
			{"id":"ig_1","type":"image_generation_call","result":"final-a","size":"3840x2160"},
			{"id":"ig_2","type":"image_generation_call","result":"final-b","size":"1024x1024"}
		]
	}`)

	require.Equal(t, 2, CountResponseOutputsFromJSONBytes(body))
	require.Equal(t, []string{"3840x2160", "1024x1024"}, CollectResponseOutputSizesFromJSONBytes(body))
}

func TestCollectResponseOutputSizesFromImagesAPIData(t *testing.T) {
	body := []byte(`{
		"data": [
			{"b64_json":"final-a","size":"2048x1152"},
			{"b64_json":"final-b","size":"2048x1152"}
		]
	}`)

	require.Equal(t, 2, CountResponseOutputsFromJSONBytes(body))
	require.Equal(t, []string{"2048x1152", "2048x1152"}, CollectResponseOutputSizesFromJSONBytes(body))
}

func TestCollectOutputSizesFromSSEBody(t *testing.T) {
	body := "data: {\"type\":\"response.output_item.done\",\"item\":{\"id\":\"ig_1\",\"type\":\"image_generation_call\",\"result\":\"final-a\",\"size\":\"3840x2160\"}}\n\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"output\":[{\"id\":\"ig_1\",\"type\":\"image_generation_call\",\"result\":\"final-a\"},{\"id\":\"ig_2\",\"type\":\"image_generation_call\",\"result\":\"final-b\",\"size\":\"1024x1024\"}]}}\n\n" +
		"data: [DONE]\n\n"

	require.Equal(t, 2, CountOutputsFromSSEBody(body))
	require.Equal(t, []string{"3840x2160", "1024x1024"}, CollectOutputSizesFromSSEBody(body))
}

func TestOutputCounterDoesNotCountTextOnlyResponses(t *testing.T) {
	body := `data: {"type":"response.output_item.done","item":{"id":"item_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello"}]}}

data: {"type":"response.completed","response":{"output":[{"id":"item_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello"}]}],"usage":{"input_tokens":10,"output_tokens":5}}}

data: [DONE]`

	require.Zero(t, CountOutputsFromSSEBody(body))
}

func TestOutputCounterCountsOnlyImageDataArrayItems(t *testing.T) {
	nonImage := []byte(`{"data":[{"id":"not_an_image","status":"done"}]}`)
	image := []byte(`{"data":[{"url":"https://example.com/img.png"}]}`)

	require.Zero(t, CountResponseOutputsFromJSONBytes(nonImage))
	require.Equal(t, 1, CountResponseOutputsFromJSONBytes(image))
}
