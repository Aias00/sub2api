package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Aias00/cloudbase/internal/billing"
	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type imageWorkspaceSettingRepo struct {
	values map[string]string
}

func (r imageWorkspaceSettingRepo) Get(ctx context.Context, key string) (*Setting, error) {
	if value, ok := r.values[key]; ok {
		return &Setting{Key: key, Value: value}, nil
	}
	return nil, ErrSettingNotFound
}

func (r imageWorkspaceSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := r.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (r imageWorkspaceSettingRepo) Set(ctx context.Context, key, value string) error {
	return nil
}

func (r imageWorkspaceSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (r imageWorkspaceSettingRepo) SetMultiple(ctx context.Context, settings map[string]string) error {
	return nil
}

func (r imageWorkspaceSettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	return r.values, nil
}

func (r imageWorkspaceSettingRepo) Delete(ctx context.Context, key string) error {
	return nil
}

func TestImageWorkspaceListModelsUsesSettingsConfig(t *testing.T) {
	svc := NewImageWorkspaceService(nil, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"custom-image","label":"Custom Image","provider":"custom","default_size":"512x512","default_quality":"fast","sizes":["512x512"],"qualities":["fast"],"cost_per_image":0.5,"cost_hint":"0.5 credit","enabled":true}]}`,
	}})

	models, err := svc.ListModels(context.Background())

	require.NoError(t, err)
	require.Len(t, models, 1)
	require.Equal(t, "custom-image", models[0].ID)
	require.Equal(t, "Custom Image", models[0].Label)
	require.Equal(t, "custom", models[0].Provider)
	require.Equal(t, "512x512", models[0].DefaultSize)
	require.Equal(t, "fast", models[0].DefaultQuality)
	require.Equal(t, 0.5, models[0].CostPerImage)
	require.Equal(t, "0.5 credit", models[0].CostHint)
	require.Equal(t, 1.5, svc.EstimateTaskCost(context.Background(), "custom-image", 3))
}

func TestImageWorkspaceListModelsFallsBackToDefaults(t *testing.T) {
	svc := NewImageWorkspaceService(nil, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{bad-json`,
	}})

	models, err := svc.ListModels(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, models)
	require.Equal(t, "gpt-image-2", models[0].ID)
}

type imageWorkspaceRepoFake struct {
	tasks     map[int64]ImageWorkspaceTask
	artifacts map[int64][]ImageWorkspaceArtifact
	templates map[int64]ImageWorkspaceTemplate
	nextID    int64
}

func newImageWorkspaceRepoFake() *imageWorkspaceRepoFake {
	return &imageWorkspaceRepoFake{
		tasks:     map[int64]ImageWorkspaceTask{},
		artifacts: map[int64][]ImageWorkspaceArtifact{},
		templates: map[int64]ImageWorkspaceTemplate{},
		nextID:    1,
	}
}

func (r *imageWorkspaceRepoFake) CreateTask(ctx context.Context, task *ImageWorkspaceTask) error {
	if task == nil {
		return nil
	}
	task.ID = r.nextID
	r.nextID++
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now
	r.tasks[task.ID] = *task
	return nil
}

