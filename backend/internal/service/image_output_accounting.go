package service

import imagecore "github.com/Aias00/cloudbase/internal/image"

type openAIImageOutputCounter = imagecore.OutputCounter

func newOpenAIImageOutputCounter() *openAIImageOutputCounter {
	return imagecore.NewOutputCounter()
}

func countOpenAIResponseImageOutputsFromJSONBytes(body []byte) int {
	return imagecore.CountResponseOutputsFromJSONBytes(body)
}

func collectOpenAIResponseImageOutputSizesFromJSONBytes(body []byte) []string {
	return imagecore.CollectResponseOutputSizesFromJSONBytes(body)
}

func countOpenAIImageOutputsFromSSEBody(body string) int {
	return imagecore.CountOutputsFromSSEBody(body)
}

func collectOpenAIImageOutputSizesFromSSEBody(body string) []string {
	return imagecore.CollectOutputSizesFromSSEBody(body)
}
