package rank_test

import (
	"errors"
	"strings"
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

	got, err := rank.BestMatch("pooling postgres", posts, rank.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Connection Pooling Deep Dive" {
		t.Fatalf("title = %q", got.Title)
	}
	if got.Explanation == "" {
		t.Fatal("expected match explanation")
	}
}

func TestBestMatchNoOverlap(t *testing.T) {
	t.Parallel()

	posts := []rank.Post{
		{ID: 1, Title: "Unrelated", Summary: "Nothing here"},
	}
	_, err := rank.BestMatch("quantum", posts, rank.DefaultConfig())
	if !errors.Is(err, rank.ErrNoStrongMatch) {
		t.Fatalf("err = %v, want ErrNoStrongMatch", err)
	}
}

func TestBestMatchBelowThreshold(t *testing.T) {
	t.Parallel()

	posts := []rank.Post{
		{ID: 1, Title: "Unrelated", Summary: "mentions quantum physics"},
	}
	_, err := rank.BestMatch("quantum", posts, rank.DefaultConfig())
	if !errors.Is(err, rank.ErrNoStrongMatch) {
		t.Fatalf("err = %v, want ErrNoStrongMatch", err)
	}
}

func TestBestMatchPrefersNewerOnNearTie(t *testing.T) {
	t.Parallel()

	old := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	posts := []rank.Post{
		{ID: 1, Title: "Rust Memory Safety", Summary: "basics", PublishedAt: &old},
		{ID: 2, Title: "Rust Memory Patterns", Summary: "advanced", PublishedAt: &recent},
	}

	got, err := rank.BestMatch("rust memory", posts, rank.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 2 {
		t.Fatalf("post id = %d, want newer post 2", got.ID)
	}
}

func TestStrongMatchesOrdersByScoreAndDropsWeak(t *testing.T) {
	t.Parallel()

	old := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	posts := []rank.Post{
		{ID: 1, Title: "Rust Memory Safety", Summary: "basics", PublishedAt: &old},
		{ID: 2, Title: "Unrelated", Summary: "mentions rust once", PublishedAt: &recent},
		{ID: 3, Title: "Rust Memory Patterns", Summary: "advanced", PublishedAt: &recent},
	}

	got, err := rank.StrongMatches("rust memory", posts, rank.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != 3 || got[1].ID != 1 {
		t.Fatalf("order = %d, %d", got[0].ID, got[1].ID)
	}
}

func TestRandomMatchChoosesAmongStrongMatches(t *testing.T) {
	t.Parallel()

	posts := []rank.Post{
		{ID: 1, Title: "Rust Memory Safety", Summary: "basics"},
		{ID: 2, Title: "Unrelated", Summary: "mentions rust once"},
		{ID: 3, Title: "Rust Ownership", Summary: "patterns"},
	}

	got, err := rank.RandomMatch("rust", posts, rank.DefaultConfig(), func(n int) int {
		if n != 2 {
			t.Fatalf("candidate count = %d, want 2", n)
		}
		return n - 1
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 3 {
		t.Fatalf("post id = %d, want last strong match 3", got.ID)
	}
	if !strings.Contains(got.Explanation, "title") {
		t.Fatalf("explanation = %q", got.Explanation)
	}
}

func TestRandomMatchNoStrongMatch(t *testing.T) {
	t.Parallel()

	posts := []rank.Post{
		{ID: 1, Title: "Unrelated", Summary: "mentions quantum physics"},
	}
	_, err := rank.RandomMatch("quantum", posts, rank.DefaultConfig(), func(int) int { return 0 })
	if !errors.Is(err, rank.ErrNoStrongMatch) {
		t.Fatalf("err = %v, want ErrNoStrongMatch", err)
	}
}

func TestBestMatchExplanationListsFields(t *testing.T) {
	t.Parallel()

	pub := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	posts := []rank.Post{
		{ID: 1, Title: "Go Channels", Summary: "concurrency", Categories: "go", PublishedAt: &pub},
	}
	got, err := rank.BestMatch("go channels", posts, rank.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Explanation, "title") {
		t.Fatalf("explanation = %q", got.Explanation)
	}
}