func (r *imageWorkspaceRepoFake) ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams, filters ImageWorkspaceTaskFilters) ([]ImageWorkspaceTask, *pagination.PaginationResult, error) {
	items := make([]ImageWorkspaceTask, 0, len(r.tasks))
	for _, task := range r.tasks {
		if task.UserID != userID {
			continue
		}
		if filters.Status != "" && task.Status != filters.Status {
			continue
		}
		items = append(items, task)
	}
	return items, &pagination.PaginationResult{Total: int64(len(items)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (r *imageWorkspaceRepoFake) GetTask(ctx context.Context, userID int64, taskID int64) (*ImageWorkspaceTask, error) {
	task, ok := r.tasks[taskID]
	if !ok || task.UserID != userID {
		return nil, ErrImageWorkspaceTaskNotFound
	}
	return &task, nil
}

func (r *imageWorkspaceRepoFake) ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]ImageWorkspaceArtifact, error) {
	items := make([]ImageWorkspaceArtifact, 0)
	for _, artifact := range r.artifacts[taskID] {
		if artifact.UserID == userID {
			items = append(items, artifact)
		}
	}
	return items, nil
}

func (r *imageWorkspaceRepoFake) GetArtifact(ctx context.Context, userID int64, artifactID int64) (*ImageWorkspaceArtifact, error) {
	for _, artifacts := range r.artifacts {
		for _, artifact := range artifacts {
			if artifact.ID == artifactID && artifact.UserID == userID {
				return &artifact, nil
			}
		}
	}
	return nil, ErrImageWorkspaceTaskNotFound
}

func (r *imageWorkspaceRepoFake) ListTemplates(ctx context.Context, userID int64) ([]ImageWorkspaceTemplate, error) {
	items := make([]ImageWorkspaceTemplate, 0, len(r.templates))
	for _, template := range r.templates {
		if template.UserID == userID {
			items = append(items, template)
		}
	}
	return items, nil
}

func (r *imageWorkspaceRepoFake) UpsertTemplate(ctx context.Context, template *ImageWorkspaceTemplate) error {
	if template == nil {
		return nil
	}
	if template.ID == 0 {
		template.ID = r.nextID
		r.nextID++
	}
	now := time.Now().UTC()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	template.UpdatedAt = now
	r.templates[template.ID] = *template
	return nil
}

func (r *imageWorkspaceRepoFake) DeleteTemplate(ctx context.Context, userID int64, templateID int64) error {
	delete(r.templates, templateID)
	return nil
}

func (r *imageWorkspaceRepoFake) ListUsageRecords(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ImageWorkspaceUsageRecord, *pagination.PaginationResult, error) {
	return []ImageWorkspaceUsageRecord{}, &pagination.PaginationResult{Total: 0, Page: params.Page, PageSize: params.PageSize, Pages: 0}, nil
}

func (r *imageWorkspaceRepoFake) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*ImageWorkspaceTask, error) {
	for id, task := range r.tasks {
		if task.Status != ImageWorkspaceTaskStatusQueued {
			continue
		}
		leaseUntil := time.Now().UTC().Add(time.Duration(leaseSeconds) * time.Second)
		task.Status = ImageWorkspaceTaskStatusRunning
		task.WorkerLeaseUntil = &leaseUntil
		task.ErrorMessage = ""
		task.UpdatedAt = time.Now().UTC()
		r.tasks[id] = task
		return &task, nil
	}
	return nil, nil
}

func (r *imageWorkspaceRepoFake) CompleteTask(ctx context.Context, taskID int64, artifacts []ImageWorkspaceArtifact, resultJSON string, cost float64) (*ImageWorkspaceTask, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, ErrImageWorkspaceTaskNotFound
	}
	if task.Status == ImageWorkspaceTaskStatusSucceeded || task.Status == ImageWorkspaceTaskStatusFailed || task.Status == ImageWorkspaceTaskStatusCancelled {
		return &task, nil
	}
	task.Status = ImageWorkspaceTaskStatusSucceeded
	task.WorkerLeaseUntil = nil
	task.CostEstimate = cost
	task.ResultJSON = resultJSON
	task.ErrorMessage = ""
	task.UpdatedAt = time.Now().UTC()
	r.tasks[taskID] = task
	for i := range artifacts {
		artifacts[i].ID = r.nextID
		r.nextID++
		artifacts[i].TaskID = task.ID
		artifacts[i].UserID = task.UserID
		artifacts[i].CreatedAt = time.Now().UTC()
		r.artifacts[taskID] = append(r.artifacts[taskID], artifacts[i])
	}
	return &task, nil
}

func (r *imageWorkspaceRepoFake) FailTask(ctx context.Context, taskID int64, message string, resultJSON string) (*ImageWorkspaceTask, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, ErrImageWorkspaceTaskNotFound
	}
	task.Status = ImageWorkspaceTaskStatusFailed
	task.WorkerLeaseUntil = nil
	task.ErrorMessage = message
	task.ResultJSON = resultJSON
	task.UpdatedAt = time.Now().UTC()
	r.tasks[taskID] = task
	return &task, nil
}

func (r *imageWorkspaceRepoFake) CancelTask(ctx context.Context, taskID int64, userID int64) (*ImageWorkspaceTask, error) {
	task, ok := r.tasks[taskID]
	if !ok || task.UserID != userID {
		return nil, ErrImageWorkspaceTaskNotFound
	}
	if task.Status != ImageWorkspaceTaskStatusQueued {
		return nil, ErrImageWorkspaceInvalidInput
	}
	task.Status = ImageWorkspaceTaskStatusCancelled
	task.WorkerLeaseUntil = nil
	task.ErrorMessage = "cancelled by user"
	task.UpdatedAt = time.Now().UTC()
	r.tasks[taskID] = task
	return &task, nil
}

