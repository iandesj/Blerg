package feed_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iandesj/blerg/internal/feed"
)

const sampleRSS = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Fixture</title>
    <item>
      <title>Post One</title>
      <link>https://example.com/posts/one</link>
      <pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
      <description>Summary one</description>
      <category>go</category>
      <category>sqlite</category>
    </item>
    <item>
      <title>Post Two</title>
      <link>https://example.com/posts/two</link>
      <description>Summary two</description>
    </item>
  </channel>
</rss>`

func TestParseRSSItems(t *testing.T) {
	t.Parallel()

	entries, err := feed.Parse(strings.NewReader(sampleRSS))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0].Title != "Post One" || entries[0].URL != "https://example.com/posts/one" {
		t.Fatalf("first entry = %+v", entries[0])
	}
	if entries[0].Summary != "Summary one" {
		t.Fatalf("summary = %q", entries[0].Summary)
	}
	if entries[0].Categories != "go, sqlite" {
		t.Fatalf("categories = %q", entries[0].Categories)
	}
	want := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if entries[0].PublishedAt == nil || !entries[0].PublishedAt.UTC().Equal(want) {
		t.Fatalf("published = %v, want %v", entries[0].PublishedAt, want)
	}
}

func TestParseAtomFromFile(t *testing.T) {
	t.Parallel()

	f, err := os.Open("../../testdata/feeds/atom.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	entries, err := feed.Parse(f)
	if err != nil {
		t.Fatalf("parse atom: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].Title != "Atom Post" || entries[0].URL != "https://example.com/atom/1" {
		t.Fatalf("entry = %+v", entries[0])
	}
}
