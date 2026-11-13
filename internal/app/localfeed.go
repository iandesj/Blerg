package app

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) localFeedPath(feedURL string) (string, bool) {
	if strings.HasPrefix(strings.ToLower(feedURL), "http://") ||
		strings.HasPrefix(strings.ToLower(feedURL), "https://") {
		return "", false
	}
	path := feedURL
	if strings.HasPrefix(feedURL, "file://") {
		path = strings.TrimPrefix(feedURL, "file://")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(a.sourcesPath), path)
	}
	return path, true
}

func (a *App) readLocalFeed(ctx context.Context, path string) (*http.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}, nil
}
