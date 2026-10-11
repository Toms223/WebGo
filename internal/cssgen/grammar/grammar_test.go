package grammar

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

var labels = map[Kind]string{
	Alt: "alt", AnyOrder: "any", AllOrder: "all", Seq: "seq", Group: "group",
	Keyword: "kw:", Type: "type:", PropRef: "ref:", Func: "fn:",
}

func dump(n *Node) string {
	if n == nil {
		return "nil"
	}
	s := labels[n.Kind] + n.Text + multiplier(n)
	if n.Kind == Keyword || n.Kind == Type || n.Kind == PropRef {
		return s
	}
	parts := make([]string, len(n.Children))
	for i, c := range n.Children {
		parts[i] = dump(c)
	}
	return s + "(" + strings.Join(parts, " ") + ")"
}

func multiplier(n *Node) string {
	switch {
	case n.Comma && n.Min == 1 && n.Max == -1:
		return "{#}"
	case n.Comma:
		return fmt.Sprintf("{#%d,%d}", n.Min, n.Max)
	case n.Min != 1 || n.Max != 1:
		return fmt.Sprintf("{%d,%d}", n.Min, n.Max)
	}
	return ""
}

func TestParseGrammar(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{"auto | <length-percentage [0,∞]> | min-content", "alt(kw:auto type:length-percentage kw:min-content)"},
		{"<'padding-top'>{1,4}", "ref:padding-top{1,4}"},
		{"[ <family-name> | <generic-family> ]#", "group{#}(alt(type:family-name type:generic-family))"},
		{"none | <filter-value-list>", "alt(kw:none type:filter-value-list)"},
		{"<length> && <color>", "all(type:length type:color)"},
		{"a || b", "any(kw:a kw:b)"},
		{"fit-content(<length>)", "fn:fit-content(type:length)"},
		{"<length>+", "type:length{1,-1}"},
		{"<length>?", "type:length{0,1}"},
		{"a b | c && d || e", "alt(seq(kw:a kw:b) any(all(kw:c kw:d) kw:e))"},
		{"[ a | b ] c", "seq(group(alt(kw:a kw:b)) kw:c)"},
		{"<length> / <length>", "seq(type:length kw:/ type:length)"},
		{"<length>*", "type:length{0,-1}"},
		{"<length>{2}", "type:length{2,2}"},
		{"<length>{2,}", "type:length{2,-1}"},
		{"<length>#{1,4}", "type:length{#1,4}"},
		{"none | [<length-percentage> | <number>]+#", "alt(kw:none group{#}(group{1,-1}(alt(type:length-percentage type:number))))"},
		{"<bg-layer>#? , <final-bg-layer>", "seq(group{0,1}(type:bg-layer{#}) kw:, type:final-bg-layer)"},
		{"[ center | [ [ start | end ]? <length-percentage>? ]! ]#", "group{#}(alt(kw:center group(seq(group{0,1}(alt(kw:start kw:end)) type:length-percentage{0,1}))))"},
		{"auto | <overflow-position>? [ normal | <self-position> ]| stretch", "alt(kw:auto seq(type:overflow-position{0,1} group(alt(kw:normal type:self-position))) kw:stretch)"},
		{"[<cursor-image>,]* <cursor-predefined>", "seq(group{0,-1}(seq(type:cursor-image kw:,)) type:cursor-predefined)"},
		{"none | <length> [0,∞]", "alt(kw:none type:length)"},
		{"auto-add | add(<integer>) | <anchor-size()> | <image-1D>", "alt(kw:auto-add fn:add(type:integer) type:anchor-size() type:image-1D)"},
		{"sibling-count()", "fn:sibling-count()"},
	}
	for _, c := range cases {
		n, err := Parse(c.src)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", c.src, err)
			continue
		}
		if got := dump(n); got != c.want {
			t.Errorf("Parse(%q) = %s, want %s", c.src, got, c.want)
		}
	}
}

func TestParseGrammarRejectsMalformedInput(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"a |",
		"| a",
		"a || || b",
		"?",
		"[ a",
		"a ]",
		"[ ]",
		"fit-content(<length>",
		"a )",
		"<length",
		"<>",
		"<'padding-top>",
		"<length [0,∞>",
		"<length foo>",
		"a{2,1}",
		"a{x}",
		"a{1,2",
		"a{-1,2}",
		"a & b",
		"a ! b",
		"a @ b",
		"(a)",
	}
	for _, src := range cases {
		if n, err := Parse(src); err == nil {
			t.Errorf("Parse(%q) = %s, want an error", src, dump(n))
		}
	}
}

