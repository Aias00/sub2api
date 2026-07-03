package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/ent/identityadoptiondecision"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryUpsertIdentityAdoptionDecisionIsIdempotentUnderConcurrency(t *testing.T) {
	repo, client := newUserEntRepo(t)
	ctx := context.Background()

	user := &service.User{
		Email:        "repo-adoption@example.com",
		Username:     "repo-adoption",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, user))

	identity, err := client.AuthIdentity.Create().
		SetUserID(user.ID).
		SetProviderType("github").
		SetProviderKey("wechat-main").
		SetProviderSubject("union-repo-adoption").
		SetMetadata(map[string]any{}).
		Save(ctx)
	require.NoError(t, err)

	session, err := client.PendingAuthSession.Create().
		SetSessionToken("pending-repo-adoption").
		SetIntent("bind_current_user").
		SetProviderType("github").
		SetProviderKey("wechat-main").
		SetProviderSubject("union-repo-adoption").
		SetExpiresAt(time.Now().UTC().Add(15 * time.Minute)).
		SetUpstreamIdentityClaims(map[string]any{"provider_subject": "union-repo-adoption"}).
		SetLocalFlowState(map[string]any{"step": "pending"}).
		Save(ctx)
	require.NoError(t, err)

	firstCreateStarted := make(chan struct{})
	releaseFirstCreate := make(chan struct{})
	var firstCreate sync.Once
	client.IdentityAdoptionDecision.Use(func(next dbent.Mutator) dbent.Mutator {
		return dbent.MutateFunc(func(ctx context.Context, m dbent.Mutation) (dbent.Value, error) {
			blocked := false
			if m.Op().Is(dbent.OpCreate) {
				firstCreate.Do(func() {
					blocked = true
					close(firstCreateStarted)
				})
			}
			if blocked {
				<-releaseFirstCreate
			}
			return next.Mutate(ctx, m)
		})
	})

	type adoptionResult struct {
		decision *dbent.IdentityAdoptionDecision
		err      error
	}

	input := IdentityAdoptionDecisionInput{
		PendingAuthSessionID: session.ID,
		IdentityID:           &identity.ID,
		AdoptDisplayName:     true,
		AdoptAvatar:          true,
	}

	results := make(chan adoptionResult, 2)
	go func() {
		decision, err := repo.UpsertIdentityAdoptionDecision(ctx, input)
		results <- adoptionResult{decision: decision, err: err}
	}()

	<-firstCreateStarted

	go func() {
		decision, err := repo.UpsertIdentityAdoptionDecision(ctx, input)
		results <- adoptionResult{decision: decision, err: err}
	}()

	time.Sleep(100 * time.Millisecond)
	close(releaseFirstCreate)

	first := <-results
	second := <-results

	require.NoError(t, first.err)
	require.NoError(t, second.err)
	require.NotNil(t, first.decision)
	require.NotNil(t, second.decision)
	require.Equal(t, first.decision.ID, second.decision.ID)

	count, err := client.IdentityAdoptionDecision.Query().
		Where(identityadoptiondecision.PendingAuthSessionIDEQ(session.ID)).
		Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	loaded, err := client.IdentityAdoptionDecision.Query().
		Where(identityadoptiondecision.PendingAuthSessionIDEQ(session.ID)).
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, loaded.IdentityID)
	require.Equal(t, identity.ID, *loaded.IdentityID)
	require.True(t, loaded.AdoptDisplayName)
	require.True(t, loaded.AdoptAvatar)
}
