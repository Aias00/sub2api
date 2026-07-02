package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/cloudbase/internal/pkg/errors"
	"github.com/Wei-Shaw/cloudbase/internal/pkg/pagination"
)

const (
	ImageWorkspaceTaskStatusQueued    = "queued"
	ImageWorkspaceTaskStatusRunning   = "running"
	ImageWorkspaceTaskStatusSucceeded = "succeeded"
	ImageWorkspaceTaskStatusFailed    = "failed"
	ImageWorkspaceTaskStatusCancelled = "cancelled"
)

const defaultImageWorkspaceModelConfig = `{
  "models": [
    {
      "id": "gpt-image-2",
      "label": "GPT Image 2",
      "provider": "openai",
      "default_size": "1024x1024",
      "default_quality": "standard",
      "sizes": ["1024x1024", "1024x1536", "1536x1024"],
      "qualities": ["standard", "hd", "high"],
      "cost_per_image": 0.04,
      "cost_hint": "0.04 / 张",
      "enabled": true
    },
    {
      "id": "gpt-image-1",
      "label": "GPT Image 1",
      "provider": "openai",
      "default_size": "1024x1024",
      "default_quality": "standard",
      "sizes": ["1024x1024", "1024x1536", "1536x1024"],
      "qualities": ["standard", "hd"],
      "cost_per_image": 0.04,
      "cost_hint": "0.04 / 张",
      "enabled": true
    },
    {
      "id": "gemini-3.1-flash-image",
      "label": "Gemini 3.1 Flash Image",
      "provider": "gemini",
      "default_size": "1024x1024",
      "default_quality": "standard",
      "sizes": ["1024x1024"],
      "qualities": ["standard"],
      "cost_per_image": 0.04,
      "cost_hint": "0.04 / 张",
      "enabled": true
    }
  ]
}`

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

type ImageWorkspaceService struct {
	repo        ImageWorkspaceRepository
	settingRepo SettingRepository
}

func NewImageWorkspaceService(repo ImageWorkspaceRepository, settingRepo SettingRepository) *ImageWorkspaceService {
	return &ImageWorkspaceService{repo: repo, settingRepo: settingRepo}
}

func (s *ImageWorkspaceService) Health() error {
	if s == nil || s.repo == nil {
		return ErrImageWorkspaceInvalidInput
	}
	return nil
}

func (s *ImageWorkspaceService) GetWorkerRuntimeConfig(ctx context.Context) (ImageWorkspaceWorkerRuntimeConfig, error) {
	if s == nil || s.settingRepo == nil {
		return defaultImageWorkspaceWorkerRuntimeConfig(), nil
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyImageWorkspaceUpstreamURL,
		SettingKeyImageWorkspaceGenerationTimeoutMS,
		SettingKeyImageWorkspaceCompletionCost,
		SettingKeyImageWorkspaceCompletionCostMapJSON,
		SettingKeyImageWorkspacePromptSafetyEnabled,
		SettingKeyImageWorkspaceAssumeWorkerReady,
		SettingKeyImageWorkspaceObjectStorageEnabled,
		SettingKeyImageWorkspaceObjectStorageProvider,
		SettingKeyImageWorkspaceObjectStorageBucket,
		SettingKeyImageWorkspaceObjectStorageRegion,
		SettingKeyImageWorkspaceObjectStoragePrefix,
		SettingKeyImageWorkspaceObjectStoragePublicBaseURL,
		SettingKeyMediaCDNBaseURL,
	})
	if err != nil {
		return ImageWorkspaceWorkerRuntimeConfig{}, err
	}
	return imageWorkspaceWorkerRuntimeConfigFromSettings(values), nil
}

