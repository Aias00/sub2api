//go:build unit

package service

import (
	"context"
	"time"
)

type userPlatformQuotaRepoStub struct {
	bulkInsertCalls [][]UserPlatformQuotaRecord
}

func (s *userPlatformQuotaRepoStub) GetByUserPlatform(context.Context, int64, string) (*UserPlatformQuotaRecord, error) {
	return nil, nil
}

func (s *userPlatformQuotaRepoStub) BulkInsertInitial(_ context.Context, records []UserPlatformQuotaRecord) error {
	copied := make([]UserPlatformQuotaRecord, len(records))
	copy(copied, records)
	s.bulkInsertCalls = append(s.bulkInsertCalls, copied)
	return nil
}

func (s *userPlatformQuotaRepoStub) IncrementUsageWithReset(context.Context, int64, string, float64, time.Time) error {
	return nil
}

func (s *userPlatformQuotaRepoStub) ListByUser(context.Context, int64) ([]UserPlatformQuotaRecord, error) {
	return nil, nil
}

func (s *userPlatformQuotaRepoStub) UpsertForUser(context.Context, int64, []UserPlatformQuotaRecord) error {
	return nil
}

func (s *userPlatformQuotaRepoStub) ResetExpiredWindow(context.Context, int64, string, string, time.Time) error {
	return nil
}

func (s *userPlatformQuotaRepoStub) BatchSnapshotUsage(context.Context, []UserPlatformQuotaSnapshot, time.Time) error {
	return nil
}
