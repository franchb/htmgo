package interactivity

import . "htmgo-site/pages/docs"
import "htmgo-site/ui"
import "github.com/franchb/htmgo/framework/v2/h"

func Swapping(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Swapping"),
			Text(`An htmx request can replace part of a page with HTML from a partial. Use an out-of-band swap to update several elements from one response.`),
			h.P(
				h.Text("The counter form uses "),
				Link("hx-swap-oob", "https://four.htmx.org/reference/attributes/hx-swap-oob/"),
				h.Text(" to replace the form and its separate status message by ID."),
			),
			Text(`Add these functions under <code>pages/</code> in the starter project and import <code>strconv</code>.`),
			ui.GoCodeSnippet(SwapExample),
			Text(`Submitting the form sends a POST to <code>CounterPartial</code>. <code>h.NoSwap</code> suppresses the main swap. <code>h.SwapManyPartial</code> marks both returned elements for out-of-band swaps. Give each replacement the same ID as the element already in the page.`),
			Text(`Use <code>h.GetPartialPathWithQs</code> to construct a partial URL with query parameters. Use <code>ctx.QueryParam</code> to read them in the receiving handler.`),
			ui.GoCodeSnippet(SwapGetPartialPathWithQsExample),
			Text(`Pass a trigger such as <code>hx.BlurEvent</code> to <code>h.PostPartial</code> to submit when an input loses focus.`),
			ui.GoCodeSnippet(SwapGetPartialPathExampleOnBlur),
			NextStep(
				"mt-4",
				PrevBlock("Loops / Dealing With Lists", DocPath("/control/loops")),
				NextBlock("Events / Commands", DocPath("/interactivity/events")),
			),
		),
	)
}

const SwapExample = `
func CounterPage(ctx *h.RequestContext) *h.Page {
	return RootPage(
		h.Div(CounterForm(0), CounterStatus(0)),
	)
}

func CounterPartial(ctx *h.RequestContext) *h.Partial {
	count, _ := strconv.Atoi(ctx.FormValue("count"))
	count++
	return h.SwapManyPartial(ctx, CounterForm(count), CounterStatus(count))
}

func CounterForm(count int) *h.Element {
	return h.Form(
		h.Id("counter-form"),
		h.PostPartial(CounterPartial),
		h.NoSwap(),
		h.Input("hidden", h.Name("count"), h.Value(count)),
		h.Button(h.Type("submit"), h.Text("Increment")),
	)
}

func CounterStatus(count int) *h.Element {
	return h.P(
		h.Id("counter-status"),
		h.TextF("Count: %d", count),
	)
}
`

const SwapGetPartialPathWithQsExample = `
func CounterPath(count int) string {
	return h.GetPartialPathWithQs(
		CounterPartial,
		h.NewQs("count", strconv.Itoa(count)),
	)
}
`

const SwapGetPartialPathExampleOnBlur = `
func CounterInput(count int) *h.Element {
	return h.Input("number",
		h.Name("count"),
		h.Value(count),
		h.PostPartial(CounterPartial, hx.BlurEvent),
		h.NoSwap(),
	)
}
`
