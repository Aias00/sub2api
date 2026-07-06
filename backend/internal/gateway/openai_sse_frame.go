package gateway

import "strings"

type OpenAICompatSSEFrame struct {
	EventType string
	Data      string
}

type OpenAICompatSSEFrameParser struct {
	eventType string
	dataLines []string
}

func (p *OpenAICompatSSEFrameParser) AddLine(line string) (OpenAICompatSSEFrame, bool) {
	if line == "" {
		return p.dispatch()
	}
	if strings.HasPrefix(line, ":") {
		return OpenAICompatSSEFrame{}, false
	}
	if eventType, ok := ExtractOpenAISSEEventLine(line); ok {
		p.eventType = eventType
		return OpenAICompatSSEFrame{}, false
	}
	if data, ok := ExtractOpenAISSEDataLine(line); ok {
		p.dataLines = append(p.dataLines, data)
	}
	return OpenAICompatSSEFrame{}, false
}

func (p *OpenAICompatSSEFrameParser) Finish() (OpenAICompatSSEFrame, bool) {
	return p.dispatch()
}

func (p *OpenAICompatSSEFrameParser) dispatch() (OpenAICompatSSEFrame, bool) {
	frame := OpenAICompatSSEFrame{
		EventType: p.eventType,
		Data:      strings.Join(p.dataLines, "\n"),
	}
	p.eventType = ""
	p.dataLines = nil
	return frame, frame.Data != ""
}
