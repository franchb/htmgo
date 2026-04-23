package urlhelper

import (
	"net/url"

	"github.com/franchb/htmgo/framework/v2/h"
)

func ToAbsoluteUrl(ctx *h.RequestContext, path string) string {
	// Define the relative path you want to add
	relativePath := path

	// Parse the current request URL from Fiber context
	currentURL, err := url.Parse(ctx.Fiber.OriginalURL())
	if err != nil {
		currentURL = &url.URL{Path: "/"}
	}

	// Set scheme and host from the request to create an absolute URL
	scheme := ctx.Fiber.Protocol()
	currentURL.Host = ctx.Fiber.Hostname()
	currentURL.Scheme = scheme

	// Combine the base URL with the relative path
	absoluteURL := currentURL.ResolveReference(&url.URL{Path: relativePath})

	// Output the full absolute URL
	return absoluteURL.String()
}
