package gateway

import "testing"

func TestBuildOpenAIEndpointURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		endpoint string
		want     string
	}{
		{
			name:     "plain base appends versioned endpoint",
			base:     "https://api.openai.com",
			endpoint: "/v1/responses/input_tokens",
			want:     "https://api.openai.com/v1/responses/input_tokens",
		},
		{
			name:     "versioned base appends relative endpoint",
			base:     "https://api.openai.com/v1",
			endpoint: "/v1/responses/input_tokens",
			want:     "https://api.openai.com/v1/responses/input_tokens",
		},
		{
			name:     "already exact endpoint is preserved",
			base:     "https://api.openai.com/v1/responses/input_tokens",
			endpoint: "/v1/responses/input_tokens",
			want:     "https://api.openai.com/v1/responses/input_tokens",
		},
		{
			name:     "already relative endpoint is preserved",
			base:     "https://proxy.example/responses/input_tokens",
			endpoint: "/v1/responses/input_tokens",
			want:     "https://proxy.example/responses/input_tokens",
		},
		{
			name:     "trims slash and whitespace",
			base:     " https://api.openai.com/v1/ ",
			endpoint: " responses/input_tokens ",
			want:     "https://api.openai.com/v1/responses/input_tokens",
		},
		{
			name:     "preview version base uses relative endpoint",
			base:     "https://example.test/api/v1beta",
			endpoint: "/v1/responses",
			want:     "https://example.test/api/v1beta/responses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildOpenAIEndpointURL(tt.base, tt.endpoint); got != tt.want {
				t.Fatalf("BuildOpenAIEndpointURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpenAIBaseURLHasVersionSuffix(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{"https://api.openai.com/v1", true},
		{"https://api.openai.com/v1/", true},
		{"https://api.openai.com/api/v1beta", true},
		{"https://api.openai.com/api/v1.2", true},
		{"proxy.local/v1preview", true},
		{"https://api.openai.com/responses", false},
		{"https://api.openai.com/v", false},
		{"https://api.openai.com/vbeta", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := OpenAIBaseURLHasVersionSuffix(tt.raw); got != tt.want {
				t.Fatalf("OpenAIBaseURLHasVersionSuffix(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestIsOpenAIAPIVersionSegment(t *testing.T) {
	tests := []struct {
		segment string
		want    bool
	}{
		{"v1", true},
		{"V1", true},
		{"v1.2", true},
		{"v1beta", true},
		{"v1preview", true},
		{"v1alpha", true},
		{"v", false},
		{"version1", false},
		{"vbeta", false},
		{"v1.", false},
		{"v1rc", false},
	}

	for _, tt := range tests {
		t.Run(tt.segment, func(t *testing.T) {
			if got := IsOpenAIAPIVersionSegment(tt.segment); got != tt.want {
				t.Fatalf("IsOpenAIAPIVersionSegment(%q) = %v, want %v", tt.segment, got, tt.want)
			}
		})
	}
}
