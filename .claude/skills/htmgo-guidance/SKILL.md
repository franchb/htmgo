---
name: htmgo-guidance
description: Use when writing Go code that uses github.com/franchb/htmgo/framework/v2, building pages, partials, or components with h.Div, h.Button, or h.Ren, wiring htmx attributes or Alpine directives, or answering questions about htmgo patterns and routing. Covers Fiber v3 integration, RequestContext, auto-routing, service locator, caching, and the Alpine compat extension.
---

# htmgo Guidance

*This guide follows the htmgo v2 source in this repository. Check `framework/go.mod` for the Go requirement and `framework/assets/js/package.json` for the bundled htmx version.*

## 1. What htmgo is

htmgo is a lightweight Go web framework for server-rendered sites. The HTML builder composes pages and fragments in Go. The bundled htmx client handles HTTP requests and DOM updates, and Alpine.js is available for client-side state.

- Use Go 1.27.0 or later and import the framework from `github.com/franchb/htmgo/framework/v2`. The HTTP layer uses Fiber v3.
- The client bundle contains htmx 4.0.2. The starter app serves it at `/public/htmgo.js` with the default asset prefix. Apps can load Alpine.js independently. The bundled `alpine-compat` extension checks for Alpine before running its swap hooks. The docs use Alpine.js 3.15.11.
- The CLI can compile Tailwind CSS when `tailwind: true` is set in `htmgo.yml`. Go asset embedding depends on the application's build and asset setup.

The Go packages used most often are `h` for rendering and routing, `hx` for htmx attributes, events, headers and triggers, and `ax` for Alpine directives. Import them below `github.com/franchb/htmgo/framework/v2`.

## 2. The `h.Ren` builder model

All renderable nodes implement the `h.Ren` interface.

```go
type Ren interface {
    Render(context *RenderContext)
}
```

Tag functions such as `h.Div` take variadic `h.Ren` children and return `*h.Element`.

```go
card := h.Div(
    h.Class("rounded-lg border p-4"),
    h.H2(
        h.Class("text-lg font-semibold"),
        h.Text("Hello"),
    ),
    h.P(h.Text("World")),
)
```

See `framework/h/tag.go` for the built-in tag helpers, including `Div`, `Button`, `TextArea`, `THead`, `TBody` and `H1` through `H6`. Some take a required argument before any children. For example, use `h.Input("search", h.Name("q"))` for a search input and `h.Script("/public/htmgo.js")` for a script tag.

Use `h.Tag(name, children...)` for custom elements.

```go
h.Tag("my-widget",
    h.Id("status-widget"),
    h.Text("Ready"),
)
```

Values returned by `h.Attribute(key, value)` implement `h.Ren`. Pass them as children and the renderer writes them on the opening tag.

```go
h.Div(
    h.Attribute("role", "dialog"),        // attribute
    h.Attribute("aria-label", "Confirm"), // attribute
    h.Class("p-4"),                       // attribute (convenience helper)
    h.Text("Body text here"),             // body child
)
```

### Attributes

- `h.Class("btn primary")` — sets `class`. Accepts multiple strings, joined with a space.
- `h.ClassX("base-classes", h.ClassMap{"active": isActive, "disabled": !canClick})` — base classes plus conditional extras from a map. Note the first arg is always the unconditional class string.
- `h.Id("my-id")` — sets `id`.
- `h.Type("submit")`, `h.Name("email")`, `h.Value("abc")`, `h.Placeholder("...")`, `h.Href("/x")`, `h.Src("/y")` — common attributes.
- `h.AttributePairs("data-foo", "bar", "data-baz", "qux")` — batch-set multiple attrs in one call (key, value, key, value, …).
- `h.Attributes(&h.AttributeMap{...})` — map-based bulk set (`AttributeMap` is `map[string]any`).

### Text and HTML

- `h.Text("literal string")` — HTML-escaped text node.
- `h.TextF("count: %d", n)` — printf-style formatting.

- `h.UnsafeRaw("<span>raw</span>")` — emits the string without escaping. Never pass user input. Use it only for trusted HTML.

### Conditional and repeated children

- `h.If(cond, ifTrue)` — renders `ifTrue` when `cond` is true and renders nothing otherwise.
- `h.IfElse(cond, ifTrue, ifFalse)` — selects between two values of the same type.
- `h.List(items, func(item T, index int) *Element { ... })` — renders a typed slice.
- `h.IterMap(m, func(key string, value T) *Element { ... })` — renders a `map[string]T` (order undefined).

### Card component

```go
func Card(title, body string, active bool) *h.Element {
    return h.Div(
        h.ClassX("rounded-lg border p-4",
            h.ClassMap{
                "border-blue-500 bg-blue-50": active,
                "border-gray-300":            !active,
            },
        ),
        h.H3(h.Class("font-semibold"), h.Text(title)),
        h.P(h.Class("mt-2 text-gray-700"), h.Text(body)),
    )
}
```

Compose components into a tree and pass the root to `h.NewPage` or `h.NewPartial`.

## 3. Pages, partials, and routes

