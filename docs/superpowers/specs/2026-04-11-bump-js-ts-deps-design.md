# Bump JS/TS Dependencies

**Date:** 2026-04-11
**Branch:** `deps/bump-js-ts-deps`
**PR target:** `master`

## Context

This fork (franchb/htmgo) is actively maintained after the original upstream maintainer stepped down. Go dependencies were bumped to latest in a prior PR. This PR brings all JS/TS dependencies to their latest versions.

The repo has two npm packages:
- `framework/assets/js/` — builds the htmgo client-side JS bundle (htmgo.js)
- `htmgo-site/` — documentation site, uses Tailwind CSS via the htmgo CLI

## Changes

### framework/assets/js/package.json

| Package | Current | Target | Type | Notes |
|---|---|---|---|---|
| `htmx.org` | `~2.0.2` | `^2.0.8` | dependency | Switch tilde to caret for future 2.x minor updates |
| `@swc/core` | `^1.7.26` | `^1.15.24` | devDependency | Transpiler used by tsup |
| `@types/node` | `^22.5.4` | `^25.6.0` | devDependency | Type definitions |
| `prettier` | `^3.3.3` | `^3.8.2` | devDependency | Code formatter |
| `shiki` | `^1.17.6` | `^4.0.2` | devDependency | Not imported anywhere in this package — appears unused but bump anyway |
| `tailwindcss` | `^3.4.11` | `^4.2.2` | devDependency | Used only in optional `tailwind:watch` npm script |
| `tsup` | `^8.2.4` | `^8.5.1` | devDependency | JS bundler |
| `typescript` | `^5.6.2` | `^6.0.2` | devDependency | Type checking (SWC handles actual compilation) |

### htmgo-site/package.json

| Package | Current | Target | Type | Notes |
|---|---|---|---|---|
| `@tailwindcss/typography` | `^0.5.15` | `^0.5.19` | devDependency | Tailwind typography plugin; site CSS already uses v4 `@plugin` syntax |

## Verification

1. **Framework JS build:** `cd framework/assets/js && npm install && npm run build` — verify `framework/assets/dist/htmgo.js` is produced
2. **Bundle diff:** inspect `git diff framework/assets/dist/` to understand output changes
3. **Site deps:** `cd htmgo-site && npm install` — verify clean resolution
4. **Go tests:** `cd framework && go test ./...` — ensure no regressions
5. **Commit rebuilt artifacts** if the JS bundle output changed

## Risk Assessment

- **TypeScript 5 -> 6:** Could introduce stricter type checking. Mitigation: tsup uses SWC for compilation, not tsc. If TS 6 breaks type-checking, pin to `^5.8.x`.
- **shiki 1 -> 4:** Major API change, but package appears unused in this location. If it breaks, remove it.
- **tailwindcss 3 -> 4 in framework/assets/js:** Only affects the optional `tailwind:watch` script. The htmgo CLI downloads its own standalone Tailwind binary for actual builds.
- **htmx 2.0.2 -> 2.0.8:** Patch-level changes. htmx has strong semver discipline. Lowest risk.
- **@swc/core 1.7 -> 1.15:** Large minor jump but within semver ^1.x range. tsup abstracts the API.

## Out of Scope

- Removing unused dependencies (shiki, tailwindcss from framework/assets/js) — separate cleanup PR
- Tailwind v4 migration of the htmgo-site CSS — already using v4 syntax, no changes needed
- Cleaning up the stale `.claude/worktrees/tailwind-v4-migration/` directory
