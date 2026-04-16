package main

import (
	"io/fs"

	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/franchb/htmgo/extensions/websocket"
	ws2 "github.com/franchb/htmgo/extensions/websocket/opts"
	"github.com/franchb/htmgo/extensions/websocket/session"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"

	"ws-example/__htmgo"
)

func main() {
	locator := service.NewLocator()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {

			app.Use(func(ctx *h.RequestContext) {
				session.CreateSession(ctx)
			})

			websocket.EnableExtension(app, ws2.ExtensionOpts{
				WsPath: "/ws",
				RoomName: func(ctx *h.RequestContext) string {
					return "all"
				},
				SessionId: func(ctx *h.RequestContext) string {
					return ctx.QueryParam("sessionId")
				},
			})

			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")

			if err != nil {
				panic(err)
			}

			app.Router.Get("/public/*", static.New("", static.Config{
				FS: sub,
			}))
			__htmgo.Register(app.Router)
		},
	})
}
