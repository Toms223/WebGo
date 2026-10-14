package demos

import (
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/style"
)

func Items() []style.Item {
	return []style.Item{DemoCompactBase, DemoPulseBox, DemoBannerHidden, DemoOutlineBase}
}

func Dynamic() []style.Item {
	return []style.Item{DemoCompact, DemoAccent, DemoPulse, DemoBanner, DemoOutline}
}

func model[T any](key string, value *T) *data.ViewModel[T] {
	m, err := data.NewViewModel(key, value)
	if err != nil {
		panic(err)
	}
	return m
}
