# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **Automatic releases.** Merges to master are tagged and released per module from Conventional Commit subjects. Each module continues from its own latest tag (`framework/v2.0.4` → `framework/v2.1.0` while the others go from v2.1.0 to v2.1.1), and the repository tag moves too. Releases attach reproducible `htmgo` binaries for Linux, macOS and Windows (amd64/arm64), with `SHA256SUMS` and GitHub build provenance. Rules are in CONTRIBUTING.md, "Releases". The planner is `tools/release`, a stdlib-only Go module.
- **`htmgo version`** reports the real release version (from the release build, or from `go install …@vX.Y.Z`), or `dev` for a source build. Before, it always printed `2.0.0`.
- **`framework/h`: opt-in CSP nonces.** Set `AppOpts.ContentSecurityPolicy` (for example `"script-src 'nonce-{nonce}' 'strict-dynamic'; object-src 'none'; base-uri 'none'"`). Each full page then gets a fresh 128-bit nonce on every `<script>` htmgo renders, a `Content-Security-Policy` header with `{nonce}` replaced, and a `<meta name="htmx-config">` that makes htmx 4 put the nonce on scripts in swapped partials. Cached components get the current request's nonce. `RequestContext.CSPNonce()` and `h.WithScriptNonce` are available for custom setups. With the option empty, output is unchanged. Not covered: inline event handlers (`h.OnClick` and friends render `onclick=` attributes, and `hx-on` needs `'unsafe-eval'`). Note that `inlineScriptNonce` makes htmx trust every script in a swapped response, so inside partials CSP no longer guards against injected scripts; escaping still does.
- **`framework/h`: `h.Validation`** with `ValidateRequired`, `ValidatePattern`, `ValidatePartial`/`ValidatePath`, `ValidateOn`, `ValidateTarget` and `ValidateSwap`. It renders only attributes: native `required`/`pattern` (+`title`) checks, plus optional server validation that `hx-post`s the field to a partial on change and swaps the reply into the next `<div>`, with `hx-sync="this:queue last"`. No JavaScript, so it is CSP-friendly. Design notes and next steps are in `docs/design/validation.md`.

## [2.2.0] - Unreleased (draft)

Draft for the next release, covering master since v2.1.0 plus the security, CI and hygiene fixes in #65–#68. Not tagged yet.

### Breaking

