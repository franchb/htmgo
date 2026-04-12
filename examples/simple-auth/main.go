package main

import (
	"io/fs"
	"net/http"

	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"

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

			app.Router.Handle("/public/*", h.StaticCacheMiddleware(http.StripPrefix("/public", http.FileServerFS(sub))))
			__htmgo.Register(app.Router)
		},
	})
}
