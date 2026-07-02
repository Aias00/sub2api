package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// Regression test: SyncAccount should reject unbound fakeid without writing articles
// This prevents orphan article records when sync is called before BindAccount
func TestWeChatExportServiceSyncAccountRejectsUnboundFakeid(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)

	// Setup: create a ready session for user 42
	repo.sessions[1] = WeChatSession{
		ID:     1,
		UserID: 42,
		Status: WeChatSessionStatusReady,
	}

	// Bind account "bound-001" for user 42
	repo.accounts[1] = WeChatAccount{
		ID:     1,
		UserID: 42,
		FakeID: "bound-001",
	}

	// Try to sync an unbound account "unbound-002"
	account, result, err := svc.SyncAccount(context.Background(), 42, "unbound-002")

	// Should return ErrWeChatAccountNotFound, not sync articles
	require.ErrorIs(t, err, ErrWeChatAccountNotFound)
	require.Nil(t, account)
	require.Nil(t, result)

	// Verify no articles were written (regression check)
	require.Empty(t, repo.articles)
}

// Regression test: SyncAccount should only sync bound accounts with exact fakeid match
func TestWeChatExportServiceSyncAccountRequiresBoundAccount(t *testing.T) {
	repo := newWeChatExportRepoFake()
	svc := NewWeChatExportService(repo, nil)

	// Setup: create a ready session for user 42
	repo.sessions[1] = WeChatSession{
		ID:     1,
		UserID: 42,
		Status: WeChatSessionStatusReady,
	}

	// Bind accounts for user 42
	repo.accounts[1] = WeChatAccount{
		ID:       1,
		UserID:   42,
		FakeID:   "biz-001",
		Nickname: "Business Account",
	}
	repo.accounts[2] = WeChatAccount{
		ID:       2,
		UserID:   42,
		FakeID:   "biz-002",
		Nickname: "Another Account",
	}

	// Try to sync an unbound fakeid that doesn't exist
	account, result, err := svc.SyncAccount(context.Background(), 42, "nonexistent-fakeid")
	require.ErrorIs(t, err, ErrWeChatAccountNotFound)
	require.Nil(t, account)
	require.Nil(t, result)

	// Verify no articles were written
	require.Empty(t, repo.articles)
}
