package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func JsHideChildrenOnClickPage(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &JsHideChildrenOnClick)
	return Index(ctx)
}
