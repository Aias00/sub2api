//go:build integration

package service

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/internal/config"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/Aias00/cloudbase/migrations"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// 本文件用 testcontainers 起真实 PostgreSQL，验证注册赠金风控的事务级 advisory lock
// （消除 TOCTOU 竞态）与首绑奖励接入风控的端到端行为。需 Docker；CI 在 integration tag 下运行。
//
// 注意：本测试为 package service（内部测试）以访问未导出的 applySignupGrantRiskControl 等，
// 故不能 import internal/repository（会形成 import cycle：repository 依赖 service）。
// 这里直接用 migrations.FS 在全新 Postgres 容器上顺序执行 SQL 迁移文件。

var (
	svcIntegrationDB   *sql.DB
	svcIntegrationEnt  *dbent.Client
	svcIntegrationOnce sync.Once
)

func ensureSvcIntegrationDB(t *testing.T) {
	t.Helper()
	svcIntegrationOnce.Do(func() {
		ctx := context.Background()
		if !dockerAvailableSvc(ctx) {
			if os.Getenv("CI") != "" {
				t.Fatalf("docker is not available (CI=true); failing integration tests")
			}
			t.Skip("docker not available; skipping integration test")
		}
		pg, err := tcpostgres.Run(ctx, "postgres:18.1-alpine3.23",
			tcpostgres.WithDatabase("cloudbase_test"),
			tcpostgres.WithUsername("postgres"),
			tcpostgres.WithPassword("postgres"),
			tcpostgres.BasicWaitStrategies(),
		)
		if err != nil {
			t.Fatalf("start postgres container: %v", err)
		}
		// 不在此 Terminate：t.Cleanup 绑定首个调用本函数的测试，该测试一结束就 Terminate，
		// 后续共享同一 svcIntegrationEnt 的测试会 connection refused。容器由 testcontainers 的
		// ryuk reaper 在测试进程退出时统一回收（默认启用，日志可见 testcontainers/ryuk 容器）。

		dsn, err := pg.ConnectionString(ctx, "sslmode=disable", "TimeZone=UTC")
		if err != nil {
			t.Fatalf("postgres dsn: %v", err)
		}
		db, err := openSvcSQLWithRetry(ctx, dsn, 30*time.Second)
		if err != nil {
			t.Fatalf("open sql: %v", err)
		}
		if err := applyMigrationsSvc(ctx, db); err != nil {
			t.Fatalf("apply migrations: %v", err)
		}
		drv := entsql.OpenDB(dialect.Postgres, db)
		svcIntegrationDB = db
		svcIntegrationEnt = dbent.NewClient(dbent.Driver(drv))
	})
	if svcIntegrationEnt == nil {
		t.Skip("integration db not available")
	}
}

func dockerAvailableSvc(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "info")
	cmd.Env = os.Environ()
	return cmd.Run() == nil
}

func openSvcSQLWithRetry(ctx context.Context, dsn string, timeout time.Duration) (*sql.DB, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = db.PingContext(pingCtx)
		cancel()
		if err != nil {
			lastErr = err
			_ = db.Close()
			time.Sleep(250 * time.Millisecond)
			continue
		}
		return db, nil
	}
	return nil, fmt.Errorf("db not ready after %s: %w", timeout, lastErr)
}

