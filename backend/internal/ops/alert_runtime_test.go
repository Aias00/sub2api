package ops

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultAlertRuntimeSettings(t *testing.T) {
	cfg := DefaultAlertRuntimeSettings("ops:leader", 30)

	if cfg.EvaluationIntervalSeconds != 60 {
		t.Fatalf("expected default evaluation interval 60, got %d", cfg.EvaluationIntervalSeconds)
	}
	if !cfg.DistributedLock.Enabled {
		t.Fatalf("expected distributed lock enabled by default")
	}
	if cfg.DistributedLock.Key != "ops:leader" {
		t.Fatalf("expected lock key from caller, got %q", cfg.DistributedLock.Key)
	}
	if cfg.DistributedLock.TTLSeconds != 30 {
		t.Fatalf("expected lock ttl from caller, got %d", cfg.DistributedLock.TTLSeconds)
	}
	if cfg.Silencing.Entries == nil || len(cfg.Silencing.Entries) != 0 {
		t.Fatalf("expected empty silencing entries slice, got %#v", cfg.Silencing.Entries)
	}
}

func TestNormalizeAlertRuntimeSettings(t *testing.T) {
	cfg := &AlertRuntimeSettings{
		EvaluationIntervalSeconds: 0,
		DistributedLock: DistributedLockSettings{
			Key:        "  ",
			TTLSeconds: 0,
		},
		Silencing: AlertSilencingSettings{
			GlobalUntilRFC3339: " 2026-01-02T03:04:05Z ",
			GlobalReason:       " maintenance ",
			Entries: []AlertSilenceEntry{
				{UntilRFC3339: " 2026-01-03T03:04:05Z ", Reason: " deploy "},
			},
		},
	}

	NormalizeAlertRuntimeSettings(cfg, "default-lock", 45)

	if cfg.EvaluationIntervalSeconds != 60 {
		t.Fatalf("expected default interval, got %d", cfg.EvaluationIntervalSeconds)
	}
	if cfg.DistributedLock.Key != "default-lock" {
		t.Fatalf("expected default lock key, got %q", cfg.DistributedLock.Key)
	}
	if cfg.DistributedLock.TTLSeconds != 45 {
		t.Fatalf("expected default lock ttl, got %d", cfg.DistributedLock.TTLSeconds)
	}
	if cfg.Silencing.GlobalUntilRFC3339 != "2026-01-02T03:04:05Z" {
		t.Fatalf("expected trimmed global until, got %q", cfg.Silencing.GlobalUntilRFC3339)
	}
	if cfg.Silencing.GlobalReason != "maintenance" {
		t.Fatalf("expected trimmed global reason, got %q", cfg.Silencing.GlobalReason)
	}
	if cfg.Silencing.Entries[0].UntilRFC3339 != "2026-01-03T03:04:05Z" {
		t.Fatalf("expected trimmed entry until, got %q", cfg.Silencing.Entries[0].UntilRFC3339)
	}
	if cfg.Silencing.Entries[0].Reason != "deploy" {
		t.Fatalf("expected trimmed entry reason, got %q", cfg.Silencing.Entries[0].Reason)
	}
}

func TestNormalizeAlertSilencingSettingsInitializesEntries(t *testing.T) {
	cfg := &AlertSilencingSettings{}

	NormalizeAlertSilencingSettings(cfg)

	if cfg.Entries == nil {
		t.Fatalf("expected entries to be initialized")
	}
}

