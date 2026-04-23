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
				The <code>alpine-compat</code> extension preserves Alpine.js
				state across htmx swaps, re-runs Alpine init on swapped-in
				content, and round-trips Alpine component state through
				htmx's history cache. It ships inside
				<code>/public/htmgo.js</code> and auto-gates on
				<code>window.Alpine</code> — if you don't load Alpine, the
				extension no-ops.
			`),

			Text(`<b>Loading Alpine alongside htmgo:</b>`),
			ui.GoCodeSnippet(AlpineLoadingExample),
			HelpText(`
				The 'defer' attribute is important — Alpine auto-initializes
				on DOMContentLoaded. htmgo.js can be loaded before or after
				Alpine; extensions self-register in htmx 4.
			`),

			Text(`<b>Prevent FOUC for x-cloak:</b>`),
			Text(`Add this CSS rule so <code>x-cloak</code>-marked elements stay hidden until Alpine initializes them:`),
			ui.GoCodeSnippet(CloakCSSExample),

			Text(`<b>Using the ax package:</b>`),
			Text(`
				htmgo's <code>framework/ax</code> package mirrors the
				<code>hx</code> shape with Alpine directive helpers. Common
				helpers: <code>ax.Data</code>, <code>ax.Show</code>,
				<code>ax.OnClick</code>, <code>ax.Model</code>,
				<code>ax.BindClass</code>, <code>ax.OnKeydownEscape</code>,
				<code>ax.Cloak</code>. See GoDoc for the full list.
			`),
			ui.GoCodeSnippet(PopoverExample),

			Text(`<b>Alpine + htmx swap interaction:</b>`),
			Text(`
				When htmx morphs content into an Alpine component's
				descendant, the extension's
				<code>htmx_before_morph_node</code> hook copies
				<code>_x_dataStack</code> to the new nodes and
				<code>Alpine.cloneNode</code> preserves bindings. State
				survives the swap. When htmx morphs the Alpine root itself,
				the extension still carries state through because it runs
				before htmx's morph op.
			`),

			Text(`<b>Gotchas:</b>`),
			HelpText(`
				• Alpine v3 only — v2 is out of support upstream.
				• Tested against Alpine v3.15.11 (released 2025-04-02). Patch
				  updates within v3.15 should be safe; 3.16 is not on the
				  public roadmap.
				• If Alpine loads after a swap, pre-Alpine widgets won't init
				  until Alpine boots. Load Alpine in the page <head> with
				  'defer'.
				• Alpine plugins (@alpinejs/persist, @alpinejs/intersect,
				  etc.) must be loaded before Alpine itself, per the plugin
				  docs.
			`),

			NextStep(
				"mt-4",
				PrevBlock("Mutation Error", DocPath("/htmx-extensions/mutation-error")),
				NextBlock("Tailwind Intellisense", DocPath("/misc/tailwind-intellisense")),
			),
		),
	)
}

const AlpineLoadingExample = `
h.Html(
    h.Head(
        // Alpine v3.15.11 — tested pin. Use alpinejs@3.15 for patch updates.
        h.Tag("script",
            h.Attribute("src", "https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js"),
            h.Attribute("defer", ""),
        ),
        h.Script("/public/htmgo.js"),
    ),
)
`

const CloakCSSExample = `
[x-cloak] { display: none !important; }
`

const PopoverExample = `
h.Div(
    h.Class("relative"),
    ax.Data("{ open: false }"),

    h.Button(
        h.Class("btn"),
        ax.OnClick("open = !open"),
        h.Text("Toggle"),
    ),

    h.Div(
        h.Class("popover"),
        ax.Show("open"),
        ax.Cloak(),
        ax.OnClickOutside("open = false"),
        ax.OnKeydownEscape("open = false"),
        h.HxGet("/popover/content"),
        h.HxTrigger("intersect once"),
        h.HxSwap(hx.SwapTypeInnerHtml),
        h.Text("Loading…"),
    ),
)
`
