package integration_test

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iandesj/blerg/internal/app"
	"github.com/iandesj/blerg/internal/cli"
)

func TestCLIFindReturnsExpectedPostAfterRefresh(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")
	fixtureBody := readTestdata(t, "testdata/feeds/engineering-posts.rss")
	transport := fixtureFeedTransport(t, "https://example.com/engineering/feed.xml", fixtureBody)

	refreshCode := cli.RunWithSettings(cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  &http.Client{Transport: transport},
		UserAgent:   app.DefaultUserAgent,
	}, []string{"--db", dbPath, "--sources", manifest, "refresh"}, nil, nil)
	if refreshCode != 0 {
		t.Fatalf("refresh exit %d", refreshCode)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	findCode := cli.RunWithSettings(cli.Settings{
		DBPath:      dbPath,
		SourcesPath: manifest,
		HTTPClient:  &http.Client{Transport: transport},
		UserAgent:   app.DefaultUserAgent,
	}, []string{"--db", dbPath, "--sources", manifest, "find", "pooling postgres"}, &out, &errOut)
	if findCode != 0 {
		t.Fatalf("find exit %d stderr=%q", findCode, errOut.String())
	}

	text := out.String()
	for _, want := range []string{
		"Connection Pooling Deep Dive",
		"https://example.com/posts/pooling",
		"Fixture Engineering Blog",
		"2024",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("stdout missing %q:\n%s", want, text)
		}
	}
}
