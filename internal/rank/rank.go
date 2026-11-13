package rank

import (
	"errors"
	"fmt"
	"math"
	mathrand "math/rand/v2"
	"slices"
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

// RandomMatch returns one post whose keyword score meets cfg.MinScore.
// pick(n) chooses an index in [0, n). A nil pick uses math/rand/v2.
// Candidates are in posts order.
func RandomMatch(query string, posts []Post, cfg Config, pick func(n int) int) (MatchResult, error) {
	tokens := Tokenize(query)
	if len(tokens) == 0 || len(posts) == 0 {
		return MatchResult{}, ErrNoStrongMatch
	}
	cfg = normalizeConfig(cfg)
	if pick == nil {
		pick = mathrand.IntN
	}

	var candidates []Post
	for i := range posts {
		p := posts[i]
		if baseScore(tokens, p.Title, p.Summary, p.Categories) >= cfg.MinScore {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return MatchResult{}, ErrNoStrongMatch
	}
	idx := pick(len(candidates))
	if idx < 0 || idx >= len(candidates) {
		return MatchResult{}, fmt.Errorf("rank: picker returned %d for %d candidates", idx, len(candidates))
	}
	chosen := candidates[idx]
	return MatchResult{
		Post:        chosen,
		Explanation: explain(tokens, chosen.Title, chosen.Summary, chosen.Categories),
	}, nil
}

func normalizeConfig(cfg Config) Config {
	if cfg.MinScore <= 0 {
		cfg.MinScore = DefaultConfig().MinScore
	}
	if cfg.HalfLife <= 0 {
		cfg.HalfLife = DefaultConfig().HalfLife
	}
	return cfg
}

func BestMatch(query string, posts []Post, cfg Config) (MatchResult, error) {
	matches, err := StrongMatches(query, posts, cfg)
	if err != nil {
		return MatchResult{}, err
	}
	return matches[0], nil
}

type scoredPost struct {
	post Post
	eff  float64
}

func StrongMatches(query string, posts []Post, cfg Config) ([]MatchResult, error) {
	tokens := Tokenize(query)
	if len(tokens) == 0 || len(posts) == 0 {
		return nil, ErrNoStrongMatch
	}
	cfg = normalizeConfig(cfg)

	var scored []scoredPost
	for i := range posts {
		p := posts[i]
		base := baseScore(tokens, p.Title, p.Summary, p.Categories)
		if base < cfg.MinScore {
			continue
		}
		scored = append(scored, scoredPost{
			post: p,
			eff:  effectiveScore(base, p.PublishedAt, cfg.HalfLife),
		})
	}
	if len(scored) == 0 {
		return nil, ErrNoStrongMatch
	}
	slices.SortStableFunc(scored, compareScored)

	out := make([]MatchResult, len(scored))
	for i, item := range scored {
		out[i] = MatchResult{
			Post:        item.post,
			Explanation: explain(tokens, item.post.Title, item.post.Summary, item.post.Categories),
		}
	}
	return out, nil
}

func compareScored(a, b scoredPost) int {
	if a.eff != b.eff {
		if a.eff > b.eff {
			return -1
		}
		return 1
	}
	if newer(a.post.PublishedAt, b.post.PublishedAt) {
		return -1
	}
	if newer(b.post.PublishedAt, a.post.PublishedAt) {
		return 1
	}
	switch {
	case a.post.ID < b.post.ID:
		return -1
	case a.post.ID > b.post.ID:
		return 1
	default:
		return 0
	}
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
