package util

import (
	"strings"

	"github.com/franchb/htmgo/framework/h"
)

func GetClientIp(ctx *h.RequestContext) string {
	// Try to get the real client IP from the 'CF-Connecting-IP' header
	if ip := ctx.Header("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// If not available, fall back to 'X-Forwarded-For'
	if ip := ctx.Header("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For may contain multiple IPs; use only the first (client) IP.
		if parts := strings.SplitN(ip, ",", 2); len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
		return ip
	}

	// Otherwise, use the default remote address
	remote := ctx.Fiber.IP()

	if remote == "::1" {
		return "localhost"
	}

	return remote
}
