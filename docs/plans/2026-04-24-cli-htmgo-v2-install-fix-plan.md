# cli/htmgo v2 install fix — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1` succeed from a clean environment, and fix the same `replace`-rule violation in `framework-ui` and `extensions/websocket`.

**Architecture:** Move all sibling-path `replace` directives out of `go.mod` files into a single root `go.work` (honored only for local dev, invisible to downstream `go install`). Rename the CLI module to the `/v2` path to satisfy Go's Semantic Import Versioning. Retag all five v2 submodules at `v2.0.1` from the fix commit so version numbers stay aligned.

**Tech Stack:** Go 1.26, Go modules, `go.work` workspaces, GitHub tag-based releases.

**Spec:** `docs/plans/2026-04-24-cli-htmgo-v2-install-fix-design.md`
**Tracking issue:** [#14](https://github.com/franchb/htmgo/issues/14)

---

## File Structure

**Create:**
- `go.work` — root workspace file listing the five v2 submodules.

**Modify:**
- `.gitignore` — add `go.work.sum`.
- `cli/htmgo/go.mod` — module path (`cli/htmgo` → `cli/htmgo/v2`) + delete two `replace` lines. The `require` pseudo-versions of `framework/v2` and `tools/html-to-htmgo/v2` are left unchanged.
- `cli/htmgo/**/*.go` — 46 internal import statements across 13 subpackage roots (listed below).
- `framework-ui/go.mod` — delete one `replace` line. `require` pseudo-version left unchanged.
- `extensions/websocket/go.mod` — delete one `replace` line. `require` pseudo-version left unchanged.
- `CHANGELOG.md` — add `v2.0.1` entry; correct the v2.0.0 note about CLI install path.
- `.github/workflows/verify-installer-works.yml` — add a job that installs via the public `cli/htmgo/v2@latest` path to guard the regression.
- `htmgo-site/pages/docs/installation.go` — change install snippet from `cli/htmgo@latest` to `cli/htmgo/v2@latest`.
- `htmgo-site/pages/docs/misc/formatter.go` — change help-text install snippet.
- All `go run github.com/franchb/htmgo/cli/htmgo@latest` occurrences in `examples/*/Taskfile.yml`, `examples/*/Dockerfile`, `htmgo-site/Dockerfile`, `templates/starter/Dockerfile` → add `/v2`.

**Unchanged (explicit non-goals):**
- `examples/*/go.mod`, `htmgo-site/go.mod`, `templates/starter/go.mod` — these are main modules; their sibling `replace` directives are legal.
- `framework/go.mod`, `tools/html-to-htmgo/go.mod` — already packaged correctly at v2.0.0. Version-bump tag only; no code change.
- `framework/assets/js/**` — unrelated to this fix.

**Critical invariant during this work:**
> **Do NOT run `go mod tidy`** in any module during this work. Tidy may rewrite require pseudo-versions based on what the proxy currently knows, and this can subtly churn the go.mod lines we're deliberately leaving untouched. The root `go.work` resolves sibling modules from disk for local builds, so tidy is not needed.

---

## Task 1: Branch setup

**Files:** none (git state only)

- [ ] **Step 1: Confirm working tree is clean**

```bash
git status --short
```

Expected: empty output (or only untracked files unrelated to this work; no modifications to tracked files).

- [ ] **Step 2: Create and switch to a feature branch**

```bash
git checkout -b fix/v2-install-issue-14 master
```

Expected: `Switched to a new branch 'fix/v2-install-issue-14'`

- [ ] **Step 3: Verify current state matches spec assumptions**

```bash
grep "^module " cli/htmgo/go.mod
grep "^replace" cli/htmgo/go.mod framework-ui/go.mod extensions/websocket/go.mod
```

Expected:
```
cli/htmgo/go.mod:module github.com/franchb/htmgo/cli/htmgo
cli/htmgo/go.mod:replace github.com/franchb/htmgo/framework/v2 => ../../framework
cli/htmgo/go.mod:replace github.com/franchb/htmgo/tools/html-to-htmgo/v2 => ../../tools/html-to-htmgo
framework-ui/go.mod:replace github.com/franchb/htmgo/framework/v2 => ../framework
extensions/websocket/go.mod:replace github.com/franchb/htmgo/framework/v2 => ../../framework
```

If these don't match exactly, stop and reconcile with the spec before proceeding.

---

## Task 2: Create the root `go.work`

**Files:**
- Create: `/home/iru/p/github.com/franchb/htmgo/go.work`
- Modify: `/home/iru/p/github.com/franchb/htmgo/.gitignore`

- [ ] **Step 1: Write `go.work`**

Create file at repo root with exact content:

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

Tabs (not spaces) between `use` parentheses — that's the `go.work` convention, matching how `go.mod` require blocks are formatted.

- [ ] **Step 2: Verify `go.work` is syntactically valid**

```bash
go env GOWORK
```

Expected: the absolute path of the new `go.work` (Go auto-discovers it from the cwd).

```bash
go work sync
```

Expected: exit 0, no output. This synchronizes the workspace metadata. If it errors, the `use` paths are wrong.

- [ ] **Step 3: Add `go.work.sum` to `.gitignore`**

Append the following line to `.gitignore`:

```
go.work.sum
```

Verify:

```bash
grep "^go.work.sum$" .gitignore
```

Expected: `go.work.sum`

- [ ] **Step 4: Verify all five modules still build under the workspace**

```bash
go build ./framework/... ./framework-ui/... ./extensions/websocket/... ./tools/html-to-htmgo/... ./cli/htmgo/...
```

Expected: exit 0, no output. This confirms `go.work` resolves the sibling modules correctly — we haven't removed the in-`go.mod` replaces yet, so both mechanisms are active and agree.

- [ ] **Step 5: Commit**

```bash
git add go.work .gitignore
git commit -m "$(cat <<'EOF'
build: add root go.work for monorepo dev builds

go.work resolves sibling modules from disk for local development. It is
only honored for the main module in a workspace and is invisible to
downstream `go install` / `go get` of published modules.

Prepares the ground for removing sibling-path `replace` directives from
cli/htmgo, framework-ui, and extensions/websocket (see #14).
EOF
)"
```

---

## Task 3: Rename the `cli/htmgo` module path to `/v2`

**Files:**
- Modify: `cli/htmgo/go.mod` (line 1 only in this task)

Only the module path here — imports are handled in Task 4, replaces in Task 5. Split for reviewability.

- [ ] **Step 1: Edit the module line**

Change `cli/htmgo/go.mod` line 1 from:

```
module github.com/franchb/htmgo/cli/htmgo
```

to:

```
module github.com/franchb/htmgo/cli/htmgo/v2
```

- [ ] **Step 2: Verify the edit**

```bash
head -1 cli/htmgo/go.mod
```

Expected: `module github.com/franchb/htmgo/cli/htmgo/v2`

- [ ] **Step 3: Confirm build is now broken (expected)**

```bash
cd cli/htmgo && go build ./... 2>&1 | head -5 && cd -
```

Expected: compile errors of the form `package github.com/franchb/htmgo/cli/htmgo/tasks/... is not in std`. This is correct — the module's own internal imports no longer match the module path. Task 4 fixes them.

**Do not commit yet.** This task's change is bundled with Task 4's to keep the tree buildable at every commit.

---

## Task 4: Rewrite internal imports in `cli/htmgo` to the `/v2` path

**Files:**
- Modify: every `*.go` file under `cli/htmgo/` that imports another `cli/htmgo` subpackage.

Scope, confirmed by pre-work grep: **46 import statements** across **13 subpackage roots**:

```
github.com/franchb/htmgo/cli/htmgo/internal
github.com/franchb/htmgo/cli/htmgo/internal/dirutil
github.com/franchb/htmgo/cli/htmgo/internal/tailwind
github.com/franchb/htmgo/cli/htmgo/tasks/astgen
github.com/franchb/htmgo/cli/htmgo/tasks/copyassets
github.com/franchb/htmgo/cli/htmgo/tasks/css
github.com/franchb/htmgo/cli/htmgo/tasks/downloadtemplate
github.com/franchb/htmgo/cli/htmgo/tasks/formatter
github.com/franchb/htmgo/cli/htmgo/tasks/module
github.com/franchb/htmgo/cli/htmgo/tasks/process
github.com/franchb/htmgo/cli/htmgo/tasks/reloader
github.com/franchb/htmgo/cli/htmgo/tasks/run
github.com/franchb/htmgo/cli/htmgo/tasks/util
```

- [ ] **Step 1: Baseline: count un-prefixed imports before the rewrite**

```bash
grep -rn "github.com/franchb/htmgo/cli/htmgo/" cli/htmgo --include="*.go" \
  | grep -v "/v2/" | wc -l
```

Expected: `46`

If this number is not 46, stop — the codebase shape differs from the plan's assumption and the rewrite count needs to be re-derived.

- [ ] **Step 2: Run the rewrite**

```bash
find cli/htmgo -name '*.go' -type f -print0 \
  | xargs -0 sed -i 's|github.com/franchb/htmgo/cli/htmgo/|github.com/franchb/htmgo/cli/htmgo/v2/|g'
```

This inserts `/v2/` after `cli/htmgo` in every occurrence inside `cli/htmgo/**/*.go`. The trailing slash in the match prevents accidental double-matching if any occurrence already has `/v2/`.

- [ ] **Step 3: Verify no un-prefixed imports remain**

```bash
grep -rn "github.com/franchb/htmgo/cli/htmgo/" cli/htmgo --include="*.go" \
  | grep -v "/v2/" \
  | grep -v "ModuleName.*framework"
```

Expected: empty output.

The `grep -v 'ModuleName.*framework'` filter is defensive against false positives in `cli/htmgo/tasks/astgen/entry.go` where `ModuleName` is the **framework's** `/v2/h` import path (line 21) — that string contains `framework/v2`, not `cli/htmgo/v2`, and should never have been touched by the sed above. Confirm with:

```bash
grep -n "ModuleName" cli/htmgo/tasks/astgen/entry.go
```

Expected: `const ModuleName = "github.com/franchb/htmgo/framework/v2/h"` (unchanged).

- [ ] **Step 4: Verify rewritten-import count matches**

```bash
grep -rn "github.com/franchb/htmgo/cli/htmgo/v2/" cli/htmgo --include="*.go" | wc -l
```

Expected: `46`

- [ ] **Step 5: Build the CLI**

```bash
cd cli/htmgo && go build ./... && cd -
```

Expected: exit 0, no output.

- [ ] **Step 6: Run the CLI tests**

```bash
cd cli/htmgo && go test ./... && cd -
```

Expected: `ok` for each package. Per CLAUDE.md, `cli/htmgo/tasks/astgen` is the largest CLI test surface — pay attention that it passes.

- [ ] **Step 7: Commit (bundles Task 3 + Task 4)**

```bash
git add cli/htmgo/go.mod cli/htmgo
git commit -m "$(cat <<'EOF'
fix(cli): declare cli/htmgo as v2 module (Go SIV compliance) — #14

- go.mod module path: cli/htmgo → cli/htmgo/v2
- Rewrite 46 internal imports across internal/**, tasks/** to /v2

Required by Go's Semantic Import Versioning rule: a submodule with its
own go.mod at major version ≥ 2 must declare the /v2 suffix in its module
path. The install invocation becomes
`go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1`. The binary
name (`htmgo`) is unchanged.

Replace directives are still present — removed in the next commit.
EOF
)"
```

---

## Task 5: Remove `replace` directives and bump required versions in `cli/htmgo/go.mod`

**Files:**
- Modify: `cli/htmgo/go.mod`

- [ ] **Step 1: Edit `go.mod`** — three changes

  (a) In the `require (` block, change:

  ```
      github.com/franchb/htmgo/framework/v2 v2.0.0-20260423190209-1102e671d216
      github.com/franchb/htmgo/tools/html-to-htmgo/v2 v2.0.0-20260423190209-1102e671d216
  ```

  to:

  ```
      github.com/franchb/htmgo/framework/v2 v2.0.1
      github.com/franchb/htmgo/tools/html-to-htmgo/v2 v2.0.1
  ```

  (b) Delete these two lines near the bottom of the file:

  ```
  replace github.com/franchb/htmgo/framework/v2 => ../../framework

  replace github.com/franchb/htmgo/tools/html-to-htmgo/v2 => ../../tools/html-to-htmgo
  ```

- [ ] **Step 2: Verify the edits**

```bash
grep -E "^replace|framework/v2 v" cli/htmgo/go.mod
```

Expected:
```
	github.com/franchb/htmgo/framework/v2 v2.0.1
	github.com/franchb/htmgo/tools/html-to-htmgo/v2 v2.0.1
```

(Zero `replace` lines. Two `v2.0.1` require lines.)

- [ ] **Step 3: Verify build still works (via go.work)**

```bash
cd cli/htmgo && go build ./... && cd -
```

Expected: exit 0. `go.work` resolves `framework/v2` and `tools/html-to-htmgo/v2` from sibling directories on disk, ignoring the unresolvable `v2.0.1` require (which doesn't exist as a tag yet).

- [ ] **Step 4: Run CLI tests**

```bash
cd cli/htmgo && go test ./... && cd -
```

Expected: all pass.

**Reminder:** do not run `go mod tidy`.

- [ ] **Step 5: Commit**

```bash
git add cli/htmgo/go.mod
git commit -m "$(cat <<'EOF'
fix(cli): remove sibling-path replaces from cli/htmgo/go.mod — #14

Go rejects non-main modules that contain local-path replace directives.
Moving this resolution to the root go.work (which only affects local
dev) unblocks downstream `go install`.

Also pins required framework/v2 and tools/html-to-htmgo/v2 at v2.0.1
so the CLI tagged at v2.0.1 transitively pulls the matching companions
rather than the broken v2.0.0 pseudo-versions.
EOF
)"
```

---

## Task 6: Fix `framework-ui/go.mod`

**Files:**
- Modify: `framework-ui/go.mod`

- [ ] **Step 1: Edit `go.mod`**

  Delete only this line:
  ```
  replace github.com/franchb/htmgo/framework/v2 => ../framework
  ```

  **Do NOT bump the `require github.com/franchb/htmgo/framework/v2 v2.0.0-…<pseudo>` line.** Go's workspace overrides the module's *implementation* but still uses the `require` version for module-graph resolution — if we bumped to `v2.0.1` before the tag exists, `go build` would try to fetch it from the proxy and fail. The existing pseudo-version points to a correctly-packaged `framework/v2.0.0` module and works fine.

- [ ] **Step 2: Verify the edits**

```bash
grep -E "^replace|framework/v2 " framework-ui/go.mod
```

Expected:
```
require github.com/franchb/htmgo/framework/v2 v2.0.0-20260423190209-1102e671d216
```

(Zero `replace` lines; the existing pseudo-version require is unchanged.)

- [ ] **Step 3: Verify build**

```bash
cd framework-ui && go build ./... && cd -
```

Expected: exit 0.

- [ ] **Step 4: Commit**

```bash
git add framework-ui/go.mod
git commit -m "$(cat <<'EOF'
fix(framework-ui): remove sibling-path replace from go.mod — #14

The sibling `replace github.com/franchb/htmgo/framework/v2 => ../framework`
directive is rejected by Go when framework-ui is fetched as a dependency.
Local monorepo dev now uses the root go.work for this resolution.

The existing pseudo-version require `v2.0.0-…1102e671d216` of framework/v2
is kept — framework/v2.0.0 is packaged correctly and resolves fine.
EOF
)"
```

---

## Task 7: Fix `extensions/websocket/go.mod`

**Files:**
- Modify: `extensions/websocket/go.mod`

- [ ] **Step 1: Edit `go.mod`**

  Delete only this line:
  ```
  replace github.com/franchb/htmgo/framework/v2 => ../../framework
  ```

  **Do NOT bump the `require github.com/franchb/htmgo/framework/v2 v2.0.0-…<pseudo>` line** in the `require (` block — same reasoning as Task 6.

- [ ] **Step 2: Verify the edits**

```bash
grep -E "^replace|framework/v2 v" extensions/websocket/go.mod
```

Expected:
```
	github.com/franchb/htmgo/framework/v2 v2.0.0-20260423190209-1102e671d216
```

(Zero `replace` lines; the existing pseudo-version require is unchanged.)

- [ ] **Step 3: Verify build**

```bash
cd extensions/websocket && go build ./... && cd -
```

Expected: exit 0.

- [ ] **Step 4: Commit**

```bash
git add extensions/websocket/go.mod
git commit -m "$(cat <<'EOF'
fix(extensions/websocket): remove sibling-path replace from go.mod — #14

Same root cause as cli/htmgo and framework-ui: sibling `replace` rule
violation blocked downstream `go get`. Root go.work handles local
resolution. Existing pseudo-version require of framework/v2 is kept
(same reasoning as Task 6).
EOF
)"
```

---

## Task 8: Full monorepo sanity build + test

**Files:** none (verification only)

- [ ] **Step 1: Build everything under the workspace**

```bash
go build ./framework/... ./framework-ui/... ./extensions/websocket/... ./tools/html-to-htmgo/... ./cli/htmgo/...
```

Expected: exit 0, no output.

- [ ] **Step 2: Run framework tests**

```bash
cd framework && go test ./... && cd -
```

Expected: all pass. Per CLAUDE.md, `h/render_test.go` is the largest suite.

- [ ] **Step 3: Run CLI tests**

```bash
cd cli/htmgo && go test ./... && cd -
```

Expected: all pass.

- [ ] **Step 4: Confirm no `go.mod` in the workspace still has sibling replaces**

```bash
grep -H "^replace" framework/go.mod framework-ui/go.mod extensions/websocket/go.mod tools/html-to-htmgo/go.mod cli/htmgo/go.mod
```

Expected: empty output.

- [ ] **Step 5: Confirm module-path lines are correct**

```bash
grep -H "^module " framework/go.mod framework-ui/go.mod extensions/websocket/go.mod tools/html-to-htmgo/go.mod cli/htmgo/go.mod
```

Expected:
```
framework/go.mod:module github.com/franchb/htmgo/framework/v2
framework-ui/go.mod:module github.com/franchb/htmgo/framework-ui/v2
extensions/websocket/go.mod:module github.com/franchb/htmgo/extensions/websocket/v2
tools/html-to-htmgo/go.mod:module github.com/franchb/htmgo/tools/html-to-htmgo/v2
cli/htmgo/go.mod:module github.com/franchb/htmgo/cli/htmgo/v2
```

All five end in `/v2`. This task has no commit — it's a checkpoint before doc/install updates.

---

## Task 9: Update install snippets in htmgo-site docs

**Files:**
- Modify: `htmgo-site/pages/docs/installation.go`
- Modify: `htmgo-site/pages/docs/misc/formatter.go`

These are Go source files that embed user-facing install instructions on `htmgo.dev`. Leaving them on the old path sends every new reader through the v1.1.1 CLI.

- [ ] **Step 1: Edit `htmgo-site/pages/docs/installation.go` line 27 and line 29**

Change:
```go
ui.SingleLineBashCodeSnippet(`GOPROXY=direct go install github.com/franchb/htmgo/cli/htmgo@latest`),
```
to:
```go
ui.SingleLineBashCodeSnippet(`GOPROXY=direct go install github.com/franchb/htmgo/cli/htmgo/v2@latest`),
```

And the Windows variant on line 29:
```go
ui.SingleLineBashCodeSnippet(`set GOPROXY=direct && go install github.com/franchb/htmgo/cli/htmgo@latest`),
```
to:
```go
ui.SingleLineBashCodeSnippet(`set GOPROXY=direct && go install github.com/franchb/htmgo/cli/htmgo/v2@latest`),
```

- [ ] **Step 2: Edit `htmgo-site/pages/docs/misc/formatter.go` line 19**

Change:
```go
HelpText(`Note: if you have previously installed htmgo, you will need to run GOPROXY=direct go install github.com/franchb/htmgo/cli/htmgo@latest to update the cli tool.`),
```
to:
```go
HelpText(`Note: if you have previously installed htmgo, you will need to run GOPROXY=direct go install github.com/franchb/htmgo/cli/htmgo/v2@latest to update the cli tool.`),
```

- [ ] **Step 3: Verify — no `cli/htmgo@latest` references remain in htmgo-site Go sources**

```bash
grep -rn "cli/htmgo@latest" htmgo-site --include="*.go"
```

Expected: empty output.

```bash
grep -rn "cli/htmgo/v2@latest" htmgo-site --include="*.go"
```

Expected: three matches (two in `installation.go`, one in `misc/formatter.go`).

- [ ] **Step 4: Verify htmgo-site still builds**

```bash
cd htmgo-site && go build ./... && cd -
```

Expected: exit 0.

- [ ] **Step 5: Commit**

```bash
git add htmgo-site/pages/docs/installation.go htmgo-site/pages/docs/misc/formatter.go
git commit -m "$(cat <<'EOF'
docs(site): install snippets use cli/htmgo/v2@latest — #14

After v2.0.1, `go install github.com/franchb/htmgo/cli/htmgo@latest`
resolves to the v1.1.1 CLI (last tag under the v1 path). Docs must
point at the /v2 path so new users land on the fixed CLI.
EOF
)"
```

---

## Task 10: Update `go run cli/htmgo@latest` in example Dockerfiles and Taskfiles

**Files (exhaustive list, from pre-work grep):**
- Modify: `examples/chat/Dockerfile`
- Modify: `examples/chat/Taskfile.yml`
- Modify: `examples/hackernews/Dockerfile`
- Modify: `examples/simple-auth/Dockerfile`
- Modify: `examples/todo-list/Taskfile.yml`
- Modify: `examples/todo-list/Dockerfile`
- Modify: `examples/ws-example/Dockerfile`
- Modify: `examples/ws-example/Taskfile.yml`
- Modify: `htmgo-site/Dockerfile`
- Modify: `templates/starter/Dockerfile`

- [ ] **Step 1: Baseline count**

```bash
grep -rn "go run github.com/franchb/htmgo/cli/htmgo@latest" examples templates htmgo-site \
  | grep -v "/v2@latest" | wc -l
```

Expected: `16`

(9 from `Taskfile.yml` across 3 example dirs × 3 targets each, plus 7 from `Dockerfile`s across `examples/chat`, `examples/hackernews`, `examples/simple-auth`, `examples/todo-list`, `examples/ws-example`, `htmgo-site`, and `templates/starter` — one `RUN go run ...` line per Dockerfile.)

- [ ] **Step 2: Run the rewrite**

```bash
find examples templates htmgo-site \
  \( -name 'Dockerfile' -o -name 'Taskfile.yml' \) -type f -print0 \
  | xargs -0 sed -i 's|github.com/franchb/htmgo/cli/htmgo@latest|github.com/franchb/htmgo/cli/htmgo/v2@latest|g'
```

- [ ] **Step 3: Verify — no `cli/htmgo@latest` refs remain outside source files**

```bash
grep -rn "cli/htmgo@latest" examples templates htmgo-site \
  --include='Dockerfile' --include='Taskfile.yml' --include='*.yml'
```

Expected: empty output.

```bash
grep -rn "cli/htmgo/v2@latest" examples templates htmgo-site \
  --include='Dockerfile' --include='Taskfile.yml' --include='*.yml' | wc -l
```

Expected: `16`

- [ ] **Step 4: Smoke-check one example builds**

```bash
cd examples/todo-list && go build ./... && cd -
```

Expected: exit 0. We only check build, not the full `task watch` flow — the full watch flow requires the CLI binary actually installed from the new path, which doesn't exist until tags are pushed.

- [ ] **Step 5: Commit**

```bash
git add examples templates/starter/Dockerfile htmgo-site/Dockerfile
git commit -m "$(cat <<'EOF'
chore(examples): go run cli/htmgo uses /v2 path — #14

Task 9 fixed docs. This covers the 16 `go run` invocations in example
Taskfiles and Dockerfiles so shipped examples don't silently fetch the
v1.1.1 CLI.
EOF
)"
```

---

## Task 11: Add a regression-guard job to `verify-installer-works.yml`

**Files:**
- Modify: `.github/workflows/verify-installer-works.yml`

The existing workflow installs the CLI from local source (`cd cli/htmgo && go install .`) which bypasses the bug entirely. We add a second job that installs via the **public** `/v2` path to fail loudly if this class of regression reappears in a future release.

Note: this new job will **not pass on PR builds** for this change — `v2.0.1` doesn't exist as a tag yet. That's OK; it passes after tags are pushed in Task 13. The job is gated to `push` on `master` only, so it doesn't block this PR's merge.

- [ ] **Step 1: Edit `.github/workflows/verify-installer-works.yml`**

Append a new job at the end of the file (after the existing `build` job). The file currently ends at line 74:

```yaml
      # Step 9: Send curl request to verify the server is running
      - name: Test server with curl
        run: |
          curl --fail http://localhost:3000 || exit 1
```

After that, add:

```yaml

  verify-public-install:
    # Runs on push to master after tags are published. Confirms the
    # public install path works. Do not run on PRs — the tag may not
    # exist yet for the version being shipped.
    if: github.event_name == 'push' && github.ref == 'refs/heads/master'
    runs-on: ubuntu-latest
    steps:
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.26'

      - name: Install htmgo CLI from public module proxy
        run: |
          GOPROXY=https://proxy.golang.org,direct \
            go install github.com/franchb/htmgo/cli/htmgo/v2@latest

      - name: Verify CLI binary runs
        run: |
          htmgo version
```

- [ ] **Step 2: Validate YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/verify-installer-works.yml'))"
```

Expected: exit 0, no output. If `python3` isn't available, use `yamllint` or skip; GitHub will validate on push.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/verify-installer-works.yml
git commit -m "$(cat <<'EOF'
ci: add public-install guard to verify-installer-works — #14

The existing job installs from local source (`go install .`) which can't
detect SIV-path or replace-directive bugs. Add a second job that installs
via the public /v2 path to catch a future regression of the bug in #14
before it reaches users.

Gated to push-on-master so it doesn't fail PR builds before the tag
exists.
EOF
)"
```

---

## Task 12: Update `CHANGELOG.md`

**Files:**
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Open `CHANGELOG.md` and locate the `## [Unreleased]` header (line ~7)**

Replace `## [Unreleased]` with a new `v2.0.1` section, keeping `## [Unreleased]` as an empty header above it for future work:

```markdown
## [Unreleased]

## [2.0.1] - 2026-04-24

### Fixed

- **`cli/htmgo` install.** `cli/htmgo/go.mod` now declares the `/v2` module path required by Go's Semantic Import Versioning rules (`module github.com/franchb/htmgo/cli/htmgo/v2`). The install invocation is now `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1`. The binary name (`htmgo`) is unchanged. See [#14](https://github.com/franchb/htmgo/issues/14).
- **Sibling `replace` directives removed.** `cli/htmgo`, `framework-ui`, and `extensions/websocket` `go.mod` files no longer contain `replace github.com/franchb/htmgo/... => ../...` lines. These violated Go's rule that non-main modules must not carry local-path replaces and blocked all three from `go get` / `go install`. Local monorepo development now uses a root `go.work` file.

### Changed

- **Install path for `go install` / `go run` users.** All docs (`htmgo-site`), Dockerfiles, and Taskfiles under `examples/`, `templates/`, and `htmgo-site/` now reference `github.com/franchb/htmgo/cli/htmgo/v2@latest` instead of `github.com/franchb/htmgo/cli/htmgo@latest`. The old path still resolves to the v1.1.1 CLI (the last tag under the v1 module path), which is incompatible with `framework/v2`.

### Chore

- `framework/v2.0.1` and `tools/html-to-htmgo/v2.0.1` are republished at the fix commit with **no code change** so all five v2 submodule versions stay aligned.
```

- [ ] **Step 2: Correct the erroneous v2.0.0 CLI-path claim**

In the `## [2.0.0] - 2026-04-23` section, find the line:

```
The CLI (`github.com/franchb/htmgo/cli/htmgo`) is a binary and keeps its path unchanged — continue using `go run github.com/franchb/htmgo/cli/htmgo@latest` (or `@v2.0.0`).
```

Replace it with:

```
> **Correction (v2.0.1):** The claim that the CLI path was unchanged was incorrect. Go's Semantic Import Versioning applies to any submodule with its own `go.mod` at major ≥ 2, and the `v2.0.0` CLI tags are un-installable as a result. See [CHANGELOG v2.0.1](#201---2026-04-24) and [#14](https://github.com/franchb/htmgo/issues/14). Use `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1` or later.
```

- [ ] **Step 3: Verify the CHANGELOG parses cleanly**

```bash
head -50 CHANGELOG.md
```

Confirm visually: `[Unreleased]` is above `[2.0.1]`, which is above `[2.0.0]`, and the v2.0.0 section has the correction blockquote in place of the old claim.

- [ ] **Step 4: Commit**

```bash
git add CHANGELOG.md
git commit -m "$(cat <<'EOF'
docs(changelog): v2.0.1 release notes — #14

- Document the SIV fix for cli/htmgo and the replace-removal for
  framework-ui and extensions/websocket.
- Correct the erroneous v2.0.0 claim that the CLI install path was
  unchanged (SIV does not permit that for submodules with their own
  go.mod at major ≥ 2).
EOF
)"
```

---

## Task 13: Open the PR and merge after CI

**Files:** none (repo operations)

- [ ] **Step 1: Push the branch**

```bash
git push -u origin fix/v2-install-issue-14
```

- [ ] **Step 2: Open the PR**

```bash
gh pr create \
  --title "fix: cli/htmgo v2 install + cross-module replace cleanup (#14)" \
  --body "$(cat <<'EOF'
## Summary

Fixes #14. Makes `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1` succeed from a clean environment and fixes the same sibling-`replace` violation in `framework-ui` and `extensions/websocket`.

## Changes

- **New file:** root `go.work` listing the five v2 submodules. Local monorepo dev now resolves sibling modules from the workspace, not from in-`go.mod` `replace` directives.
- **`cli/htmgo`:** module path renamed to `…/cli/htmgo/v2` (Go SIV compliance); 46 internal imports rewritten; two `replace` lines removed. The existing pseudo-version `require` lines of `framework/v2` and `tools/html-to-htmgo/v2` are kept — `v2.0.0` of those modules is packaged correctly, so the pseudo-versions resolve fine.
- **`framework-ui`, `extensions/websocket`:** sibling `replace` removed. Existing pseudo-version `require` of `framework/v2` kept (same reasoning).
- **Docs:** `htmgo-site` install snippets and help text now point at `cli/htmgo/v2@latest`.
- **Examples/templates:** 16 `go run cli/htmgo@latest` references updated to the `/v2` path (Taskfiles, Dockerfiles).
- **CI:** new `verify-public-install` job guards the regression post-merge.
- **CHANGELOG:** `v2.0.1` entry added; v2.0.0 CLI-path claim corrected in place.

## Non-goals

- `examples/*/go.mod`, `htmgo-site/go.mod`, `templates/starter/go.mod` still contain sibling `replace` directives. These are main modules where `replace` is legal; out of scope here.
- `framework/go.mod`, `tools/html-to-htmgo/go.mod` have no code changes — they get a version-only bump to `v2.0.1` at tag time so all v2 submodule versions stay aligned.

## Tag plan (post-merge)

Tags are pushed in this order to minimize the proxy-warmup window where a consumer might fetch `cli/htmgo/v2.0.1` before `framework/v2.0.1` is resolvable:

1. `framework/v2.0.1`
2. `tools/html-to-htmgo/v2.0.1`
3. `framework-ui/v2.0.1`
4. `extensions/websocket/v2.0.1`
5. `cli/htmgo/v2.0.1`

## Test plan

- [x] `go build ./framework/... ./framework-ui/... ./extensions/websocket/... ./tools/html-to-htmgo/... ./cli/htmgo/...`
- [x] `cd framework && go test ./...`
- [x] `cd cli/htmgo && go test ./...`
- [x] `grep -H '^replace' */go.mod */*/go.mod` returns zero matches in the five library modules
- [x] All five modules have `/v2` in their `module` line
- [ ] Post-merge: `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1` succeeds from a clean env
- [ ] Post-merge: `go get github.com/franchb/htmgo/framework-ui/v2@v2.0.1` succeeds
- [ ] Post-merge: `go get github.com/franchb/htmgo/extensions/websocket/v2@v2.0.1` succeeds
- [ ] Post-merge: regenerating `__htmgo/pages-generated.go` in a throwaway consumer emits `import "github.com/franchb/htmgo/framework/v2/h"`

Ref: `docs/plans/2026-04-24-cli-htmgo-v2-install-fix-design.md` for the full design.
EOF
)"
```

- [ ] **Step 3: Wait for CI green**

```bash
gh pr checks --watch
```

Expected: all required checks pass. The new `verify-public-install` job will be **skipped** (its `if` condition gates it to push-on-master).

- [ ] **Step 4: Merge**

Wait for user approval before merging — this is a release-blocking change and deserves an explicit go-ahead. Once approved:

```bash
gh pr merge --squash --delete-branch
```

---

## Task 14: Tag the v2.0.1 release (post-merge)

**Files:** none (git operations against `master`)

- [ ] **Step 1: Sync local master**

```bash
git checkout master
git pull --ff-only origin master
```

- [ ] **Step 2: Identify the merge commit**

```bash
git log -1 --pretty=format:"%H %s"
```

Expected: the squash-merge commit from Task 13. Save the SHA for the tagging commands below — substitute for `<SHA>`.

- [ ] **Step 3: Tag framework/v2.0.1 FIRST**

This must resolve first so downstream modules that `require framework/v2 v2.0.1` can resolve once tagged.

```bash
git tag -s framework/v2.0.1 <SHA> -m "framework v2.0.1 — alignment bump (no code change)"
git push origin framework/v2.0.1
```

Wait ~30 seconds for `proxy.golang.org` to observe the tag:

```bash
until curl -sSf "https://proxy.golang.org/github.com/franchb/htmgo/framework/v2/@v/v2.0.1.info" >/dev/null; do sleep 5; done
```

Expected: the `until` loop exits (proxy knows about the tag).

- [ ] **Step 4: Tag tools/html-to-htmgo/v2.0.1**

```bash
git tag -s tools/html-to-htmgo/v2.0.1 <SHA> -m "tools/html-to-htmgo v2.0.1 — alignment bump (no code change)"
git push origin tools/html-to-htmgo/v2.0.1
```

- [ ] **Step 5: Tag framework-ui/v2.0.1 and extensions/websocket/v2.0.1**

```bash
git tag -s framework-ui/v2.0.1 <SHA> -m "framework-ui v2.0.1 — remove sibling replace; pin framework/v2 v2.0.1"
git tag -s extensions/websocket/v2.0.1 <SHA> -m "extensions/websocket v2.0.1 — remove sibling replace; pin framework/v2 v2.0.1"
git push origin framework-ui/v2.0.1 extensions/websocket/v2.0.1
```

- [ ] **Step 6: Tag cli/htmgo/v2.0.1 LAST**

```bash
git tag -s cli/htmgo/v2.0.1 <SHA> -m "cli/htmgo v2.0.1 — SIV fix (#14), module path now cli/htmgo/v2"
git push origin cli/htmgo/v2.0.1
```

- [ ] **Step 7: Verify all five tags are visible on the remote**

```bash
git ls-remote --tags origin | grep v2.0.1
```

Expected: five lines, one per tag.

---

## Task 15: Clean-environment install verification (post-tag)

**Files:** none (verification only)

- [ ] **Step 1: Set up a clean `GOPATH`**

```bash
export VERIFY_HOME=$(mktemp -d)
export GOPATH=$VERIFY_HOME/go
export GOMODCACHE=$VERIFY_HOME/gomodcache
export GOCACHE=$VERIFY_HOME/gocache
export PATH=$GOPATH/bin:$PATH
```

- [ ] **Step 2: Install the CLI from the public proxy**

```bash
GOPROXY=https://proxy.golang.org,direct \
  go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1
```

Expected: exit 0. If it fails with "proxy: unknown revision", wait 30s and retry — the proxy may not have observed the tag yet.

- [ ] **Step 3: Confirm the binary runs**

```bash
htmgo version
```

Expected: `htmgo cli version <version-string>` where the version reflects the new CLI (not `1.0.6`, which was v1.1.1's output).

- [ ] **Step 4: Confirm `go get` of the two companion modules**

```bash
cd $VERIFY_HOME && mkdir consumer && cd consumer
go mod init verify
go get github.com/franchb/htmgo/framework-ui/v2@v2.0.1
go get github.com/franchb/htmgo/extensions/websocket/v2@v2.0.1
```

Expected: both `go get` commands exit 0.

- [ ] **Step 5: Confirm regenerate-smoke on a throwaway consumer**

```bash
cd $VERIFY_HOME
cp -r /home/iru/p/github.com/franchb/htmgo/examples/minimal-htmgo ./minimal-test
cd minimal-test
# Strip the sibling replace (simulating a real downstream user):
sed -i '/^replace /d' go.mod
go mod tidy
htmgo generate
grep '"github.com/franchb/htmgo/framework/v2/h"' __htmgo/pages-generated.go
```

Expected: the `grep` matches — regenerated code imports `framework/v2/h`, not the un-prefixed `framework/h`.

- [ ] **Step 6: Clean up**

```bash
cd $HOME
rm -rf $VERIFY_HOME
unset VERIFY_HOME GOPATH GOMODCACHE GOCACHE
```

- [ ] **Step 7: Post verification result to the PR / issue #14**

```bash
gh issue comment 14 --body "$(cat <<'EOF'
Verified v2.0.1 install works from a clean environment:

- `go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1` ✓
- `htmgo version` runs ✓
- `go get github.com/franchb/htmgo/framework-ui/v2@v2.0.1` ✓
- `go get github.com/franchb/htmgo/extensions/websocket/v2@v2.0.1` ✓
- Regenerated `__htmgo/pages-generated.go` imports `framework/v2/h` ✓

Closing.
EOF
)"
gh issue close 14
```

---

## Task 16: Amend the GitHub release notes for v2.0.0 (manual)

**Files:** GitHub UI (no local edits).

- [ ] **Step 1: Navigate to the v2.0.0 release page**

```bash
gh release view v2.0.0 --web  # or the correct tag name used on GitHub
```

If the original release was published against a non-submodule tag (e.g., `v2.0.0` as an umbrella) or per-submodule, view each.

- [ ] **Step 2: Edit the release description**

Prepend a correction notice to each v2.0.0 release body:

```markdown
> ⚠ **v2.0.0 is broken for downstream users.** The `cli/htmgo/v2.0.0`,
> `framework-ui/v2.0.0`, and `extensions/websocket/v2.0.0` tags cannot be
> fetched by `go install` / `go get`. Use **v2.0.1 or later**. See
> [#14](https://github.com/franchb/htmgo/issues/14) and
> [CHANGELOG v2.0.1](../blob/master/CHANGELOG.md#201---2026-04-24).
```

- [ ] **Step 3: Publish v2.0.1 release notes**

```bash
gh release create cli/htmgo/v2.0.1 \
  --title "cli/htmgo v2.0.1 — SIV install fix" \
  --notes-file <(cat <<'EOF'
## Fixed

`cli/htmgo/go.mod` now declares the `/v2` module path required by Go's Semantic Import Versioning rules. Install is:

```
go install github.com/franchb/htmgo/cli/htmgo/v2@v2.0.1
```

Also removes sibling `replace` directives that blocked `cli/htmgo`, `framework-ui`, and `extensions/websocket` from downstream fetch. Local monorepo dev now uses a root `go.work`.

See [CHANGELOG](../blob/master/CHANGELOG.md#201---2026-04-24) and [#14](https://github.com/franchb/htmgo/issues/14).
EOF
)
```

Repeat for the other four v2.0.1 tags with appropriate titles. Framework and tools/html-to-htmgo releases say "alignment bump, no code change."

---

## Self-Review Notes

Checked against spec `docs/plans/2026-04-24-cli-htmgo-v2-install-fix-design.md` on 2026-04-24:

**Spec coverage:**

| Spec section | Covered by |
| --- | --- |
| §3 Approach — root `go.work` + `cli/htmgo` SIV rename | Tasks 2, 3, 4 |
| §4 Per-module changes table | Tasks 3–7 |
| §5 Root `go.work` + `.gitignore` | Task 2 |
| §6 Internal-import rewrite procedure | Task 4 |
| §7 Release / tagging plan | Tasks 12, 13, 14 |
| §8 CHANGELOG entry | Task 12 |
| §9 Release-notes / install snippet amendment | Tasks 9, 10, 12, 16 |
| §10 Verification (clean-env install, regenerate smoke, CI guard) | Tasks 11, 15 |
| §11 Rollback | (text note; no implementation task — acceptable: rollback is "don't force-push, tag v2.0.2 instead") |
| §12 Out of scope | (documented in File Structure section as "Unchanged") |

**Placeholder scan:** no `TBD`, `TODO`, "add appropriate X", or "similar to Task N" references. Every code block is concrete.

**Type / name consistency:** no code-level types in this plan — it's all file edits, shell commands, and git/gh operations. Path names and tag names (`cli/htmgo/v2.0.1`, `framework/v2`, etc.) are consistent throughout.

**Scope check:** single coordinated release. All tasks produce working, reviewable state. Task 8 is a pre-doc-change sanity gate. Tasks 14–16 are post-merge release-ops and are clearly labeled as such.
