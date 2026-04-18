# htmgo-guidance Claude Code Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Write a single comprehensive Claude Code skill (`htmgo-guidance`) at `.claude/skills/htmgo-guidance/SKILL.md` that teaches AI sessions working in downstream htmgo apps how to use the framework's Go builder, routing, `hx/`, and `ax/` packages correctly. Add a README subsection explaining how consumer projects pull the skill in. Copy the skill into the primary consumer (`vulnerability_catalog`) and have the user run a real Claude Code smoke test.

**Architecture:** The skill is a single monolithic markdown file (~700-850 lines) with YAML frontmatter that auto-activates on htmgo-related Go work. Content is organized into 12 sections from foundational (`h.Ren` builder, Pages/Partials) through integration (`hx/`, `ax/`, lifecycle) to advanced (service locator, caching, CLI) plus pitfalls and upgrade pointers. Distribution is in-repo copy-paste — matches the shape of the 5 existing `htmx-*` skills already shipped in `.claude/skills/`.

**Tech Stack:** Markdown with YAML frontmatter. No code. Reference targets: Go 1.23 source in `framework/h/`, `framework/hx/`, `framework/ax/`, `framework/service/`, `framework/h/cache/`, `framework/config/`, `cli/htmgo/`. Smoke test uses Claude Code at `/home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog`.

**Reference files (the engineer should read these before starting):**
- Spec: `docs/superpowers/specs/2026-04-18-htmgo-claude-skills-design.md` (the ground truth for what goes in each section)
- Style exemplar: `.claude/skills/htmx-guidance/SKILL.md` — 644 lines, reference + example flow
- Other existing skills: `.claude/skills/htmx-{debugging,extension-authoring,migration,upgrade-from-htmx2}/SKILL.md`
- Framework source to summarize (use as ground truth for function names, signatures, defaults):
  - `framework/h/tag.go`, `framework/h/attribute.go`, `framework/h/render.go`, `framework/h/lifecycle.go`, `framework/h/page.go`, `framework/h/partial.go`, `framework/h/app.go`, `framework/h/cache.go`
  - `framework/hx/htmx.go`, `framework/hx/trigger.go`, `framework/hx/modifiers.go`
  - `framework/ax/alpine.go`, `framework/ax/builder.go`
  - `framework/service/` (all files)
  - `framework/h/cache/` (interface + examples)
  - `framework/config/` (htmgo.yml schema)
- Project context: `CLAUDE.md` (project instructions; SHOULD NOT contradict the skill)
- Changelog: `CHANGELOG.md` `[Unreleased]` + `[1.2.0-beta.1]` entries — ground truth for upgrade pointers
- Alpine compat context: `docs/superpowers/specs/2026-04-18-alpine-compat-design.md`

**Commit style:** Use Conventional Commits consistent with the repo's history. Scope `docs(skills):` for skill content; `docs:` for README changes. Trailer:
```
Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
```

---

## File Structure

**Created:**
- `.claude/skills/htmgo-guidance/SKILL.md` (new; ~700-850 lines when complete)

**Modified:**
- `README.md` (+15-20 lines for the "Claude Code skills" subsection)

**Created in a different repo (consumer):**
- `/home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog/.claude/skills/htmgo-guidance/SKILL.md` (copied from this repo; separate git commit in that repo)

**Not touched:**
- No framework code (`framework/**/*.go`) — this is a docs-only change.
- No existing skills in `.claude/skills/htmx-*` — those stay as-is.
- No new tests — skills are markdown, not code.
- No CHANGELOG entry for this repo — the skill is a docs artifact, not a framework feature.

---

## Line-count budget

The spec caps the skill at 900 lines (hard limit; partial-load risk past that). Rough budget per section:

| § | Section | Target lines | Cumulative |
|---|---------|-------------:|-----------:|
| 1 | What htmgo is | ~20 | 20 |
| 2 | Builder model | ~80 | 100 |
| 3 | Pages vs Partials | ~70 | 170 |
| 4 | RequestContext | ~60 | 230 |
| 5 | htmx integration | ~150 | 380 |
| 6 | Alpine integration | ~130 | 510 |
| 7 | Lifecycle & commands | ~50 | 560 |
| 8 | Service locator | ~30 | 590 |
| 9 | Caching | ~40 | 630 |
| 10 | Config & CLI | ~40 | 670 |
| 11 | Common pitfalls | ~50 | 720 |
| 12 | Upgrade pointers | ~25 | 745 |
| - | Header + frontmatter + version stamp + TOC | ~50 | **~795** |

Target ~800 lines; hard cap 900. If a section overshoots, trim before committing.

---

## Task 1: Scaffold skill file, frontmatter, version stamp, and §1 "What htmgo is"

**Files:**
- Create: `.claude/skills/htmgo-guidance/SKILL.md`

- [ ] **Step 1.1: Create the skill directory and initial file**

```bash
mkdir -p .claude/skills/htmgo-guidance
```

- [ ] **Step 1.2: Write the frontmatter + version stamp + §1**

Write `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown
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
````

- [ ] **Step 1.3: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): scaffold htmgo-guidance skill with frontmatter and §1

Creates the new Claude Code skill at .claude/skills/htmgo-guidance/SKILL.md
with YAML frontmatter, a "last reviewed against" version stamp, and the
opening "What htmgo is" section. Subsequent commits add sections 2-12.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: §2 The `h.Ren` builder model

**Files:**
- Modify: `.claude/skills/htmgo-guidance/SKILL.md` (append §2 after §1)

- [ ] **Step 2.1: Survey the builder surface**

Before writing, verify actual exported function names and signatures. Run:

```bash
cd /home/iru/p/github.com/franchb/htmgo
grep -n "^func [A-Z]" framework/h/tag.go | head -40
grep -n "^func " framework/h/attribute.go | head -30
grep -n "^type " framework/h/render.go framework/h/tag.go | head -10
```

Note any name that differs from what you expect (e.g. `TextF` vs `Textf`, `UnsafeRaw` vs `Raw`). The skill must use the actual exported names.

- [ ] **Step 2.2: Append §2 content**

Append to `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown

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
    h.Attribute("role", "dialog"),       // attribute
    h.Attribute("aria-label", "Confirm"), // attribute
    h.Class("p-4"),                      // attribute (convenience helper)
    h.Text("Body text here"),            // body child
)
```

**Common attribute helpers:**
- `h.Class("btn primary")` — sets `class`.
- `h.ClassX(h.ClassMap{"active": isActive, "disabled": !canClick})` — conditional classes from a map.
- `h.Id("my-id")` — sets `id`.
- `h.Type("submit")`, `h.Name("email")`, `h.Value("abc")`, `h.Placeholder("..."), `h.Href("/x")`, `h.Src("/y")` — common attributes.
- `h.AttributePairs("data-foo", "bar", "data-baz", "qux")` — batch-set multiple attrs in one call.
- `h.Attributes(&h.AttributeMap{...})` — map-based bulk set.

