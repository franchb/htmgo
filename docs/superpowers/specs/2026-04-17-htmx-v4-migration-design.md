# htmx 4 Migration — Design

Date: 2026-04-17
Branch target: `master` (single large PR, released as `htmgo v2.0-beta`)
Upstream reference: htmx `4.0.0-beta2` (2026-04-14), cloned locally at `/home/iru/p/junk/htmx`

## Goal

Migrate the `franchb/htmgo` fork from htmx 2.0.8 to htmx 4.0.0-beta2 as a clean break.
Ship as `htmgo v2.0-beta` so downstream users know this is a breaking change.
Absorb further htmx 4 beta → GA deltas in follow-up PRs.

## Non-goals

- Backward compatibility with htmx 2 (no `htmx-2-compat.js`, no `implicitInheritance` flag,
  no deprecated-alias constants in Go).
- Revisiting the chi/Fiber routing story. `framework/h/app.go` uses `c.Fiber.Get` today; that
  is the existing state and this migration does not change it.
- The downstream `vulnerability_catalog` project migration (own cycle).

## Decisions (resolved during brainstorming)

| Topic                    | Decision                                                                |
|--------------------------|-------------------------------------------------------------------------|
| Release timing           | Migrate now against beta2, track upstream until GA                      |
| Compat strategy          | Clean break — rewrite everything to htmx 4 idiomatic                    |
| Inheritance audit scope  | Framework + examples + htmgo-site + templates (thorough)                |
| Go constant renames      | Rename in place, no aliases — compile errors are the migration signal   |
| JS extensions            | Rewrite all 11 extensions 1:1 to htmx 4 `registerExtension` API         |
| Sequencing               | Single large PR to master                                               |

## Scope summary

**Changed**:
- `framework/assets/js/package.json` — htmx.org dep bump
- `framework/hx/*.go` — constants renamed/deleted
- `framework/h/app.go` — `HX-Trigger-Name` → `HX-Source` parsing
- `framework/h/lifecycle.go` — `hx-on::` emission switched to colon-form
- `framework/h/renderer.go`, `framework/h/xhr.go`, `framework/h/attribute.go` — audit + inheritance helpers
- `framework/assets/js/htmxextensions/*.ts` — all 11 files rewritten
- `examples/*/`, `htmgo-site/`, `templates/`, `framework-ui/` — inheritance audit + fixes
- `extensions/websocket/` — verify against aligned upstream WS API
- `framework/h/*_test.go`, `framework/hx/htmx_test.go` — updated assertions
- `CHANGELOG.md`, `VERSION` — v2.0.0-beta.1 bump

## Architecture

Five layers executed in order inside the single PR. Commits are sequenced layer by layer so
review is legible.

### Layer A — JS dependency and build

`framework/assets/js/package.json`:

- `"htmx.org": "^2.0.8"` → `"htmx.org": "4.0.0-beta2"` (exact pin, no caret — see Risk R1).
- `npm install` regenerates `package-lock.json`.
- Verify `tsup` still produces `htmx.js` + extension bundles. No `tsup.config.ts` change expected
  unless extension entry names change.

Skip the upstream `htmax.js` bundled distribution; the framework controls its own bundle.

### Layer B — Go `framework/hx` constants

`framework/hx/htmx.go`:

**Events** — every `htmx:camelCase` → `htmx:colon:form`. Full map:

| Old constant               | Old value                      | New value                      |
|----------------------------|--------------------------------|--------------------------------|
| `AbortEvent`               | `htmx:abort`                   | `htmx:abort` (unchanged)       |
| `AfterOnLoadEvent`         | `htmx:afterOnLoad`             | `htmx:after:init`              |
| `AfterProcessNodeEvent`    | `htmx:afterProcessNode`        | `htmx:after:init`              |
| `AfterRequestEvent`        | `htmx:afterRequest`            | `htmx:after:request`           |
| `AfterSettleEvent`         | `htmx:afterSettle`             | `htmx:after:swap` (folded)     |
| `AfterSwapEvent`           | `htmx:afterSwap`               | `htmx:after:swap`              |
| `BeforeCleanupElementEvent`| `htmx:beforeCleanupElement`    | `htmx:before:cleanup`          |
| `BeforeOnLoadEvent`        | `htmx:beforeOnLoad`            | `htmx:before:init`             |
| `BeforeProcessNodeEvent`   | `htmx:beforeProcessNode`       | `htmx:before:process`          |
| `BeforeRequestEvent`       | `htmx:beforeRequest`           | `htmx:before:request`          |
| `BeforeSwapEvent`          | `htmx:beforeSwap`              | `htmx:before:swap`             |
| `BeforeSendEvent`          | `htmx:beforeSend`              | `htmx:before:request`          |
| `ConfigRequestEvent`       | `htmx:configRequest`           | `htmx:config:request`          |
| `BeforeHistorySaveEvent`   | `htmx:beforeHistorySave`       | `htmx:before:history:update`   |
| `HistoryRestoreEvent`      | `htmx:historyRestore`          | `htmx:before:history:restore`  |
| `HistoryCacheMissEvent`    | `htmx:historyCacheMiss`        | `htmx:before:history:restore`  |
| `PushedIntoHistoryEvent`   | `htmx:pushedIntoHistory`       | `htmx:after:history:push`      |

