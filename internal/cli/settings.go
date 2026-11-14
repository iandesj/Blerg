package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/iandesj/blerg/internal/app"
	"github.com/iandesj/blerg/internal/rank"
	"github.com/spf13/cobra"
)

type Settings struct {
	DBPath        string
	SourcesPath   string
	HTTPClient    app.HTTPClient
	UserAgent     string
	Staleness     time.Duration
	MaxConcurrent int
}

func defaultSettings() (Settings, error) {
	dbPath, err := app.DefaultDBPath()
	if err != nil {
		return Settings{}, err
	}
	sourcesPath, err := app.DefaultSourcesPath()
	if err != nil {
		return Settings{}, err
	}
	return Settings{
		DBPath:        dbPath,
		SourcesPath:   sourcesPath,
		HTTPClient:    http.DefaultClient,
		UserAgent:     app.DefaultUserAgent,
		Staleness:     app.DefaultStaleness,
		MaxConcurrent: app.DefaultMaxConcurrent,
	}, nil
}

func (s *Settings) openApp(ctx context.Context) (*app.App, error) {
	if err := app.EnsureDefaultSources(s.SourcesPath); err != nil {
		return nil, err
	}
	return app.New(ctx, app.Options{
		DBPath:        s.DBPath,
		SourcesPath:   s.SourcesPath,
		HTTPClient:    s.HTTPClient,
		UserAgent:     s.UserAgent,
		Staleness:     s.Staleness,
		MaxConcurrent: s.MaxConcurrent,
	})
}

func NewRootCommand(s Settings) *cobra.Command {
	if s.HTTPClient == nil {
		s.HTTPClient = http.DefaultClient
	}
	if s.UserAgent == "" {
		s.UserAgent = app.DefaultUserAgent
	}
	settings := s

	root := &cobra.Command{
		Use:     "blerg",
		Short:   "Blerg finds a single engineering blog post for a topic",
		Version: Version,
	}
	root.PersistentFlags().StringVar(&settings.DBPath, "db", settings.DBPath, "SQLite index path")
	root.PersistentFlags().StringVar(&settings.SourcesPath, "sources", settings.SourcesPath, "Sources manifest path")
	root.PersistentFlags().DurationVar(&settings.Staleness, "stale-after", settings.Staleness, "Refresh automatically during find when the index is older than this duration")
	root.PersistentFlags().IntVar(&settings.MaxConcurrent, "max-concurrent", settings.MaxConcurrent, "Maximum concurrent feed fetches during refresh")

	root.AddCommand(newRefreshCommand(&settings))
	root.AddCommand(newFindCommand(&settings))
	root.AddCommand(newSourcesCommand(&settings))
	root.AddCommand(newUpdateCommand())
	return root
}

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Replace this blerg binary with the latest GitHub release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			if resolved, err := filepath.EvalSymlinks(exe); err == nil {
				exe = resolved
			}
			return updateBinary(cmd.Context(), http.DefaultClient, updateSpec{
				APIBase: "https://api.github.com",
				Current: Version,
				Dest:    exe,
			}, cmd.OutOrStdout())
		},
	}
}

func newRefreshCommand(s *Settings) *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Fetch feeds and update the local post index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a, err := s.openApp(cmd.Context())
			if err != nil {
				return err
			}
			defer func() { _ = a.Close() }()

			stats, err := a.Refresh(cmd.Context())
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "refresh: indexed %d post(s) from %d enabled source(s)\n",
				stats.PostsIndexed, stats.SourcesEnabled-stats.SourcesFailed)
			if stats.SourcesFailed > 0 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "refresh: %d source(s) failed (other enabled sources still indexed)\n", stats.SourcesFailed)
			}
			return nil
		},
	}
}

