package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
	"url_shortener"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func NewDB(path string) (*sql.DB, error) {
	// DSN: busy_timeout=10s — SQLite ждёт до 10s при блокировке, вместо немедленного SQLITE_BUSY
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(10000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// connection pool — read-heavy: 8 конкурентных читателей, WAL + busy_timeout это позволяют
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
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
		log.Fatalf("migration error: %v", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	// 3. Выполняем миграции "Up" до самой свежей версии
	log.Println("run migrations")

	goose.SetBaseFS(url_shortener.EmbeddedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.UpContext(context.Background(), db, "migrations"); err != nil {
		return err
	}

	log.Println("migrations applied successfully")
	return nil
}
