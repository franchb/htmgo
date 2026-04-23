package components

import "github.com/franchb/htmgo/framework/v2/h"

func FormError(error string) *h.Element {
	return h.Div(
		h.Id("form-error"),
		h.Text(error),
		h.If(
			error != "",
			h.Class("p-4 bg-rose-400 text-white rounded"),
		),
	)
}
