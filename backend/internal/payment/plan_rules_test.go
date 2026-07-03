package payment

import (
	"testing"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
)

func TestValidatePlanRequired(t *testing.T) {
	t.Parallel()

	if err := ValidatePlanRequired("Pro", 1, 9.99, 30, "days", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := ValidatePlanRequired("", 0, 0, 0, "", nil)
	if err == nil {
		t.Fatal("expected invalid plan to fail")
	}
	if appErr := infraerrors.FromError(err); appErr.Reason != "PLAN_NAME_REQUIRED" {
		t.Fatalf("reason = %q, want PLAN_NAME_REQUIRED", appErr.Reason)
	}

	negativeOriginalPrice := -1.0
	err = ValidatePlanRequired("Pro", 1, 9.99, 30, "days", &negativeOriginalPrice)
	if err == nil {
		t.Fatal("expected negative original price to fail")
	}
	if appErr := infraerrors.FromError(err); appErr.Reason != "PLAN_ORIGINAL_PRICE_INVALID" {
		t.Fatalf("reason = %q, want PLAN_ORIGINAL_PRICE_INVALID", appErr.Reason)
	}
}

func TestValidatePlanPatch(t *testing.T) {
	t.Parallel()

	if err := ValidatePlanPatch(PlanPatchInput{}); err != nil {
		t.Fatalf("unexpected nil patch error: %v", err)
	}

	name := " "
	err := ValidatePlanPatch(PlanPatchInput{Name: &name})
	if err == nil {
		t.Fatal("expected blank name patch to fail")
	}
	if appErr := infraerrors.FromError(err); appErr.Reason != "PLAN_NAME_REQUIRED" {
		t.Fatalf("reason = %q, want PLAN_NAME_REQUIRED", appErr.Reason)
	}

	price := 0.0
	err = ValidatePlanPatch(PlanPatchInput{Price: &price})
	if err == nil {
		t.Fatal("expected zero price patch to fail")
	}
	if appErr := infraerrors.FromError(err); appErr.Reason != "PLAN_PRICE_INVALID" {
		t.Fatalf("reason = %q, want PLAN_PRICE_INVALID", appErr.Reason)
	}

	originalPrice := 0.0
	if err := ValidatePlanPatch(PlanPatchInput{OriginalPrice: &originalPrice}); err != nil {
		t.Fatalf("unexpected zero original price error: %v", err)
	}
}

func TestComputeSubscriptionValidityDays(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		days int
		unit string
		want int
	}{
		{name: "days", days: 1, unit: "days", want: 1},
		{name: "week", days: 1, unit: "week", want: 7},
		{name: "weeks", days: 2, unit: "weeks", want: 14},
		{name: "month", days: 1, unit: "month", want: 30},
		{name: "months", days: 1, unit: "months", want: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ComputeSubscriptionValidityDays(tt.days, tt.unit); got != tt.want {
				t.Fatalf("ComputeSubscriptionValidityDays(%d, %q) = %d, want %d", tt.days, tt.unit, got, tt.want)
			}
		})
	}
}
