package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	WeChatSessionStatusPending = "pending"
	WeChatSessionStatusReady   = "ready"
	WeChatSessionStatusExpired = "expired"

	WeChatArticleSourceSynced     = "synced"
	WeChatArticleSourceDirectLink = "direct_link"

	WeChatExportTaskStatusQueued              = "queued"
	WeChatExportTaskStatusRunning             = "running"
	WeChatExportTaskStatusUploading           = "uploading"
	WeChatExportTaskStatusCompleted           = "completed"
	WeChatExportTaskStatusCompletedWithErrors = "completed_with_errors"
	WeChatExportTaskStatusFailed              = "failed"
	WeChatExportTaskStatusCancelled           = "cancelled"
)

var (
	ErrWeChatExportNotConfigured = infraerrors.InternalServer("WECHAT_EXPORT_NOT_CONFIGURED", "wechat export capability is not configured")
	ErrWeChatSessionNotFound     = infraerrors.NotFound("WECHAT_SESSION_NOT_FOUND", "wechat session not found")
	ErrWeChatAccountNotFound     = infraerrors.NotFound("WECHAT_ACCOUNT_NOT_FOUND", "wechat account not found")
	ErrWeChatArticleNotFound     = infraerrors.NotFound("WECHAT_ARTICLE_NOT_FOUND", "wechat article not found")
	ErrWeChatTaskNotFound        = infraerrors.NotFound("WECHAT_EXPORT_TASK_NOT_FOUND", "wechat export task not found")
	ErrWeChatInvalidInput        = infraerrors.BadRequest("WECHAT_EXPORT_INVALID_INPUT", "wechat export input is invalid")
)

type WeChatExportFormat string

const (
	WeChatExportFormatHTML     WeChatExportFormat = "html"
	WeChatExportFormatMarkdown WeChatExportFormat = "markdown"
	WeChatExportFormatJSON     WeChatExportFormat = "json"
)

