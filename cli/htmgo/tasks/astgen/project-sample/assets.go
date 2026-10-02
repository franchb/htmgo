//go:build !prod

package main

import (
	"astgen-project-sample/internal/embedded"
	"io/fs"
)

func GetStaticAssets() fs.FS {
	return embedded.NewOsFs()
}
