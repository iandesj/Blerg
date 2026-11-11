package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/iandesj/blerg/internal/app"
	"github.com/iandesj/blerg/internal/rank"
	"github.com/spf13/cobra"
)

type Settings struct {
	DBPath      string
	SourcesPath string
	HTTPClient  app.HTTPClient
	UserAgent   string
}

func defaultSettings() (Settings, error) {
	dbPath, err := app.DefaultDBPath()
	if err != nil {
		return Settings{}, err
	}
	return Settings{
		DBPath:      dbPath,
		SourcesPath: "config/sources.yaml",
		HTTPClient:  http.DefaultClient,
		UserAgent:   app.DefaultUserAgent,
	}, nil
}

func (s *Settings) openApp(ctx context.Context) (*app.App, error) {
	return app.New(ctx, app.Options{
		DBPath:      s.DBPath,
		SourcesPath: s.SourcesPath,
		HTTPClient:  s.HTTPClient,
		UserAgent:   s.UserAgent,
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
		Use:   "blerg",
		Short: "Blerg finds a single engineering blog post for a topic",
	}
	root.PersistentFlags().StringVar(&settings.DBPath, "db", settings.DBPath, "SQLite index path")
	root.PersistentFlags().StringVar(&settings.SourcesPath, "sources", settings.SourcesPath, "Sources manifest path")

	root.AddCommand(newRefreshCommand(&settings))
	root.AddCommand(newFindCommand(&settings))
	root.AddCommand(newSourcesCommand(&settings))
	return root
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
	return &cobra.Command{
		Use:   "find [query]",
		Short: "Return the single best-matching indexed post for a query",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := s.openApp(cmd.Context())
			if err != nil {
				return err
			}
			defer func() { _ = a.Close() }()

			match, err := a.Find(cmd.Context(), strings.Join(args, " "))
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), formatFindHuman(match))
			return nil
		},
	}
}

func formatFindHuman(p rank.Post) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "Title: %s\n", p.Title)
	_, _ = fmt.Fprintf(&b, "URL: %s\n", p.URL)
	_, _ = fmt.Fprintf(&b, "Source: %s\n", p.SourceName)
	if p.PublishedAt != nil {
		_, _ = fmt.Fprintf(&b, "Published: %s\n", p.PublishedAt.UTC().Format("2006-01-02"))
	} else {
		_, _ = fmt.Fprintln(&b, "Published: unknown")
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
			_, _ = fmt.Fprintln(w, "NAME\tENABLED\tFEED_URL")
			for _, src := range sources {
				_, _ = fmt.Fprintf(w, "%s\t%v\t%s\n", src.Name, src.Enabled, src.FeedURL)
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
		return 1
	}
	return 0
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
		return 1
	}
	return 0
}
