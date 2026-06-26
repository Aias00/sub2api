package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const twitterImporterUserAgent = "sub2api-twitter-importer/1.0"
const promptCatalogImageSyncUserAgent = "sub2api-prompt-catalog-image-sync/1.0"
const promptCatalogImageCacheControl = "public, max-age=31536000, immutable"

var (
	twitterStatusURLPattern = regexp.MustCompile(`(?i)^https?://(?:www\.)?(x\.com|twitter\.com|mobile\.twitter\.com)/(.+)$`)
	pbsImageURLPattern      = regexp.MustCompile(`https?://pbs\.twimg\.com/media/[^\s"'<>\\]+`)
	htmlTagPattern          = regexp.MustCompile(`<[^>]+>`)
	spacePattern            = regexp.MustCompile(`\s+`)
)

type TwitterImportService struct {
	promptService *PromptCatalogService
	httpClient    HTTPClient
	imageStorage  PromptCatalogImageStorage
	now           func() time.Time
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type PromptCatalogImageStorage interface {
	SyncPromptImages(ctx context.Context, item PromptCatalogCase, dropUnsynced bool) (PromptCatalogCase, []string, []string)
}

type TwitterImportInput struct {
	URL       string
	Prompt    string
	Title     string
	Category  string
	ImageURLs []string
	XAuto     *bool
}

type TwitterImportResult struct {
	Item         PromptCatalogCase
	ImageURLs    []string
	UploadedURLs []string
	Warnings     []string
}

type twitterStatusRef struct {
	TweetID       string
	Handle        string
	NormalizedURL string
}

type tweetPayload struct {
	Text         string
	AuthorHandle string
	AuthorName   string
	ImageURLs    []string
}

func NewTwitterImportService(promptService *PromptCatalogService) *TwitterImportService {
	return &TwitterImportService{
		promptService: promptService,
		httpClient:    http.DefaultClient,
		imageStorage:  NewEnvPromptCatalogImageStorage(nil),
		now:           func() time.Time { return time.Now().UTC() },
	}
}

func NewTwitterImportServiceForTest(
	promptService *PromptCatalogService,
	httpClient HTTPClient,
	imageStorage PromptCatalogImageStorage,
	now func() time.Time,
) *TwitterImportService {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if imageStorage == nil {
		imageStorage = noopPromptCatalogImageStorage{}
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &TwitterImportService{
		promptService: promptService,
		httpClient:    httpClient,
		imageStorage:  imageStorage,
		now:           now,
	}
}

func (s *TwitterImportService) Import(ctx context.Context, input TwitterImportInput) (*TwitterImportResult, error) {
	if s == nil || s.promptService == nil {
		return nil, ErrPromptCatalogInvalidInput
	}
	ref, err := parseTwitterStatusURL(input.URL)
	if err != nil {
		return nil, ErrPromptCatalogInvalidInput.WithCause(err)
	}

	result, err := s.Build(ctx, input, ref)
	if err != nil {
		return nil, err
	}

	item, uploadedURLs, storageWarnings := s.imageStorage.SyncPromptImages(ctx, result.Item, true)
	result.Item = item
	result.UploadedURLs = uploadedURLs
	result.Warnings = append(result.Warnings, storageWarnings...)

	importedAt := s.now()
	result.Item.ImportedAt = &importedAt
	result.Item.ImportSource = "twitter"
	if result.Item.RawJSON == "" {
		result.Item.RawJSON = "{}"
	}
	if err := s.promptService.UpsertCase(ctx, &result.Item); err != nil {
		return nil, err
	}

	result.ImageURLs = result.Item.ImageURLs
	return result, nil
}

func (s *TwitterImportService) Build(ctx context.Context, input TwitterImportInput, ref twitterStatusRef) (*TwitterImportResult, error) {
	warnings := []string{}
	shouldTryXAuto := input.XAuto == nil || *input.XAuto

	var xAuto *tweetPayload
	if shouldTryXAuto {
		payload, err := s.fetchXAuto(ctx, ref)
		if err != nil {
			warnings = append(warnings, "x-auto fetch failed: "+err.Error())
		} else {
			xAuto = payload
		}
	}

	oembed, err := s.fetchOEmbed(ctx, ref.NormalizedURL)
	if err != nil {
		warnings = append(warnings, "oEmbed fetch failed: "+err.Error())
	}
	fx, err := s.fetchFxTwitter(ctx, ref.TweetID)
	if err != nil {
		warnings = append(warnings, "FxTwitter fetch failed: "+err.Error())
	}
	page, err := s.fetchPageMetadata(ctx, ref.NormalizedURL)
	if err != nil {
		warnings = append(warnings, "Twitter page fetch failed: "+err.Error())
	}

	prompt := cleanTweetText(firstNonEmpty(
		input.Prompt,
		tweetValueOrEmpty(xAuto, func(v *tweetPayload) string { return v.Text }),
		tweetValueOrEmpty(oembed, func(v *tweetPayload) string { return v.Text }),
		tweetValueOrEmpty(fx, func(v *tweetPayload) string { return v.Text }),
		tweetValueOrEmpty(page, func(v *tweetPayload) string { return v.Text }),
	))
	if prompt == "" {
		return nil, ErrPromptCatalogInvalidInput.WithCause(fmt.Errorf("could not extract prompt text from Twitter/X URL"))
	}

	imageURLs := uniqStrings(append(
		append(
			append(
				append([]string{}, input.ImageURLs...),
				imageURLsOrEmpty(xAuto)...,
			),
			imageURLsOrEmpty(fx)...,
		),
		imageURLsOrEmpty(page)...,
	))
	if len(imageURLs) == 0 {
		warnings = append(warnings, "No image URLs were found for the Twitter/X URL")
	}

	handle := firstNonEmpty(ref.Handle, tweetValueOrEmpty(xAuto, func(v *tweetPayload) string { return v.AuthorHandle }), tweetValueOrEmpty(fx, func(v *tweetPayload) string { return v.AuthorHandle }))
	sourceLabel := "Twitter/X"
	if handle != "" {
		sourceLabel = "@" + handle
	} else if xAuto != nil && xAuto.AuthorName != "" {
		sourceLabel = xAuto.AuthorName
	}

	item := PromptCatalogCase{
		ID:            "tw-" + firstNonEmpty(ref.TweetID, md5Hex(ref.NormalizedURL)),
		Title:         firstNonEmpty(input.Title, titleFromPromptCatalog(prompt), "Twitter Prompt "+ref.TweetID),
		Prompt:        prompt,
		PromptPreview: buildPromptCatalogPreview(prompt, 240),
		Category:      firstNonEmpty(input.Category, "Twitter Imports"),
		Tags:          uniqStrings([]string{"twitter", handle}),
		SourceURL:     ref.NormalizedURL,
		ImageURLs:     imageURLs,
		SourceProject: "twitter",
		SourceType:    "case",
		SourceLabel:   sourceLabel,
		Featured:      false,
		Styles:        []string{},
		Scenes:        []string{},
		ImportSource:  "twitter",
		Status:        PromptCatalogStatusPublished,
		RawJSON:       buildTwitterRawJSON(ref, xAuto, oembed, fx, page),
	}
	if len(imageURLs) > 0 {
		item.ImageURL = imageURLs[0]
		item.ImageOriginalURL = imageURLs[0]
	}

	return &TwitterImportResult{
		Item:      item,
		ImageURLs: imageURLs,
		Warnings:  warnings,
	}, nil
}

func parseTwitterStatusURL(input string) (twitterStatusRef, error) {
	raw := strings.TrimSpace(input)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return twitterStatusRef{}, fmt.Errorf("invalid Twitter/X URL")
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	if !twitterStatusURLPattern.MatchString(raw) && host != "x.com" && host != "twitter.com" && host != "mobile.twitter.com" {
		return twitterStatusRef{}, fmt.Errorf("URL must be an x.com or twitter.com status URL")
	}
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	statusIndex := -1
	for i, segment := range segments {
		if segment == "status" || segment == "statuses" {
			statusIndex = i
			break
		}
	}
	if statusIndex < 0 || statusIndex+1 >= len(segments) || !isAllDigits(segments[statusIndex+1]) {
		return twitterStatusRef{}, fmt.Errorf("Twitter/X status ID not found in URL")
	}
	tweetID := segments[statusIndex+1]
	handle := ""
	if statusIndex > 0 {
		handle = segments[0]
	}
	normalized := "https://x.com/i/status/" + tweetID
	if handle != "" {
		normalized = "https://x.com/" + handle + "/status/" + tweetID
	}
	return twitterStatusRef{TweetID: tweetID, Handle: handle, NormalizedURL: normalized}, nil
}

func (s *TwitterImportService) fetchXAuto(ctx context.Context, ref twitterStatusRef) (*tweetPayload, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("X_AUTO_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(strings.TrimSpace(os.Getenv("X_ATUO_BASE_URL")), "/")
	}
	if baseURL == "" {
		return s.fetchXAutoPython(ctx, ref.TweetID)
	}

	if payload, err := s.fetchXAutoURL(ctx, baseURL+"/twitter/articles/"+ref.TweetID); err == nil && payloadHasContent(payload) {
		return payload, nil
	}

	searchURL := baseURL + "/twitter/search?q=" + url.QueryEscape(ref.TweetID) + "&limit=5"
	payload, err := s.fetchJSON(ctx, searchURL)
	if err != nil {
		return nil, err
	}
	items := getPayloadItems(payload)
	for _, item := range items {
		if getPayloadTweetID(item) == ref.TweetID {
			return normalizeTweetPayload(item), nil
		}
	}
	if len(items) > 0 {
		return normalizeTweetPayload(items[0]), nil
	}
	return normalizeTweetPayload(payload), nil
}

func (s *TwitterImportService) fetchXAutoPython(ctx context.Context, tweetID string) (*tweetPayload, error) {
	pythonPath := xAutoPythonPath()
	if pythonPath == "" {
		return nil, fmt.Errorf("vendored x_atuo python path not found")
	}
	configPath := firstNonEmpty(os.Getenv("X_ATUO_CONFIG_PATH"), os.Getenv("X_AUTO_CONFIG_PATH"), defaultXAutoConfigPath())
	pythonBin := firstNonEmpty(os.Getenv("X_ATUO_PYTHON_BIN"), os.Getenv("PYTHON"), "python3")
	timeout := 120 * time.Second
	if raw := strings.TrimSpace(os.Getenv("X_ATUO_TIMEOUT_SECONDS")); raw != "" {
		if parsed, err := time.ParseDuration(raw + "s"); err == nil && parsed > 0 {
			timeout = parsed
		}
	}

	script := `
import json
import os
import sys
from dataclasses import asdict, is_dataclass

tweet_id = sys.argv[1]
module_path = sys.argv[2]
config_path = sys.argv[3]
sys.path.insert(0, module_path)

from x_atuo.core.twitter_client import TwitterClient

def normalize(value):
    if hasattr(value, "model_dump"):
        return value.model_dump(mode="json")
    if is_dataclass(value):
        return asdict(value)
    if isinstance(value, dict):
        return value
    author = getattr(value, "author", None)
    return {
        "tweet_id": getattr(value, "tweet_id", None),
        "text": getattr(value, "text", None),
        "article_text": getattr(value, "article_text", None),
        "raw": getattr(value, "raw", None),
        "author": normalize(author) if author is not None else None,
    }

client = TwitterClient.from_config(
    config_path,
    proxy=os.environ.get("X_ATUO_PROXY_URL") or os.environ.get("X_AUTO_PROXY_URL") or None,
    twitter_bin=os.environ.get("X_ATUO_TWITTER_BIN") or os.environ.get("X_AUTO_TWITTER_BIN") or "twitter",
    timeout=int(os.environ.get("X_ATUO_TIMEOUT_SECONDS") or "120"),
)
print(json.dumps(normalize(client.fetch_tweet(tweet_id)), ensure_ascii=False, default=str))
`
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, pythonBin, "-c", script, tweetID, pythonPath, configPath)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("vendored x_atuo failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	text := strings.TrimSpace(string(output))
	jsonStart := strings.Index(text, "{")
	if jsonStart < 0 {
		return nil, fmt.Errorf("vendored x_atuo returned no JSON payload")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(text[jsonStart:]), &payload); err != nil {
		return nil, err
	}
	return normalizeTweetPayload(payload), nil
}

func xAutoPythonPath() string {
	if explicit := strings.TrimSpace(os.Getenv("X_ATUO_PYTHON_PATH")); explicit != "" {
		return explicit
	}
	candidates := []string{
		filepath.Join("..", "tools", "x-atuo", "src"),
		filepath.Join("tools", "x-atuo", "src"),
		filepath.Join("..", "..", "tools", "x-atuo", "src"),
		filepath.Join("..", "..", "..", "tools", "x-atuo", "src"),
	}
	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && stat.IsDir() {
			abs, err := filepath.Abs(candidate)
			if err == nil {
				return abs
			}
			return candidate
		}
	}
	return ""
}

func defaultXAutoConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".agent-reach/config.yaml"
	}
	return filepath.Join(home, ".agent-reach", "config.yaml")
}

