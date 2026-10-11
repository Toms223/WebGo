package css

import "testing"

func TestDescriptorRenders(t *testing.T) {
	d := declareDescriptor("src", `url("a.woff2") format("woff2")`)
	if d.Name() != "src" {
		t.Errorf("Name() = %q, want %q", d.Name(), "src")
	}
	if want := `url("a.woff2") format("woff2")`; d.Value() != want {
		t.Errorf("Value() = %q, want %q", d.Value(), want)
	}
}

func TestFormat(t *testing.T) {
	if got := Format(Raw("woff2")).CSS(); got != `format("woff2")` {
		t.Errorf("Format = %q, want %q", got, `format("woff2")`)
	}
}

func TestSyntaxValues(t *testing.T) {
	cases := []struct {
		v    SyntaxValue
		want string
	}{
		{SyntaxLength, `"<length>"`},
		{SyntaxNumber, `"<number>"`},
		{SyntaxPercentage, `"<percentage>"`},
		{SyntaxLengthPercentage, `"<length-percentage>"`},
		{SyntaxColor, `"<color>"`},
		{SyntaxImage, `"<image>"`},
		{SyntaxUrl, `"<url>"`},
		{SyntaxInteger, `"<integer>"`},
		{SyntaxAngle, `"<angle>"`},
		{SyntaxTime, `"<time>"`},
		{SyntaxResolution, `"<resolution>"`},
		{SyntaxTransformFunction, `"<transform-function>"`},
		{SyntaxTransformList, `"<transform-list>"`},
		{SyntaxCustomIdent, `"<custom-ident>"`},
		{SyntaxAny, `"*"`},
		{SyntaxRaw("<length> | auto"), `"<length> | auto"`},
	}
	for _, c := range cases {
		if got := c.v.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}
