// Package admin provides HTTP handlers for administrative operations.
package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Aias00/cloudbase/internal/pkg/response"
	"github.com/Aias00/cloudbase/internal/service"

	"github.com/gin-gonic/gin"
)

// StartGeminiWebLogin 启动 Gemini Web 登录会话。
// POST /api/v1/admin/accounts/:id/gemini-web/start
func (h *AccountHandler) StartGeminiWebLogin(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	var req StartGeminiWebLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.BadRequestWithError(c, err)
		return
	}
	loginMode := normalizeGeminiWebLoginMode(req.LoginMode)

	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if account == nil {
		response.BadRequest(c, "Account not found")
		return
	}
	if account.Type != service.AccountTypeGeminiWeb {
		response.BadRequest(c, "Account type must be gemini-web")
		return
	}

	extra := cloneAnyMap(account.Extra)
	loginID := strings.TrimSpace(account.GetExtraString(geminiWebSessionLoginIDKey))
	if loginID == "" {
		loginID = fmt.Sprintf("gw-%d", time.Now().UnixNano())
	}
	extra[geminiWebSessionLoginIDKey] = loginID
	extra[geminiWebSessionStatusKey] = "waiting_import"
	extra[geminiWebSessionMessageKey] = "Login session started, please import cookies"
	extra[geminiWebSessionUpdatedAtKey] = time.Now().UTC().Format(time.RFC3339)
	extra[geminiWebSessionModeKey] = loginMode
	extra[geminiWebSessionSourceKey] = "cloudbase"

	if gatewayData, gatewayErr := h.tryGeminiWebGatewayStart(c.Request.Context(), account, loginID, loginMode); gatewayErr == nil {
		mergeGeminiWebGatewaySession(extra, gatewayData)
		extra[geminiWebSessionSourceKey] = "gateway"
	}

	if _, err := h.adminService.UpdateAccount(c.Request.Context(), accountID, &service.UpdateAccountInput{Extra: extra}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"account_id":         accountID,
		"login_id":           loginID,
		"status":             extra[geminiWebSessionStatusKey],
		"message":            extra[geminiWebSessionMessageKey],
		"updated_at":         extra[geminiWebSessionUpdatedAtKey],
		"action":             geminiWebNextAction(toString(extra[geminiWebSessionStatusKey]), strings.TrimSpace(toString(extra[geminiWebSessionLoginURLKey])), strings.TrimSpace(account.GetCredential("cookies_json")) != ""),
		"login_url":          extra[geminiWebSessionLoginURLKey],
		"expires_at":         extra[geminiWebSessionExpiresAtKey],
		"login_mode":         extra[geminiWebSessionModeKey],
		"gateway_url":        account.GetGeminiWebBaseURL(),
		"local_auth_enabled": toBool(extra[geminiWebLocalAuthEnabledKey]),
	})
}

// GetGeminiWebLoginStatus 获取 Gemini Web 登录会话状态。
// GET /api/v1/admin/accounts/:id/gemini-web/status
func (h *AccountHandler) GetGeminiWebLoginStatus(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if account == nil {
		response.BadRequest(c, "Account not found")
		return
	}
	if account.Type != service.AccountTypeGeminiWeb {
		response.BadRequest(c, "Account type must be gemini-web")
		return
	}

	status := strings.TrimSpace(account.GetExtraString(geminiWebSessionStatusKey))
	if status == "" {
		status = "idle"
	}
	loginID := strings.TrimSpace(account.GetExtraString(geminiWebSessionLoginIDKey))
	message := strings.TrimSpace(account.GetExtraString(geminiWebSessionMessageKey))
	updatedAt := strings.TrimSpace(account.GetExtraString(geminiWebSessionUpdatedAtKey))
	loginURL := strings.TrimSpace(account.GetExtraString(geminiWebSessionLoginURLKey))
	expiresAt := strings.TrimSpace(account.GetExtraString(geminiWebSessionExpiresAtKey))
	loginMode := normalizeGeminiWebLoginMode(account.GetExtraString(geminiWebSessionModeKey))
	localAuthEnabled := toBool(account.Extra[geminiWebLocalAuthEnabledKey])

	extra := cloneAnyMap(account.Extra)
	if loginID != "" {
		if gatewayData, gatewayErr := h.tryGeminiWebGatewayStatus(c.Request.Context(), account, loginID); gatewayErr == nil {
			mergeGeminiWebGatewaySession(extra, gatewayData)
			extra[geminiWebSessionSourceKey] = "gateway"
			if _, err := h.adminService.UpdateAccount(c.Request.Context(), accountID, &service.UpdateAccountInput{Extra: extra}); err == nil {
				status = strings.TrimSpace(toString(extra[geminiWebSessionStatusKey]))
				message = strings.TrimSpace(toString(extra[geminiWebSessionMessageKey]))
				updatedAt = strings.TrimSpace(toString(extra[geminiWebSessionUpdatedAtKey]))
				loginURL = strings.TrimSpace(toString(extra[geminiWebSessionLoginURLKey]))
				expiresAt = strings.TrimSpace(toString(extra[geminiWebSessionExpiresAtKey]))
				loginMode = normalizeGeminiWebLoginMode(toString(extra[geminiWebSessionModeKey]))
				localAuthEnabled = toBool(extra[geminiWebLocalAuthEnabledKey])
			}
		}
	}

	response.Success(c, gin.H{
		"account_id":         accountID,
		"login_id":           loginID,
		"status":             status,
		"message":            message,
		"updated_at":         updatedAt,
		"action":             geminiWebNextAction(status, loginURL, strings.TrimSpace(account.GetCredential("cookies_json")) != ""),
		"login_url":          loginURL,
		"expires_at":         expiresAt,
		"login_mode":         loginMode,
		"has_cookies":        strings.TrimSpace(account.GetCredential("cookies_json")) != "",
		"gateway_url":        account.GetGeminiWebBaseURL(),
		"configured_key":     account.GetGeminiWebAPIKey() != "",
		"local_auth_enabled": localAuthEnabled,
	})
}