func (r *imageWorkspaceRepoFake) GetWorkerStatus(ctx context.Context) (*ImageWorkspaceWorkerStatus, error) {
	status := &ImageWorkspaceWorkerStatus{}
	var oldestQueuedAt *time.Time
	var lastTaskUpdatedAt *time.Time
	for _, task := range r.tasks {
		status.TotalCount++
		switch task.Status {
		case ImageWorkspaceTaskStatusQueued:
			status.QueuedCount++
			if oldestQueuedAt == nil || task.CreatedAt.Before(*oldestQueuedAt) {
				createdAt := task.CreatedAt
				oldestQueuedAt = &createdAt
			}
		case ImageWorkspaceTaskStatusRunning:
			status.RunningCount++
			if task.WorkerLeaseUntil == nil || task.WorkerLeaseUntil.Before(time.Now().UTC()) {
				status.StaleRunningCount++
			}
		case ImageWorkspaceTaskStatusFailed:
			status.FailedCount++
		case ImageWorkspaceTaskStatusSucceeded:
			status.SucceededCount++
		case ImageWorkspaceTaskStatusCancelled:
			status.CancelledCount++
		}
		if lastTaskUpdatedAt == nil || task.UpdatedAt.After(*lastTaskUpdatedAt) {
			updatedAt := task.UpdatedAt
			lastTaskUpdatedAt = &updatedAt
		}
	}
	for _, artifacts := range r.artifacts {
		status.ArtifactCount += int64(len(artifacts))
	}
	status.OldestQueuedAt = oldestQueuedAt
	status.LastTaskUpdatedAt = lastTaskUpdatedAt
	return status, nil
}

func TestImageWorkspaceTaskLifecycleCompletesWithArtifacts(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"custom-image","label":"Custom Image","provider":"custom","default_size":"512x512","default_quality":"fast","sizes":["512x512"],"qualities":["fast"],"cost_per_image":0.25,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt:         "  make a compact icon  ",
		NegativePrompt: "watermark",
		Model:          "custom-image",
		Provider:       "custom",
		Size:           "512x512",
		Quality:        "fast",
		Style:          "flat vector",
		BatchSize:      2,
	})
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusQueued, task.Status)
	require.Equal(t, "make a compact icon", task.Prompt)
	require.Equal(t, 0.5, task.CostEstimate)

	claimed, err := svc.ClaimNextTask(context.Background(), 30)
	require.NoError(t, err)
	require.NotNil(t, claimed)
	require.Equal(t, task.ID, claimed.ID)
	require.Equal(t, ImageWorkspaceTaskStatusRunning, claimed.Status)
	require.NotNil(t, claimed.WorkerLeaseUntil)

	completed, err := svc.CompleteTask(context.Background(), task.ID, CompleteImageWorkspaceTaskInput{
		ResultJSON: `{"provider":"custom","artifact_count":1}`,
		Cost:       0.75,
		Artifacts: []ImageWorkspaceArtifactInput{{
			StorageProvider: "local",
			StorageKey:      "/tmp/image-workspace/42/1/image-1.png",
			ImageURL:        "http://127.0.0.1/artifacts/image-1.png",
			Prompt:          "revised prompt",
			MimeType:        "image/png",
			Width:           512,
			Height:          512,
			FileSize:        68,
			Checksum:        "abc123",
			MetadataJSON:    `{"source":"mock"}`,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusSucceeded, completed.Status)
	require.Nil(t, completed.WorkerLeaseUntil)
	require.Equal(t, 0.75, completed.CostEstimate)

	loaded, err := svc.GetTask(context.Background(), 42, task.ID)
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusSucceeded, loaded.Status)
	require.Len(t, loaded.Artifacts, 1)
	require.Equal(t, "local", loaded.Artifacts[0].StorageProvider)
	require.Equal(t, int64(68), loaded.Artifacts[0].FileSize)
	require.Equal(t, "abc123", loaded.Artifacts[0].Checksum)
}

func TestImageWorkspaceArtifactInputAcceptsWorkerSnakeCaseJSON(t *testing.T) {
	var input struct {
		Artifacts []ImageWorkspaceArtifactInput `json:"artifacts"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{
		"artifacts": [{
			"storage_provider": "local",
			"storage_key": "/app/data/image-workspace/42/1/image-1.png",
			"image_url": "data:image/png;base64,abc",
			"prompt": "revised prompt",
			"mime_type": "image/png",
			"width": 1024,
			"height": 1024,
			"file_size": 68,
			"checksum": "abc123",
			"metadata_json": "{\"source\":\"worker\"}"
		}]
	}`), &input))

	require.Len(t, input.Artifacts, 1)
	artifact := input.Artifacts[0]
	require.Equal(t, "local", artifact.StorageProvider)
	require.Equal(t, "/app/data/image-workspace/42/1/image-1.png", artifact.StorageKey)
	require.Equal(t, "data:image/png;base64,abc", artifact.ImageURL)
	require.Equal(t, int64(68), artifact.FileSize)
	require.Equal(t, `{"source":"worker"}`, artifact.MetadataJSON)
}

