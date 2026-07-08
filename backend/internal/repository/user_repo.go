package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/Aias00/cloudbase/ent/apikey"
	"github.com/Aias00/cloudbase/ent/authidentity"
	"github.com/Aias00/cloudbase/ent/authidentitychannel"
	dbgroup "github.com/Aias00/cloudbase/ent/group"
	"github.com/Aias00/cloudbase/ent/identityadoptiondecision"
	"github.com/Aias00/cloudbase/ent/predicate"
	"github.com/Aias00/cloudbase/ent/schema/mixins"
	dbuser "github.com/Aias00/cloudbase/ent/user"
	"github.com/Aias00/cloudbase/ent/userallowedgroup"
	"github.com/Aias00/cloudbase/ent/usersubscription"
	"github.com/Aias00/cloudbase/internal/billing"
	"github.com/Aias00/cloudbase/internal/domain"
	"github.com/Aias00/cloudbase/internal/identity"
	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/lib/pq"

	entsql "entgo.io/ent/dialect/sql"
)

type userRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewUserRepository(client *dbent.Client, sqlDB *sql.DB) service.UserRepository {
	return newUserRepositoryWithSQL(client, sqlDB)
}

func newUserRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *userRepository {
	return &userRepository{client: client, sql: sqlq}
}

func (r *userRepository) Create(ctx context.Context, userIn *service.User) error {
	if userIn == nil {
		return nil
	}

	const maxPublicIDCreateAttempts = 5
	requestedPublicID := normalizePublicUserID(userIn.PublicID)
	if requestedPublicID != "" {
		userIn.PublicID = requestedPublicID
		return r.create(ctx, userIn)
	}

	var lastErr error
	for attempt := 0; attempt < maxPublicIDCreateAttempts; attempt++ {
		userIn.PublicID = identity.NewPublicUserID()
		err := r.create(ctx, userIn)
		if err == nil {
			return nil
		}
		if !isPublicIDUniqueConstraintViolation(err) {
			return err
		}
		lastErr = err
	}
	return fmt.Errorf("public user id collision after %d attempts: %w", maxPublicIDCreateAttempts, lastErr)
}

func (r *userRepository) create(ctx context.Context, userIn *service.User) error {
	if userIn == nil {
		return nil
	}
	rawSignupSource := strings.TrimSpace(userIn.SignupSource)

	// 统一使用 ent 的事务：保证用户与允许分组的更新原子化，
	// 并避免基于 *sql.Tx 手动构造 ent client 导致的 ExecQuerier 断言错误。
	tx, err := r.client.Tx(ctx)
	if err != nil && !errors.Is(err, dbent.ErrTxStarted) {
		return err
	}

	var txClient *dbent.Client
	txCtx := ctx
	if err == nil {
		defer func() { _ = tx.Rollback() }()
		txClient = tx.Client()
		txCtx = dbent.NewTxContext(ctx, tx)
	} else {
		// 已处于外部事务中（ErrTxStarted），复用当前事务 client 并由调用方负责提交/回滚。
		if existingTx := dbent.TxFromContext(ctx); existingTx != nil {
			txClient = existingTx.Client()
		} else {
			txClient = r.client
		}
	}

	signupSource := userSignupSourceOrDefault(userIn.SignupSource)
	publicID := normalizePublicUserID(userIn.PublicID)
	if publicID == "" {
		publicID = identity.NewPublicUserID()
	}

	releaseEmailLock, err := lockRepositoryScopedKeys(
		txCtx,
		txClient,
		txAwareSQLExecutor(txCtx, r.sql, r.client),
		normalizedEmailUniquenessLockKey(userIn.Email),
	)
	if err != nil {
		return err
	}
	defer releaseEmailLock()

	if err := ensureNormalizedEmailAvailableWithClient(txCtx, txClient, 0, userIn.Email, signupSource); err != nil {
		return err
	}

	created, err := txClient.User.Create().
		SetPublicID(publicID).
		SetEmail(userIn.Email).
		SetUsername(userIn.Username).
		SetNotes(userIn.Notes).
		SetPasswordHash(userIn.PasswordHash).
		SetRole(userIn.Role).
		SetBalance(userIn.Balance).
		SetConcurrency(userIn.Concurrency).
		SetStatus(userIn.Status).
		SetSignupSource(signupSource).
		SetNillableLastLoginAt(userIn.LastLoginAt).
		SetNillableLastActiveAt(userIn.LastActiveAt).
		SetLoginAgreementAcceptedRevision(userIn.LoginAgreementAcceptedRevision).
		SetNillableLoginAgreementAcceptedAt(userIn.LoginAgreementAcceptedAt).
		SetRpmLimit(userIn.RPMLimit).
		Save(txCtx)
	if err != nil {
		if isPublicIDUniqueConstraintViolation(err) {
			return err
		}
		return translatePersistenceError(err, nil, identity.ErrEmailExists)
	}

	if err := r.syncUserAllowedGroupsWithClient(txCtx, txClient, created.ID, userIn.AllowedGroups); err != nil {
		return err
	}
	if userIn.Balance > 0 {
		exec := txAwareSQLExecutor(txCtx, r.sql, r.client)
		if exec == nil {
			return service.ErrServiceUnavailable
		}
		if rawSignupSource != "" {
			if err := setGiftBalanceComponentWithExec(txCtx, exec, created.ID, userIn.Balance); err != nil {
				return err
			}
		} else if err := setPaidBalanceComponentWithExec(txCtx, exec, created.ID, userIn.Balance); err != nil {
			return err
		}
	}
	if err := ensureEmailAuthIdentityWithClient(txCtx, txClient, created.ID, created.Email, "user_repo_create"); err != nil {
		return err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	applyUserEntityToService(userIn, created)
	if userIn.Balance > 0 {
		if rawSignupSource != "" {
			userIn.GiftBalance = userIn.Balance
			userIn.PaidBalance = 0
		} else {
			userIn.PaidBalance = userIn.Balance
			userIn.GiftBalance = 0
		}
	}
	return nil
}

func normalizePublicUserID(publicID string) string {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" || strings.HasPrefix(publicID, identity.PublicUserIDPrefix) {
		return publicID
	}
	return identity.PublicUserIDPrefix + publicID
}

