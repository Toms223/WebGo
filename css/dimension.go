package css

type Length struct {
	amount float64
	unit   string
}

func Px(n float64) Length   { return Length{amount: n, unit: "px"} }
func Rem(n float64) Length  { return Length{amount: n, unit: "rem"} }
func Em(n float64) Length   { return Length{amount: n, unit: "em"} }
func Ch(n float64) Length   { return Length{amount: n, unit: "ch"} }
func Ex(n float64) Length   { return Length{amount: n, unit: "ex"} }
func Vw(n float64) Length   { return Length{amount: n, unit: "vw"} }
func Vh(n float64) Length   { return Length{amount: n, unit: "vh"} }
func Vmin(n float64) Length { return Length{amount: n, unit: "vmin"} }
func Vmax(n float64) Length { return Length{amount: n, unit: "vmax"} }
func Dvh(n float64) Length  { return Length{amount: n, unit: "dvh"} }
func Svh(n float64) Length  { return Length{amount: n, unit: "svh"} }
func Lvh(n float64) Length  { return Length{amount: n, unit: "lvh"} }
func Dvw(n float64) Length  { return Length{amount: n, unit: "dvw"} }
func Svw(n float64) Length  { return Length{amount: n, unit: "svw"} }
func Lvw(n float64) Length  { return Length{amount: n, unit: "lvw"} }
func Cm(n float64) Length   { return Length{amount: n, unit: "cm"} }
func Mm(n float64) Length   { return Length{amount: n, unit: "mm"} }
func In(n float64) Length   { return Length{amount: n, unit: "in"} }
func Pt(n float64) Length   { return Length{amount: n, unit: "pt"} }
func Pc(n float64) Length   { return Length{amount: n, unit: "pc"} }
func Q(n float64) Length    { return Length{amount: n, unit: "Q"} }

func (l Length) CSS() string { return formatNumber(l.amount) + l.unit }

type Angle struct {
	amount float64
	unit   string
}

func Deg(n float64) Angle  { return Angle{amount: n, unit: "deg"} }
func Rad(n float64) Angle  { return Angle{amount: n, unit: "rad"} }
func Grad(n float64) Angle { return Angle{amount: n, unit: "grad"} }
func Turn(n float64) Angle { return Angle{amount: n, unit: "turn"} }

func (a Angle) CSS() string { return formatNumber(a.amount) + a.unit }

type Time struct {
	amount float64
	unit   string
}

func S(n float64) Time  { return Time{amount: n, unit: "s"} }
func Ms(n float64) Time { return Time{amount: n, unit: "ms"} }

func (t Time) CSS() string { return formatNumber(t.amount) + t.unit }

type Resolution struct {
	amount float64
	unit   string
}

func Dpi(n float64) Resolution  { return Resolution{amount: n, unit: "dpi"} }
func Dpcm(n float64) Resolution { return Resolution{amount: n, unit: "dpcm"} }
func Dppx(n float64) Resolution { return Resolution{amount: n, unit: "dppx"} }

func (r Resolution) CSS() string { return formatNumber(r.amount) + r.unit }

type Fraction struct{ amount float64 }

func Fr(n float64) Fraction { return Fraction{amount: n} }

func (f Fraction) CSS() string { return formatNumber(f.amount) + "fr" }
