package cmd

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/Toms223/WebGo/dev"
	"github.com/spf13/cobra"
)

func newDevCommand() *cobra.Command {
	var addr string
	var output string
	command := &cobra.Command{
		Use:   "dev",
		Short: "Serve the documentation site locally and rebuild it when a file changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()
			return dev.RunDevServer(ctx, output, addr, func(dir string) *exec.Cmd {
				return exec.Command("go", "run", ".", "static", "-o", dir, "--web", filepath.Join(dir, "web"))
			})
		},
	}
	command.Flags().StringVar(&addr, "addr", "localhost:8000", "address the site is served at")
	command.Flags().StringVarP(&output, "output", "o", filepath.Join(os.TempDir(), "webgo-docs-dev"), "directory the site is built into")
	return command
}