func (s *TwitterImportService) fetchXAutoURL(ctx context.Context, fetchURL string) (*tweetPayload, error) {
	payload, err := s.fetchJSON(ctx, fetchURL)
	if err != nil {
		return nil, err
	}
	return normalizeTweetPayload(payload), nil
}

func (s *TwitterImportService) fetchOEmbed(ctx context.Context, tweetURL string) (*tweetPayload, error) {
	fetchURL := "https://publish.twitter.com/oembed?omit_script=1&dnt=1&url=" + url.QueryEscape(tweetURL)
	payload, err := s.fetchJSON(ctx, fetchURL)
	if err != nil {
		return nil, err
	}
	html := getPayloadString(payload, "html")
	return &tweetPayload{
		Text:       cleanTweetHTMLText(html),
		AuthorName: getPayloadString(payload, "author_name"),
	}, nil
}

func (s *TwitterImportService) fetchFxTwitter(ctx context.Context, tweetID string) (*tweetPayload, error) {
	payload, err := s.fetchJSON(ctx, "https://api.fxtwitter.com/status/"+tweetID)
	if err != nil {
		return nil, err
	}
	tweet := getPayloadMap(payload, "tweet")
	author := getPayloadMap(tweet, "author")
	return &tweetPayload{
		Text:         cleanTweetText(getPayloadString(tweet, "text")),
		AuthorHandle: getPayloadString(author, "screen_name"),
		AuthorName:   getPayloadString(author, "name"),
		ImageURLs:    extractImageURLsFromJSON(tweet),
	}, nil
}

