package ops

import (
	"strings"
	"testing"
)

func TestDefaultRuntimeLogConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultRuntimeLogConfig()
	if cfg.Level != "info" || cfg.StacktraceLevel != "error" || cfg.SamplingInitial != 100 || cfg.SamplingNext != 100 || cfg.RetentionDays != 30 {
		t.Fatalf("default runtime log config = %+v", cfg)
	}
	if !cfg.Caller {
		t.Fatal("caller should default to true")
	}
}

func TestNormalizeRuntimeLogConfig(t *testing.T) {
	t.Parallel()

	defaults := &RuntimeLogConfig{
		Level:           "info",
		SamplingInitial: 100,
		SamplingNext:    100,
		StacktraceLevel: "error",
		RetentionDays:   30,
	}
	cfg := &RuntimeLogConfig{
		Level:           " WARN ",
		SamplingInitial: 0,
		SamplingNext:    -1,
		StacktraceLevel: "",
		RetentionDays:   0,
	}

	NormalizeRuntimeLogConfig(cfg, defaults)

	if cfg.Level != "warn" {
		t.Fatalf("level = %q, want warn", cfg.Level)
	}
	if cfg.SamplingInitial != 100 || cfg.SamplingNext != 100 || cfg.StacktraceLevel != "error" || cfg.RetentionDays != 30 {
		t.Fatalf("normalized config = %+v", cfg)
	}
}

func TestValidateRuntimeLogConfig(t *testing.T) {
	t.Parallel()

	valid := &RuntimeLogConfig{
		Level:           "info",
		StacktraceLevel: "error",
		SamplingInitial: 1,
		SamplingNext:    1,
		RetentionDays:   30,
	}
	if err := ValidateRuntimeLogConfig(valid); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		cfg  *RuntimeLogConfig
		want string
	}{
		{name: "nil", cfg: nil, want: "invalid config"},
		{name: "bad level", cfg: &RuntimeLogConfig{Level: "trace", StacktraceLevel: "error", SamplingInitial: 1, SamplingNext: 1, RetentionDays: 1}, want: "level"},
		{name: "bad stack", cfg: &RuntimeLogConfig{Level: "info", StacktraceLevel: "warn", SamplingInitial: 1, SamplingNext: 1, RetentionDays: 1}, want: "stacktrace_level"},
		{name: "bad initial", cfg: &RuntimeLogConfig{Level: "info", StacktraceLevel: "error", SamplingInitial: 0, SamplingNext: 1, RetentionDays: 1}, want: "sampling_initial"},
		{name: "bad next", cfg: &RuntimeLogConfig{Level: "info", StacktraceLevel: "error", SamplingInitial: 1, SamplingNext: 0, RetentionDays: 1}, want: "sampling_thereafter"},
		{name: "bad retention", cfg: &RuntimeLogConfig{Level: "info", StacktraceLevel: "error", SamplingInitial: 1, SamplingNext: 1, RetentionDays: 0}, want: "retention_days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateRuntimeLogConfig(tt.cfg)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
