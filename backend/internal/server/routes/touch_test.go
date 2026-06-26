package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestLegacyTouchAPIRoutesAreNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Group("/api/v1")

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/touch/capabilities"},
		{method: http.MethodGet, path: "/api/v1/touch/prompts/cases"},
		{method: http.MethodPost, path: "/api/v1/touch/web/auth/login"},
		{method: http.MethodPost, path: "/api/v1/touch/web/payments/checkout"},
		{method: http.MethodPost, path: "/api/v1/touch/admin/auth/check"},
		{method: http.MethodPost, path: "/api/v1/touch/admin/users/sync"},
		{method: http.MethodPost, path: "/api/v1/touch/admin/payments/checkout"},
		{method: http.MethodPost, path: "/api/v1/touch/admin/subscriptions/list"},
		{method: http.MethodPost, path: "/api/v1/touch/admin/prompts/cases"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code, tc.path)
	}
}
