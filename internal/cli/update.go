package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const releaseRepoPath = "/repos/iandesj/Blerg/releases/latest"

type updateSpec struct {
	APIBase string
	Current string
	Dest    string
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func updateBinary(ctx context.Context, client *http.Client, spec updateSpec, out io.Writer) error {
	rel, err := fetchLatestRelease(ctx, client, spec.APIBase)
	if err != nil {
		return err
	}
	latest := strings.TrimPrefix(rel.TagName, "v")
	if latest == spec.Current {
		_, err := fmt.Fprintf(out, "blerg %s is up to date\n", spec.Current)
		return err
	}
	assetURL, err := assetURL(rel, releaseAssetName())
	if err != nil {
		return err
	}
	body, err := getBytes(ctx, client, assetURL)
	if err != nil {
		return err
	}
	bin, err := extractReleaseBinary(body, releaseAssetName(), releaseBinaryName())
	if err != nil {
		return err
	}
	if err := replaceFile(spec.Dest, bin); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "updated blerg %s to %s\n", spec.Current, latest)
	return err
}

func fetchLatestRelease(ctx context.Context, client *http.Client, apiBase string) (githubRelease, error) {
	body, err := getBytes(ctx, client, strings.TrimRight(apiBase, "/")+releaseRepoPath)
	if err != nil {
		return githubRelease{}, err
	}
	var rel githubRelease
	if err := json.Unmarshal(body, &rel); err != nil {
		return githubRelease{}, fmt.Errorf("update: parse release: %w", err)
	}
	if rel.TagName == "" {
		return githubRelease{}, fmt.Errorf("update: latest release has no tag")
	}
	return rel, nil
}

func assetURL(rel githubRelease, name string) (string, error) {
	for _, asset := range rel.Assets {
		if asset.Name == name {
			if asset.URL == "" {
				return "", fmt.Errorf("update: asset %s has no download URL", name)
			}
			return asset.URL, nil
		}
	}
	return "", fmt.Errorf("update: release has no %s asset", name)
}

func getBytes(ctx context.Context, client *http.Client, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Blerg (https://github.com/iandesj/Blerg)")
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update: GET %s: %s", rawURL, resp.Status)
	}
	return body, nil
}

func extractReleaseBinary(archive []byte, assetName, binaryName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractZip(archive, binaryName)
	}
	return extractTarGZ(archive, binaryName)
}

func extractTarGZ(archive []byte, binaryName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("update: read archive: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("update: archive missing %s", binaryName)
		}
		if err != nil {
			return nil, fmt.Errorf("update: read archive: %w", err)
		}
		if path.Base(hdr.Name) != binaryName || hdr.FileInfo().IsDir() {
			continue
		}
		return io.ReadAll(io.LimitReader(tr, 64<<20))
	}
}

func extractZip(archive []byte, binaryName string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("update: read archive: %w", err)
	}
	for _, file := range zr.File {
		if path.Base(file.Name) != binaryName || file.FileInfo().IsDir() {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(io.LimitReader(rc, 64<<20))
		_ = rc.Close()
		return body, err
	}
	return nil, fmt.Errorf("update: archive missing %s", binaryName)
}

func replaceFile(dest string, contents []byte) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, ".blerg-update-*")
	if err != nil {
		return fmt.Errorf("update: write %s: %w", dest, err)
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(contents); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("update: replace %s: %w", dest, err)
	}
	ok = true
	return nil
}

func releaseAssetName() string {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	return "blerg_" + runtime.GOOS + "_" + runtime.GOARCH + ext
}

func releaseBinaryName() string {
	if runtime.GOOS == "windows" {
		return "blerg.exe"
	}
	return "blerg"
}
