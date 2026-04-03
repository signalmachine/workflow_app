import { h as head, a as attr, e as escape_html, d as derived } from "../../../../chunks/index.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/url.js";
import "../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../chunks/root.js";
import "../../../../chunks/exports.js";
import { p as page } from "../../../../chunks/index2.js";
import { F as FlashBanner } from "../../../../chunks/FlashBanner.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let orgSlug = "";
    let email = "";
    let password = "";
    let submitting = false;
    let notice = derived(() => page.url.searchParams.get("notice") ?? "");
    head("iv8lg3", $$renderer2, ($$renderer3) => {
      $$renderer3.title(($$renderer4) => {
        $$renderer4.push(`<title>workflow_app sign in</title>`);
      });
    });
    $$renderer2.push(`<main class="login-shell svelte-iv8lg3"><section class="login-panel svelte-iv8lg3"><p class="eyebrow">Milestone 13 Slice 1</p> <h1 class="svelte-iv8lg3">Operator sign-in</h1> <p class="intro svelte-iv8lg3">This Svelte login surface uses the existing cookie-auth session seam at <span class="mono">POST /api/session/login</span>.</p> `);
    if (notice()) {
      $$renderer2.push("<!--[0-->");
      FlashBanner($$renderer2, { kind: "notice", message: notice() });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <form class="login-form svelte-iv8lg3"><label class="svelte-iv8lg3"><span class="svelte-iv8lg3">Org slug</span> <input${attr("value", orgSlug)} autocomplete="organization" placeholder="north-harbor" required="" class="svelte-iv8lg3"/></label> <label class="svelte-iv8lg3"><span class="svelte-iv8lg3">Email</span> <input${attr("value", email)} autocomplete="email" placeholder="admin@example.com" required="" type="email" class="svelte-iv8lg3"/></label> <label class="svelte-iv8lg3"><span class="svelte-iv8lg3">Password</span> <input${attr("value", password)} autocomplete="current-password" required="" type="password" class="svelte-iv8lg3"/></label> <button${attr("disabled", submitting, true)} type="submit" class="svelte-iv8lg3">${escape_html("Sign in")}</button></form></section></main>`);
  });
}
export {
  _page as default
};
