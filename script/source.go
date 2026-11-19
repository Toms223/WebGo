package script

import (
	"fmt"
	"regexp"

	"github.com/Toms223/WebGo/app"
)

var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)

type local struct{ name, path string }

type cdn struct{ registry, name, version string }

// Local serves the module name from the file at path, a /web/ path that must exist.
func Local(name, path string) app.ModuleSource { return local{name: name, path: path} }

// CDN serves the module name from esm.sh. The registry is npm, gh or jsr and the version is an exact
// semver without a v prefix. Handler.Build rejects anything else.
func CDN(registry, name, version string) app.ModuleSource {
	return cdn{registry: registry, name: name, version: version}
}

func (l local) ImportMap() (name, url string) { return l.name, l.path }

func (c cdn) ImportMap() (name, url string) {
	prefix := "https://esm.sh/"
	if c.registry != "npm" {
		prefix += c.registry + "/"
	}
	return c.name, prefix + c.name + "@" + c.version
}

func (c cdn) Check() error {
	switch c.registry {
	case "npm", "gh", "jsr":
	default:
		return fmt.Errorf("%w: %s: registry %q", app.ErrInvalidModule, c.name, c.registry)
	}
	if c.name == "" {
		return fmt.Errorf("%w: registry %q: empty name", app.ErrInvalidModule, c.registry)
	}
	if !versionPattern.MatchString(c.version) {
		return fmt.Errorf("%w: %s: version %q", app.ErrInvalidModule, c.name, c.version)
	}
	return nil
}