**Events — deleted** (no htmx 4 equivalent; downstream must migrate):

- `OobAfterSwapEvent`, `OobBeforeSwapEvent` → use `AfterSwapEvent`/`BeforeSwapEvent`
- `ResponseErrorEvent`, `SendErrorEvent`, `SwapErrorEvent`, `TargetErrorEvent`, `TimeoutEvent`
  → unified as new constant `ErrorEvent = "htmx:error"`
- `ValidationValidateEvent`, `ValidationFailedEvent`, `ValidationHaltedEvent` → no replacement
  (use native form validation)
- `XhrAbortEvent`, `XhrLoadEndEvent`, `XhrLoadStartEvent`, `XhrProgressEvent` → deleted (no XHR)
- `HistoryCacheErrorEvent`, `HistoryCacheMissErrorEvent`, `HistoryCacheMissLoadEvent` → deleted
- `OobErrorNoTargetEvent` → deleted
- `NoSSESourceErrorEvent`, `OnLoadErrorEvent` → deleted

**SSE/WS events**: `SseConnectedEvent`, `SseClosedEvent`, `SseErrorEvent`, `SseBeforeMessageEvent`,
`SseAfterMessageEvent`, `SseConnectingEvent` — verify new names against htmx 4's SSE extension
(extracted to `hx-sse.js` in alpha8) and WS extension (aligned in beta1). Update values to match
upstream; keep Go constant identifiers.

**New event**: `ErrorEvent Event = "htmx:error"`.

**Headers** (`framework/hx/htmx.go` const block starting line 45):

- Delete `TriggerNameHeader`, `TriggerAfterSettleHeader`, `TriggerAfterSwapHeader`, `PromptResponseHeader`.
- `TriggerIdHeader` (`HX-Trigger`) value unchanged, but consumers must handle new `tag#id` format.
- `TargetIdHeader` (`HX-Target`) value unchanged, but now also `tag#id` format.
- Add `SourceHeader Header = "HX-Source"` (new in htmx 4, carries `tag#id`).
- Add `RequestTypeHeader Header = "HX-Request-Type"` (new, value `"full"` or `"partial"`).

**Attributes**:

Order-sensitive rename (both in a single atomic commit):

```
hx-disable      → hx-ignore    (DisableAttr rename, keeping Go identifier semantics)
hx-disabled-elt → hx-disable   (DisabledEltAttr rename)
```

Do not re-order these two edits. `DisableAttr` must point to `hx-ignore` before `DisabledEltAttr`
points to `hx-disable`, otherwise the old `hx-disable` value would collide.

To reduce identifier confusion, rename the Go identifiers too in the same commit:
- `DisableAttr` → `IgnoreAttr` (value `hx-ignore`)
- `DisabledEltAttr` → `DisableAttr` (value `hx-disable`)

Delete: `DisinheritAttr`, `InheritAttr`, `ParamsAttr`, `PromptAttr`, `ExtAttr`, `HistoryAttr`,
`HistoryEltAttr`, `RequestAttr`.

Add: `ConfigAttr Attribute = "hx-config"`, `StatusAttr Attribute = "hx-status"` (new in htmx 4
for per-status targeting/swap control).

### Layer C — Go `framework/h` internals

**`app.go:101-107, 163-170` (`RequestContext`)**:

Current:
```go
func (c *RequestContext) HxTriggerName() string { return c.hxTriggerName }
cc.hxTriggerName = cc.Fiber.Get(hx.TriggerNameHeader)
```

New:
```go
func (c *RequestContext) HxSource() string { return c.hxSource }
// HxSourceID returns the id portion of HX-Source (tag#id format), empty if no id.
func (c *RequestContext) HxSourceID() string { /* split on '#', return tail */ }
cc.hxSource = cc.Fiber.Get(hx.SourceHeader)
```

Also add `HxRequestType() string` for the new `HX-Request-Type` header.

Delete `HxTriggerName()` and the backing field. Audit every caller in examples/, htmgo-site/,
framework-ui/ and rewrite to `HxSourceID()`.

**`lifecycle.go:38-52` (`LifeCycle.OnEvent`)**:

