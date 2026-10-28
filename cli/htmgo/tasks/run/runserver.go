package run

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/franchb/htmgo/cli/htmgo/v2/tasks/process"
)

const (
	serverProcessName = "run-server"
	buildDir          = "./__htmgo/temp-build"
)

// buildMu serialises the stop/clean/build phase. File-change events can start
// several Server calls at once; without the lock one call can delete the build
// directory while another is still compiling into it.
var buildMu sync.Mutex

func Server(isWatcher bool, flags ...process.RunFlag) error {
	buildMu.Lock()
	binaryPath, err := buildServer(isWatcher)
	buildMu.Unlock()
	if err != nil {
		return err
	}
	return process.Run(process.NewRawCommand(serverProcessName, fmt.Sprintf("./%s", binaryPath), flags...))
}

func buildServer(isWatcher bool) (string, error) {
	// Stop the running server before touching its directory. On Windows a
	// running executable is locked: it can't be deleted or overwritten, so
	// RemoveAll fails silently and the following Mkdir reports
	// "Cannot create a file when that file already exists".
	process.Stop(serverProcessName, 5*time.Second)

	if err := resetDir(buildDir, 5*time.Second); err != nil {
		return "", err
	}

	build := process.NewRawCommand("", fmt.Sprintf("go build -o %s", buildDir))
	if isWatcher {
		if err := process.Run(build); err != nil {
			return "", err
		}
	} else {
		process.RunOrExit(build)
	}

	return findBinary(buildDir)
}

// resetDir removes dir and recreates it empty. Removal is retried until
// timeout because Windows can keep a file locked for a short while after the
// process that used it has exited (antivirus scanners, the Search indexer).
func resetDir(dir string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		err := os.RemoveAll(dir)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("could not clear %s: %w", dir, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	return os.MkdirAll(dir, 0755)
}

func findBinary(dir string) (string, error) {
	binaryPath := ""
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		binaryPath = path
		return nil
	})
	if err != nil {
		return "", err
	}
	if binaryPath == "" {
		return "", fmt.Errorf("could not find the binary in %s", dir)
	}
	return binaryPath, nil
}
