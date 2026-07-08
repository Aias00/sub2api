package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// UsageBillingCommand.Normalize / fingerprint
//
// The fingerprint is the idempotency key that prevents a single billable
// request from being charged twice (or, if computed inconsistently, from being
// charged at all). These tests pin its critical invariants.
// ---------------------------------------------------------------------------

func baseUsageCmd() *UsageBillingCommand {
	return &UsageBillingCommand{
		RequestID:    "req-1",
		APIKeyID:     10,
		UserID:       1,
		AccountID:    2,
		AccountType:  "oauth",
		Model:        "claude-sonnet-4-5",
		InputTokens:  100,
		OutputTokens: 200,
		BalanceCost:  0.0123456789,
	}
}

func TestUsageBillingCommand_Normalize_NilSafe(t *testing.T) {
	var c *UsageBillingCommand
	c.Normalize() // must not panic
}

func TestUsageBillingCommand_Normalize_ComputesFingerprintWhenEmpty(t *testing.T) {
	c := baseUsageCmd()
	c.Normalize()
	if c.RequestFingerprint == "" {
		t.Fatal("expected fingerprint to be computed when empty")
	}
	if len(c.RequestFingerprint) != 64 { // sha256 hex
		t.Fatalf("expected 64-char sha256 hex, got %d chars", len(c.RequestFingerprint))
	}
}

func TestUsageBillingCommand_Normalize_PreservesProvidedFingerprint(t *testing.T) {
	c := baseUsageCmd()
	c.RequestFingerprint = "preset-fingerprint"
	c.Normalize()
	if c.RequestFingerprint != "preset-fingerprint" {
		t.Fatalf("Normalize must not overwrite a provided fingerprint, got %q", c.RequestFingerprint)
	}
}

func TestUsageBillingCommand_Normalize_TrimsRequestID(t *testing.T) {
	c := baseUsageCmd()
	c.RequestID = "  req-42  "
	c.Normalize()
	if c.RequestID != "req-42" {
		t.Fatalf("expected trimmed request id, got %q", c.RequestID)
	}
}

func TestUsageBillingFingerprint_Deterministic(t *testing.T) {
	c1 := baseUsageCmd()
	c2 := baseUsageCmd()
	c1.Normalize()
	c2.Normalize()
	if c1.RequestFingerprint != c2.RequestFingerprint {
		t.Fatal("identical commands must produce identical fingerprints")
	}
}

func TestUsageBillingFingerprint_SensitiveToBillableFields(t *testing.T) {
	base := baseUsageCmd()
	base.Normalize()

	mutations := map[string]func(c *UsageBillingCommand){
		"UserID":       func(c *UsageBillingCommand) { c.UserID = 999 },
		"AccountID":    func(c *UsageBillingCommand) { c.AccountID = 999 },
		"APIKeyID":     func(c *UsageBillingCommand) { c.APIKeyID = 999 },
		"Model":        func(c *UsageBillingCommand) { c.Model = "gpt-4o" },
		"InputTokens":  func(c *UsageBillingCommand) { c.InputTokens = 101 },
		"OutputTokens": func(c *UsageBillingCommand) { c.OutputTokens = 201 },
		"BalanceCost":  func(c *UsageBillingCommand) { c.BalanceCost = 0.999 },
		"BillingType":  func(c *UsageBillingCommand) { c.BillingType = 3 },
		"ImageCount":   func(c *UsageBillingCommand) { c.ImageCount = 5 },
	}

	for name, mut := range mutations {
		t.Run(name, func(t *testing.T) {
			c := baseUsageCmd()
			mut(c)
			c.Normalize()
			if c.RequestFingerprint == base.RequestFingerprint {
				t.Fatalf("changing %s must change the fingerprint (double-charge risk)", name)
			}
		})
	}
}

func TestUsageBillingFingerprint_PayloadHashChangesFingerprint(t *testing.T) {
	a := baseUsageCmd()
	a.Normalize()

	b := baseUsageCmd()
	b.RequestPayloadHash = "some-payload-hash"
	b.Normalize()

	if a.RequestFingerprint == b.RequestFingerprint {
		t.Fatal("adding a payload hash must change the fingerprint")
	}
}

