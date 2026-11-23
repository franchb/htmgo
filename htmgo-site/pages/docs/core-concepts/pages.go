package core_concepts

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

const ExcludeRootSnippet = `automatic_page_routing_ignore: ["root.go"]`

const AbstractedRootPageUsageSnippet = `func UserPage(ctx *h.RequestContext) *h.Page {
	return RootPage(
		h.Pf("User ID: %s", ctx.UrlParam("id")),
	)
}
`

const RootPageSnippet = `func RootPage(children ...h.Ren) *h.Page {
	return h.NewPage(
		h.Html(
			h.Head(
				h.Meta("viewport", "width=device-width, initial-scale=1"),
				h.Link("/public/main.css", "stylesheet"),
				h.Script("/public/htmgo.js"),
			),
			h.Body(children...),
		),
	)
}
`

const PagesSnippet = `package pages

import "github.com/franchb/htmgo/framework/v2/h"

func HelloHtmgoPage(ctx *h.RequestContext) *h.Page {
	return h.NewPage(
		h.Html(
			h.Head(
				h.Link("/public/main.css", "stylesheet"),
				h.Script("/public/htmgo.js"),
			),
			h.Body(
				h.P(h.Text("Hello, htmgo!")),
			),
		),
	)
}
`

func Pages(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Pages"),
			Text(`
				Pages are the entry point of an htmgo application.
				A simple page may look like:
			`),
			ui.GoCodeSnippet(PagesSnippet),
			Text(`A page handler accepts <code>*h.RequestContext</code>, which wraps Fiber v3's <code>fiber.Ctx</code>, and returns <code>*h.Page</code>. Put this one in <code>pages/hello.go</code> to serve it at <code>/hello</code>.`),
			autoRegistration(),
			tips(),
			NextStep(
				"mt-4",
				PrevBlock("Getting Started", DocPath("/installation")),
				NextBlock("Partials", DocPath("/core-concepts/partials")),
			),
		),
	)
}

func autoRegistration() *h.Element {
	return h.Div(
		h.Class("flex flex-col gap-2"),
		SubTitle("Automatic Routing"),
		Text(`htmgo registers GET routes with Fiber for exported functions under <code>pages/</code> with a single named <code>*h.RequestContext</code> parameter and a <code>*h.Page</code> return value. The file path determines the URL.`),
		h.Ul(
			h.Class("list-disc list-inside"),
			h.Li(h.Text("pages/index.go → /")),
			h.Li(h.Text("pages/users.go → /users")),
			h.Li(h.Text("pages/users.$id.go → /users/:id")),
		),
		HelpText(`Read the <code>id</code> route parameter with <code>ctx.UrlParam("id")</code>.`),
	)
}

func tips() *h.Element {
	return h.Div(
		h.Class("flex flex-col gap-2"),
		SubTitle("Shared page layout"),
		Text(`Use a shared function to reuse your page layout, stylesheets, and scripts. This helper belongs in <code>pages/root.go</code> in the <code>pages</code> package.`),
		ui.GoCodeSnippet(RootPageSnippet),
		Text(`Call it from <code>pages/users.$id.go</code> in the same package.`),
		ui.GoCodeSnippet(AbstractedRootPageUsageSnippet),
		Text(`The route generator skips this <code>RootPage</code> because its signature accepts children. To exclude a file explicitly, add it to <code>htmgo.yml</code>.`),
		ui.CodeSnippet(ui.CodeSnippetProps{Code: ExcludeRootSnippet, Lang: "yaml", SingleLine: true}),
	)
}
