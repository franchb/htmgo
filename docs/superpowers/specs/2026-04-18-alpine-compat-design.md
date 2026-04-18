# Design: bundle `alpine-compat` extension + add `ax/` Go helpers

**Date:** 2026-04-18
**Status:** approved, ready for implementation plan
**Target Alpine:** v3.15.11 (latest; released 2025-04-02)
**Target htmx:** 4.0.0-beta2 (currently pinned in fork)
**Target release:** v2.0.0-beta.2

## Problem

The vulnerability_catalog project (primary consumer of this htmgo fork) is mid-migration to a new design that adds tabs, popovers, split/table toggles, theme/density/accent switches, a ⌘K modal, copy-to-clipboard, and similar lightly-stateful widgets on top of an htmx-driven SSR app. A sibling brainstorming session landed on "htmx + Alpine.js for small widgets" as the recommended pattern: stay SSR-first via htmx; use a small Alpine component for pure-client UI state; no frontend build step.

For that pattern to work, Alpine components must survive htmx swaps (morph-based in htmx 4) without losing state, and Alpine must re-initialize on swapped-in content. htmx 4 ships an official extension for exactly this — `alpine-compat`, distributed as `hx-alpine-compat.js` in the `htmx.org` npm package. The htmgo fork does not currently bundle this extension, nor does it expose Go-side ergonomics for Alpine directives.

Two gaps close in this change:

1. **Runtime gap.** Projects using this fork cannot use Alpine correctly with htmx swaps without manually pulling `hx-alpine-compat.js` from a CDN or vendoring it — both inconsistent with how every other htmx extension reaches htmgo apps (bundled in `htmgo.js`).
2. **Ergonomics gap.** Writing `h.Attribute("x-on:keydown.escape.prevent", "open = false")` is verbose and error-prone versus `ax.OnKeydown("open = false", "escape", "prevent")`, and there is no Alpine equivalent to the `hx/` constants package.

## Decision

Four-part change, all targeting v2.0.0-beta.2 of this fork:

1. **Bundle `alpine-compat`** — hand-port upstream's IIFE to a TypeScript file under `framework/assets/js/htmxextensions/`, add one import line to `htmgo.ts`, rebuild `framework/assets/dist/htmgo.js`. Extension auto-gates on `window.Alpine?.*`, so non-Alpine users only pay ~3KB gz bundle cost.
2. **Add `framework/ax/` Go package** — constants + Ren-returning builders mirroring `framework/hx/`'s shape. Self-contained namespace: `ax.Data(...)`, `ax.Show(...)`, `ax.OnClick(...)`.
3. **Docs + overview update** — new `htmgo-site/pages/docs/htmx-extensions/alpine-compat.go` page covering loading Alpine, `x-cloak` FOUC rule, `ax` usage, and Alpine/htmx swap interaction. Overview page lists alpine-compat alongside existing extensions.
4. **CHANGELOG + CLAUDE.md updates** — one-line additions documenting the new extension and new package.

Explicitly out of scope for this change (can be additive later):
- Bundling Alpine itself into htmgo.js (fork stays general-purpose).
- Alpine v2 support (compat extension is v3-only by upstream design).
- Helpers for specific Alpine plugins (`@alpinejs/persist`, `@alpinejs/intersect`, `@alpinejs/focus`, etc.); users import plugin JS directly and use raw attribute strings.
- Build-flag to exclude alpine-compat from `htmgo.js`; 3KB cost does not justify the config surface.
- Server-side pre-hydration of Alpine state (not something upstream `alpine-compat` does either).
- New example app. The doc snippets cover the 80% case, and vuln_catalog itself becomes the working reference.

## Section 1 — JS bundle: `alpine-compat` extension

### File: `framework/assets/js/htmxextensions/alpine-compat.ts`

Hand-port the upstream `hx-alpine-compat.js` IIFE (see upstream source in `node_modules/htmx.org/dist/ext/hx-alpine-compat.js` at version `4.0.0-beta2`) to match the style of existing fork extensions. Specifically:

