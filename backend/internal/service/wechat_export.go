package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Aias00/cloudbase/internal/pkg/pagination"
	wechatctx "github.com/Aias00/cloudbase/internal/wechat"
)

const (
	WeChatSessionStatusPending       = wechatctx.SessionStatusPending
	WeChatSessionStatusScanConfirmed = wechatctx.SessionStatusScanConfirmed
	WeChatSessionStatusReady         = wechatctx.SessionStatusReady
	WeChatSessionStatusExpired       = wechatctx.SessionStatusExpired

	WeChatArticleSourceSynced     = wechatctx.ArticleSourceSynced
	WeChatArticleSourceDirectLink = wechatctx.ArticleSourceDirectLink

	WeChatExportTaskStatusQueued              = wechatctx.ExportTaskStatusQueued
	WeChatExportTaskStatusRunning             = wechatctx.ExportTaskStatusRunning
	WeChatExportTaskStatusUploading           = wechatctx.ExportTaskStatusUploading
	WeChatExportTaskStatusCompleted           = wechatctx.ExportTaskStatusCompleted
	WeChatExportTaskStatusCompletedWithErrors = wechatctx.ExportTaskStatusCompletedWithErrors
	WeChatExportTaskStatusFailed              = wechatctx.ExportTaskStatusFailed
	WeChatExportTaskStatusCancelled           = wechatctx.ExportTaskStatusCancelled
)

var (
	ErrWeChatExportNotConfigured = wechatctx.ErrExportNotConfigured
	ErrWeChatSessionNotFound     = wechatctx.ErrSessionNotFound
	ErrWeChatSessionNotReady     = wechatctx.ErrSessionNotReady
	ErrWeChatAccountNotFound     = wechatctx.ErrAccountNotFound
	ErrWeChatArticleNotFound     = wechatctx.ErrArticleNotFound
	ErrWeChatTaskNotFound        = wechatctx.ErrTaskNotFound
	ErrWeChatTaskConflict        = wechatctx.ErrTaskConflict
	ErrWeChatInvalidInput        = wechatctx.ErrInvalidInput
	ErrWeChatInsufficientBalance = wechatctx.ErrInsufficientBalance
	ErrWeChatArticleVerifyPage   = wechatctx.ErrArticleVerifyPage
)

const (
	// DefaultWeChatExportCostPerArticle is the default cost per article per format for wechat export billing.
	DefaultWeChatExportCostPerArticle = wechatctx.DefaultExportCostPerArticle
	// WeChatExportEngagementCostMultiplier is the cost multiplier when engagement data is included.
	WeChatExportEngagementCostMultiplier = wechatctx.ExportEngagementCostMultiplier
)

var weChatDirectImportProbe = probeWeChatDirectArticleLink
var weChatGatewayClient = wechatctx.NewGatewayClient(nil)

type WeChatExportFormat = wechatctx.ExportFormat
type WeChatSession = wechatctx.Session
type WeChatAccount = wechatctx.Account
type WeChatArticle = wechatctx.Article
type WeChatExportTask = wechatctx.ExportTask
type WeChatExportWorkerStatus = wechatctx.ExportWorkerStatus
type WeChatExportArtifact = wechatctx.ExportArtifact
type WeChatExportTaskLog = wechatctx.ExportTaskLog
type WeChatArticleEngagementResult = wechatctx.ArticleEngagementResult
type WeChatExportRepository = wechatctx.ExportRepository

const (
	WeChatExportFormatHTML     = wechatctx.ExportFormatHTML
	WeChatExportFormatMarkdown = wechatctx.ExportFormatMarkdown
)

type CreateWeChatExportTaskInput struct {
	ArticleIDs        []int64
	Formats           []string
	IncludeEngagement bool
	RetentionDays     int
}

type CompleteWeChatExportTaskInput struct {
	Artifacts          []WeChatExportArtifactInput
	ResultManifestJSON string
	FailedArticleCount int
}

