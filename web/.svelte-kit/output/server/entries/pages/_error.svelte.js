import { h as head, e as escape_html, a as attr } from "../../chunks/index.js";
import { p as page } from "../../chunks/index2.js";
import { r as routes } from "../../chunks/routes.js";
function _error($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    head("1j96wlh", $$renderer2, ($$renderer3) => {
      $$renderer3.title(($$renderer4) => {
        $$renderer4.push(`<title>Application error</title>`);
      });
    });
    $$renderer2.push(`<main class="error-page svelte-1j96wlh"><p class="eyebrow">Application state</p> <h1 class="svelte-1j96wlh">${escape_html(page.status)}: ${escape_html(page.error?.message ?? "Unexpected error")}</h1> <p>The Svelte shell could not resolve the requested route or load state for it.</p> <p><a${attr("href", routes.home)}>Return to the operator shell</a></p></main>`);
  });
}
export {
  _error as default
};
