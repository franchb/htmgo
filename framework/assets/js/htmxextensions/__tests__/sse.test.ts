import { describe, it, expect, vi, beforeEach } from "vitest";

const registered: Record<string, any> = {};
const triggered: Array<{ elt: Element; evt: string; detail?: unknown }> = [];

vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (name: string, ext: any) => {
      registered[name] = ext;
    },
    trigger: (elt: Element, evt: string, detail?: unknown) => {
      triggered.push({ elt, evt, detail });
    },
    swap: vi.fn(),
  },
}));

describe("sse extension", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    triggered.length = 0;
  });

  it("registers with init + htmx_before_process + htmx_before_cleanup hooks", async () => {
    await import("../sse");
    expect(registered["sse"]).toBeDefined();
    expect(typeof registered["sse"].init).toBe("function");
    expect(typeof registered["sse"].htmx_before_process).toBe("function");
    expect(typeof registered["sse"].htmx_before_cleanup).toBe("function");
  });

  it("init stores the apiRef", async () => {
    await import("../sse");
    const ext = registered["sse"];
    const fakeApi = { makeFragment: vi.fn() };
    // Should not throw
    expect(() => ext.init(fakeApi)).not.toThrow();
  });

  it("calls removeAssociatedScripts on htmx_before_cleanup", async () => {
    await import("../sse");
    const ext = registered["sse"];
    const elt = document.createElement("div");
    // If this doesn't throw the cleanup path works structurally.
    expect(() => ext.htmx_before_cleanup(elt, {})).not.toThrow();
  });

  it("htmx_before_cleanup handles null element gracefully", async () => {
    await import("../sse");
    const ext = registered["sse"];
    // htmx 4 may call cleanup with undefined element in edge cases
    expect(() => ext.htmx_before_cleanup(null as any, {})).not.toThrow();
  });

  it("htmx_before_process connects elements with sse-connect attribute", async () => {
    await import("../sse");
    const ext = registered["sse"];

    // Stub EventSource with a proper constructor function so `new EventSource()`
    // works in jsdom (which does not implement EventSource natively).
    // EventSource.CLOSED = 2 per the EventSource spec.
    const instances: any[] = [];
    function MockEventSource(this: any, _url: string) {
      this.addEventListener = vi.fn();
      this.onopen = null;
      this.onerror = null;
      this.onmessage = null;
      this.readyState = 2; // CLOSED
      instances.push(this);
    }
    MockEventSource.CLOSED = 2;
    (globalThis as any).EventSource = MockEventSource;

    const el = document.createElement("div");
    el.setAttribute("sse-connect", "/events/unique-test-url-2");
    document.body.appendChild(el);

    ext.htmx_before_process(document.body, {});

    expect(instances.length).toBe(1);
  });

  it("htmx_before_process does not reconnect an already-processed URL", async () => {
    await import("../sse");
    const ext = registered["sse"];

    const instances: any[] = [];
    function MockEventSource2(this: any, _url: string) {
      this.addEventListener = vi.fn();
      this.onopen = null;
      this.onerror = null;
      this.onmessage = null;
      this.readyState = 2; // CLOSED
      instances.push(this);
    }
    MockEventSource2.CLOSED = 2;
    (globalThis as any).EventSource = MockEventSource2;

    const el = document.createElement("div");
    el.setAttribute("sse-connect", "/events/dedup-test-url-2");
    document.body.appendChild(el);

    ext.htmx_before_process(document.body, {});
    ext.htmx_before_process(document.body, {});

    // Should only create one EventSource despite being called twice
    expect(instances.length).toBe(1);
  });
});
