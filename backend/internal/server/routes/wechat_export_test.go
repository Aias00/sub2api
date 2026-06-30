package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWeChatExportRoutesRegisterExpectedPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	RegisterWeChatExportRoutes(v1, &handler.Handlers{
		WeChatExport: &handler.WeChatExportHandler{},
	}, func(c *gin.Context) {}, nil)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, routeKey := range []string{
		http.MethodGet + " /api/v1/wechat/session",
		http.MethodPost + " /api/v1/wechat/session/qrcode",
		http.MethodGet + " /api/v1/wechat/session/poll/:sessionID",
		http.MethodPost + " /api/v1/wechat/session/validate",
		http.MethodPost + " /api/v1/wechat/session/logout",
		http.MethodGet + " /api/v1/wechat/accounts/search",
		http.MethodPost + " /api/v1/wechat/accounts/bind",
		http.MethodPost + " /api/v1/wechat/accounts/:accountID/sync",
		http.MethodGet + " /api/v1/wechat/articles",
		http.MethodPost + " /api/v1/wechat/articles/import-link",
		http.MethodPost + " /api/v1/wechat/tasks/quote",
		http.MethodPost + " /api/v1/wechat/tasks",
		http.MethodGet + " /api/v1/wechat/tasks",
		http.MethodGet + " /api/v1/wechat/worker/status",
		http.MethodGet + " /api/v1/wechat/tasks/:taskID",
		http.MethodPost + " /api/v1/wechat/tasks/:taskID/cancel",
		http.MethodPost + " /api/v1/wechat/tasks/:taskID/retry",
		http.MethodGet + " /api/v1/wechat/tasks/:taskID/logs",
		http.MethodGet + " /api/v1/wechat/tasks/:taskID/artifacts",
		http.MethodGet + " /api/v1/wechat/tasks/:taskID/artifacts.zip",
		http.MethodGet + " /api/v1/wechat/artifacts/:artifactID/download",
		http.MethodGet + " /api/v1/wechat/worker/health",
		http.MethodPost + " /api/v1/wechat/worker/tasks/claim",
		http.MethodPost + " /api/v1/wechat/worker/articles/:articleID/enrich",
		http.MethodPost + " /api/v1/wechat/worker/articles/:articleID/engagement",
		http.MethodPost + " /api/v1/wechat/worker/tasks/:taskID/logs",
		http.MethodPost + " /api/v1/wechat/worker/tasks/:taskID/complete",
		http.MethodPost + " /api/v1/wechat/worker/tasks/:taskID/fail",
	} {
		require.True(t, registered[routeKey], "missing route %s", routeKey)
	}
}
