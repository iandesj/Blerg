package rank

import (
	"strings"
	"time"
	"unicode"
)

type Post struct {
	ID          int64
	Title       string
	URL         string
	Summary     string
	Categories  string
	SourceName  string
	PublishedAt *time.Time
}

const (
	weightTitle      = 3
	weightSummary    = 1
	weightCategories = 2
)

func Tokenize(query string) []string {
	query = strings.ToLower(query)
	var tokens []string
	for _, f := range strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}) {
		if f != "" {
			tokens = append(tokens, f)
		}
	}
	return tokens
}

func score(tokens []string, title, summary, categories string) int {
	title = strings.ToLower(title)
	summary = strings.ToLower(summary)
	categories = strings.ToLower(categories)
	total := 0
	for _, tok := range tokens {
		if strings.Contains(title, tok) {
			total += weightTitle
		}
		if strings.Contains(summary, tok) {
			total += weightSummary
		}
		if strings.Contains(categories, tok) {
			total += weightCategories
		}
	}
	return total
}

func BestMatch(query string, posts []Post) (Post, bool) {
	tokens := Tokenize(query)
	if len(tokens) == 0 || len(posts) == 0 {
		return Post{}, false
	}

	var best *Post
	bestScore := 0
	for i := range posts {
		p := &posts[i]
		s := score(tokens, p.Title, p.Summary, p.Categories)
		if s == 0 {
			continue
		}
		if best == nil || s > bestScore || (s == bestScore && p.ID < best.ID) {
			best = p
			bestScore = s
		}
	}
	if best == nil {
		return Post{}, false
	}
	return *best, true
}
