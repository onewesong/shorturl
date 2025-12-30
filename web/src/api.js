async function request(path, options = {}) {
  const res = await fetch(path, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  if (res.status === 204) return null;

  const contentType = res.headers.get("content-type") || "";
  const data = contentType.includes("application/json") ? await res.json() : await res.text();

  if (!res.ok) {
    const message = typeof data === "string" ? data : data?.error || "请求失败";
    const err = new Error(message);
    err.status = res.status;
    err.data = data;
    throw err;
  }
  return data;
}

export function apiLogin(username, password) {
  return request("/api/login", { method: "POST", body: JSON.stringify({ username, password }) });
}

export function apiLogout() {
  return request("/api/logout", { method: "POST" });
}

export function apiMe() {
  return request("/api/me");
}

export function apiListLinks() {
  return request("/api/links");
}

export function apiCreateLink(payload) {
  return request("/api/links", { method: "POST", body: JSON.stringify(payload) });
}

export function apiGetLink(id) {
  return request(`/api/links/${id}`);
}

export function apiUpdateLink(id, payload) {
  return request(`/api/links/${id}`, { method: "PUT", body: JSON.stringify(payload) });
}

