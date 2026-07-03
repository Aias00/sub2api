package routes

import (
	"net/http"
	"testing"

	"github.com/Aias00/cloudbase/internal/handler"
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

func TestRuntimeWorkerRoutesRegisterExpectedPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	admin := router.Group("/api/v1/admin")
	registerRuntimeWorkerRoutes(admin, &handler.Handlers{
		HomeBusiness: &handler.HomeBusinessCapabilityHandler{},
	})

	routes := map[string]bool{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	require.True(t, routes[http.MethodGet+" /api/v1/admin/runtime/workers"])
	require.True(t, routes[http.MethodPost+" /api/v1/admin/runtime/workers/:id/actions/:action"])
}
