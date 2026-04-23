package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func ClickToEditExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &ClickToEditSnippet)
	return Index(ctx)
}
