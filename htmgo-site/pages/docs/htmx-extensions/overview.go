package htmx_extensions

import (
	"github.com/franchb/htmgo/framework/h"
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
				In htmx 4, extensions self-register on script import — the <code>hx-ext</code>
				attribute has been removed. htmgo bundles the built-in extensions automatically
				via the <code>/public/htmgo.js</code> script tag.
			`),
			Text(`
				The following extensions are included in htmgo's bundled script:
			`),
			Link("Trigger Children", "/docs/htmx-extensions/trigger-children"),
			Link("Mutation Error", "/docs/htmx-extensions/mutation-error"),
			Link("Alpine Compat", "/docs/htmx-extensions/alpine-compat"),
			Link("Path Deps", "https://github.com/bigskysoftware/htmx-extensions/blob/main/src/path-deps/README.md"),
			h.P(
				h.Class("mt-3"),
				h.Text("To include the htmgo script in your project:"),
				ui.GoCodeSnippet(IncludeScript),
			),
			NextStep(
				"mt-4",
				PrevBlock("Pushing Data", DocPath("/pushing-data/sse")),
				NextBlock("Trigger Children", DocPath("/htmx-extensions/trigger-children")),
			),
		),
	)
}

const IncludeScript = `
h.Html(
    h.Head(
        h.Script("/public/htmgo.js"),
    ),
)
`
