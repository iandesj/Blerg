package app

import (
	"context"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/iandesj/blerg/internal/feed"
	"github.com/iandesj/blerg/internal/store"
	"golang.org/x/sync/errgroup"
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

	var enabled []Source
	for _, src := range sources {
		if src.Enabled {
			enabled = append(enabled, src)
		}
	}
	if len(enabled) == 0 {
		return RefreshStats{}, fmt.Errorf("refresh: no enabled sources in manifest")
	}

	var stats RefreshStats
	stats.SourcesEnabled = len(enabled)

	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, a.maxConcurrent)

	for _, source := range enabled {
		src := source
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			indexed, err := a.refreshSource(gctx, src)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				stats.SourcesFailed++
				if setErr := a.db.SetSourceLastError(gctx, src.ID, err.Error()); setErr != nil {
					return setErr
				}
				return nil
			}
			stats.PostsIndexed += indexed
			if err := a.db.SetSourceRefreshSuccess(gctx, src.ID); err != nil {
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return stats, err
	}

	if stats.SourcesFailed == stats.SourcesEnabled {
		return stats, fmt.Errorf("refresh: all %d enabled source(s) failed", stats.SourcesEnabled)
	}
	return stats, nil
}

func (a *App) refreshSource(ctx context.Context, src Source) (int, error) {
	resp, err := a.fetchFeed(ctx, src)
	if err != nil {
		return 0, fmt.Errorf("fetch %s: %w", src.Name, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotModified:
		return 0, nil
	case http.StatusOK:
		if err := a.db.SetSourceHTTPValidators(ctx, src.ID, resp.Header.Get("ETag"), resp.Header.Get("Last-Modified")); err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("fetch %s: HTTP %s", src.Name, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", src.Name, err)
	}
	entries, err := feed.Parse(bytes.NewReader(body))
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
