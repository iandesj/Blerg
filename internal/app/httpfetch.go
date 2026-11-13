package app

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const maxHTTPRetries = 3

func (a *App) fetchFeed(ctx context.Context, src Source) (*http.Response, error) {
	if path, ok := a.localFeedPath(src.FeedURL); ok {
		return a.readLocalFeed(ctx, path)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.FeedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", a.userAgent)
	if src.ETag != "" {
		req.Header.Set("If-None-Match", src.ETag)
	}
	if src.LastModified != "" {
		req.Header.Set("If-Modified-Since", src.LastModified)
	}
	return a.doWithBackoff(ctx, req)
}

func (a *App) doWithBackoff(ctx context.Context, req *http.Request) (*http.Response, error) {
	backoff := time.Second
	var lastStatus int
	for attempt := 0; attempt < maxHTTPRetries; attempt++ {
		resp, err := a.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable {
			return resp, nil
		}
		lastStatus = resp.StatusCode
		_ = resp.Body.Close()
		if attempt == maxHTTPRetries-1 {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return nil, fmt.Errorf("HTTP %d after %d attempts", lastStatus, maxHTTPRetries)
}
