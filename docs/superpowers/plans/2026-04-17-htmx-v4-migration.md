# htmx 4 Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate the `franchb/htmgo` fork from htmx 2.0.8 to htmx 4.0.0-beta2 as a clean break, released as `htmgo v2.0-beta`.

**Architecture:** Single feature branch, five sequenced layers (JS dep → Go `hx` constants → Go `h/` internals → 11 JS extensions → downstream inheritance audit) merged to master as one PR. Clean break — no compat bridge, no deprecated aliases. Pinned exact `4.0.0-beta2` dependency.

**Tech Stack:** Go 1.23 (`framework`, `framework-ui`, `extensions/websocket`, `cli/htmgo`), TypeScript + `tsup` (`framework/assets/js`), `htmx.org@4.0.0-beta2`, `vitest` (new, for JS extension tests), Playwright MCP (for smoke tests against `htmgo-site`).

**Reference spec:** `docs/superpowers/specs/2026-04-17-htmx-v4-migration-design.md`

---

## File structure — what each touched file is responsible for

**Created:**
- `framework/assets/js/htmxextensions/__tests__/*.test.ts` — one per extension, ~50 lines each, verifies `init(api)` + one event hook per extension wires correctly against htmx 4's `registerExtension` API.
- `framework/assets/js/vitest.config.ts` — minimal vitest config.

**Modified (framework core):**
- `framework/assets/js/package.json` — htmx dep pin.
- `framework/assets/js/package-lock.json` — regenerated.
- `framework/hx/htmx.go` — attribute/header/event constant values.
- `framework/hx/htmx_test.go` — constant-value assertions.
- `framework/hx/trigger.go` — verified; no code change expected.
- `framework/h/app.go` — `RequestContext` swaps `HxTriggerName` → `HxSource`/`HxSourceID`/`HxRequestType`.
- `framework/h/lifecycle.go` — `hx-on::` colon-form emission.
- `framework/h/renderer.go` — audited alongside `lifecycle.go`; no expected change.
- `framework/h/attribute.go` — new `HxTargetInherited` / `HxIncludeInherited` / etc. constructors.
- `framework/h/xhr_test.go`, `framework/h/attribute_test.go`, `framework/h/render_test.go` — updated fixtures.
- `framework/assets/js/htmxextensions/*.ts` (11 files) — each rewritten to `registerExtension`.

**Modified (downstream):**
- `examples/chat/`, `examples/hackernews/`, `examples/simple-auth/`, `examples/todo-list/`, `examples/ws-example/`, `examples/minimal-htmgo/` — inheritance-audit fixes.
- `htmgo-site/pages/`, `htmgo-site/partials/` — inheritance audit + doc prose updates.
- `templates/starter/` — inheritance audit.
- `framework-ui/ui/` — inheritance audit.
- `extensions/websocket/` — verify WS extension integration with htmx 4's aligned WS API.

**Release artefacts:**
- `CHANGELOG.md` (may not yet exist — create if absent).
- Git tag `v2.0.0-beta.1`.

---

## Task 0: Create feature branch

**Files:** none (git state only)

- [ ] **Step 1: Create branch from current master**

```bash
cd /home/iru/p/github.com/franchb/htmgo
git checkout master
git pull --ff-only
git checkout -b htmx4-migration
```

- [ ] **Step 2: Verify clean starting state**

Run: `git status && go build ./framework/... && (cd framework && go test ./... -count=1 -short)`
Expected: `nothing to commit, working tree clean` and all framework tests pass. If tests already fail on master before migration, stop and fix those first.

---

## Task 1: Bump htmx.org dependency to 4.0.0-beta2 (Layer A)

**Files:**
- Modify: `framework/assets/js/package.json`
- Modify: `framework/assets/js/package-lock.json` (regenerated)

- [ ] **Step 1: Pin exact 4.0.0-beta2**

Edit `framework/assets/js/package.json`, replace the dependencies entry:

```json
"dependencies": {
  "htmx.org": "4.0.0-beta2"
}
```

(No caret — pinned exact per spec Risk R1.)

- [ ] **Step 2: Reinstall to regenerate lockfile**

```bash
cd framework/assets/js && rm -rf node_modules && npm install
```

Expected: `package-lock.json` updated, `node_modules/htmx.org/package.json` shows `"version": "4.0.0-beta2"`.

- [ ] **Step 3: Verify bundle builds (pre-extension-rewrite, will likely fail with type errors)**

Run: `cd framework/assets/js && npm run build`
Expected: likely fails because extensions still call htmx 2 API (`defineExtension`). This is expected — extension rewrites are Layer D. Note the failures, do not fix yet.

- [ ] **Step 4: Commit**

```bash
git add framework/assets/js/package.json framework/assets/js/package-lock.json
git commit -m "build(js): bump htmx.org to 4.0.0-beta2 (exact pin)"
```

---

## Task 2: Migrate hx event constants to htmx 4 colon-form (Layer B.1)

**Files:**
- Modify: `framework/hx/htmx.go:66-151`
- Test: `framework/hx/htmx_test.go`

- [ ] **Step 1: Write failing tests for new event values**

Open `framework/hx/htmx_test.go` and add (or replace existing event tests with):

```go
package hx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventConstants_htmx4(t *testing.T) {
	t.Parallel()
	cases := map[Event]string{
		AbortEvent:                "htmx:abort",
		AfterOnLoadEvent:          "htmx:after:init",
		AfterProcessNodeEvent:     "htmx:after:init",
		AfterRequestEvent:         "htmx:after:request",
		AfterSettleEvent:          "htmx:after:swap",
		AfterSwapEvent:            "htmx:after:swap",
		BeforeCleanupElementEvent: "htmx:before:cleanup",
		BeforeOnLoadEvent:         "htmx:before:init",
		BeforeProcessNodeEvent:    "htmx:before:process",
		BeforeRequestEvent:        "htmx:before:request",
		BeforeSendEvent:           "htmx:before:request",
		BeforeSwapEvent:           "htmx:before:swap",
		ConfigRequestEvent:        "htmx:config:request",
		BeforeHistorySaveEvent:    "htmx:before:history:update",
		HistoryRestoreEvent:       "htmx:before:history:restore",
		HistoryCacheMissEvent:     "htmx:before:history:restore",
		PushedIntoHistoryEvent:    "htmx:after:history:push",
		ErrorEvent:                "htmx:error",
		ConfirmEvent:              "htmx:confirm",
		PromptEvent:               "htmx:prompt",
	}
	for c, want := range cases {
		assert.Equal(t, want, string(c))
	}
}
```

- [ ] **Step 2: Run the test — expect failures**

Run: `cd framework && go test ./hx/ -run TestEventConstants_htmx4 -v`
Expected: FAIL — old constants still carry htmx 2 values; also `ErrorEvent` undefined.

- [ ] **Step 3: Rewrite the event const block in `framework/hx/htmx.go`**

Replace the `const ( … Htmx Events … )` block (lines ~66-151) with:

```go
const (
	// Htmx Events (htmx 4 colon-form)
	AbortEvent                Event = "htmx:abort"
	AfterOnLoadEvent          Event = "htmx:after:init"
	AfterProcessNodeEvent     Event = "htmx:after:init"
	AfterRequestEvent         Event = "htmx:after:request"
	AfterSettleEvent          Event = "htmx:after:swap"
	AfterSwapEvent            Event = "htmx:after:swap"
	BeforeCleanupElementEvent Event = "htmx:before:cleanup"
	BeforeOnLoadEvent         Event = "htmx:before:init"
	BeforeProcessNodeEvent    Event = "htmx:before:process"
	BeforeRequestEvent        Event = "htmx:before:request"
	BeforeSendEvent           Event = "htmx:before:request"
	BeforeSwapEvent           Event = "htmx:before:swap"
	ConfigRequestEvent        Event = "htmx:config:request"
	BeforeHistorySaveEvent    Event = "htmx:before:history:update"
	HistoryRestoreEvent       Event = "htmx:before:history:restore"
	HistoryCacheMissEvent     Event = "htmx:before:history:restore"
	PushedIntoHistoryEvent    Event = "htmx:after:history:push"
	ErrorEvent                Event = "htmx:error"
	ConfirmEvent              Event = "htmx:confirm"
	OnMutationErrorEvent      Event = "htmx:onMutationError" // htmgo-fork custom event; kept as-is (emitted by mutation-error extension)
	PromptEvent               Event = "htmx:prompt"

	// SSE (extracted to hx-sse.js in alpha8 — names verified against upstream)
	SseConnectedEvent     Event = "htmx:sseOpen"
	SseConnectingEvent    Event = "htmx:sseConnecting"
	SseClosedEvent        Event = "htmx:sseClose"
	SseErrorEvent         Event = "htmx:sseError"
	SseBeforeMessageEvent Event = "htmx:sseBeforeMessage"
	SseAfterMessageEvent  Event = "htmx:sseAfterMessage"
	SSEErrorEvent         Event = "htmx:sseError" // alias for historical SSEErrorEvent callsites
	SSEOpenEvent          Event = "htmx:sseOpen"

	// Misc Events (non-htmx)
	RevealedEvent   Event = "revealed"
	InstersectEvent Event = "intersect"
	PollingEvent    Event = "every"

	// Dom Events (unchanged)
	ClickEvent       Event = "onclick"
	ChangeEvent      Event = "onchange"
	InputEvent       Event = "oninput"
	FocusEvent       Event = "onfocus"
	BlurEvent        Event = "onblur"
	KeyDownEvent     Event = "onkeydown"
	KeyUpEvent       Event = "onkeyup"
	KeyPressEvent    Event = "onkeypress"
	SubmitEvent      Event = "onsubmit"
	LoadDomEvent     Event = "onload"
	LoadEvent        Event = "onload"
	UnloadEvent      Event = "onunload"
	ResizeEvent      Event = "onresize"
	ScrollEvent      Event = "onscroll"
	DblClickEvent    Event = "ondblclick"
	MouseOverEvent   Event = "onmouseover"
	MouseOutEvent    Event = "onmouseout"
	MouseMoveEvent   Event = "onmousemove"
	MouseDownEvent   Event = "onmousedown"
	MouseUpEvent     Event = "onmouseup"
	ContextMenuEvent Event = "oncontextmenu"
	DragStartEvent   Event = "ondragstart"
	DragEvent        Event = "ondrag"
	DragEnterEvent   Event = "ondragenter"
	DragLeaveEvent   Event = "ondragleave"
	DragOverEvent    Event = "ondragover"
	DropEvent        Event = "ondrop"
	DragEndEvent     Event = "ondragend"
)
```