A page handler returns `*h.Page` for normal navigation. Use `h.NewPage(root Ren)` with one document root. A partial handler returns `*h.Partial` for an htmx swap. Use `h.NewPartial(root *h.Element)`, or `h.NewPartialWithHeaders(headers, root)` to attach response headers.

The generator looks for exported functions with one named `*h.RequestContext` parameter and the corresponding return type. Here is a page handler for `pages/index.go`.

```go
func IndexPage(ctx *h.RequestContext) *h.Page {
    return h.NewPage(h.Html(
        h.Head(h.Title(h.Text("Home"))),
        h.Body(h.H1(h.Text("Hello"))),
    ))
}
```

### Generated routes

The CLI scans `pages/` for page handlers and registers them with `router.Get`. Page paths come from file paths.

| File | Route |
|---|---|
| `pages/index.go` | `/` |
| `pages/users/index.go` | `/users` |
| `pages/users/profile.go` | `/users/profile` |
| `pages/blog/$slug.go` | `/blog/:slug` |

Read the last example's parameter with `ctx.UrlParam("slug")`.

The generator scans the project for exported functions matching `func(ctx *h.RequestContext) *h.Partial` and registers each with `router.All`. For example, `PostIncrement` in the `partials/counter` package of module `example.com/htmgo-docs` is registered at `/example.com/htmgo-docs/partials/counter.PostIncrement`. The route uses the package's import path and the function name. Use `h.GetPartialPath(counter.PostIncrement)` to compute the URL, or pass the function directly to `h.GetPartial` or `h.PostPartial`.

Function-name prefixes such as `Get` and `Post` are descriptive. Guard the method before doing work in a state-changing partial. Check `ctx.IsHttpPost()`, `ctx.IsHttpPut()`, `ctx.IsHttpDelete()`, or compare `ctx.Fiber.Method()` with the required method. To return 405 from a generated partial, set the status and `Allow` header and return `h.EmptyPartial()`. The generated wrapper turns a nil partial into a 404. For explicit paths and method registration, register a Fiber route in `AppOpts.Register` and call `h.PartialView` to render the partial.

The CLI writes `__htmgo/pages-generated.go` and `__htmgo/partials-generated.go`. Call `__htmgo.Register(app.Router)` from `AppOpts.Register`. Regenerate after routing changes with `htmgo generate` or through the build or watch workflow. Keep the generated directory out of version control.

Exclude handlers by file glob in `htmgo.yml`.

```yaml
automatic_page_routing_ignore:
  - "**/_shared.go"
automatic_partial_routing_ignore:
  - "**/internal/*.go"
```

Use `ctx.IsHxRequest()` when a custom route serves both a document and a fragment. For separately generated handlers, point browser navigation at the page route and htmx requests at the partial route.

### Worked example — updating a counter

The examples use the module path `example.com/htmgo-docs`. In `pages/counter.go`, use the partial function to generate the request URL.

```go
package pages

import (
    "example.com/htmgo-docs/partials/counter"
    "github.com/franchb/htmgo/framework/v2/h"
    "github.com/franchb/htmgo/framework/v2/hx"
)

func CounterPage(ctx *h.RequestContext) *h.Page {
    return h.NewPage(h.Html(
        h.Head(h.Title(h.Text("Counter")), h.Script("/public/htmgo.js")),
        h.Body(
            counter.CountDisplay(),
            h.Button(
                h.Type("button"),
                h.PostPartialOnClick(counter.PostIncrement),
                h.HxTarget("#counter"),
                h.Attribute(hx.SwapAttr, hx.SwapTypeOuterHtml),
                h.Text("+"),
            ),
        ),
    ))
}
```

In `partials/counter/increment.go`, define the partial and a shared display component. This example shares one in-memory count across requests.

```go
package counter

import (
    "sync/atomic"

    "github.com/franchb/htmgo/framework/v2/h"
    "github.com/gofiber/fiber/v3"
)

var count atomic.Int64

func CountDisplay() *h.Element {
    return h.Div(h.Id("counter"), h.TextF("Count: %d", count.Load()))
}

func PostIncrement(ctx *h.RequestContext) *h.Partial {
    if !ctx.IsHttpPost() {
        ctx.Fiber.Set("Allow", fiber.MethodPost)
        ctx.Fiber.Status(fiber.StatusMethodNotAllowed)
        return h.EmptyPartial()
    }
    count.Add(1)
    return h.NewPartial(CountDisplay())
}
```

## 4. RequestContext

Route handlers receive a `*h.RequestContext` that wraps Fiber's `fiber.Ctx` with htmx-aware helpers.

### Form, query, and URL values

- `ctx.FormValue(key string) string` — POST form field (multipart or url-encoded).
- `ctx.QueryParam(key string) string` — `?foo=bar` → `"bar"`.
- `ctx.UrlParam(key string) string` — pattern param from `/users/:id`.

### htmx request headers

