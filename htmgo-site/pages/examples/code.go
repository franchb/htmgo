package examples

import (
	"strings"
	"time"

	"github.com/franchb/htmgo/framework/v2/h"

	"htmgo-site/partials/snippets"
	"htmgo-site/ui"
)

func GetGithubPath(path string) string {
	return "https://github.com/franchb/htmgo/blob/master/htmgo-site/partials/snippets/" + snippetFilename(path)
}

func snippetFilename(path string) string {
	return strings.TrimPrefix(path, "/examples/") + ".go"
}

var RenderCodeToStringCached = h.CachedPerKeyT(time.Minute*30, func(snippet *Snippet) (string, h.GetElementFunc) {
	return snippet.path, func() *h.Element {
		return renderCodeToString(snippet)
	}
})

func renderCodeToString(snippet *Snippet) *h.Element {
	source, err := snippets.Source.ReadFile(snippetFilename(snippet.path))
	if err != nil {
		return h.Empty()
	}

	return ui.CodeSnippet(ui.CodeSnippetProps{
		Code:         string(source),
		Lang:         "go",
		CustomStyles: []string{"border-radius: 0.5rem;"},
		WrapLines:    false,
	})
}
