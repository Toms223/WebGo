package main

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedFilesAreCurrent(t *testing.T) {
	dir := t.TempDir()
	generate(dir)
	for _, name := range []string{"html_gen.go", "html_gen_test.go"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		got, err := format.Source(raw)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		want, err := os.ReadFile(filepath.Join("..", "..", "app", name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s differs from the generator output; run go generate ./app", name)
		}
	}
}

func TestOnlyStylableTagsTakeRules(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "app", "html_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	if n := bytes.Count(src, []byte("var _ Styler[HTML")); n != 98 {
		t.Errorf("got %d Styler assertions, want 98", n)
	}
	for name := range unstyled {
		if bytes.Contains(src, []byte("func (e *html"+name+") Style(rules")) {
			t.Errorf("%s must not take style rules", name)
		}
		if !bytes.Contains(src, []byte("func (e *html"+name+") RawStyle(")) {
			t.Errorf("%s is missing RawStyle; is it a real tag name?", name)
		}
	}
}
