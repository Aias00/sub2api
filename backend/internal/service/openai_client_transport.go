package service

import (
	"github.com/Aias00/cloudbase/internal/gateway"
	"github.com/gin-gonic/gin"
)

type OpenAIClientTransport = gateway.OpenAIClientTransport

const (
	OpenAIClientTransportUnknown = gateway.OpenAIClientTransportUnknown
	OpenAIClientTransportHTTP    = gateway.OpenAIClientTransportHTTP
	OpenAIClientTransportWS      = gateway.OpenAIClientTransportWS
)

// SetOpenAIClientTransport 标记当前请求的客户端入站协议。
func SetOpenAIClientTransport(c *gin.Context, transport OpenAIClientTransport) {
	gateway.SetOpenAIClientTransport(c, transport)
}

// GetOpenAIClientTransport 读取当前请求的客户端入站协议。
func GetOpenAIClientTransport(c *gin.Context) OpenAIClientTransport {
	return gateway.GetOpenAIClientTransport(c)
}

func resolveOpenAIWSDecisionByClientTransport(
	decision OpenAIWSProtocolDecision,
	clientTransport OpenAIClientTransport,
) OpenAIWSProtocolDecision {
	if clientTransport == OpenAIClientTransportHTTP {
		return openAIWSHTTPDecision("client_protocol_http")
	}
	return decision
}
