import htmx from "htmx.org";
import { removeAssociatedScripts } from "./htmgo";
import { applyOobSwap } from "./oob";

const processed = new Set<string>();
const sockets = new WeakMap<Element, { socket: WebSocket; url: string }>();
export let ws: WebSocket | null = null;

htmx.registerExtension("ws", {
  init(_api: unknown) {},

  htmx_before_cleanup(elt: HTMLElement, _detail: unknown) {
    if (!elt) return;
    removeAssociatedScripts(elt);
    const entry = sockets.get(elt);
    if (entry) {
      processed.delete(entry.url);
      sockets.delete(elt);
      try {
        entry.socket.close(1000, "element removed");
      } catch {
        /* ignore close errors on already-closed sockets */
      }
    }
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

const MAX_RECONNECT_ATTEMPTS = 10;

function exponentialBackoff(attempt: number, baseDelay = 100, maxDelay = 10000) {
  // Exponential backoff with half-jitter: base * 2^attempt * (0.5 + rand/2)
  // ensures a jitter floor of 0.5 so the first retry can never hot-loop.
  const base = Math.min(baseDelay * Math.pow(2, attempt), maxDelay);
  return base * (0.5 + Math.random() * 0.5);
}

function connectWs(ele: Element, url: string, attempt = 0): WebSocket | null {
  if (!url) return null;
  if (!url.startsWith("ws://") && !url.startsWith("wss://")) {
    const isSecure = window.location.protocol === "https:";
    url = (isSecure ? "wss://" : "ws://") + window.location.host + url;
  }
  console.info("connecting to ws", url);
  htmx.trigger(ele, "htmx:before:ws:connection", { url });
  const socket = new WebSocket(url);
  ws = socket;
  sockets.set(ele, { socket, url });

  // Event names match htmx 4 colon form (see official hx-ws.js and the
  // trigger-children event list).  Listeners register as
  //   hx-on::ws:close / hx-on::after:ws:message / OnEvent("htmx:ws:close", …).
  socket.addEventListener("close", (event) => {
    htmx.trigger(ele, "htmx:ws:close", { event });
    // Don't reconnect on clean close, when the element has been detached,
    // or once we've exhausted our retry budget.
    if (event.wasClean || !ele.isConnected || attempt >= MAX_RECONNECT_ATTEMPTS) {
      return;
    }
    const delay = exponentialBackoff(attempt);
    setTimeout(() => connectWs(ele, url, attempt + 1), delay);
  });
  socket.addEventListener("open", (event) => htmx.trigger(ele, "htmx:after:ws:connection", { event }));
  socket.addEventListener("error", (event) => htmx.trigger(ele, "htmx:ws:error", { event }));
  socket.addEventListener("message", (event) => {
    htmx.trigger(ele, "htmx:before:ws:message", { event });
    applyOobSwap(ele, event.data);
    htmx.trigger(ele, "htmx:after:ws:message", { event });
  });

  return socket;
}
