# Design: `htmgo-guidance` Claude Code skill

**Date:** 2026-04-18
**Status:** approved, ready for implementation plan
**Target consumer:** Claude Code AI sessions working in downstream htmgo apps (primary user: `/home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog`)
**Target release:** same fork branch (`htmx4-migration` → PR #12), ships alongside the alpine-compat delivery

## Problem

AI coding sessions working in projects that consume this htmgo fork have no consolidated reference for htmgo's Go APIs. The repo already ships five Claude Code skills under `.claude/skills/` — `htmx-guidance`, `htmx-debugging`, `htmx-extension-authoring`, `htmx-migration`, `htmx-upgrade-from-htmx2` — but every one of them is about **htmx itself**, not about **htmgo's Go builder/routing/DI surface** that sits on top of htmx.

Concretely, a Claude Code session at `vulnerability_catalog` asked to build an htmgo page today has to infer the `h.Div`/`h.Ren`/`h.NewPage`/`RequestContext`/Fiber wiring from stale training data or by reading framework source at runtime. Errors look like:
- Mixing htmx 2 and htmx 4 event/attr constants
- Forgetting htmx 4's explicit inheritance (passing `hx-target` where `hx-target:inherited` is needed)
- Unaware of the new `framework/ax/` package and the bundled alpine-compat extension — regressing to "add an Alpine script tag + write x-on attributes manually"
- Confusing htmgo's lifecycle command DSL (`h.OnClick(js.Alert(...))`) with Alpine's event directives (`ax.OnClick("count++")`)

This skill closes the gap with a single consolidated reference that auto-loads when a session is working in an htmgo codebase.

## Decision

Create one skill: `htmgo-guidance` at `.claude/skills/htmgo-guidance/SKILL.md`. Monolithic design (~700-850 lines) covering the full consumer surface of htmgo in one file. No separate Alpine skill — `ax/` content is a full section within the monolith.

Distribution is **in-repo copy-paste**, matching the existing `htmx-*` skills: consumers clone or otherwise pull the file into their own `.claude/skills/htmgo-guidance/`. A short "Claude Code skills" subsection is added to the htmgo `README.md` (the repo's existing 48-line README is the natural discovery surface; no separate doc file needed for a short section).

Explicitly **not** built in this change:
- Separate `htmgo-upgrade` skill (upgrades are a 1-paragraph pointer section in the monolith; split out later if content grows)
- Separate `htmgo-alpine` skill (Alpine content lives inside `htmgo-guidance`)
- Claude Code plugin packaging (copy-paste is sufficient for a single primary consumer)
- A contributor-facing skill for people hacking on the htmgo framework itself
- Tailwind, DB, deployment guidance (not htmgo's concern)

## Section 1 — Skill metadata

### File: `.claude/skills/htmgo-guidance/SKILL.md`

Frontmatter:

```yaml
---
name: htmgo-guidance
description: Use when writing Go code that uses the htmgo framework (github.com/franchb/htmgo), building pages, partials, or components with h.Div/h.Button/h.Ren, wiring hx/ htmx attributes or ax/ Alpine directives, or answering questions about htmgo patterns, routing, and best practices. Covers Fiber v3 integration, RequestContext, auto-routing, service locator, caching, and the Alpine compat extension.
---
```

The `description` field is the sole auto-activation signal. Design criteria:
- Wide enough to fire on most htmgo work in consumer projects (primary trigger: "writing Go code that uses htmgo framework").
- Names the concrete module path `github.com/franchb/htmgo` so it distinguishes this fork from upstream `maddalax/htmgo`.
- Names the three builder sub-namespaces (`h/`, `hx/`, `ax/`) so codebase surveys that mention any of them will load the skill.
- Narrow enough to not fire on generic Go tasks or on unrelated frameworks.

## Section 2 — Skill content structure

Estimated 700-850 lines, shaped like `htmx-guidance.md` (reference + worked examples, not prescriptive "when X do Y" workflows). Twelve sections; sections 1-6 carry the bulk of content, 7-12 are tighter.

### 1. What htmgo is (1 paragraph)

One-paragraph model statement: Go builder functions produce HTML; Page/Partial split separates full documents from htmx fragments; Fiber v3 handles routing; no JS build step required for consumers. Import path `github.com/franchb/htmgo/framework`. Key pins: Go 1.23, htmx 4.0.0-beta2, Alpine.js 3.15.x (optional, consumer-loaded).

### 2. The `h.Ren` builder model (~80 lines)

- `Ren` interface — `Render(*RenderContext)`. Everything that renders implements it.
- `*h.Element` — the primary node type. Created via tag builders (`h.Div`, `h.Button`, `h.Form`, `h.H1`–`h.H6`, `h.P`, `h.Span`, `h.A`, `h.Img`, `h.Ul`/`h.Li`, `h.Table`/`h.Tr`/`h.Td`, `h.Input`/`h.Select`/`h.Option`, `h.Svg`/`h.Path`, etc.).
- Children are variadic `Ren`. Attributes are also `Ren` — specifically `*h.AttributeR` returned by `h.Attribute(key, value)`. Pass them as children; the renderer separates attrs from body children at render time.
- `h.Tag(name, children...)` for arbitrary HTML tags not in the built-ins.
- `h.Text(string)` / `h.TextF(format, args...)` — text nodes; text is HTML-escaped by default.
- `h.UnsafeRaw(string)` — escape hatch; use only with trusted content. Never with user input.
- `h.Class(...)` / `h.ClassX(map)` / `h.Id(string)` — common attribute helpers.
- Example: a small card component.

### 3. Pages vs Partials (~70 lines)

- `h.NewPage(root)` returns `*h.Page` — a full HTML document response.
- `h.NewPartial(root)` returns `*h.Partial` — an HTML fragment response with optional htmx response headers (via `h.NewPartialWithHeaders`).
- Route handler signatures: `func(ctx *h.RequestContext) *h.Page` (for pages) or `func(ctx *h.RequestContext) *h.Partial` (for partials).
- **Auto-routing:** `htmgo generate` scans `pages/` and `partials/` and writes `__htmgo/pages-generated.go` + `__htmgo/partials-generated.go` that register routes by file path. `pages/foo/bar.go` → `/foo/bar`. Never edit the generated files.
- Partial function-name conventions: `GET`/`POST`/`PUT`/`PATCH`/`DELETE` prefix determines the HTTP method. Multiple methods per partial file allowed.
- Example: a page with a button that POSTs to a partial and swaps in the response.

### 4. RequestContext (~60 lines)

- Wraps Fiber v3's `fiber.Ctx`. Get it inside any Fiber handler/middleware via `h.GetRequestContext(c fiber.Ctx)`. Route handlers receive it directly.
- Key methods: `FormValue(key)`, `QueryParam(key)`, `UrlParam(key)`, `IsHxRequest()`, `Redirect(url)`, `HxSource()` / `HxSourceID()`, `HxRequestType()`, `Header(name)`, cookie helpers.
- Escape hatch: `ctx.Fiber` gives the raw `fiber.Ctx` for anything not pre-wrapped.
- Example: a form-submit partial reading `FormValue` and returning validation errors with `h.NewPartialWithHeaders` + `HX-Retarget`.

### 5. htmx integration via `hx/` + `h.Hx*` helpers (~150 lines)

- `framework/hx/` package — constants only. Attribute names (`hx.GetAttr = "hx-get"`, `hx.TargetAttr`, `hx.TriggerAttr`, ...), event names in htmx 4 colon form (`hx.AfterSwapEvent = "htmx:after:swap"`), header names, swap types (`hx.SwapTypeInnerHtml = "innerHTML"`, ...).
- Builder helpers in `framework/h/`: `h.HxGet/Post/Put/Patch/Delete(url)`, `h.HxTarget(sel)`, `h.HxTrigger(...)`, `h.HxSwap(strategy)`, `h.HxInclude(sel)`, `h.HxConfirm(msg)`, `h.HxIndicator(sel)`.
- **Explicit inheritance in htmx 4.** `hx-target` no longer cascades to descendants. Use `:inherited` variants on ancestors whose children do the requests: `h.HxTargetInherited`, `h.HxIncludeInherited`, `h.HxSwapInherited`, `h.HxBoostInherited`, `h.HxConfirmInherited`, `h.HxHeadersInherited`, `h.HxIndicatorInherited`, `h.HxSyncInherited`, `h.HxEncodingInherited`, `h.HxValidateInherited`.
- Trigger spec: `h.HxTrigger(hx.TriggerEvent{...})` for structured specs; `h.HxTriggerString("click once, keyup[key=='Escape']")` for raw strings; `h.HxTriggerClick(mods...)` shortcut.
- Response headers: set via `h.NewPartialWithHeaders(root, headers...)`. Available: `HxRedirect`, `HxRetarget`, `HxReswap`, `HxReselect`, `HxLocation`, `HxPushUrl`, `HxReplaceUrl`, `HxRefresh`, `HxTrigger` (client-side event trigger from server).
- Worked example: search box with `HxPost` + `HxTrigger("keyup changed delay:300ms")` + `HxTarget("#results")`.
- "See also" pointer to the `htmx-guidance` skill for general htmx patterns and attribute semantics.

### 6. Alpine integration via `ax/` + the bundled compat extension (~130 lines)

- The **alpine-compat extension is auto-bundled** in `/public/htmgo.js` (since v1.2.0-beta.2). Consumers add only the Alpine CDN script and `x-cloak` CSS:
  ```go
  h.Tag("script",
      h.Attribute("src", "https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js"),
      h.Attribute("defer", ""),
  )
  ```
  Plus CSS: `[x-cloak] { display: none !important; }`.
- `framework/ax/` package: 18 constants mirroring `hx/` style (`ax.DataAttr = "x-data"`, etc.) + ~30 Ren-returning builders.
- Single-arg directives: `ax.Data(expr)`, `ax.Init`, `ax.Show`, `ax.Text`, `ax.Html`, `ax.Model`, `ax.Effect`, `ax.If`, `ax.For`, `ax.Id`, `ax.Ref`, `ax.Teleport`, `ax.Modelable`.
- No-arg: `ax.Cloak()`, `ax.Ignore()`, `ax.Transition()`.
- `x-bind:*` family: `ax.Bind(attr, expr)` + shortcuts `BindClass/BindStyle/BindHref/BindValue/BindDisabled/BindChecked/BindId`.
- `x-on:*` family: `ax.On(event, handler, modifiers...)` builds `x-on:{event}[.{mod}]*="{handler}"`. Shortcuts: `OnClick`, `OnSubmit`, `OnInput`, `OnChange`, `OnFocus`, `OnBlur`, `OnKeydown`, `OnKeyup` (each variadic on modifiers). Combos: `OnClickOutside`, `OnKeydownEscape`, `OnKeydownEnter`.
- `x-model` modifier variants: `ModelNumber`, `ModelLazy`, `ModelTrim`, `ModelFill`, `ModelBoolean`, `ModelDebounce(expr, duration)`.
- **When to use `ax.*` vs htmx swaps vs pure JS** (the decision-rubric readers will want):
  - Pure-client state that never needs the server → `ax.*` (theme toggle, popover open/close, keyboard shortcut overlay, copy-to-clipboard).
  - Server round-trip required → htmx (`h.HxGet/Post/...`).
  - Mixed (widget with both client state AND server-swapped content) → `ax.Data(...)` on the outer wrapper + `h.HxGet(...)` on inner elements. The alpine-compat extension carries `_x_dataStack` across the morph automatically.
  - Imperative one-shots (e.g. scroll to element, focus input) → lifecycle commands (§7), not Alpine.
- Worked popover example showing all of the above.
- Gotcha list: Alpine v3 only; load Alpine in `<head>` with `defer`; plugins load before Alpine; `x-cloak` CSS required to prevent FOUC.

### 7. Lifecycle & JS command DSL (~50 lines)

- `framework/h/lifecycle.go` helpers: `h.OnClick(cmd...)`, `h.OnLoad(cmd...)`, `h.HxOnAfterSwap(cmd...)`, `h.HxBeforeRequest(cmd...)`, SSE event handlers (`h.HxBeforeSseMessage`, etc.).
- These emit JavaScript via the `framework/js/` command DSL: `js.Alert(msg)`, `js.SetValue(sel, val)`, `js.SetInnerHTML`, `js.SetInnerText`, `js.AddClass`, `js.RemoveClass`, `js.ToggleClass`, `js.Redirect`, `js.EvalJs(raw)`.
- Two command types: `SimpleJsCommand` (one statement) and `ComplexJsCommand` (multi-statement block). Both satisfy `Command`.
- **Contrast with Alpine's `ax.OnClick`:**
  - `h.OnClick(js.Alert("hi"))` → inline onclick attribute running an htmgo-generated JS function. Use for imperative one-shots without server round-trip.
  - `ax.OnClick("open = !open")` → Alpine directive evaluated by Alpine runtime. Use for anything that manipulates Alpine state.
  - `h.HxGet("/x")` + button → htmx request. Use for server round-trips.
- Common mistake: `h.OnClick(js.Alert(...))` next to `ax.OnClick(...)` on the same element. Both fire; prefer one approach per element.

### 8. Service locator (~30 lines)

- `framework/service/` — typed DI via generics.
- `locator := service.NewLocator()` in `main.go`, passed into `h.Start(h.AppOpts{ServiceLocator: locator, ...})`.
- Register: `service.Set[*sql.DB](locator, service.Singleton, func() *sql.DB { return openDB() })`.
- Resolve: `db := service.Get[*sql.DB](locator)`. Works inside route handlers via `ctx.ServiceLocator()`.
- Lifecycles: `Singleton` (one instance per locator), `Transient` (new instance per resolve).
- Worked example: DB handle + config struct registered at boot, resolved in a partial handler.

### 9. Caching (~40 lines)

- `framework/h/cache.go` high-level API: `h.Cache(ctx, key, ttl, func() Ren { ... })` computes+caches a rendered fragment.
- `framework/h/cache/` low-level: `Store[K, V]` interface (`Set`, `GetOrCompute`, `Delete`, `Purge`, `Close`).
- Built-in: `TTLStore` (time-based eviction).
- Example: `LRUStore` in `h/cache/lru_store_example.go` for pluggable custom stores.
- When to cache: expensive DB reads, template fragments shown on many pages, third-party API responses. Avoid caching anything user-specific without a user-scoped key.

### 10. Project configuration & CLI (~40 lines)

- `htmgo.yml` / `htmgo.yaml` / `_htmgo.yaml` / `_htmgo.yml` in app root. Fields:
  - `tailwind: true/false` — run Tailwind CSS compilation.
  - `watch_ignore` / `watch_files` — glob patterns for dev-mode file watcher.
  - `automatic_page_routing_ignore` / `automatic_partial_routing_ignore` — exclude files from auto-route registration.
  - `public_asset_path` — URL prefix for static assets (default `/public`).
- CLI commands (via `go install ./cli/htmgo` or `go run github.com/franchb/htmgo/cli/htmgo@latest`):
  - `htmgo setup` — scaffold a new app
  - `htmgo build` — compile app + assets
  - `htmgo watch` / `htmgo dev` — live-reload dev server
  - `htmgo generate` — regenerate `__htmgo/*-generated.go`
  - `htmgo css` — Tailwind build
  - `htmgo format` — Go fmt + import cleanup
  - `htmgo template` — app from starter template
- Dev workflow via `task` (taskfile.yml); example apps provide working `Taskfile.yml` to copy.
- `__htmgo/` directory is generated scratch; gitignore it; never edit by hand.

### 11. Common pitfalls (~50 lines)

Concrete error patterns + fixes:
- **Missing `:inherited` after htmx 4 upgrade.** Symptom: clicks on child elements no longer swap. Fix: on the ancestor setting the attribute, change `h.HxTarget("#x")` → `h.HxTargetInherited("#x")`.
- **Returning `*h.Page` from a partial route** or `*h.Partial` from a page route. Fix: match the auto-router's expectation.
- **Mutating a route handler's input.** `ctx` is per-request; treat its state as read-only outside the framework's provided mutators.
- **Alpine loaded AFTER `htmgo.js`** (without `defer`). First swap can fire before Alpine initializes; state won't survive. Fix: `h.Attribute("defer", "")` on the Alpine `<script>` and put it in `<head>`.
- **Mixing htmx swaps that replace the outer Alpine wrapper.** Alpine state survives inner morphs (alpine-compat hooks handle it), but if htmx outer-swaps the element with `x-data`, state is lost. Fix: use `innerHTML` swap or keep `x-data` on an element outside the swap target.
- **Confusing `h.OnClick(js.Alert(...))` with `ax.OnClick("...")`** on the same element. Both fire. Pick one.
- **Forgetting `x-cloak` CSS rule.** Visible FOUC on Alpine components. Fix: add `[x-cloak] { display: none !important; }` to the site stylesheet.
- **Editing `__htmgo/*-generated.go` by hand.** Changes get wiped by `htmgo generate`. Fix: edit the source page/partial file instead.

### 12. Upgrade pointers (~20 lines)

Short section — one paragraph per migration, pointing at `CHANGELOG.md` for details:
- **htmx 2 → htmx 4 within htmgo:** see `CHANGELOG.md` `[1.2.0-beta.1]`. Key move: explicit inheritance (`:inherited` variants), `hx-ext` removed (extensions self-register), `hx-disable` semantics flipped (use `IgnoreAttr` for old behavior).
- **chi → Fiber v3:** handlers now take `fiber.Ctx` instead of `http.Handler`/`http.ResponseWriter`. Access via `ctx.Fiber`. See the same changelog entry.
- **maddalax/htmgo → franchb/htmgo fork:** replace import paths; otherwise API-compatible except for the htmx 4 changes above.

End of skill.

## Section 3 — Distribution

### In this fork

Commit `.claude/skills/htmgo-guidance/SKILL.md` on the `htmx4-migration` branch. Same layout as the existing five `htmx-*` skills, so any Claude Code session opened at the htmgo repo root auto-loads it from `.claude/skills/`.

### In `README.md`

Add a short "Claude Code skills" subsection near the bottom of `README.md`. Content:

```markdown
## Claude Code skills

This repo ships Claude Code skills under `.claude/skills/` that teach AI coding sessions how to work with htmgo's APIs:

- `htmgo-guidance` — writing htmgo apps (builder, pages, partials, hx/ax helpers, caching)
- `htmx-guidance` — htmx patterns and best practices
- `htmx-debugging`, `htmx-extension-authoring`, `htmx-migration`, `htmx-upgrade-from-htmx2` — specialized htmx skills

To use these in a downstream project:

    # From the consumer project root (e.g. your app using htmgo):
    mkdir -p .claude/skills
    cp -r /path/to/htmgo/.claude/skills/htmgo-guidance .claude/skills/

Or, if you don't have htmgo cloned locally:

    git clone --depth=1 https://github.com/franchb/htmgo.git /tmp/htmgo
    cp -r /tmp/htmgo/.claude/skills/htmgo-guidance .claude/skills/
    rm -rf /tmp/htmgo

Then run `/skills` in Claude Code to verify `htmgo-guidance` is loaded.
```

### For vuln_catalog specifically

After the skill lands on `htmx4-migration` (the branch vuln_catalog already tracks for this fork's dep), run the copy command once in the vuln_catalog repo. Confirm via `/skills` that `htmgo-guidance` appears. No other changes needed there.

Optional: add a one-line "This project uses the htmgo-guidance skill copied from franchb/htmgo" note to vuln_catalog's own `CLAUDE.md` for provenance tracking. Skip if the user doesn't care.

### No other distribution machinery

- No `plugin.json`, no install script, no version handshake.
- If a consumer wants to update, they re-copy. Version drift is accepted (matches the existing htmx-* skills' propagation model).
- Moving to a proper Claude Code plugin is a documented follow-up when the skill count or consumer count justifies the extra machinery.

## Section 4 — Testing, verification, follow-ups

### Scope boundaries (explicit exclusions)

Not in this skill:
- **Writing new htmgo framework code** — this skill is for **consumers** of htmgo, not contributors. Framework internals (`h/renderer.go` internals, the AST generator, tsup config) are out of scope.
- **Content already covered by existing skills.** `htmgo-guidance` points at them with one-line "see also" mentions rather than duplicating:
  - `htmx-guidance` for general htmx patterns
  - `htmx-debugging` for troubleshooting htmx specifically
  - `htmx-extension-authoring` for writing new htmx extensions
  - `htmx-migration` / `htmx-upgrade-from-htmx2` for htmx-level migrations
- **Tailwind usage** — vuln_catalog has its own `tailwind-best-practices` skill; not duplicated here.
- **Database / persistence patterns** — not htmgo's concern.
- **Deployment / hosting guidance** — same.

### Verification strategy

1. **Read-through review.** After first draft, careful read to catch broken code examples, dead links, contradictions with `CLAUDE.md`.
2. **Code-example compile-check.** Every Go example in the skill should compile when pasted into a minimal htmgo app. Non-compiling snippets corrode trust fast. Approach: mark intentionally-elided snippets with `// ...` clearly; spot-check substantive examples against real `framework/` API signatures (e.g. `grep -n "^func " framework/h/*.go` to verify names and signatures before writing them into the skill).
3. **Smoke test with a real consumer.** Install the skill into vuln_catalog, open a Claude Code session, give it a real task from the current design migration (e.g. "convert this React popover spec into an htmgo+Alpine widget"). Observe whether the skill activates, whether the output uses `ax.*` correctly, whether it knows the `:inherited` inheritance rules, whether it hallucinates any APIs that don't exist. This is the real acceptance test.
4. **No automated tests.** Skills are markdown; vitest/go-test don't apply.

### Acceptance criteria

The skill passes verification if:
- All Go code examples compile against `framework/` at the tip of `htmx4-migration`.
- No reference to the old `maddalax/htmgo` import path.
- No htmx 2 attribute/event constants (all htmx 4 colon form).
- ax package examples use the actual exported names (no hallucinated helpers like `ax.Transition(classes string)` which doesn't take args).
- Lifecycle section distinguishes `h.OnClick(js.Alert(...))` from `ax.OnClick("expr")` explicitly.
- Under 900 lines total (hard cap; see Risk register below).
- Smoke test on vuln_catalog: Claude Code session asked to build a new widget produces idiomatic htmgo code using `ax.*` and `h.Hx*` helpers correctly on the first attempt.

### Follow-ups (explicitly deferred)

- `htmgo-upgrade` — dedicated migration skill if §12 grows beyond its current tight scope.
- `htmgo-framework-dev` — contributor-facing skill for people hacking on htmgo itself (AST gen, renderer internals, tsup config, extension authoring patterns specific to htmgo).
- **Convert to Claude Code plugin** once there are 3+ consumer projects using the skill set, or once the skill count exceeds what's comfortable to copy-paste.
- **Version stamp / drift detection.** Once htmgo stabilizes post-v2, consider embedding a "verified against htmgo vX.Y.Z" stamp at the top of the skill and a CI check that fails when framework signatures change without the skill being updated.

### Risk register

- **Skill rot.** The skill will become wrong as htmgo evolves — especially the hx/ax constants tables and builder signatures. Mitigation: prefer short descriptions of concrete types (`hx.TargetAttr` for the `hx-target` attribute name) over copy-pasting every value; link to Go source for ground truth when a reader needs the authoritative value; add a "last reviewed against htmgo vX.Y.Z / htmx 4.0.0-beta2 / Alpine 3.15.11" stamp at the top so future readers know the freshness.
- **Too long → ignored.** Claude Code may partial-load very long skills, and sections past ~800 lines may never activate for shorter user prompts. Mitigation: hard cap at 900 lines; put heaviest-used content (Ren model §2, Pages/Partials §3, hx/ §5, ax/ §6) in the first 500 lines; put pitfalls §11 and upgrade pointers §12 at the end where they're fine to partial-load.
- **Trigger overreach.** If the `description` field is too vague ("use for Go code"), the skill fires on unrelated Go tasks. Mitigation: description explicitly names "htmgo framework (github.com/franchb/htmgo)", "h.Div/h.Button/h.Ren", "hx/", "ax/" — auto-activation only fires when those patterns are in the user prompt or surfacing codebase context.
- **Sync drift between this fork's skill and consumers.** If vuln_catalog pulls an old copy and the skill evolves, the AI at vuln_catalog will be out of date. Mitigation: accept the drift (matches the existing htmx-* propagation model); the "last reviewed against" stamp gives readers a freshness signal; revisit if this becomes a real problem.

## Implementation plan reference

See forthcoming `docs/superpowers/plans/2026-04-18-htmgo-claude-skills.md` for step-by-step implementation tasks.

## References

- Existing skills shape (structure + frontmatter pattern): `.claude/skills/htmx-guidance/SKILL.md`, `.claude/skills/htmx-debugging/SKILL.md`
- Authoritative htmgo surface to summarize:
  - `framework/h/*.go` — builder, Ren, Element, lifecycle, RequestContext
  - `framework/hx/*.go` — htmx constants + triggers
  - `framework/ax/*.go` — Alpine constants + builders (this fork's addition)
  - `framework/service/` — service locator
  - `framework/h/cache/` — pluggable caches
  - `framework/config/` — htmgo.yml schema
  - `cli/htmgo/` — CLI commands
- Alpine-compat design context: `docs/superpowers/specs/2026-04-18-alpine-compat-design.md`
- htmx 4 migration context: `CHANGELOG.md` entry `[1.2.0-beta.1]`, `docs/superpowers/specs/2026-04-17-htmx-v4-migration-design.md`
- Consumer project: `/home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog/` — primary real user; currently has the `htmx-*` skills copied over.
