package htmx_extensions

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func Overview(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("HTMX Extensions"),
			Text(`
				htmx 4 extensions register when their scripts load. The starter template loads <code>/public/htmgo.js</code>, which bundles htmx and the extensions below.
			`),
			Link("Trigger Children", "/docs/htmx-extensions/trigger-children"),
			Link("Mutation Error", "/docs/htmx-extensions/mutation-error"),
			Link("Alpine Compat", "/docs/htmx-extensions/alpine-compat"),
			Link("Path Deps", "https://github.com/franchb/htmgo/blob/master/framework/assets/js/htmxextensions/pathdeps.ts"),
			Text("Add this script to your page layout if you use a custom layout."),
			ui.GoCodeSnippet(IncludeScript),
			NextStep(
				"mt-4",
				PrevBlock("Pushing Data", DocPath("/pushing-data/sse")),
				NextBlock("Trigger Children", DocPath("/htmx-extensions/trigger-children")),
			),
		),
	)
}

const IncludeScript = `
h.Head(
	h.Script("/public/htmgo.js"),
)
`
