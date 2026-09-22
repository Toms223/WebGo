package cmd

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/spf13/cobra"
)

func Generate(dir string) error {
	return app.GenerateStaticWebsite(dir, Handler())
}

func newStaticCommand() *cobra.Command {
	var output string
	command := &cobra.Command{
		Use:   "static",
		Short: "Generate the static documentation site",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Generate(output)
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "dist", "directory the generated site is written to")
	return command
}
