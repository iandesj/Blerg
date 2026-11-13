package cli

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/iandesj/blerg/internal/app"
)

func TestVersionFlagPrintsBuildVersion(t *testing.T) {
	previous := Version
	Version = "9.9.9"
	t.Cleanup(func() { Version = previous })

	var out bytes.Buffer
	code := RunWithSettings(Settings{HTTPClient: http.DefaultClient, UserAgent: app.DefaultUserAgent}, []string{"--version"}, &out, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out.String(), "9.9.9") {
		t.Fatalf("stdout = %q", out.String())
	}
}
