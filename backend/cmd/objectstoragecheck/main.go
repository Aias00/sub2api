package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	sample := flag.Bool("sample", false, "Sample public object URLs with HTTP HEAD")
	limit := flag.Int("limit", 10, "Maximum URLs to sample")
	timeout := flag.Duration("timeout", 2*time.Minute, "Database operation timeout")
	flag.Parse()

	db, err := sql.Open("postgres", postgresDSN())
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("connect postgres: %v", err)
	}

	inventory, err := loadInventory(ctx, db)
	if err != nil {
		log.Fatalf("load object storage inventory: %v", err)
	}
	printInventory(inventory)

	if *sample {
		urls, err := sampleURLs(ctx, db, inventory, *limit)
		if err != nil {
			log.Fatalf("load url sample: %v", err)
		}
		if len(urls) == 0 {
			fmt.Println("No public HTTP(S) URLs found for sampling.")
			return
		}
		client := &http.Client{Timeout: 10 * time.Second}
		for _, url := range urls {
			req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
			if err != nil {
				fmt.Printf("HEAD %s ... failed: %v\n", url, err)
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("HEAD %s ... failed: %v\n", url, err)
				continue
			}
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				fmt.Printf("HEAD %s ... ok\n", url)
			} else {
				fmt.Printf("HEAD %s ... failed: status=%d\n", url, resp.StatusCode)
			}
		}
	}
}

type inventory struct {
	HasPromptCatalogItems      bool
	HasImageWorkspaceArtifacts bool
	HasHotItemMedia            bool

	PromptVisible            int64
	PromptWithImage          int64
	ImageArtifactURLs        int64
	ImageArtifactStorageKeys int64
	HotItemMediaRows         int64
}

