package core_concepts

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

const ComponentExampleSnippet = `func TaskCard(title, description string) *h.Element {
	return h.Div(
		h.Class("bg-white w-full rounded shadow-md p-4"),
		h.H3(h.Text(title)),
		h.P(h.Text(description)),
	)
}

func TaskPage(ctx *h.RequestContext) *h.Page {
	return h.NewPage(
		TaskCard("Review the examples", "Run each snippet before publishing."),
	)
}
`

func Components(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Components"),
			Text(`
				Components are Go functions that build reusable pieces of HTML. A component can accept data as arguments or load data when your code calls it.
				A page or partial can call the same component to render it in different parts of your application.
			`),
			ui.GoCodeSnippet(ComponentExampleSnippet),
			Text("Call the component directly in a page or in a partial. When htmx requests an update, the partial can render the component again with fresh data."),
			NextStep(
				"mt-4",
				PrevBlock("Partials", DocPath("/core-concepts/partials")),
				NextBlock("Tags and Attributes", DocPath("/core-concepts/tags-and-attributes")),
			),
		),
	)
}
