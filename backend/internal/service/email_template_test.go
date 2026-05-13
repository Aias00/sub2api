//go:build unit

package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildVerifyCodeEmailBody_UsesCursorStyleCodeTemplate(t *testing.T) {
	s := &EmailService{}
	body := s.buildVerifyCodeEmailBody("199929", "cloudbase")

	require.Contains(t, body, "Verify your email for cloudbase")
	require.Contains(t, body, "You requested a one-time verification code for cloudbase")
	require.Contains(t, body, "199929")
	require.Contains(t, body, "This code expires in 15 minutes.")
	require.Contains(t, body, "border:1px solid #dedede")
	require.NotContains(t, body, "linear-gradient")
	require.NotContains(t, body, "%!")
}

func TestBuildVerifyCodeEmailBody_EscapesSiteName(t *testing.T) {
	s := &EmailService{}
	body := s.buildVerifyCodeEmailBody("123456", `site<script>`)

	require.Contains(t, body, "site&lt;script&gt;")
	require.NotContains(t, body, "<script>")
}

func TestBuildPasswordResetEmailBody_UsesActionTemplate(t *testing.T) {
	s := &EmailService{}
	body := s.buildPasswordResetEmailBody(`https://example.com/reset?email=a@b.com&token=<x>`, "cloudbase")

	require.Contains(t, body, "Reset your cloudbase password")
	require.Contains(t, body, "Reset password")
	require.Contains(t, body, "This link expires in 30 minutes.")
	require.Contains(t, body, "https://example.com/reset?email=a@b.com&amp;token=&lt;x&gt;")
	require.NotContains(t, body, "%!")
}

func TestBuildNotifyVerifyEmailBody_UsesSharedCodeTemplate(t *testing.T) {
	body := buildNotifyVerifyEmailBody("654321", "cloudbase")

	require.Contains(t, body, "Verify notification email for cloudbase")
	require.Contains(t, body, "654321")
	require.Contains(t, body, "This code expires in 15 minutes.")
	require.NotContains(t, body, "linear-gradient")
}

func TestBuildTestEmailBody_UsesSharedTemplate(t *testing.T) {
	body := BuildTestEmailBody("cloudbase")

	require.Contains(t, body, "Test email from cloudbase")
	require.Contains(t, body, "Sent successfully")
	require.True(t, strings.Contains(body, "border-radius:22px"), "expected shared card styling")
	require.NotContains(t, body, "%!")
}
