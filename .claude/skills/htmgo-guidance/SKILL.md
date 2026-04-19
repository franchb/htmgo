---
name: htmgo-guidance
description: Use when writing Go code that uses the htmgo framework (github.com/franchb/htmgo), building pages, partials, or components with h.Div/h.Button/h.Ren, wiring hx/ htmx attributes or ax/ Alpine directives, or answering questions about htmgo patterns, routing, and best practices. Covers Fiber v3 integration, RequestContext, auto-routing, service locator, caching, and the Alpine compat extension.
---

# htmgo Guidance

*Last reviewed against: htmgo `franchb/htmgo` v1.2.0-beta.2 (htmx4-migration branch), htmx 4.0.0-beta2, Alpine.js 3.15.11. If framework signatures have changed since, check the source files under `framework/` for ground truth.*

## 1. What htmgo is

htmgo is a lightweight Go web framework for building interactive SSR websites without a JavaScript build step. You write Go code using a builder API (`h.Div(...)`, `h.Button(...)`) that produces HTML; htmx handles server round-trips; Alpine.js is optional for pure-client state.

**Key facts:**
- **Import path:** `github.com/franchb/htmgo/framework` (this fork) — upstream is `maddalax/htmgo`; APIs are close but this fork has diverged (htmx 4, Fiber v3, `framework/ax/` Alpine helpers).
- **Go version:** 1.23+.
- **HTTP layer:** Fiber v3 (`github.com/gofiber/fiber/v3`).
- **htmx version:** 4.0.0-beta2, pinned in `framework/assets/js/package.json`. Bundled into `/public/htmgo.js`.
- **Alpine.js (optional):** consumers load Alpine 3.15.11 themselves; the `alpine-compat` htmx extension is pre-bundled in `htmgo.js` and auto-gates on `window.Alpine` presence.
- **Styling:** Tailwind CSS optional (set `tailwind: true` in `htmgo.yml`).
- **Output:** single deployable Go binary with assets embedded.