type BindWeChatAccountInput struct {
	FakeID      string `json:"fakeid"`
	Nickname    string `json:"nickname"`
	Alias       string `json:"alias"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
}

type WeChatAccountSyncResult struct {
	FakeID      string `json:"fakeid"`
	SyncedCount int    `json:"synced_count"`
	PageCount   int    `json:"page_count"`
	TotalCount  int    `json:"total_count"`
	HasMore     bool   `json:"has_more"`
}

type wechatSearchBizResponse struct {
	BaseResp wechatGatewayBaseResp `json:"base_resp"`
	List     []struct {
		FakeID       string `json:"fakeid"`
		Nickname     string `json:"nickname"`
		Alias        string `json:"alias"`
		RoundHeadImg string `json:"round_head_img"`
		Signature    string `json:"signature"`
	} `json:"list"`
	Total int `json:"total"`
}

type EnrichWeChatArticleInput struct {
	ArticleID      int64
	UserID         int64
	AccountFakeID  string
	AccountName    string
	AccountAlias   string
	AccountAvatar  string
	AccountDesc    string
	Title          string
	Author         string
	Cover          string
	Digest         string
	PublishAt      *time.Time
	IsOriginal     bool
	IsPaySubscribe bool
	ContentStatus  string
	MetadataJSON   string
}

type FetchWeChatArticleEngagementInput struct {
	ArticleID    int64
	UserID       int64
	Link         string
	MetadataJSON string
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

type AddWeChatExportTaskLogInput struct {
	Event    string
	Status   string
	Message  string
	MetaJSON string
}

type WeChatExportService struct {
	repo          WeChatExportRepository
	settingGetter CostPerArticleGetter
}

// CostPerArticleGetter abstracts fetching the per-article cost setting.
type CostPerArticleGetter interface {
	GetWeChatExportCostPerArticle() float64
	GetWeChatExportWorkerRuntimeConfig(ctx context.Context) (WeChatExportWorkerRuntimeConfig, error)
}

func (s *WeChatExportService) GetWorkerRuntimeConfig(ctx context.Context) (WeChatExportWorkerRuntimeConfig, error) {
	if s == nil || s.settingGetter == nil {
		return defaultWeChatExportWorkerRuntimeConfig(), nil
	}
	return s.settingGetter.GetWeChatExportWorkerRuntimeConfig(ctx)
}

func NewWeChatExportService(repo WeChatExportRepository, settingGetter CostPerArticleGetter) *WeChatExportService {
	return &WeChatExportService{repo: repo, settingGetter: settingGetter}
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
	sessionID, err := randomWeChatLoginToken()
	if err != nil {
		return nil, "", err
	}
	cookies, qrcodeURL, err := startWeChatQRCodeLogin(ctx, sessionID)
	if err != nil {
		return nil, "", err
	}
	encryptedCookies, err := encryptWeChatCookiePayload(wechatCookiePayload{CookieHeader: cookies})
	if err != nil {
		return nil, "", err
	}
	expiresAt := time.Now().UTC().Add(2 * time.Minute)
	if err := s.repo.ExpireLoginAttemptSessions(ctx, userID); err != nil {
		return nil, "", err
	}
	session := &WeChatSession{
		UserID:           userID,
		Status:           WeChatSessionStatusPending,
		LoginToken:       sessionID,
		CookiesEncrypted: encryptedCookies,
		ExpiresAt:        &expiresAt,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, "", err
	}
	return session, qrcodeURL, nil
}

func (s *WeChatExportService) PollSession(ctx context.Context, userID int64, sessionID int64) (*WeChatSession, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || sessionID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	session, err := s.repo.GetSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != WeChatSessionStatusPending && session.Status != WeChatSessionStatusScanConfirmed {
		return session, nil
	}
	now := time.Now().UTC()
	if session.ExpiresAt != nil && session.ExpiresAt.Before(now) {
		session.Status = WeChatSessionStatusExpired
		if updateErr := s.repo.UpdateSession(ctx, session); updateErr != nil {
			return nil, updateErr
		}
		return session, nil
	}
	cookies, err := decryptWeChatCookiePayload(session.CookiesEncrypted)
	if err != nil {
		return nil, err
	}
	pollResult, err := pollWeChatQRCodeLogin(ctx, cookies.CookieHeader)
	if err != nil {
		return nil, err
	}
	lastValidatedAt := now
	session.LastValidatedAt = &lastValidatedAt
	if pollResult.Status != 1 {
		if pollResult.Status == 4 || pollResult.Status == 6 {
			session.Status = WeChatSessionStatusScanConfirmed
		} else {
			session.Status = WeChatSessionStatusPending
		}
		if err := s.repo.UpdateSession(ctx, session); err != nil {
			return nil, err
		}
		return session, nil
	}
	loginResult, err := completeWeChatQRCodeLogin(ctx, cookies.CookieHeader)
	if err != nil {
		return nil, err
	}
	readyExpiresAt := now.Add(4 * 24 * time.Hour)
	encryptedCookies, err := encryptWeChatCookiePayload(wechatCookiePayload{CookieHeader: loginResult.CookieHeader})
	if err != nil {
		return nil, err
	}
	session.Status = WeChatSessionStatusReady
	session.LoginToken = loginResult.Token
	session.CookiesEncrypted = encryptedCookies
	session.LoginAccountName = loginResult.AccountName
	session.ExpiresAt = &readyExpiresAt
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *WeChatExportService) ValidateSession(ctx context.Context, userID int64) (*WeChatSession, error) {
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
	if session == nil || session.Status != WeChatSessionStatusReady {
		return nil, ErrWeChatSessionNotReady
	}
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now().UTC()) {
		session.Status = WeChatSessionStatusExpired
		if updateErr := s.repo.UpdateSession(ctx, session); updateErr != nil {
			return nil, updateErr
		}
		return nil, ErrWeChatSessionNotReady
	}
	cookies, err := decryptWeChatCookiePayload(session.CookiesEncrypted)
	if err != nil {
		return nil, err
	}
	if err := validateWeChatReadySession(ctx, session.LoginToken, cookies.CookieHeader); err != nil {
		session.Status = WeChatSessionStatusExpired
		if updateErr := s.repo.UpdateSession(ctx, session); updateErr != nil {
			return nil, updateErr
		}
		return nil, ErrWeChatSessionNotReady
	}
	now := time.Now().UTC()
	session.LastValidatedAt = &now
	session.ExpiresAt = wechatPtrTime(now.Add(4 * 24 * time.Hour))
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
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

func (s *WeChatExportService) SearchAccounts(ctx context.Context, userID int64, query string, limit int) ([]WeChatAccount, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.SearchAccounts(ctx, userID, strings.TrimSpace(query), limit)
}

func (s *WeChatExportService) SearchRemoteAccounts(ctx context.Context, userID int64, query string, limit int) ([]WeChatAccount, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if userID <= 0 || query == "" {
		return nil, ErrWeChatInvalidInput
	}
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	session, err := s.repo.GetActiveSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Status != WeChatSessionStatusReady {
		return nil, ErrWeChatSessionNotReady
	}
	cookies, err := decryptWeChatCookiePayload(session.CookiesEncrypted)
	if err != nil {
		return nil, err
	}
	resp, err := requestWeChatGateway(ctx, http.MethodGet, "/cgi-bin/searchbiz", url.Values{
		"action": {"search_biz"},
		"begin":  {"0"},
		"count":  {strconv.Itoa(limit)},
		"query":  {query},
		"token":  {session.LoginToken},
		"lang":   {"zh_CN"},
		"f":      {"json"},
		"ajax":   {"1"},
	}, nil, cookies.CookieHeader)
	if err != nil {
		return nil, err
	}
	var payload wechatSearchBizResponse
	decodeErr := decodeWeChatJSON(resp, &payload)
	_ = resp.Body.Close()
	if decodeErr != nil {
		return nil, decodeErr
	}
	if payload.BaseResp.Ret != 0 {
		return nil, fmt.Errorf("wechat account search failed: ret=%d msg=%s", payload.BaseResp.Ret, payload.BaseResp.ErrMsg)
	}
	items := make([]WeChatAccount, 0, len(payload.List))
	for _, item := range payload.List {
		fakeID := strings.TrimSpace(item.FakeID)
		if fakeID == "" {
			continue
		}
		items = append(items, WeChatAccount{
			UserID:      userID,
			FakeID:      fakeID,
			Nickname:    strings.TrimSpace(item.Nickname),
			Alias:       strings.TrimSpace(item.Alias),
			Avatar:      strings.TrimSpace(item.RoundHeadImg),
			Description: strings.TrimSpace(item.Signature),
			IsActive:    true,
		})
	}
	return items, nil
}

func (s *WeChatExportService) BindAccount(ctx context.Context, userID int64, input BindWeChatAccountInput) (*WeChatAccount, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	fakeID := strings.TrimSpace(input.FakeID)
	if userID <= 0 || fakeID == "" {
		return nil, ErrWeChatInvalidInput
	}
	account := &WeChatAccount{
		UserID:      userID,
		FakeID:      fakeID,
		Nickname:    strings.TrimSpace(input.Nickname),
		Alias:       strings.TrimSpace(input.Alias),
		Avatar:      strings.TrimSpace(input.Avatar),
		Description: strings.TrimSpace(input.Description),
		IsActive:    true,
	}
	if account.Nickname == "" {
		account.Nickname = fakeID
	}
	if err := s.repo.UpsertAccount(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *WeChatExportService) SyncAccount(ctx context.Context, userID int64, fakeID string, beginFrom ...int) (*WeChatAccount, *WeChatAccountSyncResult, error) {
	if err := s.Health(ctx); err != nil {
		return nil, nil, err
	}
	fakeID = strings.TrimSpace(fakeID)
	if userID <= 0 || fakeID == "" {
		return nil, nil, ErrWeChatInvalidInput
	}
	// Verify account is already bound before syncing articles
	// Use exact match query to avoid false negatives from ILIKE search
	_, err := s.repo.GetAccount(ctx, userID, fakeID)
	if err != nil {
		if errors.Is(err, ErrWeChatAccountNotFound) {
			return nil, nil, ErrWeChatAccountNotFound
		}
		return nil, nil, err
	}
	session, err := s.repo.GetActiveSession(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if session == nil || session.Status != WeChatSessionStatusReady {
		return nil, nil, ErrWeChatSessionNotReady
	}
	cookies, err := decryptWeChatCookiePayload(session.CookiesEncrypted)
	if err != nil {
		return nil, nil, err
	}
	beginOffset := 0
	if len(beginFrom) > 0 && beginFrom[0] > 0 {
		beginOffset = beginFrom[0]
	}
	result, err := s.syncWeChatAccountArticles(ctx, userID, fakeID, session.LoginToken, cookies.CookieHeader, beginOffset)
	if err != nil {
		return nil, nil, err
	}
	account, err := s.repo.MarkAccountSynced(ctx, userID, fakeID)
	if err != nil {
		return nil, nil, err
	}
	return account, result, nil
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
	if !isAllowedWeChatArticleURL(parsed) {
		return nil, ErrWeChatInvalidInput
	}
	if weChatDirectImportProbe != nil {
		if err := weChatDirectImportProbe(ctx, link); err != nil {
			return nil, err
		}
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

func (s *WeChatExportService) EnrichArticle(ctx context.Context, input EnrichWeChatArticleInput) (*WeChatArticle, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if input.ArticleID <= 0 || input.UserID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	accountFakeID := strings.TrimSpace(input.AccountFakeID)
	accountName := strings.TrimSpace(input.AccountName)
	if accountFakeID != "" && accountName != "" {
		account := &WeChatAccount{
			UserID:      input.UserID,
			FakeID:      accountFakeID,
			Nickname:    accountName,
			Alias:       strings.TrimSpace(input.AccountAlias),
			Avatar:      strings.TrimSpace(input.AccountAvatar),
			Description: strings.TrimSpace(input.AccountDesc),
			IsActive:    true,
		}
		if err := s.repo.UpsertAccount(ctx, account); err != nil {
			return nil, err
		}
	}
	metadataJSON := strings.TrimSpace(input.MetadataJSON)
	if metadataJSON == "" {
		metadataJSON = "{}"
	}
	status := strings.TrimSpace(input.ContentStatus)
	if status == "" {
		status = "fetched"
	}
	article := &WeChatArticle{
		ID:             input.ArticleID,
		UserID:         input.UserID,
		AccountFakeID:  accountFakeID,
		SourceType:     WeChatArticleSourceDirectLink,
		Title:          strings.TrimSpace(input.Title),
		Author:         strings.TrimSpace(input.Author),
		Cover:          strings.TrimSpace(input.Cover),
		Digest:         strings.TrimSpace(input.Digest),
		PublishAt:      input.PublishAt,
		IsOriginal:     input.IsOriginal,
		IsPaySubscribe: input.IsPaySubscribe,
		ContentStatus:  status,
		MetadataJSON:   metadataJSON,
	}
	if err := s.repo.UpdateArticleEnrichment(ctx, article); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *WeChatExportService) FetchArticleEngagement(ctx context.Context, input FetchWeChatArticleEngagementInput) (*WeChatArticleEngagementResult, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if input.ArticleID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	var article *WeChatArticle
	if input.UserID > 0 {
		articles, err := s.repo.ListArticlesByIDs(ctx, input.UserID, []int64{input.ArticleID})
		if err != nil {
			return nil, err
		}
		if len(articles) == 0 {
			return nil, ErrWeChatArticleNotFound
		}
		article = &articles[0]
	} else {
		var err error
		article, err = s.repo.GetArticleByID(ctx, input.ArticleID)
		if err != nil {
			return nil, err
		}
	}
	session, err := s.repo.GetActiveSession(ctx, article.UserID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Status != WeChatSessionStatusReady {
		return &WeChatArticleEngagementResult{
			Status:  "unavailable",
			Message: "ready WeChat session is required for authenticated engagement metrics",
		}, nil
	}
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now().UTC()) {
		session.Status = WeChatSessionStatusExpired
		if updateErr := s.repo.UpdateSession(ctx, session); updateErr != nil {
			return nil, updateErr
		}
		return &WeChatArticleEngagementResult{
			Status:  "unavailable",
			Message: "WeChat session expired; please scan again",
		}, nil
	}
	cookies, err := decryptWeChatCookiePayload(session.CookiesEncrypted)
	if err != nil {
		return nil, err
	}
	link := strings.TrimSpace(input.Link)
	if link == "" {
		link = article.Link
	}
	metadata := mergeWeChatMetadataJSON(article.MetadataJSON, input.MetadataJSON)
	result, err := fetchWeChatArticleEngagement(ctx, link, metadata, cookies.CookieHeader)
	now := time.Now().UTC()
	session.LastValidatedAt = &now
	if err != nil {
		_ = s.repo.UpdateSession(ctx, session)
		return &WeChatArticleEngagementResult{
			Status:  "failed",
			Message: err.Error(),
		}, nil
	}
	session.ExpiresAt = wechatPtrTime(now.Add(4 * 24 * time.Hour))
	if updateErr := s.repo.UpdateSession(ctx, session); updateErr != nil {
		return nil, updateErr
	}
	return result, nil
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
		"estimated_credits":  s.estimateCost(len(articles), len(formats), input.IncludeEngagement),
	}, nil
}

// estimateCost calculates the estimated credit cost for an export task.
func (s *WeChatExportService) estimateCost(articleCount, formatCount int, includeEngagement bool) float64 {
	costPerArticle := DefaultWeChatExportCostPerArticle
	if s.settingGetter != nil {
		if v := s.settingGetter.GetWeChatExportCostPerArticle(); v > 0 {
			costPerArticle = v
		}
	}
	return wechatctx.EstimateExportCost(articleCount, formatCount, includeEngagement, costPerArticle)
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
	retentionDays := wechatctx.NormalizeExportRetentionDays(input.RetentionDays)
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
		IncludeEngagement:    input.IncludeEngagement,
		PayloadJSON:          string(payloadJSON),
		ResultManifestJSON:   "{}",
		RetentionDays:        retentionDays,
		CostEstimate:         s.estimateCost(len(articles), len(formats), input.IncludeEngagement),
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

func (s *WeChatExportService) GetWorkerStatus(ctx context.Context, userID int64) (*WeChatExportWorkerStatus, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	status, err := s.repo.GetWorkerStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	if status == nil {
		status = &WeChatExportWorkerStatus{}
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
		status.Message = "Some running tasks have expired leases and will be reclaimed by the worker."
		status.AttentionReasons = append(status.AttentionReasons, "stale_running_tasks")
	case status.QueuedCount > 0 && status.RunningCount == 0:
		status.Health = "waiting"
		status.Message = "Queued tasks are waiting for a worker to pick them up."
		if status.OldestQueuedSeconds != nil && *status.OldestQueuedSeconds >= 300 {
			status.AttentionReasons = append(status.AttentionReasons, "queued_tasks_waiting_over_5m")
		}
	case status.RunningCount > 0:
		status.Health = "active"
		status.Message = "Worker has running export tasks."
	default:
		status.Health = "idle"
		status.Message = "No queued or running export tasks."
	}
	return status, nil
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

func (s *WeChatExportService) CancelTask(ctx context.Context, userID int64, taskID int64) (*WeChatExportTask, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || taskID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.CancelTask(ctx, userID, taskID)
}

func (s *WeChatExportService) RetryTask(ctx context.Context, userID int64, taskID int64) (*WeChatExportTask, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || taskID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.RetryTask(ctx, userID, taskID)
}

func (s *WeChatExportService) AddTaskLog(ctx context.Context, taskID int64, leaseToken string, input AddWeChatExportTaskLogInput) (*WeChatExportTaskLog, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	// Phase 5：验证leaseToken非空（强制验证，用户决策）
	if strings.TrimSpace(leaseToken) == "" {
		return nil, ErrWeChatInvalidInput
	}
	event := strings.TrimSpace(input.Event)
	if taskID <= 0 || event == "" {
		return nil, ErrWeChatInvalidInput
	}
	metaJSON := strings.TrimSpace(input.MetaJSON)
	if metaJSON == "" {
		metaJSON = "{}"
	}
	return s.repo.AddTaskLog(ctx, taskID, leaseToken, WeChatExportTaskLog{
		TaskID:   taskID,
		Event:    event,
		Status:   strings.TrimSpace(input.Status),
		Message:  strings.TrimSpace(input.Message),
		MetaJSON: metaJSON,
	})
}

func (s *WeChatExportService) ListTaskLogs(ctx context.Context, userID int64, taskID int64) ([]WeChatExportTaskLog, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || taskID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.ListTaskLogs(ctx, userID, taskID)
}

func (s *WeChatExportService) ClaimNextTask(ctx context.Context, leaseSeconds int64) (*WeChatExportTask, []WeChatArticle, string, error) {
	if err := s.Health(ctx); err != nil {
		return nil, nil, "", err
	}
	if leaseSeconds <= 0 {
		leaseSeconds = 300
	}
	return s.repo.ClaimNextTask(ctx, leaseSeconds)
}

func (s *WeChatExportService) CompleteTask(ctx context.Context, taskID int64, leaseToken string, input CompleteWeChatExportTaskInput) (*WeChatExportTask, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	// Phase 5：验证leaseToken非空（强制验证）
	if strings.TrimSpace(leaseToken) == "" {
		return nil, ErrWeChatInvalidInput
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
	failedArticleCount := input.FailedArticleCount
	if failedArticleCount < 0 {
		failedArticleCount = 0
	}

	// Fetch the task to compute actual cost based on successful articles and formats.
	task, err := s.repo.GetTask(ctx, 0, taskID)
	if err != nil {
		return nil, err
	}

	// Calculate actual cost: only charge for successfully exported articles.
	successfulCount := task.SelectedArticleCount - failedArticleCount
	if successfulCount < 0 {
		successfulCount = 0
	}
	actualCost := s.estimateCost(successfulCount, len(task.Formats), task.IncludeEngagement)

	// Pass the actual cost to the repo for billing adjustment.
	// The repo handles balance reservation/refund and usage record creation.
	return s.repo.CompleteTask(ctx, taskID, leaseToken, artifacts, manifest, failedArticleCount, actualCost)
}

func (s *WeChatExportService) FailTask(ctx context.Context, taskID int64, leaseToken string, message string) (*WeChatExportTask, error) {
	if err := s.Health(ctx); err != nil {
		return nil, err
	}
	// Phase 5：验证leaseToken非空（强制验证）
	if strings.TrimSpace(leaseToken) == "" {
		return nil, ErrWeChatInvalidInput
	}
	if taskID <= 0 {
		return nil, ErrWeChatInvalidInput
	}
	return s.repo.FailTask(ctx, taskID, leaseToken, strings.TrimSpace(message))
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
	artifact, err := s.repo.GetArtifact(ctx, userID, artifactID)
	if err != nil {
		return nil, err
	}
	if artifact.ExpiresAt != nil && artifact.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrWeChatTaskNotFound
	}
	return artifact, nil
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
	formats, ok := wechatctx.NormalizeExportFormats(raw)
	if !ok {
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

func isAllowedWeChatArticleURL(parsed *url.URL) bool {
	if parsed == nil {
		return false
	}
	if parsed.Scheme != "https" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "mp.weixin.qq.com" {
		return false
	}
	path := strings.TrimRight(parsed.EscapedPath(), "/")
	return path == "/s" || strings.HasPrefix(path, "/s/") || strings.HasPrefix(path, "/mp/appmsg")
}

func probeWeChatDirectArticleLink(ctx context.Context, link string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", "https://mp.weixin.qq.com/")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil
	}
	if isWeChatVerifyPageHTML(string(body)) {
		return ErrWeChatArticleVerifyPage
	}
	return nil
}

func isWeChatVerifyPageHTML(rawHTML string) bool {
	normalized := strings.ToLower(rawHTML)
	return strings.Contains(normalized, "secitptpage/verify") ||
		strings.Contains(normalized, "mmbizwap:secitptpage/verify.html") ||
		(strings.Contains(normalized, "cap_sid") &&
			strings.Contains(normalized, "poc_token") &&
			strings.Contains(normalized, "target_url"))
}

func maxIntValue(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func wechatPtrTime(value time.Time) *time.Time {
	return &value
}

type wechatCookiePayload = wechatctx.CookiePayload

type wechatGatewayBaseResp struct {
	Ret    int    `json:"ret"`
	ErrMsg string `json:"err_msg"`
}

type wechatScanPollResponse struct {
	BaseResp wechatGatewayBaseResp `json:"base_resp"`
	Status   int                   `json:"status"`
}

type wechatLoginResponse struct {
	BaseResp    wechatGatewayBaseResp `json:"base_resp"`
	RedirectURL string                `json:"redirect_url"`
}

type wechatAppMsgPublishResponse struct {
	BaseResp    wechatGatewayBaseResp `json:"base_resp"`
	PublishPage string                `json:"publish_page"`
}

type wechatAppMsgExtResponse struct {
	BaseResp   wechatGatewayBaseResp `json:"base_resp"`
	AppMsgStat struct {
		ReadNum    int64 `json:"read_num"`
		OldLikeNum int64 `json:"old_like_num"`
		ShareNum   int64 `json:"share_num"`
		LikeNum    int64 `json:"like_num"`
		CommentNum int64 `json:"comment_count"`
	} `json:"appmsgstat"`
}

type wechatPublishPage struct {
	PublishList []struct {
		PublishInfo string `json:"publish_info"`
	} `json:"publish_list"`
	TotalCount int `json:"total_count"`
}

type wechatPublishInfo struct {
	AppMsgEx []struct {
		AID            string `json:"aid"`
		AppMsgID       int64  `json:"appmsgid"`
		AuthorName     string `json:"author_name"`
		Cover          string `json:"cover"`
		CreateTime     int64  `json:"create_time"`
		Digest         string `json:"digest"`
		IsDeleted      bool   `json:"is_deleted"`
		IsPaySubscribe int    `json:"is_pay_subscribe"`
		ItemIdx        int    `json:"itemidx"`
		Link           string `json:"link"`
		Title          string `json:"title"`
		CopyrightStat  int    `json:"copyright_stat"`
	} `json:"appmsgex"`
}

type wechatCompletedLogin struct {
	Token        string
	CookieHeader string
	AccountName  string
}

func (s *WeChatExportService) syncWeChatAccountArticles(ctx context.Context, userID int64, fakeID string, token string, cookieHeader string, beginOffset int) (*WeChatAccountSyncResult, error) {
	// Batch sync strategy:
	// - Backend fetches up to 100 pages per request (1000 articles max)
	// - Frontend auto-continues when has_more=true, with 2-second delay between batches
	// - Added 500ms delay between API requests within batch to prevent WeChat freq control (ret=200013)
	// - This design avoids overloading WeChat API while allowing unlimited total sync
	begin := beginOffset
	const pageSize = 10  // WeChat API page size
	const maxPages = 100 // Backend batch limit: 100 pages = 1000 articles per request
	// Default 500ms delay between requests to prevent WeChat API freq control (ret=200013)
	// Can be overridden via WECHAT_SYNC_REQUEST_DELAY_MS env var (e.g., "1000" for 1 second)
	requestDelay := 500 * time.Millisecond
	if delayMs := os.Getenv("WECHAT_SYNC_REQUEST_DELAY_MS"); delayMs != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(delayMs)); err == nil && parsed > 0 {
			requestDelay = time.Duration(parsed) * time.Millisecond
		}
	}
	result := &WeChatAccountSyncResult{FakeID: fakeID}
	for page := 0; page < maxPages; page++ {
		// Add delay between requests to prevent WeChat API frequency control
		// Skip delay on first request for faster initial response
		if page > 0 {
			time.Sleep(requestDelay)
		}
		resp, err := requestWeChatGateway(ctx, http.MethodGet, "/cgi-bin/appmsgpublish", url.Values{
			"sub":               {"list"},
			"search_field":      {"null"},
			"begin":             {strconv.Itoa(begin)},
			"count":             {strconv.Itoa(pageSize)},
			"query":             {""},
			"fakeid":            {fakeID},
			"type":              {"101_1"},
			"free_publish_type": {"1"},
			"sub_action":        {"list_ex"},
			"token":             {token},
			"lang":              {"zh_CN"},
			"f":                 {"json"},
			"ajax":              {"1"},
		}, nil, cookieHeader)
		if err != nil {
			return nil, err
		}
		var payload wechatAppMsgPublishResponse
		decodeErr := decodeWeChatJSON(resp, &payload)
		_ = resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}
		if payload.BaseResp.Ret != 0 {
			return nil, fmt.Errorf("wechat account sync failed: ret=%d msg=%s", payload.BaseResp.Ret, payload.BaseResp.ErrMsg)
		}
		var publishPage wechatPublishPage
		if strings.TrimSpace(payload.PublishPage) != "" {
			if err := json.Unmarshal([]byte(payload.PublishPage), &publishPage); err != nil {
				return nil, fmt.Errorf("wechat publish_page invalid json: %w", err)
			}
		}
		result.TotalCount = publishPage.TotalCount
		syncedInPage := 0
		for _, item := range publishPage.PublishList {
			if strings.TrimSpace(item.PublishInfo) == "" {
				continue
			}
			var info wechatPublishInfo
			if err := json.Unmarshal([]byte(item.PublishInfo), &info); err != nil {
				return nil, fmt.Errorf("wechat publish_info invalid json: %w", err)
			}
			for _, appmsg := range info.AppMsgEx {
				link := strings.TrimSpace(appmsg.Link)
				if link == "" {
					continue
				}
				var publishAt *time.Time
				if appmsg.CreateTime > 0 {
					parsed := time.Unix(appmsg.CreateTime, 0).UTC()
					publishAt = &parsed
				}
				metadata, err := json.Marshal(map[string]any{
					"aid":           appmsg.AID,
					"appmsgid":      appmsg.AppMsgID,
					"itemidx":       appmsg.ItemIdx,
					"createTime":    appmsg.CreateTime,
					"copyrightStat": appmsg.CopyrightStat,
					"source":        "wechat_appmsgpublish",
				})
				if err != nil {
					return nil, err
				}
				status := "normal"
				if appmsg.IsDeleted {
					status = "deleted"
				}
				if err := s.repo.UpsertArticle(ctx, &WeChatArticle{
					UserID:         userID,
					AccountFakeID:  fakeID,
					SourceType:     WeChatArticleSourceSynced,
					Title:          strings.TrimSpace(appmsg.Title),
					Author:         strings.TrimSpace(appmsg.AuthorName),
					Link:           link,
					Cover:          strings.TrimSpace(appmsg.Cover),
					Digest:         strings.TrimSpace(appmsg.Digest),
					PublishAt:      publishAt,
					IsOriginal:     appmsg.CopyrightStat == 11,
					IsPaySubscribe: appmsg.IsPaySubscribe == 1,
					ContentStatus:  status,
					MetadataJSON:   string(metadata),
				}); err != nil {
					return nil, err
				}
				syncedInPage++
			}
		}
		result.PageCount++
		result.SyncedCount += syncedInPage
		result.HasMore = result.TotalCount > begin+pageSize && syncedInPage > 0
		if syncedInPage == 0 {
			return result, nil
		}
		begin += pageSize
	}
	result.HasMore = true
	return result, nil
}

func fetchWeChatArticleEngagement(ctx context.Context, rawLink string, metadata map[string]any, cookieHeader string) (*WeChatArticleEngagementResult, error) {
	link := strings.TrimSpace(rawLink)
	params, _ := url.ParseQuery("")
	if parsed, err := url.Parse(link); err == nil && parsed != nil {
		params = parsed.Query()
	}
	biz := firstWeChatMetadataString(metadata, "__biz", "biz", "fakeid")
	if biz == "" {
		biz = strings.TrimSpace(params.Get("__biz"))
	}
	mid := firstWeChatMetadataString(metadata, "mid", "appmsgid")
	if mid == "" {
		mid = strings.TrimSpace(params.Get("mid"))
	}
	idx := firstWeChatMetadataString(metadata, "idx", "itemidx")
	if idx == "" {
		idx = strings.TrimSpace(params.Get("idx"))
	}
	if idx == "" {
		idx = "1"
	}
	sn := firstWeChatMetadataString(metadata, "sn")
	if sn == "" {
		sn = strings.TrimSpace(params.Get("sn"))
	}
	appMsgToken := firstWeChatMetadataString(metadata, "appmsg_token", "appmsgToken")
	if appMsgToken == "" {
		appMsgToken = strings.TrimSpace(params.Get("appmsg_token"))
	}
	existing := engagementResultFromMetadata(metadata)
	if biz == "" || mid == "" || idx == "" || sn == "" || appMsgToken == "" {
		existing.Status = "unavailable"
		existing.Message = "missing __biz/mid/idx/sn/appmsg_token required by getappmsgext"
		return existing, nil
	}
	resp, err := requestWeChatGatewayWithReferer(ctx, http.MethodPost, "/mp/getappmsgext", url.Values{
		"__biz":        {biz},
		"mid":          {mid},
		"idx":          {idx},
		"sn":           {sn},
		"appmsg_token": {appMsgToken},
		"x5":           {"0"},
		"f":            {"json"},
	}, url.Values{
		"is_only_read": {"1"},
		"is_temp_url":  {"0"},
		"appmsg_type":  {"9"},
	}, cookieHeader, link)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var payload wechatAppMsgExtResponse
	if err := decodeWeChatJSON(resp, &payload); err != nil {
		return nil, err
	}
	if payload.BaseResp.Ret != 0 {
		existing.Status = "failed"
		existing.Message = fmt.Sprintf("ret=%d msg=%s", payload.BaseResp.Ret, payload.BaseResp.ErrMsg)
		return existing, nil
	}
	return &WeChatArticleEngagementResult{
		ReadNum:    wechatInt64Ptr(payload.AppMsgStat.ReadNum),
		OldLikeNum: wechatInt64Ptr(payload.AppMsgStat.OldLikeNum),
		ShareNum:   wechatInt64Ptr(payload.AppMsgStat.ShareNum),
		LikeNum:    wechatInt64Ptr(payload.AppMsgStat.LikeNum),
		CommentNum: wechatInt64Ptr(payload.AppMsgStat.CommentNum),
		Status:     "fetched",
	}, nil
}

func mergeWeChatMetadataJSON(values ...string) map[string]any {
	merged := map[string]any{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		var item map[string]any
		if err := json.Unmarshal([]byte(value), &item); err != nil {
			continue
		}
		for key, field := range item {
			merged[key] = field
		}
	}
	return merged
}

func firstWeChatMetadataString(metadata map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := metadata[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		case float64:
			if typed != 0 {
				return strconv.FormatInt(int64(typed), 10)
			}
		case int64:
			if typed != 0 {
				return strconv.FormatInt(typed, 10)
			}
		case int:
			if typed != 0 {
				return strconv.Itoa(typed)
			}
		case json.Number:
			if strings.TrimSpace(typed.String()) != "" {
				return typed.String()
			}
		}
	}
	return ""
}

func engagementResultFromMetadata(metadata map[string]any) *WeChatArticleEngagementResult {
	return &WeChatArticleEngagementResult{
		ReadNum:    firstWeChatMetadataInt64(metadata, "read_num", "readNum"),
		OldLikeNum: firstWeChatMetadataInt64(metadata, "old_like_num", "oldLikeNum"),
		ShareNum:   firstWeChatMetadataInt64(metadata, "share_num", "shareNum"),
		LikeNum:    firstWeChatMetadataInt64(metadata, "like_num", "likeNum"),
		CommentNum: firstWeChatMetadataInt64(metadata, "comment_num", "commentNum", "comment_count"),
	}
}

func firstWeChatMetadataInt64(metadata map[string]any, keys ...string) *int64 {
	value := firstWeChatMetadataString(metadata, keys...)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func wechatInt64Ptr(value int64) *int64 {
	return &value
}

func startWeChatQRCodeLogin(ctx context.Context, sessionID string) (string, string, error) {
	startForm := url.Values{}
	startForm.Set("userlang", "zh_CN")
	startForm.Set("redirect_url", "")
	startForm.Set("login_type", "3")
	startForm.Set("sessionid", sessionID)
	startForm.Set("token", "")
	startForm.Set("lang", "zh_CN")
	startForm.Set("f", "json")
	startForm.Set("ajax", "1")
	startResp, err := requestWeChatGateway(ctx, http.MethodPost, "/cgi-bin/bizlogin", url.Values{"action": {"startlogin"}}, startForm, "")
	if err != nil {
		return "", "", err
	}
	defer func() { _ = startResp.Body.Close() }()
	if startResp.StatusCode < 200 || startResp.StatusCode >= 300 {
		return "", "", fmt.Errorf("wechat startlogin failed: status %d", startResp.StatusCode)
	}
	cookieHeader := cookieHeaderFromResponse(startResp)
	if cookieHeader == "" {
		return "", "", errors.New("wechat startlogin did not return cookies")
	}
	qrResp, err := requestWeChatGateway(ctx, http.MethodGet, "/cgi-bin/scanloginqrcode", url.Values{
		"action": {"getqrcode"},
		"random": {fmt.Sprintf("%d", time.Now().UnixMilli())},
	}, nil, cookieHeader)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = qrResp.Body.Close() }()
	if qrResp.StatusCode < 200 || qrResp.StatusCode >= 300 {
		return "", "", fmt.Errorf("wechat getqrcode failed: status %d", qrResp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(qrResp.Body, 2*1024*1024))
	if err != nil {
		return "", "", err
	}
	if len(body) == 0 {
		return "", "", errors.New("wechat getqrcode returned empty body")
	}
	contentType := qrResp.Header.Get("Content-Type")
	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil && mediaType != "" {
		contentType = mediaType
	}
	if contentType == "" || strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") {
		contentType = http.DetectContentType(body)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return "", "", fmt.Errorf("wechat getqrcode returned non-image content: %s", contentType)
	}
	return cookieHeader, "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(body), nil
}

func pollWeChatQRCodeLogin(ctx context.Context, cookieHeader string) (*wechatScanPollResponse, error) {
	resp, err := requestWeChatGateway(ctx, http.MethodGet, "/cgi-bin/scanloginqrcode", url.Values{
		"action": {"ask"},
		"token":  {""},
		"lang":   {"zh_CN"},
		"f":      {"json"},
		"ajax":   {"1"},
	}, nil, cookieHeader)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var payload wechatScanPollResponse
	if err := decodeWeChatJSON(resp, &payload); err != nil {
		return nil, err
	}
	if payload.BaseResp.Ret != 0 {
		return nil, fmt.Errorf("wechat poll failed: ret=%d msg=%s", payload.BaseResp.Ret, payload.BaseResp.ErrMsg)
	}
	return &payload, nil
}

func completeWeChatQRCodeLogin(ctx context.Context, cookieHeader string) (*wechatCompletedLogin, error) {
	form := url.Values{}
	form.Set("userlang", "zh_CN")
	form.Set("redirect_url", "")
	form.Set("cookie_forbidden", "0")
	form.Set("cookie_cleaned", "0")
	form.Set("plugin_used", "0")
	form.Set("login_type", "3")
	form.Set("token", "")
	form.Set("lang", "zh_CN")
	form.Set("f", "json")
	form.Set("ajax", "1")
	resp, err := requestWeChatGateway(ctx, http.MethodPost, "/cgi-bin/bizlogin", url.Values{"action": {"login"}}, form, cookieHeader)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var payload wechatLoginResponse
	if err := decodeWeChatJSON(resp, &payload); err != nil {
		return nil, err
	}
	if payload.BaseResp.Ret != 0 {
		return nil, fmt.Errorf("wechat login failed: ret=%d msg=%s", payload.BaseResp.Ret, payload.BaseResp.ErrMsg)
	}
	loginCookies := mergeCookieHeaders(cookieHeaderFromResponse(resp), cookieHeader)
	token := tokenFromWeChatRedirect(payload.RedirectURL)
	if token == "" {
		return nil, errors.New("wechat login response did not include token")
	}
	accountName := fetchWeChatAccountName(ctx, token, loginCookies)
	return &wechatCompletedLogin{Token: token, CookieHeader: loginCookies, AccountName: accountName}, nil
}

func validateWeChatReadySession(ctx context.Context, token string, cookieHeader string) error {
	if strings.TrimSpace(token) == "" || strings.TrimSpace(cookieHeader) == "" {
		return ErrWeChatSessionNotReady
	}
	resp, err := requestWeChatGateway(ctx, http.MethodGet, "/cgi-bin/home", url.Values{
		"t":     {"home/index"},
		"token": {token},
		"lang":  {"zh_CN"},
	}, nil, cookieHeader)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wechat session validate failed: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return err
	}
	text := strings.ToLower(string(body))
	if strings.Contains(text, "login") && !strings.Contains(text, "token") {
		return ErrWeChatSessionNotReady
	}
	return nil
}

func requestWeChatGateway(ctx context.Context, method string, path string, query url.Values, form url.Values, cookieHeader string) (*http.Response, error) {
	return weChatGatewayClient.Do(ctx, method, path, query, form, cookieHeader)
}

func requestWeChatGatewayWithReferer(ctx context.Context, method string, path string, query url.Values, form url.Values, cookieHeader string, referer string) (*http.Response, error) {
	return weChatGatewayClient.DoWithReferer(ctx, method, path, query, form, cookieHeader, referer)
}

func decodeWeChatJSON(resp *http.Response, target any) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wechat gateway failed: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("wechat gateway returned invalid json: %w", err)
	}
	return nil
}

func cookieHeaderFromResponse(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	parts := make([]string, 0, len(resp.Cookies()))
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "" {
			continue
		}
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

func mergeCookieHeaders(primary string, fallback string) string {
	values := map[string]string{}
	order := []string{}
	add := func(header string) {
		for _, part := range strings.Split(header, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			name, value, ok := strings.Cut(part, "=")
			name = strings.TrimSpace(name)
			if !ok || name == "" {
				continue
			}
			if _, exists := values[name]; !exists {
				order = append(order, name)
			}
			values[name] = name + "=" + strings.TrimSpace(value)
		}
	}
	add(fallback)
	add(primary)
	parts := make([]string, 0, len(order))
	for _, name := range order {
		parts = append(parts, values[name])
	}
	return strings.Join(parts, "; ")
}

func tokenFromWeChatRedirect(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return parsed.Query().Get("token")
}

var wechatNickNamePattern = regexp.MustCompile(`wx\.cgiData\.nick_name\s*=\s*"([^"]*)"`)

