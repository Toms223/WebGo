package bcd

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	dataFile      = "data.json"
	compatKey     = "__compat"
	dateLayout    = "2006-01-02"
	preview       = "preview"
	ranged        = "≤"
	statusCurrent = "current"
)

var Browsers = []string{"chrome", "chrome_android", "edge", "firefox", "firefox_android", "safari", "safari_ios"}

type Release struct {
	Date   time.Time
	Status string
}

type Statement struct {
	VersionAdded   string
	VersionRemoved string
	Added          bool
	Prefix         string
	Alternative    string
	Flagged        bool
	Partial        bool
}

type Compat struct {
	Support       map[string][]Statement
	StandardTrack bool
	Deprecated    bool
}

type Data struct {
	Releases map[string]map[string]Release
	Features map[string]*Compat
}

type document struct {
	Browsers map[string]json.RawMessage `json:"browsers"`
	CSS      struct {
		Properties map[string]json.RawMessage `json:"properties"`
		AtRules    map[string]json.RawMessage `json:"at-rules"`
		Selectors  map[string]json.RawMessage `json:"selectors"`
	} `json:"css"`
}

type rawBrowser struct {
	Releases map[string]struct {
		ReleaseDate string `json:"release_date"`
		Status      string `json:"status"`
	} `json:"releases"`
}

type rawCompat struct {
	Support map[string]json.RawMessage `json:"support"`
	Status  struct {
		StandardTrack bool `json:"standard_track"`
		Deprecated    bool `json:"deprecated"`
	} `json:"status"`
}

type rawStatement struct {
	VersionAdded    json.RawMessage   `json:"version_added"`
	VersionRemoved  json.RawMessage   `json:"version_removed"`
	Prefix          string            `json:"prefix"`
	AlternativeName string            `json:"alternative_name"`
	Flags           []json.RawMessage `json:"flags"`
	Partial         bool              `json:"partial_implementation"`
}

func Load(dir string) (*Data, error) {
	raw, err := os.ReadFile(filepath.Join(dir, dataFile))
	if err != nil {
		return nil, fmt.Errorf("bcd: %w", err)
	}

	var doc document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("bcd: %s: %w", dataFile, err)
	}

	data := &Data{
		Releases: make(map[string]map[string]Release),
		Features: make(map[string]*Compat),
	}

	for _, browser := range Browsers {
		entry, ok := doc.Browsers[browser]
		if !ok {
			continue
		}
		releases, err := decodeReleases(browser, entry)
		if err != nil {
			return nil, err
		}
		data.Releases[browser] = releases
	}

	roots := []struct {
		path     string
		children map[string]json.RawMessage
	}{
		{"css.properties", doc.CSS.Properties},
		{"css.at-rules", doc.CSS.AtRules},
		{"css.selectors", doc.CSS.Selectors},
	}
	for _, root := range roots {
		if err := collect(data.Features, root.path, root.children); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (d *Data) Has(key string) bool {
	_, ok := d.Features[key]
	return ok
}

func (d *Data) Standard(key string) bool {
	compat := d.Features[key]
	return compat != nil && compat.StandardTrack && !compat.Deprecated
}

func (d *Data) Latest(browser string) string {
	var current, highest string
	for version, release := range d.Releases[browser] {
		if highest == "" || later(version, highest) {
			highest = version
		}
		if release.Status == statusCurrent && (current == "" || later(version, current)) {
			current = version
		}
	}
	return cmp.Or(current, highest)
}

func (d *Data) Renderable(key string) bool {
	return d.FirstUnsupported(key) == ""
}

func (d *Data) FirstUnsupported(key string) string {
	compat := d.Features[key]
	for _, browser := range Browsers {
		if compat == nil {
			return browser
		}
		latest := d.Latest(browser)
		if !slices.ContainsFunc(compat.Support[browser], func(s Statement) bool { return supports(s, latest) }) {
			return browser
		}
	}
	return ""
}

func (d *Data) Prefixes(key string, floor time.Time) []string {
	compat := d.Features[key]
	if compat == nil {
		return nil
	}

	var prefixes []string
	for _, browser := range Browsers {
		statements := compat.Support[browser]
		for version, release := range d.Releases[browser] {
			if release.Date.Before(floor) {
				continue
			}
			if slices.ContainsFunc(statements, func(s Statement) bool { return s.Prefix == "" && supports(s, version) }) {
				continue
			}
			for _, s := range statements {
				if s.Prefix != "" && supports(s, version) {
					prefixes = append(prefixes, s.Prefix)
				}
			}
		}
	}

	slices.Sort(prefixes)
	return slices.Compact(prefixes)
}

func VersionLess(a, b string) bool {
	return compareVersions(a, b) < 0
}

func Covers(s Statement, version string) bool {
	if !s.Added || s.Alternative != "" {
		return false
	}
	if s.VersionAdded != "" && VersionLess(version, s.VersionAdded) {
		return false
	}
	return s.VersionRemoved == "" || VersionLess(version, s.VersionRemoved)
}

func supports(s Statement, version string) bool {
	return !s.Flagged && Covers(s, version)
}

func later(a, b string) bool {
	return cmp.Or(compareVersions(a, b), strings.Compare(a, b)) > 0
}

func compareVersions(a, b string) int {
	aNumbers, aNumeric := parseVersion(a)
	bNumbers, bNumeric := parseVersion(b)
	switch {
	case aNumeric && bNumeric:
		return slices.Compare(aNumbers, bNumbers)
	case aNumeric:
		return -1
	case bNumeric:
		return 1
	}
	return strings.Compare(a, b)
}

func parseVersion(version string) ([]int, bool) {
	if version == "" {
		return nil, true
	}
	parts := strings.Split(strings.TrimPrefix(version, ranged), ".")
	numbers := make([]int, len(parts))
	for i, part := range parts {
		number, err := strconv.ParseUint(part, 10, 31)
		if err != nil {
			return nil, false
		}
		numbers[i] = int(number)
	}
	for len(numbers) > 0 && numbers[len(numbers)-1] == 0 {
		numbers = numbers[:len(numbers)-1]
	}
	return numbers, true
}

func decodeReleases(browser string, raw json.RawMessage) (map[string]Release, error) {
	var entry rawBrowser
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, fmt.Errorf("bcd: browsers.%s: %w", browser, err)
	}

	releases := make(map[string]Release, len(entry.Releases))
	for _, version := range slices.Sorted(maps.Keys(entry.Releases)) {
		item := entry.Releases[version]
		var date time.Time
		if item.ReleaseDate != "" {
			parsed, err := time.Parse(dateLayout, item.ReleaseDate)
			if err != nil {
				return nil, fmt.Errorf("bcd: browsers.%s.releases.%s: %w", browser, version, err)
			}
			date = parsed
		}
		releases[version] = Release{Date: date, Status: item.Status}
	}
	return releases, nil
}

