package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iandesj/blerg/internal/rank"
)

var ErrNoStrongMatch = errors.New("find: no strong match for query")

type FindJSON struct {
	Title            string `json:"title"`
	URL              string `json:"url"`
	Source           string `json:"source"`
	Date             string `json:"date"`
	MatchExplanation string `json:"match_explanation"`
}

func (a *App) FindJSON(match rank.MatchResult) FindJSON {
	out := FindJSON{
		Title:            match.Title,
		URL:              match.URL,
		Source:           match.SourceName,
		MatchExplanation: match.Explanation,
	}
	if match.PublishedAt != nil {
		out.Date = match.PublishedAt.UTC().Format(time.RFC3339)
	}
	return out
}

func (a *App) Find(ctx context.Context, query string) (rank.MatchResult, error) {
	return a.find(ctx, query, false)
}

func (a *App) FindRandom(ctx context.Context, query string) (rank.MatchResult, error) {
	return a.find(ctx, query, true)
}

func (a *App) find(ctx context.Context, query string, random bool) (rank.MatchResult, error) {
	if err := a.refreshIfStale(ctx); err != nil {
		return rank.MatchResult{}, err
	}

	rows, err := a.db.ListPostsWithSources(ctx)
	if err != nil {
		return rank.MatchResult{}, err
	}
	if len(rows) == 0 {
		return rank.MatchResult{}, fmt.Errorf("find: index is empty; run blerg refresh first")
	}

	posts := make([]rank.Post, len(rows))
	for i, row := range rows {
		posts[i] = rank.Post{
			ID:          row.ID,
			Title:       row.Title,
			URL:         row.URL,
			Summary:     row.Summary,
			Categories:  row.Categories,
			SourceName:  row.SourceName,
			PublishedAt: row.PublishedAt,
		}
	}

	var match rank.MatchResult
	if random {
		match, err = rank.RandomMatch(query, posts, rank.DefaultConfig(), nil)
	} else {
		match, err = rank.BestMatch(query, posts, rank.DefaultConfig())
	}
	if errors.Is(err, rank.ErrNoStrongMatch) {
		return rank.MatchResult{}, ErrNoStrongMatch
	}
	if err != nil {
		return rank.MatchResult{}, err
	}
	return match, nil
}

func (a *App) refreshIfStale(ctx context.Context) error {
	last, err := a.db.LatestRefreshAt(ctx)
	if err != nil {
		return err
	}
	if last == nil || a.now().Sub(*last) > a.staleness {
		_, err := a.Refresh(ctx)
		return err
	}
	return nil
}
