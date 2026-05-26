package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	redeemCodeLength     = 8
	redeemCodeAlphabet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	redeemCodeRetryLimit = 12
)

type RedeemCode struct {
	ID        int64
	Code      string
	Type      string
	Value     float64
	Status    string
	UsedBy    *int64
	UsedAt    *time.Time
	Notes     string
	CreatedAt time.Time

	GroupID      *int64
	ValidityDays int

	User  *User
	Group *Group
}

func (r *RedeemCode) IsUsed() bool {
	return r.Status == StatusUsed
}

func (r *RedeemCode) CanUse() bool {
	return r.Status == StatusUnused
}

func GenerateRedeemCode() (string, error) {
	buf := make([]byte, redeemCodeLength)
	alphabetLen := byte(len(redeemCodeAlphabet))
	limit := byte(256/int(alphabetLen)) * alphabetLen
	i := 0
	for i < redeemCodeLength {
		random := make([]byte, redeemCodeLength)
		if _, err := rand.Read(random); err != nil {
			return "", fmt.Errorf("generate random bytes: %w", err)
		}
		for _, b := range random {
			if b >= limit {
				continue
			}
			buf[i] = redeemCodeAlphabet[b%alphabetLen]
			i++
			if i == redeemCodeLength {
				break
			}
		}
	}
	return string(buf), nil
}

func canonicalRedeemCodeInput(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func generateUniqueRedeemCode(ctx context.Context, repo RedeemCodeRepository) (string, error) {
	if repo == nil {
		return "", fmt.Errorf("redeem code repository is required")
	}
	for attempt := 0; attempt < redeemCodeRetryLimit; attempt++ {
		code, err := GenerateRedeemCode()
		if err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}
		_, err = repo.GetByCode(ctx, code)
		if err == nil {
			continue
		}
		if errors.Is(err, ErrRedeemCodeNotFound) {
			return code, nil
		}
		return "", fmt.Errorf("check redeem code uniqueness: %w", err)
	}
	return "", fmt.Errorf("generate unique redeem code: exhausted %d attempts", redeemCodeRetryLimit)
}
