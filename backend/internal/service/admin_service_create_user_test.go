//go:build unit

package service

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/internal/config"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestAdminService_CreateUser_Success(t *testing.T) {
	repo := &userRepoStub{nextID: 10}
	svc := &adminServiceImpl{userRepo: repo}
	balance := 12.5

	input := &CreateUserInput{
		Email:         "user@test.com",
		Password:      "strong-pass",
		Username:      "tester",
		Notes:         "note",
		Balance:       &balance,
		Concurrency:   7,
		AllowedGroups: []int64{3, 5},
	}

	user, err := svc.CreateUser(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, int64(10), user.ID)
	require.Equal(t, input.Email, user.Email)
	require.Equal(t, input.Username, user.Username)
	require.Equal(t, input.Notes, user.Notes)
	require.Equal(t, balance, user.Balance)
	require.Equal(t, input.Concurrency, user.Concurrency)
	require.Equal(t, input.AllowedGroups, user.AllowedGroups)
	require.Equal(t, RoleUser, user.Role)
	require.Equal(t, StatusActive, user.Status)
	require.True(t, user.CheckPassword(input.Password))
	require.Len(t, repo.created, 1)
	require.Equal(t, user, repo.created[0])
}

func TestAdminService_CreateUser_UsesDefaultBalanceWhenBalanceOmitted(t *testing.T) {
	repo := &userRepoStub{nextID: 11}
	cfg := &config.Config{
		Default: config.DefaultConfig{
			UserBalance: 0,
		},
	}
	settingService := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyDefaultBalance: "0.02",
	}}, cfg)
	svc := &adminServiceImpl{userRepo: repo, settingService: settingService}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "default-balance@test.com",
		Password: "strong-pass",
	})

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, 0.02, user.Balance)
	require.Len(t, repo.created, 1)
	require.Equal(t, 0.02, repo.created[0].Balance)
}

func TestAdminService_CreateUser_ExplicitZeroBalanceOverridesDefault(t *testing.T) {
	repo := &userRepoStub{nextID: 12}
	cfg := &config.Config{
		Default: config.DefaultConfig{
			UserBalance: 0,
		},
	}
	settingService := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyDefaultBalance: "0.02",
	}}, cfg)
	svc := &adminServiceImpl{userRepo: repo, settingService: settingService}
	balance := 0.0

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "zero-balance@test.com",
		Password: "strong-pass",
		Balance:  &balance,
	})

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, 0.0, user.Balance)
	require.Len(t, repo.created, 1)
	require.Equal(t, 0.0, repo.created[0].Balance)
}

func TestAdminService_CreateUser_EmailExists(t *testing.T) {
	repo := &userRepoStub{createErr: ErrEmailExists}
	svc := &adminServiceImpl{userRepo: repo}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "dup@test.com",
		Password: "password",
	})
	require.ErrorIs(t, err, ErrEmailExists)
	require.Empty(t, repo.created)
}

func TestAdminService_CreateUser_CreateError(t *testing.T) {
	createErr := errors.New("db down")
	repo := &userRepoStub{createErr: createErr}
	svc := &adminServiceImpl{userRepo: repo}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "user@test.com",
		Password: "password",
	})
	require.ErrorIs(t, err, createErr)
	require.Empty(t, repo.created)
}

func TestAdminService_CreateUser_AssignsDefaultSubscriptions(t *testing.T) {
	repo := &userRepoStub{nextID: 21}
	assigner := &defaultSubscriptionAssignerStub{}
	cfg := &config.Config{
		Default: config.DefaultConfig{
			UserBalance:     0,
			UserConcurrency: 1,
		},
	}
	settingService := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyDefaultSubscriptions: `[{"group_id":5,"validity_days":30}]`,
	}}, cfg)
	svc := &adminServiceImpl{
		userRepo:           repo,
		settingService:     settingService,
		defaultSubAssigner: assigner,
	}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "new-user@test.com",
		Password: "password",
	})
	require.NoError(t, err)
	require.Len(t, assigner.calls, 1)
	require.Equal(t, int64(21), assigner.calls[0].UserID)
	require.Equal(t, int64(5), assigner.calls[0].GroupID)
	require.Equal(t, 30, assigner.calls[0].ValidityDays)
}

// TestAdminService_CreateUser_RecordsAdminRegistrationEvent 断言后台建号会写一条
// user_registration_events 事件行：source 为哨兵 "admin"、ip_address 与 ip_prefix 为空。
// 空 IP 保证这些账号不进 /admin/user-insights 的 Top IPs 聚合（WHERE ip_address <> ''），
// 也不污染 SameIPSignupCount24h 风控信号。
func TestAdminService_CreateUser_RecordsAdminRegistrationEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer func() { _ = client.Close() }()

	repo := &userRepoStub{nextID: 99}
	svc := &adminServiceImpl{userRepo: repo, entClient: client}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_registration_events")).
		WithArgs(
			int64(99),
			"admin-created@test.com",
			"admin",
			"",
			"",
			"", // ip_address 空
			"", // ip_prefix 空
			"",
			"",
			"",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "admin-created@test.com",
		Password: "strong-pass",
		Username: "admin-created",
	})
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, int64(99), user.ID)

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestAdminService_CreateUser_NoRegistrationEventWithoutEntClient 断言 entClient 为 nil 时
// CreateUser 仍正常工作且不写注册事件（admin 侧多数单测构造无 entClient 的 service）。
func TestAdminService_CreateUser_NoRegistrationEventWithoutEntClient(t *testing.T) {
	repo := &userRepoStub{nextID: 100}
	svc := &adminServiceImpl{userRepo: repo} // 无 entClient

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "no-entclient@test.com",
		Password: "strong-pass",
	})
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, int64(100), user.ID)
}
