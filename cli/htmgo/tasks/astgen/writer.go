package astgen

import (
	"go/format"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/franchb/htmgo/cli/htmgo/tasks/process"
)

func WriteFile(path string, cb func() string) {
	currentDir := process.GetWorkingDir()

	path = filepath.Join(currentDir, path)

	slog.Debug("astgen.WriteFile", slog.String("path", path))

	dir := filepath.Dir(path)

	os.MkdirAll(dir, 0755)

	src := []byte(cb())

	slog.Debug("astgen.WriteFile", slog.String("path", path), slog.String("content", string(src)))

	formatted, err := format.Source(src)

	if err != nil {
		PanicF("failed to format generated source for %s: %v\n%s", path, err, string(src))
	}

	// Save the buffer to a file
	err = os.WriteFile(path, formatted, 0644)
	if err != nil {
		PanicF("Failed to write buffer to file: %v", err)
	}
}
