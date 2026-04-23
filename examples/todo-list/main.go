package main

import (
	"embed"
	"io/fs"

	"github.com/gofiber/fiber/v3/middleware/static"
	_ "github.com/mattn/go-sqlite3"

	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/service"

	"todolist/__htmgo"
	"todolist/ent"
	"todolist/infrastructure/db"
)

//go:embed assets/dist/*
var StaticAssets embed.FS

func main() {
	locator := service.NewLocator()

	service.Set[ent.Client](locator, service.Singleton, func() *ent.Client {
		return db.Provide()
	})

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {

			sub, err := fs.Sub(StaticAssets, "assets/dist")

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
