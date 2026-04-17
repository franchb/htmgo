import htmx from "htmx.org";

let lastVersion = "";

function livereloadURL(): string | null {
  const meta = document.querySelector('meta[name="htmgo-livereload"]');
  if (!meta) return null;
  return (meta.getAttribute("content") || "").trim() || "/dev/livereload";
}

htmx.registerExtension("livereload", {
  init(_api: unknown) {
    const url = livereloadURL();
    if (!url) return;

    console.info("livereload extension initialized:", url);
    const eventSource = new EventSource(url);
    eventSource.onmessage = (event: MessageEvent) => {
      const message = event.data;
      if (lastVersion === "") {
        lastVersion = message;
        return;
      }
      if (lastVersion !== message) {
        lastVersion = message;
        window.location.reload();
      }
    };
    eventSource.onerror = (error: Event) => {
      console.error("EventSource error:", error);
    };
  },
});
