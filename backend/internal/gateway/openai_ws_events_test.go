package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAIWSTokenEvent_TerminalEventsExcluded(t *testing.T) {
	cases := []struct {
		name      string
		eventType string
		want      bool
	}{
		{name: "empty", eventType: "", want: false},
		{name: "whitespace_trimmed_empty", eventType: "   ", want: false},
		{name: "response.created", eventType: "response.created", want: false},
		{name: "response.in_progress", eventType: "response.in_progress", want: false},
		{name: "response.output_item.added", eventType: "response.output_item.added", want: false},
		{name: "response.output_item.done", eventType: "response.output_item.done", want: false},
		{name: "terminal_response.completed", eventType: "response.completed", want: false},
		{name: "terminal_response.done", eventType: "response.done", want: false},
		{name: "terminal_response.completed_padded", eventType: "  response.completed  ", want: false},
		{name: "terminal_response.done_padded", eventType: "  response.done  ", want: false},
		{name: "delta_text", eventType: "response.output_text.delta", want: true},
		{name: "delta_audio_transcript", eventType: "response.audio_transcript.delta", want: true},
		{name: "delta_function_call_arguments", eventType: "response.function_call_arguments.delta", want: true},
		{name: "output_text_done", eventType: "response.output_text.done", want: true},
		{name: "output_text_annotation_added", eventType: "response.output_text.annotation.added", want: true},
		{name: "output_audio_done", eventType: "response.output_audio.done", want: true},
		{name: "reasoning_summary_delta", eventType: "response.reasoning_summary_text.delta", want: true},
		{name: "unrelated_event_error", eventType: "error", want: false},
		{name: "unknown_event_without_match", eventType: "response.reasoning_summary_part.added", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsOpenAIWSTokenEvent(tc.eventType)
			require.Equal(t, tc.want, got, "IsOpenAIWSTokenEvent(%q)", tc.eventType)
		})
	}
}

func TestIsOpenAIWSTokenEvent_DisjointWithTerminal(t *testing.T) {
	terminalEvents := []string{
		"response.completed",
		"response.done",
		"response.failed",
		"response.incomplete",
		"response.cancelled",
		"response.canceled",
	}
	for _, ev := range terminalEvents {
		t.Run(ev, func(t *testing.T) {
			require.True(t, IsOpenAIWSTerminalEvent(ev), "expected terminal event %q to be classified as terminal", ev)
			require.False(t, IsOpenAIWSTokenEvent(ev), "terminal event %q must not be classified as token event", ev)
		})
	}
}

func TestOpenAIWSEventMayContainModel(t *testing.T) {
	for _, eventType := range []string{
		"response.created",
		"response.in_progress",
		"response.completed",
		"response.done",
		"response.failed",
		"response.incomplete",
		"response.cancelled",
		"response.canceled",
	} {
		require.True(t, OpenAIWSEventMayContainModel(eventType), eventType)
		require.True(t, OpenAIWSEventMayContainModel("  "+eventType+"  "), eventType)
	}
	require.False(t, OpenAIWSEventMayContainModel("response.output_text.delta"))
	require.False(t, OpenAIWSEventMayContainModel(""))
}

func TestOpenAIWSEventMayContainToolCalls(t *testing.T) {
	for _, eventType := range []string{
		"response.output_item.added",
		"response.output_item.done",
		"response.completed",
		"response.done",
		"response.function_call_arguments.delta",
		"response.tool_call.delta",
	} {
		require.True(t, OpenAIWSEventMayContainToolCalls(eventType), eventType)
		require.True(t, OpenAIWSEventMayContainToolCalls("  "+eventType+"  "), eventType)
	}
	require.False(t, OpenAIWSEventMayContainToolCalls("response.output_text.delta"))
	require.False(t, OpenAIWSEventMayContainToolCalls(""))
}

func TestOpenAIWSEventShouldParseUsage(t *testing.T) {
	for _, eventType := range []string{
		"response.completed",
		"response.done",
		"response.failed",
		"response.incomplete",
		"response.cancelled",
		"response.canceled",
	} {
		require.True(t, OpenAIWSEventShouldParseUsage(eventType), eventType)
		require.True(t, OpenAIWSEventShouldParseUsage("  "+eventType+"  "), eventType)
	}
	require.False(t, OpenAIWSEventShouldParseUsage("response.output_text.delta"))
	require.False(t, OpenAIWSEventShouldParseUsage(""))
}