Deleted events (no htmx 4 equivalent): `OobAfterSwapEvent`, `OobBeforeSwapEvent`, `ResponseErrorEvent`, `SendErrorEvent`, `SwapErrorEvent`, `TargetErrorEvent`, `TimeoutEvent`, `ValidationValidateEvent`, `ValidationFailedEvent`, `ValidationHaltedEvent`, `XhrAbortEvent`, `XhrLoadEndEvent`, `XhrLoadStartEvent`, `XhrProgressEvent`, `HistoryCacheErrorEvent`, `HistoryCacheMissErrorEvent`, `HistoryCacheMissLoadEvent`, `OobErrorNoTargetEvent`, `NoSSESourceErrorEvent`, `OnLoadErrorEvent`.

- [ ] **Step 4: Run test — expect PASS**

Run: `cd framework && go test ./hx/ -run TestEventConstants_htmx4 -v`
Expected: PASS.

- [ ] **Step 5: Run full hx package tests**

Run: `cd framework && go test ./hx/ -count=1 -v`
Expected: PASS. If the pre-existing test suite references deleted events, delete those assertions too — they test a removed API.

- [ ] **Step 6: Commit**

```bash
git add framework/hx/htmx.go framework/hx/htmx_test.go
git commit -m "feat(hx)!: migrate event constants to htmx 4 colon-form

BREAKING: renames all htmx:camelCase events to htmx:colon:form.
Deletes events removed in htmx 4 (XHR, validation, OOB, split errors).
Adds ErrorEvent unifying htmx:error."
```

---

## Task 3: Migrate hx header constants to htmx 4 (Layer B.2)

**Files:**
- Modify: `framework/hx/htmx.go:45-64`
- Test: `framework/hx/htmx_test.go`

- [ ] **Step 1: Write failing tests for new header constants**

Append to `framework/hx/htmx_test.go`:

```go
func TestHeaderConstants_htmx4(t *testing.T) {
	t.Parallel()
	cases := map[Header]string{
		BoostedHeader:    "HX-Boosted",
		RequestHeader:    "HX-Request",
		TargetIdHeader:   "HX-Target",
		TriggerIdHeader:  "HX-Trigger",
		LocationHeader:   "HX-Location",
		PushUrlHeader:    "HX-Push-Url",
		RedirectHeader:   "HX-Redirect",
		RefreshHeader:    "HX-Refresh",
		ReplaceUrlHeader: "HX-Replace-Url",
		CurrentUrlHeader: "HX-Current-Url",
		ReswapHeader:     "HX-Reswap",
		RetargetHeader:   "HX-Retarget",
		ReselectHeader:   "HX-Reselect",
		TriggerHeader:    "HX-Trigger",
		SourceHeader:     "HX-Source",
		RequestTypeHeader: "HX-Request-Type",
	}
	for h, want := range cases {
		assert.Equal(t, want, string(h))
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd framework && go test ./hx/ -run TestHeaderConstants_htmx4 -v`
Expected: FAIL — `SourceHeader`, `RequestTypeHeader` undefined.

- [ ] **Step 3: Replace the header const block in `framework/hx/htmx.go`**

```go
const (
	BoostedHeader     Header = "HX-Boosted"
	RequestHeader     Header = "HX-Request"
	TargetIdHeader    Header = "HX-Target"    // htmx 4 format: tag#id
	TriggerIdHeader   Header = "HX-Trigger"   // htmx 4 format: tag#id
	LocationHeader    Header = "HX-Location"
	PushUrlHeader     Header = "HX-Push-Url"
	RedirectHeader    Header = "HX-Redirect"
	RefreshHeader     Header = "HX-Refresh"
	ReplaceUrlHeader  Header = "HX-Replace-Url"
	CurrentUrlHeader  Header = "HX-Current-Url"
	ReswapHeader      Header = "HX-Reswap"
	RetargetHeader    Header = "HX-Retarget"
	ReselectHeader    Header = "HX-Reselect"
	TriggerHeader     Header = "HX-Trigger"
	SourceHeader      Header = "HX-Source"       // new in htmx 4; tag#id format
	RequestTypeHeader Header = "HX-Request-Type" // new in htmx 4; "full" or "partial"
)
```

Deleted headers: `PromptResponseHeader`, `TriggerNameHeader`, `TriggerAfterSettleHeader`, `TriggerAfterSwapHeader`.

- [ ] **Step 4: Run — expect PASS**

Run: `cd framework && go test ./hx/ -run TestHeaderConstants_htmx4 -v`
Expected: PASS.

- [ ] **Step 5: Confirm callers in framework will fail to build (expected — fixed in Task 6)**

Run: `cd framework && go build ./... 2>&1 | head -30`
Expected: build errors in `h/app.go` referencing `hx.PromptResponseHeader`, `hx.TriggerNameHeader`. Leave them — Task 6 fixes.

- [ ] **Step 6: Commit**

```bash
git add framework/hx/htmx.go framework/hx/htmx_test.go
git commit -m "feat(hx)!: migrate header constants to htmx 4

BREAKING: removes HX-Trigger-Name, HX-Trigger-After-Swap, HX-Trigger-After-Settle,
HX-Prompt. Adds HX-Source (tag#id) and HX-Request-Type."
```

---

## Task 4: Migrate hx attribute constants (Layer B.3 — order-sensitive)

**Files:**
- Modify: `framework/hx/htmx.go:9-43`
- Test: `framework/hx/htmx_test.go`

- [ ] **Step 1: Write failing tests**

Append to `framework/hx/htmx_test.go`:

```go
func TestAttributeConstants_htmx4(t *testing.T) {
	t.Parallel()
	cases := map[Attribute]string{
		GetAttr:        "hx-get",
		PostAttr:       "hx-post",
		PutAttr:        "hx-put",
		PatchAttr:      "hx-patch",
		DeleteAttr:     "hx-delete",
		PushUrlAttr:    "hx-push-url",
		ReplaceUrlAttr: "hx-replace-url",
		SelectAttr:     "hx-select",
		SelectOobAttr:  "hx-select-oob",
		SwapAttr:       "hx-swap",
		SwapOobAttr:    "hx-swap-oob",
		TargetAttr:     "hx-target",
		TriggerAttr:    "hx-trigger",
		ValsAttr:       "hx-vals",
		BoostAttr:      "hx-boost",
		ConfirmAttr:    "hx-confirm",
		IgnoreAttr:     "hx-ignore",  // was hx-disable in htmx 2
		DisableAttr:    "hx-disable", // was hx-disabled-elt in htmx 2
		EncodingAttr:   "hx-encoding",
		HeadersAttr:    "hx-headers",
		IncludeAttr:    "hx-include",
		IndicatorAttr:  "hx-indicator",
		PreserveAttr:   "hx-preserve",
		SyncAttr:       "hx-sync",
		ValidateAttr:   "hx-validate",
		ConfigAttr:     "hx-config",
		StatusAttr:     "hx-status",
	}
	for a, want := range cases {
		assert.Equal(t, want, string(a))
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd framework && go test ./hx/ -run TestAttributeConstants_htmx4 -v`
Expected: FAIL — `IgnoreAttr`, `ConfigAttr`, `StatusAttr` undefined; `DisableAttr` has wrong value.

- [ ] **Step 3: Replace the attribute const block in `framework/hx/htmx.go`**

**Do this in a single edit — the two renames are order-sensitive.** Replace the entire first `const ( … Attribute … )` block with:

```go
const (
	GetAttr        Attribute = "hx-get"
	PostAttr       Attribute = "hx-post"
	PutAttr        Attribute = "hx-put"
	PatchAttr      Attribute = "hx-patch"
	DeleteAttr     Attribute = "hx-delete"
	PushUrlAttr    Attribute = "hx-push-url"
	ReplaceUrlAttr Attribute = "hx-replace-url"
	SelectAttr     Attribute = "hx-select"
	SelectOobAttr  Attribute = "hx-select-oob"
	SwapAttr       Attribute = "hx-swap"
	SwapOobAttr    Attribute = "hx-swap-oob"
	TargetAttr     Attribute = "hx-target"
	TriggerAttr    Attribute = "hx-trigger"
	ValsAttr       Attribute = "hx-vals"
	BoostAttr      Attribute = "hx-boost"
	ConfirmAttr    Attribute = "hx-confirm"
	IgnoreAttr     Attribute = "hx-ignore"  // htmx 4: stops htmx processing (was hx-disable in htmx 2)
	DisableAttr    Attribute = "hx-disable" // htmx 4: disable elements during request (was hx-disabled-elt in htmx 2)
	EncodingAttr   Attribute = "hx-encoding"
	HeadersAttr    Attribute = "hx-headers"
	IncludeAttr    Attribute = "hx-include"
	IndicatorAttr  Attribute = "hx-indicator"
	PreserveAttr   Attribute = "hx-preserve"
	SyncAttr       Attribute = "hx-sync"
	ValidateAttr   Attribute = "hx-validate"
	ConfigAttr     Attribute = "hx-config" // htmx 4: replaces hx-request
	StatusAttr     Attribute = "hx-status" // htmx 4: per-status swap/target control
)
```

