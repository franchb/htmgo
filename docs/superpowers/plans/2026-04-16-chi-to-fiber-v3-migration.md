# Chi to Fiber v3 Migration — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace go-chi/chi/v5 with gofiber/fiber/v3 v3.10.0 as the HTTP router throughout the htmgo monorepo.

**Architecture:** Direct replacement — Fiber replaces chi everywhere. `RequestContext` wraps `fiber.Ctx` instead of `*http.Request`/`http.ResponseWriter`. The AST code generator emits Fiber-native handlers. No compatibility layer. Full API break.

**Tech Stack:** Go 1.26, gofiber/fiber/v3 v3.10.0, gofiber/fiber/v3/middleware/static, gofiber/contrib/v3/websocket, gobwas/ws (retained for low-level WS ops in extension)

**Spec:** `docs/superpowers/specs/2026-04-16-chi-to-fiber-v3-migration-design.md`

---

## File Structure

### Files modified
- `framework/go.mod` — swap chi dep for fiber
- `framework/h/app.go` — RequestContext, App, Start, middleware, view helpers
- `framework/h/header.go` — CurrentPath: `ctx.Request.Header.Get()` → `ctx.Fiber.Get()`
- `framework/h/qs.go` — GetQueryParam: `ctx.Request.URL.Query()` → `ctx.Fiber.Query()`
- `framework/h/livereload.go` — SSE handler: net/http → Fiber streaming
- `framework/h/header_test.go` — TestCurrentPath: uses `RequestContext{Request: req}`
- `framework/h/qs_test.go` — TestGetQueryParam: uses `RequestContext{Request: req}`
- `cli/htmgo/tasks/astgen/entry.go` — code gen templates, constants, formatRoute
- `cli/htmgo/tasks/astgen/project-sample/main.go` — test app uses chi patterns
- `cli/htmgo/tasks/astgen/project-sample/go.mod` — chi indirect dep → fiber
- `extensions/websocket/init.go` — handler registration: `app.Router.Handle()` → `app.Router.Get()`
- `extensions/websocket/internal/wsutil/handler.go` — WsHttpHandler → WsHandler (fiber.Ctx)
- `extensions/websocket/go.mod` — remove chi indirect, framework brings fiber
- `htmgo-site/internal/sitemap/generate.go` — `chi.Mux.Routes()` → `fiber.App.Stack()`
- `htmgo-site/main.go` — UseWithContext removal, sitemap handler, static serving
- `htmgo-site/internal/urlhelper/ip.go` — GetClientIp: `*http.Request` → `fiber.Ctx`
- `examples/chat/main.go` — static serving, SSE route registration
- `examples/chat/sse/handler.go` — SSE handler rewrite
- `examples/chat/pages/chat.$id.go` — `chi.URLParam(ctx.Request, "id")` → `ctx.UrlParam("id")`
- `examples/chat/go.mod` — remove direct chi dep
- `examples/hackernews/main.go` — custom /item handler, static serving
- `examples/simple-auth/main.go` — static serving
- `examples/todo-list/main.go` — static serving
- `examples/ws-example/main.go` — static serving, websocket extension
- `examples/minimal-htmgo/main.go` — chi.NewRouter → fiber.New
- `examples/minimal-htmgo/render.go` — RenderPage/RenderPartial: net/http → fiber.Ctx
- `examples/minimal-htmgo/go.mod` — remove direct chi dep

---

## Task 1: Framework go.mod — Swap chi for Fiber

**Files:**
- Modify: `framework/go.mod`

- [ ] **Step 1: Remove chi, add fiber**

```bash
cd framework && go get -u github.com/gofiber/fiber/v3@v3.10.0 && go mod tidy
```

Expected: `go.mod` now has `github.com/gofiber/fiber/v3 v3.10.0` and no longer has `github.com/go-chi/chi/v5`. There will be compilation errors until we update the source files — that's expected.

- [ ] **Step 2: Verify go.mod content**

Run: `head -20 framework/go.mod`
Expected: Shows `github.com/gofiber/fiber/v3` in require block. `go-chi/chi` is gone.

Note: Do NOT run `go mod tidy` again until the source files are updated (it would re-add chi or fail). The go.sum may also need regeneration later.

- [ ] **Step 3: Commit**

```bash
git add framework/go.mod framework/go.sum
git commit -m "deps(framework): swap chi/v5 for fiber/v3 in go.mod"
```

---

## Task 2: Framework core — RequestContext and App struct

**Files:**
- Modify: `framework/h/app.go:1-398`

This is the largest single change. Replace the entire `app.go` with Fiber-based types and functions.

- [ ] **Step 1: Rewrite imports**

In `framework/h/app.go`, replace the import block:

```go
// Old imports (lines 3-19)
import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/franchb/htmgo/framework/config"
	"github.com/franchb/htmgo/framework/hx"
	"github.com/franchb/htmgo/framework/service"
)
```

Replace with:

```go
import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/franchb/htmgo/framework/config"
	"github.com/franchb/htmgo/framework/hx"
	"github.com/franchb/htmgo/framework/service"
)
```

Removed: `"context"`, `"net/http"`, `"github.com/go-chi/chi/v5"`.
Added: `"github.com/gofiber/fiber/v3"`.

- [ ] **Step 2: Replace RequestContext struct**

Replace lines 21-34:

```go
type RequestContext struct {
	Fiber             fiber.Ctx
	locator           *service.Locator
	isBoosted         bool
	currentBrowserUrl string
	hxPromptResponse  string
	isHxRequest       bool
	hxTargetId        string
	hxTriggerName     string
	hxTriggerId       string
	kv                map[string]interface{}
	parsedQuery       url.Values
}
```

- [ ] **Step 3: Replace GetRequestContext**

Replace lines 36-42:

```go
func GetRequestContext(c fiber.Ctx) *RequestContext {
	val := c.Locals(requestContextKey)
	if val == nil {
		return nil
	}
	return val.(*RequestContext)
}
```

