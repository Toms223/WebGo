package npm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/fs"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
)

type entry struct {
	name     string
	body     string
	typeflag byte
	linkname string
}

func tarball(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		typeflag := e.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		header := &tar.Header{Name: e.name, Typeflag: typeflag, Mode: 0o644, Linkname: e.linkname}
		if typeflag == tar.TypeReg {
			header.Size = int64(len(e.body))
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

type requests struct {
	mu    sync.Mutex
	paths []string
}

func (r *requests) add(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paths = append(r.paths, path)
}

func (r *requests) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.paths)
}

func serve(t *testing.T, respond http.HandlerFunc) *requests {
	t.Helper()
	seen := &requests{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.add(r.URL.Path)
		respond(w, r)
	}))
	previous := Registry
	Registry = server.URL
	t.Cleanup(func() {
		Registry = previous
		server.Close()
	})
	return seen
}

func respondWith(body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read %s: %v", path, err)
		return
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}

func entriesUnder(t *testing.T, root string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if rel != "." {
			found = append(found, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func TestFetchSkipsWhenCached(t *testing.T) {
	cache := t.TempDir()
	cached := filepath.Join(cache, "@webref_css", "1.0.0")
	if err := os.MkdirAll(cached, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cached, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	seen := serve(t, http.NotFound)

	got, err := Fetch(cache, "@webref/css", "1.0.0")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if got != cached {
		t.Errorf("Fetch = %q, want %q", got, cached)
	}
	if paths := seen.all(); len(paths) != 0 {
		t.Errorf("registry received requests %q, want none", paths)
	}
}

func TestFetchExtracts(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "not", "created", "yet")
	archive := tarball(t,
		entry{name: "package/a.json", body: `{"a":1}`},
		entry{name: "package/lib/b.json", body: `{"b":2}`},
		entry{name: "package/empty/", typeflag: tar.TypeDir},
		entry{name: "other/c.json", body: `{"c":3}`},
	)
	seen := serve(t, respondWith(archive))

	dir, err := Fetch(cache, "@webref/css", "1.0.0")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if want := filepath.Join(cache, "@webref_css", "1.0.0"); dir != want {
		t.Errorf("Fetch = %q, want %q", dir, want)
	}
	assertFile(t, filepath.Join(dir, "a.json"), `{"a":1}`)
	assertFile(t, filepath.Join(dir, "lib", "b.json"), `{"b":2}`)
	wantTree := []string{
		"@webref_css",
		"@webref_css/1.0.0",
		"@webref_css/1.0.0/a.json",
		"@webref_css/1.0.0/empty",
		"@webref_css/1.0.0/lib",
		"@webref_css/1.0.0/lib/b.json",
	}
	if got := entriesUnder(t, cache); !slices.Equal(got, wantTree) {
		t.Errorf("entries under the cache = %q, want %q", got, wantTree)
	}

	again, err := Fetch(cache, "@webref/css", "1.0.0")
	if err != nil || again != dir {
		t.Fatalf("second Fetch = %q, %v, want %q", again, err, dir)
	}
	if paths := seen.all(); len(paths) != 1 {
		t.Errorf("registry received %q, want exactly one request", paths)
	}
}

func TestFetchRequestsTarballFromRegistry(t *testing.T) {
	archive := tarball(t, entry{name: "package/a.json", body: "{}"})
	for _, tc := range []struct {
		name, version, path, cacheName string
	}{
		{"@webref/css", "8.7.6", "/@webref/css/-/css-8.7.6.tgz", "@webref_css"},
		{"@mdn/browser-compat-data", "8.1.5", "/@mdn/browser-compat-data/-/browser-compat-data-8.1.5.tgz", "@mdn_browser-compat-data"},
		{"left-pad", "1.3.0", "/left-pad/-/left-pad-1.3.0.tgz", "left-pad"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := t.TempDir()
			seen := serve(t, respondWith(archive))

			dir, err := Fetch(cache, tc.name, tc.version)
			if err != nil {
				t.Fatalf("Fetch: %v", err)
			}
			if want := filepath.Join(cache, tc.cacheName, tc.version); dir != want {
				t.Errorf("Fetch = %q, want %q", dir, want)
			}
			if got := seen.all(); !slices.Equal(got, []string{tc.path}) {
				t.Errorf("registry received %q, want %q", got, tc.path)
			}
		})
	}
}

func TestTarballURLUsesPublicRegistryByDefault(t *testing.T) {
	for _, tc := range []struct{ name, version, want string }{
		{"@webref/css", "8.7.6", "https://registry.npmjs.org/@webref/css/-/css-8.7.6.tgz"},
		{"@mdn/browser-compat-data", "8.1.5", "https://registry.npmjs.org/@mdn/browser-compat-data/-/browser-compat-data-8.1.5.tgz"},
		{"left-pad", "1.3.0", "https://registry.npmjs.org/left-pad/-/left-pad-1.3.0.tgz"},
	} {
		if got := tarballURL(tc.name, tc.version); got != tc.want {
			t.Errorf("tarballURL(%q, %q) = %q, want %q", tc.name, tc.version, got, tc.want)
		}
	}
}

func noise(seed byte, n int) string {
	var key [32]byte
	key[0] = seed
	buf := make([]byte, n)
	rand.NewChaCha8(key).Read(buf)
	return string(buf)
}

func TestFetchFailureLeavesNoPackageBehind(t *testing.T) {
	valid := tarball(t, entry{name: "package/a.json", body: "{}"})
	large := tarball(t,
		entry{name: "package/a.json", body: noise(1, 8192)},
		entry{name: "package/b.json", body: noise(2, 8192)},
		entry{name: "package/c.json", body: noise(3, 8192)},
	)
	badChecksum := slices.Clone(large)
	badChecksum[len(badChecksum)-8] ^= 0xff
	for name, respond := range map[string]http.HandlerFunc{
		"registry error": http.NotFound,
		"registry error with a package body": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write(valid)
		},
		"not a gzip stream":  respondWith([]byte("this is not a gzip stream")),
		"truncated mid-file": respondWith(large[:len(large)*3/4]),
		"checksum mismatch":  respondWith(badChecksum),
		"no package files":   respondWith(tarball(t, entry{name: "other/a.json", body: "{}"})),
	} {
		t.Run(name, func(t *testing.T) {
			cache := t.TempDir()
			serve(t, respond)

			dir, err := Fetch(cache, "@webref/css", "1.0.0")
			if err == nil {
				t.Fatalf("Fetch = %q, want an error", dir)
			}
			if left := entriesUnder(t, cache); len(left) > 0 && !slices.Equal(left, []string{"@webref_css"}) {
				t.Errorf("failed fetch left %q", left)
			}

			serve(t, respondWith(valid))
			dir, err = Fetch(cache, "@webref/css", "1.0.0")
			if err != nil {
				t.Fatalf("Fetch after the registry recovered: %v", err)
			}
			assertFile(t, filepath.Join(dir, "a.json"), "{}")
		})
	}
}

func TestFetchFailsWhenDestinationIsAFile(t *testing.T) {
	cache := t.TempDir()
	dest := filepath.Join(cache, "@webref_css", "1.0.0")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("not a package"), 0o644); err != nil {
		t.Fatal(err)
	}
	serve(t, respondWith(tarball(t, entry{name: "package/a.json", body: "{}"})))

	if dir, err := Fetch(cache, "@webref/css", "1.0.0"); err == nil {
		t.Fatalf("Fetch = %q, want an error", dir)
	}
	want := []string{"@webref_css", "@webref_css/1.0.0"}
	if got := entriesUnder(t, cache); !slices.Equal(got, want) {
		t.Errorf("entries under the cache = %q, want %q", got, want)
	}
	assertFile(t, dest, "not a package")
}

func TestFetchExtractsOnlyRegularFilesUnderPackage(t *testing.T) {
	root := t.TempDir()
	cache := filepath.Join(root, "cache")
	serve(t, respondWith(tarball(t,
		entry{name: "package/a.json", body: "{}"},
		entry{name: "./package/dotted.json", body: "{}"},
		entry{name: "package/../../escape.txt", body: "x"},
		entry{name: "../escape.txt", body: "x"},
		entry{name: "/absolute.txt", body: "x"},
		entry{name: "package/link", typeflag: tar.TypeSymlink, linkname: "/etc/passwd"},
		entry{name: "package/hard", typeflag: tar.TypeLink, linkname: "package/a.json"},
	)))

	if _, err := Fetch(cache, "@webref/css", "1.0.0"); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	want := []string{
		"cache",
		"cache/@webref_css",
		"cache/@webref_css/1.0.0",
		"cache/@webref_css/1.0.0/a.json",
		"cache/@webref_css/1.0.0/dotted.json",
	}
	if got := entriesUnder(t, root); !slices.Equal(got, want) {
		t.Errorf("entries under %s = %q, want %q", root, got, want)
	}
}
