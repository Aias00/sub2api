package service

import "github.com/Aias00/cloudbase/internal/gateway"

func buildOpenAIEndpointURL(base string, endpoint string) string {
	return gateway.BuildOpenAIEndpointURL(base, endpoint)
}

func buildOpenAIResponsesInputTokensURL(base string) string {
	return gateway.BuildOpenAIResponsesInputTokensURL(base)
}

func openAIBaseURLHasVersionSuffix(raw string) bool {
	return gateway.OpenAIBaseURLHasVersionSuffix(raw)
}

func isOpenAIAPIVersionSegment(segment string) bool {
	return gateway.IsOpenAIAPIVersionSegment(segment)
}
