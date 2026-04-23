# response-targets error-flag rewire — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove htmx-2-era `isError` plumbing from the `response-targets` extension, add a `htmgo:response:retargeted` observability event, update CHANGELOG, rebuild the JS bundle, and draft an upstream htmx issue proposing `htmx:response:error`.

**Architecture:** Single TypeScript extension file plus its vitest suite. TDD the behavior changes: update failing defaults test → remove dead code → add failing event-firing tests → implement event dispatch → add failing suppression tests → verify. Then CHANGELOG, rebuild, upstream issue draft.

**Tech Stack:** TypeScript, htmx 4.0.0-beta2, vitest, tsup.

**Spec:** `docs/superpowers/specs/2026-04-18-response-targets-error-flag-design.md`

---

## File Structure

- Modify: `framework/assets/js/htmxextensions/response-targets.ts` — drop `handleErrorFlag` + two config knobs; add `htmgo:response:retargeted` dispatch.
- Modify: `framework/assets/js/htmxextensions/__tests__/response-targets.test.ts` — update defaults test, add event-firing and suppression tests.
- Modify: `CHANGELOG.md` — document breaking change + new event under existing `[1.2.0-beta.1]` section.
- Regenerate: `framework/assets/dist/htmgo.js` — rebuild via `npm run build`.
- Create: `docs/superpowers/drafts/2026-04-18-upstream-htmx-response-error-issue.md` — draft upstream issue body for @franchb to post.

All work stays in the current branch `htmx4-migration`.

---

### Task 1: Update defaults test to drop `responseTargetSetsError` / `responseTargetUnsetsError`

**Files:**
- Modify: `framework/assets/js/htmxextensions/__tests__/response-targets.test.ts` (the `init sets defaults` test, currently lines 36–58)

The current defaults test asserts four config keys including the two we're removing. We flip it first: the test must *fail* once we assert the two knobs are NOT set, then pass after Task 2 removes their init.

- [ ] **Step 1: Rewrite the defaults test to regression-guard the removal**

Replace the existing test body (lines 36–58) with:

```ts
  it("init sets defaults in htmx.config without clobbering existing values", async () => {
    const cfg = (await import("htmx.org")).default.config as any;
    delete cfg.responseTargetUnsetsError;
    delete cfg.responseTargetSetsError;
    cfg.responseTargetPrefersExisting = undefined;
    cfg.responseTargetPrefersRetargetHeader = undefined;
    ext.init(makeApi({}));
    // Regression guard: the htmx-2 isError knobs must NOT be initialized.
    expect("responseTargetUnsetsError" in cfg).toBe(false);
    expect("responseTargetSetsError" in cfg).toBe(false);
    expect(cfg.responseTargetPrefersExisting).toBe(false);
    expect(cfg.responseTargetPrefersRetargetHeader).toBe(true);

    // Pre-set non-default values on the surviving knobs must be preserved on re-init.
    cfg.responseTargetPrefersExisting = true;
    cfg.responseTargetPrefersRetargetHeader = false;
    ext.init(makeApi({}));
    expect(cfg.responseTargetPrefersExisting).toBe(true);
    expect(cfg.responseTargetPrefersRetargetHeader).toBe(false);
  });
```

- [ ] **Step 2: Run the suite — the defaults test must fail**

Run: `cd framework/assets/js && npx vitest run htmxextensions/__tests__/response-targets.test.ts`

Expected: the "init sets defaults" test FAILS because `responseTargetUnsetsError` / `responseTargetSetsError` are still initialized by the current `response-targets.ts`. All other tests should still pass.

- [ ] **Step 3: Commit**

```bash
git add framework/assets/js/htmxextensions/__tests__/response-targets.test.ts
git commit -m "test(response-targets): regression-guard removal of isError knobs"
```

---

### Task 2: Remove `handleErrorFlag` and its config knobs

**Files:**
- Modify: `framework/assets/js/htmxextensions/response-targets.ts` (remove lines 64–72, 78–83, and the three `handleErrorFlag(ctx)` call sites inside `htmx_before_swap`)

- [ ] **Step 1: Delete the `handleErrorFlag` function**

Remove these lines (currently 64–72):

```ts
function handleErrorFlag(detail: any) {
  if (detail.isError) {
    if (config.responseTargetUnsetsError) {
      detail.isError = false;
    }
  } else if (config.responseTargetSetsError) {
    detail.isError = true;
  }
}
```

- [ ] **Step 2: Delete the two knob initializers inside `init`**

Inside the `init` body, remove:

```ts
    if (config.responseTargetUnsetsError === undefined) {
      config.responseTargetUnsetsError = true;
    }
    if (config.responseTargetSetsError === undefined) {
      config.responseTargetSetsError = false;
    }
```

