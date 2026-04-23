import htmx from "htmx.org";

function kebabEventName(str: string): string {
  return str.replace(/([a-z0-9])([A-Z])/g, "$1-$2").toLowerCase();
}

const ignoredEvents = new Set([
  "htmx:before:process",
  "htmx:after:init",
  "htmx:config:request",
  "htmx:config:response",
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
          // Dispatch with the original event name (preserving any camelCase
          // segments like `htmx:after:viewTransition`) — reconstructing from
          // the kebab-cased attribute would lose that casing and miss listeners.
          const newEvent = makeEvent(name, detail);
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
      // Explicit enumeration of htmx 4 events (verified against
      // htmx.org/dist/htmx.js + ext/hx-sse.js + ext/hx-ws.js 4.0.0-beta2).
      for (const name of [
        "htmx:abort",
        "htmx:after:cleanup",
        "htmx:after:history:push",
        "htmx:after:history:replace",
        "htmx:after:history:update",
        "htmx:after:init",
        "htmx:after:process",
        "htmx:after:request",
        "htmx:after:settle",
        "htmx:after:swap",
        "htmx:after:viewTransition",
        "htmx:before:cleanup",
        "htmx:before:history:restore",
        "htmx:before:history:update",
        "htmx:before:init",
        "htmx:before:process",
        "htmx:before:request",
        "htmx:before:response",
        "htmx:before:settle",
        "htmx:before:swap",
        "htmx:before:viewTransition",
        "htmx:config:request",
        "htmx:confirm",
        "htmx:error",
        "htmx:finally:request",
        "htmx:load",
        "htmx:prompt",
        // SSE extension (hx-sse.js)
        "htmx:after:sse:connection",
        "htmx:after:sse:message",
        "htmx:before:sse:connection",
        "htmx:before:sse:message",
        "htmx:sse:close",
        "htmx:sse:error",
        // WS extension (hx-ws.js)
        "htmx:after:ws:connection",
        "htmx:after:ws:message",
        "htmx:before:ws:connection",
        "htmx:before:ws:message",
        "htmx:ws:close",
        "htmx:ws:error",
      ]) {
        document.addEventListener(name, handler);
      }
      (document as any)[marker] = true;
    }
  },
});
