# Design: `response-targets` error-flag rewire for htmx 4

**Date:** 2026-04-18
**Status:** approved, ready for implementation plan
**Origin:** follow-up from CodeRabbit review on PR #12; referenced by `docs/plans/2026-04-18-response-targets-error-flag-followup.md`.
**Target release:** v2.0.0-beta.1

## Problem

`framework/assets/js/htmxextensions/response-targets.ts` contains htmx 2-era error-flag logic that no longer functions in htmx 4. Specifically:

- `handleErrorFlag(detail)` mutates `detail.isError`, but htmx 4's `ctx` has no `isError` property. Verified against `node_modules/htmx.org/dist/htmax.js` and `htmx.js` (beta2): zero occurrences.
- The `responseTargetSetsError` and `responseTargetUnsetsError` config knobs are initialized but gate nothing observable.
- A rename to `handleErrorFlag(detail)` does not fix it — `detail.isError` is also absent in htmx 4's `before:swap` payload.

Deeper investigation found htmx 4 has no `isError`-equivalent surface at all:

- No `htmx:responseError` event (removed from htmx 2).
- `htmx:error` fires only on JS exceptions in the request pipeline, not on HTTP 4xx/5xx.
- Default `noSwap: [204, 304]` means htmx 4 swaps 4xx/5xx bodies by default — no opt-in needed.
- No error CSS class toggle tied to status.

Consequence: the two config knobs have no clean htmx 4 mapping. "Rewiring" them would mean inventing semantics, not preserving parity.

## Decision

Three-part plan:

1. **Drop the dead config knobs and `handleErrorFlag` call sites** from `response-targets.ts`.
2. **Add one canonical observability hook** — `htmgo:response:retargeted` event — so users can react to response-targets retargets without reintroducing htmx-2-era flag semantics.
3. **File an upstream htmx issue** proposing a native `htmx:response:error` convenience event, independent of our extension changes. Not a release blocker.

Users wanting generic HTTP-error handling in htmx 4 subscribe to `htmx:before:response` and inspect `ctx.response.status >= 400` — the same surface every other htmx 4 consumer uses.

## Extension changes

**File:** `framework/assets/js/htmxextensions/response-targets.ts`

### Remove

- `handleErrorFlag` function (current lines 64–72).
- Init of `config.responseTargetUnsetsError` (lines 78–80).
- Init of `config.responseTargetSetsError` (lines 81–83).
- All three `handleErrorFlag(ctx)` call sites inside `htmx_before_swap` (current lines 108, 115, 125).

### Keep unchanged

- `responseTargetPrefersExisting` and `responseTargetPrefersRetargetHeader` config — these gate real htmx 4 behavior (whether to override an existing `mainTask.target` or honor `HX-Retarget`) and are unrelated to the `isError` issue.
- Retargeting logic operating on `mainTask.target`.

### Add

When a retarget actually happens (a new target is resolved and assigned to `mainTask.target`), emit `htmgo:response:retargeted` on `ctx.sourceElement` via `api.triggerHtmxEvent`.

## Event contract

**Name:** `htmgo:response:retargeted`

**Dispatched on:** `ctx.sourceElement ?? ctx.elt` (i.e., the same element used to resolve the target; bubbling).

**Mechanism:** `api.triggerHtmxEvent(ctx.sourceElement ?? ctx.elt, "htmgo:response:retargeted", detail)`. The `api` ref is already captured in `init`.

**Detail payload:**

```ts
{
  status: number,        // ctx.response.status
  from: Element | null,  // previous mainTask.target (may be null)
  to: Element,           // new target resolved via getRespCodeTarget
  ctx: any               // full htmx 4 swap context (escape hatch)
}
```

**Fires when:**

- `status` is non-zero and not 200
- A matching `hx-target-*` attribute was found
- A new target was assigned to `mainTask.target`

**Does not fire when:**

- Status is 0 or 200 (early return)
- `responseTargetPrefersExisting=true` and an existing `mainTask.target` was kept
- `responseTargetPrefersRetargetHeader=true` and an `HX-Retarget` response header was honored
- No matching `hx-target-*` attribute was found

**Naming:** `htmgo:` prefix signals this is the htmgo extension's event, not core htmx. Colon form matches htmx 4 conventions (`htmx:before:swap`, `htmx:after:settle`).

