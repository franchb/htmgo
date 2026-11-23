package interactivity

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/ui"
)

import . "htmgo-site/pages/docs"

func EventsAndCommands(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Events and Commands"),
			Text(`Attach an event handler from <code>h</code> to an element. Pass it commands from <code>github.com/franchb/htmgo/framework/v2/js</code> to update the page in the browser.`),
			Text("This form disables its submit button while the request is in flight and restores it when the request finishes."),
			ui.GoCodeSnippet(EventsExample1),
			h.P(
				h.Text("The bundled "),
				Link("Trigger Children extension", DocPath("/htmx-extensions/trigger-children")),
				h.Text(" forwards the form’s request events to the button."),
			),
			SubTitle("Choose an event"),
			Text(`Common helpers include <code>h.OnClick</code>, <code>h.OnSubmit</code>, <code>h.HxBeforeRequest</code>, <code>h.HxAfterRequest</code>, and <code>h.HxOnAfterSwap</code>.`),
			h.P(
				h.Text("Use h.OnEvent with an "),
				Link("event constant from hx", "https://pkg.go.dev/github.com/franchb/htmgo/framework/v2/hx"),
				h.Text(" or a DOM handler name such as onblur. htmx event names use the htmx: prefix and colon separators."),
			),
			ui.GoCodeSnippet(OnEventBlurSnippet),
			SubTitle("Run JavaScript"),
			Text(`Use <code>js.EvalJs</code> to run code inside a handler. The function receives the current element as <code>self</code> and the event as <code>event</code> (also available as <code>e</code>).`),
			ui.GoCodeSnippet(EvalArbitraryJavascriptSnippet),
			Link("Browse the js command reference", "https://pkg.go.dev/github.com/franchb/htmgo/framework/v2/js"),
			Link("View the button demo", "/examples/js-set-text-on-click"),
			NextStep(
				"mt-4",
				PrevBlock("Swapping", DocPath("/interactivity/swapping")),
				NextBlock("Eval Commands", DocPath("/interactivity/eval-commands")),
			),
		),
	)
}

const EventsExample1 = `
func MyForm() *h.Element {
	return h.Form(
		h.PostPartial(Submit),
		h.NoSwap(),
		h.Button(
			h.Type("submit"),
			h.Text("Submit"),
			h.HxBeforeRequest(
				js.SetDisabled(true),
				js.SetText("Submitting…"),
			),
			h.OnEvent("htmx:finally:request",
				js.SetDisabled(false),
				js.SetText("Submit"),
			),
		),
	)
}

func Submit(ctx *h.RequestContext) *h.Partial {
	return h.EmptyPartial()
}
`

const OnEventBlurSnippet = `
h.Input("text",
	h.OnEvent(hx.BlurEvent, js.SetValue("Input was blurred")),
)
`

const EvalArbitraryJavascriptSnippet = `
func MyButton() *h.Element {
	return h.Button(
		h.Type("button"),
		h.Text("Click me"),
		h.OnClick(
			js.EvalJs("self.textContent = 'Clicked';"),
		),
	)
}
`