// applyMigrationsSvc 在全新 Postgres 上按文件名顺序执行 embed 的 SQL 迁移。
// 复刻 internal/repository/migrations_runner.go 的关键行为：
//   - 建 schema_migrations 表并逐文件记录（filename, applied_at），因部分迁移（006b/123）依赖它
//     判定他迁移是否已应用 / 读取其 applied_at 时间窗。
//   - *_notx.sql（CREATE/DROP INDEX CONCURRENTLY）逐条语句、事务外执行——否则 lib/pq 把多语句放进
//     隐式事务块，CONCURRENTLY 会被拒绝。
//
// 本包不能 import internal/repository（import cycle），故在此内联等价实现。
func applyMigrationsSvc(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	filename   TEXT PRIMARY KEY,
	checksum   TEXT NOT NULL,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".sql" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		content, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if isNonTxMigrationName(name) {
			for i, stmt := range splitSvcSQLStatements(string(content)) {
				trimmed := strings.TrimSpace(stmt)
				if trimmed == "" || strings.TrimSpace(stripSvcSQLLineComment(trimmed)) == "" {
					continue
				}
				if _, err := db.ExecContext(ctx, trimmed); err != nil {
					return fmt.Errorf("apply migration %s (non-tx statement %d): %w", name, i+1, err)
				}
			}
		} else if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		// 记录已应用，供后续迁移（如 123 读取 110 的 applied_at）引用。
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2) ON CONFLICT (filename) DO NOTHING`, name, fmt.Sprintf("svc-%d", len(name))); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}
	return nil
}

// isNonTxMigrationName 判定迁移是否需非事务执行（与 migrations_runner.go 的 _notx.sql 约定一致）。
func isNonTxMigrationName(name string) bool {
	return strings.HasSuffix(name, "_notx.sql")
}

// splitSvcSQLStatements 按分号拆分 SQL（与 migrations_runner.splitSQLStatements 等价，本包不能 import repository）。
func splitSvcSQLStatements(content string) []string {
	parts := strings.Split(content, ";")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

// stripSvcSQLLineComment 去除 SQL 行内 -- 注释（与 migrations_runner.stripSQLLineComment 等价）。
func stripSvcSQLLineComment(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "--"); idx >= 0 {
			lines[i] = line[:idx]
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func newSvcRiskAuthService(t *testing.T, settings map[string]string) *AuthService {
	ensureSvcIntegrationDB(t)
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "svc-integration-secret"}}
	settingService := NewSettingService(&svcSettingRepoStub{values: settings}, cfg)
	return &AuthService{
		entClient:      svcIntegrationEnt,
		settingService: settingService,
		cfg:            cfg,
	}
}

// svcSettingRepoStub 是集成测试专用的最小 SettingRepository（避免依赖 unit-tag 下的 settingRepoStub）。
type svcSettingRepoStub struct {
	values map[string]string
}

func (s *svcSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	return nil, fmt.Errorf("not implemented")
}
func (s *svcSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if v, ok := s.values[key]; ok {
		return v, nil
	}
	return "", ErrSettingNotFound
}
func (s *svcSettingRepoStub) Set(ctx context.Context, key, value string) error {
	return nil
}
func (s *svcSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := s.values[k]; ok {
			result[k] = v
		}
	}
	return result, nil
}
func (s *svcSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	return nil
}
func (s *svcSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	return s.values, nil
}
func (s *svcSettingRepoStub) Delete(ctx context.Context, key string) error { return nil }

// TestIntegration_SignupGrantRisk_TxSerializesConcurrentIP
// 真正的 TOCTOU 回归：IPLimit=1，并发发起 N 个同 IP 的赠金评估，
// 断言"allowed"的 claim 恰好 1 个（修复前并发会全部通过 count → allowed > 1）。
func TestIntegration_SignupGrantRisk_TxSerializesConcurrentIP(t *testing.T) {
	svc := newSvcRiskAuthService(t, map[string]string{
		SettingKeySignupGrantRiskControlEnabled:              "true",
		SettingKeySignupGrantRiskControlIPLimit:              "1",
		SettingKeySignupGrantRiskControlEmailLimit:           "50", // 放开 email 维度，聚焦 IP 维度
		SettingKeySignupGrantRiskControlDomainLimit:          "50",
		SettingKeySignupGrantRiskControlRequireVerifiedEmail: "false",
	})

	const concurrency = 8
	ctx := WithSignupGrantRiskInput(context.Background(), SignupGrantRiskInput{RemoteIP: "198.51.100.7"})

	var wg sync.WaitGroup
	blockedCount := 0
	allowedCount := 0
	var mu sync.Mutex
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			email := fmt.Sprintf("toctou-%d@ipserial.test", i)
			_, claim := svc.applySignupGrantRiskControl(ctx, email, "email", signupGrantPlan{Balance: 1.0})
			mu.Lock()
			defer mu.Unlock()
			if claim == nil {
				return
			}
			if claim.Blocked {
				blockedCount++
			} else {
				allowedCount++
			}
		}(i)
	}
	wg.Wait()

	require.Equal(t, 1, allowedCount, "exactly one concurrent same-IP grant should be allowed")
	require.Equal(t, concurrency-1, blockedCount, "all but one should be blocked by ip_daily_limit")
}

// TestIntegration_SignupGrantRisk_RequireVerifiedEmail_DefaultOn
// 默认部署（风控总开关关闭）+ 新默认 RequireVerifiedEmail，未验证注册赠金被剥夺。
func TestIntegration_SignupGrantRisk_RequireVerifiedEmail_DefaultOn(t *testing.T) {
	svc := newSvcRiskAuthService(t, map[string]string{
		SettingKeySignupGrantRiskControlEnabled: "false", // 风控总开关关闭
		// RequireVerifiedEmail 未显式设置 → 默认 true
	})
	plan := signupGrantPlan{Balance: 2.0}
	ctx := context.Background() // 无 EmailVerified

	filtered, claim := svc.applySignupGrantRiskControl(ctx, "unverified@test.com", "email", plan)
	require.NotNil(t, claim)
	require.True(t, claim.Blocked)
	require.Equal(t, "email_not_verified", claim.Reason)
	require.Equal(t, 0.0, filtered.Balance)
}

// TestIntegration_FirstBindGate_BlockedStripsBalance
// 首绑奖励接入风控：linuxdo（非白名单）force=true 复用限额表；同 IP 第二次命中 ip_daily_limit 被剥夺。
func TestIntegration_FirstBindGate_BlockedStripsBalance(t *testing.T) {
	svc := newSvcRiskAuthService(t, map[string]string{
		SettingKeySignupGrantRiskControlEnabled:              "true",
		SettingKeySignupGrantRiskControlIPLimit:              "1",
		SettingKeySignupGrantRiskControlEmailLimit:           "50",
		SettingKeySignupGrantRiskControlDomainLimit:          "50",
		SettingKeySignupGrantRiskControlRequireVerifiedEmail: "false",
	})
	ctx := WithSignupGrantRiskInput(context.Background(), SignupGrantRiskInput{RemoteIP: "203.0.114.20"})

	// 第一次：linuxdo（非白名单）force=true → 放行
	_, claim1 := svc.applySignupGrantRiskControlEx(ctx, "firstbind-a@fb.test", "linuxdo", signupGrantPlan{Balance: 5.0}, true)
	require.NotNil(t, claim1)
	require.False(t, claim1.Blocked, "first same-IP first-bind grant should be allowed")

	// 第二次：同 IP、不同邮箱 → 命中 ip_daily_limit → 剥夺
	filtered2, claim2 := svc.applySignupGrantRiskControlEx(ctx, "firstbind-b@fb.test", "linuxdo", signupGrantPlan{Balance: 5.0}, true)
	require.NotNil(t, claim2)
	require.True(t, claim2.Blocked)
	require.Equal(t, "ip_daily_limit", claim2.Reason)
	require.Equal(t, 0.0, filtered2.Balance)
}

// svcFirstBindUserRepo 是首绑集成测试专用的最小 UserRepository：GetByID 从真实 Postgres 读取用户行
// （首绑路径用它取邮箱），其余方法返回零值/错误。不依赖 internal/repository（避免 import cycle）。
type svcFirstBindUserRepo struct {
	db *sql.DB
}

func (r *svcFirstBindUserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	if r == nil || r.db == nil {
		return nil, ErrUserNotFound
	}
	u := &User{}
	var balance, paid, gift float64
	err := r.db.QueryRowContext(ctx, `