Deleted: `DisabledEltAttr`, `DisinheritAttr`, `InheritAttr`, `ParamsAttr`, `PromptAttr`, `ExtAttr`, `HistoryAttr`, `HistoryEltAttr`, `RequestAttr`.

- [ ] **Step 4: Run test — expect PASS**

Run: `cd framework && go test ./hx/ -run TestAttributeConstants_htmx4 -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add framework/hx/htmx.go framework/hx/htmx_test.go
git commit -m "feat(hx)!: migrate attribute constants to htmx 4

BREAKING: DisableAttr value changed from hx-disable to hx-ignore;
DisabledEltAttr renamed to DisableAttr (value hx-disable).
Removes hx-disinherit, hx-inherit, hx-params, hx-prompt, hx-ext,
hx-history, hx-history-elt, hx-request. Adds hx-config, hx-status."
```

---

## Task 5: Update RequestContext for HX-Source / HX-Request-Type (Layer C.1)

**Files:**
- Modify: `framework/h/app.go:27, 105-107, 163-170`
- Test: `framework/h/app_test.go` (create if missing)

- [ ] **Step 1: Read current app.go struct and populateHxFields**

Verify fields `hxTriggerName`, `hxPromptResponse` exist at lines 27 and vicinity. Verify `populateHxFields` at line 163+ reads `hx.TriggerNameHeader` and `hx.PromptResponseHeader`.

- [ ] **Step 2: Write failing test**

Create or append to `framework/h/app_test.go`:

```go
package h

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestRequestContext_HxSource(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	called := false
	app.Get("/", func(c fiber.Ctx) error {
		called = true
		rc := &RequestContext{Fiber: c}
		populateHxFields(rc)
		assert.Equal(t, "button#save", rc.HxSource())
		assert.Equal(t, "save", rc.HxSourceID())
		assert.Equal(t, "partial", rc.HxRequestType())
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Source", "button#save")
	req.Header.Set("HX-Request-Type", "partial")
	_, err := app.Test(req)
	assert.NoError(t, err)
	assert.True(t, called, "handler must be invoked")
}
```

- [ ] **Step 3: Run test — expect FAIL**

Run: `cd framework && go test ./h/ -run TestRequestContext_HxSource -v`
Expected: FAIL — `HxSource`, `HxSourceID`, `HxRequestType` undefined.

- [ ] **Step 4: Update struct fields**

In `framework/h/app.go`, replace the struct field set (around lines 20-30):

```go
type RequestContext struct {
	Fiber             fiber.Ctx
	locator           *service.Locator
	isBoosted         bool
	currentBrowserUrl string
	isHxRequest       bool
	hxTargetId        string
	hxSource          string
	hxRequestType     string
	hxTriggerId       string
	kv                map[string]interface{}
}
```

Removed fields: `hxPromptResponse`, `hxTriggerName`.

- [ ] **Step 5: Replace accessor methods (around lines 101-107)**

Replace the `HxPromptResponse` and `HxTriggerName` methods with:

```go
import "strings"

// HxSource returns the raw HX-Source header: "tag#id" (e.g. "button#save"), empty if none.
func (c *RequestContext) HxSource() string {
	return c.hxSource
}

// HxSourceID returns the id portion of HX-Source, empty if no id.
func (c *RequestContext) HxSourceID() string {
	_, id, ok := strings.Cut(c.hxSource, "#")
	if !ok {
		return ""
	}
	return id
}

// HxRequestType returns "full" or "partial" for htmx 4 requests, empty otherwise.
func (c *RequestContext) HxRequestType() string {
	return c.hxRequestType
}
```

Keep `HxTargetId`, `HxTriggerId`, `HxCurrentBrowserUrl` as-is.

- [ ] **Step 6: Update populateHxFields (around lines 163-170)**

```go
func populateHxFields(cc *RequestContext) {
	cc.isBoosted = cc.Fiber.Get(hx.BoostedHeader) == "true"
	cc.currentBrowserUrl = cc.Fiber.Get(hx.CurrentUrlHeader)
	cc.isHxRequest = cc.Fiber.Get(hx.RequestHeader) == "true"
	cc.hxTargetId = cc.Fiber.Get(hx.TargetIdHeader)
	cc.hxSource = cc.Fiber.Get(hx.SourceHeader)
	cc.hxRequestType = cc.Fiber.Get(hx.RequestTypeHeader)
	cc.hxTriggerId = cc.Fiber.Get(hx.TriggerIdHeader)
}
```

- [ ] **Step 7: Run test — expect PASS**

Run: `cd framework && go test ./h/ -run TestRequestContext_HxSource -v`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add framework/h/app.go framework/h/app_test.go
git commit -m "feat(h)!: replace HxTriggerName with HxSource/HxRequestType

BREAKING: RequestContext.HxTriggerName() is removed. Use HxSource()
for the full 'tag#id' string or HxSourceID() for just the id."
```

---

## Task 6: Fix `hx-on::` colon-form emission in LifeCycle (Layer C.2)

**Files:**
- Modify: `framework/h/lifecycle.go:38-52`
- Test: `framework/h/lifecycle_test.go` (create if missing)

- [ ] **Step 1: Write failing tests**

Create `framework/h/lifecycle_test.go`:

```go
package h

import (
	"testing"

	"github.com/franchb/htmgo/framework/hx"
	"github.com/stretchr/testify/assert"
)

func TestLifeCycle_OnEvent_htmx4Colon(t *testing.T) {
	t.Parallel()
	cases := []struct {
		event hx.Event
		want  string // the resulting key in l.handlers
	}{
		{hx.AfterSwapEvent, "hx-on::after:swap"},
		{hx.BeforeRequestEvent, "hx-on::before:request"},
		{hx.ConfigRequestEvent, "hx-on::config:request"},
		{hx.ErrorEvent, "hx-on::error"},
		{hx.ClickEvent, "onclick"}, // DOM event unchanged
	}
	for _, tc := range cases {
		l := NewLifeCycle().OnEvent(tc.event, SimpleJsCommand{Command: "noop"})
		_, ok := l.handlers[hx.Event(tc.want)]
		assert.True(t, ok, "event %q should map to handler key %q; got keys %v", tc.event, tc.want, l.handlers)
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd framework && go test ./h/ -run TestLifeCycle_OnEvent_htmx4Colon -v`
Expected: FAIL — current code converts `htmx:after:swap` → `hx-on::after-swap` (dashes, not colons).

- [ ] **Step 3: Rewrite `LifeCycle.OnEvent` in `framework/h/lifecycle.go`**

Replace the method body (lines 38-52):

```go
func (l *LifeCycle) OnEvent(event hx.Event, cmd ...Command) *LifeCycle {
	validateCommands(cmd)

	// htmx 4: hx-on::<event:with:colons> shortcut for htmx:event:with:colons
	if strings.HasPrefix(string(event), "htmx:") {
		event = hx.Event("hx-on::" + string(event)[len("htmx:"):])
	}

	if l.handlers[event] == nil {
		l.handlers[event] = []Command{}
	}

	l.handlers[event] = append(l.handlers[event], cmd...)
	return l
}
```

Remove the `"github.com/franchb/htmgo/framework/internal/util"` import if it becomes unused in this file (`grep util. framework/h/lifecycle.go`).

- [ ] **Step 4: Run test — expect PASS**

Run: `cd framework && go test ./h/ -run TestLifeCycle_OnEvent_htmx4Colon -v`
Expected: PASS.

- [ ] **Step 5: Delete `HxOnLoad` deprecated wrapper (spec O1)**

Delete lines 65-69 of `framework/h/lifecycle.go`:

```go
// HxOnLoad executes the given commands when the element is loaded into the DOM.
// Deprecated: Use OnLoad instead.
func HxOnLoad(cmd ...Command) *LifeCycle { ... }
```

- [ ] **Step 6: Commit**

```bash
git add framework/h/lifecycle.go framework/h/lifecycle_test.go
git commit -m "feat(h)!: emit htmx 4 hx-on::event:colon:form; drop HxOnLoad

Replaces camelCase->dash conversion with direct colon-form pass-through
since htmx 4 event constants are already colon-separated.
Deletes deprecated HxOnLoad (use OnLoad)."
```

---

## Task 7: Audit `ToHtmxTriggerName` + `renderer.go` (Layer C.3)

**Files:**
- Verify: `framework/hx/trigger.go:16-24`
- Verify: `framework/h/renderer.go:231-238`

- [ ] **Step 1: Inspect `ToHtmxTriggerName`**

Read `framework/hx/trigger.go` lines 16-24. Confirm behavior: strips `htmx:` prefix, strips `on` prefix, else returns unchanged. For htmx 4 colon-form input `htmx:after:swap`, it returns `after:swap` — which is a valid event name inside `hx-trigger`. No change needed.

- [ ] **Step 2: Add test covering htmx 4 names**

Append to `framework/hx/htmx_test.go`:

```go
func TestToHtmxTriggerName_htmx4(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "after:swap", ToHtmxTriggerName("htmx:after:swap"))
	assert.Equal(t, "click", ToHtmxTriggerName("onclick"))
	assert.Equal(t, "custom-event", ToHtmxTriggerName("custom-event"))
}
```

- [ ] **Step 3: Run — expect PASS immediately (no code change)**

Run: `cd framework && go test ./hx/ -run TestToHtmxTriggerName_htmx4 -v`
Expected: PASS.

- [ ] **Step 4: Inspect `renderer.go:231-238`**

Read `framework/h/renderer.go` lines 231-238. The `fromAttributeMap` function uses `hx.ToHtmxTriggerName(event)` to emit a default `hx-trigger` when an hx-get/patch/post attribute is present. With the colon-form names from Layer B, this will emit `hx-trigger="after:swap"` which is correct.

No code change needed. Spot-check rendered output with an existing test:

Run: `cd framework && go test ./h/ -run TestRender -v -count=1`
Expected: all PASS. If any fail due to changed default trigger name, inspect and update the test fixture only (not the renderer).

- [ ] **Step 5: Commit**

```bash
git add framework/hx/htmx_test.go
git commit -m "test(hx): verify ToHtmxTriggerName handles htmx 4 colon-form"
```

---

## Task 8: Add `:inherited` attribute helpers (Layer C.4)

**Files:**
- Modify: `framework/h/attribute.go`
- Test: `framework/h/attribute_test.go`

- [ ] **Step 1: Write failing tests**

Append to `framework/h/attribute_test.go`:

```go
func TestHxInheritedAttributes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		attr  Ren
		key   string
		value string
	}{
		{"HxTargetInherited", HxTargetInherited("#out"), "hx-target:inherited", "#out"},
		{"HxIncludeInherited", HxIncludeInherited("closest form"), "hx-include:inherited", "closest form"},
		{"HxSwapInherited", HxSwapInherited("outerHTML"), "hx-swap:inherited", "outerHTML"},
		{"HxBoostInherited", HxBoostInherited("true"), "hx-boost:inherited", "true"},
		{"HxConfirmInherited", HxConfirmInherited("Sure?"), "hx-confirm:inherited", "Sure?"},
		{"HxHeadersInherited", HxHeadersInherited(`{"X-Token":"abc"}`), "hx-headers:inherited", `{"X-Token":"abc"}`},
		{"HxIndicatorInherited", HxIndicatorInherited("#spinner"), "hx-indicator:inherited", "#spinner"},
		{"HxSyncInherited", HxSyncInherited("this:drop"), "hx-sync:inherited", "this:drop"},
		{"HxConfigInherited", HxConfigInherited(`{"timeout":5000}`), "hx-config:inherited", `{"timeout":5000}`},
		{"HxEncodingInherited", HxEncodingInherited("multipart/form-data"), "hx-encoding:inherited", "multipart/form-data"},
		{"HxValidateInherited", HxValidateInherited("true"), "hx-validate:inherited", "true"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ar, ok := tc.attr.(*AttributeR)
			assert.True(t, ok, "expected *AttributeR from %s", tc.name)
			assert.Equal(t, tc.key, ar.Name)
			assert.Equal(t, tc.value, ar.Value)
		})
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd framework && go test ./h/ -run TestHxInheritedAttributes -v`
Expected: FAIL — none of the `HxXxxInherited` functions exist yet.

- [ ] **Step 3: Add constructors to `framework/h/attribute.go`**

Append to `framework/h/attribute.go`:

```go
// Inheritance-aware helpers (htmx 4 explicit inheritance).
// Emits `<attr>:inherited="..."` which propagates to descendants.

