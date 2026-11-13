package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateReplacesBinaryFromLatestRelease(t *testing.T) {
	t.Parallel()

	archive := gzipTar(t, releaseBinaryName(), []byte("new-binary"))
	var gotAsset string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/iandesj/Blerg/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tag_name":"v0.2.0","assets":[{"name":"` + releaseAssetName() + `","browser_download_url":"http://` + r.Host + `/asset"}]}`))
		case "/asset":
			gotAsset = r.URL.Path
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	dest := filepath.Join(t.TempDir(), "blerg")
	if err := os.WriteFile(dest, []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := updateBinary(context.Background(), srv.Client(), updateSpec{APIBase: srv.URL, Current: "0.1.0", Dest: dest}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if gotAsset != "/asset" {
		t.Fatalf("asset path = %q", gotAsset)
	}
	body, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "new-binary" {
		t.Fatalf("binary = %q", body)
	}
	if !strings.Contains(out.String(), "0.2.0") {
		t.Fatalf("stdout = %q", out.String())
	}
}

func TestUpdateAlreadyCurrent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v0.2.0","assets":[]}`))
	}))
	t.Cleanup(srv.Close)

	dest := filepath.Join(t.TempDir(), "blerg")
	if err := os.WriteFile(dest, []byte("same"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := updateBinary(context.Background(), srv.Client(), updateSpec{APIBase: srv.URL, Current: "0.2.0", Dest: dest}, &out)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "same" {
		t.Fatalf("binary changed: %q", body)
	}
	if !strings.Contains(out.String(), "up to date") {
		t.Fatalf("stdout = %q", out.String())
	}
}

func gzipTar(t *testing.T, name string, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zw)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(payload))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
