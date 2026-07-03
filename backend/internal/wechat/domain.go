package wechat

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
)

const (
	SessionStatusPending       = "pending"
	SessionStatusScanConfirmed = "scan_confirmed"
	SessionStatusReady         = "ready"
	SessionStatusExpired       = "expired"

	ArticleSourceSynced     = "synced"
	ArticleSourceDirectLink = "direct_link"

	ExportTaskStatusQueued              = "queued"
	ExportTaskStatusRunning             = "running"
	ExportTaskStatusUploading           = "uploading"
	ExportTaskStatusCompleted           = "completed"
	ExportTaskStatusCompletedWithErrors = "completed_with_errors"
	ExportTaskStatusFailed              = "failed"
	ExportTaskStatusCancelled           = "cancelled"
)

var (
	ErrExportNotConfigured = infraerrors.InternalServer("WECHAT_EXPORT_NOT_CONFIGURED", "wechat export capability is not configured")
	ErrSessionNotFound     = infraerrors.NotFound("WECHAT_SESSION_NOT_FOUND", "wechat session not found")
	ErrSessionNotReady     = infraerrors.BadRequest("WECHAT_SESSION_NOT_READY", "wechat session is not ready")
	ErrAccountNotFound     = infraerrors.NotFound("WECHAT_ACCOUNT_NOT_FOUND", "wechat account not found")
	ErrArticleNotFound     = infraerrors.NotFound("WECHAT_ARTICLE_NOT_FOUND", "wechat article not found")
	ErrTaskNotFound        = infraerrors.NotFound("WECHAT_EXPORT_TASK_NOT_FOUND", "wechat export task not found")
	ErrTaskConflict        = infraerrors.Conflict("WECHAT_EXPORT_TASK_CONFLICT", "wechat export task is in a conflicting state")
	ErrInvalidInput        = infraerrors.BadRequest("WECHAT_EXPORT_INVALID_INPUT", "wechat export input is invalid")
	ErrInsufficientBalance = infraerrors.BadRequest("INSUFFICIENT_BALANCE", "insufficient balance for wechat export")
	ErrArticleVerifyPage   = infraerrors.BadRequest("WECHAT_ARTICLE_VERIFY_PAGE", "微信返回验证页，请通过公众号同步导入")
)

type Session struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	Status           string     `json:"status"`
	LoginToken       string     `json:"login_token,omitempty"`
	CookiesEncrypted string     `json:"-"`
	LoginAccountName string     `json:"login_account_name"`
	LastValidatedAt  *time.Time `json:"last_validated_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Account struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	FakeID       string     `json:"fakeid"`
	Nickname     string     `json:"nickname"`
	Alias        string     `json:"alias"`
	Avatar       string     `json:"avatar"`
	Description  string     `json:"description"`
	IsActive     bool       `json:"is_active"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Article struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	AccountFakeID  string     `json:"account_fakeid"`
	SourceType     string     `json:"source_type"`
	Title          string     `json:"title"`
	Author         string     `json:"author"`
	Link           string     `json:"link"`
	Cover          string     `json:"cover"`
	Digest         string     `json:"digest"`
	PublishAt      *time.Time `json:"publish_at,omitempty"`
	IsOriginal     bool       `json:"is_original"`
	IsPaySubscribe bool       `json:"is_pay_subscribe"`
	ContentStatus  string     `json:"content_status"`
	MetadataJSON   string     `json:"metadata_json"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ExportTask struct {
	ID                     int64          `json:"id"`
	UserID                 int64          `json:"user_id"`
	Status                 string         `json:"status"`
	ArticleIDs             []int64        `json:"article_ids"`
	Formats                []ExportFormat `json:"formats"`
	SelectedArticleCount   int            `json:"selected_article_count"`
	SuccessfulArticleCount int            `json:"successful_article_count"`
	FailedArticleCount     int            `json:"failed_article_count"`
	IncludeEngagement      bool           `json:"include_engagement"`
	PayloadJSON            string         `json:"payload_json,omitempty"`
	ResultManifestJSON     string         `json:"result_manifest_json"`
	ErrorMessage           string         `json:"error_message"`
	WorkerLeaseUntil       *time.Time     `json:"worker_lease_until,omitempty"`
	WorkerLeaseToken       string         `json:"-"`
	WorkerRunID            string         `json:"-"`
	RetentionDays          int            `json:"retention_days"`
	ExpiresAt              *time.Time     `json:"expires_at,omitempty"`
	CostEstimate           float64        `json:"cost_estimate"`
	BalanceSnapshot        float64        `json:"balance_snapshot"`
	ReservedPaidBalance    float64        `json:"-"`
	ReservedGiftBalance    float64        `json:"-"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

