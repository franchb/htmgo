package control

import . "htmgo-site/pages/docs"
import "htmgo-site/ui"
import "github.com/franchb/htmgo/framework/v2/h"

func Loops(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Loops / Dealing With Lists"),
			Text(`
				Use <code>h.List</code> to render a slice in its existing order. Use <code>h.IterMap</code> to render a map when iteration order can vary.
			`),
			Text("Example: Rendering a list of tasks"),
			ui.GoCodeSnippet(ListExample),
			Text("Example: Rendering a map"),
			ui.GoCodeSnippet(MapExample),
			NextStep(
				"mt-4",
				PrevBlock("Conditionals", DocPath("/control/if-else")),
				NextBlock("Adding Interactivity", DocPath("/interactivity/swapping")),
			),
		),
	)
}

const ListExample = `
func TaskList() *h.Element {
	tasks := []string{"Write docs", "Review examples", "Ship release"}
	return h.Ul(
		h.List(tasks, func(task string, _ int) *h.Element {
			return h.Li(h.Text(task))
		}),
	)
}
`

const MapExample = `
func StatusList() *h.Element {
	statuses := map[string]string{
		"docs": "ready",
		"release": "pending",
	}
	return h.Ul(
		h.IterMap(statuses, func(name, status string) *h.Element {
			return h.Li(h.TextF("%s: %s", name, status))
		}),
	)
}
`
