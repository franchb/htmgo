package pages

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func RootPage(children ...h.Ren) h.Ren {
	return h.Html(
		h.Meta("viewport", "width=device-width, initial-scale=1"),
		h.Meta("title", "htmgo chat example"),
		h.Meta("charset", "utf-8"),
		h.Meta("author", "htmgo"),
		h.Head(
			h.Link("/public/main.css", "stylesheet"),
			h.Script("/public/htmgo.js"),
		),
		h.Body(
			h.Div(
				h.Class("flex flex-col gap-2 bg-white h-full"),
				h.Fragment(children...),
			),
		),
	)
}
