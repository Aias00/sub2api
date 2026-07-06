package gateway

import "time"

type OpenAIStreamScanEvent struct {
	Line string
	Err  error
}

type OpenAIStreamTimeoutDecision int

const (
	OpenAIStreamTimeoutContinue OpenAIStreamTimeoutDecision = iota
	OpenAIStreamTimeoutIncompleteAfterDisconnect
	OpenAIStreamTimeoutDataIntervalExceeded
)

func ResolveOpenAIStreamLoopTimings(dataIntervalSeconds, keepaliveSeconds int) (time.Duration, time.Duration) {
	var dataInterval time.Duration
	if dataIntervalSeconds > 0 {
		dataInterval = time.Duration(dataIntervalSeconds) * time.Second
	}
	var keepaliveInterval time.Duration
	if keepaliveSeconds > 0 {
		keepaliveInterval = time.Duration(keepaliveSeconds) * time.Second
	}
	return dataInterval, keepaliveInterval
}

func IsOpenAISSEDoneData(data string) bool {
	return trimSpace(data) == "[DONE]"
}

func EvaluateOpenAIStreamDataInterval(now, lastRead time.Time, interval time.Duration, clientDisconnected bool) OpenAIStreamTimeoutDecision {
	if interval <= 0 || now.Sub(lastRead) < interval {
		return OpenAIStreamTimeoutContinue
	}
	if clientDisconnected {
		return OpenAIStreamTimeoutIncompleteAfterDisconnect
	}
	return OpenAIStreamTimeoutDataIntervalExceeded
}

func ShouldSendOpenAIStreamKeepalive(now, lastDownstreamWrite time.Time, interval time.Duration, clientDisconnected, holdingClientOutput bool) bool {
	if interval <= 0 || clientDisconnected || holdingClientOutput {
		return false
	}
	return now.Sub(lastDownstreamWrite) >= interval
}

func ShouldFlushOpenAIStreamData(queueDrained, clientOutputStarted, startsClientOutput, firstTokenSeen bool) bool {
	if !queueDrained {
		return false
	}
	if !firstTokenSeen && startsClientOutput {
		return true
	}
	return clientOutputStarted || startsClientOutput
}

func ShouldFlushOpenAIStreamForwardedLine(queueDrained, clientOutputStarted bool) bool {
	return queueDrained && clientOutputStarted
}

type OpenAIStreamDownstreamWriteResult struct {
	ClientDisconnected bool
	OutputStarted      bool
	WriteSucceeded     bool
}

type OpenAIStreamDownstreamWriter interface {
	WriteString(string) (int, error)
	Flush() error
}

func ResolveOpenAIStreamDownstreamWrite(writeErr, flushErr error, startsClientOutput bool) OpenAIStreamDownstreamWriteResult {
	if writeErr != nil || flushErr != nil {
		return OpenAIStreamDownstreamWriteResult{ClientDisconnected: true}
	}
	return OpenAIStreamDownstreamWriteResult{
		OutputStarted:  startsClientOutput,
		WriteSucceeded: true,
	}
}

func WriteOpenAIStreamDownstreamFrame(w OpenAIStreamDownstreamWriter, frame string, shouldFlush, startsClientOutput bool) OpenAIStreamDownstreamWriteResult {
	if w == nil {
		return OpenAIStreamDownstreamWriteResult{ClientDisconnected: true}
	}
	_, writeErr := w.WriteString(frame)
	var flushErr error
	if writeErr == nil && shouldFlush {
		flushErr = w.Flush()
	}
	return ResolveOpenAIStreamDownstreamWrite(writeErr, flushErr, startsClientOutput)
}

func WriteOpenAIStreamDownstreamLine(w OpenAIStreamDownstreamWriter, line string, shouldFlush, startsClientOutput bool) OpenAIStreamDownstreamWriteResult {
	return WriteOpenAIStreamDownstreamFrame(w, line+"\n", shouldFlush, startsClientOutput)
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) {
		switch s[start] {
		case ' ', '\t', '\n', '\r':
			start++
		default:
			goto trimEnd
		}
	}
	return ""
trimEnd:
	end := len(s)
	for end > start {
		switch s[end-1] {
		case ' ', '\t', '\n', '\r':
			end--
		default:
			return s[start:end]
		}
	}
	return ""
}
