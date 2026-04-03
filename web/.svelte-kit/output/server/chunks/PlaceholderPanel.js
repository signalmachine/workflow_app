import { b as attr_class, e as escape_html, c as ensure_array_like } from "./index.js";
function SurfaceCard($$renderer, $$props) {
  let { children, tone = "default" } = $$props;
  $$renderer.push(`<section${attr_class("surface-card svelte-txxlo", void 0, { "muted": tone === "muted" })}>`);
  children($$renderer);
  $$renderer.push(`<!----></section>`);
}
function PlaceholderPanel($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { title, summary, points = [] } = $$props;
    SurfaceCard($$renderer2, {
      children: ($$renderer3) => {
        $$renderer3.push(`<div class="placeholder svelte-13sx0aq"><p class="eyebrow svelte-13sx0aq">Milestone 13 Slice 1</p> <h2 class="svelte-13sx0aq">${escape_html(title)}</h2> <p class="svelte-13sx0aq">${escape_html(summary)}</p> `);
        if (points.length > 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<ul class="svelte-13sx0aq"><!--[-->`);
          const each_array = ensure_array_like(points);
          for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
            let point = each_array[$$index];
            $$renderer3.push(`<li class="svelte-13sx0aq">${escape_html(point)}</li>`);
          }
          $$renderer3.push(`<!--]--></ul>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--></div>`);
      }
    });
  });
}
export {
  PlaceholderPanel as P
};
