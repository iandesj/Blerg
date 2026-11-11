package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/iandesj/blerg/internal/feed"
	"github.com/iandesj/blerg/internal/store"
)

type RefreshStats struct {
	SourcesEnabled int
	SourcesFailed  int
	PostsIndexed   int
}

func (a *App) Refresh(ctx context.Context) (RefreshStats, error) {
	sources, err := a.ListSources(ctx)
	if err != nil {
		return RefreshStats{}, err
	}

	var stats RefreshStats
	for _, src := range sources {
		if !src.Enabled {
			continue
		}
		stats.SourcesEnabled++
		indexed, err := a.refreshSource(ctx, src)
		if err != nil {
			stats.SourcesFailed++
			_ = a.db.SetSourceLastError(ctx, src.ID, err.Error())
			continue
		}
		stats.PostsIndexed += indexed
		if err := a.db.SetSourceRefreshSuccess(ctx, src.ID); err != nil {
			return stats, err
		}
	}

	if stats.SourcesEnabled == 0 {
		return stats, fmt.Errorf("refresh: no enabled sources in manifest")
	}
	if stats.SourcesFailed == stats.SourcesEnabled {
		return stats, fmt.Errorf("refresh: all %d enabled source(s) failed", stats.SourcesEnabled)
	}
	return stats, nil
}

func (a *App) refreshSource(ctx context.Context, src Source) (int, error) {
	resp, err := a.FetchFeed(ctx, src.FeedURL)
	if err != nil {
		return 0, fmt.Errorf("fetch %s: %w", src.Name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fetch %s: HTTP %s", src.Name, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", src.Name, err)
	}
	entries, err := feed.Parse(strings.NewReader(string(body)))
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", src.Name, err)
	}

	indexed := 0
	for _, e := range entries {
		draft := store.PostDraft{
			SourceID:    src.ID,
			Title:       e.Title,
			URL:         e.URL,
			PublishedAt: e.PublishedAt,
			Summary:     e.Summary,
			Categories:  e.Categories,
		}
		if err := a.UpsertPost(ctx, draft); err != nil {
			return indexed, fmt.Errorf("index %s: %w", src.Name, err)
		}
		indexed++
	}
	return indexed, nil
}