- `ctx.IsHxRequest() bool` — true if the `HX-Request: true` header is present.
- `ctx.HxSource() string` — raw `HX-Source` header (e.g. `button#submit-btn`).
- `ctx.HxSourceID() string` — just the id portion (e.g. `submit-btn`).
- `ctx.HxRequestType()` reports `"full"` when htmx targets the document body or selects content from a response. It reports `"partial"` for other htmx requests and returns an empty string when the header is absent.

### Navigation

Use `ctx.Redirect(url, code)` to write an HTTP redirect with a status such as `fiber.StatusTemporaryRedirect`. Return `h.RedirectPartial(url)` from a partial handler to have htmx navigate using the `HX-Redirect` response header.

Use `ctx.Fiber` to access Fiber directly for operations such as reading cookies with `ctx.Fiber.Cookies(...)`.

Middleware registered through `AppOpts.Register` can retrieve the request context from Fiber. Static asset requests skip context creation.

```go
func requestMarker(c fiber.Ctx) error {
    if ctx := h.GetRequestContext(c); ctx != nil {
        ctx.Set("entrypoint", "web")
    }
    return c.Next()
}
```

Retrieve request-scoped values with `ctx.Get(key)`.

### Setting response headers

Build headers with `h.NewHeaders` and pass them as the first argument to `h.NewPartialWithHeaders`.

```go
func GetStatus(ctx *h.RequestContext) *h.Partial {
    return h.NewPartialWithHeaders(
        h.NewHeaders(hx.RetargetHeader, "#status"),
        h.Div(h.Text("Ready")),
    )
}
```

- `h.ReplaceUrlHeader(url)` sets `HX-Replace-Url`.
- `h.PushUrlHeader(url)` sets `HX-Push-Url`.
- `h.CombineHeaders(headers ...*h.Headers)` merges header maps. Later maps override earlier values for the same key.
- `h.NewHeaders(key, value, ...)` accepts an even number of strings. An odd number produces an empty map.

## 5. htmx integration (`hx/` package + `h.Hx*` helpers)

### The `hx` package

The `hx` package provides attribute, header, event, and swap constants along with helpers for building triggers.

```go
import "github.com/franchb/htmgo/framework/v2/hx"

// Inside a component:
h.Attribute(hx.SwapAttr+":inherited", hx.SwapTypeInnerHtml)
```

Common constants are listed below. See `framework/hx/htmx.go` for the full set.

- **Attributes** include `GetAttr`, `PostAttr`, `TargetAttr`, `TriggerAttr`, `SwapAttr`, `IncludeAttr`, `IndicatorAttr`, `ConfirmAttr`, `ValsAttr`, `SyncAttr`, `BoostAttr`, `SwapOobAttr`, `SelectAttr`, `PushUrlAttr`, `ReplaceUrlAttr`, `EncodingAttr`, `HeadersAttr`, `ValidateAttr`, `PreserveAttr`, plus htmx 4 `ConfigAttr` (replaces `hx-request`), `StatusAttr` (per-status swap/target), `DisableAttr` (disables form elements during a request), and `IgnoreAttr` (skips htmx processing).
- **Events in htmx 4 colon form** include `AfterSwapEvent = "htmx:after:swap"`, `BeforeSwapEvent`, `AfterRequestEvent`, `BeforeRequestEvent`, `AfterOnLoadEvent = "htmx:after:init"`, `ConfigRequestEvent = "htmx:config:request"`, `ErrorEvent = "htmx:error"`, `AfterSettleEvent = "htmx:after:settle"`, `AbortEvent`.
- **Swap types** include `SwapTypeInnerHtml`, `SwapTypeOuterHtml`, `SwapTypeBeforeBegin`, `SwapTypeAfterBegin`, `SwapTypeBeforeEnd`, `SwapTypeAfterEnd`, `SwapTypeTextContent`, `SwapTypeDelete`, `SwapTypeNone`, `SwapTypeTrue`.
- **Headers** include `RequestHeader`, `RedirectHeader`, `RetargetHeader`, `ReswapHeader`, `ReselectHeader`, `TriggerHeader`, `PushUrlHeader`, `ReplaceUrlHeader`, `LocationHeader`, `CurrentUrlHeader`, `RefreshHeader`, plus htmx 4 `SourceHeader` (`HX-Source`) and `RequestTypeHeader` (`HX-Request-Type`, values `"full"`/`"partial"`).

### Setting htmx attributes directly

Use `h.Attribute` for individual attributes and the request helpers to combine a URL with a trigger.

**Individual attributes**

```go
h.Attribute(hx.GetAttr, "/search")   // hx-get="/search"
h.Attribute(hx.PostAttr, "/users")    // hx-post="/users"
h.Attribute(hx.SwapAttr, hx.SwapTypeInnerHtml)  // hx-swap="innerHTML"
```

**Request helpers**

```go
// Get(path, trigger...) → sets hx-get + hx-trigger
h.Button(h.Get("/search", "click"), h.Text("Search"))

// Post(url, trigger...) → sets hx-post + hx-trigger
h.Form(h.Post("/users/create", "submit"), h.Button(h.Type("submit"), h.Text("Create")))

// Partial-path variants (type-safe; avoid hard-coding URLs)
h.Button(h.GetPartial(search.GetSearch, "click"), h.Text("Search"))
h.Button(h.PostPartialOnClick(counter.PostIncrement), h.Text("Increment"))
```

