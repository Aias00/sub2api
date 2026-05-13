package service

import (
	"net/url"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	RegistrationNotifyProviderDingTalk = "dingtalk"
	RegistrationNotifyProviderFeishu   = "feishu"
)

func normalizeRegistrationNotifyProvider(value string) (string, error) {
	provider := strings.ToLower(strings.TrimSpace(value))
	switch provider {
	case "", "none", "disabled":
		return "", nil
	case RegistrationNotifyProviderDingTalk:
		return RegistrationNotifyProviderDingTalk, nil
	case RegistrationNotifyProviderFeishu, "lark":
		return RegistrationNotifyProviderFeishu, nil
	default:
		return "", infraerrors.BadRequest("INVALID_REGISTRATION_NOTIFY_PROVIDER", "registration notification provider must be dingtalk or feishu")
	}
}

func normalizeStoredRegistrationNotifyProvider(value string) string {
	provider, err := normalizeRegistrationNotifyProvider(value)
	if err != nil {
		return ""
	}
	return provider
}

func normalizeRegistrationNotifyWebhookURL(value string, enabled bool) (string, error) {
	webhookURL := strings.TrimSpace(value)
	if webhookURL == "" {
		if enabled {
			return "", infraerrors.BadRequest("REGISTRATION_NOTIFY_WEBHOOK_REQUIRED", "registration notification webhook url is required")
		}
		return "", nil
	}

	parsed, err := url.Parse(webhookURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", infraerrors.BadRequest("INVALID_REGISTRATION_NOTIFY_WEBHOOK_URL", "registration notification webhook url must be http or https")
	}
	return webhookURL, nil
}
