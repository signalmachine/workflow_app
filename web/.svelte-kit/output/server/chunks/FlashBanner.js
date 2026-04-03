import { b as attr_class, e as escape_html } from "./index.js";
function FlashBanner($$renderer, $$props) {
  let { kind = "notice", message } = $$props;
  if (message) {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<div${attr_class("flash svelte-1flphlw", void 0, { "error": kind === "error", "notice": kind !== "error" })}>${escape_html(message)}</div>`);
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]-->`);
}
export {
  FlashBanner as F
};
