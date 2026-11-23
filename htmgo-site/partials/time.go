package partials

import (
	"time"

	"github.com/franchb/htmgo/framework/v2/h"

	"htmgo-site/internal/sitebuild"
)

func CurrentTimeDemo() *h.Element {
	if !sitebuild.Static() {
		return h.Div(h.GetPartial(CurrentTimePartial, "load, every 1s"))
	}

	return h.Div(
		h.Class("flex flex-wrap gap-1 items-center"),
		h.Text("The current time is "),
		h.Tag("time", h.Class("font-bold")),
		h.UnsafeRawScript(`
			(() => {
				const clock = document.currentScript.previousElementSibling;
				function update() {
					if (!clock.isConnected) return;
					const timestamp = new Date().toISOString().replace(/\.\d{3}Z$/, "Z");
					clock.dateTime = timestamp;
					clock.textContent = timestamp;
					window.setTimeout(update, 1000);
				}
				update();
			})();
		`),
	)
}

func CurrentTimePartial(ctx *h.RequestContext) *h.Partial {
	now := time.Now()
	return h.NewPartial(
		h.Div(
			h.Class("flex gap-1 items-center"),
			h.Pf("The current time is "),
			h.Span(
				h.Text(now.Format(time.RFC3339)),
				h.Class("font-bold"),
			),
		),
	)
}