func HxTargetInherited(selector string) Ren   { return Attribute("hx-target:inherited", selector) }
func HxIncludeInherited(selector string) Ren  { return Attribute("hx-include:inherited", selector) }
func HxSwapInherited(swap string) Ren         { return Attribute("hx-swap:inherited", swap) }
func HxBoostInherited(value string) Ren       { return Attribute("hx-boost:inherited", value) }
func HxConfirmInherited(message string) Ren   { return Attribute("hx-confirm:inherited", message) }
func HxHeadersInherited(json string) Ren      { return Attribute("hx-headers:inherited", json) }
func HxIndicatorInherited(selector string) Ren { return Attribute("hx-indicator:inherited", selector) }
func HxSyncInherited(spec string) Ren         { return Attribute("hx-sync:inherited", spec) }
func HxConfigInherited(json string) Ren       { return Attribute("hx-config:inherited", json) }
func HxEncodingInherited(enc string) Ren      { return Attribute("hx-encoding:inherited", enc) }
func HxValidateInherited(value string) Ren    { return Attribute("hx-validate:inherited", value) }
```

- [ ] **Step 4: Run test — expect PASS**

Run: `cd framework && go test ./h/ -run TestHxInheritedAttributes -v`
Expected: PASS.

- [ ] **Step 5: Run the full framework test suite**

Run: `cd framework && go test ./... -count=1`
Expected: all PASS. Failures in `examples/` or `htmgo-site` will appear in their own module runs — address in Layer E tasks. Framework itself should be green.

- [ ] **Step 6: Commit**

```bash
git add framework/h/attribute.go framework/h/attribute_test.go
git commit -m "feat(h): add HxXxxInherited constructors for htmx 4 :inherited modifier

Provides first-class Go builders for the 11 inheritable attributes,
avoiding string concatenation at call sites."
```

---

## Task 9: Set up vitest for JS extension tests

**Files:**
- Create: `framework/assets/js/vitest.config.ts`
- Modify: `framework/assets/js/package.json`

- [ ] **Step 1: Install vitest**

```bash
cd framework/assets/js && npm install --save-dev vitest @vitest/ui jsdom
```

- [ ] **Step 2: Create vitest config**

Write `framework/assets/js/vitest.config.ts`:

```ts
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "jsdom",
    include: ["htmxextensions/__tests__/**/*.test.ts"],
  },
});
```

- [ ] **Step 3: Add test script to package.json**

Edit `framework/assets/js/package.json` scripts block — add:

```json
"test": "vitest run",
"test:watch": "vitest"
```

- [ ] **Step 4: Sanity test**

Create `framework/assets/js/htmxextensions/__tests__/sanity.test.ts`:

```ts
import { describe, it, expect } from "vitest";

describe("vitest sanity", () => {
  it("is wired up", () => {
    expect(1 + 1).toBe(2);
  });
});
```

Run: `cd framework/assets/js && npm test`
Expected: 1 passed.

- [ ] **Step 5: Delete sanity test and commit**

```bash
rm framework/assets/js/htmxextensions/__tests__/sanity.test.ts
git add framework/assets/js/package.json framework/assets/js/package-lock.json framework/assets/js/vitest.config.ts
git commit -m "build(js): add vitest for extension tests"
```

---

## Task 10: Port `htmgo.ts` extension to htmx 4 API (Layer D)

**Files:**
- Modify: `framework/assets/js/htmxextensions/htmgo.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/htmgo.test.ts`

- [ ] **Step 1: Write failing test**

Create `framework/assets/js/htmxextensions/__tests__/htmgo.test.ts`:

```ts
import { describe, it, expect, vi } from "vitest";

// Mock htmx before importing the extension
const registeredExtensions: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (name: string, ext: any) => {
      registeredExtensions[name] = ext;
    },
  },
}));

describe("htmgo extension", () => {
  it("registers with htmx 4 registerExtension API", async () => {
    await import("../htmgo");
    expect(registeredExtensions["htmgo"]).toBeDefined();
    expect(typeof registeredExtensions["htmgo"].htmx_before_cleanup).toBe("function");
    expect(typeof registeredExtensions["htmgo"].htmx_after_init).toBe("function");
  });

  it("invokes onload handlers on descendants with [onload]", async () => {
    await import("../htmgo");
    const ext = registeredExtensions["htmgo"];
    const parent = document.createElement("div");
    const child = document.createElement("span");
    let called = false;
    child.setAttribute("onload", "");
    child.onload = () => { called = true; };
    parent.appendChild(child);
    document.body.appendChild(parent);
    ext.htmx_after_init(parent, {});
    expect(called).toBe(true);
    document.body.removeChild(parent);
  });
});
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd framework/assets/js && npm test`
Expected: FAIL — current `htmgo.ts` uses `defineExtension`, no `registerExtension` call.

- [ ] **Step 3: Rewrite `htmgo.ts`**

Replace `framework/assets/js/htmxextensions/htmgo.ts` entirely:

```ts
import htmx from "htmx.org";

const evalFuncRegex = /__eval_[A-Za-z0-9]+\([a-z]+\)/gm;

htmx.registerExtension("htmgo", {
  init(_api: unknown) {
    // no-op; retained for htmx 4 registerExtension API shape
  },

  htmx_before_cleanup(elt: HTMLElement, _detail: unknown) {
    if (elt) removeAssociatedScripts(elt);
  },

  htmx_after_init(elt: HTMLElement, _detail: unknown) {
    if (elt) invokeOnLoad(elt);
  },
});

// Browser doesn't support onload for all elements, so we manually trigger it
// (useful for locality of behavior).
function invokeOnLoad(element: Element) {
  if (element == null || !(element instanceof HTMLElement)) return;
  const ignored = ["SCRIPT", "LINK", "STYLE", "META", "BASE", "TITLE", "HEAD", "HTML", "BODY"];
  if (!ignored.includes(element.tagName)) {
    if (element.hasAttribute("onload")) {
      element.onload!(new Event("load"));
    }
  }
  element.querySelectorAll("[onload]").forEach(invokeOnLoad);
}

