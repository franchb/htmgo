# Chi to Fiber v3 Migration Design

**Date:** 2026-04-16
**Goal:** Replace go-chi/chi/v5 with gofiber/fiber/v3 (v3.10.0) as the HTTP router for the htmgo framework. Motivation is performance — Fiber's fasthttp foundation provides higher throughput than net/http.
**Scope:** Full migration — framework core, code generator, websocket extension, all example apps, htmgo-site. No compatibility layer.
**API breakage:** Full break accepted. This is a major version change for htmgo.

---

## 1. RequestContext Transformation

### Current structure

```go
type RequestContext struct {
    Request           *http.Request
    Response          http.ResponseWriter
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

### New structure

```go
type RequestContext struct {
    Fiber             fiber.Ctx              // replaces Request + Response
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

The `Fiber` field is public so users can access the full Fiber context for operations not covered by RequestContext helpers (e.g., `ctx.Fiber.Body()`, `ctx.Fiber.SendStream()`).

### Method mapping

| Method | Old implementation | New implementation |
|--------|-------------------|-------------------|
| `UrlParam(key)` | `chi.URLParam(c.Request, key)` | `c.Fiber.Params(key)` |
| `QueryParam(key)` | `c.Request.URL.Query().Get(key)` | `c.Fiber.Query(key)` |
| `FormValue(key)` | `c.Request.FormValue(key)` | `c.Fiber.FormValue(key)` |
| `Header(key)` | `c.Request.Header.Get(key)` | `c.Fiber.Get(key)` |
| `IsHttpPost()` | `c.Request.Method == "POST"` | `c.Fiber.Method() == "POST"` |
| `IsHttpGet()` | `c.Request.Method == "GET"` | `c.Fiber.Method() == "GET"` |
| `IsHttpPut()` | `c.Request.Method == "PUT"` | `c.Fiber.Method() == "PUT"` |
| `IsHttpDelete()` | `c.Request.Method == "DELETE"` | `c.Fiber.Method() == "DELETE"` |
| `Redirect(path, code)` | `w.Header().Set("Location", path); w.WriteHeader(code)` | `c.Fiber.Redirect().Status(code).To(path)` |
| `SetCookie(cookie)` | `http.SetCookie(w, cookie)` — takes `*http.Cookie` | `c.Fiber.Cookie(cookie)` — takes `*fiber.Cookie` |
| `Set(key, val)` | Internal `kv` map | Internal `kv` map (unchanged) |
| `Get(key)` | Internal `kv` map | Internal `kv` map (unchanged) |
| `ServiceLocator()` | Returns `c.locator` | Returns `c.locator` (unchanged) |

### Cookie conversion

`SetCookie` currently takes `*http.Cookie`. Two options:
- Change the signature to take `*fiber.Cookie` (breaking but clean)
- Accept `*http.Cookie` and convert internally (easier migration for apps)

Decision: Change signature to `*fiber.Cookie`. Full break is accepted, and fiber.Cookie has the same essential fields (Name, Value, Path, Domain, Expires, Secure, HTTPOnly, SameSite).

### GetRequestContext

```go
// Old
func GetRequestContext(r *http.Request) *RequestContext {
    return r.Context().Value(requestContextKey).(*RequestContext)
}

// New
func GetRequestContext(c fiber.Ctx) *RequestContext {
    val := c.Locals(requestContextKey)
    if val == nil {
        return nil
    }
    return val.(*RequestContext)
}
```

### populateHxFields

```go
// Old — reads from http.Request.Header
cc.isBoosted = cc.Request.Header.Get(hx.BoostedHeader) == "true"

// New — reads from fiber.Ctx
cc.isBoosted = cc.Fiber.Get(hx.BoostedHeader) == "true"
```

All htmx header reads use `cc.Fiber.Get(headerName)` instead of `cc.Request.Header.Get(headerName)`.

### Removed fields and methods

- `Request *http.Request` — removed. Use `ctx.Fiber` for raw access.
- `Response http.ResponseWriter` — removed. Use `ctx.Fiber` for raw access.
- `RequestContextKey` (legacy string constant) — removed. Was deprecated.

---

## 2. App and Router

### Current

```go
type App struct {
    Opts   AppOpts
    Router *chi.Mux
}
```

### New

```go
type App struct {
    Opts   AppOpts
    Router *fiber.App
}
```

### Start()

```go
func Start(opts AppOpts) {
    app := fiber.New(fiber.Config{
        // Sensible defaults; can be extended via AppOpts later
    })
    instance := App{
        Opts:   opts,
        Router: app,
    }
    instance.start()
}
```

### Middleware

The RequestContext-creation middleware becomes:

```go
app.Router.Use(func(c fiber.Ctx) error {
    // Skip static file requests
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
```

### app.Use()

```go
// Old
func (app *App) Use(h func(ctx *RequestContext)) {
    app.Router.Use(func(handler http.Handler) http.Handler { ... })
}

// New
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

### app.UseWithContext() — REMOVED

This method exposed `http.ResponseWriter` and `*http.Request`. It is removed entirely. Users should use `app.Use()` with RequestContext or register Fiber middleware directly on `app.Router`.

### StaticCacheMiddleware

Rewritten as a Fiber middleware:

```go
func StaticCacheMiddleware() fiber.Handler {
    return func(c fiber.Ctx) error {
        err := c.Next()
        status := c.Response().StatusCode()
        if status == fiber.StatusOK || status == fiber.StatusNotModified {
            cacheHeader := "public, max-age=3600"
            if c.Request().URI().QueryString() != nil && len(c.Request().URI().QueryString()) > 0 {
                cacheHeader = "public, max-age=31536000, immutable"
            }
            c.Set("Cache-Control", cacheHeader)
        }
        return err
    }
}
```

Note: The middleware now runs after `c.Next()` to check the response status code before setting headers.

### Server startup

```go
// Old
http.ListenAndServe(port, app.Router)

// New
app.Router.Listen(port)
```

The port-conflict retry logic stays the same conceptually — kill the process holding the port, then `app.Router.Listen(port)` again.

### View helpers

```go
// Old
func writeHtml(w http.ResponseWriter, element Ren) error {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    _, err := fmt.Fprint(w, Render(element, WithDocType()))
    return err
}

func HtmlView(w http.ResponseWriter, page *Page) error { ... }
func PartialView(w http.ResponseWriter, partial *Partial) error { ... }
func PartialViewWithHeaders(w http.ResponseWriter, headers *Headers, partial *Partial) error { ... }

// New
func writeHtml(c fiber.Ctx, element Ren) error {
    c.Set("Content-Type", "text/html; charset=utf-8")
    return c.SendString(Render(element, WithDocType()))
}

func HtmlView(c fiber.Ctx, page *Page) error { ... }
func PartialView(c fiber.Ctx, partial *Partial) error { ... }
func PartialViewWithHeaders(c fiber.Ctx, headers *Headers, partial *Partial) error { ... }
```

All view helper functions take `fiber.Ctx` instead of `http.ResponseWriter`.

---

## 3. Code Generator (AST Gen)

File: `cli/htmgo/tasks/astgen/entry.go`

### Constants

```go
// Old
const ChiModuleName = "github.com/go-chi/chi/v5"

// New
const FiberModuleName = "github.com/gofiber/fiber/v3"
```

### Route pattern format

Fiber uses `:param` syntax (not `{param}`). The `formatRoute` function simplifies:

```go
// Old: $id → :id → {id}
parts[i] = fmt.Sprintf("{%s}", part[1:])

// New: $id → :id (stop here)
// The existing $→: replacement already produces what Fiber needs.
// Remove the {param} wrapping step.
```

### Generated RegisterPartials

```go
// Old template
func RegisterPartials(router *chi.Mux) {
    router.Handle("/path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cc := h.GetRequestContext(r)
        partial := pkg.Func(cc)
        if partial == nil {
            w.WriteHeader(404)
            return
        }
        h.PartialView(w, partial)
    }))
}

