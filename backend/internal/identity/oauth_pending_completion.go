package identity

import "strings"

const OAuthPendingChoiceStep = "choose_account_action_required"

func ClonePendingMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func PendingSessionStringValue(values map[string]any, key string) string {
	if len(values) == 0 {
		return ""
	}
	raw, ok := values[key]
	if !ok {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func ApplySuggestedProfileToCompletionResponse(payload map[string]any, upstream map[string]any) {
	if len(payload) == 0 || len(upstream) == 0 {
		return
	}

	displayName := PendingSessionStringValue(upstream, "suggested_display_name")
	avatarURL := PendingSessionStringValue(upstream, "suggested_avatar_url")

	if displayName != "" {
		if _, exists := payload["suggested_display_name"]; !exists {
			payload["suggested_display_name"] = displayName
		}
	}
	if avatarURL != "" {
		if _, exists := payload["suggested_avatar_url"]; !exists {
			payload["suggested_avatar_url"] = avatarURL
		}
	}
	if displayName != "" || avatarURL != "" {
		payload["adoption_required"] = true
	}
}

func NormalizePendingOAuthCompletionResponse(payload map[string]any) map[string]any {
	normalized := ClonePendingMap(payload)
	for _, key := range []string{"access_token", "refresh_token", "expires_in", "token_type"} {
		delete(normalized, key)
	}
	step := strings.ToLower(strings.TrimSpace(PendingSessionStringValue(normalized, "step")))
	switch step {
	case "choice", "choose_account_action", "choose_account", "choose", "email_required":
		normalized["step"] = OAuthPendingChoiceStep
	}
	if strings.EqualFold(strings.TrimSpace(PendingSessionStringValue(normalized, "step")), OAuthPendingChoiceStep) {
		normalized["adoption_required"] = true
	}
	if _, exists := normalized["adoption_required"]; !exists {
		if _, hasChoiceFields := normalized["email_binding_required"]; hasChoiceFields {
			normalized["adoption_required"] = true
		}
	}
	return normalized
}

func PendingSessionWantsInvitation(payload map[string]any) bool {
	return strings.EqualFold(strings.TrimSpace(PendingSessionStringValue(payload, "error")), "invitation_required")
}

func PendingSessionRequiresEmailCompletion(payload map[string]any) bool {
	if v, ok := payload["requires_email_completion"].(bool); ok && v {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(PendingSessionStringValue(payload, "step")), "email_completion")
}

func PendingSessionRequiresBindLogin(payload map[string]any) bool {
	return strings.EqualFold(strings.TrimSpace(PendingSessionStringValue(payload, "step")), "bind_login_required")
}

func PendingOAuthCompletionCanIssueTokenPair(intent string, targetUserID *int64, payload map[string]any) bool {
	if !strings.EqualFold(strings.TrimSpace(intent), OAuthIntentLogin) {
		return false
	}
	if targetUserID == nil || *targetUserID <= 0 {
		return false
	}
	if PendingSessionWantsInvitation(payload) {
		return false
	}
	return strings.TrimSpace(PendingSessionStringValue(payload, "step")) == ""
}

func IsPendingOAuthCompleteRegistrationSession(intent string, targetUserID *int64, payload map[string]any) bool {
	if strings.TrimSpace(intent) != OAuthIntentLogin {
		return false
	}
	if targetUserID != nil && *targetUserID > 0 {
		return false
	}
	return !PendingSessionRequiresBindLogin(payload)
}

func SanitizePendingAuthLocalFlowState(localFlowState map[string]any) map[string]any {
	sanitized := ClonePendingMap(localFlowState)
	if len(sanitized) == 0 {
		return sanitized
	}

	rawCompletion, ok := sanitized["completion_response"]
	if !ok {
		return sanitized
	}
	completion, ok := rawCompletion.(map[string]any)
	if !ok {
		return sanitized
	}

	cleanedCompletion := ClonePendingMap(completion)
	for _, key := range []string{"access_token", "refresh_token", "expires_in", "token_type"} {
		delete(cleanedCompletion, key)
	}
	sanitized["completion_response"] = cleanedCompletion
	return sanitized
}
