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
- `h.IterMap(m, func(key string, value T) *Element { ... })` — renders a `map[string]T`; iterates in insertion order.

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
