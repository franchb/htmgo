---
name: htmgo-guidance
description: Use when writing Go code that uses the htmgo framework (github.com/franchb/htmgo), building pages, partials, or components with h.Div/h.Button/h.Ren, wiring hx/ htmx attributes or ax/ Alpine directives, or answering questions about htmgo patterns, routing, and best practices. Covers Fiber v3 integration, RequestContext, auto-routing, service locator, caching, and the Alpine compat extension.
---

# htmgo Guidance

*Last reviewed against: htmgo `franchb/htmgo` v1.2.0-beta.2 (htmx4-migration branch), htmx 4.0.0-beta2, Alpine.js 3.15.11. If framework signatures have changed since, check the source files under `framework/` for ground truth.*

## 1. What htmgo is

htmgo is a lightweight Go web framework for building interactive SSR websites without a JavaScript build step. You write Go code using a builder API (`h.Div(...)`, `h.Button(...)`) that produces HTML; htmx handles server round-trips; Alpine.js is optional for pure-client state.

**Key facts:**
- **Import path:** `github.com/franchb/htmgo/framework` (this fork) — upstream is `maddalax/htmgo`; APIs are close but this fork has diverged (htmx 4, Fiber v3, `framework/ax/` Alpine helpers).
- **Go version:** 1.23+.
- **HTTP layer:** Fiber v3 (`github.com/gofiber/fiber/v3`).
- **htmx version:** 4.0.0-beta2, pinned in `framework/assets/js/package.json`. Bundled into `/public/htmgo.js`.
- **Alpine.js (optional):** consumers load Alpine 3.15.11 themselves; the `alpine-compat` htmx extension is pre-bundled in `htmgo.js` and auto-gates on `window.Alpine` presence.
- **Styling:** Tailwind CSS optional (set `tailwind: true` in `htmgo.yml`).
- **Output:** single deployable Go binary with assets embedded.

**The three main Go packages you'll touch as a consumer:**
- `framework/h` — the HTML builder, rendering, routing primitives, request context, lifecycle events, caching helpers, JS command DSL.
- `framework/hx` — constants for htmx attributes, events, headers, swap types.
- `framework/ax` — constants + builder helpers for Alpine.js directives (this fork's addition; mirrors `hx/` shape).

The rest of this skill walks each of these plus related subsystems.