**The three main Go packages you'll touch as a consumer:**
- `framework/h` — the HTML builder, rendering, routing primitives, request context, lifecycle events, caching helpers, JS command DSL.
- `framework/hx` — constants for htmx attributes, events, headers, swap types.
- `framework/ax` — constants + builder helpers for Alpine.js directives (this fork's addition; mirrors `hx/` shape).

The rest of this skill walks each of these plus related subsystems.

## 2. The `h.Ren` builder model

Everything renderable implements the `h.Ren` interface:

```go
type Ren interface {
    Render(context *RenderContext)
}
```

**`*h.Element` is the primary node type.** Built via tag functions that take variadic `Ren` children:

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

**Built-in tag helpers** (partial list — see `framework/h/tag.go` for the full set): `Div`, `Span`, `P`, `A`, `Button`, `Form`, `Input`, `Label`, `Select`, `Option`, `Textarea`, `Img`, `Ul`, `Li`, `Ol`, `Table`, `Thead`, `Tbody`, `Tr`, `Th`, `Td`, `H1`–`H6`, `Svg`, `Path`, `Nav`, `Section`, `Article`, `Header`, `Footer`, `Main`, `Details`, `Summary`, `Html`, `Head`, `Body`, `Meta`, `Link`, `Script`, `Style`.

For tags not in the built-in list, use `h.Tag(name, children...)`:

```go
h.Tag("dialog",
    h.Id("confirm-dialog"),
    h.Text("Are you sure?"),
)
```

**Attributes are also `Ren`** — specifically `*h.AttributeR`, produced by `h.Attribute(key, value)`. Pass them as children; the renderer separates attrs from body children at render time:

```go
h.Div(
    h.Attribute("role", "dialog"),        // attribute
    h.Attribute("aria-label", "Confirm"), // attribute
    h.Class("p-4"),                       // attribute (convenience helper)
    h.Text("Body text here"),             // body child
)
```

**Common attribute helpers:**
- `h.Class("btn primary")` — sets `class`. Accepts multiple strings, joined with a space.
- `h.ClassX("base-classes", h.ClassMap{"active": isActive, "disabled": !canClick})` — base classes plus conditional extras from a map. Note the first arg is always the unconditional class string.
- `h.Id("my-id")` — sets `id`.
- `h.Type("submit")`, `h.Name("email")`, `h.Value("abc")`, `h.Placeholder("...")`, `h.Href("/x")`, `h.Src("/y")` — common attributes.
- `h.AttributePairs("data-foo", "bar", "data-baz", "qux")` — batch-set multiple attrs in one call (key, value, key, value, …).
- `h.Attributes(&h.AttributeMap{...})` — map-based bulk set (`AttributeMap` is `map[string]any`).

**Text helpers:**
- `h.Text("literal string")` — HTML-escaped text node.
- `h.TextF("count: %d", n)` — printf-style formatting.

**Escape hatch — use with care:**
- `h.UnsafeRaw("<span>raw</span>")` — emits the string without escaping. Never pass user input. Useful for trusted pre-rendered HTML (e.g. server-rendered markdown output).

**Children can be conditional or iterated:**
- `h.If(cond, ifTrue)` — includes `ifTrue` only when `cond` is true; otherwise renders nothing.
- `h.IfElse(cond, ifTrue, ifFalse)` — generic ternary (works for any type `T`, not just `Ren`).
- `h.List(items, func(item T, index int) *Element { ... })` — renders a typed slice; the primary loop builder for lists.
- `h.IterMap(m, func(key string, value T) *Element { ... })` — renders a `map[string]T` (order undefined).

**Worked example — a small card component:**

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

**Key mental model:** a page is a tree of `Ren` nodes. You don't emit HTML strings directly; you compose Go functions that each return a `Ren`, then hand the root to `h.NewPage` or `h.NewPartial` (next section).

## 3. Pages vs Partials

htmgo handlers return one of two response types:

**`*h.Page`** — a full HTML document. Created with `h.NewPage(root Ren)`. Used for normal navigation / first-load responses.

**`*h.Partial`** — an HTML fragment for htmx swaps. Created with `h.NewPartial(root *Element)` or `h.NewPartialWithHeaders(headers *Headers, root *Element)` when you need to set response headers (`HX-Retarget`, `HX-Trigger`, etc.). Note the argument order: headers first, root second.

### Route handler signatures

```go
// Page handler
func IndexPage(ctx *h.RequestContext) *h.Page {
    return h.NewPage(
        h.Div(h.Text("Hello")),
    )
}

// Partial handler
func GetSearch(ctx *h.RequestContext) *h.Partial {
    q := ctx.QueryParam("q")
    return h.NewPartial(
        h.Div(h.TextF("results for: %s", q)),
    )
}
```

### Auto-routing

htmgo scans `pages/` and `partials/` and generates `__htmgo/pages-generated.go` + `__htmgo/partials-generated.go` that register all routes. Mapping:

- `pages/index.go` → `/`
- `pages/users/index.go` → `/users`
- `pages/users/profile.go` → `/users/profile`
- `pages/blog/[slug].go` → `/blog/:slug` (URL param captured; access via `ctx.UrlParam("slug")`)

Generated files are **rebuilt on every `htmgo build`, `htmgo watch`, and `htmgo generate`.** Never edit them by hand; your changes will be wiped. Add them to `.gitignore`.

Exclude files from auto-routing in `htmgo.yml`:

```yaml
automatic_page_routing_ignore:
  - "**/_shared.go"
automatic_partial_routing_ignore:
  - "**/internal/*.go"
```

### Partial function-name conventions (HTTP method)

Partial filenames map to routes; the **function name prefix** determines the HTTP method:

```go
// partials/users/create.go → route: /users/create

func GetCreate(ctx *h.RequestContext) *h.Partial { ... }   // GET
func PostCreate(ctx *h.RequestContext) *h.Partial { ... }  // POST
func PutCreate(ctx *h.RequestContext) *h.Partial { ... }   // PUT
func PatchCreate(ctx *h.RequestContext) *h.Partial { ... } // PATCH
func DeleteCreate(ctx *h.RequestContext) *h.Partial { ... }// DELETE
```

Multiple methods can live in the same file.

### Full-page vs fragment decision

- If htmx is making the request (`ctx.IsHxRequest()` returns true), the response typically comes from a **partial** so only the target element is swapped.
- If the browser navigates directly (e.g. user types the URL or does a full refresh), you return a **page**.
- For routes that serve both cases (common for list pages), the handler can branch on `ctx.IsHxRequest()` or you can define the page + partial at the same path.

### Worked example

A button that posts to a partial and swaps the response:

```go
// pages/counter.go
func CounterPage(ctx *h.RequestContext) *h.Page {
    return h.NewPage(
        h.Div(
            h.Id("counter"),
            h.Text("count: 0"),
        ),
        h.Button(
            h.Post("/counter/increment"),
            h.HxTarget("#counter"),
            h.Attribute(hx.SwapAttr, hx.SwapTypeOuterHtml),
            h.Text("+"),
        ),
    )
}

// partials/counter/increment.go
func PostIncrement(ctx *h.RequestContext) *h.Partial {
    // In real code, read current count from state / session.
    n := 42
    return h.NewPartial(
        h.Div(
            h.Id("counter"),
            h.TextF("count: %d", n),
        ),
    )
}
```

## 4. RequestContext

Route handlers receive a `*h.RequestContext` that wraps Fiber's `fiber.Ctx` with htmx-aware helpers.

**Form / query / URL:**
- `ctx.FormValue(key string) string` — POST form field (multipart or url-encoded).
- `ctx.QueryParam(key string) string` — `?foo=bar` → `"bar"`.
- `ctx.UrlParam(key string) string` — pattern param from `/users/:id`.

**htmx-specific:**
- `ctx.IsHxRequest() bool` — true if the `HX-Request: true` header is present.
- `ctx.HxSource() string` — raw `HX-Source` header (e.g. `button#submit-btn`).
- `ctx.HxSourceID() string` — just the id portion (e.g. `submit-btn`).
- `ctx.HxRequestType() string` — `"full"` or `"partial"` (htmx 4 addition).

**Navigation:**
- `ctx.Redirect(url string, code int) error` — redirect; pass an HTTP status code (e.g. `fiber.StatusTemporaryRedirect`). For htmx-friendly client-side redirects from a partial, set the `HX-Redirect` response header instead (see headers example below).

**Escape hatch:**
- `ctx.Fiber` — the raw `fiber.Ctx`. Use for anything not pre-wrapped: cookies (`ctx.Fiber.Cookies(...)`), custom headers, SSE setup, etc.

**Looking up the RequestContext inside Fiber middleware:**

```go
func authMiddleware(c fiber.Ctx) error {
    ctx := h.GetRequestContext(c)
    if ctx.FormValue("token") == "" {
        return c.SendStatus(401)
    }
    return c.Next()
}
```

**Setting response headers from a handler** — use `h.NewHeaders` (variadic key/value pairs) to build a `*h.Headers`, then pass it as the first argument to `h.NewPartialWithHeaders`:

```go
func PostLogin(ctx *h.RequestContext) *h.Partial {
    if !validCreds(ctx) {
        return h.NewPartialWithHeaders(
            h.NewHeaders(hx.RetargetHeader, "#login-error"),
            h.Div(h.Class("error"), h.Text("Invalid credentials")),
        )
    }
    return h.NewPartialWithHeaders(
        h.NewHeaders(hx.RedirectHeader, "/dashboard"),
        h.Div(h.Text("Welcome!")),
    )
}
```

**Convenience header constructors** (all return `*h.Headers`):
- `h.ReplaceUrlHeader(url string)` — sets `HX-Replace-Url`.
- `h.PushUrlHeader(url string)` — sets `HX-Push-Url`.
- `h.CombineHeaders(headers ...*Headers)` — merges multiple `*Headers` into one.
- `h.NewHeaders(key, val, key, val, ...)` — general-purpose; key/value pairs must be even-length.

## 5. htmx integration (`hx/` package + `h.Hx*` helpers)

### The `hx/` package — constants only

`framework/hx/` is a pure constants package: attribute names, header names, event names (htmx 4 colon form), swap types. Import when you need a string value:

```go
import "github.com/franchb/htmgo/framework/hx"

h.Attribute("hx-swap:inherited", string(hx.SwapTypeInnerHtml))
```

**Attribute name constants** (`framework/hx/htmx.go`):
`GetAttr`, `PostAttr`, `PutAttr`, `PatchAttr`, `DeleteAttr`, `TargetAttr`, `TriggerAttr`, `SwapAttr`, `SwapOobAttr`, `SelectAttr`, `SelectOobAttr`, `IncludeAttr`, `IndicatorAttr`, `ConfirmAttr`, `BoostAttr`, `PushUrlAttr`, `ReplaceUrlAttr`, `ValsAttr`, `ValidateAttr`, `HeadersAttr`, `EncodingAttr`, `PreserveAttr`, `SyncAttr`, `DisableAttr` (disable form elements during in-flight; htmx 4 semantics — **repurposed** from htmx 2 where it stopped htmx processing), `IgnoreAttr` (stop htmx processing within the element subtree; htmx 4 migration of htmx 2 `hx-disable`), `ConfigAttr` (replaces `hx-request`), `StatusAttr` (per-status swap/target control).

**Event name constants** (htmx 4 colon form):
`AfterSwapEvent = "htmx:after:swap"`, `BeforeSwapEvent = "htmx:before:swap"`, `AfterRequestEvent = "htmx:after:request"`, `BeforeRequestEvent = "htmx:before:request"`, `AfterOnLoadEvent = "htmx:after:init"`, `BeforeOnLoadEvent = "htmx:before:init"`, `ConfigRequestEvent = "htmx:config:request"`, `ErrorEvent = "htmx:error"`, `AfterSettleEvent = "htmx:after:settle"`, `AbortEvent = "htmx:abort"`.

**Swap type constants:** `SwapTypeInnerHtml`, `SwapTypeOuterHtml`, `SwapTypeBeforeBegin`, `SwapTypeAfterBegin`, `SwapTypeBeforeEnd`, `SwapTypeAfterEnd`, `SwapTypeTextContent`, `SwapTypeDelete`, `SwapTypeNone`, `SwapTypeTrue`.

**Header name constants:** `RequestHeader = "HX-Request"`, `RedirectHeader = "HX-Redirect"`, `RetargetHeader = "HX-Retarget"`, `ReswapHeader = "HX-Reswap"`, `ReselectHeader = "HX-Reselect"`, `TriggerHeader = "HX-Trigger"`, `PushUrlHeader = "HX-Push-Url"`, `ReplaceUrlHeader = "HX-Replace-Url"`, `LocationHeader = "HX-Location"`, `SourceHeader = "HX-Source"` (htmx 4), `RequestTypeHeader = "HX-Request-Type"` (htmx 4 — `"full"` or `"partial"`), `CurrentUrlHeader = "HX-Current-Url"`, `RefreshHeader = "HX-Refresh"`.

### Setting htmx attributes directly

There are **no** `h.HxGet`, `h.HxPost`, `h.HxPut`, `h.HxPatch`, `h.HxDelete`, or `h.HxSwap` standalone builder functions. Use one of these patterns:

**Option A — raw `Attribute` with `hx.*Attr` constants:**

```go
h.Attribute(hx.GetAttr, "/search")   // hx-get="/search"
h.Attribute(hx.PostAttr, "/users")    // hx-post="/users"
h.Attribute(hx.SwapAttr, hx.SwapTypeInnerHtml)  // hx-swap="innerHTML"
```

**Option B — `h.Get` / `h.Post` composite helpers** (set both `hx-*` and `hx-trigger` in one call):

```go
// Get(path, trigger...) → sets hx-get + hx-trigger
h.Button(h.Get("/search", "click"), h.Text("Search"))

// Post(url, trigger...) → sets hx-post + hx-trigger
h.Form(h.Post("/users/create", "submit"), ...)

// Partial-path variants (type-safe; avoid hard-coding URLs)
h.Button(h.GetPartial(SearchResults, "click"), h.Text("Search"))
h.Button(h.PostOnClick("/items/delete"), h.Text("Delete"))
```

**Other `h.*` attribute helpers that DO exist:**

```go
h.HxTarget("#results")                // hx-target="#results"
h.HxInclude("form[name=filter]")      // hx-include="..."
h.HxConfirm("Really delete?")         // hx-confirm="..."
h.HxIndicator("#spinner")             // hx-indicator="..."
```

### Explicit inheritance — htmx 4 BREAKING CHANGE

In htmx 2, `hx-target` on an ancestor cascaded to descendants automatically. **htmx 4 removed implicit inheritance.** If you set `hx-target` on a wrapper element and expect its children to use that target, you MUST use the `:inherited` form.

`h.Hx*Inherited` helpers:
- `h.HxTargetInherited("#x")` → `hx-target:inherited="#x"`
- `h.HxIncludeInherited(sel)`
- `h.HxSwapInherited(strategy)`
- `h.HxBoostInherited(value)`
- `h.HxConfirmInherited(msg)`
- `h.HxHeadersInherited(jsonStr)`
- `h.HxIndicatorInherited(sel)`
- `h.HxSyncInherited(spec)`
- `h.HxEncodingInherited(enc)`
- `h.HxValidateInherited(value)`

*`hx-config` does NOT support inheritance; configure htmx globally via `htmx.config` or a `<meta name="htmx-config">`.*

**Wrong (htmx 2 thinking):**

```go
h.Div(
    h.HxTarget("#results"),  // does NOT cascade in htmx 4
    h.Button(h.Get("/a", "click"), h.Text("A")),
    h.Button(h.Get("/b", "click"), h.Text("B")),
)
```

**Right (htmx 4):**

```go
h.Div(
    h.HxTargetInherited("#results"),  // cascades to descendants
    h.Button(h.Get("/a", "click"), h.Text("A")),
    h.Button(h.Get("/b", "click"), h.Text("B")),
)
```

### Triggers

**Simple raw-string form** (`HxTriggerString` joins multiple strings with `", "`):

```go
h.HxTriggerString("keyup changed delay:300ms", "search")
// renders: hx-trigger="keyup changed delay:300ms, search"
```

**Structured form** using `hx.TriggerEvent` values built by `hx.OnEvent` / `hx.OnClick` / `hx.OnChange` / `hx.OnLoad`:

```go
h.HxTrigger(
    hx.OnClick(hx.OnceModifier{}),
    hx.OnEvent("keyup", hx.StringModifier("keyCode==27"), hx.StringModifier("from:body")),
)
```

`h.HxTrigger` takes `...hx.TriggerEvent` and calls `hx.NewTrigger(opts...)` internally — each `TriggerEvent` is a plain struct with an event name and modifier list.

**`hx.NewTrigger`** signature: `NewTrigger(opts ...TriggerEvent) *Trigger` — collects multiple events into a `Trigger` whose `ToString()` produces the full `hx-trigger` string. You rarely call this directly; prefer `h.HxTrigger(...)`.

**Available modifiers** (`framework/hx/modifiers.go`):
- `hx.OnceModifier{}` — appends `"once"`
- `hx.Delay(n int)` — appends `"delay:<n>s"` (seconds)
- `hx.Throttle(n int)` — appends `"throttle:<n>s"` (seconds)
- `hx.StringModifier("...")` — raw modifier string (use for anything not covered above, e.g. filter expressions, `from:`, `target:`)

`Once()`, `KeyEquals()`, `From()`, and `DelayMs()` **do not exist**; use `hx.OnceModifier{}`, `hx.StringModifier("[keyCode==27]")`, `hx.StringModifier("from:body")`, and `hx.Delay(n)` respectively.

**Convenience click shortcut:**

```go
h.HxTriggerClick(hx.OnceModifier{})   // hx-trigger="click once"
// HxTriggerClick(opts ...hx.Modifier) is sugar for HxTrigger(hx.OnClick(opts...))
```

**Predefined trigger constants** in `framework/hx/triggers.go`:

```go
hx.TriggerClick          // "onclick"
hx.TriggerClickOnce      // "onclick once"
hx.TriggerKeyUpEnter     // "keyup[keyCode==13]"
hx.TriggerBlur           // "blur"
hx.TriggerEvery1s        // "every:1s"
hx.TriggerEvery5s        // "every:5s"
// ...and more
```

### Response headers

Set headers on a partial response via `h.NewPartialWithHeaders` + the `h.NewHeaders` variadic helper (see §4 for full details):

```go
h.NewPartialWithHeaders(
    h.NewHeaders(
        hx.TriggerHeader, "showToast",     // fires a client event
        hx.RetargetHeader, "#status",      // changes swap target
        hx.ReswapHeader, "afterend",       // changes swap style
    ),
    h.Div(h.Text("Saved")),
)
```

Common header uses:
- `hx.RedirectHeader` — full-page redirect (htmx handles)
- `hx.LocationHeader` — SPA-style navigation with pushState (JSON payload)
- `hx.PushUrlHeader` / `hx.ReplaceUrlHeader` — update browser URL without reload
- `hx.RefreshHeader` — force a full page refresh
- `hx.TriggerHeader` — dispatch a client-side event
- `hx.RetargetHeader` — change the swap target server-side
- `hx.ReswapHeader` — change the swap style server-side
- `hx.ReselectHeader` — change which element is selected from the response

### Worked example — search box with debounce

```go
func SearchBox() *h.Element {
    return h.Div(
        h.Input(
            h.Type("search"),
            h.Name("q"),
            h.Placeholder("Search..."),
            h.Attribute(hx.PostAttr, "/search"),
            h.HxTriggerString("keyup changed delay:300ms", "search"),
            h.HxTarget("#search-results"),
            h.Attribute(hx.SwapAttr, hx.SwapTypeInnerHtml),
        ),
        h.Div(h.Id("search-results")),
    )
}
```

### See also

For general htmx patterns (attribute semantics, event lifecycle, OOB swaps, SSE/WS), the `htmx-guidance` skill covers the htmx side in depth. This skill focuses on the Go builder layer; the two are complementary.

## 6. Alpine integration (`ax/` package + bundled `alpine-compat`)

Alpine.js is **optional**. When you need small client-side state (open/closed, theme, density, copy-to-clipboard, command-palette modals), reach for Alpine. For server round-trips, reach for htmx. For imperative one-shots without state, reach for the lifecycle command DSL (see `h/lifecycle.go`).

### Setup (two lines of HTML + one CSS rule)

1. Add the Alpine CDN script to your layout `<head>`:

```go
h.Head(
    h.Tag("script",
        h.Attribute("src", "https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js"),
        h.Attribute("defer", ""),
    ),
    h.Script("/public/htmgo.js"),
)
```

The `defer` attribute is important — Alpine initializes on `DOMContentLoaded`, so it must defer until `[x-data]` nodes are parsed. `htmgo.js` can load in either order; extensions self-register at import time in htmx 4.

2. Add the FOUC-prevention CSS rule to your stylesheet:

```css
[x-cloak] { display: none !important; }
```

The `alpine-compat` htmx extension is **already bundled** in `/public/htmgo.js` — no additional script tag, no `hx-ext` attribute. It auto-gates on `window.Alpine?.*` presence, so if Alpine isn't loaded, every hook no-ops.

### The `ax/` package — constants + builders

`framework/ax/` mirrors the `hx/` shape: string constants for directive names, plus `h.Ren`-returning builders. Import as `github.com/franchb/htmgo/framework/ax`.

**Constants** (all 18 are `ax.Attribute` which is `type Attribute = string`): `DataAttr`, `InitAttr`, `ShowAttr`, `BindAttr`, `OnAttr`, `TextAttr`, `HtmlAttr`, `ModelAttr`, `ModelableAttr`, `CloakAttr`, `RefAttr`, `IgnoreAttr`, `TeleportAttr`, `EffectAttr`, `IfAttr`, `ForAttr`, `IdAttr`, `TransitionAttr`.

**Simple single-arg directives** (Alpine expression as the value):

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

**No-arg directives** (emit the bare attribute; Alpine accepts the boolean form):

```go
ax.Cloak()       // x-cloak
ax.Ignore()      // x-ignore
ax.Transition()  // x-transition
```

For richer transitions (`x-transition:enter`, `.opacity`, `.duration.500ms`), drop to `h.Attribute("x-transition:enter", ...)` directly.

**`x-bind:*` family** — bind any attribute to an expression. `ax.Bind(attr, expr string) h.Ren` is the generic form; the seven shortcuts cover the most-used targets:

```go
ax.Bind("data-foo", "value")  // x-bind:data-foo="value"

ax.BindClass("{ active: isActive }")
ax.BindStyle("{ color: hex }")
ax.BindHref("url")
ax.BindValue("input")
ax.BindDisabled("locked")
ax.BindChecked("selected")
ax.BindId("compId")
```

**`x-on:*` family** — event handlers. `ax.On(event, handler string, modifiers ...string) h.Ren` is the generic form; eight event shortcuts and three combo shortcuts forward to it:

```go
ax.On("click", "count++")                         // x-on:click="count++"
ax.On("click", "submit()", "prevent")             // x-on:click.prevent="submit()"
ax.On("keydown", "handle()", "meta", "k", "prevent")
// emits: x-on:keydown.meta.k.prevent="handle()"

// Event shortcuts — each signature: (handler string, mods ...string) h.Ren
ax.OnClick(handler, mods...)
ax.OnSubmit(handler, mods...)
ax.OnInput(handler, mods...)
ax.OnChange(handler, mods...)
ax.OnFocus(handler, mods...)
ax.OnBlur(handler, mods...)
ax.OnKeydown(handler, mods...)
ax.OnKeyup(handler, mods...)

// Combo shortcuts — signature: (handler string) h.Ren (no modifier slot; the
// modifier is baked in)
ax.OnClickOutside("open = false")     // @click.outside
ax.OnKeydownEscape("open = false")    // @keydown.escape
ax.OnKeydownEnter("submit()")         // @keydown.enter
```

**`x-model` modifier variants:**

```go
ax.ModelNumber("age")               // x-model.number="age"
ax.ModelLazy("title")               // x-model.lazy="title"
ax.ModelTrim("name")                // x-model.trim="name"
ax.ModelFill("notes")               // x-model.fill="notes"
ax.ModelBoolean("checked")          // x-model.boolean="checked"
ax.ModelDebounce("query", "500ms")  // x-model.debounce.500ms="query"
```

Note `ax.ModelDebounce(expr, duration string)` — the duration is a plain string (`"500ms"`, `"2s"`), not a `time.Duration`. Whatever you pass is appended verbatim to the attribute name.

### When to use `ax.*` vs htmx swaps vs lifecycle commands

| Use case | Tool |
|---|---|
| Pure-client state that never needs the server (theme toggle, popover open/close, keyboard overlay, copy-to-clipboard) | `ax.*` |
| Server round-trip required (load data, save form, refresh a list) | htmx — `h.Get(path, trigger...)`, `h.Post(url, trigger...)`, or `h.Attribute(hx.GetAttr, ...)` |
| Widget with BOTH client state AND server-loaded content (popover whose body is fetched lazily) | `ax.Data(...)` on the outer wrapper + `h.Get(...)` on inner elements. The alpine-compat extension carries `_x_dataStack` across the morph swap. |
| Imperative one-shots without state (set a value, add/remove a class, fire an alert) | Lifecycle commands — `h.OnClick(js.SetValue("…"))`, not Alpine |

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

All three layers compose: Alpine manages the `open` flag, htmx lazy-loads the body on first intersect, and the bundled `alpine-compat` extension preserves Alpine state through the morph swap automatically.

### Gotchas

- **Alpine v3 only.** v2 is unsupported upstream; the compat extension uses v3 internal APIs (`closestDataStack`, `cloneNode`, `deferMutations`, `destroyTree`, `flushAndStopDeferringMutations`).
- **Load Alpine with `defer` in `<head>`.** Without `defer`, Alpine may initialize before `[x-data]` elements are parsed and silently skip them.
- **Alpine plugins load before Alpine itself.** Per upstream plugin docs — `@alpinejs/persist`, `@alpinejs/intersect`, `@alpinejs/focus`, etc. go in `<script>` tags ABOVE the main Alpine script.
- **`[x-cloak]` CSS rule is required** to prevent FOUC. Without it, Alpine-hidden elements flash visible on page load before Alpine initializes.
- **Don't outer-swap an Alpine root element** with htmx. Morphs of inner content preserve `_x_dataStack`; full replacement of the `[x-data]` root loses state. Either swap inner content only, or re-attach state via `ax.Data(...)` on the new root.

## 7. Lifecycle & JS command DSL

Not every interaction needs Alpine or a server round-trip. For **imperative one-shots** (fire an alert, set a value, add a class, mutate the DOM), htmgo provides event helpers in `framework/h/lifecycle.go` that emit generated JavaScript at render time.

### Event helpers

Bind a JS command to a DOM event:

```go
h.Button(
    h.OnClick(js.Alert("Hello!")),
    h.Text("Say hi"),
)
```

The narrow set of named helpers (see `framework/h/lifecycle.go`) — use `h.OnEvent(event, cmd...)` for anything not on this list:

- DOM events: `h.OnClick(cmd...)`, `h.OnSubmit(cmd...)`, `h.OnLoad(cmd...)`.
- Generic: `h.OnEvent(event hx.Event, cmd...)` — pass any event name, e.g. `h.OnEvent("change", js.Alert("changed"))` or `h.OnEvent("input", ...)`, `h.OnEvent("keyup", ...)`, `h.OnEvent("focus", ...)`, `h.OnEvent("blur", ...)`, `h.OnEvent("mouseover", ...)`, etc.
- htmx request lifecycle: `h.HxBeforeRequest(cmd...)`, `h.HxAfterRequest(cmd...)`, `h.HxOnAfterSwap(cmd...)`, `h.HxOnMutationError(cmd...)`.
- SSE hooks: `h.HxBeforeSseMessage(cmd...)`, `h.HxAfterSseMessage(cmd...)`, `h.HxOnSseError(cmd...)`, `h.HxOnSseClose(cmd...)`, `h.HxOnSseConnecting(cmd...)`, `h.HxOnSseOpen(cmd...)`.

Note: `h.OnLoad` works on any element because htmgo's bundled htmx extension fires a synthetic `load` event on swap-in — useful for initializing a just-swapped fragment.

### JS commands (`framework/js/`)

The `framework/js/` package is a set of variable aliases re-exporting the command constructors from `framework/h/`. Each helper takes one or more `js.Command` values. Common commands (non-exhaustive):

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

There's no `js.Redirect` or `js.Focus` — use `js.EvalJs("window.location = '/home'")` or `js.EvalJs("self.focus()")`.

Commands chain automatically — pass multiple to one handler:

```go
h.OnClick(
    js.AddClass("fade-out"),
    js.EvalJs("setTimeout(() => self.remove(), 300)"),
)
```

### Two command shapes

- `SimpleJsCommand` — single statement; inlined into the `hx-on:<event>="..."` attribute value. Examples: `Alert`, `AddClass`, `SetValue`, `SetText`, `Remove`.
- `ComplexJsCommand` — multi-statement block; htmgo emits a generated `__eval_<id>()` helper and wires the handler to invoke it. Examples: `EvalJs`, `SetTextOnChildren`, `EvalJsOnParent`, `RunAfterTimeout`.

Both satisfy the `Command` interface (aliased to `Ren`), so you can mix them freely in a single handler.

### Contrast with Alpine's `ax.OnClick` and htmx requests

Three superficially-similar things do different jobs — picking the right one matters.

| Expression | Mechanism | When to use |
|---|---|---|
| `h.OnClick(js.Alert("hi"))` | htmgo-generated JS wired via `hx-on:click` | Imperative one-shot with no state and no server call |
| `ax.OnClick("open = !open")` | `x-on:click="..."` evaluated by Alpine at runtime against the component's `x-data` scope | Mutate Alpine state |
| `h.Button(h.Get("/x", "click"))` | htmx request — click triggers a GET; server returns a fragment to swap | Server round-trip |

**Common mistake:** putting `h.OnClick(js.X)` AND `ax.OnClick("...")` on the same element. Both fire. Pick one per element, or use separate nested elements if you genuinely need both.

## 8. Service locator (DI)

For dependencies that handlers need (DB connections, config, external API clients), use `framework/service/`. The locator is generic and type-keyed — each registered type can be resolved from any `RequestContext` via `ctx.ServiceLocator()`.

### Setup in `main.go`

```go
package main

import (
    "database/sql"

    "github.com/franchb/htmgo/framework/h"
    "github.com/franchb/htmgo/framework/service"

    "myapp/__htmgo"
)

func main() {
    locator := service.NewLocator()

    service.Set[sql.DB](locator, service.Singleton, func() *sql.DB {
        return openDB()
    })
    service.Set[Config](locator, service.Singleton, loadConfig)

    h.Start(h.AppOpts{
        ServiceLocator: locator,
        LiveReload:     true,
        Register: func(app *h.App) {
            __htmgo.Register(app.Router)
        },
    })
}
```

Key signature details:
- `service.Set[T](locator, lifecycle, func() *T)` — the provider **must return a pointer** to `T`. Don't write `service.Set[*sql.DB]`; write `service.Set[sql.DB]` and the provider returns `*sql.DB`.
- `service.Get[T](locator) *T` — always returns `*T`.

### Resolving in handlers

```go
func GetUsers(ctx *h.RequestContext) *h.Partial {
    db := service.Get[sql.DB](ctx.ServiceLocator())
    // ... use db ...
    return h.NewPartial(renderUsers())
}
```

### Lifecycles

`service.Singleton` and `service.Transient` are the two string-valued lifecycle constants.

- `service.Singleton` — provider runs once; same `*T` returned for every `Get`. Cached after first resolve.
- `service.Transient` — provider runs on every `Get`; returns a fresh instance.

Pick `Singleton` for pools, long-lived clients, config. Pick `Transient` for per-request scoped state (though request-scoped values are better stored via `ctx.Set(key, value)` / `ctx.Get(key)` on `RequestContext` than via the locator).

If a `Get[T]` is called for an unregistered type, the framework calls `log.Fatalf` — registration errors fail fast at first use, not at startup.

### Worked example — DB handle + config

```go
// main.go
locator := service.NewLocator()

service.Set[Config](locator, service.Singleton, func() *Config {
    return mustLoad("config.yaml")
})

service.Set[sql.DB](locator, service.Singleton, func() *sql.DB {
    cfg := service.Get[Config](locator)
    return sql.MustOpen(cfg.DatabaseURL)
})

// partials/search/query.go
func GetQuery(ctx *h.RequestContext) *h.Partial {
    db := service.Get[sql.DB](ctx.ServiceLocator())
    q := ctx.QueryParam("q")
    rows := queryRows(db, q)
    return h.NewPartial(renderResults(rows))
}
```

Providers can resolve other services from the same locator (as `sql.DB`'s provider does with `Config` above) — the locator releases its internal lock before invoking a Singleton provider, so nested resolution doesn't deadlock.

## 9. Caching

### High-level helpers (`h.Cached` family)

htmgo's cache helpers in `framework/h/cache.go` wrap `*Element` nodes and return a *constructor* you call at render time:

```go
// Package-level: cache once, globally.
CachedHeader := h.Cached(5*time.Minute, func() *h.Element {
    return h.Header(h.H1(h.Text("Welcome")))
})
// Inside a page/partial:
return h.NewPage(h.Div(CachedHeader()))
```

Signature: `h.Cached(duration time.Duration, cb func() *h.Element, opts ...h.CacheOption) func() *h.Element`. Globally cached, not per-request — the callback runs once per `duration`.

For user-scoped or parameterized caching, use the `PerKey` variants, whose callback returns `(key, renderFunc)`:

```go
UserProfile := h.CachedPerKeyT(15*time.Minute,
    func(u User) (int, h.GetElementFunc) {
        return u.ID, func() *h.Element { return renderProfile(u) }
    })
UserProfile(currentUser)   // returns *h.Element
```

Also: `CachedT`/`T2`/`T3`/`T4` (parameterized, globally cached — NOT per-key), and `CachedPerKeyT2`/`T3`/`T4`. There is **no** `h.Cache` or `h.CacheGlobal`.

### Swapping the underlying store

The default is `cache.TTLStore` via `h.DefaultCacheProvider`. Override per-component with `h.WithCacheStore(store)`, or replace the default globally:

```go
lru := cache.NewLRUStore[any, string](1000)
UserProfile := h.CachedPerKeyT(15*time.Minute, profileCb, h.WithCacheStore(lru))

// Or app-wide:
h.DefaultCacheProvider = func() cache.Store[any, string] {
    return cache.NewLRUStore[any, string](5000)
}
```

### Low-level pluggable stores

`framework/h/cache/interface.go` defines:

```go
type Store[K comparable, V any] interface {
    Set(key K, value V, ttl time.Duration)
    Get(key K) (V, bool)
    GetOrCompute(key K, compute func() V, ttl time.Duration) V
    Delete(key K); Purge(); Close()   // Close returns no error
}
```

TTL is per call, not per store. **Built-in:** `cache.NewTTLStore[K, V]()`, plus `NewTTLStoreWithInterval(d)` and `NewTTLStoreWithMaxSize(n)` (TTL + LRU cap). **Example:** `cache.NewLRUStore[K, V](maxSize)` in `framework/h/cache/lru_store_example.go` — copy its shape for a Redis/distributed store, or see `ExampleDistributedCacheAdapter` in `framework/h/cache/example_test.go` for the adapter pattern.

### When to cache / not cache

**Cache:** expensive DB reads shown on many pages; shared nav/sidebar/footer fragments; rate-limited third-party responses.

**Don't cache:** user-specific content without a user-scoped key (use `CachedPerKey*`, never plain `Cached`); data with strict freshness requirements; content cheap enough that caching overhead dominates — measure first.

## 10. Project configuration & CLI

### `htmgo.yaml` / `htmgo.yml` / `_htmgo.yaml` / `_htmgo.yml`

Lives at the app root. The loader searches in that order and uses the first it finds (`framework/config/project.go`). Fields (yaml tags are ground truth):

```yaml
tailwind: true                    # run Tailwind CSS compilation
tailwind_version: "4"             # optional; auto-detected otherwise
watch_ignore: ["**/node_modules/**", "**/.git/**", "assets/dist/**"]  # globs watcher ignores
watch_files:  ["**/*.go", "**/*.css", "**/*.md"]     # globs watcher rebuilds on
automatic_page_routing_ignore:    ["_shared.go"]     # skip for page auto-routing
automatic_partial_routing_ignore: ["internal/*.go"]  # skip for partial auto-routing
public_asset_path: "/public"      # URL prefix for static assets
```

Defaults in `config.DefaultProjectConfig()`: `tailwind: true`, `public_asset_path: "/public"`, standard watch globs. `automatic_page_routing_ignore` entries are auto-prefixed with `pages/`; partial entries with `partials/`.

### CLI commands

Install: `cd /path/to/htmgo/cli/htmgo && go install .` — or one-shot `go run github.com/franchb/htmgo/cli/htmgo@latest <subcommand>`.

Subcommands (from `cli/htmgo/runner.go` — no `htmgo dev` alias, use `watch`):

- `htmgo template [name]` — scaffold a new app from a starter template into `./<name>`.
- `htmgo setup` — run the project setup task in the current app.
- `htmgo build` — full production build: regenerates `__htmgo/`, builds CSS, emits binary.
- `htmgo watch` — live-reload dev server: assets, CSS, `__htmgo/` + ent regen, server, watcher.
- `htmgo run` — build and run without the watcher.
- `htmgo generate` — regenerate `__htmgo/*-generated.go` + run ent codegen.
- `htmgo css` — Tailwind/CSS build only.
- `htmgo schema` — prompts for an entity name and scaffolds an ent schema.
- `htmgo format <file|.>` — format a file or the whole working directory.
- `htmgo version` — print the CLI version.

### `__htmgo/` — generated, never edit

Every `htmgo build` / `watch` / `generate` overwrites `__htmgo/*-generated.go`. Add `__htmgo/` to `.gitignore`.

Example apps wrap these commands in a `Taskfile.yml` (`task watch` / `task build` / `task run`) — copy `examples/todo-list/Taskfile.yml` when bootstrapping.
