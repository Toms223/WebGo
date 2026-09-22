package cmd

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/spf13/cobra"
)

func Generate(dir, repository string) error {
	return app.GenerateStaticWebsite(dir, Handler(repository))
}

func newStaticCommand() *cobra.Command {
	var output string
	var repository string
	command := &cobra.Command{
		Use:   "static",
		Short: "Generate the static documentation site",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Generate(output, repository)
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "dist", "directory the generated site is written to")
	command.Flags().StringVar(&repository, "github-pages", "", "repository name to serve the site from a GitHub Pages subpath, for example WebGo")
	return command
}
