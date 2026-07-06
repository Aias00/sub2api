package identity

import "testing"

func TestUserStatusRules(t *testing.T) {
	if !IsUserAdmin("admin") {
		t.Fatal("admin role should be admin")
	}
	if IsUserAdmin("user") {
		t.Fatal("user role should not be admin")
	}
	if !IsUserActive("active") {
		t.Fatal("active status should be active")
	}
	if IsUserActive("disabled") {
		t.Fatal("disabled status should not be active")
	}
}

func TestCanUserBindGroup(t *testing.T) {
	if !CanUserBindGroup(nil, 42, false) {
		t.Fatal("non-exclusive groups should be bindable")
	}
	if !CanUserBindGroup([]int64{1, 42}, 42, true) {
		t.Fatal("exclusive group should be bindable when allowed")
	}
	if CanUserBindGroup([]int64{1, 2}, 42, true) {
		t.Fatal("exclusive group should not be bindable when missing")
	}
}

func TestPasswordRules(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" || hash == "secret" {
		t.Fatalf("HashPassword() produced invalid hash %q", hash)
	}
	if !CheckPassword("secret", hash) {
		t.Fatal("CheckPassword() should accept matching password")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("CheckPassword() should reject non-matching password")
	}
}