func TestMergeAlternatives(t *testing.T) {
	cases := []struct {
		values []string
		want   string
	}{
		{[]string{"block | inline | flex", "grid | inline-flex"}, "alt(kw:block kw:inline kw:flex kw:grid kw:inline-flex)"},
		{[]string{"auto", "<length> <length>"}, "alt(kw:auto seq(type:length type:length))"},
		{[]string{"[ <family-name> | <generic-family> ]#"}, "group{#}(alt(type:family-name type:generic-family))"},
	}
	for _, c := range cases {
		n, err := MergeAlternatives(c.values)
		if err != nil {
			t.Errorf("MergeAlternatives(%q) failed: %v", c.values, err)
			continue
		}
		if got := dump(n); got != c.want {
			t.Errorf("MergeAlternatives(%q) = %s, want %s", c.values, got, c.want)
		}
	}
	for _, values := range [][]string{nil, {"auto", "[ a"}} {
		if n, err := MergeAlternatives(values); err == nil {
			t.Errorf("MergeAlternatives(%q) = %s, want an error", values, dump(n))
		}
	}
}

var propGrammars = map[string]string{
	"width":           "auto | <length-percentage [0,∞]> | min-content | max-content",
	"padding-top":     "<length-percentage [0,∞]>",
	"padding":         "<'padding-top'>{1,4}",
	"backdrop-filter": "none | <filter-value-list>",
	"font-family":     "[ <family-name> | <generic-family> ]#",
	"anchor-name":     "none | <dashed-ident>#",
	"display":         "block | inline | flex",
	"row-gap":         "normal | <length-percentage [0,∞]>",
	"column-gap":      "normal | <length-percentage [0,∞]>",
	"loop":            "<'loop'>",
	"ping":            "auto | <'pong'>",
	"pong":            "none | <'ping'>",
	"position-area":   "none | <position-area>",
}

var newValues = map[string]string{"display": "grid | inline-flex"}

var typeGrammars = map[string]string{
	"box-size":            "<length-percentage [0,∞]> | min-content | max-content | fit-content(<length-percentage [0,∞]>) | <calc-size()> | stretch",
	"line-width":          "<length [0,∞]> | thin | medium | thick",
	"display-outside":     "block | inline | run-in",
	"display-inside":      "flow | flow-root | table | flex | grid | ruby",
	"display-listitem":    "<display-outside>? && [ flow | flow-root ]? && list-item",
	"display-box":         "contents | none",
	"single-transition":   "[ none | <custom-ident> ] || <time> || <easing-function> || <time>",
	"easing-function":     "linear | ease | cubic-bezier(<number>, <number>, <number>, <number>)",
	"family-name":         "<string> | <custom-ident>+",
	"generic-family":      "serif | sans-serif | monospace",
	"animateable-feature": "scroll-position | contents | <custom-ident>",
	"self-ref":            "<self-ref> | auto",
	"self-list":           "auto | <self-list>#",
	"position-area":       "left | right",
	"calc-size()":         "calc-size-leaked",
}

func mustParse(t *testing.T, src string) *Node {
	t.Helper()
	n, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q) failed: %v", src, err)
	}
	return n
}

func propResolver(t *testing.T) func(string) (*Node, bool) {
	return func(prop string) (*Node, bool) {
		src, ok := propGrammars[prop]
		if !ok {
			return nil, false
		}
		values := []string{src}
		if extra, ok := newValues[prop]; ok {
			values = append(values, extra)
		}
		n, err := MergeAlternatives(values)
		if err != nil {
			t.Fatalf("MergeAlternatives(%q) failed: %v", values, err)
		}
		return n, true
	}
}

func typeResolver(t *testing.T) func(string) (*Node, bool) {
	defs := map[string]string{"deep-9": "found"}
	for i := 1; i < 9; i++ {
		defs[fmt.Sprintf("deep-%d", i)] = fmt.Sprintf("<deep-%d>", i+1)
	}
	for name, src := range typeGrammars {
		defs[name] = src
	}
	return func(name string) (*Node, bool) {
		src, ok := defs[name]
		if !ok {
			return nil, false
		}
		return mustParse(t, src), true
	}
}

