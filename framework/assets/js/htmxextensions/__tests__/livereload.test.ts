import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    config: {},
  },
}));

describe("livereload extension", () => {
  let originalEventSource: any;
  let eventSourceCalls: string[] = [];
  let lastInstance: any;

  beforeEach(() => {
    eventSourceCalls = [];
    originalEventSource = (globalThis as any).EventSource;
    (globalThis as any).EventSource = class MockEventSource {
      url: string;
      onmessage: ((e: any) => void) | null = null;
      onerror: ((e: any) => void) | null = null;
      constructor(url: string) {
        this.url = url;
        eventSourceCalls.push(url);
        lastInstance = this;
      }
    };
    document.head.innerHTML = "";
  });

  afterEach(() => {
    (globalThis as any).EventSource = originalEventSource;
    document.head.innerHTML = "";
  });

  it("registers with init hook", async () => {
    await import("../livereload");
    expect(registered["livereload"]).toBeDefined();
    expect(typeof registered["livereload"].init).toBe("function");
  });

  it("does NOT open EventSource when meta marker absent", async () => {
    await import("../livereload");
    registered["livereload"].init({});
    expect(eventSourceCalls.length).toBe(0);
  });

  it("opens EventSource to /dev/livereload when meta marker present (default url)", async () => {
    const meta = document.createElement("meta");
    meta.setAttribute("name", "htmgo-livereload");
    document.head.appendChild(meta);
    await import("../livereload");
    registered["livereload"].init({});
    expect(eventSourceCalls).toContain("/dev/livereload");
  });

  it("uses custom url from meta content when provided", async () => {
    const meta = document.createElement("meta");
    meta.setAttribute("name", "htmgo-livereload");
    meta.setAttribute("content", "/custom/livereload");
    document.head.appendChild(meta);
    await import("../livereload");
    registered["livereload"].init({});
    expect(eventSourceCalls).toContain("/custom/livereload");
  });
});
