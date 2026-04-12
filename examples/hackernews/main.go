package main

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"

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

			app.Router.Handle("/item", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id := r.URL.Query().Get("id")
				w.Header().Set("Location", fmt.Sprintf("/?item=%s", id))
				w.WriteHeader(302)
			}))
			app.Router.Handle("/public/*", h.StaticCacheMiddleware(http.StripPrefix("/public", http.FileServerFS(sub))))
			__htmgo.Register(app.Router)
		},
	})
}
