package astgen

import (
	"go/format"
	"log"
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

	bytes := []byte(cb())

	slog.Debug("astgen.WriteFile", slog.String("path", path), slog.String("content", string(bytes)))

	var err error
	bytes, err = format.Source(bytes)

	if err != nil {
		log.Printf("Failed to format source: %v\n", err.Error())
		data := string(bytes)
		println(data)
		return
	}

	// Save the buffer to a file
	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		PanicF("Failed to write buffer to file: %v", err)
	}
}
