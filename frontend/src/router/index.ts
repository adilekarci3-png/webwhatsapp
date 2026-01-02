import { createRouter, createWebHistory, type RouteLocationNormalized } from "vue-router";
import LoginView from "@/views/LoginView.vue";
import RegisterView from "@/views/RegisterView.vue";
import HomeView from "@/views/HomeView.vue";
import { useAuthStore } from "@/stores/auth";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/login", name: "login", component: LoginView },
    { path: "/register", name: "register", component: RegisterView },
    { path: "/", name: "home", component: HomeView, meta: { requiresAuth: true } },
  ],
});

router.beforeEach((to: RouteLocationNormalized) => {
  const auth = useAuthStore();
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: "login" };
  }
  if ((to.name === "login" || to.name === "register") && auth.isAuthenticated) {
    return { name: "home" };
  }
});

export default router;
