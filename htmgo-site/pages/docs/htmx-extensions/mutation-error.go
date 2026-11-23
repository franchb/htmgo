package htmx_extensions

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func MutationError(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Mutation Error"),
			Text(`
				The <code>mutation-error</code> extension emits <code>htmx:onMutationError</code> for POST, PUT, PATCH, and DELETE requests with an HTTP status of 400 or higher.
				Attach a handler with <code>h.HxOnMutationError</code> to respond to the event.
			`),
			Text(`<b>Example:</b>`),
			ui.GoCodeSnippet(MutationErrorExample),
			Text(`The extension also notifies elements with a mutation-error listener, such as a submit button inside the requesting form.`),
			NextStep(
				"mt-4",
				PrevBlock("Trigger Children", DocPath("/htmx-extensions/trigger-children")),
				NextBlock("Alpine Compat", DocPath("/htmx-extensions/alpine-compat")),
			),
		),
	)
}

const MutationErrorExample = `
func SaveForm() *h.Element {
	return h.Form(
		h.PostPartial(Save),
		h.NoSwap(),
		h.Button(
			h.Type("submit"),
			h.HxOnMutationError(js.Alert("Save failed.")),
			h.Text("Save"),
		),
	)
}

func Save(ctx *h.RequestContext) *h.Partial {
	ctx.Fiber.Status(400)
	return h.EmptyPartial()
}
`
