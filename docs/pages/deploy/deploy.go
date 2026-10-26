package deploy

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Handler and deployment"
	Route = "/deploy"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "deploy",
		Path:  Route,
		Title: Title,
		Lead:  "Configure app.Handler, serve static files, prerender pages, ship updates and publish a static site.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "The handler", Body: handler},
			{Title: "Static resources", Body: resources},
			{Title: "Prerendering and SEO", Body: prerendering},
			{Title: "Versions and updates", Body: updates},
			{Title: "Installing the app", Body: installing},
			{Title: "Static sites and GitHub Pages", Body: staticSites},
			{Title: "Libraries", Body: libraries},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("`app.Handler` is the server half of a WebGo app. It renders each routed page to HTML, and serves the manifest, the service worker, the JavaScript support files and everything under `/web/`. This page lists its fields, then follows an app from its static files through prerendering, updates and installation to a static site on GitHub Pages."),
		doc.P("Stylesheets, `WebDir` and `Build` are covered under ", doc.A("/applying#handler-and-build", "Handler and Build"), ". The smallest working server is in ", doc.A("/start", "Getting Started"), "."),
	}
}
