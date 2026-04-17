import { describe, it, expect, vi, beforeEach } from "vitest";

const registered: Record<string, any> = {};
const triggered: Array<{ elt: Element; evt: string }> = [];
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    trigger: (elt: Element, evt: string) => triggered.push({ elt, evt }),
    config: {},
  },
}));

describe("path-deps extension", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    triggered.length = 0;
  });

  it("registers with after_request hook", async () => {
    await import("../pathdeps");
    expect(registered["path-deps"]).toBeDefined();
    expect(typeof registered["path-deps"].htmx_after_request).toBe("function");
  });

  it("triggers path-deps event on matching element after POST", async () => {
    await import("../pathdeps");
    const ext = registered["path-deps"];
    const el = document.createElement("div");
    el.setAttribute("path-deps", "/api/users");
    document.body.appendChild(el);
    ext.htmx_after_request(document.body, {
      ctx: { request: { method: "POST", action: "/api/users" } },
    });
    expect(triggered.length).toBe(1);
    expect(triggered[0].evt).toBe("path-deps");
  });

  it("triggers on wildcard pattern match", async () => {
    await import("../pathdeps");
    const ext = registered["path-deps"];
    const el = document.createElement("div");
    el.setAttribute("path-deps", "/api/*");
    document.body.appendChild(el);
    ext.htmx_after_request(document.body, {
      ctx: { request: { method: "DELETE", action: "/api/users/5" } },
    });
    expect(triggered.length).toBe(1);
  });

  it("does not trigger on non-matching path", async () => {
    await import("../pathdeps");
    const ext = registered["path-deps"];
    const el = document.createElement("div");
    el.setAttribute("path-deps", "/api/users");
    document.body.appendChild(el);
    ext.htmx_after_request(document.body, {
      ctx: { request: { method: "POST", action: "/api/posts" } },
    });
    expect(triggered.length).toBe(0);
  });

  it("does not trigger on GET (not a mutation)", async () => {
    await import("../pathdeps");
    const ext = registered["path-deps"];
    const el = document.createElement("div");
    el.setAttribute("path-deps", "/api/users");
    document.body.appendChild(el);
    ext.htmx_after_request(document.body, {
      ctx: { request: { method: "GET", action: "/api/users" } },
    });
    expect(triggered.length).toBe(0);
  });

  it("does not trigger when path-deps is 'ignore'", async () => {
    await import("../pathdeps");
    const ext = registered["path-deps"];
    const el = document.createElement("div");
    el.setAttribute("path-deps", "ignore");
    document.body.appendChild(el);
    ext.htmx_after_request(document.body, {
      ctx: { request: { method: "POST", action: "/api/users" } },
    });
    expect(triggered.length).toBe(0);
  });
});
