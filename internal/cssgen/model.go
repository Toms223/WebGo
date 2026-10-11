package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/Toms223/WebGo/internal/cssgen/bcd"
	"github.com/Toms223/WebGo/internal/cssgen/grammar"
	"github.com/Toms223/WebGo/internal/cssgen/webref"
)

const rangeSyntaxKey = "css.at-rules.media.range_syntax"

var rangeKinds = []string{"Length", "Integer", "Number", "Resolution", "RatioValue"}

var excludedFeatures = []string{"grid"}

var excludedPseudos = []string{":first", ":left", ":right", ":blank"}

var pseudoArgs = map[string]string{
	":nth-child()":        "nth",
	":nth-last-child()":   "nth",
	":nth-of-type()":      "nth",
	":nth-last-of-type()": "nth",
	":lang()":             "string",
	":dir()":              "dir",
	"::highlight()":       "string",
}

var dirKeywords = []string{"ltr", "rtl"}

var cssWideKeywords = []string{"inherit", "initial", "unset", "revert", "revert-layer"}

const fontFormatType = "font-format"

var descriptorRules = map[string]string{"font-face": "Face", "property": "Property"}

var containerKeys = map[string]string{
	"scrollable": "scroll-state_queries",
	"scrolled":   "scroll-state_queries",
	"snapped":    "scroll-state_queries",
	"stuck":      "scroll-state_queries",
	"fallback":   "anchor_position_queries",
}

var atRules = []struct{ name, key string }{
	{"container", "css.at-rules.container"},
	{"property", "css.at-rules.property"},
	{"layer", "css.at-rules.layer"},
	{"supports", "css.at-rules.supports"},
	{"import", "css.at-rules.import"},
	{"font-face", "css.at-rules.font-face"},
	{"keyframes", "css.at-rules.keyframes"},
	{"nesting", "css.selectors.nesting"},
}

type property struct {
	css, goName string
	tier        int
	shape       grammar.Shape
	prefixes    []string
}

type descriptor struct {
	atRule, css, goName string
	tier                int
	shape               grammar.Shape
}

type keyword struct {
	css, goName                             string
	cssProps, mediaFeatures, styleUses      []string
	descriptors, containerFeatures, cssUses []string
}

type mediaFeature struct {
	css, goName string
	rangeKind   string
	discrete    *grammar.Shape
}

type pseudo struct {
	css, goName string
	element     bool
	arg         string
}

type model struct {
	properties        []property
	descriptors       []descriptor
	keywords          []keyword
	features          []mediaFeature
	containerFeatures []mediaFeature
	pseudos           []pseudo
	rangeSyntax       bool
}

type verdict struct {
	name       string
	renderable bool
}

type report struct {
	tier1, tier2                         int
	droppedProps                         map[string][]string
	droppedFeatures, droppedPseudos      []string
	descriptors                          map[string]int
	droppedDescriptors, droppedContainer []string
	atRules                              []verdict
}

type assembly struct {
	w            *webref.Data
	b            *bcd.Data
	floor        time.Time
	props, types func(name string) (*grammar.Node, bool)
	words        map[string]*keyword
	cssNames     *namespace
	model        *model
	report       report
}

func build(w *webref.Data, b *bcd.Data, floor time.Time) (*model, report, error) {
	a := &assembly{
		w:     w,
		b:     b,
		floor: floor,
		props: memoise(func(name string) (*grammar.Node, error) {
			var values []string
			if p := w.Properties[name]; p != nil {
				values = p.Values
			}
			return grammar.MergeAlternatives(values)
		}),
		types: memoise(func(name string) (*grammar.Node, error) {
			return grammar.Parse(w.Values[name])
		}),
		words:    map[string]*keyword{},
		cssNames: newNamespace("package css", reservedCSS),
		model:    &model{rangeSyntax: b.Renderable(rangeSyntaxKey) && b.Prefixes(rangeSyntaxKey, floor) == nil},
		report:   report{droppedProps: map[string][]string{}, descriptors: map[string]int{}},
	}
	for _, css := range slices.Concat(a.colorKeywords(), cssWideKeywords, dirKeywords) {
		a.word(css)
	}
	for _, step := range []func() error{a.checkAtRules, a.addProperties, a.addDescriptors, a.addFeatures, a.addContainerFeatures, a.addPseudos, a.addKeywords} {
		if err := step(); err != nil {
			return nil, report{}, err
		}
	}
	return a.model, a.report, nil
}

