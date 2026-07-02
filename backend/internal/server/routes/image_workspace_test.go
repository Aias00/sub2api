package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/cloudbase/internal/handler"
	"github.com/Wei-Shaw/cloudbase/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestImageWorkspaceRoutesRegisterExpectedPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterImageWorkspaceRoutes(v1, &handler.Handlers{
		ImageWorkspace: &handler.ImageWorkspaceHandler{},
	}, func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
		c.Next()
	}, nil)

	routes := map[string]bool{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	expected := []string{
		http.MethodPost + " /api/v1/image-workspace/tasks",
		http.MethodGet + " /api/v1/image-workspace/tasks",
		http.MethodGet + " /api/v1/image-workspace/models",
		http.MethodGet + " /api/v1/image-workspace/tasks/:taskID",
		http.MethodPost + " /api/v1/image-workspace/tasks/:taskID/cancel",
		http.MethodPost + " /api/v1/image-workspace/tasks/:taskID/retry",
		http.MethodGet + " /api/v1/image-workspace/artifacts/:artifactID/download",
		http.MethodGet + " /api/v1/image-workspace/templates",
		http.MethodPost + " /api/v1/image-workspace/templates",
		http.MethodGet + " /api/v1/image-workspace/usage-records",
		http.MethodGet + " /api/v1/image-workspace/worker/health",
		http.MethodGet + " /api/v1/image-workspace/worker/status",
		http.MethodPost + " /api/v1/image-workspace/worker/tasks/claim",
		http.MethodPost + " /api/v1/image-workspace/worker/tasks/:taskID/complete",
		http.MethodPost + " /api/v1/image-workspace/worker/tasks/:taskID/fail",
	}
	for _, path := range expected {
		require.True(t, routes[path], "expected route %s to be registered", path)
	}
}
