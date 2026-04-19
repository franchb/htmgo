# htmgo-guidance skill — resume notes (2026-04-18)

Checkpoint for continuing execution of `docs/superpowers/plans/2026-04-18-htmgo-claude-skills.md` in a fresh Claude Code session.

## State

- **Branch:** `htmx4-migration` (not pushed since these 6 commits — push to update PR #12 before merging, or hold until smoke test passes)
- **Spec:** `docs/superpowers/specs/2026-04-18-htmgo-claude-skills-design.md`
- **Plan:** `docs/superpowers/plans/2026-04-18-htmgo-claude-skills.md`
- **File being built:** `.claude/skills/htmgo-guidance/SKILL.md` (currently ~464 lines with §1–§5)

## Commits landed (in order)

```
76d470d docs(skills): scaffold htmgo-guidance skill with frontmatter and §1
c9b951f docs(skills): §2 h.Ren builder model
f63acae fix(skills): correct h.IterMap ordering note in §2
b977b9d docs(skills): §3 Pages vs Partials and §4 RequestContext
d104f20 docs(skills): §5 htmx integration via hx/ and h.Hx* helpers
d86365f fix(skills): correct phantom-API references in §3 and §5
```

## Completed tasks

- [x] Task 1: Scaffold + frontmatter + §1 "What htmgo is"
- [x] Task 2: §2 h.Ren builder model (+ fix for IterMap ordering claim)
- [x] Task 3: §3 Pages vs Partials + §4 RequestContext
- [x] Task 4: §5 htmx integration (+ fix for phantom HxGet/HxPost/HxSwap)

## Remaining tasks (in order)

- [ ] Task 5: §6 Alpine integration (ax/ + compat) — `~130 lines, largest remaining section`
- [ ] Task 6: §7 Lifecycle commands + §8 Service locator — `~80 lines combined`
- [ ] Task 7: §9 Caching + §10 Project config & CLI — `~80 lines combined`
- [ ] Task 8: §11 Common pitfalls — `~50 lines, symptom→fix pairs`
- [ ] Task 9: §12 Upgrade pointers + version stamp update + full read-through + line-count check (cap: 900)
- [ ] Task 10: Append "Claude Code skills" subsection to `README.md` (before Star History)
- [ ] Task 11: `cp -r` skill into `/home/iru/p/gitlab.etecs.ru/services/vulnerability_catalog/.claude/skills/`, commit in that repo
- [ ] Task 12: User-driven Claude Code smoke test in vuln_catalog (manual, no commit)

## Important learnings from Tasks 1–4 — apply to remaining sections

**The plan's body text has many phantom/wrong API references.** Every implementer has had to grep source and correct several names. Expect the same for remaining sections. Survey-first discipline is non-negotiable.

Concrete patterns already caught:

| Phantom in plan | Real API |
|---|---|
| `h.HxGet/HxPost/HxPut/HxPatch/HxDelete` | `h.Get(url, trigger...)` / `h.Post(url, trigger...)` (composite, in `framework/h/xhr.go`) or raw `h.Attribute(hx.GetAttr, url)` |
| `h.HxSwap(spec)` | `h.Attribute(hx.SwapAttr, spec)` |
| `h.Map` / `h.Range` (as Ren builders) | `h.List[T]` (real), `h.IterMap[T]` (real) |
| `h.NewHeader(k, v)` | `h.NewHeaders(k, v, ...)` (variadic string pairs; returns `*h.Headers`) |
| `h.NewPartialWithHeaders(root, headers)` | `h.NewPartialWithHeaders(headers *Headers, root *Element)` — args reversed |
| `ctx.Redirect(url)` | `ctx.Redirect(path, code)` — two args |
| `hx.NewTrigger("click", modifiers...)` | `hx.NewTrigger(opts ...TriggerEvent)` — takes pre-built events |
| `hx.Once()`, `hx.KeyEquals`, `hx.From`, `hx.DelayMs` | `hx.OnceModifier{}`, `hx.StringModifier(...)`, `hx.Delay(n int)` (seconds, not ms) |

**Builders that ARE real** (plan was right): `h.HxTarget`, `h.HxConfirm`, `h.HxInclude`, `h.HxIndicator`, `h.HxTrigger`, `h.HxTriggerString`, `h.HxTriggerClick`, all 10 `*Inherited` helpers.

## Sections likely to have more phantom APIs

- **§6 Alpine** — the plan names `ax.*` helpers; these SHOULD mostly be real (we shipped the ax package), but verify signatures (especially `ax.On`, `ax.Bind` variadic shape, `ax.ModelDebounce(expr, duration)`).
- **§7 Lifecycle + js/ commands** — plan names `h.OnClick(cmd...)`, `h.HxOnAfterSwap`, `h.HxBeforeRequest`, `js.Alert`, `js.SetValue`, `js.AddClass`, `js.EvalJs`, etc. GREP THESE FIRST — likely mixed real/phantom per pattern above.
- **§8 Service locator** — plan names `service.NewLocator`, `service.Set[T]`, `service.Get[T]`, `Singleton`/`Transient`, `h.AppOpts{ServiceLocator: ...}`. Verify the `h.AppOpts` shape and whether it's `ServiceLocator` or a different field name.
- **§9 Caching** — plan names `h.Cache(ctx, key, ttl, compute)`, `h.CacheGlobal`. Verify the exact signature of `h.Cache` (especially whether it takes `ctx` first or not).
- **§10 CLI** — the plan subcommands list (`htmgo setup/build/watch/generate/css/format/template/version`) should be accurate, but spot-check against `cli/htmgo/`.

## Resume prompt for a fresh Claude Code session

> I'm continuing the htmgo-guidance skill implementation from commit `d86365f` on branch `htmx4-migration`. Resume notes are at `docs/superpowers/plans/2026-04-18-htmgo-claude-skills-RESUME.md`. Plan is at `docs/superpowers/plans/2026-04-18-htmgo-claude-skills.md`. Pick up at Task 5 (§6 Alpine integration). Use the superpowers:subagent-driven-development skill, one subagent per task, and for every task the implementer should grep `framework/` source BEFORE writing any API name into the skill — there are several plan-vs-reality mismatches documented in the resume notes.

## Post-completion, before merging

After Task 12 passes:

1. `wc -l .claude/skills/htmgo-guidance/SKILL.md` — verify under 900 lines.
2. `git log --oneline d86365f..HEAD` — expect ~7–10 more commits.
3. `git push origin htmx4-migration` — pushes to PR #12.
4. Add a short comment to PR #12 summarizing the skill addition (reference the upcoming commit range).

## Context on why the split

The current session executed Tasks 1–4 over many turns with large system-reminder overhead per turn. A fresh session starts with clean context and can carry Tasks 5–12 efficiently. State is clean (branch, commits, files all on disk); no reconstruction needed.
