# Form validation in htmgo

Tracking issue: [maddalax/htmgo#60](https://github.com/maddalax/htmgo/issues/60).

## Phase 1 (this change)

- `h.Validation(opts...)` returns attributes only, so it composes with any
  `input`, `select` or `textarea`.
- Native constraint validation:
  - `h.ValidateRequired()` → `required`
  - `h.ValidatePattern(re, message)` → `pattern` (the browser anchors it with
    `^…$` and uses JavaScript regex syntax), plus `title` as the mismatch hint.
  These run in the browser with no JavaScript and block form submission.
- Server validation: `h.ValidatePartial(fn)` / `h.ValidatePath(path)` render
  `hx-post` on the field, triggered on `change` (override with `ValidateOn`).
  The response is swapped into `next div` with `innerHTML` (override with
  `ValidateTarget` / `ValidateSwap`). `hx-sync="this:replace"` makes a newer
  edit replace an in-flight request, so a slow response for an old value never
  overwrites a newer one. An empty response clears the message.
- Server validation is advisory and does not block submit. The handler that
  processes the submitted form must validate again.
- No JavaScript is emitted, so it works under a nonce-only CSP
  (`AppOpts.ContentSecurityPolicy`).

## Not in phase 1

- **Custom browser messages for `required`, and blocking submit on server
  errors.** Both need `setCustomValidity`, so a small htmx extension: on
  `htmx:after:swap` of a validation response, call `setCustomValidity(text)`
  on the source field, and clear it on `input`. Ship it as a self-registering
  htmx 4 extension in `framework/assets/js/htmxextensions/`.
- **Validating the whole form on submit with errors mapped per field.**
  Proposed: the partial returns `h.SwapMany` of per-field error elements, with
  ids derived from field names.
- **JS-expression validators** (the issue's `validation.Js()`). Rejected for
  now: they need inline JS and conflict with CSP.
- **Type-specific helpers** (`min`/`max`/`minlength`/`maxlength`/`step`).
  Trivial follow-ups that map to native attributes.
- **framework-ui `ui.Input`.** It moves to `h.Validation` in a follow-up PR,
  once the framework dependency bump that includes this change has landed
  (framework-ui pins a framework pseudo-version and CI builds with
  `GOWORK=off`).
