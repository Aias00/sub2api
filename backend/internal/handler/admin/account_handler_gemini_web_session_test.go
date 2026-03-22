package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiWebSessionAdminService struct {
	*stubAdminService
	account         service.Account
	lastUpdateInput *service.UpdateAccountInput
	updateCalls     int
}

func (s *geminiWebSessionAdminService) GetAccount(_ context.Context, id int64) (*service.Account, error) {
	if s.account.ID != id {
		return nil, errors.New("not found")
	}
	acc := s.account
	return &acc, nil
}

func (s *geminiWebSessionAdminService) UpdateAccount(_ context.Context, id int64, input *service.UpdateAccountInput) (*service.Account, error) {
	if s.account.ID != id {
		return nil, errors.New("not found")
	}
	s.lastUpdateInput = input
	s.updateCalls++
	if len(input.Credentials) > 0 {
		s.account.Credentials = input.Credentials
	}
	if len(input.Extra) > 0 {
		s.account.Extra = input.Extra
	}
	acc := s.account
	return &acc, nil
}

func setupGeminiWebSessionRouter(adminSvc service.AdminService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router.POST("/api/v1/admin/accounts/:id/gemini-web/start", handler.StartGeminiWebLogin)
	router.GET("/api/v1/admin/accounts/:id/gemini-web/status", handler.GetGeminiWebLoginStatus)
	router.POST("/api/v1/admin/accounts/:id/gemini-web/import-cookies", handler.ImportGeminiWebCookies)
	return router
}

func TestGeminiWebLoginStart_SetsSessionState(t *testing.T) {
	svc := &geminiWebSessionAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       88,
			Platform: service.PlatformGemini,
			Type:     service.AccountTypeGeminiWeb,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"base_url": "http://127.0.0.1:8000",
				"api_key":  "test-key",
			},
			Extra: map[string]any{},
		},
	}
	router := setupGeminiWebSessionRouter(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/88/gemini-web/start", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, svc.updateCalls)
	require.NotNil(t, svc.lastUpdateInput)
	require.Equal(t, "waiting_import", svc.lastUpdateInput.Extra[geminiWebSessionStatusKey])
	require.NotEmpty(t, svc.lastUpdateInput.Extra[geminiWebSessionLoginIDKey])
	require.Equal(t, "auto", svc.lastUpdateInput.Extra[geminiWebSessionModeKey])
}

func TestGeminiWebImportCookies_StoresCookiesAndReadyState(t *testing.T) {
	svc := &geminiWebSessionAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       89,
			Platform: service.PlatformGemini,
			Type:     service.AccountTypeGeminiWeb,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"api_key": "test-key",
			},
			Extra: map[string]any{},
		},
	}
	router := setupGeminiWebSessionRouter(svc)

	body, err := json.Marshal(map[string]any{
		"cookies_json": `[{"name":"SID","value":"abc"}]`,
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/89/gemini-web/import-cookies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, svc.updateCalls)
	require.NotNil(t, svc.lastUpdateInput)
	require.Equal(t, `[{"name":"SID","value":"abc"}]`, svc.lastUpdateInput.Credentials["cookies_json"])
	require.Equal(t, "ready", svc.lastUpdateInput.Extra[geminiWebSessionStatusKey])
}

func TestGeminiWebLoginStatus_ReturnsSessionState(t *testing.T) {
	svc := &geminiWebSessionAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       90,
			Platform: service.PlatformGemini,
			Type:     service.AccountTypeGeminiWeb,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"api_key":      "test-key",
				"cookies_json": "[{\"name\":\"SID\"}]",
			},
			Extra: map[string]any{
				geminiWebSessionStatusKey:    "ready",
				geminiWebSessionLoginIDKey:   "gw-123",
				geminiWebSessionMessageKey:   "Cookies imported",
				geminiWebSessionUpdatedAtKey: "2026-03-16T08:00:00Z",
			},
		},
	}
	router := setupGeminiWebSessionRouter(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/90/gemini-web/status", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "done", data["action"])
	require.Equal(t, "ready", data["status"])
	require.Equal(t, "gw-123", data["login_id"])
	require.Equal(t, true, data["has_cookies"])
	require.Equal(t, true, data["configured_key"])
}

func TestGeminiWebLoginStart_RejectsNonGeminiWebAccount(t *testing.T) {
	svc := &geminiWebSessionAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       91,
			Platform: service.PlatformGemini,
			Type:     service.AccountTypeOAuth,
			Status:   service.StatusActive,
		},
	}
	router := setupGeminiWebSessionRouter(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/91/gemini-web/start", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGeminiWebLoginStart_UsesGatewayStartEndpoint(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/auth/start", r.URL.Path)
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "remote", payload["login_mode"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"login_id":"gw-remote","status":"pending","message":"scan qr","login_url":"https://example.com/login","login_mode":"remote"}}`))
	}))
	defer gateway.Close()

	svc := &geminiWebSessionAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       92,
			Platform: service.PlatformGemini,
			Type:     service.AccountTypeGeminiWeb,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"base_url": gateway.URL,
				"api_key":  "test-key",
			},
			Extra: map[string]any{},
		},
	}
	router := setupGeminiWebSessionRouter(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/92/gemini-web/start", strings.NewReader(`{"login_mode":"remote"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "pending", svc.lastUpdateInput.Extra[geminiWebSessionStatusKey])
	require.Equal(t, "gw-remote", svc.lastUpdateInput.Extra[geminiWebSessionLoginIDKey])
	require.Equal(t, "https://example.com/login", svc.lastUpdateInput.Extra[geminiWebSessionLoginURLKey])
	require.Equal(t, "remote", svc.lastUpdateInput.Extra[geminiWebSessionModeKey])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	require.Equal(t, "open_login_url", data["action"])
	require.Equal(t, "remote", data["login_mode"])
}

func TestGeminiWebLoginStatus_UsesGatewayStatusEndpoint(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/auth/status", r.URL.Path)
		require.Equal(t, "gw-local", r.URL.Query().Get("login_id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ready","message":"ok","updated_at":"2026-03-16T12:00:00Z"}`))
	}))
	defer gateway.Close()

	svc := &geminiWebSessionAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       93,
			Platform: service.PlatformGemini,
			Type:     service.AccountTypeGeminiWeb,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"base_url": gateway.URL,
			},
			Extra: map[string]any{
				geminiWebSessionLoginIDKey: "gw-local",
			},
		},
	}
	router := setupGeminiWebSessionRouter(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/93/gemini-web/status", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, svc.lastUpdateInput)
	require.Equal(t, "ready", svc.lastUpdateInput.Extra[geminiWebSessionStatusKey])
}
