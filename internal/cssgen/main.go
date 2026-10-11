package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Toms223/WebGo/internal/cssgen/bcd"
	"github.com/Toms223/WebGo/internal/cssgen/npm"
	"github.com/Toms223/WebGo/internal/cssgen/pins"
	"github.com/Toms223/WebGo/internal/cssgen/webref"
)

var outputs = []struct {
	path string
	emit func(*model) []byte
}{
	{"css/properties_gen.go", emitCSS},
	{"css/markers_gen.go", emitCSSMarkers},
	{"kw/keywords_gen.go", emitKW},
	{"kw/markers_gen.go", emitKWMarkers},
	{"media/features_gen.go", emitMedia},
	{"style/pseudo_gen.go", emitPseudo},
	{"css/descriptors_gen.go", emitDescriptor},
	{"container/features_gen.go", emitContainer},
}

func main() {
	out := flag.String("out", ".", "repository root that receives the generated files")
	cache := flag.String("cache", defaultCache(), "directory that keeps the downloaded npm packages")
	flag.Parse()
	if err := run(*out, *cache, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func defaultCache() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "webgo-cssgen")
}

func run(out, cache string, stdout io.Writer) error {
	if cache == "" {
		return errors.New("cssgen: no user cache directory, pass -cache")
	}
	m, r, err := load(cache)
	if err != nil {
		return err
	}
	for _, o := range outputs {
		if err := writeFormatted(filepath.Join(out, filepath.FromSlash(o.path)), o.emit(m)); err != nil {
			return err
		}
	}
	printReport(stdout, m, r)
	return nil
}

func load(cache string) (*model, report, error) {
	webrefDir, err := npm.Fetch(cache, "@webref/css", pins.WebrefCSSVersion)
	if err != nil {
		return nil, report{}, err
	}
	bcdDir, err := npm.Fetch(cache, "@mdn/browser-compat-data", pins.BCDVersion)
	if err != nil {
		return nil, report{}, err
	}
	w, err := webref.Load(webrefDir)
	if err != nil {
		return nil, report{}, err
	}
	b, err := bcd.Load(bcdDir)
	if err != nil {
		return nil, report{}, err
	}
	return build(w, b, pins.FloorDate())
}

func printReport(w io.Writer, m *model, r report) {
	fmt.Fprintf(w, "properties: %d (tier 1: %d, tier 2: %d)\n", len(m.properties), r.tier1, r.tier2)
	for _, browser := range slices.Sorted(maps.Keys(r.droppedProps)) {
		fmt.Fprintf(w, "dropped for %s: %s\n", browser, strings.Join(r.droppedProps[browser], ", "))
	}
	fmt.Fprintf(w, "media features: %d (dropped: %s)\n", len(m.features), strings.Join(r.droppedFeatures, ", "))
	fmt.Fprintf(w, "pseudos: %d (dropped: %s)\n", len(m.pseudos), strings.Join(r.droppedPseudos, ", "))
	fmt.Fprintf(w, "range syntax: %t\n", m.rangeSyntax)
	var counts []string
	for _, rule := range slices.Sorted(maps.Keys(descriptorRules)) {
		counts = append(counts, fmt.Sprintf("%s: %d", rule, r.descriptors[rule]))
	}
	fmt.Fprintf(w, "descriptors: %d (%s; dropped: %s)\n", len(m.descriptors), strings.Join(counts, ", "), strings.Join(r.droppedDescriptors, ", "))
	fmt.Fprintf(w, "container features: %d (dropped: %s)\n", len(m.containerFeatures), strings.Join(r.droppedContainer, ", "))
	var verdicts []string
	for _, v := range r.atRules {
		verdicts = append(verdicts, fmt.Sprintf("%s=%t", v.name, v.renderable))
	}
	fmt.Fprintf(w, "at-rules renderable: %s\n", strings.Join(verdicts, " "))
}