**Target, include, confirmation, and indicator helpers**

```go
h.HxTarget("#results")           // hx-target="#results"
h.HxInclude("form[name=filter]") // hx-include="form[name=filter]"
h.HxConfirm("Really delete?")    // hx-confirm="Really delete?"
h.HxIndicator("#spinner")        // hx-indicator="#spinner"
```

### Explicit inheritance

With htmx 4 defaults, use the `:inherited` suffix to apply an attribute to descendants. For example, `h.HxTargetInherited("#results")` renders `hx-target:inherited="#results"`. The other helpers are `HxIncludeInherited`, `HxSwapInherited`, `HxBoostInherited`, `HxConfirmInherited`, `HxHeadersInherited`, `HxIndicatorInherited`, `HxSyncInherited`, `HxEncodingInherited`, and `HxValidateInherited`. Set global defaults with `htmx.config` or a `<meta name="htmx-config">` tag.

```go
h.Div(
    h.HxTargetInherited("#results"),
    h.Button(h.Get("/a", "click"), h.Text("A")),
    h.Button(h.Get("/b", "click"), h.Text("B")),
    h.Div(h.Id("results")),
)
```

### Triggers

`HxTriggerString` joins raw trigger strings with `", "`.

```go
h.HxTriggerString("keyup changed delay:300ms", "search")
// renders: hx-trigger="keyup changed delay:300ms, search"
```

Use `hx.OnEvent`, `hx.OnClick`, `hx.OnChange`, or `hx.OnLoad` to build structured triggers.

```go
h.HxTrigger(
    hx.OnClick(hx.OnceModifier{}),
    hx.OnEvent("keyup[keyCode==27]", hx.StringModifier("from:body")),
)
```

`h.HxTrigger` takes `...hx.TriggerEvent` and calls `hx.NewTrigger(opts...)` internally — each `TriggerEvent` is a plain struct with an event name and modifier list.

`hx.NewTrigger(opts ...TriggerEvent)` returns a `*Trigger` whose `ToString()` yields the `hx-trigger` value.

The modifiers are defined in `framework/hx/modifiers.go`.

- `hx.OnceModifier{}` — appends `"once"`
- `hx.Delay(n int)` — appends `"delay:<n>s"` (seconds)
- `hx.Throttle(n int)` — appends `"throttle:<n>s"` (seconds)
- `hx.StringModifier("...")` appends a raw modifier such as `from:body`, `target:button`, or `delay:300ms`. Put event filters directly after the event name, as in `keyup[keyCode==27]`.

`HxTriggerClick` sets a click trigger with optional modifiers.

```go
h.HxTriggerClick(hx.OnceModifier{})   // hx-trigger="click once"
// HxTriggerClick(opts ...hx.Modifier) is sugar for HxTrigger(hx.OnClick(opts...))
```

Some common trigger constants from `framework/hx/triggers.go` are shown below.

```go
hx.TriggerClick          // "onclick"
hx.TriggerClickOnce      // "onclick once"
hx.TriggerKeyUpEnter     // "keyup[keyCode==13]"
hx.TriggerBlur           // "blur"
```

Write polling triggers as `h.HxTriggerString("every 1s")`.

### Response headers

A partial can set multiple response headers in one call.

```go
h.NewPartialWithHeaders(
    h.NewHeaders(
        hx.TriggerHeader, "show-toast",     // fires a client event
        hx.RetargetHeader, "#status",      // changes swap target
        hx.ReswapHeader, "afterend",       // changes swap style
    ),
    h.Div(h.Text("Saved")),
)
```

Use these headers to control client behavior.

- `hx.RedirectHeader` — full-page redirect (htmx handles)
- `hx.LocationHeader` requests client-side navigation. Pass a path or a JSON object containing `path` and request options.
- `hx.PushUrlHeader` / `hx.ReplaceUrlHeader` — update browser URL without reload
- `hx.RefreshHeader` — force a full page refresh
- `hx.TriggerHeader` — dispatch a client-side event
- `hx.RetargetHeader` — change the swap target server-side
- `hx.ReswapHeader` — change the swap style server-side
- `hx.ReselectHeader` — change which element is selected from the response

### Worked example — search box with debounce

Put the component in `pages/search.go` and call `SearchBox()` in a page with the htmgo bundle loaded.

```go
package pages

import (
    "example.com/htmgo-docs/partials/search"
    "github.com/franchb/htmgo/framework/v2/h"
    "github.com/franchb/htmgo/framework/v2/hx"
)

func SearchBox() *h.Element {
    return h.Div(
        h.Input(
            "search",
            h.Name("q"),
            h.Placeholder("Search..."),
            h.GetPartial(search.GetSearch, "input changed delay:300ms"),
            h.HxTarget("#search-results"),
            h.Attribute(hx.SwapAttr, hx.SwapTypeInnerHtml),
        ),
        h.Div(h.Id("search-results")),
    )
}
```