Current logic converts `htmx:afterSwap` → attribute `hx-on::after-swap` using
`util.ConvertCamelToDash`. After Layer B, event constants are already in colon form
(`htmx:after:swap`), and htmx 4 restored the `hx-on::` shortcut (changelog beta1) which takes
colon-separated event names.

New logic:
- If event starts with `htmx:`, strip the prefix and emit `hx-on::<remainder>` directly.
  Example: `htmx:after:swap` → `hx-on::after:swap`.
- Otherwise (DOM events like `click`, `submit`), emit `hx-on:<event>` unchanged.

Delete the `util.ConvertCamelToDash` import path from `lifecycle.go` if it becomes dead.

**`renderer.go:234` (`ToHtmxTriggerName`)**:

Audit this helper's expected event-name format. Update its implementation and tests to match
Layer B's colon-form constants.

**`xhr.go`**:

Builders structurally unchanged. Document in CHANGELOG that `hx-delete` no longer auto-includes
the enclosing form's inputs — downstream apps relying on this must add
`hx-include="closest form"`. The htmgo builder itself does not emit the previous auto-include;
the change only affects user-authored templates.

**`attribute.go`**:

Add inheritance-aware helper constructors for each inheritable attribute. Example:

```go
// HxTargetInherited emits hx-target:inherited="selector" — propagates to descendants.
func HxTargetInherited(selector string) Ren { return Attribute("hx-target:inherited", selector) }
```

Ship helpers for: `target`, `include`, `swap`, `boost`, `confirm`, `headers`, `indicator`,
`sync`, `config`, `encoding`, `validate`. Keep the non-inherited `HxTarget` et al. unchanged.

### Layer D — JS extensions (`framework/assets/js/htmxextensions/`)

All 11 files rewritten 1:1 to htmx 4's API. General pattern:

```ts
// htmx 2
htmx.defineExtension("my-ext", {
  onEvent(name, evt) { if (name === "htmx:beforeSwap") { /* ... */ } },
});

// htmx 4
htmx.registerExtension("my-ext", {
  init(api) { this.api = api; },
  htmx_before_swap(elt, detail) { /* detail.ctx has request context */ },
});
```

Per-file notes:

| File                    | Change notes                                                           |
|-------------------------|------------------------------------------------------------------------|
| `extension.ts`          | Mechanical: `defineExtension` → `registerExtension`.                   |
| `htmgo.ts`              | `htmx:beforeCleanupElement` → `htmx_before_cleanup`; `htmx:load` → `htmx_after_init`. |
| `debug.ts`              | Mechanical rename of event hooks.                                      |
| `livereload.ts`         | Mechanical rename + verify SSE hook still fires (livereload uses SSE). |
| `mutation-error.ts`     | Rename; verify `detail.ctx` access shape.                              |
| `pathdeps.ts`           | Rename; verify `detail.ctx.request` replaces `detail.xhr`.             |
| `trigger-children.ts`   | Rename; verify `htmx.trigger` call signature unchanged.                |
| `response-targets.ts`   | **Largest rewrite.** `evt.detail.xhr` is gone — use `detail.ctx.text` and `detail.ctx.response.status`. Re-derive the `hx-target-4xx` / `4*` / `error` matching ladder against htmx 4's native `hx-status` support. See Risk R3. |
| `sse.ts`                | Port to htmx 4's extracted `hx-sse.js` API (alpha8); reuse upstream event detail shape. |
| `ws.ts`                 | Port to aligned WS API (beta1): `htmx.config.ws`, `{headers, body}` send format, `HX-Request-ID` correlation. See Risk R4. |
| `ws-event-handler.ts`   | Align with `ws.ts`; verify interaction with `extensions/websocket/` Go side. |

### Layer E — Downstream usage (examples, site, templates, framework-ui)

**Inheritance audit procedure**, per directory:

1. `grep -rn 'hx\.\(TargetAttr\|IncludeAttr\|HeadersAttr\|BoostAttr\|IndicatorAttr\|SyncAttr\|ConfirmAttr\|EncodingAttr\|ValidateAttr\|ConfigAttr\)' examples htmgo-site templates framework-ui`
2. For each hit: check whether the containing element also emits `hx.GetAttr`/`PostAttr`/
   `PutAttr`/`PatchAttr`/`DeleteAttr`. If not → it is almost certainly an inheritance parent.
3. Check whether descendant elements in the same file issue `hx-get`/`-post` without declaring
   their own copy of the inheritable attribute. If yes → confirmed inheritance parent.
