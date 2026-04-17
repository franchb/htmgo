import { describe, it, expect, vi } from "vitest";

const registered: Record<string, any> = {};
const triggered: Array<{ elt: any; evt: string; detail?: any }> = [];
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    trigger: (elt: any, evt: string, detail?: any) => {
      triggered.push({ elt, evt, detail });
    },
    config: {},
  },
}));

describe("mutation-error extension", () => {
  it("registers with after_request hook", async () => {
    await import("../mutation-error");
    expect(registered["mutation-error"]).toBeDefined();
    expect(typeof registered["mutation-error"].htmx_after_request).toBe("function");
  });

  it("fires htmx:onMutationError on 500 POST", async () => {
    await import("../mutation-error");
    const ext = registered["mutation-error"];
    const elt = document.createElement("button");
    triggered.length = 0;
    ext.htmx_after_request(elt, {
      ctx: {
        request: { method: "POST" },
        response: { status: 500 },
      },
    });
    expect(triggered.some((t) => t.evt === "htmx:onMutationError")).toBe(true);
  });

  it("does not fire on 200 POST", async () => {
    await import("../mutation-error");
    const ext = registered["mutation-error"];
    const elt = document.createElement("button");
    triggered.length = 0;
    ext.htmx_after_request(elt, {
      ctx: {
        request: { method: "POST" },
        response: { status: 200 },
      },
    });
    expect(triggered.some((t) => t.evt === "htmx:onMutationError")).toBe(false);
  });

  it("does not fire on 500 GET (GET is not a mutation)", async () => {
    await import("../mutation-error");
    const ext = registered["mutation-error"];
    const elt = document.createElement("button");
    triggered.length = 0;
    ext.htmx_after_request(elt, {
      ctx: {
        request: { method: "GET" },
        response: { status: 500 },
      },
    });
    expect(triggered.some((t) => t.evt === "htmx:onMutationError")).toBe(false);
  });
});
