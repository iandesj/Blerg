package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/iandesj/blerg/internal/rank"
)

var ErrNoStrongMatch = errors.New("find: no strong match for query")

func (a *App) Find(ctx context.Context, query string) (rank.MatchResult, error) {
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

	match, err := rank.BestMatch(query, posts, rank.DefaultConfig())
	if errors.Is(err, rank.ErrNoStrongMatch) {
		return rank.MatchResult{}, ErrNoStrongMatch
	}
	if err != nil {
		return rank.MatchResult{}, err
	}
	return match, nil
}
