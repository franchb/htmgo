package main

import (
	"github.com/gofiber/fiber/v3"

	"github.com/franchb/htmgo/framework/v2/h"
)

func RenderToString(element *h.Element) string {
	return h.Render(element)
}

func RenderPage(c fiber.Ctx, page func(ctx *h.RequestContext) *h.Page) error {
	ctx := h.RequestContext{
		Fiber: c,
	}
	return h.HtmlView(c, page(&ctx))
}

func RenderPartial(c fiber.Ctx, partial func(ctx *h.RequestContext) *h.Partial) error {
	ctx := h.RequestContext{
		Fiber: c,
	}
	return h.PartialView(c, partial(&ctx))
}
