package demos

import (
	"github.com/Toms223/WebGo/style"
)

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
