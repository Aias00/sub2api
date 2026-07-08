// Package vendorpreset provides a curated catalog of well-known upstream AI
// vendors that expose an OpenAI-compatible (or otherwise supported) API.
//
// It mirrors the "channel catalog" idea from new-api (ChannelBaseURLs /
// ChannelTypeNames): instead of forcing an operator to hand-type a base URL and
// model list every time they onboard a paid vendor, the admin UI can offer a
// dropdown that auto-fills base_url, default models and the API style.
//
// This package is intentionally side-effect free and data-only: it is a static,
// code-level catalog. The per-account instance configuration (the actual
// api_key, chosen models, model_mapping) still lives on the Account entity
// (credentials/extra JSONB), so this package introduces NO schema change and
// does not touch the forwarding hot path.
//
// Base URLs follow the OpenAI-compatible URL builder contract used by the
// gateway: a base ending in "/v1" gets "/chat/completions" appended, a bare
// domain gets "/v1/chat/completions" appended. Therefore every OpenAI-style
// preset here uses a base URL that ends in a version segment ("/v1", "/openai/v1",
// "/compatible-mode/v1", ...) so the resulting upstream URL is correct.
package vendorpreset

import "github.com/Aias00/cloudbase/internal/domain"

// APIStyle identifies the wire protocol an upstream speaks. It is informational
// metadata for the UI and future per-vendor adaptation; the current forwarding
// path treats openai/grok platforms as OpenAI-compatible.
type APIStyle string

const (
	APIStyleOpenAI    APIStyle = "openai"    // OpenAI Chat Completions compatible
	APIStyleAnthropic APIStyle = "anthropic" // Anthropic Messages compatible
	APIStyleGemini    APIStyle = "gemini"    // Google Gemini compatible
)

// VendorPreset describes an out-of-the-box upstream vendor entry.
type VendorPreset struct {
	// ID is a stable, machine-readable slug (e.g. "deepseek"). Used as the
	// selection key from the UI. Never change an existing ID.
	ID string `json:"id"`
	// DisplayName is the human-readable vendor name shown in the dropdown.
	DisplayName string `json:"display_name"`
	// Platform is the Cloudbase platform bucket this vendor maps to. Almost all
	// third-party paid vendors are OpenAI-compatible and map to PlatformOpenAI.
	Platform string `json:"platform"`
	// AccountType is the recommended account type to create for this vendor.
	// For key-based paid vendors this is "apikey".
	AccountType string `json:"account_type"`
	// APIStyle is the wire protocol the vendor speaks.
	APIStyle APIStyle `json:"api_style"`
	// BaseURL is the upstream base URL pre-filled into credentials.base_url.
	BaseURL string `json:"base_url"`
	// DefaultModels is a starter model list pre-filled into extra.models.
	DefaultModels []string `json:"default_models"`
	// DocsURL points at the vendor's API docs (optional, for the UI).
	DocsURL string `json:"docs_url,omitempty"`
}

