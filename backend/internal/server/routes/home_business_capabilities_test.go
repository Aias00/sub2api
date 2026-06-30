package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHomeBusinessCapabilityRoutesRegisterExpectedPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterHomeBusinessCapabilityRoutes(v1, &handler.Handlers{
		HomeBusiness: &handler.HomeBusinessCapabilityHandler{},
	})

	routes := map[string]bool{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	require.True(t, routes[http.MethodGet+" /api/v1/home/business-capabilities"])
}
