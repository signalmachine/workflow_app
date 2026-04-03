import { redirect } from "@sveltejs/kit";
import { g as getCurrentSession, A as APIClientError } from "../../../../chunks/session.js";
import { r as routes } from "../../../../chunks/routes.js";
const load = async ({ fetch }) => {
  try {
    await getCurrentSession(fetch);
    throw redirect(307, routes.home);
  } catch (error) {
    if (error instanceof APIClientError && error.status === 401) {
      return {};
    }
    throw error;
  }
};
export {
  load
};
