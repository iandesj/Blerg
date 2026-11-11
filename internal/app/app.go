package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
		opts.UserAgent = "blerg"
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

type Source struct {
	ID            int64
	Name          string
	FeedURL       string
	Enabled       bool
	LastSuccessAt *time.Time
	LastError     string
}

func (a *App) ListSources(ctx context.Context) ([]Source, error) {
	rows, err := a.db.ListSources(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Source, len(rows))
	for i, r := range rows {
		out[i] = Source{
			ID:            r.ID,
			Name:          r.Name,
			FeedURL:       r.FeedURL,
			Enabled:       r.Enabled,
			LastSuccessAt: r.LastSuccessAt,
			LastError:     r.LastError,
		}
	}
	return out, nil
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

func (a *App) UpsertPost(ctx context.Context, sourceID int64, title, url string, publishedAt *time.Time, summary, categories string) error {
	return a.db.UpsertPost(ctx, sourceID, title, url, publishedAt, summary, categories)
}

func (a *App) ListPosts(ctx context.Context) ([]Post, error) {
	rows, err := a.db.ListPosts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Post, len(rows))
	for i, r := range rows {
		out[i] = Post{
			ID:          r.ID,
			SourceID:    r.SourceID,
			Title:       r.Title,
			URL:         r.URL,
			PublishedAt: r.PublishedAt,
			Summary:     r.Summary,
			Categories:  r.Categories,
		}
	}
	return out, nil
}

func (a *App) HTTPGet(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", a.userAgent)
	return a.httpClient.Do(req)
}
