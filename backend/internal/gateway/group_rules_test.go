package gateway

import "testing"

func TestGroupBasicRules(t *testing.T) {
	if !IsGroupActive("active") {
		t.Fatal("active status should be active")
	}
	if IsGroupActive("disabled") {
		t.Fatal("disabled status should not be active")
	}
	if !IsGroupSubscriptionType("subscription") {
		t.Fatal("subscription type should match")
	}
	limit := 1.0
	if !HasPositiveLimit(&limit) {
		t.Fatal("positive limit should be present")
	}
	zero := 0.0
	if HasPositiveLimit(&zero) {
		t.Fatal("zero limit should not be present")
	}
}

func TestSelectGroupImagePrice(t *testing.T) {
	one, two, four := 1.0, 2.0, 4.0
	if got := SelectGroupImagePrice("1K", &one, &two, &four); got == nil || *got != 1 {
		t.Fatalf("1K price = %v", got)
	}
	if got := SelectGroupImagePrice("unknown", &one, &two, &four); got == nil || *got != 2 {
		t.Fatalf("default price = %v", got)
	}
}

func TestGroupContextValidity(t *testing.T) {
	if !IsGroupContextValid(GroupContextValidityInput{ID: 1, Hydrated: true, Platform: "openai", Status: "active"}) {
		t.Fatal("complete context should be valid")
	}
	if IsGroupContextValid(GroupContextValidityInput{ID: 1, Platform: "openai", Status: "active"}) {
		t.Fatal("non-hydrated context should be invalid")
	}
}

func TestRoutingAccountIDs(t *testing.T) {
	routing := map[string][]int64{
		"gpt-5.4":  {1, 2},
		"claude-*": {3},
		"empty-*":  {},
	}
	if got := RoutingAccountIDs(true, routing, "gpt-5.4"); len(got) != 2 || got[0] != 1 {
		t.Fatalf("exact routing = %v", got)
	}
	if got := RoutingAccountIDs(true, routing, "claude-sonnet"); len(got) != 1 || got[0] != 3 {
		t.Fatalf("wildcard routing = %v", got)
	}
	if got := RoutingAccountIDs(false, routing, "gpt-5.4"); got != nil {
		t.Fatalf("disabled routing = %v", got)
	}
	if got := RoutingAccountIDs(true, routing, "missing"); got != nil {
		t.Fatalf("missing routing = %v", got)
	}
}
