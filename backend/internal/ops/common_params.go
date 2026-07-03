package ops

import (
	"fmt"
	"strconv"
	"strings"
)

type ErrorCorrelationKey struct {
	RequestID       string
	ClientRequestID string
}

func ParsePositiveID(raw string, errMessage string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%s", errMessage)
	}
	return id, nil
}

func ParseTruthyFlag(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func CapPageSize(pageSize int, max int) int {
	if max <= 0 || pageSize <= max {
		return pageSize
	}
	return max
}

func PickErrorCorrelationKey(requestIDRaw string, clientRequestIDRaw string) ErrorCorrelationKey {
	requestID := strings.TrimSpace(requestIDRaw)
	if requestID != "" {
		return ErrorCorrelationKey{RequestID: requestID}
	}
	return ErrorCorrelationKey{ClientRequestID: strings.TrimSpace(clientRequestIDRaw)}
}

func IsInvalidInputError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "invalid")
}
