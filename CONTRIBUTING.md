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

Each module is tagged separately with its directory as prefix, for
example `framework/v2.1.0`, `cli/htmgo/v2.1.0`, plus a repository-wide
`v2.1.0` tag. After a framework change lands on master, the
"Update HTMGO Framework Dependency" workflow bumps the apps' framework
requirement automatically.
