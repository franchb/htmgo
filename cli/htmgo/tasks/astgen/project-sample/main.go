package main

import (
	"astgen-project-sample/__htmgo"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
)

func main() {
	locator := service.NewLocator()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			__htmgo.Register(app.Router)
		},
	})
}
