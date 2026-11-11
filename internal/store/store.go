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

type Source struct {
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

func (db *DB) ListSources(ctx context.Context) ([]Source, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, name, feed_url, enabled, last_success_at, COALESCE(last_error, '')
		FROM sources ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Source
	for rows.Next() {
		var row Source
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

type Post struct {
	ID          int64
	SourceID    int64
	Title       string
	URL         string
	PublishedAt *time.Time
	Summary     string
	Categories  string
}

type PostDraft struct {
	SourceID    int64
	Title       string
	URL         string
	PublishedAt *time.Time
	Summary     string
	Categories  string
}

func (db *DB) UpsertPost(ctx context.Context, draft PostDraft) error {
	var pub string
	if draft.PublishedAt != nil {
		pub = draft.PublishedAt.UTC().Format(time.RFC3339)
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
	`, draft.SourceID, draft.Title, draft.URL, pub, draft.Summary, draft.Categories)
	return err
}

func (db *DB) SetSourceRefreshSuccess(ctx context.Context, sourceID int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.sql.ExecContext(ctx, `
		UPDATE sources SET last_success_at = ?, last_error = '' WHERE id = ?
	`, now, sourceID)
	return err
}

func (db *DB) SetSourceLastError(ctx context.Context, sourceID int64, message string) error {
	_, err := db.sql.ExecContext(ctx, `
		UPDATE sources SET last_error = ? WHERE id = ?
	`, message, sourceID)
	return err
}

type PostWithSource struct {
	Post
	SourceName string
}

func (db *DB) ListPostsWithSources(ctx context.Context) ([]PostWithSource, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT p.id, p.source_id, p.title, p.url, p.published_at,
		       COALESCE(p.summary, ''), COALESCE(p.categories, ''), s.name
		FROM posts p
		JOIN sources s ON s.id = p.source_id
		ORDER BY p.id
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []PostWithSource
	for rows.Next() {
		var row PostWithSource
		var pub sql.NullString
		if err := rows.Scan(&row.ID, &row.SourceID, &row.Title, &row.URL, &pub,
			&row.Summary, &row.Categories, &row.SourceName); err != nil {
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

func (db *DB) ListPosts(ctx context.Context) ([]Post, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, source_id, title, url, published_at, COALESCE(summary, ''), COALESCE(categories, '')
		FROM posts ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Post
	for rows.Next() {
		var row Post
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