func (s *ImageWorkspaceService) CreateTask(ctx context.Context, userID int64, input CreateImageWorkspaceTaskInput) (*ImageWorkspaceTask, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(input.Prompt)
	if userID <= 0 || prompt == "" {
		return nil, ErrImageWorkspaceInvalidInput
	}
	batchSize := input.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}
	if batchSize > 4 {
		batchSize = 4
	}
	models, _ := s.ListModels(ctx)
	var matched *ImageWorkspaceModelOption
	for i := range models {
		if models[i].ID == strings.TrimSpace(input.Model) && models[i].Enabled {
			matched = &models[i]
			break
		}
	}
	if matched == nil {
		return nil, ErrImageWorkspaceModelNotFound
	}
	task := &ImageWorkspaceTask{
		UserID:         userID,
		Status:         ImageWorkspaceTaskStatusQueued,
		Prompt:         prompt,
		NegativePrompt: strings.TrimSpace(input.NegativePrompt),
		Model:          matched.ID,
		Provider:       defaultImageWorkspaceString(input.Provider, matched.Provider),
		Size:           defaultImageWorkspaceString(input.Size, matched.DefaultSize),
		Quality:        defaultImageWorkspaceString(input.Quality, matched.DefaultQuality),
		Style:          strings.TrimSpace(input.Style),
		Seed:           input.Seed,
		BatchSize:      batchSize,
		TemplateID:     input.TemplateID,
		CostEstimate:   maxImageWorkspaceFloat(matched.CostPerImage, 0) * float64(batchSize),
		ResultJSON:     "{}",
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *ImageWorkspaceService) EstimateTaskCost(ctx context.Context, modelID string, batchSize int) float64 {
	if batchSize <= 0 {
		batchSize = 1
	}
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		modelID = "gpt-image-2"
	}
	models, _ := s.ListModels(ctx)
	for _, model := range models {
		if model.ID == modelID {
			return maxImageWorkspaceFloat(model.CostPerImage, 0) * float64(batchSize)
		}
	}
	return 0
}

func (s *ImageWorkspaceService) ListModels(ctx context.Context) ([]ImageWorkspaceModelOption, error) {
	models := defaultImageWorkspaceModels()
	if s != nil && s.settingRepo != nil {
		if raw, err := s.settingRepo.GetValue(ctx, SettingKeyImageWorkspaceModelConfig); err == nil && strings.TrimSpace(raw) != "" {
			models = parseImageWorkspaceModelConfig(raw)
		}
	}
	return normalizeImageWorkspaceModels(models), nil
}

func (s *ImageWorkspaceService) ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]ImageWorkspaceTask, *pagination.PaginationResult, error) {
	if err := s.Health(); err != nil {
		return nil, nil, err
	}
	if userID <= 0 {
		return nil, nil, ErrImageWorkspaceInvalidInput
	}
	params = normalizeImageWorkspacePagination(params)
	filters := ImageWorkspaceTaskFilters{Status: strings.TrimSpace(status)}
	return s.repo.ListTasks(ctx, userID, params, filters)
}

func (s *ImageWorkspaceService) GetTask(ctx context.Context, userID int64, taskID int64) (*ImageWorkspaceTask, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	if userID <= 0 || taskID <= 0 {
		return nil, ErrImageWorkspaceInvalidInput
	}
	task, err := s.repo.GetTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	artifacts, err := s.repo.ListArtifacts(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	task.Artifacts = artifacts
	return task, nil
}

func (s *ImageWorkspaceService) GetArtifact(ctx context.Context, userID int64, artifactID int64) (*ImageWorkspaceArtifact, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	if userID <= 0 || artifactID <= 0 {
		return nil, ErrImageWorkspaceInvalidInput
	}
	return s.repo.GetArtifact(ctx, userID, artifactID)
}

func (s *ImageWorkspaceService) RetryTask(ctx context.Context, userID int64, taskID int64) (*ImageWorkspaceTask, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	if userID <= 0 || taskID <= 0 {
		return nil, ErrImageWorkspaceInvalidInput
	}
	task, err := s.repo.GetTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	if task.Status != ImageWorkspaceTaskStatusFailed && task.Status != ImageWorkspaceTaskStatusCancelled {
		return nil, ErrImageWorkspaceInvalidInput
	}
	if ImageWorkspaceFailureIsNonRetryable(task.ErrorMessage, task.ResultJSON) {
		return nil, ErrImageWorkspaceNonRetryable
	}
	return s.CreateTask(ctx, userID, CreateImageWorkspaceTaskInput{
		Prompt:         task.Prompt,
		NegativePrompt: task.NegativePrompt,
		Model:          task.Model,
		Provider:       task.Provider,
		Size:           task.Size,
		Quality:        task.Quality,
		Style:          task.Style,
		Seed:           task.Seed,
		BatchSize:      task.BatchSize,
		TemplateID:     task.TemplateID,
	})
}

func (s *ImageWorkspaceService) CancelTask(ctx context.Context, userID int64, taskID int64) (*ImageWorkspaceTask, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	if userID <= 0 || taskID <= 0 {
		return nil, ErrImageWorkspaceInvalidInput
	}
	task, err := s.repo.GetTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	// Allow cancelling queued tasks or running tasks with expired lease
	if task.Status != ImageWorkspaceTaskStatusQueued {
		if task.Status != ImageWorkspaceTaskStatusRunning {
			return nil, ErrImageWorkspaceInvalidInput
		}
		// Running task can be cancelled only if worker lease has expired
		if task.WorkerLeaseUntil != nil && task.WorkerLeaseUntil.After(time.Now()) {
			return nil, ErrImageWorkspaceInvalidInput
		}
	}
	return s.repo.CancelTask(ctx, taskID, userID)
}

func (s *ImageWorkspaceService) ListTemplates(ctx context.Context, userID int64) ([]ImageWorkspaceTemplate, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, ErrImageWorkspaceInvalidInput
	}
	return s.repo.ListTemplates(ctx, userID)
}

