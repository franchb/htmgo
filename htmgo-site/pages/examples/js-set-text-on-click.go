package examples

import (
	"github.com/franchb/htmgo/framework/h"
)

func JsSetTextOnClickPage(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &JsSetTextOnClick)
	return Index(ctx)
}
