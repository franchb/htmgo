import { describe, it, expect, beforeEach, vi } from "vitest";

const registered: Record<string, any> = {};

vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    config: {} as any,
  },
}));

describe("alpine-compat extension", () => {
  let ext: any;

  beforeEach(async () => {
    await import("../alpine-compat");
    ext = registered["alpine-compat"];
    document.body.innerHTML = "";
    // Clear any prior Alpine stub
    delete (window as any).Alpine;
  });

  it("registers with all expected hooks", () => {
    expect(ext).toBeDefined();
    expect(typeof ext.init).toBe("function");
    expect(typeof ext.htmx_before_swap).toBe("function");
    expect(typeof ext.htmx_before_morph_node).toBe("function");
    expect(typeof ext.htmx_history_cache_before_save).toBe("function");
    expect(typeof ext.htmx_history_cache_after_restore).toBe("function");
    expect(typeof ext.htmx_after_swap).toBe("function");
    expect(typeof ext.htmx_finally_request).toBe("function");
  });

  describe("init hook", () => {
    function makeApi() {
      const api: any = {
        isSoftMatch: vi.fn((oldNode: any, newNode: any) => {
          return oldNode?.tagName === newNode?.tagName;
        }),
        morph: vi.fn(),
      };
      return api;
    }

    it("captures api ref and wraps isSoftMatch", () => {
      const api = makeApi();
      const original = api.isSoftMatch;
      ext.init(api);
      expect(api.isSoftMatch).not.toBe(original);
      expect(typeof api.isSoftMatch).toBe("function");
    });

    it("delegates to original isSoftMatch when nodes lack Alpine ID bindings", () => {
      const api = makeApi();
      const original = api.isSoftMatch;
      ext.init(api);
      const oldNode = document.createElement("div");
      const newNode = document.createElement("div");
      api.isSoftMatch(oldNode, newNode);
      // The wrapped isSoftMatch must have called through to the original.
      expect(original).toHaveBeenCalled();
      expect(api.isSoftMatch(oldNode, newNode)).toBe(true);
    });

    it("treats nodes with Alpine reactive IDs as soft-matching when tag names agree", () => {
      const api = makeApi();
      ext.init(api);
      const oldNode = document.createElement("div") as any;
      oldNode._x_bindings = { id: "dyn-1" };
      oldNode.id = "dyn-1";
      const newNode = document.createElement("div") as any;
      newNode.setAttribute(":id", "something-else");
      newNode.id = "dyn-2";
      // Even though ids differ, the wrapped isSoftMatch must return true.
      expect(api.isSoftMatch(oldNode, newNode)).toBe(true);
    });

    it("treats nodes with Alpine x-bind:id longhand as soft-matching when tag names agree", () => {
      const api = makeApi();
      ext.init(api);
      const oldNode = document.createElement("div") as any;
      oldNode._x_bindings = { id: "dyn-1" };
      oldNode.id = "dyn-1";
      const newNode = document.createElement("div") as any;
      // Use the full x-bind:id longhand form instead of the :id shorthand.
      newNode.setAttribute("x-bind:id", "something-else");
      newNode.id = "dyn-2";
      // Even though ids differ, the wrapped isSoftMatch must return true.
      expect(api.isSoftMatch(oldNode, newNode)).toBe(true);
    });

    it("returns false from wrapped isSoftMatch when Alpine reactive IDs present but tag mismatches", () => {
      const api = makeApi();
      ext.init(api);
      const oldNode = document.createElement("div") as any;
      oldNode._x_bindings = { id: "dyn-1" };
      const newNode = document.createElement("span") as any;
      newNode.setAttribute(":id", "whatever");
      expect(api.isSoftMatch(oldNode, newNode)).toBe(false);
    });
  });

  describe("defer/flush flow", () => {
    let ext: any;
    let deferMutations: ReturnType<typeof vi.fn>;
    let flush: ReturnType<typeof vi.fn>;

    beforeEach(async () => {
      vi.resetModules();
      // Re-import the extension to reset module-level deferCount + api.
      await import("../alpine-compat");
      ext = registered["alpine-compat"];
      deferMutations = vi.fn();
      flush = vi.fn();
      (window as any).Alpine = {
        deferMutations,
        flushAndStopDeferringMutations: flush,
        closestDataStack: vi.fn(() => [{}]),
        cloneNode: vi.fn(),
      };
      ext.init({ isSoftMatch: (_a: any, _b: any) => true, morph: vi.fn() });
    });

    it("htmx_before_swap calls Alpine.deferMutations on the first call of a batch", () => {
      ext.htmx_before_swap(document.body, { tasks: [], ctx: {} });
      expect(deferMutations).toHaveBeenCalledTimes(1);
    });

    it("subsequent htmx_before_swap calls within the same batch do not re-defer", () => {
      ext.htmx_before_swap(document.body, { tasks: [], ctx: {} });
      ext.htmx_before_swap(document.body, { tasks: [], ctx: {} });
      ext.htmx_before_swap(document.body, { tasks: [], ctx: {} });
      expect(deferMutations).toHaveBeenCalledTimes(1);
    });

    it("htmx_after_swap flushes only when deferCount reaches zero", () => {
      const ctxA: any = {};
      const ctxB: any = {};
      ext.htmx_before_swap(document.body, { tasks: [], ctx: ctxA });
      ext.htmx_before_swap(document.body, { tasks: [], ctx: ctxB });
      // deferCount is now 2
      ext.htmx_after_swap(document.body, { ctx: ctxA });
      expect(flush).not.toHaveBeenCalled();
      ext.htmx_after_swap(document.body, { ctx: ctxB });
      expect(flush).toHaveBeenCalledTimes(1);
    });

    it("htmx_after_swap marks ctx._alpineFlushed = true", () => {
      const ctx: any = {};
      ext.htmx_before_swap(document.body, { tasks: [], ctx });
      ext.htmx_after_swap(document.body, { ctx });
      expect(ctx._alpineFlushed).toBe(true);
    });

    it("htmx_finally_request flushes if htmx_after_swap did not", () => {
      const ctx: any = {};
      ext.htmx_before_swap(document.body, { tasks: [], ctx });
      // No htmx_after_swap — simulate a short-circuited swap.
      ext.htmx_finally_request(document.body, { ctx });
      expect(flush).toHaveBeenCalledTimes(1);
    });

    it("htmx_finally_request does NOT flush when ctx._alpineFlushed is already set", () => {
      const ctx: any = {};
      ext.htmx_before_swap(document.body, { tasks: [], ctx });
      ext.htmx_after_swap(document.body, { ctx });
      expect(flush).toHaveBeenCalledTimes(1);
      flush.mockClear();
      ext.htmx_finally_request(document.body, { ctx });
      expect(flush).not.toHaveBeenCalled();
    });

    it("htmx_before_swap is a no-op when window.Alpine is missing required APIs", () => {
      (window as any).Alpine = {}; // no deferMutations / closestDataStack / cloneNode
      ext.htmx_before_swap(document.body, { tasks: [], ctx: {} });
      expect(deferMutations).not.toHaveBeenCalled();
    });

    it("finally_request of a ctx that never entered before_swap does not release another ctx's deferral", () => {
      // Concurrency regression: request A starts and enters before_swap;
      // request B fails before any swap (e.g. network error), so htmx only
      // fires finally_request on B. B's maybeFlush must NOT decrement A's
      // deferCount and prematurely flush Alpine mutations.
      const ctxA: any = {};
      const ctxB: any = {}; // never passed to before_swap
      ext.htmx_before_swap(document.body, { tasks: [], ctx: ctxA });
      ext.htmx_finally_request(document.body, { ctx: ctxB });
      expect(flush).not.toHaveBeenCalled();
      // A's own after_swap should still be able to flush.
      ext.htmx_after_swap(document.body, { ctx: ctxA });
      expect(flush).toHaveBeenCalledTimes(1);
    });
  });

  describe("htmx_before_morph_node hook", () => {
    let ext: any;
    let closestDataStack: ReturnType<typeof vi.fn>;
    let cloneNode: ReturnType<typeof vi.fn>;
    let morph: ReturnType<typeof vi.fn>;

    beforeEach(async () => {
      vi.resetModules();
      await import("../alpine-compat");
      ext = registered["alpine-compat"];
      closestDataStack = vi.fn(() => [{ count: 7 }]);
      cloneNode = vi.fn();
      morph = vi.fn();
      (window as any).Alpine = { closestDataStack, cloneNode };
      ext.init({ isSoftMatch: (_a: any, _b: any) => true, morph });
    });

    it("copies _x_dataStack from oldNode to newNode", () => {
      const oldNode: any = document.createElement("div");
      document.body.appendChild(oldNode); // isConnected = true
      const newNode: any = document.createElement("div");
      ext.htmx_before_morph_node(document.body, { oldNode, newNode });
      expect(newNode._x_dataStack).toEqual([{ count: 7 }]);
      expect(closestDataStack).toHaveBeenCalledWith(oldNode);
    });

    it("calls Alpine.cloneNode when oldNode is connected", () => {
      const oldNode: any = document.createElement("div");
      document.body.appendChild(oldNode);
      const newNode: any = document.createElement("div");
      ext.htmx_before_morph_node(document.body, { oldNode, newNode });
      expect(cloneNode).toHaveBeenCalledWith(oldNode, newNode);
    });

    it("skips Alpine.cloneNode when oldNode is disconnected (template child)", () => {
      const oldNode: any = document.createElement("div");
      // Not appended — isConnected is false
      const newNode: any = document.createElement("div");
      ext.htmx_before_morph_node(document.body, { oldNode, newNode });
      expect(cloneNode).not.toHaveBeenCalled();
    });

    it("morphs teleport target when both nodes carry _x_teleport", () => {
      const oldNode: any = document.createElement("div");
      document.body.appendChild(oldNode);
      const oldTeleport = document.createElement("div");
      const newTeleport = document.createElement("div");
      oldNode._x_teleport = oldTeleport;
      const newNode: any = document.createElement("div");
      newNode._x_teleport = newTeleport;
      ext.htmx_before_morph_node(document.body, { oldNode, newNode });
      expect(morph).toHaveBeenCalledTimes(1);
      const [target, fragment, third] = morph.mock.calls[0];
      expect(target).toBe(oldTeleport);
      expect(fragment).toBeInstanceOf(DocumentFragment);
      expect(fragment.firstChild).toBe(newTeleport);
      expect(third).toBe(false);
    });

    it("no-ops when window.Alpine lacks closestDataStack/cloneNode", () => {
      (window as any).Alpine = {};
      const oldNode: any = document.createElement("div");
      document.body.appendChild(oldNode);
      const newNode: any = document.createElement("div");
      ext.htmx_before_morph_node(document.body, { oldNode, newNode });
      expect(newNode._x_dataStack).toBeUndefined();
    });
  });

  describe("history cache hooks", () => {
    let ext: any;
    let destroyTree: ReturnType<typeof vi.fn>;

    beforeEach(async () => {
      vi.resetModules();
      await import("../alpine-compat");
      ext = registered["alpine-compat"];
      destroyTree = vi.fn();
      (window as any).Alpine = { destroyTree };
    });

    it("before_save serializes _x_dataStack[0] for each [x-data] node", () => {
      const root = document.createElement("div");
      const a: any = document.createElement("div");
      a.setAttribute("x-data", "");
      a._x_dataStack = [{ count: 1 }];
      const b: any = document.createElement("div");
      b.setAttribute("x-data", "");
      b._x_dataStack = [{ open: true }];
      root.append(a, b);
      document.body.appendChild(root);

      ext.htmx_history_cache_before_save(document.body, { target: root });

      expect(a.getAttribute("data-alpine-state")).toBe(JSON.stringify({ count: 1 }));
      expect(b.getAttribute("data-alpine-state")).toBe(JSON.stringify({ open: true }));
      expect(destroyTree).toHaveBeenCalledWith(root);
    });

    it("before_save skips [x-data] nodes without _x_dataStack", () => {
      const root = document.createElement("div");
      const a: any = document.createElement("div");
      a.setAttribute("x-data", "");
      // no _x_dataStack
      root.append(a);
      document.body.appendChild(root);

      ext.htmx_history_cache_before_save(document.body, { target: root });

      expect(a.hasAttribute("data-alpine-state")).toBe(false);
      expect(destroyTree).toHaveBeenCalledWith(root);
    });

    it("before_save no-ops when window.Alpine.destroyTree is absent", () => {
      (window as any).Alpine = {};
      const root = document.createElement("div");
      const a: any = document.createElement("div");
      a.setAttribute("x-data", "");
      a._x_dataStack = [{ count: 1 }];
      root.append(a);
      document.body.appendChild(root);

      ext.htmx_history_cache_before_save(document.body, { target: root });
      expect(a.hasAttribute("data-alpine-state")).toBe(false);
    });

    it("after_restore merges saved state into _x_dataStack[0] and removes attribute", () => {
      const a: any = document.createElement("div");
      a.setAttribute("data-alpine-state", JSON.stringify({ count: 42 }));
      a._x_dataStack = [{ count: 0, other: "keep" }];
      document.body.appendChild(a);

      ext.htmx_history_cache_after_restore(document.body, {});

      expect(a._x_dataStack[0]).toEqual({ count: 42, other: "keep" });
      expect(a.hasAttribute("data-alpine-state")).toBe(false);
    });

    it("after_restore silently skips nodes without _x_dataStack", () => {
      const a: any = document.createElement("div");
      a.setAttribute("data-alpine-state", JSON.stringify({ x: 1 }));
      document.body.appendChild(a);

      ext.htmx_history_cache_after_restore(document.body, {});

      expect(a.hasAttribute("data-alpine-state")).toBe(false);
    });

    it("after_restore no-ops when window.Alpine is absent", () => {
      delete (window as any).Alpine;
      const a: any = document.createElement("div");
      a.setAttribute("data-alpine-state", JSON.stringify({ x: 1 }));
      document.body.appendChild(a);

      ext.htmx_history_cache_after_restore(document.body, {});
      expect(a.hasAttribute("data-alpine-state")).toBe(true);
    });
  });

  describe("API-surface contract", () => {
    it("references every documented Alpine + htmx internal API symbol", async () => {
      // This is a source-level assertion. If Alpine or htmx renames an API,
      // the extension must be updated and this test updated deliberately.
      const fs = await import("node:fs/promises");
      const path = await import("node:path");
      const url = await import("node:url");
      const here = path.dirname(url.fileURLToPath(import.meta.url));
      const srcPath = path.join(here, "..", "alpine-compat.ts");
      const src = await fs.readFile(srcPath, "utf8");

      const alpineSymbols = [
        "closestDataStack",
        "cloneNode",
        "deferMutations",
        "destroyTree",
        "flushAndStopDeferringMutations",
      ];
      for (const sym of alpineSymbols) {
        expect(src, `Alpine.${sym} must remain referenced`).toMatch(sym);
      }

      const htmxApiSymbols = ["isSoftMatch", "morph"];
      for (const sym of htmxApiSymbols) {
        expect(src, `api.${sym} must remain referenced`).toMatch(sym);
      }
    });
  });
});
