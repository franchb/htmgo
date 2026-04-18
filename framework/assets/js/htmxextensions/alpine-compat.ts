import htmx from "htmx.org";

// alpine-compat extension for htmx 4 — preserves Alpine.js state across
// htmx morph swaps and round-trips state through htmx's history cache.
// Hand-port of upstream `node_modules/htmx.org/dist/ext/hx-alpine-compat.js`
// (htmx.org 4.0.0-beta2). Every hook self-gates on `window.Alpine?.*` so it
// no-ops when Alpine is not loaded.

let api: any;
let deferCount = 0;

function maybeFlush() {
  if (deferCount > 0) deferCount--;
  if (deferCount === 0 && (window as any).Alpine?.flushAndStopDeferringMutations) {
    (window as any).Alpine.flushAndStopDeferringMutations();
  }
}

htmx.registerExtension("alpine-compat", {
  init(_api: any) {
    // filled in Task 2
  },

  htmx_before_swap(_elt: any, _detail: any) {
    // filled in Task 3
  },

  htmx_before_morph_node(_elt: any, _detail: any) {
    // filled in Task 4
  },

  htmx_history_cache_before_save(_elt: any, _detail: any) {
    // filled in Task 5
  },

  htmx_history_cache_after_restore(_elt: any, _detail: any) {
    // filled in Task 5
  },

  htmx_after_swap(_elt: any, _detail: any) {
    // filled in Task 3
  },

  htmx_finally_request(_elt: any, _detail: any) {
    // filled in Task 3
  },
});
