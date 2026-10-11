package pins

import (
	"regexp"
	"testing"
	"time"
)

func TestVersionsAreExact(t *testing.T) {
	exact := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	for name, version := range map[string]string{
		"WebrefCSSVersion": WebrefCSSVersion,
		"BCDVersion":       BCDVersion,
	} {
		if !exact.MatchString(version) {
			t.Errorf("%s = %q, want an exact x.y.z version", name, version)
		}
	}
}

func TestFloorDateIsFloorMonthsBeforeDate(t *testing.T) {
	pin, err := time.Parse(time.DateOnly, Date)
	if err != nil {
		t.Fatalf("Date %q is not YYYY-MM-DD: %v", Date, err)
	}
	want := pin.AddDate(0, -FloorMonths, 0)
	got := FloorDate()
	if !got.Equal(want) {
		t.Fatalf("FloorDate() = %s, want %s", got.Format(time.DateOnly), want.Format(time.DateOnly))
	}
	if !got.Before(pin) {
		t.Fatalf("FloorDate() = %s is not before Date %s", got.Format(time.DateOnly), Date)
	}
}
