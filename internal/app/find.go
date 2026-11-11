package app

import (
	"context"
	"fmt"

	"github.com/iandesj/blerg/internal/rank"
)

func (a *App) Find(ctx context.Context, query string) (rank.Post, error) {
	rows, err := a.db.ListPostsWithSources(ctx)
	if err != nil {
		return rank.Post{}, err
	}
	if len(rows) == 0 {
		return rank.Post{}, fmt.Errorf("find: index is empty; run blerg refresh first")
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

	match, ok := rank.BestMatch(query, posts)
	if !ok {
		return rank.Post{}, fmt.Errorf("find: no matching posts for query")
	}
	return match, nil
}
