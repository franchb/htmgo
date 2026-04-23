package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func ChatExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &ChatSnippet)
	return Index(ctx)
}
