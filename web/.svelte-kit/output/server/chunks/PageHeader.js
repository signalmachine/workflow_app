import { e as escape_html } from "./index.js";
import "clsx";
function PageHeader($$renderer, $$props) {
  let { eyebrow, title, description } = $$props;
  $$renderer.push(`<header class="page-header svelte-vk4cqd">`);
  if (eyebrow) {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<p class="eyebrow">${escape_html(eyebrow)}</p>`);
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--> <h1 class="svelte-vk4cqd">${escape_html(title)}</h1> `);
  if (description) {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<p class="description svelte-vk4cqd">${escape_html(description)}</p>`);
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--></header>`);
}
export {
  PageHeader as P
};