- [ ] **Step 4: Replace SetCookie**

Replace lines 44-46:

```go
func (c *RequestContext) SetCookie(cookie *fiber.Cookie) {
	c.Fiber.Cookie(cookie)
}
```

- [ ] **Step 5: Replace Redirect**

Replace lines 48-57:

```go
func (c *RequestContext) Redirect(path string, code int) {
	if code == 0 {
		code = fiber.StatusTemporaryRedirect
	}
	if code < 300 || code > 399 {
		code = fiber.StatusTemporaryRedirect
	}
	c.Fiber.Redirect().Status(code).To(path)
}
```

- [ ] **Step 6: Replace HTTP method checks**

Replace lines 59-72:

```go
func (c *RequestContext) IsHttpPost() bool {
	return c.Fiber.Method() == fiber.MethodPost
}

func (c *RequestContext) IsHttpGet() bool {
	return c.Fiber.Method() == fiber.MethodGet
}

func (c *RequestContext) IsHttpPut() bool {
	return c.Fiber.Method() == fiber.MethodPut
}

func (c *RequestContext) IsHttpDelete() bool {
	return c.Fiber.Method() == fiber.MethodDelete
}
```

- [ ] **Step 7: Replace FormValue, Header, UrlParam, QueryParam**

Replace lines 75-92:

```go
func (c *RequestContext) FormValue(key string) string {
	return c.Fiber.FormValue(key)
}

func (c *RequestContext) Header(key string) string {
	return c.Fiber.Get(key)
}

func (c *RequestContext) UrlParam(key string) string {
	return c.Fiber.Params(key)
}

func (c *RequestContext) QueryParam(key string) string {
	return c.Fiber.Query(key)
}
```

Note: The `parsedQuery` cache field is no longer needed — Fiber caches query parsing internally. The field stays in the struct (unused) to avoid breaking test code that sets `isHxRequest` directly; we can remove it in a follow-up cleanup.

- [ ] **Step 8: Replace App struct and Start()**

Replace lines 151-164:

```go
type App struct {
	Opts   AppOpts
	Router *fiber.App
}

// Start starts the htmgo server
func Start(opts AppOpts) {
	app := fiber.New()
	instance := App{
		Opts:   opts,
		Router: app,
	}
	instance.start()
}
```

- [ ] **Step 9: Remove legacy RequestContextKey constant**

Delete lines 171-172:

```go
// Deprecated: RequestContextKey is the legacy string context key. Use GetRequestContext instead.
const RequestContextKey = "htmgo.request.context"
```

Keep the typed key (`requestContextKey` / `requestContextKeyType`).

- [ ] **Step 10: Replace populateHxFields**

Replace lines 174-182:

```go
func populateHxFields(cc *RequestContext) {
	cc.isBoosted = cc.Fiber.Get(hx.BoostedHeader) == "true"
	cc.currentBrowserUrl = cc.Fiber.Get(hx.CurrentUrlHeader)
	cc.hxPromptResponse = cc.Fiber.Get(hx.PromptResponseHeader)
	cc.isHxRequest = cc.Fiber.Get(hx.RequestHeader) == "true"
	cc.hxTargetId = cc.Fiber.Get(hx.TargetIdHeader)
	cc.hxTriggerName = cc.Fiber.Get(hx.TriggerNameHeader)
	cc.hxTriggerId = cc.Fiber.Get(hx.TriggerIdHeader)
}
```

- [ ] **Step 11: Remove UseWithContext, rewrite Use**

Delete `UseWithContext` entirely (lines 184-199). Replace `Use` (lines 201-213):

```go
func (app *App) Use(h func(ctx *RequestContext)) {
	app.Router.Use(func(c fiber.Ctx) error {
		cc := GetRequestContext(c)
		if cc == nil {
			return c.Next()
		}
		h(cc)
		return c.Next()
	})
}
```

- [ ] **Step 12: Replace StaticCacheMiddleware**

Replace lines 233-260 (remove `cacheControlWriter` struct entirely):

```go
// StaticCacheMiddleware adds Cache-Control headers for static file requests.
// Requests with a query string (e.g. ?v=hash) are treated as immutable with a
// one-year max-age; all other static requests get a one-hour max-age.
// Headers are only applied to successful responses (2xx/304) so that transient
// 404s or errors are not cached by browsers or CDNs.
func StaticCacheMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		status := c.Response().StatusCode()
		if status == fiber.StatusOK || status == fiber.StatusNotModified {
			cacheHeader := "public, max-age=3600"
			if len(c.Request().URI().QueryString()) > 0 {
				cacheHeader = "public, max-age=31536000, immutable"
			}
			c.Set("Cache-Control", cacheHeader)
		}
		return err
	}
}
```

- [ ] **Step 13: Replace app.start()**

Replace lines 262-346:

```go
func (app *App) start() {

	slog.SetLogLoggerLevel(GetLogLevel())

	publicPrefix := config.Get().PublicAssetPath
	if publicPrefix == "" {
		publicPrefix = "/public"
	}

	app.Router.Use(func(c fiber.Ctx) error {
		// Skip RequestContext creation for static file requests.
		if strings.HasPrefix(c.Path(), publicPrefix+"/") || c.Path() == publicPrefix {
			return c.Next()
		}

		cc := &RequestContext{
			locator: app.Opts.ServiceLocator,
			Fiber:   c,
		}
		populateHxFields(cc)
		c.Locals(requestContextKey, cc)
		return c.Next()
	})

	if app.Opts.Register != nil {
		app.Opts.Register(app)
	}

	if app.Opts.LiveReload && IsDevelopment() {
		app.AddLiveReloadHandler("/dev/livereload")
	}

	port := ":3000"
	isDefaultPort := true

	if os.Getenv("PORT") != "" {
		port = fmt.Sprintf(":%s", os.Getenv("PORT"))
		isDefaultPort = false
	}

	if app.Opts.Port != 0 {
		port = fmt.Sprintf(":%d", app.Opts.Port)
		isDefaultPort = false
	}

	if isDefaultPort {
		slog.Info("Using default port 3000, set PORT environment variable to change it or use AppOpts.Port")
	}

	slog.Info(fmt.Sprintf("Server started at localhost%s", port))

	if err := app.Router.Listen(port); err != nil {
		if IsDevelopment() && IsWatchMode() {
			slog.Info("Port already in use, trying to kill the process and start again")
			if runtime.GOOS == "windows" {
				cmd := exec.Command("cmd", "/C", fmt.Sprintf(`for /F "tokens=5" %%i in ('netstat -aon ^| findstr :%s') do taskkill /F /PID %%i`, port))
				cmd.Run()
			} else {
				cmd := exec.Command("bash", "-c", fmt.Sprintf("kill -9 $(lsof -ti%s)", port))
				cmd.Run()
			}

			time.Sleep(time.Millisecond * 50)

			if err := app.Router.Listen(port); err != nil {
				slog.Error("Failed to restart server", "error", err)
				panic(err)
			}
		}

		panic(err)
	}
}
```

