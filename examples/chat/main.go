package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"runtime"
	"time"

	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"

	"chat/__htmgo"
	"chat/chat"
	"chat/internal/db"
	"chat/sse"
)

func main() {
	locator := service.NewLocator()

	service.Set[db.Queries](locator, service.Singleton, db.Provide)
	service.Set[sse.SocketManager](locator, service.Singleton, func() *sse.SocketManager {
		return sse.NewSocketManager()
	})

	chatManager := chat.NewManager(locator)
	go chatManager.StartListener()

	go func() {
		for {
			count := runtime.NumGoroutine()
			fmt.Printf("goroutines: %d\n", count)
			time.Sleep(10 * time.Second)
		}
	}()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")

			if err != nil {
				panic(err)
			}

			app.Router.Handle("/public/*", h.StaticCacheMiddleware(http.StripPrefix("/public", http.FileServerFS(sub))))
			app.Router.Handle("/sse/chat/{id}", sse.Handle())

			__htmgo.Register(app.Router)
		},
	})
}