func memoise(parse func(name string) (*grammar.Node, error)) func(string) (*grammar.Node, bool) {
	cache := map[string]*grammar.Node{}
	return func(name string) (*grammar.Node, bool) {
		n, seen := cache[name]
		if !seen {
			n, _ = parse(name)
			cache[name] = n
		}
		return n, n != nil
	}
}

func (a *assembly) checkAtRules() error {
	var unsupported []string
	for _, rule := range atRules {
		renderable := a.b.Renderable(rule.key)
		a.report.atRules = append(a.report.atRules, verdict{rule.name, renderable})
		if !renderable {
			unsupported = append(unsupported, rule.name)
		}
	}
	if len(unsupported) > 0 {
		return fmt.Errorf("cssgen: at-rules not renderable at the pins: %s", strings.Join(unsupported, ", "))
	}
	return nil
}

func (a *assembly) addProperties() error {
	names := a.cssNames
	for _, name := range slices.Sorted(maps.Keys(a.w.Properties)) {
		key := "css.properties." + name
		if !a.b.Standard(key) {
			continue
		}
		if browser := a.b.FirstUnsupported(key); browser != "" {
			a.report.droppedProps[browser] = append(a.report.droppedProps[browser], name)
			continue
		}
		n, _ := a.props(name)
		p := property{
			css:      name,
			goName:   goName(name),
			tier:     2,
			shape:    grammar.Shape{Comma: commaSeparated(n)},
			prefixes: a.b.Prefixes(key, a.floor),
		}
		if shape, ok := grammar.Classify(n, a.props, a.types); ok {
			shape.Keywords = lowercase(shape.Keywords)
			p.tier, p.shape = 1, shape
		}
		if err := claimProperty(names, p); err != nil {
			return err
		}
		for _, css := range grammar.Keywords(n, a.props, a.types) {
			a.word(css)
		}
		if p.tier == 1 {
			a.report.tier1++
			a.accept(p.shape.Keywords, name)
			a.accept(cssWideKeywords, name)
			if slices.Contains(p.shape.Kinds, "ColorLiteral") {
				a.accept(a.colorKeywords(), name)
			}
		} else {
			a.report.tier2++
		}
		a.model.properties = append(a.model.properties, p)
	}
	return nil
}

func claimProperty(names *namespace, p property) error {
	item := fmt.Sprintf("property %q", p.css)
	if p.tier == 1 {
		if err := names.claim(item, p.goName+"Value"); err != nil {
			return err
		}
	}
	return names.claim(item, p.goName)
}

func commaSeparated(n *grammar.Node) bool {
	switch {
	case n == nil:
		return false
	case n.Comma:
		return true
	case n.Kind == grammar.Seq:
		return slices.ContainsFunc(n.Children, func(c *grammar.Node) bool { return c.Kind == grammar.Keyword && c.Text == "," })
	case n.Kind == grammar.Alt:
		lists := 0
		for _, a := range n.Children {
			if a.Kind == grammar.Keyword {
				continue
			}
			if !a.Comma {
				return false
			}
			lists++
		}
		return lists > 0
	}
	return false
}