func TestClassify(t *testing.T) {
	props, types := propResolver(t), typeResolver(t)
	cases := []struct {
		prop     string
		src      string
		tier1    bool
		kinds    []string
		keywords []string
		min, max int
		comma    bool
	}{
		{prop: "width", tier1: true, kinds: []string{"Length", "Percentage"}, keywords: []string{"auto", "max-content", "min-content"}, min: 1, max: 1},
		{prop: "padding", tier1: true, kinds: []string{"Length", "Percentage"}, min: 1, max: 4},
		{prop: "font-family"},
		{prop: "backdrop-filter", tier1: true, keywords: []string{"none"}, min: 1, max: 1},
		{prop: "anchor-name", tier1: true, kinds: []string{"IdentValue"}, keywords: []string{"none"}, min: 1, max: -1, comma: true},
		{prop: "display", tier1: true, keywords: []string{"block", "flex", "grid", "inline", "inline-flex"}, min: 1, max: 1},
		{src: "<length>?"},
		{src: "<length> && <color>", tier1: true, kinds: []string{"ColorLiteral", "Length"}, min: 1, max: 2},
		{src: "auto | <box-size> | <anchor-size()>", tier1: true, kinds: []string{"Length", "Percentage"}, keywords: []string{"auto", "max-content", "min-content", "stretch"}, min: 1, max: 1},
		{src: "<line-width>{1,4}", tier1: true, kinds: []string{"Length"}, keywords: []string{"medium", "thick", "thin"}, min: 1, max: 4},
		{src: "[ <display-outside> || <display-inside> ] | <display-listitem> | <display-box>", tier1: true, keywords: []string{"block", "contents", "flex", "flow", "flow-root", "grid", "inline", "list-item", "none", "ruby", "run-in", "table"}, min: 1, max: 3},
		{src: "<line-width> || <color>", tier1: true, kinds: []string{"ColorLiteral", "Length"}, keywords: []string{"medium", "thick", "thin"}, min: 1, max: 2},
		{src: "<'row-gap'> <'column-gap'>?", tier1: true, kinds: []string{"Length", "Percentage"}, keywords: []string{"normal"}, min: 1, max: 2},
		{src: "<single-transition>#"},
		{src: "[ <family-name> | <generic-family> ]#"},
		{src: "auto | <animateable-feature>#", tier1: true, kinds: []string{"IdentValue"}, keywords: []string{"auto", "contents", "scroll-position"}, min: 1, max: -1, comma: true},
		{src: "<self-ref>", tier1: true, keywords: []string{"auto"}, min: 1, max: 1},
		{src: "<self-list>", tier1: true, keywords: []string{"auto"}, min: 1, max: -1, comma: true},
		{src: "<'position-area'>", tier1: true, keywords: []string{"left", "none", "right"}, min: 1, max: 1},
		{src: "<unknown-thing>"},
		{src: "<length-percentage [0,∞]>{1,4} [ / <length-percentage [0,∞]>{1,4} ]?"},
		{src: "<length>+", tier1: true, kinds: []string{"Length"}, min: 1, max: -1},
		{src: "<length>#{1,4}", tier1: true, kinds: []string{"Length"}, min: 1, max: 4, comma: true},
		{src: "[ none | <dashed-ident> ]#", tier1: true, kinds: []string{"IdentValue"}, keywords: []string{"none"}, min: 1, max: -1, comma: true},
		{src: "none | [ <number> | <percentage> ]{1,3}", tier1: true, kinds: []string{"Number", "Percentage"}, keywords: []string{"none"}, min: 1, max: 3},
		{src: "normal | <length> | normal | <length-percentage> | <color>", tier1: true, kinds: []string{"ColorLiteral", "Length", "Percentage"}, keywords: []string{"normal"}, min: 1, max: 1},
		{src: "auto | <'padding-top'>", tier1: true, kinds: []string{"Length", "Percentage"}, keywords: []string{"auto"}, min: 1, max: 1},
		{src: "<'padding'>", tier1: true, kinds: []string{"Length", "Percentage"}, min: 1, max: 4},
		{src: "<length> <color>", tier1: true, kinds: []string{"ColorLiteral", "Length"}, min: 2, max: 2},
		{src: "<deep-2>", tier1: true, keywords: []string{"found"}, min: 1, max: 1},
		{prop: "ping", tier1: true, keywords: []string{"auto", "none"}, min: 1, max: 1},
		{src: "<deep-1>"},
		{src: "<length>*"},
		{src: "<length>{0,3}"},
		{src: "none | <length>?"},
		{src: "<length>? <color>?"},
		{src: "<length>+ || <color>"},
		{src: "[ a || b ] && c"},
		{src: "[ <length> | <number> ]+#"},
		{src: "<length># | <percentage>+"},
		{src: "<'padding'>{1,2}"},
		{src: "<length> / <length>"},
		{src: "auto | /"},
		{src: "fit-content(<length>)"},
		{src: "<'unknown'>"},
		{src: "auto | <'unknown'>"},
		{prop: "loop"},
	}
	node := func(prop, src string) *Node {
		if prop == "" {
			return mustParse(t, src)
		}
		n, ok := props(prop)
		if !ok {
			t.Fatalf("no grammar for %s", prop)
		}
		return n
	}
	for _, c := range cases {
		name := c.prop + c.src
		got, ok := Classify(node(c.prop, c.src), props, types)
		if ok != c.tier1 {
			t.Errorf("Classify(%s) tier 1 = %v, want %v (shape %+v)", name, ok, c.tier1, got)
			continue
		}
		if !ok {
			continue
		}
		if !slices.Equal(got.Kinds, c.kinds) || !slices.Equal(got.Keywords, c.keywords) || got.Min != c.min || got.Max != c.max || got.Comma != c.comma {
			t.Errorf("Classify(%s) = %+v, want kinds %v keywords %v min %d max %d comma %v", name, got, c.kinds, c.keywords, c.min, c.max, c.comma)
		}
	}
}

