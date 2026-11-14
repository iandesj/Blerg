package cli

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSourcesListInstallsDefaultManifest(t *testing.T) {
	manifest := filepath.Join(isolatedConfigHome(t), "Blerg", "sources.yaml")
	settings, dbPath := settingsForIsolatedConfig(t)

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := RunWithSettings(settings, []string{"--db", dbPath, "sources", "list"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	if _, err := os.Stat(manifest); err != nil {
		t.Fatalf("default manifest: %v", err)
	}
	if !strings.Contains(out.String(), "Netflix Tech Blog") {
		t.Fatalf("stdout = %q", out.String())
	}
}

func TestSourcesListKeepsExistingUserManifest(t *testing.T) {
	manifest := filepath.Join(isolatedConfigHome(t), "Blerg", "sources.yaml")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	const body = "sources:\n  - name: Custom Only Blog\n    feed_url: https://example.com/feed.xml\n    enabled: true\n"
	if err := os.WriteFile(manifest, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	settings, dbPath := settingsForIsolatedConfig(t)
	var out bytes.Buffer
	code := RunWithSettings(settings, []string{"--db", dbPath, "sources", "list"}, &out, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d stdout=%q", code, out.String())
	}
	if !strings.Contains(out.String(), "Custom Only Blog") {
		t.Fatalf("stdout = %q", out.String())
	}
	if strings.Contains(out.String(), "Netflix Tech Blog") {
		t.Fatalf("shipped manifest replaced user file:\n%s", out.String())
	}
	got, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("manifest = %q", got)
	}
}

func TestExplicitSourcesDoesNotInstallDefault(t *testing.T) {
	home := isolatedConfigHome(t)
	custom := filepath.Join(t.TempDir(), "sources.yaml")
	const body = "sources:\n  - name: Explicit Blog\n    feed_url: https://example.com/feed.xml\n    enabled: true\n"
	if err := os.WriteFile(custom, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	settings, dbPath := settingsForIsolatedConfig(t)
	var errOut bytes.Buffer
	code := RunWithSettings(settings, []string{"--db", dbPath, "--sources", custom, "sources", "list"}, &bytes.Buffer{}, &errOut)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, errOut.String())
	}
	defaultManifest := filepath.Join(home, "Blerg", "sources.yaml")
	if _, err := os.Stat(defaultManifest); !os.IsNotExist(err) {
		t.Fatalf("default manifest installed: %v", err)
	}
}

func settingsForIsolatedConfig(t *testing.T) (Settings, string) {
	t.Helper()
	settings, err := defaultSettings()
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(t.TempDir(), "blerg.db")
	settings.DBPath = dbPath
	settings.HTTPClient = http.DefaultClient
	return settings, dbPath
}

func isolatedConfigHome(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", root)
	case "darwin":
		t.Setenv("HOME", root)
	default:
		t.Setenv("XDG_CONFIG_HOME", root)
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
