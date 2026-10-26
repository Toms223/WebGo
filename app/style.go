package app

import "strings"

// ClassRule supplies class names to an element's Style method; the rules of the
// style package implement it. A nil rule or one without class names is skipped
// and logged.
type ClassRule interface {
	ClassNames() []string
}

// Styler is the styling API shared by the HTML elements, where E is the element
// type. Style adds class rules; RawStyle, RawStylef and RawStyles set inline
// CSS, with RawStyles applying keys in sorted order.
type Styler[E any] interface {
	UI
	Class(v ...string) E
	Style(rules ...ClassRule) E
	RawStyle(k, v string) E
	RawStylef(k, format string, v ...any) E
	RawStyles(s map[string]string) E
}

func (e *htmlElement) addClassRules(rules []ClassRule) {
	for _, r := range rules {
		var names []string
		if r != nil {
			names = r.ClassNames()
		}
		if len(names) == 0 {
			Log("app: Style skipped a rule that has no class names")
			continue
		}
		e.setAttr("class", strings.Join(names, " "))
	}
}