func TestKeywords(t *testing.T) {
	props, types := propResolver(t), typeResolver(t)
	cases := []struct {
		src  string
		want []string
	}{
		{"<single-transition>#", []string{"ease", "linear", "none"}},
		{"[ <family-name> | <generic-family> ]#", []string{"monospace", "sans-serif", "serif"}},
		{"[ <display-outside> || <display-inside> ] | <display-listitem>", []string{"block", "flex", "flow", "flow-root", "grid", "inline", "list-item", "ruby", "run-in", "table"}},
		{"none | stylistic(swash) | <'unknown'>", []string{"none"}},
		{"<length> / auto", []string{"auto"}},
		{"<'ping'>+", []string{"auto", "none"}},
		{"<length>", nil},
	}
	for _, c := range cases {
		if got := Keywords(mustParse(t, c.src), props, types); !slices.Equal(got, c.want) {
			t.Errorf("Keywords(%s) = %v, want %v", c.src, got, c.want)
		}
	}
}

func TestClassifyRejectsMissingInputs(t *testing.T) {
	if got, ok := Classify(nil, nil, nil); ok {
		t.Errorf("Classify(nil, nil, nil) = %+v, want not tier 1", got)
	}
	if got := Keywords(nil, nil, nil); got != nil {
		t.Errorf("Keywords(nil, nil, nil) = %v, want nil", got)
	}
	if got, ok := Classify(mustParse(t, "<'padding-top'>"), nil, nil); ok {
		t.Errorf("Classify with nil resolvers = %+v, want not tier 1", got)
	}
	nilChild := &Node{Kind: Alt, Children: []*Node{{Kind: Keyword, Text: "auto", Min: 1, Max: 1}, nil}, Min: 1, Max: 1}
	for _, n := range []*Node{{Kind: Group, Min: 1, Max: 1}, nilChild} {
		if got, ok := Classify(n, nil, nil); ok {
			t.Errorf("Classify(%s) = %+v, want not tier 1", dump(n), got)
		}
	}
	if got := Keywords(nilChild, nil, nil); !slices.Equal(got, []string{"auto"}) {
		t.Errorf("Keywords(%s) = %v, want [auto]", dump(nilChild), got)
	}
}

func caching(resolve func(string) (*Node, bool)) func(string) (*Node, bool) {
	cache := map[string]*Node{}
	return func(name string) (*Node, bool) {
		if n, ok := cache[name]; ok {
			return n, true
		}
		n, ok := resolve(name)
		if ok {
			cache[name] = n
		}
		return n, ok
	}
}

func TestClassifyLeavesResolvedGrammarsUntouched(t *testing.T) {
	props, types := caching(propResolver(t)), caching(typeResolver(t))
	paddingTop, _ := props("padding-top")
	lineWidth, _ := types("line-width")
	before := dump(paddingTop) + " " + dump(lineWidth)
	for _, src := range []string{"<'padding'>", "<line-width>{1,4}"} {
		Classify(mustParse(t, src), props, types)
		Keywords(mustParse(t, src), props, types)
	}
	if after := dump(paddingTop) + " " + dump(lineWidth); after != before {
		t.Errorf("resolved grammars changed from %s to %s", before, after)
	}
}
