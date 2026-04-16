package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func main() {
	app := fiber.New()

	app.Get("/public/*", static.New("./public"))

	app.Get("/", func(c fiber.Ctx) error {
		return RenderPage(c, Index)
	})

	app.Get("/current-time", func(c fiber.Ctx) error {
		return RenderPartial(c, CurrentTime)
	})

	app.Listen(":3000")
}
