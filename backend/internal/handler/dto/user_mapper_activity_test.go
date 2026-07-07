package dto

import (
	"testing"
	"time"

	"github.com/Aias00/cloudbase/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserFromServiceAdmin_MapsActivityTimestamps(t *testing.T) {
	t.Parallel()

	lastLoginAt := time.Date(2026, time.April, 20, 10, 0, 0, 0, time.UTC)
	lastActiveAt := lastLoginAt.Add(15 * time.Minute)
	lastUsedAt := lastLoginAt.Add(45 * time.Minute)

	out := UserFromServiceAdmin(&service.User{
		ID:                         42,
		Email:                      "admin@example.com",
		Username:                   "admin",
		Role:                       service.RoleAdmin,
		Status:                     service.StatusActive,
		LastActiveAt:               &lastActiveAt,
		LastUsedAt:                 &lastUsedAt,
		RegistrationIP:             "203.0.113.9",
		RegistrationUserAgent:      "Mozilla/5.0",
		RegistrationAcceptLanguage: "zh-CN,zh;q=0.9",
	})

	require.NotNil(t, out)
	require.NotNil(t, out.LastActiveAt)
	require.NotNil(t, out.LastUsedAt)
	require.WithinDuration(t, lastActiveAt, *out.LastActiveAt, time.Second)
	require.WithinDuration(t, lastUsedAt, *out.LastUsedAt, time.Second)
	require.Equal(t, "203.0.113.9", out.RegistrationIP)
	require.Equal(t, "Mozilla/5.0", out.RegistrationUserAgent)
	require.Equal(t, "zh-CN,zh;q=0.9", out.RegistrationAcceptLanguage)
}
