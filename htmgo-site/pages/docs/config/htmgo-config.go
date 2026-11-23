package config

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func HtmgoConfig(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Htmgo Config"),
			Text(`
				Certain aspects of htmgo can be configured via a htmgo.yml file in the root of your project.
				Here is an example configuration file:
			`),
			ui.CodeSnippet(ui.CodeSnippetProps{
				Code:            htmgoConfig,
				Lang:            "yaml",
				HideLineNumbers: true,
			}),
			NextStep(
				"mt-4",
				PrevBlock("Formatter", DocPath("/misc/formatter")),
				NextBlock("Examples", "/examples"),
			),
		),
	)
}

const htmgoConfig = `
# htmgo configuration

# if tailwindcss is enabled, htmgo will automatically compile your tailwind and output it to assets/dist
tailwind: true

# optional: select a Tailwind release, such as "v3.4.16", or use "latest"
# if omitted, htmgo auto-detects: projects using @tailwind directives get v3.4.16,
# otherwise defaults to latest (v4)
# tailwind_version: latest

# paths to ignore while watching files, relative to the project root
watch_ignore: [".git", "node_modules", "dist/**", "assets/dist"]

# files to watch for changes, supports glob patterns through https://github.com/bmatcuk/doublestar
watch_files: ["**/*.go", "**/*.css", "**/*.md"]

# files or directories to ignore when automatically registering routes for pages
# supports glob patterns through https://github.com/bmatcuk/doublestar
automatic_page_routing_ignore: ["root.go"]

# files or directories to ignore when automatically registering routes for partials
# supports glob patterns through https://github.com/bmatcuk/doublestar
automatic_partial_routing_ignore: []

# URL prefix for public assets
public_asset_path: "/public"
`
