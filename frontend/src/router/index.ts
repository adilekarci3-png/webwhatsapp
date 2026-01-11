// src/router/index.ts
import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import { useAuthStore } from "@/stores/auth";

import LoginView from "@/views/LoginView.vue";
import RegisterView from "@/views/RegisterView.vue";
import HomeView from "@/views/HomeView.vue";

const routes: RouteRecordRaw[] = [
  // Uygulama açılınca direkt login'e düşsün istiyorsan:
  { path: "/", redirect: "/" },

  { path: "/login", name: "login", component: LoginView, meta: { guestOnly: true } },
  { path: "/register", name: "register", component: RegisterView, meta: { guestOnly: true } },

  // Korumalı alan
  { path: "/app", name: "home", component: HomeView, meta: { requiresAuth: true } },

  // 404 fallback
  { path: "/:pathMatch(.*)*", redirect: "/login" },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// Auth guard
router.beforeEach((to) => {
  const auth = useAuthStore();

  const isAuthed = !!auth.accessToken; // store'unuzda token alanı bu isimde
  const requiresAuth = !!to.meta.requiresAuth;
  const guestOnly = !!to.meta.guestOnly;

  if (requiresAuth && !isAuthed) return { name: "login" };
  if (guestOnly && isAuthed) return { name: "home" };

  return true;
});

export default router;
