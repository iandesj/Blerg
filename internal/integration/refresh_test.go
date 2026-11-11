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

const fixtureRSS = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Fixture</title>
    <item>
      <title>Connection Pooling Deep Dive</title>
      <link>https://example.com/posts/pooling</link>
      <pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
      <description>How we pool connections.</description>
      <category>postgres</category>
    </item>
    <item>
      <title>Raft For Humans</title>
      <link>https://example.com/posts/raft</link>
      <description>Consensus without tears.</description>
    </item>
  </channel>
</rss>`

func TestCLIRefreshIndexesPostsFromFixtureFeed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")

	transport := fixtureFeedTransport(t, fixtureRSS)
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := cli.RunWithSettings(cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  &http.Client{Transport: transport},
		UserAgent:   app.DefaultUserAgent,
	}, []string{"--db", dbPath, "--sources", manifest, "refresh"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	if !strings.Contains(out.String(), "indexed 2 post(s)") {
		t.Fatalf("stdout = %q", out.String())
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
	if len(posts) != 2 {
		t.Fatalf("posts = %d, want 2", len(posts))
	}
	for _, want := range []struct {
		title, url, cats string
	}{
		{"Connection Pooling Deep Dive", "https://example.com/posts/pooling", "postgres"},
		{"Raft For Humans", "https://example.com/posts/raft", ""},
	} {
		found := false
		for _, p := range posts {
			if p.URL == want.url {
				found = true
				if p.Title != want.title {
					t.Fatalf("title = %q, want %q", p.Title, want.title)
				}
				if p.Categories != want.cats {
					t.Fatalf("categories = %q, want %q", p.Categories, want.cats)
				}
				if want.url == "https://example.com/posts/pooling" && p.Summary == "" {
					t.Fatal("expected summary on pooling post")
				}
				break
			}
		}
		if !found {
			t.Fatalf("missing post %q", want.url)
		}
	}

	sources, err := a.ListSources(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range sources {
		if src.Name == "Fixture Engineering Blog" {
			if src.LastSuccessAt == nil {
				t.Fatal("expected last_success_at on fixture source")
			}
			if src.LastError != "" {
				t.Fatalf("last_error = %q", src.LastError)
			}
		}
	}
}

func TestCLIRefreshIdempotentByURL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")
	transport := fixtureFeedTransport(t, fixtureRSS)

	run := func() {
		code := cli.RunWithSettings(cli.Settings{
			DBPath:      dbPath,
			SourcesPath: manifest,
			HTTPClient:  &http.Client{Transport: transport},
			UserAgent:   app.DefaultUserAgent,
		}, []string{"--db", dbPath, "--sources", manifest, "refresh"}, io.Discard, io.Discard)
		if code != 0 {
			t.Fatalf("refresh exit %d", code)
		}
	}
	run()
	run()

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
	if len(posts) != 2 {
		t.Fatalf("posts after double refresh = %d, want 2", len(posts))
	}
}

func TestCLIRefreshNonZeroWhenFeedFails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("error")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	code := cli.RunWithSettings(cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  &http.Client{Transport: transport},
		UserAgent:   app.DefaultUserAgent,
	}, []string{"--db", dbPath, "--sources", manifest, "refresh"}, io.Discard, io.Discard)
	if code == 0 {
		t.Fatal("expected non-zero exit when enabled feed fetch fails")
	}
}

func fixtureFeedTransport(t *testing.T, body string) roundTripFunc {
	t.Helper()
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://example.com/engineering/feed.xml" {
			t.Fatalf("unexpected URL %s", req.URL)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
}
