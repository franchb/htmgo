package htmx_extensions

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
)

func TriggerChildren(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Trigger Children"),
			Text(`
				The <code>trigger-children</code> extension forwards supported htmx events to descendants that declare matching listeners.
				For example, a button inside a form can react when the form sends a request.
			`),
			Link("View Example", "/examples/form"),
			HelpText(`The loading indicator listens for <code>htmx:before:request</code> and <code>htmx:after:request</code> to update the button during submission.`),
			NextStep(
				"mt-4",
				PrevBlock("HTMX Extensions", DocPath("/htmx-extensions/overview")),
				NextBlock("Mutation Error", DocPath("/htmx-extensions/mutation-error")),
			),
		),
	)
}
