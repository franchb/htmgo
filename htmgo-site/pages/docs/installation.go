package docs

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/ui"
)

func Installation(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Quick Start"),
			SubTitle("Prerequisites"),
			h.Ul(
				h.Class("list-disc list-inside"),
				h.Li(Link("Go 1.27 or later", "https://go.dev/doc/install")),
				h.Li(Link("Git", "https://git-scm.com/install")),
			),
			h.P(
				h.Text("The "),
				Link("htmx 4 documentation", "https://four.htmx.org/docs/"),
				h.Text(" covers the request and swap concepts used in this guide."),
			),
			StepTitle("1. Install htmgo"),
			ui.SingleLineBashCodeSnippet(`go install github.com/franchb/htmgo/cli/htmgo/v2@latest`),
			Text("Add the directory where Go installs binaries to your PATH. This is GOBIN if set, or the bin directory under GOPATH."),
			StepTitle("2. Create a project"),
			ui.BashCodeSnippet(`htmgo template myapp
cd myapp`),
			Text("The command downloads the starter template, initializes a Git repository, and sets up the app in myapp."),
			StepTitle("3. Start the development server"),
			ui.SingleLineBashCodeSnippet(`htmgo watch`),
			h.P(
				h.Text("Open "),
				Link("http://localhost:3000", "http://localhost:3000"),
				h.Text(" to view the starter app. Source changes trigger a rebuild and browser reload. Use htmgo run to start the server without watching files."),
			),
			StepTitle("4. Build for production"),
			ui.SingleLineBashCodeSnippet(`htmgo build`),
			Text("Run this from your project directory. It builds the executable into ./dist with the production assets embedded."),
			NextStep(
				"mt-4",
				PrevBlock("Introduction", DocPath("/introduction")),
				NextBlock("Core Concepts", DocPath("/core-concepts/pages")),
			),
		),
	)
}