Leave the `responseTargetPrefersExisting` and `responseTargetPrefersRetargetHeader` initializers intact.

- [ ] **Step 3: Remove the three `handleErrorFlag(ctx)` call sites from `htmx_before_swap`**

Inside `htmx_before_swap`, delete each `handleErrorFlag(ctx);` line. There are three of them; after removal the control flow is:

```ts
    if (mainTask.target) {
      if (config.responseTargetPrefersExisting) {
        return;
      }
      const headers = ctx?.response?.headers;
      const retarget =
        typeof headers?.get === "function" ? headers.get("HX-Retarget") : null;
      if (config.responseTargetPrefersRetargetHeader && retarget) {
        return;
      }
    }

    const reqElt = ctx?.sourceElement ?? ctx?.elt;
    if (!reqElt) return;

    const target = getRespCodeTarget(reqElt, status);
    if (target) {
      mainTask.target = target;
    }
```

- [ ] **Step 4: Run the suite — everything must pass**

Run: `cd framework/assets/js && npx vitest run htmxextensions/__tests__/response-targets.test.ts`

Expected: all tests PASS, including the defaults regression guard from Task 1.

- [ ] **Step 5: Commit**

```bash
git add framework/assets/js/htmxextensions/response-targets.ts
git commit -m "refactor(response-targets): drop dead htmx-2 isError plumbing

htmx 4 has no isError flag, no htmx:responseError event, and swaps 4xx/5xx
by default. The handleErrorFlag function and its two config knobs
(responseTargetSetsError, responseTargetUnsetsError) were no-ops after the
htmx 4 port. Remove them. See docs/superpowers/specs/2026-04-18-response-targets-error-flag-design.md."
```

---

### Task 3: Add failing test — event fires on successful retarget

**Files:**
- Modify: `framework/assets/js/htmxextensions/__tests__/response-targets.test.ts` (extend `makeApi` to record event triggers, add new test)

- [ ] **Step 1: Extend `makeApi` to capture `triggerHtmxEvent` calls**

Replace the existing `makeApi` helper (lines 11–20) with:

```ts
function makeApi(attrs: Record<string, string>) {
  const triggered: Array<{ elt: any; name: string; detail: any }> = [];
  const api = {
    getClosestAttributeValue: (_elt: any, name: string) => attrs[name] ?? null,
    findThisElement: (elt: any) => elt,
    querySelectorExt: (_elt: any, sel: string) => {
      const el = document.querySelector(sel);
      return el;
    },
    triggerHtmxEvent: (elt: any, name: string, detail: any) => {
      triggered.push({ elt, name, detail });
    },
  };
  return { api, triggered };
}
```

- [ ] **Step 2: Update every existing `ext.init(makeApi(...))` call site to use the new shape**

The helper now returns `{api, triggered}`. Every existing `ext.init(makeApi(...))` becomes `ext.init(makeApi(...).api)` unless the test needs `triggered`. Update:

- Line ~42: `ext.init(makeApi({}));` → `ext.init(makeApi({}).api);` (appears twice in the defaults test)
- Line ~74: `ext.init(makeApi({ "hx-target-404": "#err" }));` → `ext.init(makeApi({ "hx-target-404": "#err" }).api);`
- Line ~83: same pattern for `hx-target-4xx`
- Line ~92: same pattern for `hx-target-error`
- Line ~101: same pattern for status-200 test
- Line ~112: same pattern for `hx-target-404="this"` test

- [ ] **Step 3: Run the suite to confirm the existing tests still pass with the reshaped helper**

Run: `cd framework/assets/js && npx vitest run htmxextensions/__tests__/response-targets.test.ts`

Expected: all tests PASS.

- [ ] **Step 4: Add the failing event-firing test**

Append inside the `describe("response-targets extension", ...)` block:

```ts
  it("fires htmgo:response:retargeted when a new target is resolved", () => {
    document.body.innerHTML = `<div id="err"></div>`;
    const srcElt = document.createElement("button");
    document.body.appendChild(srcElt);
    const { api, triggered } = makeApi({ "hx-target-404": "#err" });
    ext.init(api);
    const { detail, mainTask } = makeDetail(srcElt, 404);
    ext.htmx_before_swap(srcElt, detail);

    const newTarget = document.getElementById("err")!;
    expect(mainTask.target).toBe(newTarget);

    const events = triggered.filter((e) => e.name === "htmgo:response:retargeted");
    expect(events.length).toBe(1);
    expect(events[0].elt).toBe(srcElt);
    expect(events[0].detail.status).toBe(404);
    expect(events[0].detail.from).toBe(null);
    expect(events[0].detail.to).toBe(newTarget);
    expect(events[0].detail.ctx).toBe(detail.ctx);
  });
```

