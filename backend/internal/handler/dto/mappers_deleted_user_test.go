package dto

import (
	"testing"
	"time"

	"github.com/Aias00/cloudbase/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserFromServiceShallow_MapsDeletedAt(t *testing.T) {
	ts := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)

	deleted := UserFromServiceShallow(&service.User{ID: 1, PublicID: "u_735000000000000001", Email: "d@test.com", DeletedAt: &ts})
	require.NotNil(t, deleted.DeletedAt)
	require.Equal(t, ts, *deleted.DeletedAt)
	require.Equal(t, "u_735000000000000001", deleted.PublicID)

	active := UserFromServiceShallow(&service.User{ID: 2, Email: "a@test.com"})
	require.Nil(t, active.DeletedAt, "active user must have nil DeletedAt")
}
