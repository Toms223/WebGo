package demos

import (
	"strings"

	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func All() map[string]func() app.UI {
	return map[string]func() app.UI{
		"card":        CardDemo,
		"selectors":   SelectorsDemo,
		"pseudo":      PseudoDemo,
		"nesting":     NestingDemo,
		"hover-media": HoverMediaDemo,
		"values":      ValuesDemo,
		"media":       MediaDemo,
		"container":   ContainerDemo,
		"supports":    SupportsDemo,
		"keyframes":   KeyframesDemo,
		"property":    PropertyDemo,
		"layers":      LayersDemo,
		"important":   ImportantDemo,
	}
}

func Items() []style.Item {
	return []style.Item{
		Card,
		ListItem,
		Button,
		Panel,
		Tile,
		Swatch,
		Wide,
		Frame, NarrowFrame, Row, Cell,
		Mixed,
		Pulse, Pulsing,
		Angle, Turning,
		Under, Over,
		Quiet, Loud,
	}
}

func Pins() map[string]string {
	return map[string]string{
		"card":        Card.CSS(),
		"selectors":   ListItem.CSS(),
		"pseudo":      Button.CSS(),
		"nesting":     Panel.CSS(),
		"hover-media": Tile.CSS(),
		"values":      Swatch.CSS(),
		"media":       Wide.CSS(),
		"container":   joined(Frame.CSS(), NarrowFrame.CSS(), Row.CSS(), Cell.CSS()),
		"supports":    Mixed.CSS(),
		"keyframes":   joined(Pulse.CSS(), Pulsing.CSS()),
		"property":    joined(Angle.CSS(), Turning.CSS()),
		"layers":      joined(Under.CSS(), Over.CSS()),
		"important":   joined(Quiet.CSS(), Loud.CSS()),
	}
}

func joined(texts ...string) string {
	return strings.Join(texts, "\n")
}
