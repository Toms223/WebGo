package pages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/docs/components/topic"
)

func TestEveryScreenHasItsDataFile(t *testing.T) {
	screens, err := Screens()
	if err != nil {
		t.Fatalf("building the screens failed: %v", err)
	}

	for _, screen := range screens {
		topicScreen, ok := screen.(*topic.Screen)
		if !ok {
			t.Fatalf("screen %q is a %T, not a *topic.Screen: cannot determine its data source", screen.Route(), screen)
			continue
		}

		source := topicScreen.Source()
		name := strings.TrimPrefix(source, "/data/")
		if name == source || name == "" {
			t.Errorf("screen %q has an unexpected source %q, expected it to start with /data/", screen.Route(), source)
			continue
		}

		path := filepath.Join("..", "web", "data", name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("screen %q expects %s, which is missing: %v", screen.Route(), path, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("screen %q expects %s, which is empty", screen.Route(), path)
		}
	}
}
