package util

import (
	"github.com/franchb/htmgo/framework/h"
)

func GetClientIp(ctx *h.RequestContext) string {
	// Try to get the real client IP from the 'CF-Connecting-IP' header
	if ip := ctx.Header("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// If not available, fall back to 'X-Forwarded-For'
	if ip := ctx.Header("X-Forwarded-For"); ip != "" {
		return ip
	}

	// Otherwise, use the default remote address
	remote := ctx.Fiber.IP()

	if remote == "::1" {
		return "localhost"
	}

	return remote
}
