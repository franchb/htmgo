package main

import (
	"io/fs"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/service"

	"htmgo-site/__htmgo"
	"htmgo-site/internal/cache"
	"htmgo-site/internal/markdown"
	"htmgo-site/internal/sitemap"
	"htmgo-site/internal/urlhelper"
)

func main() {
	locator := service.NewLocator()
	staticAssets := GetStaticAssets()
	markdownAssets := GetMarkdownAssets()

	service.Set(locator, service.Singleton, markdown.NewRenderer)
	service.Set(locator, service.Singleton, cache.NewSimpleCache)

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {

			app.Use(func(ctx *h.RequestContext) {
				// Log the details of the incoming request
				log.Printf("Method: %s, URL: %s, RemoteAddr: %s", ctx.Fiber.Method(), ctx.Fiber.OriginalURL(), urlhelper.GetClientIp(ctx.Fiber))
			})

			app.Use(func(ctx *h.RequestContext) {
				ctx.Set("embeddedMarkdown", markdownAssets)
			})

			sub, err := fs.Sub(staticAssets, "assets/dist")

			if err != nil {
				panic(err)
			}

			app.Router.Get("/sitemap.xml", func(c fiber.Ctx) error {
				s, err := sitemap.Generate(app.Router)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).SendString("failed to generate sitemap")
				}
				c.Set("Content-Type", "application/xml")
				return c.Send(s)
			})

			app.Router.Get("/public/*", h.StaticCacheMiddleware, static.New("", static.Config{
				FS: sub,
			}))

			__htmgo.Register(app.Router)
		},
	})
}