func TestValidateAlertRuntimeSettings(t *testing.T) {
	valid := &AlertRuntimeSettings{
		EvaluationIntervalSeconds: 60,
		DistributedLock: DistributedLockSettings{
			Enabled:    true,
			Key:        "ops:leader",
			TTLSeconds: 30,
		},
		Silencing: AlertSilencingSettings{
			Enabled: true,
			Entries: []AlertSilenceEntry{
				{UntilRFC3339: "2026-01-03T03:04:05Z"},
			},
		},
	}

	if err := ValidateAlertRuntimeSettings(valid); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateAlertRuntimeSettingsRejectsInvalidInterval(t *testing.T) {
	err := ValidateAlertRuntimeSettings(&AlertRuntimeSettings{
		EvaluationIntervalSeconds: 0,
	})

	if err == nil || !strings.Contains(err.Error(), "evaluation_interval_seconds") {
		t.Fatalf("expected interval error, got %v", err)
	}
}

func TestValidateDistributedLockSettings(t *testing.T) {
	if err := ValidateDistributedLockSettings(DistributedLockSettings{Key: " ", TTLSeconds: 30}); err == nil {
		t.Fatalf("expected empty key error")
	}
	if err := ValidateDistributedLockSettings(DistributedLockSettings{Key: "ops", TTLSeconds: 86401}); err == nil {
		t.Fatalf("expected ttl range error")
	}
}

func TestValidateAlertSilencingSettings(t *testing.T) {
	if err := ValidateAlertSilencingSettings(AlertSilencingSettings{GlobalUntilRFC3339: "not-time"}); err == nil {
		t.Fatalf("expected invalid global time error")
	}
	if err := ValidateAlertSilencingSettings(AlertSilencingSettings{
		Entries: []AlertSilenceEntry{{Reason: "maintenance"}},
	}); err == nil {
		t.Fatalf("expected missing entry until error")
	}
	if err := ValidateAlertSilencingSettings(AlertSilencingSettings{
		Entries: []AlertSilenceEntry{{UntilRFC3339: "not-time"}},
	}); err == nil {
		t.Fatalf("expected invalid entry until error")
	}
}

func TestIsAlertRuntimeSilenced(t *testing.T) {
	now := mustParseAlertRuntimeTime(t, "2026-01-02T03:04:05Z")
	ruleID := int64(42)

	tests := []struct {
		name      string
		target    AlertSilenceTarget
		silencing AlertSilencingSettings
		want      bool
	}{
		{
			name: "disabled",
			silencing: AlertSilencingSettings{
				Enabled:            false,
				GlobalUntilRFC3339: "2026-01-02T04:04:05Z",
			},
			want: false,
		},
		{
			name: "global until in future",
			silencing: AlertSilencingSettings{
				Enabled:            true,
				GlobalUntilRFC3339: "2026-01-02T04:04:05Z",
			},
			want: true,
		},
		{
			name:   "matching rule id",
			target: AlertSilenceTarget{RuleID: ruleID},
			silencing: AlertSilencingSettings{
				Enabled: true,
				Entries: []AlertSilenceEntry{
					{RuleID: &ruleID, UntilRFC3339: "2026-01-02T04:04:05Z"},
				},
			},
			want: true,
		},
		{
			name:   "different rule id",
			target: AlertSilenceTarget{RuleID: 7},
			silencing: AlertSilencingSettings{
				Enabled: true,
				Entries: []AlertSilenceEntry{
					{RuleID: &ruleID, UntilRFC3339: "2026-01-02T04:04:05Z"},
				},
			},
			want: false,
		},
		{
			name:   "matching event severity",
			target: AlertSilenceTarget{RuleID: ruleID, EventSeverity: "p1", RuleSeverity: "P2"},
			silencing: AlertSilencingSettings{
				Enabled: true,
				Entries: []AlertSilenceEntry{
					{UntilRFC3339: "2026-01-02T04:04:05Z", Severities: []string{" P1 "}},
				},
			},
			want: true,
		},
		{
			name:   "matching rule severity",
			target: AlertSilenceTarget{RuleID: ruleID, EventSeverity: "P3", RuleSeverity: "p0"},
			silencing: AlertSilencingSettings{
				Enabled: true,
				Entries: []AlertSilenceEntry{
					{UntilRFC3339: "2026-01-02T04:04:05Z", Severities: []string{"P0"}},
				},
			},
			want: true,
		},
		{
			name:   "expired entry",
			target: AlertSilenceTarget{RuleID: ruleID, EventSeverity: "P1"},
			silencing: AlertSilencingSettings{
				Enabled: true,
				Entries: []AlertSilenceEntry{
					{UntilRFC3339: "2026-01-02T02:04:05Z", Severities: []string{"P1"}},
				},
			},
			want: false,
		},
		{
			name:   "invalid entry time skipped",
			target: AlertSilenceTarget{RuleID: ruleID, EventSeverity: "P1"},
			silencing: AlertSilencingSettings{
				Enabled: true,
				Entries: []AlertSilenceEntry{
					{UntilRFC3339: "not-time", Severities: []string{"P1"}},
				},
			},
			want: false,
		},
		{
			name:   "severity mismatch",
			target: AlertSilenceTarget{RuleID: ruleID, EventSeverity: "P3", RuleSeverity: "P2"},
			silencing: AlertSilencingSettings{
				Enabled: true,
				Entries: []AlertSilenceEntry{
					{UntilRFC3339: "2026-01-02T04:04:05Z", Severities: []string{"P1"}},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAlertRuntimeSilenced(now, tt.target, tt.silencing)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func mustParseAlertRuntimeTime(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed
}

func TestAlertEmailSeverityForOps(t *testing.T) {
	tests := []struct {
		severity string
		want     string
	}{
		{severity: "P0", want: "critical"},
		{severity: " p1 ", want: "warning"},
		{severity: "P2", want: "info"},
		{severity: "unknown", want: "info"},
		{severity: "", want: "info"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			got := AlertEmailSeverityForOps(tt.severity)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestShouldSendAlertEmailByMinSeverity(t *testing.T) {
	tests := []struct {
		name         string
		minSeverity  string
		ruleSeverity string
		want         bool
	}{
		{name: "empty min allows all", minSeverity: "", ruleSeverity: "P2", want: true},
		{name: "critical allows p0", minSeverity: "critical", ruleSeverity: "P0", want: true},
		{name: "critical blocks p1", minSeverity: "critical", ruleSeverity: "P1", want: false},
		{name: "warning allows p1", minSeverity: " warning ", ruleSeverity: " p1 ", want: true},
		{name: "warning blocks info", minSeverity: "warning", ruleSeverity: "P2", want: false},
		{name: "info allows default info", minSeverity: "info", ruleSeverity: "P3", want: true},
		{name: "unknown min preserves previous permissive rank behavior", minSeverity: "unknown", ruleSeverity: "P3", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSendAlertEmailByMinSeverity(tt.minSeverity, tt.ruleSeverity)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
