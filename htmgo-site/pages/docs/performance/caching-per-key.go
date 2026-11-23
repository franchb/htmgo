package performance

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func CachingPerKey(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Caching Components Per Key"),
			Text(`
				Use <code>h.CachedPerKey</code> or one of its typed variants when the component's output varies by key. The key can be any comparable Go value, such as a product ID or a struct containing a product ID and locale.
				Declare the wrapper at package scope. Its callback returns the key and a render function. The callback runs on each call, while the render function runs when the key is missing or expired.
			`),
			ui.GoCodeSnippet(CachedPerKeyExample),
			Text(`In <code>pages/products.$id.go</code>, the page passes a product ID as its one argument to <code>h.CachedPerKeyT</code>. Requests for the same ID reuse that ID's cached HTML.`),
			Text(`Use <code>h.CachedPerKey</code> for zero arguments, <code>h.CachedPerKeyT</code> for one, and <code>h.CachedPerKeyT2</code>, <code>h.CachedPerKeyT3</code>, or <code>h.CachedPerKeyT4</code> for two to four. The key returned by the callback determines which HTML to reuse. For private data, authorize each request before reading from the cache, and include every value that affects who can see the rendered content in the key.`),
			h.P(
				h.Text("The default in-memory store coordinates misses for each key through GetOrCompute. Other backends implement that contract through "),
				Link("custom cache stores", DocPath("/performance/pluggable-caches")),
				h.Text("."),
			),
			NextStep(
				"mt-4",
				PrevBlock("Caching Globally", DocPath("/performance/caching-globally")),
				NextBlock("Custom Cache Stores", DocPath("/performance/pluggable-caches")),
			),
		),
	)
}

const CachedPerKeyExample = `
var CachedProductCard = h.CachedPerKeyT(15*time.Minute, func(productID string) (string, h.GetElementFunc) {
	return productID, func() *h.Element {
		return h.Div(
			h.H3(h.TextF("Product %s", productID)),
			h.P(h.TextF("Generated at %s", time.Now().Format(time.RFC3339))),
		)
	}
})

func ProductPage(ctx *h.RequestContext) *h.Page {
	return h.NewPage(
		CachedProductCard(ctx.UrlParam("id")),
	)
}
`
