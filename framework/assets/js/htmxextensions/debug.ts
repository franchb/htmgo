import htmx from "htmx.org";

htmx.registerExtension("debug", {
  init(_api: unknown) {
    (htmx.config as any).logAll = true;
  },
});