- [ ] **Step 14: Replace view helper functions**

Replace lines 348-397:

```go
func writeHtml(c fiber.Ctx, element Ren) error {
	if element == nil {
		return nil
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(Render(element, WithDocType()))
}

func HtmlView(c fiber.Ctx, page *Page) error {
	if page == nil {
		return nil
	}
	return writeHtml(c, page.Root)
}

func PartialViewWithHeaders(c fiber.Ctx, headers *Headers, partial *Partial) error {
	if partial == nil {
		return nil
	}

	if partial.Headers != nil {
		for s, a := range *partial.Headers {
			c.Set(s, a)
		}
	}

	if headers != nil {
		for s, a := range *headers {
			c.Set(s, a)
		}
	}

	return writeHtml(c, partial.Root)
}

func PartialView(c fiber.Ctx, partial *Partial) error {
	if partial == nil {
		return nil
	}

	if partial.Headers != nil {
		for s, a := range *partial.Headers {
			c.Set(s, a)
		}
	}

	return writeHtml(c, partial.Root)
}
```

- [ ] **Step 15: Verify the file compiles (expect errors from other files)**

Run: `cd framework && go build ./h/ 2>&1 | head -30`
Expected: Errors from `header.go`, `qs.go`, `livereload.go` (they still reference `ctx.Request`). Errors from `app.go` itself should be zero.

- [ ] **Step 16: Commit**

```bash
git add framework/h/app.go
git commit -m "feat(framework): rewrite app.go for Fiber v3 — RequestContext, App, middleware, view helpers"
```

---

## Task 3: Framework — header.go, qs.go, livereload.go

**Files:**
- Modify: `framework/h/header.go:36-43`
- Modify: `framework/h/qs.go:1-80`
- Modify: `framework/h/livereload.go:1-41`

- [ ] **Step 1: Fix header.go — CurrentPath**

Replace line 37 in `framework/h/header.go`:

```go
current := ctx.Request.Header.Get(hx.CurrentUrlHeader)
```

with:

```go
current := ctx.Fiber.Get(hx.CurrentUrlHeader)
```

- [ ] **Step 2: Fix qs.go — GetQueryParam**

Replace lines 60-80 in `framework/h/qs.go`:

```go
// GetQueryParam returns the value of the given query parameter from the request URL.
// There are two layers of priority:
// 1. The query parameter in the Fiber request URL
// 2. The current browser URL
// If the query parameter is not found in the Fiber request, it will fall back to the current browser URL if set.
// The URL from the *RequestContext would normally be the url from an XHR request through htmx,
// which is not the current browser url a visitor may be on.
func GetQueryParam(ctx *RequestContext, key string) string {
	value := ctx.Fiber.Query(key)
	if value != "" {
		return value
	}
	// Fallback: parse from the current browser URL (htmx XHR vs browser URL)
	current := ctx.currentBrowserUrl
	if current != "" {
		u, err := url.Parse(current)
		if err == nil {
			return u.Query().Get(key)
		}
	}
	return ""
}
```

Also remove the `"strings"` import if it's no longer used (check — `strings` is NOT imported in qs.go, only `"net/url"` and `"strings"`. `strings` is used by `Qs.ToString` so keep it).

- [ ] **Step 3: Rewrite livereload.go**

Replace the entire content of `framework/h/livereload.go`:

```go
package h

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var Version = uuid.NewString()

func sseHandler(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			_, err := fmt.Fprintf(w, "data: %s\n\n", Version)
			if err != nil {
				break
			}
			if err := w.Flush(); err != nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	})
	return nil
}

func (app *App) AddLiveReloadHandler(path string) {
	app.Router.Get(path, sseHandler)
}
```

- [ ] **Step 4: Verify framework/h compiles**

Run: `cd framework && go build ./h/`
Expected: Success (no errors). All files in framework/h now reference Fiber types consistently.

- [ ] **Step 5: Commit**

```bash
git add framework/h/header.go framework/h/qs.go framework/h/livereload.go
git commit -m "feat(framework): migrate header.go, qs.go, livereload.go to Fiber v3"
```

---

## Task 4: Framework tests — Fix tests that create RequestContext with Request field

**Files:**
- Modify: `framework/h/header_test.go:41-47`
- Modify: `framework/h/qs_test.go:53-76`

- [ ] **Step 1: Fix TestCurrentPath in header_test.go**

Replace `TestCurrentPath` (lines 41-47):

```go
func TestCurrentPath(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		c.Request().Header.Set(hx.CurrentUrlHeader, "https://example.com/current-path")
		ctx := &RequestContext{Fiber: c}
		path := CurrentPath(ctx)
		assert.Equal(t, "/current-path", path)
		return nil
	})
	req, _ := http.NewRequest("GET", "/test", nil)
	app.Test(req)
}
```

Add these imports to header_test.go:

```go
import (
	"github.com/gofiber/fiber/v3"
	"github.com/franchb/htmgo/framework/hx"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
```