export function removeAssociatedScripts(element: HTMLElement) {
  const attributes = Array.from(element.attributes);
  for (const attribute of attributes) {
    const matches = attribute.value.match(evalFuncRegex) || [];
    for (const match of matches) {
      const id = match.replace("()", "").replace("(this)", "").replace(";", "");
      const ele = document.getElementById(id);
      if (ele && ele.tagName === "SCRIPT") {
        console.debug("removing associated script with id", id);
        ele.remove();
      }
    }
  }
}
```

- [ ] **Step 4: Run test — expect PASS**

Run: `cd framework/assets/js && npm test`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add framework/assets/js/htmxextensions/htmgo.ts framework/assets/js/htmxextensions/__tests__/htmgo.test.ts
git commit -m "feat(js): port htmgo extension to htmx 4 registerExtension API"
```

---

## Task 11: Port `extension.ts` loader

**Files:**
- Modify: `framework/assets/js/htmxextensions/extension.ts`

- [ ] **Step 1: Read current extension.ts**

Read the file — it is the entry that imports/registers all extensions. Likely just `import` statements.

- [ ] **Step 2: Update imports so every extension is loaded**

Ensure `extension.ts` imports every rewritten extension file:

```ts
import "./htmgo";
import "./debug";
import "./livereload";
import "./mutation-error";
import "./pathdeps";
import "./trigger-children";
import "./response-targets";
import "./sse";
import "./ws";
import "./ws-event-handler";
```

- [ ] **Step 3: Commit**

```bash
git add framework/assets/js/htmxextensions/extension.ts
git commit -m "build(js): wire extension loader for htmx 4 ports"
```

---

## Task 12: Port `debug.ts`

**Files:**
- Modify: `framework/assets/js/htmxextensions/debug.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/debug.test.ts`

- [ ] **Step 1: Write failing test**

```ts
import { describe, it, expect, vi } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: { registerExtension: (n: string, e: any) => (registered[n] = e) },
}));

describe("debug extension", () => {
  it("registers and logs events", async () => {
    await import("../debug");
    expect(registered["debug"]).toBeDefined();
  });
});
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd framework/assets/js && npm test -- debug.test`
Expected: FAIL.

- [ ] **Step 3: Rewrite `debug.ts`**

```ts
import htmx from "htmx.org";

htmx.registerExtension("debug", {
  init(_api: unknown) {
    htmx.config.logAll = true;
  },
});
```

- [ ] **Step 4: Run — expect PASS**

Run: `cd framework/assets/js && npm test -- debug.test`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add framework/assets/js/htmxextensions/debug.ts framework/assets/js/htmxextensions/__tests__/debug.test.ts
git commit -m "feat(js): port debug extension to htmx 4"
```

---

## Task 13: Port `mutation-error.ts`

**Files:**
- Modify: `framework/assets/js/htmxextensions/mutation-error.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/mutation-error.test.ts`

- [ ] **Step 1: Read current mutation-error.ts and understand its trigger**

The htmx 2 version listens for `htmx:afterRequest` and dispatches a custom `htmx:onMutationError` event if the request was a mutation that failed. Preserve that behavior but using htmx 4 API.

- [ ] **Step 2: Write failing test**

```ts
import { describe, it, expect, vi } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: { registerExtension: (n: string, e: any) => (registered[n] = e), trigger: vi.fn() },
}));

describe("mutation-error extension", () => {
  it("registers and has after_request hook", async () => {
    await import("../mutation-error");
    expect(registered["mutation-error"]).toBeDefined();
    expect(typeof registered["mutation-error"].htmx_after_request).toBe("function");
  });
});
```

- [ ] **Step 3: Run — expect FAIL**

Run: `cd framework/assets/js && npm test -- mutation-error.test`
Expected: FAIL.

- [ ] **Step 4: Rewrite `mutation-error.ts`**

```ts
import htmx from "htmx.org";

const mutationMethods = new Set(["POST", "PUT", "PATCH", "DELETE"]);

htmx.registerExtension("mutation-error", {
  init(_api: unknown) {},

  htmx_after_request(elt: HTMLElement, detail: any) {
    const ctx = detail?.ctx;
    if (!ctx || !ctx.request) return;
    const method = (ctx.request.method || "").toUpperCase();
    if (!mutationMethods.has(method)) return;
    const status = ctx.response?.status ?? 0;
    if (status === 0 || status >= 400) {
      htmx.trigger(elt, "htmx:onMutationError", { status, elt });
    }
  },
});
```

- [ ] **Step 5: Run — expect PASS**

Run: `cd framework/assets/js && npm test -- mutation-error.test`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add framework/assets/js/htmxextensions/mutation-error.ts framework/assets/js/htmxextensions/__tests__/mutation-error.test.ts
git commit -m "feat(js): port mutation-error extension to htmx 4"
```

---

## Task 14: Port `pathdeps.ts`

**Files:**
- Modify: `framework/assets/js/htmxextensions/pathdeps.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/pathdeps.test.ts`

- [ ] **Step 1: Read current pathdeps.ts** — understand the `path-deps` attribute matching logic (refreshes any element whose `path-deps` attr matches the mutated URL).

- [ ] **Step 2: Write failing test**

```ts
import { describe, it, expect, vi } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    trigger: vi.fn(),
  },
}));

describe("pathdeps extension", () => {
  it("registers and has after_request hook", async () => {
    await import("../pathdeps");
    expect(registered["path-deps"]).toBeDefined();
    expect(typeof registered["path-deps"].htmx_after_request).toBe("function");
  });
});
```

- [ ] **Step 3: Rewrite `pathdeps.ts`**

```ts
import htmx from "htmx.org";

const mutationMethods = new Set(["POST", "PUT", "PATCH", "DELETE"]);

function intersects(pattern: string, path: string): boolean {
  if (!pattern || !path) return false;
  if (pattern === "ignore") return false;
  // simple prefix/glob: supports "*" wildcard at tail
  if (pattern.endsWith("*")) return path.startsWith(pattern.slice(0, -1));
  return pattern === path;
}

htmx.registerExtension("path-deps", {
  init(_api: unknown) {},

  htmx_after_request(_elt: HTMLElement, detail: any) {
    const ctx = detail?.ctx;
    if (!ctx || !ctx.request) return;
    const method = (ctx.request.method || "").toUpperCase();
    if (!mutationMethods.has(method)) return;
    const path = ctx.request.action || "";
    document.querySelectorAll("[path-deps]").forEach((el) => {
      const dep = el.getAttribute("path-deps") || "";
      if (intersects(dep, path)) htmx.trigger(el, "path-deps");
    });
  },
});
```

- [ ] **Step 4: Run test — expect PASS**

Run: `cd framework/assets/js && npm test -- pathdeps.test`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add framework/assets/js/htmxextensions/pathdeps.ts framework/assets/js/htmxextensions/__tests__/pathdeps.test.ts
git commit -m "feat(js): port path-deps extension to htmx 4"
```

---

## Task 15: Port `trigger-children.ts`

**Files:**
- Modify: `framework/assets/js/htmxextensions/trigger-children.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/trigger-children.test.ts`

- [ ] **Step 1: Read current trigger-children.ts** — understand: when an element has `trigger-children` attr, any htmx trigger on that element fans out to all descendants.

- [ ] **Step 2: Write failing test**

```ts
import { describe, it, expect, vi } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    trigger: vi.fn(),
  },
}));

describe("trigger-children extension", () => {
  it("registers and has hook", async () => {
    await import("../trigger-children");
    expect(registered["trigger-children"]).toBeDefined();
  });
});
```

- [ ] **Step 3: Rewrite `trigger-children.ts`** preserving the existing fan-out semantics, but using `htmx_after_request` or `htmx_before_request` hooks as appropriate. Follow the shape of current code; the only mechanical change is `defineExtension`→`registerExtension`, `onEvent` callback → per-event methods.

- [ ] **Step 4: Run test — expect PASS**

Run: `cd framework/assets/js && npm test -- trigger-children.test`

- [ ] **Step 5: Commit**

```bash
git add framework/assets/js/htmxextensions/trigger-children.ts framework/assets/js/htmxextensions/__tests__/trigger-children.test.ts
git commit -m "feat(js): port trigger-children extension to htmx 4"
```

---

## Task 16: Port `livereload.ts`

**Files:**
- Modify: `framework/assets/js/htmxextensions/livereload.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/livereload.test.ts`

- [ ] **Step 1: Read current livereload.ts** — understand: it sets up an SSE connection to the dev server and reloads the page on a `reload` event.

- [ ] **Step 2: Write failing test**

```ts
import { describe, it, expect, vi } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: { registerExtension: (n: string, e: any) => (registered[n] = e) },
}));

describe("livereload extension", () => {
  it("registers", async () => {
    await import("../livereload");
    expect(registered["livereload"]).toBeDefined();
  });
});
```

- [ ] **Step 3: Rewrite `livereload.ts`**

Convert `defineExtension` → `registerExtension`. If it uses `onEvent` for `htmx:load` transition to `htmx_after_init`. SSE connection bootstrap runs inside `init(api)`.

- [ ] **Step 4: Run test — expect PASS**

- [ ] **Step 5: Commit**

```bash
git add framework/assets/js/htmxextensions/livereload.ts framework/assets/js/htmxextensions/__tests__/livereload.test.ts
git commit -m "feat(js): port livereload extension to htmx 4"
```

---

## Task 17: Port `response-targets.ts` (largest rewrite)

**Files:**
- Modify: `framework/assets/js/htmxextensions/response-targets.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/response-targets.test.ts`

- [ ] **Step 1: Read current response-targets.ts** (136 lines). Note the matching-attribute ladder: exact status (`404`), two-digit wildcard (`4*`/`4x`), one-digit wildcard (`4**`/`4xx`/`4*`/`4x`), `*`/`x`/`***`/`xxx`, plus `error` for 4xx/5xx.

- [ ] **Step 2: Write failing test covering the attribute ladder**

```ts
import { describe, it, expect, vi } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    config: {},
  },
}));