func TestImageWorkspaceTaskLifecycleFailsWithWorkerMessage(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})
	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{Prompt: "make an icon", Model: "gpt-image-2"})
	require.NoError(t, err)
	_, err = svc.ClaimNextTask(context.Background(), 30)
	require.NoError(t, err)

	diagnostics := `{"failure":{"upstream_status":200,"upstream_body_preview":"{\"data\":[]}"}}`
	failed, err := svc.FailTask(context.Background(), task.ID, FailImageWorkspaceTaskInput{Message: "upstream quota exceeded", ResultJSON: diagnostics})
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusFailed, failed.Status)
	require.Equal(t, "upstream quota exceeded", failed.ErrorMessage)
	require.JSONEq(t, diagnostics, failed.ResultJSON)
	require.Nil(t, failed.WorkerLeaseUntil)
}

func TestImageWorkspaceWorkerStatusReportsAttentionForStaleTask(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	now := time.Now().UTC()
	expiredLease := now.Add(-time.Minute)
	repo.tasks[1] = ImageWorkspaceTask{
		ID:               1,
		UserID:           42,
		Status:           ImageWorkspaceTaskStatusRunning,
		Prompt:           "stale task",
		WorkerLeaseUntil: &expiredLease,
		CreatedAt:        now.Add(-10 * time.Minute),
		UpdatedAt:        now.Add(-9 * time.Minute),
	}
	repo.tasks[2] = ImageWorkspaceTask{
		ID:        2,
		UserID:    42,
		Status:    ImageWorkspaceTaskStatusQueued,
		Prompt:    "queued task",
		CreatedAt: now.Add(-6 * time.Minute),
		UpdatedAt: now.Add(-6 * time.Minute),
	}
	repo.artifacts[3] = []ImageWorkspaceArtifact{{ID: 100}, {ID: 101}}
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{}})

	status, err := svc.GetWorkerStatus(context.Background())

	require.NoError(t, err)
	require.Equal(t, "attention", status.Health)
	require.Equal(t, int64(2), status.TotalCount)
	require.Equal(t, int64(1), status.QueuedCount)
	require.Equal(t, int64(1), status.RunningCount)
	require.Equal(t, int64(1), status.StaleRunningCount)
	require.Equal(t, int64(2), status.ArtifactCount)
	require.Contains(t, status.AttentionReasons, "stale_running_tasks")
	require.NotNil(t, status.LastTaskAgeSeconds)
	require.NotNil(t, status.OldestQueuedSeconds)
}

func TestImageWorkspaceCreateTaskRejectsUnknownModel(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	_, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "a sunset",
		Model:  "nonexistent-model",
	})
	require.ErrorIs(t, err, ErrImageWorkspaceModelNotFound)
}

func TestImageWorkspaceCreateTaskRejectsDisabledModel(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-1","label":"GPT Image 1","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":false}]}`,
	}})

	_, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "a sunset",
		Model:  "gpt-image-1",
	})
	require.ErrorIs(t, err, ErrImageWorkspaceModelNotFound)
}

