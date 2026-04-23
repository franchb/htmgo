package partials

import "github.com/franchb/htmgo/framework/v2/h"

func CountersPartial(ctx *h.RequestContext) *h.Partial {
	return h.NewPartial(
		h.Div(
			h.Text("my counter"),
		),
	)
}

func SwapFormError(ctx *h.RequestContext, error string) *h.Partial {
	return h.SwapPartial(
		ctx,
		h.Div(),
	)
}
