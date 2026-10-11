package pins

import "time"

const (
	WebrefCSSVersion = "8.7.6"
	BCDVersion       = "8.1.5"
	Date             = "2026-10-10"
	FloorMonths      = 30
)

func FloorDate() time.Time {
	pinned, _ := time.Parse(time.DateOnly, Date)
	return pinned.AddDate(0, -FloorMonths, 0)
}
