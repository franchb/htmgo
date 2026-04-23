package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func JsSetTextOnClickPage(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &JsSetTextOnClick)
	return Index(ctx)
}
