//go:build unit

package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aias00/cloudbase/internal/config"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// updateSettingsCodexStatus PUT /settings 仅带给定字段，返回 HTTP 状态码（轻量 stub repo，无 DB）。
func updateSettingsCodexStatus(t *testing.T, body map[string]any) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{values: map[string]string{service.SettingKeyPromoCodeEnabled: "true"}}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	raw, err := json.Marshal(body)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateSettings(c)
	return rec.Code
}

// 白名单是双因子 AND：originator-only 条目在运行时永不命中（静默失效）。
// handler 应路由到 ValidateCodexWhitelistEntriesJSON，在写入时即拒（400）。
func TestUpdateSettings_CodexWhitelistRejectsUnmatchable(t *testing.T) {
	code := updateSettingsCodexStatus(t, map[string]any{
		"codex_cli_only_whitelist": `[{"originator":"opencode"}]`,
	})
	require.Equal(t, http.StatusBadRequest, code, "白名单 originator-only 应被拒(静默失效防护)")
}

func TestUpdateSettings_CodexWhitelistAcceptsMatchable(t *testing.T) {
	code := updateSettingsCodexStatus(t, map[string]any{
		"codex_cli_only_whitelist": `[{"originator":"opencode","ua_contains":["opencode/"]}]`,
	})
	require.Equal(t, http.StatusOK, code, "可命中白名单条目应通过")
}

// 黑名单是 OR 宽 deny：允许 originator-only。非对称——不受白名单收紧影响。
func TestUpdateSettings_CodexBlacklistAllowsOriginatorOnly(t *testing.T) {
	code := updateSettingsCodexStatus(t, map[string]any{
		"codex_cli_only_blacklist": `[{"originator":"evil"}]`,
	})
	require.Equal(t, http.StatusOK, code, "黑名单 originator-only 应允许(非对称)")
}

func TestUpdateSettings_PreservesOmittedCodexHardeningSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const fingerprintSignals = `[{"type":"header_prefix","match":["x-codex-"],"required":true}]`
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyPromoCodeEnabled:                     "true",
			service.SettingKeyMinCodexVersion:                      "0.141.0",
			service.SettingKeyMaxCodexVersion:                      "0.200.0",
			service.SettingKeyCodexCLIOnlyBlacklist:                `[{"originator":"evil"}]`,
			service.SettingKeyCodexCLIOnlyWhitelist:                `[{"originator":"opencode","ua_contains":["opencode/"]}]`,
			service.SettingKeyCodexCLIOnlyEngineFingerprintSignals: fingerprintSignals,
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	raw, err := json.Marshal(map[string]any{
		"promo_code_enabled": false,
	})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "0.141.0", repo.values[service.SettingKeyMinCodexVersion])
	require.Equal(t, "0.200.0", repo.values[service.SettingKeyMaxCodexVersion])
	require.Equal(t, `[{"originator":"evil"}]`, repo.values[service.SettingKeyCodexCLIOnlyBlacklist])
	require.Equal(t, `[{"originator":"opencode","ua_contains":["opencode/"]}]`, repo.values[service.SettingKeyCodexCLIOnlyWhitelist])
	require.Equal(t, fingerprintSignals, repo.values[service.SettingKeyCodexCLIOnlyEngineFingerprintSignals])
}

func TestUpdateSettings_AllowsExplicitCodexHardeningClear(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{
		values: map[string]string{
			service.SettingKeyPromoCodeEnabled:                     "true",
			service.SettingKeyMinCodexVersion:                      "0.141.0",
			service.SettingKeyMaxCodexVersion:                      "0.200.0",
			service.SettingKeyCodexCLIOnlyBlacklist:                `[{"originator":"evil"}]`,
			service.SettingKeyCodexCLIOnlyWhitelist:                `[{"originator":"opencode","ua_contains":["opencode/"]}]`,
			service.SettingKeyCodexCLIOnlyEngineFingerprintSignals: `[{"type":"header_prefix","match":["x-codex-"],"required":true}]`,
		},
	}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	raw, err := json.Marshal(map[string]any{
		"min_codex_version":                         "",
		"max_codex_version":                         "",
		"codex_cli_only_blacklist":                  "",
		"codex_cli_only_whitelist":                  "",
		"codex_cli_only_engine_fingerprint_signals": "",
	})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, repo.values[service.SettingKeyMinCodexVersion])
	require.Empty(t, repo.values[service.SettingKeyMaxCodexVersion])
	require.Empty(t, repo.values[service.SettingKeyCodexCLIOnlyBlacklist])
	require.Empty(t, repo.values[service.SettingKeyCodexCLIOnlyWhitelist])
	require.Empty(t, repo.values[service.SettingKeyCodexCLIOnlyEngineFingerprintSignals])
}