// New template
func RegisterPartials(router *fiber.App) {
    router.All("/path", func(c fiber.Ctx) error {
        cc := h.GetRequestContext(c)
        partial := pkg.Func(cc)
        if partial == nil {
            return c.SendStatus(404)
        }
        return h.PartialView(c, partial)
    })
}
```

### Generated RegisterPages

```go
// Old template
func RegisterPages(router *chi.Mux) {
    router.Get("/path", func(writer http.ResponseWriter, request *http.Request) {
        cc := h.GetRequestContext(request)
        h.HtmlView(writer, pkg.Func(cc))
    })
}

// New template
func RegisterPages(router *fiber.App) {
    router.Get("/path", func(c fiber.Ctx) error {
        cc := h.GetRequestContext(c)
        return h.HtmlView(c, pkg.Func(cc))
    })
}
```

### Generated Register (setup-generated.go)

```go
// Old
import "github.com/go-chi/chi/v5"
func Register(r *chi.Mux) { ... }

// New
import "github.com/gofiber/fiber/v3"
func Register(r *fiber.App) { ... }
```

### Import changes in generated files

- Remove: `"net/http"` (no longer needed in generated partial handlers)
- Remove: `"github.com/go-chi/chi/v5"`
- Add: `"github.com/gofiber/fiber/v3"`
- Keep: `"github.com/franchb/htmgo/framework/h"`

---

## 4. WebSocket Extension

File: `extensions/websocket/init.go`

### Current

```go
func EnableExtension(app *h.App, opts opts.ExtensionOpts) {
    // ... service setup ...
    app.Router.Handle(opts.WsPath, wsutil.WsHttpHandler(&opts))
}
```

### New

```go
import "github.com/gofiber/contrib/v3/websocket"

