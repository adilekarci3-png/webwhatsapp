// api.js
import axios from "axios";
import { useAuthStore } from "@/stores/auth";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE,
  withCredentials: true, // refresh cookie için şart
});

let isRefreshing = false;
let queue = [];

api.interceptors.request.use((config) => {
  const auth = useAuthStore();
  if (auth.accessToken) {
    config.headers.Authorization = `Bearer ${auth.accessToken}`;
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  async (err) => {
    const auth = useAuthStore();
    const original = err.config;

    if (err.response?.status !== 401 || original._retry) {
      return Promise.reject(err);
    }

    original._retry = true;

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        queue.push({ resolve, reject });
      }).then(() => api(original));
    }

    isRefreshing = true;
    try {
      const r = await api.post("/auth/refresh"); // cookie ile
      auth.accessToken = r.data.accessToken;

      queue.forEach((p) => p.resolve());
      queue = [];
      return api(original);
    } catch (e) {
      queue.forEach((p) => p.reject(e));
      queue = [];
      auth.logout();
      return Promise.reject(e);
    } finally {
      isRefreshing = false;
    }
  }
);