func (s *ImageWorkspaceService) UpsertTemplate(ctx context.Context, userID int64, input UpsertImageWorkspaceTemplateInput) (*ImageWorkspaceTemplate, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	title := strings.TrimSpace(input.Title)
	prompt := strings.TrimSpace(input.Prompt)
	if userID <= 0 || title == "" || prompt == "" {
		return nil, ErrImageWorkspaceInvalidInput
	}
	template := &ImageWorkspaceTemplate{
		ID:             input.ID,
		UserID:         userID,
		Title:          title,
		Description:    strings.TrimSpace(input.Description),
		Prompt:         prompt,
		NegativePrompt: strings.TrimSpace(input.NegativePrompt),
		Model:          defaultImageWorkspaceString(input.Model, "gpt-image-2"),
		Size:           defaultImageWorkspaceString(input.Size, "1024x1024"),
		Quality:        defaultImageWorkspaceString(input.Quality, "standard"),
		Style:          strings.TrimSpace(input.Style),
		IsDefault:      input.IsDefault,
	}
	if err := s.repo.UpsertTemplate(ctx, template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *ImageWorkspaceService) DeleteTemplate(ctx context.Context, userID int64, templateID int64) error {
	if err := s.Health(); err != nil {
		return err
	}
	if userID <= 0 || templateID <= 0 {
		return ErrImageWorkspaceInvalidInput
	}
	return s.repo.DeleteTemplate(ctx, userID, templateID)
}

func (s *ImageWorkspaceService) ListUsageRecords(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ImageWorkspaceUsageRecord, *pagination.PaginationResult, error) {
	if err := s.Health(); err != nil {
		return nil, nil, err
	}
	if userID <= 0 {
		return nil, nil, ErrImageWorkspaceInvalidInput
	}
	params = normalizeImageWorkspacePagination(params)
	return s.repo.ListUsageRecords(ctx, userID, params)
}

func (s *ImageWorkspaceService) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*ImageWorkspaceTask, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	if leaseSeconds <= 0 {
		leaseSeconds = 300
	}
	if leaseSeconds > 3600 {
		leaseSeconds = 3600
	}
	return s.repo.ClaimNextTask(ctx, leaseSeconds)
}

