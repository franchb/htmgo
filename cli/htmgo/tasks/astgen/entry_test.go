package astgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Generated asset constant names must not change when the title-casing
// implementation does; these are the outputs of the former strings.Title.
func TestToPascaleCase(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"assets_dist_main_css", "AssetsDistMainCss"},
		{"favicon_32x32_png", "Favicon32x32Png"},
		{"public_logo@2x_png", "PublicLogo@2xPng"},
		{"_index_html", "IndexHtml"},
		{"über_datei_svg", "ÜberDateiSvg"},
		{"a'b_c", "A'BC"},
		{"my file_png", "My FilePng"},
		{"1abc", "1abc"},
		{"ABC_def", "AbcDef"},
		{"", ""},
		{"___", ""},
		{"x_9", "X9"},
		{"straße_png", "StraßePng"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, toPascaleCase(c.in), c.in)
	}
}