func TestImageWorkspaceCreateTaskUsesModelDefaults(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"custom-image","label":"Custom Image","provider":"custom","default_size":"768x768","default_quality":"draft","sizes":["768x768","1024x1024"],"qualities":["draft","hd"],"cost_per_image":0.5,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "a beautiful landscape",
		Model:  "custom-image",
	})
	require.NoError(t, err)
	require.Equal(t, "custom-image", task.Model)
	require.Equal(t, "custom", task.Provider)
	require.Equal(t, "768x768", task.Size)
	require.Equal(t, "draft", task.Quality)
	require.Equal(t, 0.5, task.CostEstimate)
}

func TestImageWorkspaceCreateTaskUsesInputOverModelDefaults(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"custom-image","label":"Custom Image","provider":"custom","default_size":"768x768","default_quality":"draft","sizes":["768x768","1024x1024"],"qualities":["draft","hd"],"cost_per_image":0.5,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt:   "a beautiful landscape",
		Model:    "custom-image",
		Provider: "override-provider",
		Size:     "1024x1024",
		Quality:  "hd",
	})
	require.NoError(t, err)
	require.Equal(t, "custom-image", task.Model)
	require.Equal(t, "override-provider", task.Provider)
	require.Equal(t, "1024x1024", task.Size)
	require.Equal(t, "hd", task.Quality)
}

func TestImageWorkspaceCreateTaskBatchSizeClamping(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt:    "test",
		Model:     "gpt-image-2",
		BatchSize: 10,
	})
	require.NoError(t, err)
	require.Equal(t, 4, task.BatchSize)
	require.Equal(t, 0.16, task.CostEstimate)
}

func TestImageWorkspaceListTasksFiltersByStatus(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	_, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{Prompt: "task a", Model: "gpt-image-2"})
	require.NoError(t, err)
	_, err = svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{Prompt: "task b", Model: "gpt-image-2"})
	require.NoError(t, err)

	// Claim one task to make it running
	claimed, err := svc.ClaimNextTask(context.Background(), 30)
	require.NoError(t, err)
	require.NotNil(t, claimed)

	// List all tasks
	allTasks, _, err := svc.ListTasks(context.Background(), 42, pagination.PaginationParams{Page: 1, PageSize: 20}, "")
	require.NoError(t, err)
	require.Len(t, allTasks, 2)

	// List only queued tasks
	queuedTasks, _, err := svc.ListTasks(context.Background(), 42, pagination.PaginationParams{Page: 1, PageSize: 20}, ImageWorkspaceTaskStatusQueued)
	require.NoError(t, err)
	require.Len(t, queuedTasks, 1)
	require.Equal(t, ImageWorkspaceTaskStatusQueued, queuedTasks[0].Status)

	// List only running tasks
	runningTasks, _, err := svc.ListTasks(context.Background(), 42, pagination.PaginationParams{Page: 1, PageSize: 20}, ImageWorkspaceTaskStatusRunning)
	require.NoError(t, err)
	require.Len(t, runningTasks, 1)
	require.Equal(t, ImageWorkspaceTaskStatusRunning, runningTasks[0].Status)

	// List failed tasks (none)
	failedTasks, _, err := svc.ListTasks(context.Background(), 42, pagination.PaginationParams{Page: 1, PageSize: 20}, ImageWorkspaceTaskStatusFailed)
	require.NoError(t, err)
	require.Len(t, failedTasks, 0)
}

func TestImageWorkspaceCancelQueuedTask(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{Prompt: "cancel me", Model: "gpt-image-2"})
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusQueued, task.Status)

	cancelled, err := svc.CancelTask(context.Background(), 42, task.ID)
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusCancelled, cancelled.Status)
	require.Equal(t, "cancelled by user", cancelled.ErrorMessage)
}

func TestImageWorkspaceCancelRunningTaskFails(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{Prompt: "running task", Model: "gpt-image-2"})
	require.NoError(t, err)
	_, err = svc.ClaimNextTask(context.Background(), 30)
	require.NoError(t, err)

	_, err = svc.CancelTask(context.Background(), 42, task.ID)
	require.ErrorIs(t, err, ErrImageWorkspaceInvalidInput)
}

func TestImageWorkspaceCancelNonexistentTaskFails(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{})

	_, err := svc.CancelTask(context.Background(), 42, 9999)
	require.ErrorIs(t, err, ErrImageWorkspaceTaskNotFound)
}

