package dirutil

import (
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"strings"
)

func matchesAny(patterns []string, path string) bool {
	for _, pattern := range patterns {
		matched, err :=
			doublestar.Match(strings.ReplaceAll(pattern, `\`, "/"), strings.ReplaceAll(path, `\`, "/"))
		if err != nil {
			fmt.Printf("Error matching pattern: %v\n", err)
			return false
		}
		if matched {
			return true
		}
	}
	return false
}

func IsGlobExclude(path string, excludePatterns []string) bool {
	return matchesAny(excludePatterns, path)
}

func IsGlobMatch(path string, patterns []string, excludePatterns []string) bool {
	if matchesAny(excludePatterns, path) {
		return false
	}
	return matchesAny(patterns, path)
}

// BuildOutputDir is where `htmgo watch` and `htmgo run` compile the server
// binary. It is never a source of changes, so the watcher always skips it.
const BuildOutputDir = "__htmgo/temp-build"

// IsBuildOutput reports whether path is the build output dir or inside it.
// path may be relative ("./__htmgo/temp-build/app") and use either separator.
func IsBuildOutput(path string) bool {
	p := strings.TrimPrefix(strings.ReplaceAll(path, `\`, "/"), "./")
	return p == BuildOutputDir || strings.HasPrefix(p, BuildOutputDir+"/")
}
