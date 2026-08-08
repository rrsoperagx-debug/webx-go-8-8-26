
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
	"github.com/rs/zerolog/log"
)

type DB struct {
	*sql.DB
}

func Init(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	// modernc.org/sqlite DSN: file:./data/metrics.db?_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=on
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=on&cache=shared", path)

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// Run migrations
	if err := runMigrations(sqlDB); err != nil {
		return nil, err
	}

	log.Info().Str("path", path).Msg("Database initialized")
	return &DB{sqlDB}, nil
}

func runMigrations(db *sql.DB) error {
	// Try to read migration files
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value REAL NOT NULL,
			labels TEXT NOT NULL DEFAULT '{}',
			timestamp TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action TEXT NOT NULL,
			actor TEXT NOT NULL,
			details TEXT,
			created_at TEXT NOT NULL
		);`,
	}

	for _, q := range migrations {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration failed: %w query: %s", err, q)
		}
	}
	return nil
}
