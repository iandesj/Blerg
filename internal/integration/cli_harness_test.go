package integration_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iandesj/blerg/internal/app"
	"github.com/iandesj/blerg/internal/cli"
	"github.com/iandesj/blerg/internal/store"
)

func TestCLIHelpListsCommands(t *testing.T) {
	t.Parallel()

	root := cli.NewRootCommand(cli.Settings{UserAgent: app.DefaultUserAgent})
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

func TestCLISourcesList(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := cli.RunWithSettings(cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  http.DefaultClient,
		UserAgent:   "Blerg/integration-test",
	}, []string{"--db", dbPath, "--sources", manifest, "sources", "list"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Fixture Engineering Blog") {
		t.Fatalf("output missing fixture source:\n%s", out.String())
	}
}

func TestCLIRefreshUsesInjectableHTTPAndCannedFeed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources-http.yaml")

	cannedRSS, err := os.ReadFile(filepath.Join("testdata", "feeds", "smoke-item.rss"))
	if err != nil {
		t.Fatal(err)
	}
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.Header.Get("User-Agent"), "Blerg") {
			t.Errorf("User-Agent = %q, want substring Blerg", req.Header.Get("User-Agent"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(cannedRSS)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

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
	if !strings.Contains(out.String(), "indexed 1 post(s)") {
		t.Fatalf("stdout = %q", out.String())
	}
	if len(cannedRSS) == 0 {
		t.Fatal("canned RSS missing")
	}
}

func TestCLISubprocessSourcesList(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest, err := filepath.Abs("testdata/sources.yaml")
	if err != nil {
		t.Fatal(err)
	}
	bin := buildCLI(t)

	cmd := exec.Command(bin, "--db", dbPath, "--sources", manifest, "sources", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v output=%s", err, out)
	}
	if !strings.Contains(string(out), "Fixture Engineering Blog") {
		t.Fatalf("output=%s", out)
	}
}

func TestPostURLUniqueConstraintViaCLIPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")

	a, err := app.New(context.Background(), app.Options{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  http.DefaultClient,
		UserAgent:   app.DefaultUserAgent,
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

	var sourceID int64
	for _, s := range sources {
		if s.Name == "Fixture Engineering Blog" {
			sourceID = s.ID
			break
		}
	}
	if sourceID == 0 {
		t.Fatal("fixture source id missing")
	}

	if err := a.UpsertPost(ctx, store.PostDraft{
		SourceID: sourceID,
		Title:    "First",
		URL:      "https://example.com/a",
	}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := a.UpsertPost(ctx, store.PostDraft{
		SourceID: sourceID,
		Title:    "Second",
		URL:      "https://example.com/a",
	}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	posts, err := a.ListPosts(ctx)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if len(posts) != 1 || posts[0].Title != "Second" {
		t.Fatalf("posts=%+v", posts)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func buildCLI(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "blerg")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/blerg")
	cmd.Dir = mustRepoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
		return wd
	}
	return filepath.Join(wd, "..", "..")
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
