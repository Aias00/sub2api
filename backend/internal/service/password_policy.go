package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/cloudbase/internal/pkg/errors"
)

const (
	DefaultPasswordMinLength = 8
	MaxPasswordMinLength     = 128
)

func normalizePasswordMinLength(value int) int {
	if value < DefaultPasswordMinLength {
		return DefaultPasswordMinLength
	}
	if value > MaxPasswordMinLength {
		return MaxPasswordMinLength
	}
	return value
}

func NormalizePasswordMinLengthForSetting(value int) int {
	return normalizePasswordMinLength(value)
}

func parsePasswordMinLength(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return DefaultPasswordMinLength
	}
	return normalizePasswordMinLength(value)
}

func passwordTooShortError(minLength int) error {
	minLength = normalizePasswordMinLength(minLength)
	return infraerrors.BadRequest(
		"PASSWORD_TOO_SHORT",
		fmt.Sprintf("password must be at least %d characters", minLength),
	)
}

func validatePasswordMinLength(password string, minLength int) error {
	if len(password) < normalizePasswordMinLength(minLength) {
		return passwordTooShortError(minLength)
	}
	return nil
}

func (s *SettingService) GetPasswordMinLength(ctx context.Context) int {
	return passwordMinLengthFromRepository(ctx, func() SettingRepository {
		if s == nil {
			return nil
		}
		return s.settingRepo
	}())
}

func passwordMinLengthFromRepository(ctx context.Context, repo SettingRepository) int {
	if repo == nil {
		return DefaultPasswordMinLength
	}
	value, err := repo.GetValue(ctx, SettingKeyPasswordMinLength)
	if err != nil {
		return DefaultPasswordMinLength
	}
	return parsePasswordMinLength(value)
}
