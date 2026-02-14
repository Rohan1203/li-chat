import { CONFIG } from "./config.js";
import { logout } from "./auth.js";

export async function apiFetch(url, options = {}) {
  const token = localStorage.getItem("token");

  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(CONFIG.API_BASE + url, {
    ...options,
    headers,
  });

  if (res.status === 401) {
    logout();
    throw new Error("Unauthorized");
  }

  return res;
}
