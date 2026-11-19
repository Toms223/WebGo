package featureflags

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Feature flags"
	Route = "/feature-flags"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "feature-flags",
		Path:  Route,
		Title: Title,
		Lead:  "Evaluate OpenFeature flags from Go, with the behaviour of the official Web SDK.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Setup", Body: setup},
			{Title: "Evaluating", Body: evaluating},
			{Title: "Re-rendering on change", Body: rerender},
			{Title: "Providers", Body: providers},
			{Title: "Flipt behaviour", Body: fliptBehaviour},
			{Title: "Server config", Body: serverConfig},
			{Title: "Parity", Body: parity},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("OpenFeature is a vendor-neutral API for feature flags. The `openfeature` package is a Go port of its Web SDK: one global API, clients per domain, hooks, events and an evaluation context. A provider connects it to a flag service."),
		doc.P("There is no live demo on this page, because a real one needs a running Flipt server."),
	}
}

func setup() []app.UI {
	return []app.UI{
		doc.P("List the Flipt modules on the handler, then create the provider and bind it. Binding waits for the provider to initialize, so do it inside `ctx.Async`:"),
		doc.Go(`h := &app.Handler{Imports: flipt.Modules()}

func (c *nav) OnMount(ctx app.Context) {
	p, ok := flipt.New(ctx, flipt.Options{URL: "https://flags.example.com", ClientToken: token})
	if !ok {
		return
	}
	ctx.Async(func() {
		err := openfeature.SetProviderAndWait(p, openfeature.WithContext(openfeature.Context{TargetingKey: userID}))
		ctx.Dispatch(func(app.Context) { c.err = err })
	})
}`),
		doc.P("`flipt.New` reports `false` on the server and when the module did not load. `flipt.LocalModules(dir)` serves the same four files from `/web/<dir>/` for self-hosting."),
		doc.P("`SetProvider` is the non-blocking form. It returns only validation errors and logs an initialization failure."),
	}
}

func evaluating() []app.UI {
	return []app.UI{
		doc.P("A client evaluates one flag per call and returns the default when anything goes wrong:"),
		doc.Go(`c := openfeature.GetClient()

on := c.Boolean("new-nav", false)
label := c.String("cta", "Buy")
ratio := c.Number("ratio", 0.5)
limit := c.Int("limit", 10)

if o, ok := c.Object("theme", nil).(openfeature.Object); ok {
	accent, _ := o.Get("accent")
}

d := c.BooleanDetails("new-nav", false)
// d.Value, d.Variant, d.Reason, d.ErrorCode, d.ErrorMessage, d.FlagMetadata`),
		doc.List(
			doc.Item("`Int` is a number evaluation. A result that is not whole is `TYPE_MISMATCH`."),
			doc.Item("`Object` values from a JS provider are an `openfeature.Object`, an ordered list of `Field` values with a `Get` method."),
			doc.Item("`GetClientVersion(version, domain...)` and `GetClient(domain...)` pick a domain. `ForDomain` does the same for `SetProvider`."),
			doc.Item("`openfeature.ProviderStatus()` and `client.ProviderStatus()` report `Status`: `NOT_READY`, `READY`, `ERROR`, `STALE`, `FATAL` or `RECONCILING`."),
		),
	}
}

func rerender() []app.UI {
	return []app.UI{
		doc.P("Evaluations are plain calls, so a component has to ask to render again when flags change. `EventContextChanged` fires after `SetContext` settles, and `EventConfigurationChanged` when a provider reports new flags. `AddHandler` returns a function that removes the handler:"),
		doc.Go(`func (c *nav) OnMount(ctx app.Context) {
	c.client = openfeature.GetClient()
	rerender := func(openfeature.EventDetails) { ctx.Dispatch(func(app.Context) {}) }
	stopContext := c.client.AddHandler(openfeature.EventContextChanged, rerender)
	stopConfig := c.client.AddHandler(openfeature.EventConfigurationChanged, rerender)
	c.stop = func() { stopContext(); stopConfig() }
}

func (c *nav) OnDismount() { c.stop() }

func (c *nav) Render() app.UI {
	return app.If(c.client.Boolean("new-nav", false), func() app.UI { return newNav() })
}`),
		doc.P("Change the user with `openfeature.SetContext(openfeature.Context{TargetingKey: id})`. It blocks until the provider settles, so call it from `ctx.Async`."),
	}
}

