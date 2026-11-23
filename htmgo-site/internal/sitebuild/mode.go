package sitebuild

import "os"

// UpstreamOrigin hosts the upstream Go-backed tools and examples.
const UpstreamOrigin = "https://htmgo.dev"

// Static reports whether the server is rendering pages for the static export.
func Static() bool {
	return os.Getenv("HTMGO_SITE_STATIC") == "1"
}
