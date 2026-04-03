class APIClientError extends Error {
  status;
  constructor(status, message) {
    super(message);
    this.name = "APIClientError";
    this.status = status;
  }
}
async function parseResponse(response) {
  if (response.ok) {
    return await response.json();
  }
  let message = `Request failed with status ${response.status}`;
  try {
    const payload = await response.json();
    if (payload?.error) {
      message = payload.error;
    }
  } catch {
  }
  throw new APIClientError(response.status, message);
}
async function apiRequest(input, init, fetcher = fetch) {
  const response = await fetcher(input, {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      ...{},
      ...{}
    },
    ...init
  });
  return parseResponse(response);
}
function getCurrentSession(fetcher = fetch) {
  return apiRequest("/api/session", void 0, fetcher);
}
export {
  APIClientError as A,
  getCurrentSession as g
};