4. Swap emission to the `:inherited` helper (from Layer C's new `attribute.go` additions).

**Per-example audit checklist** (each may require zero or several fixes):
- `examples/chat/` — chat input, WS flow
- `examples/hackernews/` — sidebar with `hx-boost` inheritance (per grep, already shows 3 hits)
- `examples/simple-auth/` — login/input components
- `examples/todo-list/` — CRUD
- `examples/ws-example/` — repeater partial
- `examples/minimal-htmgo/` — smoke only
- `htmgo-site/pages/docs/` — doc snippets rendered to users' screens
- `htmgo-site/partials/snippets/` — code snippets (may be literal HTML strings, verify)
- `templates/starter/` — starter template seed
- `framework-ui/ui/` — particularly `input.go`

## Testing strategy

### Unit/integration tests (Go)
- `framework/h/*_test.go` — all pass after Layer B-C updates.
- `framework/h/lifecycle_test.go` — add if missing; table-test every `On*` helper → emitted
  `hx-on::` attribute.
- `framework/h/xhr_test.go` — assert inheritance helpers emit `:inherited` suffix.
- `framework/hx/htmx_test.go` — regenerate constant assertions.
- `cli/htmgo/tasks/astgen/` — no semantic change expected; keep green.

### JS extension tests
New `framework/assets/js/htmxextensions/__tests__/<ext>.test.ts` per extension. ~50 lines each,
covers `init(api)` wiring and one representative event hook with mocked `detail.ctx`. Not full
behavioral coverage; just wiring.

### Playwright MCP end-to-end (per CLAUDE.md smoke procedure)

After `cd htmgo-site && npm install && htmgo build` and `PORT=3123 ./dist/htmgo-site &`:

- Homepage, `/docs` redirect, nested doc `/docs/core-concepts/components`, `/examples` — per
  existing CLAUDE.md checks.
- `browser_console_messages` level `error` — must be **0**. htmx 4 logs legacy-attribute usage
  as errors, so this check actively catches missed renames.
- New: one interactive flow per example category — forms, interactivity, projects. Exercise
  at least one `hx-post` or `hx-get` swap per flow.

### Per-example manual verification
Run each `examples/*` under `task watch`, exercise main flow:
- `chat` — WS path (Layer D WS rewrite)
- `todo-list` — `hx-post`/`hx-delete` CRUD
- `simple-auth` — form + redirect header response
- `hackernews` — `hx-boost` sidebar (Layer E inheritance)
- `ws-example` — WS stream
- `minimal-htmgo` — boot sanity

## Rollback

Single PR → rollback is `git revert` of the merge commit. No build flag, no toggle. If htmx
4 beta3/beta4 breaks something after merge, absorb via follow-up PR.

## Risks and open questions

**R1 — htmx 4 beta API churn.** Upstream shipped beta1 (2026-04-06) and beta2 (2026-04-14)
four weeks apart with non-trivial event/config renames. Mitigation: pin exact `4.0.0-beta2`
in `package.json` (no caret). Bump deliberately per beta.

**R2 — hidden inheritance in downstream apps.** Clean break means users with their own htmgo
apps see silent regressions. Mitigation: CHANGELOG "Breaking: attribute inheritance" section
with concrete grep/fix recipe and link to htmx 4 migration docs.

**R3 — `response-targets.ts` semantics drift.** htmx 4's native status-based swap may differ
subtly from the custom extension's `4xx`/`4*`/`*`/`error` matching ladder. Mitigation: write a
vitest mirroring the current ladder before committing the rewrite; diff behavior against native
`hx-status:4xx`.

**R4 — WS extension alignment.** Upstream beta1 WS extension adds `pauseOnBackground`,
`HX-Request-ID` correlation, per-element config, exponential-backoff-with-jitter. Current
`ws.ts` has none of these. Mitigation: 1:1 port preserves current behavior only; file a
follow-up issue to adopt upstream features after this PR lands.

**R5 — `extensions/websocket/` Go module.** Separate module (`github.com/franchb/htmgo/extensions/websocket`)
may encode assumptions about the old WS event-detail shape. To-be-verified during Layer D port;
flag surfaces while writing the extension test.

**O1 — `HxOnLoad` deprecated helper** (`lifecycle.go:67`). Delete it in this PR — clean-break
policy applies. If downstream calls it, CHANGELOG covers the `OnLoad` migration.

**O2 — `framework-ui` release coordination.** `framework-ui` ships as a sibling module.
If it depends on `framework/hx` constants by identifier, Layer B renames cascade. Decision:
release `framework-ui v2.0-beta` in the same PR (or an immediate follow-up PR on the same
branch).

## Deliverables

- Single PR targeting `master`.
- Commits sequenced: Layer A → B → C → D → E → tests → CHANGELOG → VERSION.
- `CHANGELOG.md` entry under `v2.0.0-beta.1` with breaking-change summary and migration recipe.
- Git tag `v2.0.0-beta.1` at the merge commit (matches existing `vX.Y.Z` tagging;
  current HEAD is past `v1.2.3`). No `VERSION` file exists today; do not introduce one.
- Playwright MCP smoke on `htmgo-site` green, 0 console errors.
- All `examples/*` manually exercised.