type WeChatSession struct {
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

type WeChatAccount struct {
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

type WeChatArticle struct {
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

type WeChatExportTask struct {
	ID                     int64                `json:"id"`
	UserID                 int64                `json:"user_id"`
	Status                 string               `json:"status"`
	ArticleIDs             []int64              `json:"article_ids"`
	Formats                []WeChatExportFormat `json:"formats"`
	SelectedArticleCount   int                  `json:"selected_article_count"`
	SuccessfulArticleCount int                  `json:"successful_article_count"`
	FailedArticleCount     int                  `json:"failed_article_count"`
	FormatsJSON            string               `json:"formats_json,omitempty"`
	IncludeEngagement      bool                 `json:"include_engagement"`
	PayloadJSON            string               `json:"payload_json,omitempty"`
	ResultManifestJSON     string               `json:"result_manifest_json"`
	ErrorMessage           string               `json:"error_message"`
	WorkerLeaseUntil       *time.Time           `json:"worker_lease_until,omitempty"`
	RetentionDays          int                  `json:"retention_days"`
	ExpiresAt              *time.Time           `json:"expires_at,omitempty"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
}

type WeChatExportArtifact struct {
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

type CreateWeChatExportTaskInput struct {
	ArticleIDs        []int64
	Formats           []string
	IncludeEngagement bool
	RetentionDays     int
}

type CompleteWeChatExportTaskInput struct {
	Artifacts          []WeChatExportArtifactInput
	ResultManifestJSON string
}

type WeChatExportArtifactInput struct {
	Format          string `json:"format"`
	StorageProvider string `json:"storage_provider"`
	StorageKey      string `json:"storage_key"`
	DownloadURL     string `json:"download_url"`
	FileName        string `json:"file_name"`
	FileSize        int64  `json:"file_size"`
	Checksum        string `json:"checksum"`
}

type WeChatExportRepository interface {
	GetActiveSession(ctx context.Context, userID int64) (*WeChatSession, error)
	CreateSession(ctx context.Context, session *WeChatSession) error
	GetSession(ctx context.Context, userID int64, sessionID int64) (*WeChatSession, error)
	ExpireUserSessions(ctx context.Context, userID int64) error
	UpsertArticle(ctx context.Context, article *WeChatArticle) error
	ListArticles(ctx context.Context, userID int64, params pagination.PaginationParams) ([]WeChatArticle, *pagination.PaginationResult, error)
	ListArticlesByIDs(ctx context.Context, userID int64, articleIDs []int64) ([]WeChatArticle, error)
	CreateTask(ctx context.Context, task *WeChatExportTask) error
	ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams) ([]WeChatExportTask, *pagination.PaginationResult, error)
	GetTask(ctx context.Context, userID int64, taskID int64) (*WeChatExportTask, error)
	ClaimNextTask(ctx context.Context, leaseSeconds int64) (*WeChatExportTask, []WeChatArticle, error)
	CompleteTask(ctx context.Context, taskID int64, artifacts []WeChatExportArtifact, resultManifestJSON string) (*WeChatExportTask, error)
	FailTask(ctx context.Context, taskID int64, message string) (*WeChatExportTask, error)
	ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]WeChatExportArtifact, error)
	GetArtifact(ctx context.Context, userID int64, artifactID int64) (*WeChatExportArtifact, error)
}

type WeChatExportService struct {
	repo WeChatExportRepository
}

func NewWeChatExportService(repo WeChatExportRepository) *WeChatExportService {
	return &WeChatExportService{repo: repo}
}

func (s *WeChatExportService) Health(_ context.Context) error {
	if s == nil || s.repo == nil {
		return ErrWeChatExportNotConfigured
	}
	return nil
}

func (s *WeChatExportService) GetSession(ctx context.Context, userID int64) (*WeChatSession, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	session, err := s.repo.GetActiveSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrWeChatSessionNotFound
	}
	return session, nil
}

func (s *WeChatExportService) CreateQRCodeSession(ctx context.Context, userID int64) (*WeChatSession, string, error) {
	if err := s.Health(ctx); err != nil {
		return nil, "", err
	}
	if userID <= 0 {
		return nil, "", ErrWeChatInvalidInput
	}
	token, err := randomWeChatLoginToken()
	if err != nil {
		return nil, "", err
	}
	expiresAt := time.Now().UTC().Add(2 * time.Minute)
	session := &WeChatSession{
		UserID:     userID,
		Status:     WeChatSessionStatusPending,
		LoginToken: token,
		ExpiresAt:  &expiresAt,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, "", err
	}
	return session, "wechat-export://login?token=" + url.QueryEscape(token), nil
}

func (s *WeChatExportService) PollSession(ctx context.Context, userID int64, sessionID int64) (*WeChatSession, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || sessionID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.GetSession(ctx, userID, sessionID)
}

func (s *WeChatExportService) LogoutSession(ctx context.Context, userID int64) error {
	if err := s.Health(ctx); err != nil {
		return err
	}
	if userID <= 0 {
		return ErrWeChatInvalidInput
	}
	return s.repo.ExpireUserSessions(ctx, userID)
}

func (s *WeChatExportService) ImportArticleLink(ctx context.Context, userID int64, rawLink string) (*WeChatArticle, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	link := strings.TrimSpace(rawLink)
	if userID <= 0 || link == "" {
		return nil, ErrWeChatInvalidInput
	}
	parsed, err := url.ParseRequestURI(link)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, ErrWeChatInvalidInput
	}
	article := &WeChatArticle{
		UserID:        userID,
		SourceType:    WeChatArticleSourceDirectLink,
		Title:         link,
		Link:          link,
		ContentStatus: "pending",
		MetadataJSON:  "{}",
	}
	if err := s.repo.UpsertArticle(ctx, article); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *WeChatExportService) ListArticles(ctx context.Context, userID int64, params pagination.PaginationParams) ([]WeChatArticle, *pagination.PaginationResult, error) {
	if err := s.Health(ctx); err != nil {
		return nil, nil, err
	}
	if userID <= 0 {
		return nil, nil, ErrWeChatInvalidInput
	}
	params = normalizeWeChatPagination(params)
	return s.repo.ListArticles(ctx, userID, params)
}

func (s *WeChatExportService) QuoteTask(ctx context.Context, userID int64, input CreateWeChatExportTaskInput) (map[string]any, error) {
	formats, err := normalizeWeChatExportFormats(input.Formats)
	if err != nil {
		return nil, err
	}
	articles, err := s.requireTaskArticles(ctx, userID, input.ArticleIDs)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"article_count":      len(articles),
		"formats":            formats,
		"include_engagement": input.IncludeEngagement,
		"estimated_credits":  len(articles) * maxIntValue(1, len(formats)),
	}, nil
}

func (s *WeChatExportService) CreateTask(ctx context.Context, userID int64, input CreateWeChatExportTaskInput) (*WeChatExportTask, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	formats, err := normalizeWeChatExportFormats(input.Formats)
	if err != nil {
		return nil, err
	}
	articles, err := s.requireTaskArticles(ctx, userID, input.ArticleIDs)
	if err != nil {
		return nil, err
	}
	retentionDays := input.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 7
	}
	formatsJSON, err := json.Marshal(formats)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"article_ids": input.ArticleIDs,
		"formats":     formats,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := &WeChatExportTask{
		UserID:               userID,
		Status:               WeChatExportTaskStatusQueued,
		ArticleIDs:           append([]int64(nil), input.ArticleIDs...),
		Formats:              formats,
		SelectedArticleCount: len(articles),
		FormatsJSON:          string(formatsJSON),
		IncludeEngagement:    input.IncludeEngagement,
		PayloadJSON:          string(payloadJSON),
		ResultManifestJSON:   "{}",
		RetentionDays:        retentionDays,
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *WeChatExportService) ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams) ([]WeChatExportTask, *pagination.PaginationResult, error) {
	if err := s.Health(ctx); err != nil {
		return nil, nil, err
	}
	if userID <= 0 {
		return nil, nil, ErrWeChatInvalidInput
	}
	params = normalizeWeChatPagination(params)
	return s.repo.ListTasks(ctx, userID, params)
}

func (s *WeChatExportService) GetTask(ctx context.Context, userID int64, taskID int64) (*WeChatExportTask, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || taskID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.GetTask(ctx, userID, taskID)
}

func (s *WeChatExportService) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*WeChatExportTask, []WeChatArticle, error) {
	if err := s.Health(ctx); err != nil {
		return nil, nil, err
	}
	if leaseSeconds <= 0 {
		leaseSeconds = 300
	}
	return s.repo.ClaimNextTask(ctx, leaseSeconds)
}

func (s *WeChatExportService) CompleteTask(ctx context.Context, taskID int64, input CompleteWeChatExportTaskInput) (*WeChatExportTask, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if taskID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	artifacts := make([]WeChatExportArtifact, 0, len(input.Artifacts))
	for _, item := range input.Artifacts {
		format := strings.TrimSpace(item.Format)
		if format == "" || strings.TrimSpace(item.FileName) == "" {
			return nil, ErrWeChatInvalidInput
		}
		artifacts = append(artifacts, WeChatExportArtifact{
			Format:          format,
			StorageProvider: strings.TrimSpace(item.StorageProvider),
			StorageKey:      strings.TrimSpace(item.StorageKey),
			DownloadURL:     strings.TrimSpace(item.DownloadURL),
			FileName:        strings.TrimSpace(item.FileName),
			FileSize:        item.FileSize,
			Checksum:        strings.TrimSpace(item.Checksum),
		})
	}
	manifest := strings.TrimSpace(input.ResultManifestJSON)
	if manifest == "" {
		manifest = "{}"
	}
	return s.repo.CompleteTask(ctx, taskID, artifacts, manifest)
}

func (s *WeChatExportService) FailTask(ctx context.Context, taskID int64, message string) (*WeChatExportTask, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if taskID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.FailTask(ctx, taskID, strings.TrimSpace(message))
}

func (s *WeChatExportService) ListArtifacts(ctx context.Context, userID int64, taskID int64) ([]WeChatExportArtifact, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || taskID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.ListArtifacts(ctx, userID, taskID)
}

func (s *WeChatExportService) GetArtifact(ctx context.Context, userID int64, artifactID int64) (*WeChatExportArtifact, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || artifactID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.GetArtifact(ctx, userID, artifactID)
}

func (s *WeChatExportService) requireTaskArticles(ctx context.Context, userID int64, articleIDs []int64) ([]WeChatArticle, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || len(articleIDs) == 0 {
		return nil, ErrWeChatInvalidInput
	}
	seen := make(map[int64]struct{}, len(articleIDs))
	for _, id := range articleIDs {
		if id <= 0 {
			return nil, ErrWeChatInvalidInput
		}
		seen[id] = struct{}{}
	}
	articles, err := s.repo.ListArticlesByIDs(ctx, userID, articleIDs)
	if err != nil {
		return nil, err
	}
	if len(articles) != len(seen) {
		return nil, ErrWeChatArticleNotFound
	}
	return articles, nil
}

func normalizeWeChatExportFormats(raw []string) ([]WeChatExportFormat, error) {
	if len(raw) == 0 {
		raw = []string{string(WeChatExportFormatHTML), string(WeChatExportFormatMarkdown), string(WeChatExportFormatJSON)}
	}
	formats := make([]WeChatExportFormat, 0, len(raw))
	seen := make(map[WeChatExportFormat]struct{}, len(raw))
	for _, item := range raw {
		format := WeChatExportFormat(strings.ToLower(strings.TrimSpace(item)))
		switch format {
		case WeChatExportFormatHTML, WeChatExportFormatMarkdown, WeChatExportFormatJSON:
			if _, ok := seen[format]; !ok {
				seen[format] = struct{}{}
				formats = append(formats, format)
			}
		default:
			return nil, ErrWeChatInvalidInput
		}
	}
	if len(formats) == 0 {
		return nil, ErrWeChatInvalidInput
	}
	return formats, nil
}

func normalizeWeChatPagination(params pagination.PaginationParams) pagination.PaginationParams {
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

func maxIntValue(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func randomWeChatLoginToken() (string, error) {
	var buf [24]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}
