package main

import (
	"fmt"
	"io/fs"

	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/franchb/htmgo/framework/config"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
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

			_ = sub

			// change this in htmgo.yml (public_asset_path)
			app.Router.Use(cfg.PublicAssetPath, h.StaticCacheMiddleware())
			app.Router.Get(fmt.Sprintf("%s/*", cfg.PublicAssetPath), static.New("./assets/dist"))

			__htmgo.Register(app.Router)
		},
	})
}
