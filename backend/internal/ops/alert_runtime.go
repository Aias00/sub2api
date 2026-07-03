package ops

import (
	"errors"
	"strings"
	"time"
)

type DistributedLockSettings struct {
	Enabled    bool   `json:"enabled"`
	Key        string `json:"key"`
	TTLSeconds int    `json:"ttl_seconds"`
}

type AlertSilenceEntry struct {
	RuleID     *int64   `json:"rule_id,omitempty"`
	Severities []string `json:"severities,omitempty"`

	UntilRFC3339 string `json:"until_rfc3339"`
	Reason       string `json:"reason"`
}

type AlertSilencingSettings struct {
	Enabled bool `json:"enabled"`

	GlobalUntilRFC3339 string `json:"global_until_rfc3339"`
	GlobalReason       string `json:"global_reason"`

	Entries []AlertSilenceEntry `json:"entries,omitempty"`
}

type AlertRuntimeSettings struct {
	EvaluationIntervalSeconds int `json:"evaluation_interval_seconds"`

	DistributedLock DistributedLockSettings `json:"distributed_lock"`
	Silencing       AlertSilencingSettings  `json:"silencing"`
	Thresholds      MetricThresholds        `json:"thresholds"`
}

type AlertSilenceTarget struct {
	RuleID        int64
	EventSeverity string
	RuleSeverity  string
}

func DefaultAlertRuntimeSettings(defaultLockKey string, defaultLockTTLSeconds int) *AlertRuntimeSettings {
	return &AlertRuntimeSettings{
		EvaluationIntervalSeconds: 60,
		DistributedLock: DistributedLockSettings{
			Enabled:    true,
			Key:        defaultLockKey,
			TTLSeconds: defaultLockTTLSeconds,
		},
		Silencing: AlertSilencingSettings{
			Enabled:            false,
			GlobalUntilRFC3339: "",
			GlobalReason:       "",
			Entries:            []AlertSilenceEntry{},
		},
	}
}

func NormalizeDistributedLockSettings(s *DistributedLockSettings, defaultKey string, defaultTTLSeconds int) {
	if s == nil {
		return
	}
	s.Key = strings.TrimSpace(s.Key)
	if s.Key == "" {
		s.Key = defaultKey
	}
	if s.TTLSeconds <= 0 {
		s.TTLSeconds = defaultTTLSeconds
	}
}

func NormalizeAlertSilencingSettings(s *AlertSilencingSettings) {
	if s == nil {
		return
	}
	s.GlobalUntilRFC3339 = strings.TrimSpace(s.GlobalUntilRFC3339)
	s.GlobalReason = strings.TrimSpace(s.GlobalReason)
	if s.Entries == nil {
		s.Entries = []AlertSilenceEntry{}
	}
	for i := range s.Entries {
		s.Entries[i].UntilRFC3339 = strings.TrimSpace(s.Entries[i].UntilRFC3339)
		s.Entries[i].Reason = strings.TrimSpace(s.Entries[i].Reason)
	}
}

func NormalizeAlertRuntimeSettings(cfg *AlertRuntimeSettings, defaultLockKey string, defaultLockTTLSeconds int) {
	if cfg == nil {
		return
	}
	if cfg.EvaluationIntervalSeconds <= 0 {
		cfg.EvaluationIntervalSeconds = 60
	}
	NormalizeDistributedLockSettings(&cfg.DistributedLock, defaultLockKey, defaultLockTTLSeconds)
	NormalizeAlertSilencingSettings(&cfg.Silencing)
}

func ValidateDistributedLockSettings(s DistributedLockSettings) error {
	if strings.TrimSpace(s.Key) == "" {
		return errors.New("distributed_lock.key is required")
	}
	if s.TTLSeconds <= 0 || s.TTLSeconds > int((24*time.Hour).Seconds()) {
		return errors.New("distributed_lock.ttl_seconds must be between 1 and 86400")
	}
	return nil
}

