//go:build unit

package service

import (
	"testing"
)

func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		account  Account
		expected string
	}{
		{
			name: "non-apikey type returns empty",
			account: Account{
				Type:     AccountTypeOAuth,
				Platform: PlatformAnthropic,
			},
			expected: "",
		},
		{
			name: "apikey without base_url returns default anthropic",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAnthropic,
				Credentials: map[string]any{},
			},
			expected: "https://api.anthropic.com",
		},
		{
			name: "apikey with custom base_url",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAnthropic,
				Credentials: map[string]any{"base_url": "https://custom.example.com"},
			},
			expected: "https://custom.example.com",
		},
		{
			name: "antigravity apikey auto-appends /antigravity",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com"},
			},
			expected: "https://upstream.example.com/antigravity",
		},
		{
			name: "antigravity apikey trims trailing slash before appending",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com/"},
			},
			expected: "https://upstream.example.com/antigravity",
		},
		{
			name: "antigravity non-apikey returns empty",
			account: Account{
				Type:        AccountTypeOAuth,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.GetBaseURL()
			if result != tt.expected {
				t.Errorf("GetBaseURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetGeminiBaseURL(t *testing.T) {
	const defaultGeminiURL = "https://generativelanguage.googleapis.com"

	tests := []struct {
		name     string
		account  Account
		expected string
	}{
		{
			name: "apikey without base_url returns default",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformGemini,
				Credentials: map[string]any{},
			},
			expected: defaultGeminiURL,
		},
		{
			name: "apikey with custom base_url",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformGemini,
				Credentials: map[string]any{"base_url": "https://custom-gemini.example.com"},
			},
			expected: "https://custom-gemini.example.com",
		},
		{
			name: "antigravity apikey auto-appends /antigravity",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com"},
			},
			expected: "https://upstream.example.com/antigravity",
		},
		{
			name: "antigravity apikey trims trailing slash",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com/"},
			},
			expected: "https://upstream.example.com/antigravity",
		},
		{
			name: "antigravity oauth does NOT append /antigravity",
			account: Account{
				Type:        AccountTypeOAuth,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com"},
			},
			expected: "https://upstream.example.com",
		},
		{
			name: "oauth without base_url returns default",
			account: Account{
				Type:        AccountTypeOAuth,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{},
			},
			expected: defaultGeminiURL,
		},
		{
			name: "nil credentials returns default",
			account: Account{
				Type:     AccountTypeAPIKey,
				Platform: PlatformGemini,
			},
			expected: defaultGeminiURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.GetGeminiBaseURL(defaultGeminiURL)
			if result != tt.expected {
				t.Errorf("GetGeminiBaseURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetGeminiWebBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		account  Account
		expected string
	}{
		{
			name: "uses default localhost when base_url missing",
			account: Account{
				Type:        AccountTypeGeminiWeb,
				Platform:    PlatformGemini,
				Credentials: map[string]any{},
			},
			expected: "http://127.0.0.1:8000",
		},
		{
			name: "uses configured base_url when present",
			account: Account{
				Type:        AccountTypeGeminiWeb,
				Platform:    PlatformGemini,
				Credentials: map[string]any{"base_url": "https://gemini-web.example.com"},
			},
			expected: "https://gemini-web.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.GetGeminiWebBaseURL()
			if result != tt.expected {
				t.Errorf("GetGeminiWebBaseURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetGeminiWebAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		account  Account
		expected string
	}{
		{
			name: "prefers api_key",
			account: Account{
				Credentials: map[string]any{"api_key": "gw-key", "access_token": "legacy-token"},
			},
			expected: "gw-key",
		},
		{
			name: "falls back to access_token",
			account: Account{
				Credentials: map[string]any{"access_token": "legacy-token"},
			},
			expected: "legacy-token",
		},
		{
			name: "returns empty when both missing",
			account: Account{
				Credentials: map[string]any{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.GetGeminiWebAPIKey()
			if result != tt.expected {
				t.Errorf("GetGeminiWebAPIKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}
