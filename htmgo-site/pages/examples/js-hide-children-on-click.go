package examples

import (
	"github.com/franchb/htmgo/framework/h"
)

func JsHideChildrenOnClickPage(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &JsHideChildrenOnClick)
	return Index(ctx)
}
