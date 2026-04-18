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
  init(internalAPI: any) {
    api = internalAPI;

    // Override isSoftMatch to handle Alpine reactive IDs.
    // When both nodes carry Alpine-managed ID bindings (`_x_bindings.id` on the
    // old node and a `:id` / `x-bind:id` attr on the new node), ignore the id
    // mismatch and match on tagName only. Otherwise defer to the original.
    //
    // NOTE: We use explicit hasAttribute checks instead of a CSS attribute
    // selector because jsdom (and some real-browser CSS parsers) fail to match
    // `[x-bind\:id]` — the colon in the attribute name is not reliably escapable
    // inside a CSS selector string.  The shorthand `[\:id]` works fine for `:id`
    // but `[x-bind\:id]`, `[x-bind\3A id]`, and `[x-bind\3A\ id]` all return
    // false in jsdom even when the attribute is present.  hasAttribute has no
    // such limitation, making this approach more portable across environments.
    const originalIsSoftMatch = api.isSoftMatch;
    api.isSoftMatch = function (oldNode: any, newNode: any) {
      if (
        oldNode?._x_bindings?.id &&
        (newNode?.hasAttribute?.(":id") || newNode?.hasAttribute?.("x-bind:id"))
      ) {
        return oldNode instanceof Element && oldNode.tagName === newNode.tagName;
      }
      return originalIsSoftMatch(oldNode, newNode);
    };
  },

  htmx_before_swap(_elt: any, detail: any) {
    const alpine = (window as any).Alpine;
    if (!alpine?.closestDataStack || !alpine?.cloneNode || !alpine?.deferMutations) {
      return;
    }
    if (deferCount === 0) {
      alpine.deferMutations();
    }
    deferCount++;

    // Note: upstream iterates `detail.tasks` looking for innerMorph / outerMorph
    // entries and resolves string targets to DOM nodes, but the loop body is a
    // no-op in the upstream source (it just sets a local `target` that is never
    // used). Preserved as a comment so future readers know why the loop is
    // absent here.
    // for (let task of detail.tasks) { if (innerMorph/outerMorph) { ... } }
  },

  htmx_before_morph_node(_elt: any, detail: any) {
    const alpine = (window as any).Alpine;
    if (!alpine?.closestDataStack || !alpine?.cloneNode) return;

    const { oldNode, newNode } = detail;
    if (!oldNode || !newNode) return;

    newNode._x_dataStack = alpine.closestDataStack(oldNode);

    // Skip cloneNode for template children — reactive content on a disconnected
    // node throws inside Alpine's evaluator.
    if (!oldNode.isConnected) return;
    alpine.cloneNode(oldNode, newNode);

    // If both carry a teleport target, morph the teleport destinations too.
    if (oldNode._x_teleport && newNode._x_teleport) {
      const fragment = document.createDocumentFragment();
      fragment.append(newNode._x_teleport);
      api.morph(oldNode._x_teleport, fragment, false);
    }
  },

  htmx_history_cache_before_save(_elt: any, detail: any) {
    const alpine = (window as any).Alpine;
    if (!alpine?.destroyTree) return;

    detail.target.querySelectorAll("[x-data]").forEach((el: any) => {
      if (el._x_dataStack) {
        el.setAttribute("data-alpine-state", JSON.stringify(el._x_dataStack[0]));
      }
    });
    alpine.destroyTree(detail.target);
  },

  htmx_history_cache_after_restore(_elt: any, _detail: any) {
    const alpine = (window as any).Alpine;
    if (!alpine) return;

    document.querySelectorAll("[data-alpine-state]").forEach((el: any) => {
      const saved = JSON.parse(el.getAttribute("data-alpine-state"));
      el.removeAttribute("data-alpine-state");
      if (el._x_dataStack) Object.assign(el._x_dataStack[0], saved);
    });
  },

  htmx_after_swap(_elt: any, detail: any) {
    if (detail?.ctx) detail.ctx._alpineFlushed = true;
    maybeFlush();
  },

  htmx_finally_request(_elt: any, detail: any) {
    if (!detail?.ctx?._alpineFlushed) maybeFlush();
  },
});
