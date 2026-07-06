package image

import (
	"context"
	"time"

	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
)

const (
	ImageWorkspaceTaskStatusQueued    = "queued"
	ImageWorkspaceTaskStatusRunning   = "running"
	ImageWorkspaceTaskStatusSucceeded = "succeeded"
	ImageWorkspaceTaskStatusFailed    = "failed"
	ImageWorkspaceTaskStatusCancelled = "cancelled"
)

var (
	ErrImageWorkspaceInvalidInput  = infraerrors.BadRequest("IMAGE_WORKSPACE_INVALID_INPUT", "image workspace input is invalid")
	ErrImageWorkspaceTaskNotFound  = infraerrors.NotFound("IMAGE_WORKSPACE_TASK_NOT_FOUND", "image workspace task not found")
	ErrImageWorkspaceModelNotFound = infraerrors.BadRequest("IMAGE_WORKSPACE_MODEL_NOT_FOUND", "image workspace model not found or disabled")
	ErrImageWorkspaceInvalidCost   = infraerrors.BadRequest("IMAGE_WORKSPACE_INVALID_COST", "reported cost is suspiciously lower than the original estimate")
	ErrImageWorkspaceNonRetryable  = infraerrors.BadRequest("IMAGE_WORKSPACE_NON_RETRYABLE", "image workspace task failed with a non-retryable upstream policy violation")
)