func (s *TwitterImportService) fetchPageMetadata(ctx context.Context, tweetURL string) (*tweetPayload, error) {
	html, err := s.fetchText(ctx, tweetURL)
	if err != nil {
		return nil, err
	}
	description := firstNonEmpty(extractHTMLMeta(html, "og:description"), extractHTMLMeta(html, "twitter:description"))
	images := uniqStrings(append([]string{
		extractHTMLMeta(html, "og:image"),
		extractHTMLMeta(html, "twitter:image"),
	}, extractImageURLsFromText(html)...))
	return &tweetPayload{
		Text:      cleanTweetText(description),
		ImageURLs: images,
	}, nil
}

func (s *TwitterImportService) fetchJSON(ctx context.Context, fetchURL string) (map[string]any, error) {
	text, err := s.fetchText(ctx, fetchURL)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *TwitterImportService) fetchText(ctx context.Context, fetchURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", twitterImporterUserAgent)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%s returned %d", fetchURL, resp.StatusCode)
	}
	return string(body), nil
}

type EnvPromptCatalogImageStorage struct {
	httpClient HTTPClient
}

func NewEnvPromptCatalogImageStorage(httpClient HTTPClient) *EnvPromptCatalogImageStorage {
	if httpClient == nil {
		httpClient = newPromptCatalogImageHTTPClient()
	}
	return &EnvPromptCatalogImageStorage{httpClient: httpClient}
}

func newPromptCatalogImageHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if req == nil || req.URL == nil {
				return fmt.Errorf("invalid redirect image url")
			}
			if err := validatePromptCatalogImageSourceURL(req.URL.String()); err != nil {
				return fmt.Errorf("blocked image redirect: %w", err)
			}
			return nil
		},
	}
}

func (s *EnvPromptCatalogImageStorage) SyncPromptImages(ctx context.Context, item PromptCatalogCase, dropUnsynced bool) (PromptCatalogCase, []string, []string) {
	sourceURLs := getPromptCatalogImageSourceURLs(item)
	if len(sourceURLs) == 0 || allStorageURLs(sourceURLs) {
		return item, sourceURLs, nil
	}
	cfg := loadPromptCatalogR2Config()
	if !cfg.configured() {
		if dropUnsynced {
			clearPromptCatalogImages(&item)
		}
		return item, nil, []string{"R2 image sync skipped: R2 storage is not configured"}
	}
	client, err := newPromptCatalogR2Client(ctx, cfg)
	if err != nil {
		if dropUnsynced {
			clearPromptCatalogImages(&item)
		}
		return item, nil, []string{"R2 image sync skipped: " + err.Error()}
	}

	uploaded := []string{}
	warnings := []string{}
	for index, sourceURL := range sourceURLs {
		if isPromptCatalogStorageURL(sourceURL, cfg) {
			uploaded = append(uploaded, sourceURL)
			continue
		}
		bytes, contentType, err := s.fetchImage(ctx, sourceURL)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("R2 image sync failed for %s: %v", sourceURL, err))
			continue
		}
		extension := imageExtensionFromMime(contentType)
		key := buildPromptCatalogImageKey(item, cfg.uploadPath, index, extension)
		if err := putPromptCatalogR2Object(ctx, client, cfg.bucket, key, bytes, contentType); err != nil {
			warnings = append(warnings, fmt.Sprintf("R2 image sync failed for %s: %v", sourceURL, err))
			continue
		}
		uploaded = append(uploaded, cfg.publicURL(key))
	}
	if len(uploaded) == 0 {
		if dropUnsynced {
			clearPromptCatalogImages(&item)
		}
		return item, uploaded, warnings
	}
	item.ImageURL = uploaded[0]
	item.ImageURLs = uploaded
	item.ImageOriginalURL = uploaded[0]
	item.ImagePreviewURL = ""
	item.ImageThumbURL = ""
	return item, uploaded, warnings
}

