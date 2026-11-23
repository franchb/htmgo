package core_concepts

import "htmgo-site/ui"
import "github.com/franchb/htmgo/framework/v2/h"
import . "htmgo-site/pages/docs"

func TagsAndAttributes(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Tags and Attributes"),
			Text(`
				Build HTML with tag and attribute helpers from <code>github.com/franchb/htmgo/framework/v2/h</code>. Tag helpers return <code>*h.Element</code> and accept attributes and child elements.
			`),
			ui.GoCodeSnippet(TagExample),
			Text("<b>h.Tag</b> and <b>h.Attribute</b> are available to use when you need to render a tag or attribute that htmgo does not provide a method for."),
			ui.GoCodeSnippet(TagExampleUsingTagFunc),
			Text(`
			<b>Attributes</b> are one of the main ways we can add interactivity to the pages with htmx.
			htmgo provides various methods to add attributes to elements, as well as adding attributes based on a condition.
			`),
			ui.GoCodeSnippet(AttributeExample),
			HelpText("The error class is added when errorMessage is nonempty."),
			Text(`This button sends the input named <code>q</code> to <code>GET /search</code> and places the returned fragment in <code>#search-results</code>. Your server supplies the search handler.`),
			ui.GoCodeSnippet(HxAttributeExample),
			NextStep(
				"mt-4",
				PrevBlock("Components", DocPath("/core-concepts/components")),
				NextBlock("Raw HTML", DocPath("/core-concepts/raw-html")),
			),
		),
	)
}

const TagExample = `h.Div(
	h.Class("flex gap-2"),
	h.Button(
		h.Text("Submit"),
	),
)
`

const TagExampleUsingTagFunc = `h.Tag("my-custom-tag",
	h.Class("flex gap-2"),
	h.Button(
		h.Attribute("x-custom-attr", "my-value"),
		h.Text("Submit"),
	),
)
`

const AttributeExample = `h.Div(
	h.Class("flex gap-2"),
	h.Id("my-div"),
	h.ClassIf(errorMessage != "", "p-4 bg-rose-400 text-white rounded"),
)
`

const HxAttributeExample = `h.Div(
	h.Input("text",
		h.Id("search-query"),
		h.Name("q"),
		h.Placeholder("Search"),
	),
	h.Button(
		h.Type("button"),
		h.Get("/search"),
		h.HxInclude("#search-query"),
		h.HxTarget("#search-results"),
		h.Text("Search"),
	),
	h.Div(h.Id("search-results")),
)`