- [ ] **Step 5: Run the test — it must fail**

Run: `cd framework/assets/js && npx vitest run htmxextensions/__tests__/response-targets.test.ts -t "fires htmgo:response:retargeted"`

Expected: FAIL with `expect(events.length).toBe(1)` — received 0. The extension does not yet dispatch the event.

- [ ] **Step 6: Commit**

```bash
git add framework/assets/js/htmxextensions/__tests__/response-targets.test.ts
git commit -m "test(response-targets): add failing test for htmgo:response:retargeted event"
```

---

### Task 4: Implement `htmgo:response:retargeted` dispatch

**Files:**
- Modify: `framework/assets/js/htmxextensions/response-targets.ts`

- [ ] **Step 1: Dispatch the event when a new target is assigned**

Change the tail of `htmx_before_swap` from:

```ts
    const target = getRespCodeTarget(reqElt, status);
    if (target) {
      mainTask.target = target;
    }
```

to:

```ts
    const target = getRespCodeTarget(reqElt, status);
    if (target) {
      const from = mainTask.target ?? null;
      mainTask.target = target;
      api.triggerHtmxEvent(ctx.sourceElement, "htmgo:response:retargeted", {
        status,
        from,
        to: target,
        ctx,
      });
    }
```

- [ ] **Step 2: Run the test — it must pass**

Run: `cd framework/assets/js && npx vitest run htmxextensions/__tests__/response-targets.test.ts -t "fires htmgo:response:retargeted"`

Expected: PASS.

- [ ] **Step 3: Run the full suite — everything passes**

Run: `cd framework/assets/js && npx vitest run htmxextensions/__tests__/response-targets.test.ts`

Expected: all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add framework/assets/js/htmxextensions/response-targets.ts
git commit -m "feat(response-targets): fire htmgo:response:retargeted on retarget

Dispatch via api.triggerHtmxEvent on ctx.sourceElement with
{status, from, to, ctx}. Gives consumers a canonical hook for observing
response-targets retargets without resurrecting htmx-2 isError semantics."
```

---

### Task 5: Add suppression-path tests

**Files:**
- Modify: `framework/assets/js/htmxextensions/__tests__/response-targets.test.ts`

Verify the event does NOT fire in the three early-return paths and on 2xx. Each test is independent.

- [ ] **Step 1: Add the four tests**

Append inside the `describe` block:

```ts
  it("does not fire htmgo:response:retargeted when PrefersExisting keeps existing target", async () => {
    document.body.innerHTML = `<div id="err"></div>`;
    const srcElt = document.createElement("button");
    document.body.appendChild(srcElt);
    const cfg = (await import("htmx.org")).default.config as any;
    cfg.responseTargetPrefersExisting = true;
    const { api, triggered } = makeApi({ "hx-target-404": "#err" });
    ext.init(api);
    const existing = document.createElement("div");
    const { detail } = makeDetail(srcElt, 404, existing);
    ext.htmx_before_swap(srcElt, detail);

    cfg.responseTargetPrefersExisting = false;
    expect(triggered.find((e) => e.name === "htmgo:response:retargeted")).toBeUndefined();
  });

  it("does not fire htmgo:response:retargeted when HX-Retarget header is honored", async () => {
    document.body.innerHTML = `<div id="err"></div>`;
    const srcElt = document.createElement("button");
    document.body.appendChild(srcElt);
    const cfg = (await import("htmx.org")).default.config as any;
    cfg.responseTargetPrefersRetargetHeader = true;
    const { api, triggered } = makeApi({ "hx-target-404": "#err" });
    ext.init(api);
    const existing = document.createElement("div");
    const { detail } = makeDetail(srcElt, 404, existing);
    detail.ctx.response.headers = {
      get: (name: string) => (name === "HX-Retarget" ? "#other" : null),
    };
    ext.htmx_before_swap(srcElt, detail);

    expect(triggered.find((e) => e.name === "htmgo:response:retargeted")).toBeUndefined();
  });

  it("does not fire htmgo:response:retargeted when no hx-target-* matches", () => {
    const srcElt = document.createElement("button");
    document.body.appendChild(srcElt);
    const { api, triggered } = makeApi({});
    ext.init(api);
    const { detail, mainTask } = makeDetail(srcElt, 404);
    ext.htmx_before_swap(srcElt, detail);

    expect(mainTask.target).toBe(null);
    expect(triggered.find((e) => e.name === "htmgo:response:retargeted")).toBeUndefined();
  });

  it("does not fire htmgo:response:retargeted on 2xx status", () => {
    document.body.innerHTML = `<div id="err"></div>`;
    const srcElt = document.createElement("button");
    document.body.appendChild(srcElt);
    const { api, triggered } = makeApi({ "hx-target-error": "#err" });
    ext.init(api);
    const existing = document.createElement("div");
    const { detail, mainTask } = makeDetail(srcElt, 200, existing);
    ext.htmx_before_swap(srcElt, detail);

    expect(mainTask.target).toBe(existing);
    expect(triggered.find((e) => e.name === "htmgo:response:retargeted")).toBeUndefined();
  });