func (a *assembly) addDescriptors() error {
	for _, rule := range slices.Sorted(maps.Keys(descriptorRules)) {
		syntaxes := a.w.Descriptors[rule]
		for _, name := range slices.Sorted(maps.Keys(syntaxes)) {
			key := "css.at-rules." + rule + "." + name
			if !a.b.Standard(key) {
				continue
			}
			if !a.b.Renderable(key) {
				a.report.droppedDescriptors = append(a.report.droppedDescriptors, rule+"/"+name)
				continue
			}
			n, _ := grammar.Parse(syntaxes[name])
			d := descriptor{atRule: rule, css: name, goName: goName(name), tier: 2, shape: grammar.Shape{Comma: commaSeparated(n)}}
			if _, taken := a.cssNames.claimed[d.goName]; taken {
				d.goName = descriptorRules[rule] + d.goName
			}
			if shape, ok := grammar.Classify(n, a.props, a.types); ok {
				shape.Keywords = lowercase(shape.Keywords)
				d.tier, d.shape = 1, shape
			}
			if err := claimDescriptor(a.cssNames, d); err != nil {
				return err
			}
			for _, css := range grammar.Keywords(n, a.props, a.types) {
				a.word(css)
			}
			if d.tier == 1 {
				accepted := d.shape.Keywords
				if slices.Contains(d.shape.Kinds, "ColorLiteral") {
					accepted = slices.Concat(accepted, a.colorKeywords())
				}
				for _, css := range accepted {
					k := a.word(css)
					k.descriptors = append(k.descriptors, d.goName)
				}
			}
			a.report.descriptors[rule]++
			a.model.descriptors = append(a.model.descriptors, d)
		}
	}
	if formats, ok := a.types(fontFormatType); ok {
		for _, css := range grammar.Keywords(formats, a.props, a.types) {
			k := a.word(css)
			k.cssUses = append(k.cssUses, "Format")
		}
	}
	return nil
}

func claimDescriptor(names *namespace, d descriptor) error {
	item := fmt.Sprintf("@%s descriptor %q", d.atRule, d.css)
	if d.tier == 1 {
		if err := names.claim(item, d.goName+"Descriptor"); err != nil {
			return err
		}
	}
	return names.claim(item, d.goName)
}

func (a *assembly) addFeatures() error {
	features, dropped, err := a.collectFeatures("media", reservedMedia, a.w.Features, mediaKey, func(k *keyword) *[]string { return &k.mediaFeatures })
	a.model.features, a.report.droppedFeatures = features, dropped
	return err
}

func (a *assembly) addContainerFeatures() error {
	features, dropped, err := a.collectFeatures("container", reservedContainer, a.w.Descriptors["container"], containerKey, func(k *keyword) *[]string { return &k.containerFeatures })
	a.model.containerFeatures, a.report.droppedContainer = features, dropped
	return err
}

func mediaKey(name string) string {
	return "css.at-rules.media." + name
}

func containerKey(name string) string {
	if query, ok := containerKeys[name]; ok {
		return "css.at-rules.container." + query
	}
	return "css.at-rules.container"
}

func (a *assembly) collectFeatures(pkg string, reserved []string, syntaxes map[string]string, key func(string) string, uses func(*keyword) *[]string) ([]mediaFeature, []string, error) {
	names := newNamespace("package "+pkg, reserved)
	var features []mediaFeature
	var dropped []string
	for _, name := range slices.Sorted(maps.Keys(syntaxes)) {
		f, ok := a.featureOf(name, syntaxes[name], key(name))
		if !ok {
			dropped = append(dropped, name)
			continue
		}
		if err := names.claim(fmt.Sprintf("%s feature %q", pkg, name), f.goName); err != nil {
			return nil, nil, err
		}
		if f.discrete != nil {
			for _, css := range f.discrete.Keywords {
				list := uses(a.word(css))
				*list = append(*list, name)
			}
		}
		features = append(features, f)
	}
	return features, dropped, nil
}

