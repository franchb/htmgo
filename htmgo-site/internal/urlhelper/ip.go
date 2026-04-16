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
		// X-Forwarded-For may contain multiple IPs; use only the first (client) IP.
		if parts := strings.SplitN(ip, ",", 2); len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
		return ip
	}

	// Otherwise, use the default remote address (this will be Cloudflare's IP)
	remote := c.IP()

	// c.IP() returns "::1" without brackets for IPv6 loopback.
	if remote == "::1" || strings.HasPrefix(remote, "[::1]") || remote == "127.0.0.1" {
		return "localhost"
	}

	return remote
}