- [ ] **Step 2: Fix TestGetQueryParam in qs_test.go**

Replace `TestGetQueryParam` (lines 53-76):

```go
func TestGetQueryParam(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		ctx := &RequestContext{Fiber: c}

		result := GetQueryParam(ctx, "foo")
		assert.Equal(t, "bar", result)

		result = GetQueryParam(ctx, "baz")
		assert.Equal(t, "qux", result)

		result = GetQueryParam(ctx, "missing")
		assert.Equal(t, "", result)

		// Fallback to browser URL
		ctx.currentBrowserUrl = "http://localhost/?current=value"
		result = GetQueryParam(ctx, "current")
		assert.Equal(t, "value", result)

		return nil
	})
	req, _ := http.NewRequest("GET", "/test?foo=bar&baz=qux", nil)
	app.Test(req)
}
```

Add `"github.com/gofiber/fiber/v3"` to the imports in qs_test.go. Keep `"net/http"` (needed for `http.NewRequest` in `app.Test()`).

- [ ] **Step 3: Run framework tests**

Run: `cd framework && go test ./h/ -v -count=1 2>&1 | tail -40`
Expected: All tests pass. Pay attention to `TestCurrentPath` and `TestGetQueryParam` specifically.

- [ ] **Step 4: Run the full framework test suite**

Run: `cd framework && go test ./... -count=1`
Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add framework/h/header_test.go framework/h/qs_test.go
git commit -m "test(framework): update header and qs tests for Fiber v3 RequestContext"
```

---

## Task 5: Code generator — AST gen templates

**Files:**
- Modify: `cli/htmgo/tasks/astgen/entry.go:37-543`

- [ ] **Step 1: Replace constants**

Replace lines 38-39:

```go
const HttpModuleName = "net/http"
const ChiModuleName = "github.com/go-chi/chi/v5"
```

with:

```go
const FiberModuleName = "github.com/gofiber/fiber/v3"
```

Remove `HttpModuleName` — it's no longer needed in generated code.

- [ ] **Step 2: Update formatRoute — remove {param} wrapping**

In `formatRoute` (lines 315-345), delete the loop that wraps `:param` in `{param}` (lines 325-331):

```go
	parts := strings.Split(path, "/")

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = fmt.Sprintf("{%s}", part[1:])
		}
	}

	path = strings.Join(parts, "/")
```

Replace with just (Fiber uses `:param` directly, which the `$` → `:` replacement already produces):

```go
	// Fiber uses :param syntax — no wrapping needed.
	// The earlier $ → : replacement already produces the correct format.
```

So the `parts` splitting and rejoining is gone. The rest of `formatRoute` stays.

- [ ] **Step 3: Update buildGetPartialFromContext — Fiber handler template**

Replace the `routerHandlerMethod` closure in `buildGetPartialFromContext` (lines 237-248):

```go
	var routerHandlerMethod = func(path string, caller string) string {
		return fmt.Sprintf(`
			router.All("%s", func(c fiber.Ctx) error {
				cc := h.GetRequestContext(c)
				partial := %s(cc)
				if partial == nil {
					return c.SendStatus(404)
				}
				return h.PartialView(c, partial)
			})`, path, caller)
	}
```

Replace the `registerFunction` template (lines 258-262):

```go
	registerFunction := fmt.Sprintf(`
		func RegisterPartials(router *fiber.App) {
				%s
		}
	`, strings.Join(handlerMethods, "\n"))
```

- [ ] **Step 4: Update writePartialsFile — imports**

In `writePartialsFile` (lines 292-313), replace:

```go
	builder.AddImport(ChiModuleName)

	if len(partials) > 0 {
		builder.AddImport(ModuleName)
		builder.AddImport(HttpModuleName)
	}
```

with:

```go
	builder.AddImport(FiberModuleName)

	if len(partials) > 0 {
		builder.AddImport(ModuleName)
	}
```

- [ ] **Step 5: Update writePagesFile — Fiber handler template and imports**

In `writePagesFile` (lines 347-397):

Replace the import additions (lines 351-353):

```go
	builder.AddImport(HttpModuleName)
	builder.AddImport(ChiModuleName)
```

with:

```go
	builder.AddImport(FiberModuleName)
```

Replace the page handler body template (lines 374-381):

```go
		body += fmt.Sprintf(
			`
			router.Get("%s", func(c fiber.Ctx) error {
				cc := h.GetRequestContext(c)
				return h.HtmlView(c, %s(cc))
			})
			`, formatRoute(page.Path), call,
		)
```

Replace the function parameter (lines 386-388):

```go
		Parameters: []NameType{
			{Name: "router", Type: "*fiber.App"},
		},
```

- [ ] **Step 6: Update GenAst — setup-generated.go template**

In `GenAst` (lines 525-539), replace:

```go
	WriteFile("__htmgo/setup-generated.go", func() string {

		return fmt.Sprintf(`
			// Package __htmgo THIS FILE IS GENERATED. DO NOT EDIT.
			package __htmgo

			import (
				"%s"
			)

			func Register(r *fiber.App) {
				RegisterPartials(r)
				RegisterPages(r)
			}
		`, FiberModuleName)
	})
```

- [ ] **Step 7: Verify astgen compiles**

Run: `cd cli/htmgo && go build ./tasks/astgen/`
Expected: Success.

- [ ] **Step 8: Commit**

```bash
git add cli/htmgo/tasks/astgen/entry.go
git commit -m "feat(astgen): update code generator templates for Fiber v3"
```

---

## Task 6: Code generator test fixtures — project-sample

**Files:**
- Modify: `cli/htmgo/tasks/astgen/project-sample/main.go`
- Modify: `cli/htmgo/tasks/astgen/project-sample/go.mod`

- [ ] **Step 1: Update project-sample/main.go**

Replace the entire file:

```go
package main

import (
	"astgen-project-sample/__htmgo"
	"fmt"
	"io/fs"

	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/franchb/htmgo/framework/config"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
)