## Tests

**File:** `framework/assets/js/htmxextensions/__tests__/response-targets.test.ts`

Replace the existing defaults-only test with the following vitest cases. Use a minimal htmx api stub (mock `getClosestAttributeValue`, `findThisElement`, `querySelectorExt`, `triggerHtmxEvent`) following the existing file's stub pattern.

1. **Config defaults** — assert `responseTargetPrefersExisting` and `responseTargetPrefersRetargetHeader` are initialized. Assert `responseTargetUnsetsError` and `responseTargetSetsError` are NOT set (regression guard).
2. **Retarget fires event** — simulate `htmx:before:swap` with a 404 response and a matching `hx-target-404` ancestor. Assert `htmgo:response:retargeted` fires once on `ctx.sourceElement` with expected detail fields; assert `mainTask.target` reflects the new target.
3. **No match, no event** — 404 response, no `hx-target-*` anywhere. Assert event does NOT fire and `mainTask.target` is unchanged.
4. **PrefersExisting suppresses event** — `mainTask.target` already set, `responseTargetPrefersExisting=true`. Assert event does NOT fire.
5. **HX-Retarget suppresses event** — `responseTargetPrefersRetargetHeader=true` and `HX-Retarget` response header present. Assert event does NOT fire.
6. **2xx early return** — status 200; assert no event, no target mutation.

## CHANGELOG

Under the existing `1.2.0-beta.1` section, add:

> **BREAKING:** `response-targets` extension no longer reads or initializes `htmx.config.responseTargetSetsError` / `responseTargetUnsetsError`. These htmx 2 config keys have no htmx 4 equivalent (htmx 4 removed the `isError` flag and `htmx:responseError` event). If you relied on them, subscribe to `htmx:before:response` and inspect `ctx.response.status` directly.
>
> **NEW:** `htmgo:response:retargeted` event fires on the source element when `response-targets` reassigns the swap target based on a 4xx/5xx status. Detail: `{status, from, to, ctx}`.

## Rebuild

After code + test changes, rebuild the bundle:

```bash
cd framework/assets/js
npm test
npm run build
```

Commit the regenerated `framework/assets/dist/htmgo.js`.

## Upstream htmx issue (parallel, not a blocker)

**Repo:** `bigskysoftware/htmx` (verify the correct branch for htmx 4 development before filing).

**Title:** `htmx 4: no convenience event for HTTP 4xx/5xx responses (regression from htmx 2's htmx:responseError)`

**Body outline:**

1. **Context** — htmx 2 fired `htmx:responseError` on HTTP error statuses, giving consumers one canonical hook. Many extensions and user codebases rely on it.
2. **Finding** — htmx 4.0.0-beta2 core emits zero events for HTTP error status codes. `htmx:error` fires only on JS exceptions, not on status ≥ 400. Verified by enumerating `#trigger` call sites in `src/htmx.js`.
3. **Impact** — consumers wanting per-request error handling must subscribe to `htmx:before:response` or `htmx:after:request` and manually check `ctx.response.status`. Third-party extensions built for htmx 2 (e.g. `response-targets`, loading states, error toasts) lose their migration path. Concrete example: the htmgo fork's `response-targets` port to htmx 4 had to drop its `responseTargetSetsError`/`UnsetsError` config knobs because the surface they gated is gone.
4. **Proposal** — add `htmx:response:error` (colon form), fired from `#handleStatusCodes` or immediately after `htmx:before:response` when `ctx.response.raw.status >= 400`. Detail: `{ctx}`. Non-breaking; opt-in for listeners. If the team prefers a config flag (e.g. `config.emitResponseError`), that's also fine.
5. **Alternatives considered** — `hx-status:4xx` attribute-driven handling works for element-scoped cases but not for app-wide error listeners.
6. **PR offer** — willing to submit a patch if maintainers agree on the shape.

**Who files:** @franchb posts, not Claude. Maintainers respond better to humans.

**Status:** parallel track; ship the extension changes regardless of upstream outcome.

## Out of scope

- Inventing an htmgo-owned error-class toggle or per-status CSS signaling.
- Reintroducing any mutable error flag on `ctx`.
- Changes to `responseTargetPrefersExisting` / `responseTargetPrefersRetargetHeader`.
- Any other htmx extension in `framework/assets/js/htmxextensions/`.
