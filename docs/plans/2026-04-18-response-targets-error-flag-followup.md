# Follow-up: `response-targets` error-flag logic in htmx 4

**Status:** open — scope decision needed before v2.0.0 GA.

**Origin:** CodeRabbit review on PR #12 (htmx 4 migration). Finding: https://github.com/franchb/htmgo/pull/12#discussion_r3104645514

## Problem

`framework/assets/js/htmxextensions/response-targets.ts`:

```ts
function handleErrorFlag(detail: any) {
  if (detail.isError) {
    if (config.responseTargetUnsetsError) detail.isError = false;
  } else if (config.responseTargetSetsError) {
    detail.isError = true;
  }
}
```

Two issues, one on top of the other:

1. Callers pass `ctx` (the htmx 4 swap context), not `detail`.
2. Independently, `ctx` in htmx 4.0.0-beta2 has **no `isError` property at all** — verified against `node_modules/htmx.org/dist/htmax.js`, where `isError` appears zero times.

The htmx 2 error-flag branch is therefore dead code in htmx 4: mutations silently no-op, and the `htmx.config.responseTargetSetsError` / `responseTargetUnsetsError` config knobs have no observable effect.

A trivial rename to `handleErrorFlag(detail)` does not fix it — `detail.isError` is also absent in htmx 4's before:swap payload.

## Scope decision

Two directions, pick one before cutting GA:

### Option A — Drop the feature (recommended default)

Remove `handleErrorFlag` and the two config keys (`responseTargetSetsError`, `responseTargetUnsetsError`). Retargeting on 4xx/5xx still works as before; only the "flip error flag on retarget" htmx 2 side effect is lost.

- Pros: small diff, no guesswork about htmx 4 internals, honest about what we ship.
- Cons: minor BC break for anyone relying on those two config keys. Document in CHANGELOG under the v2.0.0-beta.1 section.

### Option B — Rewire to htmx 4's error path

Drive the error decision off `ctx.response.status` (or whatever htmx 4 uses downstream) instead of a mutable `isError`. Needs mapping to wherever htmx 4 surfaces "this response counts as an error" — likely the `htmx:response:error` trigger path.

- Pros: preserves full parity with htmx 2 extension semantics.
- Cons: requires understanding htmx 4's error-handling rewrite; risk of drift when beta evolves; larger test surface.

**Suggested default:** Option A for v2.0.0-beta. Revisit Option B if users ask for it.

## Checklist when resuming

- [ ] Decide A vs B (open question for @franchb)
- [ ] Implement chosen option in `framework/assets/js/htmxextensions/response-targets.ts`
- [ ] Update `framework/assets/js/htmxextensions/__tests__/response-targets.test.ts` — the current test only verifies defaults are installed, not the error-flag branch
- [ ] CHANGELOG: record whichever decision under v2.0.0-beta.1
- [ ] Rebuild `framework/assets/dist/htmgo.js`
