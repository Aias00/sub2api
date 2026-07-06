package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/pquerna/otp/totp"
)

const DefaultTotpIssuer = "Cloudbase"

func ResolveTotpIssuer(frontendURL, siteName string) string {
	if parsedURL, err := url.Parse(strings.TrimSpace(frontendURL)); err == nil {
		if host := strings.TrimSpace(parsedURL.Hostname()); host != "" {
			return host
		}
	}
	if siteName := strings.TrimSpace(siteName); siteName != "" {
		return siteName
	}
	return DefaultTotpIssuer
}

func ConstantTimeEqual(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func TotpSecretPrefix(secret string) string {
	if len(secret) >= 4 {
		return secret[:4]
	}
	return "N/A"
}

func ValidateTotpCode(code, secret string) bool {
	return totp.Validate(code, secret)
}

func GenerateRandomHexToken(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func MaskEmail(email string) string {
	if len(email) < 3 {
		return "***"
	}

	atIdx := -1
	for i, c := range email {
		if c == '@' {
			atIdx = i
			break
		}
	}

	if atIdx == -1 || atIdx < 1 {
		return email[:1] + "***"
	}

	localPart := email[:atIdx]
	domain := email[atIdx:]
	if len(localPart) <= 2 {
		return localPart[:1] + "***" + domain
	}
	return localPart[:1] + "***" + localPart[len(localPart)-1:] + domain
}

func TotpVerificationMethod(emailVerifyEnabled bool) string {
	if emailVerifyEnabled {
		return "email"
	}
	return "password"
}