In `partials/search/search.go`, read the input from the query string.

```go
package search

import "github.com/franchb/htmgo/framework/v2/h"

func GetSearch(ctx *h.RequestContext) *h.Partial {
    q := ctx.QueryParam("q")
    return h.NewPartial(h.P(h.TextF("Results for %s", q)))
}
```

## 6. Alpine integration (`ax/` package + bundled `alpine-compat`)

Alpine.js is **optional**. When you need small client-side state (open/closed, theme, density, copy-to-clipboard, command-palette modals), reach for Alpine. For server round-trips, reach for htmx. For imperative one-shots without state, reach for the lifecycle command DSL (see `h/lifecycle.go`).

### Setup

Add the Alpine CDN script to your layout’s `<head>`.

```go
h.Head(
    h.Tag("script",
        h.Attribute("src", "https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js"),
        h.Attribute("defer", ""),
    ),
    h.Script("/public/htmgo.js"),
)
```

Use `defer` so the browser parses the Alpine components before the script initializes.

Add the `[x-cloak]` rule to your stylesheet to hide cloaked elements until Alpine initializes.

```css
[x-cloak] { display: none !important; }
```

The bundled `alpine-compat` htmx extension registers when `htmgo.js` loads. Its swap hooks check for the Alpine APIs they use.

### The `ax/` package — constants + builders

`framework/ax/` mirrors the `hx/` shape: string constants for directive names, plus `h.Ren`-returning builders. Import `github.com/franchb/htmgo/framework/v2/ax`.

The directive name constants use `ax.Attribute`, an alias for `string`. Examples include `DataAttr`, `InitAttr`, `ShowAttr`, `BindAttr`, `OnAttr`, `TextAttr`, `HtmlAttr`, `ModelAttr`, `ModelableAttr`, `CloakAttr`, `RefAttr`, `IgnoreAttr`, `TeleportAttr`, `EffectAttr`, `IfAttr`, `ForAttr`, `IdAttr`, `TransitionAttr`.

Pass Alpine expressions to directive builders that take a value.

```go
ax.Data("{ open: false, count: 0 }")  // x-data="..."
ax.Init("count = 10")                  // x-init="..."
ax.Show("open")                        // x-show="..."
ax.Text("message")                     // x-text="..."
ax.Html("markup")                      // x-html="..."
ax.Model("query")                      // x-model="..."
ax.Effect("console.log(count)")        // x-effect="..."
ax.If("visible")                       // x-if="..." (template-only)
ax.For("item in items")                // x-for="..."
ax.Id("['tab']")                       // x-id="..."
ax.Ref("input")                        // x-ref="input"
ax.Teleport("body")                    // x-teleport="body"
ax.Modelable("value")                  // x-modelable="value"
```

These builders emit bare attributes that Alpine accepts as boolean directives.

```go
ax.Cloak()       // x-cloak
ax.Ignore()      // x-ignore
ax.Transition()  // x-transition
```

For richer transitions (`x-transition:enter`, `.opacity`, `.duration.500ms`), drop to `h.Attribute("x-transition:enter", ...)` directly.

`ax.Bind(attr, expr string) h.Ren` binds an HTML attribute to an expression. The shortcuts are `BindClass`, `BindStyle`, `BindHref`, `BindValue`, `BindDisabled`, `BindChecked`, and `BindId`. For example, `ax.BindClass("{ active: isActive }")` emits `x-bind:class="{ active: isActive }"`.

`ax.On(event, handler string, modifiers ...string) h.Ren` attaches an Alpine event handler. `OnClick`, `OnSubmit`, `OnInput`, `OnChange`, `OnFocus`, `OnBlur`, `OnKeydown`, and `OnKeyup` take a handler string and optional modifiers. The shortcuts `OnClickOutside(handler)`, `OnKeydownEscape(handler)`, and `OnKeydownEnter(handler)` emit `x-on:click.outside`, `x-on:keydown.escape`, and `x-on:keydown.enter` respectively.

```go
ax.On("click", "count++")                             // x-on:click="count++"
ax.On("keydown", "handle()", "meta", "k", "prevent")  // x-on:keydown.meta.k.prevent="handle()"
ax.OnClick("submit()", "prevent")                     // x-on:click.prevent="submit()"
ax.OnKeydownEscape("open = false")                    // x-on:keydown.escape="open = false"
```

The `x-model` shortcuts embed a modifier in the directive name.

```go
ax.ModelNumber("age")               // x-model.number="age"
ax.ModelLazy("title")               // x-model.lazy="title"
ax.ModelTrim("name")                // x-model.trim="name"
ax.ModelFill("notes")               // x-model.fill="notes"
ax.ModelBoolean("checked")          // x-model.boolean="checked"
ax.ModelDebounce("query", "500ms")  // x-model.debounce.500ms="query"
```

`ax.ModelDebounce(expr, duration string)` appends a duration string such as `"500ms"` or `"2s"` to the directive name.