func newFindCommand(s *Settings) *cobra.Command {
	var jsonOutput bool
	var random bool
	var interactive bool
	cmd := &cobra.Command{
		Use:   "find [query]",
		Short: "Return one indexed post for a query",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive && jsonOutput {
				return fmt.Errorf("find: --interactive cannot be combined with --json")
			}
			a, err := s.openApp(cmd.Context())
			if err != nil {
				return err
			}
			defer func() { _ = a.Close() }()

			query := strings.Join(args, " ")
			if interactive {
				return browseMatches(cmd, a, query, random)
			}
			return printFindResult(cmd, a, query, random, jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit find result as JSON")
	cmd.Flags().BoolVar(&random, "random", false, "Return a random strong match, or shuffle the interactive list")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Browse strong matches with j/k and open the selection in a browser")
	return cmd
}

func browseMatches(cmd *cobra.Command, a *app.App, query string, random bool) error {
	matches, err := a.FindMatches(cmd.Context(), query)
	if err != nil {
		if errors.Is(err, app.ErrNoStrongMatch) {
			return withExitCode(ExitNoMatch, err)
		}
		return err
	}
	if random {
		shuffleMatches(matches, nil)
	}
	return runInteractive(cmd.InOrStdin(), cmd.OutOrStdout(), func(url string) error {
		if err := openBrowser(url); err != nil {
			return err
		}
		return a.MarkOpened(cmd.Context(), url, time.Now())
	}, matches)
}

func printFindResult(cmd *cobra.Command, a *app.App, query string, random, jsonOutput bool) error {
	var match rank.MatchResult
	var err error
	if random {
		match, err = a.FindRandom(cmd.Context(), query)
	} else {
		match, err = a.Find(cmd.Context(), query)
	}
	if err != nil {
		if errors.Is(err, app.ErrNoStrongMatch) {
			return withExitCode(ExitNoMatch, err)
		}
		return err
	}
	if jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetEscapeHTML(false)
		return enc.Encode(a.FindJSON(match))
	}
	_, _ = fmt.Fprint(cmd.OutOrStdout(), formatFindHuman(match))
	return nil
}

func formatSourceTimestamp(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.UTC().Format(time.RFC3339)
}

func sanitizeTabField(s string) string {
	if s == "" {
		return "-"
	}
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func formatFindHuman(p rank.MatchResult) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "Title: %s\n", p.Title)
	_, _ = fmt.Fprintf(&b, "URL: %s\n", p.URL)
	_, _ = fmt.Fprintf(&b, "Source: %s\n", p.SourceName)
	if p.PublishedAt != nil {
		_, _ = fmt.Fprintf(&b, "Published: %s\n", p.PublishedAt.UTC().Format("2006-01-02"))
	} else {
		_, _ = fmt.Fprintln(&b, "Published: unknown")
	}
	if p.Explanation != "" {
		_, _ = fmt.Fprintf(&b, "Match: %s\n", p.Explanation)
	}
	return b.String()
}

func newSourcesCommand(s *Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Inspect configured feed sources",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List sources from the manifest and index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a, err := s.openApp(cmd.Context())
			if err != nil {
				return err
			}
			defer func() { _ = a.Close() }()

			sources, err := a.ListSources(cmd.Context())
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tENABLED\tLAST_SUCCESS\tLAST_ERROR\tFEED_URL")
			for _, src := range sources {
				_, _ = fmt.Fprintf(w, "%s\t%v\t%s\t%s\t%s\n",
					src.Name,
					src.Enabled,
					formatSourceTimestamp(src.LastSuccessAt),
					sanitizeTabField(src.LastError),
					src.FeedURL,
				)
			}
			return w.Flush()
		},
	})
	return cmd
}

func Run(args []string) int {
	settings, err := defaultSettings()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	root := NewRootCommand(settings)
	root.SetIn(os.Stdin)
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCodeFor(err)
	}
	return ExitSuccess
}

func RunWithSettings(s Settings, args []string, stdout, stderr io.Writer) int {
	root := NewRootCommand(s)
	if stdout != nil {
		root.SetOut(stdout)
	}
	if stderr != nil {
		root.SetErr(stderr)
	}
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if stderr != nil {
			_, _ = fmt.Fprintln(stderr, err)
		}
		return exitCodeFor(err)
	}
	return ExitSuccess
}