func main() {
	locator := service.NewLocator()
	cfg := config.Get()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")

			if err != nil {
				panic(err)
			}

			_ = sub

			app.Router.Get(fmt.Sprintf("%s/*", cfg.PublicAssetPath),
				static.New("./assets/dist"))

			__htmgo.Register(app.Router)
		},
	})
}
```

- [ ] **Step 2: Update project-sample/go.mod**

Run:
```bash
cd cli/htmgo/tasks/astgen/project-sample && go get github.com/gofiber/fiber/v3@v3.10.0 && go mod tidy
```

- [ ] **Step 3: Run astgen test**

Run: `cd cli/htmgo/tasks/astgen && go test -v -run TestAstGen -count=1 -timeout 30s`
Expected: Test passes — the generated code compiles and the test server responds with 200 on all routes.

- [ ] **Step 4: Commit**

```bash
git add cli/htmgo/tasks/astgen/project-sample/
git commit -m "test(astgen): update project-sample test fixture for Fiber v3"
```

---

## Task 7: WebSocket extension

**Files:**
- Modify: `extensions/websocket/init.go`
- Modify: `extensions/websocket/internal/wsutil/handler.go`
- Modify: `extensions/websocket/go.mod`

- [ ] **Step 1: Update go.mod**

Run:
```bash
cd extensions/websocket && go mod tidy
```

This should pick up the new Fiber dep from framework (via replace directive) and drop the `go-chi/chi` indirect dep.

- [ ] **Step 2: Rewrite handler.go**

Replace the entire `extensions/websocket/internal/wsutil/handler.go`:

```go
package wsutil

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	ws2 "github.com/franchb/htmgo/extensions/websocket/opts"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
	"net/http"
)

// WsHandler returns a Fiber handler that upgrades connections to WebSocket
// using gobwas/ws. We use Fiber's adaptor to get a net/http handler because
// gobwas/ws.UpgradeHTTP requires http.ResponseWriter + *http.Request.
func WsHandler(opts *ws2.ExtensionOpts) fiber.Handler {

	if opts.RoomName == nil {
		opts.RoomName = func(ctx *h.RequestContext) string {
			return "all"
		}
	}

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the RequestContext that was stored before the adaptor bridge.
		ccVal := r.Context().Value("htmgo.ws.reqctx")
		if ccVal == nil {
			http.Error(w, "websocket: missing request context", http.StatusInternalServerError)
			return
		}
		cc := ccVal.(*h.RequestContext)
		locator := cc.ServiceLocator()
		manager := service.Get[SocketManager](locator)

		sessionId := opts.SessionId(cc)

		if sessionId == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			slog.Info("failed to upgrade", slog.String("error", err.Error()))
			return
		}

		roomId := opts.RoomName(cc)
		done := make(chan bool, 1000)
		writer := make(WriterChan, 1000)

		ctx, cancel := context.WithCancel(context.Background())

		wg := sync.WaitGroup{}

		manager.Add(roomId, sessionId, writer, done)

		cleanupOnce := sync.Once{}
		cleanup := func() {
			cleanupOnce.Do(func() {
				cancel()
				conn.Close()
			})
		}

		wg.Add(1)
		go func() {
			defer manager.Disconnect(sessionId)
			defer wg.Done()
			defer cleanup()

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-done:
					fmt.Printf("closing connection: \n")
					return
				case <-ticker.C:
					manager.Ping(sessionId)
				case message := <-writer:
					err = wsutil.WriteServerMessage(conn, ws.OpText, []byte(message))
					if err != nil {
						return
					}
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cleanup()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					return
				}
				if op != ws.OpText {
					return
				}
				m := make(map[string]any)
				err = json.Unmarshal(msg, &m)
				if err != nil {
					return
				}
				manager.OnMessage(sessionId, m)
			}
		}()

		wg.Wait()
	})

	fiberHandler := adaptor.HTTPHandler(httpHandler)

	return func(c fiber.Ctx) error {
		// Get RequestContext from Fiber locals and pass it through to the
		// net/http handler via context.Value so gobwas/ws can do the upgrade.
		cc := h.GetRequestContext(c)
		if cc == nil {
			return c.Status(fiber.StatusInternalServerError).SendString("websocket: missing request context")
		}
		// Store the RequestContext in fiber locals under a key the adapted handler can read.
		c.Locals("htmgo.ws.reqctx", cc)
		return fiberHandler(c)
	}
}
```

Note: We keep `gobwas/ws` for the actual WebSocket upgrade because it's already in use and well-tested. We use `fiber/v3/middleware/adaptor` to bridge from Fiber's fasthttp to net/http for the gobwas upgrade. The RequestContext is passed through via `c.Locals()` → `r.Context().Value()`.

- [ ] **Step 3: Update init.go**

Replace line 30 in `extensions/websocket/init.go`:

```go
app.Router.Handle(opts.WsPath, wsutil.WsHttpHandler(&opts))
```

with:

```go
app.Router.Get(opts.WsPath, wsutil.WsHandler(&opts))
```

- [ ] **Step 4: Run go mod tidy again**

```bash
cd extensions/websocket && go get github.com/gofiber/fiber/v3@v3.10.0 && go mod tidy
```

- [ ] **Step 5: Verify compilation**

Run: `cd extensions/websocket && go build ./...`
Expected: Success.

- [ ] **Step 6: Commit**

```bash
git add extensions/websocket/
git commit -m "feat(websocket): migrate extension to Fiber v3 with adaptor bridge for gobwas/ws"
```

---

## Task 8: Example apps — chat (most complex)

**Files:**
- Modify: `examples/chat/main.go`
- Modify: `examples/chat/sse/handler.go`
- Modify: `examples/chat/pages/chat.$id.go`
- Modify: `examples/chat/go.mod`

- [ ] **Step 1: Rewrite examples/chat/sse/handler.go**

Replace the entire file:

```go
package sse

import (
	"bufio"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
)

