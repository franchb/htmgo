# Contributing to htmgo

Thanks for helping. This file covers how the repository is laid out and
how to build and test it. Please also read the
[Code of Conduct](CODE_OF_CONDUCT.md); security issues go through
[SECURITY.md](SECURITY.md), not public issues.

## Prerequisites

- Go 1.27.0 or newer
- Node.js 22 and npm (JS extensions, Tailwind plugins for htmgo-site)
- Optional: [Task](https://taskfile.dev) for the examples' `Taskfile.yml`
- Optional: Docker or Podman to build the release images

## Repository layout

This is a multi-module monorepo.

| Directory                | Module                                             |
| ------------------------ | -------------------------------------------------- |
| `framework/`             | `github.com/franchb/htmgo/framework/v2`            |
| `framework-ui/`          | `github.com/franchb/htmgo/framework-ui/v2`         |
| `extensions/websocket/`  | `github.com/franchb/htmgo/extensions/websocket/v2` |
| `tools/html-to-htmgo/`   | `github.com/franchb/htmgo/tools/html-to-htmgo/v2`  |
| `cli/htmgo/`             | `github.com/franchb/htmgo/cli/htmgo/v2` (the CLI)  |
| `htmgo-site/`            | documentation site (app)                           |
| `templates/starter/`     | template copied by `htmgo template` (app)          |
| `examples/*`             | example apps                                       |

### go.work and GOWORK=off

The root `go.work` lists the library modules (and the astgen test
sample), so a change in `framework/` is picked up by `framework-ui/`,
`cli/htmgo/` and the others without publishing a version.

The apps (`htmgo-site`, `templates/starter`, `examples/*`) are **not** in
`go.work`. They reach the local framework through `replace` directives in
their own `go.mod`. Run Go commands in them with `GOWORK=off`, or Go
fails with "current directory is contained in a module that is not one
of the workspace modules listed in go.work":

```bash
cd htmgo-site
GOWORK=off go mod tidy
GOWORK=off htmgo build
```

CI also runs every library module with `GOWORK=off`, so each `go.mod`
must be complete on its own. After changing dependencies, run
`GOWORK=off go mod tidy` in that module.

## Building and running

Install the CLI from your checkout:

```bash
cd cli/htmgo && go install .
```

Inside an app:

```bash
htmgo generate   # regenerate __htmgo/ (routes, assets) from the source tree
htmgo watch      # dev server with live reload
htmgo build      # production binary in ./dist
```

`__htmgo/` is generated and gitignored. Never edit it by hand, and do not
commit it.

For htmgo-site, run `npm install` first (Tailwind plugins).

## Tests

What CI runs (`.github/workflows/tests.yml`), per library module:

```bash
for m in framework framework-ui cli/htmgo extensions/websocket tools/html-to-htmgo; do
  (cd $m && GOWORK=off go vet ./... && GOWORK=off go test -race ./...) || echo "FAIL $m"
done
```

A single test:

```bash
cd framework && go test ./h/ -run TestSimpleRender
```

JS extensions (htmx 4, vitest):

```bash
cd framework/assets/js
npm ci
npm test         # vitest run
npm run build    # tsup -> framework/assets/dist/htmgo.js
```

Rebuild and commit `framework/assets/dist/htmgo.js` when you change the
TypeScript sources.

Also run before opening a PR:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
(cd <module> && GOWORK=off govulncheck ./...)
```

Tests use `stretchr/testify` (`assert`, `require`) and `t.Parallel()`
where possible. Add a regression test with every bug fix.

## Commits and pull requests

- Keep each commit to one logical change, and each PR to one topic.
- Commit subjects use a Conventional-Commits-style prefix, with an
  optional scope:
  - `feat(framework): …` new functionality
  - `fix(cli): …` bug fix
  - `test(html-to-htmgo): …` tests only
  - `docs: …`, `docs(changelog): …` documentation
  - `ci: …`, `ci(release): …` workflows, Docker
  - `chore: …` maintenance, tidying
  - `security: …` vulnerability fixes
  - `deps(<module>): …` dependency bumps (Dependabot uses this)
- Explain *why* in the body, not only what.
- Add an entry under `## [Unreleased]` in `CHANGELOG.md` for any
  user-visible change, and call out breaking changes.
- Never push directly to `master`; open a PR and wait for green CI.

## Releases

Releases are automatic. Every merge to `master` that changes a module in a
way that matters to its users is tagged and released by the
[Release workflow](.github/workflows/release.yml) once the Tests (and, for
framework changes, the dependency updater) have succeeded on it.

Each module is tagged separately, with its directory as the prefix
(`framework/v2.1.0`, `cli/htmgo/v2.1.1`; see
[the Go modules reference](https://go.dev/ref/mod#vcs-version)), and every
release also gets a repository-wide tag (`v2.2.0`). That release is marked
latest and carries the `htmgo` binaries, `SHA256SUMS` and build provenance.
Asset names have no version in them, so
`https://github.com/franchb/htmgo/releases/latest/download/htmgo_linux_amd64.tar.gz`
always points to the newest build.

How versions are chosen (`tools/release`):

- A module takes part if it has at least one `<dir>/vX.Y.Z` tag. The
  examples, the site and the templates have none and are never released.
  To release a new module, tag its first version by hand.
- Its next version comes from the commits since its latest tag on
  `master` that touch its directory (nested modules excluded). The commit
  subject (the squash-merge title, i.e. the PR title) decides the bump:
  - `type!:` or a `BREAKING CHANGE:` footer: major
  - `feat`: minor
  - `fix`, `perf`, `security`, `revert`, `deps`, `build(deps)`,
    `chore(deps-dev)`, and the updater's "Auto-update HTMGO framework
    version" commits: patch
  - anything else (`chore`, `ci`, `docs`, `test`, `refactor`, …): no release
- The repository tag gets the largest bump of the modules released, but at
  most a minor one.
- A breaking change cannot be released as v2.x, and a /v2 module path cannot
  carry v3 tags ([major version suffixes](https://go.dev/ref/mod#major-version-suffixes)).
  So a breaking commit stops the release workflow with an error until either
  the module path moves to `/v3` (the next release is then `v3.0.0`) or a
  maintainer decides the change is compatible and pushes the next tag by hand
  (later releases continue from it).
- A module whose `go.mod` has `replace` directives is not released.

Try a plan locally (it only reads the repository):

```bash
cd tools/release && GOWORK=off go run . -repo ../.. -head origin/master \
  -out /tmp/release -repo-url https://github.com/franchb/htmgo -binary cli/htmgo=htmgo
```

or run the Release workflow by hand with "dry run" checked: it plans and
builds, and creates no tags or releases.

After a framework change lands on master, the
"Update HTMGO Framework Dependency" workflow bumps the other modules'
framework requirement first, so the tags land on that commit.
