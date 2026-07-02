//go:build unit

package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsWebSessionAccountSourceAllowsCurrentLoginSources(t *testing.T) {
	require.True(t, isWebSessionAccountSource("email"))
	require.True(t, isWebSessionAccountSource("github"))
	require.True(t, isWebSessionAccountSource("google"))
}

func TestIsWebSessionAccountSourceRejectsLegacySources(t *testing.T) {
	require.False(t, isWebSessionAccountSource("touch"))
	require.False(t, isWebSessionAccountSource("linuxdo"))
	require.False(t, isWebSessionAccountSource("wechat"))
	require.False(t, isWebSessionAccountSource("oidc"))
	require.False(t, isWebSessionAccountSource("dingtalk"))
}
