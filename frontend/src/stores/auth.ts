// src/stores/auth.ts
import { defineStore } from "pinia";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: "" as string,
    user: null as null | { id: string; name: string },
  }),
  getters: {
    isAuthenticated: (s) => !!s.token,
  },
  actions: {
    setSession(token: string, user: { id: string; name: string }) {
      this.token = token;
      this.user = user;
    },
    logout() {
      this.token = "";
      this.user = null;
    },
  },
});
