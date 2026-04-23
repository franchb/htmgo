package ws

import (
	"github.com/franchb/htmgo/extensions/websocket/v2/internal/wsutil"
	"github.com/franchb/htmgo/framework/v2/h"
)

type Metrics struct {
	Manager wsutil.ManagerMetrics
	Handler HandlerMetrics
}

func MetricsFromCtx(ctx *h.RequestContext) Metrics {
	manager := ManagerFromCtx(ctx)
	return Metrics{
		Manager: manager.Metrics(),
		Handler: GetHandlerMetics(),
	}
}
