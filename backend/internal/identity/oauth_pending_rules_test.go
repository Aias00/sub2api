package identity

import "testing"

func TestIsSupportedPendingOAuthProvider(t *testing.T) {
	if !IsSupportedPendingOAuthProvider(" GitHub ") {
		t.Fatal("github should be supported")
	}
	if !IsSupportedPendingOAuthProvider("google") {
		t.Fatal("google should be supported")
	}
	if IsSupportedPendingOAuthProvider("wechat") {
		t.Fatal("wechat pending provider should not be supported")
	}
}

func TestWeChatOAuthIdentityProviderKeys(t *testing.T) {
	got := WeChatOAuthIdentityProviderKeys(" wechat-main ")
	if len(got) != 2 || got[0] != WeChatOAuthProviderKey || got[1] != WeChatOAuthLegacyProviderKey {
		t.Fatalf("WeChatOAuthIdentityProviderKeys() = %#v", got)
	}

	got = WeChatOAuthIdentityProviderKeys("custom")
	if len(got) != 1 || got[0] != "custom" {
		t.Fatalf("WeChatOAuthIdentityProviderKeys(custom) = %#v", got)
	}
}

func TestShouldBindPendingOAuthIdentity(t *testing.T) {
	tests := []struct {
		intent           string
		adoptDisplayName bool
		adoptAvatar      bool
		want             bool
	}{
		{intent: " login ", want: true},
		{intent: "bind_current_user", want: true},
		{intent: "adopt_existing_user_by_email", want: true},
		{intent: "choose_account_action_required", want: false},
		{intent: "choose_account_action_required", adoptDisplayName: true, want: true},
		{intent: "choose_account_action_required", adoptAvatar: true, want: true},
	}

	for _, tt := range tests {
		got := ShouldBindPendingOAuthIdentity(tt.intent, tt.adoptDisplayName, tt.adoptAvatar)
		if got != tt.want {
			t.Fatalf("ShouldBindPendingOAuthIdentity(%q, %v, %v) = %v, want %v", tt.intent, tt.adoptDisplayName, tt.adoptAvatar, got, tt.want)
		}
	}
}
