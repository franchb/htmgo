import htmx from "htmx.org";

function kebabEventName(str: string): string {
  return str.replace(/([a-z0-9])([A-Z])/g, "$1-$2").toLowerCase();
}

const ignoredEvents = new Set([
  "htmx:before:process",
  "htmx:after:init",
  "htmx:config:request",
  "htmx:config:response",
  "htmx:error",
  // legacy names (in case any consumer still dispatches them)
  "htmx:beforeProcessNode",
  "htmx:afterProcessNode",
  "htmx:configRequest",
  "htmx:configResponse",
  "htmx:responseError",
]);

function makeEvent(eventName: string, detail: any): CustomEvent {
  if (window.CustomEvent && typeof window.CustomEvent === "function") {
    // TODO: `composed: true` here is a hack to make global event handlers work with events in shadow DOM
    return new CustomEvent(eventName, {
      bubbles: false,
      cancelable: true,
      composed: true,
      detail,
    });
  }
  const evt = document.createEvent("CustomEvent");
  evt.initCustomEvent(eventName, true, true, detail);
  return evt;
}

function triggerChildren(
  target: HTMLElement,
  name: string,
  event: CustomEvent,
  triggered: Set<HTMLElement>,
) {
  if (ignoredEvents.has(name)) return;
  if (!target || !target.children) return;
  Array.from(target.children).forEach((child) => {
    const kebab = kebabEventName(name);
    const attrName = kebab.replace("htmx:", "hx-on::");
    if (!triggered.has(child as HTMLElement)) {
      if (child.hasAttribute(attrName)) {
        setTimeout(() => {
          const detail = {
            ...(event.detail ?? {}),
            target: child,
            meta: "trigger-children",
          };
          const newEvent = makeEvent(attrName.replace("hx-on::", "htmx:"), detail);
          child.dispatchEvent(newEvent);
          triggered.add(child as HTMLElement);
        }, 1);
      }
      if (child.children) {
        triggerChildren(child as HTMLElement, name, event, triggered);
      }
    }
  });
}

htmx.registerExtension("trigger-children", {
  init(_api: unknown) {
    // htmx 4 extension API has no catch-all event hook, so we listen at the
    // document level for every bubbling htmx event and fan it out to children.
    const handler = (evt: Event) => {
      if (!(evt instanceof CustomEvent)) return;
      if (evt.detail?.meta === "trigger-children") return;
      if (!evt.type.startsWith("htmx:")) return;
      const triggered = new Set<HTMLElement>();
      const target =
        (evt.target as HTMLElement) ?? (evt.detail?.target as HTMLElement);
      triggerChildren(target, evt.type, evt, triggered);
    };

    // Register once; use a sentinel to avoid double-registration across re-init.
    const marker = "__htmgo_trigger_children_installed__" as const;
    if (!(document as any)[marker]) {
      // "htmx:*" wildcard is not standard — left as harmless future-compat hint.
      document.addEventListener("htmx:*", handler as any);
      // Explicit enumeration is the reliable mechanism for current browsers.
      for (const name of [
        "htmx:before:init",
        "htmx:after:init",
        "htmx:before:request",
        "htmx:after:request",
        "htmx:before:swap",
        "htmx:after:swap",
        "htmx:config:request",
        "htmx:config:response",
        "htmx:before:cleanup",
        "htmx:before:history:update",
        "htmx:before:history:restore",
        "htmx:after:history:push",
        "htmx:after:history:replace",
        "htmx:before:viewTransition",
        "htmx:error",
        "htmx:abort",
        "htmx:confirm",
        "htmx:prompt",
        "htmx:sseOpen",
        "htmx:sseConnecting",
        "htmx:sseClose",
        "htmx:sseError",
        "htmx:sseBeforeMessage",
        "htmx:sseAfterMessage",
      ]) {
        document.addEventListener(name, handler);
      }
      (document as any)[marker] = true;
    }
  },
});
