import { describe, it, expect, vi, beforeEach } from "vitest";

const swap = vi.fn();
vi.mock("htmx.org", () => ({ default: { swap } }));

describe("applyOobSwap", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    swap.mockClear();
    for (const k of Object.keys(document.body.dataset)) delete document.body.dataset[k];
  });

  it("keeps elements after a literal </template> in the message", async () => {
    const { applyOobSwap } = await import("../oob");
    const ele = document.createElement("div");
    applyOobSwap(
      ele,
      '<div id="a" hx-swap-oob="true">1</div></template><div id="b" hx-swap-oob="true">2</div>',
    );
    expect(swap).toHaveBeenCalledTimes(1);
    const text: string = swap.mock.calls[0][0].text;
    expect(text).toContain('id="a"');
    expect(text).toContain('id="b"');
    expect(swap.mock.calls[0][0]).toMatchObject({ sourceElement: ele, target: ele, swap: "none" });
  });

  it("preserves table rows", async () => {
    const { applyOobSwap } = await import("../oob");
    applyOobSwap(document.createElement("div"), '<tr id="r" hx-swap-oob="true"><td>1</td></tr>');
    expect(swap.mock.calls[0][0].text).toBe('<tr id="r" hx-swap-oob="true"><td>1</td></tr>');
  });

  it("runs __eval scripts in the live document and strips them from the swap", async () => {
    const { applyOobSwap } = await import("../oob");
    applyOobSwap(
      document.createElement("div"),
      '<div id="w" hx-swap-oob="true"><script id="__eval_1" data-k="v">document.body.dataset.hit1=1</script></div>' +
        '<script id="other">document.body.dataset.hit2=1</script>',
    );
    const script = document.body.querySelector('script[id="__eval_1"]') as HTMLScriptElement;
    expect(script).not.toBeNull();
    expect(script.dataset.k).toBe("v");
    expect(script.ownerDocument).toBe(document);
    expect(document.body.dataset.hit1).toBe("1");
    expect(document.body.dataset.hit2).toBeUndefined();
    const text: string = swap.mock.calls[0][0].text;
    expect(text).not.toContain("__eval_1");
    expect(text).toContain('id="other"');
  });

  it("does not execute anything while parsing", async () => {
    const { applyOobSwap } = await import("../oob");
    applyOobSwap(
      document.createElement("div"),
      '<img src="x" onerror="document.body.dataset.hit3=1"><script id="other">document.body.dataset.hit2=1</script>',
    );
    await new Promise((r) => setTimeout(r, 20));
    expect({ ...document.body.dataset }).toEqual({});
  });
});
