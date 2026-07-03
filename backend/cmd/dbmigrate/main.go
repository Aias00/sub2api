package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Aias00/cloudbase/internal/repository"
	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := getenv("DATABASE_HOST", "127.0.0.1")
		port := getenv("DATABASE_PORT", "5432")
		user := getenv("DATABASE_USER", getenv("PGUSER", "cloudbase"))
		password := getenv("DATABASE_PASSWORD", "")
		dbname := getenv("DATABASE_DBNAME", getenv("PGDATABASE", "cloudbase"))
		sslmode := getenv("DATABASE_SSLMODE", "disable")
		if _, err := strconv.Atoi(port); err != nil {
			log.Fatalf("invalid DATABASE_PORT: %v", err)
		}
		if password == "" {
			dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", host, port, user, dbname, sslmode)
		} else {
			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
		}
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := repository.ApplyMigrations(ctx, db); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Println("database migrations applied")
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
