package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/Aias00/cloudbase/internal/worker"
)

type WeChatExportWorkerRuntimeConfig = worker.WeChatExportRuntimeConfig
type ImageWorkspaceObjectStorageRuntimeConfig = worker.ImageWorkspaceObjectStorageRuntimeConfig
type ImageWorkspaceWorkerRuntimeConfig = worker.ImageWorkspaceRuntimeConfig

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
	return worker.DefaultWeChatExportRuntimeConfig()
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
	return worker.DefaultImageWorkspaceRuntimeConfig()
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
