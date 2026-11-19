//go:build !wasm

package dev

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func site(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), "module site\n\ngo 1.21\n")
	write(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")
	write(t, filepath.Join(dir, "web", "app.css"), "html {}\n")
	write(t, filepath.Join(dir, "page.txt"), "one")
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "-mod=mod")
	return dir
}

func copyPage(srcDir string) func(dir string) *exec.Cmd {
	return func(dir string) *exec.Cmd {
		return exec.Command("cp", filepath.Join(srcDir, "page.txt"), filepath.Join(dir, "index.html"))
	}
}

func get(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
}

func TestFileServer(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "index.html"), "home")
	write(t, filepath.Join(dir, "about.html"), "about")
	write(t, filepath.Join(dir, "news", "first.html"), "first")
	write(t, filepath.Join(dir, "news.html"), "news")
	write(t, filepath.Join(dir, "web", "app.css"), "css")
	write(t, filepath.Join(filepath.Dir(dir), "secret"), "secret")
	h := fileServer(dir)

	for target, want := range map[string]string{
		"/":                 "home",
		"/about":            "about",
		"/news/first":       "first",
		"/news":             "news",
		"/web":              "home",
		"/web/app.css":      "css",
		"/not/a/page":       "home",
		"/../secret":        "home",
		"/web/../about":     "about",
		"/about?x=1#anchor": "about",
	} {
		w := get(t, h, target)
		if w.Code != http.StatusOK || w.Body.String() != want {
			t.Errorf("GET %s = %d %q, want 200 %q", target, w.Code, w.Body.String(), want)
		}
		if got := w.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("GET %s Cache-Control = %q, want no-store", target, got)
		}
	}
}

func TestFileServerAnswersNotFoundForMissingAssets(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "index.html"), "home")
	h := fileServer(dir)
	for _, target := range []string{"/web/app.wasm", "/web/missing.css", "/app.js", "/data/page.json"} {
		if w := get(t, h, target); w.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, w.Code)
		}
	}
}

func TestRebuild(t *testing.T) {
	src := site(t)
	out := filepath.Join(t.TempDir(), "out")

	if err := rebuild(context.Background(), src, out, copyPage(src)); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if got := read(t, filepath.Join(out, "index.html")); got != "one" {
		t.Errorf("index.html = %q, want one", got)
	}
	if got := read(t, filepath.Join(out, "web", "app.css")); got != "html {}\n" {
		t.Errorf("web/app.css = %q", got)
	}
	wasm := read(t, filepath.Join(out, "web", "app.wasm"))
	if !strings.HasPrefix(wasm, "\x00asm") {
		t.Error("web/app.wasm is not a WebAssembly binary")
	}

	write(t, filepath.Join(src, "page.txt"), "two")
	if err := rebuild(context.Background(), src, out, copyPage(src)); err != nil {
		t.Fatalf("second rebuild: %v", err)
	}
	if got := read(t, filepath.Join(out, "index.html")); got != "two" {
		t.Errorf("index.html after a second rebuild = %q, want two", got)
	}
	for _, leftover := range []string{out + ".next", out + ".old"} {
		if _, err := os.Stat(leftover); err == nil {
			t.Errorf("%s was left behind", leftover)
		}
	}
}

func TestRebuildKeepsTheLastGoodBuild(t *testing.T) {
	src := site(t)
	out := filepath.Join(t.TempDir(), "out")
	if err := rebuild(context.Background(), src, out, copyPage(src)); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	failing := func(string) *exec.Cmd { return exec.Command("false") }
	if err := rebuild(context.Background(), src, out, failing); err == nil {
		t.Error("rebuild with a failing generator returned nil")
	}
	write(t, filepath.Join(src, "main.go"), "package main\n\nfunc main() { broken }\n")
	if err := rebuild(context.Background(), src, out, copyPage(src)); err == nil {
		t.Error("rebuild of code that does not compile returned nil")
	}
	if got := read(t, filepath.Join(out, "index.html")); got != "one" {
		t.Errorf("index.html after failed rebuilds = %q, want one", got)
	}
	if _, err := os.Stat(filepath.Join(out, "web", "app.wasm")); err != nil {
		t.Errorf("the last good app.wasm is gone: %v", err)
	}
}

func TestRebuildWithoutAWebDirectory(t *testing.T) {
	src := site(t)
	if err := os.RemoveAll(filepath.Join(src, "web")); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if err := rebuild(context.Background(), src, out, copyPage(src)); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "web", "app.wasm")); err != nil {
		t.Errorf("app.wasm missing: %v", err)
	}
}

func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().String()
}

func body(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestRunDevServer(t *testing.T) {
	src := site(t)
	t.Chdir(src)
	for name, tempDir := range map[string]string{
		"outside the source": filepath.Join(t.TempDir(), "out"),
		"inside the source":  ".dev",
	} {
		t.Run(name, func(t *testing.T) {
			write(t, filepath.Join(src, "page.txt"), "one")
			write(t, filepath.Join(src, "main.go"), "package main\n\nfunc main() {}\n")
			write(t, filepath.Join(src, "web", "app.css"), "html {}\n")
			addr := freeAddr(t)
			url := "http://" + addr
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- RunDevServer(ctx, tempDir, addr, copyPage(src)) }()
			t.Cleanup(cancel)

			eventually(t, "the first build to be served", func() bool { return body(url+"/") == "one" })
			if got := body(url + "/any/route"); got != "one" {
				t.Errorf("an unknown route served %q, want the index", got)
			}
			if got := body(url + "/web/app.css"); got != "html {}\n" {
				t.Errorf("/web/app.css served %q", got)
			}
			time.Sleep(1500 * time.Millisecond)

			write(t, filepath.Join(src, "page.txt"), "two")
			write(t, filepath.Join(src, "main.go"), "package main\n\nfunc main() { println() }\n")
			eventually(t, "a Go change to rebuild", func() bool { return body(url+"/") == "two" })
			time.Sleep(1500 * time.Millisecond)

			write(t, filepath.Join(src, "page.txt"), "three")
			write(t, filepath.Join(src, "web", "app.css"), "body {}\n")
			eventually(t, "a web asset change to rebuild", func() bool {
				return body(url+"/") == "three" && body(url+"/web/app.css") == "body {}\n"
			})
			time.Sleep(1500 * time.Millisecond)

			write(t, filepath.Join(src, "main.go"), "package main\n\nfunc main() { broken }\n")
			time.Sleep(3 * time.Second)
			if got := body(url + "/"); got != "three" {
				t.Errorf("after a broken build the server answered %q, want the last good build", got)
			}

			cancel()
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("RunDevServer returned %v after cancel, want nil", err)
				}
			case <-time.After(15 * time.Second):
				t.Fatal("RunDevServer did not return after cancel")
			}
		})
	}
}

func TestRunDevServerFailsWhenTheFirstBuildFails(t *testing.T) {
	src := site(t)
	t.Chdir(src)
	write(t, filepath.Join(src, "main.go"), "package main\n\nfunc main() { broken }\n")
	err := RunDevServer(context.Background(), filepath.Join(t.TempDir(), "out"), freeAddr(t), copyPage(src))
	if err == nil {
		t.Fatal("RunDevServer returned nil for code that does not compile")
	}
}
