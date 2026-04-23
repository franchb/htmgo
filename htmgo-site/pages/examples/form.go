package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func FormWithLoadingState(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &FormWithLoadingStateSnippet)
	return Index(ctx)
}
