import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("htmx.org", () => ({
  default: {
    registerExtension: vi.fn(),
    trigger: vi.fn(),
    swap: vi.fn(),
    config: {},
  },
}));

// ws-event-handler.ts installs a window.addEventListener("load") hook.
// Import it, simulate load, and verify behavior.
describe("ws-event-handler extension", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
  });

  it("loads without importing a deleted module", async () => {
    // If the import itself throws (e.g. stale import from "./extension"), test fails.
    await import("../ws-event-handler");
    expect(true).toBe(true);
  });

  it("exports addWsEventHandlers function", async () => {
    const mod = await import("../ws-event-handler");
    expect(typeof mod.addWsEventHandlers).toBe("function");
  });

  it("addWsEventHandlers installs a MutationObserver on document.body", async () => {
    const observeSpy = vi.fn();
    const origMutationObserver = globalThis.MutationObserver;

    // vi.fn().mockImplementation returns an arrow function, which cannot be
    // used as a constructor.  Use a real class instead.
    class MockMutationObserver {
      observe = observeSpy;
      disconnect = vi.fn();
      constructor(_cb: MutationCallback) {}
    }
    globalThis.MutationObserver = MockMutationObserver as unknown as typeof MutationObserver;

    const mod = await import("../ws-event-handler");
    mod.addWsEventHandlers();

    expect(observeSpy).toHaveBeenCalledWith(document.body, {
      childList: true,
      subtree: true,
    });

    globalThis.MutationObserver = origMutationObserver;
  });

  it("addWsEventHandlers registers event listeners for data-handler-id elements", async () => {
    const mod = await import("../ws-event-handler");

    const btn = document.createElement("button");
    btn.setAttribute("data-handler-id", "handler-1");
    btn.setAttribute("data-handler-event", "click");
    document.body.appendChild(btn);

    const addEventListenerSpy = vi.spyOn(btn, "addEventListener");

    mod.addWsEventHandlers();

    expect(addEventListenerSpy).toHaveBeenCalledWith("click", expect.any(Function));
  });
});
