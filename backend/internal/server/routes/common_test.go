package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCommonRoutesHealthIncludesEnvFromConfig(t *testing.T) {
	t.Setenv("HEALTH_ENV_MARKER", "")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterCommonRoutes(router, &config.Config{Log: config.LogConfig{Environment: "prod"}})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, "ok", body["status"])
	require.Equal(t, "prod", body["env"])
}

func TestCommonRoutesHealthMarkerEnvOverride(t *testing.T) {
	t.Setenv("HEALTH_ENV_MARKER", "staging-marker")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterCommonRoutes(router, &config.Config{Log: config.LogConfig{Environment: "prod"}})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, "ok", body["status"])
	require.Equal(t, "staging-marker", body["env"])
}
