package cmd

import (
	"github.com/Toms223/WebGo/app"
	"github.com/spf13/cobra"
)

func Generate(dir, repository, webDir string) error {
	handler, err := Handler(repository, webDir)
	if err != nil {
		return err
	}
	return app.GenerateStaticWebsite(dir, handler)
}

func newStaticCommand() *cobra.Command {
	var output string
	var repository string
	var webDir string
	command := &cobra.Command{
		Use:   "static",
		Short: "Generate the static documentation site",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Generate(output, repository, webDir)
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "dist", "directory the generated site is written to")
	command.Flags().StringVar(&repository, "github-pages", "", "repository name to serve the site from a GitHub Pages subpath, for example WebGo")
	command.Flags().StringVar(&webDir, "web", "web", "directory served at /web/: app.css is read from it and theme.css is written into it")
	return command
}
