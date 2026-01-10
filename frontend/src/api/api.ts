import axios from "axios";
import { useAuthStore } from "@/stores/auth";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || "/api",
  withCredentials: true, // refresh cookie için
  headers: {
    "Content-Type": "application/json",
  },
});

let isRefreshing = false;
let queue: Array<{ resolve: () => void; reject: (e: any) => void }> = [];

api.interceptors.request.use((config) => {
  const auth = useAuthStore();

  if (auth.accessToken) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${auth.accessToken}`;
  }

  return config;
});

api.interceptors.response.use(
  (res) => res,
  async (err) => {
    const auth = useAuthStore();
    const original = err.config;

    // Sadece 401 ve daha önce retry edilmemiş ise refresh dene
    if (err.response?.status !== 401 || original?._retry) {
      return Promise.reject(err);
    }
    original._retry = true;

    // Refresh devam ediyorsa kuyrukla
    if (isRefreshing) {
      return new Promise<void>((resolve, reject) => {
        queue.push({ resolve, reject });
      }).then(() => api(original));
    }

    isRefreshing = true;
    try {
      const r = await api.post("/auth/refresh");
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
