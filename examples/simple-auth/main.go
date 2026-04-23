package main

import (
	"io/fs"

	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/service"

	"simpleauth/__htmgo"
	"simpleauth/internal/db"
)

func main() {
	locator := service.NewLocator()

	service.Set(locator, service.Singleton, func() *db.Queries {
		return db.Provide()
	})

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")

			if err != nil {
				panic(err)
			}

			app.Router.Use("/public", h.StaticCacheMiddleware)
			app.Router.Get("/public/*", static.New("", static.Config{
				FS: sub,
			}))
			__htmgo.Register(app.Router)
		},
	})
}
