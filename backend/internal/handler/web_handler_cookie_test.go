package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWebSessionCookiesUseGenericNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/web/auth/login", nil)

	handler := &WebHandler{}
	handler.setWebSessionCookies(c, "access-token", "refresh-token", 120)

	cookies := rec.Result().Cookies()
	require.Equal(t, "access-token", findCookieValue(cookies, webAccessTokenCookie))
	require.Equal(t, "refresh-token", findCookieValue(cookies, webRefreshTokenCookie))
}

func TestWebSessionCookieKeepsLaxForUntrustedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/web/auth/login", nil)
	c.Request.Host = "api.example.com"
	c.Request.Header.Set("Origin", "https://evil.example.com")

	handler := &WebHandler{}
	handler.setWebSessionCookies(c, "access-token", "refresh-token", 120)

	cookie := findCookie(rec.Result().Cookies(), webAccessTokenCookie)
	require.NotNil(t, cookie)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestWebSessionCookieUsesNoneForAllowedCredentialedCrossSiteOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/web/auth/login", nil)
	c.Request.Host = "api.example.com"
	c.Request.Header.Set("Origin", "https://app.example.com")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "https://app.example.com")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	handler := &WebHandler{}
	handler.setWebSessionCookies(c, "access-token", "refresh-token", 120)

	cookie := findCookie(rec.Result().Cookies(), webAccessTokenCookie)
	require.NotNil(t, cookie)
	require.Equal(t, http.SameSiteNoneMode, cookie.SameSite)
	require.True(t, cookie.Secure)
}

func TestReadWebSessionCookieIgnoresLegacyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/web/auth/me", nil)
	c.Request.AddCookie(&http.Cookie{Name: "touch_cloudbase_access_token", Value: "legacy-access-token"})

	value, err := readWebSessionCookie(c, webAccessTokenCookie)

	require.Error(t, err)
	require.Empty(t, value)
}

func TestClearWebSessionCookiesClearsGenericNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/web/auth/logout", nil)

	handler := &WebHandler{}
	handler.clearWebSessionCookies(c)

	cookies := rec.Result().Cookies()
	require.Equal(t, -1, findCookieMaxAge(cookies, webAccessTokenCookie))
	require.Equal(t, -1, findCookieMaxAge(cookies, webRefreshTokenCookie))
}

func TestWebCheckoutPaymentSourceDefaultsToGenericWebSource(t *testing.T) {
	require.Equal(t, webPaymentSource, webCheckoutPaymentSource(""))
	require.Equal(t, webPaymentSource, webCheckoutPaymentSource("   "))
	require.Equal(t, "explicit_source", webCheckoutPaymentSource(" explicit_source "))
}

func findCookieValue(cookies []*http.Cookie, name string) string {
	if cookie := findCookie(cookies, name); cookie != nil {
		return cookie.Value
	}
	return ""
}

func findCookieMaxAge(cookies []*http.Cookie, name string) int {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.MaxAge
		}
	}
	return 0
}
