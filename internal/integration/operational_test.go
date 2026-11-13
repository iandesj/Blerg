package integration_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iandesj/blerg/internal/app"
	"github.com/iandesj/blerg/internal/cli"
	"github.com/iandesj/blerg/internal/store"
)

func TestCLIRefreshSendsConditionalRequestOnSecondRefresh(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")
	fixtureBody := readTestdata(t, "testdata/feeds/smoke-item.rss")

	var fetchCount atomic.Int32
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		fetchCount.Add(1)
		if fetchCount.Load() > 1 {
			if req.Header.Get("If-None-Match") != `"feed-v1"` {
				t.Fatalf("If-None-Match = %q, want %q", req.Header.Get("If-None-Match"), `"feed-v1"`)
			}
			return &http.Response{
				StatusCode: http.StatusNotModified,
				Body:       http.NoBody,
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}
		h := make(http.Header)
		h.Set("ETag", `"feed-v1"`)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(fixtureBody)),
			Header:     h,
			Request:    req,
		}, nil
	})

	settings := cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  &http.Client{Transport: transport},
		UserAgent:   app.DefaultUserAgent,
	}
	args := []string{"--db", dbPath, "--sources", manifest, "refresh"}
	if code := cli.RunWithSettings(settings, args, nil, nil); code != 0 {
		t.Fatalf("first refresh exit %d", code)
	}
	if code := cli.RunWithSettings(settings, args, nil, nil); code != 0 {
		t.Fatalf("second refresh exit %d", code)
	}
	if fetchCount.Load() != 2 {
		t.Fatalf("fetches = %d, want 2", fetchCount.Load())
	}
}

func TestCLIFindJSONOutput(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")
	fixtureBody := readTestdata(t, "testdata/feeds/engineering-posts.rss")
	transport := fixtureFeedTransport(t, "https://example.com/engineering/feed.xml", fixtureBody)
	settings := cli.Settings{
		DBPath:        dbPath,
		SourcesPath:   manifest,
		HTTPClient:    &http.Client{Transport: transport},
		UserAgent:     app.DefaultUserAgent,
		Staleness:     365 * 24 * time.Hour,
		MaxConcurrent: 4,
	}

	if code := cli.RunWithSettings(settings, []string{"--db", dbPath, "--sources", manifest, "refresh"}, nil, nil); code != 0 {
		t.Fatalf("refresh exit %d", code)
	}

	var out bytes.Buffer
	if code := cli.RunWithSettings(settings, []string{"--db", dbPath, "--sources", manifest, "find", "--json", "pooling"}, &out, nil); code != 0 {
		t.Fatalf("find exit %d", code)
	}

	var doc struct {
		Title            string `json:"title"`
		URL              string `json:"url"`
		Source           string `json:"source"`
		Date             string `json:"date"`
		MatchExplanation string `json:"match_explanation"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v body=%q", err, out.String())
	}
	if doc.Title == "" || doc.URL == "" || doc.Source == "" {
		t.Fatalf("doc = %+v", doc)
	}
	if doc.MatchExplanation == "" {
		t.Fatal("expected match_explanation")
	}
}

func TestCLIFindRefreshesWhenIndexStale(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")
	fixtureBody := readTestdata(t, "testdata/feeds/smoke-item.rss")

	var feedFetches atomic.Int32
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == "https://example.com/engineering/feed.xml" {
			feedFetches.Add(1)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(fixtureBody)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	settings := cli.Settings{
		DBPath:        dbPath,
		SourcesPath:   manifest,
		HTTPClient:    &http.Client{Transport: transport},
		UserAgent:     app.DefaultUserAgent,
		Staleness:     time.Hour,
		MaxConcurrent: 2,
	}
	if code := cli.RunWithSettings(settings, []string{"--db", dbPath, "--sources", manifest, "refresh"}, nil, nil); code != 0 {
		t.Fatalf("refresh exit %d", code)
	}
	if feedFetches.Load() != 1 {
		t.Fatalf("after refresh fetches = %d", feedFetches.Load())
	}

	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	srcs, err := db.ListSources(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	staleAt := time.Now().UTC().Add(-48 * time.Hour)
	for _, src := range srcs {
		if src.Name == "Fixture Engineering Blog" {
			if err := db.SetSourceLastSuccessAt(t.Context(), src.ID, staleAt); err != nil {
				t.Fatal(err)
			}
		}
	}

	var out bytes.Buffer
	if code := cli.RunWithSettings(settings, []string{"--db", dbPath, "--sources", manifest, "find", "smoke"}, &out, nil); code != 0 {
		t.Fatalf("find exit %d", code)
	}
	if feedFetches.Load() != 2 {
		t.Fatalf("fetches after stale find = %d, want 2", feedFetches.Load())
	}
	if !strings.Contains(out.String(), "Smoke") {
		t.Fatalf("stdout=%q", out.String())
	}
}
