package routes

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWebRoutesExposeOnlyGenericPrimaryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	handlers := &handler.Handlers{Web: &handler.WebHandler{}}

	RegisterWebRoutes(v1, handlers)

	registered := map[string]bool{}
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, path := range []string{
		"/api/v1/web/auth/oauth/:provider/start",
		"/api/v1/web/auth/login",
		"/api/v1/web/auth/register",
		"/api/v1/web/auth/oauth/session",
		"/api/v1/web/auth/refresh",
		"/api/v1/web/auth/logout",
		"/api/v1/web/auth/me",
		"/api/v1/web/auth/credits",
		"/api/v1/web/payments/checkout",
		"/api/v1/web/prompts/import-twitter",
	} {
		method := "POST"
		if path == "/api/v1/web/auth/oauth/:provider/start" || path == "/api/v1/web/auth/me" || path == "/api/v1/web/auth/credits" {
			method = "GET"
		}
		require.True(t, registered[method+" "+path], "missing route %s %s", method, path)
	}

	for _, path := range []string{
		"/api/v1/touch/web/auth/login",
		"/api/v1/touch/web/payments/checkout",
		"/api/v1/touch/web/prompts/import-twitter",
	} {
		require.False(t, registered["POST "+path], "legacy touch web alias should not be registered: %s", path)
	}
}
