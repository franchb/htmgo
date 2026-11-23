package interactivity

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/ui"
)

import . "htmgo-site/pages/docs"

func EvalCommands(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Eval Commands"),
			Text(`Pass an element and commands to <code>js.EvalCommands</code> to run the commands against it when an event fires.`),
			SubTitle("Reference an element directly"),
			ui.GoCodeSnippet(EvalCommandsSnippet),
			Text(`The button toggles the paragraph text. <code>js.EvalCommands</code> assigns a selector to the referenced paragraph as it builds the elements.`),
			Link("View the related demo", "/examples/js-hide-children-on-click"),
			SubTitle("Use a selector"),
			Text(`Use <code>js.EvalCommandsOnSelector</code> to target an element using its CSS selector.`),
			ui.GoCodeSnippet(EvalCommandsSnippetWithSelector),
			NextStep(
				"mt-4",
				PrevBlock("Events / Commands", DocPath("/interactivity/events")),
				NextBlock("Caching Components", DocPath("/performance/caching-globally")),
			),
		),
	)
}

const EvalCommandsSnippetWithSelector = `
func ToggleTextButton() *h.Element {
	text := h.P(h.Id("my-element"), h.Text("Text Before"))
	return h.Div(
		h.Button(
			h.Type("button"),
			h.Text("Toggle Text"),
			h.OnClick(
				js.EvalCommandsOnSelector(
					"#my-element",
					js.ToggleText("Text Before", "Text After"),
				),
			),
		),
		text,
	)
}
`

const EvalCommandsSnippet = `func ToggleTextButton() *h.Element {
	text := h.P(h.Text("Text Before"))
	return h.Div(
		h.Button(
			h.Type("button"),
			h.Text("Toggle Text"),
			h.OnClick(
				js.EvalCommands(
					text,
					js.ToggleText("Text Before", "Text After"),
				),
			),
		),
		text,
	)
}`
