package websocket

import (
	"github.com/franchb/htmgo/extensions/websocket/v2/internal/wsutil"
	"github.com/franchb/htmgo/extensions/websocket/v2/opts"
	"github.com/franchb/htmgo/extensions/websocket/v2/ws"
	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/service"
)

func EnableExtension(app *h.App, opts opts.ExtensionOpts) {
	if app.Opts.ServiceLocator == nil {
		app.Opts.ServiceLocator = service.NewLocator()
	}

	if opts.WsPath == "" {
		panic("websocket: WsPath is required")
	}

	if opts.SessionId == nil {
		panic("websocket: SessionId func is required")
	}

	service.Set[wsutil.SocketManager](app.Opts.ServiceLocator, service.Singleton, func() *wsutil.SocketManager {
		manager := wsutil.NewSocketManager(&opts)
		manager.StartMetrics()
		return manager
	})
	ws.StartListener(app.Opts.ServiceLocator)
	app.Router.Get(opts.WsPath, wsutil.WsHandler(&opts))
}
