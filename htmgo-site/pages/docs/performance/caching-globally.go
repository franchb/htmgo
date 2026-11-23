package performance

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func CachingGlobally(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Caching Components Globally"),
			Text(`
				Use <code>h.Cached</code> to reuse a component's rendered HTML for a fixed duration. The first render computes and stores the HTML. Subsequent renders reuse that value until it expires.
				Declare the wrapper at package scope so it can reuse the cache across requests. Cache only content that every caller can see.
			`),
			ui.GoCodeSnippet(CachedGloballyExample),
			Text(`The timestamp remains the same until the cached value expires. Replace the callback with your own component to cache a shared render.`),
			Text(`Use <code>h.Cached</code> for zero arguments, <code>h.CachedT</code> for one, and <code>h.CachedT2</code>, <code>h.CachedT3</code>, or <code>h.CachedT4</code> for two to four. Each wrapper uses one cache entry for all calls, regardless of its arguments.`),
			h.P(
				h.Text("When the rendered output depends on an identifier, use "),
				Link("per-key caching", DocPath("/performance/caching-per-key")),
				h.Text("."),
			),
			NextStep(
				"mt-4",
				PrevBlock("Eval Commands", DocPath("/interactivity/eval-commands")),
				NextBlock("Caching Per Key", DocPath("/performance/caching-per-key")),
			),
		),
	)
}

const CachedGloballyExample = `
var CachedStatus = h.Cached(15*time.Minute, func() *h.Element {
	return h.P(h.TextF("Generated at %s", time.Now().Format(time.RFC3339)))
})

func StatusPage(ctx *h.RequestContext) *h.Page {
	return h.NewPage(
		CachedStatus(),
	)
}
`
