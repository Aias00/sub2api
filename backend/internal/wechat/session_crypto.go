package wechat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type CookiePayload struct {
	CookieHeader string `json:"cookie_header"`
}

func EncryptCookiePayload(payload CookiePayload) (string, error) {
	secret := sessionSecret()
	if secret == "" {
		return "", errors.New("wechat export session secret is not configured")
	}
	plain, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, plain, nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func DecryptCookiePayload(raw string) (CookiePayload, error) {
	var payload CookiePayload
	secret := sessionSecret()
	if secret == "" {
		return payload, errors.New("wechat export session secret is not configured")
	}
	encrypted, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return payload, err
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return payload, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return payload, err
	}
	if len(encrypted) < gcm.NonceSize() {
		return payload, errors.New("wechat cookie payload is invalid")
	}
	nonce := encrypted[:gcm.NonceSize()]
	ciphertext := encrypted[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return payload, err
	}
	if err := json.Unmarshal(plain, &payload); err != nil {
		return payload, err
	}
	if strings.TrimSpace(payload.CookieHeader) == "" {
		return payload, errors.New("wechat cookie payload is empty")
	}
	return payload, nil
}

func sessionSecret() string {
	if secret := strings.TrimSpace(os.Getenv("WECHAT_EXPORT_SESSION_SECRET")); secret != "" {
		return secret
	}
	return strings.TrimSpace(os.Getenv("JWT_SECRET"))
}
