package htmx_extensions

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func AlpineCompat(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Alpine Compat"),
			Text(`
				The <code>alpine-compat</code> extension preserves Alpine.js state and bindings on matching nodes when htmx morphs content.
				htmgo includes the extension in <code>/public/htmgo.js</code>. Its Alpine hooks run when <code>window.Alpine</code> is available.
			`),
			SubTitle("Load Alpine and htmgo"),
			Text("Include both scripts in the page head. Use defer so Alpine initializes after the document is parsed."),
			ui.GoCodeSnippet(AlpineLoadingExample),
			Text(`Add this CSS to keep <code>x-cloak</code> elements hidden until Alpine initializes them.`),
			ui.CodeSnippet(ui.CodeSnippetProps{Code: CloakCSSExample, Lang: "css"}),
			SubTitle("Load a popover on demand"),
			h.P(
				h.Text("Import Alpine directive helpers from "),
				Link("framework/v2/ax", "https://pkg.go.dev/github.com/franchb/htmgo/framework/v2/ax"),
				h.Text(". The popover loads its content when it first becomes visible."),
			),
			ui.GoCodeSnippet(PopoverExample),
			Text(`Use <code>innerMorph</code> or <code>outerMorph</code> when htmx needs to update Alpine-managed content in place. The extension copies the existing Alpine data stack when htmx matches old and new nodes.`),
			NextStep(
				"mt-4",
				PrevBlock("Mutation Error", DocPath("/htmx-extensions/mutation-error")),
				NextBlock("Tailwind Intellisense", DocPath("/misc/tailwind-intellisense")),
			),
		),
	)
}

const AlpineLoadingExample = `
h.Head(
	h.Script("/public/htmgo.js"),
	h.Tag("script",
		h.Src("https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js"),
		h.Attribute("defer", ""),
	),
)
`

const CloakCSSExample = `
[x-cloak] { display: none !important; }
`

const PopoverExample = `
func Popover() *h.Element {
	return h.Div(
		ax.Data("{ open: false }"),
		ax.OnKeydown("open = false", "escape", "window"),
		h.Button(
			h.Type("button"),
			ax.OnClick("open = !open"),
			ax.Bind("aria-expanded", "open"),
			h.Text("Toggle"),
		),
		h.Div(
			ax.Show("open"),
			ax.Cloak(),
			ax.OnClickOutside("open = false"),
			h.GetPartial(PopoverContent, "intersect once"),
			h.Attribute(hx.SwapAttr, "innerMorph"),
			h.Text("Loading…"),
		),
	)
}

func PopoverContent(ctx *h.RequestContext) *h.Partial {
	return h.NewPartial(h.P(h.Text("Popover content")))
}
`