- **Minimum Go version is now 1.27.0** in every module (#58). Downstream modules that depend on any htmgo v2 module must build with Go ≥ 1.27.0, or let `GOTOOLCHAIN=auto` fetch it.
- **`framework/h`: the unused `Entry` type is removed.** It had only unexported fields and nothing referenced it, but it was exported.

### Security

- **`golang.org/x/net` v0.62.0** in every module. Fixes GO-2026-6860, which was reachable from `tools/html-to-htmgo` through `html.Parse`, plus GO-2026-6866, GO-2026-6617, GO-2026-6612, GO-2026-6611, GO-2026-6610 and GO-2026-6603. `golang.org/x/crypto`, `x/sys`, `x/text` and `x/tools` move up with it.

### Fixed

- **`tools/html-to-htmgo`:** conversion no longer fails with `ParseErr` when a node, such as a comment, follows `</html>`. `findBody` used to keep only the last child's result.
- **`cli/htmgo`:** data races in the process runner, which `htmgo watch` hits. The running-command map is now locked, and a command is published only after `Start` has set `cmd.Process`.
- **`extensions/websocket`:** `SocketManager.Get` returns nil for an unknown connection id instead of a pointer to a zero `SocketConnection`.
- **`cli/htmgo`:** removed an unreachable duplicate `generate` branch. Behaviour is unchanged.

### Changed

- All dependencies bumped (#57). Notable: `gofiber/fiber/v3` v3.5.0, `valyala/fasthttp` v1.74.0, TypeScript 7, vitest 5, jsdom 30.
- `cli/htmgo` astgen no longer calls the deprecated `strings.Title`. Generated asset constant names are unchanged; a test now pins them.
- Docker images for htmgo-site, the starter template and the examples are built from the same commit as the framework, use Debian-based Go builders, and run on `distroless/base-debian12`.
- The README gained a "What this fork adds" section (#41).

### CI

- One `Tests` workflow runs `go vet` and `go test -race` for every library module on Go 1.27.x and stable, plus the vitest suite for the JS extensions.
- New workflows: govulncheck for every library module, and golangci-lint.
- Fixed the Cloudflare docs deploy, the GHCR release workflows and the framework-dependency updater.
- Actions are pinned to commit SHAs, and workflow tokens are read-only by default.

## [2.1.0] - 2026-07-09

Tags: `cli/htmgo/v2.1.0`, `framework-ui/v2.1.0`, `extensions/websocket/v2.1.0`, `tools/html-to-htmgo/v2.1.0`, `framework/v2.0.4` (all at the same commit) and the repository tag `v2.1.0`. `framework/v2.0.3` (2026-07-08) carried the same `FiberConfig` feature.

### Added

- **`AppOpts.FiberConfig *fiber.Config`** (#36). When non-nil, it is passed verbatim to `fiber.New`, which gives access to body limits, timeouts, proxy headers, error handlers and the rest. Nil keeps the previous defaults.

### Changed

- `gofiber/fiber/v3` v3.4.0, `valyala/fasthttp` v1.72.0, `golang.org/x/net` v0.56.0 (Dependabot).

### Fixed

- Docker images and the docs deploy pinned Go 1.23, while all modules have required Go ≥ 1.26 since v2.0.0, so every deploy failed with `go.work requires go >= 1.26` (#37).

## [2.0.2] - 2026-06-09

All five v2 modules and the repository were tagged `v2.0.2` at the same commit.

### Changed

- `gofiber/fiber/v3` v3.3.0, `valyala/fasthttp` v1.71.0, `golang.org/x/net` v0.55.0 (Dependabot).
- Dependabot is configured for the Go modules, npm and GitHub Actions.

## [2.0.1] - 2026-04-24

### Fixed

- **`cli/htmgo` install.** `cli/htmgo/go.mod` now declares the `/v2` module path required by Go's Semantic Import Versioning rules (`module github.com/franchb/htmgo/cli/htmgo/v2`). The install invocation is now `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1`. The binary name (`htmgo`) is unchanged. See [#14](https://github.com/franchb/htmgo/issues/14).
- **Sibling `replace` directives removed.** `cli/htmgo`, `framework-ui`, and `extensions/websocket` `go.mod` files no longer contain `replace github.com/franchb/htmgo/... => ../...` lines. These violated Go's rule that non-main modules must not carry local-path replaces and blocked all three from `go get` / `go install`. Local monorepo development now uses a root `go.work` file.

### Changed

- **Install path for `go install` / `go run` users.** All docs (`htmgo-site`), Dockerfiles, and Taskfiles under `examples/`, `templates/`, and `htmgo-site/` now reference `github.com/franchb/htmgo/cli/htmgo/v2@latest` instead of `github.com/franchb/htmgo/cli/htmgo@latest`. The old path still resolves to the v1.1.1 CLI (the last tag under the v1 module path), which is incompatible with `framework/v2`.

### Chore

- `framework/v2.0.1` and `tools/html-to-htmgo/v2.0.1` are republished at the fix commit with **no code change** so all five v2 submodule versions stay aligned.

## [2.0.0] - 2026-04-23

**Module path migration to `/v2` (Go Semantic Import Versioning).** All importable library modules moved to `/v2` import paths. Update your imports:

- `github.com/franchb/htmgo/framework` → `github.com/franchb/htmgo/framework/v2`
- `github.com/franchb/htmgo/framework-ui` → `github.com/franchb/htmgo/framework-ui/v2`
- `github.com/franchb/htmgo/tools/html-to-htmgo` → `github.com/franchb/htmgo/tools/html-to-htmgo/v2`
- `github.com/franchb/htmgo/extensions/websocket` → `github.com/franchb/htmgo/extensions/websocket/v2`

> **Correction (v2.0.1):** The claim that the CLI path was unchanged was incorrect. Go's Semantic Import Versioning applies to any submodule with its own `go.mod` at major ≥ 2, and the `v2.0.0` CLI tags are un-installable as a result. See [CHANGELOG v2.0.1](#201---2026-04-24) and [#14](https://github.com/franchb/htmgo/issues/14). Use `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1` or later.

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
