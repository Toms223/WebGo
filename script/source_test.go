package script_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/script"
)

type mod struct{ src app.ModuleSource }

func (m mod) Source() app.ModuleSource { return m.src }

func TestLocalReturnsNameAndPath(t *testing.T) {
	name, url := script.Local("m", "/web/m.js").ImportMap()
	if name != "m" || url != "/web/m.js" {
		t.Fatalf("got %q %q", name, url)
	}
}

func TestCDNBuildsEsmShURLs(t *testing.T) {
	tests := []struct{ registry, name, version, url string }{
		{"npm", "@datadog/browser-rum", "5.20.0", "https://esm.sh/@datadog/browser-rum@5.20.0"},
		{"gh", "user/repo", "1.2.3", "https://esm.sh/gh/user/repo@1.2.3"},
		{"jsr", "@std/path", "1.0.0", "https://esm.sh/jsr/@std/path@1.0.0"},
	}
	for _, tt := range tests {
		name, url := script.CDN(tt.registry, tt.name, tt.version).ImportMap()
		if name != tt.name || url != tt.url {
			t.Errorf("%s: got %q %q", tt.registry, name, url)
		}
	}
}

func checkOf(t *testing.T, src app.ModuleSource) error {
	t.Helper()
	c, ok := src.(interface{ Check() error })
	if !ok {
		t.Fatal("no Check method")
	}
	return c.Check()
}

func TestCDNCheck(t *testing.T) {
	for _, v := range [][3]string{{"npm", "x", "1.0.0"}, {"npm", "x", "1.0.0-beta.1"}} {
		if err := checkOf(t, script.CDN(v[0], v[1], v[2])); err != nil {
			t.Errorf("%v: %v", v, err)
		}
	}
	bad := [][3]string{
		{"pypi", "x", "1.0.0"}, {"npm", "", "1.0.0"}, {"npm", "x", "^1.0.0"}, {"npm", "x", "v1.0.0"},
		{"npm", "x", "latest"}, {"npm", "x", "1.0"}, {"npm", "x", ""},
	}
	for _, v := range bad {
		err := checkOf(t, script.CDN(v[0], v[1], v[2]))
		if !errors.Is(err, app.ErrInvalidModule) {
			t.Errorf("%v: got %v", v, err)
			continue
		}
		if !strings.Contains(err.Error(), "x") && v[1] != "" {
			t.Errorf("%v: no name in %q", v, err)
		}
		offender := v[2]
		if v[0] == "pypi" {
			offender = v[0]
		} else if v[1] == "" {
			offender = ""
		}
		if offender != "" && !strings.Contains(err.Error(), offender) {
			t.Errorf("%v: no value in %q", v, err)
		}
	}
}

func TestBuildRejectsBadCDNCoordinates(t *testing.T) {
	h := &app.Handler{Imports: []app.Module{mod{script.CDN("npm", "x", "latest")}}, WebDir: t.TempDir()}
	if err := h.Build(); !errors.Is(err, app.ErrInvalidModule) || strings.Count(err.Error(), "module source is invalid") != 1 {
		t.Fatalf("got %v", err)
	}
}
