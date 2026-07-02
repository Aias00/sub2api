package handler

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	oauthMaxRedirectLen          = 2048
	oauthMaxFragmentValueLen     = 128
	oauthDefaultRedirectTo       = "/dashboard"
	oauthIntentLogin             = "login"
	wechatOAuthProviderKey       = "wechat-main"
	wechatOAuthLegacyProviderKey = "wechat"
)

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func singleLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func sanitizeFrontendRedirectPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if len(path) > oauthMaxRedirectLen {
		return ""
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return ""
	}
	if strings.Contains(path, "://") || strings.ContainsAny(path, "\r\n") {
		return ""
	}
	return path
}

func isRequestHTTPS(c *gin.Context) bool {
	if c != nil && c.Request != nil && c.Request.TLS != nil {
		return true
	}
	proto := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")))
	return proto == "https"
}

func encodeCookieValue(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func decodeCookieValue(value string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func readCookieDecoded(c *gin.Context, name string) (string, error) {
	ck, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return decodeCookieValue(ck.Value)
}

func truncateLogValue(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if value == "" || maxLen <= 0 {
		return ""
	}
	if len(value) <= maxLen {
		return value
	}
	value = value[:maxLen]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

func truncateFragmentValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) > oauthMaxFragmentValueLen {
		value = value[:oauthMaxFragmentValueLen]
		for !utf8.ValidString(value) {
			value = value[:len(value)-1]
		}
	}
	return value
}

func clearCookieWithPath(c *gin.Context, name string, path string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func redirectOAuthError(c *gin.Context, frontendCallback string, code string, message string, description string) {
	fragment := url.Values{}
	fragment.Set("error", truncateFragmentValue(code))
	if strings.TrimSpace(message) != "" {
		fragment.Set("error_message", truncateFragmentValue(message))
	}
	if strings.TrimSpace(description) != "" {
		fragment.Set("error_description", truncateFragmentValue(description))
	}
	redirectWithFragment(c, frontendCallback, fragment)
}

func redirectOAuthTokenPair(c *gin.Context, frontendCallback string, tokenPair *service.TokenPair, redirectTo string) {
	fragment := url.Values{}
	if tokenPair != nil {
		fragment.Set("access_token", truncateFragmentValue(tokenPair.AccessToken))
		fragment.Set("refresh_token", truncateFragmentValue(tokenPair.RefreshToken))
		fragment.Set("expires_in", strconv.Itoa(tokenPair.ExpiresIn))
		fragment.Set("token_type", "Bearer")
	}
	if redirect := strings.TrimSpace(redirectTo); redirect != "" {
		originalRedirect := redirect
		for range 2 {
			decoded, err := url.QueryUnescape(redirect)
			if err != nil || decoded == redirect {
				break
			}
			redirect = decoded
		}
		if redirect != originalRedirect {
			if sanitized := sanitizeFrontendRedirectPath(redirect); sanitized != "" {
				redirect = sanitized
			} else {
				redirect = originalRedirect
			}
		}
		fragment.Set("redirect", truncateFragmentValue(redirect))
	}
	redirectWithFragment(c, frontendCallback, fragment)
}

func redirectWithFragment(c *gin.Context, frontendCallback string, fragment url.Values) {
	u, err := url.Parse(frontendCallback)
	if err != nil {
		c.Redirect(http.StatusFound, oauthDefaultRedirectTo)
		return
	}
	if u.Scheme != "" && !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		c.Redirect(http.StatusFound, oauthDefaultRedirectTo)
		return
	}
	u.Fragment = fragment.Encode()
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Redirect(http.StatusFound, u.String())
}