func (s *ImageWorkspaceService) CompleteTask(ctx context.Context, taskID int64, input CompleteImageWorkspaceTaskInput) (*ImageWorkspaceTask, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	if taskID <= 0 {
		return nil, ErrImageWorkspaceInvalidInput
	}
	resultJSON := strings.TrimSpace(input.ResultJSON)
	if resultJSON == "" {
		resultJSON = "{}"
	}
	if !json.Valid([]byte(resultJSON)) {
		return nil, ErrImageWorkspaceInvalidInput
	}
	artifacts := make([]ImageWorkspaceArtifact, 0, len(input.Artifacts))
	for _, item := range input.Artifacts {
		artifact := ImageWorkspaceArtifact{
			StorageProvider: defaultImageWorkspaceString(item.StorageProvider, "external"),
			StorageKey:      strings.TrimSpace(item.StorageKey),
			ImageURL:        strings.TrimSpace(item.ImageURL),
			Prompt:          strings.TrimSpace(item.Prompt),
			MimeType:        defaultImageWorkspaceString(item.MimeType, "image/png"),
			Width:           maxImageWorkspaceInt(item.Width, 0),
			Height:          maxImageWorkspaceInt(item.Height, 0),
			FileSize:        maxImageWorkspaceInt64(item.FileSize, 0),
			Checksum:        strings.TrimSpace(item.Checksum),
			MetadataJSON:    defaultImageWorkspaceJSON(item.MetadataJSON),
		}
		if artifact.ImageURL == "" && artifact.StorageKey == "" {
			continue
		}
		artifacts = append(artifacts, artifact)
	}
	if len(artifacts) == 0 {
		return nil, ErrImageWorkspaceInvalidInput
	}
	return s.repo.CompleteTask(ctx, taskID, artifacts, resultJSON, maxImageWorkspaceFloat(input.Cost, 0))
}

func (s *ImageWorkspaceService) FailTask(ctx context.Context, taskID int64, input FailImageWorkspaceTaskInput) (*ImageWorkspaceTask, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	if taskID <= 0 {
		return nil, ErrImageWorkspaceInvalidInput
	}
	message := strings.TrimSpace(input.Message)
	if message == "" {
		message = "image generation failed"
	}
	return s.repo.FailTask(ctx, taskID, message, defaultImageWorkspaceJSON(input.ResultJSON))
}

func (s *ImageWorkspaceService) GetWorkerStatus(ctx context.Context) (*ImageWorkspaceWorkerStatus, error) {
	if err := s.Health(); err != nil {
		return nil, err
	}
	status, err := s.repo.GetWorkerStatus(ctx)
	if err != nil {
		return nil, err
	}
	if status == nil {
		status = &ImageWorkspaceWorkerStatus{}
	}
	now := time.Now().UTC()
	if status.LastTaskUpdatedAt != nil {
		ageSeconds := int64(now.Sub(status.LastTaskUpdatedAt.UTC()).Seconds())
		if ageSeconds < 0 {
			ageSeconds = 0
		}
		status.LastTaskAgeSeconds = &ageSeconds
	}
	if status.OldestQueuedAt != nil {
		queuedSeconds := int64(now.Sub(status.OldestQueuedAt.UTC()).Seconds())
		if queuedSeconds < 0 {
			queuedSeconds = 0
		}
		status.OldestQueuedSeconds = &queuedSeconds
	}
	switch {
	case status.StaleRunningCount > 0:
		status.Health = "attention"
		status.Message = "Some running image tasks have expired leases and will be reclaimed by the worker."
		status.AttentionReasons = append(status.AttentionReasons, "stale_running_tasks")
	case status.RecentFailedCount > 0 && imageWorkspaceFailureRequiresRuntimeAttention(status.LastFailureMessage):
		status.Health = "attention"
		status.Message = "Recent image tasks failed because the worker runtime or upstream image provider needs attention."
		status.AttentionReasons = append(status.AttentionReasons, "recent_runtime_failure")
	case status.QueuedCount > 0 && status.RunningCount == 0:
		status.Health = "waiting"
		status.Message = "Queued image tasks are waiting for a worker to pick them up."
		if status.OldestQueuedSeconds != nil && *status.OldestQueuedSeconds >= 300 {
			status.AttentionReasons = append(status.AttentionReasons, "queued_tasks_waiting_over_5m")
		}
	case status.RunningCount > 0:
		status.Health = "active"
		status.Message = "Worker has running image tasks."
	default:
		status.Health = "idle"
		status.Message = "No queued or running image tasks."
	}
	return status, nil
}

