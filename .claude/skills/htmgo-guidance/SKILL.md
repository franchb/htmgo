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
            h.HxPost("/counter/increment"),
            h.HxTarget("#counter"),
            h.HxSwap(hx.SwapTypeOuterHtml),
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
