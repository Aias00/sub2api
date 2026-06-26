package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type httpClientStub struct {
	responses map[string]string
	statuses  map[string]int
}

func (s httpClientStub) Do(req *http.Request) (*http.Response, error) {
	status := s.statuses[req.URL.String()]
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(s.responses[req.URL.String()])),
		Header:     make(http.Header),
	}, nil
}

type promptCatalogImageStorageStub struct {
	item         PromptCatalogCase
	uploadedURLs []string
	warnings     []string
}

func (s promptCatalogImageStorageStub) SyncPromptImages(ctx context.Context, item PromptCatalogCase, dropUnsynced bool) (PromptCatalogCase, []string, []string) {
	if s.item.ID != "" {
		return s.item, s.uploadedURLs, s.warnings
	}
	return item, item.ImageURLs, s.warnings
}

func TestTwitterImportServiceImportUsesXAutoAndUpsertsStaticImages(t *testing.T) {
	t.Setenv("X_AUTO_BASE_URL", "https://x-auto.test")
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	repo := &promptCatalogRepoStub{}
	promptSvc := NewPromptCatalogService(repo)
	staticItem := PromptCatalogCase{
		ID:               "tw-2062285104073109762",
		Title:            "Imported",
		Prompt:           "A GPT Image 2 prompt",
		PromptPreview:    "A GPT Image 2 prompt",
		Category:         "Twitter Imports",
		SourceURL:        "https://x.com/RealHanyaHu/status/2062285104073109762",
		ImageURL:         "https://static.example/prompts/twitter/image-01.jpg",
		ImageURLs:        []string{"https://static.example/prompts/twitter/image-01.jpg"},
		ImageOriginalURL: "https://static.example/prompts/twitter/image-01.jpg",
		ModelTags:        []string{"OpenAI Image"},
		SourceProject:    "twitter",
		SourceType:       "case",
		SourceLabel:      "@RealHanyaHu",
		ImportSource:     "twitter",
		Status:           PromptCatalogStatusPublished,
	}
	importSvc := NewTwitterImportServiceForTest(
		promptSvc,
		httpClientStub{responses: map[string]string{
			"https://x-auto.test/twitter/articles/2062285104073109762": `{
				"tweet_id":"2062285104073109762",
				"text":"A GPT Image 2 prompt https://t.co/abc",
				"author":{"screen_name":"RealHanyaHu","name":"Hanya"},
				"raw":"https://pbs.twimg.com/media/HJ87P5EboAAz9Ir?format=jpg&name=orig"
			}`,
			"https://publish.twitter.com/oembed?dnt=1&omit_script=1&url=https%3A%2F%2Fx.com%2FRealHanyaHu%2Fstatus%2F2062285104073109762": `{"html":"<blockquote><p>fallback</p></blockquote>","author_name":"Hanya"}`,
			"https://api.fxtwitter.com/status/2062285104073109762":                                                                        `{"tweet":{"text":"fx fallback","media":{"photos":[{"url":"https://pbs.twimg.com/media/fx?format=jpg&name=orig"}]},"author":{"screen_name":"RealHanyaHu"}}}`,
			"https://x.com/RealHanyaHu/status/2062285104073109762":                                                                        `<html><meta property="og:description" content="page fallback"><meta property="og:image" content="https://pbs.twimg.com/media/page?format=jpg&amp;name=orig"></html>`,
		}},
		promptCatalogImageStorageStub{item: staticItem, uploadedURLs: staticItem.ImageURLs},
		func() time.Time { return now },
	)

	result, err := importSvc.Import(context.Background(), TwitterImportInput{
		URL: "https://x.com/RealHanyaHu/status/2062285104073109762",
	})

	require.NoError(t, err)
	require.Equal(t, "tw-2062285104073109762", result.Item.ID)
	require.Equal(t, staticItem.ImageURLs, result.ImageURLs)
	require.NotNil(t, repo.upsertItem)
	require.Equal(t, now, *repo.upsertItem.ImportedAt)
	require.Equal(t, "twitter", repo.upsertItem.ImportSource)
	require.Equal(t, []string{"OpenAI Image"}, repo.upsertItem.ModelTags)
	require.Equal(t, staticItem.ImageURL, repo.upsertItem.ImageURL)
}

func TestTwitterImportServiceRejectsInvalidURL(t *testing.T) {
	svc := NewTwitterImportServiceForTest(NewPromptCatalogService(&promptCatalogRepoStub{}), nil, nil, nil)

	_, err := svc.Import(context.Background(), TwitterImportInput{URL: "https://example.com/nope"})

	require.ErrorIs(t, err, ErrPromptCatalogInvalidInput)
}

func TestValidatePromptCatalogImageSourceURL(t *testing.T) {
	require.NoError(t, validatePromptCatalogImageSourceURL("https://pbs.twimg.com/media/test?format=jpg&name=orig"))
	require.NoError(t, validatePromptCatalogImageSourceURL("https://images.weserv.nl/?url=https%3A%2F%2Fpbs.twimg.com%2Fmedia%2Ftest.jpg&output=jpg"))

	for _, value := range []string{
		"file:///etc/passwd",
		"http://127.0.0.1/image.jpg",
		"http://pbs.twimg.com/media/image.jpg",
		"https://example.com/image.jpg",
		"https://pbs.twimg.com.evil.test/media/image.jpg",
		"https://images.weserv.nl/?url=http%3A%2F%2Fpbs.twimg.com%2Fmedia%2Fimage.jpg&output=jpg",
		"https://images.weserv.nl/?url=http%3A%2F%2F127.0.0.1%2Fimage.jpg&output=jpg",
	} {
		require.Error(t, validatePromptCatalogImageSourceURL(value), value)
	}
}

func TestPromptCatalogImageHTTPClientBlocksUnsafeRedirects(t *testing.T) {
	client := newPromptCatalogImageHTTPClient()
	require.NotNil(t, client.CheckRedirect)
	require.Equal(t, 20*time.Second, client.Timeout)

	allowedReq, err := http.NewRequest(http.MethodGet, "https://pbs.twimg.com/media/redirected.jpg", nil)
	require.NoError(t, err)
	require.NoError(t, client.CheckRedirect(allowedReq, nil))

	blockedReq, err := http.NewRequest(http.MethodGet, "https://example.com/image.jpg", nil)
	require.NoError(t, err)
	require.Error(t, client.CheckRedirect(blockedReq, nil))
}
