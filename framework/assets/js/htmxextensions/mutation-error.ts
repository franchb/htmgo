import htmx from "htmx.org";

const mutationMethods = new Set(["POST", "PUT", "PATCH", "DELETE"]);

// The v2 extension broadcast htmx:onMutationError to every element carrying
// the hx-on::onMutationError / hx-on::on-mutation-error attribute, not just
// the request element.  htmx 4's trigger-children does not fan out this
// htmgo-custom event, and descendants of the request element (e.g. a submit
// button inside a form) would otherwise never see it — so we replicate the
// v2 broadcast semantics here.
const HX_ON_ATTRS = [
  "hx-on::onMutationError",
  "hx-on::on-mutation-error",
];

function broadcast(status: number, exclude?: Element) {
  const seen = new Set<Element>();
  for (const attr of HX_ON_ATTRS) {
    const selector = `[${attr.replace(/:/g, "\\:")}]`;
    document.querySelectorAll(selector).forEach((el) => {
      if (seen.has(el)) return;
      // htmx.trigger dispatches bubbling CustomEvents, so ancestors of the
      // excluded element already receive the event via bubbling — skip them
      // here to avoid double-firing their hx-on listeners.
      if (exclude && (el === exclude || el.contains(exclude))) return;
      seen.add(el);
      htmx.trigger(el as HTMLElement, "htmx:onMutationError", { status, elt: el });
    });
  }
}

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
      broadcast(status, elt);
    }
  },
});
