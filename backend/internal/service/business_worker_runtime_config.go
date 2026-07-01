package service

import (
	"context"
	"strconv"
	"strings"
)

type WeChatExportWorkerRuntimeConfig struct {
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

type ImageWorkspaceWorkerRuntimeConfig struct {
	UpstreamURL         string                                   `json:"upstream_url"`
	GenerationTimeoutMS int                                      `json:"generation_timeout_ms"`
	CompletionCost      string                                   `json:"completion_cost"`
	CompletionCostMap   string                                   `json:"completion_cost_map_json"`
	PromptSafetyEnabled bool                                     `json:"prompt_safety_enabled"`
	AssumeWorkerReady   bool                                     `json:"assume_worker_ready"`
	ObjectStorage       ImageWorkspaceObjectStorageRuntimeConfig `json:"object_storage"`
	MediaCDNBaseURL     string                                   `json:"media_cdn_base_url"`
}

func (s *SettingService) GetWeChatExportWorkerRuntimeConfig(ctx context.Context) (WeChatExportWorkerRuntimeConfig, error) {
	if s == nil || s.settingRepo == nil {
		return defaultWeChatExportWorkerRuntimeConfig(), nil
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyWeChatExportFetchRetries,
		SettingKeyWeChatExportFetchTimeoutMS,
		SettingKeyWeChatExportWorkerConcurrency,
		SettingKeyWeChatExportWorkerIntervalMS,
		SettingKeyWeChatExportWorkerLeaseSeconds,
		SettingKeyWeChatExportWorkerMaxBackoffMS,
	})
	if err != nil {
		return WeChatExportWorkerRuntimeConfig{}, err
	}
	return weChatExportWorkerRuntimeConfigFromSettings(values), nil
}

func (s *SettingService) GetImageWorkspaceWorkerRuntimeConfig(ctx context.Context) (ImageWorkspaceWorkerRuntimeConfig, error) {
	if s == nil || s.settingRepo == nil {
		return defaultImageWorkspaceWorkerRuntimeConfig(), nil
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyImageWorkspaceUpstreamURL,
		SettingKeyImageWorkspaceGenerationTimeoutMS,
		SettingKeyImageWorkspaceCompletionCost,
		SettingKeyImageWorkspaceCompletionCostMapJSON,
		SettingKeyImageWorkspacePromptSafetyEnabled,
		SettingKeyImageWorkspaceAssumeWorkerReady,
		SettingKeyImageWorkspaceObjectStorageEnabled,
		SettingKeyImageWorkspaceObjectStorageProvider,
		SettingKeyImageWorkspaceObjectStorageBucket,
		SettingKeyImageWorkspaceObjectStorageRegion,
		SettingKeyImageWorkspaceObjectStoragePrefix,
		SettingKeyImageWorkspaceObjectStoragePublicBaseURL,
		SettingKeyMediaCDNBaseURL,
	})
	if err != nil {
		return ImageWorkspaceWorkerRuntimeConfig{}, err
	}
	return imageWorkspaceWorkerRuntimeConfigFromSettings(values), nil
}

func defaultWeChatExportWorkerRuntimeConfig() WeChatExportWorkerRuntimeConfig {
	return WeChatExportWorkerRuntimeConfig{
		FetchRetries:       2,
		FetchTimeoutMS:     20000,
		WorkerConcurrency:  1,
		WorkerIntervalMS:   5000,
		WorkerLeaseSeconds: 300,
		WorkerMaxBackoffMS: 60000,
	}
}

func weChatExportWorkerRuntimeConfigFromSettings(values map[string]string) WeChatExportWorkerRuntimeConfig {
	cfg := defaultWeChatExportWorkerRuntimeConfig()
	cfg.FetchRetries = parseBoundedIntSetting(values[SettingKeyWeChatExportFetchRetries], cfg.FetchRetries, 0, 5)
	cfg.FetchTimeoutMS = parseBoundedIntSetting(values[SettingKeyWeChatExportFetchTimeoutMS], cfg.FetchTimeoutMS, 1000, 120000)
	cfg.WorkerConcurrency = parseBoundedIntSetting(values[SettingKeyWeChatExportWorkerConcurrency], cfg.WorkerConcurrency, 1, 8)
	cfg.WorkerIntervalMS = parseBoundedIntSetting(values[SettingKeyWeChatExportWorkerIntervalMS], cfg.WorkerIntervalMS, 1000, 300000)
	cfg.WorkerLeaseSeconds = parseBoundedIntSetting(values[SettingKeyWeChatExportWorkerLeaseSeconds], cfg.WorkerLeaseSeconds, 60, 3600)
	cfg.WorkerMaxBackoffMS = parseBoundedIntSetting(values[SettingKeyWeChatExportWorkerMaxBackoffMS], cfg.WorkerMaxBackoffMS, cfg.WorkerIntervalMS, 300000)
	return cfg
}

func defaultImageWorkspaceWorkerRuntimeConfig() ImageWorkspaceWorkerRuntimeConfig {
	return ImageWorkspaceWorkerRuntimeConfig{
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

func imageWorkspaceWorkerRuntimeConfigFromSettings(values map[string]string) ImageWorkspaceWorkerRuntimeConfig {
	cfg := defaultImageWorkspaceWorkerRuntimeConfig()
	if value := strings.TrimSpace(values[SettingKeyImageWorkspaceUpstreamURL]); value != "" {
		cfg.UpstreamURL = value
	}
	cfg.GenerationTimeoutMS = parseBoundedIntSetting(values[SettingKeyImageWorkspaceGenerationTimeoutMS], cfg.GenerationTimeoutMS, 1000, 900000)
	if value := strings.TrimSpace(values[SettingKeyImageWorkspaceCompletionCost]); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil && parsed >= 0 {
			cfg.CompletionCost = value
		}
	}
	cfg.CompletionCostMap = normalizeJSONObjectSetting(values[SettingKeyImageWorkspaceCompletionCostMapJSON], "{}")
	cfg.PromptSafetyEnabled = parseBoolSettingWithDefault(values[SettingKeyImageWorkspacePromptSafetyEnabled], true)
	cfg.AssumeWorkerReady = parseBoolSettingWithDefault(values[SettingKeyImageWorkspaceAssumeWorkerReady], false)
	cfg.ObjectStorage.Enabled = parseBoolSettingWithDefault(values[SettingKeyImageWorkspaceObjectStorageEnabled], false)
	cfg.ObjectStorage.Provider = firstNonEmpty(values[SettingKeyImageWorkspaceObjectStorageProvider], cfg.ObjectStorage.Provider)
	cfg.ObjectStorage.Bucket = strings.TrimSpace(values[SettingKeyImageWorkspaceObjectStorageBucket])
	cfg.ObjectStorage.Region = firstNonEmpty(values[SettingKeyImageWorkspaceObjectStorageRegion], cfg.ObjectStorage.Region)
	cfg.ObjectStorage.KeyPrefix = strings.Trim(strings.TrimSpace(firstNonEmpty(values[SettingKeyImageWorkspaceObjectStoragePrefix], cfg.ObjectStorage.KeyPrefix)), "/")
	cfg.ObjectStorage.PublicBaseURL = strings.TrimSpace(values[SettingKeyImageWorkspaceObjectStoragePublicBaseURL])
	cfg.MediaCDNBaseURL = strings.TrimSpace(values[SettingKeyMediaCDNBaseURL])
	return cfg
}