func TestImageWorkspaceRetryFailedTask(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt:   "a sunny beach",
		Model:    "gpt-image-2",
		Provider: "openai",
		Size:     "1024x1024",
		Quality:  "standard",
	})
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusQueued, task.Status)

	// Simulate worker claiming and failing the task
	_, err = svc.ClaimNextTask(context.Background(), 30)
	require.NoError(t, err)
	_, err = svc.FailTask(context.Background(), task.ID, FailImageWorkspaceTaskInput{Message: "fetch failed"})
	require.NoError(t, err)

	// Verify task is failed
	loaded, err := svc.GetTask(context.Background(), 42, task.ID)
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusFailed, loaded.Status)
	require.Equal(t, "fetch failed", loaded.ErrorMessage)

	// Retry the failed task
	retried, err := svc.RetryTask(context.Background(), 42, task.ID)
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusQueued, retried.Status)
	require.NotEqual(t, task.ID, retried.ID, "retry should create a new task")
	require.Equal(t, task.Prompt, retried.Prompt)
	require.Equal(t, task.Model, retried.Model)
	require.Equal(t, task.Size, retried.Size)
	require.Equal(t, task.Quality, retried.Quality)
	require.Equal(t, task.Provider, retried.Provider)
	require.Equal(t, task.NegativePrompt, retried.NegativePrompt)
	require.Equal(t, task.Style, retried.Style)
	require.Equal(t, task.BatchSize, retried.BatchSize)
	require.Equal(t, task.CostEstimate, retried.CostEstimate)
	require.Empty(t, retried.ErrorMessage)

	// Original task remains failed
	original, err := svc.GetTask(context.Background(), 42, task.ID)
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusFailed, original.Status)
}

func TestImageWorkspaceRetryPolicyViolationTaskFails(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "policy blocked image",
		Model:  "gpt-image-2",
	})
	require.NoError(t, err)
	_, err = svc.ClaimNextTask(context.Background(), 30)
	require.NoError(t, err)
	_, err = svc.FailTask(context.Background(), task.ID, FailImageWorkspaceTaskInput{
		Message:    "upstream safety_error: request violates content policy",
		ResultJSON: `{"error":{"type":"safety_error","message":"flagged by content policy"}}`,
	})
	require.NoError(t, err)

	_, err = svc.RetryTask(context.Background(), 42, task.ID)

	require.ErrorIs(t, err, ErrImageWorkspaceNonRetryable)
	require.Equal(t, int64(2), repo.nextID, "non-retryable failures must not enqueue replacement tasks")
}

func TestImageWorkspaceRetryCancelledTask(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{Prompt: "cancel then retry", Model: "gpt-image-2"})
	require.NoError(t, err)

	_, err = svc.CancelTask(context.Background(), 42, task.ID)
	require.NoError(t, err)

	retried, err := svc.RetryTask(context.Background(), 42, task.ID)
	require.NoError(t, err)
	require.Equal(t, ImageWorkspaceTaskStatusQueued, retried.Status)
	require.NotEqual(t, task.ID, retried.ID)
	require.Equal(t, "cancel then retry", retried.Prompt)
}

func TestImageWorkspaceRetryRunningTaskFails(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{Prompt: "running task", Model: "gpt-image-2"})
	require.NoError(t, err)
	_, err = svc.ClaimNextTask(context.Background(), 30)
	require.NoError(t, err)

	_, err = svc.RetryTask(context.Background(), 42, task.ID)
	require.ErrorIs(t, err, ErrImageWorkspaceInvalidInput)
}

func TestImageWorkspaceRetrySucceededTaskFails(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{Prompt: "succeeded task", Model: "gpt-image-2"})
	require.NoError(t, err)
	_, err = svc.ClaimNextTask(context.Background(), 30)
	require.NoError(t, err)
	_, err = svc.CompleteTask(context.Background(), task.ID, CompleteImageWorkspaceTaskInput{
		ResultJSON: `{"artifact_count":1}`,
		Cost:       0.04,
		Artifacts: []ImageWorkspaceArtifactInput{{
			StorageProvider: "local",
			StorageKey:      "/tmp/test.png",
			ImageURL:        "http://127.0.0.1/test.png",
			MimeType:        "image/png",
			Prompt:          "test",
		}},
	})
	require.NoError(t, err)

	_, err = svc.RetryTask(context.Background(), 42, task.ID)
	require.ErrorIs(t, err, ErrImageWorkspaceInvalidInput)
}

