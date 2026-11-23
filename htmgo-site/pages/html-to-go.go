package pages

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/internal/sitebuild"
	"htmgo-site/pages/base"
	"htmgo-site/partials"
	"htmgo-site/ui"
)

func HtmlToGoPage(ctx *h.RequestContext) *h.Page {
	if sitebuild.Static() {
		return base.PageWithNav(ctx,
			h.Main(
				h.Class("mx-auto max-w-3xl p-8"),
				h.H1(h.Class("text-2xl font-bold"), h.Text("Convert HTML to htmgo code")),
				h.P(
					h.Class("my-4"),
					h.Text("Run this repository’s Go site locally to generate code for the franchb/htmgo v2 framework. Follow the "),
					h.A(h.Class("text-blue-700 underline"), h.Href("/docs/installation"), h.Text("installation guide")),
					h.Text(" to set up Go, Node.js, and the htmgo CLI."),
				),
				ui.BashCodeSnippet(`git clone https://github.com/franchb/htmgo.git
cd htmgo/htmgo-site
npm ci
GOWORK=off htmgo run`),
				h.P(
					h.Class("my-4"),
					h.Text("Open "),
					h.Tag("code", h.Text("http://localhost:3000/html-to-go")),
					h.Text(" in your browser."),
				),
				h.A(
					h.Class("text-blue-700 underline"),
					h.Href(sitebuild.UpstreamOrigin+"/html-to-go"),
					h.Text("Try the upstream HTML converter"),
				),
				h.P(h.Class("mt-2 text-sm text-slate-700"), h.Text("The upstream converter may generate imports for a different htmgo version.")),
			),
		)
	}

	return base.PageWithNav(
		ctx,
		h.Div(
			h.Class("flex flex-col h-screen items-center justify-center w-full pt-6"),
			h.H3(
				h.Text("Convert raw html to htmgo code"),
				h.Class("text-2xl font-bold"),
			),
			h.Div(
				h.Class("h-full w-full flex gap-4 p-8"),
				partials.HtmlInput(),
				partials.HiddenCopyOutput(""),
				partials.GoOutput(""),
			),
		),
	)
}
