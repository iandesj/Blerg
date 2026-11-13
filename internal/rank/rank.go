package rank

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
)

var ErrNoStrongMatch = errors.New("no strong match")

type Post struct {
	ID          int64
	Title       string
	URL         string
	Summary     string
	Categories  string
	SourceName  string
	PublishedAt *time.Time
}

type MatchResult struct {
	Post
	Explanation string
}

type Config struct {
	MinScore int
	HalfLife time.Duration
}

const (
	weightTitle      = 3
	weightSummary    = 1
	weightCategories = 2
	maxRecencyBoost  = 1.0
)

func DefaultConfig() Config {
	return Config{
		MinScore: 2,
		HalfLife: 180 * 24 * time.Hour,
	}
}

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

func baseScore(tokens []string, title, summary, categories string) int {
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

func effectiveScore(base int, published *time.Time, halfLife time.Duration) float64 {
	if published == nil || halfLife <= 0 {
		return float64(base)
	}
	age := time.Since(published.UTC())
	if age < 0 {
		age = 0
	}
	boost := maxRecencyBoost * math.Pow(0.5, float64(age)/float64(halfLife))
	return float64(base) + boost
}

func explain(tokens []string, title, summary, categories string) string {
	titleL := strings.ToLower(title)
	summaryL := strings.ToLower(summary)
	categoriesL := strings.ToLower(categories)
	var parts []string
	seen := map[string]struct{}{}
	for _, tok := range tokens {
		if strings.Contains(titleL, tok) {
			key := "title:" + tok
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				parts = append(parts, fmt.Sprintf("title (%s)", tok))
			}
		}
		if strings.Contains(categoriesL, tok) {
			key := "category:" + tok
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				parts = append(parts, fmt.Sprintf("category (%s)", tok))
			}
		}
		if strings.Contains(summaryL, tok) {
			key := "summary:" + tok
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				parts = append(parts, fmt.Sprintf("summary (%s)", tok))
			}
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

func BestMatch(query string, posts []Post, cfg Config) (MatchResult, error) {
	tokens := Tokenize(query)
	if len(tokens) == 0 || len(posts) == 0 {
		return MatchResult{}, ErrNoStrongMatch
	}
	if cfg.MinScore <= 0 {
		cfg.MinScore = DefaultConfig().MinScore
	}
	if cfg.HalfLife <= 0 {
		cfg.HalfLife = DefaultConfig().HalfLife
	}

	var best *Post
	bestBase := 0
	bestEffective := 0.0
	for i := range posts {
		p := &posts[i]
		base := baseScore(tokens, p.Title, p.Summary, p.Categories)
		if base == 0 {
			continue
		}
		eff := effectiveScore(base, p.PublishedAt, cfg.HalfLife)
		if best == nil || eff > bestEffective {
			best = p
			bestBase = base
			bestEffective = eff
			continue
		}
		if eff == bestEffective {
			if newer(p.PublishedAt, best.PublishedAt) {
				best = p
				bestBase = base
			} else if sameInstant(p.PublishedAt, best.PublishedAt) && p.ID < best.ID {
				best = p
				bestBase = base
			}
		}
	}
	if best == nil || bestBase < cfg.MinScore {
		return MatchResult{}, ErrNoStrongMatch
	}
	return MatchResult{
		Post:        *best,
		Explanation: explain(tokens, best.Title, best.Summary, best.Categories),
	}, nil
}

func newer(a, b *time.Time) bool {
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return a.UTC().After(b.UTC())
}

func sameInstant(a, b *time.Time) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return a.UTC().Equal(b.UTC())
	}
}
