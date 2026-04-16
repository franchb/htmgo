package urlhelper

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func GetClientIp(c fiber.Ctx) string {
	// Try to get the real client IP from the 'CF-Connecting-IP' header
	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// If not available, fall back to 'X-Forwarded-For'
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	// Otherwise, use the default remote address (this will be Cloudflare's IP)
	remote := c.IP()

	if strings.HasPrefix(remote, "[::1]") {
		return "localhost"
	}

	return remote
}
