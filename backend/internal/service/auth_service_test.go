package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsReservedEmail_InvalidDomain(t *testing.T) {
	require.True(t, isReservedEmail("user@reserved.invalid"))
	require.True(t, isReservedEmail("USER@RESERVED.INVALID")) // case-insensitive
	require.False(t, isReservedEmail("real@dingtalk.com"))
}
