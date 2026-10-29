import htmx from "htmx.org";

/**
 * Parses a WS/SSE message and applies its OOB swaps via htmx 4's htmx.swap().
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
 *
 * The message is server-rendered HTML, trusted exactly like an htmx HTTP
 * response: htmx.swap() would parse the same text if we passed it straight
 * through.  Only <script id="__eval…"> elements are executed.
 */
export function applyOobSwap(ele: Element, responseText: string) {
  // Parse into the content of a <template> owned by an inert document: no
  // scripts run, no resources load, and table rows etc. parse as in a body.
  // Assigning innerHTML (rather than parsing `<template>${text}</template>`)
  // means a "</template>" inside the message cannot end the fragment early
  // and silently drop the elements after it.
  const templateEl = document.implementation
    .createHTMLDocument("")
    .createElement("template");
  templateEl.innerHTML = responseText;
  const content = templateEl.content;

  // Extract and append __eval scripts first (htmgo-specific behaviour).
  // Use querySelectorAll so scripts nested inside OOB wrappers are still found.
  // Fragment parsing marks scripts "already started", so moving the parsed
  // node would never run it; recreate it in the live document instead,
  // keeping its attributes (and CSP nonce).
  const evalScripts = Array.from(
    content.querySelectorAll('script[id^="__eval"]'),
  ) as HTMLScriptElement[];
  for (const parsed of evalScripts) {
    const script = document.createElement("script");
    for (const attr of Array.from(parsed.attributes)) {
      script.setAttribute(attr.name, attr.value);
    }
    if (parsed.nonce) script.nonce = parsed.nonce;
    script.textContent = parsed.textContent;
    parsed.remove(); // remove from fragment before swap
    document.body.appendChild(script);
  }

  // Pass the (now __eval-stripped) HTML to htmx.swap().
  // htmx 4's swap() calls __processOOB internally, so all children carrying
  // hx-swap-oob will be handled correctly.  We use swap:'none' for the main
  // target because htmgo WS/SSE messages are entirely OOB — there is no
  // primary target to replace.
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
