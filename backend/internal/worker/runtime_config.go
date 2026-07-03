package worker

import "context"

type WeChatExportRuntimeConfig struct {
	FetchRetries       int `json:"fetch_retries"`
	FetchTimeoutMS     int `json:"fetch_timeout_ms"`
	WorkerConcurrency  int `json:"worker_concurrency"`
	WorkerIntervalMS   int `json:"worker_interval_ms"`
	WorkerLeaseSeconds int `json:"worker_lease_seconds"`
	WorkerMaxBackoffMS int `json:"worker_max_backoff_ms"`
}

type ImageWorkspaceObjectStorageRuntimeConfig struct {
	Enabled       bool   `json:"enabled"`
	Provider      string `json:"provider"`
	Bucket        string `json:"bucket"`
	Region        string `json:"region"`
	KeyPrefix     string `json:"key_prefix"`
	PublicBaseURL string `json:"public_base_url"`
}

type ImageWorkspaceRuntimeConfig struct {
	UpstreamURL         string                                   `json:"upstream_url"`
	GenerationTimeoutMS int                                      `json:"generation_timeout_ms"`
	CompletionCost      string                                   `json:"completion_cost"`
	CompletionCostMap   string                                   `json:"completion_cost_map_json"`
	PromptSafetyEnabled bool                                     `json:"prompt_safety_enabled"`
	AssumeWorkerReady   bool                                     `json:"assume_worker_ready"`
	ObjectStorage       ImageWorkspaceObjectStorageRuntimeConfig `json:"object_storage"`
	MediaCDNBaseURL     string                                   `json:"media_cdn_base_url"`
}

type WeChatExportRuntimeConfigProvider interface {
	GetWorkerRuntimeConfig(ctx context.Context) (WeChatExportRuntimeConfig, error)
}

type ImageWorkspaceRuntimeConfigProvider interface {
	GetWorkerRuntimeConfig(ctx context.Context) (ImageWorkspaceRuntimeConfig, error)
}

func DefaultWeChatExportRuntimeConfig() WeChatExportRuntimeConfig {
	return WeChatExportRuntimeConfig{
		FetchRetries:       2,
		FetchTimeoutMS:     20000,
		WorkerConcurrency:  1,
		WorkerIntervalMS:   5000,
		WorkerLeaseSeconds: 300,
		WorkerMaxBackoffMS: 60000,
	}
}

func DefaultImageWorkspaceRuntimeConfig() ImageWorkspaceRuntimeConfig {
	return ImageWorkspaceRuntimeConfig{
		UpstreamURL:         "https://api.openai.com/v1/images/generations",
		GenerationTimeoutMS: 420000,
		CompletionCost:      "0",
		CompletionCostMap:   "{}",
		PromptSafetyEnabled: true,
		ObjectStorage: ImageWorkspaceObjectStorageRuntimeConfig{
			Provider:  "r2",
			Region:    "auto",
			KeyPrefix: "image-workspace",
		},
	}
}