### When to use `ax.*` vs htmx swaps vs lifecycle commands

| Use case | Tool |
|---|---|
| Pure-client state that never needs the server (theme toggle, popover open/close, keyboard overlay, copy-to-clipboard) | `ax.*` |
| Server round-trip required (load data, save form, refresh a list) | htmx — `h.Get(path, trigger...)`, `h.Post(url, trigger...)`, or `h.Attribute(hx.GetAttr, ...)` |
| Widget with BOTH client state AND server-loaded content (popover whose body is fetched lazily) | `ax.Data(...)` on the outer wrapper and `h.Get(...)` on inner elements. The wrapper retains its state when only its inner content is swapped. |
| Imperative one-shots without state (set a value, add/remove a class, fire an alert) | Lifecycle commands such as `h.OnClick(js.SetValue("…"))` |

### Worked example — popover with server-loaded content

```go
func Popover() *h.Element {
    return h.Div(
        h.Class("relative"),
        ax.Data("{ open: false }"),

        h.Button(
            h.Class("btn"),
            ax.OnClick("open = !open"),
            h.Text("Toggle"),
        ),

        h.Div(
            h.Class("popover"),
            ax.Show("open"),
            ax.Cloak(),
            ax.OnClickOutside("open = false"),
            ax.OnKeydownEscape("open = false"),
            h.Get("/popover/content", "intersect once"),
            h.Attribute(hx.SwapAttr, hx.SwapTypeInnerHtml),
            h.Text("Loading..."),
        ),
    )
}
```

Alpine manages the `open` flag. htmx loads the body on its first intersection and replaces the contents of the popover. The `x-data` root stays in place. Use a morph swap such as `innerMorph` or `outerMorph` when the swap should reconcile existing nodes. The bundled extension preserves Alpine state on matching nodes during a morph.

### Gotchas

- The compatibility extension uses Alpine v3 internal APIs (`closestDataStack`, `cloneNode`, `deferMutations`, `destroyTree`, `flushAndStopDeferringMutations`).
- **Alpine plugins load before Alpine itself.** Per upstream plugin docs — `@alpinejs/persist`, `@alpinejs/intersect`, `@alpinejs/focus`, etc. go in `<script>` tags ABOVE the main Alpine script.
- Keep the `x-data` root in place to retain its state. An `outerHTML` swap that replaces that node initializes state from the new markup.

## 7. Lifecycle & JS command DSL

The event helpers in `framework/h/lifecycle.go` generate JavaScript for actions such as showing an alert, setting a value, or toggling a CSS class.

### Event helpers

Pass a JS command to the event helper for the element.

```go
h.Button(
    h.OnClick(js.Alert("Hello!")),
    h.Text("Say hi"),
)
```

Use the named helpers for common events and `h.OnEvent` to specify another event.

- Use `h.OnClick(cmd...)`, `h.OnSubmit(cmd...)`, and `h.OnLoad(cmd...)` for DOM events.
- Use an inline event attribute with `h.OnEvent`, for example `h.OnEvent(hx.ChangeEvent, js.Alert("changed"))` or `h.OnEvent("onmouseover", cmd)`. For a custom DOM event handled by htmx, pass the full attribute name, such as `h.OnEvent("hx-on:show-toast", cmd)`.
- Use `h.HxBeforeRequest(cmd...)`, `h.HxAfterRequest(cmd...)`, `h.HxOnAfterSwap(cmd...)`, and `h.HxOnMutationError(cmd...)` for htmx request lifecycle events.
- For SSE events, use `h.HxBeforeSseMessage(cmd...)`, `h.HxAfterSseMessage(cmd...)`, `h.HxOnSseError(cmd...)`, `h.HxOnSseClose(cmd...)`, `h.HxOnSseConnecting(cmd...)`, and `h.HxOnSseOpen(cmd...)`.

DOM helpers emit inline attributes such as `onclick`, `onsubmit`, and `onload`. Helpers for `htmx:*` events use the `hx-on::` shorthand, such as `hx-on::after:swap`. The bundled htmgo extension invokes `onload` on ordinary elements when htmx initializes them, including after a swap.

### JS commands (`framework/js/`)

The `github.com/franchb/htmgo/framework/v2/js` package re-exports command constructors from `h`. Pass their results to event helpers that accept `...h.Command`.

```go
js.Alert("Message")                         // alert('...')
js.SetValue("new value")                    // (self || this).value = ...
js.SetText("hello")                         // (self || this).innerText = ...
js.SetInnerHtml(h.Div(h.Text("x")))         // (self || this).innerHTML = <rendered Ren>
js.SetOuterHtml(h.Span(h.Text("y")))        // (self || this).outerHTML = ...
js.AddClass("active")
js.RemoveClass("hidden")
js.ToggleClass("open")
js.AddAttribute("disabled", "true")
js.RemoveAttribute("disabled")
js.Remove()                                 // (self || this).remove()
js.ConsoleLog("debug")
js.PreventDefault()                         // event.preventDefault()
js.EvalJs("customFunction()")               // raw JS escape hatch
```

