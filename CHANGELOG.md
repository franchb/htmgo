# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [2.0.0] - 2026-04-23

**Module path migration to `/v2` (Go Semantic Import Versioning).** All importable library modules moved to `/v2` import paths. Update your imports:

- `github.com/franchb/htmgo/framework` → `github.com/franchb/htmgo/framework/v2`
- `github.com/franchb/htmgo/framework-ui` → `github.com/franchb/htmgo/framework-ui/v2`
- `github.com/franchb/htmgo/tools/html-to-htmgo` → `github.com/franchb/htmgo/tools/html-to-htmgo/v2`
- `github.com/franchb/htmgo/extensions/websocket` → `github.com/franchb/htmgo/extensions/websocket/v2`

The CLI (`github.com/franchb/htmgo/cli/htmgo`) is a binary and keeps its path unchanged — continue using `go run github.com/franchb/htmgo/cli/htmgo@latest` (or `@v2.0.0`).

v2.0.0 rolls in everything from the `1.2.0-beta.1` pre-release (htmx 4 migration) plus the Alpine.js compatibility work — see the sections below.

### Breaking

- **htmx upgraded to 4.0.0-beta2.** Clean break with no compat bridge. Exact pin; deliberate bumps per upstream release.
- **`framework/hx` constants:**
  - All `htmx:camelCase` event constants are now `htmx:colon:form` (e.g. `AfterSwapEvent` = `htmx:after:swap`).
  - `AfterSettleEvent` now maps to `htmx:after:settle` (distinct from `AfterSwapEvent`). `AfterProcessNodeEvent` now maps to `htmx:after:process` (distinct from `AfterOnLoadEvent` = `htmx:after:init`). `BeforeSendEvent` is an alias for `BeforeRequestEvent` (both `htmx:before:request`). `HistoryCacheMissEvent` is deprecated (htmx 4 removed localStorage history caching) and aliased to `HistoryRestoreEvent`.
  - SSE event constants now use htmx 4's colon form: `SseConnectedEvent` = `htmx:after:sse:connection`, `SseConnectingEvent` = `htmx:before:sse:connection`, `SseClosedEvent` = `htmx:sse:close`, `SseErrorEvent` = `htmx:sse:error`, `SseBeforeMessageEvent` = `htmx:before:sse:message`, `SseAfterMessageEvent` = `htmx:after:sse:message`.
  - New `ErrorEvent = "htmx:error"` replaces `ResponseErrorEvent`, `SendErrorEvent`, `SwapErrorEvent`, `TargetErrorEvent`, `TimeoutEvent`.
  - Removed: `OobAfterSwapEvent`, `OobBeforeSwapEvent`, `Validation*Event`, `Xhr*Event`, `HistoryCache*ErrorEvent`, `OnLoadErrorEvent`, `NoSSESourceErrorEvent`, `OobErrorNoTargetEvent`.
  - **`hx-disable` semantics changed.** In htmx 2, `hx-disable` stopped htmx attribute processing. In htmx 4, `hx-disable` *disables form elements during an in-flight request* (the role formerly played by `hx-disabled-elt`). The "stop processing" role moved entirely to `hx-ignore`. `DisableAttr` now maps to the new (disable-elt) semantic; use `IgnoreAttr` for the old (stop-processing) semantic.
  - Removed attribute constants: `DisabledEltAttr`, `DisinheritAttr`, `InheritAttr`, `ParamsAttr`, `PromptAttr`, `ExtAttr`, `HistoryAttr`, `HistoryEltAttr`, `RequestAttr`.
  - New attribute constants: `ConfigAttr` (`hx-config`, replaces `hx-request`), `StatusAttr` (`hx-status`).
  - Removed header constants: `PromptResponseHeader`, `TriggerNameHeader`, `TriggerAfterSettleHeader`, `TriggerAfterSwapHeader`.
  - New header constants: `SourceHeader` (`HX-Source`), `RequestTypeHeader` (`HX-Request-Type`).
- **`RequestContext.HxTriggerName()` removed** — use `HxSource()` (raw `tag#id`) or `HxSourceID()` (just the id).
- **`RequestContext.HxPromptResponse()` removed** — htmx 4 removed `hx-prompt`; use `hx-confirm="js:..."` with a JS async prompt function.
- **`HxOnLoad` deprecated helper removed** — use `OnLoad`.
- **`HxExtension`/`HxExtensions`/`TriggerChildren`/`JoinExtensions`/`BaseExtensions` removed.** htmx 4 removed `hx-ext`; extensions self-register on script import.
- **Attribute inheritance is explicit.** htmx 4 removed implicit inheritance. If your app set `hx-target` (or any of `hx-include`, `hx-headers`, `hx-boost`, `hx-indicator`, `hx-sync`, `hx-confirm`, `hx-encoding`, `hx-validate`, `hx-config`, `hx-swap`) on a container expecting it to apply to descendants, switch to the `:inherited` form.
  New helpers: `HxTargetInherited`, `HxIncludeInherited`, `HxSwapInherited`, `HxBoostInherited`, `HxConfirmInherited`, `HxHeadersInherited`, `HxIndicatorInherited`, `HxSyncInherited`, `HxEncodingInherited`, `HxValidateInherited`. (`hx-config` does not participate in attribute inheritance; configure it globally via `htmx.config` or a `<meta name="htmx-config">` tag.)