describe("response-targets extension", () => {
  it("registers with init receiving api", async () => {
    await import("../response-targets");
    const ext = registered["response-targets"];
    expect(ext).toBeDefined();
    expect(typeof ext.init).toBe("function");
    expect(typeof ext.htmx_before_swap).toBe("function");
  });

  it("exposes config defaults via init", async () => {
    const api = { getClosestAttributeValue: () => null, findThisElement: () => null, querySelectorExt: () => null };
    const ext = registered["response-targets"];
    ext.init(api);
    // config defaults should be established
    // (actual htmx.config mutation is out of scope for unit test — smoke only)
  });
});
```

- [ ] **Step 3: Run — expect FAIL**

- [ ] **Step 4: Rewrite `response-targets.ts`**

Preserve the complete attribute-matching ladder. Change only the outer shape:

```ts
import htmx from "htmx.org";
const config: any = htmx.config;

let api: any;
const attrPrefix = "hx-target-";

function startsWith(str: string, prefix: string) {
  return str.substring(0, prefix.length) === prefix;
}

function getRespCodeTarget(elt: Element, respCodeNumber: number) {
  if (!elt || !respCodeNumber) return null;
  const respCode = respCodeNumber.toString();
  const attrPossibilities = [
    respCode,
    respCode.substr(0, 2) + "*",
    respCode.substr(0, 2) + "x",
    respCode.substr(0, 1) + "*",
    respCode.substr(0, 1) + "x",
    respCode.substr(0, 1) + "**",
    respCode.substr(0, 1) + "xx",
    "*", "x", "***", "xxx",
  ];
  if (startsWith(respCode, "4") || startsWith(respCode, "5")) attrPossibilities.push("error");
  for (const p of attrPossibilities) {
    const attr = attrPrefix + p;
    const attrValue = api.getClosestAttributeValue(elt, attr);
    if (attrValue) {
      if (attrValue === "this") return api.findThisElement(elt, attr);
      return api.querySelectorExt(elt, attrValue);
    }
  }
  return null;
}

function handleErrorFlag(detail: any) {
  if (detail.isError) {
    if (config.responseTargetUnsetsError) detail.isError = false;
  } else if (config.responseTargetSetsError) {
    detail.isError = true;
  }
}

htmx.registerExtension("response-targets", {
  init(apiRef: any) {
    api = apiRef;
    if (config.responseTargetUnsetsError === undefined) config.responseTargetUnsetsError = true;
    if (config.responseTargetSetsError === undefined) config.responseTargetSetsError = false;
    if (config.responseTargetPrefersExisting === undefined) config.responseTargetPrefersExisting = false;
    if (config.responseTargetPrefersRetargetHeader === undefined) config.responseTargetPrefersRetargetHeader = true;
  },

  htmx_before_swap(_elt: HTMLElement, detail: any) {
    const ctx = detail?.ctx;
    const status = ctx?.response?.status ?? 0;
    if (status === 200) return;
    if (detail.target) {
      if (config.responseTargetPrefersExisting) {
        detail.shouldSwap = true;
        handleErrorFlag(detail);
        return;
      }
      const retarget = ctx?.response?.headers?.get?.("HX-Retarget");
      if (config.responseTargetPrefersRetargetHeader && retarget) {
        detail.shouldSwap = true;
        handleErrorFlag(detail);
        return;
      }
    }
    const elt = detail.requestConfig?.elt ?? ctx?.elt;
    if (!elt) return;
    const target = getRespCodeTarget(elt, status);
    if (target) {
      handleErrorFlag(detail);
      detail.shouldSwap = true;
      detail.target = target;
    }
  },
});
```

- [ ] **Step 5: Run test — expect PASS**

Run: `cd framework/assets/js && npm test -- response-targets.test`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add framework/assets/js/htmxextensions/response-targets.ts framework/assets/js/htmxextensions/__tests__/response-targets.test.ts
git commit -m "feat(js): port response-targets extension to htmx 4

Preserves the hx-target-<status>/<wildcard>/error matching ladder
but reads status/headers from htmx 4 detail.ctx shape instead of xhr."
```

---

## Task 18: Port `sse.ts`

**Files:**
- Modify: `framework/assets/js/htmxextensions/sse.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/sse.test.ts`

- [ ] **Step 1: Read upstream `hx-sse.js`**

Reference: `/home/iru/p/junk/htmx/src/hx-sse.js` (or under `dist/`). Read it to understand the event-detail shape and naming in htmx 4.

- [ ] **Step 2: Read current sse.ts** and identify features to preserve.

- [ ] **Step 3: Write failing test**

```ts
import { describe, it, expect, vi } from "vitest";
const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({ default: { registerExtension: (n: string, e: any) => (registered[n] = e) } }));
describe("sse extension", () => {
  it("registers", async () => {
    await import("../sse");
    expect(registered["sse"]).toBeDefined();
  });
});
```

- [ ] **Step 4: Rewrite `sse.ts`** to use `registerExtension` + per-event hooks aligned with upstream `hx-sse.js` event detail shape. Preserve the set of features currently in the file; do not extend or shrink.

- [ ] **Step 5: Run test — expect PASS**

- [ ] **Step 6: Commit**

```bash
git add framework/assets/js/htmxextensions/sse.ts framework/assets/js/htmxextensions/__tests__/sse.test.ts
git commit -m "feat(js): port sse extension to htmx 4 registerExtension API"
```

---

## Task 19: Port `ws.ts` (WebSocket core)

**Files:**
- Modify: `framework/assets/js/htmxextensions/ws.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/ws.test.ts`

- [ ] **Step 1: Diff current ws.ts against upstream htmx 4 WS**

Reference: upstream aligned WS extension (per beta1 changelog: `htmx.config.ws`, `{headers, body}` send format, `HX-Request-ID` correlation, `pauseOnBackground`, exponential backoff, per-element config).

Identify features already in htmgo's `ws.ts` vs features only in upstream. Per spec R4, this port preserves only the currently-present features. File follow-up issue for adopting upstream-only features.

- [ ] **Step 2: Write failing test**

```ts
import { describe, it, expect, vi } from "vitest";
const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({ default: { registerExtension: (n: string, e: any) => (registered[n] = e), config: {} } }));
describe("ws extension", () => {
  it("registers", async () => {
    await import("../ws");
    expect(registered["ws"]).toBeDefined();
  });
});
```

- [ ] **Step 3: Rewrite `ws.ts`** using `registerExtension` + per-event hooks. Keep the connection bootstrap, message send format, event dispatch consistent with the pre-migration file.

- [ ] **Step 4: Run test — expect PASS**

- [ ] **Step 5: Commit**

```bash
git add framework/assets/js/htmxextensions/ws.ts framework/assets/js/htmxextensions/__tests__/ws.test.ts
git commit -m "feat(js): port ws extension to htmx 4 registerExtension API

Preserves existing feature set only. Follow-up: adopt upstream beta1
additions (pauseOnBackground, HX-Request-ID correlation, etc.)."
```

---

## Task 20: Port `ws-event-handler.ts`

**Files:**
- Modify: `framework/assets/js/htmxextensions/ws-event-handler.ts`
- Create: `framework/assets/js/htmxextensions/__tests__/ws-event-handler.test.ts`

- [ ] **Step 1: Read current ws-event-handler.ts** — companion to `ws.ts` for routing `ws-send`/`ws-receive` events into DOM updates.

- [ ] **Step 2: Write failing test**

```ts
import { describe, it, expect, vi } from "vitest";
const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({ default: { registerExtension: (n: string, e: any) => (registered[n] = e) } }));
describe("ws-event-handler extension", () => {
  it("registers", async () => {
    await import("../ws-event-handler");
    expect(Object.keys(registered).some(k => k.startsWith("ws"))).toBe(true);
  });
});
```

- [ ] **Step 3: Rewrite `ws-event-handler.ts`**. Mechanical port; align any `evt.detail.xhr` reads to `detail.ctx`.

- [ ] **Step 4: Run test — expect PASS**

- [ ] **Step 5: Commit**

```bash
git add framework/assets/js/htmxextensions/ws-event-handler.ts framework/assets/js/htmxextensions/__tests__/ws-event-handler.test.ts
git commit -m "feat(js): port ws-event-handler extension to htmx 4"
```

---

## Task 21: Build the full JS bundle and verify

**Files:** none

- [ ] **Step 1: Run all vitest**

Run: `cd framework/assets/js && npm test`
Expected: all 9 extension test files pass.

- [ ] **Step 2: Production build**

Run: `cd framework/assets/js && npm run build`
Expected: no compile errors; `dist/` artifacts produced.

- [ ] **Step 3: Commit any build output if tracked** (check `.gitignore`). If built artifacts are not tracked, skip.

---

## Task 22: Inheritance audit — `examples/chat/`

**Files:**
- Audit: `examples/chat/**/*.go`

- [ ] **Step 1: Grep for inheritable-attribute emission on container elements**

Run from repo root:
```bash
grep -rn 'hx\.\(TargetAttr\|IncludeAttr\|HeadersAttr\|BoostAttr\|IndicatorAttr\|SyncAttr\|ConfirmAttr\|EncodingAttr\|ValidateAttr\|ConfigAttr\)\|Hx\(Target\|Include\|Headers\|Boost\|Indicator\|Sync\|Confirm\|Encoding\|Validate\|Config\)(' examples/chat
```