**Text helpers:**
- `h.Text("literal string")` — HTML-escaped text node.
- `h.TextF("count: %d", n)` — printf-style formatting.

**Escape hatch — use with care:**
- `h.UnsafeRaw("<span>raw</span>")` — emits the string without escaping. Never pass user input. Useful for trusted pre-rendered HTML (e.g. server-rendered markdown output).

**Children can be conditional:**
- `h.If(cond, ifTrue)` — includes `ifTrue` only when `cond` is true; otherwise renders nothing.
- `h.IfElse(cond, ifTrue, ifFalse)` — ternary.
- `h.Map(items, func(item T) Ren { ... })` — renders a slice; common for lists.
- `h.Range(n, func(i int) Ren { ... })` — renders `n` iterations.

**Worked example — a small card component:**

```go
func Card(title, body string, active bool) *h.Element {
    return h.Div(
        h.ClassX(h.ClassMap{
            "rounded-lg border p-4":   true,
            "border-blue-500 bg-blue-50": active,
            "border-gray-300":            !active,
        }),
        h.H3(h.Class("font-semibold"), h.Text(title)),
        h.P(h.Class("mt-2 text-gray-700"), h.Text(body)),
    )
}
```

**Key mental model:** a page is a tree of `Ren` nodes. You don't emit HTML strings directly; you compose Go functions that each return a `Ren`, then hand the root to `h.NewPage` or `h.NewPartial` (next section).
````

- [ ] **Step 2.3: Spot-check**

Verify every function name mentioned above actually exists:

```bash
grep -E "^func (Div|Span|Text|TextF|UnsafeRaw|Attribute|Class|ClassX|If|IfElse|Map|Range)" framework/h/*.go | head -20
```

Any name missing? Remove it from the skill or replace with the actual name.

- [ ] **Step 2.4: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): §2 h.Ren builder model

Covers the Ren interface, *h.Element, variadic children, attributes as
Ren, common attribute/text/conditional helpers, and a worked card
component example.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: §3 Pages vs Partials + §4 RequestContext

Two related sections — routing mechanics and the request context consumed inside handlers — combined for cohesion.

**Files:**
- Modify: `.claude/skills/htmgo-guidance/SKILL.md`

- [ ] **Step 3.1: Survey**

```bash
grep -n "^func NewPage\|^func NewPartial\|^type Page \|^type Partial " framework/h/*.go
grep -n "^func .* RequestContext" framework/h/*.go | head -20
grep -n "^type RequestContext" framework/h/*.go
```

Verify: `NewPage`, `NewPartial`, `NewPartialWithHeaders` exist; `RequestContext` methods are `FormValue`, `QueryParam`, `UrlParam`, `IsHxRequest`, `Redirect`, `HxSource`, `HxSourceID`, `HxRequestType`, `.Fiber` field.

- [ ] **Step 3.2: Append §3**