type ExportWorkerStatus struct {
	Health              string     `json:"health"`
	Message             string     `json:"message"`
	TotalCount          int64      `json:"total_count"`
	QueuedCount         int64      `json:"queued_count"`
	RunningCount        int64      `json:"running_count"`
	StaleRunningCount   int64      `json:"stale_running_count"`
	FailedCount         int64      `json:"failed_count"`
	CompletedCount      int64      `json:"completed_count"`
	CancelledCount      int64      `json:"cancelled_count"`
	LastTaskUpdatedAt   *time.Time `json:"last_task_updated_at,omitempty"`
	OldestQueuedAt      *time.Time `json:"oldest_queued_at,omitempty"`
	LastTaskAgeSeconds  *int64     `json:"last_task_age_seconds,omitempty"`
	OldestQueuedSeconds *int64     `json:"oldest_queued_seconds,omitempty"`
	AttentionReasons    []string   `json:"attention_reasons,omitempty"`
}

type ExportArtifact struct {
	ID              int64      `json:"id"`
	TaskID          int64      `json:"task_id"`
	UserID          int64      `json:"user_id"`
	Format          string     `json:"format"`
	StorageProvider string     `json:"storage_provider"`
	StorageKey      string     `json:"storage_key"`
	DownloadURL     string     `json:"download_url"`
	FileName        string     `json:"file_name"`
	FileSize        int64      `json:"file_size"`
	Checksum        string     `json:"checksum"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ExportTaskLog struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Event     string    `json:"event"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	MetaJSON  string    `json:"meta_json"`
	CreatedAt time.Time `json:"created_at"`
}

type ArticleEngagementResult struct {
	ReadNum    *int64 `json:"read_num,omitempty"`
	OldLikeNum *int64 `json:"old_like_num,omitempty"`
	ShareNum   *int64 `json:"share_num,omitempty"`
	LikeNum    *int64 `json:"like_num,omitempty"`
	CommentNum *int64 `json:"comment_num,omitempty"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

type ExportRepository interface {
	GetActiveSession(ctx context.Context, userID int64) (*Session, error)
	CreateSession(ctx context.Context, session *Session) error
	UpdateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, userID int64, sessionID int64) (*Session, error)
	ExpireUserSessions(ctx context.Context, userID int64) error
	ExpireLoginAttemptSessions(ctx context.Context, userID int64) error
	SearchAccounts(ctx context.Context, userID int64, query string, limit int) ([]Account, error)
	GetAccount(ctx context.Context, userID int64, fakeID string) (*Account, error)
	UpsertAccount(ctx context.Context, account *Account) error
	MarkAccountSynced(ctx context.Context, userID int64, fakeID string) (*Account, error)
	UpsertArticle(ctx context.Context, article *Article) error
	UpdateArticleEnrichment(ctx context.Context, article *Article) error
	ListArticles(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Article, *pagination.PaginationResult, error)
	GetArticleByID(ctx context.Context, articleID int64) (*Article, error)
	ListArticlesByIDs(ctx context.Context, userID int64, articleIDs []int64) ([]Article, error)
	CreateTask(ctx context.Context, task *ExportTask) error
	ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ExportTask, *pagination.PaginationResult, error)
	GetWorkerStatus(ctx context.Context, userID int64) (*ExportWorkerStatus, error)
	GetTask(ctx context.Context, userID int64, taskID int64) (*ExportTask, error)
	CancelTask(ctx context.Context, userID int64, taskID int64) (*ExportTask, error)
	RetryTask(ctx context.Context, userID int64, taskID int64) (*ExportTask, error)
	AddTaskLog(ctx context.Context, taskID int64, leaseToken string, log ExportTaskLog) (*ExportTaskLog, error)
	ListTaskLogs(ctx context.Context, userID int64, taskID int64) ([]ExportTaskLog, error)
	ClaimNextTask(ctx context.Context, leaseSeconds int64) (task *ExportTask, articles []Article, leaseToken string, err error)
	CompleteTask(ctx context.Context, taskID int64, leaseToken string, artifacts []ExportArtifact, resultManifestJSON string, failedArticleCount int, actualCost float64) (*ExportTask, error)
	FailTask(ctx context.Context, taskID int64, leaseToken string, message string) (*ExportTask, error)
	ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]ExportArtifact, error)
	GetArtifact(ctx context.Context, userID int64, artifactID int64) (*ExportArtifact, error)
}

func GenerateWorkerLeaseToken() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func GenerateWorkerRunID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}