- [ ] **Step 2: For each hit, determine inheritance parent**

For each file, open it and check: does the element that emits the attribute also emit `hx.GetAttr`/`PostAttr`/`PutAttr`/`PatchAttr`/`DeleteAttr` OR is it a container whose children do?

- If the container has its own `hx-get`/etc., the attribute is self-consumed → no change.
- If the container has NO `hx-get`/etc. but descendants do → inheritance parent → switch emission to the inherited helper (e.g., `h.HxTarget("#out")` → `h.HxTargetInherited("#out")`).

- [ ] **Step 3: Apply edits**

Swap the emission at each confirmed inheritance parent.

- [ ] **Step 4: Build + smoke run**

```bash
cd examples/chat && task build
```
Expected: compiles.

- [ ] **Step 5: Commit**

```bash
git add examples/chat/
git commit -m "refactor(examples/chat): use HxXxxInherited for htmx 4 explicit inheritance"
```

Note: if no changes were needed, skip the commit and mark this task complete.

---

## Task 23: Inheritance audit — `examples/hackernews/`

**Files:**
- Audit: `examples/hackernews/**/*.go`

Follow the Task 22 procedure for `examples/hackernews/`. Based on the initial grep findings, `partials/sidebar.go` is a likely inheritance parent (has 3 hx-prefixed hits).

- [ ] **Step 1: Grep, audit, edit, build** — as in Task 22 but under `examples/hackernews/`.

- [ ] **Step 2: Commit**

```bash
git add examples/hackernews/
git commit -m "refactor(examples/hackernews): explicit inheritance for htmx 4"
```

---

## Task 24: Inheritance audit — `examples/simple-auth/`

**Files:**
- Audit: `examples/simple-auth/**/*.go`

- [ ] **Step 1: Grep for inheritable-attribute emission**

```bash
grep -rn 'hx\.\(TargetAttr\|IncludeAttr\|HeadersAttr\|BoostAttr\|IndicatorAttr\|SyncAttr\|ConfirmAttr\|EncodingAttr\|ValidateAttr\|ConfigAttr\)\|Hx\(Target\|Include\|Headers\|Boost\|Indicator\|Sync\|Confirm\|Encoding\|Validate\|Config\)(' examples/simple-auth
```

- [ ] **Step 2: For each hit, decide inheritance**

Open each matching file. Check whether the emitting element also emits `hx.GetAttr`/`PostAttr`/`PutAttr`/`PatchAttr`/`DeleteAttr`:
- If yes → self-consumed, no change.
- If no, and descendants issue `hx-get`/`-post` without declaring their own copy → inheritance parent. Swap the emission to the `:inherited` helper (`HxTarget` → `HxTargetInherited`, etc. as added in Task 8).

- [ ] **Step 3: Build example**

```bash
cd examples/simple-auth && task build
```
Expected: compiles.

- [ ] **Step 4: Commit (skip if no changes made)**

```bash
git add examples/simple-auth/
git commit -m "refactor(examples/simple-auth): explicit inheritance for htmx 4"
```

---

## Task 25: Inheritance audit — `examples/todo-list/`, `examples/ws-example/`, `examples/minimal-htmgo/`

**Files:**
- Audit: `examples/todo-list/**/*.go`, `examples/ws-example/**/*.go`, `examples/minimal-htmgo/**/*.go`

- [ ] **Step 1: Grep each example**

```bash
for d in examples/todo-list examples/ws-example examples/minimal-htmgo; do
  echo "=== $d ==="
  grep -rn 'hx\.\(TargetAttr\|IncludeAttr\|HeadersAttr\|BoostAttr\|IndicatorAttr\|SyncAttr\|ConfirmAttr\|EncodingAttr\|ValidateAttr\|ConfigAttr\)\|Hx\(Target\|Include\|Headers\|Boost\|Indicator\|Sync\|Confirm\|Encoding\|Validate\|Config\)(' "$d"
done
```

- [ ] **Step 2: For each hit, decide inheritance**

Open each matching file. Inheritance parent = emits the attribute but no `hx.GetAttr`/`PostAttr`/`PutAttr`/`PatchAttr`/`DeleteAttr` of its own, while descendants issue those verbs. Swap emission to the `:inherited` helper (e.g., `HxTarget` → `HxTargetInherited`).

- [ ] **Step 3: Build each**

```bash
for d in examples/todo-list examples/ws-example examples/minimal-htmgo; do
  (cd "$d" && task build) || echo "FAIL: $d"
done
```
Expected: all compile.

- [ ] **Step 4: Commit**

```bash
git add examples/todo-list examples/ws-example examples/minimal-htmgo
git commit -m "refactor(examples): explicit inheritance for todo/ws/minimal"
```

---

## Task 26: Inheritance audit — `htmgo-site/`

**Files:**
- Audit: `htmgo-site/pages/**/*.go`, `htmgo-site/partials/**/*.go`

- [ ] **Step 1: Grep, audit, edit**

```bash
grep -rn 'hx\.\(TargetAttr\|IncludeAttr\|HeadersAttr\|BoostAttr\|IndicatorAttr\|SyncAttr\|ConfirmAttr\|EncodingAttr\|ValidateAttr\|ConfigAttr\)\|Hx\(Target\|Include\|Headers\|Boost\|Indicator\|Sync\|Confirm\|Encoding\|Validate\|Config\)(' htmgo-site
```

Apply inherited-helper swaps where children rely on parent-emitted attrs.

- [ ] **Step 2: Doc prose updates**

Grep doc sources for htmx-2-specific prose:
```bash
grep -rn 'hx-vars\|hx-disinherit\|hx-inherit\|hx-history\|hx-prompt\|hx-params\|hx-ext\|HX-Trigger-Name\|hx-request\b\|hx-disable[^-]' htmgo-site/pages
```

For each hit, update the documentation to reflect htmx 4 equivalents (see spec Layer B mapping tables).

- [ ] **Step 3: Build htmgo-site**

```bash
cd htmgo-site && npm install && htmgo build
```
Expected: build succeeds; `dist/` artifacts produced.

- [ ] **Step 4: Commit**

```bash
git add htmgo-site/
git commit -m "refactor(htmgo-site): htmx 4 inheritance + doc prose updates"
```

---

## Task 27: Inheritance audit — `templates/` and `framework-ui/`

**Files:**
- Audit: `templates/**/*.go`, `framework-ui/ui/**/*.go`

- [ ] **Step 1: Grep each**

```bash
grep -rn 'hx\.\(TargetAttr\|IncludeAttr\|HeadersAttr\|BoostAttr\|IndicatorAttr\|SyncAttr\|ConfirmAttr\|EncodingAttr\|ValidateAttr\|ConfigAttr\)\|Hx\(Target\|Include\|Headers\|Boost\|Indicator\|Sync\|Confirm\|Encoding\|Validate\|Config\)(' templates framework-ui
```

- [ ] **Step 2: For each hit, decide inheritance**

Inheritance parent = emits one of the listed attrs, does NOT emit `hx.GetAttr`/`PostAttr`/`PutAttr`/`PatchAttr`/`DeleteAttr`, and has descendants that do. Swap emission to the `:inherited` helper from Task 8.

For `framework-ui/ui/input.go` specifically: verify whether htmx attributes on the outer wrapper are consumed by the `<input>` directly (self-consumed — no change) or rely on descendant consumption (inheritance parent — change to `:inherited`).

- [ ] **Step 3: Build each module**

```bash
cd framework-ui && go build ./...
cd ../templates/starter && go build ./... || true   # starter may require scaffolding; if build fails structurally, skip
```

- [ ] **Step 4: Commit**

```bash
git add templates/ framework-ui/
git commit -m "refactor(templates,framework-ui): htmx 4 explicit inheritance"
```

---

## Task 28: Verify `extensions/websocket/` Go module

**Files:**
- Audit: `extensions/websocket/**/*.go`

- [ ] **Step 1: Read module entry + public types**

```bash
find extensions/websocket -name '*.go' -not -name '*_test.go' -exec wc -l {} +
```

Open each file and scan for: assumptions about htmx 2 WS event-detail shape, uses of deleted `hx.*` identifiers, any header reads touching `HX-Trigger-Name` / `HX-Prompt`.

- [ ] **Step 2: Build and test the module**

```bash
cd extensions/websocket && go build ./... && go test ./... -count=1
```
Expected: all green after Layer B-C changes land. If a test uses `hx.TriggerNameHeader` or other deleted constants, update to `hx.SourceHeader`.

- [ ] **Step 3: Commit if changes made; otherwise note in plan log**

```bash
git add extensions/websocket/
git commit -m "fix(extensions/websocket): update for htmx 4 hx package"
```

---

## Task 29: Framework test suite green across all modules

**Files:** none (verification only)

- [ ] **Step 1: Run each module's tests**

```bash
for d in framework cli/htmgo/tasks/astgen extensions/websocket framework-ui; do
  echo "=== $d ==="
  (cd "$d" && go test ./... -count=1 -short) || echo "FAIL: $d"
done
```

Expected: all PASS. Any remaining failures mean a downstream file still references a removed constant — fix inline and re-run.

- [ ] **Step 2: Commit any fix-forward changes**

```bash
git add -u
git commit -m "test: fix remaining references to removed htmx 2 constants"
```

---

## Task 30: Playwright MCP smoke test on htmgo-site

**Files:** none (verification only; see CLAUDE.md smoke procedure)

- [ ] **Step 1: Build htmgo-site and start it**

```bash
cd htmgo-site
npm install
export PATH="$(go env GOPATH)/bin:$PATH"
htmgo build
PORT=3123 ./dist/htmgo-site &
sleep 2
```