func Handle() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Access-Control-Allow-Origin", "*")

		cc := h.GetRequestContext(c)
		locator := cc.ServiceLocator()
		manager := service.Get[SocketManager](locator)

		sessionId := c.Cookies("session_id")

		done := make(chan bool, 1000)
		writer := make(WriterChan, 1000)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer manager.Disconnect(sessionId)

			defer func() {
				for len(writer) > 0 {
					<-writer
				}
				for len(done) > 0 {
					<-done
				}
			}()

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			// We use a channel to receive write requests from the main streaming goroutine
			// This goroutine will be unblocked when the stream writer exits (client disconnect).
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if sessionId == "" {
				manager.writeCloseRaw(writer, "no session")
				return
			}

			roomId := c.Params("id")

			if roomId == "" {
				slog.Error("invalid room", slog.String("room_id", roomId))
				manager.writeCloseRaw(writer, "invalid room")
				return
			}

			manager.Add(roomId, sessionId, writer, done)
		}()

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			defer func() { done <- true }()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					manager.Ping(sessionId)
				case message := <-writer:
					_, err := fmt.Fprint(w, message)
					if err != nil {
						return
					}
					if err := w.Flush(); err != nil {
						return
					}
				}
			}
		})

		wg.Wait()
		return nil
	}
}
```

Note: The SSE streaming pattern changes significantly with Fiber. The `SetBodyStreamWriter` callback runs synchronously — when it returns, the connection is done. The goroutine coordination pattern adapts accordingly.

- [ ] **Step 2: Fix examples/chat/pages/chat.$id.go**

Replace all `chi.URLParam(ctx.Request, "id")` calls with `ctx.UrlParam("id")`:

Line 15: `roomId := chi.URLParam(ctx.Request, "id")` → `roomId := ctx.UrlParam("id")`
Line 68: `roomId := chi.URLParam(ctx.Request, "id")` → `roomId := ctx.UrlParam("id")`
Line 75: `roomId := chi.URLParam(ctx.Request, "id")` → `roomId := ctx.UrlParam("id")`

Remove the `"github.com/go-chi/chi/v5"` import.

- [ ] **Step 3: Rewrite examples/chat/main.go**

Replace:

```go
package main

import (
	"fmt"
	"io/fs"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"

	"chat/__htmgo"
	"chat/chat"
	"chat/internal/db"
	"chat/sse"
)

func main() {
	locator := service.NewLocator()

	service.Set[db.Queries](locator, service.Singleton, db.Provide)
	service.Set[sse.SocketManager](locator, service.Singleton, func() *sse.SocketManager {
		return sse.NewSocketManager()
	})

	chatManager := chat.NewManager(locator)
	go chatManager.StartListener()

	go func() {
		for {
			count := runtime.NumGoroutine()
			fmt.Printf("goroutines: %d\n", count)
			time.Sleep(10 * time.Second)
		}
	}()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")
			if err != nil {
				panic(err)
			}
			_ = sub

			app.Router.Use("/public", h.StaticCacheMiddleware())
			app.Router.Get("/public/*", static.New("./assets/dist"))
			app.Router.Get("/sse/chat/:id", sse.Handle())

			__htmgo.Register(app.Router)
		},
	})
}
```

- [ ] **Step 4: Update go.mod**

```bash
cd examples/chat && go get github.com/gofiber/fiber/v3@v3.10.0 && go mod tidy
```

This should remove the direct `github.com/go-chi/chi/v5` dependency.

- [ ] **Step 5: Verify compilation**

Run: `cd examples/chat && go build .`
Expected: Success.

- [ ] **Step 6: Commit**

```bash
git add examples/chat/
git commit -m "feat(examples/chat): migrate to Fiber v3"
```

---

## Task 9: Example apps — hackernews, simple-auth, todo-list, ws-example

**Files:**
- Modify: `examples/hackernews/main.go`
- Modify: `examples/simple-auth/main.go`
- Modify: `examples/todo-list/main.go`
- Modify: `examples/ws-example/main.go`

These all follow the same pattern. The changes are mechanical.

- [ ] **Step 1: Rewrite examples/hackernews/main.go**

```go
package main

import (
	"fmt"
	"io/fs"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"

	"hackernews/__htmgo"
)

func main() {
	locator := service.NewLocator()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")
			if err != nil {
				panic(err)
			}
			_ = sub

			app.Router.Get("/item", func(c fiber.Ctx) error {
				id := c.Query("id")
				return c.Redirect().Status(302).To(fmt.Sprintf("/?item=%s", id))
			})
			app.Router.Use("/public", h.StaticCacheMiddleware())
			app.Router.Get("/public/*", static.New("./assets/dist"))
			__htmgo.Register(app.Router)
		},
	})
}
```

- [ ] **Step 2: Rewrite examples/simple-auth/main.go**

```go
package main

import (
	"io/fs"

	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"

	"simpleauth/__htmgo"
	"simpleauth/internal/db"
)

func main() {
	locator := service.NewLocator()

	service.Set(locator, service.Singleton, func() *db.Queries {
		return db.Provide()
	})

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")
			if err != nil {
				panic(err)
			}
			_ = sub

			app.Router.Use("/public", h.StaticCacheMiddleware())
			app.Router.Get("/public/*", static.New("./assets/dist"))
			__htmgo.Register(app.Router)
		},
	})
}
```

- [ ] **Step 3: Rewrite examples/todo-list/main.go**

```go
package main

import (
	"embed"
	"io/fs"

	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
	_ "github.com/mattn/go-sqlite3"

	"todolist/__htmgo"
	"todolist/ent"
	"todolist/infrastructure/db"
)

//go:embed assets/dist/*
var StaticAssets embed.FS

