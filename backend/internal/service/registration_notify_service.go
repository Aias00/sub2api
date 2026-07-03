package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Aias00/cloudbase/internal/pkg/logger"
)

const registrationNotifyTimeout = 5 * time.Second

type registrationNotifyConfig struct {
	Enabled    bool
	Provider   string
	WebhookURL string
	Secret     string
	SiteName   string
}

func (s *AuthService) notifyUserRegistered(ctx context.Context, user *User, signupSource string) {
	if s == nil || user == nil {
		return
	}
	snapshot := *user
	go func() {
		notifyCtx, cancel := context.WithTimeout(context.Background(), registrationNotifyTimeout)
		defer cancel()
		if err := s.sendRegistrationNotification(notifyCtx, &snapshot, signupSource); err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] registration notification failed for user %d: %v", snapshot.ID, err)
		}
	}()
}

func (s *AuthService) sendRegistrationNotification(ctx context.Context, user *User, signupSource string) error {
	if s == nil || user == nil || user.ID <= 0 {
		return nil
	}
	cfg, err := s.loadRegistrationNotifyConfig(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}

	webhookURL, payload, err := buildRegistrationNotifyRequest(cfg, user, signupSource, time.Now())
	if err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal registration notification payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build registration notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send registration notification: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("registration notification webhook returned %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

func (s *AuthService) loadRegistrationNotifyConfig(ctx context.Context) (registrationNotifyConfig, error) {
	if s == nil || s.settingService == nil || s.settingService.settingRepo == nil {
		return registrationNotifyConfig{}, nil
	}
	values, err := s.settingService.settingRepo.GetMultiple(ctx, []string{
		SettingKeyRegistrationNotifyEnabled,
		SettingKeyRegistrationNotifyProvider,
		SettingKeyRegistrationNotifyWebhookURL,
		SettingKeyRegistrationNotifySecret,
		SettingKeySiteName,
	})
	if err != nil {
		return registrationNotifyConfig{}, fmt.Errorf("load registration notification settings: %w", err)
	}

	siteName := strings.TrimSpace(values[SettingKeySiteName])
	if siteName == "" {
		siteName = "Cloudbase"
	}
	enabled := values[SettingKeyRegistrationNotifyEnabled] == "true"
	if !enabled {
		return registrationNotifyConfig{Enabled: false, SiteName: siteName}, nil
	}
	provider, err := normalizeRegistrationNotifyProvider(values[SettingKeyRegistrationNotifyProvider])
	if err != nil {
		return registrationNotifyConfig{}, err
	}
	webhookURL, err := normalizeRegistrationNotifyWebhookURL(values[SettingKeyRegistrationNotifyWebhookURL], true)
	if err != nil {
		return registrationNotifyConfig{}, err
	}
	return registrationNotifyConfig{
		Enabled:    true,
		Provider:   provider,
		WebhookURL: webhookURL,
		Secret:     strings.TrimSpace(values[SettingKeyRegistrationNotifySecret]),
		SiteName:   siteName,
	}, nil
}

func buildRegistrationNotifyRequest(cfg registrationNotifyConfig, user *User, signupSource string, now time.Time) (string, map[string]any, error) {
	if cfg.Provider == "" {
		return "", nil, fmt.Errorf("registration notification provider is required")
	}
	if cfg.WebhookURL == "" {
		return "", nil, fmt.Errorf("registration notification webhook url is required")
	}

	text := buildRegistrationNotifyText(cfg.SiteName, user, signupSource, now)
	switch cfg.Provider {
	case RegistrationNotifyProviderDingTalk:
		webhookURL := cfg.WebhookURL
		if cfg.Secret != "" {
			webhookURL = signDingTalkWebhookURL(webhookURL, cfg.Secret, now)
		}
		return webhookURL, map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": "新用户注册",
				"text":  text,
			},
		}, nil
	case RegistrationNotifyProviderFeishu:
		payload := map[string]any{
			"msg_type": "text",
			"content": map[string]string{
				"text": text,
			},
		}
		if cfg.Secret != "" {
			timestamp := strconv.FormatInt(now.Unix(), 10)
			payload["timestamp"] = timestamp
			payload["sign"] = signFeishuWebhook(timestamp, cfg.Secret)
		}
		return cfg.WebhookURL, payload, nil
	default:
		return "", nil, fmt.Errorf("unsupported registration notification provider: %s", cfg.Provider)
	}
}

func buildRegistrationNotifyText(siteName string, user *User, signupSource string, now time.Time) string {
	if strings.TrimSpace(siteName) == "" {
		siteName = "Cloudbase"
	}
	username := strings.TrimSpace(user.Username)
	if username == "" {
		username = "-"
	}
	source := registrationNotifySourceLabel(signupSource)
	createdAt := now.In(time.Local).Format("2006-01-02 15:04:05 MST")
	return fmt.Sprintf(
		"### %s 新用户注册\n\n- 用户 ID：%d\n- 邮箱：%s\n- 用户名：%s\n- 注册来源：%s\n- 初始余额：$%.2f\n- 并发额度：%d\n- 时间：%s",
		siteName,
		user.ID,
		user.Email,
		username,
		source,
		user.Balance,
		user.Concurrency,
		createdAt,
	)
}

func registrationNotifySourceLabel(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "email", "":
		return "邮箱注册"
	case "google":
		return "Google 登录"
	case "github":
		return "GitHub 登录"
	case "linuxdo":
		return "LinuxDo 登录"
	case "oidc":
		return "OIDC 登录"
	case "wechat":
		return "微信登录"
	default:
		return source
	}
}

func signDingTalkWebhookURL(webhookURL string, secret string, now time.Time) string {
	parsed, err := url.Parse(webhookURL)
	if err != nil {
		return webhookURL
	}
	timestamp := strconv.FormatInt(now.UnixMilli(), 10)
	stringToSign := timestamp + "\n" + secret
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	query := parsed.Query()
	query.Set("timestamp", timestamp)
	query.Set("sign", signature)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func signFeishuWebhook(timestamp string, secret string) string {
	stringToSign := timestamp + "\n" + secret
	mac := hmac.New(sha256.New, []byte(stringToSign))
	_, _ = mac.Write([]byte{})
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
