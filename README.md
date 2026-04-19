## **htmgo**

### build simple and scalable systems with go + htmx

-------
[![Go Report Card](https://goreportcard.com/badge/github.com/franchb/htmgo)](https://goreportcard.com/report/github.com/franchb/htmgo)
![Build](https://github.com/franchb/htmgo/actions/workflows/run-framework-tests.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/franchb/htmgo/framework@v1.0.2/h.svg)](https://htmgo.dev/docs)
[![Join Discord](https://img.shields.io/badge/Join%20Discord-gray?style=flat&logo=discord&logoColor=white&link=https://htmgo.dev/discord)](https://htmgo.dev/discord)




<sup>looking for a python version? check out: https://fastht.ml</sup>

**introduction:**

htmgo is a lightweight pure go way to build interactive websites / web applications using go & htmx.

By combining the speed & simplicity of go + hypermedia attributes ([htmx](https://htmx.org)) to add interactivity to websites, all conveniently wrapped in pure go, you can build simple, fast, interactive websites without touching javascript. All compiled to a **single deployable binary**.

```go
func IndexPage(ctx *h.RequestContext) *h.Page {
  now := time.Now()
  return h.NewPage(
    h.Div(
      h.Class("flex gap-2"),
      h.TextF("the current time is %s", now.String())
    )
  )
}
```

**core features:**

1. deployable single binary
2. live reload (rebuilds css, go, ent schema, and routes upon change)
3. automatic page and partial registration based on file path
4. built in tailwindcss support, no need to configure anything by default
5. custom [htmx extensions](https://github.com/franchb/htmgo/tree/b610aefa36e648b98a13823a6f8d87566120cfcc/framework/assets/js/htmxextensions) to reduce boilerplate with common tasks

**get started:**

View documentation on [htmgo.dev](https://htmgo.dev/docs).

## Claude Code skills

This repo ships [Claude Code](https://docs.claude.com/en/docs/claude-code/overview) skills under `.claude/skills/` that teach AI coding sessions how to work with htmgo:

- `htmgo-guidance` — writing htmgo apps (builder, pages, partials, hx/ax helpers, caching, CLI).
- `htmx-guidance` — htmx 4 patterns and best practices.
- `htmx-debugging`, `htmx-extension-authoring`, `htmx-migration`, `htmx-upgrade-from-htmx2` — specialized htmx skills.

To use `htmgo-guidance` in a project that consumes this fork:

```bash
# From your consumer project root:
mkdir -p .claude/skills

# If you have htmgo cloned locally:
cp -r /path/to/htmgo/.claude/skills/htmgo-guidance .claude/skills/

# Or fetch directly:
git clone --depth=1 https://github.com/franchb/htmgo.git /tmp/htmgo
cp -r /tmp/htmgo/.claude/skills/htmgo-guidance .claude/skills/
rm -rf /tmp/htmgo
```

Run `/skills` in Claude Code to verify `htmgo-guidance` is loaded. Repeat for any other skills you want.

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=franchb/htmgo&type=Date)](https://star-history.com/#franchb/htmgo&Date)
