package wsutil

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gofiber/fiber/v3"

	ws2 "github.com/franchb/htmgo/extensions/websocket/opts"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
)

// websocketGUID is the magic GUID defined in RFC 6455 section 4.2.2
// used to compute the Sec-WebSocket-Accept header.
const websocketGUID = "258EAFA5-E914-47DA-95CA-5AB5DC525B41"

// computeAcceptKey computes the Sec-WebSocket-Accept value from
// the client's Sec-WebSocket-Key, per RFC 6455 section 4.2.2.
func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte(websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// WsHandler returns a Fiber handler that upgrades connections to WebSocket
// using gobwas/ws for the protocol framing.
//
// Because Fiber is built on fasthttp (which does not expose a standard
// net/http ResponseWriter with Hijack support), we use fasthttp's native
// Hijack mechanism:
//  1. Read the Sec-WebSocket-Key from the request.
//  2. Tell fasthttp not to send its own response (HijackSetNoResponse).
//  3. Register a HijackHandler that receives the raw net.Conn.
//  4. Inside the HijackHandler, write the 101 Switching Protocols response
//     manually (RFC 6455 opening handshake).
//  5. Run the read/write loops with gobwas/ws/wsutil on the hijacked conn.
func WsHandler(opts *ws2.ExtensionOpts) fiber.Handler {

	if opts.RoomName == nil {
		opts.RoomName = func(ctx *h.RequestContext) string {
			return "all"
		}
	}

	return func(c fiber.Ctx) error {
		cc := h.GetRequestContext(c)
		if cc == nil {
			return c.Status(fiber.StatusInternalServerError).SendString("websocket: missing request context")
		}
		locator := cc.ServiceLocator()
		manager := service.Get[SocketManager](locator)

		sessionId := opts.SessionId(cc)
		if sessionId == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		roomId := opts.RoomName(cc)

		// Read the WebSocket key before the handler returns — the request
		// data is only valid during this call.
		wsKey := c.Get("Sec-WebSocket-Key")
		if wsKey == "" {
			return c.Status(fiber.StatusBadRequest).SendString("websocket: missing Sec-WebSocket-Key header")
		}

		acceptKey := computeAcceptKey(wsKey)

		// Prevent fasthttp from writing its own response; we will write
		// the 101 Switching Protocols response ourselves in the hijack
		// handler.
		fctx := c.RequestCtx()
		fctx.HijackSetNoResponse(true)
		fctx.Hijack(func(conn net.Conn) {
			// Write the WebSocket upgrade response manually.
			upgradeResponse := "HTTP/1.1 101 Switching Protocols\r\n" +
				"Upgrade: websocket\r\n" +
				"Connection: Upgrade\r\n" +
				"Sec-WebSocket-Accept: " + acceptKey + "\r\n" +
				"\r\n"

			if _, err := conn.Write([]byte(upgradeResponse)); err != nil {
				slog.Info("failed to write websocket upgrade response", slog.String("error", err.Error()))
				conn.Close()
				return
			}

			runWebSocket(conn, manager, sessionId, roomId)
		})

		return nil
	}
}

// runWebSocket manages the read and write loops for a single WebSocket
// connection. It blocks until both loops finish.
func runWebSocket(conn net.Conn, manager *SocketManager, sessionId, roomId string) {
	/*
		Large buffer in case the client disconnects while we are writing
		we don't want to block the writer
	*/
	done := make(chan bool, 1000)
	writer := make(WriterChan, 1000)

	// Shared context for coordinating reader and writer shutdown.
	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}

	manager.Add(roomId, sessionId, writer, done)

	// cleanup closes the connection and cancels the shared context.
	// It is safe to call multiple times.
	cleanupOnce := sync.Once{}
	cleanup := func() {
		cleanupOnce.Do(func() {
			cancel()
			conn.Close()
		})
	}

	/*
	 * This goroutine is responsible for writing messages to the client
	 */
	wg.Add(1)
	go func() {
		defer manager.Disconnect(sessionId)
		defer wg.Done()
		defer cleanup()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				fmt.Printf("closing connection: \n")
				return
			case <-ticker.C:
				manager.Ping(sessionId)
			case message := <-writer:
				err := wsutil.WriteServerMessage(conn, ws.OpText, []byte(message))
				if err != nil {
					return
				}
			}
		}
	}()

	/*
	 * This goroutine is responsible for reading messages from the client
	 */
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cleanup()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				return
			}
			if op != ws.OpText {
				return
			}
			m := make(map[string]any)
			err = json.Unmarshal(msg, &m)
			if err != nil {
				return
			}
			manager.OnMessage(sessionId, m)
		}
	}()

	wg.Wait()
}