Append to `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown

## 3. Pages vs Partials

htmgo handlers return one of two response types:

**`*h.Page`** — a full HTML document. Created with `h.NewPage(root Ren)`. Used for normal navigation / first-load responses.

**`*h.Partial`** — an HTML fragment for htmx swaps. Created with `h.NewPartial(root Ren)` or `h.NewPartialWithHeaders(root Ren, headers...)` when you need to set response headers (`HX-Retarget`, `HX-Trigger`, etc.).

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
- `ctx.Redirect(url string)` — 302 redirect; correctly handles `HX-Redirect` for htmx requests.

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

**Setting response headers from a handler** — use `h.NewPartialWithHeaders`:

```go
func PostLogin(ctx *h.RequestContext) *h.Partial {
    if !validCreds(ctx) {
        return h.NewPartialWithHeaders(
            h.Div(h.Class("error"), h.Text("Invalid credentials")),
            h.NewHeader(hx.RetargetHeader, "#login-error"),
        )
    }
    return h.NewPartialWithHeaders(
        h.Div(h.Text("Welcome!")),
        h.NewHeader(hx.RedirectHeader, "/dashboard"),
    )
}
```
````

- [ ] **Step 3.3: Spot-check**

```bash
grep -E "^func (NewPage|NewPartial|NewPartialWithHeaders|NewHeader|GetRequestContext)" framework/h/*.go
grep -E "^func \(.*RequestContext\) (FormValue|QueryParam|UrlParam|IsHxRequest|Redirect|HxSource|HxSourceID|HxRequestType)" framework/h/*.go
```

If any helper is missing or renamed, adjust the skill.

- [ ] **Step 3.4: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): §3 Pages vs Partials and §4 RequestContext

Covers h.NewPage / h.NewPartial / h.NewPartialWithHeaders, the file-path
auto-routing rules (including function-name method prefixes), and the
RequestContext surface (FormValue, QueryParam, IsHxRequest, Fiber
escape hatch, response headers via NewPartialWithHeaders).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: §5 htmx integration via `hx/` + `h.Hx*` helpers

**Files:**
- Modify: `.claude/skills/htmgo-guidance/SKILL.md`

This is the largest single section (~150 lines). Biggest emphasis: htmx 4's explicit inheritance model (`:inherited`), which is the most common source of bugs when upgrading.

- [ ] **Step 4.1: Survey**

```bash
grep -n "^const\|^func\|^type" framework/hx/htmx.go | head -50
grep -n "^func Hx" framework/h/attribute.go
grep -n "Inherited" framework/h/attribute.go
grep -n "^func .*Trigger" framework/h/attribute.go framework/hx/trigger.go
```

Verify: all `HxGet`, `HxPost`, `HxPut`, `HxPatch`, `HxDelete` exist; `HxTarget`, `HxInclude`, `HxSwap`, `HxTrigger` exist; all `*Inherited` variants present; `hx.TriggerEvent` struct exists in `framework/hx/trigger.go`.

- [ ] **Step 4.2: Append §5**

Append to `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown

## 5. htmx integration (`hx/` package + `h.Hx*` helpers)

### The `hx/` package — constants only

`framework/hx/` is a pure constants package: attribute names, header names, event names (htmx 4 colon form), swap types. Import when you need a string value:

```go
import "github.com/franchb/htmgo/framework/hx"

h.Attribute("hx-swap:inherited", string(hx.SwapTypeInnerHtml))
```

**Attribute name constants** (partial — see `framework/hx/htmx.go`):
`GetAttr`, `PostAttr`, `PutAttr`, `PatchAttr`, `DeleteAttr`, `TargetAttr`, `TriggerAttr`, `SwapAttr`, `SwapOobAttr`, `SelectAttr`, `SelectOobAttr`, `IncludeAttr`, `IndicatorAttr`, `ConfirmAttr`, `BoostAttr`, `PushUrlAttr`, `ReplaceUrlAttr`, `ValsAttr`, `ValidateAttr`, `HeadersAttr`, `EncodingAttr`, `PreserveAttr`, `SyncAttr`, `DisableAttr` (form-disable during in-flight; htmx 4 semantics), `IgnoreAttr` (stop htmx processing; migrated role in htmx 4), `ConfigAttr`, `StatusAttr`.

**Event name constants** (htmx 4 colon form):
`AfterSwapEvent = "htmx:after:swap"`, `BeforeSwapEvent = "htmx:before:swap"`, `AfterRequestEvent`, `BeforeRequestEvent`, `AfterOnLoadEvent = "htmx:after:init"`, `BeforeOnLoadEvent = "htmx:before:init"`, `ConfigRequestEvent`, `ErrorEvent = "htmx:error"`, etc.

**Swap type constants:** `SwapTypeInnerHtml`, `SwapTypeOuterHtml`, `SwapTypeBeforeBegin`, `SwapTypeAfterBegin`, `SwapTypeBeforeEnd`, `SwapTypeAfterEnd`, `SwapTypeTextContent`, `SwapTypeDelete`, `SwapTypeNone`, `SwapTypeTrue`.

**Header name constants:** `RequestHeader = "HX-Request"`, `RedirectHeader = "HX-Redirect"`, `RetargetHeader`, `ReswapHeader`, `ReselectHeader`, `TriggerHeader`, `PushUrlHeader`, `ReplaceUrlHeader`, `LocationHeader`, `SourceHeader = "HX-Source"` (htmx 4), `RequestTypeHeader = "HX-Request-Type"` (htmx 4), `CurrentUrlHeader`.

### The `h.Hx*` builders — return `Ren`

Use these for common htmx attributes (instead of raw `h.Attribute(hx.TargetAttr, "...")`):

```go
h.HxGet("/search")                    // hx-get="/search"
h.HxPost("/users/create")             // hx-post="/users/create"
h.HxPut("/items/123")
h.HxPatch("/items/123")
h.HxDelete("/items/123")
h.HxTarget("#results")                // hx-target="#results"
h.HxSwap("innerHTML")                 // hx-swap="innerHTML"
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
    h.HxTarget("#results"),  // ← does NOT cascade in htmx 4
    h.Button(h.HxGet("/a"), h.Text("A")),
    h.Button(h.HxGet("/b"), h.Text("B")),
)
```

**Right (htmx 4):**

```go
h.Div(
    h.HxTargetInherited("#results"),  // ← cascades to descendants
    h.Button(h.HxGet("/a"), h.Text("A")),
    h.Button(h.HxGet("/b"), h.Text("B")),
)
```

### Triggers

Simple raw-string form:

```go
h.HxTriggerString("click once, keyup[key=='Escape'] from:body")
```

Structured form (type-checked against `hx.TriggerEvent`):

```go
h.HxTrigger(
    hx.NewTrigger("click", hx.Once()),
    hx.NewTrigger("keyup", hx.KeyEquals("Escape"), hx.From("body")),
)
```

Convenience for click with modifiers:

```go
h.HxTriggerClick(hx.Once(), hx.DelayMs(300))
```

### Response headers

Set headers on a partial response via `h.NewPartialWithHeaders`:

```go
h.NewPartialWithHeaders(
    h.Div(h.Text("Saved")),
    h.NewHeader(hx.TriggerHeader, "showToast"),          // fires a client event
    h.NewHeader(hx.RetargetHeader, "#status"),           // changes swap target
    h.NewHeader(hx.ReswapHeader, "afterend"),            // changes swap style
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
            h.HxPost("/search"),
            h.HxTriggerString("keyup changed delay:300ms, search"),
            h.HxTarget("#search-results"),
            h.HxSwap(hx.SwapTypeInnerHtml),
        ),
        h.Div(h.Id("search-results")),
    )
}
```

### See also

For general htmx patterns (attribute semantics, event lifecycle, OOB swaps, SSE/WS), the `htmx-guidance` skill covers the htmx side in depth. This skill focuses on the Go builder layer; the two are complementary.
````

- [ ] **Step 4.3: Spot-check**

```bash
grep -E "^const (Get|Post|Put|Patch|Delete|Target|Trigger|Swap|Redirect|Retarget|Reswap)Attr" framework/hx/htmx.go
grep -E "^(const|func) (AfterSwap|BeforeSwap|Error|Location)Event" framework/hx/htmx.go
grep -E "^func Hx(Target|Include|Swap|Boost|Confirm|Headers|Indicator|Sync|Encoding|Validate)Inherited" framework/h/attribute.go
```

- [ ] **Step 4.4: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): §5 htmx integration via hx/ and h.Hx* helpers

Covers the hx/ constants (attrs, events, headers, swap types), the full
h.Hx* builder set, the htmx 4 explicit-inheritance model with concrete
:inherited examples, trigger specs, response-header usage via
NewPartialWithHeaders, and a worked search-box example. Points readers at
the htmx-guidance skill for htmx-level patterns.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: §6 Alpine integration via `ax/` + bundled compat extension

**Files:**
- Modify: `.claude/skills/htmgo-guidance/SKILL.md`

Second-largest section (~130 lines). Most important content: the decision rubric (Alpine vs htmx vs pure JS) and the fact that the compat extension ships automatically.

- [ ] **Step 5.1: Survey**

```bash
grep -n "^const\|^func\|^type" framework/ax/alpine.go framework/ax/builder.go
```

Verify all 18 constants and ~30 builders listed in the spec exist with the exact names.

- [ ] **Step 5.2: Append §6**

Append to `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown

## 6. Alpine integration (`ax/` package + bundled `alpine-compat`)

Alpine.js is **optional**. When you need small client-side state (open/closed, theme, density, copy-to-clipboard, ⌘K modals), reach for Alpine. For server round-trips, reach for htmx. For imperative one-shots, reach for the lifecycle command DSL (§7).

### Setup (two lines of HTML)

1. Add the Alpine CDN script to your layout:

```go
h.Head(
    h.Tag("script",
        h.Attribute("src", "https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js"),
        h.Attribute("defer", ""),
    ),
    h.Script("/public/htmgo.js"),
)
```

The `defer` attribute is important — Alpine initializes on `DOMContentLoaded`. `htmgo.js` can load before or after; extensions self-register in htmx 4.

2. Add the FOUC-prevention CSS rule to your stylesheet:

```css
[x-cloak] { display: none !important; }
```

The `alpine-compat` htmx extension is **already bundled** in `/public/htmgo.js` — no additional script, no `hx-ext` attribute. It auto-gates on `window.Alpine` presence, so if Alpine isn't loaded, the extension no-ops.

### The `ax/` package — constants + builders

`framework/ax/` mirrors the `hx/` shape: constants for directive names, plus Ren-returning builders.

**Constants** (all attribute name strings): `DataAttr`, `InitAttr`, `ShowAttr`, `BindAttr`, `OnAttr`, `TextAttr`, `HtmlAttr`, `ModelAttr`, `ModelableAttr`, `CloakAttr`, `RefAttr`, `IgnoreAttr`, `TeleportAttr`, `EffectAttr`, `IfAttr`, `ForAttr`, `IdAttr`, `TransitionAttr`.

**Simple single-arg directives** (Alpine expression as the value):

```go
ax.Data("{ open: false, count: 0 }")  // x-data="..."
ax.Init("count = 10")                  // x-init="..."
ax.Show("open")                        // x-show="..."
ax.Text("message")                     // x-text="..."
ax.Html("markup")                      // x-html="..."
ax.Model("query")                      // x-model="..."
ax.Effect("console.log(count)")        // x-effect="..."
ax.If("visible")                       // x-if="..." (template only)
ax.For("item in items")                // x-for="..."
ax.Id("['tab']")                       // x-id="..."
ax.Ref("input")                        // x-ref="input"
ax.Teleport("body")                    // x-teleport="body"
ax.Modelable("value")                  // x-modelable="value"
```

**No-arg directives** (emit bare attributes; Alpine accepts the boolean form):

```go
ax.Cloak()       // x-cloak
ax.Ignore()      // x-ignore
ax.Transition()  // x-transition
```

**`x-bind:*` family** — bind any attribute to an expression:

```go
ax.Bind("data-foo", "value")  // generic: x-bind:data-foo="value"

// Shortcuts for the most common targets:
ax.BindClass("{ active: isActive }")
ax.BindStyle("{ color: hex }")
ax.BindHref("url")
ax.BindValue("input")
ax.BindDisabled("locked")
ax.BindChecked("selected")
ax.BindId("compId")
```

**`x-on:*` family** — event handlers with variadic modifiers:

```go
ax.On("click", "count++")                        // x-on:click="count++"
ax.On("click", "submit()", "prevent")            // x-on:click.prevent="submit()"
ax.On("keydown", "handle()", "meta", "k", "prevent")
// emits: x-on:keydown.meta.k.prevent="handle()"

// Event shortcuts (each takes variadic modifiers):
ax.OnClick(handler, mods...)
ax.OnSubmit(handler, mods...)
ax.OnInput(handler, mods...)
ax.OnChange(handler, mods...)
ax.OnFocus(handler, mods...)
ax.OnBlur(handler, mods...)
ax.OnKeydown(handler, mods...)
ax.OnKeyup(handler, mods...)

// Combo shortcuts for the three most common patterns:
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

### When to use `ax.*` vs htmx swaps vs pure JS

| Use case | Tool |
|---|---|
| Pure-client state that never needs the server (theme toggle, popover open/close, keyboard overlay, copy-to-clipboard) | `ax.*` |
| Server round-trip required (load data, save form, refresh a list) | htmx (`h.HxGet`, etc.) |
| Widget with BOTH client state AND server content (popover whose body is server-loaded) | `ax.Data(...)` on the outer wrapper + `h.HxGet(...)` on inner elements. The alpine-compat extension carries `_x_dataStack` across the morph. |
| Imperative one-shots without state (focus input, scroll to element, set a value) | Lifecycle commands (§7) — `h.OnClick(js.Focus(...))`, not Alpine |

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
            h.HxGet("/popover/content"),
            h.HxTrigger("intersect once"),
            h.HxSwap(hx.SwapTypeInnerHtml),
            h.Text("Loading…"),
        ),
    )
}
```

This composes all three layers: Alpine manages `open`, htmx lazy-loads the content, and the alpine-compat extension preserves Alpine state through the swap automatically.

### Gotchas

- **Alpine v3 only** — v2 is unsupported upstream; the compat extension uses v3 internal APIs (`closestDataStack`, `cloneNode`, etc.).
- **Load Alpine with `defer` in `<head>`.** Without defer, Alpine may init before `[x-data]` elements are parsed.
- **Alpine plugins load before Alpine itself.** Per upstream plugin docs — `@alpinejs/persist`, `@alpinejs/intersect`, `@alpinejs/focus`, etc. go in `<script>` tags ABOVE the main Alpine script.
- **`[x-cloak]` CSS rule required** to prevent FOUC. Without it, Alpine-hidden elements flash visible on page load.
- **Don't outer-swap an Alpine root element** with htmx. Morphs of inner content preserve `_x_dataStack`; full replacement of the `[x-data]` root loses state. Either swap inner content only, or re-attach state via `ax.Data(...)` on the new root.
````

- [ ] **Step 5.3: Spot-check**

```bash
grep -E "^func (Data|Init|Show|Text|Html|Model|Modelable|Cloak|Ignore|Transition|Bind|On|Ref|Teleport|If|For|Id|Effect)" framework/ax/builder.go
grep -E "^func (BindClass|BindStyle|BindHref|BindValue|BindDisabled|BindChecked|BindId)" framework/ax/builder.go
grep -E "^func (OnClick|OnSubmit|OnInput|OnChange|OnFocus|OnBlur|OnKeydown|OnKeyup|OnClickOutside|OnKeydownEscape|OnKeydownEnter)" framework/ax/builder.go
grep -E "^func (ModelNumber|ModelLazy|ModelTrim|ModelFill|ModelBoolean|ModelDebounce)" framework/ax/builder.go
```

Every name listed in the skill must appear in the grep output.

- [ ] **Step 5.4: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): §6 Alpine integration via ax/ and bundled alpine-compat

Covers Alpine setup (CDN script + x-cloak CSS), the complete ax package
surface (constants, simple/no-arg/Bind/On/Model builders), the decision
rubric for ax vs htmx vs lifecycle commands, a worked popover example
combining all three layers, and a gotcha list (v3 only, defer, plugin
ordering, outer-swap interaction).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: §7 Lifecycle & commands + §8 Service locator

Two mid-weight sections grouped into one commit.

**Files:**
- Modify: `.claude/skills/htmgo-guidance/SKILL.md`

- [ ] **Step 6.1: Survey**

```bash
grep -n "^func " framework/h/lifecycle.go | head -20
grep -n "^func \|^type" framework/service/*.go
ls framework/js/*.go
grep -n "^func " framework/js/*.go | head -20
```

Note the actual command names (e.g. `js.Alert`, `js.SetValue`, `js.SetInnerHTML`, `js.AddClass`, `js.EvalJs`). If the package has renames, use the real names.

- [ ] **Step 6.2: Append §7 + §8**

Append to `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown

## 7. Lifecycle & JS command DSL

Not every interaction needs Alpine or a server round-trip. For **imperative one-shots** (fire an alert, set a value, add a class, redirect), htmgo provides event helpers in `framework/h/` that emit generated JavaScript at render time.

### Event helpers

Bind a JS command to a standard DOM event:

```go
h.Button(
    h.OnClick(js.Alert("Hello!")),
    h.Text("Say hi"),
)
```

Available helpers (see `framework/h/lifecycle.go`):
- `h.OnClick(cmd...)`, `h.OnLoad(cmd...)`, `h.OnSubmit(cmd...)`, `h.OnChange(cmd...)`, `h.OnInput(cmd...)`, `h.OnFocus(cmd...)`, `h.OnBlur(cmd...)`, `h.OnMouseOver(cmd...)`, `h.OnMouseOut(cmd...)`, `h.OnKeyUp(cmd...)`, `h.OnKeyDown(cmd...)`.
- htmx lifecycle hooks: `h.HxOnAfterSwap(cmd...)`, `h.HxBeforeRequest(cmd...)`, `h.HxAfterRequest(cmd...)`, `h.HxOnConfirm(cmd...)`.
- SSE hooks: `h.HxBeforeSseMessage(cmd...)`, `h.HxAfterSseMessage(cmd...)`, `h.HxOnSseError(cmd...)`, `h.HxOnSseClose(cmd...)`, `h.HxOnSseConnecting(cmd...)`, `h.HxOnSseOpen(cmd...)`.

### JS commands (`framework/js/`)

Each helper takes one or more `js.Command` values. Common commands:

```go
js.Alert("Message")                         // window.alert(...)
js.SetValue("#input", "new value")          // document.querySelector.value = ...
js.SetInnerHTML("#target", "<b>x</b>")      // innerHTML =
js.SetInnerText("#target", "hello")         // innerText =
js.AddClass("#x", "active")
js.RemoveClass("#x", "hidden")
js.ToggleClass("#x", "open")
js.Redirect("/home")                        // window.location = ...
js.Focus("#search-input")
js.EvalJs("customFunction()")               // raw JS escape hatch
```

Commands chain automatically — pass multiple to one handler:

```go
h.OnClick(
    js.AddClass("#panel", "fade-out"),
    js.EvalJs("setTimeout(() => panel.remove(), 300)"),
)
```

### Two command shapes

- `SimpleJsCommand` — single statement; inlined into the `onclick="..."` attribute value.
- `ComplexJsCommand` — multi-statement block; htmgo emits an `__eval_<id>()` helper script and wires the onclick to invoke it. This is how `js.EvalJs(complex)` works internally.

Both satisfy the `Command` interface.

### Contrast with Alpine's `ax.OnClick`

Three superficially-similar things do different jobs — picking the right one matters.

| Expression | Mechanism | When to use |
|---|---|---|
| `h.OnClick(js.Alert("hi"))` | Inline `onclick="__eval_xxx(this)"` running htmgo-generated JS | Imperative one-shot with no state and no server call |
| `ax.OnClick("open = !open")` | `x-on:click="..."` evaluated by Alpine at runtime against the component's `x-data` scope | Mutate Alpine state |
| `h.Button(h.HxGet("/x"))` | htmx request (button click triggers a GET) | Server round-trip |

**Common mistake:** putting `h.OnClick(js.X)` AND `ax.OnClick("...")` on the same element. Both fire. Pick one per element, or use separate nested elements if you genuinely need both.

## 8. Service locator (DI)

For dependencies that handlers need (DB connections, config, external API clients), use `framework/service/`.

### Setup in `main.go`

```go
package main

import (
    "github.com/franchb/htmgo/framework/h"
    "github.com/franchb/htmgo/framework/service"
)

func main() {
    locator := service.NewLocator()
    service.Set[*sql.DB](locator, service.Singleton, func() *sql.DB {
        return openDB()
    })
    service.Set[*Config](locator, service.Singleton, loadConfig)

    h.Start(h.AppOpts{
        ServiceLocator: locator,
        Register:       __htmgo.Register,  // generated by htmgo
    })
}
```

### Resolving in handlers

```go
func GetUsers(ctx *h.RequestContext) *h.Partial {
    db := service.Get[*sql.DB](ctx.ServiceLocator())
    // ... use db ...
}
```

### Lifecycles

- `service.Singleton` — provider runs once; same instance for every `Get`.
- `service.Transient` — provider runs on every `Get`; returns a fresh instance.

Pick `Singleton` for pools, long-lived clients, config. Pick `Transient` for per-request scoped state (though request scope is better handled via context values than the locator).

### Worked example — DB handle + config

```go
// main.go
locator := service.NewLocator()
service.Set[*Config](locator, service.Singleton, func() *Config {
    return mustLoad("config.yaml")
})
service.Set[*sql.DB](locator, service.Singleton, func() *sql.DB {
    cfg := service.Get[*Config](locator)
    return sql.MustOpen(cfg.DatabaseURL)
})

// partials/search/query.go
func GetQuery(ctx *h.RequestContext) *h.Partial {
    db := service.Get[*sql.DB](ctx.ServiceLocator())
    q := ctx.QueryParam("q")
    rows := queryRows(db, q)
    return h.NewPartial(renderResults(rows))
}
```
````

- [ ] **Step 6.3: Spot-check**

```bash
grep -E "^func (OnClick|OnSubmit|OnInput|OnChange|OnFocus|OnBlur|OnMouseOver|OnMouseOut|OnKeyUp|OnKeyDown|OnLoad|HxOnAfterSwap|HxBeforeRequest|HxAfterRequest|HxOnConfirm)" framework/h/lifecycle.go
grep -E "^func (Alert|SetValue|SetInnerHTML|SetInnerText|AddClass|RemoveClass|ToggleClass|Redirect|Focus|EvalJs)" framework/js/*.go
grep -E "^func (NewLocator|Set|Get)" framework/service/*.go
grep -E "^(const|var) (Singleton|Transient)" framework/service/*.go
```

Adjust any names that don't match.

- [ ] **Step 6.4: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): §7 lifecycle & JS commands and §8 service locator

Covers h.On* / h.HxOn* event helpers, the js.* command DSL, the
SimpleJsCommand vs ComplexJsCommand shapes, and the critical distinction
between h.OnClick (imperative JS), ax.OnClick (Alpine), and h.HxGet
(server round-trip). Adds service locator usage with Singleton and
Transient lifecycles and a DB+config worked example.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: §9 Caching + §10 Project config & CLI

**Files:**
- Modify: `.claude/skills/htmgo-guidance/SKILL.md`

- [ ] **Step 7.1: Survey**

```bash
grep -n "^func\|^type" framework/h/cache.go framework/h/cache/*.go
cat framework/h/cache/lru_store_example.go | head -30
grep -n "^func\|^type" framework/config/*.go
```

- [ ] **Step 7.2: Append §9 + §10**

Append to `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown

## 9. Caching

### High-level helper

For caching rendered fragments, use `h.Cache`:

```go
h.Cache(ctx, "user:"+userID, 5*time.Minute, func() h.Ren {
    user := loadUser(userID)
    return renderUserProfile(user)
})
```

Signature (roughly): `h.Cache(ctx *h.RequestContext, key string, ttl time.Duration, compute func() h.Ren) h.Ren`. The rendered HTML is cached for `ttl`. Subsequent calls within the window return the cached fragment without re-invoking `compute`.

Global key form: `h.CacheGlobal(key, ttl, compute)` — not scoped to a request.

### Low-level pluggable stores

`framework/h/cache/` defines:

```go
type Store[K comparable, V any] interface {
    Set(key K, value V)
    GetOrCompute(key K, compute func() V) V
    Delete(key K)
    Purge()
    Close() error
}
```

**Built-in:** `cache.TTLStore` — simple time-based eviction. Use for most cases.

**Example:** `cache.LRUStore` is provided as an example custom implementation in `framework/h/cache/lru_store_example.go`. Follow its shape to build your own (e.g. a Redis-backed store).

### When to cache

- Expensive DB reads.
- Rendered fragments shown on many pages (navigation, sidebar, footer).
- Third-party API responses with their own rate limits.

### When NOT to cache

- User-specific content without a user-scoped key (risks leaking one user's view to another).
- Rapidly changing data with strict freshness needs.
- Content cheap to compute (caching adds overhead; measure before reaching for it).

## 10. Project configuration & CLI

### `htmgo.yml` (or `htmgo.yaml` / `_htmgo.yaml` / `_htmgo.yml`)

Lives in your app's root directory. Fields:

```yaml
tailwind: true              # run Tailwind CSS compilation during build
tailwind_version: "4"       # optional; auto-detected

watch_ignore:               # glob patterns the dev-mode watcher ignores
  - "**/*.test.go"
  - "**/vendor/**"

