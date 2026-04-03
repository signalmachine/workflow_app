import { redirect } from "@sveltejs/kit";
import { g as getCurrentSession, A as APIClientError } from "../../../chunks/session.js";
import { r as routes } from "../../../chunks/routes.js";
const load = async ({ fetch, url }) => {
  try {
    const session = await getCurrentSession(fetch);
    return { session };
  } catch (error) {
    if (error instanceof APIClientError && error.status === 401) {
      const next = encodeURIComponent(url.pathname + url.search);
      throw redirect(307, `${routes.login}?next=${next}`);
    }
    throw error;
  }
};
export {
  load
};
