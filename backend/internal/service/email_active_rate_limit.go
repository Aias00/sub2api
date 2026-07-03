package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
)

func activeEmailScopeForUserID(userID int64) string {
	if userID <= 0 {
		return ""
	}
	return fmt.Sprintf("user:%d", userID)
}

func activeEmailScopeForEmail(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return ""
	}
	return "email:" + normalized
}

func reserveActiveEmailDailyQuota(ctx context.Context, cache EmailCache, scope string) error {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return nil
	}
	if cache == nil {
		return ErrServiceUnavailable
	}

	count, err := cache.IncrActiveEmailDailyRate(ctx, scope, durationUntilNextLocalMidnight(time.Now()))
	if err != nil {
		return infraerrors.ServiceUnavailable("EMAIL_RATE_LIMIT_UNAVAILABLE", "email rate limit is temporarily unavailable").WithCause(err)
	}
	if count > activeEmailDailyLimit {
		return ErrActiveEmailDailyLimit
	}
	return nil
}

func durationUntilNextLocalMidnight(now time.Time) time.Duration {
	if now.IsZero() {
		now = time.Now()
	}
	localNow := now.In(time.Local)
	nextMidnight := time.Date(localNow.Year(), localNow.Month(), localNow.Day()+1, 0, 0, 0, 0, localNow.Location())
	ttl := nextMidnight.Sub(localNow)
	if ttl <= 0 {
		return 24 * time.Hour
	}
	return ttl
}