func collect(features map[string]*Compat, prefix string, children map[string]json.RawMessage) error {
	for _, name := range slices.Sorted(maps.Keys(children)) {
		if name == compatKey {
			continue
		}
		path := prefix + "." + name

		var node map[string]json.RawMessage
		if err := json.Unmarshal(children[name], &node); err != nil {
			return fmt.Errorf("bcd: %s: %w", path, err)
		}

		if raw, ok := node[compatKey]; ok {
			compat, err := decodeCompat(raw)
			if err != nil {
				return fmt.Errorf("bcd: %s: %w", path, err)
			}
			features[path] = compat
		}

		if err := collect(features, path, node); err != nil {
			return err
		}
	}
	return nil
}

func decodeCompat(raw json.RawMessage) (*Compat, error) {
	var decoded rawCompat
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}

	compat := &Compat{
		Support:       make(map[string][]Statement),
		StandardTrack: decoded.Status.StandardTrack,
		Deprecated:    decoded.Status.Deprecated,
	}
	for _, browser := range Browsers {
		entry, ok := decoded.Support[browser]
		if !ok {
			continue
		}
		statements, err := decodeStatements(entry)
		if err != nil {
			return nil, fmt.Errorf("support.%s: %w", browser, err)
		}
		compat.Support[browser] = statements
	}
	return compat, nil
}

func decodeStatements(raw json.RawMessage) ([]Statement, error) {
	var entries []rawStatement
	if bytes.HasPrefix(bytes.TrimSpace(raw), []byte("[")) {
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, err
		}
	} else {
		var entry rawStatement
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, err
		}
		entries = []rawStatement{entry}
	}

	statements := make([]Statement, len(entries))
	for i, entry := range entries {
		statement, err := entry.statement()
		if err != nil {
			return nil, err
		}
		statements[i] = statement
	}
	return statements, nil
}

func (r rawStatement) statement() (Statement, error) {
	added, hasAdded, err := decodeVersion(r.VersionAdded)
	if err != nil {
		return Statement{}, fmt.Errorf("version_added: %w", err)
	}
	removed, hasRemoved, err := decodeVersion(r.VersionRemoved)
	if err != nil {
		return Statement{}, fmt.Errorf("version_removed: %w", err)
	}

	removedInUnknownVersion := hasRemoved && removed == ""
	return Statement{
		VersionAdded:   added,
		VersionRemoved: removed,
		Added:          hasAdded && !removedInUnknownVersion,
		Prefix:         r.Prefix,
		Alternative:    r.AlternativeName,
		Flagged:        len(r.Flags) > 0,
		Partial:        r.Partial,
	}, nil
}

func decodeVersion(raw json.RawMessage) (string, bool, error) {
	if len(raw) == 0 {
		return "", false, nil
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false, err
	}

	switch value := value.(type) {
	case nil:
		return "", false, nil
	case bool:
		return "", value, nil
	case string:
		if !validVersion(value) {
			return "", false, fmt.Errorf("unrecognised version %q", value)
		}
		return value, true, nil
	}
	return "", false, fmt.Errorf("expected a version string or a boolean, got %s", raw)
}

func validVersion(version string) bool {
	_, numeric := parseVersion(version)
	return version != "" && (numeric || version == preview)
}
