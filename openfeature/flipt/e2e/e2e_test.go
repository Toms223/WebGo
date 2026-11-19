//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Toms223/WebGo/openfeature/flipt"
	"github.com/chromedp/chromedp"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

const fliptImage = "ghcr.io/flipt-io/flipt:v2.14.1"

const canonJS = `(() => {
  const c = (v) => Array.isArray(v) ? v.map(c)
    : v !== null && typeof v === "object" ? Object.fromEntries(Object.keys(v).sort().map((k) => [k, c(v[k])]))
    : v;
  if (window.parityError) throw new Error(window.parityError);
  return window.parityResult.map((r) => JSON.stringify(c(r))).join("\n");
})()`

func TestFliptParity(t *testing.T) {
	cli := containerCLI(t)
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatal("e2e: git not found in PATH; it is needed to commit the flag fixture for Flipt's local storage")
	}
	browser := startBrowser(t)
	fliptURL := startFlipt(t, cli)
	js := evaluate(t, browser, serveJS(t)+"/?flipt="+url.QueryEscape(fliptURL))
	goPage := evaluate(t, browser, serveGo(t, fliptURL)+"/")
	if !strings.Contains(js, `"value":"variant1"`) {
		t.Fatalf("page A never resolved flag1 to variant1, so Flipt did not serve the fixture:\n%s", js)
	}
	if js == goPage {
		t.Logf("JS and Go output byte-identical (%d bytes):\n%s", len(js), js)
		return
	}
	a, b := strings.Split(js, "\n"), strings.Split(goPage, "\n")
	for i := range max(len(a), len(b)) {
		var x, y string
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			t.Fatalf("record %d differs\njs: %s\ngo: %s", i, x, y)
		}
	}
}

func startBrowser(t *testing.T) context.Context {
	opts := chromedp.DefaultExecAllocatorOptions[:]
	if os.Geteuid() == 0 {
		opts = append(opts, chromedp.NoSandbox)
	}
	alloc, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	t.Cleanup(cancel)
	ctx, cancel := chromedp.NewContext(alloc)
	t.Cleanup(cancel)
	if err := chromedp.Do(ctx); err != nil {
		t.Fatalf("e2e: headless Chrome did not start; install Google Chrome or Chromium: %v", err)
	}
	return ctx
}

func containerCLI(t *testing.T) string {
	cli := os.Getenv("E2E_CONTAINER_CLI")
	if cli == "" {
		for _, c := range []string{"docker", "podman"} {
			if _, err := exec.LookPath(c); err == nil {
				cli = c
				break
			}
		}
	}
	if cli == "" {
		t.Fatalf("e2e: neither docker nor podman is in PATH (or set E2E_CONTAINER_CLI); this test runs Flipt from %s in a container", fliptImage)
	}
	if _, err := exec.LookPath(cli); err != nil {
		t.Fatalf("e2e: container CLI %q not found in PATH: %v", cli, err)
	}
	if out, err := exec.Command(cli, "info").CombinedOutput(); err != nil {
		t.Fatalf("e2e: %s cannot reach its engine (%s info: %v); start Docker or the podman machine\n%s", cli, cli, err, out)
	}
	return cli
}

