package examples

import (
	"github.com/franchb/htmgo/framework/h"
)

func ChatExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &ChatSnippet)
	return Index(ctx)
}