watch_files:                # additional extensions the watcher rebuilds on
  - "**/*.md"

automatic_page_routing_ignore:
  - "**/_shared.go"

automatic_partial_routing_ignore:
  - "**/internal/*.go"

public_asset_path: "/public"  # URL prefix for static assets (default: /public)
```

### CLI commands

Install:

```bash
cd /path/to/htmgo/cli/htmgo && go install .
# or: go run github.com/franchb/htmgo/cli/htmgo@latest <subcommand>
```

Subcommands:

- `htmgo setup` — scaffold a new app in the current directory.
- `htmgo build` — compile the app + regenerate `__htmgo/`, produce a `dist/<app>` binary.
- `htmgo watch` / `htmgo dev` — live-reload dev server. Rebuilds on `.go`, `.css`, `.md` changes.
- `htmgo generate` — regenerate `__htmgo/pages-generated.go` + `__htmgo/partials-generated.go` without a full build.
- `htmgo css` — Tailwind build only.
- `htmgo format` — `go fmt ./...` + import cleanup.
- `htmgo template` — create an app from a starter template.
- `htmgo version` — print CLI version.

### `__htmgo/` — generated files, never edit

The CLI writes `__htmgo/pages-generated.go`, `__htmgo/partials-generated.go`, and `__htmgo/setup-generated.go` containing auto-registered routes. These files are **fully regenerated** on every `htmgo build` / `watch` / `generate`. Manual edits are overwritten.

Add `__htmgo/` to `.gitignore`:

```gitignore
__htmgo/
```

### Dev workflow via `task`

Example apps include a `Taskfile.yml` wrapping the CLI:

```bash
task watch   # live-reload dev
task build   # production build
task run     # run the built binary
```

Copy the shape from `examples/todo-list/Taskfile.yml` when starting a new app.
````

- [ ] **Step 7.3: Spot-check**

```bash
grep -E "^func (Cache|CacheGlobal)" framework/h/cache.go
grep -E "^type (Store|TTLStore)" framework/h/cache/*.go
```

- [ ] **Step 7.4: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): §9 caching and §10 project config & CLI

Covers h.Cache / h.CacheGlobal high-level API, the Store interface with
TTLStore built-in and LRUStore example, the when/when-not cache decision
table, htmgo.yml fields, all CLI subcommands, and the rule that
__htmgo/ is generated and must not be hand-edited.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: §11 Common pitfalls

**Files:**
- Modify: `.claude/skills/htmgo-guidance/SKILL.md`

Pattern-of-errors section with concrete symptom → fix pairs.

- [ ] **Step 8.1: Append §11**

Append to `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown

## 11. Common pitfalls

### Missing `:inherited` after htmx 4 upgrade

**Symptom:** Clicks on child elements no longer hit the expected swap target. No console error; htmx just uses the default (usually the clicked element itself).

**Fix:** On the ancestor element that sets the attribute, change `h.HxTarget("#x")` → `h.HxTargetInherited("#x")`. Same pattern for `include`, `swap`, `boost`, `confirm`, `headers`, `indicator`, `sync`, `encoding`, `validate`.

### Wrong return type from a route handler

**Symptom:** Build fails with a type mismatch in `__htmgo/pages-generated.go` or `__htmgo/partials-generated.go`.

**Fix:** Page routes (file in `pages/`) must return `*h.Page`. Partial routes (file in `partials/`) must return `*h.Partial`. If you need to fall back from htmx to full-page render inside one route, put the branch in the page handler and call `h.NewPartial` inside a helper that returns `Ren` when the htmx case applies.

### Alpine loaded AFTER the first swap

**Symptom:** First page load works, but after the first htmx swap nothing Alpine-driven responds. Console shows `x-data` attributes without their expected state.

**Fix:** Put the Alpine `<script>` in `<head>` with `defer`. If a page triggers `hx-trigger="load"` before `DOMContentLoaded` fires, that first swap happens pre-Alpine; in that case wait for `Alpine.initialized` or use `hx-trigger="load delay:50ms"` to let Alpine boot.

### Outer-swapping an Alpine root element

**Symptom:** Alpine state (counter, toggle, etc.) resets every time the htmx swap lands.

**Fix:** Keep `ax.Data(...)` on an element OUTSIDE the swap target. The alpine-compat extension carries `_x_dataStack` across morphs of descendants, but full replacement of the `x-data` root loses state. Either use `hx-swap="innerHTML"` so the root stays put, or re-declare the state on the new root.

### `h.OnClick(js.Alert(...))` AND `ax.OnClick("...")` on the same element

**Symptom:** Both handlers fire on click.

**Fix:** Pick one mechanism per element. If you need both a server round-trip AND Alpine state mutation, use nested elements: Alpine handler on the outer, `HxPost` on the inner.

### Editing `__htmgo/*-generated.go` by hand

**Symptom:** Your edits disappear after the next `htmgo build` / `watch` / `generate`.

**Fix:** Edit the source file (the `pages/foo.go` or `partials/foo.go`) instead. The generated file is rebuilt from the source every time.

### Forgetting `x-cloak` CSS

**Symptom:** Alpine components briefly flash visible before `x-show="false"` hides them (FOUC).

**Fix:** Add `[x-cloak] { display: none !important; }` to your site stylesheet and apply `ax.Cloak()` to every Alpine-hidden element.

### Alpine plugin load order

**Symptom:** Plugin directives (`x-persist`, `x-intersect`, etc.) don't work; console warns about unknown directives.

**Fix:** Load the plugin script BEFORE the main Alpine script:

```go
h.Tag("script", h.Attribute("src", "https://unpkg.com/@alpinejs/persist@3.x"), h.Attribute("defer", "")),
h.Tag("script", h.Attribute("src", "https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js"), h.Attribute("defer", "")),
```

Alpine v3 requires plugins to register before it initializes.

### Calling `h.Render(el)` in request handlers

**Symptom:** Handler returns a plain string; page metadata (title, meta tags) is missing; htmx response headers don't make it through.

**Fix:** Don't manually render. Return `h.NewPage(...)` or `h.NewPartial(...)` and let the framework render. `h.Render` is for tests and debugging only.

### Mutating `ctx` state

**Symptom:** Occasional cross-request data leaks; tests flake depending on order.

**Fix:** Treat `RequestContext` as read-only from handler code. If you need to stash something per-request, use `ctx.Fiber.Locals(key, value)` (Fiber-provided, request-scoped).
````

- [ ] **Step 8.2: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): §11 common pitfalls — symptom → fix pairs

Eight concrete error patterns: missing :inherited on htmx 4, wrong
return type from a route, Alpine load order, outer-swapping an Alpine
root, double handlers (h.OnClick + ax.OnClick), editing
__htmgo/*-generated.go, missing x-cloak CSS, Alpine plugin load order,
manual h.Render calls, and mutating ctx.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: §12 Upgrade pointers + version stamp polish + line-count check

**Files:**
- Modify: `.claude/skills/htmgo-guidance/SKILL.md`

- [ ] **Step 9.1: Append §12**

Append to `.claude/skills/htmgo-guidance/SKILL.md`:

````markdown

## 12. Upgrade pointers

This section is intentionally short — it points at changelog entries and other docs for detailed migration steps.

### htmx 2 → htmx 4 within htmgo (this fork)

Done in commit `1.2.0-beta.1` (2026-04-17). Key breaking changes documented in `CHANGELOG.md`:
- **Explicit inheritance** — `hx-target` and siblings no longer cascade implicitly. Use `HxTargetInherited` etc. (see §5).
- **`hx-ext` removed** — extensions self-register on script import; remove any `h.HxExtension(...)` calls.
- **`hx-disable` semantics flipped** — now disables form elements during in-flight requests. For the old "stop htmx processing" role, use `hx.IgnoreAttr`.
- **Event names in colon form** — `htmx:afterSwap` → `htmx:after:swap`. Use `hx/` constants instead of raw strings.
- **Extensions rewritten** for the new `registerExtension` API; `detail.xhr` → `detail.ctx`.

Full migration recipe + `grep` command for finding inheritance parents: see the `[1.2.0-beta.1]` changelog entry.

### chi → Fiber v3

Also shipped in `1.2.0-beta.1`. Handler signature changed from `http.Handler` / `http.HandlerFunc` to `fiber.Ctx`-based:

**Before (chi):**
```go
func myMiddleware(next http.Handler) http.Handler { ... }
```

**After (Fiber v3):**
```go
func myMiddleware(c fiber.Ctx) error { ... }
```

Inside handlers, the raw Fiber context is at `ctx.Fiber`.

### `maddalax/htmgo` → `franchb/htmgo`

This fork's import path is `github.com/franchb/htmgo/framework`. APIs are close to upstream **except** for the htmx 4 + Fiber changes above. To move a project from upstream:

1. Replace `github.com/maddalax/htmgo/framework` → `github.com/franchb/htmgo/framework` in all Go imports and in `go.mod`.
2. Apply the htmx 4 migration recipe (inheritance, event names, extension imports).
3. Apply the chi → Fiber signature changes in middleware.
4. Run `htmgo generate` + `go build ./...` to surface any remaining breakage.

### Alpine adoption (new in v1.2.0-beta.2)

The `alpine-compat` htmx extension is bundled automatically in `htmgo.js` from v1.2.0-beta.2 onward. To adopt Alpine in an existing htmgo app:

1. Add the Alpine CDN script to your layout (see §6).
2. Add the `[x-cloak] { display: none !important; }` CSS rule.
3. Start using `ax.*` helpers on elements where you want client state.

No framework dep changes needed beyond bumping the htmgo version.

### See also

- `CHANGELOG.md` — authoritative version-by-version changes.
- `docs/superpowers/specs/2026-04-17-htmx-v4-migration-design.md` — htmx 4 migration design rationale.
- `docs/superpowers/specs/2026-04-18-alpine-compat-design.md` — Alpine bundling design rationale.
- `htmx-upgrade-from-htmx2` skill (separate) — step-by-step htmx 2 → 4 upgrade workflow.
- `htmx-migration` skill (separate) — general SPA → htmx migration patterns.
````

- [ ] **Step 9.2: Verify line count and line-count budget**

Run:

```bash
wc -l .claude/skills/htmgo-guidance/SKILL.md
```

Expected: somewhere between 700 and 850. Hard cap: 900.

If over 900: trim the longest sections (usually §5 htmx or §6 Alpine). Preserve concrete API references; drop redundant prose. Re-run `wc -l` after trimming.

If under 700: the skill is likely under-detailed. Review the spec and add missing content. Most common gap: missing worked examples in §2–§4.

- [ ] **Step 9.3: Update the version stamp if framework versions have changed**

Confirm the version stamp in the skill (line ~6) is still accurate. If `framework/assets/js/package.json` has bumped since drafting began, update the Alpine/htmx version in the stamp.

```bash
grep "htmx.org" framework/assets/js/package.json
```

The stamp currently reads `htmx 4.0.0-beta2, Alpine.js 3.15.11`. Match the actual pinned htmx version.

- [ ] **Step 9.4: Read-through check**

Read the entire SKILL.md from top to bottom. Catch:
- Dangling `TODO` / `TBD` / `FIXME`.
- References to functions that don't exist (`grep -E "^func <name>" framework/...` for each — spot-check a sampling).
- Contradictions (e.g. one section says "use x-cloak CSS rule" and another says "don't worry about FOUC" — the first is right).
- Obsolete content (e.g. mentions of `http.HandlerFunc` outside the §12 Upgrade pointers context).

Fix anything found.

- [ ] **Step 9.5: Commit**

```bash
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): §12 upgrade pointers and final polish

Adds the short §12 upgrade pointers section (htmx 2→4 within htmgo,
chi→Fiber v3, maddalax→franchb fork switch, Alpine adoption). Verifies
the skill fits within the 900-line budget and does a full read-through
for consistency and accurate API references.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 10: Update `README.md` with install instructions

**Files:**
- Modify: `README.md`

- [ ] **Step 10.1: Append the "Claude Code skills" subsection**

Edit `README.md`. Add this section BEFORE the `## Star History` heading (which is currently at line 46):

```markdown
## Claude Code skills

This repo ships [Claude Code](https://docs.claude.com/en/docs/claude-code/overview) skills under `.claude/skills/` that teach AI coding sessions how to work with htmgo:

- `htmgo-guidance` — writing htmgo apps (builder, pages, partials, hx/ax helpers, caching, CLI)
- `htmx-guidance` — htmx 4 patterns and best practices
- `htmx-debugging`, `htmx-extension-authoring`, `htmx-migration`, `htmx-upgrade-from-htmx2` — specialized htmx skills

To use the `htmgo-guidance` skill in a project that consumes this fork:

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
```

- [ ] **Step 10.2: Verify the section renders correctly**

```bash
head -70 README.md
```

Check: the new section appears between the "get started" block and "Star History". No broken Markdown (fenced code blocks matched).

- [ ] **Step 10.3: Commit**

```bash
git add README.md
git commit -m "$(cat <<'EOF'
docs: add Claude Code skills section to README

Documents the .claude/skills/ directory, lists the shipped skills
(htmgo-guidance plus the existing htmx-* set), and provides copy-paste
install commands for downstream consumer projects.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 11: Copy the skill into `vulnerability_catalog`

**Files:**
- Create: `/home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog/.claude/skills/htmgo-guidance/SKILL.md`

This commit lands in a DIFFERENT repo (`vulnerability_catalog`, not this fork). Be careful not to commit it in the htmgo repo.

- [ ] **Step 11.1: Copy the skill**

```bash
cp -r /home/iru/p/github.com/franchb/htmgo/.claude/skills/htmgo-guidance \
      /home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog/.claude/skills/
```

- [ ] **Step 11.2: Verify the copy**

```bash
ls /home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog/.claude/skills/
# expected: htmx-debugging, htmx-extension-authoring, htmx-guidance, htmx-migration, htmx-upgrade-from-htmx2, tailwind-best-practices, htmgo-guidance

diff -r /home/iru/p/github.com/franchb/htmgo/.claude/skills/htmgo-guidance \
        /home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog/.claude/skills/htmgo-guidance
# expected: no output (identical)
```

- [ ] **Step 11.3: Check vuln_catalog's working tree before staging**

```bash
cd /home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog && git status --short
```

Expected: the new `.claude/skills/htmgo-guidance/SKILL.md` appears as untracked, alongside whatever unrelated working-tree changes already exist (e.g. `.beads/` setup from a prior session). Do not stage or commit unrelated changes — only the new skill file.

- [ ] **Step 11.4: Stage + commit in vuln_catalog (NOT in htmgo)**

```bash
cd /home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog
git add .claude/skills/htmgo-guidance/SKILL.md
git commit -m "$(cat <<'EOF'
docs(skills): add htmgo-guidance skill from franchb/htmgo

Copies the htmgo-guidance Claude Code skill from the upstream
franchb/htmgo repo (v1.2.0-beta.2, htmx4-migration branch) so sessions
working in this project auto-load htmgo patterns, builder conventions,
and the new ax/ Alpine helpers.

To refresh later, re-run:
  cp -r /path/to/franchb/htmgo/.claude/skills/htmgo-guidance \
        .claude/skills/

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 11.5: Return to the htmgo working directory**

```bash
cd /home/iru/p/github.com/franchb/htmgo
```

(This task made no commits in the htmgo repo; the commit landed in vuln_catalog.)

---

## Task 12: Manual smoke test with vuln_catalog (user-driven)

This is a **manual verification step**. No commit. The user opens a Claude Code session at vuln_catalog and exercises the new skill. The agent can prepare the test case and describe what to check, but cannot run a new Claude Code session from inside the current session.

**No file changes.**

- [ ] **Step 12.1: Agent prepares the test case**

Write out a concrete test prompt (200-400 words) the user can paste into a fresh Claude Code session at vuln_catalog. Include a task that exercises the most-commonly-wrong surfaces: `:inherited` attributes, `ax.*` helpers, partial return types, `h.OnClick` vs `ax.OnClick` distinction.

Example test prompt:

> Open Claude Code at `/home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog`. Type `/skills` and confirm that `htmgo-guidance` appears in the list.
>
> Then ask Claude:
>
> *"I want to add a dropdown menu to the top nav in this app. Clicking the avatar should toggle a menu showing 3 links (Profile, Settings, Sign out). Clicking outside the menu or pressing Escape should close it. The menu content should be lazy-loaded from a partial at /nav/menu the first time it opens. Show me the Go code for the page + the partial."*
>
> Check the generated code for:
>
> 1. Uses `ax.Data("{ open: false }")` on the wrapper (not raw `h.Attribute("x-data", ...)`).
> 2. Uses `ax.OnClick("open = !open")` on the avatar button.
> 3. Uses `ax.Show("open")` or similar on the menu container, with `ax.Cloak()` to prevent FOUC.
> 4. Uses `ax.OnClickOutside("open = false")` and `ax.OnKeydownEscape("open = false")`.
> 5. Uses `h.HxGet("/nav/menu")` on the menu (not a literal `x-init` calling fetch).
> 6. Uses `h.HxTrigger("intersect once")` or similar for lazy-load.
> 7. Uses `h.HxTargetInherited` / `h.HxSwapInherited` if the wrapper sets targets for descendants.
> 8. Partial returns `*h.Partial`, uses `GetMenu` / `PostMenu` naming convention.
> 9. Does NOT use htmx 2 attribute names (`hx-ext`, `hx-disable` for "stop processing").
> 10. Does NOT hallucinate functions that don't exist (e.g. `ax.Transition(classes)` which takes no args).

- [ ] **Step 12.2: User runs the session and reports back**

The user runs the Claude Code session with the prompt above and reports to the agent:
- Did `/skills` show `htmgo-guidance`?
- Did the generated code pass all 10 checks?
- What was wrong (if anything)?

- [ ] **Step 12.3: If gaps found, file a follow-up**

If the smoke test surfaces gaps in the skill (e.g. a missing surface, wrong API name, unclear section), file them as TODO items in a new design doc at `docs/superpowers/specs/<today>-htmgo-guidance-skill-gaps.md` or directly as additional commits amending the skill.

Do NOT amend past commits — add new commits with `fix(skills):` prefix.

- [ ] **Step 12.4: No commit for the verification step itself**

This is observational only. Commit in Task 12.3 only if gaps are found and a fix is being applied.

---

## Post-implementation checks

After all tasks complete, verify the full change:

```bash
# This repo
cd /home/iru/p/github.com/franchb/htmgo
git log --oneline master..htmx4-migration | head -15
# expected: new commits for each task, in order

wc -l .claude/skills/htmgo-guidance/SKILL.md
# expected: 700-900

# README check
grep -A1 "Claude Code skills" README.md | head

# Existing skills still intact
ls .claude/skills/
# expected: htmgo-guidance, htmx-debugging, htmx-extension-authoring,
#           htmx-guidance, htmx-migration, htmx-upgrade-from-htmx2
```

In vuln_catalog:

```bash
cd /home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog
ls .claude/skills/htmgo-guidance/
# expected: SKILL.md

git log -1 --oneline .claude/skills/htmgo-guidance/
# expected: the "docs(skills): add htmgo-guidance skill from franchb/htmgo" commit
```

If all three checks pass: the delivery is complete pending the Task 12 smoke test.
