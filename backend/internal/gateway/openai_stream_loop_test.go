package gateway

import (
	"testing"
	"time"
)

func TestResolveOpenAIStreamLoopTimings(t *testing.T) {
	data, keepalive := ResolveOpenAIStreamLoopTimings(3, 7)
	if data != 3*time.Second || keepalive != 7*time.Second {
		t.Fatalf("ResolveOpenAIStreamLoopTimings() = %s, %s", data, keepalive)
	}
	data, keepalive = ResolveOpenAIStreamLoopTimings(0, -1)
	if data != 0 || keepalive != 0 {
		t.Fatalf("ResolveOpenAIStreamLoopTimings disabled = %s, %s", data, keepalive)
	}
}

func TestIsOpenAISSEDoneData(t *testing.T) {
	if !IsOpenAISSEDoneData(" [DONE]\n") {
		t.Fatal("expected [DONE] marker")
	}
	if IsOpenAISSEDoneData(`{"type":"response.completed"}`) {
		t.Fatal("json payload should not be done marker")
	}
}

func TestEvaluateOpenAIStreamDataInterval(t *testing.T) {
	now := time.Unix(100, 0)
	if got := EvaluateOpenAIStreamDataInterval(now, now.Add(-time.Second), 2*time.Second, false); got != OpenAIStreamTimeoutContinue {
		t.Fatalf("fresh read decision = %v", got)
	}
	if got := EvaluateOpenAIStreamDataInterval(now, now.Add(-3*time.Second), 2*time.Second, false); got != OpenAIStreamTimeoutDataIntervalExceeded {
		t.Fatalf("timeout decision = %v", got)
	}
	if got := EvaluateOpenAIStreamDataInterval(now, now.Add(-3*time.Second), 2*time.Second, true); got != OpenAIStreamTimeoutIncompleteAfterDisconnect {
		t.Fatalf("disconnect timeout decision = %v", got)
	}
}

func TestShouldSendOpenAIStreamKeepalive(t *testing.T) {
	now := time.Unix(100, 0)
	if !ShouldSendOpenAIStreamKeepalive(now, now.Add(-3*time.Second), 2*time.Second, false, false) {
		t.Fatal("expected keepalive after idle interval")
	}
	if ShouldSendOpenAIStreamKeepalive(now, now.Add(-time.Second), 2*time.Second, false, false) {
		t.Fatal("fresh downstream write should not keepalive")
	}
	if ShouldSendOpenAIStreamKeepalive(now, now.Add(-3*time.Second), 2*time.Second, true, false) {
		t.Fatal("disconnected client should not keepalive")
	}
	if ShouldSendOpenAIStreamKeepalive(now, now.Add(-3*time.Second), 2*time.Second, false, true) {
		t.Fatal("held client output should not keepalive")
	}
}

func TestShouldFlushOpenAIStreamData(t *testing.T) {
	if ShouldFlushOpenAIStreamData(false, true, true, false) {
		t.Fatal("undrained queue should not flush")
	}
	if !ShouldFlushOpenAIStreamData(true, false, true, false) {
		t.Fatal("first output should flush")
	}
	if !ShouldFlushOpenAIStreamData(true, true, false, true) {
		t.Fatal("started client output should flush when queue is drained")
	}
	if ShouldFlushOpenAIStreamData(true, false, false, false) {
		t.Fatal("preamble before client output should not flush")
	}
}

func TestShouldFlushOpenAIStreamForwardedLine(t *testing.T) {
	if ShouldFlushOpenAIStreamForwardedLine(false, true) {
		t.Fatal("undrained queue should not flush forwarded line")
	}
	if ShouldFlushOpenAIStreamForwardedLine(true, false) {
		t.Fatal("line before client output should not flush")
	}
	if !ShouldFlushOpenAIStreamForwardedLine(true, true) {
		t.Fatal("drained forwarded line after output should flush")
	}
}

func TestResolveOpenAIStreamDownstreamWrite(t *testing.T) {
	if got := ResolveOpenAIStreamDownstreamWrite(nil, nil, true); got.ClientDisconnected || !got.OutputStarted || !got.WriteSucceeded {
		t.Fatalf("successful write result = %+v", got)
	}
	if got := ResolveOpenAIStreamDownstreamWrite(nil, assertErr{}, true); !got.ClientDisconnected || got.OutputStarted || got.WriteSucceeded {
		t.Fatalf("flush error result = %+v", got)
	}
	if got := ResolveOpenAIStreamDownstreamWrite(assertErr{}, nil, true); !got.ClientDisconnected || got.OutputStarted || got.WriteSucceeded {
		t.Fatalf("write error result = %+v", got)
	}
}

func TestWriteOpenAIStreamDownstreamFrame(t *testing.T) {
	writer := &openAIStreamDownstreamWriterStub{}
	if got := WriteOpenAIStreamDownstreamFrame(writer, "data: {}\n", true, true); got.ClientDisconnected || !got.OutputStarted || !got.WriteSucceeded {
		t.Fatalf("successful frame write result = %+v", got)
	}
	if len(writer.writes) != 1 || writer.writes[0] != "data: {}\n" {
		t.Fatalf("writes = %#v", writer.writes)
	}
	if writer.flushes != 1 {
		t.Fatalf("flushes = %d", writer.flushes)
	}

	writer = &openAIStreamDownstreamWriterStub{flushErr: assertErr{}}
	if got := WriteOpenAIStreamDownstreamFrame(writer, "data: {}\n", true, true); !got.ClientDisconnected || got.OutputStarted || got.WriteSucceeded {
		t.Fatalf("flush error result = %+v", got)
	}

	writer = &openAIStreamDownstreamWriterStub{writeErr: assertErr{}}
	if got := WriteOpenAIStreamDownstreamFrame(writer, "data: {}\n", true, true); !got.ClientDisconnected || got.OutputStarted || got.WriteSucceeded {
		t.Fatalf("write error result = %+v", got)
	}
	if writer.flushes != 0 {
		t.Fatalf("write error should skip flush, flushes = %d", writer.flushes)
	}

	if got := WriteOpenAIStreamDownstreamFrame(nil, "data: {}\n", true, true); !got.ClientDisconnected || got.OutputStarted || got.WriteSucceeded {
		t.Fatalf("nil writer result = %+v", got)
	}
}

func TestWriteOpenAIStreamDownstreamLine(t *testing.T) {
	writer := &openAIStreamDownstreamWriterStub{}
	if got := WriteOpenAIStreamDownstreamLine(writer, "event: ping", false, false); got.ClientDisconnected || got.OutputStarted || !got.WriteSucceeded {
		t.Fatalf("line write result = %+v", got)
	}
	if len(writer.writes) != 1 || writer.writes[0] != "event: ping\n" {
		t.Fatalf("writes = %#v", writer.writes)
	}
	if writer.flushes != 0 {
		t.Fatalf("flushes = %d", writer.flushes)
	}
}

type openAIStreamDownstreamWriterStub struct {
	writes   []string
	writeErr error
	flushErr error
	flushes  int
}

func (s *openAIStreamDownstreamWriterStub) WriteString(v string) (int, error) {
	if s.writeErr != nil {
		return 0, s.writeErr
	}
	s.writes = append(s.writes, v)
	return len(v), nil
}

func (s *openAIStreamDownstreamWriterStub) Flush() error {
	s.flushes++
	return s.flushErr
}

type assertErr struct{}

func (assertErr) Error() string { return "assert" }