// imageWorkspaceBalanceReaderStub is a minimal UserBalanceReader for tests.
type imageWorkspaceBalanceReaderStub struct {
	balance float64
	err     error
	calls   int
}

func (s *imageWorkspaceBalanceReaderStub) GetUserBalance(_ context.Context, _ int64) (float64, error) {
	s.calls++
	return s.balance, s.err
}

func newImageWorkspaceServiceWithPromptFilter(t *testing.T, repo ImageWorkspaceRepository, filterJSON string) *ImageWorkspaceService {
	t.Helper()
	settingRepo := imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}}
	if filterJSON != "" {
		settingRepo.values[SettingKeyImagePromptFilterConfig] = filterJSON
	}
	svc := NewImageWorkspaceService(repo, settingRepo)
	svc.settingService = NewSettingService(settingRepo, nil)
	return svc
}

func TestCreateTask_PromptSafetyFilterBlocksAndAllows(t *testing.T) {
	filterJSON := `{"enabled":true,"explicit_keywords":["lingerie","nude"],"youth_context_keywords":["school uniform","teen"],"warning_message":"explicit blocked","youth_warning_message":"youth blocked"}`

	cases := []struct {
		name      string
		prompt    string
		wantBlock bool
		wantMsg   string
	}{
		{"explicit_plus_youth_context", "a lingerie school uniform outfit", true, "explicit blocked"},
		{"explicit_plus_young_word", "lingerie on a young model", true, "youth blocked"},
		{"clean_prompt", "a calm mountain landscape at dusk", false, ""},
		{"explicit_without_youth_context", "a lingerie catalog photo", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newImageWorkspaceRepoFake()
			svc := newImageWorkspaceServiceWithPromptFilter(t, repo, filterJSON)

			task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
				Prompt: tc.prompt,
				Model:  "gpt-image-2",
			})
			if tc.wantBlock {
				require.ErrorIs(t, err, ErrImageWorkspacePromptBlocked, "prompt=%q", tc.prompt)
				require.Nil(t, task)
				if tc.wantMsg != "" {
					require.Contains(t, err.Error(), tc.wantMsg)
				}
				require.Empty(t, repo.tasks, "blocked task must not be persisted")
				return
			}
			require.NoError(t, err, "prompt=%q", tc.prompt)
			require.NotNil(t, task)
			require.Len(t, repo.tasks, 1, "clean task must be persisted")
		})
	}
}

func TestCreateTask_PromptSafetyFilterDisabledAllowsAll(t *testing.T) {
	// Filter present but disabled → no blocking.
	filterJSON := `{"enabled":false,"explicit_keywords":["lingerie"],"youth_context_keywords":["teen"],"warning_message":"x"}`
	repo := newImageWorkspaceRepoFake()
	svc := newImageWorkspaceServiceWithPromptFilter(t, repo, filterJSON)

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "lingerie teen outfit",
		Model:  "gpt-image-2",
	})
	require.NoError(t, err)
	require.NotNil(t, task)
}

func TestCreateTask_BalancePreCheckRejectsInsufficient(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})
	// settingService stays nil → prompt filter is a no-op.
	reader := &imageWorkspaceBalanceReaderStub{balance: 0.01} // < 0.04 estimate
	svc.balanceReader = reader

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "a calm mountain landscape",
		Model:  "gpt-image-2",
	})
	require.ErrorIs(t, err, billing.ErrInsufficientBalance)
	require.Nil(t, task)
	require.Equal(t, 1, reader.calls, "balance pre-check must consult the reader")
	require.Empty(t, repo.tasks, "insufficient task must not reach the repo / open a tx")
}

func TestCreateTask_BalancePreCheckPassesWhenSufficient(t *testing.T) {
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})
	reader := &imageWorkspaceBalanceReaderStub{balance: 1.00} // >= 0.04
	svc.balanceReader = reader

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "a calm mountain landscape",
		Model:  "gpt-image-2",
	})
	require.NoError(t, err)
	require.NotNil(t, task)
	require.Len(t, repo.tasks, 1, "sufficient task must be persisted")
}

