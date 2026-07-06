package gateway

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestIsOpenAICompatResponsesTerminalEvent(t *testing.T) {
	tests := []struct {
		eventType string
		want      bool
	}{
		{"response.completed", true},
		{" response.done ", true},
		{"response.incomplete", true},
		{"response.failed", true},
		{"response.output_text.delta", false},
		{"", false},
		{"Response.Completed", false},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			if got := IsOpenAICompatResponsesTerminalEvent(tt.eventType); got != tt.want {
				t.Fatalf("IsOpenAICompatResponsesTerminalEvent(%q) = %v, want %v", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestOpenAIStreamEventOutputBoundaries(t *testing.T) {
	if !IsOpenAIStreamPreambleEvent(" response.created ") {
		t.Fatal("expected response.created to be preamble")
	}
	if IsOpenAIStreamPreambleEvent("response.output_text.delta") {
		t.Fatal("did not expect delta to be preamble")
	}
	if OpenAIStreamDataStartsClientOutput(`{"type":"response.created"}`, "response.created") {
		t.Fatal("preamble data must not start client output")
	}
	if OpenAIStreamDataStartsClientOutput(`{"type":"response.failed"}`, "response.failed") {
		t.Fatal("failed event must not start client output")
	}
	if !OpenAIStreamDataStartsClientOutput(`{"type":"response.output_text.delta"}`, "response.output_text.delta") {
		t.Fatal("delta data should start client output")
	}
}

func TestExtractOpenAISSEDataAndEventLine(t *testing.T) {
	data, ok := ExtractOpenAISSEDataLine(`data: {"type":"x"}`)
	if !ok || data != `{"type":"x"}` {
		t.Fatalf("ExtractOpenAISSEDataLine standard = (%q, %v)", data, ok)
	}
	data, ok = ExtractOpenAISSEDataLine(`data:	{"type":"x"}`)
	if !ok || data != `{"type":"x"}` {
		t.Fatalf("ExtractOpenAISSEDataLine tab = (%q, %v)", data, ok)
	}
	if _, ok = ExtractOpenAISSEDataLine(`event: message`); ok {
		t.Fatal("non-data line should not match data")
	}

	event, ok := ExtractOpenAISSEEventLine(`event: response.completed `)
	if !ok || event != "response.completed" {
		t.Fatalf("ExtractOpenAISSEEventLine = (%q, %v)", event, ok)
	}
	if _, ok = ExtractOpenAISSEEventLine(`data: {"type":"x"}`); ok {
		t.Fatal("non-event line should not match event")
	}
}

func TestForEachOpenAISSEDataPayload(t *testing.T) {
	body := "event: response.output_text.delta\n" +
		"data: {\"type\":\"response.output_text.delta\",\n" +
		"data: \"delta\":\"hi\"}\n\n" +
		"data: [DONE]\n\n" +
		"data: {\"type\":\"response.completed\"}\n\n"
	var payloads []string
	ForEachOpenAISSEDataPayload(body, func(data []byte) {
		payloads = append(payloads, string(data))
	})
	if len(payloads) != 2 {
		t.Fatalf("payload count = %d, want 2: %#v", len(payloads), payloads)
	}
	if payloads[0] != "{\"type\":\"response.output_text.delta\",\n\"delta\":\"hi\"}" {
		t.Fatalf("payload[0] = %q", payloads[0])
	}
	if payloads[1] != `{"type":"response.completed"}` {
		t.Fatalf("payload[1] = %q", payloads[1])
	}
}

func TestExtractOpenAISSETerminalEvent(t *testing.T) {
	body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"hi\"}\n\n" +
		"data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"message\":\"x\"}}}\n\n"
	eventType, payload, ok := ExtractOpenAISSETerminalEvent(body)
	if !ok {
		t.Fatal("expected terminal event")
	}
	if eventType != "response.failed" {
		t.Fatalf("eventType = %q", eventType)
	}
	if string(payload) != `{"type":"response.failed","response":{"error":{"message":"x"}}}` {
		t.Fatalf("payload = %q", string(payload))
	}

	_, _, ok = ExtractOpenAISSETerminalEvent("data: {\"type\":\"response.output_text.delta\"}\n\n")
	if ok {
		t.Fatal("did not expect terminal event")
	}
}

func TestOpenAICompatPayloadWithEventType(t *testing.T) {
	if got := OpenAICompatPayloadWithEventType(`{"delta":"hi"}`, "response.output_text.delta"); got != `{"delta":"hi","type":"response.output_text.delta"}` {
		t.Fatalf("patched payload = %s", got)
	}
	if got := OpenAICompatPayloadWithEventType(`{"type":"existing"}`, "response.output_text.delta"); got != `{"type":"existing"}` {
		t.Fatalf("existing type should be preserved: %s", got)
	}
	if got := OpenAICompatPayloadWithEventType(`[DONE]`, "response.done"); got != `[DONE]` {
		t.Fatalf("[DONE] should be preserved: %s", got)
	}
}

func TestExtractOpenAIResponseIDFromJSONBytes(t *testing.T) {
	if got := ExtractOpenAIResponseIDFromJSONBytes([]byte(`{"id":"resp_json"}`)); got != "resp_json" {
		t.Fatalf("root id = %q", got)
	}
	if got := ExtractOpenAIResponseIDFromJSONBytes([]byte(`{"type":"response.completed","response":{"id":"resp_sse"}}`)); got != "resp_sse" {
		t.Fatalf("response id = %q", got)
	}
	if got := ExtractOpenAIResponseIDFromJSONBytes([]byte(`{"response":{}}`)); got != "" {
		t.Fatalf("empty id = %q", got)
	}
	if got := ExtractOpenAIResponseIDFromJSONBytes([]byte(`not-json`)); got != "" {
		t.Fatalf("invalid json id = %q", got)
	}
}

func TestExtractCodexFinalResponse(t *testing.T) {
	body := "event: message\n" +
		"data: {\"type\":\"response.in_progress\",\"response\":{\"id\":\"resp_1\"}}\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\",\"model\":\"gpt-4o\",\"usage\":{\"input_tokens\":11}}}\n" +
		"data: [DONE]\n"

	finalResp, ok := ExtractCodexFinalResponse(body)
	if !ok {
		t.Fatal("expected final response")
	}
	if string(finalResp) != `{"id":"resp_1","model":"gpt-4o","usage":{"input_tokens":11}}` {
		t.Fatalf("final response = %s", string(finalResp))
	}
}

func TestSanitizeOpenAIResponseFailedEventForClient(t *testing.T) {
	payload := []byte(`{"type":"response.failed","response":{"id":"resp_1","instructions":"secret","output":[{"type":"message"}],"usage":{"input_tokens":1},"error":{"message":"context too long"}}}`)
	got, changed := SanitizeOpenAIResponseFailedEventForClient(payload, "response.failed")
	if !changed {
		t.Fatal("expected sanitized payload")
	}
	if gjson.GetBytes(got, "response.instructions").Exists() ||
		gjson.GetBytes(got, "response.output").Exists() ||
		gjson.GetBytes(got, "response.usage").Exists() {
		t.Fatalf("sensitive response fields still exist: %s", string(got))
	}
	if !gjson.GetBytes(got, "response.error.message").Exists() {
		t.Fatalf("error should be preserved: %s", string(got))
	}

	got, changed = SanitizeOpenAIResponseFailedEventForClient(payload, "response.completed")
	if changed || string(got) != string(payload) {
		t.Fatalf("non-failed event should be unchanged")
	}
}