func (s *EnvPromptCatalogImageStorage) fetchImage(ctx context.Context, sourceURL string) ([]byte, string, error) {
	data, contentType, err := s.fetchImageOnce(ctx, sourceURL)
	if err == nil {
		return data, contentType, nil
	}
	proxyURL := buildTwitterMediaProxyURL(sourceURL)
	if proxyURL == "" {
		return nil, "", err
	}
	data, contentType, proxyErr := s.fetchImageOnce(ctx, proxyURL)
	if proxyErr != nil {
		return nil, "", err
	}
	return data, contentType, nil
}

func (s *EnvPromptCatalogImageStorage) fetchImageOnce(ctx context.Context, sourceURL string) ([]byte, string, error) {
	if err := validatePromptCatalogImageSourceURL(sourceURL); err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", promptCatalogImageSyncUserAgent)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.Request != nil && resp.Request.URL != nil {
		if err := validatePromptCatalogImageSourceURL(resp.Request.URL.String()); err != nil {
			return nil, "", err
		}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 30*1024*1024))
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}
	contentType := detectImageMime(body)
	if contentType == "" {
		return nil, "", fmt.Errorf("unsupported image format")
	}
	return body, contentType, nil
}

func validatePromptCatalogImageSourceURL(value string) error {
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil || u == nil || u.Hostname() == "" {
		return fmt.Errorf("invalid image url")
	}
	if u.Scheme != "https" {
		return fmt.Errorf("unsupported image url scheme")
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	switch host {
	case "pbs.twimg.com":
		return nil
	case "images.weserv.nl":
		proxied := strings.TrimSpace(u.Query().Get("url"))
		proxiedURL, err := url.Parse(proxied)
		if err != nil || proxiedURL == nil || proxiedURL.Hostname() == "" {
			return fmt.Errorf("invalid proxied image url")
		}
		if proxiedURL.Scheme != "https" {
			return fmt.Errorf("unsupported proxied image url scheme")
		}
		if strings.ToLower(strings.TrimSuffix(proxiedURL.Hostname(), ".")) != "pbs.twimg.com" {
			return fmt.Errorf("unsupported proxied image url host: %s", proxiedURL.Hostname())
		}
		return nil
	default:
		return fmt.Errorf("unsupported image url host: %s", host)
	}
}

type promptCatalogR2Config struct {
	endpoint        string
	region          string
	bucket          string
	accessKeyID     string
	secretAccessKey string
	publicBaseURL   string
	uploadPath      string
}

func loadPromptCatalogR2Config() promptCatalogR2Config {
	accountID := firstNonEmpty(os.Getenv("R2_ACCOUNT_ID"), os.Getenv("CLOUDFLARE_R2_ACCOUNT_ID"))
	endpoint := strings.TrimSpace(os.Getenv("R2_ENDPOINT"))
	if endpoint == "" && accountID != "" {
		endpoint = "https://" + accountID + ".r2.cloudflarestorage.com"
	}
	return promptCatalogR2Config{
		endpoint:        endpoint,
		region:          firstNonEmpty(os.Getenv("R2_REGION"), "auto"),
		bucket:          firstNonEmpty(os.Getenv("R2_BUCKET_NAME"), os.Getenv("R2_BUCKET")),
		accessKeyID:     firstNonEmpty(os.Getenv("R2_ACCESS_KEY"), os.Getenv("R2_ACCESS_KEY_ID")),
		secretAccessKey: firstNonEmpty(os.Getenv("R2_SECRET_KEY"), os.Getenv("R2_SECRET_ACCESS_KEY")),
		publicBaseURL:   strings.TrimRight(firstNonEmpty(os.Getenv("R2_PUBLIC_BASE_URL"), os.Getenv("R2_DOMAIN")), "/"),
		uploadPath:      firstNonEmpty(os.Getenv("R2_UPLOAD_PATH"), "uploads"),
	}
}

func (c promptCatalogR2Config) configured() bool {
	return c.bucket != "" && c.accessKeyID != "" && c.secretAccessKey != "" && c.publicBaseURL != ""
}

func (c promptCatalogR2Config) publicURL(key string) string {
	base := c.publicBaseURL
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(key, "/")
}

func newPromptCatalogR2Client(ctx context.Context, cfg promptCatalogR2Config) (*s3.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.accessKeyID, cfg.secretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.endpoint)
		}
		o.APIOptions = append(o.APIOptions, v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	}), nil
}

