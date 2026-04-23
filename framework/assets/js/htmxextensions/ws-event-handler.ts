import { ws } from "./ws";

if (typeof window !== "undefined") {
  window.addEventListener("load", () => {
    if (document.querySelector("[ws-connect]")) {
      addWsEventHandlers();
    }
  });
}

function sendWs(message: Record<string, any>) {
  if (ws != null && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message));
  }
}

function walk(node: Node, cb: (node: Node) => void) {
  cb(node);
  for (const child of Array.from(node.childNodes)) {
    walk(child, cb);
  }
}

export function addWsEventHandlers() {
  const observer = new MutationObserver(register);
  observer.observe(document.body, { childList: true, subtree: true });

  const added = new Set<string>();

  function register(mutations: MutationRecord[]) {
    for (const mutation of mutations) {
      for (const removedNode of Array.from(mutation.removedNodes)) {
        walk(removedNode, (node) => {
          if (node instanceof HTMLElement) {
            const handlerId = node.getAttribute("data-handler-id");
            if (handlerId) {
              added.delete(handlerId);
              sendWs({ id: handlerId, event: "dom-element-removed" });
            }
          }
        });
      }
    }

    const ids = new Set<string>();
    document.querySelectorAll("[data-handler-id]").forEach((element) => {
      const id = element.getAttribute("data-handler-id");
      const event = element.getAttribute("data-handler-event");

      if (id == null || event == null) {
        return;
      }

      ids.add(id);
      if (added.has(id)) {
        return;
      }
      added.add(id);
      element.addEventListener(event, () => {
        sendWs({ id, event });
      });
    });
    for (const id of Array.from(added)) {
      if (!ids.has(id)) {
        added.delete(id);
      }
    }
  }

  register([]);
}
