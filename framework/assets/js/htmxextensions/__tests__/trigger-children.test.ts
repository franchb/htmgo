import { describe, it, expect, vi, beforeEach } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    config: {},
  },
}));

describe("trigger-children extension", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
  });

  it("registers with init hook", async () => {
    await import("../trigger-children");
    expect(registered["trigger-children"]).toBeDefined();
    expect(typeof registered["trigger-children"].init).toBe("function");
  });

  it("init installs document-level listeners that fan htmx events to descendants with matching hx-on:: attribute", async () => {
    await import("../trigger-children");
    const ext = registered["trigger-children"];
    ext.init({});

    const parent = document.createElement("div");
    const child = document.createElement("span");
    child.setAttribute("hx-on::after:swap", ""); // attribute present triggers fan-out
    let fired = false;
    child.addEventListener("htmx:after:swap", (e: any) => {
      if (e.detail?.meta === "trigger-children") fired = true;
    });
    parent.appendChild(child);
    document.body.appendChild(parent);

    // Dispatch a bubble event from parent
    const evt = new CustomEvent("htmx:after:swap", {
      bubbles: true,
      cancelable: true,
      detail: { target: parent },
    });
    parent.dispatchEvent(evt);

    // setTimeout(1) in the implementation — allow event loop to run
    await new Promise((r) => setTimeout(r, 20));
    expect(fired).toBe(true);
  });

  it("preserves camelCase event names when dispatching to children (e.g. htmx:after:viewTransition)", async () => {
    await import("../trigger-children");
    const ext = registered["trigger-children"];
    ext.init({});

    const parent = document.createElement("div");
    const child = document.createElement("span");
    // hx-on attribute uses kebab-case (DOM attribute name rule)
    child.setAttribute("hx-on::after:view-transition", "");
    let camelFired = false;
    let kebabFired = false;
    child.addEventListener("htmx:after:viewTransition", (e: any) => {
      if (e.detail?.meta === "trigger-children") camelFired = true;
    });
    child.addEventListener("htmx:after:view-transition", (e: any) => {
      if (e.detail?.meta === "trigger-children") kebabFired = true;
    });
    parent.appendChild(child);
    document.body.appendChild(parent);

    const evt = new CustomEvent("htmx:after:viewTransition", {
      bubbles: true,
      cancelable: true,
      detail: { target: parent },
    });
    parent.dispatchEvent(evt);

    await new Promise((r) => setTimeout(r, 20));
    expect(camelFired).toBe(true);
    expect(kebabFired).toBe(false);
  });

  it("ignores re-entrant trigger-children events (meta marker)", async () => {
    await import("../trigger-children");
    const ext = registered["trigger-children"];
    ext.init({});

    const parent = document.createElement("div");
    const child = document.createElement("span");
    child.setAttribute("hx-on::after:swap", "");
    let count = 0;
    child.addEventListener("htmx:after:swap", () => { count++; });
    parent.appendChild(child);
    document.body.appendChild(parent);

    const reentrant = new CustomEvent("htmx:after:swap", {
      bubbles: true,
      cancelable: true,
      detail: { target: parent, meta: "trigger-children" },
    });
    parent.dispatchEvent(reentrant);

    await new Promise((r) => setTimeout(r, 20));
    // The reentrant event was dispatched on parent, not child, so child's listener
    // is never invoked. The extension's guard prevents any fan-out. count stays 0.
    expect(count).toBe(0);
  });
});
