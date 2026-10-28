package dirutil

import "testing"

func TestIsBuildOutput(t *testing.T) {
	cases := map[string]bool{
		"__htmgo/temp-build":            true,
		"./__htmgo/temp-build":          true,
		"__htmgo/temp-build/myapp":      true,
		`__htmgo\temp-build\myapp.exe`:  true,
		"__htmgo/temp-build-other":      false,
		"__htmgo/pages-generated.go":    false,
		"pages/__htmgo/temp-build/x.go": false,
		"pages/index.go":                false,
	}
	for path, want := range cases {
		if got := IsBuildOutput(path); got != want {
			t.Errorf("IsBuildOutput(%q) = %v, want %v", path, got, want)
		}
	}
}
