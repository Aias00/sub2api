//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/cloudbase/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRegistrationNotificationSendsDingTalkMarkdown(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NotEmpty(t, r.URL.Query().Get("timestamp"))
		require.NotEmpty(t, r.URL.Query().Get("sign"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := newMockSettingRepo()
	repo.data[SettingKeyRegistrationNotifyEnabled] = "true"
	repo.data[SettingKeyRegistrationNotifyProvider] = RegistrationNotifyProviderDingTalk
	repo.data[SettingKeyRegistrationNotifyWebhookURL] = server.URL
	repo.data[SettingKeyRegistrationNotifySecret] = "secret"
	repo.data[SettingKeySiteName] = "cloudbase"
	authService := &AuthService{settingService: NewSettingService(repo, &config.Config{})}

	err := authService.sendRegistrationNotification(context.Background(), &User{
		ID:          42,
		Email:       "new@example.com",
		Username:    "new-user",
		Balance:     10,
		Concurrency: 5,
	}, "email")

	require.NoError(t, err)
	require.Equal(t, "markdown", received["msgtype"])
	markdown, ok := received["markdown"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, markdown["text"], "new@example.com")
	require.Contains(t, markdown["text"], "cloudbase 新用户注册")
}

func TestRegistrationNotificationSendsFeishuTextWithSignature(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := newMockSettingRepo()
	repo.data[SettingKeyRegistrationNotifyEnabled] = "true"
	repo.data[SettingKeyRegistrationNotifyProvider] = RegistrationNotifyProviderFeishu
	repo.data[SettingKeyRegistrationNotifyWebhookURL] = server.URL
	repo.data[SettingKeyRegistrationNotifySecret] = "secret"
	authService := &AuthService{settingService: NewSettingService(repo, &config.Config{})}

	err := authService.sendRegistrationNotification(context.Background(), &User{
		ID:        7,
		Email:     "oauth@example.com",
		CreatedAt: time.Now(),
	}, "google")

	require.NoError(t, err)
	require.Equal(t, "text", received["msg_type"])
	require.NotEmpty(t, received["timestamp"])
	require.NotEmpty(t, received["sign"])
	content, ok := received["content"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, content["text"], "oauth@example.com")
	require.Contains(t, content["text"], "Google 登录")
}

func TestRegistrationNotificationDisabledSkipsWebhook(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := newMockSettingRepo()
	repo.data[SettingKeyRegistrationNotifyEnabled] = "false"
	repo.data[SettingKeyRegistrationNotifyProvider] = RegistrationNotifyProviderDingTalk
	repo.data[SettingKeyRegistrationNotifyWebhookURL] = server.URL
	authService := &AuthService{settingService: NewSettingService(repo, &config.Config{})}

	err := authService.sendRegistrationNotification(context.Background(), &User{ID: 1, Email: "new@example.com"}, "email")

	require.NoError(t, err)
	require.False(t, called)
}
