package rank_test

import (
	"testing"
	"time"

	"github.com/iandesj/blerg/internal/rank"
)

func TestScorePrefersTitleAndCategoryMatches(t *testing.T) {
	t.Parallel()

	older := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	posts := []rank.Post{
		{ID: 1, Title: "Connection Pooling Deep Dive", Summary: "How we pool connections.", Categories: "postgres", PublishedAt: &older},
		{ID: 2, Title: "Raft For Humans", Summary: "Consensus without tears.", PublishedAt: &newer},
	}

	got, ok := rank.BestMatch("pooling postgres", posts)
	if !ok {
		t.Fatal("expected a match")
	}
	if got.Title != "Connection Pooling Deep Dive" {
		t.Fatalf("title = %q", got.Title)
	}
}

func TestBestMatchNoOverlap(t *testing.T) {
	t.Parallel()

	posts := []rank.Post{
		{ID: 1, Title: "Unrelated", Summary: "Nothing here"},
	}
	_, ok := rank.BestMatch("quantum", posts)
	if ok {
		t.Fatal("expected no match")
	}
}