func fetchWeChatAccountName(ctx context.Context, token string, cookieHeader string) string {
	resp, err := requestWeChatGateway(ctx, http.MethodGet, "/cgi-bin/home", url.Values{
		"t":     {"home/index"},
		"token": {token},
		"lang":  {"zh_CN"},
	}, nil, cookieHeader)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return ""
	}
	matches := wechatNickNamePattern.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return ""
	}
	return htmlEntityUnescape(matches[1])
}

func htmlEntityUnescape(value string) string {
	replacer := strings.NewReplacer(`\"`, `"`, `\u0026`, "&", `\u003c`, "<", `\u003e`, ">")
	return replacer.Replace(value)
}

func encryptWeChatCookiePayload(payload wechatCookiePayload) (string, error) {
	return wechatctx.EncryptCookiePayload(payload)
}

func decryptWeChatCookiePayload(raw string) (wechatCookiePayload, error) {
	return wechatctx.DecryptCookiePayload(raw)
}

func randomWeChatLoginToken() (string, error) {
	var buf [24]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

// Phase 2：生成Worker Lease Token（256-bit，64字符hex）- Public for repository
func GenerateWorkerLeaseToken() (string, error) {
	return wechatctx.GenerateWorkerLeaseToken()
}

// Phase 2：生成Worker Run ID（128-bit，32字符hex，可选）- Public for repository
func GenerateWorkerRunID() (string, error) {
	return wechatctx.GenerateWorkerRunID()
}
