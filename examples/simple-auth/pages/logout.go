package pages

import "github.com/franchb/htmgo/framework/h"

func LogoutPage(ctx *h.RequestContext) *h.Page {

	// clear the session cookie
	ctx.Fiber.Set(
		"Set-Cookie",
		"session_id=; Path=/; Max-Age=0",
	)

	ctx.Fiber.Set(
		"Location",
		"/login",
	)

	ctx.Fiber.Status(302)

	return nil
}
