import { api } from "@/api/api";

export const authApi = {
  register(payload: any) {
    return api.post("/auth/register", payload);
  },
  login(payload: { email: string; password: string }) {
    return api.post("/auth/login", payload);
  },
  logout() {
    return api.post("/auth/logout");
  },
  // me: backend’de yoksa şimdilik kapalı
  // me() { return api.get("/auth/me"); },
};