```

- [ ] **Step 2: Run the suite — all four new tests and everything else must pass**

Run: `cd framework/assets/js && npx vitest run htmxextensions/__tests__/response-targets.test.ts`

Expected: all tests PASS.

- [ ] **Step 3: Commit**

```bash
git add framework/assets/js/htmxextensions/__tests__/response-targets.test.ts
git commit -m "test(response-targets): cover event suppression paths

PrefersExisting, HX-Retarget header, no-match, and 2xx must not fire
htmgo:response:retargeted."
```

---

### Task 6: Update CHANGELOG

**Files:**
- Modify: `CHANGELOG.md` (under the existing `[1.2.0-beta.1]` section)

- [ ] **Step 1: Add a breaking-change bullet under `### Breaking`**

Append to the `### Breaking` list (after the current last bullet on line ~36):

```markdown
- **`response-targets` extension no longer reads or initializes `htmx.config.responseTargetSetsError` / `responseTargetUnsetsError`.** These htmx 2 config keys have no htmx 4 equivalent — htmx 4 removed the `isError` flag and the `htmx:responseError` event. If you relied on them, subscribe to `htmx:before:response` and inspect `ctx.response.status` directly.
```

- [ ] **Step 2: Add an entry under `### Added`**

Append to the `### Added` list (after the current last bullet on line ~45):

```markdown
- `htmgo:response:retargeted` event fired by the `response-targets` extension on the source element when a 4xx/5xx response triggers a retarget. Detail: `{status, from, to, ctx}`. Gives consumers a canonical hook without reintroducing htmx-2 error-flag semantics.
```

- [ ] **Step 3: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs(changelog): record response-targets isError knob removal + new event"
```

---

### Task 7: Rebuild the JS bundle

**Files:**
- Regenerate: `framework/assets/dist/htmgo.js`

- [ ] **Step 1: Run the full test suite once more as a safety net**

Run: `cd framework/assets/js && npm test`

Expected: all tests PASS across all extension suites.

- [ ] **Step 2: Rebuild the bundle**

Run: `cd framework/assets/js && npm run build`

Expected: clean build; `framework/assets/dist/htmgo.js` is regenerated.

- [ ] **Step 3: Verify the bundle reflects the changes**

Run: `grep -c "responseTargetUnsetsError\|responseTargetSetsError" framework/assets/dist/htmgo.js`

Expected: `0` (neither key should appear anywhere in the bundled output).

Run: `grep -c "htmgo:response:retargeted" framework/assets/dist/htmgo.js`

Expected: `1` or more (the new event name must appear in the bundle).

- [ ] **Step 4: Commit the rebuilt bundle**

```bash
git add framework/assets/dist/htmgo.js
git commit -m "chore(js): rebuild bundle for response-targets error-flag changes"
```

---

### Task 8: Draft upstream htmx issue

**Files:**
- Create: `docs/superpowers/drafts/2026-04-18-upstream-htmx-response-error-issue.md`

This is a text-only deliverable. @franchb posts it to `bigskysoftware/htmx` manually; it is NOT a release blocker.

- [ ] **Step 1: Ensure the drafts directory exists**

Run: `mkdir -p docs/superpowers/drafts`

- [ ] **Step 2: Write the issue draft**

Create `docs/superpowers/drafts/2026-04-18-upstream-htmx-response-error-issue.md` with the following content:

```markdown
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
```

- [ ] **Step 3: Commit the draft**

```bash
git add docs/superpowers/drafts/2026-04-18-upstream-htmx-response-error-issue.md
git commit -m "docs(drafts): upstream htmx issue draft proposing htmx:response:error"
```

---

## Self-review checklist (performed during plan authoring)

- **Spec coverage:** Extension changes → Tasks 2, 4. Event contract → Task 4. Tests → Tasks 1, 3, 5. CHANGELOG → Task 6. Rebuild → Task 7. Upstream issue → Task 8. Config keys to keep (`responseTargetPrefersExisting`, `responseTargetPrefersRetargetHeader`) are preserved in Task 2 Step 2. All spec sections have tasks.
- **Placeholder scan:** no TBDs, no "implement later", every code block is complete.
- **Type consistency:** event name `htmgo:response:retargeted` used identically in Tasks 3, 4, 5. Detail keys `{status, from, to, ctx}` used identically across the contract, the dispatcher, and all assertion sites. `makeApi` return shape `{api, triggered}` stays consistent from Task 3 onward.
