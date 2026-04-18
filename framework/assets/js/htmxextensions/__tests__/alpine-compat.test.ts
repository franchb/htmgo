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
});
