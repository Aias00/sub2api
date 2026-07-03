package billing

import (
	"context"
	"errors"
)

// ModelPricing describes normalized per-token model pricing in USD.
type ModelPricing struct {
	InputPricePerToken             float64
	InputPricePerTokenPriority     float64
	ImageInputPricePerToken        float64
	OutputPricePerToken            float64
	OutputPricePerTokenPriority    float64
	CacheCreationPricePerToken     float64
	CacheReadPricePerToken         float64
	CacheReadPricePerTokenPriority float64
	CacheCreation5mPrice           float64
	CacheCreation1hPrice           float64
	SupportsCacheBreakdown         bool
	LongContextInputThreshold      int
	LongContextInputMultiplier     float64
	LongContextOutputMultiplier    float64
	ImageOutputPricePerToken       float64
	ImageOutputPriceExplicit       bool
}

// LiteLLMModelPricing is the normalized subset of LiteLLM pricing metadata used
// by billing.
type LiteLLMModelPricing struct {
	InputCostPerToken                   float64 `json:"input_cost_per_token"`
	InputCostPerTokenPriority           float64 `json:"input_cost_per_token_priority"`
	OutputCostPerToken                  float64 `json:"output_cost_per_token"`
	OutputCostPerTokenPriority          float64 `json:"output_cost_per_token_priority"`
	CacheCreationInputTokenCost         float64 `json:"cache_creation_input_token_cost"`
	CacheCreationInputTokenCostAbove1hr float64 `json:"cache_creation_input_token_cost_above_1hr"`
	CacheReadInputTokenCost             float64 `json:"cache_read_input_token_cost"`
	CacheReadInputTokenCostPriority     float64 `json:"cache_read_input_token_cost_priority"`
	LongContextInputTokenThreshold      int     `json:"long_context_input_token_threshold,omitempty"`
	LongContextInputCostMultiplier      float64 `json:"long_context_input_cost_multiplier,omitempty"`
	LongContextOutputCostMultiplier     float64 `json:"long_context_output_cost_multiplier,omitempty"`
	SupportsServiceTier                 bool    `json:"supports_service_tier"`
	LiteLLMProvider                     string  `json:"litellm_provider"`
	Mode                                string  `json:"mode"`
	SupportsPromptCaching               bool    `json:"supports_prompt_caching"`
	OutputCostPerImage                  float64 `json:"output_cost_per_image"`
	OutputCostPerImageToken             float64 `json:"output_cost_per_image_token"`
}

// PricingProvider is the dynamic model pricing port consumed by billing.
type PricingProvider interface {
	GetModelPricing(modelName string) *LiteLLMModelPricing
	GetStatus() map[string]any
	ForceUpdate() error
}

// UsageTokens is the normalized usage shape consumed by cost calculation.
type UsageTokens struct {
	InputTokens           int
	ImageInputTokens      int
	OutputTokens          int
	CacheCreationTokens   int
	CacheReadTokens       int
	CacheCreation5mTokens int
	CacheCreation1hTokens int
	ImageOutputTokens     int
}

// CostBreakdown is the itemized billing result for one usage event.
type CostBreakdown struct {
	InputCost         float64
	OutputCost        float64
	ImageOutputCost   float64
	CacheCreationCost float64
	CacheReadCost     float64
	TotalCost         float64
	ActualCost        float64
	BillingMode       string
}

// ErrModelPricingUnavailable indicates that no configured pricing source can
// price the requested model.
var ErrModelPricingUnavailable = errors.New("pricing not found")

// PricingSource identifies which pricing source produced a resolution.
const (
	PricingSourceChannel  = "channel"
	PricingSourceLiteLLM  = "litellm"
	PricingSourceFallback = "fallback"
)

// PricingInput is the stable request contract for resolving model pricing.
type PricingInput struct {
	Model   string
	GroupID *int64
}

// PricingResolver is the billing-layer contract for components that can resolve
// model pricing. Existing service resolvers can adapt richer channel data behind
// this request shape without leaking gateway details into billing callers.
type PricingResolver[T any] interface {
	Resolve(ctx context.Context, input PricingInput) T
}
