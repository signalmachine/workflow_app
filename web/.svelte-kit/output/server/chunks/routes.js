import { b as base } from "./server.js";
import "./url.js";
import "@sveltejs/kit/internal/server";
import "./root.js";
function withBase(path) {
  if (path === "/") {
    return base || "/";
  }
  return `${base}${path}`;
}
const routes = {
  login: withBase("/login"),
  home: withBase("/"),
  operations: withBase("/operations"),
  review: withBase("/review"),
  inventory: withBase("/inventory"),
  settings: withBase("/settings"),
  admin: withBase("/admin"),
  adminAccounting: withBase("/admin/accounting"),
  adminParties: withBase("/admin/parties"),
  adminAccess: withBase("/admin/access"),
  adminInventory: withBase("/admin/inventory")
};
export {
  routes as r
};
