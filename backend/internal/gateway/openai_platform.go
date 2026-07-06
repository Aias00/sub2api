package gateway

import "github.com/Aias00/cloudbase/internal/domain"

func NormalizeOpenAICompatiblePlatform(platform string) string {
	if platform == domain.PlatformGrok {
		return domain.PlatformGrok
	}
	return domain.PlatformOpenAI
}
