package gateway

import "testing"

func TestBuildOpenAIStreamErrorEventPayload(t *testing.T) {
	got := BuildOpenAIStreamErrorEventPayload(`stream "bad"`)
	want := `{"type":"error","sequence_number":0,"error":{"type":"upstream_error","message":"stream \"bad\"","code":"stream \"bad\""}}`
	if got != want {
		t.Fatalf("payload = %s, want %s", got, want)
	}
}

func TestBuildOpenAIStreamErrorSSE(t *testing.T) {
	got := BuildOpenAIStreamErrorSSE("stream_timeout")
	want := "data: " + `{"type":"error","sequence_number":0,"error":{"type":"upstream_error","message":"stream_timeout","code":"stream_timeout"}}` + "\n\n"
	if got != want {
		t.Fatalf("sse = %q, want %q", got, want)
	}
}