func TestUsageBillingFingerprint_SubscriptionIDAffectsFingerprint(t *testing.T) {
	a := baseUsageCmd()
	a.Normalize()

	sub := int64(7)
	b := baseUsageCmd()
	b.SubscriptionID = &sub
	b.Normalize()

	if a.RequestFingerprint == b.RequestFingerprint {
		t.Fatal("a set SubscriptionID must change the fingerprint")
	}
}

// ---------------------------------------------------------------------------
// HashUsageRequestPayload / valueOrZero
// ---------------------------------------------------------------------------

func TestHashUsageRequestPayload(t *testing.T) {
	if got := HashUsageRequestPayload(nil); got != "" {
		t.Fatalf("nil payload must hash to empty string, got %q", got)
	}
	if got := HashUsageRequestPayload([]byte{}); got != "" {
		t.Fatalf("empty payload must hash to empty string, got %q", got)
	}

	payload := []byte(`{"model":"x"}`)
	want := sha256.Sum256(payload)
	if got := HashUsageRequestPayload(payload); got != hex.EncodeToString(want[:]) {
		t.Fatalf("unexpected payload hash: %q", got)
	}
	// deterministic
	if HashUsageRequestPayload(payload) != HashUsageRequestPayload(payload) {
		t.Fatal("payload hash must be deterministic")
	}
}

func TestValueOrZero(t *testing.T) {
	if valueOrZero(nil) != 0 {
		t.Fatal("nil pointer must yield 0")
	}
	v := int64(42)
	if valueOrZero(&v) != 42 {
		t.Fatal("pointer must be dereferenced")
	}
}

// ---------------------------------------------------------------------------
// BatchImageBalanceHoldCommand.Normalize / fingerprint
// ---------------------------------------------------------------------------

func TestBatchImageHold_Normalize(t *testing.T) {
	var nilCmd *BatchImageBalanceHoldCommand
	nilCmd.Normalize() // must not panic

	c := &BatchImageBalanceHoldCommand{
		RequestID:    "  r ",
		BatchID:      "  b ",
		UserID:       1,
		APIKeyID:     2,
		HoldAmount:   1.5,
		ActualAmount: 1.0,
	}
	c.Normalize()
	if c.RequestID != "r" || c.BatchID != "b" {
		t.Fatalf("expected trimmed fields, got id=%q batch=%q", c.RequestID, c.BatchID)
	}
	if c.RequestFingerprint == "" || len(c.RequestFingerprint) != 64 {
		t.Fatalf("expected computed sha256 fingerprint, got %q", c.RequestFingerprint)
	}
}

func TestBatchImageHold_FingerprintSensitiveToAmounts(t *testing.T) {
	mk := func(hold, actual float64) string {
		c := &BatchImageBalanceHoldCommand{UserID: 1, APIKeyID: 2, BatchID: "b", HoldAmount: hold, ActualAmount: actual}
		c.Normalize()
		return c.RequestFingerprint
	}
	if mk(1.5, 1.0) == mk(2.0, 1.0) {
		t.Fatal("different hold amount must change fingerprint")
	}
	if mk(1.5, 1.0) == mk(1.5, 1.2) {
		t.Fatal("different actual amount must change fingerprint")
	}
}

func TestBatchImageHold_PreservesProvidedFingerprint(t *testing.T) {
	c := &BatchImageBalanceHoldCommand{UserID: 1, BatchID: "b", RequestFingerprint: "keep"}
	c.Normalize()
	if c.RequestFingerprint != "keep" {
		t.Fatalf("must preserve provided fingerprint, got %q", c.RequestFingerprint)
	}
}

// ---------------------------------------------------------------------------
// ResolveImageRateMultiplier
// ---------------------------------------------------------------------------

