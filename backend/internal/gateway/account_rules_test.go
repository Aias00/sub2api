package gateway

import "testing"

func TestAccountBasicRules(t *testing.T) {
	if !IsAccountActive("active") {
		t.Fatal("active status should be active")
	}
	if IsAccountActive("disabled") {
		t.Fatal("disabled status should not be active")
	}

	multiplier := 0.0
	if got := AccountBillingRateMultiplier(&multiplier); got != 0 {
		t.Fatalf("zero multiplier = %v, want 0", got)
	}
	negative := -1.0
	if got := AccountBillingRateMultiplier(&negative); got != 1 {
		t.Fatalf("negative multiplier = %v, want 1", got)
	}
	if got := AccountBillingRateMultiplier(nil); got != 1 {
		t.Fatalf("nil multiplier = %v, want 1", got)
	}

	load := 5
	if got := AccountEffectiveLoadFactor(&load, 2); got != 5 {
		t.Fatalf("load factor = %d, want 5", got)
	}
	zero := 0
	if got := AccountEffectiveLoadFactor(&zero, 3); got != 3 {
		t.Fatalf("fallback concurrency = %d, want 3", got)
	}
	if got := AccountEffectiveLoadFactor(nil, 0); got != 1 {
		t.Fatalf("default load factor = %d, want 1", got)
	}
}

func TestIsAccountSchedulable(t *testing.T) {
	if !IsAccountSchedulable("active", true) {
		t.Fatal("active schedulable account should be schedulable")
	}
	if IsAccountSchedulable("active", false) {
		t.Fatal("inactive schedulable flag should block")
	}
	if IsAccountSchedulable("disabled", true) {
		t.Fatal("disabled account should not be schedulable")
	}
}
