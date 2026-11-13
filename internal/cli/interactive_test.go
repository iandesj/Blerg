package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/iandesj/blerg/internal/rank"
)

func TestShuffleMatchesUsesPicker(t *testing.T) {
	t.Parallel()

	matches := []rank.MatchResult{
		{Post: rank.Post{Title: "A"}},
		{Post: rank.Post{Title: "B"}},
		{Post: rank.Post{Title: "C"}},
	}
	shuffleMatches(matches, func(n int) int { return 0 })
	got := []string{matches[0].Title, matches[1].Title, matches[2].Title}
	want := []string{"B", "C", "A"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestInteractiveMovesAndOpens(t *testing.T) {
	t.Parallel()

	matches := []rank.MatchResult{
		{Post: rank.Post{Title: "First", URL: "https://example.com/1"}},
		{Post: rank.Post{Title: "Second", URL: "https://example.com/2"}},
	}
	var opened []string
	var out bytes.Buffer
	err := runInteractive(strings.NewReader("j\rq"), &out, func(url string) error {
		opened = append(opened, url)
		return nil
	}, matches)
	if err != nil {
		t.Fatal(err)
	}
	if len(opened) != 1 || opened[0] != "https://example.com/2" {
		t.Fatalf("opened = %v", opened)
	}
	if !strings.Contains(out.String(), "Second") {
		t.Fatalf("stdout = %s", out.String())
	}
	if !strings.Contains(out.String(), "2/2") {
		t.Fatalf("stdout missing position:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Opened: 20") {
		t.Fatalf("stdout missing opened timestamp:\n%s", out.String())
	}
}

func TestRenderFrameColorsTitleAndURL(t *testing.T) {
	t.Parallel()

	match := rank.MatchResult{Post: rank.Post{
		Title: "First",
		URL:   "https://example.com/1",
	}}
	colored := renderFrame(match, 0, 2, "opened in browser", true)
	if !strings.Contains(colored, "\x1b[1mFirst") {
		t.Fatalf("title not bold:\n%s", colored)
	}
	if !strings.Contains(colored, "\x1b[36mhttps://example.com/1") {
		t.Fatalf("url not colored:\n%s", colored)
	}
	if !strings.Contains(colored, "\x1b[32mopened in browser") {
		t.Fatalf("note not green:\n%s", colored)
	}
	plain := renderFrame(match, 0, 2, "", false)
	if strings.Contains(plain, "\x1b") {
		t.Fatalf("plain frame has color: %q", plain)
	}
}

func TestInteractivePrevStopsAtFirst(t *testing.T) {
	t.Parallel()

	matches := []rank.MatchResult{
		{Post: rank.Post{Title: "First", URL: "https://example.com/1"}},
		{Post: rank.Post{Title: "Second", URL: "https://example.com/2"}},
	}
	var opened []string
	err := runInteractive(strings.NewReader("k\rq"), io.Discard, func(url string) error {
		opened = append(opened, url)
		return nil
	}, matches)
	if err != nil {
		t.Fatal(err)
	}
	if len(opened) != 1 || opened[0] != "https://example.com/1" {
		t.Fatalf("opened = %v", opened)
	}
}
