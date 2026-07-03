package ops

import (
	"errors"
	"testing"
)

func TestParsePositiveID(t *testing.T) {
	got, err := ParsePositiveID(" 42 ", "Invalid id")
	if err != nil {
		t.Fatalf("ParsePositiveID() error = %v", err)
	}
	if got != 42 {
		t.Fatalf("id = %d, want 42", got)
	}

	for _, raw := range []string{"", "0", "-1", "bad"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParsePositiveID(raw, "Invalid id"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseTruthyFlag(t *testing.T) {
	for _, raw := range []string{"1", " true ", "YES"} {
		t.Run(raw, func(t *testing.T) {
			if !ParseTruthyFlag(raw) {
				t.Fatal("expected true")
			}
		})
	}
	for _, raw := range []string{"", "0", "false", "no", "maybe"} {
		t.Run(raw, func(t *testing.T) {
			if ParseTruthyFlag(raw) {
				t.Fatal("expected false")
			}
		})
	}
}

func TestCapPageSize(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		max      int
		want     int
	}{
		{name: "below max", pageSize: 50, max: 100, want: 50},
		{name: "at max", pageSize: 100, max: 100, want: 100},
		{name: "above max", pageSize: 500, max: 100, want: 100},
		{name: "disabled max", pageSize: 500, max: 0, want: 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CapPageSize(tt.pageSize, tt.max); got != tt.want {
				t.Fatalf("CapPageSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPickErrorCorrelationKey(t *testing.T) {
	got := PickErrorCorrelationKey(" request-1 ", " client-1 ")
	if got.RequestID != "request-1" || got.ClientRequestID != "" {
		t.Fatalf("key = %+v, want request id only", got)
	}

	got = PickErrorCorrelationKey(" ", " client-1 ")
	if got.RequestID != "" || got.ClientRequestID != "client-1" {
		t.Fatalf("key = %+v, want client request id", got)
	}

	got = PickErrorCorrelationKey(" ", " ")
	if got.RequestID != "" || got.ClientRequestID != "" {
		t.Fatalf("key = %+v, want empty", got)
	}
}

func TestIsInvalidInputError(t *testing.T) {
	if !IsInvalidInputError(errors.New("invalid sort")) {
		t.Fatal("expected invalid error")
	}
	if !IsInvalidInputError(errors.New("INVALID kind")) {
		t.Fatal("expected case-insensitive invalid error")
	}
	if IsInvalidInputError(errors.New("database failed")) {
		t.Fatal("expected non-invalid error")
	}
	if IsInvalidInputError(nil) {
		t.Fatal("expected nil error to be false")
	}
}
