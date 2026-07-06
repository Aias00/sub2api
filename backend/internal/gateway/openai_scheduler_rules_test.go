package gateway

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSelectTopKOpenAIAccountCandidates(t *testing.T) {
	candidates := []OpenAIAccountCandidateScore{
		{AccountID: 11, Priority: 2, LoadRate: 10, WaitingCount: 1, Score: 10.0},
		{AccountID: 12, Priority: 1, LoadRate: 20, WaitingCount: 1, Score: 9.5},
		{AccountID: 13, Priority: 1, LoadRate: 30, WaitingCount: 0, Score: 10.0},
		{AccountID: 14, Priority: 0, LoadRate: 40, WaitingCount: 0, Score: 8.0},
	}

	top2 := SelectTopKOpenAIAccountCandidates(candidates, 2)
	require.Len(t, top2, 2)
	require.Equal(t, int64(13), top2[0].AccountID)
	require.Equal(t, int64(11), top2[1].AccountID)

	topAll := SelectTopKOpenAIAccountCandidates(candidates, 8)
	require.Len(t, topAll, len(candidates))
	require.Equal(t, int64(13), topAll[0].AccountID)
	require.Equal(t, int64(11), topAll[1].AccountID)
	require.Equal(t, int64(12), topAll[2].AccountID)
	require.Equal(t, int64(14), topAll[3].AccountID)
}

func TestBuildOpenAIWeightedSelectionOrderDeterministicBySessionSeed(t *testing.T) {
	candidates := []OpenAIAccountCandidateScore{
		{AccountID: 101, LoadRate: 10, WaitingCount: 0, Score: 4.2},
		{AccountID: 102, LoadRate: 30, WaitingCount: 1, Score: 3.5},
		{AccountID: 103, LoadRate: 50, WaitingCount: 2, Score: 2.1},
	}
	groupID := int64(99)
	req := OpenAISelectionSeedInput{
		GroupID:        &groupID,
		SessionHash:    "session_seed_fixed",
		RequestedModel: "gpt-5.1",
	}

	first := BuildOpenAIWeightedSelectionOrder(candidates, req)
	second := BuildOpenAIWeightedSelectionOrder(candidates, req)
	require.Len(t, first, len(candidates))
	require.Len(t, second, len(candidates))
	for i := range first {
		require.Equal(t, first[i].AccountID, second[i].AccountID)
	}
}

func TestDeriveOpenAISelectionSeedNoAffinityAddsEntropy(t *testing.T) {
	req := OpenAISelectionSeedInput{RequestedModel: "gpt-5.1"}

	seed1 := DeriveOpenAISelectionSeed(req)
	time.Sleep(1 * time.Millisecond)
	seed2 := DeriveOpenAISelectionSeed(req)
	require.NotZero(t, seed1)
	require.NotZero(t, seed2)
	require.NotEqual(t, seed1, seed2)
}

func TestBuildOpenAIWeightedSelectionOrderHandlesInvalidScores(t *testing.T) {
	candidates := []OpenAIAccountCandidateScore{
		{AccountID: 901, LoadRate: 5, WaitingCount: 0, Score: math.NaN()},
		{AccountID: 902, LoadRate: 5, WaitingCount: 0, Score: math.Inf(1)},
		{AccountID: 903, LoadRate: 5, WaitingCount: 0, Score: -1},
	}

	order := BuildOpenAIWeightedSelectionOrder(candidates, OpenAISelectionSeedInput{SessionHash: "seed_invalid_scores"})
	require.Len(t, order, len(candidates))
	seen := map[int64]struct{}{}
	for _, item := range order {
		seen[item.AccountID] = struct{}{}
	}
	require.Len(t, seen, len(candidates))
}

func TestOpenAISelectionRNGSeedZeroStillWorks(t *testing.T) {
	rng := NewOpenAISelectionRNG(0)
	v1 := rng.NextUint64()
	v2 := rng.NextUint64()
	require.NotEqual(t, v1, v2)
	require.GreaterOrEqual(t, rng.NextFloat64(), 0.0)
	require.Less(t, rng.NextFloat64(), 1.0)
}

func TestSortOpenAICompactRetryCandidates(t *testing.T) {
	old := time.Unix(10, 0)
	newer := time.Unix(20, 0)
	pool := []OpenAIAccountCandidateScore{
		{AccountID: 3, Priority: 1, LoadRate: 10, WaitingCount: 0, LastUsedAt: &newer},
		{AccountID: 2, Priority: 1, LoadRate: 10, WaitingCount: 0, LastUsedAt: &old},
		{AccountID: 1, Priority: 0, LoadRate: 99, WaitingCount: 9, LastUsedAt: &newer},
		{AccountID: 4, Priority: 1, LoadRate: 10, WaitingCount: 0},
	}

	ordered := SortOpenAICompactRetryCandidates(pool)
	require.Equal(t, []int64{1, 4, 2, 3}, []int64{
		ordered[0].AccountID,
		ordered[1].AccountID,
		ordered[2].AccountID,
		ordered[3].AccountID,
	})
}
