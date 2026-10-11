package demos

import (
	"strings"

	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func All() map[string]func() app.UI {
	return map[string]func() app.UI{
		"when":      WhenDemo,
		"compute":   ComputeDemo,
		"frames":    FramesDemo,
		"initially": InitiallyDemo,
		"on":        OnDemo,
	}
}

func Items() []style.Item {
	return []style.Item{DemoCompactBase, DemoPulseBox, DemoBannerHidden, DemoOutlineBase}
}

func Dynamic() []style.Item {
	return []style.Item{DemoCompact, DemoAccent, DemoPulse, DemoBanner, DemoOutline}
}

func Pins() map[string]string {
	return map[string]string{
		"when":      joined(DemoCompactBase.CSS(), DemoCompact.CSS()),
		"compute":   DemoAccent.CSSFor(AccentKey, startAccent()),
		"frames":    joined(DemoPulse.CSSFor(SpeedKey, startSpeed()), DemoPulseBox.CSS()),
		"initially": joined(DemoBannerHidden.CSS(), DemoBanner.CSS()),
		"on":        joined(DemoOutlineBase.CSS(), DemoOutline.CSS()),
	}
}

func joined(texts ...string) string {
	return strings.Join(texts, "\n")
}

func model[T any](key string, value *T) *data.ViewModel[T] {
	m, err := data.NewViewModel(key, value)
	if err != nil {
		panic(err)
	}
	return m
}
