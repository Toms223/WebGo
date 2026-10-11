package container

import (
	"reflect"
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
)

func TestGeneratedFeatures(t *testing.T) {
	cases := []struct {
		query Query
		want  string
	}{
		{MinWidth(css.Px(400)), "(min-width: 400px)"},
		{MaxInlineSize(css.Rem(30)), "(max-inline-size: 30rem)"},
		{WidthBetween(css.Px(1), css.Px(2)), "(1px <= width <= 2px)"},
		{Orientation(kw.Portrait), "(orientation: portrait)"},
		{Orientation(css.Raw("landscape")), "(orientation: landscape)"},
		{Named("side").And(MinWidth(css.Px(400))), "side (min-width: 400px)"},
	}
	for _, c := range cases {
		if err := c.query.Err(); err != nil {
			t.Errorf("%s: Err() = %v", c.want, err)
		}
		if got := c.query.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}

func TestContainerMarkerAbsent(t *testing.T) {
	if _, ok := reflect.TypeOf(css.VarValue{}).MethodByName("ContainerOrientation"); ok {
		t.Error("css.VarValue has ContainerOrientation, so Orientation(css.Var(...)) would compile")
	}
	if _, ok := reflect.TypeOf(css.RawValue{}).MethodByName("ContainerOrientation"); !ok {
		t.Error("css.RawValue lacks ContainerOrientation, so Orientation(css.Raw(...)) would not compile")
	}
}
