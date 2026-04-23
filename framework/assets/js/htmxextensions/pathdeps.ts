import htmx from "htmx.org";

const mutationMethods = new Set(["POST", "PUT", "PATCH", "DELETE"]);

/**
 * Segment-by-segment wildcard matching.
 * Each segment of `pathSpec` must match the corresponding segment of `url`,
 * where `*` matches any single segment.
 * Returns true when all segments of `pathSpec` are consumed and matched.
 * "ignore" always returns false (opt-out sentinel).
 */
function dependsOn(pathSpec: string, url: string): boolean {
  if (pathSpec === "ignore") {
    return false;
  }
  const dependencyPath = pathSpec.split("/");
  const urlPath = url.split("/");
  for (let i = 0; i < urlPath.length; i++) {
    const dependencyElement = dependencyPath.shift();
    const pathElement = urlPath[i];
    if (dependencyElement !== pathElement && dependencyElement !== "*") {
      return false;
    }
    if (
      dependencyPath.length === 0 ||
      (dependencyPath.length === 1 && dependencyPath[0] === "")
    ) {
      return true;
    }
  }
  return false;
}

htmx.registerExtension("path-deps", {
  init(_api: unknown) {},

  htmx_after_request(_elt: HTMLElement, detail: any) {
    const ctx = detail?.ctx;
    if (!ctx || !ctx.request) return;
    const method = (ctx.request.method || "").toUpperCase();
    if (!mutationMethods.has(method)) return;
    const path = ctx.request.action || "";
    document.querySelectorAll("[path-deps]").forEach((el) => {
      const dep = el.getAttribute("path-deps") || "";
      if (dependsOn(dep, path)) htmx.trigger(el, "path-deps");
    });
  },
});
