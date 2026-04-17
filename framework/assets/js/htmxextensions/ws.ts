import htmx from "htmx.org";
import { removeAssociatedScripts } from "./htmgo";

const processed = new Set<string>();
export let ws: WebSocket | null = null;

htmx.registerExtension("ws", {
  init(_api: unknown) {},

  htmx_before_cleanup(elt: HTMLElement, _detail: unknown) {
    if (elt) removeAssociatedScripts(elt);
  },

  htmx_before_process(_elt: HTMLElement, _detail: unknown) {
    const elements = document.querySelectorAll("[ws-connect]");
    for (const element of Array.from(elements)) {
      const url = element.getAttribute("ws-connect");
      if (url && !processed.has(url)) {
        connectWs(element, url);
        processed.add(url);
      }
    }
  },
});

function exponentialBackoff(attempt: number, baseDelay = 100, maxDelay = 10000) {
  // Exponential backoff: baseDelay * (2 ^ attempt) with jitter
  const jitter = Math.random();
  return Math.min(baseDelay * Math.pow(2, attempt) * jitter, maxDelay);
}

function connectWs(ele: Element, url: string, attempt = 0): WebSocket | null {
  if (!url) return null;
  if (!url.startsWith("ws://") && !url.startsWith("wss://")) {
    const isSecure = window.location.protocol === "https:";
    url = (isSecure ? "wss://" : "ws://") + window.location.host + url;
  }
  console.info("connecting to ws", url);
  const socket = new WebSocket(url);
  ws = socket;

  socket.addEventListener("close", (event) => {
    htmx.trigger(ele, "htmx:wsClose", { event });
    const delay = exponentialBackoff(attempt);
    setTimeout(() => connectWs(ele, url, attempt + 1), delay);
  });
  socket.addEventListener("open", (event) => htmx.trigger(ele, "htmx:wsOpen", { event }));
  socket.addEventListener("error", (event) => htmx.trigger(ele, "htmx:wsError", { event }));
  socket.addEventListener("message", (event) => {
    htmx.trigger(ele, "htmx:wsBeforeMessage", { event });
    applyOobSwap(ele, event.data);
    htmx.trigger(ele, "htmx:wsAfterMessage", { event });
  });

  return socket;
}

/**
 * Parses the WS message text and applies OOB swaps via htmx 4's htmx.swap().
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
  // target because htmgo WS messages are entirely OOB — there is no primary
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
