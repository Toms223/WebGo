package npm

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var Registry = "https://registry.npmjs.org"

const packagePrefix = "package/"

func Fetch(cacheDir, name, version string) (string, error) {
	parent := filepath.Join(cacheDir, strings.ReplaceAll(name, "/", "_"))
	dest := filepath.Join(parent, version)
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		return dest, nil
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	staging, err := os.MkdirTemp(parent, version+".partial-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(staging)
	if err := download(tarballURL(name, version), staging); err != nil {
		return "", fmt.Errorf("npm: %s@%s: %w", name, version, err)
	}
	if err := os.Rename(staging, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func tarballURL(name, version string) string {
	base := name[strings.LastIndex(name, "/")+1:]
	return Registry + "/" + name + "/-/" + base + "-" + version + ".tgz"
}

func download(url, dir string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return extract(resp.Body, dir)
}

func extract(r io.Reader, dir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()
	archive := tar.NewReader(gz)
	files := 0
	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		rel, ok := strings.CutPrefix(path.Clean(header.Name), packagePrefix)
		if !ok {
			continue
		}
		target := filepath.Join(dir, filepath.FromSlash(rel))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := writeFile(target, archive); err != nil {
				return err
			}
			files++
		}
	}
	if _, err := io.Copy(io.Discard, gz); err != nil {
		return err
	}
	if files == 0 {
		return errors.New("tarball holds no files under package/")
	}
	return nil
}

func writeFile(target string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, r); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
