## **htmgo**

### build simple and scalable systems with go + htmx

-------
[![Go Report Card](https://goreportcard.com/badge/github.com/franchb/htmgo)](https://goreportcard.com/report/github.com/franchb/htmgo)
[![Tests](https://github.com/franchb/htmgo/actions/workflows/tests.yml/badge.svg)](https://github.com/franchb/htmgo/actions/workflows/tests.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/franchb/htmgo/framework/v2/h.svg)](https://pkg.go.dev/github.com/franchb/htmgo/framework/v2/h)

**introduction:**

htmgo is a lightweight pure go way to build interactive websites / web applications using go & htmx.

By combining the speed & simplicity of go + hypermedia attributes ([htmx](https://htmx.org)) to add interactivity to websites, all conveniently wrapped in pure go, you can build simple, fast, interactive websites without touching javascript. All compiled to a **single deployable binary**.

A page in `pages/index.go` can use `time` and `github.com/franchb/htmgo/framework/v2/h` to render the current time.

```go
func IndexPage(ctx *h.RequestContext) *h.Page {
	now := time.Now()
	return h.NewPage(
		h.Div(
			h.Class("flex gap-2"),
			h.TextF("the current time is %s", now.String()),
		),
	)
}
```

**core features:**

1. deployable single binary
2. live reload (rebuilds css, go, ent schema, and routes upon change)
3. file-based page routes and automatic registration of partial handlers
4. built in tailwindcss support, no need to configure anything by default
5. custom [htmx extensions](https://github.com/franchb/htmgo/tree/master/framework/assets/js/htmxextensions) to reduce boilerplate with common tasks

**get started:**

View documentation on [htmgo.franchb.com](https://htmgo.franchb.com/docs).

## What this fork adds

This is a maintained fork of upstream [`maddalax/htmgo`](https://github.com/maddalax/htmgo).
It keeps the original's design and API shape while modernizing the stack and adding
new capabilities:

- **Fiber v3 routing** — replaces `go-chi/chi/v5` with [`gofiber/fiber/v3`](https://github.com/gofiber/fiber). `RequestContext` wraps `fiber.Ctx`, and `AppOpts.FiberConfig` exposes the underlying Fiber config.
- **htmx 4** — the bundled client supports colon-form events, explicit attribute inheritance through `Hx*Inherited` helpers, and extensions that register on import. The [JS manifest](framework/assets/js/package.json) pins the bundled version.
- **Alpine.js integration** — a bundled `alpine-compat` htmx extension preserves Alpine state across morph swaps, plus a new `framework/ax/` Go package with Alpine directive builders that mirror `framework/hx/`.
- **Tailwind CSS v4** — docs site and starter template migrated to Tailwind v4 (CSS-based config, no `tailwind.config.js`).
- **Go 1.27** across all modules (up from 1.23).
- **`/v2` module paths** — all library modules follow Go Semantic Import Versioning (e.g. `github.com/franchb/htmgo/framework/v2`). Install the CLI with `go install github.com/franchb/htmgo/cli/htmgo/v2@latest`.
- **Performance audit** — WebSocket leak fixes, config-aware static-asset prefixing, buffer-pool caps, and other subsystem-wide improvements.
- **Project hygiene** — a Keep-a-Changelog [`CHANGELOG.md`](CHANGELOG.md), Dependabot for Go / GitHub Actions / npm, a `vitest` suite for the JS extensions, and Cloudflare Pages docs hosting at [htmgo.franchb.com](https://htmgo.franchb.com).

See [`CHANGELOG.md`](CHANGELOG.md) for the full 2.0.0 breaking-change list and migration recipe.

## Claude Code skill

The repository ships [htmgo-guidance](.claude/skills/htmgo-guidance/SKILL.md) for writing v2 apps. It covers pages, partials, htmx and Alpine helpers, caching, and the CLI.

Claude Code can use the skill from this checkout. To install it in a consumer project, run the following from that project's root:

```bash
mkdir -p .claude/skills/htmgo-guidance
curl -fsSL https://raw.githubusercontent.com/franchb/htmgo/master/.claude/skills/htmgo-guidance/SKILL.md \
  -o .claude/skills/htmgo-guidance/SKILL.md
```

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=franchb/htmgo&type=Date)](https://star-history.com/#franchb/htmgo&Date)
