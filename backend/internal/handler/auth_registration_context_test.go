package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegistrationHeaderSnapshotKeepsDiagnosticsAndDropsSensitiveHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept-Language", "zh-CN")
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	req.Header.Set("X-Device-Fingerprint", "device-123")
	req.Header.Set("Cookie", "session=secret")
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("x-api-key", "secret")
	c.Request = req

	got := registrationHeaderSnapshot(c)

	if got["User-Agent"] != "Mozilla/5.0" {
		t.Fatalf("User-Agent = %q", got["User-Agent"])
	}
	if got["Accept-Language"] != "zh-CN" {
		t.Fatalf("Accept-Language = %q", got["Accept-Language"])
	}
	if got["X-Forwarded-For"] != "203.0.113.10" {
		t.Fatalf("X-Forwarded-For = %q", got["X-Forwarded-For"])
	}
	if _, ok := got["Cookie"]; ok {
		t.Fatal("Cookie must not be captured")
	}
	if _, ok := got["Authorization"]; ok {
		t.Fatal("Authorization must not be captured")
	}
	if _, ok := got["x-api-key"]; ok {
		t.Fatal("x-api-key must not be captured")
	}
}
