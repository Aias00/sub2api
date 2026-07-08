package identity

import (
	"strconv"
	"sync"
	"time"
)

const (
	PublicUserIDPrefix       = "u_"
	publicUserIDEpochMillis  = int64(1704067200000) // 2024-01-01T00:00:00Z
	publicUserIDNodeBits     = uint(10)
	publicUserIDSequenceBits = uint(12)
	publicUserIDNodeID       = int64(1)
	publicUserIDMaxSequence  = int64(1<<publicUserIDSequenceBits) - 1
)

var publicUserIDState = struct {
	sync.Mutex
	lastMillis int64
	sequence   int64
}{}

// NewPublicUserID returns a Snowflake-style, roughly time-ordered public user ID.
// It is serialized as a string because the numeric value exceeds JavaScript's
// safe integer range.
func NewPublicUserID() string {
	return PublicUserIDPrefix + strconv.FormatInt(newPublicUserIDInt(time.Now), 10)
}

func newPublicUserIDInt(now func() time.Time) int64 {
	publicUserIDState.Lock()
	defer publicUserIDState.Unlock()

	currentMillis := now().UTC().UnixMilli() - publicUserIDEpochMillis
	if currentMillis < 0 {
		currentMillis = 0
	}

	if currentMillis == publicUserIDState.lastMillis {
		publicUserIDState.sequence = (publicUserIDState.sequence + 1) & publicUserIDMaxSequence
		if publicUserIDState.sequence == 0 {
			for currentMillis <= publicUserIDState.lastMillis {
				currentMillis = now().UTC().UnixMilli() - publicUserIDEpochMillis
				if currentMillis < 0 {
					currentMillis = 0
				}
			}
		}
	} else {
		publicUserIDState.sequence = 0
	}

	publicUserIDState.lastMillis = currentMillis
	return (currentMillis << (publicUserIDNodeBits + publicUserIDSequenceBits)) |
		(publicUserIDNodeID << publicUserIDSequenceBits) |
		publicUserIDState.sequence
}
