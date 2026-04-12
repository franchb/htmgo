package h

import "strings"

var (
	baseExtensionsProd = strings.Join(
		[]string{"path-deps", "response-targets", "mutation-error", "htmgo", "sse"}, ", ")
	baseExtensionsDev = strings.Join(
		[]string{"path-deps", "response-targets", "mutation-error", "htmgo", "sse", "livereload"}, ", ")
)

// BaseExtensions returns the comma-separated list of htmx extensions to load.
// The strings are pre-computed at init time; this function simply selects
// between the development and production variant based on IsDevelopment().
func BaseExtensions() string {
	if IsDevelopment() {
		return baseExtensionsDev
	}
	return baseExtensionsProd
}