// ImportGeminiWebCookies 导入 Gemini Web cookies 并将会话状态置为 ready。
// POST /api/v1/admin/accounts/:id/gemini-web/import-cookies
func (h *AccountHandler) ImportGeminiWebCookies(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return
	}

	var req ImportGeminiWebCookiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestWithError(c, err)
		return
	}

	cookiesJSON := strings.TrimSpace(req.CookiesJSON)
	if cookiesJSON == "" && req.Cookies != nil {
		raw, err := json.Marshal(req.Cookies)
		if err != nil {
			response.BadRequest(c, "Invalid cookies format")
			return
		}
		cookiesJSON = string(raw)
	}
	if cookiesJSON == "" {
		response.BadRequest(c, "cookies_json or cookies is required")
		return
	}

	account, err := h.adminService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if account == nil {
		response.BadRequest(c, "Account not found")
		return
	}
	if account.Type != service.AccountTypeGeminiWeb {
		response.BadRequest(c, "Account type must be gemini-web")
		return
	}

	credentials := cloneAnyMap(account.Credentials)
	credentials["cookies_json"] = cookiesJSON

	extra := cloneAnyMap(account.Extra)
	if strings.TrimSpace(account.GetExtraString(geminiWebSessionLoginIDKey)) == "" {
		extra[geminiWebSessionLoginIDKey] = fmt.Sprintf("gw-%d", time.Now().UnixNano())
	}
	extra[geminiWebSessionModeKey] = normalizeGeminiWebLoginMode(account.GetExtraString(geminiWebSessionModeKey))
	extra[geminiWebSessionStatusKey] = "ready"
	extra[geminiWebSessionMessageKey] = "Cookies imported"
	extra[geminiWebSessionUpdatedAtKey] = time.Now().UTC().Format(time.RFC3339)
	extra[geminiWebSessionSourceKey] = "cloudbase"

	if gatewayData, gatewayErr := h.tryGeminiWebGatewayImportCookies(c.Request.Context(), account, toString(extra[geminiWebSessionLoginIDKey]), cookiesJSON); gatewayErr == nil {
		mergeGeminiWebGatewaySession(extra, gatewayData)
		extra[geminiWebSessionSourceKey] = "gateway"
	}

	if _, err := h.adminService.UpdateAccount(c.Request.Context(), accountID, &service.UpdateAccountInput{
		Credentials: credentials,
		Extra:       extra,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"account_id":         accountID,
		"login_id":           extra[geminiWebSessionLoginIDKey],
		"status":             extra[geminiWebSessionStatusKey],
		"message":            extra[geminiWebSessionMessageKey],
		"updated_at":         extra[geminiWebSessionUpdatedAtKey],
		"action":             geminiWebNextAction(toString(extra[geminiWebSessionStatusKey]), strings.TrimSpace(toString(extra[geminiWebSessionLoginURLKey])), true),
		"login_url":          extra[geminiWebSessionLoginURLKey],
		"expires_at":         extra[geminiWebSessionExpiresAtKey],
		"login_mode":         extra[geminiWebSessionModeKey],
		"local_auth_enabled": toBool(extra[geminiWebLocalAuthEnabledKey]),
	})
}

func geminiWebNextAction(status, loginURL string, hasCookies bool) string {
	if hasCookies || strings.EqualFold(strings.TrimSpace(status), "ready") {
		return "done"
	}
	if strings.TrimSpace(loginURL) != "" {
		return "open_login_url"
	}
	return "import_cookies"
}

