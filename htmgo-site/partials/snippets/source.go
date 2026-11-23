package snippets

import "embed"

// Source contains the Go source shown beside each example.
//
//go:embed *.go
var Source embed.FS
