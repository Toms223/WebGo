package css

import "strconv"

type Nth struct {
	a, b    int
	keyword string
}

var (
	Odd  = Nth{keyword: "odd"}
	Even = Nth{keyword: "even"}
)

func An(a, b int) Nth { return Nth{a: a, b: b} }

func (n Nth) CSS() string {
	if n.keyword != "" {
		return n.keyword
	}
	if n.a == 0 {
		return strconv.Itoa(n.b)
	}
	return coefficient(n.a) + offset(n.b)
}

func coefficient(a int) string {
	switch a {
	case 1:
		return "n"
	case -1:
		return "-n"
	}
	return strconv.Itoa(a) + "n"
}

func offset(b int) string {
	switch {
	case b > 0:
		return "+" + strconv.Itoa(b)
	case b < 0:
		return strconv.Itoa(b)
	}
	return ""
}
