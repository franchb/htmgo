package pages

import (
	"github.com/franchb/htmgo/framework/h"
	"htmgo-site/pages/base"
	"htmgo-site/partials"
)

func CurrentTimePage(ctx *h.RequestContext) *h.Page {
	return base.RootPage(
		ctx,
		h.GetPartial(
			partials.CurrentTimePartial,
			"load, every 1s"),
	)
}
