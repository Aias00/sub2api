package identity

import "strings"

const (
	OAuthIntentLogin             = "login"
	OAuthIntentBindCurrentUser   = "bind_current_user"
	OAuthIntentAdoptExistingUser = "adopt_existing_user_by_email"
	WeChatOAuthProviderKey       = "wechat-main"
	WeChatOAuthLegacyProviderKey = "wechat"
)

func IsSupportedPendingOAuthProvider(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "github", "google":
		return true
	default:
		return false
	}
}

func WeChatOAuthIdentityProviderKeys(providerKey string) []string {
	providerKey = strings.TrimSpace(providerKey)
	if providerKey == WeChatOAuthProviderKey {
		return []string{WeChatOAuthProviderKey, WeChatOAuthLegacyProviderKey}
	}
	return []string{providerKey}
}

func ShouldBindPendingOAuthIdentity(intent string, adoptDisplayName, adoptAvatar bool) bool {
	switch strings.ToLower(strings.TrimSpace(intent)) {
	case OAuthIntentBindCurrentUser, OAuthIntentLogin, OAuthIntentAdoptExistingUser:
		return true
	default:
		return adoptDisplayName || adoptAvatar
	}
}
