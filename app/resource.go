package app

import (
	"net/http"
	"strings"
)

// ResourceResolver maps /web/ paths to where the files are served from. Other
// paths are returned unchanged.
type ResourceResolver interface {
	Resolve(string) string
}

// LocalDir serves /web/ from directory/web on disk. LocalDir("") is the Handler
// default and also sets the default Handler.WebDir.
func LocalDir(directory string) ResourceResolver {
	directory = strings.TrimRight(directory, "/")
	return localResourceResolver{
		Handler:   http.FileServer(http.Dir(directory)),
		directory: directory,
	}
}

type localResourceResolver struct {
	http.Handler
	directory string
}

func (r localResourceResolver) Resolve(location string) string {
	if location == "/" || location == "" {
		return "/"
	}
	if remoteLocation(location) || !webLocation(location) {
		return location
	}
	return "/" + strings.Trim(location, "/")
}

// RemoteBucket resolves /web/ paths under url, such as a storage bucket or CDN.
// The Handler then no longer serves /web/ itself.
func RemoteBucket(url string) ResourceResolver {
	return remoteResourceResolver{
		url: strings.Trim(url, "/"),
	}
}

type remoteResourceResolver struct {
	url string
}

func (r remoteResourceResolver) Resolve(location string) string {
	if location == "/" || location == "" {
		return "/"
	}
	if remoteLocation(location) || !webLocation(location) {
		return location
	}
	return r.url + "/" + strings.Trim(location, "/")
}

// PrefixedLocation resolves paths under prefix, for an app hosted in a
// subdirectory: /web/main.css becomes prefix/web/main.css and / becomes prefix.
// A Handler using it serves /web/ from ./web, as with LocalDir("").
func PrefixedLocation(prefix string) ResourceResolver {
	return prefixedResourceResolver{
		localResourceResolver: LocalDir("").(localResourceResolver),
		prefix:                prefix,
	}
}

type prefixedResourceResolver struct {
	localResourceResolver
	prefix string
}

func (r prefixedResourceResolver) Resolve(location string) string {
	if remoteLocation(location) {
		return location
	}
	location = r.localResourceResolver.Resolve(location)
	if location == "/" {
		return strings.TrimRight(r.prefix, "/")
	}
	return strings.TrimRight(r.prefix, "/") + location
}

// GitHubPages is PrefixedLocation("/"+repositoryName), for a site published at
// user.github.io/repositoryName.
func GitHubPages(repositoryName string) ResourceResolver {
	return PrefixedLocation("/" + strings.Trim(repositoryName, "/"))
}

func clientResourceResolver(resourcesLocation string) func(string) string {
	return func(location string) string {
		if remoteLocation(location) || !webLocation(location) {
			return location
		}
		location = strings.Trim(location, "/")
		return resourcesLocation + "/" + strings.TrimPrefix(location, "web/")
	}
}

func resolveOGResource(domain string, location string) string {
	if remoteLocation(location) {
		return location
	}
	return "https://" + domain + strings.TrimRight("/"+strings.Trim(location, "/"), "/")
}

func remoteLocation(location string) bool {
	return strings.HasPrefix(location, "https://") ||
		strings.HasPrefix(location, "http://")
}

func webLocation(location string) bool {
	return strings.HasPrefix(location, "/web/") ||
		location == "/web" ||
		strings.HasPrefix(location, "web/") ||
		location == "web"
}

// ProxyResource serves the file at ResourcePath, a /web/ path, at Path, such as
// /robots.txt.
type ProxyResource struct {
	Path string

	ResourcePath string
}
