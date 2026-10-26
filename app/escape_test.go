package app

import (
	"strings"
	"testing"
)

func TestAttributeValuesAreHTMLEscaped(t *testing.T) {
	got := HTMLString(Div().Title(`a"b<c>&d'e`))
	if want := `<div title="a&#34;b&lt;c&gt;&amp;d&#39;e"></div>`; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestAttributeInjectionIsInert(t *testing.T) {
	got := HTMLString(Div().Title(`"><script>window.pwned=1</script>`))
	if strings.Contains(got, "<script") {
		t.Fatalf("attribute value broke out of its quotes: %s", got)
	}
}

func TestAttributesAreSortedByName(t *testing.T) {
	ui := Div().Title("t").ID("b").Class("a").DataSet("x", 1)
	want := `<div class="a" data-x="1" id="b" title="t"></div>`
	for range 50 {
		if got := HTMLString(ui); got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	}
}

func TestRawStylesAreDeterministic(t *testing.T) {
	styles := map[string]string{"color": "red", "margin": "0", "padding": "1px", "width": "2px", "height": "3px"}
	want := `<div style="color:red;height:3px;margin:0;padding:1px;width:2px"></div>`
	for range 100 {
		if got := HTMLString(Div().RawStyles(styles)); got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	}
}
