import htmx from "htmx.org";
import { removeAssociatedScripts } from "./htmgo";
import { applyOobSwap } from "./oob";

const connections = new WeakMap<Element, EventSource>();
const processedUrls = new Map<string, Element>();

htmx.registerExtension("sse", {
  init(_api: unknown) {},

  htmx_before_cleanup(elt: HTMLElement, _detail: unknown) {
    if (!elt) return;
    removeAssociatedScripts(elt);
    const es = connections.get(elt);
    if (es) {
      es.close();
      connections.delete(elt);
      const url = elt.getAttribute("sse-connect");
      if (url && processedUrls.get(url) === elt) processedUrls.delete(url);
    }
  },

  htmx_before_process(_elt: HTMLElement, _detail: unknown) {
    const elements = document.querySelectorAll("[sse-connect]");
    for (const element of Array.from(elements)) {
      const url = element.getAttribute("sse-connect");
      if (url && !processedUrls.has(url)) {
        const es = connectEventSource(element, url);
        if (es) {
          connections.set(element, es);
          processedUrls.set(url, element);
        }
      }
    }
  },
});

function connectEventSource(ele: Element, url: string): EventSource | undefined {
  if (!url) return undefined;
  console.info("Connecting to EventSource", url);
  htmx.trigger(ele, "htmx:before:sse:connection", { url });
  const eventSource = new EventSource(url);

  // A server-sent `event: close` frame is terminal — EventSource has no native
  // "close" event, so this handler only fires when the server sends one. Treat
  // it the same as the onerror-CLOSED branch: close the source and drop the
  // element from the connection maps so the URL can be reconnected later.
  eventSource.addEventListener("close", (event) => {
    eventSource.close();
    htmx.trigger(ele, "htmx:sse:close", { event });
    connections.delete(ele);
    if (processedUrls.get(url) === ele) processedUrls.delete(url);
  });

  eventSource.onopen = (event) =>
    htmx.trigger(ele, "htmx:after:sse:connection", { event });

  eventSource.onerror = (event) => {
    htmx.trigger(ele, "htmx:sse:error", { event });
    if (eventSource.readyState === EventSource.CLOSED) {
      htmx.trigger(ele, "htmx:sse:close", { event });
      connections.delete(ele);
      if (processedUrls.get(url) === ele) processedUrls.delete(url);
    }
  };

  eventSource.onmessage = (event) => {
    htmx.trigger(ele, "htmx:before:sse:message", { event });
    applyOobSwap(ele, event.data);
    htmx.trigger(ele, "htmx:after:sse:message", { event });
  };

  return eventSource;
}
