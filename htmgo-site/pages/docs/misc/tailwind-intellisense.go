package misc

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/assets"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func TailwindIntellisense(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Tailwind IntelliSense"),
			Text("Configure the Tailwind CSS language server to recognize classes passed to htmgo's Go helpers."),
			Image("/public/tailwind-intellisense.png"),
			SubTitle("JetBrains IDEs"),
			Text("In GoLand, open Settings → Languages & Frameworks → Style Sheets → Tailwind CSS. Merge these options into the existing Tailwind language-server configuration."),
			ui.CodeSnippet(ui.CodeSnippetProps{
				Code:            assets.JetBrainsTailwindConfig,
				Lang:            "json",
				HideLineNumbers: true,
			}),
			SubTitle("Visual Studio Code"),
			Text("Merge these options into settings.json."),
			ui.CodeSnippet(ui.CodeSnippetProps{
				Code:            assets.VSCodeTailwindConfig,
				Lang:            "json",
				HideLineNumbers: true,
			}),
			NextStep(
				"mt-4",
				PrevBlock("Alpine Compat", DocPath("/htmx-extensions/alpine-compat")),
				NextBlock("Formatting blocks", DocPath("/misc/formatter")),
			),
		),
	)
}
