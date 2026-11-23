package pushing_data

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func ServerSentEvents(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Server-Sent Events (SSE)"),
			Text(`Use SSE to push updates from a server to the browser over an open HTTP connection. The bundled htmgo extension uses <code>EventSource</code> to receive messages and applies out-of-band swaps from their HTML payloads.`),
			h.P(
				h.Text("The "),
				Link("chat example", "/examples/chat"),
				h.Text(" uses SSE to deliver new messages to other participants."),
			),
			SubTitle("Connect to the stream"),
			Text(`Add <code>sse-connect</code> to an element and point it to a GET endpoint that streams <code>text/event-stream</code> responses. To update the page, send HTML elements with <code>hx-swap-oob</code> and IDs matching their targets.`),
			ui.GoCodeSnippet(SseConnectAttribute),
			SubTitle("Handle connection errors and closure"),
			Text(`Attach handlers to the connecting element. <code>h.HxOnSseError</code> receives the native EventSource error event in <code>event.detail.event</code>. The browser may reconnect after an error. An explicit <code>event: close</code> frame ends the stream and can include a reason in <code>event.detail.event.data</code> when handled by <code>h.HxOnSseClose</code>.`),
			ui.GoCodeSnippet(SseErrorHandlingExample),
			Text(`Use <code>h.HxOnSseConnecting</code> and <code>h.HxOnSseOpen</code> for connection setup. Use <code>h.HxBeforeSseMessage</code> and <code>h.HxAfterSseMessage</code> to react before or after the browser applies an update.`),
			SubTitle("Send a message"),
			Text(`Clients send messages through ordinary HTTP requests, such as a form POST. Clear the input after that request succeeds so incoming messages leave your draft intact.`),
			ui.GoCodeSnippet(SseClearInputExample),
			NextStep(
				"mt-4",
				PrevBlock("Custom Cache Stores", DocPath("/performance/pluggable-caches")),
				NextBlock("HTMX extensions", DocPath("/htmx-extensions/overview")),
			),
		),
	)
}

const SseConnectAttribute = `
func MessageFeed(streamURL string) *h.Element {
	return h.Div(
		h.Attribute("sse-connect", streamURL),
		h.Div(h.Id("messages")),
	)
}
`

const SseErrorHandlingExample = `
func MessageFeed(streamURL string) *h.Element {
	return h.Div(
		h.Attribute("sse-connect", streamURL),
		h.HxOnSseError(
			js.EvalJs("console.error('SSE connection error', event.detail.event);"),
		),
		h.HxOnSseClose(
			js.EvalJs("console.info('SSE connection closed', event.detail.event.data);"),
		),
		h.Div(h.Id("messages")),
	)
}
`

const SseClearInputExample = `
func MessageForm(sendMessage h.PartialFunc) *h.Element {
	return h.Form(
		h.PostPartial(sendMessage),
		h.NoSwap(),
		h.HxAfterRequest(
			js.EvalJs("if (event.detail.ctx.response?.ok) self.reset();"),
		),
		h.Input("text", h.Name("message"), h.Required()),
		h.Button(h.Type("submit"), h.Text("Send")),
	)
}
`
