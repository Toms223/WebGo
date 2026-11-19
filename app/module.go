package app

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"os"
)

// ModuleSource names an ES module and says where it is served from: a /web/
// path or an http://, https:// or // URL.
type ModuleSource interface {
	ImportMap() (name, url string)
}

// Module is an entry of Handler.Imports, usually a script.Script[T].
type Module interface {
	Source() ModuleSource
}

type moduleChecker interface {
	Check() error
}

var (
	// ErrMissingModule means a local module file is missing under
	// Handler.WebDir.
	ErrMissingModule = stderrors.New("app: module file not found")

	// ErrDuplicateModule means two modules share a name.
	ErrDuplicateModule = stderrors.New("app: module name used twice")

	// ErrInvalidModule means a module has no source or name, or its source
	// failed its own check.
	ErrInvalidModule = stderrors.New("app: module source is invalid")
)

type importEntry struct{ name, url string }

func resolveImports(imports []Module, webDir string) ([]importEntry, error) {
	var entries []importEntry
	seen := map[string]bool{}
	for _, m := range imports {
		if m == nil {
			continue
		}
		src := m.Source()
		if src == nil {
			return nil, withSubject(ErrInvalidModule, "nil source")
		}
		name, u := src.ImportMap()
		if name == "" {
			return nil, withSubject(ErrInvalidModule, u)
		}
		if c, ok := src.(moduleChecker); ok {
			if err := c.Check(); err != nil {
				if stderrors.Is(err, ErrInvalidModule) {
					return nil, err
				}
				return nil, fmt.Errorf("%w: %s: %w", ErrInvalidModule, name, err)
			}
		}
		if !isURL(u) {
			file, err := webPath(webDir, u)
			if err != nil {
				return nil, err
			}
			if _, err := os.Stat(file); err != nil {
				return nil, fmt.Errorf("%w: %s: %w", ErrMissingModule, u, err)
			}
		}
		if seen[name] {
			return nil, withSubject(ErrDuplicateModule, name)
		}
		seen[name] = true
		entries = append(entries, importEntry{name: name, url: u})
	}
	return entries, nil
}

func (h *Handler) importMapJSON() string {
	imports := make(map[string]string, len(h.imports))
	for _, e := range h.imports {
		imports[e.name] = h.Resources.Resolve(e.url)
	}
	b, _ := json.Marshal(struct {
		Imports map[string]string `json:"imports"`
	}{imports})
	return string(b)
}

func (h *Handler) importURLs() []string {
	urls := make([]string, len(h.imports))
	for i, e := range h.imports {
		urls[i] = e.url
	}
	return urls
}
