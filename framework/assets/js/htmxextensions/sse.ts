import htmx from "htmx.org";
import { removeAssociatedScripts } from "./htmgo";

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

/**
 * Parses the SSE message text and applies OOB swaps via htmx 4's htmx.swap().
 *
 * htmx 4 removed the internal api methods makeFragment / makeSettleInfo /
 * oobSwap / getAttributeValue that htmx 2 exposed.  The replacement is the
 * public htmx.swap(ctx) which accepts a plain ctx object and processes all
 * hx-swap-oob children automatically.
 *
 * The one htmgo-specific behaviour that htmx.swap() does NOT handle is the
 * appending of <script id="__eval…"> elements to document.body.  We extract
 * those from the parsed fragment ourselves before handing the rest off to
 * htmx.swap(), so they are not lost.
 */
function applyOobSwap(ele: Element, responseText: string) {
  // Parse the HTML into a document fragment so we can inspect children.
  // DOMParser is always available in the browser environments htmgo targets.
  const doc = new DOMParser().parseFromString(
    `<template>${responseText}</template>`,
    "text/html"
  );
  const templateEl = doc.querySelector("template");
  if (!templateEl) {
    // Fallback: hand the raw text to htmx.swap() with swap:'none' so OOB
    // attributes are still processed by htmx core.
    (htmx as any).swap({
      sourceElement: ele,
      target: ele,
      swap: "none",
      text: responseText,
      transition: false,
    });
    return;
  }

  // Extract and append __eval scripts first (htmgo-specific behaviour).
  // These must be appended to document.body so the browser executes them.
  const evalScripts: HTMLScriptElement[] = [];
  const content = templateEl.content;
  for (const child of Array.from(content.children)) {
    if (
      child.tagName === "SCRIPT" &&
      child.id.startsWith("__eval")
    ) {
      evalScripts.push(child as HTMLScriptElement);
    }
  }
  for (const script of evalScripts) {
    script.remove(); // remove from fragment before swap
    document.body.appendChild(script);
  }

  // Pass the (now __eval-stripped) HTML to htmx.swap().
  // htmx 4's swap() calls __processOOB internally, so all children carrying
  // hx-swap-oob will be handled correctly.  We use swap:'none' for the main
  // target because htmgo SSE messages are entirely OOB — there is no primary
  // target to replace.
  (htmx as any).swap({
    sourceElement: ele,
    target: ele,
    swap: "none",
    // Serialise fragment back to string for htmx.swap()'s __makeFragment.
    text: Array.from(content.children)
      .map((c) => c.outerHTML)
      .join(""),
    transition: false,
  });
}
