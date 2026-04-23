package main

import (
	"fmt"
	"io/fs"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/service"

	"hackernews/__htmgo"
)

func main() {
	locator := service.NewLocator()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")

			if err != nil {
				panic(err)
			}

			app.Router.Get("/item", func(c fiber.Ctx) error {
				id := c.Query("id")
				return c.Redirect().Status(302).To(fmt.Sprintf("/?item=%s", id))
			})
			app.Router.Use("/public", h.StaticCacheMiddleware)
			app.Router.Get("/public/*", static.New("", static.Config{
				FS: sub,
			}))
			__htmgo.Register(app.Router)
		},
	})
}
