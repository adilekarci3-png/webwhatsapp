// stores/auth.ts
import { defineStore } from "pinia";

type User = { id: string; name: string };

export const useAuthStore = defineStore("auth", {
  state: () => ({
    accessToken: localStorage.getItem("accessToken") || "",
    user: (localStorage.getItem("user")
      ? JSON.parse(localStorage.getItem("user") as string)
      : null) as User | null,
  }),
  getters: {
    isAuthenticated: (s) => !!s.accessToken,
  },
  actions: {
    setSession(accessToken: string, user: User) {
      this.accessToken = accessToken;
      this.user = user;
      localStorage.setItem("accessToken", accessToken);
      localStorage.setItem("user", JSON.stringify(user));
    },
    logout() {
      this.accessToken = "";
      this.user = null;
      localStorage.removeItem("accessToken");
      localStorage.removeItem("user");
    },
  },
});