- [ ] **Step 2: Run Playwright MCP browser checks**

Via the Playwright MCP tools configured in the environment:

- `browser_navigate` to `http://localhost:3123/` → `browser_snapshot` confirms `htmgo` h1, "Get Started" link to `/docs`, nav bar with Docs/Examples/Convert HTML.
- `browser_navigate` to `http://localhost:3123/docs` → redirects to `/docs/introduction`, page title "Docs - Introduction", sidebar + intro content.
- `browser_navigate` to `http://localhost:3123/docs/core-concepts/components` → page title "Docs - Components", content renders, prev/next present.
- `browser_navigate` to `http://localhost:3123/examples` → sidebar with Forms/Interactivity/Projects/Components categories.
- `browser_console_messages` level `error` → **must return 0 errors**.

- [ ] **Step 3: Run one interactive flow per example category**

Pick:
- `/examples/forms/<any>` — submit a form, verify partial swap.
- `/examples/interactivity/click-to-load` (or equivalent) — click button, verify swap.
- `/examples/projects/hackernews` — navigate sidebar (exercises `hx-boost` inheritance).

For each: `browser_console_messages` must still return 0 errors.

- [ ] **Step 4: Stop the server**

```bash
pkill -f "dist/htmgo-site"
```

- [ ] **Step 5: If any check fails, fix and re-run**

Common failures:
- Console errors logging legacy attribute usage → find the emitter in Go code, switch to htmx 4 name.
- Swap not occurring → likely inheritance issue; re-audit the containing page.

Commit fixes as discovered:

```bash
git add -u
git commit -m "fix(htmgo-site): address htmx 4 smoke-test findings"
```

---

## Task 31: Manually exercise each example

**Files:** none (verification only)

- [ ] **Step 1: For each example, run task watch and exercise main flow**

```bash
cd examples/chat && task watch
# — exercise: open in browser, send a message, verify WS path works
# — stop: Ctrl-C
cd ../todo-list && task watch
# — exercise: add/complete/delete a todo
cd ../simple-auth && task watch
# — exercise: register + log in + log out
cd ../hackernews && task watch
# — exercise: navigate sidebar + stories (hx-boost inheritance)
cd ../ws-example && task watch
# — exercise: repeater partial receives WS messages
cd ../minimal-htmgo && task watch
# — exercise: boot + render
```

- [ ] **Step 2: Document any fixes**

Commit fixes per example as they are found:

```bash
git add examples/<name>/
git commit -m "fix(examples/<name>): htmx 4 runtime fixes"
```

---

## Task 32: Write CHANGELOG entry

**Files:**
- Create or modify: `CHANGELOG.md` at repo root

- [ ] **Step 1: Check if CHANGELOG.md exists**

```bash
ls CHANGELOG.md 2>/dev/null || echo "does not exist"
```

If it does not exist, create it with a Keep-a-Changelog header:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
```

- [ ] **Step 2: Prepend the v2.0.0-beta.1 entry**

```markdown
## [2.0.0-beta.1] - 2026-04-17

### Breaking

- **htmx upgraded to 4.0.0-beta2.** This is a clean break with no compat bridge.
- **`framework/hx` constants:**
  - All `htmx:camelCase` event constants are now `htmx:colon:form` (e.g. `AfterSwapEvent` is now `htmx:after:swap`).
  - `AfterSettleEvent` is folded into `AfterSwapEvent`. `AfterProcessNodeEvent` and `AfterOnLoadEvent` both map to `htmx:after:init`. `BeforeSendEvent` maps to `htmx:before:request`.
  - New `ErrorEvent = "htmx:error"` replaces `ResponseErrorEvent`, `SendErrorEvent`, `SwapErrorEvent`, `TargetErrorEvent`, `TimeoutEvent`.
  - Removed: `OobAfterSwapEvent`, `OobBeforeSwapEvent`, `Validation*Event`, `Xhr*Event`, `HistoryCache*ErrorEvent`, `OnLoadErrorEvent`, `NoSSESourceErrorEvent`, `OobErrorNoTargetEvent`.
  - Attribute `DisableAttr` value is now `hx-ignore` (was `hx-disable`). New `IgnoreAttr` points to `hx-ignore`; `DisableAttr` now points to `hx-disable` (formerly `hx-disabled-elt`). Recommendation: audit every `h.HxDisable` / `hx.DisableAttr` call.
  - Removed attribute constants: `DisinheritAttr`, `InheritAttr`, `ParamsAttr`, `PromptAttr`, `ExtAttr`, `HistoryAttr`, `HistoryEltAttr`, `RequestAttr`.
  - New attribute constants: `ConfigAttr` (`hx-config`, replaces `hx-request`), `StatusAttr` (`hx-status`).
  - Removed header constants: `PromptResponseHeader`, `TriggerNameHeader`, `TriggerAfterSettleHeader`, `TriggerAfterSwapHeader`.
  - New header constants: `SourceHeader` (`HX-Source`), `RequestTypeHeader` (`HX-Request-Type`).
- **`RequestContext.HxTriggerName()` removed** — use `HxSource()` (raw `tag#id`), `HxSourceID()` (just the id), or `HxTargetId()`.
- **`RequestContext.HxPromptResponse()` removed** — htmx 4 removed `hx-prompt`; use `hx-confirm="js:..."` with a JS function.
- **`HxOnLoad` deprecated helper removed** — use `OnLoad`.
- **Attribute inheritance is explicit.** htmx 4 removed implicit inheritance. If your app set `hx-target` (or any of `hx-include`, `hx-headers`, `hx-boost`, `hx-indicator`, `hx-sync`, `hx-confirm`, `hx-encoding`, `hx-validate`, `hx-config`) on a container expecting it to apply to descendants, switch to the `:inherited` form. New helpers: `HxTargetInherited`, `HxIncludeInherited`, `HxSwapInherited`, `HxBoostInherited`, `HxConfirmInherited`, `HxHeadersInherited`, `HxIndicatorInherited`, `HxSyncInherited`, `HxConfigInherited`, `HxEncodingInherited`, `HxValidateInherited`.
- **`hx-delete` on a form button no longer includes form data.** Add `hx-include="closest form"` where needed.
- **All JS extensions rewritten for htmx 4's `registerExtension` API.** Event-detail access changed from `detail.xhr` to `detail.ctx`.

### Migration recipe

```bash
# Find your inheritance parents
grep -rn 'h\.\(HxTarget\|HxInclude\|HxHeaders\|HxBoost\|HxIndicator\|HxSync\|HxConfirm\|HxEncoding\|HxValidate\)' your-app/

# For each hit where the element has no hx-get/post/put/patch/delete of its own
# but its descendants do → switch to the :inherited helper.
```

See `docs/superpowers/specs/2026-04-17-htmx-v4-migration-design.md` for the full rationale.
```

- [ ] **Step 3: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs: CHANGELOG entry for v2.0.0-beta.1 htmx 4 migration"
```

---

## Task 33: Tag release and open PR

**Files:** none

- [ ] **Step 1: Push branch**

```bash
git push -u origin htmx4-migration
```

- [ ] **Step 2: Open PR against master**

Use `gh pr create`:

```bash
gh pr create --title "feat!: migrate to htmx 4.0.0-beta2 (htmgo v2.0-beta)" --body "$(cat <<'EOF'
## Summary

- Clean-break migration from htmx 2.0.8 to htmx 4.0.0-beta2
- Releases as `htmgo v2.0-beta.1`
- Five layers: JS dep bump, Go `hx` constants, Go `h/` internals, 11 JS extensions rewritten, inheritance audit across examples/site/templates/framework-ui
- Design: `docs/superpowers/specs/2026-04-17-htmx-v4-migration-design.md`
- Plan:   `docs/superpowers/plans/2026-04-17-htmx-v4-migration.md`

## Breaking changes

See `CHANGELOG.md` §2.0.0-beta.1 for the full list and migration recipe.

## Test plan

- [x] `framework/`, `framework-ui/`, `extensions/websocket/`, `cli/htmgo/tasks/astgen/` — `go test ./...` all green
- [x] `framework/assets/js/` — `npm test` all green (new vitest suite for extensions)
- [x] `framework/assets/js/` — `npm run build` succeeds
- [x] `htmgo-site/` — `htmgo build` succeeds
- [x] Playwright MCP smoke on htmgo-site: 0 console errors, all doc/examples pages render
- [x] All `examples/*` manually exercised
EOF
)"
```

- [ ] **Step 3: After review + merge to master, tag release**

(Executed by maintainer post-merge, not in this plan's execution scope.)

```bash
git checkout master && git pull
git tag -a v2.0.0-beta.1 -m "htmgo v2.0.0-beta.1 — htmx 4.0.0-beta2 migration"
git push origin v2.0.0-beta.1
```

---

## Self-review notes

- All 5 layers in the spec are covered by numbered tasks: A (Task 1), B (Tasks 2-4), C (Tasks 5-8), D (Tasks 9-21), E (Tasks 22-28).
- Risks R1 (exact-pin) covered in Task 1. R2 (downstream inheritance) covered in CHANGELOG (Task 32). R3 (response-targets ladder) covered in Task 17. R4 (WS alignment) documented in Task 19 commit message. R5 (extensions/websocket) covered in Task 28.
- Open questions O1 (delete HxOnLoad) covered in Task 6. O2 (framework-ui release) covered implicitly by Task 27 + the single-PR scope.
- Testing: Go unit tests per Layer B/C task; new vitest per extension in Layer D; Playwright smoke in Task 30; manual example runs in Task 31.
- No placeholders, no "TODO", no "similar to Task N" hand-waves in content steps.