Most commands act on the element they're attached to (`this` / `self`). To target a different element, the `*OnParent`, `*OnChildren`, `*OnSibling` variants exist (`js.SetClassOnChildren`, `js.ToggleClassOnParent`, `js.EvalJsOnSibling`, etc.).

Use `js.EvalJs("window.location = '/home'")` or `js.EvalJs("self.focus()")` for navigation or focus inside a client-side handler.

Commands chain automatically — pass multiple to one handler:

```go
h.OnClick(
    js.AddClass("fade-out"),
    js.EvalJs("setTimeout(() => self.remove(), 300)"),
)
```

### Two command shapes

- `SimpleJsCommand` inlines a statement in the event attribute. Examples include `Alert`, `AddClass`, `SetValue`, `SetText`, and `Remove`.
- `ComplexJsCommand` emits a script with a generated `__eval_<id>` function. The event attribute calls it with the element and event. Examples include `EvalJs`, `SetTextOnChildren`, `EvalJsOnParent`, and `RunAfterTimeout`.

Both satisfy `h.Command`, an alias for `h.Ren`. Keep interpolated command arguments trusted because the constructors generate JavaScript source.

### Click handlers and requests

| Expression | Mechanism | When to use |
|---|---|---|
| `h.OnClick(js.Alert("hi"))` | htmgo-generated JS wired via `onclick` | Run a client-side command |
| `ax.OnClick("open = !open")` | `x-on:click="..."` evaluated by Alpine at runtime against the component's `x-data` scope | Mutate Alpine state |
| `h.Button(h.Get("/x", "click"))` | The click triggers an htmx GET for a fragment | Server round-trip |

Both handlers fire if you attach `h.OnClick` and `ax.OnClick` to the same element.

## 8. Service locator

Use `github.com/franchb/htmgo/framework/v2/service` for dependencies that handlers retrieve through `ctx.ServiceLocator()`. Create a locator at startup, register providers before serving requests, and pass it to `h.AppOpts.ServiceLocator`.

```go
type Settings struct {
    SiteName string
}

func start() {
    settings := &Settings{SiteName: "My app"}
    locator := service.NewLocator()
    service.Set[Settings](locator, service.Singleton, func() *Settings {
        return settings
    })

    h.Start(h.AppOpts{
        ServiceLocator: locator,
        Register: func(app *h.App) {
            __htmgo.Register(app.Router)
        },
    })
}
```

With `service.Set[T](locator, lifecycle, func() *T)`, the type argument is the underlying type. For example, register a database provider with `service.Set[sql.DB](...)` and return a `*sql.DB`. Resolve the provider from a handler with `service.Get[sql.DB](ctx.ServiceLocator())`. An unregistered type causes a fatal log at lookup.

- `service.Singleton` caches the first result published for a type. Concurrent first lookups can run the provider more than once because provider construction happens outside the lock. For resources that must be created exactly once, construct them at startup as in the example, resolve them before serving requests, or protect initialization with `sync.OnceValue`.
- `service.Transient` runs the provider on each lookup. Providers can resolve other registered services. Use `ctx.Set` and `ctx.Get` for state scoped to one request.

## 9. Caching

### High-level helpers (`h.Cached` family)

The cache helpers in `framework/h/cache.go` wrap `*h.Element` nodes and return constructors you call when building the render tree.

```go
// Package-level: cache once, globally.
var CachedHeader = h.Cached(5*time.Minute, func() *h.Element {
    return h.Header(h.H1(h.Text("Welcome")))
})
// Inside a page/partial:
return h.NewPage(h.Div(CachedHeader()))
```

`h.Cached(duration, cb, opts...)` returns a constructor of type `func() *h.Element`. Reuse the same constructor across requests. It renders on the first cache miss and shares that output until the entry expires.

For user-scoped or parameterized caching, use the `PerKey` variants. Their callback returns `(key, renderFunc)`.

```go
var UserProfile = h.CachedPerKeyT(15*time.Minute,
    func(u User) (int, h.GetElementFunc) {
        return u.ID, func() *h.Element { return renderProfile(u) }
    })

// In a request handler, call UserProfile(currentUser) to get a *h.Element.
```

`CachedT`, `CachedT2`, `CachedT3`, and `CachedT4` accept arguments and share one cache entry per constructor. Use the `CachedPerKey` family to select a separate entry for each key.

### Swapping the underlying store

The default is a fresh `cache.TTLStore` per cached component. Override a component with `h.WithCacheStore(store)` or replace `h.DefaultCacheProvider` before constructing cached components. Give each component a separate store or key namespace to avoid collisions.

```go
lru := cache.NewLRUStore[any, string](1000)
UserProfile := h.CachedPerKeyT(15*time.Minute, profileCb, h.WithCacheStore(lru))

// Or app-wide:
h.DefaultCacheProvider = func() cache.Store[any, string] {
    return cache.NewLRUStore[any, string](5000)
}
```

### Low-level pluggable stores

`framework/h/cache/interface.go` defines the store interface.

