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
});
