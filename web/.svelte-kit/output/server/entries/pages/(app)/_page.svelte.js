import "clsx";
import { P as PageHeader } from "../../../chunks/PageHeader.js";
import { P as PlaceholderPanel } from "../../../chunks/PlaceholderPanel.js";
function _page($$renderer) {
  PageHeader($$renderer, {
    eyebrow: "Operator home",
    title: "Role-aware home shell",
    description: "This Slice 1 route establishes the Svelte home entry point, shell state, and shared session bootstrap on the same Go auth origin."
  });
  $$renderer.push(`<!----> <div class="stack svelte-h7bcrl">`);
  PlaceholderPanel($$renderer, {
    title: "Home migration placeholder",
    summary: "Milestone 13 Slice 2 will replace this foundation placeholder with dashboard and workload parity built from shared reporting snapshots instead of Go template rendering.",
    points: [
      "Session bootstrap already resolves through GET /api/session.",
      "The shell uses the approved left-sidebar and fixed top-bar structure.",
      "Route groups are in place for protected app routes and public auth routes."
    ]
  });
  $$renderer.push(`<!----></div>`);
}
export {
  _page as default
};