func putPromptCatalogR2Object(ctx context.Context, client *s3.Client, bucket, key string, body []byte, contentType string) error {
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(bucket),
		Key:                aws.String(key),
		Body:               bytes.NewReader(body),
		ContentType:        aws.String(contentType),
		CacheControl:       aws.String(promptCatalogImageCacheControl),
		ContentDisposition: aws.String("inline"),
	})
	return err
}

type noopPromptCatalogImageStorage struct{}

func (noopPromptCatalogImageStorage) SyncPromptImages(ctx context.Context, item PromptCatalogCase, dropUnsynced bool) (PromptCatalogCase, []string, []string) {
	return item, item.ImageURLs, nil
}

func normalizeTweetPayload(payload map[string]any) *tweetPayload {
	tweet := payload
	if nested := getPayloadMap(payload, "tweet"); len(nested) > 0 {
		tweet = nested
	}
	author := getPayloadMap(tweet, "author")
	text := firstNonEmpty(
		getPayloadString(tweet, "article_text"),
		getPayloadString(tweet, "articleText"),
		getPayloadString(tweet, "text"),
		getPayloadString(tweet, "full_text"),
	)
	return &tweetPayload{
		Text:         cleanTweetText(text),
		AuthorHandle: firstNonEmpty(getPayloadString(author, "screen_name"), getPayloadString(author, "screenName")),
		AuthorName:   getPayloadString(author, "name"),
		ImageURLs:    extractImageURLsFromJSON(tweet),
	}
}