func main() {
	locator := service.NewLocator()

	service.Set[ent.Client](locator, service.Singleton, func() *ent.Client {
		return db.Provide()
	})

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(StaticAssets, "assets/dist")
			if err != nil {
				panic(err)
			}
			_ = sub

			app.Router.Use("/public", h.StaticCacheMiddleware())
			app.Router.Get("/public/*", static.New("./assets/dist"))

			__htmgo.Register(app.Router)
		},
	})
}
```

- [ ] **Step 4: Rewrite examples/ws-example/main.go**

```go
package main

import (
	"io/fs"

	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/franchb/htmgo/extensions/websocket"
	ws2 "github.com/franchb/htmgo/extensions/websocket/opts"
	"github.com/franchb/htmgo/extensions/websocket/session"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
	"ws-example/__htmgo"
)

func main() {
	locator := service.NewLocator()

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {

			app.Use(func(ctx *h.RequestContext) {
				session.CreateSession(ctx)
			})

			websocket.EnableExtension(app, ws2.ExtensionOpts{
				WsPath: "/ws",
				RoomName: func(ctx *h.RequestContext) string {
					return "all"
				},
				SessionId: func(ctx *h.RequestContext) string {
					return ctx.QueryParam("sessionId")
				},
			})

			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")
			if err != nil {
				panic(err)
			}
			_ = sub

			app.Router.Get("/public/*", static.New("./assets/dist"))
			__htmgo.Register(app.Router)
		},
	})
}
```

- [ ] **Step 5: Run go mod tidy on all four examples**

```bash
for dir in examples/hackernews examples/simple-auth examples/todo-list examples/ws-example; do
  (cd "$dir" && go get github.com/gofiber/fiber/v3@v3.10.0 && go mod tidy)
done
```

- [ ] **Step 6: Commit**

```bash
git add examples/hackernews/ examples/simple-auth/ examples/todo-list/ examples/ws-example/
git commit -m "feat(examples): migrate hackernews, simple-auth, todo-list, ws-example to Fiber v3"
```

---

## Task 10: Example app — minimal-htmgo (standalone, no framework Start)

**Files:**
- Modify: `examples/minimal-htmgo/main.go`
- Modify: `examples/minimal-htmgo/render.go`
- Modify: `examples/minimal-htmgo/go.mod`

This example is unique — it creates a `chi.NewRouter()` directly without using `h.Start()`.

- [ ] **Step 1: Rewrite examples/minimal-htmgo/main.go**

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func main() {
	app := fiber.New()

	app.Get("/public/*", static.New("./public"))

	app.Get("/", func(c fiber.Ctx) error {
		return RenderPage(c, Index)
	})

	app.Get("/current-time", func(c fiber.Ctx) error {
		return RenderPartial(c, CurrentTime)
	})

	app.Listen(":3000")
}
```

- [ ] **Step 2: Rewrite examples/minimal-htmgo/render.go**

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/franchb/htmgo/framework/h"
)

func RenderToString(element *h.Element) string {
	return h.Render(element)
}

func RenderPage(c fiber.Ctx, page func(ctx *h.RequestContext) *h.Page) error {
	ctx := &h.RequestContext{
		Fiber: c,
	}
	return h.HtmlView(c, page(ctx))
}

