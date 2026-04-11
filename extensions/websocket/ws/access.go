package ws

import (
	"github.com/franchb/htmgo/extensions/websocket/internal/wsutil"
	"github.com/franchb/htmgo/framework/h"
)

func ManagerFromCtx(ctx *h.RequestContext) *wsutil.SocketManager {
	return wsutil.SocketManagerFromCtx(ctx)
}
