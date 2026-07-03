package routes

import (
	"net/http"
	"testing"

	"github.com/Aias00/cloudbase/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHotContentRoutesRegisterExpectedPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterHotContentRoutes(v1, &handler.Handlers{HotContent: &handler.HotContentHandler{}})

	routes := map[string]bool{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	require.True(t, routes[http.MethodGet+" /api/v1/hot/sources"])
	require.True(t, routes[http.MethodGet+" /api/v1/hot/items"])
	require.True(t, routes[http.MethodGet+" /api/v1/hot/run-events"])
	require.False(t, routes[http.MethodGet+" /api/v1/hot/feed-items"])
	require.False(t, routes[http.MethodGet+" /api/v1/hot/daily-issues"])
	require.False(t, routes[http.MethodGet+" /api/v1/hot/daily-issues/:issueDate"])
	require.False(t, routes[http.MethodGet+" /api/v1/hot/mp-entries"])
}
