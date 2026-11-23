package core_concepts

import "htmgo-site/ui"
import "github.com/franchb/htmgo/framework/v2/h"
import . "htmgo-site/pages/docs"

func RawHtml(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Raw HTML"),
			Text(`
				In some cases, you may want to render raw html instead of using htmgo's functions.
		    This can be done by using the following methods:
			`),
			Text("Rendering raw html:"),
			ui.GoCodeSnippetSingleLine(RawHtmlExample),
			Text("Rendering with formatting:"),
			ui.GoCodeSnippetSingleLine(RawHtmlExample2),
			Text("Rendering a script:"),
			ui.GoCodeSnippetSingleLine(RawHtmlExample3),
			Text(`
				Use <code>h.Text</code> to render untrusted values as escaped text. When you intend to render user-supplied HTML markup, sanitize it before passing it to the raw HTML helpers.
			`),
			h.P(
				h.Text("For example, "),
				Link("bluemonday", "https://github.com/microcosm-cc/bluemonday"),
				h.Text(" lets you define which HTML elements and attributes to allow."),
			),
			NextStep(
				"mt-4",
				PrevBlock("Tags and Attributes", DocPath("/core-concepts/tags-and-attributes")),
				NextBlock("Conditionals", DocPath("/control/if-else")),
			),
		),
	)
}

const RawHtmlExample = `h.UnsafeRaw("<div>Raw HTML</div>")`
const RawHtmlExample2 = `h.UnsafeRawF("<div>%s</div>", "Raw HTML")`
const RawHtmlExample3 = `h.UnsafeRawScript("alert('Hello World')")`
