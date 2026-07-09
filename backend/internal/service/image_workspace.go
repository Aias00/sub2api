package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/Aias00/cloudbase/internal/billing"
	contentimage "github.com/Aias00/cloudbase/internal/image"
	infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
)

const (
	ImageWorkspaceTaskStatusQueued    = contentimage.ImageWorkspaceTaskStatusQueued
	ImageWorkspaceTaskStatusRunning   = contentimage.ImageWorkspaceTaskStatusRunning
	ImageWorkspaceTaskStatusSucceeded = contentimage.ImageWorkspaceTaskStatusSucceeded
	ImageWorkspaceTaskStatusFailed    = contentimage.ImageWorkspaceTaskStatusFailed
	ImageWorkspaceTaskStatusCancelled = contentimage.ImageWorkspaceTaskStatusCancelled
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
	ErrImageWorkspaceInvalidInput  = contentimage.ErrImageWorkspaceInvalidInput
	ErrImageWorkspaceTaskNotFound  = contentimage.ErrImageWorkspaceTaskNotFound
	ErrImageWorkspaceModelNotFound = contentimage.ErrImageWorkspaceModelNotFound
	ErrImageWorkspaceInvalidCost   = contentimage.ErrImageWorkspaceInvalidCost
	ErrImageWorkspaceNonRetryable  = contentimage.ErrImageWorkspaceNonRetryable
	// ErrImageWorkspacePromptBlocked marks a task creation rejected by the
	// server-side prompt safety filter. The concrete message returned to the
	// caller carries the configured warning text; errors.Is matches on
	// code+reason so callers can branch without string matching.
	ErrImageWorkspacePromptBlocked = infraerrors.BadRequest("IMAGE_WORKSPACE_PROMPT_BLOCKED", "image workspace prompt blocked by safety filter")
)

type ImageWorkspaceTask = contentimage.ImageWorkspaceTask
type ImageWorkspaceArtifact = contentimage.ImageWorkspaceArtifact
type ImageWorkspaceTemplate = contentimage.ImageWorkspaceTemplate
type ImageWorkspaceUsageRecord = contentimage.ImageWorkspaceUsageRecord
type ImageWorkspaceModelOption = contentimage.ImageWorkspaceModelOption
type ImageWorkspaceWorkerStatus = contentimage.ImageWorkspaceWorkerStatus
type ImageWorkspaceModelConfig = contentimage.ImageWorkspaceModelConfig
type CreateImageWorkspaceTaskInput = contentimage.CreateImageWorkspaceTaskInput
type UpsertImageWorkspaceTemplateInput = contentimage.UpsertImageWorkspaceTemplateInput
type CompleteImageWorkspaceTaskInput = contentimage.CompleteImageWorkspaceTaskInput
type FailImageWorkspaceTaskInput = contentimage.FailImageWorkspaceTaskInput
type ImageWorkspaceArtifactInput = contentimage.ImageWorkspaceArtifactInput
type ImageWorkspaceTaskFilters = contentimage.ImageWorkspaceTaskFilters
type ImageWorkspaceRepository = contentimage.ImageWorkspaceRepository

// UserBalanceReader reads a user's current balance. It is backed in production
// by *BillingCacheService (cached balance, DB fallback). Used only for a
// fast-fail pre-check in CreateTask; the authoritative balance gate remains
// the repo-layer reservation (pg_advisory_xact_lock + FOR UPDATE).
type UserBalanceReader interface {
	GetUserBalance(ctx context.Context, userID int64) (float64, error)
}

type ImageWorkspaceService struct {
	repo           ImageWorkspaceRepository
	settingRepo    SettingRepository
	settingService *SettingService
	balanceReader  UserBalanceReader
}

// NewImageWorkspaceService is the base constructor used by tests. It leaves the
// optional runtime deps (settingService / balanceReader) nil; the prompt-safety
// filter and balance pre-check degrade to no-ops in that case. Production wiring
// goes through ProvideImageWorkspaceService, which sets both.
func NewImageWorkspaceService(repo ImageWorkspaceRepository, settingRepo SettingRepository) *ImageWorkspaceService {
	return &ImageWorkspaceService{repo: repo, settingRepo: settingRepo}
}