- `import htmx from "htmx.org"` (matches `response-targets.ts`, `pathdeps.ts`, etc.).
- `htmx.registerExtension("alpine-compat", { ... })` with all seven hooks from upstream preserved:
  - `init(api)` — captures `api` ref; patches `api.isSoftMatch` to ignore ID mismatch when both nodes have Alpine reactive ID bindings (`_x_bindings.id` / `[:id]` / `[x-bind:id]`).
  - `htmx_before_swap(elt, detail)` — if `window.Alpine.deferMutations` exists, increments `deferCount` and calls `Alpine.deferMutations()` on the first swap of a batch.
  - `htmx_before_morph_node(elt, detail)` — copies `_x_dataStack` from `oldNode` to `newNode` via `Alpine.closestDataStack(oldNode)`; runs `Alpine.cloneNode(oldNode, newNode)` if `oldNode.isConnected`; morphs the teleport target when both nodes have `_x_teleport`.
  - `htmx_history_cache_before_save(elt, detail)` — serializes each `[x-data]` root's `_x_dataStack[0]` into a `data-alpine-state` attribute, then calls `Alpine.destroyTree(detail.target)`.
  - `htmx_history_cache_after_restore(elt, detail)` — reads `data-alpine-state` back into the re-initialized `_x_dataStack[0]`, then removes the attribute.
  - `htmx_after_swap(elt, detail)` — sets `detail.ctx._alpineFlushed = true` and calls `maybeFlush()` (which decrements `deferCount` and invokes `Alpine.flushAndStopDeferringMutations()` when the count reaches zero).
  - `htmx_finally_request(elt, detail)` — safety net: if `_alpineFlushed` was not set (e.g., swap was short-circuited), call `maybeFlush()` anyway.
- Type discipline matches precedent: `any` for Alpine internals and `api` fields, optional chaining on every `window.Alpine?.*` read. No `@types/alpinejs` dependency added.
- Closure-scoped `deferCount` and `maybeFlush` helpers preserved from upstream.

### Register in `htmgo.ts`

Append one import line (order inside this block is not load-bearing — extensions self-register):

```ts
// framework/assets/js/htmgo.ts
import "./htmxextensions/pathdeps";
import "./htmxextensions/trigger-children";
import "./htmxextensions/debug";
import "./htmxextensions/response-targets";
import "./htmxextensions/mutation-error";
import "./htmxextensions/livereload";
import "./htmxextensions/htmgo";
import "./htmxextensions/sse";
import "./htmxextensions/ws";
import "./htmxextensions/ws-event-handler";
import "./htmxextensions/alpine-compat"; // NEW
```

### Bundle

Rebuild: `cd framework/assets/js && npm run build` → produces `framework/assets/dist/htmgo.js`. Commit the regenerated bundle in the same commit as the TypeScript source (established fork pattern, per `docs/superpowers/plans/2026-04-18-response-targets-error-flag.md`).

### Tests

**File:** `framework/assets/js/htmxextensions/__tests__/alpine-compat.test.ts`

Mirror the vitest pattern used in `response-targets.test.ts`. Mock `htmx.org`, import the extension, retrieve it from the `registered` map, exercise hooks.

Required cases:
1. **Registers with expected hooks.** `init`, `htmx_before_swap`, `htmx_before_morph_node`, `htmx_history_cache_before_save`, `htmx_history_cache_after_restore`, `htmx_after_swap`, `htmx_finally_request` all typeof `"function"`.
2. **Self-gates without Alpine.** Each hook invoked with `window.Alpine` undefined is a no-op (no throws, no mutations).
3. **`htmx_before_morph_node` copies `_x_dataStack`.** Stub `window.Alpine.closestDataStack` and `window.Alpine.cloneNode`; pass `{oldNode, newNode}` with `oldNode.isConnected === true`; assert `newNode._x_dataStack` === stub result and `cloneNode` was called once.
4. **Teleport morph.** When both nodes carry `_x_teleport`, assert `api.morph` is called with the old teleport as target and a fragment containing the new teleport.
5. **`htmx_history_cache_before_save` serializes Alpine state.** Build a DOM with two `[x-data]` nodes having `_x_dataStack = [{...state}]`; assert each gets a `data-alpine-state` attribute with JSON-serialized state; assert `Alpine.destroyTree` was called on `detail.target`.
6. **`htmx_history_cache_after_restore` deserializes Alpine state.** Build a DOM with `[data-alpine-state]` nodes; assert `_x_dataStack[0]` gets the parsed state merged in and the attribute is removed.
7. **`htmx_after_swap` flushes defer counter.** Invoke `htmx_before_swap` twice (deferCount=2), then `htmx_after_swap` twice; assert `Alpine.flushAndStopDeferringMutations` is called exactly once (on the second flush when counter reaches zero).
8. **API-surface contract test.** Read the built `alpine-compat.ts` source as a string and assert it references every documented internal symbol: `closestDataStack`, `cloneNode`, `deferMutations`, `destroyTree`, `flushAndStopDeferringMutations` (Alpine) + `isSoftMatch`, `morph` (htmx). Protects against silent API drift — if Alpine or htmx renames a symbol and we port the rename, this test will need updating deliberately and future readers see the dependency contract in one place.

