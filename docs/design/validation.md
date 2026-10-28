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
  `ValidateTarget` / `ValidateSwap`). `hx-sync="this:queue last"` allows one
  request per field at a time and collapses edits made meanwhile into one
  follow-up for the latest value, so a slow response for an old value never
  overwrites a newer one. (`this:replace` was rejected: in htmx 4.0.0-beta2
  `ReqQ.finish()` clears the current slot when the *aborted* request finishes,
  so a third request can overlap the second.) An empty **200** response clears
  the message; htmx does not swap 204/304.
- Inside a `<form>`, htmx builds a POST body from the whole form, so the
  validation endpoint receives every field, not just the one being validated.
  htmx 4 has no `hx-params` to narrow it; switching to GET would send only the
  field but put values in URLs and access logs. Open question for phase 2.
- `pattern` only applies to text-like inputs (text, search, url, tel, email,
  password); browsers ignore it on `textarea`, `select` and e.g. `number`.
- Native constraints block a native submit or an `hx-post` on the `<form>`.
  htmx does not validate requests from other elements (an `hx-post` button)
  unless they carry `hx-validate="true"`.
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
