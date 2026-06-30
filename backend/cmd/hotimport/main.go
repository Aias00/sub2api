package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var hotTables = []string{
	"hot_sources",
	"hot_runs",
	"hot_checkpoints",
	"hot_items",
	"hot_item_media",
	"hot_media_assets",
	"hot_run_events",
	"hot_feed_items",
	"hot_daily_issues",
	"hot_daily_sections",
	"hot_daily_stories",
	"hot_mp_entries",
}

func main() {
	checkOnly := flag.Bool("check-only", false, "Deprecated compatibility flag; the checker is always read-only")
	timeout := flag.Duration("timeout", 5*time.Minute, "Database operation timeout")
	flag.Parse()
	_ = *checkOnly

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

	if err := printHotCounts(ctx, db); err != nil {
		log.Fatalf("verify hot tables: %v", err)
	}
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

func printHotCounts(ctx context.Context, db *sql.DB) error {
	fmt.Println("Hot PostgreSQL table inventory")
	for _, table := range hotTables {
		exists, err := tableExists(ctx, db, table)
		if err != nil {
			return err
		}
		if !exists {
			fmt.Printf("%s=missing\n", table)
			continue
		}
		count, err := tableCount(ctx, db, table)
		if err != nil {
			return err
		}
		fmt.Printf("%s=%d\n", table, count)
	}
	return nil
}

func tableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&exists); err != nil {
		return false, fmt.Errorf("check table %s: %w", table, err)
	}
	return exists, nil
}

func tableCount(ctx context.Context, db *sql.DB, table string) (int64, error) {
	if !isKnownHotTable(table) {
		return 0, fmt.Errorf("unknown hot table %q", table)
	}
	var count int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table).Scan(&count); err != nil {
		return 0, fmt.Errorf("count table %s: %w", table, err)
	}
	return count, nil
}

func isKnownHotTable(table string) bool {
	for _, known := range hotTables {
		if table == known {
			return true
		}
	}
	return false
}