func EnableExtension(app *h.App, opts opts.ExtensionOpts) {
    // ... service setup ...
    app.Router.Get(opts.WsPath, wsutil.WsHandler(&opts))
}
```

### WsHttpHandler → WsHandler

The handler function in `extensions/websocket/internal/wsutil/handler.go` changes:
- Returns `fiber.Handler` instead of `http.HandlerFunc`
- Uses Fiber's contrib websocket package for the upgrade
- Retrieves `*h.RequestContext` from `c.Locals()` instead of `r.Context().Value()`
- The `gorilla/websocket.Upgrader` is replaced by Fiber's `websocket.New()` middleware

### go.mod changes

- `extensions/websocket/go.mod`: add `github.com/gofiber/fiber/v3 v3.10.0` and `github.com/gofiber/contrib/websocket/v3` (latest compatible with fiber v3.10.0)
- Remove `github.com/go-chi/chi/v5` (currently indirect, will be removed when framework drops it)

---

## 5. Sitemap Generator (htmgo-site)

File: `htmgo-site/internal/sitemap/generate.go`

### Current

```go
func Generate(router *chi.Mux) ([]byte, error) {
    routes := router.Routes()
    // ... iterate routes, check route.Pattern
}
```

### New

```go
func Generate(router *fiber.App) ([]byte, error) {
    stack := router.Stack()
    // stack is [][]fiber.Route — one slice per HTTP method
    // Iterate and extract route.Path for pattern matching
    for _, routes := range stack {
        for _, route := range routes {
            // route.Path contains the pattern (e.g., "/docs/:slug")
        }
    }
}
```

Fiber's `Stack()` returns routes grouped by HTTP method index. Each `fiber.Route` has `Path`, `Method`, `Name` fields.

---

## 6. Example Apps

All examples follow the same pattern. Changes are mechanical:

### Static file serving

```go
// Old
app.Router.Handle("/public/*", h.StaticCacheMiddleware(
    http.StripPrefix("/public/", http.FileServer(http.Dir("assets/dist"))),
))

// New — using Fiber's static middleware with cache middleware
app.Router.Use("/public", h.StaticCacheMiddleware())
app.Router.Get("/public/*", static.New("assets/dist"))
```

### Auto-generated route registration

```go
// Old
__htmgo.Register(app.Router)  // Register takes *chi.Mux

// New
__htmgo.Register(app.Router)  // Register takes *fiber.App — same call, different type
```

### Chat app — direct chi.URLParam calls

```go
// Old
id := chi.URLParam(ctx.Request, "id")

// New
id := ctx.UrlParam("id")  // or ctx.Fiber.Params("id")
```

### Chat app — SSE handler

The SSE handler in `examples/chat/sse/handler.go` uses `http.ResponseWriter` and `http.Flusher` directly. This is rewritten to use Fiber streaming.

### htmgo-site main.go

```go
// Old
app.Use(func(ctx *h.RequestContext) { ... })
app.UseWithContext(func(w http.ResponseWriter, r *http.Request, context map[string]any) { ... })

// New — UseWithContext removed, convert to app.Use()
app.Use(func(ctx *h.RequestContext) { ... })
```

---

## 7. Live Reload (SSE)

File: `framework/h/livereload.go`

### Current

Uses `http.Flusher` interface:

```go
func sseHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    flusher, ok := w.(http.Flusher)
    // ... write events in loop, flusher.Flush()
}

