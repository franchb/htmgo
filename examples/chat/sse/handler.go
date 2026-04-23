package sse

import (
	"bufio"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/service"
)

func Handle() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Set the necessary headers
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")

		cc := h.GetRequestContext(c)
		locator := cc.ServiceLocator()
		manager := service.Get[SocketManager](locator)

		sessionId := c.Cookies("session_id")
		roomId := c.Params("id")

		return c.SendStreamWriter(func(w *bufio.Writer) {
			/*
				Large buffer in case the client disconnects while we are writing
				we don't want to block the writer
			*/
			done := make(chan bool, 1000)
			writer := make(WriterChan, 1000)

			wg := sync.WaitGroup{}
			wg.Add(1)

			/*
			 * This goroutine is responsible for writing messages to the client
			 */
			go func() {
				defer wg.Done()
				defer manager.Disconnect(sessionId)

				defer func() {
					for len(writer) > 0 {
						<-writer
					}
					for len(done) > 0 {
						<-done
					}
				}()

				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-done:
						fmt.Printf("closing connection: \n")
						return
					case <-ticker.C:
						manager.Ping(sessionId)
					case message := <-writer:
						_, err := fmt.Fprint(w, message)
						if err != nil {
							done <- true
						} else {
							if flushErr := w.Flush(); flushErr != nil {
								done <- true
							}
						}
					}
				}
			}()

			/**
			 * This goroutine is responsible for adding the client to the room
			 */
			wg.Add(1)
			go func() {
				defer wg.Done()
				if sessionId == "" {
					manager.writeCloseRaw(writer, "no session")
					return
				}

				if roomId == "" {
					slog.Error("invalid room", slog.String("room_id", roomId))
					manager.writeCloseRaw(writer, "invalid room")
					return
				}

				manager.Add(roomId, sessionId, writer, done)
			}()

			wg.Wait()
		})
	}
}
