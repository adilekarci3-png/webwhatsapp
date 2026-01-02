import { defineStore } from "pinia";

type User = { id: string; name: string };

export const useAuthStore = defineStore("auth", {
  state: () => ({
    accessToken: "" as string,
    user: null as User | null,
  }),
  getters: {
    isAuthenticated: (s) => !!s.accessToken,
  },
  actions: {
    setSession(accessToken: string, user: User) {
      this.accessToken = accessToken;
      this.user = user;
    },
    logout() {
      this.accessToken = "";
      this.user = null;
    },
  },
});
