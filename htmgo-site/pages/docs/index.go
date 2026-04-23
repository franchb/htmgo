package docs

import "github.com/franchb/htmgo/framework/v2/h"

func Index(ctx *h.RequestContext) *h.Page {
	_ = ctx.Redirect("/docs/introduction", 302)
	return h.EmptyPage()
}
