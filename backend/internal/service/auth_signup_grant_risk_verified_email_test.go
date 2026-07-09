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

// newRiskTestAuthService 构造一个带 SettingService（控制风控配置）但 entClient=nil 的 AuthService，
// 用于测试在 DB nil-check 之前触发的 RequireVerifiedEmail 规则。
func newRiskTestAuthService(values map[string]string) *AuthService {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-secret"}}
	settingService := NewSettingService(&settingRepoStub{values: values}, cfg)
	return &AuthService{
		settingService: settingService,
		cfg:            cfg,
	}
}

// TestApplySignupGrantRiskControl_RequireVerifiedEmail_DefaultBlocks
// 验证 RequireVerifiedEmail 默认开启、且 EmailVerified 未设置时，即便风控总开关关闭也剥夺赠金。
// 这是"独立于风控总开关"的关键回归：默认部署（风控关）仍强制邮箱验证。
func TestApplySignupGrantRiskControl_RequireVerifiedEmail_DefaultBlocks(t *testing.T) {
	// 风控总开关关闭、未显式设置 RequireVerifiedEmail（默认 true）
	svc := newRiskTestAuthService(map[string]string{
		SettingKeySignupGrantRiskControlEnabled: "false",
	})
	plan := signupGrantPlan{Balance: 3.5}
	ctx := context.Background() // 无 EmailVerified

	filtered, claim := svc.applySignupGrantRiskControl(ctx, "user@test.com", "email", plan)
	require.True(t, claim.Blocked)
	require.Equal(t, "email_not_verified", claim.Reason)
	require.Equal(t, 0.0, filtered.Balance)
}

// TestApplySignupGrantRiskControl_RequireVerifiedEmail_VerifiedPasses
// 验证 EmailVerified=true 时放行（风控总开关关闭 → 直接返回原 plan）。
func TestApplySignupGrantRiskControl_RequireVerifiedEmail_VerifiedPasses(t *testing.T) {
	svc := newRiskTestAuthService(map[string]string{
		SettingKeySignupGrantRiskControlEnabled: "false",
	})
	plan := signupGrantPlan{Balance: 3.5}
	ctx := WithSignupGrantRiskInput(context.Background(), SignupGrantRiskInput{
		EmailVerified: signupGrantEmailVerified(true),
	})

	filtered, claim := svc.applySignupGrantRiskControl(ctx, "user@test.com", "email", plan)
	require.Nil(t, claim)
	require.Equal(t, 3.5, filtered.Balance)
}

// TestApplySignupGrantRiskControl_RequireVerifiedEmail_DisabledKeepsGrant
// 验证运营显式关闭 RequireVerifiedEmail 时，未验证注册仍发赠金（逃生阀）。
func TestApplySignupGrantRiskControl_RequireVerifiedEmail_DisabledKeepsGrant(t *testing.T) {
	svc := newRiskTestAuthService(map[string]string{
		SettingKeySignupGrantRiskControlEnabled:              "false",
		SettingKeySignupGrantRiskControlRequireVerifiedEmail: "false",
	})
	plan := signupGrantPlan{Balance: 3.5}
	ctx := context.Background() // 无 EmailVerified

	filtered, claim := svc.applySignupGrantRiskControl(ctx, "user@test.com", "email", plan)
	require.Nil(t, claim)
	require.Equal(t, 3.5, filtered.Balance)
}

// TestMergeSignupGrantRiskInput_EmailVerifiedBoolPointer
// 验证 *bool 合并不丢失 false（plain bool + 非空判断会吞 false）。
func TestMergeSignupGrantRiskInput_EmailVerifiedBoolPointer(t *testing.T) {
	t.Run("patch true overrides base false", func(t *testing.T) {
		base := SignupGrantRiskInput{EmailVerified: signupGrantEmailVerified(false)}
		patch := SignupGrantRiskInput{EmailVerified: signupGrantEmailVerified(true)}
		merged := mergeSignupGrantRiskInput(base, patch)
		require.NotNil(t, merged.EmailVerified)
		require.True(t, *merged.EmailVerified)
	})
	t.Run("patch nil keeps base false", func(t *testing.T) {
		base := SignupGrantRiskInput{EmailVerified: signupGrantEmailVerified(false)}
		merged := mergeSignupGrantRiskInput(base, SignupGrantRiskInput{})
		require.NotNil(t, merged.EmailVerified)
		require.False(t, *merged.EmailVerified)
	})
	t.Run("patch false overrides base true", func(t *testing.T) {
		base := SignupGrantRiskInput{EmailVerified: signupGrantEmailVerified(true)}
		patch := SignupGrantRiskInput{EmailVerified: signupGrantEmailVerified(false)}
		merged := mergeSignupGrantRiskInput(base, patch)
		require.NotNil(t, merged.EmailVerified)
		require.False(t, *merged.EmailVerified)
	})
}

