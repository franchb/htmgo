package ws

import (
	"github.com/franchb/htmgo/extensions/websocket/v2/internal/wsutil"
	"github.com/franchb/htmgo/framework/v2/h"
)

func ManagerFromCtx(ctx *h.RequestContext) *wsutil.SocketManager {
	return wsutil.SocketManagerFromCtx(ctx)
}
