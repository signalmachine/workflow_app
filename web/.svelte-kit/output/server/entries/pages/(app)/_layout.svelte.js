import { a as attr, b as attr_class, c as ensure_array_like, e as escape_html, h as head, d as derived } from "../../../chunks/index.js";
import "@sveltejs/kit/internal";
import "../../../chunks/url.js";
import "../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../chunks/root.js";
import "../../../chunks/exports.js";
import { p as page } from "../../../chunks/index2.js";
import { F as FlashBanner } from "../../../chunks/FlashBanner.js";
import "clsx";
import { r as routes } from "../../../chunks/routes.js";
function SideNav($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { currentPath, isOpen } = $$props;
    const primaryItems = [
      {
        label: "Home",
        href: routes.home,
        match: (pathname) => pathname === routes.home
      },
      {
        label: "Operations",
        href: routes.operations,
        match: (pathname) => pathname.startsWith(routes.operations)
      },
      {
        label: "Review",
        href: routes.review,
        match: (pathname) => pathname.startsWith(routes.review)
      },
      {
        label: "Inventory",
        href: routes.inventory,
        match: (pathname) => pathname.startsWith(routes.inventory)
      }
    ];
    const utilityItems = [
      {
        label: "Settings",
        href: routes.settings,
        match: (pathname) => pathname.startsWith(routes.settings)
      },
      {
        label: "Admin",
        href: routes.admin,
        match: (pathname) => pathname.startsWith(routes.admin)
      }
    ];
    function active(item) {
      return item.match(currentPath);
    }
    $$renderer2.push(`<div${attr("aria-hidden", !isOpen)}${attr_class("sidebar-backdrop svelte-kvbt81", void 0, { "open": isOpen })}></div> <aside${attr_class("sidebar svelte-kvbt81", void 0, { "open": isOpen })}><nav><div class="nav-group svelte-kvbt81"><!--[-->`);
    const each_array = ensure_array_like(primaryItems);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let item = each_array[$$index];
      $$renderer2.push(`<a${attr("aria-current", active(item) ? "page" : void 0)}${attr("href", item.href)}${attr_class("svelte-kvbt81", void 0, { "active": active(item) })}>${escape_html(item.label)}</a>`);
    }
    $$renderer2.push(`<!--]--></div> <div class="nav-divider svelte-kvbt81"></div> <div class="nav-group svelte-kvbt81"><!--[-->`);
    const each_array_1 = ensure_array_like(utilityItems);
    for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
      let item = each_array_1[$$index_1];
      $$renderer2.push(`<a${attr("aria-current", active(item) ? "page" : void 0)}${attr("href", item.href)}${attr_class("svelte-kvbt81", void 0, { "active": active(item) })}>${escape_html(item.label)}</a>`);
    }
    $$renderer2.push(`<!--]--> <div class="nav-subgroup svelte-kvbt81"><a${attr("href", routes.adminAccounting)} class="svelte-kvbt81">Accounting setup</a> <a${attr("href", routes.adminParties)} class="svelte-kvbt81">Party setup</a> <a${attr("href", routes.adminAccess)} class="svelte-kvbt81">Access controls</a> <a${attr("href", routes.adminInventory)} class="svelte-kvbt81">Inventory setup</a></div></div></nav></aside>`);
  });
}
function TopBar($$renderer, $$props) {
  let { userDisplayName, orgName } = $$props;
  $$renderer.push(`<header class="topbar svelte-kbslt5"><div class="brand-row svelte-kbslt5"><button aria-label="Toggle navigation" class="menu-button svelte-kbslt5" type="button"><span class="svelte-kbslt5"></span> <span class="svelte-kbslt5"></span> <span class="svelte-kbslt5"></span></button> <div><div class="brand-mark svelte-kbslt5">Workflow App</div> <div class="org-name svelte-kbslt5">${escape_html(orgName)}</div></div></div> <div class="user-row svelte-kbslt5"><div class="user-copy svelte-kbslt5"><div>${escape_html(userDisplayName)}</div> <div class="role-hint svelte-kbslt5">Operator shell</div></div> <button class="logout-button svelte-kbslt5" type="button">Sign out</button></div></header>`);
}
function AppShell($$renderer, $$props) {
  let { children, currentPath, orgName, userDisplayName } = $$props;
  let navOpen = false;
  TopBar($$renderer, { orgName, userDisplayName });
  $$renderer.push(`<!----> `);
  SideNav($$renderer, { currentPath, isOpen: navOpen });
  $$renderer.push(`<!----> <div class="shell svelte-1il99g0"><main class="content svelte-1il99g0">`);
  children($$renderer);
  $$renderer.push(`<!----></main></div>`);
}
function _layout($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, children } = $$props;
    let currentPath = derived(() => page.url.pathname);
    let notice = derived(() => page.url.searchParams.get("notice") ?? "");
    let error = derived(() => page.url.searchParams.get("error") ?? "");
    head("1v2axqk", $$renderer2, ($$renderer3) => {
      $$renderer3.title(($$renderer4) => {
        $$renderer4.push(`<title>workflow_app</title>`);
      });
    });
    AppShell($$renderer2, {
      currentPath: currentPath(),
      orgName: data.session.org_name,
      userDisplayName: data.session.user_display_name,
      children: ($$renderer3) => {
        if (notice()) {
          $$renderer3.push("<!--[0-->");
          FlashBanner($$renderer3, { kind: "notice", message: notice() });
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        if (error()) {
          $$renderer3.push("<!--[0-->");
          FlashBanner($$renderer3, { kind: "error", message: error() });
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        children($$renderer3);
        $$renderer3.push(`<!---->`);
      }
    });
  });
}
export {
  _layout as default
};
