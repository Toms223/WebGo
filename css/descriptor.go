package css

// Descriptor is one declaration inside @font-face or @property, made by a descriptor function such as
// Src or Syntax.
type Descriptor interface {
	Name() string
	Value() string
	descriptor()
}

type descriptorDecl struct {
	name  string
	value string
}

func (d descriptorDecl) Name() string  { return d.name }
func (d descriptorDecl) Value() string { return d.value }
func (d descriptorDecl) descriptor()   {}

func declareDescriptor(name, value string) Descriptor {
	return descriptorDecl{name: name, value: value}
}

type FormatValue interface {
	Value
	CSSFormat()
}

func (RawValue) CSSFormat() {}

func Format(f FormatValue) Value {
	return Raw("format(" + quote(valueText(f)) + ")")
}

type SyntaxValue string

func (s SyntaxValue) CSS() string { return quote(string(s)) }

func (SyntaxValue) DescriptorSyntax() {}

func SyntaxRaw(s string) SyntaxValue { return SyntaxValue(s) }

const (
	SyntaxLength            SyntaxValue = "<length>"
	SyntaxNumber            SyntaxValue = "<number>"
	SyntaxPercentage        SyntaxValue = "<percentage>"
	SyntaxLengthPercentage  SyntaxValue = "<length-percentage>"
	SyntaxColor             SyntaxValue = "<color>"
	SyntaxImage             SyntaxValue = "<image>"
	SyntaxUrl               SyntaxValue = "<url>"
	SyntaxInteger           SyntaxValue = "<integer>"
	SyntaxAngle             SyntaxValue = "<angle>"
	SyntaxTime              SyntaxValue = "<time>"
	SyntaxResolution        SyntaxValue = "<resolution>"
	SyntaxTransformFunction SyntaxValue = "<transform-function>"
	SyntaxTransformList     SyntaxValue = "<transform-list>"
	SyntaxCustomIdent       SyntaxValue = "<custom-ident>"
	SyntaxAny               SyntaxValue = "*"
)
