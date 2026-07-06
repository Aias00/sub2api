package identity

import "testing"

func TestNormalizePendingOAuthCompletionResponse(t *testing.T) {
	got := NormalizePendingOAuthCompletionResponse(map[string]any{
		"step":          "choice",
		"access_token":  "secret",
		"refresh_token": "refresh",
		"expires_in":    3600,
		"token_type":    "Bearer",
	})
	if got["step"] != OAuthPendingChoiceStep {
		t.Fatalf("step = %v", got["step"])
	}
	if got["adoption_required"] != true {
		t.Fatalf("adoption_required = %v", got["adoption_required"])
	}
	for _, key := range []string{"access_token", "refresh_token", "expires_in", "token_type"} {
		if _, exists := got[key]; exists {
			t.Fatalf("%s should be removed", key)
		}
	}
}

func TestApplySuggestedProfileToCompletionResponse(t *testing.T) {
	payload := map[string]any{"step": "choose"}
	ApplySuggestedProfileToCompletionResponse(payload, map[string]any{
		"suggested_display_name": " Alice ",
		"suggested_avatar_url":   "https://example.com/a.png",
	})
	if payload["suggested_display_name"] != "Alice" {
		t.Fatalf("suggested_display_name = %v", payload["suggested_display_name"])
	}
	if payload["suggested_avatar_url"] != "https://example.com/a.png" {
		t.Fatalf("suggested_avatar_url = %v", payload["suggested_avatar_url"])
	}
	if payload["adoption_required"] != true {
		t.Fatalf("adoption_required = %v", payload["adoption_required"])
	}
}

func TestPendingOAuthCompletionCanIssueTokenPair(t *testing.T) {
	userID := int64(12)
	if !PendingOAuthCompletionCanIssueTokenPair(OAuthIntentLogin, &userID, map[string]any{}) {
		t.Fatal("expected login with target user and no step to issue token pair")
	}
	if PendingOAuthCompletionCanIssueTokenPair(OAuthIntentBindCurrentUser, &userID, map[string]any{}) {
		t.Fatal("bind intent should not issue login token pair")
	}
	if PendingOAuthCompletionCanIssueTokenPair(OAuthIntentLogin, &userID, map[string]any{"error": "invitation_required"}) {
		t.Fatal("invitation-required payload should not issue token pair")
	}
	if PendingOAuthCompletionCanIssueTokenPair(OAuthIntentLogin, &userID, map[string]any{"step": "email_completion"}) {
		t.Fatal("completion step should not issue token pair")
	}
}

func TestIsPendingOAuthCompleteRegistrationSession(t *testing.T) {
	if !IsPendingOAuthCompleteRegistrationSession(OAuthIntentLogin, nil, map[string]any{}) {
		t.Fatal("expected empty login session to allow registration completion")
	}
	userID := int64(12)
	if IsPendingOAuthCompleteRegistrationSession(OAuthIntentLogin, &userID, map[string]any{}) {
		t.Fatal("targeted login session should not allow registration completion")
	}
	if IsPendingOAuthCompleteRegistrationSession(OAuthIntentLogin, nil, map[string]any{"step": "bind_login_required"}) {
		t.Fatal("bind-login-required session should not allow registration completion")
	}
}

func TestSanitizePendingAuthLocalFlowState(t *testing.T) {
	got := SanitizePendingAuthLocalFlowState(map[string]any{
		"completion_response": map[string]any{
			"access_token":  "secret",
			"refresh_token": "refresh",
			"safe":          "value",
		},
	})
	completion, ok := got["completion_response"].(map[string]any)
	if !ok {
		t.Fatalf("completion_response = %#v", got["completion_response"])
	}
	if completion["safe"] != "value" {
		t.Fatalf("safe = %v", completion["safe"])
	}
	if _, exists := completion["access_token"]; exists {
		t.Fatal("access_token should be removed")
	}
	if _, exists := completion["refresh_token"]; exists {
		t.Fatal("refresh_token should be removed")
	}
}
