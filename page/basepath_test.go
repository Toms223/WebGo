package page

import "testing"

func routes(paths ...string) []Screen {
	screens := make([]Screen, 0, len(paths))
	for _, path := range paths {
		screens = append(screens, newScreen(path, path))
	}
	return screens
}

func TestDeduceBasePathFindsTheMountFromADeepLink(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		routes []string
		want   string
	}{
		{"served at the root", "/markdown", []string{"/", "/markdown", "/state"}, ""},
		{"served at the root landing on root", "/", []string{"/", "/markdown", "/state"}, ""},
		{"mounted deep link", "/WebGo/markdown", []string{"/", "/markdown", "/state"}, "/WebGo"},
		{"mounted landing on the mount", "/WebGo", []string{"/", "/markdown", "/state"}, "/WebGo"},
		{"mounted landing with a trailing slash", "/WebGo/", []string{"/", "/markdown", "/state"}, "/WebGo"},
		{"nested mount", "/a/b/state", []string{"/", "/markdown", "/state"}, "/a/b"},
		{"longest route wins", "/mount/machine/state", []string{"/state", "/machine/state"}, "/mount"},
		{"no screens", "/WebGo/markdown", nil, "/WebGo/markdown"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := deduceBasePath(normalizePath(test.path), routes(test.routes...))
			if got != test.want {
				t.Errorf("expected base %q, got %q", test.want, got)
			}
		})
	}
}

func TestDeduceBasePathDoesNotMatchAPartialSegment(t *testing.T) {
	got := deduceBasePath(normalizePath("/webhooks"), routes("/", "/hooks"))
	if got == "/web" {
		t.Errorf("expected /hooks not to be treated as a suffix of /webhooks, got base %q", got)
	}
}

func TestTrimBasePathReturnsTheScreenRoute(t *testing.T) {
	cases := []struct {
		name string
		path string
		base string
		want string
	}{
		{"no mount", "/markdown", "", "/markdown"},
		{"no mount at root", "/", "", "/"},
		{"mounted page", "/WebGo/markdown", "/WebGo", "/markdown"},
		{"mounted root", "/WebGo", "/WebGo", "/"},
		{"path outside the mount", "/other/markdown", "/WebGo", "/other/markdown"},
		{"mount is a partial segment", "/webhooks", "/web", "/webhooks"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := trimBasePath(normalizePath(test.path), test.base)
			if got != test.want {
				t.Errorf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestDeduceAndTrimAreInverses(t *testing.T) {
	screens := routes("/", "/markdown", "/state")
	for _, mount := range []string{"", "/WebGo", "/a/b"} {
		for _, route := range []string{"/", "/markdown", "/state"} {
			path := mount + route
			if route == "/" && mount != "" {
				path = mount
			}
			base := deduceBasePath(normalizePath(path), screens)
			if got := trimBasePath(normalizePath(path), base); got != route {
				t.Errorf("mount %q route %q: expected %q, got %q", mount, route, route, got)
			}
		}
	}
}

func TestNormalizePathStripsTrailingSlashesButKeepsTheRoot(t *testing.T) {
	cases := map[string]string{
		"":             "/",
		"/":            "/",
		"//":           "/",
		"/markdown":    "/markdown",
		"/markdown/":   "/markdown",
		"markdown":     "/markdown",
		"/WebGo/state": "/WebGo/state",
	}

	for input, want := range cases {
		if got := normalizePath(input); got != want {
			t.Errorf("normalizePath(%q): expected %q, got %q", input, want, got)
		}
	}
}

func TestMountDeducesTheBasePathOnce(t *testing.T) {
	ctx, _ := newContext(t)
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{newScreen("/", "home")})

	base.OnMount(ctx)

	if !base.baseKnown {
		t.Fatal("expected mounting to record that the base path is known")
	}
	if base.BasePath() != "" {
		t.Errorf("expected an empty base path at the root, got %q", base.BasePath())
	}
}

func TestNavigatingNormalizesTheScreenRouteBeforeComparing(t *testing.T) {
	home := newScreen("//", "home")
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{home})
	ctx, engine := mountInTree(t, base)

	base.OnNav(ctx)
	engine.ConsumeAll()

	if base.selected != 0 {
		t.Fatalf("expected a route with a trailing slash to match the root path, got %d", base.selected)
	}
}

func TestNavigatingRestoresTheMountInTheUrl(t *testing.T) {
	home := newScreen("/", "home")
	markdown := newScreen("/markdown", "markdown")
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{home, markdown})
	base.basePath = "/WebGo"
	base.baseKnown = true
	ctx, engine := mountInTree(t, base)

	base.OnNav(ctx)
	engine.ConsumeAll()

	if base.selected != 0 {
		t.Fatalf("expected the root screen to be selected, got %d", base.selected)
	}
	if got := ctx.Page().URL().Path; got != "/WebGo" {
		t.Fatalf("expected the mount to be restored in the url, got %q", got)
	}
}

func TestNavigatingLeavesACorrectUrlAlone(t *testing.T) {
	home := newScreen("/", "home")
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{home})
	ctx, engine := mountInTree(t, base)

	base.OnNav(ctx)
	engine.ConsumeAll()

	if got := ctx.Page().URL().Path; got != "/" {
		t.Fatalf("expected an already correct url to be left alone, got %q", got)
	}
}

func TestBrowserPathIsTheInverseOfTrimBasePath(t *testing.T) {
	base := &Base[appState]{basePath: "/WebGo"}
	for _, route := range []string{"/", "/markdown", "/state"} {
		browser := base.browserPath(route)
		if got := trimBasePath(normalizePath(browser), base.basePath); got != route {
			t.Errorf("route %q: browserPath gave %q which trims back to %q", route, browser, got)
		}
	}
}
