package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/iandesj/blerg/internal/config"
	"github.com/iandesj/blerg/internal/store"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Options struct {
	DBPath      string
	SourcesPath string
	HTTPClient  HTTPClient
	UserAgent   string
}

type Source = store.Source
type Post = store.Post

type App struct {
	db          *store.DB
	sourcesPath string
	httpClient  HTTPClient
	userAgent   string
}

func New(ctx context.Context, opts Options) (*App, error) {
	if opts.DBPath == "" {
		return nil, fmt.Errorf("db path is required")
	}
	if opts.SourcesPath == "" {
		return nil, fmt.Errorf("sources manifest path is required")
	}
	if opts.HTTPClient == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if opts.UserAgent == "" {
		opts.UserAgent = DefaultUserAgent
	}
	if err := os.MkdirAll(filepath.Dir(opts.DBPath), 0o755); err != nil {
		return nil, err
	}

	db, err := store.Open(opts.DBPath)
	if err != nil {
		return nil, err
	}

	a := &App{
		db:          db,
		sourcesPath: opts.SourcesPath,
		httpClient:  opts.HTTPClient,
		userAgent:   opts.UserAgent,
	}
	if err := a.syncSources(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return a, nil
}

func (a *App) Close() error {
	return a.db.Close()
}

func (a *App) syncSources(ctx context.Context) error {
	sources, err := config.LoadSources(a.sourcesPath)
	if err != nil {
		return err
	}
	for _, s := range sources {
		if _, err := a.db.UpsertSource(ctx, s.Name, s.FeedURL, s.Enabled); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) ListSources(ctx context.Context) ([]Source, error) {
	return a.db.ListSources(ctx)
}

func (a *App) UpsertPost(ctx context.Context, draft store.PostDraft) error {
	return a.db.UpsertPost(ctx, draft)
}

func (a *App) ListPosts(ctx context.Context) ([]Post, error) {
	return a.db.ListPosts(ctx)
}

func (a *App) HTTPGet(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", a.userAgent)
	return a.httpClient.Do(req)
}

func (a *App) FetchFeed(ctx context.Context, feedURL string) (*http.Response, error) {
	return a.HTTPGet(ctx, feedURL)
}
