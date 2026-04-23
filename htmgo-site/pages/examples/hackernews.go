package examples

import "github.com/franchb/htmgo/framework/v2/h"

func HackerNewsExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &HackerNewsSnippet)
	return Index(ctx)
}
