package gateway

import (
	"testing"
	"time"
)

func TestAPIKeyStatusAndQuotaRules(t *testing.T) {
	if !IsAPIKeyActive(StatusAPIKeyActive) {
		t.Fatal("active status should be active")
	}
	if IsAPIKeyActive(StatusAPIKeyDisabled) {
		t.Fatal("disabled status should not be active")
	}
	if IsAPIKeyQuotaExhausted(0, 100) {
		t.Fatal("zero quota should be unlimited")
	}
	if !IsAPIKeyQuotaExhausted(10, 10) {
		t.Fatal("quota should be exhausted at limit")
	}
	if got := APIKeyQuotaRemaining(0, 100); got != -1 {
		t.Fatalf("unlimited remaining = %v, want -1", got)
	}
	if got := APIKeyQuotaRemaining(10, 3); got != 7 {
		t.Fatalf("remaining = %v, want 7", got)
	}
	if got := APIKeyQuotaRemaining(10, 12); got != 0 {
		t.Fatalf("overused remaining = %v, want 0", got)
	}
}

func TestAPIKeyExpiryRules(t *testing.T) {
	now := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	future := now.Add(49 * time.Hour)
	past := now.Add(-time.Second)

	if IsAPIKeyExpired(nil, now) {
		t.Fatal("nil expiry should not be expired")
	}
	if IsAPIKeyExpired(&future, now) {
		t.Fatal("future expiry should not be expired")
	}
	if !IsAPIKeyExpired(&past, now) {
		t.Fatal("past expiry should be expired")
	}
	if got := APIKeyDaysUntilExpiry(nil, now); got != -1 {
		t.Fatalf("nil expiry days = %d, want -1", got)
	}
	if got := APIKeyDaysUntilExpiry(&future, now); got != 2 {
		t.Fatalf("future expiry days = %d, want 2", got)
	}
	if got := APIKeyDaysUntilExpiry(&past, now); got != 0 {
		t.Fatalf("past expiry days = %d, want 0", got)
	}
}
