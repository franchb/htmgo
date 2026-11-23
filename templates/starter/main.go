package main

import (
	"fmt"
	"io/fs"

	"github.com/franchb/htmgo/framework/v2/config"
	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/service"
	"github.com/gofiber/fiber/v3/middleware/static"
	"starter-template/__htmgo"
)

func main() {
	locator := service.NewLocator()
	cfg := config.Get()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")

			if err != nil {
				panic(err)
			}

			// change this in htmgo.yml (public_asset_path)
			app.Router.Use(cfg.PublicAssetPath, h.StaticCacheMiddleware)
			app.Router.Get(fmt.Sprintf("%s/*", cfg.PublicAssetPath), static.New("", static.Config{FS: sub}))

			__htmgo.Register(app.Router)
		},
	})
}
