package astgen

import (
	"fmt"
	"github.com/franchb/htmgo/cli/htmgo/v2/internal/dirutil"
	"github.com/franchb/htmgo/cli/htmgo/v2/tasks/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAstGen(t *testing.T) {
	t.Parallel()

	workingDir, err := filepath.Abs("./project-sample")

	assert.NoError(t, err)
	process.SetWorkingDir(workingDir)
	assert.NoError(t, os.Chdir(workingDir))

	err = dirutil.DeleteDir(filepath.Join(process.GetWorkingDir(), "__htmgo"))
	assert.NoError(t, err)
	err = process.Run(process.NewRawCommand("", "go build ."))
	assert.Error(t, err)
	err = GenAst()
	assert.NoError(t, err)

	go func() {
		// project was buildable after astgen, confirmed working
		runErr := process.Run(process.NewRawCommand("server", "go run ."))
		assert.NoError(t, runErr)
	}()

	defer func() {
		assert.NotNil(t, process.GetProcessByName("server"))
		// Stop marks the kill as deliberate, so Run returns nil on Windows too,
		// where taskkill makes the process exit with status 1.
		process.Stop("server", 10*time.Second)
	}()

	// `go run .` compiles the sample first, which can take well over a second
	// on a cold CI runner; wait until the server answers instead of sleeping.
	require.Eventually(t, func() bool {
		resp, e := http.Get("http://localhost:3000/")
		if e != nil {
			return false
		}
		resp.Body.Close()
		return true
	}, 2*time.Minute, 250*time.Millisecond, "server did not start")

	urls := []string{
		"/astgen-project-sample/partials.CountersPartial",
		"/",
		"/astgen-project-sample/pages.TestPartial",
	}

	wg := sync.WaitGroup{}

	for _, url := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// ensure we can get a 200 response on the partials
			resp, e := http.Get(fmt.Sprintf("http://localhost:3000%s", url))
			if !assert.NoError(t, e) {
				return
			}
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("%s was not a 200 response", url))
		}()
	}

	wg.Wait()
}
