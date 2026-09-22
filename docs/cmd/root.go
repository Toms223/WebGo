package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "docs",
		Short:         "WebGo documentation site",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.AddCommand(newStaticCommand())
	return root
}

func Execute() {
	if err := newRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
