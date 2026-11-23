package control

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func IfElse(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("If / Else Statements"),
			Text(`
				If / else statements are useful when you want to conditionally render attributes or elements / components.
				htmgo provides a couple of utilities to do so:
			`),
			Text("Render an icon if the task is complete."),
			ui.GoCodeSnippet(IfElseExample),
			Text("Choose which partial handles a request with <code>h.Ternary</code>."),
			ui.GoCodeSnippet(TernaryExample),
			Text(`Apply multiple classes based on a condition.`),
			ui.GoCodeSnippet(ConditionalClassExample),
			Text("Apply a single class based on a condition."),
			ui.GoCodeSnippet(ClassIfElseExample),
			Text("Render different elements based on a condition."),
			ui.GoCodeSnippetSingleLine(IfElseExample2),
			Text(`
				Go evaluates both <code>EditTaskForm()</code> and <code>ViewTask()</code> before calling <code>h.IfElse</code>.
				Pass the functions to <code>h.IfElseLazy</code> to call only the selected one.
			`),
			ui.GoCodeSnippetSingleLine(IfElseExample3),
			NextStep(
				"mt-4",
				PrevBlock("Raw HTML", DocPath("/core-concepts/raw-html")),
				NextBlock("Rendering Lists", DocPath("/control/loops")),
			),
		),
	)
}

const IfElseExample = `
h.Div(
	h.If(task.CompletedAt != nil, h.Span(h.Text("✓"))),
)
`

const TernaryExample = `h.Button(
	h.Type("button"),
	h.PostPartialWithQs(
		h.Ternary(editing, SaveEditing, StartEditing),
		h.NewQs("id", recordID),
	),
	h.Text(h.Ternary(editing, "Save", "Edit")),
)
`

const ConditionalClassExample = `h.ClassX("w-10 h-10 border rounded-full", h.ClassMap{
	"border-green-500": task.CompletedAt != nil,
	"border-slate-400": task.CompletedAt == nil,
})`

const IfElseExample2 = `h.IfElse(editing, EditTaskForm(), ViewTask())`

const IfElseExample3 = `h.IfElseLazy(editing, EditTaskForm, ViewTask)`

const ClassIfElseExample = `
h.Div(
	h.ClassIf(task.CompletedAt != nil, "border-green-500"),
)
`