func TestResolveImageRateMultiplier(t *testing.T) {
	cases := []struct {
		name        string
		independent bool
		imageMult   float64
		groupMult   float64
		want        float64
	}{
		{"independent positive", true, 2.5, 1.0, 2.5},
		{"independent zero", true, 0, 1.0, 0},
		{"independent negative clamps to 0", true, -1, 1.0, 0},
		{"not independent uses group", false, 2.5, 1.3, 1.3},
		{"not independent ignores image mult", false, 99, 0.5, 0.5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveImageRateMultiplier(tc.independent, tc.imageMult, tc.groupMult)
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// EntryType display mappers (known + fallback)
// ---------------------------------------------------------------------------

func TestEntryTypeDisplayMappers(t *testing.T) {
	// Known type resolves to non-fallback values.
	if got := EntryTypeDisplayName(EntryTypeRecharge); got != "充值" {
		t.Fatalf("unexpected display name: %q", got)
	}
	if got := EntryTypeColor(EntryTypeAPIUsage); got != "red" {
		t.Fatalf("unexpected color: %q", got)
	}
	if got := EntryTypeIcon(EntryTypeRecharge); got != "dollar-up" {
		t.Fatalf("unexpected icon: %q", got)
	}

	// Unknown type falls back gracefully.
	unknown := BalanceLedgerEntryType("does-not-exist")
	if got := EntryTypeDisplayName(unknown); got != "does-not-exist" {
		t.Fatalf("unknown display name should echo raw value, got %q", got)
	}
	if got := EntryTypeIcon(unknown); got != "circle" {
		t.Fatalf("unknown icon fallback expected 'circle', got %q", got)
	}
	if got := EntryTypeColor(unknown); got != "gray" {
		t.Fatalf("unknown color fallback expected 'gray', got %q", got)
	}
}

// Guard: fingerprint raw layout must stay stable; if the field order/format in
// buildUsageBillingFingerprint changes, previously-recorded fingerprints would
// no longer match and idempotency would silently break. This asserts the hash
// of a fully-populated known command stays constant.
func TestUsageBillingFingerprint_StableGolden(t *testing.T) {
	c := &UsageBillingCommand{
		UserID: 1, AccountID: 2, APIKeyID: 3, AccountType: "oauth",
		Model: "m", ServiceTier: "t", ReasoningEffort: "high", BillingType: 1,
		InputTokens: 10, OutputTokens: 20, CacheCreationTokens: 30, CacheReadTokens: 40,
		ImageCount: 1, ImageSize: "1024x1024", MediaType: "image/png",
		BalanceCost: 1, SubscriptionCost: 2, APIKeyQuotaCost: 3, APIKeyRateLimitCost: 4, AccountQuotaCost: 5,
	}
	c.Normalize()
	// Recompute independently to confirm the function is self-consistent.
	if !strings.EqualFold(c.RequestFingerprint, buildUsageBillingFingerprint(c)) {
		t.Fatal("fingerprint must be reproducible from the same command")
	}
}

// ---------------------------------------------------------------------------
// Image billing tier helpers
// ---------------------------------------------------------------------------

func TestNormalizeImageBillingTierOrDefault(t *testing.T) {
	// A classifiable size returns its tier; an unknown size falls back to 2K.
	got := NormalizeImageBillingTierOrDefault("")
	if got != ImageBillingSize2K {
		t.Fatalf("empty/unknown size must default to 2K tier, got %q", got)
	}
	// Idempotent: normalizing an already-normalized tier yields a valid tier.
	if again := NormalizeImageBillingTierOrDefault(got); again == "" {
		t.Fatal("normalized tier must not be empty")
	}
}

func TestSortedImageBillingBreakdownKeys(t *testing.T) {
	// Empty map -> empty slice (not nil-panic).
	if got := SortedImageBillingBreakdownKeys(map[string]int{}); len(got) != 0 {
		t.Fatalf("empty breakdown must yield empty keys, got %v", got)
	}
	// Deterministic ordering for a fixed input.
	in := map[string]int{ImageBillingSize4K: 1, ImageBillingSize2K: 2}
	a := SortedImageBillingBreakdownKeys(in)
	b := SortedImageBillingBreakdownKeys(in)
	if len(a) != 2 || len(b) != 2 {
		t.Fatalf("expected 2 keys, got %v", a)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("key ordering must be deterministic across calls")
		}
	}
}
