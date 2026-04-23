package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func TodoListExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &TodoListSnippet)
	return Index(ctx)
}
