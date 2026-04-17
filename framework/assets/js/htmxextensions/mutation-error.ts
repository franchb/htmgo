import htmx from "htmx.org";

const mutationMethods = new Set(["POST", "PUT", "PATCH", "DELETE"]);

htmx.registerExtension("mutation-error", {
  init(_api: unknown) {},

  htmx_after_request(elt: HTMLElement, detail: any) {
    const ctx = detail?.ctx;
    if (!ctx || !ctx.request) return;
    const method = (ctx.request.method || "").toUpperCase();
    if (!mutationMethods.has(method)) return;
    const status = ctx.response?.status ?? 0;
    if (status === 0 || status >= 400) {
      htmx.trigger(elt, "htmx:onMutationError", { status, elt });
    }
  },
});
