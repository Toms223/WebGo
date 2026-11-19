package deploy

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func local() []app.UI {
	return []app.UI{
		doc.P("The `dev` package runs the static site on your machine and rebuilds it when you save a file. It is one function:"),
		doc.Go(`func RunDevServer(ctx context.Context, tempDir, addr string, generate func(dir string) *exec.Cmd) error`),
		doc.List(
			doc.Item("`tempDir` is the folder the site is built into and served from."),
			doc.Item("`addr` is the address it listens on, such as `localhost:8000`."),
			doc.Item("`generate` returns the command that writes your static site into `dir`. It must use `dir/web` as the handler's `WebDir`."),
		),
		doc.P("Call it from the same `main` that generates the site, in the folder that holds that `main`:"),
		doc.Go(`func main() {
	app.Route("/", func() app.Composer { return &page{} })
	app.RunWhenOnBrowser()

	static := flag.String("static", "", "write the static site to this directory")
	flag.Parse()

	if *static != "" {
		handler := &app.Handler{Name: "Site", WebDir: filepath.Join(*static, "web")}
		if err := app.GenerateStaticWebsite(*static, handler); err != nil {
			log.Fatal(err)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := dev.RunDevServer(ctx, ".dev", "localhost:8000", func(dir string) *exec.Cmd {
		return exec.Command("go", "run", ".", "-static", dir)
	})
	if err != nil {
		log.Fatal(err)
	}
}`),
		doc.Sh(`go run .`),
		doc.H("What a build does"),
		doc.Ordered(
			doc.Item("It copies your `web` folder into a staging folder next to `tempDir`. A project with no `web` folder is fine."),
			doc.Item("It builds `app.wasm` into the staged `web` folder with `GOOS=js GOARCH=wasm go build`."),
			doc.Item("It runs `generate` on the staging folder. The wasm is already there, so the handler's computed version includes it."),
			doc.Item("It swaps the staging folder in as `tempDir`."),
		),
		doc.P("The first build runs before the server starts listening. If it fails, `RunDevServer` returns the error. A later build that fails is logged, and the server keeps serving the last build that worked."),
		doc.H("What triggers a rebuild"),
		doc.P("The working directory is watched, with its subfolders. A change to a `.go` file, to `go.mod` or `go.sum`, or to any file under a `web` folder starts a build. Changes within one second count as one. `tempDir` is not watched, so it can live inside the project; add it to `.gitignore`."),
		doc.P("Every build produces a new version, so an open tab picks it up the way it picks up a deploy, under ", doc.A("#versions-and-updates", "Versions and updates"), ". Switching back to the tab is the quickest trigger."),
		doc.H("What it serves"),
		doc.List(
			doc.Item("A file that exists is served as it is."),
			doc.Item("A path with no extension gets its `.html` file: `/about` is `about.html`. When there is none, it gets `index.html`, and the app routes the URL in the browser."),
			doc.Item("A missing path with an extension, such as `/web/logo.png`, is a 404."),
			doc.Item("Every response has `Cache-Control: no-store`."),
		),
		doc.H("Limits"),
		doc.List(
			doc.Item("Only the working directory is watched. A module you point at with a `replace` directive elsewhere on disk does not trigger a build."),
			doc.Item("The site is served from the root. To try a path prefix, generate with `app.GitHubPages` and serve the folder under that prefix yourself."),
			doc.Item("In the browser build `RunDevServer` returns an error and does nothing, so one `main` can import the package for both targets."),
			doc.Item("`RunDevServer` returns `nil` when `ctx` is done."),
		),
		doc.P("This site runs the same way, from its `docs` folder: `go run . dev`."),
	}
}
