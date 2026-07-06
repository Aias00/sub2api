package gateway

import (
	"container/heap"
	"hash/fnv"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type OpenAIAccountCandidateScore struct {
	Index        int
	AccountID    int64
	Priority     int
	LoadRate     float64
	WaitingCount int
	LastUsedAt   *time.Time
	Score        float64
}

type OpenAISelectionSeedInput struct {
	GroupID            *int64
	SessionHash        string
	PreviousResponseID string
	RequestedModel     string
}

func IsOpenAIAccountCandidateBetter(left OpenAIAccountCandidateScore, right OpenAIAccountCandidateScore) bool {
	if left.Score != right.Score {
		return left.Score > right.Score
	}
	if left.Priority != right.Priority {
		return left.Priority < right.Priority
	}
	if left.LoadRate != right.LoadRate {
		return left.LoadRate < right.LoadRate
	}
	if left.WaitingCount != right.WaitingCount {
		return left.WaitingCount < right.WaitingCount
	}
	return left.AccountID < right.AccountID
}

type openAIAccountCandidateHeap []OpenAIAccountCandidateScore

func (h openAIAccountCandidateHeap) Len() int {
	return len(h)
}

func (h openAIAccountCandidateHeap) Less(i, j int) bool {
	return IsOpenAIAccountCandidateBetter(h[j], h[i])
}

func (h openAIAccountCandidateHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *openAIAccountCandidateHeap) Push(x any) {
	candidate, ok := x.(OpenAIAccountCandidateScore)
	if !ok {
		panic("openAIAccountCandidateHeap: invalid element type")
	}
	*h = append(*h, candidate)
}

func (h *openAIAccountCandidateHeap) Pop() any {
	old := *h
	n := len(old)
	last := old[n-1]
	*h = old[:n-1]
	return last
}

func SelectTopKOpenAIAccountCandidates(candidates []OpenAIAccountCandidateScore, topK int) []OpenAIAccountCandidateScore {
	if len(candidates) == 0 {
		return nil
	}
	if topK <= 0 {
		topK = 1
	}
	if topK >= len(candidates) {
		ranked := append([]OpenAIAccountCandidateScore(nil), candidates...)
		sort.Slice(ranked, func(i, j int) bool {
			return IsOpenAIAccountCandidateBetter(ranked[i], ranked[j])
		})
		return ranked
	}

	best := make(openAIAccountCandidateHeap, 0, topK)
	for _, candidate := range candidates {
		if len(best) < topK {
			heap.Push(&best, candidate)
			continue
		}
		if IsOpenAIAccountCandidateBetter(candidate, best[0]) {
			best[0] = candidate
			heap.Fix(&best, 0)
		}
	}

	ranked := make([]OpenAIAccountCandidateScore, len(best))
	copy(ranked, best)
	sort.Slice(ranked, func(i, j int) bool {
		return IsOpenAIAccountCandidateBetter(ranked[i], ranked[j])
	})
	return ranked
}

type OpenAISelectionRNG struct {
	state uint64
}

func NewOpenAISelectionRNG(seed uint64) OpenAISelectionRNG {
	if seed == 0 {
		seed = 0x9e3779b97f4a7c15
	}
	return OpenAISelectionRNG{state: seed}
}

func (r *OpenAISelectionRNG) NextUint64() uint64 {
	x := r.state
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	r.state = x
	return x * 2685821657736338717
}

func (r *OpenAISelectionRNG) NextFloat64() float64 {
	return float64(r.NextUint64()>>11) / (1 << 53)
}

func DeriveOpenAISelectionSeed(req OpenAISelectionSeedInput) uint64 {
	hasher := fnv.New64a()
	writeValue := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		_, _ = hasher.Write([]byte(trimmed))
		_, _ = hasher.Write([]byte{0})
	}

	writeValue(req.SessionHash)
	writeValue(req.PreviousResponseID)
	writeValue(req.RequestedModel)
	if req.GroupID != nil {
		_, _ = hasher.Write([]byte(strconv.FormatInt(*req.GroupID, 10)))
	}

	seed := hasher.Sum64()
	if strings.TrimSpace(req.SessionHash) == "" && strings.TrimSpace(req.PreviousResponseID) == "" {
		seed ^= uint64(time.Now().UnixNano())
	}
	if seed == 0 {
		seed = uint64(time.Now().UnixNano()) ^ 0x9e3779b97f4a7c15
	}
	return seed
}

func BuildOpenAIWeightedSelectionOrder(candidates []OpenAIAccountCandidateScore, req OpenAISelectionSeedInput) []OpenAIAccountCandidateScore {
	if len(candidates) <= 1 {
		return append([]OpenAIAccountCandidateScore(nil), candidates...)
	}

	pool := append([]OpenAIAccountCandidateScore(nil), candidates...)
	weights := make([]float64, len(pool))
	minScore := pool[0].Score
	for i := 1; i < len(pool); i++ {
		if pool[i].Score < minScore {
			minScore = pool[i].Score
		}
	}
	for i := range pool {
		weight := (pool[i].Score - minScore) + 1.0
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight <= 0 {
			weight = 1.0
		}
		weights[i] = weight
	}

	order := make([]OpenAIAccountCandidateScore, 0, len(pool))
	rng := NewOpenAISelectionRNG(DeriveOpenAISelectionSeed(req))
	for len(pool) > 0 {
		total := 0.0
		for _, w := range weights {
			total += w
		}

		selectedIdx := 0
		if total > 0 {
			r := rng.NextFloat64() * total
			acc := 0.0
			for i, w := range weights {
				acc += w
				if r <= acc {
					selectedIdx = i
					break
				}
			}
		} else {
			selectedIdx = int(rng.NextUint64() % uint64(len(pool)))
		}

		order = append(order, pool[selectedIdx])
		pool = append(pool[:selectedIdx], pool[selectedIdx+1:]...)
		weights = append(weights[:selectedIdx], weights[selectedIdx+1:]...)
	}
	return order
}

func SortOpenAICompactRetryCandidates(pool []OpenAIAccountCandidateScore) []OpenAIAccountCandidateScore {
	if len(pool) == 0 {
		return nil
	}
	ordered := append([]OpenAIAccountCandidateScore(nil), pool...)
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.Priority != b.Priority {
			return a.Priority < b.Priority
		}
		if a.LoadRate != b.LoadRate {
			return a.LoadRate < b.LoadRate
		}
		if a.WaitingCount != b.WaitingCount {
			return a.WaitingCount < b.WaitingCount
		}
		switch {
		case a.LastUsedAt == nil && b.LastUsedAt != nil:
			return true
		case a.LastUsedAt != nil && b.LastUsedAt == nil:
			return false
		case a.LastUsedAt == nil && b.LastUsedAt == nil:
			return false
		default:
			return a.LastUsedAt.Before(*b.LastUsedAt)
		}
	})
	return ordered
}
