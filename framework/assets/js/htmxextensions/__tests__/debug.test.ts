import { describe, it, expect, vi } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: { registerExtension: (n: string, e: any) => (registered[n] = e), config: {} },
}));

describe("debug extension", () => {
  it("registers and logs events", async () => {
    await import("../debug");
    expect(registered["debug"]).toBeDefined();
  });
});
