import { describe, it, expect, vi, beforeEach } from "vitest";

const registered: Record<string, any> = {};
vi.mock("htmx.org", () => ({
  default: {
    registerExtension: (n: string, e: any) => (registered[n] = e),
    config: {} as any,
  },
}));

function makeApi(attrs: Record<string, string>) {
  return {
    getClosestAttributeValue: (_elt: any, name: string) => attrs[name] ?? null,
    findThisElement: (elt: any) => elt,
    querySelectorExt: (_elt: any, sel: string) => {
      const el = document.querySelector(sel);
      return el;
    },
  };
}

describe("response-targets extension", () => {
  let ext: any;
  beforeEach(async () => {
    await import("../response-targets");
    ext = registered["response-targets"];
    document.body.innerHTML = "";
  });

  it("registers with init and htmx_before_swap", () => {
    expect(ext).toBeDefined();
    expect(typeof ext.init).toBe("function");
    expect(typeof ext.htmx_before_swap).toBe("function");
  });

  it("init sets defaults in htmx.config without clobbering existing values", async () => {
    const cfg = (await import("htmx.org")).default.config as any;
    delete cfg.responseTargetUnsetsError;
    delete cfg.responseTargetSetsError;
    cfg.responseTargetPrefersExisting = undefined;
    cfg.responseTargetPrefersRetargetHeader = undefined;
    ext.init(makeApi({}));
    // Regression guard: the htmx-2 isError knobs must NOT be initialized.
    expect("responseTargetUnsetsError" in cfg).toBe(false);
    expect("responseTargetSetsError" in cfg).toBe(false);
    expect(cfg.responseTargetPrefersExisting).toBe(false);
    expect(cfg.responseTargetPrefersRetargetHeader).toBe(true);

    // Pre-set non-default values on the surviving knobs must be preserved on re-init.
    cfg.responseTargetPrefersExisting = true;
    cfg.responseTargetPrefersRetargetHeader = false;
    ext.init(makeApi({}));
    expect(cfg.responseTargetPrefersExisting).toBe(true);
    expect(cfg.responseTargetPrefersRetargetHeader).toBe(false);
  });

  function makeDetail(srcElt: HTMLElement, status: number, initialTarget: HTMLElement | null = null) {
    const mainTask: any = { type: "main", target: initialTarget };
    return {
      detail: {
        ctx: { response: { status }, sourceElement: srcElt },
        tasks: [mainTask],
      } as any,
      mainTask,
    };
  }

  it("retargets to hx-target-404 when status is 404", () => {
    document.body.innerHTML = `<div id="err"></div>`;
    const srcElt = document.createElement("button");
    ext.init(makeApi({ "hx-target-404": "#err" }));
    const { detail, mainTask } = makeDetail(srcElt, 404);
    ext.htmx_before_swap(srcElt, detail);
    expect((mainTask.target as HTMLElement)?.id).toBe("err");
  });

  it("falls back to hx-target-4xx when hx-target-404 is absent", () => {
    document.body.innerHTML = `<div id="err4"></div>`;
    const srcElt = document.createElement("button");
    ext.init(makeApi({ "hx-target-4xx": "#err4" }));
    const { detail, mainTask } = makeDetail(srcElt, 404);
    ext.htmx_before_swap(srcElt, detail);
    expect((mainTask.target as HTMLElement)?.id).toBe("err4");
  });

  it("falls back to hx-target-error for 4xx/5xx", () => {
    document.body.innerHTML = `<div id="errany"></div>`;
    const srcElt = document.createElement("button");
    ext.init(makeApi({ "hx-target-error": "#errany" }));
    const { detail, mainTask } = makeDetail(srcElt, 500);
    ext.htmx_before_swap(srcElt, detail);
    expect((mainTask.target as HTMLElement)?.id).toBe("errany");
  });

  it("does nothing when status is 200", () => {
    document.body.innerHTML = `<div id="err"></div>`;
    const srcElt = document.createElement("button");
    ext.init(makeApi({ "hx-target-error": "#err" }));
    const existing = document.createElement("div");
    const { detail, mainTask } = makeDetail(srcElt, 200, existing);
    ext.htmx_before_swap(srcElt, detail);
    expect(mainTask.target).toBe(existing);
  });

  it("hx-target-404='this' retargets to the request element", () => {
    const srcElt = document.createElement("button");
    srcElt.id = "btn";
    document.body.appendChild(srcElt);
    ext.init(makeApi({ "hx-target-404": "this" }));
    const { detail, mainTask } = makeDetail(srcElt, 404);
    ext.htmx_before_swap(srcElt, detail);
    expect(mainTask.target).toBe(srcElt);
  });
});
