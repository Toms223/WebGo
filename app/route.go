package app

import (
	"regexp"
	"sync"
)

type router struct {
	mu               sync.RWMutex
	routes           map[string]routeEntry
	routesWithRegexp []regexpRoute
	served           Composer
}

type routeEntry struct {
	factory func() Composer
	cached  bool
}

func makeRouter() router {
	return router{
		routes: make(map[string]routeEntry),
	}
}

func (r *router) route(path string, newComponent func() Composer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes[path] = routeEntry{factory: newComponent}
}

func (r *router) serve(paths []string, build func() Composer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.served = nil
	for _, path := range paths {
		r.routes[path] = routeEntry{factory: build, cached: true}
	}
}

func (r *router) routeWithRegexp(pattern string, newComponent func() Composer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routesWithRegexp = append(r.routesWithRegexp, regexpRoute{
		regexp:       regexp.MustCompile(pattern),
		newComponent: newComponent,
	})
}

func (r *router) routed(path string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, routed := r.routes[path]; routed {
		return true
	}

	for _, rwr := range r.routesWithRegexp {
		if rwr.regexp.MatchString(path) {
			return true
		}
	}

	return false
}

func (r *router) createComponent(path string) (Composer, bool) {
	r.mu.RLock()
	entry, routed := r.routes[path]
	regexps := r.routesWithRegexp
	r.mu.RUnlock()

	if routed {
		if !entry.cached || !isClient {
			return entry.factory(), true
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.served == nil {
			r.served = entry.factory()
		}
		return r.served, true
	}

	for _, rwr := range regexps {
		if rwr.regexp.MatchString(path) {
			return rwr.newComponent(), true
		}
	}

	return nil, false
}

type regexpRoute struct {
	regexp       *regexp.Regexp
	newComponent func() Composer
}
