package service

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestOpenAIWSCyberPolicyMark_ResponseFailed 验证 WS 路径 response.failed cyber_policy 标记逻辑。
//
// 全量转发循环（forwardOpenAIWSV2 / sendAndRelay）依赖真实 WebSocket 连接，
// 无法在单元测试中驱动。本测试通过直接调用转发循环内使用的两个函数
// detectOpenAICyberPolicy + MarkOpsCyberPolicy，覆盖「从 response.failed 帧
// 到 gin context 写入」的完整调用序列，等同于循环体内对应代码段的逻辑验证。
// 全量 WS 端到端覆盖由后续集成测试（Task 12 handler 编排）承担。
func TestOpenAIWSCyberPolicyMark_ResponseFailed(t *testing.T) {
	// 构造一个真实的 response.failed 帧（cyber_policy 命中路径）。
	payload := []byte(`{"type":"response.failed","response":{"id":"resp_abc","status":"failed","error":{"code":"cyber_policy","message":"Request blocked by content policy."}}}`)

	// 验证 detectOpenAICyberPolicy 能从 response.error.code 路径识别。
	hit, code, msg := detectOpenAICyberPolicy(payload)
	require.True(t, hit, "detectOpenAICyberPolicy should return true for cyber_policy payload")
	require.Equal(t, "cyber_policy", code)
	require.Equal(t, "Request blocked by content policy.", msg)

	// 构造 gin test context，模拟转发循环调用 MarkOpsCyberPolicy。
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	usage := OpenAIUsage{InputTokens: 42, OutputTokens: 7}
	MarkOpsCyberPolicy(c, CyberPolicyMark{
		Code:           code,
		Message:        msg,
		Body:           truncateString(string(payload), 4096),
		UpstreamStatus: 200,
		UpstreamInTok:  usage.InputTokens,
		UpstreamOutTok: usage.OutputTokens,
	})

	mark := GetOpsCyberPolicy(c)
	require.NotNil(t, mark, "GetOpsCyberPolicy should return non-nil after MarkOpsCyberPolicy")
	require.Equal(t, "cyber_policy", mark.Code)
	require.Equal(t, "Request blocked by content policy.", mark.Message)
	require.Equal(t, 200, mark.UpstreamStatus)
	require.Equal(t, 42, mark.UpstreamInTok)
	require.Equal(t, 7, mark.UpstreamOutTok)

	// 验证幂等性：再次标记不覆盖首个。
	MarkOpsCyberPolicy(c, CyberPolicyMark{Code: "cyber_policy", Message: "second call"})
	require.Equal(t, "Request blocked by content policy.", GetOpsCyberPolicy(c).Message, "second MarkOpsCyberPolicy call must not overwrite first")
}

// TestOpenAIWSCyberPolicyMark_NonCyberPayload 验证非 cyber_policy 的 response.failed 不触发标记。
func TestOpenAIWSCyberPolicyMark_NonCyberPayload(t *testing.T) {
	payload := []byte(`{"type":"response.failed","response":{"id":"resp_xyz","status":"failed","error":{"code":"server_error","message":"Internal error"}}}`)

	hit, _, _ := detectOpenAICyberPolicy(payload)
	require.False(t, hit, "detectOpenAICyberPolicy should return false for non-cyber_policy error code")
}