SELECT id, COALESCE(email,''), COALESCE(balance,0), COALESCE(paid_balance,0), COALESCE(gift_balance,0)
FROM users WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&u.ID, &u.Email, &balance, &paid, &gift)
	if err != nil {
		return nil, ErrUserNotFound
	}
	u.Balance, u.PaidBalance, u.GiftBalance = balance, paid, gift
	return u, nil
}

// SetGiftBalanceComponent 镜像 internal/repository 的 SET 语义
// （gift_balance = $amount, paid_balance = GREATEST(balance - $amount, 0)），
// 使本 stub 满足 signupGrantGiftBalanceSetter 接口。这样首绑路径在 markGift=true（旧的双记 bug）
// 时会真的 SET gift_balance，随后 ApplyBalanceChangeCtx 的 ADD 再 +B → gift_balance=2B，
// 集成测试从而能捕获双记回归；markGift=false（修复）则跳过本方法，gift_balance 只由 ADD 写一次。
func (r *svcFirstBindUserRepo) SetGiftBalanceComponent(ctx context.Context, id int64, amount float64) error {
	if r == nil || r.db == nil || amount <= 0 {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE users
SET gift_balance = $1,
	paid_balance = GREATEST(balance - $1, 0),
	updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL`, amount, id)
	return err
}

func (r *svcFirstBindUserRepo) Create(context.Context, *User) error                       { return ErrServiceUnavailable }
func (r *svcFirstBindUserRepo) GetByIDIncludeDeleted(context.Context, int64) (*User, error) { return nil, ErrUserNotFound }
func (r *svcFirstBindUserRepo) GetByEmail(context.Context, string) (*User, error)          { return nil, ErrUserNotFound }
func (r *svcFirstBindUserRepo) GetFirstAdmin(context.Context) (*User, error)               { return nil, ErrUserNotFound }
func (r *svcFirstBindUserRepo) Update(context.Context, *User) error                        { return ErrServiceUnavailable }
func (r *svcFirstBindUserRepo) Delete(context.Context, int64) error                        { return ErrServiceUnavailable }
func (r *svcFirstBindUserRepo) GetUserAvatar(context.Context, int64) (*UserAvatar, error)  { return nil, ErrUserNotFound }
func (r *svcFirstBindUserRepo) UpsertUserAvatar(context.Context, int64, UpsertUserAvatarInput) (*UserAvatar, error) {
	return nil, ErrServiceUnavailable
}
func (r *svcFirstBindUserRepo) DeleteUserAvatar(context.Context, int64) error { return ErrServiceUnavailable }
func (r *svcFirstBindUserRepo) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *svcFirstBindUserRepo) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *svcFirstBindUserRepo) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	return nil, nil
}
func (r *svcFirstBindUserRepo) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) { return nil, nil }
func (r *svcFirstBindUserRepo) UpdateUserLastActiveAt(context.Context, int64, time.Time) error     { return nil }
func (r *svcFirstBindUserRepo) UpdateBalance(context.Context, int64, float64) error                { return nil }
func (r *svcFirstBindUserRepo) DeductBalance(context.Context, int64, float64) error                { return nil }
func (r *svcFirstBindUserRepo) UpdateConcurrency(context.Context, int64, int) error                { return nil }
func (r *svcFirstBindUserRepo) BatchSetConcurrency(context.Context, []int64, int) (int, error)     { return 0, nil }
func (r *svcFirstBindUserRepo) BatchAddConcurrency(context.Context, []int64, int) (int, error)     { return 0, nil }
func (r *svcFirstBindUserRepo) ExistsByEmail(context.Context, string) (bool, error)                { return false, nil }
func (r *svcFirstBindUserRepo) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) { return 0, nil }
func (r *svcFirstBindUserRepo) AddGroupToAllowedGroups(context.Context, int64, int64) error        { return nil }
func (r *svcFirstBindUserRepo) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	return nil
}
func (r *svcFirstBindUserRepo) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	return nil, nil
}
func (r *svcFirstBindUserRepo) UnbindUserAuthProvider(context.Context, int64, string) error { return nil }
func (r *svcFirstBindUserRepo) UpdateTotpSecret(context.Context, int64, *string) error      { return nil }
func (r *svcFirstBindUserRepo) EnableTotp(context.Context, int64) error                    { return nil }
func (r *svcFirstBindUserRepo) DisableTotp(context.Context, int64) error                   { return nil }

// newSvcFirstBindAuthService 构造一个带真实 entClient + 最小 userRepo 的 AuthService，
// 用于端到端验证首绑奖励路径的余额/gift 不变式。
func newSvcFirstBindAuthService(t *testing.T, settings map[string]string) *AuthService {
	ensureSvcIntegrationDB(t)
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "svc-integration-secret"}}
	settingService := NewSettingService(&svcSettingRepoStub{values: settings}, cfg)
	return &AuthService{
		entClient:      svcIntegrationEnt,
		userRepo:       &svcFirstBindUserRepo{db: svcIntegrationDB},
		settingService: settingService,
		cfg:            cfg,
	}
}

// readUserBalanceFromSvcDB 从真实 Postgres 读取用户的 balance/gift_balance/paid_balance。
func readUserBalanceFromSvcDB(t *testing.T, userID int64) (balance, gift, paid float64) {
	t.Helper()
	err := svcIntegrationDB.QueryRowContext(context.Background(), `
SELECT COALESCE(balance,0), COALESCE(gift_balance,0), COALESCE(paid_balance,0)
FROM users WHERE id = $1`, userID).Scan(&balance, &gift, &paid)
	require.NoError(t, err)
	return balance, gift, paid
}

// TestIntegration_FirstBind_Allowed_GiftBalanceNotDoubleCounted
// 回归 #1 引入的 gift_balance 双记 bug：首绑放行时，余额由 ApplyBalanceChangeCtx 的 ADD+GiftDelta
// 追加（本测试 ledgerService=nil → 走 ADD 回退：balance += B, gift_balance += B），gift 组件由该 ADD
// 单独负责；attachSignupGrantClaim(markGift=false) 不再 SET gift_balance。
// 断言放行后：gift_balance == B（非 2B）、balance == B、paid_balance == 0、balance == paid + gift。
func TestIntegration_FirstBind_Allowed_GiftBalanceNotDoubleCounted(t *testing.T) {
	const grantBalance = 5.0
	svc := newSvcFirstBindAuthService(t, map[string]string{
		// google 渠道首绑发放 5 余额
		SettingKeyAuthSourceDefaultGoogleBalance:          strconv.FormatFloat(grantBalance, 'f', 8, 64),
		SettingKeyAuthSourceDefaultGoogleGrantOnFirstBind: "true",
		SettingKeyDefaultBalance:                          strconv.FormatFloat(grantBalance, 'f', 8, 64),
		// 风控开启但限额放开，确保首绑放行（claim allowed → 走 attachSignupGrantClaim markGift=false 路径）
		SettingKeySignupGrantRiskControlEnabled:              "true",
		SettingKeySignupGrantRiskControlIPLimit:              "50",
		SettingKeySignupGrantRiskControlEmailLimit:           "50",
		SettingKeySignupGrantRiskControlDomainLimit:          "50",
		SettingKeySignupGrantRiskControlDeviceLimit:          "50",
		SettingKeySignupGrantRiskControlRequireVerifiedEmail: "false",
	})

	// 建一个零余额真实用户（public_id 受 VARCHAR(32) 限制，用 unixnano 十六进制缩短）
	email := fmt.Sprintf("firstbind-gift-%d@fb.test", time.Now().UnixNano())
	publicID := fmt.Sprintf("fb%x", time.Now().UnixNano())
	var userID int64
	err := svcIntegrationDB.QueryRowContext(context.Background(), `
INSERT INTO users (public_id, email, password_hash, role, balance, paid_balance, gift_balance, created_at, updated_at)
VALUES ($1, $2, 'hash', 'user', 0, 0, 0, NOW(), NOW()) RETURNING id`,
		publicID, email).Scan(&userID)
	require.NoError(t, err)

	// 唯一 IP + User-Agent，避免与其它共享容器的集成测试在 ip/device 维度累计计数而误判。
	ctx := WithSignupGrantRiskInput(context.Background(), SignupGrantRiskInput{
		RemoteIP:   "203.0.115.30",
		UserAgent:  fmt.Sprintf("GiftDoubleCountTest/%d", time.Now().UnixNano()),
	})
	require.NoError(t, svc.ApplyProviderDefaultSettingsOnFirstBind(ctx, userID, "google", email, true))

	balance, gift, paid := readUserBalanceFromSvcDB(t, userID)
	require.Equal(t, grantBalance, balance, "balance should be granted exactly once")
	require.Equal(t, grantBalance, gift, "gift_balance must not be double-counted (== B, not 2B)")
	require.Equal(t, 0.0, paid, "paid_balance stays 0 for a pure gift grant")
	require.Equal(t, balance, paid+gift, "invariant: balance == paid_balance + gift_balance")
}
