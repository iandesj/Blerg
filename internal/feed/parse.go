package feed

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type Entry struct {
	Title       string
	URL         string
	PublishedAt *time.Time
	Summary     string
	Categories  string
}

func Parse(r io.Reader) ([]Entry, error) {
	fp := gofeed.NewParser()
	doc, err := fp.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}
	out := make([]Entry, 0, len(doc.Items))
	for _, item := range doc.Items {
		url := item.Link
		if url == "" && item.GUID != "" {
			url = item.GUID
		}
		if url == "" || item.Title == "" {
			continue
		}
		var published *time.Time
		if item.PublishedParsed != nil {
			t := item.PublishedParsed.UTC()
			published = &t
		} else if item.UpdatedParsed != nil {
			t := item.UpdatedParsed.UTC()
			published = &t
		}
		summary := item.Description
		if summary == "" {
			summary = item.Content
		}
		out = append(out, Entry{
			Title:       item.Title,
			URL:         url,
			PublishedAt: published,
			Summary:     strings.TrimSpace(summary),
			Categories:  joinCategories(item.Categories),
		})
	}
	return out, nil
}

func joinCategories(cats []string) string {
	if len(cats) == 0 {
		return ""
	}
	return strings.Join(cats, ", ")
}
