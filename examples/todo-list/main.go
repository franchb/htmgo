package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
	_ "github.com/mattn/go-sqlite3"

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

			app.Router.Handle("/public/*", h.StaticCacheMiddleware(http.StripPrefix("/public", http.FileServerFS(sub))))

			__htmgo.Register(app.Router)
		},
	})
}
