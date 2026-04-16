package h

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var Version = uuid.NewString()

func sseHandler(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")

	return c.SendStreamWriter(func(w *bufio.Writer) {
		for {
			_, err := fmt.Fprintf(w, "data: %s\n\n", Version)
			if err != nil {
				break
			}
			if err := w.Flush(); err != nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	})
}

func (app *App) AddLiveReloadHandler(path string) {
	app.Router.Get(path, sseHandler)
}
