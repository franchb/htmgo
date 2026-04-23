package examples

import (
	"bytes"
	"fmt"
	"github.com/franchb/htmgo/framework/v2/h"
	"htmgo-site/ui"
	"io"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

func GetGithubPath(path string) string {
	path = strings.ReplaceAll(path, "/examples/", "/snippets/")
	return fmt.Sprintf("https://github.com/franchb/htmgo/tree/master/htmgo-site/partials%s.go", path)
}

func GetGithubRawPath(path string) string {
	path = strings.ReplaceAll(path, "/examples/", "/snippets/")
	return fmt.Sprintf("https://raw.githubusercontent.com/franchb/htmgo/master/htmgo-site/partials%s.go", path)
}

var RenderCodeToStringCached = h.CachedPerKeyT(time.Minute*30, func(snippet *Snippet) (string, h.GetElementFunc) {
	return snippet.path, func() *h.Element {
		return renderCodeToString(snippet)
	}
})

func renderCodeToString(snippet *Snippet) *h.Element {
	source := ""
	// in development, use the local file
	if h.IsDevelopment() {
		ptr := reflect.ValueOf(snippet.partial).Pointer()
		fnInfo := runtime.FuncForPC(ptr)
		if fnInfo == nil {
			return h.Empty()
		}
		file, _ := fnInfo.FileLine(ptr)
		b, err := os.ReadFile(file)
		if err != nil {
			return h.Empty()
		}
		source = string(b)
	} else {
		url := GetGithubRawPath(snippet.path)
		slog.Info("getting snippet source code", slog.String("url", url))
		resp, err := http.Get(url)
		if err != nil {
			return h.Empty()
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return h.Empty()
		}
		out := bytes.NewBuffer(nil)
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return h.Empty()
		}
		source = out.String()
	}

	return ui.CodeSnippet(ui.CodeSnippetProps{
		Code:         source,
		Lang:         "go",
		CustomStyles: []string{"border-radius: 0.5rem;"},
		WrapLines:    false,
	})
}