func (a *assembly) featureOf(name, syntax, key string) (mediaFeature, bool) {
	if strings.HasPrefix(name, "-") || slices.Contains(excludedFeatures, name) {
		return mediaFeature{}, false
	}
	n, _ := grammar.Parse(syntax)
	shape, ok := grammar.Classify(n, a.props, a.types)
	if !ok || !a.b.Renderable(key) {
		return mediaFeature{}, false
	}
	f := mediaFeature{css: name, goName: goName(name), rangeKind: rangeKind(shape)}
	switch {
	case f.rangeKind != "":
		return f, true
	case len(shape.Kinds) == 0:
		shape.Keywords = lowercase(shape.Keywords)
		f.discrete = &shape
		return f, true
	}
	return mediaFeature{}, false
}

func rangeKind(s grammar.Shape) string {
	if len(s.Kinds) == 1 && s.Min == 1 && s.Max == 1 && slices.Contains(rangeKinds, s.Kinds[0]) {
		return s.Kinds[0]
	}
	return ""
}

func (a *assembly) addPseudos() error {
	names := newNamespace("style.Rule", reservedRule)
	listed := slices.Compact(slices.Sorted(slices.Values(a.w.Selectors)))
	for _, name := range listed {
		if !strings.HasPrefix(name, ":") {
			continue
		}
		p, ok := a.pseudoOf(name, listed)
		if !ok {
			a.report.droppedPseudos = append(a.report.droppedPseudos, name)
			continue
		}
		if err := names.claim(fmt.Sprintf("pseudo %q", name), p.goName); err != nil {
			return err
		}
		if p.arg == "dir" {
			for _, css := range dirKeywords {
				k := a.word(css)
				k.styleUses = append(k.styleUses, p.goName)
			}
		}
		a.model.pseudos = append(a.model.pseudos, p)
	}
	return nil
}

func (a *assembly) pseudoOf(name string, listed []string) (pseudo, bool) {
	arg, known := pseudoArgs[name]
	_, legacy := slices.BinarySearch(listed, ":"+name)
	key := "css.selectors." + strings.TrimSuffix(strings.TrimLeft(name, ":"), "()")
	if (strings.HasSuffix(name, "()") && !known) || legacy || slices.Contains(excludedPseudos, name) || !a.b.Renderable(key) {
		return pseudo{}, false
	}
	return pseudo{css: name, goName: goName(name), element: strings.HasPrefix(name, "::"), arg: arg}, true
}

func (a *assembly) addKeywords() error {
	names := newNamespace("package kw", nil)
	for _, css := range slices.Sorted(maps.Keys(a.words)) {
		k := a.words[css]
		if err := names.claim(fmt.Sprintf("keyword %q", css), k.goName); err != nil {
			return err
		}
		k.cssProps = sortedSet(k.cssProps)
		k.mediaFeatures = sortedSet(k.mediaFeatures)
		k.styleUses = sortedSet(k.styleUses)
		k.descriptors = sortedSet(k.descriptors)
		k.containerFeatures = sortedSet(k.containerFeatures)
		k.cssUses = sortedSet(k.cssUses)
		a.model.keywords = append(a.model.keywords, *k)
	}
	return nil
}

func (a *assembly) colorKeywords() []string {
	return slices.Concat(a.w.NamedColors, []string{"currentcolor", "transparent"})
}

func (a *assembly) word(css string) *keyword {
	css = strings.ToLower(css)
	k, ok := a.words[css]
	if !ok {
		k = &keyword{css: css, goName: goName(css)}
		a.words[css] = k
	}
	return k
}

func (a *assembly) accept(words []string, prop string) {
	for _, css := range words {
		k := a.word(css)
		k.cssProps = append(k.cssProps, prop)
	}
}

func lowercase(words []string) []string {
	var lower []string
	for _, w := range words {
		lower = append(lower, strings.ToLower(w))
	}
	return sortedSet(lower)
}

func sortedSet(s []string) []string {
	slices.Sort(s)
	return slices.Compact(s)
}
