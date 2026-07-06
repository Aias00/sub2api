package gateway

import (
	"testing"

	"github.com/Aias00/cloudbase/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNormalizeOpenAICompatiblePlatform(t *testing.T) {
	require.Equal(t, domain.PlatformOpenAI, NormalizeOpenAICompatiblePlatform(""))
	require.Equal(t, domain.PlatformOpenAI, NormalizeOpenAICompatiblePlatform(domain.PlatformOpenAI))
	require.Equal(t, domain.PlatformGrok, NormalizeOpenAICompatiblePlatform(domain.PlatformGrok))
	require.Equal(t, domain.PlatformOpenAI, NormalizeOpenAICompatiblePlatform(domain.PlatformAnthropic))
}
