package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/Aias00/cloudbase/internal/config"
	"github.com/Aias00/cloudbase/internal/util/responseheaders"
	"github.com/Aias00/cloudbase/internal/util/urlvalidator"
)

const geminiStickySessionTTL = time.Hour

const (
	geminiMaxRetries     = 5
	geminiRetryBaseDelay = 1 * time.Second
	geminiRetryMaxDelay  = 16 * time.Second
)

// Gemini tool calling now requires `thoughtSignature` in parts that include `functionCall`.
// Many clients don't send it; we inject a known dummy signature to satisfy the validator.
// Ref: https://ai.google.dev/gemini-api/docs/thought-signatures
const geminiDummyThoughtSignature = "skip_thought_signature_validator"

type GeminiMessagesCompatService struct {
	accountRepo               AccountRepository
	groupRepo                 GroupRepository
	cache                     GatewayCache
	schedulerSnapshot         *SchedulerSnapshotService
	tokenProvider             *GeminiTokenProvider
	rateLimitService          *RateLimitService
	httpUpstream              HTTPUpstream
	antigravityGatewayService *AntigravityGatewayService
	cfg                       *config.Config
	responseHeaderFilter      *responseheaders.CompiledHeaderFilter
}

func (s *GeminiMessagesCompatService) readUpstreamErrorBody(resp *http.Response) []byte {
	if resp == nil || resp.Body == nil {
		return nil
	}
	limit := gatewayUpstreamErrorBodyReadLimit
	if s != nil && s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody && s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes > int(limit) {
		limit = int64(s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, limit))
	return body
}

func NewGeminiMessagesCompatService(
	accountRepo AccountRepository,
	groupRepo GroupRepository,
	cache GatewayCache,
	schedulerSnapshot *SchedulerSnapshotService,
	tokenProvider *GeminiTokenProvider,
	rateLimitService *RateLimitService,
	httpUpstream HTTPUpstream,
	antigravityGatewayService *AntigravityGatewayService,
	cfg *config.Config,
) *GeminiMessagesCompatService {
	return &GeminiMessagesCompatService{
		accountRepo:               accountRepo,
		groupRepo:                 groupRepo,
		cache:                     cache,
		schedulerSnapshot:         schedulerSnapshot,
		tokenProvider:             tokenProvider,
		rateLimitService:          rateLimitService,
		httpUpstream:              httpUpstream,
		antigravityGatewayService: antigravityGatewayService,
		cfg:                       cfg,
		responseHeaderFilter:      compileResponseHeaderFilter(cfg),
	}
}

// GetTokenProvider returns the token provider for OAuth accounts
func (s *GeminiMessagesCompatService) GetTokenProvider() *GeminiTokenProvider {
	return s.tokenProvider
}

// GetAntigravityGatewayService 返回 AntigravityGatewayService
func (s *GeminiMessagesCompatService) GetAntigravityGatewayService() *AntigravityGatewayService {
	return s.antigravityGatewayService
}

func (s *GeminiMessagesCompatService) validateUpstreamBaseURL(raw string) (string, error) {
	if s.cfg != nil && !s.cfg.Security.URLAllowlist.Enabled {
		normalized, err := urlvalidator.ValidateURLFormat(raw, s.cfg.Security.URLAllowlist.AllowInsecureHTTP)
		if err != nil {
			return "", fmt.Errorf("invalid base_url: %w", err)
		}
		return normalized, nil
	}
	normalized, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     s.cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}
	return normalized, nil
}

var (
	sensitiveQueryParamRegex = regexp.MustCompile(`(?i)([?&](?:key|client_secret|access_token|refresh_token)=)[^&"\s]+`)
	retryInRegex             = regexp.MustCompile(`Please retry in ([0-9.]+)s`)
)

type claudeErrorMapping struct {
	Type       string
	Message    string
	StatusCode int
}

type geminiStreamResult struct {
	usage        *ClaudeUsage
	firstTokenMs *int
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type geminiNativeStreamResult struct {
	usage        *ClaudeUsage
	firstTokenMs *int
}

type UpstreamHTTPResult struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func asInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case int64:
		return int(t), true
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	default:
		return 0, false
	}
}