func imageWorkspaceFailureRequiresRuntimeAttention(message string) bool {
	if ImageWorkspaceFailureIsNonRetryable(message, "") {
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return false
	}
	for _, marker := range []string{
		"upstream",
		"上游",
		"api_key",
		"api key",
		"鉴权",
		"unauthorized",
		"forbidden",
		"401",
		"403",
		"404",
		"endpoint",
		"images/generations",
		"image provider",
		"provider",
		"timeout",
		"econnrefused",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func ImageWorkspaceFailureIsNonRetryable(message string, resultJSON string) bool {
	normalized := strings.ToLower(strings.TrimSpace(message + " " + resultJSON))
	if normalized == "" {
		return false
	}
	for _, marker := range []string{
		"safety_error",
		"policy_violation",
		"content_policy",
		"content policy",
		"moderation",
		"moderation_blocked",
		"blocked by policy",
		"policy blocked",
		"safety system",
		"safety filter",
		"safety violation",
		"violates policy",
		"violated policy",
		"flagged",
		"responsible ai",
		"risk policy",
		"违规",
		"违反",
		"安全策略",
		"内容安全",
		"风控",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func defaultImageWorkspaceString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func normalizeImageWorkspacePagination(params pagination.PaginationParams) pagination.PaginationParams {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	return params
}

func defaultImageWorkspaceJSON(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !json.Valid([]byte(value)) {
		return "{}"
	}
	return value
}

func maxImageWorkspaceInt(value int, min int) int {
	if value < min {
		return min
	}
	return value
}

func maxImageWorkspaceInt64(value int64, min int64) int64 {
	if value < min {
		return min
	}
	return value
}

func maxImageWorkspaceFloat(value float64, min float64) float64 {
	if value < min {
		return min
	}
	return value
}

func defaultImageWorkspaceModels() []ImageWorkspaceModelOption {
	return parseImageWorkspaceModelConfig(defaultImageWorkspaceModelConfig)
}

func parseImageWorkspaceModelConfig(raw string) []ImageWorkspaceModelOption {
	var config ImageWorkspaceModelConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return nil
	}
	return config.Models
}

func normalizeImageWorkspaceModels(models []ImageWorkspaceModelOption) []ImageWorkspaceModelOption {
	normalized := make([]ImageWorkspaceModelOption, 0, len(models))
	for _, model := range models {
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}
		provider := defaultImageWorkspaceString(model.Provider, "openai")
		sizes := normalizeImageWorkspaceStringList(model.Sizes)
		if len(sizes) == 0 {
			sizes = []string{"1024x1024"}
		}
		qualities := normalizeImageWorkspaceStringList(model.Qualities)
		if len(qualities) == 0 {
			qualities = []string{"standard"}
		}
		defaultSize := defaultImageWorkspaceString(model.DefaultSize, sizes[0])
		if !imageWorkspaceContainsString(sizes, defaultSize) {
			sizes = append([]string{defaultSize}, sizes...)
		}
		defaultQuality := defaultImageWorkspaceString(model.DefaultQuality, qualities[0])
		if !imageWorkspaceContainsString(qualities, defaultQuality) {
			qualities = append([]string{defaultQuality}, qualities...)
		}
		normalized = append(normalized, ImageWorkspaceModelOption{
			ID:             id,
			Label:          defaultImageWorkspaceString(model.Label, id),
			Provider:       provider,
			DefaultSize:    defaultSize,
			DefaultQuality: defaultQuality,
			Sizes:          sizes,
			Qualities:      qualities,
			CostPerImage:   maxImageWorkspaceFloat(model.CostPerImage, 0),
			CostHint:       strings.TrimSpace(model.CostHint),
			Enabled:        model.Enabled,
		})
	}
	if len(normalized) == 0 {
		return []ImageWorkspaceModelOption{
			{
				ID:             "gpt-image-2",
				Label:          "GPT Image 2",
				Provider:       "openai",
				DefaultSize:    "1024x1024",
				DefaultQuality: "standard",
				Sizes:          []string{"1024x1024", "1024x1536", "1536x1024"},
				Qualities:      []string{"standard", "hd", "high"},
				CostPerImage:   0.04,
				CostHint:       "0.04 / 张",
				Enabled:        true,
			},
		}
	}
	return normalized
}

func normalizeImageWorkspaceStringList(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	return result
}

func imageWorkspaceContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
