package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, err
	}

	return database, nil
}

func Init(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
		);`,
		`CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			target_url TEXT NOT NULL,
			remark TEXT NOT NULL DEFAULT '',
			tags_json TEXT NOT NULL DEFAULT '[]',
			enabled INTEGER NOT NULL DEFAULT 1,
			click_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
			updated_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP)
		);`,
		`CREATE TABLE IF NOT EXISTS link_visits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link_id INTEGER NOT NULL,
			ip TEXT NOT NULL DEFAULT '',
			referer TEXT NOT NULL DEFAULT '',
			referer_host TEXT NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT '',
			client_name TEXT NOT NULL DEFAULT '',
			client_type TEXT NOT NULL DEFAULT '',
			device_type TEXT NOT NULL DEFAULT '',
			os TEXT NOT NULL DEFAULT '',
			visited_at DATETIME NOT NULL DEFAULT (CURRENT_TIMESTAMP),
			FOREIGN KEY (link_id) REFERENCES links(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_links_code ON links(code);`,
		`CREATE INDEX IF NOT EXISTS idx_link_visits_link_visited_at ON link_visits(link_id, visited_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_link_visits_link_referer_host ON link_visits(link_id, referer_host);`,
		`CREATE INDEX IF NOT EXISTS idx_link_visits_link_client_name ON link_visits(link_id, client_name);`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	if err := ensureColumn(ctx, db, "links", "remark", `ALTER TABLE links ADD COLUMN remark TEXT NOT NULL DEFAULT ''`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "links", "tags_json", `ALTER TABLE links ADD COLUMN tags_json TEXT NOT NULL DEFAULT '[]'`); err != nil {
		return err
	}

	return nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func ensureColumn(ctx context.Context, db *sql.DB, table string, column string, alterSQL string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &primaryKey); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, alterSQL)
	return err
}
