package docs

import "github.com/franchb/htmgo/framework/h"

func Index(ctx *h.RequestContext) *h.Page {
	ctx.Redirect("/docs/introduction", 302)
	return h.EmptyPage()
}