// ProvideImageWorkspaceService is the wire provider that injects the optional
// runtime deps on top of the base constructor: the prompt-safety SettingService
// and the cached balance reader used for the CreateTask fast-fail pre-check.
func ProvideImageWorkspaceService(repo ImageWorkspaceRepository, settingRepo SettingRepository, settingService *SettingService, balanceReader UserBalanceReader) *ImageWorkspaceService {
	svc := NewImageWorkspaceService(repo, settingRepo)
	svc.settingService = settingService
	svc.balanceReader = balanceReader
	return svc
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
	// Server-side prompt safety filter. Mirrors the frontend rules in
	// ImageGeneratorView.vue so that direct API calls (which bypass the UI)
	// cannot skip content moderation. Returns nil when the filter is off or
	// the prompt is clean.
	if err := s.evaluatePromptSafety(ctx, input); err != nil {
		return nil, err
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
	// Fast-fail balance pre-check. This is purely an optimization/UX gate so
	// that obviously-insufficient users don't spin up a transaction + advisory
	// lock just to be rejected. The authoritative invariant is still enforced
	// in the repo: pg_advisory_xact_lock serializes concurrent creates per user
	// and reserveUserBalanceWithComponents re-checks balance >= amount under
	// FOR UPDATE. A stale cache read here therefore cannot drive the balance
	// negative — at worst it lets a request through that the repo then rejects
	// with billing.ErrInsufficientBalance, or rejects one that would have
	// succeeded (user retries).
	if task.CostEstimate > 0 && s.balanceReader != nil {
		if balance, rerr := s.balanceReader.GetUserBalance(ctx, userID); rerr == nil && balance < task.CostEstimate {
			return nil, billing.ErrInsufficientBalance
		}
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// evaluatePromptSafety applies the admin-configured image prompt filter
// (image_prompt_filter_config) server-side. The matching semantics are kept
// in lockstep with the frontend implementation in ImageGeneratorView.vue:
//
//   - text = prompt + "\n" + negative_prompt + "\n" + style, lower-cased
//   - a keyword matches via ASCII word boundary (\bkw\b), case-insensitive
//   - block when (explicit && youth_context) OR (explicit && "young")
//
// Two gates must both be on for the filter to apply:
//   - the runtime toggle image_workspace_prompt_safety_enabled (the same key the
//     worker runtime config下发为 prompt_safety_enabled; admin "Prompt 安全检查"
//     switch). When the admin disables it, the worker stops checking — this
//     server-side gate must skip too, otherwise direct API calls would still be
//     filtered while the worker path is not.
//   - filter.Enabled inside image_prompt_filter_config.
//
// On block it returns an ErrImageWorkspacePromptBlocked carrying the configured
// warning message; nil when either gate is off or the prompt is clean.
func (s *ImageWorkspaceService) evaluatePromptSafety(ctx context.Context, input CreateImageWorkspaceTaskInput) error {
	if s == nil || s.settingService == nil {
		return nil
	}
	if !s.promptSafetyEnabled(ctx) {
		return nil
	}
	filter, err := s.settingService.GetImagePromptFilterConfig(ctx)
	if err != nil || filter == nil || !filter.Enabled {
		return nil
	}
	text := strings.ToLower(input.Prompt + "\n" + input.NegativePrompt + "\n" + input.Style)
	// Normalize keywords: trim and drop empty so malformed config entries
	// (leading/trailing spaces, blanks) cannot create dead or over-broad
	// patterns. This diverges from the literal frontend wordMatch (which does
	// no trimming), but only in the safe direction for a moderation gate — a
	// spacey keyword that the frontend would silently never match becomes a
	// clean match server-side.
	explicitKeywords := normalizeImageKeywords(filter.ExplicitKeywords)
	youthContextKeywords := normalizeImageKeywords(filter.YouthContextKeywords)

	hasExplicit := anyImageKeywordMatch(text, explicitKeywords)
	if !hasExplicit {
		return nil
	}
	hasYouthContext := anyImageKeywordMatch(text, youthContextKeywords)
	if hasYouthContext {
		return imagePromptBlockedError(filter.WarningMessage)
	}
	if imageWordMatch(text, "young") {
		return imagePromptBlockedError(filter.YouthWarningMessage)
	}
	return nil
}

// imagePromptBlockedError builds a per-request error sharing the
// IMAGE_WORKSPACE_PROMPT_BLOCKED reason so errors.Is matches the sentinel
// while the message reflects the configured warning text. An empty configured
// message degrades to a generic explanation so the rejection is still useful
// to the caller.
func imagePromptBlockedError(message string) error {
	msg := strings.TrimSpace(message)
	if msg == "" {
		msg = "prompt blocked by safety filter"
	}
	return infraerrors.BadRequest("IMAGE_WORKSPACE_PROMPT_BLOCKED", msg)
}

// imageWordMatch reports whether term appears in text on an ASCII word
// boundary, case-insensitively. Both text and term are lower-cased, mirroring
// the JS `text.toLowerCase()` + `new RegExp("\\b"+escaped+"\\b")` behavior used
// by the frontend filter. term is treated as a literal (regex metacharacters
// are escaped).
func imageWordMatch(text, term string) bool {
	term = strings.TrimSpace(term)
	if term == "" {
		return false
	}
	pattern := "\\b" + regexp.QuoteMeta(strings.ToLower(term)) + "\\b"
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(strings.ToLower(text))
}

func anyImageKeywordMatch(text string, keywords []string) bool {
	for _, kw := range keywords {
		if imageWordMatch(text, kw) {
			return true
		}
	}
	return false
}

// normalizeImageKeywords trims each keyword and drops empty/whitespace-only
// entries. Keeping malformed entries (e.g. "  ", "") would either produce a
// dead pattern or, for imageWordMatch's empty guard, be skipped anyway — so
// normalizing here keeps the keyword list semantically clean and stable
// regardless of how the admin saved the config.
func normalizeImageKeywords(keywords []string) []string {
	out := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if trimmed := strings.TrimSpace(kw); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// promptSafetyEnabled reads the image_workspace_prompt_safety_enabled runtime
// toggle (the admin "Prompt 安全检查" switch, also下发 to the worker as
// prompt_safety_enabled). Defaults to true when the setting is unset/unreadable,
// matching the worker runtime config and setting service defaults, so a config
// or DB hiccup does not silently disable the server-side filter.
func (s *ImageWorkspaceService) promptSafetyEnabled(ctx context.Context) bool {
	if s == nil || s.settingRepo == nil {
		return true
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyImageWorkspacePromptSafetyEnabled)
	if err != nil {
		return true
	}
	return parseBoolSettingWithDefault(raw, true)
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
