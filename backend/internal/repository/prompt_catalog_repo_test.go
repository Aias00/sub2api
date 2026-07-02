//go:build unit

package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Wei-Shaw/cloudbase/internal/pkg/pagination"
	"github.com/Wei-Shaw/cloudbase/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPromptCatalogOrderBy(t *testing.T) {
	tests := []struct {
		name      string
		sortBy    string
		sortOrder string
		expected  string
	}{
		{"title asc", "title", "asc", "title ASC, imported_at DESC NULLS LAST, id ASC"},
		{"title desc", "title", "desc", "title DESC, imported_at DESC NULLS LAST, id ASC"},
		{"created_at asc", "created_at", "asc", "created_at ASC NULLS LAST, title ASC, id ASC"},
		{"updated_at desc", "updated_at", "desc", "updated_at DESC NULLS LAST, title ASC, id ASC"},
		{"default falls to imported_at", "", "", "imported_at DESC NULLS LAST, title ASC, id ASC"},
		{"invalid sort_by falls to imported_at", "raw_sql", "asc", "imported_at ASC NULLS LAST, title ASC, id ASC"},
		{"whitespace-padded sort_by", " title ", " asc ", "title ASC, imported_at DESC NULLS LAST, id ASC"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := pagination.PaginationParams{SortBy: tc.sortBy, SortOrder: tc.sortOrder}
			got := promptCatalogOrderBy(params)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestDecodePromptCatalogStringArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []string
	}{
		{"nil input", nil, []string{}},
		{"empty input", []byte{}, []string{}},
		{"empty JSON array", []byte("[]"), []string{}},
		{"single element", []byte(`["portrait"]`), []string{"portrait"}},
		{"multiple elements", []byte(`["a","b","c"]`), []string{"a", "b", "c"}},
		{"invalid JSON", []byte("not json"), []string{}},
		{"JSON null", []byte("null"), []string{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := decodePromptCatalogStringArray(tc.input)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestEncodePromptCatalogStringArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"nil input", nil, "[]"},
		{"empty slice", []string{}, "[]"},
		{"single element", []string{"portrait"}, `["portrait"]`},
		{"multiple elements", []string{"a", "b"}, `["a","b"]`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := encodePromptCatalogStringArray(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestScanPromptCatalogCase(t *testing.T) {
	now := time.Now()

	t.Run("full row scan", func(t *testing.T) {
		row := &mockScanner{values: []any{
			"case-1",                             // id
			"Title",                              // title
			"Prompt body",                        // prompt
			"Preview",                            // prompt_preview
			"portrait",                           // category
			[]byte(`["tag1","tag2"]`),            // tags
			[]byte(`["openai-image"]`),           // model_tags
			"https://x.com/123",                  // source_url
			"https://img.example.com/1.jpg",      // image_url
			[]byte(`["url1","url2"]`),            // image_urls
			"https://img.example.com/orig",       // image_original_url
			"https://img.example.com/prev",       // image_preview_url
			"https://img.example.com/thumb",      // image_thumb_url
			"twitter",                            // source_project
			"case",                               // source_type
			"X",                                  // source_label
			"https://github.com/repo",            // github_url
			true,                                 // featured
			[]byte(`["style1"]`),                 // styles
			[]byte(`["scene1"]`),                 // scenes
			"twitter",                            // import_source
			`{"key":"value"}`,                    // raw_json (COALESCE'd to '{}')
			"published",                          // status
			sql.NullTime{Time: now, Valid: true}, // imported_at
			now,                                  // created_at
			now,                                  // updated_at
		}}

		item, err := scanPromptCatalogCase(row)
		require.NoError(t, err)
		require.Equal(t, "case-1", item.ID)
		require.Equal(t, "Title", item.Title)
		require.Equal(t, []string{"tag1", "tag2"}, item.Tags)
		require.Equal(t, []string{"openai-image"}, item.ModelTags)
		require.Equal(t, []string{"url1", "url2"}, item.ImageURLs)
		require.Equal(t, []string{"style1"}, item.Styles)
		require.Equal(t, []string{"scene1"}, item.Scenes)
		require.True(t, item.Featured)
		require.NotNil(t, item.ImportedAt)
		require.Equal(t, now, *item.ImportedAt)
	})

	t.Run("null imported_at", func(t *testing.T) {
		row := &mockScanner{values: []any{
			"case-2",                   // id
			"Title",                    // title
			"Prompt body",              // prompt
			"Preview",                  // prompt_preview
			"general",                  // category
			[]byte("[]"),               // tags
			[]byte("[]"),               // model_tags
			"",                         // source_url
			"",                         // image_url
			[]byte("[]"),               // image_urls
			"",                         // image_original_url
			"",                         // image_preview_url
			"",                         // image_thumb_url
			"manual",                   // source_project
			"case",                     // source_type
			"",                         // source_label
			"",                         // github_url
			false,                      // featured
			[]byte("[]"),               // styles
			[]byte("[]"),               // scenes
			"catalog",                  // import_source
			"{}",                       // raw_json
			"published",                // status
			sql.NullTime{Valid: false}, // imported_at (NULL)
			now,                        // created_at
			now,                        // updated_at
		}}

		item, err := scanPromptCatalogCase(row)
		require.NoError(t, err)
		require.Nil(t, item.ImportedAt)
	})

	t.Run("scan error propagates", func(t *testing.T) {
		row := &mockScanner{err: sql.ErrNoRows}
		_, err := scanPromptCatalogCase(row)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestPromptCatalogFacetCounts(t *testing.T) {
	t.Run("sorts by count descending then value ascending", func(t *testing.T) {
		counts := map[string]int64{
			"portrait":  5,
			"landscape": 10,
			"abstract":  5,
		}
		labels := map[string]string{
			"portrait":  "Portrait",
			"landscape": "Landscape",
		}

		facets := promptCatalogFacetCounts(counts, labels, true)

		require.Equal(t, "landscape", facets[0].Value)
		require.Equal(t, int64(10), facets[0].Count)
		require.Equal(t, "Landscape", facets[0].Label)
		// portrait and abstract both have count 5, sorted by value
		require.Equal(t, "abstract", facets[1].Value)
		require.Equal(t, "portrait", facets[2].Value)
	})

	t.Run("sorts by value only when countFirst is false", func(t *testing.T) {
		counts := map[string]int64{
			"b": 10,
			"a": 5,
		}
		labels := map[string]string{}

		facets := promptCatalogFacetCounts(counts, labels, false)
		require.Equal(t, "a", facets[0].Value)
		require.Equal(t, "b", facets[1].Value)
	})
}

type mockScanner struct {
	values []any
	err    error
	idx    int
}

func (s *mockScanner) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	for i, d := range dest {
		if s.idx+i >= len(s.values) {
			break
		}
		src := s.values[s.idx+i]
		switch dst := d.(type) {
		case *string:
			if v, ok := src.(string); ok {
				*dst = v
			}
		case *bool:
			if v, ok := src.(bool); ok {
				*dst = v
			}
		case *[]byte:
			if v, ok := src.([]byte); ok {
				*dst = v
			}
		case *time.Time:
			if v, ok := src.(time.Time); ok {
				*dst = v
			}
		case *sql.NullTime:
			if v, ok := src.(sql.NullTime); ok {
				*dst = v
			}
		}
	}
	s.idx += len(dest)
	return nil
}

// Verify the mock scanner satisfies the interface at compile time.
var _ promptCatalogScanner = (*mockScanner)(nil)

// Verify service.PromptCatalogCase has the fields we test.
var _ service.PromptCatalogCase = service.PromptCatalogCase{}
