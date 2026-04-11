package partials

import (
	"chat/components"
	"chat/sse"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
)

func SendMessage(ctx *h.RequestContext) *h.Partial {
	locator := ctx.ServiceLocator()
	socketManager := service.Get[sse.SocketManager](locator)

	sessionCookie, err := ctx.Request.Cookie("session_id")

	if err != nil {
		return h.SwapPartial(ctx, components.FormError("Session not found"))
	}

	message := ctx.Request.FormValue("message")

	if message == "" {
		return h.SwapPartial(ctx, components.FormError("Message is required"))
	}

	if len(message) > 200 {
		return h.SwapPartial(ctx, components.FormError("Message is too long"))
	}

	socketManager.OnMessage(sessionCookie.Value, map[string]any{
		"message": message,
	})

	return h.EmptyPartial()
}
