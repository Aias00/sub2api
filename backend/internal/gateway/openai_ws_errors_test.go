package gateway

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyOpenAIWSErrorEvent(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		errType     string
		message     string
		wantReason  string
		wantRecover bool
	}{
		{
			name:        "upgrade code",
			code:        "upgrade_required",
			message:     "Upgrade required",
			wantReason:  "upgrade_required",
			wantRecover: true,
		},
		{
			name:        "previous response missing",
			code:        "previous_response_not_found",
			message:     "not found",
			wantReason:  "previous_response_not_found",
			wantRecover: true,
		},
		{
			name:        "rate limit stays non recoverable",
			code:        "rate_limit_exceeded",
			errType:     "rate_limit_error",
			message:     "rate limit exceeded",
			wantReason:  "upstream_rate_limited",
			wantRecover: false,
		},
		{
			name:        "server error is recoverable",
			code:        "server_error",
			errType:     "server_error",
			message:     "server failed",
			wantReason:  "upstream_error_event",
			wantRecover: true,
		},
		{
			name:        "unknown event error",
			code:        "unknown",
			message:     "unknown",
			wantReason:  "event_error",
			wantRecover: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotReason, gotRecover := ClassifyOpenAIWSErrorEvent(tc.code, tc.errType, tc.message)
			require.Equal(t, tc.wantReason, gotReason)
			require.Equal(t, tc.wantRecover, gotRecover)
		})
	}
}

func TestOpenAIWSErrorHTTPStatus(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		errType string
		want    int
	}{
		{name: "invalid request", code: "invalid_request", errType: "invalid_request_error", want: http.StatusBadRequest},
		{name: "auth", code: "invalid_api_key", errType: "authentication_error", want: http.StatusUnauthorized},
		{name: "permission", code: "forbidden", errType: "permission_error", want: http.StatusForbidden},
		{name: "rate limit", code: "rate_limit_exceeded", errType: "rate_limit_error", want: http.StatusTooManyRequests},
		{name: "usage limit type", errType: "usage_limit_reached", want: http.StatusTooManyRequests},
		{name: "server fallback", code: "server_error", errType: "server_error", want: http.StatusBadGateway},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, OpenAIWSErrorHTTPStatus(tc.code, tc.errType))
		})
	}
}

func TestIsOpenAIWSRateLimitError(t *testing.T) {
	require.True(t, IsOpenAIWSRateLimitError("insufficient_quota", "", ""))
	require.True(t, IsOpenAIWSRateLimitError("", "usage_limit_reached", ""))
	require.True(t, IsOpenAIWSRateLimitError("", "", "usage limit reached"))
	require.True(t, IsOpenAIWSRateLimitError("", "", "rate limit exceeded"))
	require.False(t, IsOpenAIWSRateLimitError("server_error", "server_error", "server failed"))
}
