import { describe, it, expect, vi } from "vitest";

// Mock htmx before importing the extension
const registeredExtensions: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (name: string, ext: any) => {
      registeredExtensions[name] = ext;
    },
  },
}));

describe("htmgo extension", () => {
  it("registers with htmx 4 registerExtension API", async () => {
    await import("../htmgo");
    expect(registeredExtensions["htmgo"]).toBeDefined();
    expect(typeof registeredExtensions["htmgo"].htmx_before_cleanup).toBe("function");
    expect(typeof registeredExtensions["htmgo"].htmx_after_init).toBe("function");
  });

  it("invokes onload handlers on descendants with [onload]", async () => {
    await import("../htmgo");
    const ext = registeredExtensions["htmgo"];
    const parent = document.createElement("div");
    const child = document.createElement("span");
    let called = false;
    child.setAttribute("onload", "");
    child.onload = () => { called = true; };
    parent.appendChild(child);
    document.body.appendChild(parent);
    ext.htmx_after_init(parent, {});
    expect(called).toBe(true);
    document.body.removeChild(parent);
  });
});
