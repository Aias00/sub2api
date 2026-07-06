package gateway

import (
	"time"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
)

const (
	StatusAPIKeyActive         = "active"
	StatusAPIKeyDisabled       = "disabled"
	StatusAPIKeyQuotaExhausted = "quota_exhausted"
	StatusAPIKeyExpired        = "expired"
)

const (
	RateLimitWindow5h = 5 * time.Hour
	RateLimitWindow1d = 24 * time.Hour
	RateLimitWindow7d = 7 * 24 * time.Hour
)

var (
	ErrAPIKeyNotFound            = infraerrors.NotFound("API_KEY_NOT_FOUND", "api key not found")
	ErrAPIKeyExists              = infraerrors.Conflict("API_KEY_EXISTS", "api key already exists")
	ErrAPIKeyTooShort            = infraerrors.BadRequest("API_KEY_TOO_SHORT", "api key must be at least 16 characters")
	ErrAPIKeyInvalidChars        = infraerrors.BadRequest("API_KEY_INVALID_CHARS", "api key can only contain letters, numbers, underscores, and hyphens")
	ErrAPIKeyRateLimited         = infraerrors.TooManyRequests("API_KEY_RATE_LIMITED", "too many failed attempts, please try again later")
	ErrAPIKeyExpired             = infraerrors.Forbidden("API_KEY_EXPIRED", "api key 已过期")
	ErrAPIKeyQuotaExhausted      = infraerrors.TooManyRequests("API_KEY_QUOTA_EXHAUSTED", "api key 额度已用完")
	ErrAPIKeyRateLimit5hExceeded = infraerrors.TooManyRequests("API_KEY_RATE_5H_EXCEEDED", "api key 5小时限额已用完")
	ErrAPIKeyRateLimit1dExceeded = infraerrors.TooManyRequests("API_KEY_RATE_1D_EXCEEDED", "api key 日限额已用完")
	ErrAPIKeyRateLimit7dExceeded = infraerrors.TooManyRequests("API_KEY_RATE_7D_EXCEEDED", "api key 7天限额已用完")
)

func IsWindowExpired(windowStart *time.Time, duration time.Duration) bool {
	return windowStart == nil || time.Since(*windowStart) >= duration
}

type APIKeyListFilters struct {
	Search  string
	Status  string
	GroupID *int64
}

type APIKeyRateLimitData struct {
	Usage5h       float64
	Usage1d       float64
	Usage7d       float64
	Window5hStart *time.Time
	Window1dStart *time.Time
	Window7dStart *time.Time
}

func (d *APIKeyRateLimitData) EffectiveUsage5h() float64 {
	if IsWindowExpired(d.Window5hStart, RateLimitWindow5h) {
		return 0
	}
	return d.Usage5h
}

func (d *APIKeyRateLimitData) EffectiveUsage1d() float64 {
	if IsWindowExpired(d.Window1dStart, RateLimitWindow1d) {
		return 0
	}
	return d.Usage1d
}

func (d *APIKeyRateLimitData) EffectiveUsage7d() float64 {
	if IsWindowExpired(d.Window7dStart, RateLimitWindow7d) {
		return 0
	}
	return d.Usage7d
}

type APIKeyQuotaUsageState struct {
	QuotaUsed float64
	Quota     float64
	Key       string
	Status    string
}

func IsAPIKeyActive(status string) bool {
	return status == StatusAPIKeyActive
}

func IsAPIKeyExpired(expiresAt *time.Time, now time.Time) bool {
	return expiresAt != nil && now.After(*expiresAt)
}

func IsAPIKeyQuotaExhausted(quota, quotaUsed float64) bool {
	return quota > 0 && quotaUsed >= quota
}

func APIKeyQuotaRemaining(quota, quotaUsed float64) float64 {
	if quota <= 0 {
		return -1
	}
	remaining := quota - quotaUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func APIKeyDaysUntilExpiry(expiresAt *time.Time, now time.Time) int {
	if expiresAt == nil {
		return -1
	}
	duration := expiresAt.Sub(now)
	if duration < 0 {
		return 0
	}
	return int(duration.Hours() / 24)
}
