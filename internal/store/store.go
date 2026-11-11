package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	sql *sql.DB
}

func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error {
	return db.sql.Close()
}

func (db *DB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			feed_url TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			etag TEXT,
			last_modified TEXT,
			last_success_at TEXT,
			last_error TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			published_at TEXT,
			summary TEXT,
			categories TEXT
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.sql.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

type SourceRow struct {
	ID            int64
	Name          string
	FeedURL       string
	Enabled       bool
	LastSuccessAt *time.Time
	LastError     string
}

func (db *DB) UpsertSource(ctx context.Context, name, feedURL string, enabled bool) (int64, error) {
	en := 0
	if enabled {
		en = 1
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO sources (name, feed_url, enabled)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			feed_url = excluded.feed_url,
			enabled = excluded.enabled
	`, name, feedURL, en)
	if err != nil {
		return 0, err
	}
	var id int64
	err = db.sql.QueryRowContext(ctx, `SELECT id FROM sources WHERE name = ?`, name).Scan(&id)
	return id, err
}

func (db *DB) ListSources(ctx context.Context) ([]SourceRow, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, name, feed_url, enabled, last_success_at, COALESCE(last_error, '')
		FROM sources ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SourceRow
	for rows.Next() {
		var row SourceRow
		var en int
		var lastSuccess sql.NullString
		if err := rows.Scan(&row.ID, &row.Name, &row.FeedURL, &en, &lastSuccess, &row.LastError); err != nil {
			return nil, err
		}
		row.Enabled = en == 1
		if lastSuccess.Valid && lastSuccess.String != "" {
			if t, err := time.Parse(time.RFC3339, lastSuccess.String); err == nil {
				row.LastSuccessAt = &t
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

type PostRow struct {
	ID          int64
	SourceID    int64
	Title       string
	URL         string
	PublishedAt *time.Time
	Summary     string
	Categories  string
}

func (db *DB) UpsertPost(ctx context.Context, sourceID int64, title, url string, publishedAt *time.Time, summary, categories string) error {
	var pub string
	if publishedAt != nil {
		pub = publishedAt.UTC().Format(time.RFC3339)
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO posts (source_id, title, url, published_at, summary, categories)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			source_id = excluded.source_id,
			title = excluded.title,
			published_at = excluded.published_at,
			summary = excluded.summary,
			categories = excluded.categories
	`, sourceID, title, url, pub, summary, categories)
	return err
}

func (db *DB) ListPosts(ctx context.Context) ([]PostRow, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, source_id, title, url, published_at, COALESCE(summary, ''), COALESCE(categories, '')
		FROM posts ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PostRow
	for rows.Next() {
		var row PostRow
		var pub sql.NullString
		if err := rows.Scan(&row.ID, &row.SourceID, &row.Title, &row.URL, &pub, &row.Summary, &row.Categories); err != nil {
			return nil, err
		}
		if pub.Valid && pub.String != "" {
			if t, err := time.Parse(time.RFC3339, pub.String); err == nil {
				row.PublishedAt = &t
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