func TestCreateTask_BalancePreCheckSkippedOnReaderError(t *testing.T) {
	// When the balance reader errors (e.g. cache/DB hiccup) the pre-check must
	// degrade to letting the repo's authoritative gate decide, not reject.
	repo := newImageWorkspaceRepoFake()
	svc := NewImageWorkspaceService(repo, imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig: `{"models":[{"id":"gpt-image-2","label":"GPT Image 2","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
	}})
	reader := &imageWorkspaceBalanceReaderStub{balance: 0, err: context.Canceled}
	svc.balanceReader = reader

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "a calm mountain landscape",
		Model:  "gpt-image-2",
	})
	require.NoError(t, err, "reader error must fall through to repo, not reject")
	require.NotNil(t, task)
	require.Len(t, repo.tasks, 1)
}

func TestImageWordMatch(t *testing.T) {
	cases := []struct {
		text string
		term string
		want bool
	}{
		{"a lingerie dress", "lingerie", true},
		{"LINGERIE dress", "lingerie", true},       // case-insensitive
		{"a lingeire dress", "lingerie", false},    // typo, no match
		{"pre-lingerie-post", "lingerie", true},    // hyphens are non-word → boundary
		{"school uniform", "school uniform", true}, // multi-word literal
		{"schooluniform", "school uniform", false}, // missing space
		{"a young girl", "young", true},
		{"younger girl", "young", false}, // word boundary prevents substring
		{"", "lingerie", false},
		{"dress", "", false}, // empty term never matches
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, imageWordMatch(tc.text, tc.term), "text=%q term=%q", tc.text, tc.term)
	}
}

func TestCreateTask_PromptSafetyRuntimeToggleDisablesFilter(t *testing.T) {
	// When image_workspace_prompt_safety_enabled is "false", the server-side
	// filter must skip even though image_prompt_filter_config.enabled is true —
	// matching the worker, which also stops checking under that toggle.
	filterJSON := `{"enabled":true,"explicit_keywords":["lingerie"],"youth_context_keywords":["teen"],"warning_message":"blocked"}`
	repo := newImageWorkspaceRepoFake()
	settingRepo := imageWorkspaceSettingRepo{values: map[string]string{
		SettingKeyImageWorkspaceModelConfig:         `{"models":[{"id":"gpt-image-2","label":"G","provider":"openai","default_size":"1024x1024","default_quality":"standard","sizes":["1024x1024"],"qualities":["standard"],"cost_per_image":0.04,"enabled":true}]}`,
		SettingKeyImagePromptFilterConfig:           filterJSON,
		SettingKeyImageWorkspacePromptSafetyEnabled: "false",
	}}
	svc := NewImageWorkspaceService(repo, settingRepo)
	svc.settingService = NewSettingService(settingRepo, nil)

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "lingerie teen outfit",
		Model:  "gpt-image-2",
	})
	require.NoError(t, err, "runtime toggle off must skip the filter")
	require.NotNil(t, task)
	require.Len(t, repo.tasks, 1, "task must be persisted when filter is toggled off")
}

func TestCreateTask_PromptSafetyFilterNormalizesMalformedKeywords(t *testing.T) {
	// Malformed keyword entries (leading/trailing spaces, blanks) must be
	// trimmed and dropped, so a spacey keyword still matches server-side.
	filterJSON := `{"enabled":true,"explicit_keywords":["  lingerie  ","", "   "],"youth_context_keywords":["teen"],"warning_message":"blocked"}`
	repo := newImageWorkspaceRepoFake()
	svc := newImageWorkspaceServiceWithPromptFilter(t, repo, filterJSON)

	task, err := svc.CreateTask(context.Background(), 42, CreateImageWorkspaceTaskInput{
		Prompt: "a lingerie teen outfit",
		Model:  "gpt-image-2",
	})
	require.ErrorIs(t, err, ErrImageWorkspacePromptBlocked, "trimmed 'lingerie' must still match")
	require.Nil(t, task)
	require.Empty(t, repo.tasks)
}

func TestNormalizeImageKeywords(t *testing.T) {
	require.Equal(t, []string{"lingerie", "nude"}, normalizeImageKeywords([]string{"  lingerie  ", "nude", "", "   "}))
	require.Equal(t, []string{"school uniform"}, normalizeImageKeywords([]string{" school uniform "}))
	require.Empty(t, normalizeImageKeywords([]string{"", "  ", "\t"}))
	require.Empty(t, normalizeImageKeywords(nil))
}
