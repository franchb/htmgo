import { describe, it, expect, vi, beforeEach } from "vitest";

const registered: Record<string, any> = {};
const triggered: Array<{ elt: any; evt: string; detail?: any }> = [];
// Mock matches htmx 4 semantics: htmx.trigger dispatches a bubbling CustomEvent
// by default. Tests that only care about call-counts read `triggered`; tests
// that exercise ancestor-bubble behavior rely on the real DOM dispatch.
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    trigger: (elt: any, evt: string, detail?: any) => {
      triggered.push({ elt, evt, detail });
      if (elt && typeof elt.dispatchEvent === "function") {
        elt.dispatchEvent(
          new CustomEvent(evt, { bubbles: true, cancelable: true, detail }),
        );
      }
    },
    config: {},
  },
}));

describe("mutation-error extension", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    triggered.length = 0;
  });

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

  it("does not double-fire ancestor hx-on listeners via bubble + broadcast", async () => {
    await import("../mutation-error");
    const ext = registered["mutation-error"];

    const ancestor = document.createElement("div");
    ancestor.setAttribute("hx-on::on-mutation-error", "");
    const requestElt = document.createElement("button");
    ancestor.appendChild(requestElt);
    document.body.appendChild(ancestor);

    let ancestorHits = 0;
    ancestor.addEventListener("htmx:onMutationError", () => {
      ancestorHits++;
    });

    ext.htmx_after_request(requestElt, {
      ctx: {
        request: { method: "POST" },
        response: { status: 500 },
      },
    });

    // Ancestor should see the event exactly once — via bubbling from the
    // request element — not a second time from broadcast().
    expect(ancestorHits).toBe(1);
    // And broadcast() must not have called htmx.trigger on the ancestor.
    expect(
      triggered.filter((t) => t.elt === ancestor && t.evt === "htmx:onMutationError").length,
    ).toBe(0);
  });
});
