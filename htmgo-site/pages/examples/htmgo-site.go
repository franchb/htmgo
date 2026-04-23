package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func HtmgoSiteExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &HtmgoSiteSnippet)
	return Index(ctx)
}
