import htmx from "htmx.org";

const evalFuncRegex = /__eval_[A-Za-z0-9]+\([a-z]+\)/gm;

htmx.registerExtension("htmgo", {
  init(_api: unknown) {
    // no-op; retained for htmx 4 registerExtension API shape
  },

  htmx_before_cleanup(elt: HTMLElement, _detail: unknown) {
    if (elt) removeAssociatedScripts(elt);
  },

  htmx_after_init(elt: HTMLElement, _detail: unknown) {
    if (elt) invokeOnLoad(elt);
  },
});

// Browser doesn't support onload for all elements, so we manually trigger it
// (useful for locality of behavior).
function invokeOnLoad(element: Element) {
  if (element == null || !(element instanceof HTMLElement)) return;
  const ignored = ["SCRIPT", "LINK", "STYLE", "META", "BASE", "TITLE", "HEAD", "HTML", "BODY"];
  if (!ignored.includes(element.tagName)) {
    if (element.hasAttribute("onload")) {
      element.onload!(new Event("load"));
    }
  }
  element.querySelectorAll("[onload]").forEach(invokeOnLoad);
}

export function removeAssociatedScripts(element: HTMLElement) {
  const attributes = Array.from(element.attributes);
  for (const attribute of attributes) {
    const matches = attribute.value.match(evalFuncRegex) || [];
    for (const match of matches) {
      const id = match.replace("()", "").replace("(this)", "").replace(";", "");
      const ele = document.getElementById(id);
      if (ele && ele.tagName === "SCRIPT") {
        console.debug("removing associated script with id", id);
        ele.remove();
      }
    }
  }
}
