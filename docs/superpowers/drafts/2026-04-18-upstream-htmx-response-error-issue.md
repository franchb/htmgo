# Upstream htmx issue draft — `htmx:response:error` convenience event

**Target repo:** `bigskysoftware/htmx`. Verify the correct htmx 4 development branch before posting.

**Posted by:** @franchb (human account). Not to be filed by automated tooling.

---

## Title

htmx 4: no convenience event for HTTP 4xx/5xx responses (regression from htmx 2's `htmx:responseError`)

## Body

### Context

htmx 2 fired `htmx:responseError` on HTTP error statuses, giving consumers a single canonical hook for "request completed, got an error status." Many extensions and downstream user codebases rely on it for logging, error toasts, loading-state cleanup, and per-request error handling.

### Finding

htmx 4.0.0-beta2 core emits **zero** events for HTTP error status codes. The only `htmx:error` event (triggered from the request pipeline's `catch` block) fires on JavaScript exceptions, not on responses with status ≥ 400.

Verified by enumerating every `#trigger` call site in `src/htmx.js` on beta2. Event surface in order: `htmx:config:request`, `htmx:confirm`, `htmx:before:request`, `htmx:before:response`, `htmx:after:request`, `htmx:error` (JS exceptions only), `htmx:finally:request`, `htmx:before:process`, `htmx:after:process`, `htmx:before:cleanup`, `htmx:after:cleanup`, `htmx:before:swap`, `htmx:after:swap`, `htmx:before:settle`, `htmx:after:settle`, history + viewTransition + morph events. No HTTP-error-status event anywhere.

Default `noSwap` is `[204, 304]`, so htmx 4 swaps 4xx/5xx bodies by default — no opt-in needed for display; the missing piece is the observability hook.

### Impact

Every consumer wanting per-request HTTP-error handling must now subscribe to `htmx:before:response` or `htmx:after:request` and manually check `ctx.response.status >= 400`. Third-party extensions built for htmx 2 — `response-targets`, loading indicators, error-toast libraries — lose their migration path.

Concrete example: porting `response-targets` to htmx 4 for the htmgo framework required dropping its `responseTargetSetsError` / `responseTargetUnsetsError` config knobs entirely, because the `isError` flag they mutated no longer exists and no event fires in its place. See [htmgo spec doc](#) for the full analysis.

### Proposal

Add `htmx:response:error` (colon form, matching htmx 4 naming conventions).

- **Fire from:** `#handleStatusCodes`, or immediately after the existing `htmx:before:response` trigger, when `ctx.response.raw.status >= 400`.
- **Detail payload:** `{ctx}` — consistent with surrounding events.
- **Non-breaking:** opt-in listener surface. No existing code paths change.

If the team prefers a config flag gate (e.g. `config.emitResponseError: true`), that works too. Opinion on the default is welcome; the ask is a canonical hook at all, not a specific ergonomic choice.

### Alternatives considered

- **`hx-status:4xx` attribute-driven handling** — works for element-scoped responses but requires annotating every request origin; no app-wide listener surface.
- **Manual status check on `htmx:before:response` / `htmx:after:request`** — what everyone does today. It works but fragments the ecosystem: every library re-implements the same status check with slightly different semantics.

### PR offer

Happy to submit a patch if maintainers agree on the shape. Will match existing code style, add a test in `test/core/events.js` (or wherever the htmx 4 event tests live), and update the event reference docs.
