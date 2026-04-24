# cli/htmgo v2 install fix + cross-module go.mod cleanup

**Date:** 2026-04-24
**Tracking issue:** [#14](https://github.com/franchb/htmgo/issues/14)
**Status:** Design (approved)
**Target release:** `v2.0.1` across all v2 submodules

## 1. Problem

`cli/htmgo/v2.0.0` cannot be installed by any downstream user:

```
$ go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.0
go: ... invalid version: module contains a go.mod file, so module path must
    match major version ("github.com/franchb/htmgo/cli/htmgo/v2")

$ go install github.com/franchb/htmgo/cli/htmgo@<pseudo>
go: ... The go.mod file for the module providing named packages contains one
    or more replace directives. It must not contain directives that would
    cause it to be interpreted differently than if it were the main module.
```

Two independent Go module rules are violated in `cli/htmgo/go.mod`:

1. **SIV path mismatch.** `module github.com/franchb/htmgo/cli/htmgo` does not
   carry the `/v2` suffix required at major ≥ 2.
2. **`replace` on a non-main module.** Two `replace` lines point at sibling
   directories (`../../framework`, `../../tools/html-to-htmgo`); Go rejects
   non-main modules that use local-path replaces.

Repo survey confirms two additional tagged v2 submodules have the same
`replace`-rule violation:

- `framework-ui/v2.0.0` — replaces `framework/v2`
- `extensions/websocket/v2.0.0` — replaces `framework/v2`

These three broken modules, plus `framework/v2.0.0` and
`tools/html-to-htmgo/v2.0.0` which are packaged correctly, all point at the
same commit `1102e671d216b9afb844763caa2d1a50fb0416f5`.

Net effect: downstream projects pinned on `framework/v2` cannot install a
matching CLI, so regenerated files (`__htmgo/pages-generated.go`,
`__htmgo/partials-generated.go`) cannot be produced with correct `/v2`
imports. v2 adoption is blocked.

## 2. Goals / non-goals

**Goals.**

- Make `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1` succeed from
  a clean environment.
- Make `go get` of `framework-ui/v2` and `extensions/websocket/v2` succeed
  from a clean environment.
- Keep local monorepo development (`go build ./cli/htmgo`, example apps,
  docs site) working with no change to the developer workflow beyond the
  presence of a root `go.work`.
- Preserve version-number alignment across the five v2 library/CLI
  submodules by republishing all of them at `v2.0.1` at a single commit.
- Guard the regression in CI so a future release cannot silently ship a
  broken `cli/htmgo` again.

**Non-goals.**

- No framework API changes.
- No generated-code shape changes. `astgen` already emits the `/v2` path
  for the framework import — that's the behavior that made this bug
  externally visible. It keeps emitting `/v2`.
- No changes to `examples/*`, `htmgo-site`, or `templates/starter`
  `go.mod` files in this PR. Those are main modules where `replace` is
  legal; a root `go.work` makes their sibling replaces redundant for
  local dev but we leave them in place to avoid expanding blast radius.

## 3. Approach

Move all local-path resolution out of `go.mod` files and into a single
`go.work` at the repo root. `go.work` is honored only for the main module
in a workspace, is ignored by `go install` / `go get` / `go mod download`
of *installed* dependencies, and never leaks into published modules. This
is the idiomatic fix for monorepos that publish multiple Go submodules.

Separately, fix `cli/htmgo`'s SIV violation by renaming its module path
to `github.com/franchb/htmgo/cli/htmgo/v2` and rewriting every internal
import inside `cli/htmgo/` to carry the `/v2` segment.

Then retag all five v2 submodules at the fix commit as `v2.0.1`, matching
the existing multi-module release pattern (all five `v2.0.0` tags
currently point at the same commit).

### Approaches considered and rejected

- **Build-tagged or environment-conditional `replace`.** Go's `go.mod`
  grammar does not support conditional directives. The only supported
  mechanism for "local-dev only" resolution is `go.work`.
- **Rename the CLI install path back to `.../cli/htmgo` (no `/v2`).**
  Impossible for a submodule with its own `go.mod` at major ≥ 2; SIV is
  mandatory. The v2.0.0 release-notes claim that the CLI install path is
  unchanged is wrong and must be amended.
- **Force-update `v2.0.0` tags in place.** Go's module proxy caches the
  original `v2.0.0` indefinitely. Some users would hit the stale cached
  module, some the fixed one. Not safe. A clean `v2.0.1` bump avoids
  this.

## 4. Per-module changes

| Module | `go.mod` | Go source | Tag |
| --- | --- | --- | --- |
| `cli/htmgo` | `module …/cli/htmgo` → `module …/cli/htmgo/v2`; delete two `replace` lines; change `require` pseudo-versions `v2.0.0-20260423190209-1102e671d216` of `framework/v2` and `tools/html-to-htmgo/v2` → `v2.0.1` | Rewrite all internal imports `…/cli/htmgo/<pkg>` → `…/cli/htmgo/v2/<pkg>` across runner.go, watcher.go, signals.go, every `tasks/*` package, `internal/**`, and tests | `cli/htmgo/v2.0.1` |
| `framework-ui` | delete `replace github.com/franchb/htmgo/framework/v2 => ../framework`; change `require github.com/franchb/htmgo/framework/v2 v2.0.0-20260423190209-1102e671d216` → `v2.0.1` | none (SIV path already correct) | `framework-ui/v2.0.1` |
| `extensions/websocket` | delete `replace github.com/franchb/htmgo/framework/v2 => ../../framework`; change `require github.com/franchb/htmgo/framework/v2 v2.0.0-20260423190209-1102e671d216` → `v2.0.1` | none | `extensions/websocket/v2.0.1` |
| `framework` | version-only republish | none | `framework/v2.0.1` |
| `tools/html-to-htmgo` | version-only republish | none | `tools/html-to-htmgo/v2.0.1` |

Required-version bumps on `framework/v2` across the three consumer
submodules ensure the CLI-plus-companions bundle pulls the fixed
`framework/v2.0.1` rather than pseudo-versions of the pre-fix commit.

## 5. Root `go.work`

New file at repo root:

```
go 1.26

use (
    ./framework
    ./framework-ui
    ./extensions/websocket
    ./tools/html-to-htmgo
    ./cli/htmgo
)
```

`.gitignore` gains a `go.work.sum` entry. Go writes `go.work.sum` alongside
`go.work` and its contents are machine-local (sum database for the
workspace) — it should not be checked in.

`examples/*`, `htmgo-site`, and `templates/starter` are deliberately not
in the workspace. Their existing `replace` directives continue to resolve
sibling paths from disk for their developers. Including them in the
workspace would be a larger change with no benefit to the install-bug
fix; it can be considered later.

## 6. Internal-import rewrite (risk area)

The CLI has ~30+ import statements of the form
`github.com/franchb/htmgo/cli/htmgo/(internal|tasks|…)`. All must gain the
`/v2` segment.

Procedure:

1. Enumerate subpackage roots: `internal`, `tasks/astgen`, `tasks/copyassets`,
   `tasks/css`, `tasks/downloadtemplate`, `tasks/formatter`, `tasks/module`,
   `tasks/process`, `tasks/reloader`, `tasks/run`.
2. Run a single `gofmt -r` pass per root, or a scripted `sed -i` over all
   `*.go` under `cli/htmgo`, replacing
   `github.com/franchb/htmgo/cli/htmgo/<pkg>` with
   `github.com/franchb/htmgo/cli/htmgo/v2/<pkg>`.
3. Verify afterwards by grep that no un-prefixed
   `github.com/franchb/htmgo/cli/htmgo/(internal|tasks)` import remains.
4. Build-and-test gate: `cd cli/htmgo && go build ./... && go test ./...`
   must pass before tagging. Run the same inside `cli/htmgo/tasks/astgen`
   specifically, since its test suite is the largest CLI-side test
   surface.

Embedded-string audit: `cli/htmgo/tasks/astgen` emits framework import
paths (`ModuleName = "github.com/franchb/htmgo/framework/v2/h"` at
`cli/htmgo/tasks/astgen/entry.go:21`) into generated files. That constant
points at the framework module and is unchanged by this fix. No other
embedded self-reference to the CLI's own module path is expected;
confirm via `grep -rn 'github.com/franchb/htmgo/cli/htmgo' cli/htmgo
--include='*.go'` before and after the rewrite (non-import string
literals should be zero).

## 7. Release / tagging plan

Single fix commit on `master`. Commit contents:

1. `go.work` (new).
2. `cli/htmgo/go.mod` — module path rename, replaces removed, dep versions
   pinned to `v2.0.1`.
3. `cli/htmgo/**/*.go` — internal import rewrite.
4. `framework-ui/go.mod` — replace removed, framework dep bumped to
   `v2.0.1`.
5. `extensions/websocket/go.mod` — replace removed, framework dep bumped
   to `v2.0.1`.
6. `.gitignore` — add `go.work.sum`.
7. `CHANGELOG.md` — entry per §8.
8. README / release-notes amendment per §9.

Push five tags at the fix commit:

- `framework/v2.0.1`
- `framework-ui/v2.0.1`
- `extensions/websocket/v2.0.1`
- `tools/html-to-htmgo/v2.0.1`
- `cli/htmgo/v2.0.1`

The two "clean" submodules (`framework`, `tools/html-to-htmgo`) get a
version-number bump with no code change so all five v2 submodules stay
aligned on one version number going forward. This matches the
single-commit-multi-tag pattern already used for `v2.0.0`.

The old `v2.0.0` tags are left in place. They remain accessible for
anyone who fetched them before the proxy ejected them, but they will no
longer be recommended in docs; the README install snippet and binny
examples point at `v2.0.1`.

## 8. CHANGELOG entry (v2.0.1)

- **Fix:** `cli/htmgo` `go.mod` now declares the `/v2` module path
  required by Go's Semantic Import Versioning rules. Install invocation
  becomes `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1`. The
  binary name (`htmgo`) is unchanged.
- **Fix:** removed sibling-path `replace` directives from `cli/htmgo`,
  `framework-ui`, and `extensions/websocket` `go.mod` files. These
  directives violated Go's rule that non-main modules must not contain
  local-path replaces, and blocked downstream `go get`/`go install` of
  all three. Local monorepo development now uses a root `go.work` file.
- **Chore:** `framework/v2.0.1` and `tools/html-to-htmgo/v2.0.1` are
  republished at the fix commit with no code change so all five v2
  submodule versions stay aligned.

## 9. Release-notes / README amendment

The v2.0.0 release notes state *"The CLI binary
(`github.com/franchb/htmgo/cli/htmgo`) retains its original path."*
This is incorrect — Go's SIV rule does not permit a submodule at major
≥ 2 to keep its v1 install path when it has its own `go.mod`.

Amend to:

> The CLI binary is still named `htmgo`. Starting at `v2.0.1` the install
> invocation is:
>
> ```
> go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1
> ```
>
> The `/v2` suffix is required by Go's Semantic Import Versioning rules
> for any submodule with its own `go.mod` at major version ≥ 2. The
> `v2.0.0` tags were mis-packaged and cannot be installed by downstream
> users; use `v2.0.1` or later.

Update the framework README's install snippet accordingly. Update any
binny example from:

```yaml
- name: htmgo
  version: { want: v2.0.0 }
  method: go-install
  with: { module: github.com/franchb/htmgo/cli/htmgo }
```

to:

```yaml
- name: htmgo
  version: { want: v2.0.1 }
  method: go-install
  with: { module: github.com/franchb/htmgo/cli/htmgo/v2 }
```

## 10. Verification

### Clean-environment install check

From a directory outside the monorepo, with `GOPATH` pointing at an empty
dir to defeat local caches:

```
$ GOPROXY=https://proxy.golang.org,direct go install \
    github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1
$ htmgo version
```

Both invocations must succeed. Repeat with the binny YAML above.

### Downstream regenerate smoke test

On a throwaway copy of `examples/minimal-htmgo` with its local `replace`
removed:

```
$ htmgo generate
$ grep '"github.com/franchb/htmgo/framework/v2/h"' \
    __htmgo/pages-generated.go
```

The generated file must import `framework/v2/h`. This confirms end-to-end
that a downstream project can (a) install the CLI, (b) regenerate, and
(c) produce code that compiles against `framework/v2`.

### CI guard

`.github/workflows/verify-installer-works.yml` currently installs the
CLI. Update it to install via `.../cli/htmgo/v2@<tag>` and assert
`htmgo version` exits zero. This prevents a silent regression of the
install path in any future v2.x release.

### Framework-ui and extensions/websocket check

```
$ go get github.com/franchb/htmgo/framework-ui/v2@v2.0.1
$ go get github.com/franchb/htmgo/extensions/websocket/v2@v2.0.1
```

Both must succeed from a directory whose `go.mod` does not contain a
sibling `replace`.

## 11. Rollback

If the release proves broken after tags are pushed, retag `v2.0.2` at a
corrective commit. Do not force-update `v2.0.1` in place — same proxy
caching hazard as §3.

## 12. Out of scope (acknowledged follow-ups)

- `examples/*/go.mod`, `htmgo-site/go.mod`, `templates/starter/go.mod`
  keep their sibling `replace` directives. Legal for main modules; a
  tidy-up can be done later if desired, but is unrelated to the install
  bug.
- `templates/starter` is copied by `htmgo setup` / `htmgo template` to
  create new projects. Its `replace` line is stripped by that pipeline
  in normal use; audit this separately before changing template
  contents.
