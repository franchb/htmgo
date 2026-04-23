# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is htmgo

A lightweight Go web framework for building interactive websites using Go + htmx, compiled to a single deployable binary. No JavaScript required. Original upstream: `github.com/franchb/htmgo`.

## Repository Layout (multi-module monorepo)

- **framework/** — Core framework module (`github.com/franchb/htmgo/framework`). The main package is `h/` which contains the HTML builder, renderer, request context, lifecycle events, caching, and swap logic. Sub-packages: `hx/` (htmx constants), `service/` (service locator DI), `config/`, `h/cache/` (pluggable cache stores), `assets/`, `js/`.
- **framework-ui/** — Pre-built UI component library built on top of the framework.
- **cli/htmgo/** — CLI tool. Commands: `template`, `run`, `watch`, `build`, `setup`, `css`, `schema`, `generate`, `format`, `version`. Key task packages live under `cli/htmgo/tasks/` (astgen, copyassets, css, reloader, run, etc.).
- **extensions/websocket/** — WebSocket extension module.
- **examples/** — Example apps (chat, hackernews, simple-auth, todo-list, ws-example, minimal-htmgo). Each has its own `Taskfile.yml`.
- **htmgo-site/** — Documentation website source.
- **templates/** — Starter project templates.

## Build & Test Commands

### Framework tests (the primary test suite)
```bash
cd framework && go test ./...
```

### CLI tests (AST generation)
```bash
cd cli/htmgo/tasks/astgen && go test ./...
```

### Run a single test
```bash
cd framework && go test ./h/ -run TestSimpleRender
```

### Running examples (requires `task` CLI — https://taskfile.dev)
```bash
cd examples/todo-list && task watch   # live-reload dev mode
cd examples/todo-list && task build   # production build
cd examples/todo-list && task run     # run built binary
```

Examples use `go run github.com/franchb/htmgo/cli/htmgo@latest <command>` under the hood. The htmgo-site Taskfile calls `htmgo` directly (requires `go install ./cli/htmgo`).

### Install htmgo CLI from source
```bash
cd cli/htmgo && go install .
```

### Building htmgo-site (documentation)
```bash
cd htmgo-site && npm install && htmgo build   # or: task build
```
Projects using Tailwind CSS plugins (e.g., `@tailwindcss/typography`) need `npm install` before building.

## Project Configuration (htmgo.yml)

Each htmgo app has an `htmgo.yml` (or `htmgo.yaml`, `_htmgo.yaml`, `_htmgo.yml`) that controls:
- `tailwind: true/false` — enable Tailwind CSS compilation
- `watch_ignore` / `watch_files` — glob patterns for live reload
- `automatic_page_routing_ignore` / `automatic_partial_routing_ignore` — exclude files from auto-routing
- `public_asset_path` — static asset URL prefix (default: `/public`)

## Core Architecture Concepts

### Ren interface & Element
Everything renderable implements the `Ren` interface (`Render(*RenderContext)`). The `Element` struct is the primary node type — created via builder functions in `h/tag.go` (e.g., `h.Div(...)`, `h.Button(...)`). Children and attributes are passed as variadic `Ren` arguments.

### Page vs Partial
- `Page` — Full HTML page response. Created with `h.NewPage(root)`. Route handlers return `*h.Page`.
- `Partial` — HTMX fragment response with optional response headers. Created with `h.NewPartial(root)`. Used for htmx partial updates.
- Routes are auto-registered from file paths via AST generation (`cli/htmgo/tasks/astgen/`).

### RequestContext
Wraps a Fiber v3 `fiber.Ctx` (exposed as `RequestContext.Fiber`) with htmx-aware helpers: `FormValue`, `QueryParam`, `UrlParam`, `IsHxRequest`, `Redirect`, header access, cookie management. Use `GetRequestContext(c fiber.Ctx)` to retrieve it inside Fiber handlers/middleware.

### Service Locator (DI)
`service.NewLocator()` provides dependency injection with `Singleton` and `Transient` lifecycles. Register with `service.Set[T](locator, lifecycle, provider)`, resolve with `service.Get[T](locator)`.

### Lifecycle & Commands
`h/lifecycle.go` — Event handlers (`OnClick`, `OnLoad`, etc.) that emit JavaScript `Command` objects (either `SimpleJsCommand` or `ComplexJsCommand`). These translate to htmx event attributes.

### HTMX Integration
Targets **htmx 4.0.0-beta2** (exact pin in `framework/assets/js/package.json`). `hx/` package defines constants for htmx attributes, events (colon form, e.g. `htmx:after:swap`), swap strategies, and headers. Attribute inheritance is **explicit** — use `HxTargetInherited`, `HxIncludeInherited`, etc. to propagate to descendants. `hx-ext` is gone: extensions self-register on import. Custom TypeScript extensions live in `framework/assets/js/htmxextensions/` (`htmgo`, `debug`, `livereload`, `mutation-error`, `path-deps`, `trigger-children`, `response-targets`, `sse`, `ws`, `ws-event-handler`, `alpine-compat`), built with `tsup` and tested with `vitest`. See `CHANGELOG.md` 1.2.0-beta.1 for the full breaking list and migration recipe.

### JS asset build & tests
```bash
cd framework/assets/js
npm run build        # tsup → framework/assets/dist/htmgo.js
npm test             # vitest run
```

### Routing
Uses **Fiber v3** (`github.com/gofiber/fiber/v3`) — migrated from `go-chi/chi/v5`. `App.Router` is a `*fiber.App`. App startup in `h/app.go` via `h.Start(h.AppOpts{...})`. The auto-generated `__htmgo.Register(router)` function registers all page/partial routes based on file structure. Downstream consumers that previously embedded chi handlers must port to Fiber's `fiber.Ctx`-based handler signature.

### Caching
`h/cache.go` provides high-level caching helpers. `h/cache/` defines the `Store[K,V]` interface with `Set`, `GetOrCompute`, `Delete`, `Purge`, `Close` methods. Built-in: `TTLStore`. `LRUStore` is provided as an example implementation in `h/cache/lru_store_example.go`.

### Auto-generated Code
The `__htmgo/` directory in each app is generated by the CLI (`htmgo generate` / `astgen`). Contains `pages-generated.go`, `partials-generated.go`, and `setup-generated.go`. Do not edit these files manually — they are regenerated on build/watch.

## Go Version

Go 1.23.0 across all modules.

## Testing Conventions

- Uses `stretchr/testify` for assertions (`assert`, `require`).
- Tests use `t.Parallel()` where possible.
- The largest test file is `h/render_test.go` — covers HTML rendering extensively.

## Smoke Testing with Playwright MCP

When a Playwright MCP server is available, validate the htmgo-site builds and renders correctly.

### 1. Build and start the site
```bash
cd htmgo-site
npm install                         # needed for Tailwind CSS plugins
export PATH="$(go env GOPATH)/bin:$PATH"
htmgo build
PORT=3123 ./dist/htmgo-site &       # background; use non-default port to avoid conflicts
```

### 2. Browser checks (via Playwright MCP tools)
- **Homepage** (`http://localhost:3123/`): `browser_snapshot` should show "htmgo" h1 heading, "Get Started" link to `/docs`, nav bar with Docs/Examples/Convert HTML links
- **Docs redirect** (`http://localhost:3123/docs`): should redirect to `/docs/introduction` — page title "Docs - Introduction", sidebar with all doc sections, Introduction content with code example
- **Nested docs** (`http://localhost:3123/docs/core-concepts/components`): page title "Docs - Components", content renders with code examples and prev/next navigation
- **Examples** (`http://localhost:3123/examples`): sidebar with example categories (Forms, Interactivity, Projects, Components), each with links
- **Console errors** (`browser_console_messages` level `error`): should return 0 errors

### 3. Stop the server
```bash
pkill -f "dist/htmgo-site"
```
