import htmx from "htmx.org";
const config: any = htmx.config;

let api: any;
const attrPrefix = "hx-target-";

// IE11 doesn't support string.startsWith
function startsWith(str: string, prefix: string) {
  return str.substring(0, prefix.length) === prefix;
}

/**
 * @param {HTMLElement} elt
 * @param respCodeNumber
 * @returns {HTMLElement | null}
 */
function getRespCodeTarget(elt: Element, respCodeNumber: number) {
  if (!elt || !respCodeNumber) return null;

  const respCode = respCodeNumber.toString();

  // '*' is the original syntax, as the obvious character for a wildcard.
  // The 'x' alternative was added for maximum compatibility with HTML
  // templating engines, due to ambiguity around which characters are
  // supported in HTML attributes.
  //
  // Start with the most specific possible attribute and generalize from
  // there.
  const attrPossibilities = [
    respCode,

    respCode.slice(0, 2) + "*",
    respCode.slice(0, 2) + "x",

    respCode.slice(0, 1) + "*",
    respCode.slice(0, 1) + "x",
    respCode.slice(0, 1) + "**",
    respCode.slice(0, 1) + "xx",

    "*",
    "x",
    "***",
    "xxx",
  ];
  if (startsWith(respCode, "4") || startsWith(respCode, "5")) {
    attrPossibilities.push("error");
  }

  for (const p of attrPossibilities) {
    const attr = attrPrefix + p;
    const attrValue = api.getClosestAttributeValue(elt, attr);
    if (attrValue) {
      if (attrValue === "this") {
        return api.findThisElement(elt, attr);
      } else {
        return api.querySelectorExt(elt, attrValue);
      }
    }
  }

  return null;
}

htmx.registerExtension("response-targets", {
  init(apiRef: any) {
    api = apiRef;

    if (config.responseTargetPrefersExisting === undefined) {
      config.responseTargetPrefersExisting = false;
    }
    if (config.responseTargetPrefersRetargetHeader === undefined) {
      config.responseTargetPrefersRetargetHeader = true;
    }
  },

  // htmx 4: `htmx:before:swap` extension hooks receive `{ctx, tasks}`.
  // Retargeting happens by mutating the main task's `target`, not by setting
  // `detail.target`/`detail.shouldSwap` (those are htmx 2 fields and are
  // ignored by the core in htmx 4).
  htmx_before_swap(_elt: HTMLElement, detail: any) {
    const ctx = detail?.ctx;
    const tasks = detail?.tasks;
    const status = ctx?.response?.status ?? 0;
    if (status === 0 || status === 200) return;
    if (!Array.isArray(tasks)) return;

    const mainTask = tasks.find((t: any) => t?.type === "main");
    if (!mainTask) return;

    if (mainTask.target) {
      if (config.responseTargetPrefersExisting) {
        return;
      }
      const headers = ctx?.response?.headers;
      const retarget =
        typeof headers?.get === "function" ? headers.get("HX-Retarget") : null;
      if (config.responseTargetPrefersRetargetHeader && retarget) {
        return;
      }
    }

    const reqElt = ctx?.sourceElement ?? ctx?.elt;
    if (!reqElt) return;

    const target = getRespCodeTarget(reqElt, status);
    if (target) {
      mainTask.target = target;
    }
  },
});
