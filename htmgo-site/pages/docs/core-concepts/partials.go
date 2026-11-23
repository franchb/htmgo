package core_concepts

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/internal/sitebuild"
	. "htmgo-site/pages/docs"
	"htmgo-site/partials"
	"htmgo-site/ui"
)

const PartialsSnippet = `package partials

import (
	"time"

	"github.com/franchb/htmgo/framework/v2/h"
)

func CurrentTimePartial(ctx *h.RequestContext) *h.Partial {
	return h.NewPartial(
		h.Pf("The current time is %s", time.Now().Format(time.RFC3339)),
	)
}
`

const examplePageSnippet = `package pages

import (
	"github.com/franchb/htmgo/framework/v2/h"

	"myapp/partials"
)

func CurrentTimePage(ctx *h.RequestContext) *h.Page {
	return RootPage(
		h.Div(
			h.GetPartial(partials.CurrentTimePartial, "load, every 1s"),
		),
	)
}
`

func Partials(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Partials"),
			Text(`
				A partial handles an HTTP request and returns an HTML fragment. It can also set response headers, for example to redirect an htmx request.
				Place this handler in <code>partials/time.go</code>.
			`),
			ui.GoCodeSnippet(PartialsSnippet),
			Text(`
				htmgo automatically registers exported handlers that accept a single named <code>*h.RequestContext</code> and return <code>*h.Partial</code>.
				The generated route uses the module, package import path, and function name. Pass the function to <code>h.GetPartial</code> or <code>h.PostPartial</code> to generate the matching URL.
			`),
			example(),
			NextStep(
				"mt-4",
				PrevBlock("Pages", DocPath("/core-concepts/pages")),
				NextBlock("Components", DocPath("/core-concepts/components")),
			),
		),
	)
}

func example() *h.Element {
	return h.Div(
		h.Class("flex flex-col gap-2"),
		SubTitle("Polling a partial"),
		Text(`Put this page in <code>pages/time.go</code> in the starter project. It uses the starter's <code>RootPage</code> helper, which loads <code>htmgo.js</code>. If you chose a different module name, update the <code>myapp/partials</code> import to match your <code>go.mod</code>.`),
		ui.GoCodeSnippet(examplePageSnippet),
		Text("The page requests the partial on load and again every second. htmx swaps the returned HTML into the div."),
		h.If(sitebuild.Static(), HelpText("This hosted preview uses your browser clock. The code above polls a server.")),
		partials.CurrentTimeDemo(),
	)
}
