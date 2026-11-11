package integration_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iandesj/blerg/internal/app"
	"github.com/iandesj/blerg/internal/cli"
)

func TestCLIHelpListsCommands(t *testing.T) {
	t.Parallel()

	root := cli.NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}

	help := out.String()
	for _, sub := range []string{"refresh", "find", "sources"} {
		if !strings.Contains(help, sub) {
			t.Fatalf("help missing subcommand %q:\n%s", sub, help)
		}
	}
}

func TestHarnessLoadsManifestAndMigratesSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")

	a, err := app.New(context.Background(), app.Options{
		DBPath:       dbPath,
		SourcesPath:  manifest,
		HTTPClient:   http.DefaultClient,
		UserAgent:    "blerg-test",
	})
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	sources, err := a.ListSources(context.Background())
	if err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if len(sources) != 2 {
		t.Fatalf("sources count = %d, want 2", len(sources))
	}
	var fixture *app.Source
	for i := range sources {
		if sources[i].Name == "Fixture Engineering Blog" {
			fixture = &sources[i]
			break
		}
	}
	if fixture == nil {
		t.Fatal("missing Fixture Engineering Blog in sources")
	}
	if !fixture.Enabled {
		t.Fatal("fixture source should be enabled")
	}
}

func TestHarnessUsesInjectableHTTP(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("User-Agent") != "blerg-test" {
			t.Errorf("User-Agent = %q, want blerg-test", req.Header.Get("User-Agent"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`<rss version="2.0"></rss>`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	a, err := app.New(context.Background(), app.Options{
		DBPath:       dbPath,
		SourcesPath:  manifest,
		HTTPClient:   &http.Client{Transport: transport},
		UserAgent:    "blerg-test",
	})
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	if _, err := a.HTTPGet(context.Background(), "https://example.com/feed.xml"); err != nil {
		t.Fatalf("http get: %v", err)
	}
}

func TestPostURLUniqueConstraint(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")

	a, err := app.New(context.Background(), app.Options{
		DBPath:       dbPath,
		SourcesPath:  manifest,
		HTTPClient:   http.DefaultClient,
		UserAgent:    "blerg-test",
	})
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	ctx := context.Background()
	sources, err := a.ListSources(ctx)
	if err != nil || len(sources) == 0 {
		t.Fatalf("sources: %v", err)
	}
	sourceID := sources[0].ID

	if err := a.UpsertPost(ctx, sourceID, "First", "https://example.com/a", nil, "", ""); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := a.UpsertPost(ctx, sourceID, "Second", "https://example.com/a", nil, "", ""); err != nil {
		t.Fatalf("second upsert same url: %v", err)
	}

	posts, err := a.ListPosts(ctx)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("posts count = %d, want 1", len(posts))
	}
	if posts[0].Title != "Second" {
		t.Fatalf("title = %q, want Second", posts[0].Title)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err := os.Chdir(filepath.Join(wd, "..", "..")); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
