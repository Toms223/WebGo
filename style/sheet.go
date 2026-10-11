package style

import (
	"errors"
	"fmt"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var errNoSheet = errors.New("the browser created no sheet for the style element (a Content-Security-Policy that blocks inline styles does this)")

type sheet interface {
	insert(text string, index int) bool
	remove(index int)
	length() int
}

type jsSheet struct {
	document app.Value
	value    app.Value
	created  bool
}

func (s *jsSheet) insert(text string, index int) bool {
	if err := s.insertRule(text, index); err != nil {
		app.Logf("style: cannot insert rule %q: %v", text, err)
		return false
	}
	return true
}

func (s *jsSheet) remove(index int) {
	if err := s.deleteRule(index); err != nil {
		app.Logf("style: cannot delete rule %d: %v", index, err)
	}
}

func (s *jsSheet) length() int {
	n, err := s.countRules()
	if err != nil {
		app.Logf("style: cannot count rules: %v", err)
	}
	return n
}

func (s *jsSheet) insertRule(text string, index int) (err error) {
	defer recoverInto(&err)
	s.create()
	if !s.usable() {
		return errNoSheet
	}
	s.value.Call("insertRule", text, index)
	return nil
}

func (s *jsSheet) deleteRule(index int) (err error) {
	defer recoverInto(&err)
	if s.usable() {
		s.value.Call("deleteRule", index)
	}
	return nil
}

func (s *jsSheet) countRules() (n int, err error) {
	defer recoverInto(&err)
	if !s.usable() {
		return 0, nil
	}
	return s.value.Get("cssRules").Length(), nil
}

func recoverInto(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("%v", r)
	}
}

func (s *jsSheet) usable() bool {
	return s.created && s.value.Truthy()
}

func (s *jsSheet) create() {
	if s.created {
		return
	}
	doc := s.document
	if doc == nil {
		doc = app.Window().Get("document")
	}
	el := doc.Call("createElement", "style")
	el.Call("setAttribute", "data-webgo", "dynamic")
	doc.Get("head").Call("appendChild", el)
	s.value = el.Get("sheet")
	s.created = true
}