func isPublicIDUniqueConstraintViolation(err error) bool {
	if err == nil || !isUniqueConstraintViolation(err) {
		return false
	}
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Constraint == "users_public_id_key"
	}
	return strings.Contains(strings.ToLower(err.Error()), "public_id")
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*service.User, error) {
	m, err := r.client.User.Query().Where(dbuser.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}

	out := userEntityToService(m)
	if err := r.hydrateUserBalanceBuckets(ctx, out); err != nil {
		return nil, err
	}
	groups, err := r.loadAllowedGroups(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	if v, ok := groups[id]; ok {
		out.AllowedGroups = v
	}
	return out, nil
}

func (r *userRepository) GetByIDIncludeDeleted(ctx context.Context, id int64) (*service.User, error) {
	ctx = mixins.SkipSoftDelete(ctx)
	m, err := r.client.User.Query().Where(dbuser.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	out := userEntityToService(m)
	if err := r.hydrateUserBalanceBuckets(ctx, out); err != nil {
		return nil, err
	}
	groups, err := r.loadAllowedGroups(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	if v, ok := groups[id]; ok {
		out.AllowedGroups = v
	}
	return out, nil
}

func (r *userRepository) GetByPublicID(ctx context.Context, publicID string, includeDeleted bool) (*service.User, error) {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return nil, identity.ErrUserNotFound
	}
	if includeDeleted {
		ctx = mixins.SkipSoftDelete(ctx)
	}
	m, err := r.client.User.Query().Where(dbuser.PublicIDEQ(publicID)).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	out := userEntityToService(m)
	if err := r.hydrateUserBalanceBuckets(ctx, out); err != nil {
		return nil, err
	}
	groups, err := r.loadAllowedGroups(ctx, []int64{m.ID})
	if err != nil {
		return nil, err
	}
	if v, ok := groups[m.ID]; ok {
		out.AllowedGroups = v
	}
	return out, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*service.User, error) {
	matches, err := r.client.User.Query().
		Where(userEmailLookupPredicate(email)).
		Order(dbent.Asc(dbuser.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, identity.ErrUserNotFound
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("normalized email lookup matched multiple users for %q", strings.TrimSpace(email))
	}
	m := matches[0]

	out := userEntityToService(m)
	if err := r.hydrateUserBalanceBuckets(ctx, out); err != nil {
		return nil, err
	}
	groups, err := r.loadAllowedGroups(ctx, []int64{m.ID})
	if err != nil {
		return nil, err
	}
	if v, ok := groups[m.ID]; ok {
		out.AllowedGroups = v
	}
	return out, nil
}

func (r *userRepository) GetByEmailAndSignupSource(ctx context.Context, email string, signupSource string) (*service.User, error) {
	m, err := r.client.User.Query().
		Where(
			userNormalizedEmailLookupPredicate(email),
			dbuser.SignupSourceEQ(userSignupSourceOrDefault(signupSource)),
		).
		Order(dbent.Asc(dbuser.FieldID)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, identity.ErrUserNotFound
		}
		return nil, err
	}

	out := userEntityToService(m)
	groups, err := r.loadAllowedGroups(ctx, []int64{m.ID})
	if err != nil {
		return nil, err
	}
	if v, ok := groups[m.ID]; ok {
		out.AllowedGroups = v
	}
	return out, nil
}

func (r *userRepository) Update(ctx context.Context, userIn *service.User) error {
	if userIn == nil {
		return nil
	}

	// 使用 ent 事务包裹用户更新与 allowed_groups 同步，避免跨层事务不一致。
	tx, err := r.client.Tx(ctx)
	if err != nil && !errors.Is(err, dbent.ErrTxStarted) {
		return err
	}

	var txClient *dbent.Client
	txCtx := ctx
	if err == nil {
		defer func() { _ = tx.Rollback() }()
		txClient = tx.Client()
		txCtx = dbent.NewTxContext(ctx, tx)
	} else {
		// 已处于外部事务中（ErrTxStarted），复用当前事务 client 并由调用方负责提交/回滚。
		if existingTx := dbent.TxFromContext(ctx); existingTx != nil {
			txClient = existingTx.Client()
		} else {
			txClient = r.client
		}
	}

	existing, err := clientFromContext(txCtx, txClient).User.Get(txCtx, userIn.ID)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	oldEmail := existing.Email
	signupSource := userSignupSourceOrDefault(userIn.SignupSource)
	if strings.TrimSpace(userIn.SignupSource) == "" {
		signupSource = userSignupSourceOrDefault(existing.SignupSource)
	}

	releaseEmailLock, err := lockRepositoryScopedKeys(
		txCtx,
		txClient,
		txAwareSQLExecutor(txCtx, r.sql, r.client),
		normalizedEmailUniquenessLockKey(userIn.Email),
	)
	if err != nil {
		return err
	}
	defer releaseEmailLock()

	if err := ensureNormalizedEmailAvailableWithClient(txCtx, txClient, userIn.ID, userIn.Email, signupSource); err != nil {
		return err
	}

	updateOp := txClient.User.UpdateOneID(userIn.ID).
		SetEmail(userIn.Email).
		SetUsername(userIn.Username).
		SetNotes(userIn.Notes).
		SetPasswordHash(userIn.PasswordHash).
		SetRole(userIn.Role).
		SetBalance(userIn.Balance).
		SetConcurrency(userIn.Concurrency).
		SetStatus(userIn.Status).
		SetLoginAgreementAcceptedRevision(userIn.LoginAgreementAcceptedRevision).
		SetBalanceNotifyEnabled(userIn.BalanceNotifyEnabled).
		SetBalanceNotifyThresholdType(userIn.BalanceNotifyThresholdType).
		SetNillableBalanceNotifyThreshold(userIn.BalanceNotifyThreshold).
		SetBalanceNotifyExtraEmails(marshalExtraEmails(userIn.BalanceNotifyExtraEmails)).
		SetTotalRecharged(userIn.TotalRecharged).
		SetRpmLimit(userIn.RPMLimit).
		SetSignupSource(signupSource)
	if userIn.LastLoginAt != nil {
		updateOp = updateOp.SetLastLoginAt(*userIn.LastLoginAt)
	}
	if userIn.LastActiveAt != nil {
		updateOp = updateOp.SetLastActiveAt(*userIn.LastActiveAt)
	}
	if userIn.BalanceNotifyThreshold == nil {
		updateOp = updateOp.ClearBalanceNotifyThreshold()
	}
	if userIn.LoginAgreementAcceptedAt == nil {
		updateOp = updateOp.ClearLoginAgreementAcceptedAt()
	} else {
		updateOp = updateOp.SetLoginAgreementAcceptedAt(*userIn.LoginAgreementAcceptedAt)
	}
	updated, err := updateOp.Save(txCtx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, identity.ErrEmailExists)
	}

	if err := r.syncUserAllowedGroupsWithClient(txCtx, txClient, updated.ID, userIn.AllowedGroups); err != nil {
		return err
	}
	if err := replaceEmailAuthIdentityWithClient(txCtx, txClient, updated.ID, oldEmail, updated.Email, "user_repo_update"); err != nil {
		return err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	userIn.UpdatedAt = updated.UpdatedAt
	return nil
}

func ensureEmailAuthIdentityWithClient(ctx context.Context, client *dbent.Client, userID int64, email string, source string) error {
	client = clientFromContext(ctx, client)
	if client == nil || userID <= 0 {
		return nil
	}

	subject := normalizeEmailAuthIdentitySubject(email)
	if subject == "" {
		return nil
	}

	if err := client.AuthIdentity.Create().
		SetUserID(userID).
		SetProviderType("email").
		SetProviderKey("email").
		SetProviderSubject(subject).
		SetVerifiedAt(time.Now().UTC()).
		SetMetadata(map[string]any{"source": source}).
		OnConflictColumns(
			authidentity.FieldProviderType,
			authidentity.FieldProviderKey,
			authidentity.FieldProviderSubject,
		).
		DoNothing().
		Exec(ctx); err != nil {
		if !isSQLNoRowsError(err) {
			return err
		}
	}

	identity, err := client.AuthIdentity.Query().
		Where(
			authidentity.ProviderTypeEQ("email"),
			authidentity.ProviderKeyEQ("email"),
			authidentity.ProviderSubjectEQ(subject),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil
		}
		return err
	}
	if identity.UserID != userID {
		return ErrAuthIdentityOwnershipConflict
	}
	return nil
}

func (r *userRepository) MarkEmailIdentitySupportsSignIn(ctx context.Context, userID int64, email string, source string) error {
	client := clientFromContext(ctx, r.client)
	if client == nil || userID <= 0 {
		return nil
	}

	subject := normalizeEmailAuthIdentitySubject(email)
	if subject == "" {
		return nil
	}

	metadata := map[string]any{"source": strings.TrimSpace(source)}

	if err := client.AuthIdentity.Create().
		SetUserID(userID).
		SetProviderType("email").
		SetProviderKey("email").
		SetProviderSubject(subject).
		SetVerifiedAt(time.Now().UTC()).
		SetMetadata(metadata).
		OnConflictColumns(
			authidentity.FieldProviderType,
			authidentity.FieldProviderKey,
			authidentity.FieldProviderSubject,
		).
		UpdateNewValues().
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func replaceEmailAuthIdentityWithClient(ctx context.Context, client *dbent.Client, userID int64, oldEmail, newEmail string, source string) error {
	newSubject := normalizeEmailAuthIdentitySubject(newEmail)
	if err := ensureEmailAuthIdentityWithClient(ctx, client, userID, newEmail, source); err != nil {
		return err
	}

	oldSubject := normalizeEmailAuthIdentitySubject(oldEmail)
	if oldSubject == "" || oldSubject == newSubject {
		return nil
	}

	_, err := clientFromContext(ctx, client).AuthIdentity.Delete().
		Where(
			authidentity.UserIDEQ(userID),
			authidentity.ProviderTypeEQ("email"),
			authidentity.ProviderKeyEQ("email"),
			authidentity.ProviderSubjectEQ(oldSubject),
		).
		Exec(ctx)
	return err
}

func normalizeEmailAuthIdentitySubject(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return ""
	}
	if strings.HasSuffix(normalized, ".invalid") {
		return ""
	}
	return normalized
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	// 复用 context 中已存在的事务（如 AdminService.DeleteUser 把删 Key 与删 User 包在同一事务中），
	// 由调用方负责提交/回滚，保证两者的原子性。
	if existingTx := dbent.TxFromContext(ctx); existingTx != nil {
		return r.deleteUser(ctx, existingTx.Client(), id)
	}

	tx, err := r.client.Tx(ctx)
	if err != nil && !errors.Is(err, dbent.ErrTxStarted) {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	exec := r.client
	if err == nil {
		defer func() { _ = tx.Rollback() }()
		exec = tx.Client()
	}
	// err == dbent.ErrTxStarted 时复用当前事务（exec = r.client）。

	if err := r.deleteUser(ctx, exec, id); err != nil {
		return err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return translatePersistenceError(err, identity.ErrUserNotFound, nil)
		}
	}
	return nil
}

// deleteUser 在给定 client（可能是外部事务 client）上删除用户及其身份关联记录，自身不开启/提交事务。
func (r *userRepository) deleteUser(ctx context.Context, exec *dbent.Client, id int64) error {
	identityIDs, err := exec.AuthIdentity.Query().
		Where(authidentity.UserIDEQ(id)).
		IDs(ctx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	if len(identityIDs) > 0 {
		if _, err := exec.IdentityAdoptionDecision.Update().
			Where(identityadoptiondecision.IdentityIDIn(identityIDs...)).
			ClearIdentityID().
			Save(ctx); err != nil {
			return translatePersistenceError(err, identity.ErrUserNotFound, nil)
		}
		if _, err := exec.AuthIdentityChannel.Delete().
			Where(authidentitychannel.IdentityIDIn(identityIDs...)).
			Exec(ctx); err != nil {
			return translatePersistenceError(err, identity.ErrUserNotFound, nil)
		}
		if _, err := exec.AuthIdentity.Delete().
			Where(authidentity.UserIDEQ(id)).
			Exec(ctx); err != nil {
			return translatePersistenceError(err, identity.ErrUserNotFound, nil)
		}
	}

	affected, err := exec.User.Delete().Where(dbuser.IDEQ(id)).Exec(ctx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	if affected == 0 {
		return identity.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, params pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	return r.ListWithFilters(ctx, params, identity.UserListFilters{})
}

func (r *userRepository) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters identity.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	// SkipSoftDelete 仅作用于 User 身份解析（下方 Count/All）；订阅、分组等关联实体沿用原始 ctx，避免穿透到这些同样带软删除的实体而带出已删除行。
	userCtx := ctx
	if filters.IncludeDeleted {
		userCtx = mixins.SkipSoftDelete(ctx)
	}

	q := r.client.User.Query()

	if filters.Status != "" {
		q = q.Where(dbuser.StatusEQ(filters.Status))
	}
	if filters.Role != "" {
		q = q.Where(dbuser.RoleEQ(filters.Role))
	}
	if filters.Search != "" {
		q = q.Where(
			dbuser.Or(
				dbuser.EmailContainsFold(filters.Search),
				dbuser.UsernameContainsFold(filters.Search),
				dbuser.NotesContainsFold(filters.Search),
				dbuser.HasAPIKeysWith(apikey.KeyContainsFold(filters.Search)),
			),
		)
	}

	if filters.GroupName != "" {
		q = q.Where(dbuser.HasAllowedGroupsWith(
			dbgroup.NameContainsFold(filters.GroupName),
		))
	}

	if filters.APIKeyGroupID > 0 {
		// 按"API Key 实际绑定的分组"过滤：用户只要有任意一个未软删除的 API Key
		// 绑定到该分组即命中（EXISTS 语义）。
		// 注意：SoftDeleteMixin 的拦截器不会自动下沉到 HasAPIKeysWith 子查询，
		// 必须显式加 apikey.DeletedAtIsNil()，否则已软删除的 key 会污染过滤结果。
		q = q.Where(dbuser.HasAPIKeysWith(
			apikey.GroupIDEQ(filters.APIKeyGroupID),
			apikey.DeletedAtIsNil(),
		))
	}

	// If attribute filters are specified, we need to filter by user IDs first
	var allowedUserIDs []int64
	if len(filters.Attributes) > 0 {
		var attrErr error
		allowedUserIDs, attrErr = r.filterUsersByAttributes(ctx, filters.Attributes)
		if attrErr != nil {
			return nil, nil, attrErr
		}
		if len(allowedUserIDs) == 0 {
			// No users match the attribute filters
			return []service.User{}, paginationResultFromTotal(0, params), nil
		}
		q = q.Where(dbuser.IDIn(allowedUserIDs...))
	}

	total, err := q.Clone().Count(userCtx)
	if err != nil {
		return nil, nil, err
	}

	usersQuery := q.
		Offset(params.Offset()).
		Limit(params.Limit())
	for _, order := range userListOrder(params) {
		usersQuery = usersQuery.Order(order)
	}

	users, err := usersQuery.All(userCtx)
	if err != nil {
		return nil, nil, err
	}

	outUsers := make([]service.User, 0, len(users))
	if len(users) == 0 {
		return outUsers, paginationResultFromTotal(int64(total), params), nil
	}

	userIDs := make([]int64, 0, len(users))
	userMap := make(map[int64]*service.User, len(users))
	for i := range users {
		userIDs = append(userIDs, users[i].ID)
		u := userEntityToService(users[i])
		outUsers = append(outUsers, *u)
		userMap[u.ID] = &outUsers[len(outUsers)-1]
	}
	if err := r.hydrateUserBalanceBucketsMap(ctx, userMap); err != nil {
		return nil, nil, err
	}

	shouldLoadSubscriptions := filters.IncludeSubscriptions == nil || *filters.IncludeSubscriptions
	if shouldLoadSubscriptions {
		// Batch load active subscriptions with groups to avoid N+1.
		subs, err := r.client.UserSubscription.Query().
			Where(
				usersubscription.UserIDIn(userIDs...),
				usersubscription.StatusEQ(service.SubscriptionStatusActive),
			).
			WithGroup().
			All(ctx)
		if err != nil {
			return nil, nil, err
		}

		for i := range subs {
			if u, ok := userMap[subs[i].UserID]; ok {
				u.Subscriptions = append(u.Subscriptions, *userSubscriptionEntityToService(subs[i]))
			}
		}
	}

	allowedGroupsByUser, err := r.loadAllowedGroups(ctx, userIDs)
	if err != nil {
		return nil, nil, err
	}
	for id, u := range userMap {
		if groups, ok := allowedGroupsByUser[id]; ok {
			u.AllowedGroups = groups
		}
	}

	return outUsers, paginationResultFromTotal(int64(total), params), nil
}

func userListOrder(params pagination.PaginationParams) []func(*entsql.Selector) {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortOrder := params.NormalizedSortOrder(pagination.SortOrderDesc)

	if sortBy == "last_used_at" {
		return userLastUsedAtOrder(sortOrder)
	}

	var field string
	defaultField := true
	nullsLastField := false
	switch sortBy {
	case "public_id":
		field = dbuser.FieldPublicID
		defaultField = false
	case "email":
		field = dbuser.FieldEmail
		defaultField = false
	case "username":
		field = dbuser.FieldUsername
		defaultField = false
	case "role":
		field = dbuser.FieldRole
		defaultField = false
	case "balance":
		field = dbuser.FieldBalance
		defaultField = false
	case "concurrency":
		field = dbuser.FieldConcurrency
		defaultField = false
	case "status":
		field = dbuser.FieldStatus
		defaultField = false
	case "created_at":
		field = dbuser.FieldCreatedAt
		defaultField = false
	case "last_active_at":
		field = dbuser.FieldLastActiveAt
		defaultField = false
		nullsLastField = true
	default:
		field = dbuser.FieldID
	}

	if sortOrder == pagination.SortOrderAsc {
		if defaultField && field == dbuser.FieldID {
			return []func(*entsql.Selector){dbent.Asc(dbuser.FieldID)}
		}
		if nullsLastField {
			return []func(*entsql.Selector){
				entsql.OrderByField(field, entsql.OrderNullsLast()).ToFunc(),
				dbent.Asc(dbuser.FieldID),
			}
		}
		return []func(*entsql.Selector){dbent.Asc(field), dbent.Asc(dbuser.FieldID)}
	}
	if defaultField && field == dbuser.FieldID {
		return []func(*entsql.Selector){dbent.Desc(dbuser.FieldID)}
	}
	if nullsLastField {
		return []func(*entsql.Selector){
			entsql.OrderByField(field, entsql.OrderDesc(), entsql.OrderNullsLast()).ToFunc(),
			dbent.Desc(dbuser.FieldID),
		}
	}
	return []func(*entsql.Selector){dbent.Desc(field), dbent.Desc(dbuser.FieldID)}
}

func (r *userRepository) GetLatestUsedAtByUserIDs(ctx context.Context, userIDs []int64) (map[int64]*time.Time, error) {
	result := make(map[int64]*time.Time, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	if r.sql == nil {
		return nil, fmt.Errorf("sql executor is not configured")
	}

	const query = `
		SELECT user_id, MAX(created_at) AS last_used_at
		FROM usage_logs
		WHERE user_id = ANY($1)
		GROUP BY user_id
	`

	rows, err := r.sql.QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			userID     int64
			lastUsedAt time.Time
		)
		if scanErr := rows.Scan(&userID, &lastUsedAt); scanErr != nil {
			return nil, scanErr
		}
		ts := lastUsedAt.UTC()
		result[userID] = &ts
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *userRepository) GetLatestUsedAtByUserID(ctx context.Context, userID int64) (*time.Time, error) {
	latestByUserID, err := r.GetLatestUsedAtByUserIDs(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}
	return latestByUserID[userID], nil
}

func userLastUsedAtOrder(sortOrder string) []func(*entsql.Selector) {
	orderExpr := func(direction, nulls string, tieOrder func(string) string) func(*entsql.Selector) {
		return func(s *entsql.Selector) {
			subquery := fmt.Sprintf("(SELECT MAX(created_at) FROM usage_logs WHERE user_id = %s)", s.C(dbuser.FieldID))
			s.OrderExpr(entsql.Expr(subquery + " " + direction + " NULLS " + nulls))
			s.OrderBy(tieOrder(s.C(dbuser.FieldID)))
		}
	}

	if sortOrder == pagination.SortOrderAsc {
		return []func(*entsql.Selector){
			orderExpr("ASC", "FIRST", entsql.Asc),
		}
	}
	return []func(*entsql.Selector){
		orderExpr("DESC", "LAST", entsql.Desc),
	}
}

// filterUsersByAttributes returns user IDs that match ALL the given attribute filters
func (r *userRepository) filterUsersByAttributes(ctx context.Context, attrs map[int64]string) ([]int64, error) {
	if len(attrs) == 0 {
		return nil, nil
	}

	if r.sql == nil {
		return nil, fmt.Errorf("sql executor is not configured")
	}

	clauses := make([]string, 0, len(attrs))
	args := make([]any, 0, len(attrs)*2+1)
	argIndex := 1
	for attrID, value := range attrs {
		clauses = append(clauses, fmt.Sprintf("(attribute_id = $%d AND value ILIKE $%d)", argIndex, argIndex+1))
		args = append(args, attrID, "%"+value+"%")
		argIndex += 2
	}

	query := fmt.Sprintf(
		`SELECT user_id
		 FROM user_attribute_values
		 WHERE %s
		 GROUP BY user_id
		 HAVING COUNT(DISTINCT attribute_id) = $%d`,
		strings.Join(clauses, " OR "),
		argIndex,
	)
	args = append(args, len(attrs))

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if scanErr := rows.Scan(&userID); scanErr != nil {
			return nil, scanErr
		}
		result = append(result, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *userRepository) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	if r == nil {
		return service.ErrServiceUnavailable
	}
	if amount == 0 {
		return nil
	}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	if exec == nil {
		return service.ErrServiceUnavailable
	}
	query := `
		UPDATE users
		SET balance = balance + $1,
			paid_balance = paid_balance + $1,
			total_recharged = CASE WHEN $1 > 0 THEN total_recharged + $1 ELSE total_recharged END,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	if amount < 0 {
		query = `
			UPDATE users
			SET balance = balance + $1,
				gift_balance = LEAST(gift_balance, GREATEST(balance + $1, 0)),
				paid_balance = (balance + $1) - LEAST(gift_balance, GREATEST(balance + $1, 0)),
				updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL
		`
	}
	res, err := exec.ExecContext(ctx, query, amount, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return identity.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) SetGiftBalanceComponent(ctx context.Context, id int64, amount float64) error {
	if r == nil || amount <= 0 {
		return nil
	}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	if exec == nil {
		return service.ErrServiceUnavailable
	}
	return setGiftBalanceComponentWithExec(ctx, exec, id, amount)
}

func setGiftBalanceComponentWithExec(ctx context.Context, exec sqlQueryExecutor, id int64, amount float64) error {
	res, err := exec.ExecContext(ctx, `
		UPDATE users
		SET gift_balance = $1,
			paid_balance = GREATEST(balance - $1, 0),
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, amount, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return identity.ErrUserNotFound
	}
	return nil
}

func setPaidBalanceComponentWithExec(ctx context.Context, exec sqlQueryExecutor, id int64, amount float64) error {
	res, err := exec.ExecContext(ctx, `
		UPDATE users
		SET paid_balance = $1,
			gift_balance = 0,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, amount, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return identity.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) AddGiftBalance(ctx context.Context, id int64, amount float64) error {
	if r == nil || amount <= 0 {
		return nil
	}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	if exec == nil {
		return service.ErrServiceUnavailable
	}
	res, err := exec.ExecContext(ctx, `
		UPDATE users
		SET balance = balance + $1,
			gift_balance = gift_balance + $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, amount, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return identity.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) ListSignupGrantRiskClaims(ctx context.Context, limit, offset int, filter service.SignupGrantRiskClaimFilter) ([]service.SignupGrantRiskClaimRecord, int64, error) {
	if r == nil || r.sql == nil {
		return nil, 0, service.ErrServiceUnavailable
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	clauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	if strings.TrimSpace(filter.Decision) != "" {
		args = append(args, strings.TrimSpace(filter.Decision))
		clauses = append(clauses, "c.decision = $"+strconv.Itoa(len(args)))
	}
	if filter.UserID > 0 {
		args = append(args, filter.UserID)
		clauses = append(clauses, "c.user_id = $"+strconv.Itoa(len(args)))
	}
	if strings.TrimSpace(filter.SubjectQuery) != "" {
		if rawColumn, hashColumn := signupGrantRiskClaimSubjectColumns(filter.SubjectType); rawColumn != "" && hashColumn != "" {
			query := strings.TrimSpace(filter.SubjectQuery)
			args = append(args, query, "%"+query+"%")
			exactIndex := strconv.Itoa(len(args) - 1)
			likeIndex := strconv.Itoa(len(args))
			clauses = append(clauses, "(c."+hashColumn+" = $"+exactIndex+" OR c."+rawColumn+" ILIKE $"+likeIndex+")")
		}
	}
	if strings.TrimSpace(filter.Reason) != "" {
		args = append(args, "%"+strings.TrimSpace(filter.Reason)+"%")
		clauses = append(clauses, "c.reason ILIKE $"+strconv.Itoa(len(args)))
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM signup_grant_claims c "+where, args, &total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(args, limit, offset)
	rows, err := r.sql.QueryContext(ctx, `
		SELECT c.id, c.user_id, COALESCE(u.public_id, ''), c.email, c.email_domain, c.ip_address,
		       c.email_hash, c.email_domain_hash, c.ip_hash, c.user_agent_hash,
		       c.signup_source, c.provider_type, c.provider_subject, c.provider_subject_hash, c.decision, c.reason,
		       c.grant_balance, c.created_at, c.updated_at
		FROM signup_grant_claims c
		LEFT JOIN users u ON u.id = c.user_id
		`+where+`
		ORDER BY c.created_at DESC
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2),
		queryArgs...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	records := make([]service.SignupGrantRiskClaimRecord, 0)
	for rows.Next() {
		var rec service.SignupGrantRiskClaimRecord
		var userID sql.NullInt64
		var userPublicID string
		if err := rows.Scan(
			&rec.ID, &userID, &userPublicID, &rec.Email, &rec.EmailDomain, &rec.IPAddress,
			&rec.EmailHash, &rec.EmailDomainHash, &rec.IPHash, &rec.UserAgentHash,
			&rec.SignupSource, &rec.ProviderType, &rec.ProviderSubject, &rec.ProviderSubjectHash, &rec.Decision, &rec.Reason,
			&rec.GrantBalance, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if userID.Valid {
			rec.UserID = &userID.Int64
		}
		rec.UserPublicID = userPublicID
		records = append(records, rec)
	}
	return records, total, rows.Err()
}

func signupGrantRiskClaimSubjectColumns(subjectType string) (string, string) {
	switch strings.TrimSpace(strings.ToLower(subjectType)) {
	case "email":
		return "email", "email_hash"
	case "email_domain":
		return "email_domain", "email_domain_hash"
	case "ip":
		return "ip_address", "ip_hash"
	case "oauth_identity":
		return "provider_subject", "provider_subject_hash"
	case "device":
		return "user_agent_hash", "user_agent_hash"
	default:
		return "", ""
	}
}

func (r *userRepository) GetSignupGrantRiskUserSummary(ctx context.Context, userID int64) (*service.SignupGrantRiskUserSummary, error) {
	if r == nil || r.sql == nil {
		return nil, service.ErrServiceUnavailable
	}
	summary := &service.SignupGrantRiskUserSummary{UserID: userID}
	var createdAt time.Time
	var updatedAt time.Time
	err := scanSingleRow(ctx, r.sql, `
		SELECT decision, reason, grant_balance, created_at, updated_at
		FROM signup_grant_claims
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, []any{userID}, &summary.Decision, &summary.Reason, &summary.GrantBalance, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return summary, nil
	}
	if err != nil {
		return nil, err
	}
	summary.HasClaim = true
	summary.CreatedAt = &createdAt
	summary.UpdatedAt = &updatedAt
	return summary, nil
}

func (r *userRepository) UpsertSignupGrantRiskOverride(ctx context.Context, input service.SignupGrantRiskOverrideInput) error {
	if r == nil || r.sql == nil {
		return service.ErrServiceUnavailable
	}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	_, err := exec.ExecContext(ctx, `
		INSERT INTO signup_grant_risk_overrides (subject_type, subject_value, subject_hash, action, reason, created_by, updated_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, 0), NOW())
		ON CONFLICT (subject_type, subject_hash)
		DO UPDATE SET action = EXCLUDED.action,
		              subject_value = EXCLUDED.subject_value,
		              reason = EXCLUDED.reason,
		              created_by = EXCLUDED.created_by,
		              updated_at = NOW()
	`, input.SubjectType, input.SubjectValue, input.Subject, input.Action, input.Reason, input.CreatedBy)
	return err
}

func (r *userRepository) DeleteSignupGrantRiskOverride(ctx context.Context, id int64, adminID int64) error {
	if r == nil || r.sql == nil {
		return service.ErrServiceUnavailable
	}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	var rec service.SignupGrantRiskOverrideRecord
	var createdBy sql.NullInt64
	var expiresAt sql.NullTime
	if err := scanSingleRow(ctx, exec, `
		SELECT id, subject_type, subject_value, subject_hash, action, reason, created_by, expires_at, created_at, updated_at
		FROM signup_grant_risk_overrides
		WHERE id = $1
	`, []any{id}, &rec.ID, &rec.SubjectType, &rec.SubjectValue, &rec.SubjectHash, &rec.Action, &rec.Reason, &createdBy, &expiresAt, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return infraerrors.NotFound("SIGNUP_GRANT_RISK_OVERRIDE_NOT_FOUND", "signup grant risk override not found")
		}
		return err
	}
	if createdBy.Valid {
		rec.CreatedBy = &createdBy.Int64
	}
	if expiresAt.Valid {
		rec.ExpiresAt = &expiresAt.Time
	}
	res, err := exec.ExecContext(ctx, `DELETE FROM signup_grant_risk_overrides WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return infraerrors.NotFound("SIGNUP_GRANT_RISK_OVERRIDE_NOT_FOUND", "signup grant risk override not found")
	}
	var adminIDPtr *int64
	if adminID > 0 {
		adminIDPtr = &adminID
	}
	if err := r.CreateSignupGrantAdminAuditLog(ctx, service.SignupGrantAdminAuditLog{
		Operation:    "risk_override_delete",
		SubjectType:  rec.SubjectType,
		SubjectValue: rec.SubjectValue,
		SubjectHash:  rec.SubjectHash,
		Action:       rec.Action,
		Reason:       rec.Reason,
		AdminID:      adminIDPtr,
		Metadata: map[string]any{
			"override_id": id,
		},
	}); err != nil {
		return err
	}
	return nil
}

func (r *userRepository) ListSignupGrantRiskOverrides(ctx context.Context, limit, offset int, filter service.SignupGrantRiskOverrideFilter) ([]service.SignupGrantRiskOverrideRecord, int64, error) {
	if r == nil || r.sql == nil {
		return nil, 0, service.ErrServiceUnavailable
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if strings.TrimSpace(filter.SubjectType) != "" {
		args = append(args, strings.TrimSpace(filter.SubjectType))
		clauses = append(clauses, "subject_type = $"+strconv.Itoa(len(args)))
	}
	if strings.TrimSpace(filter.Action) != "" {
		args = append(args, strings.TrimSpace(filter.Action))
		clauses = append(clauses, "action = $"+strconv.Itoa(len(args)))
	}
	if strings.TrimSpace(filter.SubjectQuery) != "" {
		query := strings.TrimSpace(filter.SubjectQuery)
		args = append(args, query, "%"+query+"%")
		exactIndex := strconv.Itoa(len(args) - 1)
		likeIndex := strconv.Itoa(len(args))
		clauses = append(clauses, "(subject_hash = $"+exactIndex+" OR subject_value ILIKE $"+likeIndex+")")
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM signup_grant_risk_overrides "+where, args, &total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(args, limit, offset)
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, subject_type, subject_value, subject_hash, action, reason, created_by, expires_at, created_at, updated_at
		FROM signup_grant_risk_overrides
		`+where+`
		ORDER BY updated_at DESC
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2),
		queryArgs...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	records := make([]service.SignupGrantRiskOverrideRecord, 0)
	for rows.Next() {
		var rec service.SignupGrantRiskOverrideRecord
		var createdBy sql.NullInt64
		var expiresAt sql.NullTime
		if err := rows.Scan(
			&rec.ID, &rec.SubjectType, &rec.SubjectValue, &rec.SubjectHash, &rec.Action, &rec.Reason,
			&createdBy, &expiresAt, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if createdBy.Valid {
			rec.CreatedBy = &createdBy.Int64
		}
		if expiresAt.Valid {
			rec.ExpiresAt = &expiresAt.Time
		}
		records = append(records, rec)
	}
	return records, total, rows.Err()
}

func (r *userRepository) CreateSignupGrantAdminAuditLog(ctx context.Context, input service.SignupGrantAdminAuditLog) error {
	if r == nil || r.sql == nil {
		return service.ErrServiceUnavailable
	}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	var metadata []byte
	var err error
	if input.Metadata != nil {
		metadata, err = json.Marshal(input.Metadata)
		if err != nil {
			return err
		}
	} else {
		metadata = []byte(`{}`)
	}
	var targetUserID any
	if input.TargetUserID != nil && *input.TargetUserID > 0 {
		targetUserID = *input.TargetUserID
	}
	var adminID any
	if input.AdminID != nil && *input.AdminID > 0 {
		adminID = *input.AdminID
	}
	_, err = exec.ExecContext(ctx, `
		INSERT INTO signup_grant_admin_audit_logs
			(operation, target_user_id, subject_type, subject_value, subject_hash, action, amount, reason, admin_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
	`, strings.TrimSpace(input.Operation), targetUserID, strings.TrimSpace(input.SubjectType), strings.TrimSpace(input.SubjectValue), strings.TrimSpace(input.SubjectHash),
		strings.TrimSpace(input.Action), input.Amount, strings.TrimSpace(input.Reason), adminID, string(metadata))
	return err
}

func (r *userRepository) ListSignupGrantAdminAuditLogs(ctx context.Context, limit, offset int, filter service.SignupGrantAdminAuditLogFilter) ([]service.SignupGrantAdminAuditLog, int64, error) {
	if r == nil || r.sql == nil {
		return nil, 0, service.ErrServiceUnavailable
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if strings.TrimSpace(filter.Operation) != "" {
		args = append(args, strings.TrimSpace(filter.Operation))
		clauses = append(clauses, "a.operation = $"+strconv.Itoa(len(args)))
	}
	if filter.AdminID > 0 {
		args = append(args, filter.AdminID)
		clauses = append(clauses, "a.admin_id = $"+strconv.Itoa(len(args)))
	}
	if filter.TargetUserID > 0 {
		args = append(args, filter.TargetUserID)
		clauses = append(clauses, "a.target_user_id = $"+strconv.Itoa(len(args)))
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	var total int64
	if err := scanSingleRow(ctx, r.sql, "SELECT COUNT(*) FROM signup_grant_admin_audit_logs a "+where, args, &total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(args, limit, offset)
	rows, err := r.sql.QueryContext(ctx, `
		SELECT a.id, a.operation, a.target_user_id, COALESCE(u.public_id, ''), a.subject_type, a.subject_value, a.subject_hash, a.action, a.amount, a.reason, a.admin_id, a.metadata, a.created_at
		FROM signup_grant_admin_audit_logs a
		LEFT JOIN users u ON u.id = a.target_user_id
		`+where+`
		ORDER BY a.created_at DESC
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2),
		queryArgs...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	records := make([]service.SignupGrantAdminAuditLog, 0)
	for rows.Next() {
		var rec service.SignupGrantAdminAuditLog
		var targetUserID sql.NullInt64
		var targetUserPublicID string
		var adminID sql.NullInt64
		var metadataRaw []byte
		if err := rows.Scan(
			&rec.ID, &rec.Operation, &targetUserID, &targetUserPublicID, &rec.SubjectType, &rec.SubjectValue, &rec.SubjectHash,
			&rec.Action, &rec.Amount, &rec.Reason, &adminID, &metadataRaw, &rec.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if targetUserID.Valid {
			rec.TargetUserID = &targetUserID.Int64
		}
		rec.TargetUserPublicID = targetUserPublicID
		if adminID.Valid {
			rec.AdminID = &adminID.Int64
		}
		if len(metadataRaw) > 0 {
			_ = json.Unmarshal(metadataRaw, &rec.Metadata)
		}
		records = append(records, rec)
	}
	return records, total, rows.Err()
}

func (r *userRepository) hydrateUserBalanceBuckets(ctx context.Context, user *service.User) error {
	if user == nil {
		return nil
	}
	users := map[int64]*service.User{user.ID: user}
	return r.hydrateUserBalanceBucketsMap(ctx, users)
}

func (r *userRepository) hydrateUserBalanceBucketsMap(ctx context.Context, users map[int64]*service.User) error {
	if r == nil || r.sql == nil || len(users) == 0 {
		return nil
	}
	for id, user := range users {
		var paidBalance, giftBalance float64
		if err := scanSingleRow(ctx, r.sql, `
			SELECT paid_balance, gift_balance
			FROM users
			WHERE id = $1
		`, []any{id}, &paidBalance, &giftBalance); err != nil {
			return err
		}
		user.PaidBalance = paidBalance
		user.GiftBalance = giftBalance
	}
	return nil
}

// DeductBalance 扣除用户余额
// 透支策略：允许余额变为负数，确保当前请求能够完成
// 中间件会阻止余额 <= 0 的用户发起后续请求
func (r *userRepository) DeductBalance(ctx context.Context, id int64, amount float64) error {
	if amount == 0 {
		return nil
	}
	return r.UpdateBalance(ctx, id, -amount)
}

// DeductBalanceIfEnough deducts balance only when the current balance can cover the amount.
func (r *userRepository) DeductBalanceIfEnough(ctx context.Context, id int64, amount float64) error {
	client := clientFromContext(ctx, r.client)
	n, err := client.User.Update().
		Where(dbuser.IDEQ(id), dbuser.BalanceGTE(amount)).
		AddBalance(-amount).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	if n > 0 {
		return nil
	}
	exists, err := client.User.Query().Where(dbuser.IDEQ(id)).Exist(ctx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	if !exists {
		return identity.ErrUserNotFound
	}
	return billing.ErrInsufficientBalance
}

func (r *userRepository) UpdateConcurrency(ctx context.Context, id int64, amount int) error {
	client := clientFromContext(ctx, r.client)
	n, err := client.User.Update().Where(dbuser.IDEQ(id)).AddConcurrency(amount).Save(ctx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	if n == 0 {
		return identity.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) BatchSetConcurrency(ctx context.Context, userIDs []int64, value int) (int, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	if value < 0 {
		value = 0
	}
	res, err := r.sql.ExecContext(ctx,
		"UPDATE users SET concurrency = $1, updated_at = NOW() WHERE id = ANY($2) AND deleted_at IS NULL",
		value, pq.Array(userIDs))
	if err != nil {
		return 0, fmt.Errorf("batch set concurrency: %w", err)
	}
	affected, _ := res.RowsAffected()
	return int(affected), nil
}

func (r *userRepository) BatchAddConcurrency(ctx context.Context, userIDs []int64, delta int) (int, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	res, err := r.sql.ExecContext(ctx,
		"UPDATE users SET concurrency = GREATEST(concurrency + $1, 0), updated_at = NOW() WHERE id = ANY($2) AND deleted_at IS NULL",
		delta, pq.Array(userIDs))
	if err != nil {
		return 0, fmt.Errorf("batch add concurrency: %w", err)
	}
	affected, _ := res.RowsAffected()
	return int(affected), nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.client.User.Query().Where(userEmailLookupPredicate(email)).Exist(ctx)
}

func (r *userRepository) ExistsByEmailAndSignupSource(ctx context.Context, email string, signupSource string) (bool, error) {
	return r.client.User.Query().
		Where(
			userNormalizedEmailLookupPredicate(email),
			dbuser.SignupSourceEQ(userSignupSourceOrDefault(signupSource)),
		).
		Exist(ctx)
}

func ensureNormalizedEmailAvailableWithClient(ctx context.Context, client *dbent.Client, userID int64, email string, signupSource string) error {
	client = clientFromContext(ctx, client)
	if client == nil {
		return nil
	}

	matches, err := client.User.Query().
		Where(userNormalizedEmailLookupPredicate(email)).
		All(ctx)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if match.ID == userID {
			continue
		}
		return identity.ErrEmailExists
	}
	return nil
}

func userEmailLookupPredicate(email string) predicate.User {
	return userNormalizedEmailLookupPredicate(email)
}

func userNormalizedEmailLookupPredicate(email string) predicate.User {
	normalized := normalizeEmailLookupValue(email)
	if normalized == "" {
		return dbuser.EmailEQ(email)
	}
	return predicate.User(func(s *entsql.Selector) {
		s.Where(entsql.P(func(b *entsql.Builder) {
			b.WriteString("LOWER(TRIM(").
				Ident(s.C(dbuser.FieldEmail)).
				WriteString(")) = ").
				Arg(normalized)
		}))
	})
}

func normalizeEmailLookupValue(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizedEmailUniquenessLockKey(email string) string {
	normalized := normalizeEmailLookupValue(email)
	if normalized == "" {
		return ""
	}
	return "users:normalized-email:" + normalized
}

func (r *userRepository) AddGroupToAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	client := clientFromContext(ctx, r.client)
	err := client.UserAllowedGroup.Create().
		SetUserID(userID).
		SetGroupID(groupID).
		OnConflictColumns(userallowedgroup.FieldUserID, userallowedgroup.FieldGroupID).
		DoNothing().
		Exec(ctx)
	if isSQLNoRowsError(err) {
		return nil
	}
	return err
}

func (r *userRepository) RemoveGroupFromAllowedGroups(ctx context.Context, groupID int64) (int64, error) {
	// 仅操作 user_allowed_groups 联接表，legacy users.allowed_groups 列已弃用。
	affected, err := r.client.UserAllowedGroup.Delete().
		Where(userallowedgroup.GroupIDEQ(groupID)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return int64(affected), nil
}

// RemoveGroupFromUserAllowedGroups 移除单个用户的指定分组权限
func (r *userRepository) RemoveGroupFromUserAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.UserAllowedGroup.Delete().
		Where(userallowedgroup.UserIDEQ(userID), userallowedgroup.GroupIDEQ(groupID)).
		Exec(ctx)
	return err
}

func (r *userRepository) GetFirstAdmin(ctx context.Context) (*service.User, error) {
	m, err := r.client.User.Query().
		Where(
			dbuser.RoleEQ(domain.RoleAdmin),
			dbuser.StatusEQ(domain.StatusActive),
		).
		Order(dbent.Asc(dbuser.FieldID)).
		First(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}

	out := userEntityToService(m)
	groups, err := r.loadAllowedGroups(ctx, []int64{m.ID})
	if err != nil {
		return nil, err
	}
	if v, ok := groups[m.ID]; ok {
		out.AllowedGroups = v
	}
	return out, nil
}

func (r *userRepository) loadAllowedGroups(ctx context.Context, userIDs []int64) (map[int64][]int64, error) {
	out := make(map[int64][]int64, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}

	rows, err := r.client.UserAllowedGroup.Query().
		Where(userallowedgroup.UserIDIn(userIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	for i := range rows {
		out[rows[i].UserID] = append(out[rows[i].UserID], rows[i].GroupID)
	}

	for userID := range out {
		sort.Slice(out[userID], func(i, j int) bool { return out[userID][i] < out[userID][j] })
	}

	return out, nil
}

// syncUserAllowedGroupsWithClient 在 ent client/事务内同步用户允许分组：
// 仅操作 user_allowed_groups 联接表，legacy users.allowed_groups 列已弃用。
func (r *userRepository) syncUserAllowedGroupsWithClient(ctx context.Context, client *dbent.Client, userID int64, groupIDs []int64) error {
	if client == nil {
		return nil
	}

	// Keep join table as the source of truth for reads.
	if _, err := client.UserAllowedGroup.Delete().Where(userallowedgroup.UserIDEQ(userID)).Exec(ctx); err != nil {
		return err
	}

	unique := make(map[int64]struct{}, len(groupIDs))
	for _, id := range groupIDs {
		if id <= 0 {
			continue
		}
		unique[id] = struct{}{}
	}

	if len(unique) > 0 {
		creates := make([]*dbent.UserAllowedGroupCreate, 0, len(unique))
		for groupID := range unique {
			creates = append(creates, client.UserAllowedGroup.Create().SetUserID(userID).SetGroupID(groupID))
		}
		if err := client.UserAllowedGroup.
			CreateBulk(creates...).
			OnConflictColumns(userallowedgroup.FieldUserID, userallowedgroup.FieldGroupID).
			DoNothing().
			Exec(ctx); err != nil {
			if isSQLNoRowsError(err) {
				return nil
			}
			return err
		}
	}

	return nil
}

func applyUserEntityToService(dst *service.User, src *dbent.User) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.PublicID = src.PublicID
	dst.SignupSource = src.SignupSource
	dst.LastLoginAt = src.LastLoginAt
	dst.LastActiveAt = src.LastActiveAt
	dst.WelcomeEmailSentAt = src.WelcomeEmailSentAt
	dst.MarketingEmailsUnsubscribedAt = src.MarketingEmailsUnsubscribedAt
	dst.LoginAgreementAcceptedRevision = src.LoginAgreementAcceptedRevision
	dst.LoginAgreementAcceptedAt = src.LoginAgreementAcceptedAt
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}

func userSignupSourceOrDefault(signupSource string) string {
	switch strings.TrimSpace(strings.ToLower(signupSource)) {
	case "", "email":
		return "email"
	case "github", "google":
		return strings.TrimSpace(strings.ToLower(signupSource))
	default:
		return "email"
	}
}

// marshalExtraEmails serializes notify email entries to JSON for storage.
func marshalExtraEmails(entries []service.NotifyEmailEntry) string {
	return service.MarshalNotifyEmails(entries)
}

// UpdateTotpSecret 更新用户的 TOTP 加密密钥
func (r *userRepository) UpdateTotpSecret(ctx context.Context, userID int64, encryptedSecret *string) error {
	client := clientFromContext(ctx, r.client)
	update := client.User.UpdateOneID(userID)
	if encryptedSecret == nil {
		update = update.ClearTotpSecretEncrypted()
	} else {
		update = update.SetTotpSecretEncrypted(*encryptedSecret)
	}
	_, err := update.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	return nil
}

// EnableTotp 启用用户的 TOTP 双因素认证
func (r *userRepository) EnableTotp(ctx context.Context, userID int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.User.UpdateOneID(userID).
		SetTotpEnabled(true).
		SetTotpEnabledAt(time.Now()).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	return nil
}

// DisableTotp 禁用用户的 TOTP 双因素认证
func (r *userRepository) DisableTotp(ctx context.Context, userID int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.User.UpdateOneID(userID).
		SetTotpEnabled(false).
		ClearTotpEnabledAt().
		ClearTotpSecretEncrypted().
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, identity.ErrUserNotFound, nil)
	}
	return nil
}