func providers() []app.UI {
	return []app.UI{
		doc.List(
			doc.Item("`memprovider`: flags held in memory, a port of the Web SDK's in-memory provider. Use it in tests and for local development."),
			doc.Item("`jsprovider`: `jsprovider.New(value)` wraps any JavaScript OpenFeature web provider you loaded as a module."),
			doc.Item("`flipt`: the official Flipt web provider, loaded and wrapped for you."),
		),
		doc.Go(`p := memprovider.New(memprovider.Flag{
	Key:            "new-nav",
	Variants:       map[string]any{"on": true, "off": false},
	DefaultVariant: "on",
})
openfeature.SetProviderAndWait(p)

p.PutConfiguration(memprovider.Flag{Key: "new-nav", Variants: map[string]any{"off": false}, DefaultVariant: "off"})`),
		doc.P("`PutConfiguration` replaces every flag and emits `EventConfigurationChanged`, listing the changed keys. A flag may also set `Disabled` or a `ContextEvaluator` that picks the variant from the context."),
		doc.P("Optional steps are found by exact method signature: `Initialize(ctx Context, domain string) (Pending, error)` and `Close() (Pending, error)`. A `Close() error` method is not a close step."),
		doc.P("A provider's identity is Go `==`, which decides whether two bindings share one instance. Use pointer types, since two equal struct values count as the same provider."),
	}
}

func fliptBehaviour() []app.UI {
	return []app.UI{
		doc.P("The Flipt provider is the upstream one, wrapped, and its behaviour is documented rather than patched:"),
		doc.List(
			doc.Item("`TargetingKey` is required. An empty one makes every evaluation `GENERAL` with \"entityId cannot be empty\"."),
			doc.Item("Flags refresh only on `SetContext`. There are no events, polling or streaming. `openfeature.SetContext(openfeature.GetContext())` is the refresh."),
			doc.Item("Attributes go through `toString()`: numbers and bools become strings, objects become `[object Object]`."),
			doc.Item("Boolean flags: Flipt's default and a match both report `TARGETING_MATCH`. A disabled flag returns the default with `DEFAULT`."),
			doc.Item("Variant flags: a string is the variant key, a number is `parseFloat(key)`, an object is `JSON.parse(attachment)`. Disabled is `DISABLED`, no match is `DEFAULT`."),
			doc.Item("There is no `Variant` and no flag metadata. A missing flag is `GENERAL`, never `FLAG_NOT_FOUND`."),
			doc.Item("An initialization failure is `PROVIDER_FATAL`, so the status is `FATAL`."),
			doc.Item("The provider has no close step, so the Flipt client is never closed."),
		),
	}
}

func serverConfig() []app.UI {
	return []app.UI{
		doc.P("The browser calls Flipt directly, so the server needs:"),
		doc.List(
			doc.Item("Flipt v1.47 or later, or v2."),
			doc.Item("CORS enabled for your origin, or Flipt served from the same origin."),
			doc.Item("On v1, `x-flipt-environment` and `If-None-Match` added to `cors.allowed_headers`."),
			doc.Item("A client token in `ClientToken` (or a `JWT`, not both) when authentication is on."),
		),
		doc.Plain(`cors:
  enabled: true
  allowed_origins: ["https://app.example.com"]
  allowed_headers:
    - Authorization
    - Content-Type
    - Accept
    - x-flipt-environment
    - If-None-Match`),
	}
}

func parity() []app.UI {
	return []app.UI{
		doc.P("The core is tested one to one against `@openfeature/web-sdk` 1.10.0: the same scenarios run on both and their traces of evaluations, events, hook calls and log lines must match. These are the deliberate differences:"),
		doc.List(
			doc.Item("D1: JS has `removeHandler(fn)` and `AbortSignal`. Go uses the `remove` function `AddHandler` returns, because Go funcs are not comparable."),
			doc.Item("D2: in JS a synchronous throw in `onContextChange` drives the pending counter to -1 and mutes later events. In Go it sets `ERROR`, emits the error event and leaves the counter alone. Upstream does not test it."),
			doc.Item("D3: in JS, after rebinding a provider (domain or default) or `ClearProviders`, the old forwarding keeps feeding domain emitters and API handlers. In Go forwarding is dropped on every rebind, `ClearProviders` and `Reset`."),
			doc.Item("D4: JS freezes the context only when hooks exist, and shallowly. Go always takes a snapshot."),
			doc.Item("D5: JS has one number type. Go has `Number` and `Int`, as the spec's rule on integer evaluation applies to Go."),
			doc.Item("D6: client metadata says `sdk: 'js-web'` and has a `framework`. Go says `SDK: \"webgo\"` and has no framework field."),
			doc.Item("D7: in JS an `initialize` that returns a non-promise leaves `NOT_READY`. In Go `Initialize` returns `(Pending, error)`, and a nil `Pending` means it finished at once, so the status becomes `READY`."),
			doc.Item("D8: JS exports `MultiProvider`, its strategies, `ErrorWithCode` and `AggregateError`. They are not ported in v1."),
		),
	}
}