- **`hx-delete` on a form button no longer includes form data.** Add `hx-include="closest form"` where needed.
- **All JS extensions rewritten for htmx 4's `registerExtension` API.** Event-detail access changed from `detail.xhr` to `detail.ctx`. The following extensions ship in the bundle: `htmgo`, `debug`, `livereload`, `mutation-error`, `path-deps`, `trigger-children`, `response-targets`, `sse`, `ws`, `ws-event-handler`.
- **`trigger-children` fans out only a fixed allowlist of `htmx:*` events.** htmx 2's version listened via `onEvent(name, evt)` — a catch-all that received every htmx event including custom ones emitted by `htmx.trigger` / `HX-Trigger` response headers. htmx 4's extension API has no equivalent hook, and the DOM has no wildcard event listener, so the new implementation enumerates the events it fans out at `document`-level (see `framework/assets/js/htmxextensions/trigger-children.ts`). Custom events dispatched via `htmx.trigger` or `HX-Trigger` will NOT reach descendants via `trigger-children` unless their name is added to that list.
- **`livereload` extension now gates on `<meta name="htmgo-livereload">`** (emitted automatically by the framework in dev mode).
- **`response-targets` extension no longer reads or initializes `htmx.config.responseTargetSetsError` / `responseTargetUnsetsError`.** These htmx 2 config keys have no htmx 4 equivalent — htmx 4 removed the `isError` flag and the `htmx:responseError` event. If you relied on them, subscribe to `htmx:before:response` and inspect `ctx.response.status` directly.

### Added

- `HxSource()`, `HxSourceID()`, `HxRequestType()` on `RequestContext` (htmx 4 `HX-Source` and `HX-Request-Type` headers).
- `StatusAttr` constant (`hx-status`) for htmx 4's per-status swap/target control.
- `ConfigAttr` constant (`hx-config`).
- 10 `Hx*Inherited` attribute helpers for explicit-inheritance emission.
- `vitest` test suite for JS extensions (`framework/assets/js/htmxextensions/__tests__/`).
- `<meta name="htmgo-livereload">` injected into page `<head>` in dev mode so the livereload extension can detect dev server presence.
- `htmgo:response:retargeted` event fired by the `response-targets` extension on the source element when a 4xx/5xx response triggers a retarget. Detail: `{status, from, to, ctx}`. Gives consumers a canonical hook without reintroducing htmx-2 error-flag semantics.
- **`alpine-compat` htmx extension bundled into `htmgo.js`.** Auto-preserves Alpine.js state across htmx morph swaps and re-initializes Alpine on swapped content. Auto-gates on `window.Alpine` presence; extension no-ops when Alpine is not loaded (zero runtime cost, ~3KB gz bundle cost). Tested against Alpine v3.15.11.
- **`framework/ax/` Go package.** Constants and Ren-returning builder helpers for Alpine.js directives, mirroring the shape of `framework/hx/`. Covers `Data`, `Show`, `Init`, `Text`, `Html`, `Model` (plus `.number`/`.lazy`/`.trim`/`.fill`/`.boolean`/`.debounce` variants), `Bind` + 7 shortcuts (`BindClass`, `BindStyle`, `BindHref`, `BindValue`, `BindDisabled`, `BindChecked`, `BindId`), `On` + 8 event shortcuts + 3 combos (`OnClickOutside`, `OnKeydownEscape`, `OnKeydownEnter`), `Cloak`, `Ignore`, `Ref`, `Teleport`, `Effect`, `If`, `For`, `Id`, `Modelable`, `Transition`.
- **New doc page:** `/docs/htmx-extensions/alpine-compat` — covers loading Alpine, FOUC prevention, `ax` package usage, and Alpine/htmx swap interaction.

### Migration recipe

Find inheritance parents:
```bash
grep -rn 'h\.HxTarget\|h\.HxInclude\|h\.HxConfirm\|h\.HxIndicator\|h\.Boost\|hx\.TargetAttr\|hx\.IncludeAttr\|hx\.HeadersAttr\|hx\.BoostAttr\|hx\.IndicatorAttr\|hx\.SyncAttr\|hx\.ConfirmAttr\|hx\.EncodingAttr\|hx\.ValidateAttr\|hx\.ConfigAttr' your-app/
```

For each hit where the element has no `h.Get`/`h.Post`/`h.Put`/`h.Patch`/`h.Delete` of its own but its descendants do → switch to the `:inherited` helper.

See `docs/superpowers/specs/2026-04-17-htmx-v4-migration-design.md` for the full rationale and `docs/superpowers/plans/2026-04-17-htmx-v4-migration.md` for the implementation plan.