// TestApplySignupGrantRiskControlEx_ForceBypassesSourceWhitelist
// 验证 force=true 时，非白名单来源（如 linuxdo）仍跑风控限额逻辑（首绑奖励接入复用此入口）。
// 这里风控开启、RequireVerifiedEmail 关闭、DB 不可用（entClient=nil）→ fail-closed 剥夺赠金。
// 若 force 未生效（被白名单早返回），则 plan 不变、claim 为 nil —— 用以区分。
func TestApplySignupGrantRiskControlEx_ForceBypassesSourceWhitelist(t *testing.T) {
	svc := newRiskTestAuthService(map[string]string{
		SettingKeySignupGrantRiskControlEnabled:              "true",
		SettingKeySignupGrantRiskControlRequireVerifiedEmail: "false",
	})
	plan := signupGrantPlan{Balance: 5.0}

	// force=false：linuxdo 非白名单 → 早返回，plan 不变
	filtered, claim := svc.applySignupGrantRiskControlEx(context.Background(), "user@linuxdo.test", "linuxdo", plan, false)
	require.Nil(t, claim)
	require.Equal(t, 5.0, filtered.Balance)

	// force=true：绕过白名单，进入风控；entClient=nil → DB 不可用 → fail-closed blocked
	filtered2, claim2 := svc.applySignupGrantRiskControlEx(context.Background(), "user@linuxdo.test", "linuxdo", plan, true)
	require.NotNil(t, claim2)
	require.True(t, claim2.Blocked)
	require.Equal(t, "risk_check_unavailable", claim2.Reason)
	require.Equal(t, 0.0, filtered2.Balance)
}

// TestApplySignupGrantRiskControl_AdvisoryLockFailureFailsClosed
// 验证 advisory lock 失败时 fail-closed：剥夺赠金、返回 nil claim（非合成 blocked claim）。
func TestApplySignupGrantRiskControl_AdvisoryLockFailureFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer func() { _ = client.Close() }()

	svc := &AuthService{
		entClient: client,
		settingService: NewSettingService(&settingRepoStub{values: map[string]string{
			SettingKeySignupGrantRiskControlEnabled:              "true",
			SettingKeySignupGrantRiskControlRequireVerifiedEmail: "false",
		}}, &config.Config{JWT: config.JWTConfig{Secret: "test-secret"}}),
		cfg: &config.Config{JWT: config.JWTConfig{Secret: "test-secret"}},
	}

	ctx := WithSignupGrantRiskInput(context.Background(), SignupGrantRiskInput{RemoteIP: "203.0.113.9"})
	plan := signupGrantPlan{Balance: 1.0}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SELECT pg_advisory_xact_lock($1)")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("lock unavailable"))

	filtered, claim := svc.applySignupGrantRiskControl(ctx, "user@test.com", "email", plan)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Nil(t, claim) // fail-closed 返回 nil claim
	require.Equal(t, 0.0, filtered.Balance)
}

// giftBalanceSetterStub 记录 SetGiftBalanceComponent 调用次数，用于断言 attachSignupGrantClaim
// 的 markGift 参数是否触发 gift 余额标记。嵌入 userRepoStub 以同时满足 UserRepository 接口。
type giftBalanceSetterStub struct {
	*userRepoStub
	setGiftCalls int
	setGiftArg   float64
}

func (s *giftBalanceSetterStub) SetGiftBalanceComponent(_ context.Context, _ int64, amount float64) error {
	s.setGiftCalls++
	s.setGiftArg = amount
	return nil
}

// TestAttachSignupGrantClaim_MarkGiftFlag
// 回归 #1 引入的 gift_balance 双记 bug：首绑路径用 markGift=false（gift 组件交给 ApplyBalanceChangeCtx
// 的 ADD+GiftDelta），注册路径用 markGift=true（balance 由 Create 一次性写入，此处仅幂等标记 gift）。
// 断言：markGift=false 时 SetGiftBalanceComponent 不被调用；markGift=true 且 allowed+GrantBalance>0 时被调用一次。
func TestAttachSignupGrantClaim_MarkGiftFlag(t *testing.T) {
	const userID int64 = 42
	const grantBalance = 5.0

	t.Run("markGift false skips setter (first-bind path)", func(t *testing.T) {
		stub := &giftBalanceSetterStub{userRepoStub: &userRepoStub{}}
		svc := &AuthService{userRepo: stub}
		claim := &signupGrantRiskClaim{
			ID:           1,
			Decision:     signupGrantRiskDecisionAllowed,
			GrantBalance: grantBalance,
		}
		svc.attachSignupGrantClaim(context.Background(), claim, userID, false)
		require.Equal(t, 0, stub.setGiftCalls, "markGift=false must not call SetGiftBalanceComponent (gift handled by ApplyBalanceChangeCtx)")
	})

	t.Run("markGift true calls setter (registration path)", func(t *testing.T) {
		stub := &giftBalanceSetterStub{userRepoStub: &userRepoStub{}}
		svc := &AuthService{userRepo: stub}
		claim := &signupGrantRiskClaim{
			ID:           2,
			Decision:     signupGrantRiskDecisionAllowed,
			GrantBalance: grantBalance,
		}
		svc.attachSignupGrantClaim(context.Background(), claim, userID, true)
		require.Equal(t, 1, stub.setGiftCalls, "markGift=true must call SetGiftBalanceComponent once")
		require.Equal(t, grantBalance, stub.setGiftArg)
	})

	t.Run("blocked claim never marks gift regardless of markGift", func(t *testing.T) {
		stub := &giftBalanceSetterStub{userRepoStub: &userRepoStub{}}
		svc := &AuthService{userRepo: stub}
		blocked := &signupGrantRiskClaim{
			ID:           3,
			Blocked:      true,
			Decision:     signupGrantRiskDecisionBlocked,
			GrantBalance: grantBalance,
		}
		svc.attachSignupGrantClaim(context.Background(), blocked, userID, true)
		require.Equal(t, 0, stub.setGiftCalls, "blocked claim must not mark gift even with markGift=true")
	})
}