func cleanTweetText(value string) string {
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = regexp.MustCompile(`(?i)\s+pic\.twitter\.com/\S+`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)\s+https?://t\.co/\S+`).ReplaceAllString(value, "")
	return strings.TrimSpace(spacePattern.ReplaceAllString(value, " "))
}

func cleanTweetHTMLText(html string) string {
	text := htmlTagPattern.ReplaceAllString(html, " ")
	return cleanTweetText(htmlUnescape(text))
}

func htmlUnescape(value string) string {
	return strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'").Replace(value)
}

func extractImageURLsFromJSON(value any) []string {
	raw, _ := json.Marshal(value)
	return extractImageURLsFromText(string(raw))
}

func extractImageURLsFromText(value string) []string {
	normalized := strings.NewReplacer(`\u0026`, "&", `\\u0026`, "&", `\/`, "/", `\\\/`, "/").Replace(value)
	matches := pbsImageURLPattern.FindAllString(normalized, -1)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		out = append(out, cleanImageURL(match))
	}
	return uniqStrings(out)
}

func cleanImageURL(value string) string {
	cleaned := strings.TrimSpace(strings.NewReplacer("&amp;", "&", `\u0026`, "&", `\/`, "/").Replace(value))
	u, err := url.Parse(cleaned)
	if err != nil || u.Hostname() != "pbs.twimg.com" {
		return cleaned
	}
	path := u.Path
	extension := ""
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		candidate := strings.ToLower(strings.TrimPrefix(path[idx:], "."))
		if candidate == "jpg" || candidate == "jpeg" || candidate == "png" || candidate == "webp" {
			extension = candidate
			path = path[:idx]
		}
	}
	format := firstNonEmpty(u.Query().Get("format"), extension)
	if format == "jpeg" {
		format = "jpg"
	}
	name := firstNonEmpty(u.Query().Get("name"), "orig")
	next := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: path}
	q := next.Query()
	if format != "" {
		q.Set("format", format)
	}
	q.Set("name", name)
	next.RawQuery = q.Encode()
	return next.String()
}

func extractHTMLMeta(html, property string) string {
	patterns := []string{
		`(?is)<meta[^>]+property=["']` + regexp.QuoteMeta(property) + `["'][^>]+content=["']([^"']*)["']`,
		`(?is)<meta[^>]+content=["']([^"']*)["'][^>]+property=["']` + regexp.QuoteMeta(property) + `["']`,
		`(?is)<meta[^>]+name=["']` + regexp.QuoteMeta(property) + `["'][^>]+content=["']([^"']*)["']`,
		`(?is)<meta[^>]+content=["']([^"']*)["'][^>]+name=["']` + regexp.QuoteMeta(property) + `["']`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			return htmlUnescape(matches[1])
		}
	}
	return ""
}

func getPayloadMap(payload map[string]any, key string) map[string]any {
	if payload == nil {
		return map[string]any{}
	}
	if value, ok := payload[key].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func getPayloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	if value, ok := payload[key].(string); ok {
		return value
	}
	return ""
}

