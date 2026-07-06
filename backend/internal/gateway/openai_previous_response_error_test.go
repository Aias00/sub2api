package gateway

import (
	"net/http"
	"testing"
)

func TestIsOpenAICompatPreviousResponseNotFound(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		msg        string
		body       string
		want       bool
	}{
		{name: "code in body", statusCode: http.StatusBadRequest, body: `{"error":{"code":"previous_response_not_found"}}`, want: true},
		{name: "message in body", statusCode: http.StatusNotFound, body: `{"error":{"message":"previous response not found"}}`, want: true},
		{name: "unsupported parameter compatibility", statusCode: http.StatusBadRequest, msg: "unsupported parameter previous_response_id", want: true},
		{name: "wrong status", statusCode: http.StatusInternalServerError, msg: "previous response not found", want: false},
		{name: "unrelated bad request", statusCode: http.StatusBadRequest, msg: "model required", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOpenAICompatPreviousResponseNotFound(tt.statusCode, tt.msg, []byte(tt.body))
			if got != tt.want {
				t.Fatalf("IsOpenAICompatPreviousResponseNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsOpenAICompatPreviousResponseUnsupported(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		msg        string
		body       string
		want       bool
	}{
		{name: "unsupported parameter", statusCode: http.StatusBadRequest, msg: "unsupported parameter: previous_response_id", want: true},
		{name: "only websocket", statusCode: http.StatusBadRequest, body: `{"error":{"message":"previous_response_id only supported on responses websocket"}}`, want: true},
		{name: "not supported", statusCode: http.StatusBadRequest, body: `{"error":{"message":"previous_response_id is not supported"}}`, want: true},
		{name: "wrong status", statusCode: http.StatusNotFound, msg: "previous_response_id is not supported", want: false},
		{name: "missing marker", statusCode: http.StatusBadRequest, msg: "unsupported parameter: stream", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOpenAICompatPreviousResponseUnsupported(tt.statusCode, tt.msg, []byte(tt.body))
			if got != tt.want {
				t.Fatalf("IsOpenAICompatPreviousResponseUnsupported() = %v, want %v", got, tt.want)
			}
		})
	}
}
