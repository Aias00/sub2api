package service

import "github.com/Aias00/cloudbase/internal/gateway"

const (
	OpenAIPreviousResponseIDKindEmpty      = gateway.OpenAIPreviousResponseIDKindEmpty
	OpenAIPreviousResponseIDKindResponseID = gateway.OpenAIPreviousResponseIDKindResponseID
	OpenAIPreviousResponseIDKindMessageID  = gateway.OpenAIPreviousResponseIDKindMessageID
	OpenAIPreviousResponseIDKindUnknown    = gateway.OpenAIPreviousResponseIDKindUnknown
)

// ClassifyOpenAIPreviousResponseIDKind classifies previous_response_id to improve diagnostics.
func ClassifyOpenAIPreviousResponseIDKind(id string) string {
	return gateway.ClassifyOpenAIPreviousResponseIDKind(id)
}

func IsOpenAIPreviousResponseIDLikelyMessageID(id string) bool {
	return gateway.IsOpenAIPreviousResponseIDLikelyMessageID(id)
}