type ImageWorkspaceTask struct {
	ID                  int64                    `json:"id"`
	UserID              int64                    `json:"user_id"`
	Status              string                   `json:"status"`
	Prompt              string                   `json:"prompt"`
	NegativePrompt      string                   `json:"negative_prompt"`
	Model               string                   `json:"model"`
	Provider            string                   `json:"provider"`
	Size                string                   `json:"size"`
	Quality             string                   `json:"quality"`
	Style               string                   `json:"style"`
	Seed                *int64                   `json:"seed,omitempty"`
	BatchSize           int                      `json:"batch_size"`
	TemplateID          *int64                   `json:"template_id,omitempty"`
	WorkerLeaseUntil    *time.Time               `json:"worker_lease_until,omitempty"`
	CostEstimate        float64                  `json:"cost_estimate"`
	BalanceSnapshot     float64                  `json:"balance_snapshot"`
	ReservedPaidBalance float64                  `json:"-"`
	ReservedGiftBalance float64                  `json:"-"`
	ErrorMessage        string                   `json:"error_message"`
	ResultJSON          string                   `json:"result_json"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
	Artifacts           []ImageWorkspaceArtifact `json:"artifacts,omitempty"`
}

type ImageWorkspaceArtifact struct {
	ID              int64     `json:"id"`
	TaskID          int64     `json:"task_id"`
	UserID          int64     `json:"user_id"`
	StorageProvider string    `json:"storage_provider"`
	StorageKey      string    `json:"storage_key"`
	ImageURL        string    `json:"image_url"`
	Prompt          string    `json:"prompt"`
	MimeType        string    `json:"mime_type"`
	Width           int       `json:"width"`
	Height          int       `json:"height"`
	FileSize        int64     `json:"file_size"`
	Checksum        string    `json:"checksum"`
	MetadataJSON    string    `json:"metadata_json"`
	CreatedAt       time.Time `json:"created_at"`
}

type ImageWorkspaceTemplate struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Prompt         string    `json:"prompt"`
	NegativePrompt string    `json:"negative_prompt"`
	Model          string    `json:"model"`
	Size           string    `json:"size"`
	Quality        string    `json:"quality"`
	Style          string    `json:"style"`
	IsDefault      bool      `json:"is_default"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ImageWorkspaceUsageRecord struct {
	ID              int64     `json:"id"`
	TaskID          int64     `json:"task_id"`
	UserID          int64     `json:"user_id"`
	Provider        string    `json:"provider"`
	Model           string    `json:"model"`
	Size            string    `json:"size"`
	Quality         string    `json:"quality"`
	ImageCount      int       `json:"image_count"`
	ReservedCost    float64   `json:"reserved_cost"`
	ActualCost      float64   `json:"actual_cost"`
	BalanceSnapshot float64   `json:"balance_snapshot"`
	BillingStatus   string    `json:"billing_status"`
	MetadataJSON    string    `json:"metadata_json"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ImageWorkspaceModelOption struct {
	ID             string   `json:"id"`
	Label          string   `json:"label"`
	Provider       string   `json:"provider"`
	DefaultSize    string   `json:"default_size"`
	DefaultQuality string   `json:"default_quality"`
	Sizes          []string `json:"sizes"`
	Qualities      []string `json:"qualities"`
	CostPerImage   float64  `json:"cost_per_image"`
	CostHint       string   `json:"cost_hint"`
	Enabled        bool     `json:"enabled"`
}

type ImageWorkspaceWorkerStatus struct {
	Health              string     `json:"health"`
	Message             string     `json:"message"`
	TotalCount          int64      `json:"total_count"`
	QueuedCount         int64      `json:"queued_count"`
	RunningCount        int64      `json:"running_count"`
	StaleRunningCount   int64      `json:"stale_running_count"`
	FailedCount         int64      `json:"failed_count"`
	RecentFailedCount   int64      `json:"recent_failed_count"`
	SucceededCount      int64      `json:"succeeded_count"`
	CancelledCount      int64      `json:"cancelled_count"`
	ArtifactCount       int64      `json:"artifact_count"`
	LastTaskUpdatedAt   *time.Time `json:"last_task_updated_at,omitempty"`
	LastTaskAgeSeconds  *int64     `json:"last_task_age_seconds,omitempty"`
	LastFailedAt        *time.Time `json:"last_failed_at,omitempty"`
	LastFailureMessage  string     `json:"last_failure_message,omitempty"`
	OldestQueuedAt      *time.Time `json:"oldest_queued_at,omitempty"`
	OldestQueuedSeconds *int64     `json:"oldest_queued_seconds,omitempty"`
	AttentionReasons    []string   `json:"attention_reasons,omitempty"`
}

type ImageWorkspaceModelConfig struct {
	Models []ImageWorkspaceModelOption `json:"models"`
}

type CreateImageWorkspaceTaskInput struct {
	Prompt         string
	NegativePrompt string
	Model          string
	Provider       string
	Size           string
	Quality        string
	Style          string
	Seed           *int64
	BatchSize      int
	TemplateID     *int64
}

type UpsertImageWorkspaceTemplateInput struct {
	ID             int64
	Title          string
	Description    string
	Prompt         string
	NegativePrompt string
	Model          string
	Size           string
	Quality        string
	Style          string
	IsDefault      bool
}

type CompleteImageWorkspaceTaskInput struct {
	Artifacts  []ImageWorkspaceArtifactInput
	ResultJSON string
	Cost       float64
}

type FailImageWorkspaceTaskInput struct {
	Message    string
	ResultJSON string
}

type ImageWorkspaceArtifactInput struct {
	StorageProvider string `json:"storage_provider"`
	StorageKey      string `json:"storage_key"`
	ImageURL        string `json:"image_url"`
	Prompt          string `json:"prompt"`
	MimeType        string `json:"mime_type"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	FileSize        int64  `json:"file_size"`
	Checksum        string `json:"checksum"`
	MetadataJSON    string `json:"metadata_json"`
}

type ImageWorkspaceTaskFilters struct {
	Status string
}

type ImageWorkspaceRepository interface {
	CreateTask(ctx context.Context, task *ImageWorkspaceTask) error
	ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams, filters ImageWorkspaceTaskFilters) ([]ImageWorkspaceTask, *pagination.PaginationResult, error)
	GetTask(ctx context.Context, userID int64, taskID int64) (*ImageWorkspaceTask, error)
	ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]ImageWorkspaceArtifact, error)
	GetArtifact(ctx context.Context, userID int64, artifactID int64) (*ImageWorkspaceArtifact, error)
	ListTemplates(ctx context.Context, userID int64) ([]ImageWorkspaceTemplate, error)
	UpsertTemplate(ctx context.Context, template *ImageWorkspaceTemplate) error
	DeleteTemplate(ctx context.Context, userID int64, templateID int64) error
	ListUsageRecords(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ImageWorkspaceUsageRecord, *pagination.PaginationResult, error)
	ClaimNextTask(ctx context.Context, leaseSeconds int64) (*ImageWorkspaceTask, error)
	CompleteTask(ctx context.Context, taskID int64, artifacts []ImageWorkspaceArtifact, resultJSON string, cost float64) (*ImageWorkspaceTask, error)
	FailTask(ctx context.Context, taskID int64, message string, resultJSON string) (*ImageWorkspaceTask, error)
	CancelTask(ctx context.Context, taskID int64, userID int64) (*ImageWorkspaceTask, error)
	GetWorkerStatus(ctx context.Context) (*ImageWorkspaceWorkerStatus, error)
}