func (h *AccountHandler) tryGeminiWebGatewayStart(ctx context.Context, account *service.Account, loginID, loginMode string) (map[string]any, error) {
	body := map[string]any{
		"login_id":   loginID,
		"account_id": account.ID,
		"login_mode": normalizeGeminiWebLoginMode(loginMode),
	}
	for _, endpoint := range []string{"/auth/start", "/session/start", "/login/start"} {
		if data, err := callGeminiWebGatewayJSON(ctx, http.MethodPost, endpoint, account, body); err == nil {
			return data, nil
		}
	}
	return nil, errors.New("no start endpoint available")
}

func (h *AccountHandler) tryGeminiWebGatewayStatus(ctx context.Context, account *service.Account, loginID string) (map[string]any, error) {
	for _, endpoint := range []string{"/auth/status", "/session/status", "/login/status"} {
		query := endpoint + "?login_id=" + url.QueryEscape(loginID)
		if data, err := callGeminiWebGatewayJSON(ctx, http.MethodGet, query, account, nil); err == nil {
			return data, nil
		}
	}
	return nil, errors.New("no status endpoint available")
}

func (h *AccountHandler) tryGeminiWebGatewayImportCookies(ctx context.Context, account *service.Account, loginID, cookiesJSON string) (map[string]any, error) {
	body := map[string]any{"login_id": loginID, "cookies_json": cookiesJSON}
	for _, endpoint := range []string{"/auth/import-cookies", "/session/import-cookies", "/import-cookies"} {
		if data, err := callGeminiWebGatewayJSON(ctx, http.MethodPost, endpoint, account, body); err == nil {
			return data, nil
		}
	}
	return nil, errors.New("no import endpoint available")
}

func callGeminiWebGatewayJSON(ctx context.Context, method, endpoint string, account *service.Account, payload any) (map[string]any, error) {
	baseURL := strings.TrimRight(account.GetGeminiWebBaseURL(), "/")
	if baseURL == "" {
		return nil, errors.New("empty gateway base url")
	}
	fullURL := baseURL + endpoint

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey := strings.TrimSpace(account.GetGeminiWebAPIKey()); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gateway status %d", resp.StatusCode)
	}

	if len(rawBody) == 0 {
		return map[string]any{}, nil
	}

	var payloadMap map[string]any
	if err := json.Unmarshal(rawBody, &payloadMap); err != nil {
		return map[string]any{"raw": string(rawBody)}, nil
	}
	if data, ok := payloadMap["data"].(map[string]any); ok {
		return data, nil
	}
	return payloadMap, nil
}

func mergeGeminiWebGatewaySession(extra map[string]any, data map[string]any) {
	if extra == nil || len(data) == 0 {
		return
	}

	if loginID := firstNonEmptyString(data, "login_id", "session_id", "id"); loginID != "" {
		extra[geminiWebSessionLoginIDKey] = loginID
	}
	if status := firstNonEmptyString(data, "status", "state"); status != "" {
		extra[geminiWebSessionStatusKey] = status
	}
	if message := firstNonEmptyString(data, "message", "detail"); message != "" {
		extra[geminiWebSessionMessageKey] = message
	}
	if loginURL := firstNonEmptyString(data, "login_url", "url", "auth_url"); loginURL != "" {
		extra[geminiWebSessionLoginURLKey] = loginURL
	}
	if expiresAt := firstNonEmptyString(data, "expires_at", "expire_at"); expiresAt != "" {
		extra[geminiWebSessionExpiresAtKey] = expiresAt
	}
	if localAuthEnabled, ok := data["local_auth_enabled"]; ok {
		extra[geminiWebLocalAuthEnabledKey] = toBool(localAuthEnabled)
	}
	if loginMode := normalizeGeminiWebLoginMode(firstNonEmptyString(data, "login_mode", "mode")); loginMode != "" {
		extra[geminiWebSessionModeKey] = loginMode
	}
	if updatedAt := firstNonEmptyString(data, "updated_at"); updatedAt != "" {
		extra[geminiWebSessionUpdatedAtKey] = updatedAt
	} else {
		extra[geminiWebSessionUpdatedAtKey] = time.Now().UTC().Format(time.RFC3339)
	}
}

func normalizeGeminiWebLoginMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return "auto"
	case "remote":
		return "remote"
	case "local":
		return "local"
	default:
		return "auto"
	}
}

func firstNonEmptyString(data map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(toString(data[key])); value != "" {
			return value
		}
	}
	return ""
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(t, 10)
	case int:
		return strconv.Itoa(t)
	default:
		return ""
	}
}

func toBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(t))
		return err == nil && parsed
	case json.Number:
		return t.String() == "1"
	case float64:
		return t != 0
	case int64:
		return t != 0
	case int:
		return t != 0
	default:
		return false
	}
}

func cloneAnyMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
