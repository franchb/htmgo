## **htmgo**

### build simple and scalable systems with go + htmx

-------
[![Go Report Card](https://goreportcard.com/badge/github.com/franchb/htmgo)](https://goreportcard.com/report/github.com/franchb/htmgo)
![Build](https://github.com/franchb/htmgo/actions/workflows/run-framework-tests.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/franchb/htmgo/framework/v2/h.svg)](https://htmgo.dev/docs)
[![Join Discord](https://img.shields.io/badge/Join%20Discord-gray?style=flat&logo=discord&logoColor=white&link=https://htmgo.dev/discord)](https://htmgo.dev/discord)




<sup>looking for a python version? check out: https://fastht.ml</sup>

**introduction:**

htmgo is a lightweight pure go way to build interactive websites / web applications using go & htmx.

By combining the speed & simplicity of go + hypermedia attributes ([htmx](https://htmx.org)) to add interactivity to websites, all conveniently wrapped in pure go, you can build simple, fast, interactive websites without touching javascript. All compiled to a **single deployable binary**.

```go
func IndexPage(ctx *h.RequestContext) *h.Page {
  now := time.Now()
  return h.NewPage(
    h.Div(
      h.Class("flex gap-2"),
      h.TextF("the current time is %s", now.String())
    )
  )
}
```

**core features:**

1. deployable single binary
2. live reload (rebuilds css, go, ent schema, and routes upon change)
3. automatic page and partial registration based on file path
4. built in tailwindcss support, no need to configure anything by default
5. custom [htmx extensions](https://github.com/franchb/htmgo/tree/b610aefa36e648b98a13823a6f8d87566120cfcc/framework/assets/js/htmxextensions) to reduce boilerplate with common tasks

**get started:**

View documentation on [htmgo.dev](https://htmgo.dev/docs).

## What this fork adds

This is a maintained fork of upstream [`maddalax/htmgo`](https://github.com/maddalax/htmgo).
It keeps the original's design and API shape while modernizing the stack and adding
new capabilities:

- **Fiber v3 routing** — replaces `go-chi/chi/v5` with [`gofiber/fiber/v3`](https://github.com/gofiber/fiber). `RequestContext` wraps `fiber.Ctx`, and `AppOpts.FiberConfig` exposes the underlying Fiber config.
- **htmx 4** — upgraded to htmx `4.0.0-beta2` (from htmx 2). Colon-form event constants, explicit attribute inheritance via `Hx*Inherited` helpers, self-registering extensions (no more `hx-ext`), and all JS extensions rewritten for htmx 4's `registerExtension` API. New helpers: `HxSource`, `HxSourceID`, `HxRequestType`, `StatusAttr`, `ConfigAttr`.
- **Alpine.js integration** — a bundled `alpine-compat` htmx extension preserves Alpine state across morph swaps, plus a new `framework/ax/` Go package with Alpine directive builders that mirror `framework/hx/`.
- **Tailwind CSS v4** — docs site and starter template migrated to Tailwind v4 (CSS-based config, no `tailwind.config.js`).
- **Go 1.26** across all modules (up from 1.23).
- **`/v2` module paths** — all library modules follow Go Semantic Import Versioning (e.g. `github.com/franchb/htmgo/framework/v2`). Install the CLI with `go install github.com/franchb/htmgo/cli/htmgo/v2@latest`.
- **Performance audit** — WebSocket leak fixes, config-aware static-asset prefixing, buffer-pool caps, and other subsystem-wide improvements.
- **Project hygiene** — a Keep-a-Changelog [`CHANGELOG.md`](CHANGELOG.md), Dependabot for Go / GitHub Actions / npm, a `vitest` suite for the JS extensions, and Cloudflare Pages docs hosting at [htmgo.franchb.com](https://htmgo.franchb.com).

See [`CHANGELOG.md`](CHANGELOG.md) for the full 2.0.0 breaking-change list and migration recipe.

## Claude Code skills

This repo ships [Claude Code](https://docs.claude.com/en/docs/claude-code/overview) skills under `.claude/skills/` that teach AI coding sessions how to work with htmgo:

- `htmgo-guidance` — writing htmgo apps (builder, pages, partials, hx/ax helpers, caching, CLI).
- `htmx-guidance` — htmx 4 patterns and best practices.
- `htmx-debugging`, `htmx-extension-authoring`, `htmx-migration`, `htmx-upgrade-from-htmx2` — specialized htmx skills.

To use `htmgo-guidance` in a project that consumes this fork:

```bash
# From your consumer project root:
mkdir -p .claude/skills

# If you have htmgo cloned locally:
cp -r /path/to/htmgo/.claude/skills/htmgo-guidance .claude/skills/

# Or fetch directly:
git clone --depth=1 https://github.com/franchb/htmgo.git /tmp/htmgo
cp -r /tmp/htmgo/.claude/skills/htmgo-guidance .claude/skills/
rm -rf /tmp/htmgo
```

Run `/skills` in Claude Code to verify `htmgo-guidance` is loaded. Repeat for any other skills you want.

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=franchb/htmgo&type=Date)](https://star-history.com/#franchb/htmgo&Date)