func startFlipt(t *testing.T, cli string) string {
	repo, err := os.MkdirTemp("", "flipt-flags-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(repo) })
	if err := os.CopyFS(repo, os.DirFS("testdata/flags")); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "init", "-q", "-b", "main")
	git(t, repo, "add", "-A")
	git(t, repo, "-c", "user.name=e2e", "-c", "user.email=e2e@localhost", "commit", "-q", "-m", "flags")
	if err := filepath.WalkDir(repo, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		mode := fs.FileMode(0o666)
		if d.IsDir() {
			mode = 0o777
		}
		return os.Chmod(p, mode)
	}); err != nil {
		t.Fatal(err)
	}
	run(t, "", cli, "pull", fliptImage)
	created := strings.Split(strings.TrimSpace(run(t, "", cli, "create", "-p", "127.0.0.1::8080",
		"-e", "FLIPT_META_CHECK_FOR_UPDATES=false",
		"-e", "FLIPT_META_TELEMETRY_ENABLED=false",
		"-e", "FLIPT_STORAGE_DEFAULT_BACKEND_TYPE=local",
		"-e", "FLIPT_STORAGE_DEFAULT_BACKEND_PATH=/flags",
		"-e", "FLIPT_STORAGE_DEFAULT_BRANCH=main",
		"-e", "FLIPT_CORS_ENABLED=true",
		fliptImage)), "\n")
	id := strings.TrimSpace(created[len(created)-1])
	t.Cleanup(func() {
		if t.Failed() {
			logs, _ := exec.Command(cli, "logs", id).CombinedOutput()
			t.Logf("flipt container logs:\n%s", logs)
		}
		exec.Command(cli, "rm", "-f", id).Run()
	})
	run(t, "", cli, "cp", repo, id+":/flags")
	run(t, "", cli, "start", id)
	addr, _, _ := strings.Cut(strings.TrimSpace(run(t, "", cli, "port", id, "8080/tcp")), "\n")
	if _, after, ok := strings.Cut(addr, "-> "); ok {
		addr = after
	}
	base := "http://" + addr
	err = waitFor(2*time.Minute, func() bool {
		req, err := http.NewRequest(http.MethodGet, base+"/internal/v1/evaluation/snapshot/namespace/default", nil)
		if err != nil {
			return false
		}
		req.Header.Set("x-flipt-accept-server-version", "1.47.0")
		req.Header.Set("x-flipt-environment", "default")
		res, err := httpClient.Do(req)
		if err != nil {
			return false
		}
		defer res.Body.Close()
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		return res.StatusCode == http.StatusOK && strings.Contains(b.String(), "flag1")
	})
	if err != nil {
		t.Fatalf("e2e: Flipt did not serve the default namespace snapshot at %s: %v", base, err)
	}
	return base
}

func serveJS(t *testing.T) string {
	page, err := os.ReadFile("testdata/js/index.html")
	if err != nil {
		t.Fatal(err)
	}
	imports := map[string]string{}
	for _, m := range flipt.Modules() {
		name, u := m.Source().ImportMap()
		imports[name] = u
	}
	importMap, err := json.Marshal(map[string]any{"imports": imports})
	if err != nil {
		t.Fatal(err)
	}
	page = bytes.Replace(page, []byte("__IMPORTMAP__"), importMap, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func serveGo(t *testing.T, fliptURL string) string {
	dir := t.TempDir()
	build(t, filepath.Join(dir, "web", "app.wasm"), "GOOS=js", "GOARCH=wasm")
	build(t, filepath.Join(dir, "server"))
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()
	var logs bytes.Buffer
	cmd := exec.Command(filepath.Join(dir, "server"))
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "ADDR="+addr, "FLIPT_URL="+fliptURL)
	cmd.Stdout, cmd.Stderr = &logs, &logs
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	base := "http://" + addr
	if err := waitFor(30*time.Second, func() bool {
		res, err := httpClient.Get(base + "/")
		if err == nil {
			res.Body.Close()
		}
		return err == nil
	}); err != nil {
		t.Fatalf("e2e: page B server did not start: %v\n%s", err, logs.String())
	}
	return base
}

func build(t *testing.T, out string, env ...string) {
	cmd := exec.Command("go", "build", "-o", out, "./testdata/go")
	cmd.Env = append(os.Environ(), env...)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building page B (%v): %v\n%s", env, err, b)
	}
}

func evaluate(t *testing.T, browser context.Context, u string) string {
	ctx, cancel := context.WithTimeout(browser, 3*time.Minute)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	if err := chromedp.Do(ctx, chromedp.Navigate(u)); err != nil {
		t.Fatalf("loading %s: %v", u, err)
	}
	if _, err := chromedp.Run(ctx, chromedp.Poll[bool]("window.parityDone === true", chromedp.WithPollingTimeout(2*time.Minute))); err != nil {
		t.Fatalf("%s never finished the matrix: %v", u, err)
	}
	out, err := chromedp.Run(ctx, chromedp.Evaluate[string](canonJS))
	if err != nil {
		t.Fatalf("%s: %v", u, err)
	}
	return out
}

func git(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func run(t *testing.T, dir string, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s%s", name, strings.Join(args, " "), err, out, stderr.Bytes())
	}
	return string(out)
}

func waitFor(d time.Duration, ok func() bool) error {
	for end := time.Now().Add(d); time.Now().Before(end); time.Sleep(500 * time.Millisecond) {
		if ok() {
			return nil
		}
	}
	return fmt.Errorf("timed out after %s", d)
}
