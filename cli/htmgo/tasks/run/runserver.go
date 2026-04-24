package run

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/franchb/htmgo/cli/htmgo/v2/tasks/process"
)

func Server(isWatcher bool, flags ...process.RunFlag) error {
	buildDir := "./__htmgo/temp-build"
	_ = os.RemoveAll(buildDir)
	err := os.Mkdir(buildDir, 0755)

	if err != nil {
		return err
	}

	if isWatcher {
		err := process.Run(process.NewRawCommand("", fmt.Sprintf("go build -o %s", buildDir)))
		if err != nil {
			return err
		}
	} else {
		process.RunOrExit(process.NewRawCommand("", fmt.Sprintf("go build -o %s", buildDir)))
	}
	binaryPath := ""

	// find the binary that was built
	err = filepath.WalkDir(buildDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		binaryPath = path
		return nil
	})

	if err != nil {
		return err
	}

	if binaryPath == "" {
		return fmt.Errorf("could not find the binary")
	}

	return process.Run(process.NewRawCommand("run-server", fmt.Sprintf("./%s", binaryPath), flags...))
}
