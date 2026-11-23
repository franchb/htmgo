package docs

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/ui"
)

const IntroSnippet = `func IndexPage(ctx *h.RequestContext) *h.Page {
	now := time.Now()
	return h.NewPage(
		h.Div(
			h.Class("flex gap-2"),
			h.TextF("the current time is %s", now.String()),
		),
	)
}
`

func Introduction(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-2"),
			Title("Introduction"),
			Text(`
				htmgo lets you build interactive websites with Go and htmx. Use Go functions to render HTML components, then add htmx attributes to send requests and update parts of the page.
			`),
			Text("This page uses the standard library’s <code>time</code> package and <code>github.com/franchb/htmgo/framework/v2/h</code>."),
			ui.GoCodeSnippet(IntroSnippet),
			h.P(
				Link("The site you are reading now", "https://github.com/franchb/htmgo/tree/master/htmgo-site"),
				h.Text(" was written with htmgo!"),
			),
			NextStep(
				"mt-4",
				h.Div(),
				NextBlock("Getting Started", DocPath("/installation")),
			),
		),
	)
}
