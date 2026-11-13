package store_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/iandesj/blerg/internal/store"
	_ "modernc.org/sqlite"
)

func TestUpsertPreservesOpenedAt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)
	sourceID := mustSource(t, db, ctx)
	const url = "https://example.com/post"
	if err := db.UpsertPost(ctx, store.PostDraft{
		SourceID: sourceID,
		Title:    "Original",
		URL:      url,
	}); err != nil {
		t.Fatal(err)
	}
	opened := time.Date(2026, 11, 13, 5, 53, 0, 0, time.UTC)
	if err := db.MarkOpened(ctx, url, opened); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertPost(ctx, store.PostDraft{
		SourceID: sourceID,
		Title:    "Refreshed",
		URL:      url,
		Summary:  "new summary",
	}); err != nil {
		t.Fatal(err)
	}

	posts, err := db.ListPosts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("posts = %d", len(posts))
	}
	if posts[0].Title != "Refreshed" || posts[0].Summary != "new summary" {
		t.Fatalf("post = %+v", posts[0])
	}
	if posts[0].OpenedAt == nil || !posts[0].OpenedAt.Equal(opened) {
		t.Fatalf("opened_at = %v", posts[0].OpenedAt)
	}
}

func TestOpenMigratesExistingPostsTable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "old.db")
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(`
		CREATE TABLE sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			feed_url TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			published_at TEXT,
			summary TEXT,
			categories TEXT
		);
		INSERT INTO sources (name, feed_url, enabled) VALUES ('Fixture', 'file://feed', 1);
		INSERT INTO posts (source_id, title, url) VALUES (1, 'Old', 'https://example.com/old');
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	opened := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := db.MarkOpened(ctx, "https://example.com/old", opened); err != nil {
		t.Fatal(err)
	}
	posts, err := db.ListPostsWithSources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].OpenedAt == nil || !posts[0].OpenedAt.Equal(opened) {
		t.Fatalf("posts = %+v", posts)
	}
}

func openTestDB(t *testing.T) *store.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "blerg.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func mustSource(t *testing.T, db *store.DB, ctx context.Context) int64 {
	t.Helper()
	id, err := db.UpsertSource(ctx, "Fixture", "file://feed", true)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
