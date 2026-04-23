package snippets

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/js"
)

func JsHideChildrenOnClick(ctx *h.RequestContext) *h.Partial {
	text := h.Pf("- Parent")
	return h.NewPartial(
		h.Div(
			text,
			h.Class("cursor-pointer"),
			h.Id("js-test"),
			h.OnClick(
				js.ToggleClassOnChildren("div", "hidden"),
				js.EvalCommands(
					text,
					js.ToggleText("+ Parent", "- Parent"),
				),
			),
			h.Div(
				h.Class("ml-4"),
				h.Text("Child 1"),
			),
			h.Div(
				h.Class("ml-4"),
				h.Text("Child 2"),
			),
		),
	)
}
