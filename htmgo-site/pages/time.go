package pages

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/pages/base"
	"htmgo-site/partials"
)

func CurrentTimePage(ctx *h.RequestContext) *h.Page {
	return base.RootPage(
		ctx,
		partials.CurrentTimeDemo(),
	)
}
