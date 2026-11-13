package integration_test

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iandesj/blerg/internal/app"
	"github.com/iandesj/blerg/internal/cli"
)

func TestCLIFindReturnsExpectedPostAfterRefresh(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")

	refreshCode := cli.RunWithSettings(cli.Settings{
		DBPath:        dbPath,
		SourcesPath:   manifest,
		HTTPClient:    http.DefaultClient,
		UserAgent:     app.DefaultUserAgent,
		Staleness:     365 * 24 * time.Hour,
		MaxConcurrent: 4,
	}, []string{"--db", dbPath, "--sources", manifest, "refresh"}, nil, nil)
	if refreshCode != 0 {
		t.Fatalf("refresh exit %d", refreshCode)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	findCode := cli.RunWithSettings(cli.Settings{
		DBPath:        dbPath,
		SourcesPath:   manifest,
		HTTPClient:    http.DefaultClient,
		UserAgent:     app.DefaultUserAgent,
		Staleness:     365 * 24 * time.Hour,
		MaxConcurrent: 4,
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
		"Match:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("stdout missing %q:\n%s", want, text)
		}
	}
}

func TestCLIFindNoStrongMatchExitCode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "blerg.db")
	manifest := filepath.Join("testdata", "sources.yaml")
	refreshCode := cli.RunWithSettings(cli.Settings{
		DBPath:        dbPath,
		SourcesPath:   manifest,
		HTTPClient:    http.DefaultClient,
		Staleness:     365 * 24 * time.Hour,
		MaxConcurrent: 4,
		UserAgent:   app.DefaultUserAgent,
	}, []string{"--db", dbPath, "--sources", manifest, "refresh"}, io.Discard, io.Discard)
	if refreshCode != 0 {
		t.Fatalf("refresh exit %d", refreshCode)
	}

	var errOut bytes.Buffer
	findCode := cli.RunWithSettings(cli.Settings{
		DBPath:        dbPath,
		SourcesPath:   manifest,
		HTTPClient:    http.DefaultClient,
		UserAgent:     app.DefaultUserAgent,
		Staleness:     365 * 24 * time.Hour,
		MaxConcurrent: 4,
	}, []string{"--db", dbPath, "--sources", manifest, "find", "tears"}, io.Discard, &errOut)
	if findCode != cli.ExitNoMatch {
		t.Fatalf("find exit %d, want %d stderr=%q", findCode, cli.ExitNoMatch, errOut.String())
	}
	if !strings.Contains(errOut.String(), "no strong match") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}
