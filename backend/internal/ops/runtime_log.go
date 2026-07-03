package ops

import (
	"errors"
	"strings"
)

type RuntimeLogConfig struct {
	Level           string         `json:"level"`
	EnableSampling  bool           `json:"enable_sampling"`
	SamplingInitial int            `json:"sampling_initial"`
	SamplingNext    int            `json:"sampling_thereafter"`
	Caller          bool           `json:"caller"`
	StacktraceLevel string         `json:"stacktrace_level"`
	RetentionDays   int            `json:"retention_days"`
	Source          string         `json:"source,omitempty"`
	UpdatedAt       string         `json:"updated_at,omitempty"`
	UpdatedByUserID int64          `json:"updated_by_user_id,omitempty"`
	Extra           map[string]any `json:"extra,omitempty"`
}

func DefaultRuntimeLogConfig() *RuntimeLogConfig {
	return &RuntimeLogConfig{
		Level:           "info",
		EnableSampling:  false,
		SamplingInitial: 100,
		SamplingNext:    100,
		Caller:          true,
		StacktraceLevel: "error",
		RetentionDays:   30,
	}
}

func NormalizeRuntimeLogConfig(cfg *RuntimeLogConfig, defaults *RuntimeLogConfig) {
	if cfg == nil || defaults == nil {
		return
	}
	cfg.Level = strings.ToLower(strings.TrimSpace(cfg.Level))
	if cfg.Level == "" {
		cfg.Level = defaults.Level
	}
	cfg.StacktraceLevel = strings.ToLower(strings.TrimSpace(cfg.StacktraceLevel))
	if cfg.StacktraceLevel == "" {
		cfg.StacktraceLevel = defaults.StacktraceLevel
	}
	if cfg.SamplingInitial <= 0 {
		cfg.SamplingInitial = defaults.SamplingInitial
	}
	if cfg.SamplingNext <= 0 {
		cfg.SamplingNext = defaults.SamplingNext
	}
	if cfg.RetentionDays <= 0 {
		cfg.RetentionDays = defaults.RetentionDays
	}
}

func ValidateRuntimeLogConfig(cfg *RuntimeLogConfig) error {
	if cfg == nil {
		return errors.New("invalid config")
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Level)) {
	case "debug", "info", "warn", "error":
	default:
		return errors.New("level must be one of: debug/info/warn/error")
	}
	switch strings.ToLower(strings.TrimSpace(cfg.StacktraceLevel)) {
	case "none", "error", "fatal":
	default:
		return errors.New("stacktrace_level must be one of: none/error/fatal")
	}
	if cfg.SamplingInitial <= 0 {
		return errors.New("sampling_initial must be positive")
	}
	if cfg.SamplingNext <= 0 {
		return errors.New("sampling_thereafter must be positive")
	}
	if cfg.RetentionDays < 1 || cfg.RetentionDays > 3650 {
		return errors.New("retention_days must be between 1 and 3650")
	}
	return nil
}
