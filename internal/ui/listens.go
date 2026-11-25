package ui

import (
	"net/url"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

// Listens displays the API response, loading state, and retryable errors.
type Listens struct {
	app.Compo

	// mountCtx is bound to this component; event contexts are bound to their
	// source element and drop dispatches once that element unmounts.
	mountCtx app.Context

	Failure error
	Result  ListenResult
	Loading bool
}

// RegisterRoutes registers the root page for WASM and static export.
func RegisterRoutes() {
	app.Route("/", func() app.Composer { return &Listens{} })
}

// OnMount loads listens in the browser, never during static export.
//
//nolint:gocritic // go-app requires app.Context by value in lifecycle methods.
func (page *Listens) OnMount(ctx app.Context) {
	if app.IsClient {
		page.mountCtx = ctx
		page.load()
	}
}

// Render describes the page without accessing browser globals.
func (page *Listens) Render() app.UI {
	return app.Main().Body(
		app.H1().Text("Listens"),
		app.P().Text("Response from GET /listens"),
		app.If(page.Loading, func() app.UI {
			return app.P().Role("status").Text("Loading listens…")
		}),
		app.If(page.Failure != nil, func() app.UI {
			return app.Div().Role("alert").Body(
				app.P().Text("Could not load listens. "+page.Failure.Error()),
				app.Button().Type("button").Disabled(page.Loading).Text("Try again").OnClick(page.retry),
			)
		}),
		app.If(page.Result.Empty, func() app.UI {
			return app.P().Text("No listens yet.")
		}),
		app.If(page.Result.JSON != "", func() app.UI {
			return app.Pre().Aria("label", "Listens response").Text(page.Result.JSON)
		}),
	)
}

//nolint:gocritic // go-app requires app.Context by value in event handlers.
func (page *Listens) retry(_ app.Context, _ app.Event) {
	page.load()
}

func (page *Listens) load() {
	ctx := page.mountCtx
	page.Loading = true
	page.Failure = nil
	page.Result = ListenResult{}
	location := ctx.Page().URL()
	origin := (&url.URL{Scheme: location.Scheme, Host: location.Host}).String()
	ctx.Async(func() {
		result, err := LoadListens(ctx, origin)
		ctx.Dispatch(func(app.Context) {
			page.Result = result
			page.Failure = err
			page.Loading = false
		})
	})
}
