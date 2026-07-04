package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
	"url_shortener"

	_ "github.com/mattn/go-sqlite3"

	"github.com/pressly/goose/v3"
)

func NewDB(path string) (*sql.DB, error) {
	// DSN: busy_timeout=10s — SQLite ждёт до 10s при блокировке, вместо немедленного SQLITE_BUSY
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(10000)", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// connection pool — read-heavy: 32 конкурентных читателей, WAL + busy_timeout это позволяют
	db.SetMaxOpenConns(32)
	db.SetMaxIdleConns(32)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(15 * time.Minute)

	// pragma
	sql := `
	PRAGMA busy_timeout=10000;
	PRAGMA foreign_keys=ON;
	PRAGMA journal_mode=WAL;
	PRAGMA synchronous = NORMAL;
	PRAGMA auto_vacuum = INCREMENTAL;
	PRAGMA journal_size_limit = 67110000;
	PRAGMA temp_store = MEMORY;
	PRAGMA cache_size = -65536;
	PRAGMA page_size = 4096;
	`
	if _, err := db.Exec(sql); err != nil {
		return nil, fmt.Errorf("pragma error: %w", err)
	}

	// goose up
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migration: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	// 3. Выполняем миграции "Up" до самой свежей версии
	slog.Info("run migrations")

	goose.SetBaseFS(url_shortener.EmbeddedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return err
	}

	slog.Info("migrations applied successfully")
	return nil
}