// presets is the curated catalog. Ordering here defines UI ordering.
//
// Only vendors whose OpenAI-compatible endpoint base ends in a "/vN" version
// segment (or is handled correctly by the URL builder) are included, so the
// existing forwarding path produces correct chat/completions URLs without any
// per-vendor code. Vendors needing non-OpenAI request/response adaptation are
// deferred to a later phase.
var presets = []VendorPreset{
	{
		ID:          "deepseek",
		DisplayName: "DeepSeek",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://api.deepseek.com/v1",
		DefaultModels: []string{
			"deepseek-chat",
			"deepseek-reasoner",
		},
		DocsURL: "https://api-docs.deepseek.com",
	},
	{
		ID:          "moonshot",
		DisplayName: "Moonshot (Kimi)",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://api.moonshot.cn/v1",
		DefaultModels: []string{
			"kimi-k2-0711-preview",
			"moonshot-v1-8k",
			"moonshot-v1-32k",
			"moonshot-v1-128k",
		},
		DocsURL: "https://platform.moonshot.cn/docs",
	},
	{
		ID:          "qwen",
		DisplayName: "Qwen (DashScope)",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		// DashScope's OpenAI-compatible mode base ends in /v1.
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		DefaultModels: []string{
			"qwen-max",
			"qwen-plus",
			"qwen-turbo",
			"qwen3-coder-plus",
		},
		DocsURL: "https://help.aliyun.com/zh/model-studio/developer-reference/compatibility-of-openai-with-dashscope",
	},
	{
		ID:          "siliconflow",
		DisplayName: "SiliconFlow",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://api.siliconflow.cn/v1",
		DefaultModels: []string{
			"deepseek-ai/DeepSeek-V3",
			"deepseek-ai/DeepSeek-R1",
			"Qwen/Qwen2.5-72B-Instruct",
		},
		DocsURL: "https://docs.siliconflow.cn",
	},
	{
		ID:          "openrouter",
		DisplayName: "OpenRouter",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://openrouter.ai/api/v1",
		DefaultModels: []string{
			"openai/gpt-4o",
			"anthropic/claude-3.5-sonnet",
			"google/gemini-2.0-flash-001",
		},
		DocsURL: "https://openrouter.ai/docs",
	},
	{
		ID:          "mistral",
		DisplayName: "Mistral AI",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://api.mistral.ai/v1",
		DefaultModels: []string{
			"mistral-large-latest",
			"mistral-small-latest",
			"codestral-latest",
		},
		DocsURL: "https://docs.mistral.ai",
	},
	{
		ID:          "groq",
		DisplayName: "Groq",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://api.groq.com/openai/v1",
		DefaultModels: []string{
			"llama-3.3-70b-versatile",
			"llama-3.1-8b-instant",
		},
		DocsURL: "https://console.groq.com/docs",
	},
	{
		ID:          "together",
		DisplayName: "Together AI",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://api.together.xyz/v1",
		DefaultModels: []string{
			"meta-llama/Llama-3.3-70B-Instruct-Turbo",
			"deepseek-ai/DeepSeek-V3",
		},
		DocsURL: "https://docs.together.ai",
	},
	{
		ID:          "minimax",
		DisplayName: "MiniMax",
		Platform:    domain.PlatformOpenAI,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://api.minimax.chat/v1",
		DefaultModels: []string{
			"abab6.5s-chat",
		},
		DocsURL: "https://platform.minimaxi.com/document",
	},
	{
		ID:          "xai",
		DisplayName: "xAI (Grok API)",
		Platform:    domain.PlatformGrok,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleOpenAI,
		BaseURL:     "https://api.x.ai/v1",
		DefaultModels: []string{
			"grok-4",
			"grok-3",
			"grok-3-mini",
		},
		DocsURL: "https://docs.x.ai",
	},

	// ── Anthropic-style vendors (Claude Messages compatible) ──────────────
	// These map to PlatformAnthropic + apikey, so the gateway routes them
	// through the existing Anthropic API-key passthrough path, which appends
	// "/v1/messages" to the base URL. Therefore each BaseURL below is a ROOT
	// (no "/v1" / "/v1/messages" suffix).
	{
		ID:          "anthropic",
		DisplayName: "Anthropic (Claude API)",
		Platform:    domain.PlatformAnthropic,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleAnthropic,
		BaseURL:     "https://api.anthropic.com",
		DefaultModels: []string{
			"claude-opus-4-1",
			"claude-sonnet-4-5",
			"claude-haiku-4-5",
		},
		DocsURL: "https://docs.anthropic.com",
	},
	{
		ID:          "zhipu-anthropic",
		DisplayName: "Zhipu GLM (Anthropic 兼容)",
		Platform:    domain.PlatformAnthropic,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleAnthropic,
		BaseURL:     "https://open.bigmodel.cn/api/anthropic",
		DefaultModels: []string{
			"glm-4.6",
			"glm-4.5",
		},
		DocsURL: "https://docs.bigmodel.cn",
	},
	{
		ID:          "zai-anthropic",
		DisplayName: "Z.ai GLM (Anthropic, Intl)",
		Platform:    domain.PlatformAnthropic,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleAnthropic,
		BaseURL:     "https://api.z.ai/api/anthropic",
		DefaultModels: []string{
			"glm-4.6",
		},
		DocsURL: "https://docs.z.ai",
	},
	{
		ID:          "moonshot-anthropic",
		DisplayName: "Moonshot Kimi (Anthropic 兼容)",
		Platform:    domain.PlatformAnthropic,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleAnthropic,
		BaseURL:     "https://api.moonshot.cn/anthropic",
		DefaultModels: []string{
			"kimi-k2-0711-preview",
		},
		DocsURL: "https://platform.moonshot.cn/docs",
	},
	{
		ID:          "deepseek-anthropic",
		DisplayName: "DeepSeek (Anthropic 兼容)",
		Platform:    domain.PlatformAnthropic,
		AccountType: domain.AccountTypeAPIKey,
		APIStyle:    APIStyleAnthropic,
		BaseURL:     "https://api.deepseek.com/anthropic",
		DefaultModels: []string{
			"deepseek-chat",
			"deepseek-reasoner",
		},
		DocsURL: "https://api-docs.deepseek.com",
	},
}

// All returns a copy of the full vendor catalog in UI display order.
// The returned slice is a shallow copy; callers must not mutate the shared
// DefaultModels slices.
func All() []VendorPreset {
	out := make([]VendorPreset, len(presets))
	copy(out, presets)
	return out
}

// ByID returns the preset with the given ID and whether it was found.
func ByID(id string) (VendorPreset, bool) {
	for _, p := range presets {
		if p.ID == id {
			return p, true
		}
	}
	return VendorPreset{}, false
}
