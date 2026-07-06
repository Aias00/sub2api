package gateway

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

// OpenAITransportErrorClass describes how to react to a transport-level upstream
// failure: the HTTP round trip never completed, so no HTTP status code exists.
type OpenAITransportErrorClass struct {
	// Persistent marks failures where retrying the same proxy/account is
	// pointless: rejected proxy credentials, a dead endpoint, or DNS/routing
	// failure. Service adapters can use this to temporarily remove candidates
	// from scheduling without putting side effects in this package.
	Persistent bool
}

var openAIPersistentTransportErrorMarkers = []string{
	"authentication failed",         // SOCKS5 RFC1929 / proxy credentials rejected.
	"proxy authentication required", // HTTP proxy 407.
	"connection refused",
	"no route to host",
	"network is unreachable",
	"no such host",
}

// ClassifyOpenAITransportError decides whether a transport-level upstream error
// is durable or transient. It intentionally classifies only pure error shapes;
// account eviction, logging, failover, and alert side effects stay in adapters.
func ClassifyOpenAITransportError(err error) OpenAITransportErrorClass {
	if err == nil {
		return OpenAITransportErrorClass{}
	}

	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EHOSTUNREACH) ||
		errors.Is(err, syscall.ENETUNREACH) {
		return OpenAITransportErrorClass{Persistent: true}
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return OpenAITransportErrorClass{Persistent: true}
	}

	msg := strings.ToLower(err.Error())
	for _, marker := range openAIPersistentTransportErrorMarkers {
		if strings.Contains(msg, marker) {
			return OpenAITransportErrorClass{Persistent: true}
		}
	}
	return OpenAITransportErrorClass{}
}
