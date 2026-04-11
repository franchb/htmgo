package examples

import "github.com/franchb/htmgo/framework/h"

func HackerNewsExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &HackerNewsSnippet)
	return Index(ctx)
}
