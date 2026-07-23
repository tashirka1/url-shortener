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
	// DSN-параметры применяются к КАЖДОМУ новому соединению в пуле (в отличие от db.Exec,
	// который действует только на одном соединении)
	dsn := fmt.Sprintf(
		"file:%s"+
			"?_pragma=busy_timeout(10000)"+
			"&_pragma=foreign_keys(ON)"+
			"&_pragma=journal_mode(WAL)"+
			"&_pragma=synchronous(NORMAL)"+
			"&_pragma=temp_store(MEMORY)"+
			"&_pragma=cache_size(-65536)"+
			"&_pragma=auto_vacuum(INCREMENTAL)"+
			"&_pragma=journal_size_limit(67110000)"+
			"&_pragma=page_size(4096)",
		path,
	)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// connection pool — read-heavy: 64 коннектов для 8 ядер, WAL + mmap снимают блокировки
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(64)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(15 * time.Minute)

	// goose up
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migration: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
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
