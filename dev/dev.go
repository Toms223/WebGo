//go:build !wasm

package dev

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"time"

	"github.com/sgtdi/fswatcher/v2"
)

// RunDevServer serves the site built from the working directory at addr and
// rebuilds it into tempDir whenever a Go file, go.mod, go.sum or a web/ asset
// changes. generate returns the command that writes the static site into dir
// with its web directory at dir/web. It returns when ctx is done.
func RunDevServer(ctx context.Context, tempDir, addr string, generate func(dir string) *exec.Cmd) error {
	srcDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if tempDir, err = filepath.Abs(tempDir); err != nil {
		return err
	}
	if err := rebuild(ctx, srcDir, tempDir, generate); err != nil {
		return err
	}

	watcher, err := fswatcher.New(
		fswatcher.WithPath(srcDir),
		fswatcher.WithCooldown(time.Second),
		fswatcher.WithIncRegex(`\.go$`, `go\.(mod|sum)$`, `/web/`),
		fswatcher.WithExcRegex("^"+regexp.QuoteMeta(realPath(tempDir))),
	)
	if err != nil {
		return err
	}
	watchErr := make(chan error, 1)
	go func() { watchErr <- watcher.Watch(ctx) }()

	srv := &http.Server{Addr: addr, Handler: fileServer(tempDir)}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()
	log.Printf("dev server on %s", addr)

	for {
		select {
		case <-ctx.Done():
			return srv.Shutdown(context.Background())
		case err := <-serveErr:
			return err
		case err := <-watchErr:
			srv.Close()
			return err
		case ev := <-watcher.Events():
			for len(watcher.Events()) > 0 {
				<-watcher.Events()
			}
			log.Printf("rebuilding: %s changed", ev.Path)
			if err := rebuild(ctx, srcDir, tempDir, generate); err != nil {
				log.Printf("rebuild failed, serving last good build: %v", err)
				continue
			}
			log.Print("rebuilt")
		}
	}
}

func rebuild(ctx context.Context, srcDir, tempDir string, generate func(dir string) *exec.Cmd) error {
	stage := tempDir + ".next"
	if err := os.RemoveAll(stage); err != nil {
		return err
	}
	webDir := filepath.Join(stage, "web")
	if err := os.MkdirAll(webDir, 0o755); err != nil {
		return err
	}
	if err := os.CopyFS(webDir, os.DirFS(filepath.Join(srcDir, "web"))); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	// The wasm must exist before generating, since the handler hashes it into its version.
	if err := run(exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(webDir, "app.wasm"), "."), srcDir, "GOOS=js", "GOARCH=wasm"); err != nil {
		return err
	}
	if err := run(generate(stage), srcDir); err != nil {
		return err
	}

	old := tempDir + ".old"
	os.RemoveAll(old)
	if err := os.Rename(tempDir, old); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.Rename(stage, tempDir); err != nil {
		return err
	}
	return os.RemoveAll(old)
}

func run(cmd *exec.Cmd, dir string, env ...string) error {
	if cmd.Dir == "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func fileServer(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		clean := path.Clean("/" + r.URL.Path)
		p := filepath.Join(dir, filepath.FromSlash(clean))
		if path.Ext(clean) == "" {
			if _, err := os.Stat(p + ".html"); err == nil {
				r.URL.Path = clean + ".html"
			} else if info, err := os.Stat(p); err != nil || info.IsDir() {
				r.URL.Path = "/"
			}
		}
		files.ServeHTTP(w, r)
	})
}

func realPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}
