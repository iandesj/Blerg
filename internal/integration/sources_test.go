package integration_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iandesj/blerg/internal/app"
	"github.com/iandesj/blerg/internal/cli"
)

func TestCLIRefreshSkipsDisabledSources(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")

	code := cli.RunWithSettings(cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  http.DefaultClient,
		UserAgent:   app.DefaultUserAgent,
	}, []string{"--db", dbPath, "--sources", manifest, "refresh"}, nil, nil)
	if code != 0 {
		t.Fatalf("refresh exit %d", code)
	}

	a, err := app.New(context.Background(), app.Options{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  http.DefaultClient,
		UserAgent:   app.DefaultUserAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	sources, err := a.ListSources(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range sources {
		if src.Name == "Disabled Blog" && src.LastSuccessAt != nil {
			t.Fatal("disabled source should not have last_success_at")
		}
	}
}

func TestCLISourcesListShowsIngestStatus(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources-two-enabled.yaml")
	goodBody := readTestdata(t, "testdata/feeds/smoke-item.rss")

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://example.com/engineering/feed.xml":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(goodBody)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		case "https://example.com/second/feed.xml":
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader("bad")),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		default:
			t.Fatalf("unexpected URL %s", req.URL)
			return nil, nil
		}
	})

	settings := cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  &http.Client{Transport: transport},
		UserAgent:   app.DefaultUserAgent,
	}
	if code := cli.RunWithSettings(settings, []string{"--db", dbPath, "--sources", manifest, "refresh"}, nil, nil); code != 0 {
		t.Fatalf("refresh exit %d", code)
	}

	var out bytes.Buffer
	if code := cli.RunWithSettings(settings, []string{"--db", dbPath, "--sources", manifest, "sources", "list"}, &out, nil); code != 0 {
		t.Fatalf("sources list exit %d", code)
	}
	text := out.String()
	if !strings.Contains(text, "LAST_SUCCESS") || !strings.Contains(text, "LAST_ERROR") {
		t.Fatalf("missing status columns:\n%s", text)
	}
	if !strings.Contains(text, "Fixture Engineering Blog") {
		t.Fatalf("missing good source:\n%s", text)
	}
	if !strings.Contains(text, "Second Engineering Blog") {
		t.Fatalf("missing failed source:\n%s", text)
	}
	if !strings.Contains(text, "202") { // RFC3339 year fragment from last success
		t.Fatalf("expected last success timestamp on good source:\n%s", text)
	}
	if !strings.Contains(text, "HTTP") {
		t.Fatalf("expected last error on failed source:\n%s", text)
	}
}

func TestCLIRefreshDedupesPostURLAcrossSources(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources-dedup.yaml")
	bodyA := readTestdata(t, "testdata/feeds/dedup-feed-a.rss")
	bodyB := readTestdata(t, "testdata/feeds/dedup-feed-b.rss")

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://example.com/alpha/feed.xml":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(bodyA)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		case "https://example.com/beta/feed.xml":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(bodyB)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		default:
			t.Fatalf("unexpected URL %s", req.URL)
			return nil, nil
		}
	})

	code := cli.RunWithSettings(cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  &http.Client{Transport: transport},
		UserAgent:   app.DefaultUserAgent,
	}, []string{"--db", dbPath, "--sources", manifest, "refresh"}, nil, nil)
	if code != 0 {
		t.Fatalf("refresh exit %d", code)
	}

	a, err := app.New(context.Background(), app.Options{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  &http.Client{Transport: transport},
		UserAgent:   app.DefaultUserAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })

	posts, err := a.ListPosts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("posts = %d, want 1 deduped URL", len(posts))
	}
	if posts[0].URL != "https://example.com/posts/shared" {
		t.Fatalf("url = %q", posts[0].URL)
	}
}