func RenderPartial(c fiber.Ctx, partial func(ctx *h.RequestContext) *h.Partial) error {
	ctx := &h.RequestContext{
		Fiber: c,
	}
	return h.PartialView(c, partial(ctx))
}
```

- [ ] **Step 3: Update go.mod**

```bash
cd examples/minimal-htmgo && go get github.com/gofiber/fiber/v3@v3.10.0 && go mod tidy
```

This should remove the direct `github.com/go-chi/chi/v5` dependency.

- [ ] **Step 4: Verify compilation**

Run: `cd examples/minimal-htmgo && go build .`
Expected: Success.

- [ ] **Step 5: Commit**

```bash
git add examples/minimal-htmgo/
git commit -m "feat(examples/minimal): migrate minimal-htmgo to Fiber v3"
```

---

## Task 11: htmgo-site — sitemap, main.go, urlhelper

**Files:**
- Modify: `htmgo-site/internal/sitemap/generate.go`
- Modify: `htmgo-site/main.go`
- Modify: `htmgo-site/internal/urlhelper/ip.go`

- [ ] **Step 1: Rewrite sitemap/generate.go**

Replace the entire file:

```go
package sitemap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type URL struct {
	Loc        string  `xml:"loc"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float32 `xml:"priority,omitempty"`
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XmlNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

func NewSitemap(urls []URL) *URLSet {
	return &URLSet{
		XmlNS: "https://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
}

func serialize(sitemap *URLSet) ([]byte, error) {
	buffer := bytes.Buffer{}
	enc := xml.NewEncoder(&buffer)
	enc.Indent("", "  ")
	if err := enc.Encode(sitemap); err != nil {
		return make([]byte, 0), fmt.Errorf("could not encode sitemap: %w", err)
	}
	return buffer.Bytes(), nil
}

func Generate(router *fiber.App) ([]byte, error) {
	stack := router.Stack()
	urls := []URL{
		{
			Loc:        "/",
			Priority:   0.5,
			ChangeFreq: "weekly",
		},
		{
			Loc:        "/docs",
			Priority:   1.0,
			ChangeFreq: "daily",
		},
		{
			Loc:        "/examples",
			Priority:   0.7,
			ChangeFreq: "daily",
		},
		{
			Loc:        "/html-to-go",
			Priority:   0.5,
			ChangeFreq: "weekly",
		},
	}

	seen := make(map[string]bool)
	for _, routes := range stack {
		for _, route := range routes {
			if seen[route.Path] {
				continue
			}
			seen[route.Path] = true

			if strings.HasPrefix(route.Path, "/docs/") {
				urls = append(urls, URL{
					Loc:        route.Path,
					Priority:   1.0,
					ChangeFreq: "weekly",
				})
			}

			if strings.HasPrefix(route.Path, "/examples/") {
				urls = append(urls, URL{
					Loc:        route.Path,
					Priority:   0.7,
					ChangeFreq: "weekly",
				})
			}
		}
	}

	for i, url := range urls {
		urls[i].Loc = fmt.Sprintf("%s%s", "https://htmgo.dev", url.Loc)
	}

	sitemap := NewSitemap(urls)
	return serialize(sitemap)
}
```

Note: `fiber.App.Stack()` returns `[][]fiber.Route` grouped by HTTP method. The same path can appear multiple times (once per method). We use a `seen` map to deduplicate.

- [ ] **Step 2: Rewrite urlhelper/ip.go**

The htmgo-site `main.go` calls `urlhelper.GetClientIp(r)` where `r` is `ctx.Request`. Since `Request` is gone, change `GetClientIp` to accept `fiber.Ctx`:

```go
package urlhelper

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func GetClientIp(c fiber.Ctx) string {
	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	if ip := c.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	remote := c.IP()

	if strings.HasPrefix(remote, "[::1]") || remote == "127.0.0.1" {
		return "localhost"
	}

	return remote
}
```

- [ ] **Step 3: Rewrite htmgo-site/main.go**

```go
package main

import (
	"io/fs"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"

	"htmgo-site/__htmgo"
	"htmgo-site/internal/cache"
	"htmgo-site/internal/markdown"
	"htmgo-site/internal/sitemap"
	"htmgo-site/internal/urlhelper"
)

func main() {
	locator := service.NewLocator()
	staticAssets := GetStaticAssets()
	markdownAssets := GetMarkdownAssets()

	service.Set(locator, service.Singleton, markdown.NewRenderer)
	service.Set(locator, service.Singleton, cache.NewSimpleCache)

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {

			app.Use(func(ctx *h.RequestContext) {
				log.Printf("Method: %s, URL: %s, RemoteAddr: %s",
					ctx.Fiber.Method(), ctx.Fiber.OriginalURL(), urlhelper.GetClientIp(ctx.Fiber))
			})

			// Store embedded markdown in RequestContext kv for templates.
			// Replaces the removed UseWithContext which gave raw net/http access.
			app.Use(func(ctx *h.RequestContext) {
				ctx.Set("embeddedMarkdown", markdownAssets)
			})

			sub, err := fs.Sub(staticAssets, "assets/dist")
			if err != nil {
				panic(err)
			}
			_ = sub

			app.Router.Get("/sitemap.xml", func(c fiber.Ctx) error {
				s, err := sitemap.Generate(app.Router)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).SendString("failed to generate sitemap")
				}
				c.Set("Content-Type", "application/xml")
				return c.Send(s)
			})

			app.Router.Use("/public", h.StaticCacheMiddleware())
			app.Router.Get("/public/*", static.New("./assets/dist"))

			__htmgo.Register(app.Router)
		},
	})
}
```

- [ ] **Step 4: Update go.mod**

```bash
cd htmgo-site && go get github.com/gofiber/fiber/v3@v3.10.0 && go mod tidy
```

- [ ] **Step 5: Verify compilation**

Run: `cd htmgo-site && go build .`
Expected: Success.

- [ ] **Step 6: Commit**

```bash
git add htmgo-site/
git commit -m "feat(htmgo-site): migrate to Fiber v3 — sitemap, urlhelper, main"
```

---

## Task 12: go mod tidy across all modules and final verification

**Files:**
- All `go.mod` / `go.sum` files across the monorepo

- [ ] **Step 1: Run go mod tidy on every module**

```bash
for dir in framework cli/htmgo extensions/websocket framework-ui templates/starter examples/chat examples/hackernews examples/simple-auth examples/todo-list examples/ws-example examples/minimal-htmgo htmgo-site cli/htmgo/tasks/astgen/project-sample; do
  echo "=== $dir ===" && (cd "$dir" && go mod tidy) || echo "FAILED: $dir"
done
```

- [ ] **Step 2: Verify no chi references remain**

Run: `grep -r "go-chi/chi" --include="*.go" --include="go.mod" --include="go.sum" .`
Expected: No results (or only in `.claude/worktrees/` which are ignored).

- [ ] **Step 3: Run framework tests**

Run: `cd framework && go test ./... -count=1`
Expected: All tests pass.

- [ ] **Step 4: Run astgen tests**

Run: `cd cli/htmgo/tasks/astgen && go test -v -run TestAstGen -count=1 -timeout 30s`
Expected: Test passes.

- [ ] **Step 5: Verify all modules compile**

```bash
for dir in framework cli/htmgo extensions/websocket framework-ui htmgo-site examples/chat examples/hackernews examples/simple-auth examples/todo-list examples/ws-example examples/minimal-htmgo; do
  echo "=== $dir ===" && (cd "$dir" && go build ./...) || echo "FAILED: $dir"
done
```

Expected: All modules compile successfully.

- [ ] **Step 6: Commit any remaining go.mod/go.sum changes**

```bash
git add -A '*.mod' '*.sum'
git commit -m "chore: go mod tidy across all modules after Fiber v3 migration"
```

---

## Task 13: Verify htmgo-site builds and check for regressions

**Files:** None (verification only)

- [ ] **Step 1: Check for any remaining net/http references in framework/h/**

Run: `grep -n "net/http" framework/h/app.go framework/h/header.go framework/h/qs.go framework/h/livereload.go`
Expected: No results. All net/http usage in these files should be gone.

- [ ] **Step 2: Check framework-ui compiles**

Run: `cd framework-ui && go build ./...`
Expected: Success. framework-ui only uses RequestContext methods (not fields), so it should work.

- [ ] **Step 3: Verify the htmgo-site builds (if npm is available)**

```bash
cd htmgo-site && npm install 2>/dev/null; export PATH="$(go env GOPATH)/bin:$PATH" && go install ../cli/htmgo && htmgo build
```

Expected: Builds successfully. If npm or htmgo CLI has issues, note them but the Go compilation is the primary gate.

- [ ] **Step 4: Final commit message summarizing the migration**

If there are any uncommitted changes:

```bash
git add -A
git commit -m "chore: finalize chi-to-fiber-v3 migration — all modules verified"
```
