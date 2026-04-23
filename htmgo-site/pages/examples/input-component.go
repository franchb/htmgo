package examples

import "github.com/franchb/htmgo/framework/v2/h"

func InputComponentExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &InputComponentSnippet)
	return Index(ctx)
}
