package app

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// StyleKind tells Build how to handle a StyleSource.
type StyleKind int

const (
	// StaticStyle links an existing file under /web/, or an http://, https://
	// or // URL, which skips the file check.
	StaticStyle StyleKind = iota + 1

	// GeneratedStyle writes Content to Path, a .css path under /web/, and links
	// it.
	GeneratedStyle

	// DynamicStyle is ignored by Build: no file and no link. Its CSS is applied
	// at run time.
	DynamicStyle
)

// StyleSource is what a Stylesheet resolves to: a Kind, a /web/ path or URL,
// and the CSS of a generated stylesheet.
type StyleSource struct {
	Kind    StyleKind
	Path    string
	Content []byte
}

// Stylesheet is an entry of Handler.Styles. The style package's Static,
// Generated and Dynamic return one; implement it to supply CSS from elsewhere.
type Stylesheet interface {
	StyleSource() (StyleSource, error)
}

var (
	// ErrMissingStyle means a static stylesheet is missing under
	// Handler.WebDir.
	ErrMissingStyle = stderrors.New("app: stylesheet file not found")

	// ErrPathOutsideWeb means a local stylesheet path is not under /web/.
	ErrPathOutsideWeb = stderrors.New("app: path is not under /web/")

	// ErrNotCSSPath means a generated stylesheet path does not end in .css.
	ErrNotCSSPath = stderrors.New("app: generated path must end in .css")

	// ErrDuplicatePath means two stylesheets resolve to the same path or URL.
	ErrDuplicatePath = stderrors.New("app: stylesheet path used twice")
)

const (
	defaultWebDir = "web"
	webPrefix     = "/web/"
	cssExtension  = ".css"
)

// Build resolves Styles: it checks that static files exist under WebDir and
// writes generated stylesheets there, skipping unchanged files. It is optional,
// since ServeHTTP and GenerateStaticWebsite run it, but calling it at startup
// reports errors before the first request. Errors wrap ErrMissingStyle,
// ErrPathOutsideWeb, ErrNotCSSPath or ErrDuplicatePath, so match them with
// errors.Is.
func (h *Handler) Build() error {
	plan, err := resolveStyles(h.Styles, h.webDir())
	if err != nil {
		return err
	}
	if err := plan.apply(); err != nil {
		return err
	}
	h.styleLinks = plan.links
	h.built = true
	return nil
}

func (h *Handler) webDir() string {
	if h.WebDir != "" {
		return h.WebDir
	}
	if r, ok := h.Resources.(localResourceResolver); ok && r.directory != "" {
		return filepath.Join(r.directory, defaultWebDir)
	}
	return defaultWebDir
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

func resolveStyles(styles []Stylesheet, webDir string) (*buildPlan, error) {
	plan := &buildPlan{}
	for _, s := range styles {
		if s == nil {
			continue
		}
		source, err := s.StyleSource()
		if err != nil {
			return nil, err
		}
		if err := plan.add(webDir, source); err != nil {
			return nil, err
		}
	}
	return plan, nil
}

func (b *buildPlan) add(webDir string, source StyleSource) error {
	switch source.Kind {
	case StaticStyle:
		return b.addStatic(webDir, source.Path)
	case GeneratedStyle:
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
			return fmt.Errorf("app: %w", err)
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
