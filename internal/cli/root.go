package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "blerg",
		Short: "Find a single engineering blog post for a topic",
	}
	root.AddCommand(newRefreshCommand())
	root.AddCommand(newFindCommand())
	root.AddCommand(newSourcesCommand())
	return root
}

func newRefreshCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Fetch feeds and update the local post index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("refresh: not implemented yet (see issue #3)")
		},
	}
}

func newFindCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "find [query]",
		Short: "Return the single best-matching indexed post for a query",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("find: not implemented yet (see issue #4)")
		},
	}
}

func newSourcesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Inspect configured feed sources",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List sources from the manifest and index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("sources list: not implemented yet (see issue #6)")
		},
	})
	return cmd
}

func Main() int {
	root := NewRootCommand()
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