func loadInventory(ctx context.Context, db *sql.DB) (inventory, error) {
	var inv inventory
	var err error
	if inv.HasPromptCatalogItems, err = tableExists(ctx, db, "prompt_catalog_items"); err != nil {
		return inv, err
	}
	if inv.HasImageWorkspaceArtifacts, err = tableExists(ctx, db, "image_workspace_artifacts"); err != nil {
		return inv, err
	}
	if inv.HasHotItemMedia, err = tableExists(ctx, db, "hot_item_media"); err != nil {
		return inv, err
	}
	if inv.HasPromptCatalogItems {
		if err := db.QueryRowContext(ctx, `
WITH prompt_visible AS (
  SELECT *
  FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
), prompt_flags AS (
  SELECT
    id,
    NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
      OR jsonb_array_length(COALESCE(image_urls, '[]'::jsonb)) > 0 AS has_image
  FROM prompt_visible
)
SELECT
  (SELECT COUNT(*) FROM prompt_visible) AS prompt_visible,
  (SELECT COUNT(*) FROM prompt_flags WHERE has_image) AS prompt_with_image
`).Scan(&inv.PromptVisible, &inv.PromptWithImage); err != nil {
			return inv, fmt.Errorf("count prompt catalog objects: %w", err)
		}
	}
	if inv.HasImageWorkspaceArtifacts {
		if err := db.QueryRowContext(ctx, `
SELECT
  COUNT(*) FILTER (WHERE NULLIF(TRIM(image_url), '') IS NOT NULL) AS image_artifact_urls,
  COUNT(*) FILTER (WHERE NULLIF(TRIM(storage_key), '') IS NOT NULL) AS image_artifact_storage_keys
FROM image_workspace_artifacts
`).Scan(&inv.ImageArtifactURLs, &inv.ImageArtifactStorageKeys); err != nil {
			return inv, fmt.Errorf("count image workspace objects: %w", err)
		}
	}
	if inv.HasHotItemMedia {
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM hot_item_media`).Scan(&inv.HotItemMediaRows); err != nil {
			return inv, fmt.Errorf("count hot item media: %w", err)
		}
	}
	return inv, nil
}

func printInventory(inv inventory) {
	fmt.Println("Object storage PostgreSQL inventory")
	fmt.Printf("has_prompt_catalog_items=%t\n", inv.HasPromptCatalogItems)
	fmt.Printf("has_image_workspace_artifacts=%t\n", inv.HasImageWorkspaceArtifacts)
	fmt.Printf("has_hot_item_media=%t\n", inv.HasHotItemMedia)
	fmt.Printf("prompt_visible=%d\n", inv.PromptVisible)
	fmt.Printf("prompt_with_image=%d\n", inv.PromptWithImage)
	fmt.Printf("image_artifact_urls=%d\n", inv.ImageArtifactURLs)
	fmt.Printf("image_artifact_storage_keys=%d\n", inv.ImageArtifactStorageKeys)
	fmt.Printf("hot_item_media_rows=%d\n", inv.HotItemMediaRows)
}

func sampleURLs(ctx context.Context, db *sql.DB, inv inventory, limit int) ([]string, error) {
	if limit <= 0 {
		return nil, nil
	}
	urls := make([]string, 0, limit)
	addQuery := func(query string) error {
		if len(urls) >= limit {
			return nil
		}
		rows, err := db.QueryContext(ctx, query, limit-len(urls))
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var url string
			if err := rows.Scan(&url); err != nil {
				return err
			}
			url = strings.TrimSpace(url)
			if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
				urls = append(urls, url)
			}
		}
		return rows.Err()
	}
	if inv.HasPromptCatalogItems {
		if err := addQuery(`
WITH urls AS (
  SELECT image_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
  UNION
  SELECT image_original_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
  UNION
  SELECT image_preview_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
  UNION
  SELECT image_thumb_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
  UNION
  SELECT jsonb_array_elements_text(COALESCE(image_urls, '[]'::jsonb)) AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
)
SELECT url FROM urls
WHERE url LIKE 'http://%' OR url LIKE 'https://%'
ORDER BY url
LIMIT $1
`); err != nil {
			return nil, fmt.Errorf("sample prompt catalog urls: %w", err)
		}
	}
	if inv.HasImageWorkspaceArtifacts {
		if err := addQuery(`
SELECT image_url AS url FROM image_workspace_artifacts
WHERE NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
  AND (image_url LIKE 'http://%' OR image_url LIKE 'https://%')
ORDER BY image_url
LIMIT $1
`); err != nil {
			return nil, fmt.Errorf("sample image workspace urls: %w", err)
		}
	}
	if inv.HasHotItemMedia {
		if err := addQuery(`
SELECT original_url AS url FROM hot_item_media
WHERE NULLIF(TRIM(COALESCE(original_url, '')), '') IS NOT NULL
  AND (original_url LIKE 'http://%' OR original_url LIKE 'https://%')
ORDER BY original_url
LIMIT $1
`); err != nil {
			return nil, fmt.Errorf("sample hot item media urls: %w", err)
		}
	}
	return urls, nil
}

func postgresDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn
	}
	host := envOrDefault("PGHOST", "127.0.0.1")
	port := envOrDefault("PGPORT", "5432")
	user := envOrDefault("PGUSER", "sub2api")
	dbname := envOrDefault("PGDATABASE", "sub2api")
	sslmode := envOrDefault("PGSSLMODE", "disable")
	password := os.Getenv("PGPASSWORD")
	parts := []string{
		"host=" + quoteConnValue(host),
		"port=" + quoteConnValue(port),
		"user=" + quoteConnValue(user),
		"dbname=" + quoteConnValue(dbname),
		"sslmode=" + quoteConnValue(sslmode),
	}
	if password != "" {
		parts = append(parts, "password="+quoteConnValue(password))
	}
	return strings.Join(parts, " ")
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func quoteConnValue(value string) string {
	if value == "" {
		return "''"
	}
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	if strings.ContainsAny(escaped, " \t\n'\\") {
		return "'" + escaped + "'"
	}
	return escaped
}

func tableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&exists); err != nil {
		return false, fmt.Errorf("check table %s: %w", table, err)
	}
	return exists, nil
}