func getPayloadItems(payload map[string]any) []map[string]any {
	raw, ok := payload["items"].([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func getPayloadTweetID(payload map[string]any) string {
	return firstNonEmpty(getPayloadString(payload, "tweet_id"), getPayloadString(payload, "id"), getPayloadString(payload, "rest_id"))
}

func payloadHasContent(payload *tweetPayload) bool {
	return payload != nil && (payload.Text != "" || len(payload.ImageURLs) > 0)
}

func imageURLsOrEmpty(payload *tweetPayload) []string {
	if payload == nil {
		return nil
	}
	return payload.ImageURLs
}

func tweetValueOrEmpty[T any](payload *T, getter func(*T) string) string {
	if payload == nil {
		return ""
	}
	return getter(payload)
}

func uniqStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func titleFromPromptCatalog(prompt string) string {
	lines := strings.Split(prompt, "\n")
	candidate := prompt
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			candidate = line
			break
		}
	}
	candidate = cleanTweetText(candidate)
	if len(candidate) > 80 {
		return strings.TrimSpace(candidate[:77]) + "..."
	}
	return candidate
}

func buildPromptCatalogPreview(prompt string, limit int) string {
	normalized := strings.TrimSpace(spacePattern.ReplaceAllString(prompt, " "))
	if len(normalized) > limit {
		return strings.TrimSpace(normalized[:limit]) + "..."
	}
	return normalized
}

func buildTwitterRawJSON(ref twitterStatusRef, payloads ...*tweetPayload) string {
	raw := map[string]any{
		"tweet_id":       ref.TweetID,
		"normalized_url": ref.NormalizedURL,
	}
	names := []string{"x_auto", "oembed", "fxtwitter", "page"}
	for i, payload := range payloads {
		if payload != nil && i < len(names) {
			raw[names[i]] = payload
		}
	}
	encoded, _ := json.Marshal(raw)
	return string(encoded)
}

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func isAllDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func getPromptCatalogImageSourceURLs(item PromptCatalogCase) []string {
	values := append([]string{}, item.ImageURLs...)
	values = append(values, item.ImageURL)
	return uniqStrings(values)
}

func allStorageURLs(values []string) bool {
	for _, value := range values {
		if !strings.Contains(value, "static.") && !strings.Contains(value, ".r2.") {
			return false
		}
	}
	return len(values) > 0
}

func isPromptCatalogStorageURL(value string, cfg promptCatalogR2Config) bool {
	return cfg.publicBaseURL != "" && strings.HasPrefix(value, strings.TrimRight(cfg.publicBaseURL, "/")+"/")
}

func clearPromptCatalogImages(item *PromptCatalogCase) {
	item.ImageURL = ""
	item.ImageURLs = []string{}
	item.ImageOriginalURL = ""
	item.ImagePreviewURL = ""
	item.ImageThumbURL = ""
}

func buildPromptCatalogImageKey(item PromptCatalogCase, uploadPath string, index int, extension string) string {
	parts := []string{
		sanitizePromptCatalogKeyPart(uploadPath),
		"prompts",
		sanitizePromptCatalogKeyPart(firstNonEmpty(item.SourceProject, "manual")),
		sanitizePromptCatalogKeyPart(item.ID),
		fmt.Sprintf("image-%02d.%s", index+1, extension),
	}
	return strings.Join(parts, "/")
}

func sanitizePromptCatalogKeyPart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "misc"
	}
	value = regexp.MustCompile(`[^a-zA-Z0-9._-]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "misc"
	}
	if len(value) > 120 {
		return value[:120]
	}
	return value
}

func buildTwitterMediaProxyURL(value string) string {
	u, err := url.Parse(value)
	if err != nil || u.Hostname() != "pbs.twimg.com" {
		return ""
	}
	format := firstNonEmpty(u.Query().Get("format"), "jpg")
	if format == "jpeg" {
		format = "jpg"
	}
	sourceURL := u.Scheme + "://" + u.Host + u.Path
	if !regexp.MustCompile(`(?i)\.(jpg|jpeg|png|webp)$`).MatchString(sourceURL) {
		sourceURL += "." + format
	}
	return "https://images.weserv.nl/?url=" + url.QueryEscape(sourceURL) + "&output=jpg"
}

func detectImageMime(data []byte) string {
	if len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
		return "image/jpeg"
	}
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		return "image/png"
	}
	if len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp"
	}
	return ""
}

func imageExtensionFromMime(contentType string) string {
	switch contentType {
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	default:
		return "jpg"
	}
}
