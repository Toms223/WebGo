package serve

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	defaultWebDir = "web"
	webPrefix     = "/web/"
	cssExtension  = ".css"
)

type Handler struct {
	app.Handler
	Styles []style.Style
	WebDir string

	base  []string
	built bool
}

type buildPlan struct {
	links    []string
	required []string
	outputs  []generatedFile
}

type generatedFile struct {
	file    string
	content []byte
}

func (h *Handler) Build() (*app.Handler, error) {
	plan, err := h.resolve(cmp.Or(h.WebDir, defaultWebDir))
	if err != nil {
		return nil, err
	}
	if err := plan.apply(); err != nil {
		return nil, err
	}
	if !h.built {
		h.base = slices.Clone(h.Handler.Styles)
		h.built = true
	}
	h.Handler.Styles = append(plan.links, h.base...)
	return &h.Handler, nil
}

func (h *Handler) resolve(webDir string) (*buildPlan, error) {
	plan := &buildPlan{}
	for _, s := range h.Styles {
		source, err := style.Resolve(s)
		if err != nil {
			return nil, err
		}
		if err := plan.add(webDir, source); err != nil {
			return nil, err
		}
	}
	return plan, nil
}

func (b *buildPlan) add(webDir string, source style.Source) error {
	switch source.Kind {
	case style.StaticStyle:
		return b.addStatic(webDir, source.Path)
	case style.GeneratedStyle:
		return b.addGenerated(webDir, source.Path, source.Content)
	}
	return nil
}

func (b *buildPlan) addStatic(webDir, p string) error {
	if isURL(p) {
		return b.claim(p)
	}
	file, err := webPath(webDir, p)
	if err != nil {
		return err
	}
	if err := b.claim(path.Clean(p)); err != nil {
		return err
	}
	b.required = append(b.required, file)
	return nil
}

func (b *buildPlan) addGenerated(webDir, p string, content []byte) error {
	file, err := webPath(webDir, p)
	if err != nil {
		return err
	}
	link := path.Clean(p)
	if !strings.HasSuffix(link, cssExtension) {
		return withSubject(ErrNotCSSPath, p)
	}
	if err := b.claim(link); err != nil {
		return err
	}
	b.outputs = append(b.outputs, generatedFile{file: file, content: content})
	return nil
}

func (b *buildPlan) claim(link string) error {
	if slices.Contains(b.links, link) {
		return withSubject(ErrDuplicatePath, link)
	}
	b.links = append(b.links, link)
	return nil
}

func (b *buildPlan) apply() error {
	for _, file := range b.required {
		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("%w: %w", ErrMissingStyle, err)
		}
	}
	for _, out := range b.outputs {
		if err := writeIfChanged(out.file, out.content); err != nil {
			return fmt.Errorf("serve: %w", err)
		}
	}
	return nil
}

func withSubject(sentinel error, subject string) error {
	return fmt.Errorf("%w: %s", sentinel, subject)
}

func webPath(webDir, p string) (string, error) {
	if strings.HasPrefix(p, webPrefix) {
		if rest, ok := strings.CutPrefix(path.Clean(p), webPrefix); ok && filepath.IsLocal(rest) {
			return filepath.Join(webDir, rest), nil
		}
	}
	return "", withSubject(ErrPathOutsideWeb, p)
}

func isURL(p string) bool {
	return strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") || strings.HasPrefix(p, "//")
}

func writeIfChanged(file string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	if current, err := os.ReadFile(file); err == nil && bytes.Equal(current, content) {
		return nil
	}
	return os.WriteFile(file, content, 0o644)
}