## Section 2 — `framework/ax/` Go package

### Package layout

```
framework/ax/
  alpine.go       — type alias + attribute name constants
  builder.go      — Ren-returning builder functions
  alpine_test.go  — table tests covering every builder
```

Namespace chosen: `ax` (short, mirrors `hx` style; reads naturally at the call site: `ax.Data(...)`).

### `alpine.go` — constants

```go
package ax

type Attribute = string

const (
    DataAttr       Attribute = "x-data"
    InitAttr       Attribute = "x-init"
    ShowAttr       Attribute = "x-show"
    BindAttr       Attribute = "x-bind"       // use with a sub-attr, e.g. x-bind:class
    OnAttr         Attribute = "x-on"         // use with a sub-attr, e.g. x-on:click
    TextAttr       Attribute = "x-text"
    HtmlAttr       Attribute = "x-html"
    ModelAttr      Attribute = "x-model"
    ModelableAttr  Attribute = "x-modelable"
    CloakAttr      Attribute = "x-cloak"
    RefAttr        Attribute = "x-ref"
    IgnoreAttr     Attribute = "x-ignore"
    TeleportAttr   Attribute = "x-teleport"
    EffectAttr     Attribute = "x-effect"
    IfAttr         Attribute = "x-if"
    ForAttr        Attribute = "x-for"
    IdAttr         Attribute = "x-id"
    TransitionAttr Attribute = "x-transition"
)
```

### `builder.go` — Ren helpers

Imports `github.com/franchb/htmgo/framework/h`. Each builder wraps `h.Attribute(name, value)` and returns `h.Ren`.

**Simple single-arg directives** (value passed through verbatim as the Alpine expression):
- `Data(expr string) h.Ren`
- `Init(expr string) h.Ren`
- `Show(expr string) h.Ren`
- `Text(expr string) h.Ren`
- `Html(expr string) h.Ren`
- `Model(expr string) h.Ren`
- `Effect(expr string) h.Ren`
- `Modelable(expr string) h.Ren`
- `If(expr string) h.Ren`
- `For(expr string) h.Ren`
- `Id(scopes string) h.Ren`
- `Ref(name string) h.Ren`
- `Teleport(selector string) h.Ren`