func ValidateAlertSilencingSettings(s AlertSilencingSettings) error {
	parse := func(raw string) error {
		if strings.TrimSpace(raw) == "" {
			return nil
		}
		if _, err := time.Parse(time.RFC3339, raw); err != nil {
			return errors.New("silencing time must be RFC3339")
		}
		return nil
	}

	if err := parse(s.GlobalUntilRFC3339); err != nil {
		return err
	}
	for _, entry := range s.Entries {
		if strings.TrimSpace(entry.UntilRFC3339) == "" {
			return errors.New("silencing.entries.until_rfc3339 is required")
		}
		if _, err := time.Parse(time.RFC3339, entry.UntilRFC3339); err != nil {
			return errors.New("silencing.entries.until_rfc3339 must be RFC3339")
		}
	}
	return nil
}

func ValidateAlertRuntimeSettings(cfg *AlertRuntimeSettings) error {
	if cfg == nil {
		return errors.New("invalid config")
	}
	if cfg.EvaluationIntervalSeconds < 1 || cfg.EvaluationIntervalSeconds > int((24*time.Hour).Seconds()) {
		return errors.New("evaluation_interval_seconds must be between 1 and 86400")
	}
	if cfg.DistributedLock.Enabled {
		if err := ValidateDistributedLockSettings(cfg.DistributedLock); err != nil {
			return err
		}
	}
	if cfg.Silencing.Enabled {
		if err := ValidateAlertSilencingSettings(cfg.Silencing); err != nil {
			return err
		}
	}
	return nil
}

func IsAlertRuntimeSilenced(now time.Time, target AlertSilenceTarget, silencing AlertSilencingSettings) bool {
	if !silencing.Enabled {
		return false
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if strings.TrimSpace(silencing.GlobalUntilRFC3339) != "" {
		if t, err := time.Parse(time.RFC3339, strings.TrimSpace(silencing.GlobalUntilRFC3339)); err == nil {
			if now.Before(t) {
				return true
			}
		}
	}

	for _, entry := range silencing.Entries {
		untilRaw := strings.TrimSpace(entry.UntilRFC3339)
		if untilRaw == "" {
			continue
		}
		until, err := time.Parse(time.RFC3339, untilRaw)
		if err != nil {
			continue
		}
		if now.After(until) {
			continue
		}
		if entry.RuleID != nil && target.RuleID > 0 && *entry.RuleID != target.RuleID {
			continue
		}
		if len(entry.Severities) > 0 && !alertSilenceSeverityMatches(entry.Severities, target) {
			continue
		}
		return true
	}

	return false
}

func alertSilenceSeverityMatches(severities []string, target AlertSilenceTarget) bool {
	eventSeverity := strings.TrimSpace(target.EventSeverity)
	ruleSeverity := strings.TrimSpace(target.RuleSeverity)
	for _, severity := range severities {
		normalized := strings.TrimSpace(severity)
		if strings.EqualFold(normalized, eventSeverity) || strings.EqualFold(normalized, ruleSeverity) {
			return true
		}
	}
	return false
}

func ShouldSendAlertEmailByMinSeverity(minSeverity string, ruleSeverity string) bool {
	minSeverity = strings.ToLower(strings.TrimSpace(minSeverity))
	if minSeverity == "" {
		return true
	}

	eventLevel := AlertEmailSeverityForOps(ruleSeverity)
	minLevel := strings.ToLower(minSeverity)
	return alertEmailSeverityRank(eventLevel) >= alertEmailSeverityRank(minLevel)
}

func AlertEmailSeverityForOps(severity string) string {
	switch strings.ToUpper(strings.TrimSpace(severity)) {
	case "P0":
		return "critical"
	case "P1":
		return "warning"
	default:
		return "info"
	}
}

func alertEmailSeverityRank(level string) int {
	switch level {
	case "critical":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}