func (app *App) AddLiveReloadHandler(path string) {
    app.Router.Get(path, sseHandler)
}
```

### New

Fiber supports streaming via the response writer:

```go
func sseHandler(c fiber.Ctx) error {
    c.Set("Content-Type", "text/event-stream")
    c.Set("Cache-Control", "no-cache")
    c.Set("Connection", "keep-alive")
    c.Set("Access-Control-Allow-Origin", "*")

    c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
        for {
            fmt.Fprintf(w, "data: %s\n\n", Version)
            w.Flush()
            time.Sleep(500 * time.Millisecond)
        }
    })
    return nil
}

func (app *App) AddLiveReloadHandler(path string) {
    app.Router.Get(path, sseHandler)
}
```

---

## 8. go.mod Changes

### framework/go.mod

- Remove: `github.com/go-chi/chi/v5 v5.2.5`
- Add: `github.com/gofiber/fiber/v3 v3.10.0`

### extensions/websocket/go.mod

- Add: `github.com/gofiber/contrib/v3/websocket` (version TBD — latest compatible)
- Remove: `github.com/go-chi/chi/v5` (indirect)

### cli/htmgo/go.mod

- Remove: `github.com/go-chi/chi/v5` (indirect, pulled from framework)
- Will get `fiber` as new indirect dependency from framework

### All example go.mod files and htmgo-site/go.mod

- Run `go mod tidy` after framework changes to pick up fiber transitively

---

## 9. Testing Impact

### Unchanged test areas

- `framework/h/render_test.go` — HTML rendering tests. No HTTP dependency.
- `framework/h/cache/` — cache store tests. No HTTP dependency.
- `framework/service/` — DI tests. No HTTP dependency.

### Tests that need updates

- Any test creating `RequestContext` with `http.Request`/`http.ResponseWriter` — use Fiber's `app.Test()` helper or create a Fiber test context.
- `cli/htmgo/tasks/astgen/` tests — expected generated output changes (fiber imports, handler signatures, route patterns).

### New test considerations

- Fiber provides `app.Test(req)` for integration testing — create a `*http.Request`, pass it to the Fiber app, get back `*http.Response`. This can be used to test the full request pipeline without starting a server.

---

## 10. Files Changed Summary

**Framework core (must change):**
- `framework/h/app.go` — RequestContext, App, Start, middleware, view helpers
- `framework/h/livereload.go` — SSE handler
- `framework/go.mod` — dependency swap

**Code generator (must change):**
- `cli/htmgo/tasks/astgen/entry.go` — templates, constants, formatRoute
- `cli/htmgo/tasks/astgen/*_test.go` — expected output

**WebSocket extension (must change):**
- `extensions/websocket/init.go` — handler registration
- `extensions/websocket/internal/wsutil/handler.go` — handler rewrite
- `extensions/websocket/go.mod` — add fiber contrib dep

**htmgo-site (must change):**
- `htmgo-site/internal/sitemap/generate.go` — route introspection
- `htmgo-site/main.go` — middleware registration

**Example apps (must change):**
- `examples/chat/main.go`
- `examples/chat/pages/chat.$id.go` — remove direct chi.URLParam
- `examples/chat/sse/handler.go` — SSE handler rewrite
- `examples/hackernews/main.go`
- `examples/simple-auth/main.go`
- `examples/todo-list/main.go`
- `examples/ws-example/main.go`
- `examples/minimal-htmgo/main.go`

**Unchanged:**
- `framework/h/render.go`, `tag.go`, `lifecycle.go`, `cache.go`, `page.go`, `partial.go`
- `framework/hx/` — htmx constants
- `framework/service/` — DI
- `framework/config/` — config
- `framework/assets/` — static assets / JS extensions
- `framework-ui/` — UI components (depends on h.RequestContext but only uses methods, not fields)

---

## 11. Migration Order

1. Framework core (`framework/h/app.go`, `livereload.go`, `go.mod`)
2. Code generator (`cli/htmgo/tasks/astgen/`)
3. WebSocket extension
4. Example apps (mechanical, can be parallelized)
5. htmgo-site
6. Run `go mod tidy` across all modules
7. Run tests: `cd framework && go test ./...` and `cd cli/htmgo/tasks/astgen && go test ./...`