**No-arg directives:**
- `Cloak() h.Ren` — emits bare `x-cloak` (htmgo's renderer omits `=""` for empty attribute values; Alpine accepts the boolean-attribute form identically)
- `Ignore() h.Ren` — emits bare `x-ignore`

**`x-bind:*` family:**
- `Bind(attr, expr string) h.Ren` — generic; emits `x-bind:{attr}="{expr}"`
- `BindClass(expr string) h.Ren` — `x-bind:class`
- `BindStyle(expr string) h.Ren` — `x-bind:style`
- `BindHref(expr string) h.Ren` — `x-bind:href`
- `BindValue(expr string) h.Ren` — `x-bind:value`
- `BindDisabled(expr string) h.Ren` — `x-bind:disabled`
- `BindChecked(expr string) h.Ren` — `x-bind:checked`
- `BindId(expr string) h.Ren` — `x-bind:id`

**`x-on:*` family (event + optional modifiers):**
- `On(event, handler string, modifiers ...string) h.Ren` — generic; joins modifiers with `.`, emits `x-on:{event}.{mod1}.{mod2}="{handler}"`.
- Common-event shortcuts (all forward `modifiers ...string` to `On`):
  - `OnClick(handler string, modifiers ...string) h.Ren`
  - `OnSubmit(handler string, modifiers ...string) h.Ren`
  - `OnInput(handler string, modifiers ...string) h.Ren`
  - `OnChange(handler string, modifiers ...string) h.Ren`
  - `OnFocus(handler string, modifiers ...string) h.Ren`
  - `OnBlur(handler string, modifiers ...string) h.Ren`
  - `OnKeydown(handler string, modifiers ...string) h.Ren`
  - `OnKeyup(handler string, modifiers ...string) h.Ren`
- Combo shortcuts (for the most common modifier patterns):
  - `OnClickOutside(handler string) h.Ren` — `x-on:click.outside`
  - `OnKeydownEscape(handler string) h.Ren` — `x-on:keydown.escape`
  - `OnKeydownEnter(handler string) h.Ren` — `x-on:keydown.enter`

**`x-model` modifier variants:**
- `ModelNumber(expr string) h.Ren`
- `ModelLazy(expr string) h.Ren`
- `ModelTrim(expr string) h.Ren`
- `ModelFill(expr string) h.Ren`
- `ModelBoolean(expr string) h.Ren`
- `ModelDebounce(expr, duration string) h.Ren` — emits `x-model.debounce.{duration}`, e.g. `ModelDebounce("q", "500ms")` → `x-model.debounce.500ms="q"`

**`x-transition` (kept minimal):**
- `Transition() h.Ren` — bare `x-transition` (same renderer-empty-value behavior as `Cloak`/`Ignore` above)
- Richer transition shapes (`x-transition:enter`, `:enter-start`, `:enter-end`, `.opacity`, `.duration.500ms`, etc.) are left to raw `h.Attribute("x-transition:enter", ...)` callsites for now. If patterns repeat in real use, revisit.

### Tests — `alpine_test.go`

Table-driven tests mirroring `hx/htmx_test.go`:

| Builder input | Expected attribute output |
|---|---|
| `ax.Data("{ open: false }")` | `x-data="{ open: false }"` |
| `ax.Show("open")` | `x-show="open"` |
| `ax.Cloak()` | `x-cloak=""` |
| `ax.Bind("class", "{ active: isActive }")` | `x-bind:class="{ active: isActive }"` |
| `ax.BindClass("{ active: isActive }")` | `x-bind:class="{ active: isActive }"` |
| `ax.OnClick("count++")` | `x-on:click="count++"` |
| `ax.OnClick("submit()", "prevent")` | `x-on:click.prevent="submit()"` |
| `ax.OnKeydown("close()", "escape", "prevent")` | `x-on:keydown.escape.prevent="close()"` |
| `ax.OnClickOutside("open = false")` | `x-on:click.outside="open = false"` |
| `ax.ModelDebounce("query", "500ms")` | `x-model.debounce.500ms="query"` |

Render by wrapping each builder in `h.Div(...)` and calling `h.Render(div)` (same pattern as `framework/h/render_test.go`'s `TestRenderAttributes_1`), then assert the rendered string contains the expected attribute substring. Example:

```go
func TestAxData(t *testing.T) {
    div := h.Div(Data("{ open: false }"))
    assert.Contains(t, h.Render(div), `x-data="{ open: false }"`)
}
```

### Interop with `h/` and `hx/`

- `ax` imports `h` (one-way dependency; `h` does NOT import `ax`). Matches how `hx` is independent but `h` imports `hx` — reversed here because the builders live inside `ax`.
- Call-site mixing of `ax.*`, `h.*`, and `hx.*` is idiomatic:

```go
h.Div(
    h.Class("relative"),
    ax.Data(`{ open: false }`),

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
```

## Section 3 — Docs + site changes

### New doc page: `htmgo-site/pages/docs/htmx-extensions/alpine-compat.go`

Follow the shape of `htmgo-site/pages/docs/htmx-extensions/overview.go`. Content outline:

1. **What it does** — one short paragraph: alpine-compat preserves Alpine state across htmx morph swaps, re-runs Alpine init on swapped-in content, and round-trips Alpine component state through the htmx history cache. Bundled automatically in `/public/htmgo.js`; no `hx-ext` attribute needed (htmx 4 extensions self-register).
2. **Loading Alpine alongside htmgo** — code snippet with recommended version pin:

   ```go
   h.Html(
       h.Head(
           h.Script(
               "https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js",
               h.Attribute("defer", ""),
           ),
           h.Script("/public/htmgo.js"),
       ),
   )
   ```

   Note: `defer` on Alpine is important (Alpine auto-initializes on `DOMContentLoaded`); `htmgo.js` can be loaded before or after, order is not load-bearing.
3. **FOUC prevention** — one-liner CSS rule users should add to their stylesheet:
   ```css
   [x-cloak] { display: none !important; }
   ```
4. **`ax` package quick tour** — compact cheatsheet: `ax.Data`, `ax.Show`, `ax.OnClick`, `ax.Model`, `ax.BindClass`, `ax.OnKeydownEscape`, `ax.Cloak`. Link to GoDoc for the full list.
5. **Worked example** — the popover snippet from Section 2's interop block.
6. **Alpine + htmx swap interaction** — brief explanation: when htmx morphs content into an Alpine component's descendant, alpine-compat's `htmx_before_morph_node` copies `_x_dataStack` to the new nodes and `Alpine.cloneNode` preserves bindings; state survives. When htmx morphs the Alpine root itself, alpine-compat still carries state through because it runs before htmx 4's morph op.
7. **Gotchas:**
   - Alpine v3 only (v2 is out of support upstream).
   - Recommended pin: `alpinejs@3.15.11` (tested version). Patch updates are safe; 3.16 not on the roadmap.
   - If Alpine loads AFTER a swap, pre-Alpine widgets won't init until Alpine boots — load Alpine in `<head>` with `defer`.
   - Alpine plugins (`@alpinejs/persist`, etc.) need to be loaded BEFORE Alpine itself, per upstream plugin docs.

### Update: `htmgo-site/pages/docs/htmx-extensions/overview.go`

Add `Link("Alpine Compat", "/docs/htmx-extensions/alpine-compat")` to the extensions list. Update the intro prose to mention alpine-compat as a bundled extension.

### Update: `CHANGELOG.md` under `## [Unreleased]`

Two lines added:
- **Added:** `alpine-compat` htmx extension bundled into `htmgo.js` — auto-preserves Alpine.js state across htmx morph swaps and re-initializes Alpine on swapped content. Auto-gates on `window.Alpine` presence; no bundle cost if Alpine isn't loaded.
- **Added:** `framework/ax/` Go package — constants + builder helpers for Alpine.js directives (`ax.Data`, `ax.Show`, `ax.OnClick`, `ax.Model`, `ax.Bind`, etc.) mirroring the `hx/` package shape.

### Update: `CLAUDE.md`

Extend the "Custom TypeScript extensions live in `framework/assets/js/htmxextensions/`" sentence to include `alpine-compat` in the listed extension names.

### Starter template untouched

`examples/*/pages/.../root.go` and the starter template already include `h.Script("/public/htmgo.js")`, which after this change bundles alpine-compat. Users who want Alpine add the Alpine CDN script tag themselves. No starter-template change needed.

## Section 4 — Testing, verification, and Alpine v3.15.11 support

### Test matrix

| Layer | Tool | Coverage |
|---|---|---|
| JS extension unit | vitest | hook registration, state transitions, self-gating, API-surface contract (see Section 1 tests) |
| Go `ax` package | `go test ./ax/` | every builder renders the expected attribute string; modifier composition order is correct |
| Bundle integrity | manual + CI-friendly grep | `grep -c "alpine-compat" framework/assets/dist/htmgo.js` > 0 |
| Integration smoke | Playwright MCP on htmgo-site | new doc page renders, no console errors, code snippets highlighted |
| E2E Alpine + htmx | manual Playwright MCP | with Alpine v3.15.11 injected: `x-data` mounts, state survives an `hx-swap`, no console errors |

### Alpine v3.15.11 API verification (done during design)

Verified by reading `packages/alpinejs/src/alpine.js` at tag `v3.15.11`:

| Symbol | Present | Notes |
|---|---|---|
| `Alpine.closestDataStack` | ✓ | Exposed on the Alpine object |
| `Alpine.cloneNode` | ✓ | Marked `// INTERNAL` upstream |
| `Alpine.deferMutations` | ✓ | Defined in `packages/alpinejs/src/mutation.js`, re-exported on Alpine |
| `Alpine.destroyTree` | ✓ | Exposed on the Alpine object |
| `Alpine.flushAndStopDeferringMutations` | ✓ | Defined in `packages/alpinejs/src/mutation.js`, re-exported on Alpine |

htmx 4 internal API verification (from `node_modules/htmx.org/dist/htmx.js` 4.0.0-beta2):

| Symbol | Present | Notes |
|---|---|---|
| `api.isSoftMatch` | ✓ | Used by alpine-compat's `init` to patch morph ID handling |
| `api.morph` | ✓ | Used for teleport-target morphing |
| `detail.tasks` with `swapSpec.style` of `innerMorph`/`outerMorph` | ✓ | Matches response-targets' task-based access pattern |

### Version-pin recommendation in docs

Doc page recommends `alpinejs@3.15.11` (current latest, released 2025-04-02). Also shows `alpinejs@3.15` looser range for users who want patch updates. Fork does not ship Alpine itself, so no `package.json` pin is added.

### Risk register

- **Alpine v3.x internal API drift** — Low. Compat extension uses optional chaining on every call into Alpine internals (`window.Alpine?.xxx`), so removed APIs degrade to no-op rather than crash. Alpine v3 is in patch-maintenance mode (v3.15.11 is 2025-04-02; no 3.16 on the public roadmap). `cloneNode` is marked `// INTERNAL` upstream, but Alpine's own morph plugin and Laravel Livewire rely on the same surface. Mitigation: recommend version pin in docs; vitest API-surface test detects drift on this fork's own dep bumps.
- **htmx 4 internal API drift before stable** — Medium until htmx 4.0 stable. Extension reads `task.swapSpec.style`, `api.isSoftMatch`, `api.morph`. Mitigation: fork pins `htmx.org@4.0.0-beta2` (same mitigation as every other extension in the fork per CHANGELOG 1.2.0-beta.1); audit on each htmx bump.
- **User loads Alpine after a swap** — Low. If a page first render triggers an `hx-trigger="load"` before Alpine boots, the first swap's Alpine state handling is a no-op. Document the `<head>` + `defer` recommendation. Typical apps load Alpine before any user-initiated swap.
- **`defer` ordering gotcha** — Low. If user omits `defer`, Alpine runs before `DOMContentLoaded`, potentially before `[x-data]` roots are parsed in deferred script scenarios. Recommended snippet uses `defer`. Doc page calls this out explicitly.

### Vulnerability_catalog migration path (post-release)

1. Bump the `github.com/franchb/htmgo/framework` dep in vuln_catalog to the commit or tag containing this change.
2. In the root layout, add Alpine before `htmgo.js`:
   ```go
   h.Script("https://unpkg.com/alpinejs@3.15.11/dist/cdn.min.js", h.Attribute("defer", ""))
   h.Script("/public/htmgo.js")
   ```
3. Add `[x-cloak] { display: none !important; }` to the site stylesheet (or Tailwind global layer).
4. Convert client-state widgets to Alpine using `ax.*`:
   - theme / density / accent toggles → `ax.Data`, `ax.Model`, `$persist` magic for `localStorage` (requires `@alpinejs/persist` plugin; separate opt-in)
   - popovers / dropdowns → `ax.Data`, `ax.Show`, `ax.OnClickOutside`, `ax.OnKeydownEscape`, `ax.Cloak`
   - split / table toggles → `ax.Data`, `ax.BindClass`
   - ⌘K modal → `ax.Data`, `ax.OnKeydown("open = true", "meta", "k", "prevent")` on `<body>` (each key-modifier is a separate arg — emits `x-on:keydown.meta.k.prevent`), `ax.Show`, `ax.Ref` for focus mgmt
   - copy-to-clipboard → `ax.OnClick("navigator.clipboard.writeText($el.dataset.copy)")`
5. Widgets that need htmx + Alpine co-existence (e.g., tabs that both swap content AND toggle client state): put `ax.Data(...)` on the outer wrapper; htmx swaps inner content; alpine-compat's `htmx_before_morph_node` hook carries the outer `_x_dataStack` through the morph automatically.

## Implementation plan reference

See forthcoming `docs/superpowers/plans/2026-04-18-alpine-compat.md` for step-by-step implementation tasks.

## References

- Upstream alpine-compat source: `node_modules/htmx.org/dist/ext/hx-alpine-compat.js` (htmx.org 4.0.0-beta2)
- htmx 4 docs: <https://four.htmx.org/docs/extensions/alpine-compat>
- Alpine.js v3.15.11 release: <https://github.com/alpinejs/alpine/releases/tag/v3.15.11>
- Alpine internal API source: <https://github.com/alpinejs/alpine/blob/v3.15.11/packages/alpinejs/src/alpine.js>
- Alpine Morph plugin docs: <https://alpinejs.dev/plugins/morph> (pattern reference; not used by alpine-compat itself)
- htmgo fork extension authoring pattern: `framework/assets/js/htmxextensions/response-targets.ts`
- htmgo `hx/` package (mirrored shape): `framework/hx/htmx.go`
