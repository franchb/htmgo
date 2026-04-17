# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [2.0.0-beta.1] - 2026-04-17

### Breaking

- **htmx upgraded to 4.0.0-beta2.** Clean break with no compat bridge. Exact pin; deliberate bumps per upstream release.
- **`framework/hx` constants:**
  - All `htmx:camelCase` event constants are now `htmx:colon:form` (e.g. `AfterSwapEvent` = `htmx:after:swap`).
  - `AfterSettleEvent` folded into `AfterSwapEvent`. `AfterOnLoadEvent`/`AfterProcessNodeEvent` both map to `htmx:after:init`. `BeforeSendEvent` maps to `htmx:before:request`.
  - New `ErrorEvent = "htmx:error"` replaces `ResponseErrorEvent`, `SendErrorEvent`, `SwapErrorEvent`, `TargetErrorEvent`, `TimeoutEvent`.
  - Removed: `OobAfterSwapEvent`, `OobBeforeSwapEvent`, `Validation*Event`, `Xhr*Event`, `HistoryCache*ErrorEvent`, `OnLoadErrorEvent`, `NoSSESourceErrorEvent`, `OobErrorNoTargetEvent`.
  - Attribute `DisableAttr` now maps to `"hx-disable"` (disable-elt semantics). New `IgnoreAttr` maps to `"hx-ignore"` (formerly `hx-disable` in htmx 2).
  - Removed attribute constants: `DisabledEltAttr`, `DisinheritAttr`, `InheritAttr`, `ParamsAttr`, `PromptAttr`, `ExtAttr`, `HistoryAttr`, `HistoryEltAttr`, `RequestAttr`.
  - New attribute constants: `ConfigAttr` (`hx-config`, replaces `hx-request`), `StatusAttr` (`hx-status`).
  - Removed header constants: `PromptResponseHeader`, `TriggerNameHeader`, `TriggerAfterSettleHeader`, `TriggerAfterSwapHeader`.
  - New header constants: `SourceHeader` (`HX-Source`), `RequestTypeHeader` (`HX-Request-Type`).
- **`RequestContext.HxTriggerName()` removed** — use `HxSource()` (raw `tag#id`) or `HxSourceID()` (just the id).
- **`RequestContext.HxPromptResponse()` removed** — htmx 4 removed `hx-prompt`; use `hx-confirm="js:..."` with a JS async prompt function.
- **`HxOnLoad` deprecated helper removed** — use `OnLoad`.
- **`HxExtension`/`HxExtensions`/`TriggerChildren`/`JoinExtensions`/`BaseExtensions` removed.** htmx 4 removed `hx-ext`; extensions self-register on script import.
- **Attribute inheritance is explicit.** htmx 4 removed implicit inheritance. If your app set `hx-target` (or any of `hx-include`, `hx-headers`, `hx-boost`, `hx-indicator`, `hx-sync`, `hx-confirm`, `hx-encoding`, `hx-validate`, `hx-config`, `hx-swap`) on a container expecting it to apply to descendants, switch to the `:inherited` form.
  New helpers: `HxTargetInherited`, `HxIncludeInherited`, `HxSwapInherited`, `HxBoostInherited`, `HxConfirmInherited`, `HxHeadersInherited`, `HxIndicatorInherited`, `HxSyncInherited`, `HxConfigInherited`, `HxEncodingInherited`, `HxValidateInherited`.
- **`hx-delete` on a form button no longer includes form data.** Add `hx-include="closest form"` where needed.
- **All JS extensions rewritten for htmx 4's `registerExtension` API.** Event-detail access changed from `detail.xhr` to `detail.ctx`. The following extensions ship in the bundle: `htmgo`, `debug`, `livereload`, `mutation-error`, `path-deps`, `trigger-children`, `response-targets`, `sse`, `ws`, `ws-event-handler`.
- **`livereload` extension now gates on `<meta name="htmgo-livereload">`** (emitted automatically by the framework in dev mode).

### Added

- `HxSource()`, `HxSourceID()`, `HxRequestType()` on `RequestContext` (htmx 4 `HX-Source` and `HX-Request-Type` headers).
- `StatusAttr` constant (`hx-status`) for htmx 4's per-status swap/target control.
- `ConfigAttr` constant (`hx-config`).
- 11 `Hx*Inherited` attribute helpers for explicit-inheritance emission.
- `vitest` test suite for JS extensions (`framework/assets/js/htmxextensions/__tests__/`).
- `<meta name="htmgo-livereload">` injected into page `<head>` in dev mode so the livereload extension can detect dev server presence.

### Migration recipe

Find inheritance parents:
```bash
grep -rn 'h\.HxTarget\|h\.HxInclude\|h\.HxConfirm\|h\.HxIndicator\|h\.Boost\|hx\.TargetAttr\|hx\.IncludeAttr\|hx\.HeadersAttr\|hx\.BoostAttr\|hx\.IndicatorAttr\|hx\.SyncAttr\|hx\.ConfirmAttr\|hx\.EncodingAttr\|hx\.ValidateAttr\|hx\.ConfigAttr' your-app/
```

For each hit where the element has no `h.Get`/`h.Post`/`h.Put`/`h.Patch`/`h.Delete` of its own but its descendants do → switch to the `:inherited` helper.

See `docs/superpowers/specs/2026-04-17-htmx-v4-migration-design.md` for the full rationale and `docs/superpowers/plans/2026-04-17-htmx-v4-migration.md` for the implementation plan.
