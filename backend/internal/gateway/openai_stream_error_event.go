package gateway

import "strconv"

func BuildOpenAIStreamErrorEventPayload(reason string) string {
	quoted := strconv.Quote(reason)
	return `{"type":"error","sequence_number":0,"error":{"type":"upstream_error","message":` + quoted + `,"code":` + quoted + `}}`
}

func BuildOpenAIStreamErrorSSE(reason string) string {
	return "data: " + BuildOpenAIStreamErrorEventPayload(reason) + "\n\n"
}
