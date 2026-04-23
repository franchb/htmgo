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

describe("ws extension", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    triggered.length = 0;
  });

  it("registers with init + htmx_before_process + htmx_before_cleanup hooks", async () => {
    await import("../ws");
    expect(registered["ws"]).toBeDefined();
    expect(typeof registered["ws"].init).toBe("function");
    expect(typeof registered["ws"].htmx_before_process).toBe("function");
    expect(typeof registered["ws"].htmx_before_cleanup).toBe("function");
  });

  it("init accepts an api ref without throwing", async () => {
    await import("../ws");
    const ext = registered["ws"];
    const fakeApi = { makeFragment: vi.fn() };
    expect(() => ext.init(fakeApi)).not.toThrow();
  });

  it("calls removeAssociatedScripts on htmx_before_cleanup", async () => {
    await import("../ws");
    const ext = registered["ws"];
    const elt = document.createElement("div");
    expect(() => ext.htmx_before_cleanup(elt, {})).not.toThrow();
  });

  it("htmx_before_cleanup handles null element gracefully", async () => {
    await import("../ws");
    const ext = registered["ws"];
    expect(() => ext.htmx_before_cleanup(null as any, {})).not.toThrow();
  });

  it("htmx_before_process connects elements with ws-connect attribute", async () => {
    await import("../ws");
    const ext = registered["ws"];

    const instances: any[] = [];
    function MockWebSocket(this: any, _url: string) {
      this.addEventListener = vi.fn();
      this.readyState = WebSocket.CONNECTING;
      instances.push(this);
    }
    (globalThis as any).WebSocket = MockWebSocket;
    (MockWebSocket as any).OPEN = 1;
    (MockWebSocket as any).CONNECTING = 0;

    const el = document.createElement("div");
    el.setAttribute("ws-connect", "/ws/unique-test-url-1");
    document.body.appendChild(el);

    ext.htmx_before_process(document.body, {});

    expect(instances.length).toBe(1);
  });

  it("htmx_before_process does not reconnect an already-processed URL", async () => {
    await import("../ws");
    const ext = registered["ws"];

    const instances: any[] = [];
    function MockWebSocket2(this: any, _url: string) {
      this.addEventListener = vi.fn();
      this.readyState = WebSocket.CONNECTING;
      instances.push(this);
    }
    (globalThis as any).WebSocket = MockWebSocket2;
    (MockWebSocket2 as any).OPEN = 1;
    (MockWebSocket2 as any).CONNECTING = 0;

    const el = document.createElement("div");
    el.setAttribute("ws-connect", "/ws/dedup-test-url-1");
    document.body.appendChild(el);

    ext.htmx_before_process(document.body, {});
    ext.htmx_before_process(document.body, {});

    // Should only create one WebSocket despite being called twice
    expect(instances.length).toBe(1);
  });
});
