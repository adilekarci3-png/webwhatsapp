<template>
  <div class="auth-page">
    <form class="card" @submit.prevent="onSubmit">
      <h2>Giriş Yap</h2>

      <label>E-posta</label>
      <input v-model.trim="email" type="email" required />

      <label>Şifre</label>
      <input v-model="password" type="password" required />

      <button type="submit" :disabled="loading">
        {{ loading ? "Giriş yapılıyor..." : "Giriş Yap" }}
      </button>

      <p class="err" v-if="error">{{ error }}</p>

      <p class="hint">
        Hesabın yok mu?
        <router-link to="/register">Kayıt Ol</router-link>
      </p>
    </form>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { authApi } from "@/api/auth";

const router = useRouter();
const auth = useAuthStore();

const email = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");

async function onSubmit() {
  if (loading.value) return; // double submit guard

  error.value = "";
  loading.value = true;

  try {
    const r = await authApi.login({
      email: email.value.trim(),
      password: password.value,
    });

    auth.setSession(r.data.accessToken, r.data.user);
    await router.push("/app");
  } catch (e) {
    const data = e?.response?.data;
    error.value =
      (typeof data === "string" && data) ||
      data?.message ||
      "Giriş başarısız.";
  } finally {
    loading.value = false;
  }
}
</script>


<style scoped>
.auth-page { min-height: 100vh; display:flex; align-items:center; justify-content:center; padding:24px; }
.card { width: 360px; display:flex; flex-direction:column; gap:10px; padding:18px; border:1px solid #ddd; border-radius:12px; background:#fff; }
input { padding:10px; border:1px solid #ccc; border-radius:10px; }
button { padding:10px; border-radius:10px; border:0; cursor:pointer; }
.err { color:#b00020; }
.hint { font-size: 14px; }
</style>