```go
type Store[K comparable, V any] interface {
    Set(key K, value V, ttl time.Duration)
    Get(key K) (V, bool)
    GetOrCompute(key K, compute func() V, ttl time.Duration) V
    Delete(key K)
    Purge()
    Close()
}
```

The caller supplies TTL for each entry. Available constructors include `NewTTLStore`, `NewTTLStoreWithInterval`, `NewTTLStoreWithMaxSize`, and `NewLRUStore` in the `cache` package. See `ExampleDistributedCacheAdapter` in `framework/h/cache/example_test.go` for an adapter pattern.

### Choosing what to cache

Cache expensive shared fragments such as navigation or results backed by a rate-limited service. Use a user-scoped key that covers all inputs to the rendered output when caching personalized content. Choose a TTL that meets the content’s freshness requirements.

## 10. Project configuration & CLI

### `htmgo.yaml` / `htmgo.yml` / `_htmgo.yaml` / `_htmgo.yml`

Keep the configuration at the app root. The loader in `framework/config/project.go` uses the first available file in the order shown above.

```yaml
tailwind: true                    # run Tailwind CSS compilation
tailwind_version: "v4.3.5"        # optional; use a full release tag or "latest"
watch_ignore: ["**/node_modules/**", "**/.git/**", "assets/dist/**"]  # globs watcher ignores
watch_files:  ["**/*.go", "**/*.css", "**/*.md"]     # globs watcher rebuilds on
automatic_page_routing_ignore:    ["_shared.go"]     # skip for page auto-routing
automatic_partial_routing_ignore: ["internal/*.go"]  # skip for partial auto-routing
public_asset_path: "/public"      # URL prefix for static assets
```

Without a config file, Tailwind is enabled by default. In a config file, set `tailwind: true` explicitly to enable it. If `tailwind_version` is omitted, the CLI selects `v3.4.16` for `assets/css/input.css` containing `@tailwind` directives and `latest` otherwise. Tailwind v3 also requires `tailwind.config.js` at the project root. The default asset prefix is `/public`. The config loader prefixes page ignore globs with `pages/` and partial ignore globs with `partials/`.

### CLI commands

Install with `go install github.com/franchb/htmgo/cli/htmgo/v2@latest`. Run an individual command with `go run github.com/franchb/htmgo/cli/htmgo/v2@latest <subcommand>`, or use `go install .` from `cli/htmgo` in this checkout.

The subcommands are defined in `cli/htmgo/runner.go`.

- `htmgo template [name]` — scaffold a new app from a starter template into `./<name>`.
- `htmgo setup` — run the project setup task in the current app.
- `htmgo build` generates routes, builds assets, and writes the production binary to `dist/`.
- `htmgo watch` builds assets, regenerates routes and Ent code when configured, and runs the live-reload development server.
- `htmgo run` — build and run without the watcher.
- `htmgo generate` — regenerate `__htmgo/*-generated.go` + run ent codegen.
- `htmgo css` — Tailwind/CSS build only.
- `htmgo schema` — prompts for an entity name and scaffolds an ent schema.
- `htmgo format <file|.>` — format a file or the whole working directory.
- `htmgo version` — print the CLI version.

Starter apps wrap routine commands in a `Taskfile.yml`. See `examples/todo-list/Taskfile.yml` for `task watch`, `task build`, and `task run`.

## 11. Troubleshooting

- If a descendant's request swaps into the wrong element, check whether the intended target on its ancestor uses `h.HxTargetInherited(...)`.
- If a route is missing, check that the handler is exported and uses the named `*h.RequestContext` parameter and expected return type. Regenerate routes and check the ignore globs.
- For a generated route, return `h.NewPage(root)` or `h.NewPartial(root)` so the route wrapper handles the response. `h.Render(root)` is available when you need an HTML string.
- A strict CSP can block the inline handlers emitted by `h.OnClick` and the evaluation used by htmx `hx-on`. `AppOpts.ContentSecurityPolicy` adds script nonces to scripts that the framework renders. For a strict policy, use event listeners registered from an allowed script and use declarative request attributes for htmx requests.

## 12. Upgrading to v2

- Update the framework and CLI to their `/v2` module paths and build with Go 1.27.0 or later. Use `htmgo generate` followed by `go build ./...` from the app root.
- Port middleware from chi and `net/http` to `func(fiber.Ctx) error`. In htmgo handlers, the Fiber context is available as `ctx.Fiber`.
- Apply htmx 4's explicit inheritance where needed. Remove `hx-ext` and calls to `h.HxExtension`. Bundled extensions register themselves. Use `hx-ignore` to skip htmx processing and `hx-disable` to disable form controls during a request. Use the colon-form htmx events and read `detail.ctx` when porting custom extensions.
- Load Alpine v3 separately when the app uses it and add the `[x-cloak]` rule from the Alpine setup section. The htmgo v2 bundle includes `alpine-compat`.

See [CHANGELOG.md](../../../CHANGELOG.md) for the v2 module migration and the htmx 4, Fiber v3, and Alpine changes.
